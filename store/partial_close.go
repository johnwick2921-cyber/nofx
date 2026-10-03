package store

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ── PARTIAL CLOSE (2026-10-03, mentor mode, behind the #309 knob) ───────────
//
// Two tables. position_reductions is the per-partial ledger: who asked, how
// many, at what fill, with what remaining. stop_resizes is the protective-stop
// resize lifecycle for the SAME #309 report regime: a leg cancel is DONE only
// on the AddOn's positive terminal report for the leg's order id; until then
// the resize is pending and the old stop is assumed live. A resize that cannot
// be confirmed FAILS CLOSED and the caller flattens the remainder — never a
// blind re-place.

// PositionReduction is ONE partial exit. FillPrice=0 and Remaining=-1 while
// the fill has not been reported (absent ≠ 0 law: an unreported fill is NOT a
// zero fill).
type PositionReduction struct {
	ID        int64   `gorm:"primaryKey;autoIncrement"`
	TraderID  string  `gorm:"index"`
	Symbol    string  `gorm:"index"`
	Side      string  `json:"side"` // "long" | "short"
	ClientID  string  `gorm:"index"`
	Quantity  int     `json:"quantity"`
	FillPrice float64 `json:"fill_price"`
	Remaining int     `json:"remaining"`
	Who       string  `json:"who"`
	CreatedMs int64   `json:"created_ms"`
	UpdatedMs int64   `json:"updated_ms"`
}

// StopResizeState is one protective-stop resize's state machine.
const (
	StopResizePending   = "pending"   // leg cancel requested, report not yet recorded
	StopResizeConfirmed = "confirmed" // the leg's terminal report is recorded; the replacement may place
	StopResizeDone      = "done"      // the replacement stop was placed at the new quantity
	StopResizeFailed    = "failed"    // the report never came — the caller flattened the remainder
)

// StopResize is one protective-stop resize: cancel the leg named LegSignalID
// (the full "<signal>-sl" name) and re-place at Quantity when the cancel is
// CONFIRMED by the AddOn's report. Same column semantics as armed_orders'
// cancel-report columns (request time + report time/state, F9 predates check).
type StopResize struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	TraderID     string `gorm:"index"`
	Symbol       string
	Side         string  // "long" | "short" — the position side the stop protects
	LegSignalID  string  `gorm:"index"` // the leg's full order name, e.g. "<signal>-sl"
	Quantity     int     // the stop quantity to re-place at (= remaining position)
	StopPrice    float64 // the OLD stop's price, captured before the cancel
	State        string  `gorm:"index"`
	RequestMs    int64
	ReportMs     int64  `gorm:"default:0"`
	ReportState  string `gorm:"default:''"`
	Attempts     int    `gorm:"default:0"`
	AttemptsBoot string `gorm:"default:''"`
	CreatedMs    int64
	UpdatedMs    int64
}

type partialCloseStore struct {
	db *gorm.DB
}

// Migrate creates both tables. sqlite gets the exact DDL (no guessing
// migration), every other dialect gets AutoMigrate.
func (s *partialCloseStore) Migrate() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if s.db.Dialector.Name() == "sqlite" {
		if err := s.db.Exec(`
			CREATE TABLE IF NOT EXISTS position_reductions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				trader_id TEXT NOT NULL DEFAULT '',
				symbol TEXT NOT NULL DEFAULT '',
				side TEXT NOT NULL DEFAULT '',
				client_id TEXT NOT NULL DEFAULT '',
				quantity INTEGER NOT NULL DEFAULT 0,
				fill_price REAL NOT NULL DEFAULT 0,
				remaining INTEGER NOT NULL DEFAULT -1,
				who TEXT NOT NULL DEFAULT '',
				created_ms INTEGER NOT NULL DEFAULT 0,
				updated_ms INTEGER NOT NULL DEFAULT 0
			);
			CREATE INDEX IF NOT EXISTS idx_position_reductions_trader ON position_reductions(trader_id);
			CREATE INDEX IF NOT EXISTS idx_position_reductions_client ON position_reductions(client_id);
		`).Error; err != nil {
			return err
		}
		if err := s.db.Exec(`
			CREATE TABLE IF NOT EXISTS stop_resizes (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				trader_id TEXT NOT NULL DEFAULT '',
				symbol TEXT NOT NULL DEFAULT '',
				side TEXT NOT NULL DEFAULT '',
				leg_signal_id TEXT NOT NULL DEFAULT '',
				quantity INTEGER NOT NULL DEFAULT 0,
				stop_price REAL NOT NULL DEFAULT 0,
				state TEXT NOT NULL DEFAULT '',
				request_ms INTEGER NOT NULL DEFAULT 0,
				report_ms INTEGER NOT NULL DEFAULT 0,
				report_state TEXT NOT NULL DEFAULT '',
				attempts INTEGER NOT NULL DEFAULT 0,
				attempts_boot TEXT NOT NULL DEFAULT '',
				created_ms INTEGER NOT NULL DEFAULT 0,
				updated_ms INTEGER NOT NULL DEFAULT 0
			);
			CREATE INDEX IF NOT EXISTS idx_stop_resizes_state ON stop_resizes(state);
			CREATE INDEX IF NOT EXISTS idx_stop_resizes_leg ON stop_resizes(leg_signal_id);
		`).Error; err != nil {
			return err
		}
		return nil
	}
	return s.db.AutoMigrate(&PositionReduction{}, &StopResize{})
}

