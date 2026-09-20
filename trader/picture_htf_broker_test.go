package trader

import (
	"strings"
	"testing"

	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// W-PICTURE-HTF (2026-09-20) — broker-state consumer + reconciliation sweep.
// These prove the two execution-readiness gaps the CTO named: RECEIVED entry/
// rejection/fill/protective-order events update the CORRECT opportunity, and
// pending/ambiguous submissions recover across restarts WITHOUT another entry.

func pictureSeedBrokerRows(t *testing.T, st *store.Store) (entryRow, otherRow *store.PictureHtfOpportunityDB) {
	t.Helper()
	r1 := &store.PictureHtfOpportunityDB{
		OppKey: "broker|1", TraderID: "trader-1", Stage: "place_pending", Direction: "long",
		SignalID: "sig-entry", StopPx: 98.25, TargetPx: 110, EntryRef: 101.49,
	}
	r2 := &store.PictureHtfOpportunityDB{
		OppKey: "broker|2", TraderID: "trader-1", Stage: "place_pending", Direction: "long",
		SignalID: "sig-other", StopPx: 98.25, TargetPx: 110, EntryRef: 101.49,
	}
	if _, fresh, err := st.PictureHtfClaim(r1); err != nil || !fresh {
		t.Fatalf("claim r1: %v %v", err, fresh)
	}
	if _, fresh, err := st.PictureHtfClaim(r2); err != nil || !fresh {
		t.Fatalf("claim r2: %v %v", err, fresh)
	}
	// Both rows own their submission.
	if _, err := st.PictureHtfClaimSubmission(r1.OppKey, "sig-entry"); err != nil {
		t.Fatalf("own r1: %v", err)
	}
	if _, err := st.PictureHtfClaimSubmission(r2.OppKey, "sig-other"); err != nil {
		t.Fatalf("own r2: %v", err)
	}
	return r1, r2
}

func TestPictureHtfConsumeOrderUpdateEntryLifecycle(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{})
	r1, _ := pictureSeedBrokerRows(t, st)

	// working — moves forward, no fill yet.
	pictureHtfConsumeOrderUpdate(at, ntwire.OrderUpdatePayload{SignalID: "sig-entry", OrderName: "sig-entry", State: "working"})
	got, _, _ := st.PictureHtfGet(r1.OppKey)
	if got.Stage != "working" {
		t.Fatalf("working must move the row forward: %+v", got)
	}
	// fill at 101.60 → filled, actual R:R computed from the row geometry.
	pictureHtfConsumeOrderUpdate(at, ntwire.OrderUpdatePayload{SignalID: "sig-entry", OrderName: "sig-entry", State: "filled", FillPrice: 101.60, Quantity: 1})
	got, _, _ = st.PictureHtfGet(r1.OppKey)
	if got.Stage != "filled" || got.FillPrice != 101.60 || got.FillQty != 1 {
		t.Fatalf("fill must stamp stage/price/qty: %+v", got)
	}
	wantRR := (110 - 101.60) / (101.60 - 98.25)
	if got.FillRR < wantRR-1e-9 || got.FillRR > wantRR+1e-9 {
		t.Fatalf("actual-fill R:R must be computed from the row geometry: got %.4f want %.4f", got.FillRR, wantRR)
	}
	// A duplicate/late working event must NEVER downgrade a filled row.
	pictureHtfConsumeOrderUpdate(at, ntwire.OrderUpdatePayload{SignalID: "sig-entry", OrderName: "sig-entry", State: "working"})
	got, _, _ = st.PictureHtfGet(r1.OppKey)
	if got.Stage != "filled" {
		t.Fatalf("a late event must never downgrade: %+v", got)
	}
}

func TestPictureHtfConsumeOrderUpdateRejectionCarriesReason(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{})
	r1, _ := pictureSeedBrokerRows(t, st)
	pictureHtfConsumeOrderUpdate(at, ntwire.OrderUpdatePayload{
		SignalID: "sig-entry", OrderName: "sig-entry", State: "rejected", Reason: "sim: no market data",
	})
	got, _, _ := st.PictureHtfGet(r1.OppKey)
	if got.Stage != "rejected" || !strings.Contains(got.RejectReason, "no market data") {
		t.Fatalf("the rejection reason must ride the row: %+v", got)
	}
}

