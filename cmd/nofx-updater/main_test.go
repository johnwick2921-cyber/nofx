package main

import (
	"bytes"
	"context"
	"errors"
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
	// ONE adapter landed and the other not (the fold may land activation
	// before U3, or the reverse): still refused, still nothing written
	defer func() {
		newLibrary = func() (updaterworker.Library, error) { return nil, updaterworker.ErrNotWired }
		newReverifier = func(updaterworker.Target) (updaterworker.Reverifier, error) { return nil, updaterworker.ErrNotWired }
	}()
	newLibrary = func() (updaterworker.Library, error) { return testLib{}, nil }
	if rc, _, errs := runCLI(t, nil, "--install-dir", inst, "serve"); rc != 2 || !strings.Contains(errs, "not wired yet") {
		t.Fatalf("serve with the library but no re-proof adapter = %d %q", rc, errs)
	}
	newLibrary = func() (updaterworker.Library, error) { return nil, updaterworker.ErrNotWired }
	newReverifier = func(updaterworker.Target) (updaterworker.Reverifier, error) { return testRel{}, nil }
	if rc, _, errs := runCLI(t, nil, "--install-dir", inst, "serve"); rc != 2 || !strings.Contains(errs, "not wired yet") {
		t.Fatalf("serve with the re-proof but no library adapter = %d %q", rc, errs)
	}
	if _, err := os.Stat(filepath.Join(data, "updater")); !os.IsNotExist(err) {
		t.Fatalf("a half-wired serve created %s (%v)", filepath.Join(data, "updater"), err)
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

// testLib / testRel stand in for the two adapters the fold lands; every
// side effect refuses (serve here only proves its wiring).
type testLib struct{}

func (testLib) Resolve(string) (updaterworker.Release, error) { return updaterworker.Release{}, errNo }
func (testLib) Stage(updaterworker.Release) (updaterworker.Receipt, error) {
	return updaterworker.Receipt{}, errNo
}
func (testLib) Backup(string, string) (updaterworker.Receipt, error) {
	return updaterworker.Receipt{}, errNo
}
func (testLib) Snapshot(updaterworker.Release, string) (updaterworker.Receipt, error) {
	return updaterworker.Receipt{}, errNo
}
func (testLib) Activate(updaterworker.Release, updaterworker.Release, updaterworker.Identity) (updaterworker.Identity, updaterworker.Receipt, error) {
	return updaterworker.Identity{}, updaterworker.Receipt{}, errNo
}
func (testLib) Watch(updaterworker.Release, updaterworker.Identity, updaterworker.WatchOpts) (updaterworker.Receipt, error) {
	return updaterworker.Receipt{}, errNo
}
func (testLib) RollbackTo(updaterworker.Release, updaterworker.Release, updaterworker.Identity) (updaterworker.Identity, updaterworker.Receipt, error) {
	return updaterworker.Identity{}, updaterworker.Receipt{}, errNo
}
func (testLib) CurrentIdentity() (updaterworker.Identity, error) {
	return updaterworker.Identity{}, errNo
}

type testRel struct{}

func (testRel) Verdict(string) (updaterworker.Verdict, error) { return updaterworker.Verdict{}, errNo }
func (testRel) Rehash(updaterworker.Verdict) (int, error)     { return 0, errNo }
func (testRel) Reverify(updaterworker.Verdict) (updaterworker.ReleaseFacts, error) {
	return updaterworker.ReleaseFacts{}, errNo
}

var errNo = errors.New("test adapter: refused")

// serve, once its adapters exist, starts the worker (sweep), listens on the
// installation's socket, prints its start line READ (n/a when absent),
// answers the verbs, refuses an install whose release has no verdict, and
// ends cleanly on the signal — removing the socket.
func TestServeWiresTheWorkerBehindTheSocket(t *testing.T) {
	inst, data := install(t)
	t.Setenv(updaterworker.CutoverTokenEnv, "tok-serve-never-printed-9e1f")
	t.Setenv("HOME", t.TempDir())
	checkProcess = func() error { return nil }
	newLibrary = func() (updaterworker.Library, error) { return testLib{}, nil }
	newReverifier = func(updaterworker.Target) (updaterworker.Reverifier, error) { return testRel{}, nil }
	ctx, cancel := context.WithCancel(context.Background())
	serveContext = func() (context.Context, context.CancelFunc) { return ctx, cancel }
	t.Cleanup(func() {
		checkProcess = updaterworker.CheckProcess
		newLibrary = func() (updaterworker.Library, error) { return nil, updaterworker.ErrNotWired }
		newReverifier = func(updaterworker.Target) (updaterworker.Reverifier, error) { return nil, updaterworker.ErrNotWired }
		serveContext = func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) }
	})
	var errb syncBuffer
	rc := make(chan int, 1)
	go func() { rc <- run([]string{"--install-dir", inst, "serve"}, strings.NewReader(""), io.Discard, &errb) }()
	var c *updaterwire.Client
	for deadline := time.Now().Add(10 * time.Second); c == nil; time.Sleep(10 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatalf("serve never listened: %s", errb.String())
		}
		c, _ = updaterwire.DialWorker(data)
	}
	defer c.Close()
	if resp, err := c.Do(updaterwire.NewStatus("")); err != nil || resp.State != "idle" {
		t.Fatalf("status = %+v, %v", resp, err)
	}
	if resp, err := c.Do(updaterwire.NewInstall("v1.2.0", testJob)); err != nil || resp.Error != "release not verified" {
		t.Fatalf("install without a verdict = %+v, %v", resp, err)
	}
	cancel()
	if got := <-rc; got != 0 {
		t.Fatalf("serve ended %d: %s", got, errb.String())
	}
	line := errb.String()
	if !strings.Contains(line, "🔧 nofx-updater: serving ") || !strings.Contains(line, "active=n/a · recovery_needed=n/a · stale_at_start=n/a") {
		t.Fatalf("start line: %q", line)
	}
	if strings.Contains(line, "tok-serve-never-printed") {
		t.Fatal("serve printed the token")
	}
	if p, _ := updaterwire.SocketPath(data); fileExists(p) {
		t.Fatal("the socket survives serve")
	}
}

type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func fileExists(p string) bool { _, err := os.Lstat(p); return err == nil }
