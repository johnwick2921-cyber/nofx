package updaterjob

import (
	"errors"
	"os"
	"testing"
	"time"
)

// refusedAtBothCallSites: a crafted job is refused by the production writer
// (Write) AND, planted on disk in the writer's exact byte form (so only the
// content rules can refuse it), by the production reader (Read).
func refusedAtBothCallSites(t *testing.T, name string, j Job) {
	t.Helper()
	if err := Write(t.TempDir(), j); !errors.Is(err, ErrCorrupt) {
		t.Errorf("%s: Write = %v, want ErrCorrupt", name, err)
	}
	dd := t.TempDir()
	seed := mustNew(t, j.JobID, j.ReleaseID, j.CreatedAt)
	mustWrite(t, dd, seed)
	p, _ := Path(dd, j.JobID)
	b, err := encode(j)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := Read(dd, j.JobID); !errors.Is(err, ErrCorrupt) {
		t.Errorf("%s: Read = %s/%s, %v — want ErrCorrupt", name, got.State, got.Phase, err)
	}
}

// crafted builds a job directly (no Enter/Finish), in the shape Read returns.
func crafted(state State, phase Phase, attempts int, tr []Transition, rc []Receipt) Job {
	last := tr[len(tr)-1].At
	return Job{Schema: SchemaVersion, JobID: "job-0100", ReleaseID: "v1.2.0", State: state, Phase: phase, Attempts: attempts,
		CreatedAt: t0, UpdatedAt: last, Transitions: tr, Receipts: rc}
}

// TestDoneStepCarriesItsReceipt (U1 verifier defect 2, probe H3): a step with
// a side effect becomes DONE only with a receipt appended since its STARTED
// transition. Each transition RECORDS the receipt count it was written with
// (canon 35: counters record, never infer), so the rule survives a Read:
//
//   - Finish (the runner's call) refuses when no receipt was added since the
//     state started — a receipt from an earlier step (the park's resume
//     receipt) does not count — and leaves the job unchanged;
//   - Validate (every Write and every Read) refuses a history whose DONE
//     transition records no more receipts than its STARTED one, whose
//     counts run backwards (two finished steps sharing one receipt), or
//     whose counts claim receipts the file does not hold.
func TestDoneStepCarriesItsReceipt(t *testing.T) {
	restoreSeams(t)
	// H3 at the production call site: Enter, Write, Finish with no receipt.
	dd := t.TempDir()
	now := t0
	j := mustNew(t, "job-0100", "v1.2.0", now)
	mustWrite(t, dd, j)
	now = now.Add(time.Second)
	if err := j.Enter(StateDownloaded, now); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, dd, j)
	before := len(j.Transitions)
	if err := j.Finish(now.Add(time.Second)); !errors.Is(err, ErrNoReceipt) {
		t.Errorf("Finish(downloaded/started, no receipt) = %v, want ErrNoReceipt", err)
	}
	if j.Phase != PhaseStarted || len(j.Transitions) != before {
		t.Errorf("a refused Finish changed the job: %s, %d transitions", j.Phase, len(j.Transitions))
	}
	// A receipt from an EARLIER step does not finish this one: the park's
	// resume receipt is on file before activated starts.
	p := t.TempDir()
	pn := t0
	pj := mustNew(t, "job-0101", "v1.2.0", pn)
	mustWrite(t, p, pj)
	for _, s := range []State{StateDownloaded, StateVerified, StatePreflightOK, StateMaintenanceHeld, StateDrainedAcked, StateGateOK, StateBackupDone} {
		step(t, p, &pj, &pn, s)
	}
	pj.NT8 = &NT8Decision{Decision: NT8Updated, Reason: "ninjascript/*.cs changed"}
	step(t, p, &pj, &pn, StateNT8Updated)
	pn = pn.Add(time.Minute)
	resumed := pn
	pj.ResumedAt = &resumed
	if err := pj.AddReceipt(Receipt{Step: "resume", StartedAt: pn, EndedAt: pn, OK: true}, pn); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, p, pj)
	withRollbackInputs(&pj)
	pn = pn.Add(time.Second)
	if err := pj.Enter(StateActivated, pn); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, p, pj)
	if err := pj.Finish(pn.Add(time.Second)); !errors.Is(err, ErrNoReceipt) {
		t.Errorf("Finish(activated/started) on the resume receipt alone = %v, want ErrNoReceipt", err)
	}
	// positive control: the activate receipt finishes it, and the file reads.
	if err := pj.AddReceipt(Receipt{Step: "activate", StartedAt: pn, EndedAt: pn.Add(time.Second), OK: true}, pn.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := pj.Finish(pn.Add(time.Second)); err != nil {
		t.Fatalf("Finish with its receipt: %v", err)
	}
	mustWrite(t, p, pj)
	if got, err := Read(p, pj.JobID); err != nil || got.State != StateActivated || got.Phase != PhaseDone {
		t.Fatalf("positive control: %s/%s %v", got.State, got.Phase, err)
	}

	// Crafted histories (Validate at Write and at Read).
	r := func(step string) Receipt { return Receipt{Step: step, StartedAt: t0, EndedAt: t0, OK: true} }
	req := Transition{State: StateRequested, Phase: PhaseDone, At: t0}
	at := func(n int) time.Time { return t0.Add(time.Duration(n) * time.Second) }
	refusedAtBothCallSites(t, "H3: downloaded done with no receipt since it started", crafted(StateDownloaded, PhaseDone, 1,
		[]Transition{req, {State: StateDownloaded, Phase: PhaseStarted, At: at(1)}, {State: StateDownloaded, Phase: PhaseDone, At: at(2)}},
		[]Receipt{}))
	refusedAtBothCallSites(t, "a done count claiming a receipt the file does not hold", crafted(StateDownloaded, PhaseDone, 1,
		[]Transition{req, {State: StateDownloaded, Phase: PhaseStarted, At: at(1)}, {State: StateDownloaded, Phase: PhaseDone, At: at(2), Receipts: 1}},
		[]Receipt{}))
	refusedAtBothCallSites(t, "two finished steps sharing one receipt (counts run backwards)", crafted(StateVerified, PhaseDone, 1,
		[]Transition{req, {State: StateDownloaded, Phase: PhaseStarted, At: at(1)}, {State: StateDownloaded, Phase: PhaseDone, At: at(2), Receipts: 1},
			{State: StateVerified, Phase: PhaseStarted, At: at(3)}, {State: StateVerified, Phase: PhaseDone, At: at(4), Receipts: 1}},
		[]Receipt{r("download")}))
	// control: the same shapes with honest counts read back.
	ok := crafted(StateVerified, PhaseDone, 1,
		[]Transition{req, {State: StateDownloaded, Phase: PhaseStarted, At: at(1)}, {State: StateDownloaded, Phase: PhaseDone, At: at(2), Receipts: 1},
			{State: StateVerified, Phase: PhaseStarted, At: at(3), Receipts: 1}, {State: StateVerified, Phase: PhaseDone, At: at(4), Receipts: 2}},
		[]Receipt{r("download"), r("verify")})
	if err := ok.Validate(); err != nil {
		t.Fatalf("control: honest counts refused: %v", err)
	}
}

