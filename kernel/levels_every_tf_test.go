// W-TF — every detector, every timeframe. RED-first pins.
//
// The owner's standing rule is that every timeframe the store holds feeds every
// computation. What existed before this wave: DetectHTFLevels already looped
// per timeframe (G2/G3, 2026-08-24), but isHTFDetectionTF gated it to 15m…12h,
// so 1d/3d/1w could not reach the map at all — measured on 2026-09-10, 0 of 297
// stored plans carried a daily or weekly level.
//
// These tests pin the four things this wave changes and nothing else:
//
//	E1  a daily timeframe reaches the map, tagged tf=1d
//	E3  timeframe is part of a level's identity — same kind and price on two
//	    timeframes are TWO levels, not one dedupe casualty
//	E4  a timeframe with too few bars for the definition emits nothing AND says
//	    so — never a level from a partial window
//	C6  the tier a daily level is classified into, and the HTF flag it carries
//
// A28: this file owns ONE clock. Every bar timestamp and every `now` derives
// from tfTestNow() — a moment inside a regular session and outside every band,
// so no time predicate reads differently between two assertions.
//
// A24: no fixture retypes a production constant. Tier names, multipliers and
// the detection-gate membership are read from the code under test.
package kernel

import (
	"strings"
	"testing"
	"time"

	"nofx/market"
)

// tfTestNow is this file's single clock: 2026-09-10 13:30 CT, a Thursday inside
// the regular session. Bars are built backwards from it so every bar is closed.
func tfTestNow() time.Time {
	return time.Date(2026, 9, 10, 13, 30, 0, 0, CTLocation())
}

const (
	tfTestMinuteMs = int64(60 * 1000)
	tfTestDayMs    = 24 * 60 * tfTestMinuteMs
)

// barsWithTwoEqualPivotHighs builds n closed bars on a `stepMs` grid ending
// before `end`, shaped so that EqualHighsLows finds at least two strict pivot
// highs at the SAME price — the cheapest input that makes a detector fire on
// any timeframe. The shape is identical whatever the timeframe, which is the
// point: 12e says hold the family definition constant across timeframes.
func barsWithTwoEqualPivotHighs(n int, base float64, stepMs int64, end time.Time) []market.Kline {
	out := make([]market.Kline, 0, n)
	endMs := end.UnixMilli()
	for i := 0; i < n; i++ {
		openMs := endMs - int64(n-i)*stepMs
		o := base
		h := base + 2
		l := base - 2
		// Two isolated peaks at the same high, far enough apart to both be
		// strict pivots with k=2 (indices 3 and 8 of a >=12 bar window).
		if i == 3 || i == 8 {
			h = base + 20
		}
		out = append(out, market.Kline{
			OpenTime:  openMs,
			Open:      o,
			High:      h,
			Low:       l,
			Close:     base,
			Volume:    1000,
			CloseTime: openMs + stepMs - 1,
		})
	}
	return out
}

// fetchFor returns a fetch func that serves the same bar shape for one
// timeframe and nothing for any other, so a test states exactly which
// timeframe it is exercising.
func fetchFor(tf string, bars []market.Kline) func(string, int) []market.Kline {
	return func(want string, _ int) []market.Kline {
		if want == tf {
			return bars
		}
		return nil
	}
}

// ---------------------------------------------------------------- E1

// TestE1_DailyTimeframeReachesTheMap — RED before this wave: isHTFDetectionTF
// rejects "1d", so DetectHTFLevels returns nothing however good the bars are.
func TestE1_DailyTimeframeReachesTheMap(t *testing.T) {
	now := tfTestNow()
	daily := barsWithTwoEqualPivotHighs(30, 29500, tfTestDayMs, now)

	got := DetectHTFLevels(fetchFor("1d", daily), []string{"1d"}, "MNQ", now)
	if len(got) == 0 {
		t.Fatalf("no level detected on 1d from %d daily bars — the daily timeframe never reaches the map", len(daily))
	}
	for _, l := range got {
		if l.TF != "1d" {
			t.Errorf("level %s @%.2f carries TF=%q, want %q — a level must name the timeframe it formed on", l.Kind, l.Price, l.TF, "1d")
		}
	}
}

