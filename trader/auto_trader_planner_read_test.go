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

// P0-relax (2026-08-27) — write-site: a machine-caused thin side must WARN,
// write the plan, and stamp the thin_side note onto the stored doc (the card
// renders it). The machine map is what the prompt displayed (seated + owner +
// HTF rows, price-keyed).
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

func TestRunPlannerReadMachineThinWritesWithNote(t *testing.T) {
	at := plannerTestTrader(t)
	facts := kernel.PlanFacts{Price: 29614, DATR: 300} // PDH/PDL 0 → gap rules skipped
	machine := map[float64]string{30000: "RN 30000"}   // the map itself is thin above
	ver, lc, err := at.runPlannerReadCoreWithFactsGrades("ASIA", "2026-08-26", "owner_reset",
		"deepseek-v4-pro", "hashQ", "", "", "", facts, nil, machine, nil, true, 2,
		func() (string, error) { return thinAbovePlanJSON, nil })
	if err != nil || ver != 1 || lc != "active" {
		t.Fatalf("machine-thin write: ver=%d lc=%q err=%v want active", ver, lc, err)
	}
	row, _ := at.store.Plan().GetLatestPlanForSession("2026-08-26", "ASIA")
	if row == nil || row.TriggerReason != "owner_reset" {
		t.Fatalf("stored row wrong: %+v", row)
	}
	var doc kernel.PlanDoc
	if err := json.Unmarshal([]byte(row.Doc), &doc); err != nil {
		t.Fatalf("doc unmarshal: %v", err)
	}
	if !strings.Contains(doc.ThinSide, "above") || !strings.Contains(doc.ThinSide, "machine map 1") {
		t.Fatalf("thin_side note not stamped: %q", doc.ThinSide)
	}
}

func TestRunPlannerReadAIOmissionFailsClosed(t *testing.T) {
	at := plannerTestTrader(t)
	facts := kernel.PlanFacts{Price: 29614, DATR: 300}
	machine := map[float64]string{ // map HAS 3 above — the plan carries only 1
		29500: "PDL", 29550: "RN 29550", 29600: "RN 29600",
		30000: "RN 30000", 30100: "RN 30100", 30200: "RN 30200",
	}
	_, lc, err := at.runPlannerReadCoreWithFactsGrades("ASIA", "2026-08-26", "owner_reset",
		"deepseek-v4-pro", "hashQ2", "", "", "", facts, nil, machine, nil, true, 2,
		func() (string, error) { return thinAbovePlanJSON, nil })
	// The fail-closed NO-TRADE WRITE succeeds (err nil) — the outcome is the
	// lifecycle, never an error return.
	if lc != "no_trade" || err != nil {
		t.Fatalf("AI omission must fail-closed: lc=%q err=%v", lc, err)
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
