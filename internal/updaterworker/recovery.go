package updaterworker

import (
	"fmt"
	"strings"

	"nofx/internal/updaterjob"
)

// ── recovery_needed: the job-specific manual steps (C16 as ruled) ──────────
//
// deploy/RESTORE.md kills with pgrep, proves boots with journalctl and
// rebuilds from the source tree; none of that is safe here. These steps name
// THIS job's snapshot, backup and shas, restart by the unit's MainPID only
// (kill -9: SIGTERM exits 0 and systemd's Restart=on-failure would not
// relaunch), prove the boot from the data log + /api/health, and clear the
// hold LAST — until then every entry stays refused. RESTORE.md is cited only
// for its database-restore section. TestRollbackFailureIsRecoveryNeededAndStops
// pins the text free of pgrep, journalctl and the version-control tool.

// RecoveryText renders the manual steps for a recovery_needed job.
func RecoveryText(j updaterjob.Job, t Target) string {
	var b strings.Builder
	p := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }
	p("job %s — release %s — state %s", j.JobID, j.ReleaseID, j.State)
	p("reason: %s", orNA(j.RecoveryReason))
	if j.LastGoodReceipt != nil && *j.LastGoodReceipt < len(j.Receipts) {
		r := j.Receipts[*j.LastGoodReceipt]
		p("last good receipt: #%d %s (ended %s)", *j.LastGoodReceipt, r.Step, r.EndedAt.UTC().Format("2006-01-02T15:04:05Z"))
	} else {
		p("last good receipt: n/a")
	}
	if j.State != updaterjob.StateRecoveryNeeded {
		p("this job is not recovery_needed; nothing to do.")
		return b.String()
	}
	hold := HoldFileForDisplay(t.DataDir)
	p("")
	p("The installation hold is KEPT (%s): every new entry stays refused until the last step.", hold)
	oldSHA, newSHA := "n/a", "n/a"
	if j.Install != nil {
		oldSHA = j.Install.SHA
	}
	if j.Release != nil {
		newSHA = j.Release.SHA
	}
	p("Pre-update build: %s   release being installed: %s", oldSHA, newSHA)
	p("")
	// What is running decides the steps (verifier D3): a job that never
	// reached the activate installed nothing; a proven rollback or a proven
	// release must NOT be rolled back from the snapshot. Only an activate
	// that was never proven either way restores the snapshot.
	reach := recoveryReach(j)
	prove, restart := oldSHA, true
	step := 1
	switch reach {
	case reachNothingInstalled:
		p("%d. Nothing was installed: the job stopped before the activate, so the pre-update build %s still runs. Do NOT restore anything.", step, oldSHA)
		restart = false
	case reachRolledBack:
		p("%d. The rollback to the pre-update build %s was PROVEN (rolled_back reached). Do NOT restore anything.", step, oldSHA)
		restart = false
	case reachReleaseProven:
		p("%d. The release %s was PROVEN running (boot_verified finished). Do NOT restore the snapshot: that would roll back a proven release.", step, newSHA)
		prove, restart = newSHA, false
	default:
		if j.Snapshot != nil && j.Install != nil {
			s, in := *j.Snapshot, *j.Install
			p("%d. Restore the pre-update install from this job's snapshot (copy to a temp name, then mv -f over the live one):", step)
			p("     cp -p %s %s.recovery.tmp && mv -f %s.recovery.tmp %s", s.Binary, in.Binary, in.Binary, in.Binary)
			p("     cp -p %s %s.recovery.tmp && mv -f %s.recovery.tmp %s", s.ReleaseFile, in.ReleaseFile, in.ReleaseFile, in.ReleaseFile)
			// The failed dist is moved aside under a UNIQUE name (job id +
			// a nanosecond timestamp), never deleted — it is the evidence —
			// and mv -T refuses to move it INTO an existing directory, so a
			// repeat run can never nest the live dist in an earlier leftover
			// (U4 re-verify note 7, TestRecoveryRestoreIsSafeToRunTwice).
			p("     rm -rf %s.recovery.tmp && cp -a %s %s.recovery.tmp && mv -T %s %s.failed.%s.$(date +%%Y%%m%%dT%%H%%M%%S.%%N) && mv -T %s.recovery.tmp %s",
				in.Dist, s.Dist, in.Dist, in.Dist, in.Dist, j.JobID, in.Dist, in.Dist)
		} else {
			p("%d. The activate ran but this job recorded no snapshot of the install: the install cannot be restored from it — stop and restore it by hand.", step)
		}
	}
	step++
	short := prove
	if len(short) > 12 {
		short = short[:12]
	}
	if restart {
		p("%d. Restart the bot by its identity, never by name (systemd's Restart=on-failure relaunches it):", step)
	} else {
		p("%d. ONLY if health (step %d) does not serve %s: restart the bot by its identity, never by name:", step, step+1, short)
	}
	// MainPID is 0 for a unit that is not running, and kill -9 0 signals
	// the operator's whole process group: never kill a pid below 2.
	p("     %s", restartLine)
	step++
	p("%d. Prove the boot: the data log must show the OK line for %s, and health must serve it:", step, short)
	p(`     grep -a "BOOT INTEGRITY OK — rev %s ·" %s/nofx_$(date +%%F).log`, short, t.LogDir)
	p("     curl -s http://127.0.0.1:%d/api/health    (revision must be %s)", t.Port, short)
	if j.BackupPath != "" {
		p("   If the database itself must be restored, use this job's backup %s with deploy/RESTORE.md's database-restore section only.", j.BackupPath)
	}
	step++
	p("%d. LAST, once the boot is proven, clear this job's hold:", step)
	p("     maintenance-hold --install-dir %s clear --job %s", t.InstallDir, j.JobID)
	p("")
	p("Then restart nofx-updater (the restart is the acknowledgement; install stays refused until then).")
	return b.String()
}

// restartLine restarts the unit by its MainPID and refuses a pid below 2 (a
// stopped unit reports 0; kill -9 0 would signal the whole process group).
const restartLine = `pid=$(systemctl show -p MainPID --value nofx); if [ "$pid" -gt 1 ] 2>/dev/null; then kill -9 "$pid"; else echo "nofx is not running (MainPID '$pid'); start it: sudo systemctl start nofx"; fi`

type reach int

const (
	reachUnproven         reach = iota // the activate ran; neither build is proven
	reachNothingInstalled              // the job never entered activated
	reachRolledBack                    // rolled_back was entered: the old build is proven
	reachReleaseProven                 // boot_verified finished, no rollback began
)

// recoveryReach reads the job's HISTORY (never its current state alone: a
// recovery_needed job's state is recovery_needed).
func recoveryReach(j updaterjob.Job) reach {
	activated, rollingBack, rolledBack, proven := false, false, false, false
	for _, tr := range j.Transitions {
		switch {
		case tr.State == updaterjob.StateActivated:
			activated = true
		case tr.State == updaterjob.StateRollingBack:
			rollingBack = true
		case tr.State == updaterjob.StateRolledBack:
			rolledBack = true
		case tr.State == updaterjob.StateBootVerified && tr.Phase == updaterjob.PhaseDone,
			tr.State == updaterjob.StateComplete:
			proven = true
		}
	}
	switch {
	case !activated:
		return reachNothingInstalled
	case rolledBack:
		return reachRolledBack
	case proven && !rollingBack:
		return reachReleaseProven
	}
	return reachUnproven
}
