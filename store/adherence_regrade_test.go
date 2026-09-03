package store

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func regradeStore(t *testing.T) *Store {
	t.Helper()
	st, err := New(filepath.Join(t.TempDir(), "rg.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func seedGraded(t *testing.T, st *Store, id int64, source, scenario, grade string, matched, version int) {
	t.Helper()
	now := time.Now().UnixMilli()
	p := &TraderPosition{
		TraderID: "hoang", ExchangeID: "nt8", ExchangePositionID: "x" + scenario + grade + source,
		Symbol: "MNQ", Side: "SHORT", Quantity: 1, EntryQuantity: 1, EntryPrice: 29000,
		EntryTime: now, Leverage: 1, Status: "OPEN", Source: source, CreatedAt: now, UpdatedAt: now,
	}
	if err := st.Position().CreateOpenPosition(p); err != nil {
		t.Fatal(err)
	}
	if err := st.GormDB().Exec(`UPDATE trader_positions SET id=?, status='CLOSED', adherence_grade=?,
plan_matched=?, cited_scenario_id=?, plan_version=? WHERE id=?`, id, grade, matched, scenario, version, p.ID).Error; err != nil {
		t.Fatal(err)
	}
}

// The migration regrades EXACTLY the four late-stamped reconcile closes, and
// leaves every row that legitimately earned its D alone.
func TestAdherenceRegradeScopeIsExactlyFour(t *testing.T) {
	st := regradeStore(t)
	// The four in scope.
	seedGraded(t, st, 575, "reconcile", "S2", "D", 1, 4)
	seedGraded(t, st, 584, "reconcile", "S2", "D", 1, 6)
	seedGraded(t, st, 586, "reconcile", "S3", "D", 1, 5)
	seedGraded(t, st, 591, "reconcile", "S1", "D", 1, 2)
	// The three the first count wrongly included.
	seedGraded(t, st, 530, "system", "off-plan", "D", 0, 1)         // genuinely off-plan
	seedGraded(t, st, 572, "e7_farside_test", "TEST-E7", "D", 1, 3) // test seam
	seedGraded(t, st, 582, "armed_entry", "S2", "D", 0, 3)          // direction mismatch

	if got := len(AdherenceRegradeIDs); got != 4 {
		t.Fatalf("the ruled scope is 4 rows, the list holds %d", got)
	}
	st.RegradeStuckAdherence()

	grade := func(id int64) string {
		var g string
		st.GormDB().Raw(`SELECT COALESCE(adherence_grade,'') FROM trader_positions WHERE id=?`, id).Scan(&g)
		return g
	}
	for _, id := range []int64{575, 584, 586, 591} {
		if g := grade(id); g != "" {
			t.Errorf("id %d must be cleared for regrading, still %q", id, g)
		}
	}
	for _, id := range []int64{530, 572, 582} {
		if g := grade(id); g != "D" {
			t.Errorf("id %d earned its D and must keep it, got %q", id, g)
		}
	}
}

// Idempotent, and a row already regraded is skipped by NAME rather than
// silently re-cleared.
func TestAdherenceRegradeIsIdempotentAndNamesSkips(t *testing.T) {
	st := regradeStore(t)
	seedGraded(t, st, 575, "reconcile", "S2", "D", 1, 4)
	seedGraded(t, st, 584, "reconcile", "S2", "B", 1, 6) // already regraded → skip
	st.RegradeStuckAdherence()
	var g string
	st.GormDB().Raw(`SELECT COALESCE(adherence_grade,'') FROM trader_positions WHERE id=584`).Scan(&g)
	if g != "B" {
		t.Errorf("an already-regraded row must not be re-cleared, got %q", g)
	}
	// Second run changes nothing.
	st.RegradeStuckAdherence()
	st.GormDB().Raw(`SELECT COALESCE(adherence_grade,'') FROM trader_positions WHERE id=584`).Scan(&g)
	if g != "B" {
		t.Errorf("second run touched a skipped row: %q", g)
	}
}

// The boot line reports the count from the LIST, never a typed number.
func TestAdherenceRegradeBootLineIsCounted(t *testing.T) {
	line := AdherenceRegradeBootLine()
	if !strings.Contains(line, "adherence regraded=4") {
		t.Errorf("boot line must report the ruled count: %s", line)
	}
	for _, id := range []string{"575", "584", "586", "591"} {
		if !strings.Contains(line, id) {
			t.Errorf("boot line must name the ids it will touch, missing %s: %s", id, line)
		}
	}
	if !strings.Contains(line, "keeps its D") {
		t.Errorf("boot line must say what is NOT regraded: %s", line)
	}
	t.Logf("boot: %s", line)
}
