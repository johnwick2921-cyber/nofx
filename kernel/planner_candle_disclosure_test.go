package kernel

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"nofx/market"
)

// ── D1 PINS (wave BARS HORIZON, 2026-09-09) ─────────────────────────────────
//
// THE HEADING LIES, AND THE MODEL IS TOLD TO TRUST IT.
//
// BuildPlannerCandleTablesAt truncates only when the slice is LONG and bakes
// the count into the title as a literal, so "daily session candles (last 8)"
// stood over 3 rows. MEASURED on the stored prompts (planner_rejected_prompts,
// n=54 carrying a Candles block, ids 70..142): 15m rendered 12 in 54/54, 1h 12
// in 54/54, 4h 8 in 54/54 — and the daily table rendered 2 rows in 19 and 3
// rows in 35. NEVER 8, in 0 of 54.
//
// Worse, the OLDEST daily row is a PARTIAL session candle presented as a whole
// one: DailySessionBars takes the first bar it sees as the session Open with no
// completeness check. Live example, prompt id 142: the oldest daily row is
// stamped "09-07 02:39" — the session it claims opened at 09-06 17:00 CT.
//
// And kernel/planner_prompt.go tells the model "On conflict, trust the candles".
//
// A24 ABSOLUTE: these pins require the partial row to be MARKED. Nothing here
// may synthesise, interpolate or carry forward a bar to complete it.

// d1Tape builds 1m bars at every OPEN CME minute in [from, to), so a fixture
// never has to hand-model the daily break or the weekend.
func d1Tape(from, to time.Time) []market.Kline {
	var out []market.Kline
	i := 0
	for t := from; t.Before(to); t = t.Add(time.Minute) {
		if !IsCMEOpen(t) {
			continue
		}
		o := 29500.0 + float64(i%40)*0.25
		out = append(out, market.Kline{
			OpenTime: t.UnixMilli(), Open: o, High: o + 2, Low: o - 2, Close: o + 0.5, Volume: 12,
		})
		i++
	}
	return out
}

// d1Fixture reproduces the live shape of prompt id 142: a tape that starts
// MID-SESSION (02:39 CT) and ends mid-session, so the oldest daily row is
// front-truncated and the newest is still forming.
func d1Fixture(t *testing.T) ([]market.Kline, time.Time) {
	t.Helper()
	from := time.Date(2026, 9, 8, 2, 39, 0, 0, CTLocation())
	now := time.Date(2026, 9, 9, 13, 18, 13, 0, CTLocation())
	bars := d1Tape(from, now)
	if len(bars) == 0 {
		t.Fatalf("fixture built no bars")
	}
	return bars, now
}

func d1Section(t *testing.T, table, heading string) string {
	t.Helper()
	i := strings.Index(table, heading)
	if i < 0 {
		t.Fatalf("section %q absent from table:\n%s", heading, table)
	}
	rest := table[i:]
	if j := strings.Index(rest[len(heading):], "\n### "); j >= 0 {
		rest = rest[:len(heading)+j]
	}
	return rest
}

// d1Rows returns ONLY the rendered candle rows of a section (lines stamped
// "MM-DD HH:MM"), so an assertion about a ROW MARKER can never be satisfied by
// the heading — which names ⚠PARTIAL / ⏳FORMING / ❓COVERAGE-UNKNOWN in its own
// legend clause. The first version of these pins DID match the heading and a
// mutation that disabled the whole UNCOVERED branch passed (class 89).
func d1Rows(sec string) string {
	var out []string
	for _, ln := range strings.Split(sec, "\n") {
		if len(ln) > 11 && ln[2] == '-' && ln[5] == ' ' && ln[8] == ':' {
			out = append(out, ln)
		}
	}
	return strings.Join(out, "\n")
}

// PIN D1-A — THE HEADING STATES HELD-vs-CLAIMED, ON EVERY TABLE.
// A disclosure that only appears on failure is one the reader learns to skim
// past, so the COMPLETE tables must say so too.
func TestCandleHeadingsStateHeldVsClaimed(t *testing.T) {
	bars, now := d1Fixture(t)
	table := BuildPlannerCandleTablesAt(bars, 12000, now)

	for _, want := range []string{
		"### 15m — HELD 12 of 12 requested rows",
		"### 1h — HELD 12 of 12 requested rows",
		"### 4h — HELD 8 of 8 requested rows",
		"### daily session candles — HELD 2 of 8 requested rows",
	} {
		if !strings.Contains(table, want) {
			t.Errorf("heading missing %q\n--- rendered ---\n%s", want, table)
		}
	}
	// The literal count must be GONE from every heading — it is the lie.
	for _, gone := range []string{"(last 12)", "(last 8)"} {
		if strings.Contains(table, gone) {
			t.Errorf("heading still carries the baked literal %q — a heading cannot say 8 over 2 rows", gone)
		}
	}
	// SHORT is named, with the shortfall, and the absent rows are not inferable.
	if !strings.Contains(table, "SHORT BY 6") {
		t.Errorf("daily heading does not name the shortfall (want \"SHORT BY 6\")\n%s", d1Section(t, table, "### daily session candles"))
	}
}

