package ninjatrader

import (
	"sync/atomic"
	"testing"
	"time"
)

// ClosedBarsOnly: closed bars persist, forming bars never do.
func TestClosedBarsOnly(t *testing.T) {
	// 1m bars with OPEN time T: a bar closes at T+60s.
	now := int64(1_800_000_000_000)
	closed := Bar{T: now - 90_000}
	forming := Bar{T: now - 10_000}
	got := ClosedBarsOnly([]Bar{closed, forming}, "1m", now)
	if len(got) != 1 || got[0].T != closed.T {
		t.Fatalf("ClosedBarsOnly = %+v want only the closed bar", got)
	}
}

// fanOutBarPersist: a panicking persister is recovered and never propagates;
// the drain loop stays alive (the warn sink fires).
func TestFanOutBarPersistRecoversPanic(t *testing.T) {
	var warned atomic.Int64
	warn := func(msg string, kv ...interface{}) { warned.Add(1) }
	SetBarPersister(func(historical bool, symbol, tf string, bars []Bar) {
		panic("injected persister failure")
	})
	defer SetBarPersister(nil)
	fanOutBarPersist(warn, false, "MNQ", "1m", []Bar{{T: 1}})
	deadline := time.Now().Add(2 * time.Second)
	for warned.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if warned.Load() == 0 {
		t.Fatal("panic was not recovered / warn sink never fired")
	}
}

// fanOutBarPersist: no persister installed → no-op, no panic.
func TestFanOutBarPersistNoop(t *testing.T) {
	SetBarPersister(nil)
	fanOutBarPersist(func(string, ...interface{}) {}, false, "MNQ", "1m", []Bar{{T: 1}})
}
