package trader

import (
	"fmt"
	"time"

	"nofx/kernel"
	"nofx/store"
)

// G6 (regime wave, 2026-08-21) — LOSS-STREAK PAUSE observer. Runs per cycle:
// N consecutive losing closes (pnl_corrected < 0) within one CME session-day
// pause new entries for LOSS_STREAK_PAUSE_MIN minutes (capped at the active
// session's end). Master-independent — the guardrail master toggle never
// enters this path. A winner resets the streak; the timer resumes entries.
// While paused: P0 dashboard banner (once per session-day) + WARN → log_events.

// observeLossStreak wires ctx for the executor gate.
func (at *AutoTrader) observeLossStreak(ctx *kernel.Context) {
	ctx.LossStreakPaused = false
	if at.store == nil {
		return
	}
	n := 0
	if sc := at.GetStrategyConfig(); sc != nil {
		n = sc.LossStreakNValue()
	}
	if n <= 0 {
		return // off
	}
	rows, err := at.store.Position().GetClosedPositions(at.id, n+1)
	if err != nil || len(rows) == 0 {
		return
	}
	now := time.Now()
	day := kernel.CMESessionDayStart(time.UnixMilli(rows[0].ExitTime))
	streak := 0
	var lastLoseAt int64
	for _, r := range rows {
		if r.ExitTime < day.UnixMilli() {
			break // older session — streak is session-scoped
		}
		if r.EffectivePnL() >= 0 {
			break // winner resets the streak
		}
		streak++
		lastLoseAt = r.ExitTime // the streak-closing close (oldest in the streak)
	}
	if streak < n {
		return
	}
	until := time.UnixMilli(lastLoseAt).Add(time.Duration(kernel.LossStreakPauseMin()) * time.Minute)
	// Cap at the active session's end (whichever is first).
	if sess, ok := at.sessionRegistry(now).ActiveSession(now); ok {
		if end, perr := parseCTClock(sess.WindowEndCT, now); perr == nil && end.Before(until) {
			until = end
		}
	}
	if !now.Before(until) {
		return // timer expired
	}
	ctx.LossStreakPaused = true
	ctx.LossStreakMsg = kernel.LossStreakRefusal(n, until)

	key := fmt.Sprintf("loss-streak:%d", day.UnixMilli())
	if at.lastLossStreakAlertKey != key {
		at.lastLossStreakAlertKey = key
		at.logWarnf("🧊 %s", ctx.LossStreakMsg)
		at.emitAlert("P0", "loss-streak", "loss-streak:"+key,
			"LOSS-STREAK PAUSE — new entries paused",
			fmt.Sprintf("%d consecutive losers in this session — new entries paused until %s.", n, kernel.FormatCT(until)))
		if at.store != nil {
			at.store.LogEvent().Enqueue(store.LogEventDB{
				TsUTC:     now.UnixMilli(),
				Level:     "WARN",
				Component: "loss_streak_pause",
				TraderID:  at.id,
				Message:   ctx.LossStreakMsg,
			})
		}
	}
}

// parseCTClock resolves a "HH:MM" CT window edge on the given day (an edge
// earlier than now belongs to TOMORROW — the overnight sessions wrap midnight).
func parseCTClock(hhmm string, now time.Time) (time.Time, error) {
	var h, m int
	if _, err := fmt.Sscanf(hhmm, "%d:%d", &h, &m); err != nil {
		return time.Time{}, err
	}
	t := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, kernel.CTLocation())
	if t.Before(now) {
		t = t.Add(24 * time.Hour)
	}
	return t, nil
}
