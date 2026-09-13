package trader

import (
	"errors"
	"nofx/market"
	"nofx/store"
	"testing"
	"time"
)

// Exercise the production placement loop, not just the stop-side predicate.
func TestStopRefusalDoesNotRetireOtherScenarios(t *testing.T) {
	for _, tc := range []struct {
		name  string
		price float64
	}{
		{"accepted through", 29602}, {"unknown price", 0}, {"unproven addon", 29599},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("STOP_ENTRY_SEAM", "on")
			now := time.Date(2026, 9, 14, 15, 0, 0, 0, time.UTC)
			at, st, sigs, _ := shadowWireHarnessAt(t, store.StrategyConfig{}, now)
			ledger := st.ArmedOrders()
			first := store.ArmedOrderDB{TraderID: at.id, PlanID: "refusal", Version: 1, Scenario: "S1", State: store.StateArmed, Side: "long", Kind: "stop_entry", Condition: "reclaim", EntryPx: 29600, StopPx: 29590, TargetPx: 29630}
			other := store.ArmedOrderDB{TraderID: at.id, PlanID: "refusal", Version: 1, Scenario: "S2", State: store.StateArmed, Side: "long", Kind: "limit", EntryPx: 29000, StopPx: 28990, TargetPx: 29030}
			for _, row := range []*store.ArmedOrderDB{&first, &other} {
				if err := ledger.UpsertArm(row); err != nil {
					t.Fatal(err)
				}
			}
			at.runArmedPlacementAt([]market.Kline{{Close: tc.price}}, now.Add(-time.Hour).UnixMilli(), now)
			if err := ledger.DB().First(&other, other.ID).Error; err != nil {
				t.Fatal(err)
			}
			if other.State != store.StateArmed {
				t.Fatalf("refused stop retired unrelated scenario: state=%s reason=%s", other.State, other.StateReason)
			}
			select {
			case p := <-sigs:
				t.Fatalf("refusal sent an order: %+v", p)
			default:
			}
		})
	}
}

func TestStopAttemptRemainsCommittedAfterAmbiguousSend(t *testing.T) {
	at := &AutoTrader{id: "t1"}
	row := armRow(1, "LONG", 29600)
	ledger := &fakeLedger{}
	placer := &fakePlacer{afterRegisterErr: errors.New("synthetic send failed after registration")}
	d := decideStopEntry(row.Side, row.EntryPx, testOffset(), testTick, 29599)
	if !at.placeOneStopEntry(placer, ledger, row, d, 29599, time.Now(), freeSlot()) {
		t.Fatal("registered attempt must retain the in-pass commitment despite send error")
	}
	if ledger.signals[row.ID] == "" {
		t.Fatal("test did not exercise registration")
	}
}
