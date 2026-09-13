package trader

import (
	"nofx/kernel"
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
	nttrader "nofx/trader/ninjatrader"
	"strings"
	"testing"
)

// Keep the real adapter GetPositions, while recording any unintended send.
// The wrapper bypasses the independent connection-status gate so this test
// reaches the production account-state gate specifically.
type firstSnapshotAdapter struct {
	*nttrader.TCPTrader
	opens int
}

func (a *firstSnapshotAdapter) OpenLong(string, float64, int) (map[string]interface{}, error) {
	a.opens++
	return nil, nil
}

func TestRealAdapterUnknownFirstSnapshotIsNotFlatAtCaller(t *testing.T) {
	srv := ntwire.NewTCPServer(nil)
	adapter := &firstSnapshotAdapter{TCPTrader: nttrader.NewTCPTrader(srv, "MNQ", "Sim101")}
	at := &AutoTrader{id: "first-snapshot", exchange: "ninjatrader", trader: adapter}
	side, err := at.ntHeldPosition("MNQ")
	if err == nil {
		t.Fatalf("real adapter absence became flat: held=%q error=%v", side, err)
	}
}
func TestProductionOpenRefusesRealAdapterBeforeFirstSnapshot(t *testing.T) {
	srv := ntwire.NewTCPServer(nil)
	adapter := &firstSnapshotAdapter{TCPTrader: nttrader.NewTCPTrader(srv, "MNQ", "Sim101")}
	at := &AutoTrader{id: "first-snapshot", exchange: "ninjatrader", trader: adapter}
	err := at.executeOpenLongWithRecord(&kernel.Decision{Symbol: "MNQ"}, &store.DecisionAction{})
	if err == nil || !strings.Contains(err.Error(), "account positions unknown") || adapter.opens != 0 {
		t.Fatalf("unknown did not refuse at actual open callsite: %v opens=%d", err, adapter.opens)
	}
	srv.SeedPositionsForTest("Sim101", []ntwire.OpenPosition{})
	if held, err := at.ntHeldPosition("MNQ"); err != nil || held != "" {
		t.Fatalf("explicit fresh empty unavailable: %q %v", held, err)
	}
}
