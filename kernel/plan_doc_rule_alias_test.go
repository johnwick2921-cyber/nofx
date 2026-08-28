package kernel

import (
	"testing"
)

// F1b (LONDON-FORENSICS 2026-08-28) — alias completion fixtures.
//
// The 02:03/02:08 planner attempts were rejected for flip.rule "2x5m_close"
// (15 rejects this week) and the 02:23 attempt for confirm.rule "5m_close"
// (2 rejects). NormalizePlanDocRules canonicalizes those spellings at parse
// time so a truncation-adjacent spelling never burns a retry again — and an
// unknown spelling is still rejected honestly.

func validAliasDoc() *PlanDoc {
	return &PlanDoc{
		Reasoning: "fixture",
		Bias: PlanBias{
			Direction:     "short",
			Conviction:    "low",
			FlipCondition: "15m close above 29671.88 flips bias to long",
		},
		Levels: []PlanLevel{
			{Price: 29591.50, Label: "ONL", Grade: "A", Instruction: "fade"},
		},
		Scenarios: []PlanScenario{
			{
				ID: "S1", Trigger: "reject at 29642.00", Condition: "reject",
				Direction: "short", TargetChain: []float64{29550},
				Invalid: "invalid above 29642.00", Quality: "B",
				Confirm: &PlanConfirm{Rule: "5m_close", RefPrice: 29642.00, Side: "below"},
			},
		},
		DeathCondition: "15m close above 29707.50 voids the plan",
		FlipStructured: &PlanCondition{Price: 29671.88, Side: "above", Rule: "2x5m_close", FlipTo: "long"},
	}
}

func TestPlanDocRuleAliasesNormalizeAndPass(t *testing.T) {
	d := validAliasDoc()
	if err := ValidatePlanDocWithCaps(d, 0, 0); err != nil {
		t.Fatalf("the exact 02:03/02:08 failing shapes must now pass: %v", err)
	}
	if d.Scenarios[0].Confirm.Rule != "1x5m_close" {
		t.Fatalf("confirm.rule normalized = %q, want 1x5m_close", d.Scenarios[0].Confirm.Rule)
	}
	if d.FlipStructured.Rule != "2x5m" {
		t.Fatalf("flip.rule normalized = %q, want 2x5m", d.FlipStructured.Rule)
	}
}

func TestPlanDocRuleAliasUnknownStillRejected(t *testing.T) {
	d := validAliasDoc()
	d.Scenarios[0].Confirm.Rule = "9x5m_close"
	if err := ValidatePlanDocWithCaps(d, 0, 0); err == nil {
		t.Fatal("an unknown confirm.rule spelling must still be rejected")
	}
	d = validAliasDoc()
	d.FlipStructured.Rule = "9x5m"
	if err := ValidatePlanDocWithCaps(d, 0, 0); err == nil {
		t.Fatal("an unknown flip.rule spelling must still be rejected")
	}
}

func TestPlanDocRuleAliasFullVocabulary(t *testing.T) {
	for in, want := range map[string]string{
		"5m_close": "1x5m_close", "5m-close": "1x5m_close", "5mclose": "1x5m_close", "1x5m": "1x5m_close",
		"15m": "15m_close", "15m-close": "15m_close", "15mclose": "15m_close",
		"2x5m": "2x5m_close", "2x_5m": "2x5m_close",
		"touch": "touch", "1x5m_close": "1x5m_close", "2x5m_close": "2x5m_close", "15m_close": "15m_close",
	} {
		if got := NormalizeConfirmRule(in); got != want {
			t.Errorf("NormalizeConfirmRule(%q) = %q, want %q", in, got, want)
		}
	}
	for in, want := range map[string]string{
		"2x5m_close": "2x5m", "2x_5m": "2x5m", "2x5": "2x5m",
		"15m": "15m_close", "15m-close": "15m_close", "15mclose": "15m_close",
		"1x5m_close": "5m_close", "1x5m": "5m_close", "5m-close": "5m_close", "5mclose": "5m_close",
		"2x5m": "2x5m", "15m_close": "15m_close", "5m_close": "5m_close",
	} {
		if got := NormalizeConditionRule(in); got != want {
			t.Errorf("NormalizeConditionRule(%q) = %q, want %q", in, got, want)
		}
	}
}

// F4 (LONDON-FORENSICS 2026-08-28) — arm feasibility WARN math mirrors the
// gate-at-arm chain: the LONDON v1 arms that printed ~120 REFUSED lines must
// each produce a warning, and the ASIA v12 arm that FILLED must produce none.
func TestArmFeasibilityWarningsMatchTheLiveRefusals(t *testing.T) {
	d := &PlanDoc{Scenarios: []PlanScenario{
		{ID: "S1", Condition: "reject", Direction: "short",
			Arm: &PlanArmSpec{Enabled: true, Entry: 29640, Stop: 29650, Target: 29619.5}},
		{ID: "S2", Condition: "sweep_reclaim", Direction: "short",
			Arm: &PlanArmSpec{Enabled: true, Entry: 29676, Stop: 29712.5, Target: 29642, WaitConfirm: true}},
		{ID: "S3", Condition: "breakout_retest", Direction: "short",
			Arm: &PlanArmSpec{Enabled: true, Entry: 29589, Stop: 29598, Target: 29576.5}},
		{ID: "S4", Condition: "reject", Direction: "short",
			Arm: &PlanArmSpec{Enabled: true, Entry: 29621.01, Stop: 29642, Target: 29576.5}},
	}}
	w := ArmFeasibilityWarnings(d, 16.49, 2.0, 1.0)
	byScenario := map[string]int{}
	for _, s := range w {
		byScenario[s[:2]]++
	}
	if byScenario["S1"] == 0 || byScenario["S2"] == 0 || byScenario["S3"] == 0 {
		t.Fatalf("the three live-refused arms must all warn: %v", w)
	}
	if byScenario["S4"] != 0 {
		t.Fatalf("the ASIA v12 arm that filled must NOT warn: %v", w)
	}
}
