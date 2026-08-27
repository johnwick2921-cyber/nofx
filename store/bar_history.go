package store

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// BAR PERSISTENCE (2026-08-26) — the unblock for replay/calibration.
//
// Every CLOSED OHLCV bar the NT8 BarCache receives is persisted here,
// idempotently (INSERT OR IGNORE on the composite PK). This table feeds:
//   - the volume-levels wave validation replay (VWAP/POC/VAH/VAL/naked-POC),
//   - the swing-k / MSS-FVG / trail-mult calibration queue,
//   - the min-SL replay (joined with structure_json.atr by timestamp).
//
// Retention: BAR_RETENTION_DAYS (env, default 90). 1m MNQ+ES ≈ ~50 bytes/row
// ≈ ~2,900 rows/day → ≈1.5 MB/day → ≈130 MB at the 90-day cap. Trivial.

// BarHistoryDB is one closed OHLCV bar. OpenTimeMs is the bar's OPEN time
// (epoch ms UTC) — the BarCache canonical contract.
type BarHistoryDB struct {
	Symbol     string  `gorm:"column:symbol;primaryKey"`
	TF         string  `gorm:"column:tf;primaryKey"`
	OpenTimeMs int64   `gorm:"column:open_time_ms;primaryKey"`
	O          float64 `gorm:"column:o"`
	H          float64 `gorm:"column:h"`
	L          float64 `gorm:"column:l"`
	C          float64 `gorm:"column:c"`
	V          float64 `gorm:"column:v"`
}

// TableName is the bars table (spec name).
func (BarHistoryDB) TableName() string { return "bars" }

// BarRetentionDays resolves the bars retention window (env BAR_RETENTION_DAYS,
// default 90). A value ≤ 0 means "keep forever".
func BarRetentionDays() int {
	if v := os.Getenv("BAR_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return 90
}

// BarHistoryStore persists closed bars (gorm-backed, like the other sub-stores).
type BarHistoryStore struct {
	db *gorm.DB
}

// NewBarHistoryStore constructs the sub-store.
func NewBarHistoryStore(db *gorm.DB) *BarHistoryStore { return &BarHistoryStore{db: db} }

// Migrate creates the bars table + the natural-key unique index (additive +
// idempotent). On SQLite it also runs the ONE-SHOT integrity migration:
//
//	(1) a safety copy `bars_pre_dedupe_<date>` of the pre-fix table (if absent),
//	(2) DELETE keeping max(rowid) per (symbol, tf, open_time_ms) — 2026-08-26
//	    live table held 17,695 duplicate revisions (F5, 2026-08-27-london-
//	    drought.md), written by the old INSERT OR IGNORE against a table with
//	    no real unique constraint,
//	(3) CREATE UNIQUE INDEX on the natural key so the constraint is real,
//	(4) DELETE tf != '1m' rows — 5m/15m are DERIVED ON READ from 1m (the
//	    stored NT8 aggregates were inconsistent with their 1m constituents).
//
// Idempotent: the heavy steps run only while the unique index is absent; every
// later boot is a no-op.
func (s *BarHistoryStore) Migrate() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if err := s.db.AutoMigrate(&BarHistoryDB{}); err != nil {
		return err
	}
	if s.db.Dialector.Name() == "sqlite" {
		var hasUnique int64
		if err := s.db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_bars_sym_tf_time_unique'").Scan(&hasUnique).Error; err != nil {
			return err
		}
		if hasUnique == 0 {
			// (1) pre-dedupe safety copy (besides the systemd-timer backups).
			today := time.Now().Format("2006-01-02")
			backup := "bars_pre_dedupe_" + today
			if err := s.db.Exec(`CREATE TABLE IF NOT EXISTS "` + backup + `" AS SELECT * FROM bars`).Error; err != nil {
				return err
			}
			// (2) keep max(rowid) per natural key.
			if err := s.db.Exec("DELETE FROM bars WHERE rowid NOT IN (SELECT MAX(rowid) FROM bars GROUP BY symbol, tf, open_time_ms)").Error; err != nil {
				return err
			}
			// (4) 1m-only storage — aggregates derive on read.
			if err := s.db.Exec("DELETE FROM bars WHERE tf <> '1m'").Error; err != nil {
				return err
			}
		}
		// (3) the REAL unique constraint (idempotent; replaces the old plain index).
		if err := s.db.Exec("DROP INDEX IF EXISTS idx_bars_sym_tf_time").Error; err != nil {
			return err
		}
		if err := s.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_bars_sym_tf_time_unique ON bars(symbol, tf, open_time_ms)").Error; err != nil {
			return err
		}
		return nil
	}
	return s.db.Exec("CREATE INDEX IF NOT EXISTS idx_bars_sym_tf_time ON bars(symbol, tf, open_time_ms DESC)").Error
}

