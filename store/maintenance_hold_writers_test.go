package store

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"nofx/internal/censuswalk"
)

// ── W-ONE-BUTTON M2 site 6 — nothing in the trading API writes the hold ────
//
// The hold is written and cleared ONLY by the local operator CLI (and, from
// M4, the updater worker — admitted BY NAME in M3, below). No API route —
// resume, clear-freeze, anything future — may lift an installation-wide hold.
// This scans every non-test Go file in the module for calls to the three
// writers.
//
// M3 admission (design note §4.5, fail-closed default Q5): the worker's hold
// writer is admitted as the exact file internal/updaterworker/hold.go, which
// does not exist yet. An admitted-but-absent path grants nothing; when M4
// creates it, only that one file may write, and
// TestTradingAppNeverLinksTheUpdaterWorkerSide keeps the trading app from
// ever importing it.
var (
	holdWriterFiles = map[string]bool{
		"store/maintenance_hold.go":      true, // the definitions
		"internal/holdcli/holdcli.go":    true, // cmd/maintenance-hold
		"internal/updaterworker/hold.go": true, // M4 updater worker (admitted M3, by name)
	}
	// M2.1 (review 3 F13): names alone let os.Remove / os.WriteFile on the hold
	// path through. The path itself is confined too: MaintenanceHoldPath only in
	// the store, the CLI and the worker, the literal "hold.json" only in the store.
	holdPathUserFiles = map[string]bool{
		"store/maintenance_hold.go":      true,
		"internal/holdcli/holdcli.go":    true,
		"internal/updaterworker/hold.go": true, // M4 updater worker (admitted M3, by name)
	}
)

func TestOnlyTheOperatorCLIWritesTheMaintenanceHold(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	offenders, scanned, err := holdWriterOffenders(root)
	if err != nil {
		t.Fatal(err)
	}
	if scanned < 100 {
		t.Fatalf("scan saw only %d files — the walk is not covering the module", scanned)
	}
	if len(offenders) > 0 {
		t.Fatalf("the maintenance hold may be written/cleared only by the operator CLI (and the M4 updater, when it lands):\n%s", strings.Join(offenders, "\n"))
	}
}

