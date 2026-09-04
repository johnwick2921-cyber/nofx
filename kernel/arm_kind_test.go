// D3 — one function decides limit vs stop-entry, and both the prompt table and
// the executor read it.
//
// The mapping is MECHANICS, not a market belief: a play that rests AT a level
// is a limit; a play that only becomes valid once price travels BEYOND a
// trigger is a stop-entry. Encoding it once is what stops the prompt promising
// one thing while the composer does another.

package kernel

import "testing"

func TestArmKindForRestsAtALevelIsLimit(t *testing.T) {
	for _, c := range []string{"reject", "fvg_entry", "sweep_reclaim"} {
		if got := ArmKindFor(c); got != ArmKindLimit {
			t.Errorf("ArmKindFor(%q) = %q, want %q — it rests AT a price", c, got, ArmKindLimit)
		}
	}
}

func TestArmKindForContinuationIsStopEntry(t *testing.T) {
	// The continuation plays are the ones that make a LONG armable in a trend:
	// breakup_continue only becomes valid once price travels beyond the trigger.
	for _, c := range []string{"breakup_continue", "breakdown_continue"} {
		if got := ArmKindFor(c); got != ArmKindStopEntry {
			t.Errorf("ArmKindFor(%q) = %q, want %q — it triggers BEYOND a price", c, got, ArmKindStopEntry)
		}
	}
}

// Case and padding must not change the answer: the condition arrives from model
// JSON (canon 28 — canonicalize where the value enters).
func TestArmKindForIsCanonicalized(t *testing.T) {
	if ArmKindFor("  BreakUp_Continue  ") != ArmKindStopEntry {
		t.Error("ArmKindFor must canonicalize case and padding")
	}
}

// A condition that cannot be armed at all has no kind — the caller must not get
// a plausible default it would then act on.
func TestArmKindForNonArmableIsEmpty(t *testing.T) {
	for _, c := range []string{"acceptance", "hold", "reclaim", "", "nonsense"} {
		if got := ArmKindFor(c); got != "" {
			t.Errorf("ArmKindFor(%q) = %q, want \"\" — non-armable conditions have no kind", c, got)
		}
	}
}

// Every armable condition must have a kind, or the table and the whitelist have
// drifted apart — which is exactly how the prompt and the composer disagree.
func TestEveryArmableConditionHasAKind(t *testing.T) {
	for _, c := range []string{"fvg_entry", "reject", "breakdown_continue", "breakup_continue", "sweep_reclaim"} {
		if !ArmableCondition(c) && c != "sweep_reclaim" {
			t.Fatalf("test premise wrong: %q is not armable", c)
		}
		if ArmKindFor(c) == "" {
			t.Errorf("%q is armable but ArmKindFor gives no kind — table drifted from the whitelist", c)
		}
	}
}

// The mismatch the owner asked the executor to refuse.
func TestArmKindMismatchIsNamed(t *testing.T) {
	err := ArmKindMismatch("reject", "stop_entry")
	if err == nil {
		t.Fatal("a stop_entry authored for a reject fade must be refused")
	}
	if got := err.Error(); got == "" {
		t.Fatal("the refusal must name what was wrong")
	}
	if ArmKindMismatch("breakup_continue", "stop_entry") != nil {
		t.Error("the correct kind must not be refused")
	}
	// An unset kind is not a mismatch — the composer derives it.
	if ArmKindMismatch("reject", "") != nil {
		t.Error("an absent kind is derived, not refused")
	}
}
