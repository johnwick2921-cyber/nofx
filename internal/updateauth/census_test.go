package updateauth

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"nofx/internal/censuswalk"
)

// ── W-ONE-BUTTON M3 census: who may touch the enrollment, and who may mint ──
//
// Scans every non-test Go file in the module (the ONE root-only walk,
// internal/censuswalk — M3 fold M5):
//
//  1. the enrollment/seen file names are spelled ONLY in paths.go — judged on
//     every string literal AND every constant-folded concatenation run
//     ("device"+".key", a const + "_ids.json", dir + "/ad" + "min.json"),
//     and the key file's name may not be spelled even in FRAGMENTS a
//     variable could join (a path element "device" / "device.*" / "*.key");
//  2. only EXACT importers may import this package — the update handler, the
//     Server wiring, the attended CLI's one file, and the M4 worker package
//     by its exact directory (never a prefix: "internal/updater" also covered
//     the APP-linked internal/updaterwire — red-team 3 #1(b)); never a dot or
//     blank import;
//  3. every exported identifier an outside file references is CLASSIFIED:
//     restricted ones (enroll, mint, load the key/admin, consume, verify,
//     and every enrollment path helper) are admitted per FILE; open ones
//     (types, errors, limits, validators) to any admitted importer; an
//     unclassified one is refused until someone classifies it (CTO ruling
//     Q1(a): nothing API-side mints a MAC; M3 spec: NO API creates, resets
//     or reads the enrollment);
//  4. crypto/hmac is imported ONLY by this package among the packages that
//     reach the updater data dir (an updater/installpath/holdcli import, a
//     data-dir or updater-dir helper, or a path element "updater") — a MAC
//     minted beside the key's directory is a mint, whatever it calls.
//
// WHAT THIS CANNOT PROVE (M3 fold M4 — stated, not implied): it is a
// syntactic census over identifiers, imports and constant strings. A file
// that builds the path at run time (fmt.Sprintf with a non-literal, byte
// arithmetic, a directory listing), receives the key bytes or a path through
// an interface or a function value handed to it by an admitted file,
// hand-rolls HMAC over crypto/sha256, or reaches the updater dir through a
// package the census does not relate to it, passes. The app process runs as
// the same uid that owns device.key, so nothing but review and this tripwire
// stops app code from reading the key; the census makes the direct
// spellings and the likely drift (a helper reused, a prefix admission, a MAC
// beside the key) fail loudly.
var (
	// updateAuthImporterFiles may import the package (exact files).
	updateAuthImporterFiles = map[string]bool{
		"api/handler_updates.go":                 true,
		"api/server.go":                          true,
		"internal/updaterbootstrap/bootstrap.go": true,
	}
	// updateAuthImporterPackages may import the package (exact directories —
	// never a prefix). The M4 worker gets no restricted identifier.
	updateAuthImporterPackages = map[string]bool{
		"internal/updaterworker": true,
	}
	// updateAuthRestricted: identifier → the ONLY files outside the package
	// that may reference it. An empty set = nobody outside.
	updateAuthRestricted = map[string]map[string]bool{
		"Enroll":        {"internal/updaterbootstrap/bootstrap.go": true},
		"Authorize":     {"internal/updaterbootstrap/bootstrap.go": true},
		"ComputeMAC":    {},
		"Message":       {},
		"LoadDeviceKey": {"api/handler_updates.go": true},
		"LoadAdmin":     {"api/handler_updates.go": true},
		"VerifyMAC":     {"api/handler_updates.go": true},
		"Consume":       {"api/handler_updates.go": true},
		"Dir":           {"internal/updaterbootstrap/bootstrap.go": true},
		"AdminPath":     {"internal/updaterbootstrap/bootstrap.go": true},
		"DeviceKeyPath": {"internal/updaterbootstrap/bootstrap.go": true},
		"SeenPath":      {},
	}
	// updateAuthOpen: identifiers any admitted importer may reference.
	updateAuthOpen = map[string]bool{
		"Admin": true, "Grant": true, "Manifest": true, "Verifier": true, "StubVerifier": true,
		"ErrAlreadyEnrolled": true, "ErrBadClock": true, "ErrExpired": true, "ErrMalformed": true,
		"ErrNoDataDir": true, "ErrNoVerifiedManifest": true, "ErrNotEnrolled": true, "ErrPrunedReplay": true,
		"ErrReplay": true, "ErrSeenCorrupt": true, "ErrSeenFull": true, "ErrUnsafe": true,
		"DeviceKeyLen": true, "MaxAuthorizationWindow": true, "MaxEmailLen": true, "MaxJobIDLen": true,
		"MaxReleaseIDLen": true, "MaxSeenEntries": true, "MaxUserIDLen": true, "SeenRetention": true,
		"ValidReleaseID": true, "ValidJobID": true, "NewJobID": true, "ParseInstallRequest": true, "CheckExpiry": true,
	}
	// updaterAreaImports: importing one of these (or a subpackage) reaches the
	// updater data dir.
	updaterAreaImports = []string{
		"internal/updateauth", "internal/updaterwire", "internal/updaterworker",
		"internal/updaterbootstrap", "internal/installpath", "internal/holdcli",
	}
	// updaterAreaIdents: referencing one of these names reaches the data dir
	// or the updater dir.
	updaterAreaIdents = map[string]bool{
		"MaintenanceDataDir": true, "UpdaterDirName": true, "SocketPath": true,
		"SocketPathFor": true, "MaintenanceHoldPath": true, "DataDirFor": true,
	}
)