// The census is proved on a synthetic module (M3): the worker's hold writer
// is admitted by its EXACT repo path. internal/updaterworker/hold.go calling
// the writers and MaintenanceHoldPath is clean (positive control); the same
// code under any other path — another file in the worker package, a deeper
// file with the same base name, the worker binary's main, the app's update
// handler, the wire — offends; and even the admitted file may not name
// "hold.json" (that literal stays in the store).
func TestHoldWriterCensusAdmitsTheWorkerOnlyByName(t *testing.T) {
	const worker = "package updaterworker\n\nimport \"nofx/store\"\n\n" +
		"func HoldForJob(d string) error {\n\t_ = store.MaintenanceHoldPath(d)\n\treturn store.WriteMaintenanceHold(d, store.MaintenanceHold{})\n}\n\n" +
		"func ReleaseJob(d string) error { return store.ClearMaintenanceHold(d) }\n"
	write := func(root, rel, body string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	base := func() string {
		root := t.TempDir()
		write(root, "go.mod", "module nofx\n\ngo 1.25\n")
		write(root, "store/maintenance_hold.go", "package store\n\nconst holdFile = \"hold.json\"\n\n"+
			"func MaintenanceHoldPath(d string) string { return d + \"/\" + holdFile }\n\n"+
			"func WriteMaintenanceHold(d string, h any) error { return nil }\n")
		write(root, "internal/holdcli/holdcli.go", "package holdcli\n\nimport \"nofx/store\"\n\nfunc Set(d string) error { return store.WriteMaintenanceHold(d, nil) }\n")
		write(root, "internal/updaterworker/hold.go", worker)
		write(root, "api/handler_updates.go", "package api\n")
		return root
	}
	// positive control: the admitted worker file writes, clears and resolves the path
	root := base()
	off, scanned, err := holdWriterOffenders(root)
	if err != nil || len(off) != 0 || scanned != 4 {
		t.Fatalf("clean synthetic module: offenders=%v scanned=%d err=%v (want none, 4 files)", off, scanned, err)
	}
	for name, c := range map[string]struct{ rel, body, want string }{
		"other file in the worker package": {"internal/updaterworker/other.go", strings.Replace(worker, "HoldForJob", "H2", 1), "internal/updaterworker/other.go: WriteMaintenanceHold"},
		"same base name, deeper path":      {"internal/updaterworker/sub/hold.go", worker, "internal/updaterworker/sub/hold.go: WriteMaintenanceHold"},
		"the worker binary's main":         {"cmd/nofx-updater/main.go", "package main\n\nimport \"nofx/store\"\n\nfunc main() { store.ClearMaintenanceHold(\"d\") }\n", "cmd/nofx-updater/main.go: ClearMaintenanceHold"},
		"the app's update handler":         {"api/handler_updates.go", "package api\n\nimport \"nofx/store\"\n\nfunc clear() { store.ForceClearMaintenanceHold(\"d\") }\n", "api/handler_updates.go: ForceClearMaintenanceHold"},
		"the wire resolving the hold path": {"internal/updaterwire/dial.go", "package updaterwire\n\nimport \"nofx/store\"\n\nvar p = store.MaintenanceHoldPath(\"d\")\n", "internal/updaterwire/dial.go: references MaintenanceHoldPath"},
		"admitted file naming hold.json":   {"internal/updaterworker/hold.go", worker + "\nvar raw = \"updater/hold.json\"\n", "internal/updaterworker/hold.go: names the hold file"},
	} {
		t.Run(name, func(t *testing.T) {
			root := base()
			write(root, c.rel, c.body)
			off, _, err := holdWriterOffenders(root)
			if err != nil {
				t.Fatal(err)
			}
			hit := false
			for _, o := range off {
				hit = hit || strings.HasPrefix(o, c.want)
			}
			if !hit {
				t.Fatalf("offenders = %v, want one starting %q", off, c.want)
			}
		})
	}
}

// holdWriterOffenders scans every non-test .go file under root and reports
// each call to a hold writer, each MaintenanceHoldPath reference and each
// "hold.json" literal outside the admitted files (repo-relative slash paths).
func holdWriterOffenders(root string) (offenders []string, scanned int, err error) {
	allowed := holdWriterFiles
	writers := map[string]bool{"WriteMaintenanceHold": true, "ClearMaintenanceHold": true, "ForceClearMaintenanceHold": true}
	pathUsers := holdPathUserFiles
	// M3 fold M5: the ONE root-only walk (censuswalk) — a skip-named dir below
	// the root (api/web, internal/node_modules/x, …) is a compiled package.
	files, err := censuswalk.NonTestGoFiles(root)
	if err != nil {
		return nil, 0, err
	}
	for _, file := range files {
		p, rel := file.Path, file.Rel
		f, perr := parser.ParseFile(token.NewFileSet(), p, nil, 0)
		if perr != nil {
			offenders = append(offenders, rel+": cannot be parsed, so it cannot be checked ("+perr.Error()+")")
			continue
		}
		scanned++
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := ""
			switch fn := call.Fun.(type) {
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			case *ast.Ident:
				name = fn.Name
			}
			if writers[name] && !allowed[rel] {
				offenders = append(offenders, rel+": "+name)
			}
			return true
		})
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				if x.Sel.Name == "MaintenanceHoldPath" && !pathUsers[rel] {
					offenders = append(offenders, rel+": references MaintenanceHoldPath")
				}
			case *ast.Ident:
				if x.Name == "MaintenanceHoldPath" && !pathUsers[rel] {
					offenders = append(offenders, rel+": references MaintenanceHoldPath")
				}
			case *ast.BasicLit:
				if strings.Contains(x.Value, "hold.json") && rel != "store/maintenance_hold.go" {
					offenders = append(offenders, rel+": names the hold file (\"hold.json\")")
				}
			}
			return true
		})
	}
	return offenders, scanned, nil
}

// The admission list is pinned exactly: widening it is a reviewed act, not a
// drive-by line in some other PR.
func TestHoldWriterAdmissionsArePinned(t *testing.T) {
	want := []string{"internal/holdcli/holdcli.go", "internal/updaterworker/hold.go", "store/maintenance_hold.go"}
	for name, m := range map[string]map[string]bool{"writers": holdWriterFiles, "path users": holdPathUserFiles} {
		var got []string
		for k, v := range m {
			if v {
				got = append(got, k)
			}
		}
		sort.Strings(got)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("hold %s admitted = %v, want exactly %v", name, got, want)
		}
	}
	for _, api := range []string{"api/handler_updates.go", "api/handler_maintenance.go", "internal/updaterwire/dial.go", "internal/updaterwire/wireserver/server.go"} {
		if holdWriterFiles[api] || holdPathUserFiles[api] {
			t.Fatalf("%s must never be admitted as a hold writer — the app never writes the hold", api)
		}
	}
}