// TestE1b_WeeklyTimeframeReachesTheMap — the top of the range the owner ruled.
func TestE1b_WeeklyTimeframeReachesTheMap(t *testing.T) {
	now := tfTestNow()
	weekly := barsWithTwoEqualPivotHighs(30, 29500, 7*tfTestDayMs, now)

	got := DetectHTFLevels(fetchFor("1w", weekly), []string{"1w"}, "MNQ", now)
	if len(got) == 0 {
		t.Fatalf("no level detected on 1w from %d weekly bars", len(weekly))
	}
	for _, l := range got {
		if l.TF != "1w" {
			t.Errorf("level %s @%.2f carries TF=%q, want %q", l.Kind, l.Price, l.TF, "1w")
		}
	}
}

// TestE1c_SubFifteenStaysOut — the owner's ruling kept the existing gate's
// reason: intraday noise below 15m adds nothing to HTF detection, and base
// swings continue to serve 5m/15m. This pins the gate so a later widening is a
// decision rather than an accident.
func TestE1c_SubFifteenStaysOut(t *testing.T) {
	now := tfTestNow()
	for _, tf := range []string{"1m", "3m", "5m"} {
		bars := barsWithTwoEqualPivotHighs(30, 29500, tfTestMinuteMs, now)
		if got := DetectHTFLevels(fetchFor(tf, bars), []string{tf}, "MNQ", now); len(got) != 0 {
			t.Errorf("tf %s produced %d HTF level(s); sub-15m must stay out of HTF detection", tf, len(got))
		}
	}
}

// ---------------------------------------------------------------- E3

// TestE3_TimeframeIsPartOfIdentity — two levels of the SAME kind at the SAME
// price on DIFFERENT timeframes are two distinct references, not one. Before
// this wave dedupeSameKind keyed on (kind, price±tick) alone, so a 1d order
// block and a 1h order block at one price collapsed to whichever the detector
// emitted first — and the survivor kept the loser's timeframe silently.
func TestE3_TimeframeIsPartOfIdentity(t *testing.T) {
	in := []DetectedLevel{
		{Kind: KindOB, Price: 29500, Lo: 29500, Hi: 29500, Label: "OB·1h", TF: "1h"},
		{Kind: KindOB, Price: 29500, Lo: 29500, Hi: 29500, Label: "OB·1d", TF: "1d"},
	}
	out := dedupeSameKind(in)
	if len(out) != 2 {
		t.Fatalf("dedupeSameKind collapsed %d levels to %d — a 1h and a 1d level at one price are two references, not a duplicate", len(in), len(out))
	}
	seen := map[string]bool{}
	for _, l := range out {
		seen[l.TF] = true
	}
	if !seen["1h"] || !seen["1d"] {
		t.Errorf("survivors carry timeframes %v, want both 1h and 1d", seen)
	}
}

// TestE3b_SameTimeframeStillDedupes — the other half of the identity change.
// Widening the key must not stop the thing the key was added for (register S4:
// the dual nPOC emission paths could seat one POC twice).
func TestE3b_SameTimeframeStillDedupes(t *testing.T) {
	in := []DetectedLevel{
		{Kind: KindOB, Price: 29500.00, Lo: 29500, Hi: 29500, Label: "OB·1h", TF: "1h"},
		{Kind: KindOB, Price: 29500.10, Lo: 29500.10, Hi: 29500.10, Label: "OB·1h", TF: "1h"},
	}
	if out := dedupeSameKind(in); len(out) != 1 {
		t.Fatalf("two same-kind same-tf levels within a tick produced %d survivors, want 1", len(out))
	}
}

// ---------------------------------------------------------------- E4

// TestE4_PartialWindowEmitsNothingAndSaysSo — a timeframe holding fewer bars
// than the definition needs must produce NO level and must record why. Silence
// is the failure mode this pins: before the wave the loop `continue`d with no
// record, so "1w produced nothing" and "1w was never read" were the same
// observation from outside.
func TestE4_PartialWindowEmitsNothingAndSaysSo(t *testing.T) {
	now := tfTestNow()
	tooFew := barsWithTwoEqualPivotHighs(3, 29500, tfTestDayMs, now)

	levels, report := DetectHTFLevelsReport(fetchFor("1d", tooFew), []string{"1d"}, "MNQ", now)
	if len(levels) != 0 {
		t.Fatalf("a %d-bar window produced %d level(s); a partial window must emit nothing", len(tooFew), len(levels))
	}
	skip, ok := report.Skipped["1d"]
	if !ok || skip == "" {
		t.Fatalf("no skip reason recorded for 1d; report=%+v — a timeframe that emits nothing must say why", report)
	}
	// The reason must be the PARTIAL WINDOW, not "outside the detection set".
	// Before the gate admitted 1d this assertion passed for the wrong reason —
	// a skip was recorded, but because the timeframe was rejected outright. A
	// test that cannot tell those two apart would go green the day someone
	// narrows the gate again.
	if !strings.Contains(skip, "closed bars") {
		t.Errorf("1d skipped with %q; want the partial-window reason naming the bar shortfall, not a gate rejection", skip)
	}
	if _, counted := report.Counts["1d"]; counted {
		t.Errorf("1d appears in Counts as well as Skipped; a timeframe belongs to exactly one so that 'produced nothing' and 'was never read' stay distinguishable")
	}
}

