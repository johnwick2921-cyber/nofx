// W117 F2 trader half — the ordered consumer applies execution evidence EXACTLY
// ONCE: the read goroutine runs handleFill before fanout, and the channel
// consumer's copy is a no-op via OrderedHandled. RED = remove the
// `if fill.OrderedHandled { return }` guard in handleFill → the same fill is
// applied twice (ordered pass + channel pass) and the ring grows to 2.

package ninjatrader

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

func TestOrderedFillAppliedExactlyOnce(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "ordered.db"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	srv := ntwire.NewTCPServer(nil)
	srv.SetAddrForTest("127.0.0.1:0")
	srv.SetAccountsList([]ntwire.AccountInfo{{Name: "Sim101", IsSim: true}}, "Sim101")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := srv.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer srv.Stop()

	tr := NewTCPTrader(srv, "MNQ", "Sim101")
	orders := make(chan ntwire.OrderUpdatePayload, 4)
	tr.InstallOrderedExecutions("t1", "x", "ninjatrader", st, func(u ntwire.OrderUpdatePayload) {
		orders <- u
	})

	conn, err := net.Dial("tcp", srv.ListenAddrForTest().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// One filled fill — the durable path must apply it exactly once (seq==0 is
	// legacy-tolerated by the echo verify).
	if err := ntwire.WriteFrame(conn, ntwire.FrameFill, ntwire.FillPayload{
		SignalID: "s1", Symbol: "MNQ", Account: "Sim101",
		Side: "long", Quantity: 1, FillPrice: 100, Status: "filled",
	}); err != nil {
		t.Fatalf("write fill: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		tr.mu.Lock()
		hasFill, n := tr.hasFill, len(tr.recentFills)
		tr.mu.Unlock()
		if hasFill && n > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the fill never reached the durable cache")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Give the channel consumer time to receive the fanout copy too.
	time.Sleep(100 * time.Millisecond)
	tr.mu.Lock()
	n := len(tr.recentFills)
	sig := tr.lastFill.SignalID
	tr.mu.Unlock()
	if n != 1 {
		t.Fatalf("the same fill must be applied EXACTLY once: ring=%d (ordered pass + channel pass double-applied it)", n)
	}
	if sig != "s1" {
		t.Fatalf("lastFill signal = %q, want s1", sig)
	}
}
