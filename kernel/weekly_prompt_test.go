package kernel

import (
	"strings"
	"testing"
	"time"

	"nofx/market"
)

// ── fixture helpers ────────────────────────────────────────────────────────

func wkDoc() *WeeklyDoc {
	return &WeeklyDoc{
		Bias: "bull", Conviction: "high",
		Draw:         WeeklyDraw{Name: "PWH", Px: 30500.25},
		Invalidation: WeeklyInvalidation{Px: 30300.00, Basis: "1h close beyond 30300.00"},
		WeeklyLevels: []WeeklyLevel{{Name: "PWH", Px: 30500.25}, {Name: "PWL", Px: 29980.00}},
		Narrative:    "auction accepted above the prior high\nholding above weekly open\nfailure below the draw voids the read",
	}
}

func refsForDoc() []float64 { return []float64{30500.25, 29980.00, 30600.00, 29900.00} }

func weeklyPxBar(openTimeMs int64, o, h, l, c float64) market.Kline {
	return market.Kline{OpenTime: openTimeMs, Open: o, High: h, Low: l, Close: c, Volume: 1}
}

// ── validator r1-r6 ────────────────────────────────────────────────────────

func TestWeeklyValidatorR1Enum(t *testing.T) {
	d := wkDoc()
	if got := ValidateWeeklyDoc(d, refsForDoc(), false); got != "" {
		t.Fatalf("valid doc rejected: %s", got)
	}
	d.Bias = "sideways"
	if got := ValidateWeeklyDoc(d, refsForDoc(), false); !strings.Contains(got, "r1") {
		t.Fatalf("proving line: r1 bias enum — got %q", got)
	}
	d = wkDoc()
	d.Conviction = "extreme"
	if got := ValidateWeeklyDoc(d, refsForDoc(), false); !strings.Contains(got, "r1") {
		t.Fatalf("proving line: r1 conviction enum — got %q", got)
	}
}

func TestWeeklyValidatorR2Invalidation(t *testing.T) {
	d := wkDoc()
	d.Invalidation.Px = 0
	if got := ValidateWeeklyDoc(d, refsForDoc(), false); !strings.Contains(got, "r2") {
		t.Fatalf("proving line: r2 px>0 — got %q", got)
	}
	d = wkDoc()
	d.Invalidation.Basis = ""
	if got := ValidateWeeklyDoc(d, refsForDoc(), false); !strings.Contains(got, "r2") {
		t.Fatalf("proving line: r2 basis non-empty — got %q", got)
	}
}

func TestWeeklyValidatorR3DrawReference(t *testing.T) {
	d := wkDoc()
	d.Draw.Px = 30333.33 // matches nothing in refsForDoc
	if got := ValidateWeeklyDoc(d, refsForDoc(), false); !strings.Contains(got, "r3") {
		t.Fatalf("proving line: r3 draw-not-a-reference — got %q", got)
	}
	d = wkDoc()
	d.Draw.Px = 30500.30 // 0.05 from PWH — inside the ±1 tick (0.25) band
	if got := ValidateWeeklyDoc(d, refsForDoc(), false); got != "" {
		t.Fatalf("proving line: r3 draw within ±1 tick accepted — got %q", got)
	}
}

func TestWeeklyValidatorR4NarrativeLines(t *testing.T) {
	d := wkDoc()
	d.Narrative = "l1\nl2\nl3\nl4"
	if got := ValidateWeeklyDoc(d, refsForDoc(), false); !strings.Contains(got, "r4") {
		t.Fatalf("proving line: r4 narrative >3 lines — got %q", got)
	}
}

func TestWeeklyValidatorR5DayOfWeekTokens(t *testing.T) {
	d := wkDoc()
	d.Narrative = "Monday seasonality favors longs"
	if got := ValidateWeeklyDoc(d, refsForDoc(), false); !strings.Contains(got, "r5") {
		t.Fatalf("proving line: r5 day-of-week token — got %q", got)
	}
}

func TestWeeklyValidatorR6ThinHistoryConviction(t *testing.T) {
	d := wkDoc() // conviction high
	if got := ValidateWeeklyDoc(d, refsForDoc(), true); !strings.Contains(got, "r6") {
		t.Fatalf("proving line: r6 thin_history + high conviction — got %q", got)
	}
	d.Conviction = "low"
	if got := ValidateWeeklyDoc(d, refsForDoc(), true); got != "" {
		t.Fatalf("proving line: r6 thin_history + low conviction accepted — got %q", got)
	}
}

