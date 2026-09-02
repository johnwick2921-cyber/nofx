package kernel

import (
	"strings"
	"testing"
)

// ── REPAIR-PARSE (2026-09-02) — THE REAL DEFECTS, REPLAYED ───────────────────
// The dispatch's C1 said "repair output unparseable 59%". Measured over every
// repair attempt in the journals (n=28, 2026-09-01 → 09-02): 18 rejected at
// the parse/schema step (64%), but only ONE was a packaging failure and it was
// a TYPE error (`cannot unmarshal number 0.5 into … PlanArmLeg.size of type
// int`) — not fences, not prose, not a fragment. The other 17 PARSED FINE and
// were rejected on field values, and 10 of those 17 are confirm-rule
// VOCABULARY errors that the repair prompt never gave the model the words for.
// These pins hold the real defects.

// realRepairRejects are the verbatim validator reasons that drove attempt ≥2
// into a repair call, with the count observed in the journals.
var realRepairRejects = []struct {
	reason string
	n      int
	want   string // the law excerpt the repair prompt MUST carry for it
}{
	{`scenario[0].confirm.rule "1x5m_close" — fade_requires_touch (a reject fade enters on the touch at the level, never on a close-confirm: touch-entry at the level (limit), stop behind structure by ≥2 ticks)`, 5, "ENTRY-LAW CONFIRM"},
	{`scenario[3].confirm2.rule "2x5m" invalid (touch|1x5m_close|2x5m_close|1m_mss|time_hold)`, 2, "CONFIRM-RULE VOCABULARY"},
	{`scenario[3].confirm2.rule "displacement" invalid (touch|1x5m_close|2x5m_close|1m_mss|time_hold)`, 3, "CONFIRM-RULE VOCABULARY"},
	{`scenario[1].confirm2.rule "1m_mss" not allowed for breakdown_continue — entry law: 1 confirming close + displacement ≥ BD_MIN_DISP_ATR`, 2, "ENTRY-LAW CONFIRM"},
	{`arm legs on reject — arm_legs_sweep_reclaim_only (the split entry is the sweep_reclaim contract; other conditions arm single)`, 1, "ARM-SPLIT LAW"},
	{`arm on S2 needs EXACTLY 2 legs (split contract), got 1`, 2, "ARM-SPLIT LAW"},
	{`arm on S1 top-level entry/stop/target must equal leg 1's (legacy readers read the top-level)`, 1, "ARM-SPLIT LAW"},
	{`S1 breakdown_continue: a close came back across 29021.25 — the breakdown is void; author a ` + "`reject`" + ` play instead`, 1, "BREAKDOWN-CONTINUE LAW"},
}

// F1 — every real defect gets a RELEVANT law excerpt. Today lawExcerptsFor is
// a first-match switch whose cases miss "invalid (" and "fade_requires_touch",
// so 10 of 18 repairs received a generic excerpt about level labels and
// targets while the defect was a confirm-rule token.
func TestRepairParsePinRealHeads(t *testing.T) {
	generic := "Copy the machine table's labels and prices"
	misrouted, total := 0, 0
	for _, c := range realRepairRejects {
		total += c.n
		got := lawExcerptsFor(c.reason)
		if !strings.Contains(got, c.want) {
			misrouted += c.n
			t.Errorf("REPAIR-PARSE: %d× defect got the WRONG law.\n  reason: %s\n  want excerpt: %s\n  got: %s",
				c.n, rpTrunc(c.reason, 90), c.want, rpTrunc(got, 120))
		}
		if strings.Contains(got, generic) && c.want != "" {
			t.Errorf("REPAIR-PARSE: %d× defect fell through to the GENERIC excerpt (levels/labels/targets) while the defect is %s", c.n, c.want)
		}
	}
	t.Logf("REPAIR-PARSE F1: %d of %d repair defects misrouted to an irrelevant law excerpt", misrouted, total)
}

// F2 — the repair prompt must carry the class-34 vocabulary line. It never has:
// LiveConditionsLine is appended only by the re-author tail
// (trader/auto_trader_planner.go plannerRejectBlock), so the DEFAULT retry has
// run without the condition vocabulary since class 34 shipped.
func TestRepairPromptCarriesConditionVocabulary(t *testing.T) {
	live := []string{"reclaim", "hold", "reject", "sweep_reclaim"}
	p := BuildPlannerRepairPrompt("{}", `scenario[0].condition "reject_retest" invalid`, live)
	if !strings.Contains(p, "Valid conditions: [reclaim, hold, reject, sweep_reclaim]") {
		t.Fatalf("repair prompt lacks the class-34 vocabulary line:\n%s", rpTrunc(p, 400))
	}
	if !strings.Contains(p, "do NOT combine condition names") {
		t.Fatalf("vocabulary line lost its instruction")
	}
}

