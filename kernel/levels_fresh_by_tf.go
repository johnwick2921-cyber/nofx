package kernel

import (
	"strings"
	"time"

	"nofx/market"
)

// S2 — TIMEFRAME-AWARE FRESHNESS (2026-09-16, structure-first planner wave).
//
// Today (W7/W11b) a level's freshness is the persisted 1m-touch ladder
// (A→B→C→done, store.LevelState). A 4h zone is "tested" when a single 1m bar
// trades into its band, which is why fresh HTF zones collapse to grade C after
// one touch and the 12-seat table stops seating them (CTO WHY block, 2026-09-16
// 16:31 CT: 244 HTF levels detected, 3 seated).
//
// This file grades the freshness of an HTF level on ITS OWN timeframe bars:
// a bar of the level's TF that traded into [Lo,Hi] after the level's origin
// date is one test. Grade by test count: 0 fresh · 1 tested-1 · 2 tested-2 ·
// >=3 stale. The knob is `day_plan.levels_fresh_by_tf` (default OFF) — with it
// off this file is never on the grading path and every score/prompt stays
// byte-identical (proved by the parity + golden tests).

// htfFreshTFSet is the set of timeframes graded on their own bars when the S2
// knob is ON. Outside this set a level keeps today's 1m-touch grading.
var htfFreshTFSet = map[string]bool{
	"1h": true, "2h": true, "4h": true, "6h": true, "8h": true, "12h": true,
	"1d": true, "3d": true, "1w": true,
}

// IsHTFFreshTF reports whether tf is one of the S2 own-timeframe grading TFs.
func IsHTFFreshTF(tf string) bool {
	return htfFreshTFSet[strings.ToLower(strings.TrimSpace(tf))]
}

// LevelFreshnessByTF grades an HTF level's freshness on its own timeframe
// bars. Caller routes: only levels with DetectedLevel.HTF==true and a TF in
// htfFreshTFSet may enter. Bars MUST be that same timeframe. A bar "tests" the
// level when it traded into [Lo,Hi] (bar.Low <= Hi && bar.High >= Lo) at or
// after the level's origin. Origin resolution: OriginDate (YYYY-MM-DD), else
// FormedAtMs; unknown origin → 0 tests (fresh) — the level's own bars cannot
// contradict an unrecorded origin.
//
// Returns the S2 display grade ("fresh" | "tested-1" | "tested-2" | "stale")
// and the test count. The display grade is what Research.Freshness carries;
// scoring maps it onto the unchanged freshMult/zoneFreshMult ladders via
// normalizeByTFGrade (levels_score.go).
func LevelFreshnessByTF(l DetectedLevel, now time.Time, bars []market.Kline) (string, int) {
	origin, ok := levelOriginTime(l)
	if !ok || origin.IsZero() {
		return "fresh", 0
	}
	lo, hi := l.Lo, l.Hi
	if hi < lo {
		lo, hi = hi, lo
	}
	tests := 0
	for _, b := range bars {
		if b.OpenTime < origin.UnixMilli() {
			continue // before the level existed — cannot test it
		}
		if b.OpenTime > now.UnixMilli() {
			continue // future bar — not evidence
		}
		if b.Low <= hi && b.High >= lo {
			tests++
		}
	}
	switch tests {
	case 0:
		return "fresh", 0
	case 1:
		return "tested-1", 1
	case 2:
		return "tested-2", 2
	default:
		return "stale", tests
	}
}

// levelOriginTime resolves a level's origin instant: OriginDate first, then
// FormedAtMs. ok=false means no origin is recorded at all.
func levelOriginTime(l DetectedLevel) (time.Time, bool) {
	if d := strings.TrimSpace(l.OriginDate); d != "" {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			return t, true
		}
	}
	if l.FormedAtMs > 0 {
		return time.UnixMilli(l.FormedAtMs), true
	}
	return time.Time{}, false
}

// normalizeByTFGrade maps the S2 display vocabulary onto the canonical ladder
// strings the UNCHANGED freshMult/zoneFreshMult tables already understand
// (tested-1 → b · tested-2 → c · stale → done). Every pre-existing freshness
// string passes through unchanged, so with the knob OFF the multipliers are
// byte-identical to today.
func normalizeByTFGrade(f string) string {
	switch strings.ToLower(strings.TrimSpace(f)) {
	case "tested-1":
		return "b"
	case "tested-2":
		return "c"
	case "stale":
		return "done"
	default:
		return f
	}
}
