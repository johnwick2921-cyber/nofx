package ninjatrader

import (
	"nofx/kernel"
	ntwire "nofx/provider/ninjatrader"
	"strings"
	"testing"
	"time"
)

// Actual adapter methods used by decision, resting arm and debug entry paths.
// No NT8 connection is opened; refusal must precede account checks/ledger writes.
func TestBootRefusalCoversEveryTCPEntryMethod(t *testing.T) {
	oldReason, oldRefused := kernel.TradingRefused()
	t.Cleanup(func() { kernel.SetTradingRefusedForTest(oldRefused, oldReason) })
	kernel.SetTradingRefusedForTest(true, "offline boot mismatch")
	server := ntwire.NewTCPServer(nil)
	server.SeedPositionsForTest("Sim101", []ntwire.OpenPosition{}) // explicit known flat fixture, not absence
	tr := NewTCPTrader(server, "MNQ", "Sim101")
	registered := 0
	register := func(string) error { registered++; return nil }
	cases := []struct {
		name string
		call func() error
	}{
		{"market-long", func() error { _, e := tr.OpenLong("MNQ", 1, 1); return e }},
		{"market-short", func() error { _, e := tr.OpenShort("MNQ", 1, 1); return e }},
		{"limit-long", func() error { _, e := tr.PlaceLimitEntry("MNQ", "long", 1, 20000, 19980, 20040, register); return e }},
		{"limit-short", func() error { _, e := tr.PlaceLimitEntry("MNQ", "short", 1, 20000, 20020, 19960, register); return e }},
		{"stop-long", func() error { _, e := tr.PlaceStopEntry("MNQ", "long", 1, 20000, 19980, 20040, register); return e }},
		{"stop-short", func() error { _, e := tr.PlaceStopEntry("MNQ", "short", 1, 20000, 20020, 19960, register); return e }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if e := tc.call(); e == nil || !strings.Contains(e.Error(), "offline boot mismatch") {
				t.Errorf("boot refusal missing: %v", e)
			}
		})
	}
	if registered != 0 || len(tr.pending) != 0 {
		t.Fatalf("refused entry changed placement state: registered=%d pending=%d", registered, len(tr.pending))
	}
	// Read/protection settings remain usable; close/cancel reach their transport
	// (disconnected here) rather than being refused by boot integrity.
	if e := tr.SetStopLoss("MNQ", "long", 1, 19980); e != nil {
		t.Fatal(e)
	}
	if _, e := tr.GetPositions(); e != nil {
		t.Fatal(e)
	}
	if _, e := tr.CloseLong("MNQ", 1); e != nil && strings.Contains(e.Error(), "offline boot mismatch") {
		t.Fatal("boot blocked close")
	}
	if e := tr.CancelOrder("test-entry"); e != nil && strings.Contains(e.Error(), "offline boot mismatch") {
		t.Fatal("boot blocked cancel")
	}
}

func TestBootRefusalOnReadyLoopbackWire(t *testing.T) {
	oldReason, oldRefused := kernel.TradingRefused()
	t.Cleanup(func() { kernel.SetTradingRefusedForTest(oldRefused, oldReason) })
	s, _, conn, frames := stopEntryServer(t)
	if err := ntwire.WriteFrame(conn, ntwire.FrameHeartbeat, ntwire.HeartbeatPayload{BuildID: ntwire.MinAddonBuildStopSlot}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for !ntwire.FarSideProven(s.FarSideBuildID(), ntwire.MinAddonBuildStopSlot) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !ntwire.FarSideProven(s.FarSideBuildID(), ntwire.MinAddonBuildStopSlot) {
		t.Fatal("offline peer handshake missing")
	}
	tr := NewTCPTrader(s, "MNQ", "Sim101")
	tr.guard = nil // Isolate the boot gate; dedup has its own tests.
	kernel.SetTradingRefusedForTest(true, "offline boot mismatch")
	if _, err := tr.PlaceLimitEntry("MNQ", "long", 1, 20000, 19980, 20040); err == nil || !strings.Contains(err.Error(), "offline boot mismatch") {
		t.Errorf("ready wire bypassed boot refusal: %v", err)
	}
	select {
	case p := <-frames:
		t.Errorf("entry escaped boot refusal: %s", p.SignalID)
	case <-time.After(50 * time.Millisecond):
	}
	kernel.SetTradingRefusedForTest(false, "")
	if _, err := tr.PlaceLimitEntry("MNQ", "long", 1, 20000, 19980, 20040); err != nil {
		t.Fatal(err)
	}
	select {
	case p := <-frames:
		if p.OrderType != "limit" {
			t.Fatalf("wrong allowed frame: %+v", p)
		}
	case <-time.After(time.Second):
		t.Fatal("allowed entry never reached offline peer")
	}
}
