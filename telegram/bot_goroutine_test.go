package telegram

// M3 ha verifier defect 3: runBot answers every AI message on its own
// goroutine, while the main loop's ident.refresh() — at start, on /start and
// before every AI call — REASSIGNS the identity's fields (agents, token,
// userID, email) whenever the bot re-mints. M3 H2 made the re-mint routine
// (every owner password change, every 24 h expiry), so a goroutine that reads
// a field of the identity races the next message's refresh.
//
// The rule, pinned where runBot and refresh actually run (canon 53 — the
// production source, parsed; no copy of the loop):
//   - runBot: no closure captures the identity variable, and no go statement
//     hands the identity (the pointer) to the goroutine. The goroutine gets
//     what it needs — the agent manager — as a value captured on the main
//     loop, BEFORE the go statement.
//   - botIdentity's methods: no closure they build reads the receiver. The
//     LLM factory refresh hands to agent.NewManager runs on the per-message
//     goroutine (Manager.Run → agent.New / Agent.Run), so it must close over
//     locals, never b.<field>.
//
// The -race reproduction of the second half is
// TestRaceBotRefreshAgainstInFlightManager (bot_refresh_race_test.go).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"sort"
	"strings"
	"testing"
)

type pkgFunc struct {
	file string
	decl *ast.FuncDecl
}

// parsePackageFuncs parses every non-test .go file of this package and
// returns its function declarations keyed "name" or "(*Recv).name".
func parsePackageFuncs(t *testing.T) (*token.FileSet, map[string][]pkgFunc) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	funcs := map[string][]pkgFunc{}
	parsed := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, e.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		parsed++
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			name := fd.Name.Name
			if fd.Recv != nil && len(fd.Recv.List) == 1 {
				if s, ok := fd.Recv.List[0].Type.(*ast.StarExpr); ok {
					if id, ok := s.X.(*ast.Ident); ok {
						name = "(*" + id.Name + ")." + name
					}
				}
			}
			funcs[name] = append(funcs[name], pkgFunc{file: e.Name(), decl: fd})
		}
	}
	if parsed == 0 {
		t.Fatal("no package source parsed — the pin walked nothing")
	}
	return fset, funcs
}

// refsTo lists every use of the named variables inside n, as "file:line: expr"
// — a selector on one of them (ident.agents) or the bare name (ident).
func refsTo(fset *token.FileSet, n ast.Node, names map[string]bool) []string {
	var out []string
	var visit func(ast.Node) bool
	visit = func(x ast.Node) bool {
		switch v := x.(type) {
		case *ast.SelectorExpr:
			if id, ok := v.X.(*ast.Ident); ok && names[id.Name] {
				out = append(out, fset.Position(v.Pos()).String()+": "+types.ExprString(v))
				return false
			}
			ast.Inspect(v.X, visit) // never v.Sel: a field NAMED like the variable is not a use
			return false
		case *ast.Ident:
			if names[v.Name] {
				out = append(out, fset.Position(v.Pos()).String()+": "+v.Name)
			}
		}
		return true
	}
	ast.Inspect(n, visit)
	return out
}

func onlyDecl(t *testing.T, funcs map[string][]pkgFunc, name string) pkgFunc {
	t.Helper()
	got := funcs[name]
	if len(got) != 1 {
		t.Fatalf("%s declared %d times in the package — want exactly 1", name, len(got))
	}
	return got[0]
}