// macPackageHome is the ONE package that may compute or verify a MAC beside
// the updater dir.
const macPackageHome = "internal/updateauth"

func TestUpdateAuthCensus(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	offenders, scanned, err := updateAuthOffenders(root)
	if err != nil {
		t.Fatal(err)
	}
	if scanned < 100 {
		t.Fatalf("scan saw only %d files — the walk is not covering the module", scanned)
	}
	if len(offenders) > 0 {
		t.Fatalf("update-authorization census:\n%s", strings.Join(offenders, "\n"))
	}
}

// The admission tables are pinned: widening any of them is a reviewed act,
// and every name in them is a real exported identifier of this package (a
// typo would admit nothing and restrict nothing).
func TestUpdateAuthCensusTablesArePinned(t *testing.T) {
	keys := func(m map[string]bool) string {
		var out []string
		for k, v := range m {
			if v {
				out = append(out, k)
			}
		}
		sort.Strings(out)
		return strings.Join(out, ",")
	}
	if got, want := keys(updateAuthImporterFiles), "api/handler_updates.go,api/server.go,internal/updaterbootstrap/bootstrap.go"; got != want {
		t.Fatalf("importer files = %s, want exactly %s", got, want)
	}
	if got, want := keys(updateAuthImporterPackages), "internal/updaterworker"; got != want {
		t.Fatalf("importer packages = %s, want exactly %s", got, want)
	}
	var restricted []string
	for name, files := range updateAuthRestricted {
		restricted = append(restricted, name+"="+keys(files))
	}
	sort.Strings(restricted)
	want := "AdminPath=internal/updaterbootstrap/bootstrap.go;Authorize=internal/updaterbootstrap/bootstrap.go;ComputeMAC=;" +
		"Consume=api/handler_updates.go;DeviceKeyPath=internal/updaterbootstrap/bootstrap.go;Dir=internal/updaterbootstrap/bootstrap.go;" +
		"Enroll=internal/updaterbootstrap/bootstrap.go;LoadAdmin=api/handler_updates.go;LoadDeviceKey=api/handler_updates.go;" +
		"Message=;SeenPath=;VerifyMAC=api/handler_updates.go"
	if got := strings.Join(restricted, ";"); got != want {
		t.Fatalf("restricted identifiers =\n%s\nwant exactly\n%s", got, want)
	}
	// every classified name exists; nothing is both restricted and open
	exported := map[string]bool{}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			switch x := d.(type) {
			case *ast.FuncDecl:
				if x.Recv == nil {
					exported[x.Name.Name] = x.Name.IsExported()
				}
			case *ast.GenDecl:
				for _, s := range x.Specs {
					switch sp := s.(type) {
					case *ast.TypeSpec:
						exported[sp.Name.Name] = sp.Name.IsExported()
					case *ast.ValueSpec:
						for _, n := range sp.Names {
							exported[n.Name] = n.IsExported()
						}
					}
				}
			}
		}
	}
	for name := range updateAuthRestricted {
		if !exported[name] {
			t.Errorf("restricted %s is not an exported identifier of updateauth", name)
		}
		if updateAuthOpen[name] {
			t.Errorf("%s is both restricted and open", name)
		}
	}
	for name := range updateAuthOpen {
		if !exported[name] {
			t.Errorf("open %s is not an exported identifier of updateauth", name)
		}
	}
}

