package ninjatrader

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestPositionsWireDistinguishesAbsentAndEmptyAndRecordsReceipt(t *testing.T) {
	srv := NewTCPServer(nil)
	srv.SetAddrForTest("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := srv.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()
	conn, err := net.Dial("tcp", srv.ListenAddrForTest().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for _, p := range []map[string]any{{"account": "absent"}, {"account": "null", "positions": nil}, {"account": "Sim101", "positions": []OpenPosition{}}} {
		if err := WriteFrame(conn, FramePositions, p); err != nil {
			t.Fatal(err)
		}
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		ps, received, ok := srv.PositionsForReceived("Sim101")
		if ok {
			if ps == nil || len(ps) != 0 || received.IsZero() || received.After(time.Now()) {
				t.Fatalf("invalid explicit empty receipt: %+v %v", ps, received)
			}
			for _, acct := range []string{"absent", "null"} {
				if _, _, ok := srv.PositionsForReceived(acct); ok {
					t.Fatal("absent/null book accepted")
				}
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("valid snapshot did not arrive")
}
func TestPositionsReceiptGetterReturnsPairedDefensiveCopy(t *testing.T) {
	srv := NewTCPServer(nil)
	at := time.Now()
	srv.SeedPositionsAtForTest("Sim101", []OpenPosition{{Symbol: "MNQ", Quantity: 2}}, at)
	ps, received, ok := srv.PositionsForReceived("Sim101")
	if !ok || !received.Equal(at) || ps[0].Quantity != 2 {
		t.Fatal("payload/receipt mismatch")
	}
	ps[0].Quantity = 99
	again, _ := srv.PositionsFor("Sim101")
	if again[0].Quantity != 2 {
		t.Fatal("caller mutated cache")
	}
}
