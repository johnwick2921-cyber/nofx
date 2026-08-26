package ninjatrader

import (
	"sync/atomic"
)

// Bar persistence hook (2026-08-26) — the unblock for replay/calibration.
//
// A single late-bound persister is installed by the trader layer once the
// store exists. drainBarIngest calls it AFTER the cache write, in its own
// goroutine, so a slow or failing DB can NEVER stall the bar drain (the
// backpressure invariant) or the socket read loop. Failures are the
// persister's to log (WARN) — this package only fans the call out.

var barPersister atomic.Value // func(historical bool, symbol, tf string, bars []Bar)

// SetBarPersister installs the persistence callback (nil to detach).
func SetBarPersister(fn func(historical bool, symbol, tf string, bars []Bar)) {
	if fn == nil {
		barPersister.Store((func(bool, string, string, []Bar))(nil))
		return
	}
	barPersister.Store(fn)
}

// ClosedBarsOnly keeps the bars whose CLOSE time (T + tf duration) has passed
// at nowMs — forming bars are never persisted (the spec: zero writes on
// forming bars). T is the bar's OPEN time per the canonical cache contract.
func ClosedBarsOnly(bars []Bar, tf string, nowMs int64) []Bar {
	dur := timeframeMs(tf)
	if dur <= 0 {
		dur = 60_000
	}
	out := make([]Bar, 0, len(bars))
	for _, b := range bars {
		if b.T+dur <= nowMs {
			out = append(out, b)
		}
	}
	return out
}

// ClosedCacheTail returns up to `window` of the most-recent CLOSED bars from
// the cache (sorted ascending by T, oldest first). Live bar_update frames
// only ever carry the FORMING bar — NT8 does not re-emit the just-closed bar
// at the minute boundary — so the live persistence path reads the final
// closed bars from the cache itself (which always holds them) instead of the
// frame. A window of 8 covers multi-minute gaps between frames.
func ClosedCacheTail(get func(symbol, tf string) []Bar, symbol, tf string, nowMs int64, window int) []Bar {
	dur := timeframeMs(tf)
	if dur <= 0 {
		dur = 60_000
	}
	all := get(symbol, tf)
	closed := make([]Bar, 0, window)
	for i := len(all) - 1; i >= 0 && len(closed) < window; i-- {
		if all[i].T+dur <= nowMs {
			closed = append(closed, all[i])
		}
	}
	// Reverse to restore ascending order (InsertBars expects any order, but
	// ascending keeps logs sane).
	for l, r := 0, len(closed)-1; l < r; l, r = l+1, r-1 {
		closed[l], closed[r] = closed[r], closed[l]
	}
	return closed
}

// fanOutBarPersist invokes the installed persister in its own goroutine with
// a panic guard, so a persister failure can never propagate into the drain
// loop. warn is the package's WARN sink (injected for tests).
func fanOutBarPersist(warn func(msg string, kv ...interface{}), historical bool, symbol, tf string, bars []Bar) {
	fn, ok := barPersister.Load().(func(bool, string, string, []Bar))
	if !ok || fn == nil {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				warn("tcp_server: bar persister panicked (recovered)", "symbol", symbol, "timeframe", tf, "panic", r)
			}
		}()
		fn(historical, symbol, tf, bars)
	}()
}
