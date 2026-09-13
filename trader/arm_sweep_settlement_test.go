package trader

import (
	"testing"
	"time"

	nt "nofx/provider/ninjatrader"
	"nofx/store"
	nttrader "nofx/trader/ninjatrader"
)

func TestArmSweepLeavesPendingCancelForSnapshotConfirmation(t *testing.T) {
	at := class33Trader(t)
	at.trader = nttrader.NewTCPTrader(nt.NewTCPServer(nil), "MNQ", "SimArmSweep")
	ledger := at.store.ArmedOrders()
	now := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	class33Seed(t, at, "pending-at-boot", "pending-entry", "prior-process", store.StateWorking)
	var row store.ArmedOrderDB
	if err := ledger.DB().Where("scenario = ?", "pending-at-boot").First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if err := ledger.RequestCancel(row.ID, "original cancel reason", now.UnixMilli()); err != nil {
		t.Fatal(err)
	}
	sends := 0
	wire := func(string) error { sends++; return nil }
	swept := at.sweepPreBootArmsWith(ledger, wire)
	if err := ledger.DB().First(&row, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if swept != 0 || sends != 0 || row.State != store.StateCancelPending || row.CancelSettledSnapshotID != 0 {
		t.Fatalf("boot bypassed confirmation: swept=%d sends=%d state=%s snapshot=%d", swept, sends, row.State, row.CancelSettledSnapshotID)
	}
	// The real settlement pass must pick up this prior-process row even while
	// the broker book is unavailable, and leave it pending without evidence.
	if settled, pending, re := at.confirmPendingCancels(ledger, wire, now.Add(time.Second)); settled != 0 || pending != 1 || re != 0 {
		t.Fatalf("no-book settlement: got %d/%d/%d, want 0/1/0", settled, pending, re)
	}
	// A fresh but pre-request empty book cannot prove the cancellation.
	oldBook := &store.NT8OrderSnapshot{Account: "SimArmSweep", OrdersJSON: "[]", ReceivedMs: now.Add(-time.Millisecond).UnixMilli()}
	if err := at.store.NT8OrderSnapshots().Insert(oldBook); err != nil {
		t.Fatal(err)
	}
	if settled, pending, re := at.confirmPendingCancels(ledger, wire, now.Add(time.Second)); settled != 0 || pending != 1 || re != 0 {
		t.Fatalf("pre-request book settled a later cancel: got %d/%d/%d", settled, pending, re)
	}
	book := &store.NT8OrderSnapshot{Account: "SimArmSweep", OrdersJSON: `[{"name":"pending-entry","state":"Working","symbol":"MNQ"}]`, ReceivedMs: now.Add(time.Second).UnixMilli()}
	if err := at.store.NT8OrderSnapshots().Insert(book); err != nil {
		t.Fatal(err)
	}
	if settled, pending, re := at.confirmPendingCancels(ledger, wire, now.Add(2*time.Second)); settled != 0 || pending != 1 || re != 0 {
		t.Fatalf("still-resting settlement: got %d/%d/%d, want 0/1/0", settled, pending, re)
	}
	book = &store.NT8OrderSnapshot{Account: "SimArmSweep", OrdersJSON: "[]", ReceivedMs: now.Add(2 * time.Second).UnixMilli()}
	if err := at.store.NT8OrderSnapshots().Insert(book); err != nil {
		t.Fatal(err)
	}
	if settled, pending, re := at.confirmPendingCancels(ledger, wire, now.Add(3*time.Second)); settled != 1 || pending != 0 || re != 0 {
		t.Fatalf("fresh-empty settlement: got %d/%d/%d, want 1/0/0", settled, pending, re)
	}
	if err := ledger.DB().First(&row, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if row.State != store.StateCancelled || book.ID <= 0 || row.CancelSettledSnapshotID != book.ID {
		t.Fatalf("settlement lost evidence: state=%s snapshot=%d want=%d", row.State, row.CancelSettledSnapshotID, book.ID)
	}
	t.Logf("boot swept=0; no book pending=1; resting book pending=1; empty book cancelled with snapshot=%d", book.ID)
}

func TestBootSweepSendDoesNotSettleWorkingEntry(t *testing.T) {
	at := class33Trader(t)
	class33Seed(t, at, "working-at-boot", "working-entry", "prior-process", store.StateWorking)
	ledger := at.store.ArmedOrders()
	sent := 0
	at.sweepPreBootArmsWith(ledger, func(string) error { sent++; return nil })
	var row store.ArmedOrderDB
	if err := ledger.DB().Where("scenario = ?", "working-at-boot").First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if sent != 1 || row.State != store.StateCancelPending || row.CancelSettledSnapshotID != 0 {
		t.Fatalf("send is not settlement: sent=%d state=%s snapshot=%d", sent, row.State, row.CancelSettledSnapshotID)
	}
	at.onArmedOrderUpdate(nt.OrderUpdatePayload{SignalID: "working-entry", OrderName: "working-entry", State: "Cancelled"}, ledger)
	if err := ledger.DB().First(&row, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if row.State != store.StateCancelPending {
		t.Fatalf("receipt bypassed book confirmation: %s", row.State)
	}
	if got, _ := store.BootSweptCount(at.store); got != 0 {
		t.Fatalf("unconfirmed cancellation counted as completed: %d", got)
	}
}