// TestE4b_ReportCountsPerTimeframe — D6's boot line reads per-tf counts from
// this report rather than recomputing them, so the line cannot drift from the
// detection that produced it (A11: boot lines are READ, never literal).
func TestE4b_ReportCountsPerTimeframe(t *testing.T) {
	now := tfTestNow()
	daily := barsWithTwoEqualPivotHighs(30, 29500, tfTestDayMs, now)

	levels, report := DetectHTFLevelsReport(fetchFor("1d", daily), []string{"1d"}, "MNQ", now)
	if report.Counts["1d"] != len(levels) {
		t.Errorf("report says 1d=%d, detection returned %d levels — the boot line would print a number nothing produced", report.Counts["1d"], len(levels))
	}
	if len(levels) == 0 {
		t.Fatalf("fixture produced no levels; this test cannot distinguish a correct zero from a broken fetch")
	}
}

// ---------------------------------------------------------------- C6 / D5

// TestC6_DailyInheritsTheFourHourTier — the owner's ruling. Before the wave,
// zoneTierFor("1d") fell through to the "1m" noise floor: multiplier 1.0,
// KindOB evidence 0.40 against 4h's 0.72, and the zone grader's default clause
// forced grade C. A daily level would have been the weakest thing on the map.
//
// The tier VALUES are not asserted here — this wave changes no weight. What is
// asserted is which tier the classification lands in, read from the production
// table rather than retyped (A24).
func TestC6_DailyInheritsTheFourHourTier(t *testing.T) {
	for _, tf := range []string{"1d", "3d", "1w"} {
		if got := zoneTierFor(tf); got != "4h" {
			t.Errorf("zoneTierFor(%q) = %q, want %q — a daily/weekly level must not be graded at the 1m noise floor", tf, got, "4h")
		}
	}
	// The tier must be a real key in the production multiplier table, or the
	// classification resolves to a missing-map zero.
	if _, ok := zoneTFMult["4h"]; !ok {
		t.Fatalf("zoneTFMult has no %q key — the tier this wave classifies daily levels into does not exist", "4h")
	}
}

// TestC6b_KnownTiersUnmoved — the classification change must touch ONLY inputs
// that were previously impossible. Every timeframe that already had an answer
// keeps it.
func TestC6b_KnownTiersUnmoved(t *testing.T) {
	for tf, want := range map[string]string{
		"":    "1m",
		"1m":  "1m",
		"3m":  "1m",
		"5m":  "1m",
		"15m": "15m",
		"30m": "15m",
		"1h":  "1h",
		"2h":  "1h",
		"4h":  "4h",
		"6h":  "4h",
		"8h":  "4h",
		"12h": "4h",
	} {
		if got := zoneTierFor(tf); got != want {
			t.Errorf("zoneTierFor(%q) = %q, want %q — this wave must not move a timeframe that already had a tier", tf, got, want)
		}
	}
}

// TestC6c_DailyCarriesTheHTFFlag — the 1.2× fires on DetectedLevel.HTF, which
// tagHTFLevel sets for every timeframe the detection gate admits. Daily levels
// join that set by the same route as 4h, with no change to the multiplier
// itself. Round 12 (12c) says the 1.2× is untested; this asserts only that a
// daily level is treated like the other higher timeframes, not that it should
// be.
func TestC6c_DailyCarriesTheHTFFlag(t *testing.T) {
	now := tfTestNow()
	daily := barsWithTwoEqualPivotHighs(30, 29500, tfTestDayMs, now)
	got := DetectHTFLevels(fetchFor("1d", daily), []string{"1d"}, "MNQ", now)
	if len(got) == 0 {
		t.Fatalf("no daily level produced; cannot assert the HTF flag")
	}
	for _, l := range got {
		if !l.HTF {
			t.Errorf("daily level %s @%.2f has HTF=false — it would miss the multiplier every other higher timeframe receives", l.Kind, l.Price)
		}
	}
}
