package ninjatrader

import (
	nt "nofx/provider/ninjatrader"
	"nofx/store"
	"path/filepath"
	"testing"
	"time"
)

func TestBoundBalanceCannotBorrowOtherAccount(t *testing.T) {
	s := nt.NewTCPServer(nil)
	s.SeedAccountBalanceForTest("SimOther", nt.AccountBalancePayload{Account: "SimOther", NetLiquidation: 12345})
	s.SetCurrentAccountForTest("SimOther")
	for _, account := range []string{"SimMissing", ""} {
		tr := NewTCPTrader(s, "MNQ", account)
		if b, err := tr.GetBalance(); err == nil {
			t.Errorf("missing bound balance must be unavailable: bound=%q got=%+v", account, b)
		}
	}
}

func TestReconcileOtherAccountCannotMaskHeldPosition(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "accounts.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Position().CreateOpenPosition(&store.TraderPosition{TraderID: "foreign-trader", ExchangeType: "ninjatrader", ExchangePositionID: "foreign", Symbol: "MNQ", Side: "LONG", Quantity: 1, EntryQuantity: 1, EntryPrice: 29400, Status: "OPEN", Account: "SimOther"}); err != nil {
		t.Fatal(err)
	}
	s := nt.NewTCPServer(nil)
	s.SeedPositionsForTest("Sim101", []nt.OpenPosition{{Symbol: "MNQ", Side: "LONG", Quantity: 1, AvgPrice: 29500}})
	tr := NewTCPTrader(s, "MNQ", "Sim101")
	tr.reconcilePositions("own-trader", "nt", "ninjatrader", st)
	tr.untrackedSince["MNQ|LONG"] = time.Now().UnixMilli() - untrackedGraceMs - 1
	tr.reconcilePositions("own-trader", "nt", "ninjatrader", st)
	rows, err := st.Position().GetOpenPositions("own-trader")
	if err != nil || len(rows) != 1 || rows[0].Account != "Sim101" {
		t.Fatalf("foreign row masked bound position: rows=%+v err=%v", rows, err)
	}
	foreign, err := st.Position().GetOpenPositions("foreign-trader")
	if err != nil || len(foreign) != 1 || foreign[0].Account != "SimOther" {
		t.Fatalf("foreign row changed: %+v %v", foreign, err)
	}
}

func TestAccountResetHonorsPendingOwnershipLock(t *testing.T) {
	tr := NewTCPTrader(nt.NewTCPServer(nil), "MNQ", "Sim101")
	tr.pendingMu.Lock()
	started, done := make(chan struct{}), make(chan struct{})
	go func() { close(started); tr.ResetAccountState(); close(done) }()
	<-started
	completed := false
	select {
	case <-done:
		completed = true
	case <-time.After(30 * time.Millisecond):
	}
	tr.pendingMu.Unlock()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("reset stuck after pending lock released")
	}
	if completed {
		t.Fatal("reset replaced pending maps without acquiring their ownership lock")
	}
}
