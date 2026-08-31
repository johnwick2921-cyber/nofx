package trader

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"nofx/kernel"
	"nofx/store"
)

const validTraderPlanJSON = `{
  "reasoning": "Balance below PDH; fade edges, long the reclaim.",
  "bias": {"direction": "long", "conviction": "medium", "flip_condition": "2x5m < 15480"},
  "levels": [
    {"price": 15480, "label": "PWL", "grade": "A", "instruction": "fade"},
    {"price": 15520, "label": "RN 15525", "grade": "B", "instruction": "fade"},
    {"price": 15575, "label": "RN 15575", "grade": "B", "instruction": "fade"},
    {"price": 15620, "label": "PDH", "grade": "A", "instruction": "fade"},
    {"price": 15650, "label": "RN 15650", "grade": "B", "instruction": "fade"},
    {"price": 15700, "label": "RN 15700", "grade": "B", "instruction": "fade"}
  ],
  "scenarios": [{"id": "S1", "trigger": "sweep 15480 reclaim", "condition": "sweep_reclaim", "direction": "long", "target_chain": [15550, 15620], "invalid": "2x5m<15470", "quality": "A"}],
  "no_trade": ["first 5m"],
  "death_condition": "acceptance above 15620",
  "death": {"price": 15620, "side": "above", "rule": "2x5m"},
  "flip": {"price": 15480, "side": "below", "rule": "2x5m", "flip_to": "short"},
  "day_type": "balance"
}`

func plannerTestTrader(t *testing.T) *AutoTrader {
	t.Helper()
	st, err := store.New(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { st.Plan().Close(); _ = st.Close() })
	return &AutoTrader{
		id: "t1", exchange: "ninjatrader", store: st,
		config: AutoTraderConfig{StrategyConfig: &store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}}},
	}
}

func TestRunPlannerReadCoreSuccess(t *testing.T) {
	at := plannerTestTrader(t)
	ver, lc, err := at.runPlannerReadCore("NY", "2026-08-14", "deepseek-reasoner", "hashA", "", "",
		func() (string, error) { return validTraderPlanJSON, nil })
	if err != nil || ver != 1 || lc != "active" {
		t.Fatalf("success: ver=%d lc=%q err=%v", ver, lc, err)
	}
	got, _ := at.store.Plan().GetLatestPlanForSession("2026-08-14", "NY")
	if got == nil || got.Lifecycle != "active" || got.ModelID != "deepseek-reasoner" {
		t.Fatalf("stored plan wrong: %+v", got)
	}
}

func TestRunPlannerReadCoreFailClosed(t *testing.T) {
	at := plannerTestTrader(t)
	ver, lc, err := at.runPlannerReadCore("NY", "2026-08-14", "deepseek-reasoner", "hashB", "", "",
		func() (string, error) { return "", errors.New("timeout") })
	if err != nil || lc != "no_trade" {
		t.Fatalf("fail-closed: ver=%d lc=%q err=%v want no_trade", ver, lc, err)
	}
	got, _ := at.store.Plan().GetLatestPlanForSession("2026-08-14", "NY")
	if got == nil || got.Lifecycle != "no_trade" || got.TriggerReason != "planner_fail_closed" {
		t.Fatalf("fail-closed plan not written correctly: %+v", got)
	}
}

// Owner ruling 2026-08-31 — the per-side COUNT concept is deleted: a plan with
// 1 level above while the machine map offered 3 writes ACTIVE with NO WARN, NO
// note, and no thin_side key anywhere in the stored doc. (Previously this exact
// shape fail-closed ASIA 3×, then WARNed — both behaviors are gone.)
const thinAbovePlanJSON = `{
  "reasoning": "thin above: price sits at the top of the stack; only one level above in the map",
  "bias": {"direction": "long", "conviction": "low", "flip_condition": "2x5m < 29500"},
  "levels": [
    {"price": 29500, "label": "PDL", "grade": "A", "instruction": "reclaim"},
    {"price": 29550, "label": "RN 29550", "grade": "B", "instruction": "reclaim"},
    {"price": 29600, "label": "RN 29600", "grade": "B", "instruction": "fade"},
    {"price": 30000, "label": "RN 30000", "grade": "B", "instruction": "fade"}
  ],
  "scenarios": [{"id": "S1", "trigger": "hold 29550", "condition": "hold", "direction": "long", "target_chain": [29700], "invalid": "2x5m<29540", "quality": "B", "confirm": {"rule": "time_hold", "ref_price": 29550, "side": "above"}}],
  "no_trade": ["first 5m"],
  "death_condition": "acceptance above 30000",
  "death": {"price": 30000, "side": "above", "rule": "2x5m"},
  "flip": {"price": 29500, "side": "below", "rule": "2x5m", "flip_to": "short"},
  "day_type": "balance"
}`