// ── prompt sections + facts hash ───────────────────────────────────────────

func TestWeeklyPromptSectionsAndHash(t *testing.T) {
	now := time.Date(2026, 8, 30, 16, 45, 0, 0, CTLocation())
	f := WeeklyFacts{
		Now: now, Price: 30200.5,
		Weeks: []WeekCandle{
			{WeekStart: "2026-08-03", Open: 29800, High: 30200, Low: 29700, Close: 30150, Volume: 1000, StructTag: "HH"},
			{WeekStart: "2026-08-10", Open: 30150, High: 30400, Low: 29900, Close: 30300, Volume: 1100, StructTag: "HH"},
		},
		Refs: WeeklyRefs{WeeklyOpen: 30120, PWH: 30400, PWL: 29900, PWC: 30300}, RefsOK: true,
		NWOGs:       []NWOG{{Born: "2026-08-30", Hi: 30310, Lo: 30290, CE: 30300, Filled: false}},
		IPDA:        []IPDARange{{Days: 20, High: 30600, Low: 29500, PosPct: 0.63}},
		ThinHistory: false,
	}
	f.SectionsText = RenderWeeklyFactsSections(f)
	f.FactsHash = WeeklyFactsHash(f)

	for _, want := range []string{"## Weekly candles (12 completed weeks, oldest → latest)",
		"## Weekly references", "## NWOG (last 5 weekend gaps, oldest → latest)",
		"## IPDA (trailing dealing ranges)", "## Prior week recap (facts only)",
		"weekly_open 30120.00 · PWH 30400.00 · PWL 29900.00 · PWC 30300.00",
		"2026-08-03  29800.00  30200.00  29700.00  30150.00  1000  HH"} {
		if !strings.Contains(f.SectionsText, want) {
			t.Fatalf("proving line: facts sections missing %q", want)
		}
	}
	if f.FactsHash == "" || f.FactsHash != WeeklyFactsHash(f) {
		t.Fatalf("proving line: facts hash must be deterministic sha256 hex — %q", f.FactsHash)
	}
	prompt := BuildWeeklyPrompt(f)
	for _, want := range []string{"## Instructions (THE RULES — violations are rejected)",
		"Tier-A evidence ONLY", "draw_on_liquidity MUST equal", `"invalidation":{"px":<n>,"basis":"1h close beyond <px>"}`,
		"Day-of-week reasoning is FORBIDDEN"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("proving line: weekly prompt missing %q", want)
		}
	}
}

// ── W4 invalidation twins ──────────────────────────────────────────────────

func TestWeeklyInvalidationTwins(t *testing.T) {
	base := int64(1725000000000)
	// bull doc invalidates on a CLOSED bar below px.
	bars := []market.Kline{
		weeklyPxBar(base, 30200, 30210, 30190, 30205),
		weeklyPxBar(base+60000, 30205, 30215, 30180, 30180), // closes below 30200
	}
	if !WeeklyInvalidationCrossed("bull", 30200.0, bars) {
		t.Fatalf("proving line: bull invalidation high — close 30180 < 30200 must cross")
	}
	// bear doc invalidates on a CLOSED bar above px.
	bars = []market.Kline{
		weeklyPxBar(base, 30100, 30110, 30090, 30105),
		weeklyPxBar(base+60000, 30105, 30140, 30100, 30140), // closes above 30120
	}
	if !WeeklyInvalidationCrossed("bear", 30120.0, bars) {
		t.Fatalf("proving line: bear invalidation low — close 30140 > 30120 must cross")
	}
	// the mirror directions do NOT cross: bull with closes ABOVE px.
	if WeeklyInvalidationCrossed("bull", 30080.0, bars) {
		t.Fatalf("proving line: bull doc must not cross on closes ABOVE px")
	}
	// neutral / px≤0 never crosses.
	if WeeklyInvalidationCrossed("neutral", 30200.0, bars) || WeeklyInvalidationCrossed("bull", 0, bars) {
		t.Fatalf("proving line: neutral/zero-px never cross")
	}
}

// ── W3 injection render — all 4 states ─────────────────────────────────────

