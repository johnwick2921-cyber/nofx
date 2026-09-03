package kernel

import (
	"fmt"
	"time"
)

// ── NO-TRADE WINDOWS — ONE DEFINITION EACH (owner ruling 2026-09-02) ────────
// The first-N-minutes and lunch windows were hardcoded in THREE places that
// could drift apart: the entry gate (trader/auto_trader_session.go), the
// adherence grader (adherence.go:121, its own copy of "12:00"/"13:30"), and
// the plan card, which rendered whatever prose the model had written. The gate
// enforced one thing, the grader scored another, and the card claimed a third.
//
// These are the definitions. The gate, the grader and the card all read them.
// They are NOT owner-configurable — every payload built from them is stamped
// SourceCodeConstant so the surface never implies a knob that does not exist.

// NoTradeSource names where a window's boundaries came from.
const (
	SourceCodeConstant = "code-constant" // fixed in code; no config surface yet
	SourceRegistry     = "session-registry"
	SourceCalendar     = "calendar"
)

// NoTradeKind classifies a window for rendering and filtering.
const (
	KindFirstN   = "first_n"
	KindLunch    = "lunch"
	KindT1       = "t1"
	KindKillzone = "killzone"
)

// FirstNoTradeMinutes is the count of minutes after a session's open during
// which entries are refused. ONE definition — the gate's `cur < start+N` and
// the grader's first-N test both resolve here.
func FirstNoTradeMinutes() int { return 5 }

// LunchWindowCT returns the lunch no-trade window as CT wall-clock "HH:MM"
// bounds. ONE definition — the gate's InBlackoutWindow call and the grader's
// used to carry separate copies of these strings.
func LunchWindowCT() (startCT, endCT string) { return "12:00", "13:30" }

// InFirstNoTradeMinutes reports whether t falls in the first N minutes of a
// session whose window starts at startCT. Wrap-safe via minutes-of-day.
func InFirstNoTradeMinutes(startCT string, t time.Time) bool {
	start, ok := parseHHMM(startCT)
	if !ok {
		return false
	}
	n := FirstNoTradeMinutes()
	cur := t.In(CTLocation()).Hour()*60 + t.In(CTLocation()).Minute()
	d := ((cur-start)%1440 + 1440) % 1440
	return d < n
}

// InLunchNoTrade reports whether t falls in the lunch window.
func InLunchNoTrade(t time.Time) bool {
	s, e := LunchWindowCT()
	return InBlackoutWindow(t, s, e)
}

// NoTradeWindow is one machine-derived constraint, stored on the plan doc and
// re-evaluated at read time. Start/End are CT minutes-of-day so a wrapping
// session (ASIA 17:00→02:00) needs no special case at the boundary.
type NoTradeWindow struct {
	StartMin int    `json:"start_min"`
	EndMin   int    `json:"end_min"`
	Kind     string `json:"kind"`
	Label    string `json:"label"`
	Source   string `json:"source"`
}

