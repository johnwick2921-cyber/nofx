package ninjatrader

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

// ── 101 D1'(3a) [O "I want full data", 2026-09-16] — after a confirmed scale
// break, ask NT8 for the history again ─────────────────────────────────────
//
// The AddOn's N4 path already DISPOSES + RECREATES its BarsRequest on a repeat
// bars_subscribe ("a fresh BarsRequest re-reads NT8's DB over the full
// barsBack lookback"), so a Go-side re-send IS the re-request — no AddOn
// change. Three pins: one frame per break; not twice inside the window (a
// flapping feed must not storm the AddOn); nothing at all while the feed is
// down (that is exactly how the 09-16 reconnects produced bars=0).

func readSubscribeFrame(t *testing.T, cli net.Conn, wait time.Duration) (BarsSubscribePayload, bool) {
	t.Helper()
	_ = cli.SetReadDeadline(time.Now().Add(wait))
	env, err := ReadFrame(cli)
	if err != nil {
		return BarsSubscribePayload{}, false
	}
	if env.Type != FrameBarsSubscribe {
		t.Fatalf("frame type %q, want %q", env.Type, FrameBarsSubscribe)
	}
	var p BarsSubscribePayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return p, true
}

func TestConfirmedBreakRequestsAFreshReplayOnce(t *testing.T) {
	s := NewTCPServer(nil)
	srv, cli := net.Pipe()
	defer srv.Close()
	defer cli.Close()
	s.conn = srv
	s.barsSubscribe.Timeframes = []string{"1m", "5m", "1h"}
	s.barsSubscribe.BarsBack = 2000
	s.markFeedConnectedForTest(true)

	now := time.Date(2026, 9, 16, 9, 22, 16, 0, time.UTC)
	done := make(chan BarsSubscribePayload, 2)
	go func() {
		for {
			p, ok := readSubscribeFrame(t, cli, 500*time.Millisecond)
			if !ok {
				close(done)
				return
			}
			done <- p
		}
	}()

	if err := s.RequestHistoryReplayAt("MNQ", now); err != nil {
		t.Fatalf("first re-request: %v", err)
	}
	// a second break on another tf of the same symbol, seconds later
	if err := s.RequestHistoryReplayAt("MNQ", now.Add(5*time.Second)); err == nil {
		t.Errorf("a second re-request inside the window must be refused (rate limit), got nil")
	}
	var got []BarsSubscribePayload
	for p := range done {
		got = append(got, p)
	}
	if len(got) != 1 {
		t.Fatalf("want exactly ONE bars_subscribe frame, got %d", len(got))
	}
	if got[0].Symbol != "MNQ" || got[0].BarsBack != 2000 || len(got[0].Timeframes) != 3 {
		t.Errorf("the re-request must carry the subscribe's own symbol/tfs/bars_back: %+v", got[0])
	}
}

func TestReplayIsNotRequestedWhileTheFeedIsDown(t *testing.T) {
	s := NewTCPServer(nil)
	srv, cli := net.Pipe()
	defer srv.Close()
	defer cli.Close()
	s.conn = srv
	s.barsSubscribe.Timeframes = []string{"5m"}
	s.barsSubscribe.BarsBack = 2000
	s.markFeedConnectedForTest(false)

	if err := s.RequestHistoryReplayAt("MNQ", time.Now()); err == nil {
		t.Fatalf("feed down: the re-request must be refused with a reason, got nil")
	}
	if _, ok := readSubscribeFrame(t, cli, 100*time.Millisecond); ok {
		t.Fatalf("feed down: no frame may be sent — a BarsRequest that runs while the feed is down returns bars=0 (09-16)")
	}
}

// Past the window the same symbol may ask again.
func TestReplayRequestReArmsAfterTheWindow(t *testing.T) {
	s := NewTCPServer(nil)
	srv, cli := net.Pipe()
	defer srv.Close()
	defer cli.Close()
	s.conn = srv
	s.barsSubscribe.Timeframes = []string{"5m"}
	s.barsSubscribe.BarsBack = 2000
	s.markFeedConnectedForTest(true)
	go func() {
		for {
			if _, ok := readSubscribeFrame(t, cli, 2*time.Second); !ok {
				return
			}
		}
	}()
	t0 := time.Date(2026, 9, 16, 9, 22, 16, 0, time.UTC)
	if err := s.RequestHistoryReplayAt("MNQ", t0); err != nil {
		t.Fatal(err)
	}
	if err := s.RequestHistoryReplayAt("MNQ", t0.Add(historyReplayMinInterval+time.Second)); err != nil {
		t.Fatalf("past the window the symbol may ask again: %v", err)
	}
}
