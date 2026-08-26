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

// Migrate creates the bars table + the time-ordered index (additive + idempotent).
func (s *BarHistoryStore) Migrate() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if err := s.db.AutoMigrate(&BarHistoryDB{}); err != nil {
		return err
	}
	return s.db.Exec("CREATE INDEX IF NOT EXISTS idx_bars_sym_tf_time ON bars(symbol, tf, open_time_ms DESC)").Error
}

// InsertBars upserts closed bars with INSERT OR IGNORE (idempotent across
// restarts/backfills — a re-seeded historical batch never duplicates). A
// failure returns the error; the caller logs WARN and never blocks the
// trade loop on it.
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
			placeholders = append(placeholders, "(?,?,?,?,?,?,?,?)")
			args = append(args, r.Symbol, r.TF, r.OpenTimeMs, r.O, r.H, r.L, r.C, r.V)
		}
		q := "INSERT OR IGNORE INTO bars(symbol, tf, open_time_ms, o, h, l, c, v) VALUES " + strings.Join(placeholders, ",")
		if err := s.db.Exec(q, args...).Error; err != nil {
			return err
		}
	}
	return nil
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
