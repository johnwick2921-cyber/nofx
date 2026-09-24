package updaterworker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"nofx/internal/censuswalk"
)

// ── census: the hold writer seams live in hold.go only ─────────────────────
//
// store/maintenance_hold_writers_test.go admits internal/updaterworker/hold.go
// BY NAME to call the store's hold writers. hold.go reaches them through two
// package vars (so a test can observe the file at the instant of the write);
// a call through a var is not a call the store census can see by name. So
// this census pins the rest: no other non-test file of the worker package
// names writeMaintenanceHold / clearMaintenanceHold — every hold write and
// clear goes through HoldForJob / ReleaseJob — and those two are called only
// from the steps that the state table says write or clear (steps.go).
func TestWorkerHoldSeamCensus(t *testing.T) {
	seams := map[string]bool{"writeMaintenanceHold": true, "clearMaintenanceHold": true}
	callers := map[string]map[string]bool{"HoldForJob": {}, "ReleaseJob": {}}
	ents, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	scanned := 0
	for _, e := range ents {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		scanned++
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				if seams[x.Name] && name != "hold.go" {
					t.Errorf("%s names the hold writer seam %s — only hold.go may", name, x.Name)
				}
			case *ast.CallExpr:
				if id, ok := x.Fun.(*ast.Ident); ok {
					if m, ok := callers[id.Name]; ok {
						m[name] = true
					}
				}
			}
			return true
		})
	}
	if scanned < 10 {
		t.Fatalf("scanned %d files — the census is not covering the package", scanned)
	}
	for fn, files := range callers {
		got := fmt.Sprint(keysOf(files))
		if got != "[steps.go]" {
			t.Errorf("%s is called from %s, want exactly [steps.go] (stepHold / stepReleaseHold)", fn, got)
		}
	}
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ── census: nothing outside the worker calls its hold writers ──────────────
//
// hold.go is admitted by the store census to call the store's writers, and it
// EXPORTS HoldForJob / ReleaseJob — a call to those is a hold write the store
// census cannot see (it matches the store's names). So the whole module's
// non-test files are scanned: any file importing nofx/internal/updaterworker
// (under any name, or dot-imported) that names HoldForJob or ReleaseJob is an
// offender — cmd/nofx-updater included (the CLI never touches the hold; its
// recovery text tells the OPERATOR to clear it with maintenance-hold).
func TestWorkerHoldWritersCensusModuleWide(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	off, scanned, err := workerHoldCallers(root)
	if err != nil {
		t.Fatal(err)
	}
	if scanned < 100 {
		t.Fatalf("scanned %d files — not the module", scanned)
	}
	if len(off) > 0 {
		t.Fatalf("the worker's hold writers are called from outside it:\n%s", strings.Join(off, "\n"))
	}
	// the census itself, on a synthetic module: a CLI calling it, a renamed
	// import, a dot import are all caught; a read of the display path is not
	dir := t.TempDir()
	put := func(rel, body string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(body), 0o644)
	}
	put("go.mod", "module nofx\n\ngo 1.25\n")
	put("internal/updaterworker/hold.go", "package updaterworker\n\nfunc HoldForJob() {}\nfunc ReleaseJob() {}\nfunc HoldFileForDisplay() string { return \"\" }\n")
	put("cmd/nofx-updater/main.go", "package main\n\nimport \"nofx/internal/updaterworker\"\n\nfunc main() { updaterworker.ReleaseJob() }\n")
	put("api/a.go", "package api\n\nimport uw \"nofx/internal/updaterworker\"\n\nvar _ = uw.HoldForJob\n")
	put("api/b.go", "package api\n\nimport . \"nofx/internal/updaterworker\"\n\nfunc b() { HoldForJob() }\n")
	put("api/c.go", "package api\n\nimport \"nofx/internal/updaterworker\"\n\nvar _ = updaterworker.HoldFileForDisplay\n")
	off, _, err = workerHoldCallers(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(off, "|"); got != "api/a.go: names updaterworker.HoldForJob|api/b.go: names updaterworker.HoldForJob|cmd/nofx-updater/main.go: names updaterworker.ReleaseJob" {
		t.Fatalf("synthetic census offenders = %v", off)
	}
}

func workerHoldCallers(root string) (offenders []string, scanned int, err error) {
	files, err := censuswalk.NonTestGoFiles(root)
	if err != nil {
		return nil, 0, err
	}
	module, err := censuswalk.ModulePath(root)
	if err != nil {
		return nil, 0, err
	}
	pkg := module + "/internal/updaterworker"
	writers := map[string]bool{"HoldForJob": true, "ReleaseJob": true}
	for _, file := range files {
		if path.Dir(file.Rel) == "internal/updaterworker" {
			continue // the package itself: TestWorkerHoldSeamCensus
		}
		f, perr := parser.ParseFile(token.NewFileSet(), file.Path, nil, parser.SkipObjectResolution)
		if perr != nil {
			offenders = append(offenders, file.Rel+": cannot be parsed")
			continue
		}
		scanned++
		local, dot := map[string]bool{}, false
		for _, imp := range f.Imports {
			if p, _ := strconv.Unquote(imp.Path.Value); p == pkg {
				switch {
				case imp.Name == nil:
					local["updaterworker"] = true
				case imp.Name.Name == ".":
					dot = true
				case imp.Name.Name != "_":
					local[imp.Name.Name] = true
				}
			}
		}
		if len(local) == 0 && !dot {
			continue
		}
		hit := map[string]bool{}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				if id, ok := x.X.(*ast.Ident); ok && local[id.Name] && writers[x.Sel.Name] {
					hit[x.Sel.Name] = true
				}
			case *ast.Ident:
				if dot && writers[x.Name] {
					hit[x.Name] = true
				}
			}
			return true
		})
		for _, name := range keysOf(hit) {
			offenders = append(offenders, file.Rel+": names updaterworker."+name)
		}
	}
	sort.Strings(offenders)
	return offenders, scanned, nil
}
