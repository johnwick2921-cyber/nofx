package kernel

import (
	"testing"
	"time"
)

func ctAt(hh, mm int) time.Time {
	return time.Date(2026, 9, 2, hh, mm, 0, 0, CTLocation())
}

// THE RULING'S TEST: the gate, the grader and the card must return the SAME
// window for the same clock. Before this the grader carried its own copies of
// "12:00"/"13:30" and its own first-5m helper, so it could score against
// windows the gate did not enforce.
func TestNoTradeWindowsOneDefinitionForGateGraderCard(t *testing.T) {
	ls, le := LunchWindowCT()
	n := FirstNoTradeMinutes()

	// The CARD's payload, built from the same definitions.
	sess := SessionDef{Name: SessionNY, WindowStartCT: "08:30", WindowEndCT: "14:45"}
	wins := BuildMachineNoTradeWindows(sess)
	var lunch, firstN *NoTradeWindow
	for i := range wins {
		switch wins[i].Kind {
		case KindLunch:
			lunch = &wins[i]
		case KindFirstN:
			firstN = &wins[i]
		}
	}
	if lunch == nil || firstN == nil {
		t.Fatalf("card payload missing a kind: %+v", wins)
	}
	if HHMM(lunch.StartMin) != ls || HHMM(lunch.EndMin) != le {
		t.Fatalf("card lunch %s–%s ≠ definition %s–%s", HHMM(lunch.StartMin), HHMM(lunch.EndMin), ls, le)
	}
	if firstN.EndMin-firstN.StartMin != n {
		t.Fatalf("card first-N spans %d min, definition says %d", firstN.EndMin-firstN.StartMin, n)
	}

	// The GRADER and the CARD must agree at every minute across both windows.
	for _, tc := range []struct {
		at   time.Time
		want bool
		why  string
	}{
		{ctAt(8, 30), true, "first minute after the open"},
		{ctAt(8, 34), true, "last minute of the first-N window"},
		{ctAt(8, 35), false, "one minute past first-N"},
		{ctAt(12, 0), true, "lunch open"},
		{ctAt(13, 29), true, "lunch last minute"},
		{ctAt(13, 30), false, "lunch close is exclusive"},
		{ctAt(11, 59), false, "one minute before lunch"},
	} {
		graderInNoTrade := InLunchNoTrade(tc.at) || InFirstNoTradeMinutes(sess.WindowStartCT, tc.at)
		cardInNoTrade := false
		cur := tc.at.In(CTLocation()).Hour()*60 + tc.at.In(CTLocation()).Minute()
		for _, w := range wins {
			if cur >= w.StartMin && cur < w.EndMin {
				cardInNoTrade = true
			}
		}
		if graderInNoTrade != tc.want || cardInNoTrade != tc.want {
			t.Errorf("%s at %s: grader=%v card=%v want %v",
				tc.why, tc.at.Format("15:04"), graderInNoTrade, cardInNoTrade, tc.want)
		}
	}
}

// Every machine window is stamped with an honest source. The two fixed windows
// say code-constant so the surface never implies a knob that does not exist.
func TestNoTradeWindowsCarryHonestSource(t *testing.T) {
	wins := BuildMachineNoTradeWindows(SessionDef{Name: SessionAsia, WindowStartCT: "17:00", WindowEndCT: "02:00"})
	for _, w := range wins {
		if w.Source != SourceCodeConstant {
			t.Errorf("%s claims source %q — neither window is owner-configurable yet", w.Kind, w.Source)
		}
		if w.Label == "" {
			t.Errorf("%s has no label", w.Kind)
		}
	}
	// A wrapping session's first-N must not wrap incorrectly.
	asia := wins[0]
	if asia.Kind != KindFirstN || HHMM(asia.StartMin) != "17:00" || HHMM(asia.EndMin) != "17:05" {
		t.Fatalf("ASIA first-N wrong: %s–%s", HHMM(asia.StartMin), HHMM(asia.EndMin))
	}
	// And the first-N test must work across midnight for that session.
	if !InFirstNoTradeMinutes("17:00", ctAt(17, 2)) {
		t.Fatal("17:02 is inside ASIA's first-N")
	}
	if InFirstNoTradeMinutes("17:00", ctAt(16, 59)) {
		t.Fatal("16:59 is before the open")
	}
}

