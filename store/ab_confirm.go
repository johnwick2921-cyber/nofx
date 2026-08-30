package store

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AB-CONFIRM SHADOW (E8, entry-mechanics 2026-08-30) — the Sep-9 courtroom
// table. Per armed/authored setup, the machine logs 4 COUNTERFACTUAL fills
// (touch · 1x5m_close · 2x5m_close · 1m_mss): what each entry rule would have
// filled, its MFE/MAE, its target/stop outcome, and the time-to-fill.
// ZERO effect on real paths — this table is written by nothing except the
// shadow logger and read by nothing except replays/reports.

type AbConfirmLogDB struct {
	ID int64 `gorm:"primaryKey;autoIncrement"`

	TraderID string `gorm:"index"`
	PlanID   string `gorm:"index"`
	Version  int
	Session  string
	Scenario string
	Rule     string // touch | 1x5m_close | 2x5m_close | 1m_mss

	FillPx      float64 // counterfactual fill price
	MFE         float64 // max favorable excursion (pts, favorable sign)
	MAE         float64 // max adverse excursion (pts)
	Outcome     string  // target | stop | open (neither by snapshot end)
	TimeToFillMs int64  // entry bar open − plan birth

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (AbConfirmLogDB) TableName() string { return "ab_confirm_log" }

const abConfirmDDL = `
CREATE TABLE IF NOT EXISTS ab_confirm_log (
	id              INTEGER PRIMARY KEY AUTOINCREMENT,
	trader_id       TEXT    NOT NULL DEFAULT '',
	plan_id         TEXT    NOT NULL DEFAULT '',
	version         INTEGER NOT NULL DEFAULT 0,
	session         TEXT    NOT NULL DEFAULT '',
	scenario        TEXT    NOT NULL DEFAULT '',
	rule            TEXT    NOT NULL DEFAULT '',
	fill_px         REAL    NOT NULL DEFAULT 0,
	mfe             REAL    NOT NULL DEFAULT 0,
	mae             REAL    NOT NULL DEFAULT 0,
	outcome         TEXT    NOT NULL DEFAULT 'open',
	time_to_fill_ms INTEGER NOT NULL DEFAULT 0,
	created_at      DATETIME,
	updated_at      DATETIME
)`

// AbConfirmStore persists the shadow A/B table (E8).
type AbConfirmStore struct {
	db *gorm.DB
}

func NewAbConfirmStore(db *gorm.DB) *AbConfirmStore { return &AbConfirmStore{db: db} }

// Migrate creates the table + the per-(plan,scenario,rule) unique index so the
// logger is idempotent by construction.
func (s *AbConfirmStore) Migrate() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if s.db.Dialector.Name() == "sqlite" {
		if err := s.db.Exec(abConfirmDDL).Error; err != nil {
			return err
		}
		return s.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_ab_confirm_key ON ab_confirm_log(plan_id, version, scenario, rule)").Error
	}
	return s.db.AutoMigrate(&AbConfirmLogDB{})
}

// Upsert writes one counterfactual row idempotently (same key = same row).
func (s *AbConfirmStore) Upsert(row *AbConfirmLogDB) error {
	if row == nil || row.PlanID == "" || row.Scenario == "" || row.Rule == "" {
		return fmt.Errorf("plan_id, scenario and rule required")
	}
	var existing AbConfirmLogDB
	err := s.db.Where("plan_id = ? AND version = ? AND scenario = ? AND rule = ?",
		row.PlanID, row.Version, row.Scenario, row.Rule).First(&existing).Error
	if err == nil {
		row.ID = existing.ID
		return s.db.Model(&existing).Updates(map[string]any{
			"trader_id": row.TraderID, "session": row.Session,
			"fill_px": row.FillPx, "mfe": row.MFE, "mae": row.MAE,
			"outcome": row.Outcome, "time_to_fill_ms": row.TimeToFillMs,
			"updated_at": row.UpdatedAt,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return s.db.Create(row).Error
}

// Has reports whether the (plan,version,scenario,rule) shadow row exists —
// the executor's once-per-plan dedup check.
func (s *AbConfirmStore) Has(planID string, version int, scenario, rule string) bool {
	var n int64
	if err := s.db.Model(&AbConfirmLogDB{}).Where(
		"plan_id = ? AND version = ? AND scenario = ? AND rule = ?",
		planID, version, scenario, rule).Count(&n).Error; err != nil {
		return false
	}
	return n > 0
}