func TestRunPlannerReadThinAboveWritesCleanNoArtifacts(t *testing.T) {
	at := plannerTestTrader(t)
	facts := kernel.PlanFacts{Price: 29614, DATR: 300} // PDH/PDL 0 → gap rules skipped
	machine := map[float64]string{                     // rich map: 3 below + 3 above; the plan carries only 1 above
		29500: "PDL", 29550: "RN 29550", 29600: "RN 29600",
		30000: "RN 30000", 30100: "RN 30100", 30200: "RN 30200",
	}
	ver, lc, err := at.runPlannerReadCoreWithFactsGrades("ASIA", "2026-08-26", "owner_reset",
		"deepseek-v4-pro", "hashQ", "", "", "", facts, nil, machine, nil, true,
		func(rejectBlock string) (string, error) { return thinAbovePlanJSON, nil })
	// Count is deleted (owner ruling 2026-08-31): no WARN, no note, plan writes.
	if err != nil || ver != 1 || lc != "active" {
		t.Fatalf("thin-above + rich map must write active: ver=%d lc=%q err=%v", ver, lc, err)
	}
	row, _ := at.store.Plan().GetLatestPlanForSession("2026-08-26", "ASIA")
	if row == nil || row.TriggerReason != "owner_reset" {
		t.Fatalf("stored row wrong: %+v", row)
	}
	if strings.Contains(row.Doc, "thin_side") || strings.Contains(row.Doc, "thin-side") {
		t.Fatalf("no side-count artifact may be stamped, got %s", row.Doc)
	}
	var doc kernel.PlanDoc
	if err := json.Unmarshal([]byte(row.Doc), &doc); err != nil {
		t.Fatalf("doc unmarshal: %v", err)
	}
}

// TestRunPlannerReadRetryAppendRejectBlock — CHANGE 2 (owner ruling 2026-08-31):
// attempt ≥2 carries the previous rejection VERBATIM; a clean attempt 1 sees no
// block.
func TestRunPlannerReadRetryAppendRejectBlock(t *testing.T) {
	at := plannerTestTrader(t)
	facts := kernel.PlanFacts{Price: 15550, DATR: 300}
	machine := map[float64]string{15480: "PWL", 15700: "RN 15700"}
	blocks := []string{}
	_, lc, err := at.runPlannerReadCoreWithFactsGrades("ASIA", "2026-08-26", "owner_reset",
		"deepseek-v4-pro", "hashQB", "", "", "", facts, nil, machine, nil, true,
		func(rejectBlock string) (string, error) {
			blocks = append(blocks, rejectBlock)
			if len(blocks) == 1 {
				return "not json", nil // attempt 1 fails the parse gate
			}
			return validTraderPlanJSON, nil
		})
	if err != nil || lc != "active" {
		t.Fatalf("retry-then-success: lc=%q err=%v", lc, err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(blocks))
	}
	if blocks[0] != "" {
		t.Fatalf("attempt 1 must see NO reject block, got %q", blocks[0])
	}
	if !strings.Contains(blocks[1], "## PREVIOUS ATTEMPT REJECTED / Validator reason (verbatim)") ||
		!strings.Contains(blocks[1], "no JSON object found in planner output") ||
		!strings.Contains(blocks[1], "Fix ONLY this defect") {
		t.Fatalf("attempt 2 must carry the verbatim reject block, got %q", blocks[1])
	}
}

func TestRunPlannerReadCoreRetryThenSuccess(t *testing.T) {
	at := plannerTestTrader(t)
	n := 0
	_, lc, err := at.runPlannerReadCore("NY", "2026-08-14", "m", "hashC", "", "", func() (string, error) {
		n++
		if n < 3 {
			return "not json", nil // 2 invalid → retried
		}
		return validTraderPlanJSON, nil
	})
	if err != nil || lc != "active" {
		t.Fatalf("retry-then-success: lc=%q err=%v (calls=%d)", lc, err, n)
	}
	if n != 3 {
		t.Fatalf("expected 3 attempts (1+2 retries), got %d", n)
	}
}