// InsertBars upserts closed 1m bars on the natural key — a revision of an
// already-stored bar UPDATES (close/volume move on forming-bar snapshots), and
// the unique index makes duplication impossible (F5: the old INSERT OR IGNORE
// wrote 17,695 duplicate revisions because the table had no real constraint).
// Only tf="1m" rows are stored: 5m/15m aggregates are DERIVED ON READ from 1m.
func (s *BarHistoryStore) InsertBars(rows []BarHistoryDB) error {
	if s == nil || s.db == nil || len(rows) == 0 {
		return nil
	}
	const batch = 200
	for start := 0; start < len(rows); start += batch {
		end := start + batch
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[start:end]
		placeholders := make([]string, 0, len(chunk))
		args := make([]interface{}, 0, len(chunk)*8)
		for _, r := range chunk {
			if r.TF != "1m" {
				continue // derive-on-read: only 1m is stored
			}
			placeholders = append(placeholders, "(?,?,?,?,?,?,?,?)")
			args = append(args, r.Symbol, r.TF, r.OpenTimeMs, r.O, r.H, r.L, r.C, r.V)
		}
		if len(placeholders) == 0 {
			continue
		}
		q := "INSERT INTO bars(symbol, tf, open_time_ms, o, h, l, c, v) VALUES " +
			strings.Join(placeholders, ",") +
			" ON CONFLICT(symbol, tf, open_time_ms) DO UPDATE SET o=excluded.o, h=excluded.h, l=excluded.l, c=excluded.c, v=excluded.v"
		if err := s.db.Exec(q, args...).Error; err != nil {
			return err
		}
	}
	return nil
}

// ClearSince deletes rows with open_time_ms >= sinceMs for (symbol, tf) — the
// BAR-TRUTH backfill wipes the window BEFORE a deep replay repopulates it, so
// previously-misstamped rows can never survive as spurious extras.
func (s *BarHistoryStore) ClearSince(symbol, tf string, sinceMs int64) (int64, error) {
	if s == nil || s.db == nil {
		return 0, nil
	}
	res := s.db.Where("symbol = ? AND tf = ? AND open_time_ms >= ?", symbol, tf, sinceMs).
		Delete(&BarHistoryDB{})
	return res.RowsAffected, res.Error
}

// BarsIntegrity returns the nightly integrity triple: duplicate natural-key
// groups (must be 0), the distinct tfs present (must be {"1m"}), and total rows.
func (s *BarHistoryStore) BarsIntegrity() (dups int64, tfs []string, total int64, err error) {
	if s == nil || s.db == nil {
		return 0, nil, 0, fmt.Errorf("store required")
	}
	if err = s.db.Raw("SELECT COUNT(*) FROM (SELECT symbol, tf, open_time_ms FROM bars GROUP BY symbol, tf, open_time_ms HAVING COUNT(*) > 1)").Scan(&dups).Error; err != nil {
		return 0, nil, 0, err
	}
	if err = s.db.Raw("SELECT DISTINCT tf FROM bars ORDER BY tf").Scan(&tfs).Error; err != nil {
		return 0, nil, 0, err
	}
	if err = s.db.Model(&BarHistoryDB{}).Count(&total).Error; err != nil {
		return 0, nil, 0, err
	}
	return dups, tfs, total, nil
}

// Count returns the persisted bar count (boot line + growth proof).
func (s *BarHistoryStore) Count() (int64, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("store required")
	}
	var n int64
	err := s.db.Model(&BarHistoryDB{}).Count(&n).Error
	return n, err
}

// SymbolTFCount returns the (symbol, tf) pair count (boot line).
func (s *BarHistoryStore) SymbolTFCount() (int64, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("store required")
	}
	var n int64
	err := s.db.Raw("SELECT COUNT(*) FROM (SELECT DISTINCT symbol, tf FROM bars)").Scan(&n).Error
	return n, err
}

// PruneOlderThan deletes bars older than cutoffMs (RETENTION). Returns the
// deleted count.
func (s *BarHistoryStore) PruneOlderThan(cutoffMs int64) (int64, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("store required")
	}
	res := s.db.Exec("DELETE FROM bars WHERE open_time_ms < ?", cutoffMs)
	return res.RowsAffected, res.Error
}

// BarsBetween is the one-line replay read: bars for (symbol, tf) in
// [fromMs, toMs). Join with structure_json.atr by timestamp in scripts.
func (s *BarHistoryStore) BarsBetween(symbol, tf string, fromMs, toMs int64) ([]BarHistoryDB, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("store required")
	}
	var out []BarHistoryDB
	err := s.db.Where("symbol = ? AND tf = ? AND open_time_ms >= ? AND open_time_ms < ?",
		symbol, tf, fromMs, toMs).Order("open_time_ms").Find(&out).Error
	return out, err
}

// RetentionCutoffMs is the prune boundary for the resolved retention window.
func RetentionCutoffMs(now time.Time) int64 {
	days := BarRetentionDays()
	if days <= 0 {
		return 0 // keep forever
	}
	return now.Add(-time.Duration(days) * 24 * time.Hour).UnixMilli()
}
