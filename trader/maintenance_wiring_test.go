package trader

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ntwire "nofx/provider/ninjatrader"
	ntTrader "nofx/trader/ninjatrader"
)

// ── W-ONE-BUTTON M2 — THE WIRING IS THE THING UNDER TEST (canon 53) ────────
//
// A permit nobody installs is a hold that does not exist. NewAutoTrader cannot
// run here (its NT8 branch binds the live port 36974), so the wiring is proven
// three ways (review 3 F2 — the old whole-body AST walk passed with every call
// under 'if false'):
//  1. BEHAVIOUR: wireNT8Maintenance on a real TCPTrader over a real server
//     (here, and every drop test through newDropWire).
//  2. NewAutoTrader calls it DIRECTLY and UNCONDITIONALLY inside its
//     `if nt, ok := at.trader.(*ntTrader.TCPTrader); ok` block, which is a
//     top-level statement of NewAutoTrader.
//  3. It is the ONLY production caller of the four setters.

func parseFuncBody(t *testing.T, file, name string) []ast.Stmt {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name && fd.Recv == nil {
			return fd.Body.List
		}
	}
	t.Fatalf("%s not found in %s", name, file)
	return nil
}

func TestNewAutoTraderCallsWireNT8MaintenanceUnconditionally(t *testing.T) {
	var block *ast.IfStmt
	for _, st := range parseFuncBody(t, "auto_trader.go", "NewAutoTrader") {
		ifs, ok := st.(*ast.IfStmt)
		if !ok || ifs.Init == nil {
			continue
		}
		as, ok := ifs.Init.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != 2 || len(as.Rhs) != 1 {
			continue
		}
		ta, ok := as.Rhs[0].(*ast.TypeAssertExpr)
		if !ok || !isSelector(ta.X, "at", "trader") {
			continue
		}
		if star, ok := ta.Type.(*ast.StarExpr); ok && isSelector(star.X, "ntTrader", "TCPTrader") {
			block = ifs
		}
	}
	if block == nil {
		t.Fatal("NewAutoTrader has no top-level `if nt, ok := at.trader.(*ntTrader.TCPTrader); ok` block")
	}
	for _, st := range block.Body.List {
		es, ok := st.(*ast.ExprStmt)
		if !ok {
			continue
		}
		c, ok := es.X.(*ast.CallExpr)
		if !ok {
			continue
		}
		if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "wireNT8Maintenance" && len(c.Args) == 2 {
			a0, ok0 := c.Args[0].(*ast.Ident)
			a1, ok1 := c.Args[1].(*ast.Ident)
			if ok0 && ok1 && a0.Name == "at" && a1.Name == "nt" {
				return
			}
		}
	}
	t.Fatal("inside that block, NewAutoTrader must call wireNT8Maintenance(at, nt) as a direct, unconditional statement")
}

func isSelector(e ast.Expr, x, sel string) bool {
	s, ok := e.(*ast.SelectorExpr)
	if !ok || s.Sel.Name != sel {
		return false
	}
	id, ok := s.X.(*ast.Ident)
	return ok && id.Name == x
}

// wireNT8Maintenance itself calls all four setters as DIRECT, UNCONDITIONAL
// statements with the production arguments (M02: the calls under if false).
func TestWireNT8MaintenanceCallsAllFourSettersUnconditionally(t *testing.T) {
	want := map[string]string{
		"SetEntryPermit":       "MaintenanceEntryPermit",
		"SetEntryHoldCheck":    "maintenanceQueueHeld",
		"SetMaintenanceSource": "maintenanceWireState",
		"SetDroppedEntrySink":  "at.onMaintenanceDroppedEntry",
	}
	got := map[string]string{}
	for _, st := range parseFuncBody(t, "maintenance_wiring.go", "wireNT8Maintenance") {
		es, ok := st.(*ast.ExprStmt)
		if !ok {
			continue
		}
		c, ok := es.X.(*ast.CallExpr)
		if !ok || len(c.Args) != 1 {
			continue
		}
		s, ok := c.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if x, ok := s.X.(*ast.Ident); !ok || x.Name != "nt" {
			continue
		}
		switch a := c.Args[len(c.Args)-1].(type) {
		case *ast.Ident:
			got[s.Sel.Name] = a.Name
		case *ast.SelectorExpr:
			if x, ok := a.X.(*ast.Ident); ok {
				got[s.Sel.Name] = x.Name + "." + a.Sel.Name
			}
		}
	}
	for setter, arg := range want {
		if got[setter] != arg {
			t.Errorf("wireNT8Maintenance must call nt.%s(%s) as a direct, unconditional statement; found %q", setter, arg, got[setter])
		}
	}
}