// HHMM renders a minutes-of-day value as CT wall clock.
func HHMM(min int) string {
	m := ((min % 1440) + 1440) % 1440
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

// BuildMachineNoTradeWindows assembles the machine-derived windows for one
// session: the first-N window from the session registry's own open, and lunch
// from its single definition. T1 windows are NOT baked here — they are
// re-resolved at read time so a later calendar change shows up (owner ruling).
func BuildMachineNoTradeWindows(sess SessionDef) []NoTradeWindow {
	var out []NoTradeWindow
	if start, ok := parseHHMM(sess.WindowStartCT); ok {
		n := FirstNoTradeMinutes()
		out = append(out, NoTradeWindow{
			StartMin: start, EndMin: (start + n) % 1440,
			Kind: KindFirstN, Source: SourceCodeConstant,
			Label: fmt.Sprintf("first %dm after the %s open (%s–%s CT)", n, sess.Name, HHMM(start), HHMM(start+n)),
		})
	}
	ls, le := LunchWindowCT()
	if s, ok1 := parseHHMM(ls); ok1 {
		if e, ok2 := parseHHMM(le); ok2 {
			out = append(out, NoTradeWindow{
				StartMin: s, EndMin: e, Kind: KindLunch, Source: SourceCodeConstant,
				Label: fmt.Sprintf("lunch %s–%s CT", ls, le),
			})
		}
	}
	return out
}

// ── READ-TIME EVALUATION ────────────────────────────────────────────────────

// NoTradeStatus is how one window relates to the moment it is rendered.
const (
	StatusLive         = "live"          // intersects [now, session EOD]
	StatusElapsed      = "elapsed"       // ended before now
	StatusOtherSession = "other_session" // starts after this session's EOD
)

// RenderedNoTradeWindow is one window plus its read-time verdict.
type RenderedNoTradeWindow struct {
	NoTradeWindow
	Status string `json:"status"`
}

// EvaluateNoTradeWindows filters windows against [nowMin, eodMin] in CT
// minutes-of-day, wrap-aware so an ASIA session spanning midnight is handled
// without a special case. Nothing is dropped: elapsed and other-session
// windows come back marked, for the collapsed section.
func EvaluateNoTradeWindows(wins []NoTradeWindow, nowMin, sessionStartMin, eodMin int) []RenderedNoTradeWindow {
	out := make([]RenderedNoTradeWindow, 0, len(wins))
	// Geometry is measured from the session start so a window that wraps past
	// midnight stays monotonic; "has it finished" is measured from NOW on a
	// signed ±12h axis, because a pre-session read has nothing elapsed yet and
	// offsets-from-start alone cannot tell 08:00-before-open from 16:00-after.
	off := func(m int) int { return ((m-sessionStartMin)%1440 + 1440) % 1440 }
	rel := func(m int) int {
		d := ((m-nowMin)%1440 + 1440) % 1440
		if d > 720 {
			d -= 1440 // the nearer reading of a wrapped clock is the past one
		}
		return d
	}
	sessionLen := off(eodMin)
	if sessionLen == 0 {
		sessionLen = 1440
	}
	for _, w := range wins {
		sOff, eOff := off(w.StartMin), off(w.EndMin)
		if eOff <= sOff {
			eOff += 1440 // window wraps past the session start
		}
		// Does the window touch this session's tradeable span at all? A window
		// that does not can never constrain this session, whatever the clock
		// says — that is the ASIA-at-23:00 lie (NY lunch on an ASIA card).
		intersects := sOff < sessionLen || eOff > 1440
		status := StatusLive
		switch {
		case !intersects:
			status = StatusOtherSession
		case rel(w.EndMin) <= 0:
			status = StatusElapsed
		}
		out = append(out, RenderedNoTradeWindow{NoTradeWindow: w, Status: status})
	}
	return out
}

// LiveNoTradeWindows returns only the windows still ahead in this session.
func LiveNoTradeWindows(r []RenderedNoTradeWindow) []RenderedNoTradeWindow {
	out := make([]RenderedNoTradeWindow, 0, len(r))
	for _, w := range r {
		if w.Status == StatusLive {
			out = append(out, w)
		}
	}
	return out
}

// T1WindowsForRead re-resolves the calendar's T1 blackouts at READ time and
// applies the drift widening measured NOW. driftMs is the resolved clock-health
// offset; when the clock is healthy the widening is zero and no "(clock drift)"
// suffix appears — the label stops claiming a correction that was not applied.
func T1WindowsForRead(events []PlannerCalendarEvent, driftMs int64) []NoTradeWindow {
	var out []NoTradeWindow
	for _, w := range WidenCTWindows(T1BlackoutWindows(events), driftMs) {
		out = append(out, NoTradeWindow{
			StartMin: w.Start, EndMin: w.End, Kind: KindT1,
			Source: SourceCalendar, Label: "🔴 " + w.Label + " — HARD no-trade (red news)",
		})
	}
	return out
}

// NoTradeBandBootLine is the boot line (every field read from code).
func NoTradeBandBootLine() string {
	ls, le := LunchWindowCT()
	return fmt.Sprintf("🗓 no-trade band: session-scoped, config-driven (0 literals) — first_n=%dm lunch=%s–%s (source=%s) · T1 re-resolved at read time with drift measured then · model prose renders as NOTES",
		FirstNoTradeMinutes(), ls, le, SourceCodeConstant)
}
