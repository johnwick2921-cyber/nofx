package updaterworker

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"nofx/internal/updaterjob"
)

// PIN (dispatch §4(4)): a Watch that never proves the new build rolls back to
// the job's SNAPSHOT of the install (RollbackTo(snapshot, install, id)), then
// watches the OLD sha with since = the persisted rollback kill instant; the
// job reads rolled_back, the install's halves are the old ones again, and
// the hold is cleared.
func TestWatchTimeoutRollsBackToTheSnapshotAndReadsRolledBack(t *testing.T) {
	r := newRig(t)
	r.watchFail[boxNew] = true
	j := r.runToEnd(t)
	r.noViolations(t)
	if j.State != updaterjob.StateRolledBack || j.Phase != updaterjob.PhaseDone {
		t.Fatalf("job %s/%s (error %q), want rolled_back/done\n%v", j.State, j.Phase, j.Error, states(j))
	}
	steps := receiptSteps(j)
	tail := strings.Join(steps[len(steps)-6:], ",")
	if tail != "activate,watch(fail),rollback,watch,boot_verify,release_hold" {
		t.Fatalf("receipts end %s\nwant activate,watch(fail),rollback,watch,boot_verify,release_hold", tail)
	}
	if len(r.rollbackArg) != 1 {
		t.Fatalf("rollback ran %d times", len(r.rollbackArg))
	}
	prev, inst := r.rollbackArg[0][0], r.rollbackArg[0][1]
	wantSnap := filepath.Join(r.backupRoot, boxJobID, "install")
	if prev != *j.Snapshot || prev.Dir != wantSnap || prev.SHA != boxOld || prev.ReleaseFile != filepath.Join(wantSnap, "RELEASE") {
		t.Fatalf("RollbackTo restored from %+v, want the job's snapshot at %s (103 layout, sha %s)", prev, wantSnap, boxOld)
	}
	if inst != *j.Install || inst.Dir != r.inst {
		t.Fatalf("RollbackTo restored INTO %+v, want the install %s", inst, r.inst)
	}
	if got := r.rollbackIDs[0]; got != *j.IdentityRollback {
		t.Fatalf("RollbackTo killed %v, want the persisted identity_rollback %v", got, *j.IdentityRollback)
	}
	if len(r.watchSHAs) != 2 || r.watchSHAs[0] != boxNew || r.watchSHAs[1] != boxOld {
		t.Fatalf("Watch ran on %v, want [new, OLD]", r.watchSHAs)
	}
	if got := r.watchOpts[1]; !got.Since.Equal(*j.RollbackWatchSince) || got.LogPath != j.RollbackLogPath {
		t.Fatalf("the rollback's Watch since=%v log=%s, want the persisted kill instant %v and %s", got.Since, got.LogPath, *j.RollbackWatchSince, j.RollbackLogPath)
	}
	if sha, _ := parseBinaryBody(filepath.Join(r.inst, "nofx-bin")); sha != boxOld {
		t.Fatalf("the install binary is %s after the rollback, want %s", sha, boxOld)
	}
	if b, _ := os.ReadFile(filepath.Join(r.inst, "web", "dist", "index.html")); string(b) != "<html>old</html>\n" {
		t.Fatalf("the install dist is %q after the rollback", b)
	}
	if r.hold().Present {
		t.Fatalf("hold survives rolled_back: %+v", r.hold())
	}
}

// PIN (dispatch §4(5)): a rollback that fails is recovery_needed — the worker
// STOPS (no step runs again, install is refused until restart), the job names
// its last GOOD receipt, the hold is KEPT, and the recovery text is this
// job's steps: kill by MainPID only, never pgrep / journalctl / the
// version-control tool, the hold cleared last.
func TestRollbackFailureIsRecoveryNeededAndStops(t *testing.T) {
	r := newRig(t)
	r.watchFail[boxNew] = true
	r.rollbackFail = true
	j := r.runToEnd(t)
	r.noViolations(t)
	if j.State != updaterjob.StateRecoveryNeeded {
		t.Fatalf("job %s (error %q), want recovery_needed\n%v", j.State, j.Error, states(j))
	}
	if j.LastGoodReceipt == nil || !j.Receipts[*j.LastGoodReceipt].OK {
		t.Fatalf("last_good_receipt %v does not name an OK receipt (%v)", j.LastGoodReceipt, receiptSteps(j))
	}
	if got := j.Receipts[*j.LastGoodReceipt].Step; got != "activate" {
		t.Fatalf("last good receipt is %q, want activate (the last step that succeeded)", got)
	}
	if !strings.Contains(j.RecoveryReason, "rollback failed") {
		t.Fatalf("recovery_reason %q", j.RecoveryReason)
	}
	if s, _ := ReadHoldFor(r.data, boxJobID); s != HoldOurs {
		t.Fatalf("the hold is %s after a failed rollback — it must be KEPT", s)
	}
	calls := len(r.calls)
	if err := r.drive(); err != nil {
		t.Fatal(err)
	}
	if len(r.calls) != calls {
		t.Fatalf("the worker kept going after recovery_needed: %v", r.calls[calls:])
	}
	if resp := r.w.Handle(updaterwireInstallOf("job-u4-0003abcd")); resp.OK || resp.Error != "recovery needed" {
		t.Fatalf("install after recovery_needed = %+v", resp)
	}
	if resp := r.w.Handle(updaterwireStatusOf("")); resp.State != StatusStopped {
		t.Fatalf("status = %+v", resp)
	}
	text := RecoveryText(j, r.cfg.Target)
	for _, bad := range []*regexp.Regexp{regexp.MustCompile(`pgrep`), regexp.MustCompile(`journalctl`), regexp.MustCompile(`\bgit\b`)} {
		if bad.MatchString(text) {
			t.Fatalf("recovery text uses %s:\n%s", bad, text)
		}
	}
	for _, want := range []string{`kill -9 "$(systemctl show -p MainPID --value nofx)"`, j.Snapshot.Binary, "BOOT INTEGRITY OK — rev " + boxOld[:12],
		"maintenance-hold --install-dir " + r.inst + " clear --job " + boxJobID, j.BackupPath} {
		if !strings.Contains(text, want) {
			t.Fatalf("recovery text lacks %q:\n%s", want, text)
		}
	}
	if strings.Index(text, "clear --job") < strings.Index(text, "kill -9") {
		t.Fatalf("the hold is not cleared LAST:\n%s", text)
	}
	t.Logf("recovery text:\n%s", text)
}
