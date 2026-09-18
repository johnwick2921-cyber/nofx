package trader

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/market"
	"nofx/store"
)

// W-WRITE-TIME-FEASIBILITY (2026-09-18) — call-site tests.
//
// These drive runPlannerReadCoreWithFactsGrades (the real planner call site,
// with prompt visibility) with a plan whose arm the gate-at-arm chain would
// refuse, through a stubbed FuturesBarsProvider that yields a known 5m ATR so
// the min-SL floor is deterministic.

// feasBars returns synthetic 5m bars with a true range of exactly 8 points each
// → ExportCalculateATR(14) = 8.0, so the min-SL floor is 1.5 × 8.0 = 12.0.
func feasBars(symbol, tf string, n int) []market.Kline {
	out := make([]market.Kline, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, market.Kline{Open: 100, High: 108, Low: 100, Close: 108})
	}
	return out
}

func feasStubBars(t *testing.T) {
	t.Helper()
	old := market.FuturesBarsProvider
	market.FuturesBarsProvider = feasBars
	t.Cleanup(func() { market.FuturesBarsProvider = old })
}

// infeasibleFeasPlanJSON: entry 15550 / stop 15540 → stop distance 10.00 <
// 12.00 floor (1.5 × ATR5m 8.0). The feasible variant widens the stop to 15530.
const infeasibleFeasPlanJSON = `{
  "reasoning": "Balance below PDH; fade edges, long the reclaim.",
  "bias": {"direction": "long", "conviction": "medium", "flip_condition": "2x5m < 15480"},
  "levels": [
    {"price": 15480, "label": "PWL", "grade": "A", "instruction": "fade"},
    {"price": 15520, "label": "RN 15525", "grade": "B", "instruction": "fade"},
    {"price": 15620, "label": "PDH", "grade": "A", "instruction": "fade"}
  ],
  "scenarios": [{"id": "S1", "trigger": "sweep 15480 reclaim", "condition": "sweep_reclaim", "direction": "long", "target_chain": [15550, 15620], "invalid": "2x5m<15470", "quality": "A", "confirm":{"rule":"touch","ref_price":15480,"side":"below"},"economics":{"entry_zone":[15480,15480],"geometry":{"entry":15550,"stop":15540,"target":15620},"first_obstacle":{"price":15560,"level":"fixture reference","family":"reference","response":"pass_through"},"r_to_obstacle":1.0,"r_to_arm_target":7.0},"arm":{"enabled":true,"entry":15550,"stop":15540,"target":15620,"wait_confirm":true}}],
  "no_trade": ["first 5m"],
  "death_condition": "acceptance above 15620",
  "death": {"price": 15620, "side": "above", "rule": "2x5m"},
  "flip": {"price": 15480, "side": "below", "rule": "2x5m", "flip_to": "short"},
  "day_type": "balance"
}`

func feasPlannerTrader(t *testing.T, writeFeas *bool) *AutoTrader {
	t.Helper()
	at := plannerTestTrader(t)
	at.config.StrategyConfig.DayPlan.WriteTimeFeasibility = writeFeas
	return at
}

// reclaimFeasPlanJSON (CTO stop-side amendment, 2026-09-18): a LONG reclaim
// whose level (arm entry 15480) is already BELOW the read-time price 15550, so
// the executor's stop-side guard would cancel the stop entry at placement
// (trigger 15480.50 ≤ price). The stop is 20pt wide so the min-SL floor
// (1.5×ATR5m 8.0 = 12.0) does NOT fire first — the stop-side predicate is the
// only refusal.
const reclaimFeasPlanJSON = `{
  "reasoning": "Reclaim the broken level.",
  "bias": {"direction": "long", "conviction": "medium", "flip_condition": "2x5m < 15470"},
  "levels": [
    {"price": 15480, "label": "PWL", "grade": "A", "instruction": "reclaim"},
    {"price": 15620, "label": "PDH", "grade": "A", "instruction": "fade"}
  ],
  "scenarios": [{"id": "S1", "trigger": "reclaim 15480", "condition": "reclaim", "direction": "long", "target_chain": [15550, 15620], "invalid": "2x5m<15470", "quality": "A", "confirm":{"rule":"1x5m_close","ref_price":15480,"side":"above"},"economics":{"entry_zone":[15480,15480],"geometry":{"entry":15480,"stop":15460,"target":15620},"first_obstacle":{"price":15550,"level":"fixture reference","family":"reference","response":"pass_through"},"r_to_obstacle":3.5,"r_to_arm_target":7.0},"arm":{"enabled":true,"entry":15480,"stop":15460,"target":15620}}],
  "no_trade": ["first 5m"],
  "death_condition": "acceptance above 15620",
  "death": {"price": 15620, "side": "above", "rule": "2x5m"},
  "flip": {"price": 15470, "side": "below", "rule": "2x5m", "flip_to": "short"},
  "day_type": "balance"
}`

