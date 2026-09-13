package store

import (
	"errors"
	"gorm.io/gorm"
	"path/filepath"
	"testing"
)

func TestTraderDeleteOwnershipAndRollback(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "delete.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Trader().Create(&Trader{ID: "target", UserID: "owner", Name: "Target", AIModelID: "m", ExchangeID: "e"}); err != nil {
		t.Fatal(err)
	}
	if err := st.Equity().Save(&EquitySnapshot{TraderID: "target", TotalEquity: 123}); err != nil {
		t.Fatal(err)
	}
	assertRows := func(want int64) {
		t.Helper()
		var traders, equity int64
		st.GormDB().Model(&Trader{}).Where("id = ?", "target").Count(&traders)
		st.GormDB().Model(&EquitySnapshot{}).Where("trader_id = ?", "target").Count(&equity)
		if traders != want || equity != want {
			t.Fatalf("trader=%d equity=%d want=%d", traders, equity, want)
		}
	}
	if err := st.Trader().Delete("other", "target"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("foreign delete must fail, got %v", err)
	}
	assertRows(1)
	if err := st.GormDB().Exec(`CREATE TRIGGER refuse_equity_delete BEFORE DELETE ON trader_equity_snapshots BEGIN SELECT RAISE(ABORT, 'offline injected failure'); END`).Error; err != nil {
		t.Fatal(err)
	}
	if err := st.Trader().Delete("owner", "target"); err == nil {
		t.Error("child deletion failure must be returned")
	}
	assertRows(1)
	if err := st.GormDB().Exec("DROP TRIGGER refuse_equity_delete").Error; err != nil {
		t.Fatal(err)
	}
	if err := st.Trader().Delete("owner", "target"); err != nil {
		t.Fatal(err)
	}
	assertRows(0)
}
