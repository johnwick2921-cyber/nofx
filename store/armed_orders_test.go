package store

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newArmedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	st := NewArmedOrderStore(db)
	if err := st.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// PRE-REOPEN F3 (2026-08-28) — the dead re-arm fix: a TERMINAL row for the
// same (plan, scenario) must come back to life when a new authorization
// arrives. The old Assign-based upsert left terminal rows terminal forever.
func TestUpsertArmReauthorizesTerminalRow(t *testing.T) {
	db := newArmedTestDB(t)
	st := NewArmedOrderStore(db)
	now := time.Now()

	orig := &ArmedOrderDB{
		TraderID: "t1", PlanID: "2026-08-28:planX", Version: 3, Session: "RTH",
		Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 99, TargetPx: 102,
		State: "armed", CreatedAt: now, UpdatedAt: now,
	}
	if err := st.UpsertArm(orig); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	// Terminal: cancelled with lineage from the prior life.
	if err := st.SetState(orig.ID, "cancelled", "gate changed: stale_reeval"); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if err := st.SetSignal(orig.ID, "sig-42"); err != nil {
		t.Fatalf("signal: %v", err)
	}

	// New authorization for the same scenario arrives.
	re := &ArmedOrderDB{
		TraderID: "t1", PlanID: "2026-08-28:planX", Version: 4, Session: "ETH",
		Scenario: "S1", Side: "short", EntryPx: 105, StopPx: 106, TargetPx: 102,
		CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute),
	}
	if err := st.UpsertArm(re); err != nil {
		t.Fatalf("re-arm upsert: %v", err)
	}

	row, err := st.ListNonTerminal()
	if err != nil {
		t.Fatalf("list non-terminal: %v", err)
	}
	if len(row) != 1 {
		t.Fatalf("non-terminal rows = %d, want 1", len(row))
	}
	var got ArmedOrderDB
	if err := db.Where("plan_id = ? AND scenario = ?", re.PlanID, re.Scenario).First(&got).Error; err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if got.State != "armed" {
		t.Fatalf("state = %q, want armed", got.State)
	}
	if got.StateReason != "" {
		t.Fatalf("state_reason = %q, want cleared", got.StateReason)
	}
	if got.SignalID != "" {
		t.Fatalf("signal_id = %q, want cleared on re-arm", got.SignalID)
	}
	if got.FillPrice != 0 || got.FillQuantity != 0 {
		t.Fatalf("fill lineage not cleared: price=%v qty=%d", got.FillPrice, got.FillQuantity)
	}
	if got.Side != "short" || got.EntryPx != 105 || got.StopPx != 106 || got.Version != 4 {
		t.Fatalf("fresh prices not applied: %+v", got)
	}
	if got.ID != orig.ID {
		t.Fatalf("identity not preserved: id %d want %d", got.ID, orig.ID)
	}
}

// Non-terminal rows keep their identity and only refresh prices.
func TestUpsertArmPreservesNonTerminalIdentity(t *testing.T) {
	db := newArmedTestDB(t)
	st := NewArmedOrderStore(db)
	now := time.Now()

	orig := &ArmedOrderDB{
		TraderID: "t1", PlanID: "2026-08-28:planY", Version: 1, Session: "RTH",
		Scenario: "S2", Side: "long", EntryPx: 200, StopPx: 198, TargetPx: 205,
		State: "armed", CreatedAt: now, UpdatedAt: now,
	}
	if err := st.UpsertArm(orig); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := st.SetState(orig.ID, "working", "placed"); err != nil {
		t.Fatalf("working: %v", err)
	}
	if err := st.SetSignal(orig.ID, "sig-7"); err != nil {
		t.Fatalf("signal: %v", err)
	}

	upd := &ArmedOrderDB{
		TraderID: "t1", PlanID: "2026-08-28:planY", Version: 2, Session: "RTH",
		Scenario: "S2", Side: "long", EntryPx: 201, StopPx: 199, TargetPx: 206,
		CreatedAt: now, UpdatedAt: now.Add(time.Minute),
	}
	if err := st.UpsertArm(upd); err != nil {
		t.Fatalf("refresh upsert: %v", err)
	}

	var got ArmedOrderDB
	if err := db.Where("id = ?", orig.ID).First(&got).Error; err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if got.State != "working" || got.SignalID != "sig-7" {
		t.Fatalf("working identity mutated: state=%q signal=%q", got.State, got.SignalID)
	}
	if got.EntryPx != 201 || got.StopPx != 199 || got.TargetPx != 206 || got.Version != 2 {
		t.Fatalf("prices not refreshed: %+v", got)
	}
}