// ── W-ONE-BUTTON M3 — the trading app never links the worker side ─────────
//
// The app DIALS the updater worker (nofx/internal/updaterwire); only the
// worker binary may LISTEN (nofx/internal/updaterwire/wireserver) or hold
// the worker's hold writer (nofx/internal/updaterworker, M4). If api/,
// trader/, kernel/, agent/, telegram/, store/ or the root main package could
// reach either — directly or through any chain of module packages — an
// app-side bug could serve forged worker verbs or write the hold. (store/ is
// in the design note's list; the worker's hold.go imports store, so the
// reverse edge must never appear.) Build tags are ignored (every non-test
// file counts), which can only over-report.
//
// M3 fold M4 (red-team 4 #2(b)): the attended CLI's package
// nofx/internal/updaterbootstrap is forbidden too — its Run reaches
// updateauth.Authorize → ComputeMAC, the one door that mints an install MAC,
// and nothing on the app side may mint (CTO ruling Q1(a)). Only its own
// binary, cmd/updater-bootstrap, links it.
var (
	tradingAppDirs          = []string{"api", "trader", "kernel", "agent", "telegram", "store"}
	forbiddenWorkerPackages = []string{"nofx/internal/updaterwire/wireserver", "nofx/internal/updaterworker", "nofx/internal/updaterbootstrap",
		// M4 3b-B U5b (f), OQ-8/C2 (CTO 1790259689740): the kill/restart
		// library — only the worker binary links it
		// (TestWorkerImportGuardRefusesTheActivationLibrary).
		"nofx/internal/activation"}
)

func TestTradingAppNeverLinksTheUpdaterWorkerSide(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	offenders, guarded, err := workerImportOffenders(root)
	if err != nil {
		t.Fatal(err)
	}
	// every guarded root must actually have been walked — a walk that found
	// nothing (wrong module name, a skipped dir) would otherwise pass vacuously
	seen := map[string]bool{}
	for _, g := range guarded {
		seen[g] = true
	}
	for _, want := range append([]string{"nofx"}, prefixed("nofx/", tradingAppDirs)...) {
		if !seen[want] {
			t.Fatalf("guarded root %s was not walked (walked %d packages: %v) — the guard is not covering the app", want, len(guarded), guarded)
		}
	}
	if len(offenders) > 0 {
		t.Fatalf("the trading app must never link the updater worker side:\n%s", strings.Join(offenders, "\n"))
	}
}

