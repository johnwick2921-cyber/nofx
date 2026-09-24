package updaterworker

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nofx/internal/updaterjob"
)

// PIN (D8b, CTO 1790279155144 (4) + 1790280128238): a fetch killed between
// its rename and its verdict link leaves <release_root>/<sha> with no
// verdict. That dir is NEVER activated (no verdict ⇒ the install verb refuses;
// a verdict gone mid-job ⇒ the downloaded re-proof refuses), and the next
// fetch's recovery QUARANTINES it (renamed, bytes kept — never deleted) so a
// clean re-fetch can take the name. A dir a verdict names is untouched; an
// unreadable verdict or a fetch in flight (the release-root lock) refuses the
// whole recovery.
func TestInterruptedFetchIsQuarantinedAndNeverActivated(t *testing.T) {
	r := newRig(t)
	root := filepath.Dir(r.relDir) // <tmp>/releases: the release dir exists, and no verdict names it
	r.verdictMissing = true        // the re-proof finds no verdict, as the real reader would

	// never activated: the install verb refuses; no job file is born
	if resp := r.w.Handle(updaterwireInstallOf(boxJobID)); resp.OK || resp.Error != "release not verified" {
		t.Fatalf("install of a verdict-less release = %+v", resp)
	}
	if _, err := updaterjob.Read(r.data, boxJobID); !errors.Is(err, updaterjob.ErrNotFound) {
		t.Fatalf("a refused install wrote a job: %v", err)
	}
	// a verdict that vanishes after the install: the job's own re-proof refuses
	r.verdictMissing = false
	r.install()
	r.verdictMissing = true
	if err := r.drive(); err != nil {
		t.Fatal(err)
	}
	if j := r.job(); j.State != updaterjob.StateRefused || r.callCount("activate") != 0 {
		t.Fatalf("a job whose verdict vanished ended %s (activate ran %d times)", j.State, r.callCount("activate"))
	}

	// a verified neighbour (a verdict names it), a staging leftover, a symlink
	other := filepath.Join(root, strings.Repeat("e5", 20))
	writeFile(t, filepath.Join(other, "nofx-bin"), binaryBody(strings.Repeat("e5", 20)))
	writeFile(t, filepath.Join(r.data, "updater", "verdicts", "v1.1.0.json"), `{"schema":1,"release_id":"v1.1.0","release_dir":"`+other+`"}`)
	writeFile(t, filepath.Join(root, ".fetch-v1.3.0-123", "x"), "staging")
	os.Symlink(other, filepath.Join(root, strings.Repeat("f6", 20)))

	// a fetch in flight holds the lock: refused, nothing moved
	unlock, err := LockReleaseRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	if moved, err := QuarantineInterruptedFetches(root, r.data, r.clock.Now()); !errors.Is(err, ErrFetchInFlight) || len(moved) != 0 {
		t.Fatalf("recovery during a fetch = %v, %v", moved, err)
	}
	unlock()
	// an unreadable verdict: refused, nothing moved
	bad := filepath.Join(r.data, "updater", "verdicts", "v0.9.0.json")
	writeFile(t, bad, `{"release_id":"v0.9.0"}`)
	if moved, err := QuarantineInterruptedFetches(root, r.data, r.clock.Now()); err == nil || len(moved) != 0 {
		t.Fatalf("recovery with an unreadable verdict = %v, %v", moved, err)
	}
	os.Remove(bad)

	moved, err := QuarantineInterruptedFetches(root, r.data, r.clock.Now())
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, ".orphan-"+boxNew+"-")
	if len(moved) != 1 || !strings.HasPrefix(moved[0], want) {
		t.Fatalf("quarantined %v, want exactly the verdict-less %s", moved, r.relDir)
	}
	if _, err := os.Stat(r.relDir); !os.IsNotExist(err) {
		t.Fatalf("the verdict-less dir still blocks the name: %v", err)
	}
	if sha, _ := parseBinaryBody(filepath.Join(moved[0], "nofx-bin")); sha != boxNew {
		t.Fatal("the quarantined release lost its bytes")
	}
	for _, keep := range []string{other, filepath.Join(root, ".fetch-v1.3.0-123"), filepath.Join(root, strings.Repeat("f6", 20))} {
		if _, err := os.Lstat(keep); err != nil {
			t.Fatalf("%s was touched: %v", keep, err)
		}
	}
	// idempotent: nothing more to move
	if again, err := QuarantineInterruptedFetches(root, r.data, r.clock.Now()); err != nil || len(again) != 0 {
		t.Fatalf("second recovery = %v, %v", again, err)
	}
}