// runBot's per-message goroutine reads no field of the bot identity: refresh
// on the main loop may replace every one of them while a message is answered.
func TestRunBotGoroutinesReadNoBotIdentityField(t *testing.T) {
	fset, funcs := parsePackageFuncs(t)
	rb := onlyDecl(t, funcs, "runBot").decl

	// The identity variable(s): whatever runBot assigns newBotIdentity(...) to.
	idents := map[string]bool{}
	ast.Inspect(rb.Body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Rhs) != 1 {
			return true
		}
		if call, ok := as.Rhs[0].(*ast.CallExpr); ok {
			if fn, ok := call.Fun.(*ast.Ident); ok && fn.Name == "newBotIdentity" {
				for _, l := range as.Lhs {
					if id, ok := l.(*ast.Ident); ok && id.Name != "_" {
						idents[id.Name] = true
					}
				}
			}
		}
		return true
	})
	if len(idents) == 0 {
		t.Fatal("runBot assigns newBotIdentity(...) to no variable — the pin has nothing to guard (re-anchor it)")
	}

	var bad []string
	goStmts, aiGoroutines := 0, 0
	ast.Inspect(rb.Body, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncLit:
			// Any closure in runBot: a closure is how a field read leaves
			// the main loop (go func(){…}(), or f := func(){…}; go f()).
			for _, r := range refsTo(fset, v.Body, idents) {
				bad = append(bad, r+" — read inside a closure in runBot")
			}
		case *ast.GoStmt:
			goStmts++
			if fl, ok := v.Call.Fun.(*ast.FuncLit); ok {
				ast.Inspect(fl.Body, func(m ast.Node) bool {
					if sel, ok := m.(*ast.SelectorExpr); ok && sel.Sel.Name == "Run" {
						aiGoroutines++
					}
					return true
				})
			} else {
				// go ident.method(...) / go helper(...): the function value's
				// body runs on the goroutine.
				for _, r := range refsTo(fset, v.Call.Fun, idents) {
					bad = append(bad, r+" — the go statement's function reads the identity")
				}
			}
			// Arguments are evaluated on the main loop (Go spec), so
			// ident.agents THERE is the capture — but handing over the
			// identity pointer itself lets the goroutine read its fields.
			for _, a := range v.Call.Args {
				if id, ok := a.(*ast.Ident); ok && idents[id.Name] {
					bad = append(bad, fset.Position(id.Pos()).String()+": "+id.Name+" — the identity pointer handed to a goroutine")
				}
			}
		}
		return true
	})
	if goStmts == 0 || aiGoroutines == 0 {
		t.Fatalf("runBot has %d go statements, %d of them calling .Run — the AI goroutine the pin guards is gone (re-anchor it)", goStmts, aiGoroutines)
	}
	if len(bad) > 0 {
		sort.Strings(bad)
		t.Fatalf("runBot's per-message goroutine reads the bot identity, which refresh() on the main loop reassigns (data race; M3 made the re-mint routine) — capture what it needs on the main loop and pass it in:\n  %s", strings.Join(bad, "\n  "))
	}
}

// The closures botIdentity's methods build read no field of the receiver:
// refresh hands its LLM factory to agent.NewManager, and the manager calls it
// on the per-message goroutine while the next refresh rewrites b.userID.
func TestBotIdentityClosuresReadNoReceiverField(t *testing.T) {
	fset, funcs := parsePackageFuncs(t)
	var bad []string
	methods, closures := 0, 0
	for name, decls := range funcs {
		if !strings.HasPrefix(name, "(*botIdentity).") {
			continue
		}
		for _, d := range decls {
			methods++
			recv := map[string]bool{}
			for _, n := range d.decl.Recv.List[0].Names {
				if n.Name != "_" {
					recv[n.Name] = true
				}
			}
			ast.Inspect(d.decl.Body, func(n ast.Node) bool {
				if fl, ok := n.(*ast.FuncLit); ok {
					closures++
					for _, r := range refsTo(fset, fl.Body, recv) {
						bad = append(bad, r+" — in a closure built by "+name)
					}
				}
				return true
			})
		}
	}
	if methods == 0 || closures == 0 {
		t.Fatalf("botIdentity has %d methods building %d closures — the LLM factory the pin guards is gone (re-anchor it)", methods, closures)
	}
	if len(bad) > 0 {
		sort.Strings(bad)
		t.Fatalf("a closure botIdentity hands out reads the receiver, which the next refresh() rewrites on the main loop while the per-message goroutine runs the closure (data race) — close over locals:\n  %s", strings.Join(bad, "\n  "))
	}
}
