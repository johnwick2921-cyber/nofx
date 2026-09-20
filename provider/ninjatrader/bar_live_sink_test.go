package ninjatrader

import (
	"context"
	"testing"
	"time"
)

// LIVE BAR SINK (W-PICTURE-HTF, 2026-09-20) — pins for the deterministic
// evaluator fan-out: live frames are delivered, historical replays are NOT
// (a replay receipt can never mint a real opportunity).

func TestFanOutLiveBarsDelivers(t *testing.T) {
	got := make(chan liveSinkMsg, 2)
	SetLiveBarSink(func(symbol, tf string, bars []Bar, receivedAt time.Time) {
		got <- liveSinkMsg{symbol: symbol, tf: tf, bars: bars, receivedAt: receivedAt}
	})
	t.Cleanup(func() { SetLiveBarSink(nil) })

	fanOutLiveBars("MNQ", "5m", []Bar{{T: 1, C: 101.5}})
	select {
	case m := <-got:
		if m.symbol != "MNQ" || m.tf != "5m" || len(m.bars) != 1 || m.bars[0].C != 101.5 {
			t.Fatalf("delivered frame mismatch: %+v", m)
		}
		if m.receivedAt.IsZero() {
			t.Fatalf("receivedAt must be stamped at the drain instant")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("live frame never delivered")
	}
}

func TestFanOutLiveBarsEmptyNoOp(t *testing.T) {
	SetLiveBarSink(func(symbol, tf string, bars []Bar, receivedAt time.Time) {
		t.Fatalf("empty batch must not fan out")
	})
	t.Cleanup(func() { SetLiveBarSink(nil) })
	fanOutLiveBars("MNQ", "5m", nil)
	time.Sleep(50 * time.Millisecond)
}

func TestDrainBarIngestHistoricalNeverFansOut(t *testing.T) {
	live := make(chan string, 1)
	SetLiveBarSink(func(symbol, tf string, bars []Bar, receivedAt time.Time) {
		live <- tf
	})
	t.Cleanup(func() { SetLiveBarSink(nil) })

	s := NewTCPServer(nil)
	s.barIngestCh = make(chan barIngestMsg, 4)
	s.wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() { cancel(); s.wg.Wait() }()
	go s.drainBarIngest(ctx)

	s.barIngestCh <- barIngestMsg{historical: true, symbol: "MNQ", timeframe: "5m", bars: []Bar{{T: 1}}}
	select {
	case tf := <-live:
		t.Fatalf("historical replay fanned out (%s) — replay receipts must never mint opportunities", tf)
	case <-time.After(200 * time.Millisecond):
	}
	s.barIngestCh <- barIngestMsg{historical: false, symbol: "MNQ", timeframe: "5m", bars: []Bar{{T: 2}}}
	select {
	case tf := <-live:
		if tf != "5m" {
			t.Fatalf("live frame delivered with tf %q", tf)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("live frame never delivered")
	}
}