// The guard itself is proved on a synthetic module: a direct import, a
// transitive import and a root-main import are each caught, and the same
// tree with the offending import removed is clean (positive control).
func TestWorkerImportGuardCatchesDirectAndTransitiveImports(t *testing.T) {
	write := func(root, rel, body string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	base := func() string {
		root := t.TempDir()
		write(root, "go.mod", "module nofx\n\ngo 1.25\n")
		write(root, "internal/updaterwire/dial.go", "package updaterwire\n")
		write(root, "internal/updaterwire/wireserver/server.go", "package wireserver\nimport _ \"nofx/internal/updaterwire\"\n")
		write(root, "internal/updaterworker/hold.go", "package updaterworker\n")
		write(root, "internal/helper/h.go", "package helper\n")
		write(root, "api/server.go", "package api\nimport _ \"nofx/internal/updaterwire\"\nimport _ \"nofx/internal/helper\"\n")
		write(root, "trader/t.go", "package trader\n")
		write(root, "main.go", "package main\nimport _ \"nofx/api\"\n")
		write(root, "cmd/updater-worker/main.go", "package main\nimport _ \"nofx/internal/updaterwire/wireserver\"\nimport _ \"nofx/internal/updaterworker\"\n")
		return root
	}
	// positive control: the app dials, the worker binary listens — clean
	root := base()
	if off, guarded, err := workerImportOffenders(root); err != nil || len(off) != 0 || strings.Join(guarded, ",") != "nofx,nofx/api,nofx/trader" {
		t.Fatalf("clean synthetic module: offenders=%v guarded=%v err=%v (want none; guarded nofx, nofx/api, nofx/trader)", off, guarded, err)
	}
	for name, c := range map[string]struct{ rel, body, want string }{
		"direct api":        {"api/worker.go", "package api\nimport _ \"nofx/internal/updaterwire/wireserver\"\n", "api"},
		"transitive helper": {"internal/helper/h.go", "package helper\nimport _ \"nofx/internal/updaterworker\"\n", "api"},
		"trader":            {"trader/w.go", "package trader\nimport w \"nofx/internal/updaterworker\"\nvar _ = w.X\n", "trader"},
		"root main":         {"main_worker.go", "package main\nimport _ \"nofx/internal/updaterwire/wireserver\"\n", "(root main)"},
		"store":             {"store/s.go", "package store\nimport _ \"nofx/internal/updaterworker\"\n", "store"},
	} {
		t.Run(name, func(t *testing.T) {
			root := base()
			write(root, c.rel, c.body)
			off, _, err := workerImportOffenders(root)
			if err != nil {
				t.Fatal(err)
			}
			hit := false
			for _, o := range off {
				hit = hit || strings.HasPrefix(o, c.want+":")
			}
			if !hit {
				t.Fatalf("offenders = %v, want one starting %q", off, c.want+":")
			}
		})
	}
	// a _test.go file linking the worker is not production linkage
	root = base()
	write(root, "api/worker_test.go", "package api\nimport _ \"nofx/internal/updaterwire/wireserver\"\n")
	if off, _, _ := workerImportOffenders(root); len(off) != 0 {
		t.Fatalf("a test-only import must not count as production linkage: %v", off)
	}
}

// workerImportOffenders builds the module's package import graph from every
// non-test .go file under root (module path from root/go.mod) and reports,
// for each trading-app package, any path to a forbidden worker package.
// guarded lists the trading-app packages that were checked, sorted.
func workerImportOffenders(root string) (offenders []string, guarded []string, err error) {
	modBytes, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, nil, err
	}
	module := ""
	for _, line := range strings.Split(string(modBytes), "\n") {
		if f := strings.Fields(line); len(f) == 2 && f[0] == "module" {
			module = f[1]
		}
	}
	if module == "" {
		return nil, nil, fmt.Errorf("no module line in go.mod")
	}
	imports := map[string]map[string]bool{} // import path → module-internal imports
	// M3 fold M5: the ONE root-only walk (censuswalk). testdata is walked too:
	// a package under x/testdata/ is importable and linked like any other.
	files, err := censuswalk.NonTestGoFiles(root)
	if err != nil {
		return nil, nil, err
	}
	for _, file := range files {
		p := file.Path
		relDir, _ := filepath.Rel(root, filepath.Dir(p))
		pkg := module
		if relDir != "." {
			pkg = module + "/" + filepath.ToSlash(relDir)
		}
		f, perr := parser.ParseFile(token.NewFileSet(), p, nil, parser.ImportsOnly)
		if perr != nil {
			offenders = append(offenders, file.Rel+": cannot be parsed, so its imports cannot be checked")
			continue
		}
		if imports[pkg] == nil {
			imports[pkg] = map[string]bool{}
		}
		for _, im := range f.Imports {
			ip, _ := strconv.Unquote(im.Path.Value)
			if ip == module || strings.HasPrefix(ip, module+"/") {
				imports[pkg][ip] = true
			}
		}
	}
	forbidden := isForbiddenWorkerPackage
	var starts []string
	for pkg := range imports {
		if pkg == module {
			starts = append(starts, pkg)
			continue
		}
		for _, d := range tradingAppDirs {
			if pkg == module+"/"+d || strings.HasPrefix(pkg, module+"/"+d+"/") {
				starts = append(starts, pkg)
				break
			}
		}
	}
	sort.Strings(starts)
	for _, start := range starts {
		// BFS with parent links so the report names the chain
		parent := map[string]string{start: ""}
		queue := []string{start}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			if forbidden(cur) {
				chain := []string{cur}
				for p := parent[cur]; p != ""; p = parent[p] {
					chain = append([]string{p}, chain...)
				}
				label := strings.TrimPrefix(strings.TrimPrefix(start, module), "/")
				if label == "" {
					label = "(root main)"
				}
				offenders = append(offenders, label+": "+strings.Join(chain, " → "))
				break
			}
			var next []string
			for ip := range imports[cur] {
				next = append(next, ip)
			}
			sort.Strings(next)
			for _, ip := range next {
				if _, seen := parent[ip]; !seen {
					parent[ip] = cur
					queue = append(queue, ip)
				}
			}
		}
	}
	return offenders, starts, nil
}

