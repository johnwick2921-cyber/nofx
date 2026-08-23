package kernel

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// G6 (regime wave, 2026-08-21) — LOSS-STREAK PAUSE: N consecutive losing closes
// (pnl_corrected < 0) within one CME session-day pause NEW entries for
// LOSS_STREAK_PAUSE_MIN minutes (or until session end, whichever is first).
// MASTER-INDEPENDENT armor: the gate reads only the strategy config — it never
// consults the guardrail master toggle. Position management and the watcher are
// untouched. Studio: loss_streak_n (default 4; 0/absent → off — note: the
// absent=off convention matches the other armor knobs; the SHIPPED default of 4
// lives in the resolver, so an absent key means "the default", while an
// explicit 0 means OFF).

const (
	DefaultLossStreakN            = 4
	DefaultLossStreakPauseMinutes = 60
)

// LossStreakPauseMin resolves the pause window (env LOSS_STREAK_PAUSE_MIN,
// default 60 minutes).
func LossStreakPauseMin() int {
	if v := os.Getenv("LOSS_STREAK_PAUSE_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return DefaultLossStreakPauseMinutes
}

// LossStreakRefusal renders the gate message: "loss_streak: 4 consecutive
// losers — paused until 12:30 CT".
func LossStreakRefusal(n int, until time.Time) string {
	return fmt.Sprintf("loss_streak: %d consecutive losers — paused until %s", n, ClockCT(until))
}
