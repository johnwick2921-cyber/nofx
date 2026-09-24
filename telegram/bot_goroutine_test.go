package telegram

// M3 ha verifier defect 3: runBot answers every AI message on its own
// goroutine, while the main loop's ident.refresh() — at start, on /start and
// before every AI call — REASSIGNS the identity's fields (agents, token,
// userID, email) whenever the bot re-mints. M3 H2 made the re-mint routine
// (every owner password change, every 24 h expiry), so a goroutine that reads
// a field of the identity races the next message's refresh.
//
// M3 hc (verifier ha2 defect 2): these pins are TYPE-based. Their first
// version tracked the identity by VARIABLE NAME, so `id := ident; go func(){
// agents := id.agents … }()` restored the race with both pins green, and an
// LLM factory handed out as the method value b.llmClient was caught only by a
// vacuity guard. Every rule below asks go/types what an expression IS, never
// what it is called: package telegram's production files are type-checked
// (imports from the go command's own export data), and "the identity" is any
// value whose type is botIdentity or *botIdentity — through an alias, or a
// named type declared from it.
//
// The rules, pinned where runBot and refresh actually run (canon 53 — the
// production source, type-checked; no copy of the loop):
//
//   - TestRunBotGoroutinesReadNoBotIdentityField — no closure outside
//     botIdentity's methods uses the identity: not a selector through it
//     (ident.agents, w.id.agents, (*ident).x), not a field or method promoted
//     through an embedded one (w.agents), not the bare value (an alias, <-ch,
//     a conversion, a copy). A go statement binds no method of the identity
//     (go ident.refresh()), starts no package function that uses it
//     (go f(…)), and hands the goroutine no value that holds it. A go
//     statement's ARGUMENTS are evaluated on the main loop (Go spec), so
//     `}(ident.agents, chatID, text)` is the capture, and it is allowed.
//   - TestBotIdentityClosuresReadNoReceiverField — no closure a botIdentity
//     method builds uses the identity, and no method value is bound to the
//     identity anywhere in the package (b.llmClient, get := ident.refresh):
//     a method value IS a closure over its receiver. The LLM factory refresh
//     hands to agent.NewManager runs on the per-message goroutine
//     (Manager.Run → agent.New / Agent.Run), so it closes over locals.
//   - TestBotIdentityEscapesNoOtherWay — no package-level variable holds the
//     identity, and no value holding it is converted to an interface (or
//     unsafe.Pointer): through either, any goroutine could reach its fields
//     by a road the two rules above cannot see.
//
// The -race reproduction of the factory half is
// TestRaceBotRefreshAgainstInFlightManager (bot_refresh_race_test.go).

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// typedTelegram is package telegram's production source, type-checked.
type typedTelegram struct {
	fset        *token.FileSet
	files       []*ast.File
	info        *types.Info
	pkg         *types.Package
	identNamed  types.Type                     // botIdentity
	identStruct types.Type                     // its underlying struct
	decls       map[types.Object]*ast.FuncDecl // package funcs and methods → declaration
}

var (
	typedTelegramOnce sync.Once
	typedTelegramPkg  *typedTelegram
	typedTelegramErr  error
)

func loadTypedTelegram(t *testing.T) *typedTelegram {
	t.Helper()
	typedTelegramOnce.Do(func() { typedTelegramPkg, typedTelegramErr = typeCheckTelegram() })
	if typedTelegramErr != nil {
		t.Fatalf("the pin type-checks package telegram or it guards nothing: %v", typedTelegramErr)
	}
	return typedTelegramPkg
}

