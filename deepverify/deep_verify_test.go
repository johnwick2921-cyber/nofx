// deep-verify-22 fixtures (2026-08-27, rev c21ad24a) — standalone fixture tests
// that drive the SHIPPED code paths cold. Committed as the fixture delta; never
// call the function under test to COMPUTE expectations (R2).
package deepverify_test

import (
	"strings"
	"testing"

	"nofx/kernel"
	"nofx/telemetry"
)

// NOTE: the market-entry R:R fixture lives in kernel/deepverify_fixture_test.go
// (package kernel — validateDecision is unexported); the ARM-floor fixture in
// trader/deepverify_fixture_test.go (armGateVerdict is unexported).

// G1 1.3 — FantasyTargetWarnings: R=7 warns, R=5.9 silent, and the warning can
// NEVER fail a write (ValidatePlanDocWithCaps returns nil on a R=7 arm).
func TestG1FantasyTargetWarningsNeverFail(t *testing.T) {
	doc := kernel.PlanDoc{Reasoning: "fixture", Bias: kernel.PlanBias{Direction: "long", Conviction: "low", FlipCondition: "n/a"}, DeathCondition: "n/a",
		Scenarios: []kernel.PlanScenario{
			{ID: "S1", Trigger: "t", Condition: "reject", Direction: "long", Quality: "A",
				Confirm: &kernel.PlanConfirm{Rule: "touch", RefPrice: 100, Side: "above"}, Invalid: "1x5m close above 100",
				TargetChain: []float64{120},
			},
		}}
	// R=7 arm (long: entry 100, stop 95, target 135 → 7.0)
	doc.Scenarios[0].Arm = &kernel.PlanArmSpec{Enabled: true, Entry: 100, Stop: 95, Target: 135}
	w := kernel.FantasyTargetWarnings(doc)
	if len(w) != 1 || !strings.Contains(w[0], "S1") || !strings.Contains(w[0], "fantasy-target") {
		t.Fatalf("R=7 must WARN once, got %v", w)
	}
	// R=5.9 → silent
	doc.Scenarios[0].Arm = &kernel.PlanArmSpec{Enabled: true, Entry: 100, Stop: 95, Target: 129.5}
	if w := kernel.FantasyTargetWarnings(doc); len(w) != 0 {
		t.Fatalf("R=5.9 must be silent, got %v", w)
	}
	// NEVER fails the write: a R=7 arm passes the validator chain
	doc.Scenarios[0].Arm = &kernel.PlanArmSpec{Enabled: true, Entry: 100, Stop: 95, Target: 135}
	doc.Scenarios[0].Confirm = &kernel.PlanConfirm{Rule: "touch", RefPrice: 100, Side: "above"}
	if err := kernel.ValidatePlanDocWithCaps(&doc, 0, 0); err != nil {
		t.Fatalf("a fantasy-target arm must still WRITE (WARN-only): %v", err)
	}
}

// G1 1.4 — the leak counters exist and increment via the gate-block table.
func TestG1LeakCountersIncrement(t *testing.T) {
	telemetry.IncGateBlock("fixture-deepverify", "decline_fresh_met")
	telemetry.IncGateBlock("fixture-deepverify", "arm_authored")
	telemetry.IncGateBlock("fixture-deepverify", "arm_authored")
	_, table := telemetry.GateBlockSnapshot()
	row := table["fixture-deepverify"]
	if row == nil || row["decline_fresh_met"] < 1 || row["arm_authored"] < 2 {
		t.Fatalf("counters missing from the snapshot: %v", row)
	}
}

// G2 2.1 — ClassifyCitation matrix: {matched, direction-mismatch, unknown-id,
// empty-cited} × {open_long, open_short} — 8 verdicts.
func TestG2ClassifyCitationMatrix(t *testing.T) {
	doc := kernel.PlanDoc{Scenarios: []kernel.PlanScenario{
		{ID: "S1", Direction: "long"},
		{ID: "S2", Direction: "short"},
	}}
	type tc struct {
		action, cited string
		want          string // cited|matched|offplan
	}
	cases := []tc{
		{"open_long", "S1", "S1|true|false"},
		{"open_short", "S2", "S2|true|false"},
		{"open_long", "S2", "S2|false|false"},  // direction mismatch
		{"open_short", "S1", "S1|false|false"}, // direction mismatch
		{"open_long", "S9", "off-plan|false|true"},
		{"open_short", "S9", "off-plan|false|true"},
		{"open_long", "", "off-plan|false|true"},
		{"open_short", "", "off-plan|false|true"},
	}
	for _, c := range cases {
		r := kernel.ClassifyCitation(c.action, c.cited, doc)
		got := r.Cited + "|" + bools(r.Matched) + "|" + bools(r.OffPlan)
		if got != c.want {
			t.Errorf("%s %q → %s, want %s", c.action, c.cited, got, c.want)
		}
	}
	// no-plan (empty doc): any citation resolves off-plan
	if r := kernel.ClassifyCitation("open_long", "S1", kernel.PlanDoc{}); !r.OffPlan {
		t.Fatalf("empty-doc citation must be off-plan, got %+v", r)
	}
}

func bools(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