// feasClock returns the fixed authoring clock the seam walk (class 60/113)
// requires of every test that drives the planner.
func feasClock() func() time.Time {
	now := time.Date(2026, 9, 8, 19, 0, 0, 0, time.UTC)
	return func() time.Time { return now }
}

// feasATR5m recomputes the 5m ATR exactly the way the write site does.
func feasATR5m(t *testing.T) float64 {
	t.Helper()
	b5 := market.FuturesBarsProvider("MNQ", "5m", kernel.AISVPBarCount)
	if len(b5) == 0 {
		t.Fatalf("stub provider returned no bars")
	}
	return market.ExportCalculateATR(b5, 14)
}

// expectedFeasReason runs the SAME verdict function the write site runs, with
// the same inputs, so the test asserts parity instead of hardcoding a predicate
// (canon 53).
func expectedFeasReason(t *testing.T, at *AutoTrader, atr5m float64) string {
	t.Helper()
	var doc kernel.PlanDoc
	if err := json.Unmarshal([]byte(infeasibleFeasPlanJSON), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	sc := doc.Scenarios[0]
	if sc.Arm == nil {
		t.Fatalf("fixture scenario must carry an arm")
	}
	leg := kernel.PlanArmLeg{Entry: sc.Arm.Entry, Stop: sc.Arm.Stop, Target: sc.Arm.Target}
	var mq string
	if at.config.StrategyConfig != nil && at.config.StrategyConfig.DayPlan != nil {
		mq = at.config.StrategyConfig.DayPlan.MinGradeFor("NY")
	}
	return at.armGateVerdictFor(sc, leg, biasDirectionFor(doc.Bias.Direction), nil, atr5m, mq, at.config.StrategyConfig, "NY", false)
}

// (b) HINT RED→GREEN at the planner call site: the write-time verdict feeds the
// repair prompt (verbatim refusal + fix vocabulary); the next attempt (wider
// stop) writes active.
func TestWriteTimeFeasibilityHintRedToGreen(t *testing.T) {
	at := feasPlannerTrader(t, nil) // nil = ON (owner default)
	feasStubBars(t)
	blocks := []string{}
	_, lc, err := at.runPlannerReadCoreWithFactsGradesClock(feasClock(), "NY", "2026-08-14", "owner_reset",
		"deepseek-v4-pro", "hashFeas1", "", "", "", "FULLPROMPT",
		kernel.PlanFacts{Price: 15550, DATR: 300}, nil, map[float64]string{15480: "PWL", 15620: "PDH"}, nil, true,
		func(userPrompt string) (string, error) {
			blocks = append(blocks, userPrompt)
			if len(blocks) == 1 {
				return infeasibleFeasPlanJSON, nil
			}
			r := strings.ReplaceAll(infeasibleFeasPlanJSON, `"stop":15540`, `"stop":15530`)
			r = strings.Replace(r, `"r_to_obstacle":1.0`, `"r_to_obstacle":0.5`, 1)
			return strings.Replace(r, `"r_to_arm_target":7.0`, `"r_to_arm_target":3.5`, 1), nil
		})
	if err != nil || lc != "active" {
		t.Fatalf("repair-then-success: lc=%q err=%v", lc, err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(blocks))
	}
	if blocks[0] != "FULLPROMPT" {
		t.Fatalf("attempt 1 must be the full author prompt, got %q", blocks[0])
	}
	// The repair prompt carries the refusal verbatim plus the fix vocabulary.
	for _, frag := range []string{
		"write-time feasibility:",
		"S1 would be refused at arm —",
		"widen the stop past the min-SL floor",
		"raise the arm R:R",
		"pick a mapped level with an id",
	} {
		if !strings.Contains(blocks[1], frag) {
			t.Fatalf("repair prompt missing %q:\n%s", frag, blocks[1])
		}
	}
	row, _ := at.store.Plan().GetLatestPlanForSession("2026-08-14", "NY")
	if row == nil {
		t.Fatalf("no stored plan")
	}
	var doc kernel.PlanDoc
	if err := json.Unmarshal([]byte(row.Doc), &doc); err != nil {
		t.Fatalf("doc unmarshal: %v", err)
	}
	if doc.Scenarios[0].Arm == nil || !doc.Scenarios[0].Arm.Enabled {
		t.Fatalf("feasible re-write must keep the arm enabled, got %s", row.Doc)
	}
}

// (c) LAST ATTEMPT → arm.enabled=false + arm_disabled_reason + counter, never a
// silent write.
func TestWriteTimeFeasibilityLastAttemptDisablesArm(t *testing.T) {
	at := feasPlannerTrader(t, nil)
	feasStubBars(t)
	atr5m := feasATR5m(t)
	want := expectedFeasReason(t, at, atr5m)
	if want == "" {
		t.Fatalf("fixture arm must be refused at write (atr5m=%.2f)", atr5m)
	}
	_, lc, err := at.runPlannerReadCoreWithFactsGradesClock(feasClock(), "NY", "2026-08-14", "owner_reset",
		"deepseek-v4-pro", "hashFeas2", "", "", "", "FULLPROMPT",
		kernel.PlanFacts{Price: 15550, DATR: 300}, nil, map[float64]string{15480: "PWL", 15620: "PDH"}, nil, true,
		func(userPrompt string) (string, error) { return infeasibleFeasPlanJSON, nil })
	if err != nil || lc != "active" {
		t.Fatalf("last-attempt write: lc=%q err=%v", lc, err)
	}
	row, _ := at.store.Plan().GetLatestPlanForSession("2026-08-14", "NY")
	if row == nil {
		t.Fatalf("no stored plan")
	}
	var doc kernel.PlanDoc
	if err := json.Unmarshal([]byte(row.Doc), &doc); err != nil {
		t.Fatalf("doc unmarshal: %v", err)
	}
	if len(doc.Scenarios) == 0 || doc.Scenarios[0].Arm == nil {
		t.Fatalf("doc lost its arm: %s", row.Doc)
	}
	if doc.Scenarios[0].Arm.Enabled {
		t.Fatalf("last attempt must write the arm DISABLED, got enabled: %s", row.Doc)
	}
	if !strings.Contains(doc.Scenarios[0].Arm.DisabledReason, "too close") {
		t.Fatalf("arm_disabled_reason must name the refusal, got %q", doc.Scenarios[0].Arm.DisabledReason)
	}
	key := "arm_disabled_at_write:t1:2026-08-14:NY:" + want
	if n, err := store.SystemCounter(at.store, key); err != nil || n != 1 {
		t.Fatalf("counter %q = %d, %v (want 1)", key, n, err)
	}
}

// (d) KNOB OFF → byte-identical behaviour: the arm writes enabled on attempt 1,
// no repair, no disabled_reason stamp.
func TestWriteTimeFeasibilityOffIsByteIdentical(t *testing.T) {
	off := false
	at := feasPlannerTrader(t, &off)
	feasStubBars(t)
	blocks := []string{}
	_, lc, err := at.runPlannerReadCoreWithFactsGradesClock(feasClock(), "NY", "2026-08-14", "owner_reset",
		"deepseek-v4-pro", "hashFeas3", "", "", "", "FULLPROMPT",
		kernel.PlanFacts{Price: 15550, DATR: 300}, nil, map[float64]string{15480: "PWL", 15620: "PDH"}, nil, true,
		func(userPrompt string) (string, error) {
			blocks = append(blocks, userPrompt)
			return infeasibleFeasPlanJSON, nil
		})
	if err != nil || lc != "active" {
		t.Fatalf("OFF must write attempt 1: lc=%q err=%v", lc, err)
	}
	if len(blocks) != 1 {
		t.Fatalf("OFF must not repair, got %d attempts", len(blocks))
	}
	row, _ := at.store.Plan().GetLatestPlanForSession("2026-08-14", "NY")
	var doc kernel.PlanDoc
	if err := json.Unmarshal([]byte(row.Doc), &doc); err != nil {
		t.Fatalf("doc unmarshal: %v", err)
	}
	if len(doc.Scenarios) == 0 || doc.Scenarios[0].Arm == nil || !doc.Scenarios[0].Arm.Enabled {
		t.Fatalf("OFF must keep the arm enabled (today's WARN-only behaviour), got %s", row.Doc)
	}
	if doc.Scenarios[0].Arm.DisabledReason != "" {
		t.Fatalf("OFF must not stamp arm_disabled_reason, got %q", doc.Scenarios[0].Arm.DisabledReason)
	}
}

// TestWriteTimeFeasibilityStopSideHintText pins the amendment's hint words
// (CTO 2026-09-18): the trigger, its relation to price, and the two fixes.
func TestWriteTimeFeasibilityStopSideHintText(t *testing.T) {
	hint := writeTimeFeasibilityHint([]writeTimeFeasibilityIssue{
		{Scenario: "S1", Cond: "reclaim", Kind: "stop_side", Trigger: 15480.50, Price: 15550.00, Side: "long", Reason: "stop_side_wrong"},
	})
	for _, frag := range []string{
		"write-time feasibility:",
		"S1 reclaim trigger 15480.50 is already below price 15550.00",
		"a stop entry there fills at market on placement",
		"author the trigger ahead of price",
		"author a reject/limit at the level",
	} {
		if !strings.Contains(hint, frag) {
			t.Fatalf("hint missing %q: %s", frag, hint)
		}
	}
	if strings.Contains(hint, "would be refused at arm") {
		t.Fatalf("stop-side hint must not carry the gate-kind suffix: %s", hint)
	}
}

// TestWriteTimeFeasibilityStopSideLastAttemptDisables — the amendment's (c): a
// reclaim whose stop trigger is already behind the read-time price rides the
// repair hint, and the last attempt writes arm.enabled=false with
// disabled_reason stop_side_wrong + the counter.
func TestWriteTimeFeasibilityStopSideLastAttemptDisables(t *testing.T) {
	at := feasPlannerTrader(t, nil)
	feasStubBars(t)
	blocks := []string{}
	_, lc, err := at.runPlannerReadCoreWithFactsGradesClock(feasClock(), "NY", "2026-08-14", "owner_reset",
		"deepseek-v4-pro", "hashFeas4", "", "", "", "FULLPROMPT",
		kernel.PlanFacts{Price: 15550, DATR: 300}, nil, map[float64]string{15480: "PWL", 15620: "PDH"}, nil, true,
		func(userPrompt string) (string, error) {
			blocks = append(blocks, userPrompt)
			return reclaimFeasPlanJSON, nil
		})
	if err != nil || lc != "active" {
		t.Fatalf("last-attempt write: lc=%q err=%v", lc, err)
	}
	if len(blocks) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(blocks))
	}
	for _, frag := range []string{
		"trigger 15480.50 is already below price 15550.00",
		"a stop entry there fills at market on placement",
	} {
		if !strings.Contains(blocks[1], frag) {
			t.Fatalf("repair prompt missing %q:\n%s", frag, blocks[1])
		}
	}
	row, _ := at.store.Plan().GetLatestPlanForSession("2026-08-14", "NY")
	if row == nil {
		t.Fatalf("no stored plan")
	}
	var doc kernel.PlanDoc
	if err := json.Unmarshal([]byte(row.Doc), &doc); err != nil {
		t.Fatalf("doc unmarshal: %v", err)
	}
	if doc.Scenarios[0].Arm == nil || doc.Scenarios[0].Arm.Enabled {
		t.Fatalf("last attempt must write the arm DISABLED, got %s", row.Doc)
	}
	if doc.Scenarios[0].Arm.DisabledReason != "stop_side_wrong" {
		t.Fatalf("disabled_reason must be stop_side_wrong, got %q", doc.Scenarios[0].Arm.DisabledReason)
	}
	key := "arm_disabled_at_write:t1:2026-08-14:NY:stop_side_wrong"
	if n, err := store.SystemCounter(at.store, key); err != nil || n != 1 {
		t.Fatalf("counter %q = %d, %v (want 1)", key, n, err)
	}
}
