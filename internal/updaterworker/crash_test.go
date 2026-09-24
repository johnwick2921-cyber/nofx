package updaterworker

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"nofx/internal/updaterjob"
)

// crashPanic is a crash played at a boundary: the runner stops mid-flight and
// writes nothing more (the runner has no deferred writes).
type crashPanic struct{ point string }

// runCrashing installs (once) and drives, resuming a park once; it reports
// whether the crash seam fired.
func (r *rig) runCrashing(t *testing.T) (crashed bool) {
	t.Helper()
	defer func() {
		if v := recover(); v != nil {
			if _, ok := v.(crashPanic); !ok {
				panic(v)
			}
			crashed = true
		}
	}()
	if _, err := updaterjob.Read(r.data, boxJobID); err != nil {
		r.install()
	}
	if err := r.drive(); err != nil {
		t.Fatalf("drive: %v", err)
	}
	if j := r.job(); j.State == updaterjob.StateNT8Updated && j.Phase == updaterjob.PhaseDone {
		r.f5()
		if resp := r.w.Handle(resumeRequest(boxJobID)); !resp.OK {
			t.Fatalf("resume: %+v", resp)
		}
		if err := r.drive(); err != nil {
			t.Fatalf("drive after resume: %v", err)
		}
	}
	return false
}

// restart is a FRESH worker over the same dirs: the start sweep, then the
// runner's drive of whatever it found (and the attended resume of a park).
func (r *rig) restart(t *testing.T) updaterjob.Job {
	t.Helper()
	r.w = r.newWorker()
	rep, err := r.w.sweep()
	if err != nil {
		t.Fatalf("start sweep: %v", err)
	}
	if rep.Active != "" {
		if rep.Active != boxJobID {
			t.Fatalf("the sweep found job %q", rep.Active)
		}
		r.runCrashing(t)
	}
	return r.job()
}

// holdHeldBetween: from maintenance_held DONE until complete/rolled_back
// STARTED, this job's hold must be on disk (a crash never drops it).
func holdMustBeHeld(j updaterjob.Job) bool {
	switch j.State {
	case updaterjob.StateMaintenanceHeld:
		return j.Phase == updaterjob.PhaseDone
	case updaterjob.StateDrainedAcked, updaterjob.StateGateOK, updaterjob.StateBackupDone, updaterjob.StateNT8Skipped,
		updaterjob.StateNT8Updated, updaterjob.StateActivated, updaterjob.StateBooted, updaterjob.StateBootVerified,
		updaterjob.StateRollingBack:
		return true
	}
	return false
}

// PIN (dispatch §4(3)): crash at EVERY boundary — after a state is persisted
// started, after its side effect, after it is persisted done — for both AddOn
// branches and the rollback path. A FRESH worker from the same dirs resumes
// to the right terminal state, re-running at most the one step it crashed in,
// and the hold reads held on disk at every point from maintenance_held done
// through complete (checked at the crash, and by every fake side effect after
// the hold write).
func TestCrashAtEveryBoundaryResumesToTheRightState(t *testing.T) {
	type path struct {
		name  string
		opts  []rigOpt
		setup func(r *rig)
		final updaterjob.State
	}
	for _, p := range []path{
		{name: "nt8_skipped", final: updaterjob.StateComplete},
		{name: "nt8_updated", opts: []rigOpt{withCSChanged()}, final: updaterjob.StateComplete},
		{name: "rollback", setup: func(r *rig) { r.watchFail[boxNew] = true }, final: updaterjob.StateRolledBack},
	} {
		// the boundaries a clean run passes, in order (with repeats numbered)
		clean := newRig(t, p.opts...)
		if p.setup != nil {
			p.setup(clean)
		}
		var points []string
		seen := map[string]int{}
		clean.w.crash = func(pt string) {
			seen[pt]++
			points = append(points, fmt.Sprintf("%s#%d", pt, seen[pt]))
		}
		clean.runCrashing(t)
		if j := clean.job(); j.State != p.final {
			t.Fatalf("%s: the clean run ended %s", p.name, j.State)
		}
		if len(points) < 30 {
			t.Fatalf("%s: only %d boundaries on the clean run: %v", p.name, len(points), points)
		}
		for _, point := range points {
			pt, occ := point, 1
			if i := strings.LastIndex(point, "#"); i > 0 {
				pt = point[:i]
				fmt.Sscan(point[i+1:], &occ)
			}
			t.Run(p.name+"/"+point, func(t *testing.T) {
				r := newRig(t, p.opts...)
				if p.setup != nil {
					p.setup(r)
				}
				hit := 0
				r.w.crash = func(q string) {
					if q == pt {
						hit++
						if hit == occ {
							panic(crashPanic{q})
						}
					}
				}
				if !r.runCrashing(t) {
					t.Fatalf("the crash at %s never fired", point)
				}
				at := r.job()
				if holdMustBeHeld(at) {
					if s, _ := ReadHoldFor(r.data, boxJobID); s != HoldOurs {
						t.Fatalf("crashed at %s with the job at %s/%s: the hold is %s, want this job's", point, at.State, at.Phase, s)
					}
				}
				j := r.restart(t)
				r.noViolations(t)
				if j.State != p.final || j.Phase != updaterjob.PhaseDone {
					t.Fatalf("crash at %s (job at %s/%s): the fresh worker ended %s/%s (error %q, blocker %q)\nhistory %v",
						point, at.State, at.Phase, j.State, j.Phase, j.Error, j.Blocker, states(j))
				}
				if n := r.callCount("activate"); n > 2 || (n == 2 && pt != "activated/effect") {
					t.Fatalf("crash at %s: activate ran %d times", point, n)
				}
				if n := r.callCount("rollback"); n > 2 {
					t.Fatalf("crash at %s: rollback ran %d times", point, n)
				}
				if r.hold().Present {
					t.Fatalf("crash at %s: the hold survives the finished job", point)
				}
				for _, tr := range j.Transitions {
					if tr.State == updaterjob.StateRecoveryNeeded {
						t.Fatalf("crash at %s went through recovery_needed", point)
					}
				}
			})
		}
	}
}

