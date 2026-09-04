// TREE-GUARD tests.
//
// These live in Go, not in a standalone shell script, for one reason: a shell
// test nobody runs is not a test. `go test ./...` is what every lane runs before
// every merge, so the guard's contract is checked there or it is not checked.
//
// The contract under test is mostly ABOUT ABSENCE — the guard must alarm when
// something is missing and must never write. Both are easy to get wrong in a way
// that still looks like it works.

package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const guardScript = "nofx-tree-guard.sh"

// runGuard runs the script against a fixture and returns its combined output.
// Every path the guard touches is an override so a test can never reach the
// real deploy tree — a guard test that read /home/hoang/nofx would be a test
// that changes its answer depending on the machine's mood.
func runGuard(t *testing.T, env map[string]string) (string, int) {
	t.Helper()
	abs, err := filepath.Abs(guardScript)
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	if _, serr := os.Stat(abs); serr != nil {
		t.Fatalf("guard script missing at %s: %v", abs, serr)
	}
	cmd := exec.Command("bash", abs, "--once")
	cmd.Env = append(os.Environ(), "TREE_GUARD_TEST=1")
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	out, err := cmd.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run guard: %v (%s)", err, out)
	}
	return string(out), code
}

// fixtureRepo builds a throwaway git repo standing in for the deploy tree.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	if err := os.MkdirAll(filepath.Join(dir, "trader"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A file carrying the shipped symbols the canary looks for.
	body := "package trader\nfunc composeArmStop() {}\nfunc normalizeArmLegs() {}\nvar CorrectedPnL = 1\n"
	if err := os.WriteFile(filepath.Join(dir, "trader", "shipped.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "deploy"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "deploy", "RELEASE"), []byte("abc123\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-qm", "seed")
	return dir
}

func symbolsFile(t *testing.T, syms ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "symbols.txt")
	if err := os.WriteFile(p, []byte(strings.Join(syms, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func baseEnv(t *testing.T, tree string) map[string]string {
	t.Helper()
	return map[string]string{
		"TREE_GUARD_TREE": tree,
		// BOTH lock paths are overridden to absent temp paths. Overriding only
		// the legacy one let the guard fall through to the MACHINE'S REAL
		// ~/nofx-main.lock.d, so a verdict depended on whether another lane
		// happened to hold the lock while the suite ran. It did: a
		// 'rebrand-merge' lock appeared mid-run and silently downgraded an
		// expected ALARM to INFO. A test that reads ambient state nobody stated
		// is a test whose result is a coincidence.
		"TREE_GUARD_LOCK":         filepath.Join(t.TempDir(), "absent.lock"),
		"TREE_GUARD_LOCK_DIR":     filepath.Join(t.TempDir(), "absent.lock.d"),
		"TREE_GUARD_STATE":        filepath.Join(t.TempDir(), "state"),
		"TREE_GUARD_SYMBOLS":      symbolsFile(t, "composeArmStop", "normalizeArmLegs", "CorrectedPnL"),
		"TREE_GUARD_SKIP_REMOTE":  "1", // no origin in a fixture
		"TREE_GUARD_SKIP_RUNNING": "1",
	}
}

// ── E1: porcelain ───────────────────────────────────────────────────────

func TestE1DirtyTreeWithNoLockAlarmsAndNamesTheFile(t *testing.T) {
	tree := fixtureRepo(t)
	// The 08:46 signature: a file rewritten with nobody holding the lock.
	if err := os.WriteFile(filepath.Join(tree, "trader", "shipped.go"),
		[]byte("package trader\n// everything gone\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, code := runGuard(t, baseEnv(t, tree))
	if !strings.Contains(out, "ALARM") {
		t.Fatalf("a dirty tree with no lock must ALARM:\n%s", out)
	}
	if !strings.Contains(out, "trader/shipped.go") {
		t.Errorf("the alarm must NAME the file — an alarm you have to investigate to understand is half an alarm:\n%s", out)
	}
	if code == 0 {
		t.Errorf("exit code must be non-zero on ALARM, got 0")
	}
}

// SUPERSEDED AND MIGRATED, not weakened. This case originally used a legacy
// ~/nofx-main.lock file naming a LIVE pid and expected INFO. The lock changed
// under this wave (ec2dd8f7): it is now an atomic directory keyed by session
// with a heartbeat and NO pid, because kill -0 was the wrong liveness test — a
// pid died while its holder kept working. So the legacy FILE deliberately no
// longer confers liveness, and asserting that it does not is the point.
func TestLegacyLockFileNoLongerConfersLiveness(t *testing.T) {
	tree := fixtureRepo(t)
	if err := os.WriteFile(filepath.Join(tree, "trader", "shipped.go"), []byte("package trader\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := baseEnv(t, tree)
	lock := filepath.Join(t.TempDir(), "legacy.lock")
	// OUR pid — unambiguously alive. Under the old contract this suppressed the
	// alarm; under the new one it must not.
	body := fmt.Sprintf("owner=hoang session=test pid=%d task=cutover heartbeat=%s\n",
		os.Getpid(), time.Now().UTC().Format(time.RFC3339))
	if err := os.WriteFile(lock, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	env["TREE_GUARD_LOCK"] = lock
	env["TREE_GUARD_LOCK_DIR"] = filepath.Join(t.TempDir(), "absent.lock.d")

	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM porcelain") {
		t.Fatalf("a legacy pid-file must NOT suppress the alarm — that liveness test is exactly what the new lock removes:\n%s", out)
	}
	if !strings.Contains(out, "legacy") {
		t.Errorf("and its presence must still be surfaced:\n%s", out)
	}
}

// A lock whose pid is dead does NOT suppress the alarm. This is the trap that
// nearly cost a live holder their lock, inverted: a dead pid must not be able to
// silence the guard either.
func TestE1DirtyTreeUnderADeadPidLockStillAlarms(t *testing.T) {
	tree := fixtureRepo(t)
	if err := os.WriteFile(filepath.Join(tree, "trader", "shipped.go"), []byte("package trader\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := baseEnv(t, tree)
	lock := filepath.Join(t.TempDir(), "dead.lock")
	// pid 2^22-ish: valid syntax, certainly not running.
	if err := os.WriteFile(lock, []byte("owner=hoang session=ghost pid=4194301 task=cutover\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env["TREE_GUARD_LOCK"] = lock
	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM") {
		t.Fatalf("a dead-pid lock must NOT suppress the alarm:\n%s", out)
	}
}

// ── E2: the canary — the incident-specific pin ──────────────────────────

func TestE2CommittedRevertAlarmsWhileTheTreeIsClean(t *testing.T) {
	tree := fixtureRepo(t)
	// Remove a shipped symbol AND COMMIT it: porcelain now says clean, which is
	// exactly why check 1 alone would have missed the 08:46 class had it been
	// committed.
	if err := os.WriteFile(filepath.Join(tree, "trader", "shipped.go"),
		[]byte("package trader\nfunc normalizeArmLegs() {}\nvar CorrectedPnL = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-qm", "revert composeArmStop"}} {
		c := exec.Command("git", args...)
		c.Dir = tree
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
	out, code := runGuard(t, baseEnv(t, tree))
	if !strings.Contains(out, "PASS porcelain") {
		t.Errorf("the tree IS clean after the commit; porcelain must pass:\n%s", out)
	}
	if !strings.Contains(out, "ALARM canary") {
		t.Fatalf("a committed revert must ALARM the canary:\n%s", out)
	}
	if !strings.Contains(out, "composeArmStop") {
		t.Errorf("the canary must name the MISSING symbol:\n%s", out)
	}
	if code == 0 {
		t.Errorf("exit must be non-zero when the canary alarms")
	}
}

func TestE2AllSymbolsPresentPasses(t *testing.T) {
	tree := fixtureRepo(t)
	out, _ := runGuard(t, baseEnv(t, tree))
	if !strings.Contains(out, "PASS canary") {
		t.Fatalf("all symbols present must PASS:\n%s", out)
	}
}

// A symbols file that is missing or empty must ALARM, never silently pass. An
// empty canary list is a check that cannot fail.
func TestE2AnEmptySymbolListIsItselfAnAlarm(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	env["TREE_GUARD_SYMBOLS"] = filepath.Join(t.TempDir(), "nonexistent.txt")
	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM canary") {
		t.Fatalf("a missing symbol list must ALARM — an empty canary cannot fail:\n%s", out)
	}
}

// ── E3: RELEASE vs running ──────────────────────────────────────────────

func TestE3ReleaseMismatchAlarmsAndQuotesBoth(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	delete(env, "TREE_GUARD_SKIP_RUNNING")
	env["TREE_GUARD_RUNNING_REV"] = "def456"
	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM release") {
		t.Fatalf("RELEASE abc123 vs running def456 must ALARM:\n%s", out)
	}
	if !strings.Contains(out, "abc123") || !strings.Contains(out, "def456") {
		t.Errorf("the alarm must quote BOTH values, not just say mismatch:\n%s", out)
	}
}

func TestE3ReleaseMatchPasses(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	delete(env, "TREE_GUARD_SKIP_RUNNING")
	env["TREE_GUARD_RUNNING_REV"] = "abc123"
	out, _ := runGuard(t, env)
	if !strings.Contains(out, "PASS release") {
		t.Fatalf("equal revs must PASS:\n%s", out)
	}
}

// ── E5: the contract — the guard NEVER writes ───────────────────────────

func TestE5ScriptContainsNoWritingGitCommand(t *testing.T) {
	src, err := os.ReadFile(guardScript)
	if err != nil {
		t.Fatalf("read guard: %v", err)
	}
	s := string(src)
	// A31 is the whole point of the wave: a repairing guard is a second actor
	// mutating the deploy tree — the disease, not the cure.
	for _, forbidden := range []string{
		"git checkout", "git restore", "git reset", "git stash",
		"git clean", "git commit", "git push", "git pull", "git merge",
	} {
		if strings.Contains(s, forbidden) {
			t.Errorf("the guard must never run a writing git command, found %q", forbidden)
		}
	}
}

func TestE5ScriptDeclaresItsReadOnlyContract(t *testing.T) {
	src, err := os.ReadFile(guardScript)
	if err != nil {
		t.Fatalf("read guard: %v", err)
	}
	if !strings.Contains(strings.ToLower(string(src)), "never repair") {
		t.Errorf("the contract must be stated in the script itself, not only in a spec")
	}
}

// ── E4: staleness ───────────────────────────────────────────────────────
//
// The A19 class: three deployed waves lived only on disk for hours because the
// marker was never pushed. "Ahead of origin/dev" is that state, seen from the
// tree.

// fixtureWithRemote gives the fixture a real `origin` so ls-remote resolves.
func fixtureWithRemote(t *testing.T) (tree, origin string) {
	t.Helper()
	tree = fixtureRepo(t)
	origin = t.TempDir()
	gitIn := func(dir string, args ...string) {
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v in %s: %v (%s)", args, dir, err, out)
		}
	}
	gitIn(origin, "init", "-q", "--bare", "-b", "dev")
	gitIn(tree, "remote", "add", "origin", origin)
	gitIn(tree, "push", "-q", "origin", "main:dev")
	return tree, origin
}

func TestE4HeadAheadOfOriginDevAlarmsUnpushed(t *testing.T) {
	tree, _ := fixtureWithRemote(t)
	// One local commit that was never pushed — the unpushed-marker signature.
	if err := os.WriteFile(filepath.Join(tree, "deploy", "RELEASE"), []byte("newrev\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-qm", "deploy: marker"}} {
		c := exec.Command("git", args...)
		c.Dir = tree
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
	env := baseEnv(t, tree)
	delete(env, "TREE_GUARD_SKIP_REMOTE")
	out, code := runGuard(t, env)
	if !strings.Contains(out, "ALARM staleness") {
		t.Fatalf("an unpushed commit must ALARM:\n%s", out)
	}
	if !strings.Contains(out, "AHEAD") || !strings.Contains(out, "unpushed") {
		t.Errorf("the alarm must name the direction and the class:\n%s", out)
	}
	if code == 0 {
		t.Errorf("exit must be non-zero")
	}
}

func TestE4HeadEqualToOriginDevPasses(t *testing.T) {
	tree, _ := fixtureWithRemote(t)
	env := baseEnv(t, tree)
	delete(env, "TREE_GUARD_SKIP_REMOTE")
	out, _ := runGuard(t, env)
	if !strings.Contains(out, "PASS staleness") {
		t.Fatalf("HEAD == origin/dev must PASS:\n%s", out)
	}
}

// The guard must never FETCH: a fetch writes into the .git of the tree it is
// supposed only to observe.
func TestE4GuardNeverFetches(t *testing.T) {
	src, err := os.ReadFile(guardScript)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Contains(string(src), "git fetch") || strings.Contains(string(src), "git -C \"$TREE\" fetch") {
		t.Errorf("the guard must use ls-remote, never fetch — a fetch writes into the observed tree's .git")
	}
}

// ── E6: the unit files ──────────────────────────────────────────────────

func TestE6UnitPointsAtTheGuardAndTimerIsSixtySeconds(t *testing.T) {
	svc, err := os.ReadFile(filepath.Join("systemd-user", "nofx-tree-guard.service"))
	if err != nil {
		t.Fatalf("service unit: %v", err)
	}
	if !strings.Contains(string(svc), "nofx-tree-guard.sh --once") {
		t.Errorf("the unit must invoke the guard in --once mode:\n%s", svc)
	}
	tim, err := os.ReadFile(filepath.Join("systemd-user", "nofx-tree-guard.timer"))
	if err != nil {
		t.Fatalf("timer unit: %v", err)
	}
	if !strings.Contains(string(tim), "OnUnitActiveSec=60s") {
		t.Errorf("the timer must fire every 60s (the spec's discovery-window target):\n%s", tim)
	}
}

// The installer must not be the thing that dirties the tree it installs a guard
// for. It may only write under $HOME/.config and chmod the script.
func TestE6InstallerWritesNothingIntoTheTree(t *testing.T) {
	src, err := os.ReadFile("install-tree-guard.sh")
	if err != nil {
		t.Fatalf("installer: %v", err)
	}
	s := string(src)
	for _, forbidden := range []string{"git ", "> $REPO", ">$REPO", "rm -rf"} {
		if strings.Contains(s, forbidden) {
			t.Errorf("installer must not run %q — it would write into the observed tree", forbidden)
		}
	}
	if !strings.Contains(s, "UNIT_DIR") || !strings.Contains(s, ".config/systemd/user") {
		t.Errorf("installer must target the user unit dir")
	}
}

// The symbols file is the canary's vocabulary and the guard OWNS it. An empty or
// comment-only file would be a check that cannot fail.
func TestE6SymbolsFileHasRealEntries(t *testing.T) {
	src, err := os.ReadFile("tree-guard-symbols.txt")
	if err != nil {
		t.Fatalf("symbols: %v", err)
	}
	n := 0
	for _, ln := range strings.Split(string(src), "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		n++
	}
	if n < 5 {
		t.Fatalf("the canary needs real symbols, found %d", n)
	}
	// The four symbols the 2026-09-02 incident actually destroyed must be in it.
	for _, must := range []string{"composeArmStop", "normalizeArmLegs", "CorrectedPnL", "IncStopUnanchored"} {
		if !strings.Contains(string(src), must) {
			t.Errorf("the canary omits %q — one of the symbols the 08:46 Save-All deleted", must)
		}
	}
}

// ── THE LOCK CHANGED UNDER THIS WAVE ────────────────────────────────────
//
// I read the tree-guard spec at ~21:54 and built against its lock model:
// ~/nofx-main.lock, a file with a pid, liveness by kill -0. At 21:48:36 —
// SIX MINUTES BEFORE I READ IT — another lane landed ec2dd8f7, which replaced
// that model with an atomic directory ~/nofx-main.lock.d keyed by SESSION with a
// heartbeat, no pid at all, and edited this very spec to say so. My branch is
// based on a commit that predates the edit, so I read the superseded version.
//
// Left alone, the guard would have found no legacy lock file during a cutover
// under the new lock, concluded "no live holder", and ALARMED on a legitimately
// dirty tree — a false alarm at the exact moment the guard is supposed to be
// trusted. It would have kept running and kept printing the whole time.

func newLockDir(t *testing.T, session, task string, heartbeatAge time.Duration) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "nofx-main.lock.d")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	hb := time.Now().Add(-heartbeatAge)
	meta := fmt.Sprintf("session=%s\ntask=%s\nexpiry=%s\nheartbeat=%s\nheartbeat_epoch=%d\n",
		session, task, hb.Add(45*time.Minute).Format(time.RFC3339),
		hb.Format(time.RFC3339), hb.Unix())
	if err := os.WriteFile(filepath.Join(dir, "meta"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLockDirWithFreshHeartbeatDowngradesToInfo(t *testing.T) {
	tree := fixtureRepo(t)
	if err := os.WriteFile(filepath.Join(tree, "deploy", "RELEASE"), []byte("newrev\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := baseEnv(t, tree)
	env["TREE_GUARD_LOCK_DIR"] = newLockDir(t, "nofx-63", "merge+boot", 30*time.Second)

	out, _ := runGuard(t, env)
	if strings.Contains(out, "ALARM porcelain") {
		t.Fatalf("a dirty tree under a FRESH-heartbeat lock dir is INFO, not ALARM:\n%s", out)
	}
	if !strings.Contains(out, "nofx-63") {
		t.Errorf("the INFO line must name the holding SESSION (there is no pid any more):\n%s", out)
	}
}

// STALE, NEVER DEAD — and stale does not buy silence.
func TestLockDirWithStaleHeartbeatStillAlarms(t *testing.T) {
	tree := fixtureRepo(t)
	if err := os.WriteFile(filepath.Join(tree, "trader", "shipped.go"), []byte("package trader\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := baseEnv(t, tree)
	env["TREE_GUARD_LOCK_DIR"] = newLockDir(t, "nofx-ghost", "abandoned", 20*time.Minute)

	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM") {
		t.Fatalf("a heartbeat 20 minutes old must not suppress the alarm:\n%s", out)
	}
	if !strings.Contains(out, "STALE") {
		t.Errorf("the guard must say STALE — never DEAD; it does not get to declare a session dead:\n%s", out)
	}
}

// The transition's real hazard: a legacy pid-file lock lying around after the
// lock moved to a directory. It must be SURFACED, not silently honoured and not
// silently ignored.
func TestLegacyLockFileIsSurfacedAsAHazard(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	legacy := filepath.Join(t.TempDir(), "nofx-main.lock")
	if err := os.WriteFile(legacy, []byte("owner=hoang pid=4194301 task=old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env["TREE_GUARD_LOCK"] = legacy
	env["TREE_GUARD_LOCK_DIR"] = newLockDir(t, "nofx-63", "merge", 10*time.Second)

	out, _ := runGuard(t, env)
	if !strings.Contains(out, "legacy") {
		t.Fatalf("a leftover legacy lock file must be surfaced:\n%s", out)
	}
}

// With the new lock absent AND no legacy file, a dirty tree is still the 08:46
// signature. The guard must not treat "I don't recognise any lock" as consent.
func TestNoLockOfEitherKindStillAlarmsOnDirt(t *testing.T) {
	tree := fixtureRepo(t)
	if err := os.WriteFile(filepath.Join(tree, "trader", "shipped.go"), []byte("package trader\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := baseEnv(t, tree)
	env["TREE_GUARD_LOCK_DIR"] = filepath.Join(t.TempDir(), "absent.lock.d")
	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM porcelain") {
		t.Fatalf("no lock of either kind + dirty tree must ALARM:\n%s", out)
	}
}

// ── CHECK 5: the canon files ────────────────────────────────────────────
//
// Owner ruling 2026-09-03: CLAUDE.md's canon moved into a TRACKED file
// (docs/superpowers/CLAUDE-canon.md) because a law that cannot travel on a branch
// or be guarded is not canon. The pointer file stays gitignored — and therefore
// stays INVISIBLE to check 1, which is exactly why it needs its own check.
//
// The baseline is recorded in the guard's own state file. A change under a live
// lock is accepted and re-baselined; a change with NO live lock ALARMS and does
// NOT re-baseline, so an unexplained edit to the standing laws keeps shouting
// until a human resolves it rather than going quiet after one tick.

func writeCanonPair(t *testing.T, dir, pointer, canon string) (string, string) {
	t.Helper()
	p := filepath.Join(dir, "CLAUDE.md")
	c := filepath.Join(dir, "CLAUDE-canon.md")
	if err := os.WriteFile(p, []byte(pointer), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c, []byte(canon), 0o644); err != nil {
		t.Fatal(err)
	}
	return p, c
}

func TestCheck5FirstRunRecordsTheBaselineAndPasses(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	dir := t.TempDir()
	p, c := writeCanonPair(t, dir, "pointer v1", "canon v1")
	env["TREE_GUARD_CANON_FILES"] = p + ":" + c
	state := filepath.Join(t.TempDir(), "state")
	env["TREE_GUARD_STATE"] = state

	out, _ := runGuard(t, env)
	if !strings.Contains(out, "PASS canon") {
		t.Fatalf("a first run has nothing to compare and must PASS while recording:\n%s", out)
	}
	b, err := os.ReadFile(state)
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if !strings.Contains(string(b), "canon_md5") {
		t.Errorf("the baseline must be recorded in the state file:\n%s", b)
	}
}

func TestCheck5UnchangedCanonPasses(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	dir := t.TempDir()
	p, c := writeCanonPair(t, dir, "pointer v1", "canon v1")
	env["TREE_GUARD_CANON_FILES"] = p + ":" + c
	env["TREE_GUARD_STATE"] = filepath.Join(t.TempDir(), "state")

	runGuard(t, env) // baseline
	out, code := runGuard(t, env)
	if !strings.Contains(out, "PASS canon") || code != 0 {
		t.Fatalf("an unchanged canon must PASS on the second run:\n%s", out)
	}
}

// THE RULING'S POINT: an edit to the standing laws with nobody holding the lock.
func TestCheck5ChangedWithNoLockAlarmsAndNamesTheFile(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	dir := t.TempDir()
	p, c := writeCanonPair(t, dir, "pointer v1", "canon v1")
	env["TREE_GUARD_CANON_FILES"] = p + ":" + c
	env["TREE_GUARD_STATE"] = filepath.Join(t.TempDir(), "state")

	runGuard(t, env) // baseline
	if err := os.WriteFile(c, []byte("canon v2 — a law quietly deleted"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, code := runGuard(t, env)
	if !strings.Contains(out, "ALARM canon") {
		t.Fatalf("a canon change with no live lock must ALARM:\n%s", out)
	}
	if !strings.Contains(out, "CLAUDE-canon.md") {
		t.Errorf("the alarm must NAME the changed file:\n%s", out)
	}
	if code == 0 {
		t.Errorf("exit must be non-zero")
	}
}

// It must keep alarming. A one-shot alarm that re-baselines itself turns an
// unexplained edit into the new normal after a single 60s tick.
func TestCheck5KeepsAlarmingUntilResolved(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	dir := t.TempDir()
	p, c := writeCanonPair(t, dir, "pointer v1", "canon v1")
	env["TREE_GUARD_CANON_FILES"] = p + ":" + c
	env["TREE_GUARD_STATE"] = filepath.Join(t.TempDir(), "state")

	runGuard(t, env)
	if err := os.WriteFile(c, []byte("canon v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 3; i++ {
		out, _ := runGuard(t, env)
		if !strings.Contains(out, "ALARM canon") {
			t.Fatalf("run %d went quiet — an unexplained canon edit must not become the new baseline:\n%s", i, out)
		}
	}
}

// A change UNDER A LIVE LOCK is a legitimate edit: INFO, and re-baselined so the
// next run is clean.
func TestCheck5ChangeUnderALiveLockIsAcceptedAndRebaselined(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	dir := t.TempDir()
	p, c := writeCanonPair(t, dir, "pointer v1", "canon v1")
	env["TREE_GUARD_CANON_FILES"] = p + ":" + c
	env["TREE_GUARD_STATE"] = filepath.Join(t.TempDir(), "state")

	runGuard(t, env) // baseline with no lock
	if err := os.WriteFile(c, []byte("canon v2 — edited on purpose"), 0o644); err != nil {
		t.Fatal(err)
	}
	env["TREE_GUARD_LOCK_DIR"] = newLockDir(t, "nofx-canon", "edit the canon", 20*time.Second)

	out, _ := runGuard(t, env)
	if strings.Contains(out, "ALARM canon") {
		t.Fatalf("a canon edit under a live lock is legitimate:\n%s", out)
	}
	if !strings.Contains(out, "INFO canon") {
		t.Errorf("it must still be reported, not silent:\n%s", out)
	}
	// re-baselined: with the lock gone, the next run is clean
	env["TREE_GUARD_LOCK_DIR"] = filepath.Join(t.TempDir(), "absent.lock.d")
	out2, _ := runGuard(t, env)
	if !strings.Contains(out2, "PASS canon") {
		t.Fatalf("an accepted edit must become the new baseline:\n%s", out2)
	}
}

// A canon file that VANISHES is the loudest case of all.
func TestCheck5MissingCanonFileAlarms(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	dir := t.TempDir()
	p, c := writeCanonPair(t, dir, "pointer v1", "canon v1")
	env["TREE_GUARD_CANON_FILES"] = p + ":" + c
	env["TREE_GUARD_STATE"] = filepath.Join(t.TempDir(), "state")

	runGuard(t, env)
	if err := os.Remove(c); err != nil {
		t.Fatal(err)
	}
	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM canon") || !strings.Contains(strings.ToLower(out), "missing") {
		t.Fatalf("a vanished canon file must ALARM and say missing:\n%s", out)
	}
}

// ── THE RULING ITSELF: canon must be tracked, the pointer must point ────

func TestCanonFileIsTrackedAndCarriesTheLaws(t *testing.T) {
	canon := filepath.Join("..", "docs", "superpowers", "CLAUDE-canon.md")
	b, err := os.ReadFile(canon)
	if err != nil {
		t.Fatalf("the canon file must exist at docs/superpowers/CLAUDE-canon.md: %v", err)
	}
	s := string(b)
	// Every standing law the repo runs on must have survived the move.
	for _, law := range []string{
		"WORKTREE LAW", "MAIN-TREE LOCK LAW", "SPEC-FRESHNESS LAW",
		"NO UNATTENDED DEPLOYS", "GUIDE CONTENT LAW", "AUDIT-PLAYBOOK LAW",
	} {
		if !strings.Contains(s, law) {
			t.Errorf("the canon lost %q in the move", law)
		}
	}
	if len(strings.Split(s, "\n")) < 200 {
		t.Errorf("the canon is suspiciously short (%d lines) — did the move truncate it?",
			len(strings.Split(s, "\n")))
	}
}

// ── SHORT SHA vs FULL SHA — my own guard's false alarm ──────────────────
//
// Found on the guard's SECOND live run, on a healthy tree, minutes after boot 7:
//
//   🚨 ALARM release: RELEASE=530009ff · running=530009ffd5402c206d1ef7...
//      — the file and the binary disagree with no cutover in flight
//
// They do not disagree. deploy/RELEASE held the 8-char short form and the binary
// reports the full 40-char sha; a string equality test calls that a mismatch.
// Earlier deploys happened to write the full sha, so the bug stayed hidden until
// a lane wrote the short one — which is the normal, correct thing to write.
//
// This is precisely the failure this wave's own report warned about: a guard that
// cries wolf on a healthy deploy is worse than no guard, because the next real
// alarm is the one everyone scrolls past. Git compares shas by prefix; so must this.

// setReleaseCommitted writes deploy/RELEASE and COMMITS it, so a comparison test
// exercises only the comparison — not the porcelain check and not the
// "marker never committed" check, both of which would otherwise fire and mask it.
func setReleaseCommitted(t *testing.T, tree, rev string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(tree, "deploy", "RELEASE"), []byte(rev+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-qm", "deploy: RELEASE=" + rev}} {
		c := exec.Command("git", args...)
		c.Dir = tree
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
}

func TestReleaseComparisonAcceptsAShortShaAgainstAFullOne(t *testing.T) {
	tree := fixtureRepo(t)
	full := "530009ffd5402c206d1ef7fbdcc0e5b14f8814ad"
	setReleaseCommitted(t, tree, "530009ff")
	env := baseEnv(t, tree)
	delete(env, "TREE_GUARD_SKIP_RUNNING")
	env["TREE_GUARD_RUNNING_REV"] = full

	out, code := runGuard(t, env)
	if strings.Contains(out, "ALARM release") {
		t.Fatalf("a short RELEASE against the full running sha is the SAME rev:\n%s", out)
	}
	if !strings.Contains(out, "PASS release") {
		t.Errorf("expected PASS:\n%s", out)
	}
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
}

// The other direction, and the case that must STILL alarm: a genuine mismatch
// that merely shares no prefix. Prefix-matching must not become "any overlap".
func TestReleaseComparisonStillCatchesARealMismatch(t *testing.T) {
	tree := fixtureRepo(t)
	setReleaseCommitted(t, tree, "530009ff")
	env := baseEnv(t, tree)
	delete(env, "TREE_GUARD_SKIP_RUNNING")
	env["TREE_GUARD_RUNNING_REV"] = "89673ccc5984eef8ad8168edda24984fd1da50cb"

	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM release") {
		t.Fatalf("530009ff vs 89673ccc… is a REAL mismatch and must still ALARM:\n%s", out)
	}
}

// A one-character RELEASE must not prefix-match everything.
func TestReleaseComparisonRefusesAnAbsurdlyShortPrefix(t *testing.T) {
	tree := fixtureRepo(t)
	setReleaseCommitted(t, tree, "5")
	env := baseEnv(t, tree)
	delete(env, "TREE_GUARD_SKIP_RUNNING")
	env["TREE_GUARD_RUNNING_REV"] = "530009ffd5402c206d1ef7fbdcc0e5b14f8814ad"

	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM release") {
		t.Fatalf("a 1-char RELEASE is not a sha and must not silently match:\n%s", out)
	}
}

// ── CHECK 5b: a POINTER THAT POINTS AT NOTHING ──────────────────────────
//
// nofx-63's catch, and they were right that check 5 as shipped would have stayed
// silent through it. Check 5 alarms when a canon file CHANGES. It says nothing
// about whether the file still leads anywhere.
//
// That is not hypothetical: the repo-root CLAUDE.md was turned into a two-line
// @import at 22:59 while its target existed only on an unmerged branch. For ~35
// minutes every reader of the main tree got a dangling reference instead of the
// standing laws — and the file's own comment told them to go read a path that
// did not exist. An unchanged pointer to nothing is the worst case: stable,
// silent, and empty.

func TestCheck5AlarmsOnADanglingImportTarget(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	dir := t.TempDir()
	pointer := "# canon lives elsewhere\n\n@docs/superpowers/CLAUDE-canon.md\n"
	p, c := writeCanonPair(t, dir, pointer, "canon v1")
	env["TREE_GUARD_CANON_FILES"] = p + ":" + c
	env["TREE_GUARD_STATE"] = filepath.Join(t.TempDir(), "state")

	out, code := runGuard(t, env)
	if !strings.Contains(out, "ALARM canon") {
		t.Fatalf("an @import whose target does not exist must ALARM:\n%s", out)
	}
	if !strings.Contains(out, "CLAUDE-canon.md") {
		t.Errorf("the alarm must name the unresolved target:\n%s", out)
	}
	if code == 0 {
		t.Errorf("exit must be non-zero")
	}
}

// It must alarm on the FIRST run, before any baseline exists — a pointer that was
// born dangling never "changes", so a change-only check never fires on it.
func TestCheck5DanglingImportAlarmsEvenWithNoBaseline(t *testing.T) {
	tree := fixtureRepo(t)
	env := baseEnv(t, tree)
	dir := t.TempDir()
	p, c := writeCanonPair(t, dir, "@nowhere/at/all.md\n", "canon v1")
	env["TREE_GUARD_CANON_FILES"] = p + ":" + c
	env["TREE_GUARD_STATE"] = filepath.Join(t.TempDir(), "state")

	out, _ := runGuard(t, env)
	if !strings.Contains(out, "ALARM canon") {
		t.Fatalf("a born-dangling pointer must alarm on the first run:\n%s", out)
	}
}

// A resolvable import is fine, and a file with no import at all is fine.
func TestCheck5ResolvableAndImportlessCanonPass(t *testing.T) {
	tree := fixtureRepo(t)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "docs", "superpowers"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "docs", "superpowers", "CLAUDE-canon.md")
	if err := os.WriteFile(target, []byte("the laws"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(p, []byte("@docs/superpowers/CLAUDE-canon.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := baseEnv(t, tree)
	env["TREE_GUARD_CANON_FILES"] = p + ":" + target
	env["TREE_GUARD_STATE"] = filepath.Join(t.TempDir(), "state")

	out, code := runGuard(t, env)
	if strings.Contains(out, "ALARM canon") {
		t.Fatalf("a resolvable import must not alarm:\n%s", out)
	}
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
}
