// W117 F3 store pins — ApplyNT8Exit applies exactly one owned exit, and a valid
// exit that beats the later cumulative entry update is RETAINED as pending
// (never discarded, never invented). RED: drop the `in.Quantity > row.Quantity`
// pending branch (restore the old error) → the pending pin fails with a hard
// error instead of Pending=true.

package store

import (
	"path/filepath"
	"testing"
	"time"
)

func openNT8Row(t *testing.T, st *Store, qty float64) *TraderPosition {
	t.Helper()
	nowMs := time.Now().UTC().UnixMilli()
	row := &TraderPosition{
		TraderID: "t1", ExchangeType: "ninjatrader",
		ExchangePositionID: "pos-1", Symbol: "MNQ", Side: "LONG",
		Quantity: qty, EntryQuantity: qty, EntryPrice: 100, EntryTime: nowMs - 400_000,
		EntryOrderID: "entry-1", Leverage: 1,
		Status: "OPEN", Source: "armed_entry", Account: "Sim101",
		CreatedAt: nowMs - 400_000, UpdatedAt: nowMs - 400_000,
	}
	if err := st.Position().CreateOpenPosition(row); err != nil {
		t.Fatalf("create open row: %v", err)
	}
	return row
}

func receiptFor(row *TraderPosition, qty, price, pv float64) NT8ExitReceipt {
	return NT8ExitReceipt{
		ID: "receipt-1", Account: row.Account, Symbol: row.Symbol, Side: row.Side,
		SignalID: row.EntryOrderID, // a bracket exit's frame correlates by the ENTRY order id
		TraderID: row.TraderID, ExchangeID: row.ExchangeID,
		ExchangeType: row.ExchangeType, Reason: "tp",
		Quantity: qty, Price: price, PointValue: pv,
		ExitMs: row.EntryTime + 300_000, ReceivedMs: row.EntryTime + 301_000,
	}
}

func TestApplyNT8ExitAppliesAndCloses(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "f3apply.db"))
	if err != nil {
		t.Fatal(err)
	}
	row := openNT8Row(t, st, 2)
	out, err := st.Position().ApplyNT8Exit(receiptFor(row, 2, 105, 2))
	if err != nil {
		t.Fatalf("ApplyNT8Exit: %v", err)
	}
	if !out.Applied || !out.Closed || out.Pending {
		t.Fatalf("a covering exit must apply and close, got %+v", out)
	}
	if out.RealizedPnL != 20 { // (105−100)×2 contracts×$2/pt = +$20
		t.Fatalf("pnl must be +20, got %.2f", out.RealizedPnL)
	}
	var after TraderPosition
	if err := st.GormDB().First(&after, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if after.Status != "CLOSED" || after.CloseReason != ExitCauseFromBroker("tp") {
		t.Fatalf("row must close with the broker's reason, got %+v", after)
	}
	var fills []TraderFill
	st.GormDB().Where("exchange_trade_id = ?", "receipt-1").Find(&fills)
	if len(fills) != 1 || fills[0].RealizedPnL != 20 {
		t.Fatalf("one fill with the exit pnl, got %+v", fills)
	}
}

func TestApplyNT8ExitRetainsPendingWhenRowIncomplete(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "f3pending.db"))
	if err != nil {
		t.Fatal(err)
	}
	// The row's residual is STALE at 1 — the cumulative entry update to 2 has
	// not landed yet. The exit for 2 is valid broker evidence and must be
	// RETAINED, not refused, not trimmed.
	row := openNT8Row(t, st, 1)
	out, err := st.Position().ApplyNT8Exit(receiptFor(row, 2, 105, 2))
	if err != nil {
		t.Fatalf("an incomplete row must not HARD-FAIL the exit (the receipt is retained): %v", err)
	}
	if !out.Pending || out.Applied {
		t.Fatalf("an exit above the stale residual must be PENDING, got %+v", out)
	}
	var after TraderPosition
	if err := st.GormDB().First(&after, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if after.Quantity != 1 || after.Status != "OPEN" {
		t.Fatalf("the pending exit must not mutate the incomplete row, got %+v", after)
	}
	pending, err := st.Position().PendingNT8Exits("Sim101")
	if err != nil || len(pending) != 1 || pending[0].Applied {
		t.Fatalf("the receipt must be retained unapplied, got %+v err=%v", pending, err)
	}
}

func TestApplyNT8ExitRetryAppliesAfterRowCatchesUp(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "f3retry.db"))
	if err != nil {
		t.Fatal(err)
	}
	row := openNT8Row(t, st, 1)
	if out, err := st.Position().ApplyNT8Exit(receiptFor(row, 2, 105, 2)); err != nil || !out.Pending {
		t.Fatalf("first pass must park pending, got %+v err=%v", out, err)
	}
	// The later cumulative entry update lands: residual now covers the exit.
	if err := st.GormDB().Model(&TraderPosition{}).Where("id = ?", row.ID).Updates(map[string]any{"quantity": 2, "entry_quantity": 2}).Error; err != nil {
		t.Fatal(err)
	}
	out, err := st.Position().ApplyNT8Exit(receiptFor(row, 2, 105, 2))
	if err != nil {
		t.Fatalf("the retry must apply once the row catches up: %v", err)
	}
	if !out.Applied || !out.Closed {
		t.Fatalf("the retry must close the row, got %+v", out)
	}
	// A second retry of the SAME receipt is a no-op (idempotent).
	if out2, err := st.Position().ApplyNT8Exit(receiptFor(row, 2, 105, 2)); err != nil || out2.Applied || out2.Pending {
		t.Fatalf("a replayed applied receipt must be a no-op, got %+v err=%v", out2, err)
	}
}

func TestLatestNT8ExitReceiptMs(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "f3fence.db"))
	if err != nil {
		t.Fatal(err)
	}
	row := openNT8Row(t, st, 2)
	if _, err := st.Position().ApplyNT8Exit(receiptFor(row, 2, 105, 2)); err != nil {
		t.Fatal(err)
	}
	ms, err := st.Position().LatestNT8ExitReceiptMs("Sim101")
	if err != nil {
		t.Fatal(err)
	}
	if ms == 0 {
		t.Fatal("the receipt fence must report the receipt's received_ms")
	}
}
