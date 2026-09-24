package censuswalk

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// PIN (M3 fold M5, red-team 4 #1): the skip names bind ONLY at the module
// root. Every probe dir below the root (api/web, internal/node_modules/p,
// api/.git, x/testdata/y, _x, api/.hidden, …) is walked; the root-level skip
// dirs are not; _test.go files never are.
func TestWalkSkipsOnlyAtTheModuleRoot(t *testing.T) {
	root := t.TempDir()
	var want []string
	for _, n := range RootSkips() {
		write(t, root, n+"/a.go", "package a\n")
		write(t, root, n+"/deeper/b.go", "package deeper\n")
	}
	for _, d := range NestedProbeDirs() {
		write(t, root, d+"/a.go", "package "+PackageName(d)+"\n")
		write(t, root, d+"/a_test.go", "package "+PackageName(d)+"\n")
		want = append(want, d+"/a.go")
	}
	write(t, root, "web.go", "package main\n") // a FILE named like a skip dir is ordinary
	want = append(want, "web.go")
	files, err := NonTestGoFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, f := range files {
		got = append(got, f.Rel)
		if f.Path != filepath.Join(root, filepath.FromSlash(f.Rel)) {
			t.Fatalf("Path %q does not match Rel %q", f.Path, f.Rel)
		}
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("walked:\n%s\nwant exactly:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// The root-level skip set is pinned: widening it is a reviewed act.
func TestRootSkipsArePinned(t *testing.T) {
	want := ".Codex,.claude,.git,.understand-anything,node_modules,vendor,web"
	if got := strings.Join(RootSkips(), ","); got != want {
		t.Fatalf("root-level skips = %s, want exactly %s", got, want)
	}
}

// Ask the tool: every module package the toolchain links from anything
// `go list ./...` matches (outside the root-level skips) lies inside the walk.
// This is what makes a skip that hides a compiled package impossible to add
// silently — wherever it is and however it is spelled.
func TestWalkCoversEveryPackageTheToolchainLinks(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	uncovered, roots, err := ToolchainUncovered(root)
	if err != nil {
		t.Fatal(err)
	}
	if roots < 50 {
		t.Fatalf("go list ./... matched only %d packages — the toolchain is not seeing the module", roots)
	}
	if len(uncovered) > 0 {
		t.Fatalf("packages the build links but the census walk does not cover:\n%s", strings.Join(uncovered, "\n"))
	}
}

// ToolchainUncovered has teeth: on a synthetic module, a linked package under
// a root-level skip dir (web/evil) is reported; linked packages in every
// nested probe dir — including _x and x/testdata/y, which `go list ./...`
// itself never matches — are covered (walked). Negative control first.
func TestToolchainUncoveredControls(t *testing.T) {
	root := t.TempDir()
	write(t, root, "go.mod", "module nofx\n\ngo 1.25\n")
	var imports []string
	for _, d := range NestedProbeDirs() {
		if strings.Contains("/"+d+"/", "/vendor/") {
			continue // nested vendor is not importable in module mode
		}
		write(t, root, d+"/a.go", "package "+PackageName(d)+"\n")
		imports = append(imports, "import _ \"nofx/"+d+"\"\n")
	}
	write(t, root, "api/api.go", "package api\n\n"+strings.Join(imports, ""))
	write(t, root, "main.go", "package main\n\nimport _ \"nofx/api\"\n\nfunc main() {}\n")
	uncovered, roots, err := ToolchainUncovered(root)
	if err != nil || len(uncovered) != 0 || roots == 0 {
		t.Fatalf("every nested probe dir is walked: uncovered=%v roots=%d err=%v", uncovered, roots, err)
	}
	write(t, root, "web/evil/e.go", "package evil\n")
	write(t, root, "api/evil.go", "package api\n\nimport _ \"nofx/web/evil\"\n")
	uncovered, _, err = ToolchainUncovered(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(uncovered) != 1 || !strings.HasPrefix(uncovered[0], "nofx/web/evil ") {
		t.Fatalf("a linked package under the root-level web/ must be reported, got %v", uncovered)
	}
}

// nonTestImporters asks the toolchain which packages matched by ./... import
// target from a NON-test file (.Imports never carries TestImports or
// XTestImports). A failing go command is an error (fail closed).
func nonTestImporters(root, target string) (importers []string, matched int, err error) {
	cmd := goCommand(root, "list", "-e", "-f", "{{.ImportPath}}{{range .Imports}} {{.}}{{end}}", "./...")
	out, err := cmd.Output()
	if err != nil {
		return nil, 0, err
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		matched++
		for _, ip := range f[1:] {
			if ip == target {
				importers = append(importers, f[0])
			}
		}
	}
	sort.Strings(importers)
	return importers, matched, nil
}

// PIN (M3 census repair, verifier N1): the package doc says "Test tooling
// only: nothing but _test.go files may import this package" — asked of the
// toolchain, not asserted: no non-test file of the module imports censuswalk,
// and neither the app nor any cmd/ binary links it.
func TestCensusWalkIsTestToolingOnly(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	module, err := ModulePath(root)
	if err != nil {
		t.Fatal(err)
	}
	target := module + "/internal/censuswalk"
	importers, matched, err := nonTestImporters(root, target)
	if err != nil {
		t.Fatal(err)
	}
	if matched < 50 {
		t.Fatalf("go list ./... matched only %d packages — the toolchain is not seeing the module", matched)
	}
	if len(importers) > 0 {
		t.Fatalf("non-test code imports %s (test tooling only):\n%s", target, strings.Join(importers, "\n"))
	}
	linked, err := ListPackages(root, true, ".", "./cmd/...")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range linked {
		if p.ImportPath == target {
			t.Fatalf("the app or a cmd/ binary links %s (go list -deps . ./cmd/...)", target)
		}
	}
}

// nonTestImporters has teeth: a non-test importer is reported, a _test.go
// importer is not.
func TestNonTestImportersControls(t *testing.T) {
	root := t.TempDir()
	write(t, root, "go.mod", "module nofx\n\ngo 1.25\n")
	write(t, root, "internal/censuswalk/w.go", "package censuswalk\n")
	write(t, root, "api/api.go", "package api\n")
	write(t, root, "api/api_test.go", "package api\n\nimport _ \"nofx/internal/censuswalk\"\n")
	got, _, err := nonTestImporters(root, "nofx/internal/censuswalk")
	if err != nil || len(got) != 0 {
		t.Fatalf("a _test.go importer is test tooling: got %v err %v", got, err)
	}
	write(t, root, "api/uses.go", "package api\n\nimport _ \"nofx/internal/censuswalk\"\n")
	got, _, err = nonTestImporters(root, "nofx/internal/censuswalk")
	if err != nil || len(got) != 1 || got[0] != "nofx/api" {
		t.Fatalf("a non-test importer must be reported: got %v err %v", got, err)
	}
}
