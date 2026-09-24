package updaterworker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
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
