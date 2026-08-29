package ninjatrader

import (
	"sync"
	"testing"
	"time"
)

// TestC3SyntheticBurstClosesDroppedZero (P2/C3, pre-live-fire) — feeds N
// synthetic CLOSED bars through the REAL fanOutBarPersist queue path with a
// healthy persister and asserts the Sunday-proof invariant: closes_dropped=0
// and queue_drops=0 with every bar flushed. This is the clean-path twin of
// TestFanOutClosesSurviveWriterStall (which exercises a short stall).
func TestC3SyntheticBurstClosesDroppedZero(t *testing.T) {
	var mu sync.Mutex
	flushed := 0
	SetBarPersister(func(historical bool, symbol, tf string, bars []Bar) {
		mu.Lock()
		flushed += len(bars)
		mu.Unlock()
	})
	defer SetBarPersister(nil)
	persistDropped.Store(0)
	persistDroppedCloses.Store(0)
	persistFlushed.Store(0)

	const n = 600 // 600 concurrent close-carrying frames ≫ worker batch cap 256
	base := time.Now().Add(-2 * time.Minute).UnixMilli()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fanOutBarPersist(nil, false, "MNQ", "1m", []Bar{{
				T: base + int64(i)*60_000, O: 1, H: 2, L: 0.5, C: 1.5,
			}})
		}(i)
	}
	wg.Wait()

	deadline := time.Now().Add(5 * time.Second)
	for {
		mu.Lock()
		f := flushed
		mu.Unlock()
		if f >= n || time.Now().After(deadline) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	mu.Lock()
	f := flushed
	mu.Unlock()
	if f != n {
		t.Fatalf("persist flush shortfall: flushed=%d want %d (queue_drops=%d closes_dropped=%d)",
			f, n, persistDropped.Load(), persistDroppedCloses.Load())
	}
	if persistDroppedCloses.Load() != 0 {
		t.Fatalf("closes_dropped=%d, want 0 — closes are sacred (bar_persist.go:209 ERROR path must never fire on a healthy drain)", persistDroppedCloses.Load())
	}
	if persistDropped.Load() != 0 {
		t.Fatalf("queue_drops=%d, want 0 (clean drain through the 4096-cap queue)", persistDropped.Load())
	}
}
