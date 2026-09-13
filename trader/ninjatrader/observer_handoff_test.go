package ninjatrader

import (
	nt "nofx/provider/ninjatrader"
	"testing"
	"time"
)

func TestActualObserverStartReplacementRetiresOldReconcileOnly(t *testing.T) {
	old, st, _ := partialCloseFixture(t)
	srv := nt.NewTCPServer(nil)
	old.server = srv
	old.StartCloseSync("owned", "ex", "ninjatrader", st)
	old.StartPositionReconcile("owned", "ex", "ninjatrader", st)
	old.mu.Lock()
	stopped := old.reconcileStopped
	old.mu.Unlock()
	foreign := &TCPTrader{server: srv, symbol: "MNQ", boundAccount: "SimOther"}
	foreign.StartCloseSync("other", "ex", "ninjatrader", st)
	select {
	case <-old.observerLifetime():
		t.Fatal("other account retired observer")
	default:
	}
	// A disconnected bridge does not mean the observer was replaced.
	if err := srv.Stop(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-old.observerLifetime():
		t.Fatal("disconnect retired observer")
	default:
	}
	next := &TCPTrader{server: srv, symbol: "MNQ", boundAccount: "Sim101"}
	next.StartCloseSync("owned", "ex", "ninjatrader", st)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("old reconcile worker survived successful replacement")
	}
	// Re-entrant Start on the retired instance cannot steal the replacement.
	old.StartCloseSync("owned", "ex", "ninjatrader", st)
	select {
	case <-next.observerLifetime():
		t.Fatal("old Start stole replacement subscription")
	default:
	}
	// Actual router replacement drains all observers created by this fixture.
	srv.SubscribeClosesFor("MNQ", "Sim101")
	srv.SubscribeRejectsFor("MNQ", "Sim101")
	srv.SubscribeInstrumentInfoFor("MNQ", "Sim101")
	srv.SubscribeClosesFor("MNQ", "SimOther")
	srv.SubscribeRejectsFor("MNQ", "SimOther")
	srv.SubscribeInstrumentInfoFor("MNQ", "SimOther")
}
