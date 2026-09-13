package ninjatrader

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOrderedExecutionOwnerReplacementAndIsolation(t *testing.T) {
	s := NewTCPServer(nil)
	var got []string
	var mu sync.Mutex
	handler := func(name string) OrderedExecutionHandlers {
		return OrderedExecutionHandlers{Order: func(p OrderUpdatePayload) {
			// Callback can query router/cache state: dispatch holds neither lock.
			s.PositionsFor(p.Account)
			s.SubscribeOrderUpdatesFor("NQ", "Sim-other")
			if !p.OrderedArrival || p.OrderedHandled {
				t.Error("invalid sink flags")
			}
			mu.Lock()
			got = append(got, name)
			mu.Unlock()
		}}
	}
	releaseOld := s.RegisterOrderedExecutionsFor("MNQ", "Sim101", handler("old"))
	p := OrderUpdatePayload{Symbol: "MNQ", Account: "Sim101"}
	s.dispatchOrderedOrder(&p)
	releaseNew := s.RegisterOrderedExecutionsFor("MNQ", "Sim101", handler("new"))
	releaseOld() // obsolete disposal must not erase replacement
	p = OrderUpdatePayload{Symbol: "MNQ", Account: "Sim101"}
	s.dispatchOrderedOrder(&p)
	alien := OrderUpdatePayload{Symbol: "MNQ", Account: "Sim102"}
	s.dispatchOrderedOrder(&alien)
	if alien.OrderedHandled {
		t.Fatal("foreign account routed")
	}
	releaseNew()
	p = OrderUpdatePayload{Symbol: "MNQ", Account: "Sim101"}
	s.dispatchOrderedOrder(&p)
	if p.OrderedHandled || strings.Join(got, ",") != "old,new" {
		t.Fatalf("ownership: %v %+v", got, p)
	}
}
func TestOrderedExecutionFlagsCannotArriveOnWire(t *testing.T) {
	var p OrderUpdatePayload
	if err := json.Unmarshal([]byte(`{"OrderedArrival":true,"OrderedHandled":true}`), &p); err != nil {
		t.Fatal(err)
	}
	if p.OrderedArrival || p.OrderedHandled {
		t.Fatal("wire injected internal authority")
	}
}

func TestOrderedCallbackCanSendSignalOnActualConnection(t *testing.T) {
	s := NewTCPServer(nil)
	s.SetAddrForTest("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	callback := make(chan error, 1)
	s.RegisterOrderedExecutionsFor("MNQ", "Sim101", OrderedExecutionHandlers{Order: func(OrderUpdatePayload) {
		callback <- s.SendSignal(SignalPayload{SignalID: "synthetic-reentrant", Symbol: "MNQ", Account: "Sim101", Quantity: 1, Side: "long", Timestamp: time.Now().UTC().Format(time.RFC3339Nano)})
	}})
	conn, err := net.Dial("tcp", s.ListenAddrForTest().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err := WriteFrame(conn, FrameOrderUpdate, OrderUpdatePayload{Symbol: "MNQ", Account: "Sim101"}); err != nil {
		t.Fatal(err)
	}
	for {
		frame, err := ReadFrame(conn)
		if err != nil {
			t.Fatal(err)
		}
		if frame.Type == FrameSignal {
			break
		}
	}
	select {
	case err := <-callback:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("callback deadlocked")
	}
}