// F1 — THE PIN. The ASIA card at 23:00 CT showed three constraints when one
// applied: first-5m (elapsed six hours earlier), the NY lunch window (cannot
// apply to an ASIA session), and a 09:00 T1 blackout that had passed fourteen
// hours before. Session-scoped evaluation collapses all three correctly.
func TestNoTradeBandAsiaAt2300(t *testing.T) {
	asia := SessionDef{Name: SessionAsia, WindowStartCT: "17:00", WindowEndCT: "02:00"}
	wins := BuildMachineNoTradeWindows(asia)
	// A 09:00 CT T1 event, as the calendar would resolve it at read time.
	wins = append(wins, NoTradeWindow{StartMin: 8*60 + 45, EndMin: 9*60 + 15,
		Kind: KindT1, Source: SourceCalendar, Label: "🔴 09:00 ISM PMI — HARD no-trade (red news)"})

	now := 23 * 60   // 23:00 CT
	start := 17 * 60 // ASIA opens 17:00
	eod := 2 * 60    // ASIA flat 02:00 (wraps midnight)
	got := EvaluateNoTradeWindows(wins, now, start, eod)

	byKind := map[string]string{}
	for _, w := range got {
		byKind[w.Kind] = w.Status
	}
	if byKind[KindFirstN] != StatusElapsed {
		t.Errorf("first-5m opened at 17:00 and is six hours gone at 23:00 — want elapsed, got %q", byKind[KindFirstN])
	}
	if byKind[KindLunch] != StatusOtherSession {
		t.Errorf("lunch 12:00–13:30 is an NY window and cannot apply to ASIA — want other_session, got %q", byKind[KindLunch])
	}
	if byKind[KindT1] != StatusOtherSession {
		t.Errorf("a 09:00 T1 is not in ASIA's remaining window — want other_session, got %q", byKind[KindT1])
	}
	if n := len(LiveNoTradeWindows(got)); n != 0 {
		t.Fatalf("ASIA at 23:00 has NO live constraint, got %d: %+v", n, LiveNoTradeWindows(got))
	}
}

// F2 — NY at 08:00: first-5m and lunch are both ahead; an 07:30 T1 is gone.
func TestNoTradeBandNYAt0800(t *testing.T) {
	ny := SessionDef{Name: SessionNY, WindowStartCT: "08:30", WindowEndCT: "14:45"}
	wins := BuildMachineNoTradeWindows(ny)
	wins = append(wins,
		NoTradeWindow{StartMin: 7*60 + 15, EndMin: 7*60 + 45, Kind: KindT1, Source: SourceCalendar, Label: "🔴 07:30 event"},
		NoTradeWindow{StartMin: 9*60 + 45, EndMin: 10*60 + 15, Kind: KindT1, Source: SourceCalendar, Label: "🔴 10:00 event"})
	got := EvaluateNoTradeWindows(wins, 8*60, 8*60+30, 14*60+45)

	live := map[string]bool{}
	for _, w := range LiveNoTradeWindows(got) {
		live[w.Label] = true
	}
	if len(live) != 3 {
		t.Fatalf("want first-5m + lunch + the 10:00 T1 live, got %d: %+v", len(live), live)
	}
	// The 07:15–07:45 blackout never overlaps NY's 08:30–14:45 span, so it is
	// scoped out rather than merely spent. Either way it must not render live.
	for _, w := range got {
		if w.Label == "🔴 07:30 event" && w.Status == StatusLive {
			t.Errorf("an 07:30 T1 cannot constrain an 08:30 session, got %q", w.Status)
		}
	}
}

// F3 — a window straddling "now" is LIVE, not elapsed.
func TestNoTradeBandStraddlingNowIsLive(t *testing.T) {
	got := EvaluateNoTradeWindows(
		[]NoTradeWindow{{StartMin: 12 * 60, EndMin: 13*60 + 30, Kind: KindLunch}},
		12*60+45, 8*60+30, 14*60+45)
	if got[0].Status != StatusLive {
		t.Fatalf("12:45 is inside 12:00–13:30 — want live, got %q", got[0].Status)
	}
}

// F5 — a healthy clock produces NO widening and NO "(clock drift)" label. The
// suffix used to be baked in at plan time whether or not it applied.
func TestT1WindowsForReadDriftLabel(t *testing.T) {
	evs := []PlannerCalendarEvent{{Impact: "T1", TimeCT: "09:00", Title: "ISM PMI"}}
	healthy := T1WindowsForRead(evs, 108) // +108 ms, the audited NTP offset
	if len(healthy) == 0 {
		t.Skip("no T1 window produced by this calendar shape")
	}
	for _, w := range healthy {
		if contains(w.Label, "clock drift") {
			t.Fatalf("a healthy clock must not claim the machine is drifting: %q", w.Label)
		}
		// The minute of boundary protection is KEPT — only the claim is dropped.
		if w.EndMin != (9*60+15+1)%1440 {
			t.Fatalf("boundary widening must survive the label change, end=%d", w.EndMin)
		}
		if w.Source != SourceCalendar {
			t.Fatalf("T1 windows come from the calendar, got %q", w.Source)
		}
	}
	skewed := T1WindowsForRead(evs, 90_000) // 90 s of real skew
	for _, w := range skewed {
		if !contains(w.Label, "clock drift") {
			t.Fatalf("a real skew must be stated in the label: %q", w.Label)
		}
	}
}
