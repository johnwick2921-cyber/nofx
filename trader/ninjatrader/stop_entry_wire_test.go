package ninjatrader

import (
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// ENTRY-MECHANICS E7 (2026-08-30) — stop-entry orders. Far-side frame law:
// the stop_entry frame is PROVEN by receiving it on the loopback TCP server
// (the same harness the move_stop wire proof uses) — order_type=stop_entry,
// stop_price = the rounded trigger, identity stamp intact. The C# AddOn's
// parse of the new fields is additive (old Go never sends it); the real NT8
// far-side proof is the D-rule cutover gate.

func stopEntryServer(t *testing.T) (*ntwire.TCPServer, *store.Store, net.Conn, chan ntwire.SignalPayload) {
	st, err := store.New(filepath.Join(t.TempDir(), "se.db"))
	if err != nil {
		t.Fatal(err)
	}
	s := ntwire.NewTCPServer(nil)
	s.SetAddrForTest("127.0.0.1:0")
	// The tradeable guard needs Sim101 registered as SIM.
	s.SetAccountsList([]ntwire.AccountInfo{{Name: "Sim101", IsSim: true}}, "Sim101")
	ctx, cancel := context.WithCancel(context.Background())
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Stop(); cancel(); _ = st.Close() })

	conn, err := net.Dial("tcp", s.ListenAddrForTest().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	frames := make(chan ntwire.SignalPayload, 4)
	go func() {
		for {
			env, err := ntwire.ReadFrame(conn)
			if err != nil {
				return
			}
			if env.Type != ntwire.FrameSignal {
				continue
			}
			var p ntwire.SignalPayload
			if json.Unmarshal(env.Payload, &p) == nil {
				frames <- p
			}
		}
	}()
	return s, st, conn, frames
}

// TestPlaceStopEntryFrameOnLoopback — twin long/short: the stop-entry signal
// frame arrives with order_type=stop_entry + stop_price and the identity stamp.
func TestPlaceStopEntryFrameOnLoopback(t *testing.T) {
	for _, tc := range []struct {
		side    string
		trigger float64
		sl      float64
		tp      float64
	}{
		{"long", 29670.50, 29650.00, 29720.00},
		{"short", 29400.25, 29420.75, 29350.00},
	} {
		s, st, _, frames := stopEntryServer(t)

		tr := NewTCPTrader(s, "MNQ", "Sim101")
		tr.mu.Lock()
		tr.st = st
		tr.mu.Unlock()

		var sid string
		var err error
		for i := 0; i < 50; i++ {
			sid, err = tr.PlaceStopEntry("MNQ", tc.side, 1, tc.trigger, tc.sl, tc.tp)
			if err == nil || !strings.Contains(err.Error(), "no NT client connected") {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if err != nil {
			t.Fatalf("%s: place stop-entry failed: %v", tc.side, err)
		}
		select {
		case p := <-frames:
			if p.OrderType != "stop_entry" {
				t.Fatalf("%s: order_type=%q, want stop_entry", tc.side, p.OrderType)
			}
			if p.StopPrice != tc.trigger {
				t.Fatalf("%s: stop_price=%.2f, want %.2f", tc.side, p.StopPrice, tc.trigger)
			}
			if p.SignalID != sid {
				t.Fatalf("%s: signal_id mismatch", tc.side)
			}
			if p.Account != "Sim101" || p.TraderID != tr.traderID {
				t.Fatalf("%s: identity stamp missing: acct=%q trader=%q", tc.side, p.Account, p.TraderID)
			}
			if p.StopLoss != tc.sl || p.TakeProfit != tc.tp {
				t.Fatalf("%s: bracket wrong: SL %.2f TP %.2f", tc.side, p.StopLoss, p.TakeProfit)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("%s: no stop_entry frame arrived on the loopback", tc.side)
		}
	}
}

// TestStopEntryFrameIsAdditiveJSON — a pre-E7 Go binary never emits the new
// fields; the limit frame must remain byte-shape-identical apart from its own
// fields (backward-compat proof).
func TestStopEntryFrameIsAdditiveJSON(t *testing.T) {
	lim := ntwire.SignalPayload{Symbol: "MNQ", Side: "long", Quantity: 1,
		Entry: 100.25, StopLoss: 99.00, TakeProfit: 102.00, SignalID: "s1",
		OrderType: "limit", LimitPrice: 100.25}
	b1, _ := json.Marshal(lim)
	var m1 map[string]any
	_ = json.Unmarshal(b1, &m1)
	if _, has := m1["stop_price"]; has {
		t.Fatal("a limit frame must NOT carry stop_price")
	}
	if _, has := m1["limit_price"]; !has {
		t.Fatal("limit frame must carry limit_price")
	}
	se := lim
	se.OrderType = "stop_entry"
	se.StopPrice = 101.00
	se.LimitPrice = 0
	b2, _ := json.Marshal(se)
	var m2 map[string]any
	_ = json.Unmarshal(b2, &m2)
	if m2["order_type"] != "stop_entry" || m2["stop_price"] != 101.00 {
		t.Fatalf("stop_entry frame malformed: %s", b2)
	}
}
