package store

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// This grep/AST guard includes executable audit scripts wherever they live,
// plus new untracked source files. Captured JSON and historical prose are
// receipts, not executable readers. Broker order states are a different domain.
func armStateListViolations(path string, data []byte) []string {
	if path == "store/arm_state.go" {
		return nil
	} // the single definition
	if path == "provider/ninjatrader/order_state.go" || path == "docs/superpowers/reports/2026-09-05-vet-08-stretch-data/complete-0905/reaper.py" {
		return nil
	} // broker classifier / historical broker replay
	ext := filepath.Ext(path)
	if ext != ".go" && ext != ".py" && ext != ".sh" && ext != ".sql" {
		return nil
	}
	states := map[string]bool{}
	for _, state := range ArmStateNames() {
		states[state] = true
	}
	bad := []string{}
	sqlList := regexp.MustCompile(`(?is)\bstate\s+(?:not\s+)?in\s*\(`)
	words := strings.Join(ArmStateNames(), "|")
	pair := regexp.MustCompile(`(?i)['"](?:` + words + `)['"]\s*,\s*['"](?:` + words + `)['"]`)
	checkSQL := func(s string) {
		if sqlList.MatchString(s) || pair.MatchString(s) {
			bad = append(bad, "SQL state list; use store.NonTerminalArmStateSQL / cmd/arm-state-sql")
		}
	}
	if ext != ".go" {
		checkSQL(string(data))
		if pair.Match(data) {
			bad = append(bad, "copied arm state set")
		}
		return bad
	}
	f, err := parser.ParseFile(token.NewFileSet(), path, data, 0)
	if err != nil {
		return []string{"cannot inspect Go source: " + err.Error()}
	}
	stateName := func(e ast.Expr) string {
		if v, ok := e.(*ast.BasicLit); ok && v.Kind == token.STRING {
			s, _ := strconv.Unquote(v.Value)
			s = strings.ToLower(strings.TrimSpace(s))
			if states[s] {
				return s
			}
		}
		var name string
		switch v := e.(type) {
		case *ast.Ident:
			name = v.Name
		case *ast.SelectorExpr:
			name = v.Sel.Name
		}
		if strings.HasPrefix(name, "State") {
			name = strings.TrimPrefix(name, "State")
			for state := range states {
				if strings.EqualFold(strings.ReplaceAll(state, "_", ""), name) {
					return state
				}
			}
		}
		return ""
	}
	checkSet := func(exprs []ast.Expr) {
		seen := map[string]bool{}
		for _, e := range exprs {
			if name := stateName(e); name != "" {
				seen[name] = true
			} else {
				return
			}
		}
		if len(seen) > 1 {
			bad = append(bad, "copied arm state set; call store.IsTerminalArmState")
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.BasicLit:
			if v.Kind == token.STRING {
				s, _ := strconv.Unquote(v.Value)
				checkSQL(s)
			}
		case *ast.CaseClause:
			checkSet(v.List)
		case *ast.CompositeLit:
			if typ, ok := v.Type.(*ast.MapType); ok {
				if value, ok := typ.Value.(*ast.Ident); !ok || value.Name != "bool" {
					break
				}
			}
			es := []ast.Expr{}
			for _, e := range v.Elts {
				if kv, ok := e.(*ast.KeyValueExpr); ok {
					es = append(es, kv.Key)
				} else {
					es = append(es, e)
				}
			}
			checkSet(es)
		case *ast.BinaryExpr:
			if v.Op == token.LOR {
				es := []ast.Expr{}
				ast.Inspect(v, func(n ast.Node) bool {
					if b, ok := n.(*ast.BinaryExpr); ok && (b.Op == token.EQL || b.Op == token.NEQ) {
						// Broker Status aliases are not an armed_orders.State predicate.
						if sel, ok := b.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "Status" {
							return true
						}
						if sel, ok := b.Y.(*ast.SelectorExpr); ok && sel.Sel.Name == "Status" {
							return true
						}
						if stateName(b.X) != "" {
							es = append(es, b.X)
						}
						if stateName(b.Y) != "" {
							es = append(es, b.Y)
						}
					}
					return true
				})
				checkSet(es)
			}
		}
		return true
	})
	return bad
}

func TestArmStateNoRetypedLists(t *testing.T) {
	root := ".."
	cmd := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	cmd.Dir = root
	files, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range strings.Split(string(files), "\x00") {
		if path == "" {
			continue
		}
		switch filepath.Ext(path) {
		case ".go", ".py", ".sh", ".sql":
		default:
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		for _, violation := range armStateListViolations(path, data) {
			t.Errorf("%s: %s", path, violation)
		}
	}
}

func TestArmStateGrepRejectsRetypedListAnywhere(t *testing.T) {
	// Build the mutation from the canonical names; the mutation artifact really
	// contains a literal list, while the test itself owns no second state set.
	names := ArmStateNames()
	sql := "SELECT id FROM armed_orders WHERE " + "state " + "IN ('" + strings.Join(names, "','") + "')"
	goSource := "package stray\nfunc copied(s string) bool { switch s { case " + strconv.Quote(names[0]) + ", " + strconv.Quote(names[1]) + ": return true }; return false }"
	for _, path := range []string{"new-watch.py", "scripts/stray.sql", "docs/reports/audit.sh", "unrelated/new_reader.go"} {
		input := sql
		if strings.HasSuffix(path, ".go") {
			input = goSource
		}
		if got := armStateListViolations(path, []byte(input)); len(got) == 0 {
			t.Fatalf("retyped list escaped at %s: %s", path, input)
		}
	}
	for _, source := range []string{
		"package stray\nfunc copied(s string) bool { return s == " + strconv.Quote(names[0]) + " || s == " + strconv.Quote(names[1]) + " }",
		strings.ReplaceAll(goSource, strconv.Quote(names[0]), strconv.Quote(strings.ToUpper(names[0]))),
	} {
		if got := armStateListViolations("elsewhere/copied.go", []byte(source)); len(got) == 0 {
			t.Fatalf("copied classifier escaped: %s", source)
		}
	}
}