// The four setters are called in production ONLY from wireNT8Maintenance (and
// forwarded to the server by the TCPTrader's own definitions).
func TestOnlyWireNT8MaintenanceCallsTheMaintenanceSetters(t *testing.T) {
	setters := map[string]bool{"SetEntryPermit": true, "SetEntryHoldCheck": true, "SetMaintenanceSource": true, "SetDroppedEntrySink": true}
	allowed := map[string]bool{"trader/maintenance_wiring.go": true, "trader/ninjatrader/tcp_trader.go": true}
	root, _ := filepath.Abs("..")
	var offenders []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "web", "vendor", ".claude", ".Codex", ".understand-anything":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		f, perr := parser.ParseFile(token.NewFileSet(), p, nil, 0)
		if perr != nil {
			return nil
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if s, ok := c.Fun.(*ast.SelectorExpr); ok && setters[s.Sel.Name] && !allowed[rel] {
					offenders = append(offenders, rel+": "+s.Sel.Name)
				}
			}
			return true
		})
		return nil
	})
	if len(offenders) > 0 {
		t.Fatalf("the maintenance setters may be called only from wireNT8Maintenance:\n%s", strings.Join(offenders, "\n"))
	}
}

// BEHAVIOUR: after wireNT8Maintenance, a hold refuses the entry at the permit
// and the AddOn is told it on connect.
func TestWireNT8MaintenanceInstallsThePermitAndTheWireSource(t *testing.T) {
	w := newDropWire(t) // calls wireNT8Maintenance(at, nt) — the production wiring
	_ = w.nt.SetStopLoss("MNQ", "LONG", 1, 28950)
	_ = w.nt.SetTakeProfit("MNQ", "LONG", 1, 29100)
	setHold(t, w.dir, "job-wire")
	// The PERMIT's own refusal — not the queue backstop (ErrEntryHeld), which
	// would also satisfy IsMaintenanceHold and hide an unwired permit.
	if _, err := w.nt.OpenLong("MNQ", 1, 1); !errors.Is(err, ntTrader.ErrMaintenanceHold) {
		t.Fatalf("site 4: with the hold wired, an entry must be refused AT THE PERMIT (ErrMaintenanceHold), got %v", err)
	}
	c := dialRaw(t, w.addr)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		_ = c.SetReadDeadline(deadline)
		env, err := ntwire.ReadFrame(c)
		if err != nil {
			break
		}
		if env.Type == ntwire.FrameMaintenance {
			return
		}
	}
	t.Fatal("site 7: with the hold wired, a connecting AddOn must be told the hold")
}

// maintenanceQueueHeld is exactly MaintenanceHeld's boolean.
func TestMaintenanceQueueHeldFollowsTheHoldFile(t *testing.T) {
	dir := withMaintenanceDir(t)
	if maintenanceQueueHeld() {
		t.Fatal("absent → not held")
	}
	setHold(t, dir, "job-q")
	if !maintenanceQueueHeld() {
		t.Fatal("held file → queue held")
	}
}

// maintenanceWireState is what the AddOn is told: absent → not held; held →
// the job id; a corrupt file → held with no job (fail-closed).
func TestMaintenanceWireStateFollowsTheHoldFile(t *testing.T) {
	dir := withMaintenanceDir(t)
	if held, job := maintenanceWireState(); held || job != "" {
		t.Fatalf("absent → (false, \"\"), got (%v, %q)", held, job)
	}
	setHold(t, dir, "job-wire")
	if held, job := maintenanceWireState(); !held || job != "job-wire" {
		t.Fatalf("held → (true, job-wire), got (%v, %q)", held, job)
	}
	if err := writeRaw(dir, "{not json"); err != nil {
		t.Fatal(err)
	}
	if held, job := maintenanceWireState(); !held || job != "" {
		t.Fatalf("corrupt → (true, \"\"), got (%v, %q)", held, job)
	}
}
