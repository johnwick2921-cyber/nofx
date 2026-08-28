package store

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ARMED ORDERS (Wave 2, 2026-08-27) — the durable ledger of scenario-arm
// authorizations the AI granted. The LLM stays the authorizer; Go manages
// WHEN a working order exists (placement/cancel/fill lineage). Every state
// transition is a row update with a reason — nothing armed is ever dropped.

// ArmedOrderDB is one armed scenario (one row per scenario-arm; upserted on
// plan version change, re-armed only by a NEW authorization).
type ArmedOrderDB struct {
	ID int64 `gorm:"primaryKey;autoIncrement"`

	TraderID string `gorm:"index"`
	PlanID   string `gorm:"index"`
	Version  int
	Session  string
	Scenario string // S1, S2, …

	Side     string  // long | short
	EntryPx  float64 // resting limit price
	StopPx   float64 // bracket stop
	TargetPx float64 // bracket target

	// State: armed (authorized, not yet placed) | working (live in NT8) |
	// filled | cancelled | expired. Only the terminal states carry a reason.
	State        string `gorm:"index"`
	StateReason  string
	EntryClass   string // armed_fill when filled (fills bypass stale_reeval)
	SignalID     string // the wire signal_id once working
	FillPrice    float64
	FillQuantity int

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName is the armed_orders table (spec name).
func (ArmedOrderDB) TableName() string { return "armed_orders" }

const armedOrdersDDL = `
CREATE TABLE IF NOT EXISTS armed_orders (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	trader_id     TEXT    NOT NULL DEFAULT '',
	plan_id       TEXT    NOT NULL DEFAULT '',
	version       INTEGER NOT NULL DEFAULT 0,
	session       TEXT    NOT NULL DEFAULT '',
	scenario      TEXT    NOT NULL DEFAULT '',
	side          TEXT    NOT NULL DEFAULT '',
	entry_px      REAL    NOT NULL DEFAULT 0,
	stop_px       REAL    NOT NULL DEFAULT 0,
	target_px     REAL    NOT NULL DEFAULT 0,
	state         TEXT    NOT NULL DEFAULT 'armed',
	state_reason  TEXT    NOT NULL DEFAULT '',
	entry_class   TEXT    NOT NULL DEFAULT '',
	signal_id     TEXT    NOT NULL DEFAULT '',
	fill_price    REAL    NOT NULL DEFAULT 0,
	fill_quantity INTEGER NOT NULL DEFAULT 0,
	created_at    DATETIME,
	updated_at    DATETIME
)`

// ArmedOrderStore persists the armed ledger.
type ArmedOrderStore struct {
	db *gorm.DB
}

// NewArmedOrderStore constructs the sub-store.
func NewArmedOrderStore(db *gorm.DB) *ArmedOrderStore { return &ArmedOrderStore{db: db} }

// Migrate creates the table (sqlite: exact DDL; else AutoMigrate).
func (s *ArmedOrderStore) Migrate() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if s.db.Dialector.Name() == "sqlite" {
		if err := s.db.Exec(armedOrdersDDL).Error; err != nil {
			return err
		}
		return s.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_armed_orders_plan_scenario ON armed_orders(plan_id, scenario)").Error
	}
	return s.db.AutoMigrate(&ArmedOrderDB{})
}

// UpsertArm writes/refreshes the arm row for (plan_id, scenario). Same key =
// same row (state reset to armed only when the spec CHANGED materially —
// entry/stop/target diff >= 2 ticks — the caller decides and passes reset).
func (s *ArmedOrderStore) UpsertArm(row *ArmedOrderDB) error {
	if row == nil || row.PlanID == "" || row.Scenario == "" {
		return fmt.Errorf("plan_id and scenario required")
	}
	// PRE-REOPEN F3 (2026-08-28) — dead re-arm fix: a TERMINAL row for the same
	// (plan, scenario) is re-authorized as a fresh armed row (new identity, no
	// stale fill); a non-terminal row keeps its identity and only its prices
	// are refreshed. The old Assign-based upsert left terminal rows terminal
	// forever, so a legit same-scenario re-arm was impossible and the executor
	// re-logged the dead row every cycle.
	var existing ArmedOrderDB
	err := s.db.Where("plan_id = ? AND scenario = ?", row.PlanID, row.Scenario).First(&existing).Error
	if err == nil {
		if existing.State == "armed" || existing.State == "working" {
			row.ID = existing.ID
			return s.db.Model(&existing).Updates(map[string]any{
				"version": row.Version, "session": row.Session,
				"side": row.Side, "entry_px": row.EntryPx, "stop_px": row.StopPx,
				"target_px": row.TargetPx, "updated_at": row.UpdatedAt,
			}).Error
		}
		// Terminal → RE-AUTHORIZE: fresh armed state, fresh lineage.
		row.ID = existing.ID
		return s.db.Model(&existing).Updates(map[string]any{
			"state": "armed", "state_reason": "", "signal_id": "",
			"entry_class": "", "fill_price": 0, "fill_quantity": 0,
			"trader_id": row.TraderID, "version": row.Version, "session": row.Session,
			"side": row.Side, "entry_px": row.EntryPx, "stop_px": row.StopPx,
			"target_px": row.TargetPx, "created_at": row.CreatedAt, "updated_at": row.UpdatedAt,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return s.db.Create(row).Error
}

// ListNonTerminal returns every armed order that is NOT in a terminal state.
func (s *ArmedOrderStore) ListNonTerminal() ([]ArmedOrderDB, error) {
	var out []ArmedOrderDB
	err := s.db.Where("state IN ('armed','working')").Order("id").Find(&out).Error
	return out, err
}

// SetState transitions one row's state with a reason (the ledger rule: a
// terminal state change is never silent).
func (s *ArmedOrderStore) SetState(id int64, state, reason string) error {
	return s.db.Model(&ArmedOrderDB{}).Where("id = ?", id).
		Updates(map[string]any{"state": state, "state_reason": reason}).Error
}

// SetSignal records the wire signal_id once the resting limit is placed
// (armed → working transition).
func (s *ArmedOrderStore) SetSignal(id int64, signalID string) error {
	return s.db.Model(&ArmedOrderDB{}).Where("id = ?", id).
		Update("signal_id", signalID).Error
}

// ListForPlan returns every armed row of one plan chain (card render + API).
func (s *ArmedOrderStore) ListForPlan(planID string) ([]ArmedOrderDB, error) {
	var out []ArmedOrderDB
	err := s.db.Where("plan_id = ?", planID).Order("id").Find(&out).Error
	return out, err
}

// ListFilled returns one trader's most recent FILLED rows, newest first
// (F3, LONDON-FORENSICS 2026-08-28 — the lineage-repair matcher for
// reconcile-materialized positions reads this).
func (s *ArmedOrderStore) ListFilled(traderID string, limit int) ([]ArmedOrderDB, error) {
	if limit <= 0 {
		limit = 20
	}
	var out []ArmedOrderDB
	err := s.db.Where("trader_id = ? AND state = 'filled'", traderID).
		Order("updated_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

// Touch refreshes UpdatedAt (the stale-working reconnect safety net reads it).
func (s *ArmedOrderStore) Touch(id int64) error {
	return s.db.Model(&ArmedOrderDB{}).Where("id = ?", id).
		Update("updated_at", time.Now()).Error
}
