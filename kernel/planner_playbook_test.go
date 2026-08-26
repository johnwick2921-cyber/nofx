package kernel

import (
	"strings"
	"testing"
)

// A1-A5 (planner-contract wave 2026-08-26) — render tests for the new prompt
// sections + the chain_after WARN validator. Advisory only: WARN, never fail.

func TestBuildPlannerPromptPlaybookSections(t *testing.T) {
	in := samplePlannerInput()
	in.BiasCtxFacts = BiasContext{Price: 15600, PDC: 15550, VAH: 15650, VAL: 15400, NearestLiquidity: "ONH (+40.0)"}
	p := BuildPlannerPrompt(in)
	for _, want := range []string{
		"## BIAS-TREE", "bull-continuation", "premium/discount", "draw-on-liquidity",
		"## Priority setup — THE CHAIN", "chain_after", "raw-FVG", "displacement → FVG retrace",
		"## No-trade gates", "balance-day", "1.2×ATR", "10:30 ET", "Tier-1 news",
		"## Killzone weighting", "premium FVG window", "Thursday/Friday",
		"## STOP-DOING", "0% win evidence",
		`"chain_after"`, "bias-tree",
	} {
		if !strings.Contains(p, want) {
			t.Fatalf("planner prompt missing %q", want)
		}
	}
	// No-trade gates must stay ≤8 lines (6 bullets + header + blank = 8).
	block := p[strings.Index(p, "## No-trade gates"):]
	if i := strings.Index(block, "\n## "); i > 0 {
		block = block[:i]
	}
	if got := strings.Count(block, "\n"); got > 8 {
		t.Fatalf("no-trade gates block is %d lines (>8)", got)
	}
}

func TestRenderBiasTreeBranches(t *testing.T) {
	levels := []ScoredLevel{
		{DetectedLevel: DetectedLevel{Kind: KindPDH, Price: 15620, Label: "PDH"}, Grade: "A"},
		{DetectedLevel: DetectedLevel{Kind: KindPDL, Price: 15450, Label: "PDL"}, Grade: "A"},
		{DetectedLevel: DetectedLevel{Kind: KindPDC, Price: 15550, Label: "PDC"}, Grade: "A"},
	}
	// Branch 1: close > PDH.
	out := RenderBiasTree(15640, levels, BiasContext{Price: 15640})
	if !strings.Contains(out, "facts match branch 1 (close > PDH)") {
		t.Fatalf("branch 1 not flagged:\n%s", out)
	}
	// Branch 3: inside day, close above PDC → long LOW.
	out = RenderBiasTree(15580, levels, BiasContext{Price: 15580})
	if !strings.Contains(out, "facts match branch 3 (inside day; close long PDC → long LOW)") {
		t.Fatalf("branch 3 not flagged:\n%s", out)
	}
	// Premium: price above the 50% mark of the value range.
	out = RenderBiasTree(15600, levels, BiasContext{Price: 15600, VAH: 15650, VAL: 15400})
	if !strings.Contains(out, "PREMIUM — longs disallowed by branch 5") {
		t.Fatalf("premium not flagged:\n%s", out)
	}
	// Discount: below the 50% mark.
	out = RenderBiasTree(15420, levels, BiasContext{Price: 15420, VAH: 15650, VAL: 15400})
	if !strings.Contains(out, "DISCOUNT — shorts disallowed by branch 5") {
		t.Fatalf("discount not flagged:\n%s", out)
	}
}

func TestChainWarnings(t *testing.T) {
	doc := PlanDoc{
		Levels: []PlanLevel{{Price: 15620, Label: "PDH", Grade: "A"}, {Price: 15500, Label: "RN", Grade: "C"}},
		Scenarios: []PlanScenario{
			{ID: "S1", Condition: "sweep_reclaim", Direction: "long"},
			{ID: "S2", Condition: "fvg_entry", Direction: "long",
				Fvg: &PlanFvgEntry{Lo: 15510, Hi: 15530, OriginLevel: "RN", Direction: "long"}},
		},
	}
	// No chain_after + origin C → WARN.
	w := ChainWarnings(doc)
	if len(w) != 1 || !strings.Contains(w[0], "no sweep_reclaim precursor") {
		t.Fatalf("expected 1 chain warning, got %v", w)
	}
	// chain_after → sweep_reclaim → clean.
	doc.Scenarios[1].ChainAfter = "S1"
	if w := ChainWarnings(doc); len(w) != 0 {
		t.Fatalf("chained fvg must not warn, got %v", w)
	}
	// No chain but A-grade origin → clean.
	doc.Scenarios[1].ChainAfter = ""
	doc.Scenarios[1].Fvg.OriginLevel = "PDH"
	if w := ChainWarnings(doc); len(w) != 0 {
		t.Fatalf("A-origin fvg must not warn, got %v", w)
	}
}
