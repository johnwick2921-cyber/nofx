package store

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"strings"

	"nofx/logger"
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
	// Version is the LAST plan version that touched this row, not the one it was
	// armed under: UpsertArm overwrites it on every re-authorization. Documented
	// 2026-09-02 after the cadence audit trusted it as "armed under" and could
	// not defend the reading. Use ArmedUnderVersion for attribution.
	Version int
	// ArmedUnderVersion is set ONCE, when the arm is first authorized, and is
	// never overwritten. This is the version the arm actually belongs to.
	ArmedUnderVersion int `gorm:"index"`
	Session           string
	Scenario          string // S1, S2, …

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

	// E4 (entry-mechanics 2026-08-30) — split-entry legs: a two-leg arm writes
	// TWO rows sharing (plan_id, scenario) distinguished by LegIndex. LegCount
	// = the pair size (2 for split arms, 0 for legacy single arms). Kind =
	// "limit" (default) | "stop_entry" (E7).
	// BootID (CLASS 33, 2026-09-02) — the process that AUTHORED this row
	// (store.ProcessBootID). A non-terminal row whose BootID differs from the
	// running process was placed by a DEAD process: its broker order has no
	// listener, so the boot sweep cancels it before anything is re-armed.
	// Never refreshed on a same-identity re-arm — the stamp must survive an
	// upsert or the sweep would lose its evidence.
	BootID string `gorm:"index"`

	LegIndex int    `gorm:"default:0"`
	LegCount int    `gorm:"default:0"`
	Kind     string `gorm:"default:''"`
	// Condition (arms-follow-bias 2026-09-04) — the scenario condition this arm
	// was authored from. The executor needs it to tell a PRIMARY stop-entry
	// (reclaim) from the E7 no-retest FALLBACK, which share a Kind. Legacy rows
	// carry '' = UNKNOWN, which is never treated as a condition.
	Condition string `gorm:"default:''"`

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
	leg_index     INTEGER NOT NULL DEFAULT 0,
	leg_count     INTEGER NOT NULL DEFAULT 0,
	kind          TEXT    NOT NULL DEFAULT '',
	boot_id       TEXT    NOT NULL DEFAULT '',
	created_at    DATETIME,
	updated_at    DATETIME
)`

// ArmedOrderStore persists the armed ledger.
type ArmedOrderStore struct {
	db *gorm.DB
}

// NewArmedOrderStore constructs the sub-store.
func NewArmedOrderStore(db *gorm.DB) *ArmedOrderStore { return &ArmedOrderStore{db: db} }

// Migrate creates the table (sqlite: exact DDL; else AutoMigrate). E4 adds
// the leg/kind columns idempotently for EXISTING databases (ALTER TABLE ADD
// COLUMN is a no-op-safe guarded by pragma table_info).
func (s *ArmedOrderStore) Migrate() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if s.db.Dialector.Name() == "sqlite" {
		if err := s.db.Exec(armedOrdersDDL).Error; err != nil {
			return err
		}
		for _, col := range []struct{ name, decl string }{
			{"leg_index", "INTEGER NOT NULL DEFAULT 0"},
			{"leg_count", "INTEGER NOT NULL DEFAULT 0"},
			{"kind", "TEXT NOT NULL DEFAULT ''"},
			{"boot_id", "TEXT NOT NULL DEFAULT ''"},   // class 33 — pre-boot decidability
			{"condition", "TEXT NOT NULL DEFAULT ''"}, // arms-follow-bias 2026-09-04
			// ATTRIBUTION (2026-09-02): the version the arm was FIRST authorized
			// under. 0 on legacy rows; UpsertArm adopts their current version
			// once, so the table self-heals without a guessing migration.
			{"armed_under_version", "INTEGER NOT NULL DEFAULT 0"},
		} {
			var n int64
			if err := s.db.Raw("SELECT COUNT(*) FROM pragma_table_info('armed_orders') WHERE name = ?", col.name).Scan(&n).Error; err != nil {
				return err
			}
			if n == 0 {
				if err := s.db.Exec(fmt.Sprintf("ALTER TABLE armed_orders ADD COLUMN %s %s", col.name, col.decl)).Error; err != nil {
					return err
				}
			}
		}
		// E4 (entry-mechanics 2026-08-30): the split entry writes TWO rows per
		// (plan_id, scenario) distinguished by leg_index — the legacy 2-column
		// unique index would reject the second leg. Replace it with the
		// 3-column form (idempotent: DROP IF EXISTS + CREATE IF NOT EXISTS).
		if err := s.db.Exec("DROP INDEX IF EXISTS idx_armed_orders_plan_scenario").Error; err != nil {
			return err
		}
		return s.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_armed_orders_plan_scenario ON armed_orders(plan_id, scenario, leg_index)").Error
	}
	return s.db.AutoMigrate(&ArmedOrderDB{})
}

// UpsertArm writes/refreshes the arm row for (plan_id, scenario, leg_index).
// Same key = same row (state reset to armed only when the spec CHANGED
// materially — entry/stop/target diff >= 2 ticks — the caller decides and
// passes reset). E4: leg rows of a split arm share (plan_id, scenario) and
// differ by LegIndex.
func (s *ArmedOrderStore) UpsertArm(row *ArmedOrderDB) error {
	if row == nil || row.PlanID == "" || row.Scenario == "" {
		return fmt.Errorf("plan_id and scenario required")
	}
	// CANONICAL CASING AT THE WRITE (class 28, owner ruling 2026-09-03). This
	// table stored lowercase (long 19 / short 17) while trader_positions stores
	// uppercase (LONG 280 / SHORT 304), and the fill handler's side-keyed
	// lookup compared them literally — so it could never match. UPPER() on the
	// read makes existing rows work; canonicalizing HERE, where the value
	// enters, is what stops the two tables disagreeing at all.
	row.Side = strings.ToUpper(strings.TrimSpace(row.Side))
	// PRE-REOPEN F3 (2026-08-28) — dead re-arm fix: a TERMINAL row for the same
	// (plan, scenario) is re-authorized as a fresh armed row (new identity, no
	// stale fill); a non-terminal row keeps its identity and only its prices
	// are refreshed. The old Assign-based upsert left terminal rows terminal
	// forever, so a legit same-scenario re-arm was impossible and the executor
	// re-logged the dead row every cycle.
	var existing ArmedOrderDB
	err := s.db.Where("plan_id = ? AND scenario = ? AND leg_index = ?", row.PlanID, row.Scenario, row.LegIndex).First(&existing).Error
	if err == nil {
		if existing.State == "armed" || existing.State == "working" {
			row.ID = existing.ID
			// ATTRIBUTION (2026-09-02): armed_under_version is NOT in this map —
			// it belongs to the first authorization and must survive every
			// re-authorization. "version" continues to mean last-touch.
			row.ArmedUnderVersion = existing.ArmedUnderVersion
			if row.ArmedUnderVersion == 0 {
				// A row armed before the column existed: adopt its current
				// version once, so the backfill is self-healing rather than a
				// migration that has to guess.
				row.ArmedUnderVersion = existing.Version
				if err := s.db.Model(&existing).Update("armed_under_version", row.ArmedUnderVersion).Error; err != nil {
					return err
				}
			}
			return s.db.Model(&existing).Updates(map[string]any{
				"version": row.Version, "session": row.Session,
				"side": row.Side, "entry_px": row.EntryPx, "stop_px": row.StopPx,
				"target_px": row.TargetPx, "updated_at": row.UpdatedAt,
				"leg_count": row.LegCount, "kind": row.Kind,
			}).Error
		}
		// MANUAL-CANCEL-WINS (2026-08-30 E7 incident): a TERMINAL row is
		// re-authorized ONLY on a plan VERSION change. The old
		// re-authorize-every-cycle behavior was the re-place loop:
		// terminal → armed → marketable fill → stop-out → terminal → armed…
		// forever while the confirm stayed MET, so an owner/NT8 cancel
		// never won. Same version + terminal = the row STAYS terminal.
		// 0B (owner ruling 2026-09-02) — RE-ARM AFTER BOOT SWEEP. The
		// manual-cancel-wins law exists so the OWNER's cancels stick. A boot
		// sweep is the machine's own housekeeping: it cancels pre-boot orders
		// because the process that owned them died, not because anyone judged
		// the setup dead. Leaving those rows sticky killed the live setup
		// until the next plan version — on 09-02 00:16 that rule would have
		// meant no position 587. Swept rows (state_reason prefixed
		// "boot_sweep") re-authorize under the SAME version; every other
		// terminal row stays terminal.
		if existing.Version == row.Version && !IsBootSweepReason(existing.StateReason) {
			return nil
		}
		// 0B — the re-arm is LOUD: a swept row coming back under the SAME version
		// is the machine undoing its own boot housekeeping, and the journal must
		// say so with the dead broker identity it replaces.
		if existing.Version == row.Version && IsBootSweepReason(existing.StateReason) {
			logger.Warnf("⚖ re-armed after boot sweep: %s %s leg %d — signal %s → (fresh arm, awaiting placement) · same plan version v%d",
				row.Session, row.Scenario, row.LegIndex+1, signalOrNone(existing.SignalID), row.Version)
		}
		// New plan version → RE-AUTHORIZE: fresh armed state, fresh lineage.
		row.ID = existing.ID
		return s.db.Model(&existing).Updates(map[string]any{
			"state": "armed", "state_reason": "", "signal_id": "",
			"entry_class": "", "fill_price": 0, "fill_quantity": 0,
			"trader_id": row.TraderID, "version": row.Version, "session": row.Session,
			// ATTRIBUTION: a terminal row re-authorized under a NEW plan version
			// is a NEW arm reusing the row id — its first authorization is now.
			"armed_under_version": row.Version,
			"side":                row.Side, "entry_px": row.EntryPx, "stop_px": row.StopPx,
			"target_px": row.TargetPx, "leg_count": row.LegCount, "kind": row.Kind,
			"created_at": row.CreatedAt, "updated_at": row.UpdatedAt,
			// class 33: a re-authorized row belongs to THIS process.
			"boot_id": ProcessBootID(),
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	// class 33 — a freshly created row belongs to THIS process, so the boot
	// sweep never mistakes it for an orphan of a dead one.
	if row.BootID == "" {
		row.BootID = ProcessBootID()
	}
	// ATTRIBUTION: first authorization stamps the version the arm belongs to.
	if row.ArmedUnderVersion == 0 {
		row.ArmedUnderVersion = row.Version
	}
	return s.db.Create(row).Error
}

// ListNonTerminal returns ONE TRADER's armed orders that are NOT in a
// terminal state. PRE-SUNDAY F4 (2026-08-28): the old unscoped scan crossed
// trader boundaries the moment more than one trader runs.
func (s *ArmedOrderStore) ListNonTerminal(traderID string) ([]ArmedOrderDB, error) {
	var out []ArmedOrderDB
	err := s.db.Where("trader_id = ? AND state IN ('armed','working')", traderID).Order("id").Find(&out).Error
	return out, err
}

// SetState transitions one row's state with a reason (the ledger rule: a
// terminal state change is never silent).
func (s *ArmedOrderStore) SetState(id int64, state, reason string) error {
	return s.db.Model(&ArmedOrderDB{}).Where("id = ?", id).
		Updates(map[string]any{"state": state, "state_reason": reason}).Error
}

// SetFillPrice records the actual fill price on a FILLED row (PRE-SUNDAY F2 —
// the lineage matcher keys on this; entry_px drifts on re-arm and is NOT the
// fill).
func (s *ArmedOrderStore) SetFillPrice(id int64, fillPrice float64) error {
	return s.db.Model(&ArmedOrderDB{}).Where("id = ?", id).
		Update("fill_price", fillPrice).Error
}

// SetSignal records the wire signal_id once the resting limit is placed
// (armed → working transition).
func (s *ArmedOrderStore) SetSignal(id int64, signalID string) error {
	return s.db.Model(&ArmedOrderDB{}).Where("id = ?", id).
		Update("signal_id", signalID).Error
}

// SetFillQuantity stamps the contracts a fill actually delivered
// (invalidation-wired, 2026-09-03).
//
// armed row 35 (2026-09-03, NY v2 S1) read state=filled with fill_quantity=0
// while trader_positions carried quantity 1. Nothing wrote the column, so the
// ledger could say a row filled and not how much — and 0 is also a legal
// "nothing filled", so the row could not be read either way. WHERE-scoped and
// idempotent; a zero quantity is never written over a real one.
func (s *ArmedOrderStore) SetFillQuantity(id int64, qty int) error {
	if s == nil || s.db == nil || id == 0 || qty <= 0 {
		return nil
	}
	return s.db.Model(&ArmedOrderDB{}).Where("id = ?", id).
		Update("fill_quantity", qty).Error
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

// BootSweepReasonPrefix marks a ledger row cancelled by the class-33 boot
// sweep (the machine's own housekeeping), as opposed to an owner/NT8 cancel.
const BootSweepReasonPrefix = "boot_sweep"

// IsBootSweepReason reports whether a terminal row was swept at boot — the ONE
// terminal class that re-authorizes under the same plan version (0B).
func IsBootSweepReason(reason string) bool {
	return strings.HasPrefix(strings.TrimSpace(reason), BootSweepReasonPrefix)
}

// signalOrNone renders a possibly-empty broker signal id for the re-arm line.
func signalOrNone(id string) string {
	if strings.TrimSpace(id) == "" {
		return "(never placed)"
	}
	return id
}
