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
		"TREE_GUARD_TREE":    tree,
		"TREE_GUARD_LOCK":    filepath.Join(t.TempDir(), "absent.lock"),
		"TREE_GUARD_STATE":   filepath.Join(t.TempDir(), "state"),
		"TREE_GUARD_SYMBOLS": symbolsFile(t, "composeArmStop", "normalizeArmLegs", "CorrectedPnL"),
		"TREE_GUARD_SKIP_REMOTE": "1", // no origin in a fixture
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

func TestE1DirtyTreeUnderALiveLockIsInfoNotAlarm(t *testing.T) {
	tree := fixtureRepo(t)
	if err := os.WriteFile(filepath.Join(tree, "trader", "shipped.go"),
		[]byte("package trader\nfunc composeArmStop() {}\nfunc normalizeArmLegs() {}\nvar CorrectedPnL = 1\n// RELEASE bump\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := baseEnv(t, tree)
	lock := filepath.Join(t.TempDir(), "live.lock")
	// A lock naming OUR pid, which is by definition alive, with a fresh heartbeat.
	body := fmt.Sprintf("owner=hoang session=test pid=%d task=cutover heartbeat=%s\n",
		os.Getpid(), time.Now().UTC().Format(time.RFC3339))
	if err := os.WriteFile(lock, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	env["TREE_GUARD_LOCK"] = lock

	out, _ := runGuard(t, env)
	if strings.Contains(out, "ALARM porcelain") {
		t.Fatalf("a cutover legitimately dirties the tree — under a LIVE lock this is INFO:\n%s", out)
	}
	if !strings.Contains(out, "INFO") {
		t.Errorf("expected an INFO line naming the lock holder:\n%s", out)
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
