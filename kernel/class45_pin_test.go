package kernel

import (
	"strings"
	"testing"

	"nofx/market"
)

// ── CLASS 45 (2026-09-02) — THE PROMPT WITHHOLDS WHAT THE VALIDATOR KNOWS ────
//
// LONDON 2026-09-02, rows 92/93/94 of planner_rejected_prompts:
//
//   01:32:54  attempt 1 (25,903 ch)  S1 breakdown_continue @29021.25 → "a close
//                                    came back across 29021.25 — void"
//   01:35:04  attempt 2 ( 3,673 ch)  repair obeyed the hint (author a reject
//                                    fade) → killed by fade_requires_touch
//   01:37:44  attempt 3 (26,407 ch)  full RE-AUTHOR → S2 breakdown_continue at
//                                    the SAME 29021.25 → void again → fail-closed
//
// Why attempt 3 regressed, measured from the rendered prompts:
//   · the standing MUST ("If price sits BELOW PDL you MUST write a continuation
//     short") sits at char 18,544 of 26,408 — 70% in, full weight.
//   · the reject block is the LAST 239 chars — 59 tokens of 6,602, under 1%, at
//     99% depth — and it carried attempt 2's fade defect, not attempt 1's void.
//   · the facts block names 29021.25 SIX times and never once says it is void:
//     "VOID" / "reclaimed" / "came back across" each appear 0 times.
//
// So the model was ordered to write a continuation short, was never told the
// only breakdown level was dead, and was corrected about something else. It
// complied. The validator knew all along — BreakdownContinueState computes
// Reclaimed — and told nobody until the write.

// pinLondonBars rebuilds the shape of that tape: a breakdown through 29021.25
// followed by a close back across it (the reclaim that voids the play).
func pinLondonBars() []market.Kline {
	const lvl = 29021.25
	var out []market.Kline
	t := int64(1_756_800_000_000)
	add := func(c float64) {
		out = append(out, market.Kline{OpenTime: t, CloseTime: t + 59_000, Open: c, High: c + 3, Low: c - 3, Close: c})
		t += 60_000
	}
	add(lvl + 6) // above
	add(lvl - 8) // the breakdown close
	add(lvl - 12)
	add(lvl + 4) // ← a close came back ACROSS: the breakdown is VOID
	add(lvl + 7)
	return out
}

// TestClass45PinLondon0132 is the wave's pin. It asserts the three things the
// dispatch fixes, using only surfaces that exist on the pre-45 tree, so it
// compiles there and FAILS.
func TestClass45PinLondon0132(t *testing.T) {
	const lvl = 29021.25
	bars := pinLondonBars()

	// The validator's own function already knows the level is void.
	sc := PlanScenario{
		ID: "S1", Condition: "breakdown_continue", Direction: "short",
		Breakdown: &PlanBreakdownContinue{Level: lvl, EntryMode: "pullback"},
	}
	st := BreakdownContinueState(sc, bars, bars[0].OpenTime, bars[len(bars)-1].CloseTime)
	if !st.Reclaimed {
		t.Fatalf("fixture: the tape must reclaim %.2f (the validator's own predicate says void)", lvl)
	}

	// (E1) The MUST line must ask for a DIRECTION, not a condition. As shipped it
	// names "a continuation short", which the prompt binds to
	// breakdown_continue/breakup_continue — an instruction the validator then
	// refuses when the only breakdown level is void.
	prompt := BuildPlannerPrompt(PlannerInput{})
	must := "If price sits BELOW PDL you MUST write a continuation short"
	if strings.Contains(prompt, must) {
		t.Errorf("E1: the prompt still orders a CONDITION (%q). It must ask for a SHORT-DIRECTION scenario and name the legal conditions, so the order stays satisfiable when every breakdown level is void", must)
	}
	if !strings.Contains(prompt, "SHORT-direction scenario") {
		t.Errorf("E1: the prompt must ask for a SHORT-direction scenario (any legal condition)")
	}

	// E1b (plan_doc's message), E2 (the void line) and E3 (the floor line) need
	// mechanisms that do not exist yet; they are pinned in
	// class45_feeds_forward_test.go, which lands WITH the implementation. This
	// test deliberately stays compilable on the pre-45 tree so its failure is
	// BEHAVIOURAL, not a build error.
}