// goCommand is the go command running this test (PATH, else $GOROOT/bin).
func goCommand() (string, error) {
	if p, err := exec.LookPath("go"); err == nil {
		return p, nil
	}
	if root := os.Getenv("GOROOT"); root != "" {
		p := filepath.Join(root, "bin", "go")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", errors.New("no go command on PATH or in $GOROOT/bin")
}

// typeCheckTelegram type-checks the production files of the package in the
// working directory (package telegram, under go test). The file list and
// every import's export data come from ONE `go list -export -deps` — the go
// command's own build view, so the files checked are the files compiled.
func typeCheckTelegram() (*typedTelegram, error) {
	goBin, err := goCommand()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(goBin, "list", "-export", "-deps", "-f",
		"{{.ImportPath}}\t{{.Export}}\t{{.DepOnly}}\t{{join .GoFiles \",\"}}\t{{join .CgoFiles \",\"}}\t{{join .IgnoredGoFiles \",\"}}", ".")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list -export -deps .: %v\n%s", err, stderr.String())
	}
	exports := map[string]string{}
	var self []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		f := strings.Split(sc.Text(), "\t")
		if len(f) != 6 {
			return nil, fmt.Errorf("go list: unexpected line %q", sc.Text())
		}
		exports[f[0]] = f[1]
		if f[2] == "false" {
			if self != nil {
				return nil, errors.New("go list named more than one package for \".\"")
			}
			self = f
		}
	}
	if self == nil || self[0] != "nofx/telegram" {
		return nil, fmt.Errorf("go list did not name nofx/telegram for \".\" (got %q)", self)
	}
	if self[4] != "" {
		return nil, fmt.Errorf("cgo files %s are not type-checked by this pin — extend it", self[4])
	}
	for _, n := range strings.Split(self[5], ",") {
		if n != "" && !strings.HasSuffix(n, "_test.go") {
			return nil, fmt.Errorf("production file %s sits behind a build constraint this pin does not type-check — extend it", n)
		}
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, n := range strings.Split(self[3], ",") {
		if n == "" {
			continue
		}
		f, err := parser.ParseFile(fset, n, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		return nil, errors.New("no production file listed — the pin would walk nothing")
	}
	imp := importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		p := exports[path]
		if p == "" {
			return nil, fmt.Errorf("no export data for %q", path)
		}
		return os.Open(p)
	})
	info := &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}
	pkg, err := (&types.Config{Importer: imp}).Check("nofx/telegram", fset, files, info)
	if err != nil {
		return nil, err // any type error: the pin cannot trust what it did record
	}
	tn, ok := pkg.Scope().Lookup("botIdentity").(*types.TypeName)
	if !ok {
		return nil, errors.New("package telegram declares no type botIdentity — the identity the pins guard is gone (re-anchor them)")
	}
	tp := &typedTelegram{fset: fset, files: files, info: info, pkg: pkg,
		identNamed: tn.Type(), identStruct: tn.Type().Underlying(), decls: map[types.Object]*ast.FuncDecl{}}
	for _, f := range files {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				tp.decls[info.Defs[fd.Name]] = fd
			}
		}
	}
	return tp, nil
}

// isIdentity: T is botIdentity or *botIdentity — through an alias, a named
// type declared from it (type twin botIdentity), or a named pointer type.
func (tp *typedTelegram) isIdentity(T types.Type) bool {
	if T == nil {
		return false
	}
	T = types.Unalias(T)
	if p, ok := T.Underlying().(*types.Pointer); ok {
		T = types.Unalias(p.Elem())
	}
	return types.Identical(T.Underlying(), tp.identStruct)
}

// holdsIdentity: a value of type T carries the identity — it is one, or
// reaches one through pointers, struct fields, arrays, slices, maps, channels
// or a generic type's arguments. (Interfaces: TestBotIdentityEscapesNoOtherWay.)
func (tp *typedTelegram) holdsIdentity(T types.Type) bool {
	seen := map[types.Type]bool{}
	var reach func(types.Type) bool
	reach = func(T types.Type) bool {
		if T == nil {
			return false
		}
		T = types.Unalias(T)
		if tp.isIdentity(T) {
			return true
		}
		if seen[T] {
			return false
		}
		seen[T] = true
		if n, ok := T.(*types.Named); ok {
			for i := 0; i < n.TypeArgs().Len(); i++ {
				if reach(n.TypeArgs().At(i)) {
					return true
				}
			}
		}
		switch u := T.Underlying().(type) {
		case *types.Pointer:
			return reach(u.Elem())
		case *types.Struct:
			for i := 0; i < u.NumFields(); i++ {
				if reach(u.Field(i).Type()) {
					return true
				}
			}
		case *types.Slice:
			return reach(u.Elem())
		case *types.Array:
			return reach(u.Elem())
		case *types.Map:
			return reach(u.Key()) || reach(u.Elem())
		case *types.Chan:
			return reach(u.Elem())
		}
		return false
	}
	return reach(T)
}

// valueType is the type of e when e denotes a VALUE — nil for a type, a
// package, a function name, or a struct literal's field key.
func (tp *typedTelegram) valueType(e ast.Expr) types.Type {
	if id, ok := e.(*ast.Ident); ok {
		if v, ok := tp.info.Uses[id].(*types.Var); ok && !v.IsField() {
			return v.Type()
		}
		return nil
	}
	if tv, ok := tp.info.Types[e]; ok && tv.IsValue() {
		return tv.Type
	}
	return nil
}