func TestPictureHtfConsumeProtectiveLegsUpdateProtectionNotFill(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{})
	r1, _ := pictureSeedBrokerRows(t, st)
	// Entry fills first.
	pictureHtfConsumeOrderUpdate(at, ntwire.OrderUpdatePayload{SignalID: "sig-entry", OrderName: "sig-entry", State: "filled", FillPrice: 101.60, Quantity: 1})
	// The SL leg fills (stop-out): the row must NOT lose its entry fill, but
	// the protection event is recorded.
	pictureHtfConsumeOrderUpdate(at, ntwire.OrderUpdatePayload{SignalID: "sig-entry", OrderName: "sig-entry-sl", State: "filled", FillPrice: 98.25, Quantity: 1})
	got, _, _ := st.PictureHtfGet(r1.OppKey)
	if got.Stage != "filled" || got.FillPrice != 101.60 {
		t.Fatalf("a protective-leg fill must never clobber the entry fill: %+v", got)
	}
	if !strings.Contains(got.BrokerStatus, "protection_sl_filled") {
		t.Fatalf("the SL-leg fill must be recorded as protection evidence: %+v", got)
	}
	// A rejected TP leg keeps the entry fill and records the rejection.
	pictureHtfConsumeOrderUpdate(at, ntwire.OrderUpdatePayload{SignalID: "sig-entry", OrderName: "sig-entry-tp", State: "rejected", Reason: "cancelled by OCO"})
	got, _, _ = st.PictureHtfGet(r1.OppKey)
	if got.Stage != "filled" || got.FillPrice != 101.60 {
		t.Fatalf("protection updates must not touch the entry fill: %+v", got)
	}
	if !strings.Contains(got.BrokerStatus, "protection_tp_rejected") {
		t.Fatalf("the TP-leg rejection must be recorded: %+v", got)
	}
}

func TestPictureHtfConsumeOrderUpdateIsolation(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{})
	r1, r2 := pictureSeedBrokerRows(t, st)
	pictureHtfConsumeOrderUpdate(at, ntwire.OrderUpdatePayload{SignalID: "sig-other", OrderName: "sig-other", State: "filled", FillPrice: 99.9, Quantity: 1})
	got1, _, _ := st.PictureHtfGet(r1.OppKey)
	got2, _, _ := st.PictureHtfGet(r2.OppKey)
	if got1.Stage != "place_pending" {
		t.Fatalf("another signal's event must not touch this row: %+v", got1)
	}
	if got2.Stage != "filled" || got2.FillPrice != 99.9 {
		t.Fatalf("the named row must get the event: %+v", got2)
	}
}

func TestPictureHtfReconcileRecoversFilledAcrossRestart(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{})
	r1, _ := pictureSeedBrokerRows(t, st)
	orig := pictureHtfBrokerLookup
	pictureHtfBrokerLookup = func(a *AutoTrader, signalID string) (ntwire.NT8Order, bool) {
		if signalID == "sig-entry" {
			return ntwire.NT8Order{OrderID: "nt8-1", Name: signalID, State: "Filled", Filled: 1, LimitPrice: 101.60}, true
		}
		return ntwire.NT8Order{}, false
	}
	defer func() { pictureHtfBrokerLookup = orig }()
	pictureHtfReconcilePending(at)
	got, _, _ := st.PictureHtfGet(r1.OppKey)
	if got.Stage != "filled" || got.BrokerOrderID != "nt8-1" || got.FillPrice != 101.60 {
		t.Fatalf("the sweep must recover the fill from the broker book: %+v", got)
	}
	// The recovered row can never be re-submitted (atomic claim requires
	// confirmed; the row is filled).
	if won, _ := st.PictureHtfClaimSubmission(r1.OppKey, "retry"); won {
		t.Fatalf("a recovered row must block re-entry")
	}
}

func TestPictureHtfReconcileRejectedAndAbsent(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{})
	r1, r2 := pictureSeedBrokerRows(t, st)
	orig := pictureHtfBrokerLookup
	pictureHtfBrokerLookup = func(a *AutoTrader, signalID string) (ntwire.NT8Order, bool) {
		if signalID == "sig-entry" {
			return ntwire.NT8Order{OrderID: "nt8-2", Name: signalID, State: "Rejected"}, true
		}
		return ntwire.NT8Order{}, false // sig-other absent from the book
	}
	defer func() { pictureHtfBrokerLookup = orig }()
	pictureHtfReconcilePending(at)
	got1, _, _ := st.PictureHtfGet(r1.OppKey)
	got2, _, _ := st.PictureHtfGet(r2.OppKey)
	if got1.Stage != "rejected" || !strings.Contains(got1.RejectReason, "reconciled") {
		t.Fatalf("rejected must reconcile: %+v", got1)
	}
	if got2.Stage != "place_pending" {
		t.Fatalf("an ABSENT order must stay place_pending (ambiguous — never resent): %+v", got2)
	}
}
