package trader

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── A29 STANDING GATE — "reported wired, called by nobody" ───────────────────
//
// Checklist 63. Three instances inside 24 hours (2026-09-02/03):
//   1. the seam migration — a grade "excluded" by a reader nothing called;
//   2. the 1B detector hook — detector_record.go's header declared itself
//      "THE PRODUCTION CALL PATH" while zero production code called it, so
//      touch_outcomes and candidate_pool booted 0/0 on every boot and the
//      wake-predicate wave that depended on them had no substrate;
//   3. ab_confirm_log's "usable" figure — healthy-looking because what it
//      dropped was never counted.
//
// The shared shape is a claim of being wired that no test could falsify. This
// gate falsifies it: any function whose OWN doc comment claims it is a
// production call path must have at least one production (non-test) call site.
// It reads the claim from the source, so a future function inherits the gate
// for free — there is no list to remember to update.
//
// It fails, it does not warn (A24: a check that prints but does not gate is
// not a check).

// wiringClaim matches a doc comment asserting production wiring.
var wiringClaims = []string{"production call path", "the production call site"}

// alsoRequireWired is the maintained escape hatch for functions that must be
// wired but whose doc does not say so in those words. Keep it SHORT; prefer
// putting the claim in the function's own comment.
var alsoRequireWired = []string{"recordDetectorOutputs"}

func TestEveryClaimedProductionPathHasACallSite(t *testing.T) {
	root := ".."
	claimed := map[string]string{}      // funcName -> where the claim was made
	fileClaims := map[string][]string{} // file that claims wiring -> funcs it declares
	for _, n := range alsoRequireWired {
		claimed[n] = "(maintained list)"
	}

	fset := token.NewFileSet()
	var goFiles []string
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if p == root {
				return nil // never skip the root itself (its Name() is "..")
			}
			base := info.Name()
			if base == ".git" || base == "node_modules" || base == "web" || base == "vendor" ||
				strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(p, ".go") {
			goFiles = append(goFiles, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(goFiles) < 50 {
		t.Fatalf("only %d .go files walked — the gate is not seeing the repo", len(goFiles))
	}

	// Pass 1: collect the claims (a claim in a _test.go file does not count).
	for _, p := range goFiles {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		f, perr := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if perr != nil {
			continue
		}
		// (a) the claim in a function's OWN doc comment.
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Doc == nil {
				continue
			}
			doc := strings.ToLower(fn.Doc.Text())
			for _, claim := range wiringClaims {
				if strings.Contains(doc, claim) {
					claimed[fn.Name.Name] = p
					break
				}
			}
		}
		// (b) the claim in a FILE-LEVEL banner. This is the shape 1B used —
		// detector_record.go's "── 1B — THE PRODUCTION CALL PATH ──" is a file
		// comment, so a func-doc-only scan misses precisely the case this gate
		// was written for. A file that declares itself a production call path
		// must have at least one of its own funcs called from production.
		var fileClaim bool
		for _, cg := range f.Comments {
			low := strings.ToLower(cg.Text())
			for _, claim := range wiringClaims {
				if strings.Contains(low, claim) {
					fileClaim = true
					break
				}
			}
		}
		if fileClaim {
			var declared []string
			for _, d := range f.Decls {
				if fn, ok := d.(*ast.FuncDecl); ok {
					declared = append(declared, fn.Name.Name)
				}
			}
			if len(declared) > 0 {
				fileClaims[p] = declared
			}
		}
	}

	// Pass 2: count PRODUCTION call sites (non-test files, excluding the decl).
	sites := map[string]int{}
	for _, p := range goFiles {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		f, perr := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if perr != nil {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			var name string
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				name = fn.Name
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			}
			if name != "" {
				sites[name]++
			}
			return true
		})
	}

	for file, funcs := range fileClaims {
		wired := 0
		for _, fn := range funcs {
			wired += sites[fn]
		}
		if wired == 0 {
			t.Errorf("A29 WIRING GATE: %s declares itself a production call path but NONE of its %d "+
				"function(s) has a production call site — the file is reported wired and called by nobody.", file, len(funcs))
			continue
		}
		t.Logf("A29 ok: %s (file claim) → %d production call site(s) across %d func(s)", file, wired, len(funcs))
	}

	for name, where := range claimed {
		if sites[name] == 0 {
			t.Errorf("A29 WIRING GATE: %q claims a production call path (%s) but has 0 production call sites — "+
				"it is reported wired and called by nobody. Wire it or drop the claim.", name, where)
			continue
		}
		t.Logf("A29 ok: %s → %d production call site(s) (%s)", name, sites[name], where)
	}
}
