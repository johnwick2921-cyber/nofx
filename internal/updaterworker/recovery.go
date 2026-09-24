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
	p("The installation hold is KEPT (%s): every new entry stays refused until step 4.", hold)
	oldSHA, newSHA := "n/a", "n/a"
	if j.Install != nil {
		oldSHA = j.Install.SHA
	}
	if j.Release != nil {
		newSHA = j.Release.SHA
	}
	p("Pre-update build: %s   release being installed: %s", oldSHA, newSHA)
	p("")
	step := 1
	if j.Snapshot != nil && j.Install != nil {
		s, in := *j.Snapshot, *j.Install
		p("%d. Restore the pre-update install from this job's snapshot (copy to a temp name, then mv -f over the live one):", step)
		p("     cp -p %s %s.recovery.tmp && mv -f %s.recovery.tmp %s", s.Binary, in.Binary, in.Binary, in.Binary)
		p("     cp -p %s %s.recovery.tmp && mv -f %s.recovery.tmp %s", s.ReleaseFile, in.ReleaseFile, in.ReleaseFile, in.ReleaseFile)
		p("     rm -rf %s.recovery.tmp && cp -a %s %s.recovery.tmp && mv %s %s.failed.%s && mv %s.recovery.tmp %s",
			in.Dist, s.Dist, in.Dist, in.Dist, in.Dist, j.JobID, in.Dist, in.Dist)
	} else {
		p("%d. No snapshot was taken (the job stopped before backup_done): nothing was installed — skip to step %d.", step, step+1)
	}
	step++
	p("%d. Restart the bot by its identity, never by name (systemd's Restart=on-failure relaunches it):", step)
	p(`     kill -9 "$(systemctl show -p MainPID --value nofx)"`)
	step++
	short := oldSHA
	if len(short) > 12 {
		short = short[:12]
	}
	p("%d. Prove the boot: the data log must show the OK line for the pre-update build, and health must serve it:", step)
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