// PIN D1-B — A PARTIAL SESSION CANDLE IS MARKED, AND NO BAR IS INVENTED.
func TestPartialSessionCandleIsMarkedNeverSynthesised(t *testing.T) {
	bars, now := d1Fixture(t)
	table := BuildPlannerCandleTablesAt(bars, 12000, now)
	sec := d1Section(t, table, "### daily session candles")

	rows := d1Rows(sec)
	if !strings.Contains(rows, "⚠PARTIAL — holds ") {
		t.Errorf("the front-truncated session ROW is NOT marked partial:\n%s", sec)
	}
	if !strings.Contains(rows, "⏳FORMING — the window has not closed") {
		t.Errorf("the still-open session ROW is NOT marked forming:\n%s", sec)
	}
	// A24: marking must not change the DATA. The rendered OHLC rows must still
	// be exactly what DailySessionBars produced — same count, same open time.
	want := DailySessionBars(bars)
	if len(want) > 8 {
		want = want[len(want)-8:]
	}
	if n := len(strings.Split(rows, "\n")); n != len(want) {
		t.Errorf("rendered %d daily rows, DailySessionBars produced %d — a row was invented or dropped", n, len(want))
	}
	for _, k := range want {
		stamp := TableTimeCT(time.UnixMilli(k.OpenTime).In(CTLocation()))
		if !strings.Contains(sec, stamp) {
			t.Errorf("daily row %q absent — the rendered rows are not DailySessionBars' rows", stamp)
		}
	}
	// The partial row must say WHY its Open is not the session open.
	if !strings.Contains(rows, "first bar HELD") {
		t.Errorf("the partial row does not disclose that its Open is the first bar HELD, not the session open:\n%s", sec)
	}
}

// PIN D1-C — THE PROMPT DISCLOSES ITS TAPE DEPTH. It disclosed it nowhere.
func TestCandleBlockDisclosesItsTape(t *testing.T) {
	bars, now := d1Fixture(t)
	table := BuildPlannerCandleTablesAt(bars, 12000, now)
	first := strings.SplitN(table, "\n", 2)[0]

	for _, want := range []string{"TAPE:", "1m bars HELD of 12000 requested", "oldest", "gaps"} {
		if !strings.Contains(first, want) {
			t.Errorf("tape line missing %q — got %q", want, first)
		}
	}
	if !strings.Contains(table, "NEVER interpolated") {
		t.Errorf("the marker legend does not state that a missing bar is never interpolated")
	}
}

// PIN D1-D — A COMPLETE, CLOSED SESSION IS NOT MARKED. A marker on everything
// is a marker on nothing.
func TestCompleteSessionRowCarriesNoMarker(t *testing.T) {
	// Two WHOLE session-days, ending exactly at a session roll so nothing forms.
	from := time.Date(2026, 9, 7, 17, 0, 0, 0, CTLocation())
	now := time.Date(2026, 9, 9, 17, 0, 0, 0, CTLocation())
	bars := d1Tape(from, now)
	table := BuildPlannerCandleTablesAt(bars, 12000, now)
	sec := d1Section(t, table, "### daily session candles")
	if r := d1Rows(sec); strings.Contains(r, "PARTIAL") || strings.Contains(r, "FORMING") || strings.Contains(r, "UNKNOWN") {
		t.Errorf("a complete closed session ROW was marked:\n%s", r)
	}
	if !strings.Contains(sec, "all held rows COMPLETE") {
		t.Errorf("a complete table does not SAY it is complete:\n%s", sec)
	}
}

// PIN D1-E — AN UNCLASSIFIABLE DAY IS UNKNOWN, NEVER GUESSED (calendar
// discipline, inherited from kernel/session_calendar.json).
func TestUncoveredCalendarYearRendersUnknownNotZero(t *testing.T) {
	yr := 2031
	if SessionCalendarCoversYear(yr) {
		t.Skipf("calendar now covers %d — pick an uncovered year", yr)
	}
	from := time.Date(yr, 9, 8, 2, 39, 0, 0, CTLocation())
	now := time.Date(yr, 9, 9, 13, 18, 0, 0, CTLocation())
	bars := d1Tape(from, now)
	if len(bars) == 0 {
		t.Skip("no open minutes generated for the uncovered year")
	}
	sec := d1Section(t, BuildPlannerCandleTablesAt(bars, 12000, now), "### daily session candles")
	rows := d1Rows(sec)
	want := fmt.Sprintf("❓COVERAGE-UNKNOWN — %d is not covered by kernel/session_calendar.json", yr)
	if !strings.Contains(rows, want) {
		t.Errorf("a session-day in a year the calendar does not cover must render %q on the ROW, never a computed-looking number:\n%s", want, rows)
	}
	// A24 — and it must NOT also print a plausible count beside the UNKNOWN.
	if strings.Contains(rows, "open 1m intervals in this window") {
		t.Errorf("an UNKNOWN row also printed a computed interval count:\n%s", rows)
	}
}

