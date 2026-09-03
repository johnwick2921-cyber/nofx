package trader

import (
	"path/filepath"
	"testing"

	"nofx/store"
)

// T5 — armed row 35 (2026-09-03, NY v2 S1) read state=filled with
// fill_quantity=0 beside a trader_positions row of quantity 1. Nothing wrote
// the column, and 0 is also a legal "nothing filled", so the ledger could not
// be read either way.
func TestArmedFillQuantityStamped(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ao := st.ArmedOrders()

	row := store.ArmedOrderDB{
		TraderID: "t1", PlanID: "2026-09-03:NY:t1", Version: 2, Session: "NY",
		Scenario: "S1", Side: "short", EntryPx: 29285, StopPx: 29362.5, TargetPx: 29130,
		State: "armed", EntryClass: "armed_fill",
	}
	if err := ao.UpsertArm(&row); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if row.ID == 0 {
		t.Fatal("no id after upsert")
	}
	if err := ao.SetFillQuantity(row.ID, 1); err != nil {
		t.Fatalf("stamp: %v", err)
	}
	got, err := ao.ListForPlan(row.PlanID)
	if err != nil || len(got) != 1 {
		t.Fatalf("read back: %v n=%d", err, len(got))
	}
	if got[0].FillQuantity != 1 {
		t.Errorf("fill_quantity = %d, want 1 (row 35 read 0 beside a position of 1)", got[0].FillQuantity)
	}
	// A zero is never written over a real quantity — 0 means "not stamped".
	if err := ao.SetFillQuantity(row.ID, 0); err != nil {
		t.Fatalf("zero stamp: %v", err)
	}
	got, _ = ao.ListForPlan(row.PlanID)
	if got[0].FillQuantity != 1 {
		t.Errorf("a zero overwrote a measured quantity: %d", got[0].FillQuantity)
	}
	// ATTRIBUTION: the version the arm BELONGS to, stamped once by UpsertArm.
	if got[0].ArmedUnderVersion != 2 {
		t.Errorf("armed_under_version = %d, want 2", got[0].ArmedUnderVersion)
	}
}

// T5b — a filled or cancelled row never re-logs "⚔️ armed". On 2026-09-03 that
// line fired 4× after row 35 had already filled, because ListNonTerminal
// cannot see a terminal row and the authored branch ran again.
func TestArmedAuthoredNeverRelogsATerminalRow(t *testing.T) {
	for state, want := range map[string]bool{
		"armed": true, "working": true,
		"filled": false, "cancelled": false, "expired": false,
		"FILLED": false, " cancelled ": false,
	} {
		if got := armAuthoredLoggable(state); got != want {
			t.Errorf("armAuthoredLoggable(%q) = %v, want %v", state, got, want)
		}
	}
}

// The fill line and the card must name the version the arm BELONGS to. Rows
// created before the attribution column was written carry 0, and 0 must fall
// back to the mutable Version rather than rendering "armed under v0".
func TestArmedUnderVersionFallsBackHonestly(t *testing.T) {
	if got := armedUnderVersionOf(store.ArmedOrderDB{Version: 3, ArmedUnderVersion: 2}); got != 2 {
		t.Errorf("with attribution stamped, want 2, got %d", got)
	}
	// armed rows 35 and 36 are exactly this shape: created 09:02 and 09:20,
	// before the attribution wave booted at 10:28.
	if got := armedUnderVersionOf(store.ArmedOrderDB{Version: 2, ArmedUnderVersion: 0}); got != 2 {
		t.Errorf("unstamped row must fall back to Version, want 2, got %d", got)
	}
}
