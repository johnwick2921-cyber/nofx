// FLAT MEANS FLAT (2026-09-09, dispatch 104 D1).
//
// The session close must leave the BOOK flat, not merely the position. The
// research's "do not" is explicit: "do not assume a 14:45 CT exit prevents
// trend-day damage… Canceling remaining entries is part of being flat."
//
// The S-LIST CLOSER (2026-08-27) already cancels WORKING arms before flattening,
// and TestSListEODFlatCancelsArmsBeforeFlatten pins that. These two pins cover
// the holes it left, both of which leave a live order at the broker while our
// ledger believes the book is flat.

package trader

import (
	"strings"
	"sync"
	"testing"
	"time"

	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// flatFixture builds the EOD scene with independent control over whether a
// POSITION exists and what STATE the arm is in — the two axes the existing
// fixture holds constant.
func flatFixture(t *testing.T, now time.Time, withPosition bool, armState, signalID string,
	acks <-chan ntwire.OrderUpdatePayload) (*AutoTrader, *wireRecorder) {
	t.Helper()
	offset := 15
	trueV := true
	cfg := store.StrategyConfig{DayPlan: &store.DayPlanConfig{
		PlanEnabled: true,
		Sessions:    []store.DayPlanSessionOverride{{Session: "NY", Enable: &trueV, EODFlatOffsetMin: &offset}},
	}}
	at, st := resetTrader(t, cfg)
	rt := &wireRecorder{MockTrader: &MockTrader{}}
	at.trader = rt
	at.armedSyncSeam = &armedSyncSeam{
		Cancel:  func(sid string) error { rt.record("cancel:" + sid); return nil },
		Stream:  func() <-chan ntwire.OrderUpdatePayload { return acks },
		Timeout: 200 * time.Millisecond,
	}
	if withPosition {
		if err := st.Position().Create(&store.TraderPosition{
			TraderID: at.id, Symbol: "MNQ", Side: "LONG", Account: "Sim101",
			ExchangeType: "ninjatrader", EntryQuantity: 1, Quantity: 1,
			EntryPrice: 30000, EntryTime: now.Add(-2 * time.Hour).UnixMilli(),
			Status: "OPEN",
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.ArmedOrders().UpsertArm(&store.ArmedOrderDB{
		TraderID: at.id, PlanID: "2026-08-18:NY:trader-1", Version: 1, Session: "NY",
		Scenario: "S1", Side: "long", EntryPx: 29950, StopPx: 29970, TargetPx: 29910,
		State: armState, SignalID: signalID,
	}); err != nil {
		t.Fatal(err)
	}
	return at, rt
}

func cancelsFor(rt *wireRecorder, sid string) int {
	n := 0
	for _, e := range rt.snapshot() {
		if e == "cancel:"+sid {
			n++
		}
	}
	return n
}

// TestFlatWithNoPositionStillCancelsRestingArms — E1(a).
//
// enforceEODFlatAt reads the open positions and returns false when there are
// none — BEFORE it reaches cancelArmedOrdersSync. So a session that ends flat
// BY LUCK (nothing filled) leaves every resting arm alive at the broker, past
// the close, into the next session's tape. A day that ends flat by luck is not
// flat, and nothing in the process says otherwise.
func TestFlatWithNoPositionStillCancelsRestingArms(t *testing.T) {
	now := time.Date(2026, 8, 18, 14, 30, 0, 0, chicagoLoc())
	acks := make(chan ntwire.OrderUpdatePayload, 4)
	at, rt := flatFixture(t, now, false, store.StateWorking, "sig-noposition", acks)
	go func() {
		time.Sleep(20 * time.Millisecond)
		acks <- ntwire.OrderUpdatePayload{SignalID: "sig-noposition", State: "cancelled"}
	}()

	at.enforceEODFlatAt(now)

	if got := cancelsFor(rt, "sig-noposition"); got == 0 {
		t.Errorf("NO wire cancel was sent for a resting arm at the close because no position was open — "+
			"the flatten returned on len(positions)==0 before reaching the cancel. The arm is still "+
			"live at the broker while the book reads flat. wire=%v", rt.snapshot())
	}
	rows, err := at.store.ArmedOrders().ListNonTerminal(at.id)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Errorf("%d arm(s) still non-terminal after the close with no position open: %+v", len(rows), rows)
	}
}

// TestFlatCancelsPlacePendingOnTheWire — E1(b).
//
// cancelArmedOrdersSyncWith refuses any row whose state is not exactly
// "working" and writes 'cancelled' into the ledger WITHOUT sending anything:
//
//	if r.State != "working" || r.SignalID == "" || cancelFn == nil || src == nil {
//	    _ = ledger.SetState(r.ID, "cancelled", reason)
//	    continue
//	}
//
// BeginPlacement sets signal_id AND state=place_pending in ONE update, before
// the order reaches the broker — so a place_pending row ALWAYS carries a signal
// id and its order may already be resting. This marks it dead without asking.
// Class 81 (a send read as a settlement) reached by omitting the send entirely.
func TestFlatCancelsPlacePendingOnTheWire(t *testing.T) {
	now := time.Date(2026, 8, 18, 14, 30, 0, 0, chicagoLoc())
	acks := make(chan ntwire.OrderUpdatePayload, 4)
	at, rt := flatFixture(t, now, true, store.StatePlacePending, "sig-pending", acks)
	go func() {
		time.Sleep(20 * time.Millisecond)
		acks <- ntwire.OrderUpdatePayload{SignalID: "sig-pending", State: "cancelled"}
	}()

	at.enforceEODFlatAt(now)

	if got := cancelsFor(rt, "sig-pending"); got == 0 {
		t.Errorf("a place_pending arm was written 'cancelled' with NO wire cancel. BeginPlacement had "+
			"already assigned its signal id, so the order may be resting at the broker right now — "+
			"the ledger says dead and the book may say working. wire=%v", rt.snapshot())
	}
	rows, _ := at.store.ArmedOrders().ListForPlan("2026-08-18:NY:trader-1")
	for _, r := range rows {
		if r.State == store.StateCancelled && !strings.Contains(strings.ToLower(r.StateReason), "confirm") &&
			cancelsFor(rt, r.SignalID) == 0 {
			t.Errorf("row %d reads cancelled with no cancel ever sent (reason %q)", r.ID, r.StateReason)
		}
	}
}

var _ = sync.Mutex{}
