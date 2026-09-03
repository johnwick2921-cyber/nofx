package store

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ── PLANNER READ FACTS (VOID PARITY wave, 2026-09-02) ────────────────────────
//
// Until now a rendered prompt was persisted ONLY in planner_rejected_prompts,
// i.e. only when a read FAILED. Class 45's VOID / stop-floor sections could
// therefore be proven live only because a read happened to be rejected —
// "does the fix work" and "is the fix present" were the same query, and if the
// fix worked perfectly the evidence disappeared.
//
// This table records what the model was TOLD on EVERY read, accepted or
// rejected: the void list, the stop floor and its ATR, the bias labels, the
// prompt hash and size. It is deliberately small (no prompt text — that stays
// in the rejected store) so the cap can be generous.
type PlannerReadFact struct {
	ID         uint   `gorm:"primaryKey"`
	TraderID   string `gorm:"index"`
	TradeDate  string `gorm:"index"`
	Session    string `gorm:"index"`
	PlanID     string `gorm:"index"` // "" when the read has not yet written a plan
	Version    int    `gorm:"index"` // 0 = unknown at render time
	PromptHash string `gorm:"index"`
	// VoidLevels is the rendered VOID list as JSON ([] = computed and empty,
	// which is a REAL answer; NULL/"" = not computed). The distinction matters:
	// an absent list and an empty list mean different things.
	VoidLevels   string  `gorm:"type:text"`
	VoidCount    int     `gorm:"index"`
	StopFloorPts float64 // 0 = no ATR this cycle → the prompt rendered no floor
	ATR5m        float64
	StopFloorMlt float64
	BiasAI       string
	BiasTree     string
	BiasRegime   string
	TokensIn     int
	// Scope is the resolved void scope both prompt and validator read, recorded
	// so a later reader can tell WHICH tape produced this list.
	ScopeSinceMs int64
	ScopeBars    int
	ScopeIntv    string
	CreatedAt    time.Time `gorm:"index"`
}

// TableName is explicit so the cap-trim SQL never guesses.
func (PlannerReadFact) TableName() string { return "planner_read_facts" }

// PlannerReadFactsStore persists one row per planner read.
type PlannerReadFactsStore struct {
	db *gorm.DB
}

// NewPlannerReadFactsStore wires the table via AutoMigrate.
func NewPlannerReadFactsStore(db *gorm.DB) *PlannerReadFactsStore {
	if db != nil {
		_ = db.AutoMigrate(&PlannerReadFact{})
	}
	return &PlannerReadFactsStore{db: db}
}

// PlannerReadFactsCap bounds the table. Rows are small (a few hundred bytes),
// so 500 spans several days of reads at the observed ~30-minute cadence.
const PlannerReadFactsCap = 500

// VoidLevelRecord is one entry of the rendered VOID list, stored verbatim so a
// later reader sees exactly what the model saw.
type VoidLevelRecord struct {
	Price       float64 `json:"price"`
	Short       bool    `json:"short"`
	ReclaimedAt string  `json:"reclaimed_at,omitempty"`
}

// SaveReadFact writes one row and trims to the newest PlannerReadFactsCap.
// Never returns an error to the caller's critical path on a nil store — a
// missing telemetry table must not fail a planner read (A10).
func (s *PlannerReadFactsStore) SaveReadFact(f *PlannerReadFact) error {
	if s == nil || s.db == nil || f == nil {
		return nil
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}
	if err := s.db.Create(f).Error; err != nil {
		return err
	}
	var count int64
	if err := s.db.Model(&PlannerReadFact{}).Count(&count).Error; err == nil && count > PlannerReadFactsCap {
		_ = s.db.Exec("DELETE FROM planner_read_facts WHERE id NOT IN (SELECT id FROM planner_read_facts ORDER BY id DESC LIMIT ?)", PlannerReadFactsCap).Error
	}
	return nil
}

// EncodeVoidLevels renders the list to JSON. A nil/empty list encodes as "[]" —
// computed AND empty, which is a real answer and must not read as "not
// computed" (that is the empty string).
func EncodeVoidLevels(v []VoidLevelRecord) string {
	if v == nil {
		return "[]"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// LatestReadFact returns the newest row (nil, gorm.ErrRecordNotFound when empty).
func (s *PlannerReadFactsStore) LatestReadFact() (*PlannerReadFact, error) {
	if s == nil || s.db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var row PlannerReadFact
	if err := s.db.Order("id DESC").First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// ReadFactsCount is the row count (for the cap fixture and the boot line).
func (s *PlannerReadFactsStore) ReadFactsCount() int64 {
	if s == nil || s.db == nil {
		return 0
	}
	var n int64
	_ = s.db.Model(&PlannerReadFact{}).Count(&n).Error
	return n
}