// updateAuthOffenders is the census over every non-test .go file under root.
func updateAuthOffenders(root string) (offenders []string, scanned int, err error) {
	module, err := censuswalk.ModulePath(root)
	if err != nil {
		return nil, 0, err
	}
	const literalHome = "internal/updateauth/paths.go"
	fileNames := []string{adminFileName, deviceKeyName, seenFileName, enrollLockName, seenLockName}
	importers := func(rel string) bool {
		return updateAuthImporterFiles[rel] || updateAuthImporterPackages[path.Dir(rel)]
	}
	type pkgFacts struct {
		hmacFiles  []string
		updaterWhy string // first reason the package reaches the updater dir; "" = none
	}
	pkgs := map[string]*pkgFacts{}
	seen := map[string]bool{}
	offend := func(line string) {
		if !seen[line] {
			seen[line] = true
			offenders = append(offenders, line)
		}
	}

	files, err := censuswalk.NonTestGoFiles(root)
	if err != nil {
		return nil, 0, err
	}
	for _, file := range files {
		rel := file.Rel
		f, perr := parser.ParseFile(token.NewFileSet(), file.Path, nil, 0)
		if perr != nil {
			offend(rel + ": cannot be parsed, so it cannot be checked (" + perr.Error() + ")")
			continue
		}
		scanned++
		dir := path.Dir(rel)
		facts := pkgs[dir]
		if facts == nil {
			facts = &pkgFacts{}
			pkgs[dir] = facts
		}
		reach := func(why string) {
			if facts.updaterWhy == "" {
				facts.updaterWhy = rel + " " + why
			}
		}

		// 2. imports; 4. crypto/hmac and updater-area imports
		alias := ""
		for _, im := range f.Imports {
			ip, _ := strconv.Unquote(im.Path.Value)
			if ip == "crypto/hmac" {
				facts.hmacFiles = append(facts.hmacFiles, rel)
			}
			for _, a := range updaterAreaImports {
				if ip == module+"/"+a || strings.HasPrefix(ip, module+"/"+a+"/") {
					reach("imports " + ip)
				}
			}
			if ip != module+"/internal/updateauth" {
				continue
			}
			if !importers(rel) {
				offend(rel + ": imports nofx/internal/updateauth")
			}
			alias = "updateauth"
			if im.Name != nil {
				alias = im.Name.Name
				if alias == "." || alias == "_" {
					offend(rel + ": " + alias + "-imports nofx/internal/updateauth")
				}
			}
		}

		// 1. literals and constant-folded runs; 4. a path element "updater"
		for _, v := range constantStrings(f) {
			if rel != literalHome {
				for _, name := range fileNames {
					if strings.Contains(v, name) {
						offend(rel + ": spells " + name)
					}
				}
			}
			for _, e := range strings.FieldsFunc(v, func(r rune) bool { return r == '/' || r == '\\' }) {
				if rel != literalHome && (e == "device" || strings.HasPrefix(e, "device.") || strings.HasSuffix(e, ".key")) {
					offend(rel + ": spells a fragment of device.key (" + strconv.Quote(e) + ")")
				}
				if e == updaterDirName {
					reach("names the path element \"updater\"")
				}
			}
		}

		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident: // 4. data-dir / updater-dir helpers (bare, or a selector's Sel)
				if updaterAreaIdents[x.Name] {
					reach("references " + x.Name)
				}
			case *ast.SelectorExpr: // 3. classified references (any reference, not only calls)
				if id, ok := x.X.(*ast.Ident); ok && alias != "" && id.Name == alias {
					name := x.Sel.Name
					if allowed, restricted := updateAuthRestricted[name]; restricted {
						if !allowed[rel] {
							offend(rel + ": references updateauth." + name)
						}
					} else if !updateAuthOpen[name] {
						offend(rel + ": references unclassified updateauth." + name + " — classify it in the census tables")
					}
				}
			}
			return true
		})
	}

	// 4. crypto/hmac beside the updater dir, per package
	var dirs []string
	for d := range pkgs {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	for _, d := range dirs {
		facts := pkgs[d]
		if len(facts.hmacFiles) == 0 || facts.updaterWhy == "" || d == macPackageHome {
			continue
		}
		ip := module
		if d != "." {
			ip = module + "/" + d
		}
		for _, h := range facts.hmacFiles {
			offend(h + ": imports crypto/hmac in package " + ip + ", which references the updater data dir (" + facts.updaterWhy + ") — only " + macPackageHome + " computes a MAC there")
		}
	}
	return offenders, scanned, nil
}

