package store

import "testing"

func TestEntryNotionalMigrationPreservesUnknownLegacyBasis(t *testing.T) {
	st := newPlanTestStore(t)
	p := &TraderPosition{TraderID: "legacy", Symbol: "MNQ", Account: "Sim101", Side: "LONG", Quantity: 1, EntryQuantity: 2, EntryPrice: 100, EntryTime: 1, Status: "OPEN"}
	if err := st.Position().Create(p); err != nil {
		t.Fatal(err)
	}
	// Recreate the pre-field schema, then invoke the actual migration twice.
	if err := st.GormDB().Exec("ALTER TABLE trader_positions DROP COLUMN entry_notional").Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := st.Position().InitTables(); err != nil {
			t.Fatal(err)
		}
	}
	var got TraderPosition
	if err := st.GormDB().First(&got, p.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.EntryNotional != nil || got.EntryQuantity != 2 || got.Quantity != 1 || got.EntryPrice != 100 {
		t.Fatalf("migration fabricated or changed legacy basis: %+v", got)
	}
}
