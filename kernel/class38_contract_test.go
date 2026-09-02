package kernel

import (
	"regexp"
	"strings"
	"testing"
)

// CLASS 38 (2026-09-01) — PROMPT/VALIDATOR CONTRACT MISMATCH.
//
// Evidence: planner_rejected_prompts rows 78, 79, 80 — one ASIA read, three
// attempts, three DIFFERENT rejects, each one a place where the prompt (or a
// hint it quotes) offers something the validator refuses:
//
//	78  17:47:32 attempt 1  scenario[0].confirm2.rule "1m_mss" not allowed for
//	                        breakdown_continue — entry law: … "2x5m legal ONLY here"
//	                        → the hint names the DEATH/FLIP spelling of the token
//	                          in a message about the CONFIRM field.
//	79  17:53:33 attempt 2  scenario[0].confirm2.rule "2x5m" invalid
//	                        (touch|1x5m_close|2x5m_close|1m_mss|time_hold)
//	                        → the model copied the hint's token verbatim.
//	80  18:00:41 attempt 3  arm legs on breakdown_continue —
//	                        arm_legs_sweep_reclaim_only
//	                        → the schema line offers "legs":[…] on EVERY scenario
//	                          with no condition qualifier; the sweep_reclaim-only
//	                          rule lives ONLY in plan_doc.go.
//
// The model is not the defect; the contract is. These assertions fail on the
// pre-class-38 tree.

// bareRuleToken matches a confirm-rule token spelled in a form the confirm enum
// does NOT contain. \b…\b means "2x5m_close" never matches \b2x5m\b (the "_" is
// a word character), so only genuinely bare spellings trip this.
var bareRuleToken = regexp.MustCompile(`\b(2x5m|1x5m|2x|15m|5m_close|5m-close|2x_5m)\b`)

// TestClass38PinRows78to80 is the row-78/79/80 pin: the three exact defects,
// asserted against the live entry-law table and the live rendered prompt.
func TestClass38PinRows78to80(t *testing.T) {
	// ---- row 78 + 79: the confirm-field hint must speak the confirm enum ----
	// entry_law.go Style strings are quoted VERBATIM into the rejection the
	// model reads ("… not allowed for %s — entry law: %s"). A Style naming a
	// token that confirm.rule/confirm2.rule cannot hold is an instruction the
	// model is punished for following.
	for _, cond := range KnownConditions() {
		law, ok := EntryLawFor(cond)
		if !ok {
			continue
		}
		if m := bareRuleToken.FindString(law.Style); m != "" {
			t.Errorf("entry law Style for %q names %q — not a confirm.rule enum member "+
				"(touch|1x5m_close|2x5m_close|1m_mss|time_hold). Row 78/79: the model copied this token into confirm2.rule and was rejected.\n  Style: %s",
				cond, m, law.Style)
		}
	}

	// The breakdown law must name the token in its enum form explicitly.
	bd, ok := EntryLawFor("breakdown_continue")
	if !ok {
		t.Fatal("breakdown_continue missing from the entry law table")
	}
	if !strings.Contains(bd.Style, "2x5m_close") {
		t.Errorf("breakdown_continue Style must name 2x5m_close (the confirm enum form), got: %s", bd.Style)
	}

	// The repair excerpt the model reads on attempt 2 (row 79's prompt) must
	// speak the same enum.
	if m := bareRuleToken.FindString(RepairEntryConfirmLaw); m != "" {
		t.Errorf("RepairEntryConfirmLaw names bare token %q — row 79 is exactly this token echoed back into confirm2.rule.\n  %s", m, RepairEntryConfirmLaw)
	}

	// ---- row 80: the schema must qualify legs, and the prose must forbid it ----
	prompt := plannerOutputContract(8, 5, true, true)

	idx := strings.Index(prompt, `"legs"`)
	if idx < 0 {
		t.Fatal(`the rendered output contract has no "legs" field — locate the schema line before editing`)
	}
	// The qualifier must ride WITH the field, the way fvg{}/breakdown{} carry
	// theirs. Checking the whole line would pass on the unrelated
	// "condition": "…|sweep_reclaim|…" enum earlier in the same line, so the
	// window starts at the field itself.
	legsWindow := prompt[idx:min(idx+400, len(prompt))]
	if !strings.Contains(legsWindow, "ONLY if condition is sweep_reclaim") {
		t.Errorf(`the "legs" schema field carries NO condition qualifier, while its siblings fvg{} and breakdown{} on the same line DO ("REQUIRED iff …"). `+
			`plan_doc.go rejects legs on every non-sweep_reclaim condition (arm_legs_sweep_reclaim_only) — row 80.`+"\n  field renders as: %s", legsWindow[:min(200, len(legsWindow))])
	}

	// The ARMED ORDERS prose pushes breakdown_continue toward arm{}; it must
	// also say those arms are SINGLE.
	for _, want := range []string{"arm SINGLE", "no legs"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the rendered prompt never says %q — breakdown_continue/breakup_continue arms are single-leg, and nothing in the prompt states it (row 80: 24 breakdown + 11 reject instances in 72h)", want)
		}
	}

	// The tokens the dispatch counted as absent from the rendered prompt
	// (counts before this wave: "EXACTLY 2 legs" 0 · "split contract" 0).
	for _, want := range []string{"EXACTLY 2 legs", "split contract"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the rendered prompt never states %q — the split contract is invisible to the author", want)
		}
	}
}

// TestClass38DeathFlipVocabularyIsDeclared — death/flip.rule genuinely has its
// OWN enum (2x5m|5m_close, plan_doc.go conditionRules) which is NOT the confirm
// enum. Two vocabularies for one concept is the trap that produced row 79, so
// the prompt must say so beside the death/flip lines rather than leave the
// reader to infer it.
func TestClass38DeathFlipVocabularyIsDeclared(t *testing.T) {
	prompt := plannerOutputContract(8, 5, true, true)
	if !strings.Contains(prompt, `"rule": "2x5m|5m_close"`) {
		t.Fatal("death/flip schema line not found — locate before editing")
	}
	if !strings.Contains(prompt, "death/flip rules use their OWN vocabulary") {
		t.Error("the prompt offers death/flip rule tokens 2x5m|5m_close and confirm rule tokens touch|1x5m_close|2x5m_close|1m_mss|time_hold with no statement that they are DIFFERENT enums — row 79 is the model moving a token between them")
	}
}
