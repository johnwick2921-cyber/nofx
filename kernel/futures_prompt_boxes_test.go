package kernel

import (
	"os"
	"strings"
	"testing"

	"nofx/store"
)

const futuresEmptyGolden = "testdata/futures_mnq_empty.golden"

// emptyBoxFuturesEngine is futuresTestEngine with explicitly EMPTY prompt boxes
// (the back-compat case: the owner edited nothing).
func emptyBoxFuturesEngine() *StrategyEngine {
	e := futuresTestEngine()
	e.config.PromptSections = store.PromptSectionsConfig{}
	return e
}

// boxedFuturesEngine fills all 4 prompt boxes with sentinel text.
func boxedFuturesEngine() *StrategyEngine {
	e := futuresTestEngine()
	e.config.PromptSections = store.PromptSectionsConfig{
		RoleDefinition:   "ROLE_BOX_SENTINEL — you are a custom futures desk.",
		TradingFrequency: "FREQ_BOX_SENTINEL — at most 3 trades/session.",
		EntryStandards:   "ENTRY_BOX_SENTINEL — only A+ confluence.",
		DecisionProcess:  "DECISION_BOX_SENTINEL — checklist 1-2-3.",
	}
	return e
}

// TestFuturesPromptEmptyBoxesByteIdentical is the Change-4 BACK-COMPAT proof:
// with all 4 prompt boxes EMPTY, BuildFuturesDecisionSystemPrompt must produce
// EXACTLY the pre-change fixed futures prompt (golden captured before the
// builder read the boxes). So existing futures strategies do not change.
// Recapture with: UPDATE_GOLDEN=1 go test ./kernel -run EmptyBoxesByteIdentical
func TestFuturesPromptEmptyBoxesByteIdentical(t *testing.T) {
	got := emptyBoxFuturesEngine().BuildFuturesDecisionSystemPrompt("MNQ", 50000)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(futuresEmptyGolden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Skip("golden captured/updated")
	}
	want, err := os.ReadFile(futuresEmptyGolden)
	if err != nil {
		t.Fatalf("read golden (capture first with UPDATE_GOLDEN=1): %v", err)
	}
	if got != string(want) {
		t.Fatalf("empty-box futures prompt DIFFERS from the pre-change golden — back-compat broken")
	}
}

// TestFuturesPromptBoxesHonored proves the Chunk-1 (Option B) split: on futures,
// Trading Frequency + Entry Standards boxes ARE honored, but Role Definition +
// Decision Process are ALWAYS the fixed CME text (NOT box-driven). So a strategy
// whose Role/Decision boxes hold crypto-default text cannot leak crypto framing.
func TestFuturesPromptBoxesHonored(t *testing.T) {
	boxed := boxedFuturesEngine().BuildFuturesDecisionSystemPrompt("MNQ", 50000)

	// Frequency + Entry boxes ARE injected.
	for _, s := range []string{"FREQ_BOX_SENTINEL", "ENTRY_BOX_SENTINEL"} {
		if !strings.Contains(boxed, s) {
			t.Errorf("futures prompt should honor the %q box", s)
		}
	}
	// Role + Decision boxes are NOT injected (fixed CME text instead).
	for _, s := range []string{"ROLE_BOX_SENTINEL", "DECISION_BOX_SENTINEL"} {
		if strings.Contains(boxed, s) {
			t.Errorf("futures prompt must NOT inject %q (Role/Decision are fixed on futures)", s)
		}
	}
	// FIXED markers present regardless of boxes.
	for _, s := range []string{
		"professional CME", "Symbol: MNQ", "# Decision Process", "<reasoning>", "<decision>", "Hard Constraints (Risk Control)",
	} {
		if !strings.Contains(boxed, s) {
			t.Errorf("boxed futures prompt lost FIXED marker %q", s)
		}
	}

	// Empty boxes never inject the Frequency/Entry sections.
	empty := emptyBoxFuturesEngine().BuildFuturesDecisionSystemPrompt("MNQ", 50000)
	for _, s := range []string{"FREQ_BOX_SENTINEL", "ENTRY_BOX_SENTINEL"} {
		if strings.Contains(empty, s) {
			t.Errorf("empty-box futures prompt unexpectedly contains %q", s)
		}
	}
}

// TestFuturesPromptNoCryptoLeak is the Chunk-1 regression proof: a strategy whose
// Role + Decision boxes hold the CRYPTO-default text (the real-world case that
// leaked "professional cryptocurrency AI" / "候选币种" into the futures prompt
// after Change 4) must render the FIXED CME role + decision, with NO crypto line.
func TestFuturesPromptNoCryptoLeak(t *testing.T) {
	e := futuresTestEngine()
	e.config.PromptSections = store.PromptSectionsConfig{
		RoleDefinition:  "# You are a professional cryptocurrency trading AI\n\nYour task is to make trading decisions.",
		DecisionProcess: "# 决策流程\n1. 扫描候选币种 + 多时间框架",
	}
	p := e.BuildFuturesDecisionSystemPrompt("MNQ", 50000)
	if strings.Contains(strings.ToLower(p), "cryptocurrency") {
		t.Error("futures prompt LEAKED 'cryptocurrency' from the Role box")
	}
	if strings.Contains(p, "候选币种") {
		t.Error("futures prompt LEAKED '候选币种' from the Decision box")
	}
	if !strings.Contains(p, "professional CME") {
		t.Error("futures prompt must lead with the FIXED CME role")
	}
}

// TestFuturesSubModes proves Chunk-2: the 3 futures sub-modes route to the
// futures builder; "futures"/"futures-balanced" == the golden (byte-identical);
// aggressive/conservative add ONLY their "## Mode:" block (TEXT-only — nothing
// else in the prompt changes, so Risk Control / output format are untouched).
func TestFuturesSubModes(t *testing.T) {
	e := emptyBoxFuturesEngine() // empty boxes → only the mode block can vary
	balanced := e.BuildSystemPrompt(50000, "futures", "MNQ")
	balancedExplicit := e.BuildSystemPrompt(50000, "futures-balanced", "MNQ")
	aggressive := e.BuildSystemPrompt(50000, "futures-aggressive", "MNQ")
	conservative := e.BuildSystemPrompt(50000, "futures-conservative", "MNQ")

	golden, err := os.ReadFile(futuresEmptyGolden)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if balanced != string(golden) {
		t.Error("variant 'futures' must equal the golden (no-block default)")
	}
	if balancedExplicit != balanced {
		t.Error("'futures-balanced' must equal 'futures'")
	}
	if !strings.Contains(aggressive, "## Mode: Aggressive") {
		t.Error("futures-aggressive missing the Aggressive block")
	}
	if !strings.Contains(conservative, "## Mode: Conservative") {
		t.Error("futures-conservative missing the Conservative block")
	}
	// still the FUTURES prompt (CME role), never crypto.
	for _, p := range []string{aggressive, conservative} {
		if !strings.Contains(p, "professional CME") || strings.Contains(strings.ToLower(p), "cryptocurrency") {
			t.Error("a futures sub-mode is not the futures prompt")
		}
	}
	// TEXT-ONLY proof: removing exactly the mode block yields the balanced prompt.
	if strings.Replace(aggressive, futuresModeAggressive, "", 1) != balanced {
		t.Error("futures-aggressive must equal balanced + ONLY the Aggressive block")
	}
	if strings.Replace(conservative, futuresModeConservative, "", 1) != balanced {
		t.Error("futures-conservative must equal balanced + ONLY the Conservative block")
	}
}