// constantStrings returns every string the file spells as a constant: each
// string literal, and each run of adjacent constant operands in every
// maximal `+` chain (so "device"+".key", a file-local const + "_ids.json" and
// the "/ad"+"min.json" inside dir+"/ad"+"min.json" all fold). Names bound in
// the file to a foldable string (const, var, :=, =) fold too — by name, not
// by scope, which can only over-report.
func constantStrings(f *ast.File) []string {
	env := map[string]string{}
	bind := func(names []*ast.Ident, values []ast.Expr) bool {
		changed := false
		if len(names) != len(values) {
			return false
		}
		for i, n := range names {
			if v, ok := foldString(values[i], env); ok {
				if old, had := env[n.Name]; !had || old != v {
					env[n.Name] = v
					changed = true
				}
			}
		}
		return changed
	}
	for pass := 0; pass < 4; pass++ {
		changed := false
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.ValueSpec:
				changed = bind(x.Names, x.Values) || changed
			case *ast.AssignStmt:
				var names []*ast.Ident
				for _, l := range x.Lhs {
					id, ok := l.(*ast.Ident)
					if !ok {
						return true
					}
					names = append(names, id)
				}
				changed = bind(names, x.Rhs) || changed
			}
			return true
		})
		if !changed {
			break
		}
	}
	var out []string
	inner := map[*ast.BinaryExpr]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if b, ok := n.(*ast.BinaryExpr); ok && b.Op == token.ADD {
			for _, side := range []ast.Expr{b.X, b.Y} {
				if c, ok := unparen(side).(*ast.BinaryExpr); ok && c.Op == token.ADD {
					inner[c] = true
				}
			}
		}
		return true
	})
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				if v, err := strconv.Unquote(x.Value); err == nil {
					out = append(out, v)
				}
			}
		case *ast.BinaryExpr:
			if x.Op != token.ADD || inner[x] {
				return true
			}
			run, have := "", false
			for _, op := range flattenAdd(x) {
				if v, ok := foldString(op, env); ok {
					run, have = run+v, true
					continue
				}
				if have {
					out = append(out, run)
				}
				run, have = "", false
			}
			if have {
				out = append(out, run)
			}
		}
		return true
	})
	return out
}

func unparen(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}

func flattenAdd(e ast.Expr) []ast.Expr {
	if b, ok := unparen(e).(*ast.BinaryExpr); ok && b.Op == token.ADD {
		return append(flattenAdd(b.X), flattenAdd(b.Y)...)
	}
	return []ast.Expr{e}
}

func foldString(e ast.Expr, env map[string]string) (string, bool) {
	switch x := unparen(e).(type) {
	case *ast.BasicLit:
		if x.Kind == token.STRING {
			v, err := strconv.Unquote(x.Value)
			return v, err == nil
		}
	case *ast.Ident:
		v, ok := env[x.Name]
		return v, ok
	case *ast.BinaryExpr:
		if x.Op == token.ADD {
			a, ok1 := foldString(x.X, env)
			b, ok2 := foldString(x.Y, env)
			if ok1 && ok2 {
				return a + b, true
			}
		}
	}
	return "", false
}