// RecordReduce inserts the request-time row. The fill update (FillPrice,
// Remaining) upserts by client_id latest-wins so a part-fill then full-fill
// pair applies exactly once per progress step.
func (s *partialCloseStore) RecordReduce(r *PositionReduction) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if strings.TrimSpace(r.ClientID) == "" || r.Quantity <= 0 {
		return fmt.Errorf("a reduce needs a client id and a positive quantity")
	}
	nowMs := time.Now().UTC().UnixMilli()
	if r.CreatedMs <= 0 {
		r.CreatedMs = nowMs
	}
	r.UpdatedMs = nowMs
	return s.db.Create(r).Error
}

// ApplyReduceFill records the fill side of a partial: fill price + remaining,
// upserted by client_id latest-wins (idempotent under C# re-reports).
func (s *partialCloseStore) ApplyReduceFill(clientID string, fillPrice float64, remaining int) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if strings.TrimSpace(clientID) == "" {
		return fmt.Errorf("a reduce fill needs its client id")
	}
	res := s.db.Model(&PositionReduction{}).Where("client_id = ?", clientID).Updates(map[string]any{
		"fill_price": fillPrice,
		"remaining":  remaining,
		"updated_ms": time.Now().UTC().UnixMilli(),
	})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// ListReductions returns one trader's partials, newest first.
func (s *partialCloseStore) ListReductions(traderID string) ([]PositionReduction, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	var out []PositionReduction
	err := s.db.Where("trader_id = ?", traderID).Order("id DESC").Find(&out).Error
	return out, err
}

// RecordStopResizeRequest opens a pending resize: the leg is cancelled, and
// the replacement quantity is the remaining position.
func (s *partialCloseStore) RecordStopResizeRequest(r *StopResize) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if strings.TrimSpace(r.LegSignalID) == "" || r.Quantity <= 0 {
		return fmt.Errorf("a stop resize needs the leg order id and a positive quantity")
	}
	nowMs := time.Now().UTC().UnixMilli()
	r.State = StopResizePending
	if r.RequestMs <= 0 {
		r.RequestMs = nowMs
	}
	r.CreatedMs = nowMs
	r.UpdatedMs = nowMs
	return s.db.Create(r).Error
}

// RecordStopResizeReport records the AddOn's positive terminal report for the
// leg's order id — the SAME evidence rule as armed_orders.RecordCancelReport
// (monotonic latest-wins; the C# echo and the real transition both report).
func (s *partialCloseStore) RecordStopResizeReport(id int64, reportMs int64, state string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	if reportMs <= 0 || strings.TrimSpace(state) == "" {
		return fmt.Errorf("a stop-resize report requires its receipt time and state")
	}
	res := s.db.Model(&StopResize{}).
		Where("id = ? AND (report_ms = 0 OR report_ms < ?)", id, reportMs).
		Updates(map[string]any{
			"report_ms":    reportMs,
			"report_state": state,
			"updated_ms":   time.Now().UTC().UnixMilli(),
		})
	return res.Error
}

// ConfirmStopResizeByReport promotes a pending resize to confirmed — ONLY on
// the recorded report: terminal ('cancelled'), postdating the request (F9).
// A 'filled' report never confirms a cancel here (the leg filled is a
// different path). Returns the confirmed row's replacement quantity.
func (s *partialCloseStore) ConfirmStopResizeByReport(id int64) (*StopResize, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("store required")
	}
	var out *StopResize
	err := immediateOrPlainTxAny(s.db, func(tx *gorm.DB) error {
		var row StopResize
		if err := tx.First(&row, id).Error; err != nil {
			return err
		}
		if row.State != StopResizePending {
			return fmt.Errorf("stop resize %d is %s, not pending", id, row.State)
		}
		if row.ReportMs <= 0 ||
			strings.ToLower(strings.TrimSpace(row.ReportState)) != "cancelled" {
			return fmt.Errorf("stop resize %d has no qualifying terminal report (report_ms=%d state=%q)", id, row.ReportMs, row.ReportState)
		}
		if row.RequestMs <= 0 || row.ReportMs < row.RequestMs {
			return fmt.Errorf("stop resize %d report predates the request (report %d < request %d)", id, row.ReportMs, row.RequestMs)
		}
		res := tx.Model(&StopResize{}).Where("id = ? AND state = ?", id, StopResizePending).
			Updates(map[string]any{"state": StopResizeConfirmed, "updated_ms": time.Now().UTC().UnixMilli()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("stop resize %d changed during confirmation", id)
		}
		out = &row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MarkStopResizeDone records the replacement stop as placed.
func (s *partialCloseStore) MarkStopResizeDone(id int64) error {
	return s.setStopResizeState(id, StopResizeDone)
}

// MarkStopResizeFailed records the fail-closed outcome (the remainder was
// flattened because the leg cancel never confirmed).
func (s *partialCloseStore) MarkStopResizeFailed(id int64) error {
	return s.setStopResizeState(id, StopResizeFailed)
}

func (s *partialCloseStore) setStopResizeState(id int64, state string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store required")
	}
	return s.db.Model(&StopResize{}).Where("id = ?", id).
		Updates(map[string]any{"state": state, "updated_ms": time.Now().UTC().UnixMilli()}).Error
}

// ListStopResizePending returns this trader's resize rows awaiting the leg
// cancel confirmation, oldest request first.
func (s *partialCloseStore) ListStopResizePending(traderID string) ([]StopResize, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	var out []StopResize
	err := s.db.Where("trader_id = ? AND state = ?", traderID, StopResizePending).
		Order("request_ms ASC, id ASC").Find(&out).Error
	return out, err
}
