package kernel

import "testing"

// C-gate (2026-08-24, owner "B and up only"): decisions citing C-quality
// scenarios are refused; B and off-plan pass.
func TestRequireScenarioQualityAboveC(t *testing.T) {
	plan := &ActivePlan{Doc: PlanDoc{Scenarios: []PlanScenario{
		{ID: "S1", Quality: "C", Consumed: true},
		{ID: "S2", Quality: "B"},
	}}}
	fd := func(cite string) *FullDecision {
		return &FullDecision{Decisions: []Decision{{Action: "open_short", CitedScenario: cite}}}
	}
	if err := requireScenarioQualityAboveC(fd("S1"), plan); err == nil {
		t.Fatal("C citation must be refused")
	}
	if err := requireScenarioQualityAboveC(fd("S2"), plan); err != nil {
		t.Fatalf("B citation must pass: %v", err)
	}
	if err := requireScenarioQualityAboveC(fd("off-plan"), plan); err != nil {
		t.Fatalf("off-plan must pass: %v", err)
	}
	// non-open actions are never gated
	if err := requireScenarioQualityAboveC(&FullDecision{Decisions: []Decision{{Action: "close_long", CitedScenario: "S1"}}}, plan); err != nil {
		t.Fatalf("close actions must pass: %v", err)
	}
}

// The plan block must SAY C scenarios are not tradeable (the model re-judges
// instead of retry-looping against the gate).
func TestRenderPlanBlockMarksCAsNotTradeable(t *testing.T) {
	doc := PlanDoc{
		Bias:          PlanBias{Direction: "short", Conviction: "medium", FlipCondition: "n/a"},
		DeathCondition: "15m above PDL",
		Scenarios: []PlanScenario{
			{ID: "S1", Quality: "C", Consumed: true, Condition: "acceptance", Direction: "short", Trigger: "below EQL", Invalid: "n/a"},
			{ID: "S2", Quality: "B", Condition: "reject", Direction: "short", Trigger: "rally to PWL", Invalid: "n/a"},
		},
	}
	block := RenderPlanBlock(doc, "NY")
	if !containsStr(block, "C·CONSUMED (NOT tradeable)") {
		t.Fatalf("plan block must mark C scenarios untradeable:\n%s", block)
	}
	if !containsStr(block, "S2 [B]") {
		t.Fatalf("B scenario tag wrong:\n%s", block)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
