package store

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

// ── D2 — TOUCH OUTCOMES (1B, 2026-09-03) ─────────────────────────────────────
//
// One row per D1′ episode, written at episode CLOSE. This is the table that
// replaces two biased instruments: touch_telemetry's "rejection" (the close is
// still on its starting side — ≈0.69 on IID noise BY CONSTRUCTION) and
// level_stats_calc's "reacted" (any ≥reactPts move away, so a blast-through
// scored as a reaction). Every reaction rate ever published from those — 84%,
// 70.3%, 75.1% — is an artifact of the predicate.
//
// AMBIGUOUS ROWS ARE WRITTEN, FLAGGED AND EXCLUDED FROM THE RATE, never
// dropped: a rate that silently discards its hard cases lies about its base.
type TouchOutcomeRow struct {
	ID       uint   `gorm:"primaryKey"`
	TraderID string `gorm:"index"`
	Symbol   string `gorm:"index"`
	// The level this episode belongs to.
	LevelPrice float64 `gorm:"index"`
	LevelKind  string  `gorm:"index"`
	// Was the level SEATED in the plan, or only a candidate? The selection
	// question is off-policy and needs the excluded pool too (B2/B3).
	CandidateSeated bool   `gorm:"index"`
	PlanID          string `gorm:"index"`
	PlanVersion     int
	Session         string `gorm:"index"`
	// Ordinal comes from the STORE, never an in-memory counter (C4): touch
	// numbering must survive a restart.
	Ordinal int `gorm:"index"`
	// The resolved detector scope this episode was judged under, recorded so a
	// later reader knows which instrument produced the verdict.
	K       float64
	Delta   float64
	BandPts float64
	Horizon int
	ExitOn  string
	// The episode itself.
	EntrySide  string // "below" | "above" — the side price came FROM
	ExitSide   string // "below" | "above" | "" when ambiguous
	Outcome    string `gorm:"index"` // hold | break | ambiguous_span | ambiguous_horizon
	Ambiguous  bool   `gorm:"index"`
	BarsToExit int
	MFEPts     float64
	MAEPts     float64
	OpenedAtMs int64 `gorm:"index"`
	ClosedAtMs int64
	CreatedAt  time.Time `gorm:"index"`
}

func (TouchOutcomeRow) TableName() string { return "touch_outcomes" }

// TouchOutcomeStore persists one row per closed episode.
type TouchOutcomeStore struct{ db *gorm.DB }

// NewTouchOutcomeStore wires the table via AutoMigrate.
func NewTouchOutcomeStore(db *gorm.DB) *TouchOutcomeStore {
	if db != nil {
		_ = db.AutoMigrate(&TouchOutcomeRow{})
	}
	return &TouchOutcomeStore{db: db}
}

// NextOrdinal reads the next touch ordinal for a level FROM THE STORE (C4), so
// numbering survives a restart. Scoped per (trader, symbol, level, session-day).
func (s *TouchOutcomeStore) NextOrdinal(traderID, symbol string, level float64, sessionDayMs int64) int {
	if s == nil || s.db == nil {
		return 1
	}
	var maxOrd int
	if err := s.db.Model(&TouchOutcomeRow{}).
		Where("trader_id = ? AND symbol = ? AND level_price = ? AND opened_at_ms >= ?",
			traderID, symbol, level, sessionDayMs).
		Select("COALESCE(MAX(ordinal), 0)").Scan(&maxOrd).Error; err != nil {
		return 1
	}
	return maxOrd + 1
}

