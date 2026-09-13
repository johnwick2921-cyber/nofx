package kernel

import (
	"strings"
	"testing"
)

func TestPlannerPromptSeparatesStructuralFadeFromLegacyFloor(t *testing.T) {
	p := BuildPlannerPrompt(PlannerInput{StopFloorATR5m: 20, StopFloorMult: 1.5})
	for _, required := range []string{"reject fades use the frozen zone", "does not widen a structural fade stop", "No-trade is valid", "Non-reject arms"} {
		if !strings.Contains(p, required) {
			t.Errorf("planner missing structural contract: %s", required)
		}
	}
	for _, forbidden := range []string{"If your setup cannot meet BOTH", "## Minimum stop distance this cycle"} {
		if strings.Contains(p, forbidden) {
			t.Errorf("planner still publishes universal floor/bypass: %s", forbidden)
		}
	}
}

func TestAuthoredRejectGeometryDoesNotProduceLegacyFloorWarnings(t *testing.T) {
	d := &PlanDoc{Scenarios: []PlanScenario{{ID: "S1", Condition: "reject", Direction: "long", Arm: &PlanArmSpec{Enabled: true, Entry: 29600, Stop: 29598, Target: 29601}}}}
	if w := ArmFeasibilityWarnings(d, 20, 2, 1.5); len(w) != 0 {
		t.Fatalf("authored values do not represent composed structural geometry: %v", w)
	}
}
