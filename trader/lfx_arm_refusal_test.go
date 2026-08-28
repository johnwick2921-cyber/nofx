package trader

import (
	"testing"
)

// F4 (LONDON-FORENSICS 2026-08-28) — the every-cycle REFUSED spam dedup: one
// log line per arm-spec verdict, silent until the spec (or the refusal reason)
// changes. Pure decision, tested directly.

func TestArmRefusalChangedDedup(t *testing.T) {
	var last map[string]string
	if !armRefusalChanged(&last, "plan:v12:S1", "R:R 0.93 below arm min 2.00") {
		t.Fatal("first refusal must log")
	}
	if armRefusalChanged(&last, "plan:v12:S1", "R:R 0.93 below arm min 2.00") {
		t.Fatal("identical re-refusal must stay silent")
	}
	if !armRefusalChanged(&last, "plan:v12:S1", "stop too close") {
		t.Fatal("a CHANGED refusal reason must log again")
	}
	if !armRefusalChanged(&last, "plan:v13:S1", "R:R 0.93 below arm min 2.00") {
		t.Fatal("a new plan version (spec change) must log again")
	}
	if armRefusalChanged(&last, "plan:v13:S1", "R:R 0.93 below arm min 2.00") {
		t.Fatal("same verdict on the same version stays silent")
	}
}

// F1a (LONDON-FORENSICS 2026-08-28) — the planner token knob default.
func TestAIPlanMaxTokensDefault(t *testing.T) {
	t.Setenv("AI_PLAN_MAX_TOKENS", "")
	if got := aiPlanMaxTokens(); got != 65536 {
		t.Fatalf("default = %d, want 65536 (2× the observed 32768 truncation ceiling)", got)
	}
	t.Setenv("AI_PLAN_MAX_TOKENS", "100000")
	if got := aiPlanMaxTokens(); got != 100000 {
		t.Fatalf("override = %d, want 100000", got)
	}
	t.Setenv("AI_PLAN_MAX_TOKENS", "0") // invalid → default
	if got := aiPlanMaxTokens(); got != 65536 {
		t.Fatalf("zero override = %d, want 65536 fallback", got)
	}
}
