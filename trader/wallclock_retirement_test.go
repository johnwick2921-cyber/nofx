package trader

import (
	"nofx/market"
	nt "nofx/provider/ninjatrader"
	"nofx/store"
	nttrader "nofx/trader/ninjatrader"
	"testing"
	"time"
)

func TestFrozenTapeDoesNotSkipSessionArmRetirement(t *testing.T) {
	at, st := class32ClockTrader(t)
	now := ctTime(t, 2026, 8, 18, 16, 5)
	previous := market.FuturesBarsProvider
	market.FuturesBarsProvider = class32FrozenBars(now)
	t.Cleanup(func() { market.FuturesBarsProvider = previous })
	testNow = func() time.Time { return now }
	t.Cleanup(func() { testNow = nil })
	server := nt.NewTCPServer(nil)
	server.SetAccountsList([]nt.AccountInfo{{Name: "Sim101", IsSim: true}}, "Sim101")
	at.trader = nttrader.NewTCPTrader(server, "MNQ", "Sim101")
	row := store.ArmedOrderDB{TraderID: at.id, PlanID: "expired", Version: 1, Session: "NY", Scenario: "S1", Side: "long", EntryPx: 15600, StopPx: 15590, TargetPx: 15630, State: store.StateArmed}
	if err := st.ArmedOrders().UpsertArm(&row); err != nil {
		t.Fatal(err)
	}
	at.skipNoNewData(now)
	if !at.skipNoNewData(now) {
		t.Fatal("fixture did not freeze the tape")
	}
	at.tickOnce(false)
	if err := st.GormDB().First(&row, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if row.State != store.StateCancelled {
		t.Fatalf("session-ended authorization survived frozen-tape tick: %s", row.State)
	}
}
