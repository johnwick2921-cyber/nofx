package trader

import (
	"math"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nofx/store"
)

// E1's RED is the base commit itself: at e28b604d there is no api/handler_desk.go,
// no trader/desk_facts.go and no DeskStrip type, so none of these assertions could
// even compile. `git show e28b604d:trader/desk_facts.go` → "path does not exist".

func deskStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.New(filepath.Join(t.TempDir(), "desk.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

// THE 09-04 FIXTURE, from the store: armed_orders 35's ledger stop against the
// stop NT8 actually accepted. 3.3715271 points apart.
const (
	arm35Ledger   = 29351.6284728996
	arm35Accepted = 29355.0
)

func openPos(signal string, entry, mark float64) map[string]interface{} {
	return map[string]interface{}{
		"side": "SHORT", "quantity": 1.0, "entryPrice": entry, "markPrice": mark,
		"signal_id": signal,
	}
}

// E2 — THE LEDGER PRICE IS NEVER SHOWN AS THE BROKER'S.
// PROTECTION renders the ACCEPTED stop; DRIFT renders the difference; and the
// ledger number must not appear anywhere in the PROTECTION row.
func TestProtectionShowsTheAcceptedStopAndDriftShowsTheDifference(t *testing.T) {
	st := deskStore(t)
	accepted := arm35Accepted
	if err := st.AcceptedRisk().Append(&store.AcceptedRisk{
		TraderID: "hoang", SignalID: "f2b1eb20", OrderName: "f2b1eb20-sl", Side: "SHORT",
		AcceptedStopPx: &accepted, LedgerStopPx: arm35Ledger, AcceptedAtMs: 1788442433811,
	}); err != nil {
		t.Fatal(err)
	}
	at := &AutoTrader{id: "hoang", store: st}
	pos := openPos("f2b1eb20", 29285, 29300)

	prot := at.deskProtection(pos, 0.25, 2, time.Now().UnixMilli())
	if prot.State != "ok" {
		t.Fatalf("PROTECTION should render from the accepted record, got state %q reason %q", prot.State, prot.Reason)
	}
	if !strings.Contains(prot.Text, "29355") {
		t.Fatalf("PROTECTION must show the ACCEPTED stop 29355: %s", prot.Text)
	}
	if strings.Contains(prot.Text, "29351.6") {
		t.Fatalf("E2: the LEDGER price must never be rendered as your stop: %s", prot.Text)
	}

	drift := at.deskDrift(pos, time.Now().UnixMilli())
	if !strings.Contains(drift.Text, "3.371527") {
		t.Fatalf("DRIFT must state the difference to the point: %s", drift.Text)
	}
	t.Logf("PROTECTION: %s", prot.Text)
	t.Logf("DRIFT:      %s", drift.Text)
}

// E2b — WITH NO ACCEPTED RECORD, PROTECTION IS UNKNOWN WITH A REASON.
// It must never fall back to the ledger's stop, which is the whole defect.
func TestProtectionIsUnknownRatherThanFallingBackToTheLedger(t *testing.T) {
	st := deskStore(t)
	at := &AutoTrader{id: "hoang", store: st}
	prot := at.deskProtection(openPos("no-such-signal", 29285, 29300), 0.25, 2, time.Now().UnixMilli())
	if prot.State != "unknown" {
		t.Fatalf("with no accepted record PROTECTION must be UNKNOWN, got %q", prot.State)
	}
	if strings.TrimSpace(prot.Reason) == "" {
		t.Fatal("an UNKNOWN with no reason is the same silence it replaced")
	}
	if prot.Text != "UNKNOWN" || strings.Contains(prot.Text, "0.00") {
		t.Fatalf("UNKNOWN must not render as a number or a dash: %q", prot.Text)
	}
	t.Logf("reason: %s", prot.Reason)
}

// E4 — FLAT IS NOT UNKNOWN AND NOT ZERO. Three different facts, three renders.
func TestFlatRendersFlatNotUnknownAndNotZero(t *testing.T) {
	st := deskStore(t)
	at := &AutoTrader{id: "hoang", store: st}
	now := time.Now().UnixMilli()
	for _, l := range []DeskLine{
		at.deskProtection(nil, 0.25, 2, now),
		at.deskTarget(nil, 0.25, 2, now),
		at.deskDrift(nil, now),
		at.deskPositionLine(nil, nil, 2, now),
	} {
		if l.State != "flat" {
			t.Fatalf("line %d (%s) with no position must read FLAT, got %q", l.N, l.Key, l.State)
		}
		if strings.Contains(l.Text, "0.00") || l.Text == "-" || l.Text == "UNKNOWN" {
			t.Fatalf("line %d rendered a zero/dash/UNKNOWN for FLAT: %q", l.N, l.Text)
		}
	}
}

// E1 — EVERY LINE IS DATED OR EXPLICITLY UNKNOWN WITH A REASON, and no line
// ever renders a bare zero or a dash for something it did not compute.
func TestEveryLineIsDatedOrUnknownWithAReason(t *testing.T) {
	st := deskStore(t)
	at := &AutoTrader{id: "hoang", store: st, config: AutoTraderConfig{NinjaTraderSymbol: "MNQ"}}
	s := at.DeskStripAt(time.Now())

	if len(s.Lines) != 12 {
		t.Fatalf("the strip is defined as 12 lines, got %d", len(s.Lines))
	}
	for _, l := range s.Lines {
		switch l.State {
		case "unknown", "stale":
			if strings.TrimSpace(l.Reason) == "" {
				t.Fatalf("line %d (%s) is %s with NO reason", l.N, l.Key, l.State)
			}
		case "ok", "flat":
			if l.AsOfMs == 0 {
				t.Fatalf("line %d (%s) rendered a value with no as-of instant — nothing renders undated", l.N, l.Key)
			}
		default:
			t.Fatalf("line %d (%s) has an unknown state %q", l.N, l.Key, l.State)
		}
		if strings.TrimSpace(l.Text) == "" || strings.TrimSpace(l.Text) == "-" {
			t.Fatalf("line %d (%s) rendered an empty value or a dash: %q", l.N, l.Key, l.Text)
		}
		if strings.TrimSpace(l.Source) == "" {
			t.Fatalf("line %d (%s) names no source", l.N, l.Key)
		}
	}
	t.Logf("12 lines, %d unknown, %d stale, cadence %dms", s.UnknownCount, s.StaleCount, s.CadenceMs)
}

// E6 — CADENCE IS RESOLVED FROM WHAT IS LIVE, never a file default.
func TestCadenceIsFiveSecondsOnlyWhenSomethingIsLive(t *testing.T) {
	st := deskStore(t)
	at := &AutoTrader{id: "hoang", store: st, config: AutoTraderConfig{NinjaTraderSymbol: "MNQ"}}
	if s := at.DeskStripAt(time.Now()); s.CadenceMs != deskCadenceIdleMs {
		t.Fatalf("flat and armless must poll at %dms, got %d", deskCadenceIdleMs, s.CadenceMs)
	}
	// A resting arm makes it live.
	row := &store.ArmedOrderDB{
		TraderID: "hoang", PlanID: "P", Session: "NY", Scenario: "S1", Side: "short",
		EntryPx: 29720, StopPx: 29755, TargetPx: 29635, State: store.StateArmed,
	}
	if err := st.ArmedOrders().UpsertArm(row); err != nil {
		t.Fatal(err)
	}
	if s := at.DeskStripAt(time.Now()); s.CadenceMs != deskCadenceLiveMs {
		t.Fatalf("a resting arm must poll at %dms, got %d", deskCadenceLiveMs, s.CadenceMs)
	}
}

// A10 / class 23 — a broken dependency renders UNKNOWN rows, never a panic and
// never a blank strip.
func TestStripSurvivesABrokenStoreAndStillReturnsTwelveLines(t *testing.T) {
	st := deskStore(t)
	at := &AutoTrader{id: "hoang", store: st, config: AutoTraderConfig{NinjaTraderSymbol: "MNQ"}}
	_ = st.Close() // the harshest realistic failure

	var s DeskStrip
	if r := recoverOf(func() { s = at.DeskStripAt(time.Now()) }); r != nil {
		t.Fatalf("the strip panicked through to the caller: %v", r)
	}
	if len(s.Lines) != 12 {
		t.Fatalf("a broken store must still render all 12 rows, got %d", len(s.Lines))
	}
	for _, l := range s.Lines {
		if l.State == "unknown" && strings.TrimSpace(l.Reason) == "" {
			t.Fatalf("line %d is UNKNOWN with no reason under failure", l.N)
		}
	}
	t.Logf("with the store closed: %d unknown of 12", s.UnknownCount)
}

// The boot line reads its fields (A11) and states the rule it enforces.
func TestDeskBootLineReadsItsFields(t *testing.T) {
	line := DeskBootLine(12, 7)
	for _, want := range []string{"desk strip:", "lines=12", "unknown-at-boot=7", "cadence=5s live / 15s idle", "book-age-bound=", "never a zero"} {
		if !strings.Contains(line, want) {
			t.Fatalf("boot line missing %q:\n%s", want, line)
		}
	}
	t.Logf("%s", line)
}

func TestDeskAgeNeverRendersNegative(t *testing.T) {
	if got := deskAge(-5000); got != "0s" {
		t.Fatalf("a negative age is a clock artifact, not a fact: %q", got)
	}
	if got := deskAge(int64(90 * time.Second / time.Millisecond)); got != "1m30s" {
		t.Fatalf("age rendering changed: %q", got)
	}
	_ = math.Abs
}