// viaEmbeddedIdentity: the selection reaches its field or method through an
// embedded field of the identity's type (w.agents, w embedding *botIdentity) —
// a use of the identity no expression in the source spells.
func (tp *typedTelegram) viaEmbeddedIdentity(sel *types.Selection) bool {
	T := sel.Recv()
	idx := sel.Index()
	for _, i := range idx[:len(idx)-1] {
		T = types.Unalias(T)
		if p, ok := T.Underlying().(*types.Pointer); ok {
			T = p.Elem()
		}
		st, ok := types.Unalias(T).Underlying().(*types.Struct)
		if !ok {
			return false
		}
		f := st.Field(i)
		if tp.isIdentity(f.Type()) {
			return true
		}
		T = f.Type()
	}
	return false
}

// boundToIdentity: sel is a method bound to an identity receiver — directly
// (ident.refresh) or promoted through an embedded identity (w.refresh).
func (tp *typedTelegram) boundToIdentity(sel *types.Selection) bool {
	return sel != nil && sel.Kind() == types.MethodVal && (tp.isIdentity(sel.Recv()) || tp.viaEmbeddedIdentity(sel))
}

func (tp *typedTelegram) typeString(T types.Type) string {
	return types.TypeString(T, types.RelativeTo(tp.pkg))
}

func (tp *typedTelegram) at(pos token.Pos) string { return tp.fset.Position(pos).String() }

// identityUses lists every use of the identity inside n, one line each: a
// selector through it (reported whole: ident.agents), a field or method
// promoted through an embedded one, or the bare value (an alias, <-ch, a
// conversion, *ident).
func (tp *typedTelegram) identityUses(n ast.Node) []string {
	var out []string
	var visit func(ast.Node) bool
	visit = func(x ast.Node) bool {
		switch v := x.(type) {
		case *ast.SelectorExpr:
			if T := tp.valueType(v.X); tp.isIdentity(T) {
				out = append(out, fmt.Sprintf("%s: %s — through a %s", tp.at(v.X.Pos()), types.ExprString(v), tp.typeString(T)))
				return false
			}
			if sel := tp.info.Selections[v]; sel != nil && tp.viaEmbeddedIdentity(sel) {
				out = append(out, fmt.Sprintf("%s: %s — promoted through an embedded %s", tp.at(v.Pos()), types.ExprString(v), tp.typeString(tp.identNamed)))
				return false
			}
			ast.Inspect(v.X, visit) // never v.Sel: a field NAMED like a variable is not a use of it
			return false
		case ast.Expr:
			if T := tp.valueType(v); tp.isIdentity(T) {
				out = append(out, fmt.Sprintf("%s: %s — a %s value", tp.at(v.Pos()), types.ExprString(v), tp.typeString(T)))
				return false
			}
		}
		return true
	}
	ast.Inspect(n, visit)
	return out
}

// isIdentityMethod: fd is a method of the identity (pointer or value receiver).
func (tp *typedTelegram) isIdentityMethod(fd *ast.FuncDecl) bool {
	if fd == nil || fd.Recv == nil {
		return false
	}
	fn, ok := tp.info.Defs[fd.Name].(*types.Func)
	if !ok {
		return false
	}
	return tp.isIdentity(fn.Type().(*types.Signature).Recv().Type())
}

// eachFunc calls fn for every function body in the package's production
// files — every declared function or method — and for every package-level
// declaration (a func literal in a var initialiser) as "package scope".
func (tp *typedTelegram) eachFunc(fn func(name string, fd *ast.FuncDecl, body ast.Node)) {
	for _, f := range tp.files {
		for _, d := range f.Decls {
			switch v := d.(type) {
			case *ast.FuncDecl:
				if v.Body == nil {
					continue
				}
				name := v.Name.Name
				if obj, ok := tp.info.Defs[v.Name].(*types.Func); ok {
					name = strings.ReplaceAll(obj.FullName(), tp.pkg.Path()+".", "")
				}
				fn(name, v, v.Body)
			case *ast.GenDecl:
				fn("package scope", nil, v)
			}
		}
	}
}