// withRollbackInputs sets what activated requires (the rollback inputs) with
// SHAs that agree: release = source_sha, install = snapshot = the old build.
func withRollbackInputs(j *Job) {
	sha := "0123456789abcdef0123456789abcdef01234567"
	old := "89abcdef0123456789abcdef0123456789abcdef"
	j.SourceSHA = sha
	j.Release = &Release{Dir: "/r/" + sha, SHA: sha, Binary: "/r/" + sha + "/nofx-bin", Dist: "/r/" + sha + "/web/dist", ReleaseFile: "/r/" + sha + "/RELEASE", ManifestPath: "/r/" + sha + "/manifest.json"}
	j.Install = &Release{Dir: "/i", SHA: old, Binary: "/i/nofx-bin", Dist: "/i/web/dist", ReleaseFile: "/i/deploy/RELEASE"}
	j.Snapshot = &Release{Dir: "/b/install", SHA: old, Binary: "/b/install/nofx-bin", Dist: "/b/install/web/dist", ReleaseFile: "/b/install/deploy/RELEASE"}
	j.BackupPath = "/b/data.db"
	j.IdentityBefore = &Identity{PID: 172, StartTicks: 23987}
}

// walkTo steps a fresh job through states with the production API (step =
// Enter, Write, receipt, Finish, Write), setting the AddOn decision each nt8
// state requires.
func walkTo(t *testing.T, dd, jobID string, states ...State) (Job, time.Time) {
	t.Helper()
	now := t0
	j := mustNew(t, jobID, "v1.2.0", now)
	mustWrite(t, dd, j)
	for _, s := range states {
		switch s {
		case StateNT8Skipped:
			j.NT8 = &NT8Decision{Decision: NT8Skipped}
		case StateNT8Updated:
			j.NT8 = &NT8Decision{Decision: NT8Updated, Reason: "ninjascript/*.cs changed"}
		}
		step(t, dd, &j, &now, s)
	}
	return j, now
}

