package trader

import (
	"os"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/market"
)

// ── VOID PARITY (2026-09-02) — THE PROMPT MUST LIST WHAT THE VALIDATOR REJECTS ─
//
// The 20:58 CT read on rev 575e9c05 was rejected "S2 breakdown_continue: a close
// came back across 29141.25 — the breakdown is void". 29141.25 IS a seated level
// (ONL, on the ranked list at 0.00 pts) and it was NOT in the prompt's VOID list.
//
// Cause: the two sides read DIFFERENT tape.
//   render   auto_trader_planner.go:2304 — sinceMs = CMESessionDayStart, 12,000 bars
//   validate breakdown_continue.go:248   — sinceMs = 0 (whole slice), 2,000 bars
// A level broken and reclaimed BEFORE the session-day boundary is void to the
// validator and invisible to the prompt. ONL is exactly that: an overnight low,
// broken overnight.
//
// The class-45 parity fixture passed 40/40 because it fed BOTH sides the same
// sinceMs. It pinned the functions; it never pinned the CALL SITES.

// onlTape reproduces the shape: a break-and-reclaim of `level` BEFORE the 17:00
// CT session-day boundary, then quiet bars above it up to `now`.
func onlTape(level float64, now time.Time) []market.Kline {
	ct := now.Location()
	start := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, ct) // overnight, pre-boundary
	var bars []market.Kline
	add := func(t time.Time, o, h, l, c float64) {
		bars = append(bars, market.Kline{OpenTime: t.UnixMilli(), Open: o, High: h, Low: l, Close: c,
			CloseTime: t.UnixMilli() + 60_000})
	}
	t := start
	// base ABOVE the level
	for i := 0; i < 20; i++ {
		add(t, level+18, level+22, level+14, level+17)
		t = t.Add(time.Minute)
	}
	// displacement DOWN through the level (the break) — several closes beyond
	for i := 0; i < 14; i++ {
		p := level - float64(4+3*i)
		add(t, p+3, p+4, p-4, p)
		t = t.Add(time.Minute)
	}
	// the RECLAIM: a close back across, still before 17:00 CT
	add(t, level-6, level+9, level-7, level+6)
	t = t.Add(time.Minute)
	// quiet bars above the level all the way to now (crosses the 17:00 boundary)
	for t.Before(now) {
		add(t, level+8, level+12, level+5, level+9)
		t = t.Add(time.Minute)
	}
	return bars
}

// sessionDayScope is the DELETED production window, kept here so the mechanism
// test can still demonstrate what it did. Never used by production code again.
func sessionDayScope(bars []market.Kline, now time.Time) kernel.VoidScope {
	start := kernel.CMESessionDayStart(now).UnixMilli()
	if len(bars) > 0 && bars[0].OpenTime > start {
		start = bars[0].OpenTime
	}
	sc := kernel.VoidScopeOf(bars)
	sc.SinceMs = start
	return sc
}

func voidListHas(v []kernel.VoidBreakdownLevel, price float64, short bool) bool {
	for _, x := range v {
		if x.Price == price && x.Short == short {
			return true
		}
	}
	return false
}

// THE PIN: whatever scope the prompt renders from, the level the validator will
// reject MUST appear in the VOID list. Same tape, same level, same instant.
func TestVoidParityCallSitesAgree_ONL29141(t *testing.T) {
	ct := kernel.CTLocation()
	now := time.Date(2026, 9, 2, 20, 58, 0, 0, ct)
	const onl = 29141.25
	bars := onlTape(onl, now)
	nowMs := now.UnixMilli()

	// --- the VALIDATOR's verdict, through the production entry point ---
	doc := &kernel.PlanDoc{Scenarios: []kernel.PlanScenario{{
		ID: "S2", Condition: "breakdown_continue", Direction: "short",
		Breakdown: &kernel.PlanBreakdownContinue{Level: onl, EntryMode: "pullback"},
	}}}
	verr := kernel.ValidateBreakdownContinueScenarios(doc, kernel.VoidScopeOf(bars), 20, onl-5, nowMs)
	if verr == nil || !strings.Contains(verr.Error(), "the breakdown is void") {
		t.Fatalf("fixture is wrong: the validator must void this level, got %v", verr)
	}
	t.Logf("validator: %v", verr)

	// --- the PROMPT's list, through the production render scope ---
	levels := []kernel.ScoredLevel{{DetectedLevel: kernel.DetectedLevel{Price: onl, Label: "ONL"}}}
	rendered := kernel.ComputeVoidBreakdownLevels(levels, kernel.VoidScopeOf(bars), nowMs)

	if !voidListHas(rendered, onl, true) {
		t.Errorf("PARITY BROKEN: the validator voids %.2f but the prompt's VOID list omits it (%d entries) — the model is told nothing and authors straight into the reject", onl, len(rendered))
	}
}

