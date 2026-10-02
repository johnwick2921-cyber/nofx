package updaterworker

// WORKER-TAKES-THE-LOCK (owner order 10-02 12:1x CT, CTO fold 12:40 CT) — the
// worker acquires the main-tree lock itself right before preflight, as session
// "updater-<job id first 12>", with the SAME atomic acquire humans use. It
// never reclaims, never takes a held or stale lock, never clears-incomplete;
// ANY held lock refuses naming the holder (an attended install is just a button
// install — humans never hold the lock across one); the lock is released at
// complete/rolled_back/refused and KEPT + named in the job on recovery_needed;
// a restart mid-job recognises its own lock and never double-acquires.
//
// These tests pin the PRODUCTION CALL SITES: ensureMainTreeLock inside
// stepPreflight (driven by the real runner through the rig), settleMainTreeLock
// inside finished(), and OSHost against the REAL deploy/vl-lock.sh in a
// sandbox HOME.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vl/internal/updaterjob"
)

// The free-lock path: preflight ACQUIRES the lock as our deterministic
// session, and a later refusal releases it again (the terminal edge).
func TestWorkerAcquiresAFreeLockAndReleasesOnRefused(t *testing.T) {
	r := newRig(t)
	r.flat = false // refuse AFTER the lock step: exercises acquire + the refused release edge
	r.install()
	if err := r.drive(); err != nil {
		t.Fatal(err)
	}
	j := r.job()
	if j.State != updaterjob.StateRefused {
		t.Fatalf("job %s: want refused, got %s", j.State, states(j))
	}
	r.mu.Lock()
	acq, sessions, rel := r.box.lockAcquireCalls, append([]string(nil), r.box.lockAcquireSessions...), r.box.lockReleaseCalls
	holder := r.box.lockHolder
	r.mu.Unlock()
	if acq != 1 || len(sessions) != 1 || sessions[0] != lockSessionFor(boxJobID) {
		t.Fatalf("preflight must acquire exactly once as %q (calls %d, sessions %v)",
			lockSessionFor(boxJobID), acq, sessions)
	}
	if rel != 1 || holder != "" {
		t.Fatalf("the refused terminal edge must release our lock (releases %d, holder %q)", rel, holder)
	}
}

