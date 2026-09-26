package trader

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// NOTE-leak (armed_event_pass.go:67-72) RED→GREEN — the old close() armed a
// 5s time.After on EVERY call, and once it fired with no receiver left (the
// select had already returned via done), the timer's delivery goroutine
// blocked forever: one permanent goroutine per Stop. With the fast path, no
// timer is allocated at all for an already-finished loop; with the slow path,
// the timer is stopped on return. RED: wait past the timer's firing time and
// the leaked delivery goroutine must not appear.
func TestArmedEventLoopCloseDoesNotLeakTimer(t *testing.T) {
	l := &armedEventLoop{kick: make(chan struct{}, 1), stop: make(chan struct{}), done: make(chan struct{})}
	close(l.done) // the pass already finished — the fast path
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	base := runtime.NumGoroutine()
	l.close()

	// The old code's timer fires at +5s; the new code has no timer at all.
	time.Sleep(5*time.Second + 500*time.Millisecond)
	if got := runtime.NumGoroutine(); got > base+2 {
		t.Fatalf("armed-event close leaked the 5s timer's delivery goroutine: %d goroutines (base %d)", got, base)
	}
}

// NOTE-leak (auto_trader.go:1081) RED→GREEN — the drawdown monitor must stop
// when the per-Run cancel fires (the grid-init error path calls the same
// cancelStopMonitor as Stop). Start the monitor, cancel, and the goroutine
// count must return to baseline — the old channel could only be closed by
// Stop, so the grid-init error path orphaned the monitor.
func TestDrawdownMonitorStopsOnCancel(t *testing.T) {
	at := &AutoTrader{}
	at.stopMonitorMu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	at.stopMonitorCtx = ctx
	at.stopMonitorCancel = cancel
	at.stopMonitorMu.Unlock()

	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	base := runtime.NumGoroutine()

	at.startDrawdownMonitor()
	time.Sleep(50 * time.Millisecond) // let the monitor reach its select
	at.cancelStopMonitor()

	deadline := time.Now().Add(10 * time.Second)
	for {
		if got := runtime.NumGoroutine(); got <= base+1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("drawdown monitor ignored the cancel: %d goroutines (base %d)", runtime.NumGoroutine(), base)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
