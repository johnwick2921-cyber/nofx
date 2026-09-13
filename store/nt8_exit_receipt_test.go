package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNT8ExitReceiptsPersistPendingAndAccumulateFees(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pending.db")
	st, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	receipt := NT8ExitReceipt{ID: "receipt1", Account: "Sim101", Symbol: "MNQ", Side: "LONG", SignalID: "exit1", ExchangeID: "nt", ExchangeType: "ninjatrader", Reason: "limit", Quantity: 1, Price: 110, PointValue: 2, Fee: .5, ExitMs: time.Now().UnixMilli()}
	r, err := st.Position().ApplyNT8Exit(receipt)
	if err != nil || !r.Pending || r.Applied {
		t.Fatalf("pending: %+v %v", r, err)
	}
	st.Close()
	st, err = New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	pending, err := st.Position().PendingNT8Exits("Sim101")
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending receipt lost on reopen: %+v %v", pending, err)
	}
	row := &TraderPosition{TraderID: "t", Account: "Sim101", Symbol: "MNQ", Side: "LONG", Quantity: 3, EntryQuantity: 3, EntryPrice: 100, EntryTime: receipt.ExitMs - 1, Status: "OPEN", EntryOrderID: "entry", Fee: 1.25}
	if err := st.Position().Create(row); err != nil {
		t.Fatal(err)
	}
	r, err = st.Position().ApplyNT8Exit(pending[0])
	if err != nil || !r.Applied || r.Closed || r.Position.Quantity != 2 || r.Position.Fee != 1.75 || r.RealizedPnL != 20 {
		t.Fatalf("first execution: %+v %v", r, err)
	}
	receipt.ID = "receipt2"
	receipt.SignalID = "exit2"
	receipt.Quantity = 2
	receipt.Price = 120
	receipt.Fee = .75
	r, err = st.Position().ApplyNT8Exit(receipt)
	if err != nil || !r.Closed || r.Position.Fee != 2.5 || r.Position.RealizedPnL != 100 || r.Position.PnlCorrected == nil || *r.Position.PnlCorrected != 100 {
		t.Fatalf("final fees/pnl: %+v %v", r, err)
	}
	r, err = st.Position().ApplyNT8Exit(receipt)
	if err != nil || r.Applied {
		t.Fatalf("duplicate: %+v %v", r, err)
	}
	var fills []TraderFill
	st.GormDB().Order("id").Find(&fills)
	if len(fills) != 2 || fills[0].Commission != .5 || fills[1].Commission != .75 {
		t.Fatalf("execution fees wrong: %+v", fills)
	}
}
func TestNT8PendingExitCannotConsumeLaterPosition(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "later.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	x := NT8ExitReceipt{ID: "old", Account: "Sim101", Symbol: "MNQ", Side: "LONG", SignalID: "oldexit", Quantity: 1, Price: 110, PointValue: 2, ExitMs: 100}
	if _, err := st.Position().ApplyNT8Exit(x); err != nil {
		t.Fatal(err)
	}
	row := &TraderPosition{TraderID: "t", Account: "Sim101", Symbol: "MNQ", Side: "LONG", Quantity: 1, EntryQuantity: 1, EntryPrice: 100, EntryTime: 101, Status: "OPEN"}
	st.Position().Create(row)
	r, err := st.Position().ApplyNT8Exit(x)
	if err != nil || r.Applied || !r.Pending {
		t.Fatalf("old receipt consumed future row: %+v %v", r, err)
	}
}

func TestFinalExitRestoresCumulativeEntryBasisForHistory(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "basis.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	notional := 200.0
	row := &TraderPosition{TraderID: "t", Account: "Sim101", Symbol: "MNQ", Side: "LONG", Quantity: 2, EntryQuantity: 2, EntryNotional: &notional, EntryPrice: 100, EntryTime: 1, Status: "OPEN"}
	st.Position().Create(row)
	x := NT8ExitReceipt{ID: "basis1", Account: "Sim101", Symbol: "MNQ", Side: "LONG", SignalID: "exit1", Quantity: 1, Price: 110, PointValue: 2, ExitMs: 100}
	if _, err := st.Position().ApplyNT8Exit(x); err != nil {
		t.Fatal(err)
	}
	if err := st.GormDB().Model(&TraderPosition{}).Where("id = ?", row.ID).Updates(map[string]any{"quantity": 2, "entry_quantity": 3, "entry_notional": 310, "entry_price": 105}).Error; err != nil {
		t.Fatal(err)
	}
	x.ID = "basis2"
	x.SignalID = "exit2"
	x.Quantity = 2
	x.Price = 120
	r, err := st.Position().ApplyNT8Exit(x)
	if err != nil || !r.Closed || r.Position.RealizedPnL != 80 || r.Position.EntryPrice != 310.0/3 || r.Position.EntryNotional == nil || *r.Position.EntryNotional != 310 {
		t.Fatalf("final history basis: %+v %v", r, err)
	}
}

func TestExcessExitReceiptWaitsForCumulativeEntryWithoutLosingEvidence(t *testing.T) {
	st := newPlanTestStore(t)
	p := &TraderPosition{TraderID: "t", Account: "Sim101", Symbol: "MNQ", Side: "LONG", EntryOrderID: "entry", Quantity: 1, EntryQuantity: 1, EntryPrice: 100, EntryTime: 1, Status: "OPEN"}
	if err := st.Position().Create(p); err != nil {
		t.Fatal(err)
	}
	receipt := NT8ExitReceipt{ID: "exit-after-entry-partial", Account: "Sim101", Symbol: "MNQ", Side: "LONG", SignalID: "entry", TraderID: "t", Reason: "tp", Quantity: 2, Price: 110, PointValue: 2, ExitMs: 10, ReceivedMs: 20}
	result, err := st.Position().ApplyNT8Exit(receipt)
	if err != nil || !result.Pending || result.Applied {
		t.Fatalf("excess evidence discarded/applied: %+v %v", result, err)
	}
	pending, err := st.Position().PendingNT8Exits("Sim101")
	if err != nil || len(pending) != 1 || pending[0].Quantity != 2 || pending[0].ReceivedMs != 20 {
		t.Fatalf("original evidence lost: %+v %v", pending, err)
	}
	var got TraderPosition
	st.GormDB().First(&got, p.ID)
	if got.Quantity != 1 || got.RealizedPnL != 0 || got.Status != "OPEN" {
		t.Fatalf("pending receipt invented fill: %+v", got)
	}
	// This isolated store test supplies subsequently known entry quantity. The
	// trader tests separately exercise the real cumulative entry callback.
	if err := st.GormDB().Model(p).Updates(map[string]any{"quantity": 2, "entry_quantity": 2}).Error; err != nil {
		t.Fatal(err)
	}
	result, err = st.Position().ApplyNT8Exit(pending[0])
	if err != nil || !result.Applied || !result.Closed || result.RealizedPnL != 40 {
		t.Fatalf("retained receipt not applied after entry: %+v %v", result, err)
	}
	result, err = st.Position().ApplyNT8Exit(pending[0])
	if err != nil || result.Applied {
		t.Fatalf("retry duplicated exit: %+v %v", result, err)
	}
}
