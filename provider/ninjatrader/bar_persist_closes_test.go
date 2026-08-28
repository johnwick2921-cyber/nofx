package ninjatrader

import (
	"sync"
	"testing"
	"time"
)

// F2 (LONDON-FORENSICS 2026-08-28) — CLOSES ARE SACRED fixtures.
//
// The 06:09 event dropped 8 CLOSED bars because the queue-full path dropped
// close-carrying batches while the GORM writer was stalled. The fixed path
// bounded-blocks (backpressure) for close-carrying batches and only counts a
// close drop after a LOUD last resort. A writer stall shorter than the retry
// window must therefore produce ZERO close drops.

func TestFanOutClosesSurviveWriterStall(t *testing.T) {
	stalled := make(chan struct{})
	release := make(chan struct{})
	var first sync.Once
	releaseOnce := sync.Once{}
	SetBarPersister(func(historical bool, symbol, tf string, bars []Bar) {
		first.Do(func() {
			close(stalled)
			<-release // only the FIRST flush stalls
		})
	})
	defer func() { releaseOnce.Do(func() { close(release) }); SetBarPersister(nil) }()
	persistDropped.Store(0)
	persistDroppedCloses.Store(0)

	closed := []Bar{{T: time.Now().Add(-2 * time.Minute).UnixMilli(), O: 1, H: 1, L: 1, C: 1}}
	var wg sync.WaitGroup
	for i := 0; i < barPersistQueueCap+20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fanOutBarPersist(nil, false, "MNQ", "1m", closed)
		}()
	}
	// The worker picks the first batch and stalls on it → the queue fills and
	// the overflow enters the bounded-blocking retry.
	<-stalled
	time.Sleep(500 * time.Millisecond)
	releaseOnce.Do(func() { close(release) }) // the stall resolves well inside the retry window
	wg.Wait()
	time.Sleep(500 * time.Millisecond) // let the worker drain

	if got := persistDroppedCloses.Load(); got != 0 {
		t.Fatalf("closes_dropped = %d, want 0 (a writer stall < the retry window must never drop closes)", got)
	}
	if got := persistDropped.Load(); got != 0 {
		t.Fatalf("queue_drops = %d, want 0 (close-carrying batches must wait, not drop)", got)
	}
}

func TestFanOutClosesLastResortIsHonest(t *testing.T) {
	// A stall LONGER than the retry window must eventually drop with the LOUD
	// ERROR path — but the counter says exactly what happened (never silent).
	block := make(chan struct{})
	SetBarPersister(func(historical bool, symbol, tf string, bars []Bar) { <-block })
	defer func() { close(block); SetBarPersister(nil) }()
	persistDropped.Store(0)
	persistDroppedCloses.Store(0)

	closed := []Bar{{T: time.Now().Add(-2 * time.Minute).UnixMilli(), O: 1, H: 1, L: 1, C: 1}}
	// Fill the queue past cap (the worker eats up to 256 into its batch, then
	// stalls) so the close-carrying frame genuinely meets a full queue.
	for i := 0; i < barPersistQueueCap+512; i++ {
		fanOutBarPersist(nil, false, "MNQ", "1m", []Bar{{T: time.Now().UnixMilli()}}) // forming → drop path is bypassed
	}
	// One close-carrying frame now: the queue is full → it waits 2s×3 then drops.
	done := make(chan struct{})
	go func() { fanOutBarPersist(nil, false, "MNQ", "1m", closed); close(done) }()
	select {
	case <-done:
		if got := persistDroppedCloses.Load(); got != 1 {
			t.Fatalf("closes_dropped = %d, want exactly 1 after the last resort", got)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("close-carrying frame did not resolve within the retry window + margin")
	}
}
