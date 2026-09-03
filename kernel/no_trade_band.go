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
