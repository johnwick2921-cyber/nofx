package ninjatrader

import (
	"math"
	nt "nofx/provider/ninjatrader"
	"nofx/store"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func partialCloseFixture(t *testing.T) (*TCPTrader, *store.Store, *store.TraderPosition) {
	t.Helper()
	st, err := store.New(filepath.Join(t.TempDir(), "partial.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	row := &store.TraderPosition{TraderID: "owned", Account: "Sim101", Symbol: "MNQ", Side: "LONG", Quantity: 3, EntryQuantity: 3, EntryPrice: 100, EntryTime: 1, Status: "OPEN", EntryOrderID: "entry-owned", PlanID: "plan-owned", PlanVersion: 7, Fee: 1.25}
	if err := st.Position().Create(row); err != nil {
		t.Fatal(err)
	}
	tr := &TCPTrader{symbol: "MNQ", boundAccount: "Sim101", hasFill: true, lastFill: nt.FillPayload{SignalID: "entry-owned", Symbol: "MNQ", Account: "Sim101", Side: "long", Quantity: 3, FillPrice: 100, Status: "filled"}}
	return tr, st, row
}
func closeFixtureFrame() nt.PositionClosePayload {
	return nt.PositionClosePayload{ExitOrderID: "broker-exit-one", SignalID: "exit-one", Symbol: "MNQ", Account: "Sim101", PositionSide: "long", Quantity: 1, ExitPrice: 110, ExitReason: "limit", ExitTime: time.Now().UTC().Format(time.RFC3339Nano)}
}
func TestPartialCloseActualCallbackPreservesResidualAndLineage(t *testing.T) {
	tr, st, row := partialCloseFixture(t)
	pb := store.NewPositionBuilder(st.Position())
	p := closeFixtureFrame()
	previous := OnPositionClosed
	closedCalls := 0
	OnPositionClosed = func(string, int64) { closedCalls++ }
	t.Cleanup(func() { OnPositionClosed = previous })
	foreign := &store.TraderPosition{TraderID: "owned", Account: "SimOther", Symbol: "MNQ", Side: "LONG", Quantity: 4, EntryQuantity: 4, EntryPrice: 90, EntryTime: 2, Status: "OPEN"}
	if err := st.Position().Create(foreign); err != nil {
		t.Fatal(err)
	}
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	var got store.TraderPosition
	st.GormDB().First(&got, row.ID)
	if got.Status != "OPEN" || got.Quantity != 2 || got.RealizedPnL != 20 || got.Fee != 1.25 || got.EntryOrderID != "entry-owned" || got.PlanID != "plan-owned" || got.PlanVersion != 7 {
		t.Fatalf("partial corrupted owned row: %+v", got)
	}
	if closedCalls != 0 || tr.CloseConfirmedSince("MNQ", "LONG", 1) || !tr.hasFill || tr.lastFill.Quantity != 2 {
		t.Fatal("partial claimed flat or lost residual cache")
	}
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	st.GormDB().First(&got, row.ID)
	if got.Quantity != 2 || got.RealizedPnL != 20 {
		t.Fatal("duplicate reduced twice")
	}
	p.SignalID = "exit-two"
	p.ExitOrderID = "broker-exit-two"
	p.Quantity = 2
	p.ExitPrice = 120
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	st.GormDB().First(&got, row.ID)
	if got.Status != "CLOSED" || got.Quantity != 3 || got.RealizedPnL != 100 || got.PnlCorrected == nil || *got.PnlCorrected != 100 || math.Abs(got.ExitPrice-350.0/3) > 1e-10 || got.Fee != 1.25 {
		t.Fatalf("final totals wrong: %+v", got)
	}
	if closedCalls != 1 || tr.CloseConfirmedSince("MNQ", "LONG", 1) || tr.hasFill {
		t.Fatal("terminal evidence/hook missing")
	}
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	if closedCalls != 1 {
		t.Fatal("duplicate terminal fired hook")
	}
	var fills []store.TraderFill
	st.GormDB().Order("id").Find(&fills)
	if len(fills) != 2 || fills[0].Quantity != 1 || fills[1].Quantity != 2 || fills[0].RealizedPnL != 20 || fills[1].RealizedPnL != 80 {
		t.Fatalf("actual execution fills wrong: %+v", fills)
	}
	got = store.TraderPosition{}
	st.GormDB().First(&got, foreign.ID)
	if got.Status != "OPEN" || got.Quantity != 4 {
		t.Fatal("same trader different account changed")
	}
}
func TestExitReceiptRefusesUnknownExcessAndConflictingEvidence(t *testing.T) {
	for _, kind := range []string{"zero", "excess", "missing identity", "missing account", "wrong side", "nonfinite", "lineage"} {
		t.Run(kind, func(t *testing.T) {
			tr, st, row := partialCloseFixture(t)
			p := closeFixtureFrame()
			switch kind {
			case "zero":
				p.Quantity = 0
			case "excess":
				p.Quantity = 4
			case "missing identity":
				p.SignalID = ""
			case "missing account":
				p.Account = ""
			case "wrong side":
				p.PositionSide = "garbage"
			case "nonfinite":
				p.ExitPrice = math.NaN()
			case "lineage":
				p.ExitReason = "sl"
				p.SignalID = "different-entry"
			}
			tr.recordClose("owned", "ex", "ninjatrader", st, store.NewPositionBuilder(st.Position()), p)
			var got store.TraderPosition
			st.GormDB().First(&got, row.ID)
			if got.Quantity != 3 || got.Status != "OPEN" || !tr.hasFill || tr.CloseConfirmedSince("MNQ", "LONG", 1) {
				t.Fatal("invalid evidence changed position")
			}
		})
	}
	tr, st, row := partialCloseFixture(t)
	p := closeFixtureFrame()
	pb := store.NewPositionBuilder(st.Position())
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	p.Quantity = 2
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	var got store.TraderPosition
	st.GormDB().First(&got, row.ID)
	if got.Quantity != 2 {
		t.Fatal("conflicting receipt altered residual")
	}
}
func TestPartialExitReceiptAndFillRollbackTogether(t *testing.T) {
	tr, st, row := partialCloseFixture(t)
	if err := st.GormDB().Exec("CREATE TRIGGER fail_exit_fill BEFORE INSERT ON trader_fills BEGIN SELECT RAISE(ABORT, 'synthetic fill failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	p := closeFixtureFrame()
	pb := store.NewPositionBuilder(st.Position())
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	var got store.TraderPosition
	st.GormDB().First(&got, row.ID)
	var count int64
	st.GormDB().Model(&store.NT8ExitReceipt{}).Count(&count)
	if got.Quantity != 3 || got.RealizedPnL != 0 || count != 0 || tr.lastFill.Quantity != 3 {
		t.Fatal("failed fill did not roll back position and receipt")
	}
	st.GormDB().Exec("DROP TRIGGER fail_exit_fill")
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	st.GormDB().First(&got, row.ID)
	if got.Quantity != 2 {
		t.Fatal("same receipt could not retry after rollback")
	}
}
func TestConcurrentDuplicateExitCallbackReducesOnce(t *testing.T) {
	tr, st, row := partialCloseFixture(t)
	p := closeFixtureFrame()
	pb := store.NewPositionBuilder(st.Position())
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); tr.recordClose("owned", "ex", "ninjatrader", st, pb, p) }()
	}
	wg.Wait()
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	var got store.TraderPosition
	st.GormDB().First(&got, row.ID)
	var count int64
	st.GormDB().Model(&store.TraderFill{}).Count(&count)
	if got.Quantity != 2 || got.RealizedPnL != 20 || count != 1 {
		t.Fatalf("duplicate race: qty=%g pnl=%g fills=%d", got.Quantity, got.RealizedPnL, count)
	}
}

func TestBrokerExitIdentityDistinguishesManualOrdersAndIgnoresEchoSeq(t *testing.T) {
	tr, st, row := partialCloseFixture(t)
	pb := store.NewPositionBuilder(st.Position())
	p := closeFixtureFrame()
	p.SignalID = "Close"
	p.ExitReason = "manual"
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	p.Seq++
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	p.ExitOrderID = "broker-second-manual"
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	var got store.TraderPosition
	st.GormDB().First(&got, row.ID)
	if got.Quantity != 1 || got.RealizedPnL != 40 {
		t.Fatalf("separate manual orders / retransmission wrong: %+v", got)
	}
	p.ExitOrderID = ""
	p.Seq++
	tr.recordClose("owned", "ex", "ninjatrader", st, pb, p)
	st.GormDB().First(&got, row.ID)
	if got.Quantity != 1 {
		t.Fatalf("generic legacy name consumed residual: %+v", got)
	}
}

func TestExactBracketReceiptSurvivesDelayedEntryMaterialization(t *testing.T) {
	tr, st, row := partialCloseFixture(t)
	p := closeFixtureFrame()
	p.SignalID = row.EntryOrderID
	p.ExitReason = "tp"
	if err := st.GormDB().Model(&store.TraderPosition{}).Where("id = ?", row.ID).Update("entry_time", time.Now().Add(time.Minute).UnixMilli()).Error; err != nil {
		t.Fatal(err)
	}
	tr.recordClose("owned", "ex", "ninjatrader", st, store.NewPositionBuilder(st.Position()), p)
	var got store.TraderPosition
	st.GormDB().First(&got, row.ID)
	if got.Quantity != 2 || got.RealizedPnL != 20 {
		t.Fatalf("exact delayed entry receipt not applied: %+v", got)
	}
}
