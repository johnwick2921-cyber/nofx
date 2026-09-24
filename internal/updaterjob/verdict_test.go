package updaterjob

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// The verdict reader is app-linkable, so it must not be able to WRITE a
// verdict: no call in verdict.go may create, link, rename, chmod or remove a
// file, and it imports no package that could (the worker's FetchRelease is the
// one writer — internal/updaterworker, which the trading app cannot link).
// The behaviour (FetchRelease writes, ReadVerdict reads) is pinned at the
// worker's call sites in internal/updaterworker/release_test.go.
func TestVerdictFileHasNoWriter(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "verdict.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	allowedImports := map[string]bool{
		"bytes": true, "encoding/json": true, "errors": true, "fmt": true, "io": true, "io/fs": true,
		"os": true, "path/filepath": true, "regexp": true, "strings": true, "syscall": true, "time": true,
		"nofx/internal/updaterwire": true,
	}
	for _, im := range f.Imports {
		p, _ := strconv.Unquote(im.Path.Value)
		if !allowedImports[p] {
			t.Errorf("verdict.go imports %s — the verdict reader imports only the standard library it reads with and updaterwire", p)
		}
	}
	writers := map[string]bool{
		"WriteFile": true, "Create": true, "CreateTemp": true, "Link": true, "Symlink": true, "Rename": true,
		"Mkdir": true, "MkdirAll": true, "MkdirTemp": true, "Remove": true, "RemoveAll": true, "Chmod": true,
		"Chown": true, "Lchown": true, "Chtimes": true, "Truncate": true, "Write": true, "WriteString": true,
		"Mkfifo": true, "Mknod": true, "Unlink": true,
	}
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if ok && writers[sel.Sel.Name] {
			t.Errorf("verdict.go references %s — the app-linkable verdict reader may not write", sel.Sel.Name)
		}
		if id, ok := n.(*ast.Ident); ok && (id.Name == "O_WRONLY" || id.Name == "O_RDWR" || id.Name == "O_CREATE" || id.Name == "O_CREAT" || id.Name == "O_TRUNC" || id.Name == "O_APPEND") {
			t.Errorf("verdict.go opens with %s — the app-linkable verdict reader may not write", id.Name)
		}
		return true
	})
}
