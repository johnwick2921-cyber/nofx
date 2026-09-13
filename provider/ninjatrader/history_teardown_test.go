package ninjatrader

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

// Exercises actual frame dispatch while the importer tears down/replaces its
// subscription. All traffic stays on a random-port loopback fixture.
func TestHistoryFrameDeliveryConcurrentTeardown(t *testing.T) {
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
	srv.SubscribeBarsHistoryFor("race")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20000; i++ {
			srv.UnsubscribeBarsHistoryFor("race")
			srv.SubscribeBarsHistoryFor("race")
		}
	}()
	for i := 0; i < 1000; i++ {
		if err := WriteFrame(conn, FrameBarsHistoryData, BarsHistoryDataPayload{RequestID: "race", Contract: "MNQ 09-26"}); err != nil {
			t.Error(err)
			break
		}
		if err := WriteFrame(conn, FrameBarsHistoryError, BarsHistoryErrorPayload{RequestID: "race", Contract: "MNQ 09-26", Reason: "fixture"}); err != nil {
			t.Error(err)
			break
		}
	}
	wg.Wait()
	data, _ := srv.SubscribeBarsHistoryFor("sentinel")
	if err := WriteFrame(conn, FrameBarsHistoryData, BarsHistoryDataPayload{RequestID: "sentinel", Contract: "MNQ 09-26"}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-data:
		if got.RequestID != "sentinel" {
			t.Fatal(got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("read loop did not deliver sentinel")
	}
}
