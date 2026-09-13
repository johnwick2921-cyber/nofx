package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// Loading persisted running traders starts goroutines immediately. Verify the
// production main call order, rather than a separate illustrative boot helper.
func TestBootIntegrityPrecedesTraderAutostart(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var checked, loaded token.Pos
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" {
			continue
		}
		for _, stmt := range fn.Body.List {
			ast.Inspect(stmt, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if sel.Sel.Name == "AssertBootIntegrity" {
					checked = call.Pos()
				}
				if sel.Sel.Name == "LoadTradersFromStore" {
					loaded = call.Pos()
				}
				return true
			})
		}
	}
	if checked == token.NoPos || loaded == token.NoPos || checked >= loaded {
		t.Fatalf("boot integrity at %v must precede autostart load at %v", fs.Position(checked), fs.Position(loaded))
	}
}