// The held-by-anyone path: preflight REFUSES naming the holder and NEVER
// calls acquire (the CTO's mutant 'acquire when held' must fail here).
func TestWorkerRefusesAHeldLockNamingTheHolder(t *testing.T) {
	r := newRig(t)
	r.lockHolder = "other-lane"
	r.install()
	if err := r.drive(); err != nil {
		t.Fatal(err)
	}
	j := r.job()
	if j.State != updaterjob.StateRefused {
		t.Fatalf("job %s: want refused, got %s", j.State, states(j))
	}
	if !strings.Contains(j.Error, `held by "other-lane"`) {
		t.Fatalf("the refusal must name the holder, got %q", j.Error)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.box.lockAcquireCalls != 0 {
		t.Fatal("a held lock must never be acquired")
	}
	if r.box.lockHolder != "other-lane" {
		t.Fatalf("the stranger's lock must stay untouched (holder %q)", r.box.lockHolder)
	}
}

// A worker restart mid-job finds its own lock by the deterministic session
// name and continues — never a second acquire.
func TestWorkerRestartFindsItsOwnLockWithoutDoubleAcquire(t *testing.T) {
	r := newRig(t)
	r.lockHolder = lockSessionFor(boxJobID) // the prior run acquired it
	r.flat = false
	r.install()
	if err := r.drive(); err != nil {
		t.Fatal(err)
	}
	j := r.job()
	if j.State != updaterjob.StateRefused {
		t.Fatalf("job %s: want refused, got %s", j.State, states(j))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.box.lockAcquireCalls != 0 {
		t.Fatal("a restart must recognise its own lock, never double-acquire")
	}
}

// The complete edge releases the lock (the CTO's mutant 'skip release on
// complete' must fail here).
func TestWorkerReleasesOnComplete(t *testing.T) {
	r := newRig(t)
	r.install()
	if err := r.drive(); err != nil {
		t.Fatal(err)
	}
	j := r.job()
	if j.State != updaterjob.StateComplete {
		t.Fatalf("job %s: want complete, got %s", j.State, states(j))
	}
	r.mu.Lock()
	rel, holder := r.box.lockReleaseCalls, r.box.lockHolder
	r.mu.Unlock()
	if rel != 1 || holder != "" {
		t.Fatalf("complete must release our lock exactly once (releases %d, holder %q)", rel, holder)
	}
	// the job names the release
	found := false
	for _, rc := range j.Receipts {
		if rc.Step == "main_tree_lock" && rc.Evidence["outcome"] == "released" {
			found = true
		}
	}
	if !found {
		t.Fatal("the job must record the main_tree_lock release receipt")
	}
}

// recovery_needed KEEPS the lock and names it in the job (the CTO's mutant
// 'release on recovery_needed' must fail here).
func TestWorkerKeepsTheLockOnRecoveryNeeded(t *testing.T) {
	r := newRig(t)
	r.watchFail[boxNew], r.rollbackFail = true, true // watch fails -> rollback fails -> recovery_needed
	r.install()
	if err := r.drive(); err != nil {
		t.Fatal(err)
	}
	j := r.job()
	if j.State != updaterjob.StateRecoveryNeeded {
		t.Fatalf("job %s: want recovery_needed, got %s", j.State, states(j))
	}
	r.mu.Lock()
	rel, holder := r.box.lockReleaseCalls, r.box.lockHolder
	r.mu.Unlock()
	if rel != 0 {
		t.Fatalf("recovery_needed must KEEP the lock (releases %d)", rel)
	}
	if holder != lockSessionFor(boxJobID) {
		t.Fatalf("the lock must still be held by our session (holder %q)", holder)
	}
	kept := false
	for _, rc := range j.Receipts {
		if rc.Step == "main_tree_lock" && rc.Evidence["outcome"] == "kept" &&
			rc.Evidence["session"] == lockSessionFor(boxJobID) {
			kept = true
		}
	}
	if !kept {
		t.Fatal("the job must NAME the kept lock (session + a human looks first)")
	}
}

// OSHost against the REAL deploy/vl-lock.sh in a sandbox HOME: the same
// atomic acquire/release verbs humans use, parsed exactly as production does.
func TestOSHostLockVerbsAgainstTheRealScript(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	script := filepath.Join(t.TempDir(), "vl-lock.sh")
	src, err := os.ReadFile(filepath.Join("..", "..", "deploy", "vl-lock.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, src, 0o755); err != nil {
		t.Fatal(err)
	}
	h := OSHost{LockScript: script}

	// free -> ours
	acq, holder, err := h.LockAcquire("updater-test-1", "worker job test", 1)
	if err != nil || !acq || holder != "" {
		t.Fatalf("acquire on a free lock: acquired=%v holder=%q err=%v", acq, holder, err)
	}
	if got, err := h.LockHolder(); err != nil || got != "updater-test-1" {
		t.Fatalf("holder round-trip: %q, %v", got, err)
	}

	// a stranger's acquire is refused naming us
	acq, holder, err = h.LockAcquire("other-lane", "squatter", 1)
	if err != nil || acq || holder != "updater-test-1" {
		t.Fatalf("a held lock must refuse naming the holder: acquired=%v holder=%q err=%v", acq, holder, err)
	}

	// only the holder may release; a stranger's release refuses
	if err := h.LockRelease("other-lane"); err == nil {
		t.Fatal("a stranger must not release our lock")
	}
	if err := h.LockRelease("updater-test-1"); err != nil {
		t.Fatalf("the holder's release: %v", err)
	}
	if got, err := h.LockHolder(); err != nil || got != "" {
		t.Fatalf("after release the lock must be free: %q, %v", got, err)
	}
	// idempotent: a second release on a free lock succeeds (the script's no-op)
	if err := h.LockRelease("updater-test-1"); err != nil {
		t.Fatalf("a no-lock release must be a no-op success: %v", err)
	}
}