func isForbiddenWorkerPackage(ip string) bool {
	for _, f := range forbiddenWorkerPackages {
		if ip == f || strings.HasPrefix(ip, f+"/") {
			return true
		}
	}
	return false
}

// ── M3 fold M5: the same guard, answered by the toolchain ─────────────────
//
// workerImportOffenders models the import graph from source (every non-test
// file, build tags ignored — it can only over-report). This asks the Go
// toolchain what the trading app ACTUALLY links in the default build context
// (`go list -deps` of the root main package and every package under the
// trading-app dirs): whatever the source model misses — a directory the walk
// skipped, an import it mis-resolved — the linker's own answer cannot.
// patterns is what was asked (a vacuity check for callers).
func toolchainWorkerLinkOffenders(root string) (offenders []string, patterns []string, err error) {
	module, err := censuswalk.ModulePath(root)
	if err != nil {
		return nil, nil, err
	}
	patterns = []string{"."}
	for _, d := range tradingAppDirs {
		if fi, serr := os.Stat(filepath.Join(root, d)); serr == nil && fi.IsDir() {
			patterns = append(patterns, "./"+d+"/...")
		}
	}
	pkgs, err := censuswalk.ListPackages(root, true, patterns...)
	if err != nil {
		return nil, patterns, err
	}
	for _, p := range pkgs {
		if p.Module != module {
			continue
		}
		if p.Error != "" && !strings.Contains(p.Error, "build constraints exclude all Go files") {
			offenders = append(offenders, p.ImportPath+": the toolchain could not load it, so its links are unknown ("+p.Error+")")
			continue
		}
		if isForbiddenWorkerPackage(p.ImportPath) {
			offenders = append(offenders, p.ImportPath+": linked by the trading app (go list -deps "+strings.Join(patterns, " ")+")")
		}
	}
	sort.Strings(offenders)
	return offenders, patterns, nil
}

func TestTradingAppLinkageFromTheToolchain(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	offenders, patterns, err := toolchainWorkerLinkOffenders(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) != 1+len(tradingAppDirs) {
		t.Fatalf("asked the toolchain for %v — every trading-app dir must exist and be asked about", patterns)
	}
	if len(offenders) > 0 {
		t.Fatalf("the toolchain says the trading app links the updater worker side:\n%s", strings.Join(offenders, "\n"))
	}
}

func prefixed(prefix string, xs []string) []string {
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		out = append(out, prefix+x)
	}
	return out
}

// ── M4 3b-B U5b (g) — only the operator CLI ever SETS withdraw_entries ─────
//
// Owner rule (dispatch §0/§2, CTO 1790258770876): the updater never cancels
// orders, so the worker's hold NEVER carries withdraw_entries. The hold-writer
// census above admits internal/updaterworker/hold.go as a WRITER by exact
// name; that admission does not extend to this field. Only the operator CLI
// (internal/holdcli/holdcli.go — cmd/maintenance-hold --withdraw) may set it;
// store/maintenance_hold.go only DEFINES it (the struct tag). Reading it
// (trader/withdraw.go, the worker's own foreign-hold check) is fine.
//
// A "set" is any of: a composite-literal key WithdrawEntries / a
// "withdraw_entries" key; an assignment (any operator) to x.WithdrawEntries;
// taking &x.WithdrawEntries; a POSITIONAL MaintenanceHold{...} literal (it sets
// every field); or a string literal naming withdraw_entries (raw JSON) outside
// the two admitted files.
var (
	withdrawSetterFiles  = map[string]bool{"internal/holdcli/holdcli.go": true}
	withdrawLiteralFiles = map[string]bool{"internal/holdcli/holdcli.go": true, "store/maintenance_hold.go": true}
)

func TestOnlyTheOperatorCLISetsWithdrawEntries(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	offenders, scanned, err := withdrawSetterOffenders(root)
	if err != nil {
		t.Fatal(err)
	}
	if scanned < 100 {
		t.Fatalf("scan saw only %d files — the walk is not covering the module", scanned)
	}
	if len(offenders) > 0 {
		t.Fatalf("withdraw_entries may be set only by the operator CLI (internal/holdcli/holdcli.go):\n%s", strings.Join(offenders, "\n"))
	}
}

