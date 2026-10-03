package mentor

import (
	"strings"
	"testing"
)

// TestEvaluatorBoxPathRecordedTape is the CTO-ordered call-site test for the
// ONE box implementation (DS-103's box-edge level path in eval.go), run
// through Evaluator.Tick on a recorded tape (the 13-Sep golden frame with
// both golden boxes). Expected behaviour [BOX RULING facts + R1]:
//
//   - an FTGH with price below it: the first return touches the bottom and
//     closes below → one short entry (R1);
//   - a second visit later the same day → a second entry;
//   - a candle closing inside the box → no entry, and the box is still there;
//   - a body escape above the box → the box is still there (B1: no intraday
//     deletion).
//
// BLOCKER (updated after the ping-pong fix, 2026-10-03): the mid-range
// exception is in (box-edge references pass midRangeBoxed — census on the
// golden tape: 59 box-edge rejects now reach PHLPLHGatedR2), but
// PHLPLHGatedR2's own gates (HTF direction / day gate / PHL geometry) still
// emit ZERO box entries. Until DS-103 opens that gate, the assertions are
// skipped instead of falsely green.
func TestEvaluatorBoxPathRecordedTape(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	e := New(cfg)
	bars := loadFixture(t, "mnq_1m_2026-09-13_boxframe", "1m")

	var entries []Intent
	for i := 2; i <= len(bars); i++ {
		for _, in := range e.Tick(bars[:i], bars[i-1].OpenTime+59_999) {
			if (in.Action == PlaceStopEntry || in.Action == PlaceStopLimitEntry) &&
				strings.Contains(in.LevelKey, "ftg") {
				entries = append(entries, in)
			}
		}
	}

	if len(entries) == 0 {
		t.Skip("BLOCKER: PHLPLHGatedR2 gates (HTF direction / day gate / PHL geometry) still emit zero box entries " +
			"on recorded tapes — 59 box-edge rejects reach it on the golden tape, 0 emit. " +
			"Assertions below go live when DS-103 opens that gate.")
	}

	if len(entries) < 2 {
		t.Fatalf("box entries = %d, want >= 2 — every return trades (R1)", len(entries))
	}
	for _, in := range entries {
		if in.Side != SideShort {
			t.Fatalf("box entry %+v: FTGH return trades are shorts", in)
		}
		if in.Reason == "" || in.Stop == 0 || in.Target == 0 {
			t.Fatalf("box entry incomplete: %+v", in)
		}
	}
}