// closureUses: every identity use inside a func literal in body, deduped (a
// use inside nested literals is one use).
func (tp *typedTelegram) closureUses(body ast.Node) []string {
	seen := map[string]bool{}
	var out []string
	ast.Inspect(body, func(n ast.Node) bool {
		if fl, ok := n.(*ast.FuncLit); ok {
			for _, u := range tp.identityUses(fl.Body) {
				if !seen[u] {
					seen[u] = true
					out = append(out, u)
				}
			}
		}
		return true
	})
	return out
}

// runBot's per-message goroutine uses no part of the bot identity: refresh on
// the main loop may replace every field of it while a message is answered.
func TestRunBotGoroutinesReadNoBotIdentityField(t *testing.T) {
	tp := loadTypedTelegram(t)
	var bad []string
	var runBot *ast.FuncDecl
	tp.eachFunc(func(name string, fd *ast.FuncDecl, body ast.Node) {
		if fd != nil && fd.Recv == nil && fd.Name.Name == "runBot" {
			runBot = fd
		}
		// Closures outside botIdentity's methods (those are the next test's).
		if !tp.isIdentityMethod(fd) {
			for _, u := range tp.closureUses(body) {
				bad = append(bad, u+" — used inside a closure in "+name)
			}
		}
		// Every go statement: what it binds, what it starts, what it hands over.
		ast.Inspect(body, func(n ast.Node) bool {
			g, ok := n.(*ast.GoStmt)
			if !ok {
				return true
			}
			switch fun := ast.Unparen(g.Call.Fun).(type) {
			case *ast.FuncLit:
				// its body: the closure rule above (or the next test's)
			case *ast.SelectorExpr:
				sel := tp.info.Selections[fun]
				if tp.boundToIdentity(sel) {
					bad = append(bad, fmt.Sprintf("%s: go %s — binds a method of the identity; its body runs on the goroutine (in %s)", tp.at(fun.Pos()), types.ExprString(fun), name))
				} else if sel != nil && sel.Kind() == types.MethodVal {
					if decl := tp.decls[sel.Obj()]; decl != nil {
						for _, u := range tp.identityUses(decl.Body) {
							bad = append(bad, u+" — in "+types.ExprString(fun)+", which `go` runs on a goroutine (in "+name+")")
						}
					}
				}
			case *ast.Ident:
				if obj, ok := tp.info.Uses[fun].(*types.Func); ok {
					if decl := tp.decls[obj]; decl != nil {
						for _, u := range tp.identityUses(decl.Body) {
							bad = append(bad, u+" — in "+fun.Name+", which `go` runs on a goroutine (in "+name+")")
						}
					}
				}
			}
			// Arguments are evaluated on the main loop (Go spec): a field
			// read THERE is the capture. A value that still HOLDS the
			// identity hands the goroutine the fields refresh rewrites.
			for _, a := range g.Call.Args {
				if T := tp.valueType(a); tp.holdsIdentity(T) {
					bad = append(bad, fmt.Sprintf("%s: %s — a %s handed to a goroutine (in %s)", tp.at(a.Pos()), types.ExprString(a), tp.typeString(T), name))
				}
			}
			return true
		})
	})

	// Vacuity: runBot, its identity variable, and the AI goroutine are where
	// the pin says they are.
	if runBot == nil {
		t.Fatal("package telegram declares no runBot — the loop the pin guards is gone (re-anchor it)")
	}
	idents := map[string]bool{}
	for id, obj := range tp.info.Defs {
		if v, ok := obj.(*types.Var); ok && !v.IsField() && id.Pos() >= runBot.Body.Pos() && id.Pos() < runBot.Body.End() && tp.isIdentity(v.Type()) {
			idents[id.Name] = true
		}
	}
	aiGoroutines := 0
	ast.Inspect(runBot.Body, func(n ast.Node) bool {
		g, ok := n.(*ast.GoStmt)
		if !ok {
			return true
		}
		ast.Inspect(g.Call.Fun, func(m ast.Node) bool {
			if se, ok := m.(*ast.SelectorExpr); ok {
				if fn, ok := tp.info.Uses[se.Sel].(*types.Func); ok && fn.Name() == "Run" && fn.Pkg() != nil && fn.Pkg().Path() == "nofx/telegram/agent" {
					aiGoroutines++
				}
			}
			return true
		})
		return true
	})

	if len(bad) > 0 {
		sort.Strings(bad)
		t.Errorf("runBot's per-message goroutine can reach the bot identity, which refresh() on the main loop reassigns (data race; M3 made the re-mint routine) — capture what it needs on the main loop and pass THAT in:\n  %s", strings.Join(bad, "\n  "))
	}
	if len(idents) == 0 {
		t.Errorf("runBot declares no variable of type %s — the identity the pin guards is gone (re-anchor it)", tp.typeString(tp.identNamed))
	}
	if aiGoroutines == 0 {
		t.Errorf("runBot has no go statement calling (*agent.Manager).Run — the AI goroutine the pin guards is gone (re-anchor it)")
	}
}