// PIN (C3 as ruled): a crash INSIDE activated — after the kill, before its
// receipt — resumes by re-reading the CURRENT identity and re-running
// Activate against it (all three halves reinstalled, one extra restart),
// never against the identity the first run already killed. The re-read
// identity and a new kill instant are persisted before the re-run.
func TestResumeAtActivatedRereadsTheIdentity(t *testing.T) {
	r := newRig(t)
	r.w.crash = func(q string) {
		if q == "activated/effect" {
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("no crash")
	}
	killed := *r.job().IdentityBefore
	r.mu.Lock()
	current := r.id
	r.mu.Unlock()
	if current == killed {
		t.Fatal("the first activate did not restart the process")
	}
	firstSince := *r.job().WatchSince
	r.clock.Advance(time.Minute) // the worker comes back a minute later
	j := r.restart(t)
	r.noViolations(t)
	if len(r.activateIDs) != 2 || r.activateIDs[0] != killed || r.activateIDs[1] != current {
		t.Fatalf("activate ran against %v, want [%v (the first run) %v (the CURRENT identity)]", r.activateIDs, killed, current)
	}
	if j.State != updaterjob.StateComplete {
		t.Fatalf("resumed job ended %s (error %q)", j.State, j.Error)
	}
	if *j.IdentityBefore != current || !j.WatchSince.After(firstSince) {
		t.Fatalf("identity_before %v / watch_since %v not re-read before the re-run (current %v, first since %v)", *j.IdentityBefore, *j.WatchSince, current, firstSince)
	}
	if got := r.watchOpts[0].Since; !got.Equal(*j.WatchSince) {
		t.Fatalf("Watch since %v, want the persisted (second) kill instant %v", got, *j.WatchSince)
	}
}

// PIN (C21 as ruled): at worker start, an unfinished job older than 30
// minutes is marked recovery_needed — never resumed. Its hold is untouched
// (the operator clears it, attended), nothing on the box runs, and install is
// refused until the worker restarts (OQ-3).
func TestStaleJobAtStartIsRecoveryNeededNeverResumed(t *testing.T) {
	r := newRig(t)
	r.w.crash = func(q string) {
		if q == "gate_ok/started" {
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("no crash")
	}
	callsAtCrash := len(r.calls)
	r.clock.Advance(31 * time.Minute)
	r.w = r.newWorker()
	rep, err := r.w.sweep()
	if err != nil {
		t.Fatal(err)
	}
	if rep.Active != "" || len(rep.StaleAtStart) != 1 || rep.StaleAtStart[0] != boxJobID {
		t.Fatalf("start sweep = %+v, want the job stale and nothing active", rep)
	}
	j := r.job()
	if j.State != updaterjob.StateRecoveryNeeded || !strings.Contains(j.RecoveryReason, "stale at start") {
		t.Fatalf("stale job is %s (reason %q)", j.State, j.RecoveryReason)
	}
	if s, _ := ReadHoldFor(r.data, boxJobID); s != HoldOurs {
		t.Fatalf("the stale job's hold is %s — recovery_needed never touches the hold", s)
	}
	if err := r.w.drive(t.Context(), boxJobID); err != nil {
		t.Fatal(err)
	}
	if len(r.calls) != callsAtCrash {
		t.Fatalf("a stale job was resumed: calls %v", r.calls[callsAtCrash:])
	}
	if resp := r.w.Handle(updaterwireInstallOf("job-u4-0002abcd")); resp.OK || resp.Error != "recovery needed" {
		t.Fatalf("install after recovery_needed = %+v, want refused until restart", resp)
	}
	if resp := r.w.Handle(updaterwireStatusOf("")); resp.State != StatusStopped {
		t.Fatalf("status = %+v, want stopped", resp)
	}

	// the same rule at the resume verb: a park older than 30 minutes
	p := newRig(t, withCSChanged())
	p.install()
	if err := p.drive(); err != nil {
		t.Fatal(err)
	}
	if j := p.job(); j.State != updaterjob.StateNT8Updated {
		t.Fatalf("no park: %s", j.State)
	}
	p.clock.Advance(31 * time.Minute)
	if resp := p.w.Handle(resumeRequest(boxJobID)); resp.OK || resp.Error != "stale job recovery needed" {
		t.Fatalf("resume of a stale park = %+v", resp)
	}
	if j := p.job(); j.State != updaterjob.StateRecoveryNeeded || p.callCount("activate") != 0 {
		t.Fatalf("stale park: %s, activate ran %d times", j.State, p.callCount("activate"))
	}
}
