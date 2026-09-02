package trader

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"nofx/kernel"
)

// ── CLASS 47 (2026-09-02) — WAKE CADENCE ────────────────────────────────────
//
// Measured on the 7-day tape and today's session (a5a53bec cadence audit,
// re-measured in this wave):
//
//   · 60 level_event wake re-plans in 7 days; 33 arm rows written, 23 ever
//     placed, 9 ever working/filled. Most wakes buy a plan version nobody
//     trades.
//   · Today's wakes fire at 08:42:30 · 09:12:30 · 09:42:30 · 10:14:30 ·
//     10:44:30 · 11:15 · 11:45 · 12:16:30 · 12:48:29 · 13:18:29 · 13:48:29 ·
//     14:20:29 — a clean ~30-minute drumbeat. The wake CONDITION is
//     continuously true; the existing wake_min_interval_min throttle is the
//     only thing pacing them. NY produced 12 plan versions on that pattern.
//   · A wake at 14:20:29 sits 10 minutes from the NY last-entry cutoff (14:30)
//     and 25 from the flat: a max-reasoning read whose plan can never be
//     entered.
//
// EVERYTHING HERE EXCEPT THE ARM EXPIRY IS WARN-FIRST (owner ruling): the wake
// still runs. We are measuring what suppression WOULD have cost before anyone
// suppresses anything. The counters are recorded (class-35 law) so a week of
// real numbers, not a week of impressions, backs the eventual ruling.

// WakeCutoffMinDefault — a wake this close to the session flat cannot produce a
// tradeable plan: the last-entry cutoff is flat−15, and the p90 planner call is
// ~9.3 min, so a read starting inside ~25 min of the flat lands after the gate
// has already closed. Env WAKE_CUTOFF_MIN; 0 disables the check.
const WakeCutoffMinDefault = 25

// WakeCooldownMinDefault — minutes since the last wake-AUTHORED plan version.
// Deliberately distinct from wake_min_interval_min (which paces wake ATTEMPTS
// from the last attempt): this measures from the last version a wake actually
// WROTE, so a wake whose read failed or kept the active plan does not start the
// clock. Env WAKE_COOLDOWN_MIN; 0 disables.
const WakeCooldownMinDefault = 30

func wakeCutoffMinutes() int   { return envPosInt("WAKE_CUTOFF_MIN", WakeCutoffMinDefault) }
func wakeCooldownMinutes() int { return envPosInt("WAKE_COOLDOWN_MIN", WakeCooldownMinDefault) }

// envPosInt resolves a non-negative integer knob (0 = feature off).
func envPosInt(name string, def int) int {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return def
}

// minutesToSessionFlat returns whole minutes from now to this session's window
// end (the flat), wrap-aware for midnight-spanning sessions. ok=false outside
// the window or on malformed registry times — never invent a deadline.
func minutesToSessionFlat(now time.Time, sess *kernel.SessionDef) (int, bool) {
	if sess == nil {
		return 0, false
	}
	startMin, okS := hhmmToMin(sess.WindowStartCT)
	endMin, okE := hhmmToMin(sess.WindowEndCT)
	if !okS || !okE {
		return 0, false
	}
	sessLen := ((endMin-startMin)%1440 + 1440) % 1440
	nowOff := ((ctMinutesNow(now)-startMin)%1440 + 1440) % 1440
	if nowOff >= sessLen {
		return 0, false // outside the window
	}
	return sessLen - nowOff, true
}

// anyPlannerStreamOpen reports whether ANY planner read is in flight, for any
// trading session and any trader in this process, and names one.
//
// F3: the claim key is per (trader, trade_date, session)
// (auto_trader_planner.go — `claimPlannerRead(store.MakePlanIDForTrader(...))`),
// so a LONDON read and an NY read hold different claims and stream
// concurrently. Today 08:01:06 opened a second max-reasoning stream while the
// 07:51:06 one was still running — the LONDON→NY handover overlap.
//
// A WAKE defers on this. A SCHEDULED read never does: the naive "defer on any
// stream" rule would have parked today's NY session read behind LONDON's, which
// is strictly worse than an overlap.
func anyPlannerStreamOpen() (string, bool) {
	held := ""
	plannerReadInFlight.Range(func(k, _ any) bool {
		if s, ok := k.(string); ok {
			held = s
			return false
		}
		return true
	})
	return held, held != ""
}

// wakeCutoffLine / wakeCooldownLine / wakeStreamDeferLine are the pure log-line
// builders (A9: loud, and fixture-tested for wording).
func wakeCutoffLine(session, desc string, minsToFlat, cutoff int, n int) string {
	return fmt.Sprintf("⏱ wake would_skip: %d min to flat (cutoff %dm) — %s on %s. WARN-first: the wake PROCEEDS. Recorded n=%d (class 47)",
		minsToFlat, cutoff, desc, session, n)
}

func wakeCooldownLine(session, desc string, sinceMin, cooldown int, n int) string {
	return fmt.Sprintf("⏱ wake would_skip: cooldown %d min since the last wake-authored version (cooldown %dm) — %s on %s. WARN-first: the wake PROCEEDS. Recorded n=%d (class 47)",
		sinceMin, cooldown, desc, session, n)
}

func wakeStreamDeferLine(session, desc, heldKey string) string {
	return fmt.Sprintf("⏱ wake DEFERRED: a planner stream is already open (%s) — %s on %s. Scheduled reads never defer; this is a wake (class 47)",
		heldKey, desc, session)
}

// WakeCadenceBootLine (F5) — every value READ from its resolver (A12/A24: no
// literals in a boot line).
func WakeCadenceBootLine() string {
	cutoff, cooldown := wakeCutoffMinutes(), wakeCooldownMinutes()
	return fmt.Sprintf("wakes: cutoff=%dm cooldown=%dm cross-session=%s stale-arm-expiry=%s (class 47)",
		cutoff, cooldown, onOffWord(cutoff >= 0), onOffWord(true))
}

func onOffWord(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