// SaveOutcome writes one episode. Telemetry may WARN, never panic (A10) — a
// failed write must not stop the loop.
func (s *TouchOutcomeStore) SaveOutcome(r *TouchOutcomeRow) error {
	if s == nil || s.db == nil || r == nil {
		return nil
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	return s.db.Create(r).Error
}

// CountOutcomes is the boot line's figure — READ, never a literal.
func (s *TouchOutcomeStore) CountOutcomes() int64 {
	if s == nil || s.db == nil {
		return 0
	}
	var n int64
	_ = s.db.Model(&TouchOutcomeRow{}).Count(&n).Error
	return n
}

// HoldRateBy returns hold/(hold+break) with n and the excluded ambiguous count
// for a filtered slice of the table. group is a column name ("level_kind",
// "session", "ordinal") or "" for the whole table.
type OutcomeRate struct {
	Group     string
	Hold      int
	Break     int
	Ambiguous int
}

// N is the rate's base — hold+break, ambiguous EXCLUDED.
func (r OutcomeRate) N() int { return r.Hold + r.Break }

// P is hold/(hold+break); 0 at n=0, and callers must check N before quoting it.
func (r OutcomeRate) P() float64 {
	if r.N() == 0 {
		return 0
	}
	return float64(r.Hold) / float64(r.N())
}

// RatesBy groups the table and returns hold/break/ambiguous per group.
func (s *TouchOutcomeStore) RatesBy(column string) ([]OutcomeRate, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	sel := "'' AS grp"
	grp := ""
	if column != "" {
		sel = column + " AS grp"
		grp = column
	}
	type row struct {
		Grp     string
		Outcome string
		N       int
	}
	var rows []row
	q := s.db.Model(&TouchOutcomeRow{}).Select(sel + ", outcome, COUNT(*) AS n")
	if grp != "" {
		q = q.Group(grp + ", outcome").Order(grp)
	} else {
		q = q.Group("outcome")
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	byGroup := map[string]*OutcomeRate{}
	var order []string
	for _, r := range rows {
		g, ok := byGroup[r.Grp]
		if !ok {
			g = &OutcomeRate{Group: r.Grp}
			byGroup[r.Grp] = g
			order = append(order, r.Grp)
		}
		switch r.Outcome {
		case "hold":
			g.Hold += r.N
		case "break":
			g.Break += r.N
		default:
			g.Ambiguous += r.N
		}
	}
	out := make([]OutcomeRate, 0, len(order))
	for _, g := range order {
		out = append(out, *byGroup[g])
	}
	return out, nil
}

// ── D6 — THE READ-ONLY REPORT ────────────────────────────────────────────────
//
// p(hold) per kind / session / ordinal, each with n, a Wilson interval and the
// ambiguous share. Below the floor every line says DESCRIPTIVE ONLY: at n<200 a
// rate is a description of a sample, not an estimate of a property.
const TouchRateFloor = 200

// DetectorReport renders the D6 table. Empty table → says so, never a zero rate.
func (s *TouchOutcomeStore) DetectorReport() string {
	if s == nil || s.db == nil {
		return "detector report: store unavailable"
	}
	var b []byte
	add := func(f string, a ...any) { b = append(b, []byte(sprintf(f, a...))...) }
	total := s.CountOutcomes()
	add("touch_outcomes: %d row(s)\n", total)
	if total == 0 {
		add("  (empty — the detector has recorded no episodes yet; every figure below would be a plausible zero)\n")
		return string(b)
	}
	for _, dim := range []struct{ label, col string }{
		{"ALL", ""}, {"by kind", "level_kind"}, {"by session", "session"}, {"by ordinal", "ordinal"},
	} {
		rates, err := s.RatesBy(dim.col)
		if err != nil {
			add("  %s: unavailable (%v)\n", dim.label, err)
			continue
		}
		add("  %s:\n", dim.label)
		for _, r := range rates {
			g := r.Group
			if g == "" {
				g = "(all)"
			}
			if r.N() == 0 {
				add("    %-14s n=0 — no resolved episodes (ambiguous=%d)\n", g, r.Ambiguous)
				continue
			}
			lo, hi := wilson(r.P(), r.N())
			note := ""
			if r.N() < TouchRateFloor {
				note = sprintf("  — n<%d, DESCRIPTIVE ONLY", TouchRateFloor)
			}
			ambShare := float64(r.Ambiguous) / float64(r.N()+r.Ambiguous) * 100
			add("    %-14s p(hold)=%.3f [%.3f, %.3f] n=%d · ambiguous=%d (%.1f%%)%s\n",
				g, r.P(), lo, hi, r.N(), r.Ambiguous, ambShare, note)
		}
	}
	return string(b)
}

// TouchOutcomesBootLine reports the recorded row count — READ, never a literal.
func (s *TouchOutcomeStore) TouchOutcomesBootLine() int64 { return s.CountOutcomes() }

func sprintf(f string, a ...any) string { return fmt.Sprintf(f, a...) }

// wilson is the 95% score interval — the same formula the detector uses, kept
// here so the report cannot quote a bare proportion.
func wilson(p float64, n int) (lo, hi float64) {
	if n <= 0 {
		return 0, 0
	}
	const z = 1.959963984540054
	nf := float64(n)
	den := 1 + z*z/nf
	c := (p + z*z/(2*nf)) / den
	h := z * math.Sqrt(p*(1-p)/nf+z*z/(4*nf*nf)) / den
	return c - h, c + h
}
