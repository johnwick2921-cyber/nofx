// Package sqldriverpin holds ONE pin, in a test-only package that links
// NOTHING of the worker set on purpose: the failure it guards is an init-time
// panic ("sql: Register called twice for driver sqlite"), and a pin living in
// a test binary that links the worker set would die of that panic before it
// could say which binary links which two drivers. From here the toolchain is
// asked, and the answer names them.
package sqldriverpin_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// The two packages that register the database/sql driver "sqlite". The repo's
// ONE registration site is store/sqlitedriver: modernc in the default build,
// glebarez under -tags cgofree — never both in one binary.
const (
	modernc  = "modernc.org/sqlite"
	glebarez = "github.com/glebarez/go-sqlite"
)

// PIN (D4, CTO ruling 1790302009885): no binary linking the worker set
// registers the "sqlite" driver twice. internal/activation (#201) blank-
// imported glebarez/go-sqlite beside store/sqlitedriver's modernc; any binary
// linking both — the updater linking nofx/store (the hold) AND activation (the
// adapter) — panicked at init. For each binary the worker set reaches, the
// toolchain's own dependency list must not carry both registrations (and must
// carry one: a probe that sees neither proves nothing).
//
// The in-process half — this runs inside the updaterworker test binary itself
// — is updaterworker's TestTheWorkerTestBinaryRegistersSQLiteOnce.
func TestNoBinaryLinkingTheWorkerSetRegistersADuplicateSQLDriver(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root %s: %v", root, err)
	}
	for _, c := range []struct {
		binary string
		args   []string
	}{
		{"the nofx-updater binary (./cmd/nofx-updater)", []string{"list", "-deps", "-f", "{{.ImportPath}}", "./cmd/nofx-updater"}},
		{"the updaterworker test binary (./internal/updaterworker)", []string{"list", "-deps", "-test", "-f", "{{.ImportPath}}", "./internal/updaterworker"}},
		{"the api test binary (./api)", []string{"list", "-deps", "-test", "-f", "{{.ImportPath}}", "./api"}},
	} {
		t.Run(c.binary, func(t *testing.T) {
			deps := goList(t, root, c.args...)
			var drivers []string
			for _, d := range []string{modernc, glebarez} {
				if deps[d] {
					drivers = append(drivers, d)
				}
			}
			sort.Strings(drivers)
			switch len(drivers) {
			case 0:
				t.Fatalf("%s links neither %s nor %s — the probe is not seeing the binary (%d deps)", c.binary, modernc, glebarez, len(deps))
			case 2:
				t.Fatalf("%s links BOTH %s — two registrations of the database/sql driver \"sqlite\"; it panics at init "+
					"(\"sql: Register called twice for driver sqlite\"). Every package imports nofx/store/sqlitedriver, never a driver.",
					c.binary, strings.Join(drivers, " AND "))
			}
		})
	}
}

// goList runs the toolchain that built this test (GOTOOLCHAIN=local, offline,
// read-only go.mod — the census environment of internal/censuswalk) and
// returns the import paths it prints.
func goList(t *testing.T, root string, args ...string) map[string]bool {
	t.Helper()
	bin := filepath.Join(runtime.GOROOT(), "bin", "go")
	env := append(os.Environ(), "GOPROXY=off", "GOFLAGS=-mod=readonly", "GOWORK=off")
	if _, err := os.Stat(bin); err == nil {
		env = append(env, "GOTOOLCHAIN=local")
	} else {
		bin = "go"
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir, cmd.Env = root, env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, stderr.String())
	}
	deps := map[string]bool{}
	for _, ln := range strings.Split(string(out), "\n") {
		if ln = strings.TrimSpace(ln); ln != "" {
			deps[ln] = true
		}
	}
	return deps
}
