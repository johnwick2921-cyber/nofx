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

// TestFuturesPromptBoxesOverride proves each set box reaches the futures prompt,
// and that empty boxes never inject the box-only sections.
func TestFuturesPromptBoxesOverride(t *testing.T) {
	boxed := boxedFuturesEngine().BuildFuturesDecisionSystemPrompt("MNQ", 50000)
	for _, s := range []string{
		"ROLE_BOX_SENTINEL", "FREQ_BOX_SENTINEL", "ENTRY_BOX_SENTINEL", "DECISION_BOX_SENTINEL",
	} {
		if !strings.Contains(boxed, s) {
			t.Errorf("futures prompt with boxes set is missing %q", s)
		}
	}

	empty := emptyBoxFuturesEngine().BuildFuturesDecisionSystemPrompt("MNQ", 50000)
	for _, s := range []string{
		"ROLE_BOX_SENTINEL", "FREQ_BOX_SENTINEL", "ENTRY_BOX_SENTINEL", "DECISION_BOX_SENTINEL",
	} {
		if strings.Contains(empty, s) {
			t.Errorf("empty-box futures prompt unexpectedly contains %q", s)
		}
	}

	// FIXED parts must survive regardless of boxes (instrument + output format +
	// risk rules are NOT box-driven).
	for _, s := range []string{"Symbol: MNQ", "<reasoning>", "<decision>", "Hard Constraints (Risk Control)"} {
		if !strings.Contains(boxed, s) {
			t.Errorf("boxed futures prompt lost FIXED section marker %q", s)
		}
	}
}
