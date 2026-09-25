package updaterworker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

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
	for _, want := range []string{`systemctl show -p MainPID --value nofx`, j.Snapshot.Binary, "BOOT INTEGRITY OK — rev " + boxOld[:12],
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

// PIN (verifier D3): the recovery text's restart never becomes "kill -9 0"
// (MainPID is 0 for a unit that is not running — exactly when recovery is
// needed — and kill -9 0 signals the operator's whole process group). The
// restart line is RUN here under sh with a stub systemctl and a kill
// function that only records: MainPID 0 and an empty MainPID kill nothing;
// a live MainPID is killed exactly once, by pid.
//
// U4 re-verify note 6: running the REAL line must never be able to signal a
// real process, even if the line drifts. So the "live" MainPID is 4194305 —
// above Linux's pid_max ceiling (4194304), a pid no process can hold — a
// recording stub `kill` EXECUTABLE sits first in PATH (it catches an
// `env kill` / `exec kill` drift the shell function cannot), and the line
// may name neither `command kill` (bypasses the function for the builtin)
// nor `/bin/kill` / `/usr/bin/kill` (bypasses both).
func TestRecoveryRestartNeverKillsTheProcessGroup(t *testing.T) {
	r := newRig(t)
	r.watchFail[boxNew] = true
	r.rollbackFail = true
	j := r.runToEnd(t)
	text := RecoveryText(j, r.cfg.Target)
	var line string
	for _, l := range strings.Split(text, "\n") {
		if strings.Contains(l, "systemctl show -p MainPID --value nofx") {
			line = strings.TrimSpace(l)
		}
	}
	if line == "" {
		t.Fatalf("no restart line:\n%s", text)
	}
	for _, bypass := range []string{"command kill", "/bin/kill", "/usr/bin/kill", "env kill", "exec kill"} {
		if strings.Contains(line, bypass) {
			t.Fatalf("the restart line names %q, which bypasses the recording kill in this test:\n%s", bypass, line)
		}
	}
	const impossiblePID = "4194305" // > pid_max's ceiling (4194304): no process can hold it
	for _, tc := range []struct{ mainPID, want string }{{"0", ""}, {"", ""}, {"1", ""}, {impossiblePID, "KILL -9 " + impossiblePID}} {
		stub := t.TempDir()
		for name, body := range map[string]string{
			"systemctl": "#!/bin/sh\necho '" + tc.mainPID + "'\n",
			"kill":      "#!/bin/sh\necho \"KILL-EXE $*\"\n", // records; never signals
		} {
			writeFile(t, filepath.Join(stub, name), body)
			if err := os.Chmod(filepath.Join(stub, name), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command("/bin/sh", "-c", `kill() { echo "KILL $*"; }; `+line)
		cmd.Env = []string{"PATH=" + stub + ":/usr/bin:/bin"}
		out, _ := cmd.CombinedOutput()
		var kills []string
		for _, l := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(l, "KILL-EXE") {
				t.Fatalf("MainPID %q: the restart line ran a kill EXECUTABLE from PATH (%q), not the shell's kill\nline: %s", tc.mainPID, l, line)
			}
			if strings.HasPrefix(l, "KILL") {
				kills = append(kills, l)
			}
		}
		got := strings.Join(kills, ";")
		if got != tc.want {
			t.Fatalf("MainPID %q: the restart line ran %q, want %q\nline: %s\noutput: %s", tc.mainPID, got, tc.want, line, out)
		}
	}
}

// PIN (verifier D3, low): the steps follow what the job PROVED. A job that
// never reached the activate installed nothing; one whose release was proven
// (boot_verified done) must never be rolled back from the snapshot; only an
// unproven activate restores it. Each case is a real job file (a stale crash
// swept to recovery_needed at start, or a failed rollback).
func TestRecoveryTextFollowsWhatTheJobProved(t *testing.T) {
	staleAt := func(t *testing.T, point string) (*rig, updaterjob.Job) {
		r := newRig(t)
		r.w.crash = func(q string) {
			if q == point {
				panic(crashPanic{q})
			}
		}
		if !r.runCrashing(t) {
			t.Fatalf("no crash at %s", point)
		}
		r.clock.Advance(31 * time.Minute)
		r.w = r.newWorker()
		if _, err := r.w.sweep(); err != nil {
			t.Fatal(err)
		}
		j := r.job()
		if j.State != updaterjob.StateRecoveryNeeded {
			t.Fatalf("stale job at %s is %s", point, j.State)
		}
		return r, j
	}
	restore := func(j updaterjob.Job) string { return "cp -p " + j.Snapshot.Binary }
	for _, tc := range []struct {
		name        string
		build       func(t *testing.T) (*rig, updaterjob.Job)
		wantRestore bool
		prove       string
		say         string
	}{
		{"never activated", func(t *testing.T) (*rig, updaterjob.Job) { return staleAt(t, "nt8_skipped/done") }, false, boxOld, "Nothing was installed"},
		{"release proven", func(t *testing.T) (*rig, updaterjob.Job) { return staleAt(t, "boot_verified/done") }, false, boxNew, "Do NOT restore the snapshot"},
		{"release proven, complete started", func(t *testing.T) (*rig, updaterjob.Job) { return staleAt(t, "complete/started") }, false, boxNew, "Do NOT restore the snapshot"},
		{"activate unproven", func(t *testing.T) (*rig, updaterjob.Job) { return staleAt(t, "activated/effect") }, true, boxOld, "Restore the pre-update install"},
		{"rollback failed", func(t *testing.T) (*rig, updaterjob.Job) {
			r := newRig(t)
			r.watchFail[boxNew] = true
			r.rollbackFail = true
			return r, r.runToEnd(t)
		}, true, boxOld, "Restore the pre-update install"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, j := tc.build(t)
			if j.Snapshot == nil {
				t.Fatalf("no snapshot recorded (history %v)", states(j))
			}
			text := RecoveryText(j, r.cfg.Target)
			if got := strings.Contains(text, restore(j)); got != tc.wantRestore {
				t.Fatalf("restore-from-snapshot in the text = %v, want %v (history %v):\n%s", got, tc.wantRestore, states(j), text)
			}
			if !strings.Contains(text, "BOOT INTEGRITY OK — rev "+tc.prove[:12]) || !strings.Contains(text, "revision must be "+tc.prove[:12]) {
				t.Fatalf("the text does not prove %s:\n%s", tc.prove[:12], text)
			}
			if !strings.Contains(text, tc.say) {
				t.Fatalf("the text lacks %q:\n%s", tc.say, text)
			}
			if strings.Index(text, "clear --job") < strings.Index(text, "MainPID") {
				t.Fatalf("the hold is not cleared LAST:\n%s", text)
			}
		})
	}
}

// PIN (verifier D6): rolling_back entered by a failure edge runs at once as
// its FIRST attempt — never counted as a crash retry. So a rollback that
// crashes inside its effect twice still gets its third run (the attempts cap
// is about runs of the step, and RollbackTo is called exactly once per run),
// and the job reads rolled_back.
func TestAFailureEdgeIntoRollingBackIsItsFirstAttempt(t *testing.T) {
	r := newRig(t)
	r.watchFail[boxNew] = true
	crashes := 0
	r.w.crash = func(q string) {
		if q == "rolling_back/effect" && crashes < updaterjob.MaxAttempts-1 {
			crashes++
			panic(crashPanic{q})
		}
	}
	if !r.runCrashing(t) {
		t.Fatal("no crash in the rollback")
	}
	for crashes < updaterjob.MaxAttempts-1 {
		w := r.newWorker()
		w.crash = r.w.crash
		r.w = w
		if _, err := r.w.sweep(); err != nil {
			t.Fatal(err)
		}
		r.runCrashing(t)
	}
	j := r.restart(t)
	r.noViolations(t)
	want := make([]int, updaterjob.MaxAttempts)
	for i := range want {
		want[i] = i + 1
	}
	if fmt.Sprint(r.rollbackAtt) != fmt.Sprint(want) {
		t.Fatalf("RollbackTo ran at attempts %v, want %v (the failure edge is attempt 1)", r.rollbackAtt, want)
	}
	if j.State != updaterjob.StateRolledBack {
		t.Fatalf("job %s (reason %q), want rolled_back after %d crashed runs", j.State, j.RecoveryReason, crashes)
	}
}
