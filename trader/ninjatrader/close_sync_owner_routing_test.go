package ninjatrader

import (
	"path/filepath"
	"testing"

	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// TestRecordClose_OwnerRouting locks the fix: a position_close frame received by a
// trader that does NOT own the open row is still matched to the OWNING trader (by
// account+symbol+side across ALL traders) and the P&L is PERSISTED — instead of being
// silently dropped (the old "No matching open position, skipping" → reconcile pnl=0).
func TestRecordClose_OwnerRouting(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "own.db"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}

	const ownerID = "trader-A"
	// Owner A holds an open MNQ SHORT on Sim101, entry 29804.50.
	if err := st.Position().Create(&store.TraderPosition{
		TraderID: ownerID, Account: "Sim101", Symbol: "MNQ", Side: "SHORT",
		Quantity: 1, EntryQuantity: 1, EntryPrice: 29804.50, EntryTime: 1, Status: "OPEN",
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	// The frame is received by a DIFFERENT trader's close-sync (trader B, other account).
	s := ntwire.NewTCPServer(nil)
	trB := NewTCPTrader(s, "MNQ", "SimAccountX")
	pb := store.NewPositionBuilder(st.Position())
	trB.recordClose("trader-B", "ex", "ninjatrader", st, pb, ntwire.PositionClosePayload{
		ExitOrderID: "broker-x", SignalID: "x", Symbol: "MNQ", PositionSide: "short", ExitPrice: 29718.00,
		Quantity: 1, ExitReason: "manual", Account: "Sim101",
	})

	// Owner A's row must now be CLOSED (no open row) with the real exit + P&L.
	if open, _ := st.Position().GetOpenPositionByAccountSymbol("Sim101", "MNQ", "SHORT"); open != nil {
		t.Fatalf("owner's row should be CLOSED after owner-routed record; still open id=%d", open.ID)
	}
	closed, err := st.Position().GetClosedPositions(ownerID, 10)
	if err != nil || len(closed) != 1 {
		t.Fatalf("want 1 closed row for owner, got %d (err=%v)", len(closed), err)
	}
	// SHORT P&L = (entry - exit) * qty * $2 = (29804.50 - 29718.00) * 1 * 2 = 173.0.
	if closed[0].ExitPrice != 29718.00 {
		t.Fatalf("exit_price = %.2f, want 29718.00 (real fill, not entry)", closed[0].ExitPrice)
	}
	if closed[0].RealizedPnL != 173.0 {
		t.Fatalf("realized_pnl = %.2f, want 173.0 (×$2 point value)", closed[0].RealizedPnL)
	}
}

// Unmatched evidence retains quantity durably; a price-only park could later
// fabricate a full close. A later entry row must precede the receipt in time.
func TestRecordClose_NoOpenRowPreservesCompleteReceipt(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "pending.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	tr := &TCPTrader{symbol: "MNQ", boundAccount: "Sim101"}
	tr.recordClose("owned", "ex", "ninjatrader", st, store.NewPositionBuilder(st.Position()), ntwire.PositionClosePayload{ExitOrderID: "broker-pending-exit", SignalID: "pending-exit", Symbol: "MNQ", Account: "Sim101", PositionSide: "long", Quantity: 1, ExitPrice: 110})
	receipts, err := st.Position().PendingNT8Exits("Sim101")
	if err != nil || len(receipts) != 1 || receipts[0].Quantity != 1 {
		t.Fatalf("receipt lost: %+v %v", receipts, err)
	}
	if _, ok := takePricedClose("Sim101", "MNQ", "LONG", 1); ok {
		t.Fatal("unsafe price-only fallback populated")
	}
	row := &store.TraderPosition{TraderID: "owned", Account: "Sim101", Symbol: "MNQ", Side: "LONG", Quantity: 3, EntryQuantity: 3, EntryPrice: 100, EntryTime: 1, Status: "OPEN", PlanID: "original"}
	if err := st.Position().Create(row); err != nil {
		t.Fatal(err)
	}
	tr.retryPendingNT8Exits(st)
	var got store.TraderPosition
	st.GormDB().First(&got, row.ID)
	if got.Status != "OPEN" || got.Quantity != 2 || got.PlanID != "original" || got.RealizedPnL != 20 {
		t.Fatalf("pending partial applied incorrectly: %+v", got)
	}
}
