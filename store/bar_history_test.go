package store

import (
	"path/filepath"
	"testing"
	"time"
)

func newBarStore(t *testing.T) *BarHistoryStore {
	t.Helper()
	st, err := New(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	bh := st.BarHistory()
	if err := bh.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return bh
}

func mkBar(sym, tf string, tMs int64, c float64) BarHistoryDB {
	return BarHistoryDB{Symbol: sym, TF: tf, OpenTimeMs: tMs, O: c, H: c, L: c, C: c, V: 1}
}

func TestBarInsertAndDedup(t *testing.T) {
	st := newBarStore(t)
	rows := []BarHistoryDB{mkBar("MNQ", "1m", 1000, 1.0), mkBar("MNQ", "1m", 2000, 2.0)}
	if err := st.InsertBars(rows); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := st.InsertBars(rows); err != nil { // restart backfill → OR IGNORE
		t.Fatalf("re-insert: %v", err)
	}
	n, _ := st.Count()
	if n != 2 {
		t.Fatalf("count = %d want 2 (dedup on restart)", n)
	}
	got, err := st.BarsBetween("MNQ", "1m", 0, 3000)
	if err != nil || len(got) != 2 {
		t.Fatalf("BarsBetween = %d rows err=%v", len(got), err)
	}
}

func TestBarPrune(t *testing.T) {
	st := newBarStore(t)
	now := time.Now().UnixMilli()
	old := now - 200*24*time.Hour.Milliseconds()
	if err := st.InsertBars([]BarHistoryDB{mkBar("MNQ", "1m", old, 1.0), mkBar("MNQ", "1m", now, 2.0)}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	n, err := st.PruneOlderThan(RetentionCutoffMs(time.Now()))
	if err != nil || n != 1 {
		t.Fatalf("prune = %d err=%v want 1", n, err)
	}
	c, _ := st.Count()
	if c != 1 {
		t.Fatalf("count after prune = %d want 1", c)
	}
}

func TestBarRetentionEnv(t *testing.T) {
	t.Setenv("BAR_RETENTION_DAYS", "7")
	if got := BarRetentionDays(); got != 7 {
		t.Fatalf("retention = %d want 7", got)
	}
	t.Setenv("BAR_RETENTION_DAYS", "")
	if got := BarRetentionDays(); got != 90 {
		t.Fatalf("default retention = %d want 90", got)
	}
}
