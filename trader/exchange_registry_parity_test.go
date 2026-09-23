package trader

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"
	"testing"

	"nofx/store"
)

// store's supported-exchange registry is what every load/probe/update path
// consults FIRST; NewAutoTrader's broker switch is what actually constructs.
// They must name exactly the same types — a type in the registry with no case
// would pass the gate and then die in the switch default; a case missing from
// the registry would be refused before it could construct.
func TestExchangeRegistryMatchesTheBrokerSwitch(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "auto_trader.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var cases []string
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "NewAutoTrader" || fd.Body == nil {
			continue
		}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			sw, ok := n.(*ast.SwitchStmt)
			if !ok {
				return true
			}
			sel, ok := sw.Tag.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Exchange" {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); !ok || x.Name != "config" {
				return true
			}
			for _, st := range sw.Body.List {
				cc := st.(*ast.CaseClause)
				for _, e := range cc.List {
					lit, ok := e.(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						t.Fatalf("%s: non-literal broker case", fset.Position(e.Pos()))
					}
					v, _ := strconv.Unquote(lit.Value)
					cases = append(cases, v)
				}
			}
			return false
		})
	}
	if len(cases) == 0 {
		t.Fatal("NewAutoTrader's `switch config.Exchange` not found — the parity check went vacuous")
	}
	reg := store.SupportedExchangeTypes()
	sort.Strings(cases)
	sort.Strings(reg)
	if strings.Join(cases, ",") != strings.Join(reg, ",") {
		t.Fatalf("broker switch cases %v != store registry %v", cases, reg)
	}
}

// ExchangeRefusal names the stored value and keeps the classifier substring;
// a registry type is never refused.
func TestExchangeRefusalIsNamed(t *testing.T) {
	for _, typ := range store.SupportedExchangeTypes() {
		if err := ExchangeRefusal("t", typ); err != nil {
			t.Errorf("%q is supported, refused: %v", typ, err)
		}
	}
	if err := ExchangeRefusal("t", "  "); err == nil || !strings.Contains(err.Error(), "no exchange configured") {
		t.Errorf("an empty exchange must be refused by name, got %v", err)
	}
	err := ExchangeRefusal("t", "retired-venue")
	if err == nil || !strings.Contains(err.Error(), `unsupported trading platform "retired-venue"`) || !strings.Contains(err.Error(), `unsupported exchange_type "retired-venue"`) {
		t.Errorf("an unknown exchange must be refused naming the stored value, got %v", err)
	}
	// Production call site: NewAutoTrader refuses before building anything.
	if at, err := NewAutoTrader(AutoTraderConfig{Name: "t", Exchange: "retired-venue"}, nil, "u"); err == nil || at != nil {
		t.Fatalf("NewAutoTrader must refuse an unknown exchange, got at=%v err=%v", at, err)
	}
}