// The closures botIdentity's methods build — and every method value bound to
// the identity, anywhere — read no field of it: refresh hands its LLM factory
// to agent.NewManager, and the manager calls it on the per-message goroutine
// while the next refresh rewrites b.userID.
func TestBotIdentityClosuresReadNoReceiverField(t *testing.T) {
	tp := loadTypedTelegram(t)
	var bad []string
	methods, handoffs := 0, 0
	tp.eachFunc(func(name string, fd *ast.FuncDecl, body ast.Node) {
		if tp.isIdentityMethod(fd) {
			methods++
			for _, u := range tp.closureUses(body) {
				bad = append(bad, u+" — in a closure built by "+name)
			}
			ast.Inspect(body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok {
						if fn, ok := tp.info.Uses[se.Sel].(*types.Func); ok && fn.Name() == "NewManager" && fn.Pkg() != nil && fn.Pkg().Path() == "nofx/telegram/agent" {
							handoffs++
						}
					}
				}
				return true
			})
		}
		// A method value is a closure over its receiver. Called on the spot
		// it runs here; anywhere else it carries the identity with it. (A
		// `go` statement's call runs elsewhere: the previous test.)
		called := map[*ast.SelectorExpr]bool{}
		ast.Inspect(body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok {
					called[se] = true
				}
			}
			return true
		})
		ast.Inspect(body, func(n ast.Node) bool {
			if se, ok := n.(*ast.SelectorExpr); ok && !called[se] && tp.boundToIdentity(tp.info.Selections[se]) {
				bad = append(bad, fmt.Sprintf("%s: %s — a method value bound to the identity (a closure over it) in %s", tp.at(se.Pos()), types.ExprString(se), name))
			}
			return true
		})
	})
	if len(bad) > 0 {
		sort.Strings(bad)
		t.Errorf("a closure the bot hands out reads the identity, which the next refresh() rewrites on the main loop while the per-message goroutine runs the closure (data race) — close over locals:\n  %s", strings.Join(bad, "\n  "))
	}
	if methods == 0 || handoffs == 0 {
		t.Errorf("botIdentity has %d methods, %d calls to agent.NewManager among them — the LLM-factory hand-off the pin guards is gone (re-anchor it)", methods, handoffs)
	}
}

