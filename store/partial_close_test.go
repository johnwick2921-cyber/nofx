package store

import (
	"path/filepath"
	"strings"
	"testing"
)

// PARTIAL-CLOSE (2026-10-03) — the store halves at the production methods.

func newPartialCloseTestStore(t *testing.T) *partialCloseStore {
	t.Helper()
	st, err := New(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st.PartialClose()
}

func TestRecordReduceAndApplyFillUpsertLatestWins(t *testing.T) {
	pc := newPartialCloseTestStore(t)
	r := &PositionReduction{TraderID: "t1", Symbol: "MNQ", Side: "long", ClientID: "rx-1", Quantity: 3, Remaining: -1, Who: "mentor"}
	if err := pc.RecordReduce(r); err != nil {
		t.Fatalf("record: %v", err)
	}
	// The fill is applied by client_id — a part-fill then full-fill pair
	// upserts, never duplicates.
	if err := pc.ApplyReduceFill("rx-1", 100.25, 2); err != nil {
		t.Fatalf("apply fill: %v", err)
	}
	if err := pc.ApplyReduceFill("rx-1", 100.25, 2); err != nil {
		t.Fatalf("apply fill again (idempotent): %v", err)
	}
	rows, err := pc.ListReductions("t1")
	if err != nil || len(rows) != 1 {
		t.Fatalf("list: %v %d", err, len(rows))
	}
	if rows[0].FillPrice != 100.25 || rows[0].Remaining != 2 || rows[0].Quantity != 3 {
		t.Fatalf("want fill=100.25 remaining=2 qty=3, got %+v", rows[0])
	}
	if rows[0].Who != "mentor" {
		t.Fatalf("the ledger row must carry who asked, got %q", rows[0].Who)
	}
}

func TestRecordReduceRejectsGarbage(t *testing.T) {
	pc := newPartialCloseTestStore(t)
	if err := pc.RecordReduce(&PositionReduction{TraderID: "t1", Quantity: 1}); err == nil {
		t.Fatalf("a reduce without a client id must be refused")
	}
	if err := pc.RecordReduce(&PositionReduction{TraderID: "t1", ClientID: "rx-1"}); err == nil {
		t.Fatalf("a reduce without a positive quantity must be refused")
	}
}

func TestStopResizeLifecycleConfirmsOnlyOnTheReport(t *testing.T) {
	pc := newPartialCloseTestStore(t)
	r := &StopResize{TraderID: "t1", Symbol: "MNQ", Side: "long", LegSignalID: "sig-1-sl", Quantity: 2, StopPrice: 29590.5, RequestMs: 1000}
	if err := pc.RecordStopResizeRequest(r); err != nil {
		t.Fatalf("request: %v", err)
	}
	// No report: confirmation refuses and the row stays pending.
	if _, err := pc.ConfirmStopResizeByReport(r.ID); err == nil {
		t.Fatalf("a confirmation with no recorded report must be refused")
	}
	// A report that predates the request is refused (F9).
	if err := pc.RecordStopResizeReport(r.ID, 500, "cancelled"); err != nil {
		t.Fatalf("record old report: %v", err)
	}
	if _, err := pc.ConfirmStopResizeByReport(r.ID); err == nil {
		t.Fatalf("a report predating the request must be refused")
	}
	// A FILLED report never confirms a cancel here.
	if err := pc.RecordStopResizeReport(r.ID, 1500, "filled"); err != nil {
		t.Fatalf("record filled report: %v", err)
	}
	if _, err := pc.ConfirmStopResizeByReport(r.ID); err == nil {
		t.Fatalf("a filled report is the filled path's word, never a confirmation")
	}
	// The qualifying terminal report confirms, carrying the replacement facts.
	if err := pc.RecordStopResizeReport(r.ID, 1600, "cancelled"); err != nil {
		t.Fatalf("record report: %v", err)
	}
	confirmed, err := pc.ConfirmStopResizeByReport(r.ID)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirmed.Quantity != 2 || confirmed.StopPrice != 29590.5 || confirmed.Side != "long" {
		t.Fatalf("the confirmed row must carry the replacement facts, got %+v", confirmed)
	}
	if err := pc.MarkStopResizeDone(r.ID); err != nil {
		t.Fatalf("done: %v", err)
	}
	// A done row is not pending any more.
	rows, err := pc.ListStopResizePending("t1")
	if err != nil || len(rows) != 0 {
		t.Fatalf("pending after done: %v %d", err, len(rows))
	}
}

func TestStopResizeFailedStateIsExplicit(t *testing.T) {
	pc := newPartialCloseTestStore(t)
	r := &StopResize{TraderID: "t1", Symbol: "MNQ", Side: "short", LegSignalID: "sig-2-sl", Quantity: 1, StopPrice: 29590.5, RequestMs: 1000}
	if err := pc.RecordStopResizeRequest(r); err != nil {
		t.Fatalf("request: %v", err)
	}
	if err := pc.MarkStopResizeFailed(r.ID); err != nil {
		t.Fatalf("fail: %v", err)
	}
	rows, err := pc.ListStopResizePending("t1")
	if err != nil || len(rows) != 0 {
		t.Fatalf("a failed resize is not pending: %v %d", err, len(rows))
	}
	if got := rowsState(t, pc, r.ID); !strings.Contains(got, "failed") {
		t.Fatalf("the failed state must be explicit, got %q", got)
	}
}

func rowsState(t *testing.T, pc *partialCloseStore, id int64) string {
	t.Helper()
	var row StopResize
	if err := pc.db.First(&row, id).Error; err != nil {
		t.Fatalf("read: %v", err)
	}
	return row.State
}
