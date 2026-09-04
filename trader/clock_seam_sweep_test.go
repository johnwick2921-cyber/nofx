// CLASS 60 SWEEP — the remaining wall-clock seams.
//
// Class 60's sweep listed eight test files that "build a fixed time.Date clock
// and call an entry point that reads time.Now()". Measured function by function
// (real brace-matched bodies, not a line window), the list resolves to:
//
//   REAL SEAMS (2)      latestClosedPrimaryBarMs · observeTransitionStanddown
//   LEGITIMATE (1)      ForceReset — its two time.Now() calls are a wall-clock
//                       TIMEOUT (deadline := now+max; for ... Before(deadline)).
//                       Converting that to At(now) would break the wait: a fixed
//                       clock never advances, so the loop would spin forever or
//                       not at all. A timeout measures elapsed REAL time and is
//                       supposed to read the wall clock.
//   FALSE POSITIVES (5) eodFlatCT, lastEntryCT, MaybeResetDaily, CanForceReset,
//                       PlannerReadInFlight, SetClockContext, BuildSystemPrompt
//                       — all either take their clock already or read none.
//
// These pin the two real ones: the …At(now) variant must honour the clock it is
// GIVEN, which is the only way to tell a fixed-clock test from a test that
// happens to pass at this hour.

package trader

import (
	"testing"
	"time"

	"nofx/market"
)

func TestLatestClosedPrimaryBarHonoursTheGivenClock(t *testing.T) {
	base := time.Date(2026, 9, 3, 14, 0, 0, 0, time.UTC)
	// Three bars closing at +1m, +2m, +3m from base.
	bars := []market.Kline{
		{CloseTime: base.Add(1 * time.Minute).UnixMilli()},
		{CloseTime: base.Add(2 * time.Minute).UnixMilli()},
		{CloseTime: base.Add(3 * time.Minute).UnixMilli()},
	}
	prev := market.FuturesBarsProvider
	market.FuturesBarsProvider = func(sym, tf string, n int) []market.Kline { return bars }
	t.Cleanup(func() { market.FuturesBarsProvider = prev })

	at := &AutoTrader{}

	// At base+2m30s exactly two bars have CLOSED, so the answer is the +2m bar.
	got, ok := at.latestClosedPrimaryBarMsAt(base.Add(150 * time.Second))
	if !ok {
		t.Fatalf("expected a closed bar at base+2m30s")
	}
	if want := bars[1].CloseTime; got != want {
		t.Errorf("at base+2m30s got %d, want the +2m bar %d", got, want)
	}

	// Move the CALLER's clock forward and the answer must move with it. If the
	// function read time.Now() this assertion would depend on the hour the suite
	// runs, which is exactly the class-60 defect.
	got, ok = at.latestClosedPrimaryBarMsAt(base.Add(10 * time.Minute))
	if !ok || got != bars[2].CloseTime {
		t.Errorf("at base+10m got %d/%v, want the +3m bar %d", got, ok, bars[2].CloseTime)
	}

	// Before any bar has closed there is no answer — and "no answer" must be
	// ok=false, never a plausible zero.
	if v, ok := at.latestClosedPrimaryBarMsAt(base); ok {
		t.Errorf("before any close, want ok=false, got %d", v)
	}
}

// The entry point still reads the wall clock exactly once and delegates — that
// is the shape class 60 asks for, not the removal of time.Now() from the process.
func TestLatestClosedPrimaryBarEntryPointStillWorks(t *testing.T) {
	bars := []market.Kline{{CloseTime: time.Now().Add(-time.Minute).UnixMilli()}}
	prev := market.FuturesBarsProvider
	market.FuturesBarsProvider = func(sym, tf string, n int) []market.Kline { return bars }
	t.Cleanup(func() { market.FuturesBarsProvider = prev })

	at := &AutoTrader{}
	if got, ok := at.latestClosedPrimaryBarMs(); !ok || got != bars[0].CloseTime {
		t.Errorf("entry point = %d/%v, want %d/true", got, ok, bars[0].CloseTime)
	}
}

// A nil provider is "warming", not "no bars ever" — ok=false, no panic.
func TestLatestClosedPrimaryBarSurvivesANilProvider(t *testing.T) {
	prev := market.FuturesBarsProvider
	market.FuturesBarsProvider = nil
	t.Cleanup(func() { market.FuturesBarsProvider = prev })
	at := &AutoTrader{}
	if _, ok := at.latestClosedPrimaryBarMsAt(time.Now()); ok {
		t.Errorf("a nil provider must report ok=false")
	}
}
