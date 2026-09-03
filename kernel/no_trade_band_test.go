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
