package trader

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/store"
)

// ── CLASS 47 — F4: STALE-ARM EXPIRY (owner-ruled; the only non-WARN change) ──
//
// 2026-09-02: one NEVER-PLACED v5 arm stayed non-terminal across six plan
// versions and held the class-33 cutover gate's leg 4 shut for ~5 hours. A row
// with no broker signal id has nothing at the broker to orphan.
//
// This test uses ONLY pre-class-47 surfaces (UpsertArm + ListNonTerminal), so it
// compiles on the old tree and FAILS there: nothing retires the stale row.
func TestClass47StaleArmExpiry(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "c47.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ledger := st.ArmedOrders()
	const tid, planID = "t1", "2026-09-02:NY:t1"
	now := time.Now()

	seed := func(scenario string, version int, state, signal string) int64 {
		row := &store.ArmedOrderDB{
			TraderID: tid, PlanID: planID, Version: version, Session: "NY",
			Scenario: scenario, Side: "long", EntryPx: 29044, StopPx: 29014, TargetPx: 29104,
			State: state, SignalID: signal, EntryClass: "armed_fill",
			CreatedAt: now, UpdatedAt: now, LegIndex: 0, LegCount: 1, Kind: "limit",
		}
		if err := ledger.UpsertArm(row); err != nil {
			t.Fatal(err)
		}
		return row.ID
	}
	stale := seed("S1", 5, "armed", "")            // the v5 arm that held the gate
	placed := seed("S2", 5, "working", "sig-live") // placed at the broker — MUST survive
	current := seed("S3", 6, "armed", "")          // the CURRENT version — MUST survive

	ids, err := ledger.SupersedeUnplacedArms(tid, planID, 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != stale {
		t.Fatalf("exactly the unplaced v5 row must be retired, got %v (stale=%d)", ids, stale)
	}

	byID := map[int64]store.ArmedOrderDB{}
	rows, _ := ledger.ListForPlan(planID)
	for _, r := range rows {
		byID[r.ID] = r
	}
	if got := byID[stale]; got.State != "superseded" || !strings.Contains(got.StateReason, "never placed") {
		t.Fatalf("the stale arm must be terminal `superseded` with a reason, got state=%q reason=%q", got.State, got.StateReason)
	}
	if got := byID[placed]; got.State != "working" || got.SignalID != "sig-live" {
		t.Fatalf("a PLACED row must be untouched (stop-line), got %+v", got)
	}
	if got := byID[current]; got.State != "armed" {
		t.Fatalf("the CURRENT version's arm must be untouched, got %q", got.State)
	}

	// The gate's leg 4 reads ListNonTerminal — the stale row must no longer hold it.
	nt, _ := ledger.ListNonTerminal(tid)
	for _, r := range nt {
		if r.ID == stale {
			t.Fatal("a superseded row must leave the cutover gate's non-terminal set (leg 4)")
		}
	}
	if len(nt) != 2 {
		t.Fatalf("leg 4 should see exactly the placed + current rows, got %d", len(nt))
	}
	// Idempotent: a second pass retires nothing more.
	if again, _ := ledger.SupersedeUnplacedArms(tid, planID, 6); len(again) != 0 {
		t.Fatalf("second pass must be a no-op, retired %v", again)
	}
}

// ── F1/F2 — WARN-FIRST: the line is logged and the wake still PROCEEDS ──────

func TestClass47CutoffAndCooldownAreWarnFirst(t *testing.T) {
	if got := wakeCutoffMinutes(); got != 25 {
		t.Fatalf("resolved cutoff = %d, want the shipped 25", got)
	}
	if got := wakeCooldownMinutes(); got != 30 {
		t.Fatalf("resolved cooldown = %d, want the shipped 30", got)
	}
	t.Setenv("WAKE_CUTOFF_MIN", "40")
	if got := wakeCutoffMinutes(); got != 40 {
		t.Fatalf("env override ignored, got %d", got)
	}
	t.Setenv("WAKE_CUTOFF_MIN", "0")
	if got := wakeCutoffMinutes(); got != 0 {
		t.Fatalf("0 must disable the check, got %d", got)
	}

	// The lines say would_skip AND say the wake proceeds — the WARN-first contract.
	cut := wakeCutoffLine("NY", "seated OB(bull)·1h invalidated", 10, 25, 3)
	for _, want := range []string{"⏱ wake would_skip", "10 min to flat", "cutoff 25m", "WARN-first: the wake PROCEEDS", "n=3", "class 47"} {
		if !strings.Contains(cut, want) {
			t.Errorf("cutoff line %q missing %q", cut, want)
		}
	}
	cool := wakeCooldownLine("NY", "fresh FVG", 12, 30, 7)
	for _, want := range []string{"cooldown 12 min since the last wake-authored version", "cooldown 30m", "WARN-first: the wake PROCEEDS", "n=7"} {
		if !strings.Contains(cool, want) {
			t.Errorf("cooldown line %q missing %q", cool, want)
		}
	}
	if strings.Contains(cut, "SKIPPED") || strings.Contains(cool, "SKIPPED") {
		t.Fatal("WARN-first lines must never claim the wake was skipped")
	}
}

// minutesToSessionFlat drives F1 — pinned against the real registry windows,
// including today's two would-be skips and the ASIA midnight wrap.
func TestClass47MinutesToFlat(t *testing.T) {
	reg := kernel.DefaultSessionRegistry()
	get := func(name string) *kernel.SessionDef {
		for i := range reg.Sessions {
			if reg.Sessions[i].Name == name {
				return &reg.Sessions[i]
			}
		}
		t.Fatalf("session %s missing", name)
		return nil
	}
	ny, london, asia := get("NY"), get("LONDON"), get("ASIA")

	// Today's NY wake at 14:20:29 — 25 min to the 14:45 flat: AT the cutoff.
	if m, ok := minutesToSessionFlat(ctTime(t, 2026, 9, 2, 14, 20), ny); !ok || m != 25 {
		t.Fatalf("NY 14:20 → %d min (ok=%v), want 25", m, ok)
	}
	// Today's LONDON wake at 08:12:30 — 17 min to the 08:30 flat: inside it.
	if m, ok := minutesToSessionFlat(ctTime(t, 2026, 9, 2, 8, 12), london); !ok || m != 18 {
		t.Fatalf("LONDON 08:12 → %d min (ok=%v), want 18", m, ok)
	}
	// Mid-window: nowhere near.
	if m, ok := minutesToSessionFlat(ctTime(t, 2026, 9, 2, 10, 0), ny); !ok || m != 285 {
		t.Fatalf("NY 10:00 → %d min (ok=%v), want 285", m, ok)
	}
	// ASIA wraps midnight: 17:00→02:00, so 01:30 is 30 min out.
	if m, ok := minutesToSessionFlat(ctTime(t, 2026, 9, 2, 1, 30), asia); !ok || m != 30 {
		t.Fatalf("ASIA 01:30 → %d min (ok=%v), want 30", m, ok)
	}
	// Outside the window → ok=false, never an invented deadline.
	if _, ok := minutesToSessionFlat(ctTime(t, 2026, 9, 2, 15, 30), ny); ok {
		t.Fatal("outside the NY window must return ok=false")
	}
}

// ── F3 — a WAKE defers on any open stream; a SCHEDULED read never does ──────

func TestClass47CrossSessionClaimDefersWakesOnly(t *testing.T) {
	// Nothing open.
	if _, open := anyPlannerStreamOpen(); open {
		t.Fatal("fixture: no stream should be open at the start")
	}
	// Today's overlap: LONDON's read is in flight when NY's wake arrives.
	londonKey := store.MakePlanIDForTrader("t1", "2026-09-02", "LONDON")
	if !claimPlannerRead(londonKey) {
		t.Fatal("fixture: could not claim")
	}
	t.Cleanup(func() { releasePlannerRead(londonKey) })

	held, open := anyPlannerStreamOpen()
	if !open || held != londonKey {
		t.Fatalf("an open LONDON stream must be visible globally, got %q open=%v", held, open)
	}
	// The claim is per (trader, date, session): NY's own key is still FREE —
	// which is exactly why the two streamed concurrently today at 08:01.
	nyKey := store.MakePlanIDForTrader("t1", "2026-09-02", "NY")
	if nyKey == londonKey {
		t.Fatal("fixture: the two session keys must differ")
	}
	if !claimPlannerRead(nyKey) {
		t.Fatal("the per-session claim does NOT block a different session — the overlap this wave defers wakes on")
	}
	releasePlannerRead(nyKey)

	line := wakeStreamDeferLine("NY", "fresh FVG", londonKey)
	for _, want := range []string{"⏱ wake DEFERRED", "a planner stream is already open", "Scheduled reads never defer", "class 47"} {
		if !strings.Contains(line, want) {
			t.Errorf("defer line %q missing %q", line, want)
		}
	}
}

// ── F5 — the boot line reads every value from its resolver (A24) ────────────

func TestClass47BootLineReadsResolvers(t *testing.T) {
	t.Setenv("WAKE_CUTOFF_MIN", "17")
	t.Setenv("WAKE_COOLDOWN_MIN", "41")
	line := WakeCadenceBootLine()
	for _, want := range []string{"cutoff=17m", "cooldown=41m", "cross-session=on", "stale-arm-expiry=on", "class 47"} {
		if !strings.Contains(line, want) {
			t.Fatalf("boot line %q missing %q — every field must be READ, never a literal", line, want)
		}
	}
}

// ── counters: recorded, per session, malformed reads as 0 ──────────────────

func TestClass47CountersRecorded(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "c47c.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	const tid, date = "t1", "2026-09-02"
	for i := 1; i <= 3; i++ {
		if n, err := store.IncWakeCounter(st, tid, date, "NY", store.WakeWouldSkipCutoffKind); err != nil || n != i {
			t.Fatalf("cutoff bump %d → %d (%v)", i, n, err)
		}
	}
	if _, err := store.IncWakeCounter(st, tid, date, "NY", store.WakeWouldSkipCooldownKind); err != nil {
		t.Fatal(err)
	}
	if _, err := store.IncWakeCounter(st, tid, date, "LONDON", store.WakeWouldSkipCutoffKind); err != nil {
		t.Fatal(err)
	}
	if got := store.WakeCounterCount(st, tid, date, "NY", store.WakeWouldSkipCutoffKind); got != 3 {
		t.Fatalf("NY cutoff = %d, want 3", got)
	}
	if got := store.WakeCounterCount(st, tid, date, "NY", store.WakeWouldSkipCooldownKind); got != 1 {
		t.Fatalf("kinds must be separable, got %d", got)
	}
	if got := store.WakeCounterCount(st, tid, date, "LONDON", store.WakeWouldSkipCutoffKind); got != 1 {
		t.Fatalf("sessions must not bleed, got %d", got)
	}
	_ = st.SetSystemConfig(store.WakeCounterKey(tid, date, "ASIA", store.WakeWouldSkipCutoffKind), "not-a-number")
	if got := store.WakeCounterCount(st, tid, date, "ASIA", store.WakeWouldSkipCutoffKind); got != 0 {
		t.Fatalf("malformed must read 0, got %d", got)
	}
}
