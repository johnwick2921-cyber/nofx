// W117 F2 (9b379c8c) — the owning durable consumers run on the TCP read
// goroutine in RECEIVE order, before advisory fanout. This pin drives the real
// read loop: interleaved order/fill/close frames must reach the registered
// handler in wire order. RED = remove the three dispatchOrdered* calls in
// tcp_server.go → the handler never fires and the pin times out.

package ninjatrader

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestOrderedExecutionConsumesInReceiveOrder(t *testing.T) {
	srv := NewTCPServer(nil)
	addr := freeEphemeralAddr(t)
	srv.SetAddrForTest(addr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := srv.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer srv.Stop()

	seen := make(chan string, 8)
	unregister := srv.RegisterOrderedExecutionsFor("MNQ", "Sim101", OrderedExecutionHandlers{
		Order: func(p OrderUpdatePayload) {
			if !p.OrderedArrival {
				seen <- "order:no-arrival-flag"
				return
			}
			seen <- "order:" + p.State
		},
		Fill: func(p FillPayload) {
			seen <- fmt.Sprintf("fill:%s:%d", p.Side, p.Quantity)
		},
		Close: func(p PositionClosePayload) {
			seen <- "close:" + p.ExitReason
		},
	})
	defer unregister()

	client := NewMockTCPClient(addr, 50*time.Millisecond)
	if err := client.Start(ctx); err != nil {
		t.Fatalf("client Start: %v", err)
	}
	defer client.Stop()
	waitForConnected(t, srv, 2*time.Second)

	// Interleaved wire order: order → fill → close → order.
	for _, f := range []struct {
		ft  FrameType
		any any
	}{
		{FrameOrderUpdate, OrderUpdatePayload{SignalID: "s1", OrderName: "s1", State: "partfilled", Quantity: 1, Symbol: "MNQ", Account: "Sim101"}},
		{FrameFill, FillPayload{SignalID: "s1", Symbol: "MNQ", Account: "Sim101", Side: "long", Quantity: 1, FillPrice: 100}},
		{FramePositionClose, PositionClosePayload{SignalID: "s1", Symbol: "MNQ", Account: "Sim101", ExitReason: "tp", Quantity: 1, PositionSide: "long"}},
		{FrameOrderUpdate, OrderUpdatePayload{SignalID: "s1", OrderName: "s1", State: "cancelled", Quantity: 0, Symbol: "MNQ", Account: "Sim101"}},
	} {
		if err := writeFromMock(client, f.ft, f.any); err != nil {
			t.Fatalf("write %s: %v", f.ft, err)
		}
	}

	want := []string{"order:partfilled", "fill:long:1", "close:tp", "order:cancelled"}
	deadline := time.After(3 * time.Second)
	for i, w := range want {
		select {
		case got := <-seen:
			if got != w {
				t.Fatalf("receive order broken: step %d want %q got %q", i, w, got)
			}
		case <-deadline:
			t.Fatalf("step %d (%q) never dispatched — the ordered consumers are not running in receive order", i, w)
		}
	}
}
