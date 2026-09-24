package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"nofx/internal/updaterjob"
	"nofx/internal/updaterwire"
	"nofx/internal/updaterwire/wireserver"
	"nofx/internal/updaterworker"
)

const testJob = "job-u4-cli0abcd"

// install makes a short temp installation (the socket path must fit sun_path)
// with a bot database, and returns its dir and data dir.
func install(t *testing.T) (string, string) {
	t.Helper()
	t.Setenv("DB_PATH", "x")
	os.Unsetenv("DB_PATH") // the operator's shell exports none
	root, err := os.MkdirTemp("", "u4c-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(root) })
	if err := os.MkdirAll(filepath.Join(root, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "data.db"), []byte("SQLite format 3\x00"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, filepath.Join(root, "data")
}

func runCLI(t *testing.T, stdin io.Reader, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	if stdin == nil {
		stdin = strings.NewReader("")
	}
	rc := run(args, stdin, &out, &errb)
	return rc, out.String(), errb.String()
}

// L4: until the production adapters land, serve and fetch refuse and write
// NOTHING — no updater dir, no socket, no job, no hold — so an installed
// nofx-updater binary cannot change the installation.
func TestServeAndFetchRefuseUntilTheAdaptersLand(t *testing.T) {
	inst, data := install(t)
	t.Setenv(updaterworker.CutoverTokenEnv, "tok-cli-never-printed-51c2")
	checkProcess = func() error { return nil } // not root, not the bot's cgroup, no TZ
	t.Cleanup(func() { checkProcess = updaterworker.CheckProcess })
	rc, out, errs := runCLI(t, nil, "--install-dir", inst, "serve")
	if rc != 2 || !strings.Contains(errs, "not wired yet") || out != "" {
		t.Fatalf("serve = %d %q %q", rc, out, errs)
	}
	if strings.Contains(errs, "tok-cli-never-printed") {
		t.Fatal("serve printed the cutover token")
	}
	rc, _, errs = runCLI(t, nil, "--install-dir", inst, "fetch", "v1.2.0")
	if rc != 2 || !strings.Contains(errs, "not wired yet") {
		t.Fatalf("fetch = %d %q", rc, errs)
	}
	if _, err := os.Stat(filepath.Join(data, "updater")); !os.IsNotExist(err) {
		t.Fatalf("a refused serve/fetch created %s (%v)", filepath.Join(data, "updater"), err)
	}
	// and serve without the token refuses before anything else
	os.Unsetenv(updaterworker.CutoverTokenEnv)
	if rc, _, errs := runCLI(t, nil, "--install-dir", inst, "serve"); rc != 2 || !strings.Contains(errs, updaterworker.CutoverTokenEnv+" is not set") {
		t.Fatalf("serve without a token = %d %q", rc, errs)
	}
}

// PIN (the worker never runs as root): every subcommand refuses root before
// it resolves anything; serve also runs the worker's process refusals (root,
// the bot's cgroup, TZ) at its call site.
func TestEverySubcommandRefusesRoot(t *testing.T) {
	inst, _ := install(t)
	geteuid = func() int { return 0 }
	t.Cleanup(func() { geteuid = os.Geteuid })
	for _, args := range [][]string{{"serve"}, {"fetch", "v1.2.0"}, {"status"}, {"status", testJob}, {"resume", testJob}, {"recovery", testJob}} {
		rc, _, errs := runCLI(t, nil, append([]string{"--install-dir", inst}, args...)...)
		if rc != 2 || !strings.Contains(errs, "refusing to run as root") {
			t.Fatalf("%v as root = %d %q", args, rc, errs)
		}
	}
	geteuid = os.Geteuid
	checkProcess = func() error { return updaterworker.ErrBotCgroup }
	t.Cleanup(func() { checkProcess = updaterworker.CheckProcess })
	t.Setenv(updaterworker.CutoverTokenEnv, "tok")
	if rc, _, errs := runCLI(t, nil, "--install-dir", inst, "serve"); rc != 2 || !strings.Contains(errs, "nofx.service control group") {
		t.Fatalf("serve inside the bot's cgroup = %d %q", rc, errs)
	}
}

// resume is attended: a terminal and the job id typed back, then exactly one
// resume frame for that job over the worker socket. A mistyped id, or no
// terminal, sends nothing.
func TestResumeSendsOneFrameOnlyAfterTheTypedJobID(t *testing.T) {
	inst, data := install(t)
	j, err := updaterjob.New(testJob, "v1.2.0", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := updaterjob.Write(data, j); err != nil {
		t.Fatal(err)
	}
	path, err := updaterwire.SocketPath(data)
	if err != nil {
		t.Fatal(err)
	}
	ln, err := wireserver.Listen(path, func(string, ...any) {})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	var mu sync.Mutex
	var got []updaterwire.Request
	go ln.Serve(func(r updaterwire.Request) updaterwire.Response {
		mu.Lock()
		got = append(got, r)
		mu.Unlock()
		return updaterwire.Response{OK: true, State: "resuming"}
	})
	isTerminal = func(io.Reader) bool { return true }
	t.Cleanup(func() { isTerminal = stdinIsTerminal })

	if rc, _, errs := runCLI(t, strings.NewReader("job-u4-wrong000\n"), "--install-dir", inst, "resume", testJob); rc != 2 || !strings.Contains(errs, "does not match") {
		t.Fatalf("mistyped resume = %d %q", rc, errs)
	}
	isTerminal = func(io.Reader) bool { return false }
	if rc, _, errs := runCLI(t, strings.NewReader(testJob+"\n"), "--install-dir", inst, "resume", testJob); rc != 2 || !strings.Contains(errs, "from a terminal") {
		t.Fatalf("resume without a terminal = %d %q", rc, errs)
	}
	mu.Lock()
	if len(got) != 0 {
		t.Fatalf("a refused resume sent %v", got)
	}
	mu.Unlock()
	isTerminal = func(io.Reader) bool { return true }
	rc, out, errs := runCLI(t, strings.NewReader(testJob+"\n"), "--install-dir", inst, "resume", testJob)
	if rc != 0 || !strings.HasSuffix(out, "resuming\n") {
		t.Fatalf("resume = %d %q %q", rc, out, errs)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 || got[0].Verb != updaterwire.VerbResume || got[0].Resume == nil || got[0].Resume.JobID != testJob {
		t.Fatalf("the worker received %+v, want exactly one resume for %s", got, testJob)
	}
}

// status <job> and recovery <job> read the job file (no worker needed);
// absent values print n/a.
func TestStatusAndRecoveryReadTheJobFile(t *testing.T) {
	inst, data := install(t)
	j, _ := updaterjob.New(testJob, "v1.2.0", time.Now())
	if err := updaterjob.Write(data, j); err != nil {
		t.Fatal(err)
	}
	rc, out, errs := runCLI(t, nil, "--install-dir", inst, "status", testJob)
	if rc != 0 || !strings.Contains(out, `"state": "requested"`) || !strings.Contains(out, `"blocker": "n/a"`) {
		t.Fatalf("status job = %d %q %q", rc, out, errs)
	}
	rc, out, _ = runCLI(t, nil, "--install-dir", inst, "recovery", testJob)
	if rc != 0 || !strings.Contains(out, "not recovery_needed; nothing to do") {
		t.Fatalf("recovery of a live job = %d %q", rc, out)
	}
	if rc, _, errs := runCLI(t, nil, "--install-dir", inst, "status"); rc != 1 || !strings.Contains(errs, "no worker answers") {
		t.Fatalf("status with no worker = %d %q", rc, errs)
	}
	if rc, _, errs := runCLI(t, nil, "--install-dir", inst, "status", "../../etc/passwd"); rc != 1 || !strings.Contains(errs, "invalid job id") {
		t.Fatalf("status of a forged id = %d %q", rc, errs)
	}
}