// Proved on a synthetic module: the CLI's set and every read are clean
// (positive control), and each way of setting the field elsewhere — including
// in the worker's census-ADMITTED hold writer — is caught.
func TestWithdrawSetterCensusCatchesEveryForm(t *testing.T) {
	const worker = "package updaterworker\n\nimport \"nofx/store\"\n\n" +
		"func holdFor(j string) store.MaintenanceHold {\n\treturn store.MaintenanceHold{Held: true, JobID: j, Owner: \"updater\"}\n}\n\n" +
		"func ours(st store.MaintenanceHoldState, j string) bool {\n\treturn st.Hold.JobID == j && !st.Hold.WithdrawEntries\n}\n"
	base := func() string {
		root := t.TempDir()
		censusWrite(t, root, "go.mod", "module nofx\n\ngo 1.25\n")
		censusWrite(t, root, "store/maintenance_hold.go", "package store\n\ntype MaintenanceHold struct {\n\tHeld bool `json:\"held\"`\n\tJobID string `json:\"job_id\"`\n\tOwner string `json:\"owner,omitempty\"`\n\tWithdrawEntries bool `json:\"withdraw_entries,omitempty\"`\n}\n\ntype MaintenanceHoldState struct{ Hold MaintenanceHold }\n")
		censusWrite(t, root, "internal/holdcli/holdcli.go", "package holdcli\n\nimport (\n\t\"fmt\"\n\t\"nofx/store\"\n)\n\nfunc Set(w bool) store.MaintenanceHold {\n\th := store.MaintenanceHold{Held: true, WithdrawEntries: w}\n\tfmt.Printf(\"withdraw_entries=%v\\n\", h.WithdrawEntries)\n\th.WithdrawEntries = w\n\treturn h\n}\n")
		censusWrite(t, root, "internal/updaterworker/hold.go", worker)
		censusWrite(t, root, "trader/withdraw.go", "package trader\n\nimport \"nofx/store\"\n\nfunc wants(st store.MaintenanceHoldState) bool { return st.Hold.Held && st.Hold.WithdrawEntries }\n")
		return root
	}
	root := base()
	if off, scanned, err := withdrawSetterOffenders(root); err != nil || len(off) != 0 || scanned != 4 {
		t.Fatalf("clean synthetic module: offenders=%v scanned=%d err=%v (want none, 4 files)", off, scanned, err)
	}
	for name, c := range map[string]struct{ rel, body, want string }{
		"the admitted hold writer setting it in its literal": {"internal/updaterworker/hold.go", strings.Replace(worker, `Owner: "updater"}`, `Owner: "updater", WithdrawEntries: true}`, 1), "internal/updaterworker/hold.go: composite literal sets WithdrawEntries"},
		"an assignment in the worker":                        {"internal/updaterworker/set.go", "package updaterworker\n\nimport \"nofx/store\"\n\nfunc f(h *store.MaintenanceHold) { h.WithdrawEntries = true }\n", "internal/updaterworker/set.go: assigns WithdrawEntries"},
		"an op-assignment":                                   {"internal/updaterworker/set.go", "package updaterworker\n\nimport \"nofx/store\"\n\nfunc f(h *store.MaintenanceHold, b bool) { h.WithdrawEntries = h.WithdrawEntries || b }\n", "internal/updaterworker/set.go: assigns WithdrawEntries"},
		"the field's address taken":                          {"cmd/nofx-updater/main.go", "package main\n\nimport \"nofx/store\"\n\nfunc main() { var h store.MaintenanceHold; p := &h.WithdrawEntries; *p = true }\n", "cmd/nofx-updater/main.go: takes the address of WithdrawEntries"},
		"a positional literal":                               {"internal/updaterworker/pos.go", "package updaterworker\n\nimport \"nofx/store\"\n\nvar h = store.MaintenanceHold{true, \"j\", \"updater\", true}\n", "internal/updaterworker/pos.go: positional MaintenanceHold literal"},
		"raw JSON naming the key":                            {"cmd/nofx-updater/main.go", "package main\n\nconst raw = `{\"held\":true,\"withdraw_entries\":true}`\n\nfunc main() {}\n", "cmd/nofx-updater/main.go: names withdraw_entries"},
		"a map literal keyed withdraw_entries":               {"api/handler_updates.go", "package api\n\nvar m = map[string]any{\"withdraw_entries\": true}\n", "api/handler_updates.go: composite literal sets withdraw_entries"},
	} {
		t.Run(name, func(t *testing.T) {
			root := base()
			censusWrite(t, root, c.rel, c.body)
			off, _, err := withdrawSetterOffenders(root)
			if err != nil {
				t.Fatal(err)
			}
			hit := false
			for _, o := range off {
				hit = hit || strings.HasPrefix(o, c.want)
			}
			if !hit {
				t.Fatalf("offenders = %v, want one starting %q", off, c.want)
			}
		})
	}
}

