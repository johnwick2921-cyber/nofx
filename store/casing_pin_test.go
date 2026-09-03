package store

import (
	"path/filepath"
	"testing"
	"time"
)

// ── CLASS 28, LIVE INSTANCE (owner ruling 2026-09-03) ────────────────────────
//
// armed_orders stored lowercase, trader_positions stores uppercase, and
// GetOpenPositionBySymbol compared `side = ?` case-sensitively. Against
// position 591 the armed side 'short' matched 0 rows and 'SHORT' matched 1, so
// the fill-time lookup could NEVER find the position it had just filled —
// every armed fill read fill_quantity=0 (10 of 10 on the live table).
//
// Two halves: the reads compare UPPER=UPPER, and the WRITE canonicalizes so the
// tables stop disagreeing at rest.
func TestCasingPinPosition591Shape(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "casing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ps := st.Position()
	now := time.Now().UnixMilli()
	// A position row as trader_positions actually stores it: UPPERCASE.
	if err := ps.CreateOpenPosition(&TraderPosition{
		TraderID: "hoang", ExchangeID: "nt8", ExchangePositionID: "p591", Symbol: "MNQ",
		Side: "SHORT", Quantity: 1, EntryQuantity: 1, EntryPrice: 29200, EntryTime: now,
		Leverage: 1, Status: "OPEN", Account: "Sim101", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	// Looked up by the ARMED side, which is lowercase — the exact live shape.
	got, err := ps.GetOpenPositionBySymbol("hoang", "MNQ", "short")
	if err != nil || got == nil {
		t.Fatalf("lookup by the armed (lowercase) side must return 1 row, got %v err=%v", got, err)
	}
	if got.Side != "SHORT" {
		t.Errorf("the stored row keeps its own casing: %q", got.Side)
	}
	// And the account-scoped sibling, which close-sync routes through.
	if p, err := ps.GetOpenPositionByAccountSymbol("Sim101", "MNQ", "short"); err != nil || p == nil {
		t.Errorf("account-scoped lookup must be case-insensitive too: %v err=%v", p, err)
	}
}

// The write chokepoint: an arm authored lowercase is STORED uppercase, so the
// two tables agree at rest and not merely at each comparison.
func TestCasingCanonicalizesAtTheWriteChokepoint(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "canon.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	l := st.ArmedOrders()
	now := time.Now()
	if err := l.UpsertArm(&ArmedOrderDB{
		TraderID: "hoang", PlanID: "p1", Scenario: "S1", Side: "short", Version: 1,
		EntryPx: 29200, State: "armed", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	got, err := l.GetArm("p1", "S1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.Side != "SHORT" {
		t.Errorf("armed_orders must store the canonical side, got %q — the tables disagree at rest again", got.Side)
	}
	// An unknown side stays visible rather than being coerced to a legal one.
	if s := CanonicalSide(" sideways "); s != "SIDEWAYS" {
		t.Errorf("unknown sides pass through uppercased, got %q", s)
	}
}
