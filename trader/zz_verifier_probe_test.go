package trader

// VERIFIER PROBE — temporary, deleted after the run. Not part of the fold.

import (
	"testing"
	"time"

	"nofx/store"
	ntTrader "nofx/trader/ninjatrader"
)

// ── FOLD-11 negative: a future-stamped fill on ANOTHER account / instrument
// must still explain NOTHING (the position is flattened as before).

func (w *reconcileWire) probeMustFlatten(t *testing.T, what string) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- w.at.reconcileBeforeOpenNT("MNQ", "long") }()
	if !w.closeSent(3 * time.Second) {
		<-done
		t.Fatalf("%s: a future-stamped row that is NOT on this account/instrument was taken as owner — NOT flattened", what)
	}
	w.s.SeedPositionsForTest("Sim101", nil)
	if err := <-done; err != nil {
		t.Fatalf("%s: after the flatten the open proceeds: %v", what, err)
	}
}

func probeFutureArmedFill(t *testing.T, w *reconcileWire, traderID, sid string) {
	t.Helper()
	led := w.st.ArmedOrders()
	r := &store.ArmedOrderDB{TraderID: traderID, PlanID: "p-probe-" + traderID, Scenario: "S1", Version: 1, State: "armed", Side: "short", EntryPx: 29000, StopPx: 29010, TargetPx: 28970}
	if err := led.UpsertArm(r); err != nil {
		t.Fatal(err)
	}
	if err := led.BeginPlacement(r.ID, sid); err != nil {
		t.Fatal(err)
	}
	if err := led.SetState(r.ID, store.StateFilled, "fixture"); err != nil {
		t.Fatal(err)
	}
	if err := led.DB().Model(&store.ArmedOrderDB{}).Where("id = ?", r.ID).
		UpdateColumn("updated_at", time.Now().Add(30*time.Second)).Error; err != nil {
		t.Fatal(err)
	}
}

func TestProbeFold11FutureFillOtherAccountNotOwned(t *testing.T) {
	w := newReconcileWire(t)
	other := &AutoTrader{id: "probe-other-acct", store: w.st, trader: ntTrader.NewTCPTrader(w.s, "MNQ", "Sim202")}
	other.config.NinjaTraderSymbol = "MNQ"
	registerPostExitDispatch(other)
	t.Cleanup(func() { unregisterPostExitDispatch(other) })
	probeFutureArmedFill(t, w, other.id, "sig-probe-acct")
	w.probeMustFlatten(t, "armed, other account")
}

func TestProbeFold11FutureFillOtherInstrumentNotOwned(t *testing.T) {
	w := newReconcileWire(t)
	other := &AutoTrader{id: "probe-other-inst", store: w.st, trader: ntTrader.NewTCPTrader(w.s, "ES", "Sim101")}
	other.config.NinjaTraderSymbol = "ES"
	registerPostExitDispatch(other)
	t.Cleanup(func() { unregisterPostExitDispatch(other) })
	probeFutureArmedFill(t, w, other.id, "sig-probe-inst")
	w.probeMustFlatten(t, "armed, other instrument")
}

func TestProbeFold11FuturePictureFillOtherAccountNotOwned(t *testing.T) {
	for _, c := range []struct{ name, acct, sym string }{{"acct", "Sim202", "MNQ"}, {"inst", "Sim101", "ES"}} {
		t.Run(c.name, func(t *testing.T) {
			w := newReconcileWire(t)
			key := "pic-probe-" + c.name
			if _, _, err := w.st.PictureHtfClaim(&store.PictureHtfOpportunityDB{OppKey: key, TraderID: w.at.id, Account: c.acct, Symbol: c.sym, Direction: "short", Stage: "confirmed"}); err != nil {
				t.Fatal(err)
			}
			if err := w.st.PictureHtfTransition(key, store.StateFilled, "fixture"); err != nil {
				t.Fatal(err)
			}
			if err := w.st.GormDB().Model(&store.PictureHtfOpportunityDB{}).Where("opp_key = ?", key).
				UpdateColumn("updated_at", time.Now().Add(30*time.Second)).Error; err != nil {
				t.Fatal(err)
			}
			w.probeMustFlatten(t, "picture "+c.name)
		})
	}
}

// ── FOLD-12: a row that is NOT cancel_pending is never paced, even with a
// pace record and a recent request stamp on it.
func TestProbeFold12NonPendingRowIsNeverPaced(t *testing.T) {
	for _, st := range []string{store.StateWorking, store.StatePlacePending} {
		t.Run(st, func(t *testing.T) {
			r, sid := windowSweepRig(t, "probe-fold12-nonpending-"+st)
			sweep := r.windowPass(e13Lead, sid)
			if sweep == 0 {
				t.Fatal("fixture: first sweep must send")
			}
			got := r.row("S1")
			if got.State != store.StateCancelPending {
				t.Fatalf("fixture: %s", cancelBrief(got))
			}
			if err := r.st.ArmedOrders().DB().Model(&store.ArmedOrderDB{}).Where("id = ?", got.ID).UpdateColumn("state", st).Error; err != nil {
				t.Fatal(err)
			}
			if n := r.windowPass(e13Lead.Add(4*time.Second), sid); n != sweep {
				t.Fatalf("a %s row (pace record 4s old) was paced: sent %d, want %d", st, n, sweep)
			}
		})
	}
}

// ── FOLD-12: a lost cancel is re-sent every >=30 s over a long window (no
// starvation), never twice within 30 s; report cancel_attempts vs the cap.
func TestProbeFold12LongWindowCadence(t *testing.T) {
	r, sid := windowSweepRig(t, "probe-fold12-cadence")
	var sends []time.Duration
	for dt := time.Duration(0); dt <= 150*time.Second; dt += 5 * time.Second {
		if n := r.windowPass(e13Lead.Add(dt), sid); n > 0 {
			sends = append(sends, dt)
		}
	}
	got := r.row("S1")
	t.Logf("sends at %v; %s; settlement cap=%d", sends, cancelBrief(got), cancelReRequestMax())
	want := []time.Duration{0, 30 * time.Second, 60 * time.Second, 90 * time.Second, 120 * time.Second, 150 * time.Second}
	if len(sends) != len(want) {
		t.Fatalf("send cadence %v, want %v", sends, want)
	}
	for i := range want {
		if sends[i] != want[i] {
			t.Fatalf("send cadence %v, want %v", sends, want)
		}
	}
}

// ── FOLD-12: "requested" includes another path's request (ledger stamp).
func TestProbeFold12OtherPathRequestAge(t *testing.T) {
	for _, c := range []struct {
		age  time.Duration
		send bool
	}{{10 * time.Second, false}, {60 * time.Second, true}} {
		t.Run(c.age.String(), func(t *testing.T) {
			r, sid := windowSweepRig(t, "probe-fold12-other-"+c.age.String())
			got := r.row("S1")
			if err := r.st.ArmedOrders().RequestCancel(got.ID, "other path", e13Lead.Add(-c.age).UnixMilli()); err != nil {
				t.Fatal(err)
			}
			n := r.windowPass(e13Lead, sid)
			if (n > 0) != c.send {
				t.Fatalf("request %s old by another path: sent %d frame(s), want send=%v", c.age, n, c.send)
			}
		})
	}
}
