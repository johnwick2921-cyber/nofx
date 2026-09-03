package store

import (
	"path/filepath"
	"testing"
)

func excStore(t *testing.T) *TradeExcursionStore {
	t.Helper()
	st, err := New(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st.TradeExcursions()
}

// F4 — NULL ≠ 0. This is the whole reason the wave exists: on trader_positions
// a computed zero and a value that was never computed are the same bit pattern
// (D15; 521 of 586 closed rows read mae=0). A fresh excursion row must carry
// NULLs, and a genuine zero must survive as 0.
func TestExcursionFreshRowIsNullNotZero(t *testing.T) {
	es := excStore(t)
	id, err := es.Open(TradeExcursion{
		PositionID: 900, PlanID: "p", Version: 3, Session: "NY", Scenario: "S1",
		Condition: "reject", Side: "LONG", EntryPx: 100, EntryTs: 1_000_000,
		StopPxInitial: 90, TargetPx: 120, Size: 1, ATR5mAtEntry: 8,
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	row, err := es.Get(id)
	if err != nil || row == nil {
		t.Fatalf("get: %v", err)
	}
	for name, p := range map[string]*float64{"mae_pts": row.MAEPts, "mfe_pts": row.MFEPts, "pnl_corrected": row.PnlCorrected} {
		if p != nil {
			t.Errorf("%s = %v on a fresh row — it must be NULL until computed", name, *p)
		}
	}
	for name, p := range map[string]*int64{"mae_ts": row.MAETs, "mfe_ts": row.MFETs, "exit_ts": row.ExitTs} {
		if p != nil {
			t.Errorf("%s = %v on a fresh row — it must be NULL until computed", name, *p)
		}
	}
	if row.BarsHeld != nil || row.AmbiguousBars != nil {
		t.Error("bars_held / ambiguous_bars must be NULL before any bar is seen")
	}
	if row.Resolution != "" {
		t.Errorf("resolution = %q on a fresh row, want empty until a tape is chosen", row.Resolution)
	}

	// A genuine zero survives as zero, distinct from the NULL above.
	if err := es.UpdatePath(id, TradeExcursionPath{
		MAEPts: 0, MAETs: 1_000_000, MAEBars: 0,
		MFEPts: 12.5, MFETs: 1_060_000, MFEBars: 1,
		BarsHeld: 2, AmbiguousBars: 0, Resolution: "1m",
	}); err != nil {
		t.Fatalf("update path: %v", err)
	}
	row, _ = es.Get(id)
	if row.MAEPts == nil {
		t.Fatal("a COMPUTED zero must be stored as 0, not left NULL")
	}
	if *row.MAEPts != 0 {
		t.Errorf("mae_pts = %v, want a stored 0", *row.MAEPts)
	}
	if row.AmbiguousBars == nil || *row.AmbiguousBars != 0 {
		t.Error("a computed zero ambiguity is 0, not NULL")
	}
}

// A22 — the close copies pnl_corrected, never raw realized_pnl.
func TestExcursionCloseCopiesCorrectedPnL(t *testing.T) {
	es := excStore(t)
	id, _ := es.Open(TradeExcursion{PositionID: 901, Side: "SHORT", EntryPx: 200, EntryTs: 5, Size: 1})
	corrected := -155.0
	if err := es.Close(id, TradeExcursionClose{
		ExitPx: 190, ExitTs: 99, ExitReason: "stop", StopPxFinal: 210,
		PnlCorrected: &corrected, Ambiguous: true,
	}); err != nil {
		t.Fatalf("close: %v", err)
	}
	row, _ := es.Get(id)
	if row.PnlCorrected == nil || *row.PnlCorrected != -155 {
		t.Fatalf("pnl_corrected not copied: %+v", row.PnlCorrected)
	}
	if row.ExitTs == nil || *row.ExitTs != 99 || row.ExitReason != "stop" {
		t.Errorf("exit fields not stored: ts=%v reason=%q", row.ExitTs, row.ExitReason)
	}
	if !row.AmbiguousExit {
		t.Error("an ambiguous close must be flagged on the row")
	}

	// A close with no corrected P&L leaves the column NULL — it never falls
	// back to realized_pnl (A22).
	id2, _ := es.Open(TradeExcursion{PositionID: 902, Side: "LONG", EntryPx: 10, EntryTs: 5, Size: 1})
	if err := es.Close(id2, TradeExcursionClose{ExitPx: 11, ExitTs: 6, ExitReason: "target"}); err != nil {
		t.Fatalf("close 2: %v", err)
	}
	row2, _ := es.Get(id2)
	if row2.PnlCorrected != nil {
		t.Errorf("pnl_corrected = %v with none supplied — must stay NULL", *row2.PnlCorrected)
	}
}