var toPark = []State{StateDownloaded, StateVerified, StatePreflightOK, StateMaintenanceHeld, StateDrainedAcked, StateGateOK, StateBackupDone, StateNT8Updated}

// TestLeavingTheParkNeedsAnAttendedResume (U1 verifier defect 3, probes H20
// and H18): the job leaves the nt8_updated park for activated only with a
// resumed_at recorded AFTER the park was done and no later than the move
// out of it (the attended `nofx-updater resume` is the only way out); and a
// resumed_at on a job that never parked is refused.
func TestLeavingTheParkNeedsAnAttendedResume(t *testing.T) {
	restoreSeams(t)
	sha := "0123456789abcdef0123456789abcdef01234567"
	leave := func(t *testing.T, resumed *time.Time) (Job, string, []byte) {
		t.Helper()
		dd := t.TempDir()
		j, now := walkTo(t, dd, "job-0110", toPark...)
		withRollbackInputs(&j)
		p, _ := Path(dd, j.JobID)
		before, _ := os.ReadFile(p)
		j.ResumedAt = resumed
		if err := j.Enter(StateActivated, now.Add(time.Minute)); err != nil {
			t.Fatal(err)
		}
		return j, dd, before
	}
	parkDone := t0.Add(16 * time.Second) // walkTo: 8 steps × 2 s
	for name, resumed := range map[string]*time.Time{
		"H20: no resumed_at":             nil,
		"resumed_at at the park instant": ptr(parkDone),
		"resumed_at before the park":     ptr(parkDone.Add(-time.Second)),
		"resumed_at after it left":       ptr(parkDone.Add(2 * time.Minute)),
	} {
		t.Run(name, func(t *testing.T) {
			j, dd, before := leave(t, resumed)
			if err := Write(dd, j); !errors.Is(err, ErrCorrupt) {
				t.Errorf("Write(nt8_updated → activated, %s) = %v, want ErrCorrupt", name, err)
			}
			p, _ := Path(dd, j.JobID)
			if b, _ := os.ReadFile(p); string(b) != string(before) {
				t.Error("a refused leave changed the file")
			}
			refusedAtBothCallSites(t, name, j)
		})
	}
	// H18: a resume time on a job that never parked — mid-flight, and on the
	// nt8_skipped path.
	for name, states := range map[string][]State{
		"H18: downloaded":      {StateDownloaded},
		"nt8_skipped, no park": {StateDownloaded, StateVerified, StatePreflightOK, StateMaintenanceHeld, StateDrainedAcked, StateGateOK, StateBackupDone, StateNT8Skipped},
		"nt8_updated not done": {StateDownloaded, StateVerified, StatePreflightOK, StateMaintenanceHeld, StateDrainedAcked, StateGateOK, StateBackupDone},
	} {
		t.Run(name, func(t *testing.T) {
			dd := t.TempDir()
			j, now := walkTo(t, dd, "job-0111", states...)
			if name == "nt8_updated not done" {
				j.NT8 = &NT8Decision{Decision: NT8Updated}
				if err := j.Enter(StateNT8Updated, now.Add(time.Second)); err != nil {
					t.Fatal(err)
				}
				mustWrite(t, dd, j)
			}
			j.ResumedAt = ptr(now.Add(time.Minute))
			j.UpdatedAt = *j.ResumedAt
			if err := Write(dd, j); !errors.Is(err, ErrCorrupt) {
				t.Errorf("Write(resumed_at, %s) = %v, want ErrCorrupt", name, err)
			}
			refusedAtBothCallSites(t, name, j)
		})
	}
	// positive control: resumed inside the park window, then activated.
	j, dd, _ := leave(t, ptr(parkDone.Add(30*time.Second)))
	mustWrite(t, dd, j)
	if got, err := Read(dd, j.JobID); err != nil || got.State != StateActivated || got.ResumedAt == nil || got.Release.SHA != sha {
		t.Fatalf("positive control: %+v %v", got, err)
	}
	// and the park itself may carry the resume before the runner moves
	k := t.TempDir()
	pj, now := walkTo(t, k, "job-0112", toPark...)
	pj.ResumedAt = ptr(now.Add(time.Second))
	pj.UpdatedAt = *pj.ResumedAt
	mustWrite(t, k, pj)
}

func ptr[T any](v T) *T { return &v }