// The admissions are pinned exactly: the worker (any file) is never one.
func TestWithdrawSetterAdmissionsArePinned(t *testing.T) {
	for name, pair := range map[string]struct {
		m    map[string]bool
		want string
	}{
		"setters":  {withdrawSetterFiles, "internal/holdcli/holdcli.go"},
		"literals": {withdrawLiteralFiles, "internal/holdcli/holdcli.go,store/maintenance_hold.go"},
	} {
		var got []string
		for k, v := range pair.m {
			if v {
				got = append(got, k)
			}
		}
		sort.Strings(got)
		if strings.Join(got, ",") != pair.want {
			t.Fatalf("withdraw %s admitted = %v, want exactly %s", name, got, pair.want)
		}
	}
	for f := range withdrawSetterFiles {
		if strings.HasPrefix(f, "internal/updaterworker/") || strings.HasPrefix(f, "cmd/nofx-updater/") || strings.HasPrefix(f, "internal/updaterjob/") || strings.HasPrefix(f, "api/") {
			t.Fatalf("%s must never be admitted to set withdraw_entries — the updater never cancels orders", f)
		}
	}
}

func withdrawSetterOffenders(root string) (offenders []string, scanned int, err error) {
	files, err := censuswalk.NonTestGoFiles(root)
	if err != nil {
		return nil, 0, err
	}
	isField := func(e ast.Expr) bool {
		sel, ok := e.(*ast.SelectorExpr)
		return ok && sel.Sel.Name == "WithdrawEntries"
	}
	isHoldType := func(e ast.Expr) bool {
		switch x := e.(type) {
		case *ast.SelectorExpr:
			return x.Sel.Name == "MaintenanceHold"
		case *ast.Ident:
			return x.Name == "MaintenanceHold"
		}
		return false
	}
	for _, file := range files {
		rel := file.Rel
		f, perr := parser.ParseFile(token.NewFileSet(), file.Path, nil, 0)
		if perr != nil {
			offenders = append(offenders, rel+": cannot be parsed, so it cannot be checked ("+perr.Error()+")")
			continue
		}
		scanned++
		setter := withdrawSetterFiles[rel]
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CompositeLit:
				for _, el := range x.Elts {
					kv, ok := el.(*ast.KeyValueExpr)
					if !ok {
						if isHoldType(x.Type) && !setter {
							offenders = append(offenders, rel+": positional MaintenanceHold literal (sets every field, withdraw_entries included)")
							return true
						}
						continue
					}
					switch k := kv.Key.(type) {
					case *ast.Ident:
						if k.Name == "WithdrawEntries" && !setter {
							offenders = append(offenders, rel+": composite literal sets WithdrawEntries")
						}
					case *ast.BasicLit:
						if strings.Contains(k.Value, "withdraw_entries") && !setter {
							offenders = append(offenders, rel+": composite literal sets withdraw_entries")
						}
					}
				}
			case *ast.AssignStmt:
				for _, l := range x.Lhs {
					if isField(l) && !setter {
						offenders = append(offenders, rel+": assigns WithdrawEntries")
					}
				}
			case *ast.IncDecStmt:
				if isField(x.X) && !setter {
					offenders = append(offenders, rel+": assigns WithdrawEntries")
				}
			case *ast.UnaryExpr:
				if x.Op == token.AND && isField(x.X) && !setter {
					offenders = append(offenders, rel+": takes the address of WithdrawEntries")
				}
			case *ast.BasicLit:
				if x.Kind == token.STRING && strings.Contains(x.Value, "withdraw_entries") && !withdrawLiteralFiles[rel] {
					offenders = append(offenders, rel+": names withdraw_entries in a string literal")
				}
			}
			return true
		})
	}
	return offenders, scanned, nil
}