// The confirm-rule enum must be stated whenever a confirm-rule token was
// rejected — the model cannot pick a legal token it was never shown.
func TestRepairPromptStatesConfirmEnumOnRuleDefects(t *testing.T) {
	p := BuildPlannerRepairPrompt("{}", `scenario[3].confirm2.rule "displacement" invalid (touch|1x5m_close|2x5m_close|1m_mss|time_hold)`, nil)
	for _, tok := range []string{"touch", "1x5m_close", "2x5m_close", "1m_mss", "time_hold"} {
		if !strings.Contains(p, tok) {
			t.Fatalf("repair prompt must state the confirm enum; missing %q", tok)
		}
	}
	if !strings.Contains(p, "confirm") || !strings.Contains(p, "death") {
		t.Fatalf("the prompt must separate the confirm enum from the death/flip enum — that is the row-79 defect")
	}
}

// E1 — the return contract is stated at the TOP and the BOTTOM (lost-in-the-
// middle): one complete plan JSON document, no prose, no fences.
func TestRepairPromptRepeatsTheReturnContract(t *testing.T) {
	p := BuildPlannerRepairPrompt("{}", "x", nil)
	head, tail := p[:rpMin(len(p), 500)], p[rpMax(0, len(p)-500):]
	for name, part := range map[string]string{"head": head, "tail": tail} {
		if !strings.Contains(part, "COMPLETE") || !strings.Contains(part, "JSON") {
			t.Errorf("%s must restate the full-document contract:\n%s", name, rpTrunc(part, 200))
		}
		if !strings.Contains(strings.ToLower(part), "no prose") || !strings.Contains(part, "```") {
			t.Errorf("%s must forbid prose and fences explicitly:\n%s", name, rpTrunc(part, 200))
		}
	}
}

func rpTrunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func rpMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func rpMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// E7 / R0 — THE RIDER PIN, from chain 4 (2026-09-02 14:23 CT). The incoming
// error is a VOID BREAKDOWN, not a confirm defect; the document carries a
// confirm object; the repair prompt must still state the confirm enum. Before
// the rider it did not, and the repair answered with an illegal confirm rule.
func TestClass46RiderConfirmEnumOnEveryRepairWithConfirm(t *testing.T) {
	chain4Doc := `{"reasoning":"r","bias":{"direction":"short"},"levels":[],"scenarios":[` +
		`{"id":"S2","condition":"breakdown_continue","direction":"short",` +
		`"confirm":{"rule":"1x5m_close","ref_price":29167.66,"side":"below"}}]}`
	chain4Err := "S2 breakdown_continue: a close came back across 29167.66 — the breakdown is void; author a `reject` play instead"

	p := BuildPlannerRepairPrompt(chain4Doc, chain4Err, []string{"reject", "hold"})
	if !strings.Contains(p, "BREAKDOWN-CONTINUE LAW") {
		t.Fatal("the incoming defect's own law must still be attached")
	}
	if !strings.Contains(p, "CONFIRM-RULE VOCABULARY") {
		t.Fatalf("RIDER: the document carries a confirm object, so the confirm enum MUST be attached even though the error was a breakdown defect.\n%s", rpTrunc(p, 600))
	}
	for _, tok := range []string{"touch", "1x5m_close", "2x5m_close", "1m_mss", "time_hold"} {
		if !strings.Contains(p, tok) {
			t.Fatalf("confirm enum incomplete, missing %q", tok)
		}
	}
	// A document with NO confirm object must not pay the ~60 tokens.
	noConfirm := `{"reasoning":"r","bias":{"direction":"short"},"levels":[],"scenarios":[{"id":"S1","condition":"hold"}]}`
	if q := BuildPlannerRepairPrompt(noConfirm, chain4Err, nil); strings.Contains(q, "CONFIRM-RULE VOCABULARY") {
		t.Fatal("no confirm object in the document — the enum must not be attached")
	}
	// No double-attach when the error itself already routed it.
	confErr := `scenario[1].confirm.rule "displacement" invalid (touch|1x5m_close|2x5m_close|1m_mss|time_hold)`
	r := BuildPlannerRepairPrompt(chain4Doc, confErr, nil)
	if n := strings.Count(r, "CONFIRM-RULE VOCABULARY"); n != 1 {
		t.Fatalf("the enum must appear exactly once, got %d", n)
	}
}