func TestWeeklyContextLineAllStates(t *testing.T) {
	if got := WeeklyContextLine(nil, 0); got != "WEEKLY: none" {
		t.Fatalf("proving line: no-doc state — got %q", got)
	}
	d := wkDoc()
	if got := WeeklyContextLine(d, 0); got != "WEEKLY: bull/high · draw PWH 30500.25 · invalid 30300.00 (1h close beyond 30300.00)" {
		t.Fatalf("proving line: active state — got %q", got)
	}
	inv := wkDoc()
	inv.InvalidatedAt = "2026-08-28 10:15 CT"
	if got := WeeklyContextLine(inv, 0); got != "WEEKLY: neutral (invalidated 2026-08-28 10:15 CT)" {
		t.Fatalf("proving line: invalidated state — got %q", got)
	}
	thin := wkDoc()
	thin.ThinHistory = true
	thin.Conviction = "low"
	if got := WeeklyContextLine(thin, 3); got != "WEEKLY: thin history (3w) — low conviction" {
		t.Fatalf("proving line: thin-history state — got %q", got)
	}
	if got := WeeklyExecutorLine(d); got != "WEEKLY: bull/high · draw 30500.25" {
		t.Fatalf("proving line: executor line — got %q", got)
	}
	if got := WeeklyExecutorLine(inv); got != "WEEKLY: neutral (invalidated 2026-08-28 10:15 CT)" {
		t.Fatalf("proving line: invalidated doc renders a neutral line — got %q", got)
	}
}

// TestWeeklyExecutorLineInvalidatedRendersNeutral — F1 (2026-08-30): an
// invalidated weekly doc must stay VISIBLE in the executor prompt as
// "WEEKLY: neutral (invalidated …)", never silent (the silent "" hid the
// invalidated-bear doc through the −204pt sell-off).
func TestWeeklyExecutorLineInvalidatedRendersNeutral(t *testing.T) {
	doc := &WeeklyDoc{Bias: "neutral", Conviction: "low", Draw: WeeklyDraw{Name: "PWL", Px: 28947.75},
		Invalidation: WeeklyInvalidation{Px: 29535, Basis: "1h close beyond 29535.00"}, InvalidatedAt: "2026-08-30 17:07 CT"}
	if got := WeeklyExecutorLine(doc); got != "WEEKLY: neutral (invalidated 2026-08-30 17:07 CT)" {
		t.Fatalf("invalidated line = %q", got)
	}
	active := &WeeklyDoc{Bias: "bear", Conviction: "low", Draw: WeeklyDraw{Name: "PWL", Px: 28947.75},
		Invalidation: WeeklyInvalidation{Px: 29535, Basis: "1h close beyond 29535.00"}}
	if got := WeeklyExecutorLine(active); !strings.Contains(got, "WEEKLY: bear/low") {
		t.Fatalf("active line = %q", got)
	}
	if got := WeeklyExecutorLine(nil); got != "" {
		t.Fatalf("nil doc line = %q, want empty", got)
	}
}

// TestApplyWeeklyDOAStampsNeutralAtWrite — F5 (2026-08-30): a bias whose own
// invalidation basis is ALREADY crossed at write is stamped neutral + stamp
// time, never written stillborn (the 17:07:15 bear lived 3 seconds).
func TestApplyWeeklyDOAStampsNeutralAtWrite(t *testing.T) {
	now := time.Date(2026, 8, 30, 22, 7, 15, 0, time.UTC)
	doc := &WeeklyDoc{Bias: "bear", Conviction: "low", Invalidation: WeeklyInvalidation{Px: 29535, Basis: "1h close beyond 29535.00"}}
	breached := []market.Kline{{OpenTime: 1, Close: 29541.75}} // 1h close beyond 29535
	if !ApplyWeeklyDOA(doc, breached, now) {
		t.Fatal("DOA should stamp neutral on a breach-at-write")
	}
	if doc.Bias != "neutral" || doc.InvalidatedAt == "" {
		t.Fatalf("doc = bias %q invalidated_at %q, want neutral + stamp", doc.Bias, doc.InvalidatedAt)
	}
	// Not crossed → untouched; already invalidated → untouched; neutral bias → untouched.
	doc2 := &WeeklyDoc{Bias: "bull", Invalidation: WeeklyInvalidation{Px: 100, Basis: "1h close below 100.00"}}
	if ApplyWeeklyDOA(doc2, []market.Kline{{OpenTime: 1, Close: 101}}, now) {
		t.Fatal("no breach → no DOA stamp")
	}
	if ApplyWeeklyDOA(doc, breached, now) { // already invalidated
		t.Fatal("already-invalidated doc must not re-stamp")
	}
	doc3 := &WeeklyDoc{Bias: "neutral", Invalidation: WeeklyInvalidation{Px: 100}}
	if ApplyWeeklyDOA(doc3, []market.Kline{{OpenTime: 1, Close: 50}}, now) {
		t.Fatal("neutral bias → no DOA stamp")
	}
}