// PIN D1-F — THE HEADING AND THE ROWS MUST AGREE.
//
// Found by MUTATION, not by design: forcing RowCoverage.Marked() to false left
// every pin above green while the daily heading said "all held rows COMPLETE"
// over a row that carried "⚠PARTIAL". A heading that contradicts its own rows
// is the exact defect this wave exists to remove, one layer up.
func TestHeadingCompletenessAgreesWithTheRows(t *testing.T) {
	bars, now := d1Fixture(t)
	table := BuildPlannerCandleTablesAt(bars, 12000, now)

	for _, h := range []string{"### 15m", "### 1h", "### 4h", "### daily session candles"} {
		sec := d1Section(t, table, h)
		head := strings.SplitN(sec, "\n", 2)[0]
		rows := d1Rows(sec)
		markedRows := 0
		for _, ln := range strings.Split(rows, "\n") {
			if strings.Contains(ln, "⚠PARTIAL") || strings.Contains(ln, "⏳FORMING") || strings.Contains(ln, "❓COVERAGE-UNKNOWN") || strings.Contains(ln, "ℹ️HELD") {
				markedRows++
			}
		}
		claimsComplete := strings.Contains(head, "all held rows COMPLETE")
		if claimsComplete && markedRows > 0 {
			t.Errorf("%s heading claims \"all held rows COMPLETE\" over %d MARKED row(s):\n%s\n%s", h, markedRows, head, rows)
		}
		if !claimsComplete && markedRows == 0 {
			t.Errorf("%s heading withholds the COMPLETE claim although no row is marked:\n%s\n%s", h, head, rows)
		}
		if markedRows > 0 && !strings.Contains(head, fmt.Sprintf("%d of ", markedRows)) {
			t.Errorf("%s heading does not count its %d marked rows:\n%s", h, markedRows, head)
		}
	}
}

// PIN D1-G — A HOLE INSIDE A WINDOW IS NOT A TRUNCATED FRONT.
//
// Found in the LIVE render, not by design. Against the store's real
// 2026-09-09 01:28→13:05 CT hole the 15m row read:
//
//	⚠PARTIAL — holds 14 of the 15 open 1m intervals in this window; its Open is
//	the first bar HELD (01:15 CT), not the window open (01:15 CT)
//
// The two clock times are IDENTICAL: that row's Open is exactly the window
// open, and the missing minute is inside it. A disclosure that states a false
// reason is the same class of defect as a heading that states a false count.
func TestInteriorHoleIsNotReportedAsATruncatedOpen(t *testing.T) {
	// One whole 15m bucket, minus a single minute in its middle. The bucket
	// starts exactly on the tape's first bar, so nothing is truncated.
	start := time.Date(2026, 9, 8, 20, 0, 0, 0, CTLocation())
	var bars []market.Kline
	for i := 0; i < 15; i++ {
		if i == 7 {
			continue // the hole
		}
		ts := start.Add(time.Duration(i) * time.Minute)
		bars = append(bars, market.Kline{OpenTime: ts.UnixMilli(), Open: 29500, High: 29502, Low: 29498, Close: 29501, Volume: 10})
	}
	now := start.Add(15 * time.Minute)
	sec := d1Section(t, BuildPlannerCandleTablesAt(bars, 12000, now), "### 15m")
	rows := d1Rows(sec)
	if !strings.Contains(rows, "⚠PARTIAL") {
		t.Fatalf("a 14-of-15 window is not marked partial:\n%s", rows)
	}
	if strings.Contains(rows, "its Open is the first bar HELD") {
		t.Errorf("an INTERIOR hole was reported as a truncated front — the row's Open IS the window open:\n%s", rows)
	}
	if !strings.Contains(rows, "missing INSIDE") {
		t.Errorf("an interior hole must say the missing intervals are INSIDE the window:\n%s", rows)
	}
}

// PIN D1-H — THE TAPE AGE IS READABLE, NOT SUB-SECOND NOISE.
// The live render printed "320h36m50.387s ago"; a millisecond-precise age in a
// prompt invites a model to reason about precision that means nothing here.
func TestTapeAgeIsMinuteResolution(t *testing.T) {
	bars, _ := d1Fixture(t)
	// The clock MUST carry a sub-second part, or this pin cannot fail: the
	// first version used d1Fixture's whole-second clock and a mutation that
	// restored msDurTxt(raw) passed green (class 89).
	now := time.Date(2026, 9, 9, 13, 18, 13, 387_000_000, CTLocation())
	first := strings.SplitN(BuildPlannerCandleTablesAt(bars, 12000, now), "\n", 2)[0]
	if strings.Contains(first, ".") && strings.Contains(first, "s ago") {
		i := strings.Index(first, "s ago")
		if j := strings.LastIndex(first[:i], "."); j >= 0 && i-j <= 5 {
			t.Errorf("tape age carries sub-second noise: %q", first)
		}
	}
	if !strings.Contains(first, "m ago") && !strings.Contains(first, "h") {
		t.Errorf("tape age does not render as a duration: %q", first)
	}
}
