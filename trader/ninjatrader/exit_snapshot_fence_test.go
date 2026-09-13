package ninjatrader

import (
	nt "nofx/provider/ninjatrader"
	"nofx/store"
	"testing"
	"time"
)

func TestExitRequiresFreshPostReceiptAccountSnapshotAcrossRestart(t *testing.T) {
	tr, st, _ := partialCloseFixture(t)
	srv := nt.NewTCPServer(nil)
	tr.server = srv
	srv.SeedPositionsForTest("Sim101", []nt.OpenPosition{}) // pre-exit empty is not evidence
	p := closeFixtureFrame()
	tr.recordClose("owned", "ex", "ninjatrader", st, store.NewPositionBuilder(st.Position()), p)
	if ps, err := tr.GetPositions(); err == nil {
		t.Fatalf("pre-exit empty accepted: %+v", ps)
	}
	restarted := &TCPTrader{symbol: "MNQ", boundAccount: "Sim101", server: srv}
	restarted.loadExitSnapshotFence(st)
	if restarted.positionsAfterMs == 0 {
		t.Fatal("local receipt fence not durable")
	}
	if _, err := restarted.GetPositions(); err == nil {
		t.Fatal("restart lost exit fence")
	}
	srv.SeedPositionsAtForTest("SimOther", []nt.OpenPosition{}, time.Now().Add(time.Second))
	if _, err := tr.GetPositions(); err == nil {
		t.Fatal("foreign snapshot cleared fence")
	}
	time.Sleep(3 * time.Millisecond)
	srv.SeedPositionsForTest("Sim101", []nt.OpenPosition{{Symbol: "MNQ", Side: "LONG", Quantity: 2, AvgPrice: 100}})
	ps, err := tr.GetPositions()
	if err != nil || len(ps) != 1 || ps[0]["quantity"] != float64(2) {
		t.Fatalf("fresh residual missing: %+v %v", ps, err)
	}
	p.SignalID = "exit-final"
	p.ExitOrderID = "broker-exit-final"
	p.Quantity = 2
	tr.recordClose("owned", "ex", "ninjatrader", st, store.NewPositionBuilder(st.Position()), p)
	if tr.CloseConfirmedSince("MNQ", "LONG", 1) {
		t.Fatal("owned row terminal was labeled account-flat")
	}
	if _, err := tr.GetPositions(); err == nil {
		t.Fatal("pre-final snapshot accepted after final execution")
	}
	time.Sleep(3 * time.Millisecond)
	srv.SeedPositionsForTest("Sim101", []nt.OpenPosition{})
	if ps, err := tr.GetPositions(); err != nil || len(ps) != 0 {
		t.Fatalf("explicit later flat unavailable: %+v %v", ps, err)
	}
	tr.positionsAfterMs = time.Now().Add(-2 * time.Minute).UnixMilli()
	srv.SeedPositionsAtForTest("Sim101", []nt.OpenPosition{}, time.Now().Add(-61*time.Second))
	if _, err := tr.GetPositions(); err == nil {
		t.Fatal("stale post-receipt empty accepted")
	}
}