// The session-day window is the mechanism, isolated: identical bars, identical
// level, only sinceMs differs.
func TestVoidParityWindowIsTheMechanism(t *testing.T) {
	ct := kernel.CTLocation()
	now := time.Date(2026, 9, 2, 20, 58, 0, 0, ct)
	const onl = 29141.25
	bars := onlTape(onl, now)
	levels := []kernel.ScoredLevel{{DetectedLevel: kernel.DetectedLevel{Price: onl, Label: "ONL"}}}

	whole := kernel.ComputeVoidBreakdownLevels(levels, kernel.VoidScopeOf(bars), now.UnixMilli())
	dayOnly := kernel.ComputeVoidBreakdownLevels(levels, sessionDayScope(bars, now), now.UnixMilli())

	if !voidListHas(whole, onl, true) {
		t.Fatalf("sinceMs=0 (the validator's scope) must see the reclaim")
	}
	if voidListHas(dayOnly, onl, true) {
		t.Fatalf("fixture is wrong: the session-day window was supposed to MISS the pre-boundary reclaim")
	}
	t.Logf("whole-slice=%d entries · session-day-window=%d entries — the window is the bug", len(whole), len(dayOnly))
}

// SOURCE PIN — neither call site may choose its own window or bar slice again.
// The class-45 parity fixture pinned the two FUNCTIONS and passed 40/40 while
// production fed them different tape; this pins the CALL SITES.
func TestVoidParityCallSitesUseTheResolver(t *testing.T) {
	for _, f := range []string{"auto_trader_planner.go", "rootfix_shadow_ab.go"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		for _, banned := range []string{
			`ComputeVoidBreakdownLevels(scored, bars1m,`,
			`ValidateBreakdownContinueScenarios(d, bdBars,`,
			`ValidateBreakdownContinueScenarios(d, bars,`,
			`ValidateBreakdownContinueScenarios(d, scope.Bars,`,
			`ComputeVoidBreakdownLevels(scored, bars`,
		} {
			if strings.Contains(src, banned) {
				t.Errorf("%s still picks its own void tape: %q", f, banned)
			}
		}
	}
	// Positively: every void call site names the resolver.
	pb, _ := os.ReadFile("auto_trader_planner.go")
	if n := strings.Count(string(pb), "kernel.ResolveVoidScope("); n < 2 {
		t.Errorf("both void call sites must resolve the scope, found %d in auto_trader_planner.go", n)
	}
	sb, _ := os.ReadFile("rootfix_shadow_ab.go")
	if !strings.Contains(string(sb), "kernel.ResolveVoidScope(") {
		t.Error("the shadow A/B validator must read the same resolved scope")
	}

	// The deleted session-day window must not come back into production code.
	b, _ := os.ReadFile("auto_trader_planner.go")
	if strings.Contains(string(b), "func voidWindowStartMs(") {
		t.Error("voidWindowStartMs is deleted — the window belongs to the resolver")
	}
	// The resolver's scope is the VALIDATOR's: whole slice.
	if kernel.VoidScopeSinceMs() != 0 {
		t.Errorf("the validator judges the whole slice; the prompt must too (got sinceMs=%d)", kernel.VoidScopeSinceMs())
	}
	if kernel.VoidScopeBarCount() != kernel.AISVPBarCount {
		t.Errorf("bar depth must match the validator's historical slice, got %d", kernel.VoidScopeBarCount())
	}
	line := kernel.VoidScopeBootLine()
	for _, want := range []string{"void scope:", "whole-slice", "1m×2000", "parity"} {
		if !strings.Contains(line, want) {
			t.Errorf("boot line missing %q: %s", want, line)
		}
	}
	t.Logf("boot: %s", line)
}
