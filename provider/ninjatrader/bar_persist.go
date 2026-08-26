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
