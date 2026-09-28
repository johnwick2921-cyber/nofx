package updaterworker

// UPDATER-NT8-CLOSED — the nt8_absent drain path, at the production call
// sites: the REAL runner, job file, hold file and loopback app (the harness).
// NT8 closed = the safest moment to update. The drain then passes on the
// gate's nt8_absent verdict (every ledger leg on its own evidence + the SIM
// predicate), records the path and each leg's evidence, and the job file says
// "nt8_absent (link down since <ts>)". A reconnect revokes the verdict —
// the normal ack/census legs apply again from that moment, never
// grandfathered.

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nofx/internal/updaterjob"
)

const absentDown = 90 * time.Second

func withNT8Down(d time.Duration) rigOpt {
	return func(b *box) {
		b.addonDownSince = d
		b.addonConnected = false
	}
}

// atDrained drives to drained_acked/done (held + drained).
func atDrained(t *testing.T, opts ...rigOpt) (*rig, updaterjob.Job) {
	t.Helper()
	r := newRig(t, opts...)
	r.w.crash = func(q string) {
		if q == "drained_acked/done" {
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("never reached drained_acked/done")
	}
	r.w.crash = nil
	j := r.job()
	if j.State != updaterjob.StateDrainedAcked || j.Phase != updaterjob.PhaseDone {
		t.Fatalf("at %s/%s", j.State, j.Phase)
	}
	return r, j
}

func mustReceipt(t *testing.T, j updaterjob.Job, step string) Receipt {
	t.Helper()
	for _, rc := range j.Receipts {
		if rc.Step == step {
			return rc
		}
	}
	t.Fatalf("no %q receipt", step)
	return Receipt{}
}

func TestDrainNt8AbsentPassesAndRecordsPathAndLegs(t *testing.T) {
	r, j := atDrained(t, withNT8Down(absentDown))
	if j.DrainPath != drainPathNT8Absent {
		t.Fatalf("drain_path=%q", j.DrainPath)
	}
	if j.LinkDownSince == "" {
		t.Fatal("link_down_since not recorded")
	}
	rc := mustReceipt(t, j, "drain")
	if !rc.OK {
		t.Fatalf("drain receipt failed: %s", rc.Err)
	}
	if rc.Evidence["drain_path"] != drainPathNT8Absent {
		t.Fatalf("receipt drain_path=%q", rc.Evidence["drain_path"])
	}
	for _, leg := range []string{"hold", "go_drained", "in_flight_sends", "queued_signals", "planner_in_flight", "ledger_exposure", "sim_accounts"} {
		if v := rc.Evidence["leg_"+leg]; !strings.HasPrefix(v, "PASS: ") {
			t.Fatalf("leg_%s evidence=%q — the receipt records each leg's evidence", leg, v)
		}
	}
	r.noViolations(t)
}

func TestDrainNt8AbsentLinkDownBelowWindowRefuses(t *testing.T) {
	r := newRig(t, withNT8Down(30*time.Second))
	if r.runCrashing(t) {
		t.Fatal("drain passed with the link down only 30 s")
	}
	j := r.job()
	// the absent verdict is not eligible → the normal path applies → the ack
	// blocker names the real reason and the drain ends in recovery (the hold
	// is kept, nothing installed)
	if j.State != updaterjob.StateRecoveryNeeded {
		t.Fatalf("state=%s, want recovery_needed", j.State)
	}
	if !strings.Contains(j.Error, "addon_ack") {
		t.Fatalf("error %q must name the missing ack", j.Error)
	}
}

func TestDrainNt8AbsentNonSIMAccountRefuses(t *testing.T) {
	r := newRig(t, withNT8Down(absentDown), func(b *box) { b.absentSim = false })
	if r.runCrashing(t) {
		t.Fatal("drain passed although the bound account is not SIM-tradeable")
	}
	j := r.job()
	if j.State != updaterjob.StateRecoveryNeeded {
		t.Fatalf("state=%s", j.State)
	}
	if !strings.Contains(j.Error, "sim_accounts") {
		t.Fatalf("error %q must name the sim_accounts leg", j.Error)
	}
}

// A reconnect mid-DRAIN revokes the verdict: the normal ack/census path
// applies from that moment, never grandfathered.
func TestDrainNt8AbsentReconnectMidDrainRevokes(t *testing.T) {
	r := newRig(t, withNT8Down(absentDown), func(b *box) { b.absentQueued = 1 })
	// crash the instant the drain STARTS (a queued signal keeps the absent
	// verdict not-ready), then the AddOn reconnects
	r.w.crash = func(q string) {
		if q == "drained_acked/started" {
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("drain never started")
	}
	r.w.crash = nil
	r.set(func() { r.addonDownSince, r.addonConnected, r.absentQueued = 0, true, 0 })
	r.w.crash = func(q string) {
		if q == "drained_acked/done" {
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("the normal drain never passed after the reconnect")
	}
	j := r.job()
	rc := mustReceipt(t, j, "drain")
	if rc.Evidence["drain_path"] != "normal" {
		t.Fatalf("the reconnect must revoke the absent path; drain_path=%q", rc.Evidence["drain_path"])
	}
	if j.DrainPath != "" {
		t.Fatalf("a normal drain must not stamp drain_path (got %q)", j.DrainPath)
	}
}

// A reconnect AFTER the drain (at the gate) revokes too: the gate runs the
// normal two-ack path instead of the absent fast path.
func TestGateAfterAbsentDrainReconnectRunsTheNormalAckPath(t *testing.T) {
	r, j := atDrained(t, withNT8Down(absentDown))
	if j.DrainPath != drainPathNT8Absent {
		t.Fatalf("drain_path=%q", j.DrainPath)
	}
	r.set(func() { r.addonDownSince, r.addonConnected = 0, true })
	r.w.crash = func(q string) {
		if q == "gate_ok/done" {
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("the gate never passed on the normal ack path")
	}
	j = r.job()
	rc := mustReceipt(t, j, "gate")
	if _, ok := rc.Evidence["gate_path"]; ok {
		t.Fatalf("the gate must not take the absent fast path after a reconnect: %+v", rc.Evidence)
	}
	if _, ok := rc.Evidence["ack_1_age_ms"]; !ok {
		t.Fatalf("the normal gate must record ack ages: %+v", rc.Evidence)
	}
}

func TestDecideNT8AbsentNoCSChangeSkips(t *testing.T) {
	r, j := atBackupDone(t, withNT8Down(absentDown))
	r.w.crash = func(q string) {
		if q == "nt8_skipped/done" {
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("never reached nt8_skipped/done")
	}
	j = r.job()
	if j.NT8 == nil || j.NT8.Decision != updaterjob.NT8Skipped || !j.NT8.Absent {
		t.Fatalf("nt8 decision %+v", j.NT8)
	}
	if j.NT8.F5Owed {
		t.Fatalf("C# unchanged — no F5 is owed: %+v", j.NT8)
	}
}

// With a .cs change the job does NOT park (NT8 is closed — nobody to F5 now):
// the Go side completes and the owed F5 is RECORDED.
func TestDecideNT8AbsentCSChangeRecordsF5OwedAndDoesNotPark(t *testing.T) {
	r, _ := atBackupDone(t, withNT8Down(absentDown))
	writeFile(r.t, filepath.Join(r.relDir, "ninjascript", "VLTrader.cs"), "// C# v2\n")
	r.w.crash = func(q string) {
		if q == "nt8_skipped/done" {
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("never reached nt8_skipped/done (must not park)")
	}
	j := r.job()
	if j.NT8 == nil || j.NT8.Decision != updaterjob.NT8Skipped {
		t.Fatalf("nt8 decision %+v — the absent path must not park", j.NT8)
	}
	if !j.NT8.Absent || !j.NT8.F5Owed {
		t.Fatalf("nt8 decision %+v — the owed F5 must be recorded", j.NT8)
	}
	if !strings.Contains(j.NT8.Reason, "AddOn F5 owed at next NT8 start") {
		t.Fatalf("reason %q", j.NT8.Reason)
	}
}

// A reconnect before the decision revokes the absent mode: the normal
// decision (which reads the ack) applies again.
func TestDecideNT8AbsentReconnectedUsesTheNormalDecision(t *testing.T) {
	r, _ := atBackupDone(t, withNT8Down(absentDown))
	r.set(func() { r.addonDownSince, r.addonConnected = 0, true })
	r.w.crash = func(q string) {
		if q == "nt8_skipped/done" {
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("never reached nt8_skipped/done")
	}
	j := r.job()
	if j.NT8 == nil || j.NT8.Absent {
		t.Fatalf("the reconnect must revoke the absent decision: %+v", j.NT8)
	}
	if j.NT8.Decision != updaterjob.NT8Skipped {
		t.Fatalf("nt8 decision %+v", j.NT8)
	}
}

// The full absent run completes, and boot_verify SKIPS the AddOn-ack wait
// with that fact recorded (never faked).
func TestNt8AbsentFullRunCompletesAndBootVerifySkipsTheAckWait(t *testing.T) {
	r := newRig(t, withNT8Down(absentDown))
	j := r.runToEnd(t)
	r.noViolations(t)
	if j.State != updaterjob.StateComplete {
		t.Fatalf("job %s\n%v", j.State, states(j))
	}
	if j.DrainPath != drainPathNT8Absent {
		t.Fatalf("drain_path=%q", j.DrainPath)
	}
	rc := mustReceipt(t, j, "boot_verify")
	if rc.Evidence["addon_ack_wait"] != "skipped (drain path nt8_absent)" {
		t.Fatalf("boot_verify must record the skipped ack wait: %+v", rc.Evidence)
	}
	if _, ok := rc.Evidence["acked_build_id"]; ok {
		t.Fatalf("no ack comparison may be faked in absent mode: %+v", rc.Evidence)
	}
}