// No package-level variable holds the identity, and no value holding it is
// converted to an interface or unsafe.Pointer: either is a road to its
// fields from any goroutine that the two tests above cannot see.
func TestBotIdentityEscapesNoOtherWay(t *testing.T) {
	tp := loadTypedTelegram(t)
	var bad []string
	for _, n := range tp.pkg.Scope().Names() {
		if v, ok := tp.pkg.Scope().Lookup(n).(*types.Var); ok && tp.holdsIdentity(v.Type()) {
			bad = append(bad, fmt.Sprintf("%s: var %s %s — a package-level variable holding the identity", tp.at(v.Pos()), n, tp.typeString(v.Type())))
		}
	}
	opaque := func(T types.Type) bool {
		if T == nil {
			return false
		}
		T = types.Unalias(T)
		if _, ok := T.(*types.TypeParam); ok {
			return false
		}
		if b, ok := T.Underlying().(*types.Basic); ok {
			return b.Kind() == types.UnsafePointer
		}
		return types.IsInterface(T)
	}
	slots := 0
	check := func(val ast.Expr, slot types.Type, what, where string) {
		if val == nil || !opaque(slot) {
			return
		}
		slots++
		if T := tp.valueType(val); T != nil && !opaque(T) && tp.holdsIdentity(T) {
			bad = append(bad, fmt.Sprintf("%s: %s — a %s converted to %s (%s, in %s)", tp.at(val.Pos()), types.ExprString(val), tp.typeString(T), tp.typeString(slot), what, where))
		}
	}
	var walk func(body ast.Node, sig *types.Signature, where string)
	walk = func(body ast.Node, sig *types.Signature, where string) {
		ast.Inspect(body, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.FuncLit:
				s, _ := tp.valueType(v).(*types.Signature)
				walk(v.Body, s, where)
				return false
			case *ast.CallExpr:
				tv := tp.info.Types[v.Fun]
				switch {
				case tv.IsType():
					if len(v.Args) == 1 {
						check(v.Args[0], tv.Type, "conversion", where)
					}
				case tv.IsBuiltin():
					if id, ok := ast.Unparen(v.Fun).(*ast.Ident); ok && id.Name == "append" && len(v.Args) > 1 && !v.Ellipsis.IsValid() {
						if T := tp.valueType(v.Args[0]); T != nil {
							if s, ok := T.Underlying().(*types.Slice); ok {
								for _, a := range v.Args[1:] {
									check(a, s.Elem(), "append", where)
								}
							}
						}
					}
				case tv.Type != nil:
					s, ok := tv.Type.Underlying().(*types.Signature)
					if !ok {
						break
					}
					ps := s.Params()
					for i, a := range v.Args {
						var pt types.Type
						switch {
						case s.Variadic() && i >= ps.Len()-1:
							last := ps.At(ps.Len() - 1).Type()
							if v.Ellipsis.IsValid() {
								pt = last
							} else if sl, ok := last.Underlying().(*types.Slice); ok {
								pt = sl.Elem()
							}
						case i < ps.Len():
							pt = ps.At(i).Type()
						}
						check(a, pt, "argument", where)
					}
				}
			case *ast.AssignStmt:
				if v.Tok == token.ASSIGN && len(v.Lhs) == len(v.Rhs) {
					for i := range v.Lhs {
						check(v.Rhs[i], tp.valueType(v.Lhs[i]), "assignment", where)
					}
				}
			case *ast.ValueSpec:
				for i, nm := range v.Names {
					if i < len(v.Values) && v.Type != nil {
						if obj := tp.info.Defs[nm]; obj != nil {
							check(v.Values[i], obj.Type(), "declaration", where)
						}
					}
				}
			case *ast.ReturnStmt:
				if sig != nil && len(v.Results) == sig.Results().Len() {
					for i, r := range v.Results {
						check(r, sig.Results().At(i).Type(), "return", where)
					}
				}
			case *ast.CompositeLit:
				T := tp.valueType(v)
				if T == nil {
					break
				}
				switch u := T.Underlying().(type) {
				case *types.Struct:
					for i, el := range v.Elts {
						if kv, ok := el.(*ast.KeyValueExpr); ok {
							if k, ok := kv.Key.(*ast.Ident); ok {
								if f, ok := tp.info.Uses[k].(*types.Var); ok {
									check(kv.Value, f.Type(), "field", where)
								}
							}
						} else if i < u.NumFields() {
							check(el, u.Field(i).Type(), "field", where)
						}
					}
				case *types.Slice:
					for _, el := range v.Elts {
						if kv, ok := el.(*ast.KeyValueExpr); ok {
							el = kv.Value
						}
						check(el, u.Elem(), "element", where)
					}
				case *types.Array:
					for _, el := range v.Elts {
						if kv, ok := el.(*ast.KeyValueExpr); ok {
							el = kv.Value
						}
						check(el, u.Elem(), "element", where)
					}
				case *types.Map:
					for _, el := range v.Elts {
						if kv, ok := el.(*ast.KeyValueExpr); ok {
							check(kv.Key, u.Key(), "map key", where)
							check(kv.Value, u.Elem(), "map value", where)
						}
					}
				}
			case *ast.SendStmt:
				if T := tp.valueType(v.Chan); T != nil {
					if ch, ok := T.Underlying().(*types.Chan); ok {
						check(v.Value, ch.Elem(), "send", where)
					}
				}
			}
			return true
		})
	}
	tp.eachFunc(func(name string, fd *ast.FuncDecl, body ast.Node) {
		var sig *types.Signature
		if fd != nil {
			if fn, ok := tp.info.Defs[fd.Name].(*types.Func); ok {
				sig = fn.Type().(*types.Signature)
			}
		}
		walk(body, sig, name)
	})
	if len(bad) > 0 {
		sort.Strings(bad)
		t.Errorf("the bot identity escapes to where any goroutine can reach its fields (refresh() rewrites them on the main loop):\n  %s", strings.Join(bad, "\n  "))
	}
	if slots == 0 {
		t.Errorf("the interface-conversion walk checked 0 interface slots in package telegram — it walked nothing (re-anchor it)")
	}
}
