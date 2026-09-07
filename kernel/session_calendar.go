package kernel

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ── THE CME SESSION CALENDAR (2026-09-07) ────────────────────────────────────
//
// SHORTENED SESSIONS ARE TRADING SESSIONS (owner ruling 2026-09-07).
//
// This replaced a boolean. isCMEHoliday() answered yes/no and the gate treated
// every holiday as a FULL closure — the old code said so itself: "for v1 we
// treat them as full closures and refuse to trade. Refine in Plan 3 if it
// becomes restrictive." On 2026-09-07 it became restrictive. MNQ traded 980
// bars across 153.50 points on Labor Day while the gate called the market shut,
// every cycle was skipped, no LONDON plan was ever read, and the 3-minute
// closed-market backoff logged 156 "overruns" against a 2-minute interval.
//
// A date is now one of three things, and the third is the one the boolean could
// not express: CLOSED, SHORTENED (trades until its stated close, then flat), or
// NORMAL (absent from the calendar — the weekly rules alone decide).
//
// THE CALENDAR IS DATA. It lives in session_calendar.json, embedded so it ships
// with the binary and cannot go missing at runtime, and every date cites a
// source that says whether it was verified or decided.

//go:embed session_calendar.json
var sessionCalendarJSON []byte

// SessionClass is what a date IS.
type SessionClass string

const (
	// SessionClosed — no trading at all.
	SessionClosed SessionClass = "closed"
	// SessionShortened — trades normally until CloseCT, then flat.
	SessionShortened SessionClass = "shortened"
	// SessionNormal — the weekly rules alone decide. Never stored; returned
	// for any date the calendar does not list.
	SessionNormal SessionClass = "normal"
)

// SessionDay is one calendar entry, with the provenance that lets a reader tell
// a published fact from a decision.
type SessionDay struct {
	Date    string       `json:"date"`
	Class   SessionClass `json:"class"`
	CloseCT string       `json:"close_ct,omitempty"`
	Name    string       `json:"name"`
	Source  string       `json:"source"`
}

type sessionCalendarFile struct {
	CoveredYears []int        `json:"covered_years"`
	Dates        []SessionDay `json:"dates"`
}

var sessionCal sessionCalendarFile

func init() {
	if err := json.Unmarshal(sessionCalendarJSON, &sessionCal); err != nil {
		// The calendar is embedded, so this can only fail if someone shipped
		// invalid JSON — which the loader test catches before it ever boots.
		panic("kernel: session_calendar.json is not valid JSON: " + err.Error())
	}
}

// SessionCalendarCoversYear reports whether the calendar has been maintained
// for this year. An uncovered year is NOT a year of normal days — it is a year
// nobody has checked, and the boot line says so out loud.
func SessionCalendarCoversYear(year int) bool {
	for _, y := range sessionCal.CoveredYears {
		if y == year {
			return true
		}
	}
	return false
}

// SessionDayFor returns the calendar entry for this date, if it has one.
func SessionDayFor(t time.Time) (SessionDay, bool) {
	key := t.In(CTLocation()).Format("2006-01-02")
	for i := range sessionCal.Dates {
		if sessionCal.Dates[i].Date == key {
			return sessionCal.Dates[i], true
		}
	}
	return SessionDay{Date: key, Class: SessionNormal}, false
}

// parseCloseCT turns "12:00" into that instant on the given CT date. A malformed
// or absent time is reported, never guessed: a shortened day whose close cannot
// be read must not silently become a full trading day.
func parseCloseCT(ct time.Time, hhmm string) (time.Time, bool) {
	parts := strings.SplitN(strings.TrimSpace(hhmm), ":", 2)
	if len(parts) != 2 {
		return time.Time{}, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return time.Time{}, false
	}
	loc := CTLocation()
	c := ct.In(loc)
	return time.Date(c.Year(), c.Month(), c.Day(), h, m, 0, 0, loc), true
}

// weeklyCMEOpen is the ordinary week: the rules that applied before any
// calendar existed, unchanged.
func weeklyCMEOpen(ct time.Time) bool {
	switch ct.Weekday() {
	case time.Saturday:
		return false
	case time.Sunday:
		return ct.Hour() >= 17
	case time.Friday:
		return ct.Hour() < 16
	default: // Mon-Thu
		return ct.Hour() != 16
	}
}

// SessionCalendarBootLine names TODAY's classification and, when it has one, the
// time it closes. Every field READ from the calendar the gate actually consults
// (A11) — including the source, so the owner can see whether today's answer was
// published or decided.
func SessionCalendarBootLine(now time.Time) string {
	ct := now.In(CTLocation())
	covered := "covered"
	if !SessionCalendarCoversYear(ct.Year()) {
		covered = fmt.Sprintf("YEAR %d NOT COVERED — no one has checked it; the weekly rules alone are deciding", ct.Year())
	}
	day, listed := SessionDayFor(ct)
	if !listed {
		return fmt.Sprintf("session calendar: today %s is NORMAL (not listed) · %d date(s) listed for %v · %s",
			ct.Format("2006-01-02"), len(sessionCal.Dates), sessionCal.CoveredYears, covered)
	}
	switch day.Class {
	case SessionClosed:
		return fmt.Sprintf("session calendar: today %s is CLOSED — %s · no trading · source: %s · %s",
			day.Date, day.Name, day.Source, covered)
	case SessionShortened:
		closeTxt := day.CloseCT + " CT"
		if _, ok := parseCloseCT(ct, day.CloseCT); !ok {
			closeTxt = fmt.Sprintf("UNREADABLE close_ct %q — treated as CLOSED rather than guessed", day.CloseCT)
		}
		return fmt.Sprintf("session calendar: today %s is SHORTENED — %s · TRADES until %s, then flat · source: %s · %s",
			day.Date, day.Name, closeTxt, day.Source, covered)
	}
	return fmt.Sprintf("session calendar: today %s is NORMAL · %s", day.Date, covered)
}
