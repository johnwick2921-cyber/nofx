package updaterwire

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"nofx/internal/censuswalk"
)

// ── M4 3b-B U2: a resume is built ONLY by the attended updater CLI ─────────
//
// resume continues a job the worker parked for an attended step (the owner's
// F5 at nt8_updated); dispatch §0: it "resumes only on an attended
// `nofx-updater resume <job>`". The app never sends one: the API side hands
// off an install and reads status, nothing else. So the names that BUILD a
// resume request may appear in exactly two directories:
//
//   - internal/updaterwire — the wire package, which defines them;
//   - cmd/nofx-updater     — the CLI the owner types the job id into.
//
// Both are EXACT directories, never prefixes (CTO ruling 1790258770876:
// census admissions extend by exact names): cmd/nofx-updaterx, a
// subdirectory of cmd/nofx-updater and internal/updaterwire/wireserver are
// all outside.
//
// "Build" is judged by NAME, fail closed: any reference to NewResume,
// VerbResume or ResumePayload of the wire package (under its own name, a
// renamed import or a dot import), and any string literal that spells a
// resume frame ("verb":"resume"). ResumePayload is in the set because a
// resume Request validates only with a non-nil *ResumePayload — naming the
// type is the one other way to build one. A READ of the verb counts too: a
// worker-side handler dispatches on the payload pointer (r.Resume != nil),
// which Validate makes equivalent, and names none of these.

// resumeBuilders are the wire package's names that build a resume request.
var resumeBuilders = map[string]bool{"NewResume": true, "VerbResume": true, "ResumePayload": true}

// resumeAdmittedDirs are the EXACT module-relative directories whose non-test
// files may name a resume builder, each with its reason.
var resumeAdmittedDirs = map[string]string{
	"internal/updaterwire": "the wire package defines the resume verb",
	"cmd/nofx-updater":     "the attended `nofx-updater resume <job>` CLI (M4 3b-B dispatch §0/§3) — the only sender",
}

// resumeWireDir is where the builders must be DEFINED; the census checks they
// are, so a rename cannot leave it judging names that no longer exist.
const resumeWireDir = "internal/updaterwire"

// resumeFrameRe matches a hand-spelled resume frame inside a string literal.
var resumeFrameRe = regexp.MustCompile(`"verb"\s*:\s*"resume"`)

type resumeCensus struct {
	offenders []string
	scanned   int
	defined   map[string]bool // builder names declared at top level of resumeWireDir
}

// resumeBuilderCensus scans every non-test .go file under root (the ONE
// root-only walk, internal/censuswalk) and reports each file outside
// resumeAdmittedDirs that names a resume builder of <module>/internal/updaterwire
// or spells a resume frame.
func resumeBuilderCensus(root string) (resumeCensus, error) {
	c := resumeCensus{defined: map[string]bool{}}
	module, err := censuswalk.ModulePath(root)
	if err != nil {
		return c, err
	}
	wirePath := module + "/" + resumeWireDir
	files, err := censuswalk.NonTestGoFiles(root)
	if err != nil {
		return c, err
	}
	hits := map[string]bool{}
	for _, file := range files {
		rel := file.Rel
		f, perr := parser.ParseFile(token.NewFileSet(), file.Path, nil, parser.SkipObjectResolution)
		if perr != nil {
			hits[rel+": cannot be parsed"] = true
			continue
		}
		c.scanned++
		dir := path.Dir(rel)
		if dir == resumeWireDir {
			for _, d := range f.Decls {
				switch d := d.(type) {
				case *ast.FuncDecl:
					if d.Recv == nil && resumeBuilders[d.Name.Name] {
						c.defined[d.Name.Name] = true
					}
				case *ast.GenDecl:
					for _, s := range d.Specs {
						switch s := s.(type) {
						case *ast.TypeSpec:
							if resumeBuilders[s.Name.Name] {
								c.defined[s.Name.Name] = true
							}
						case *ast.ValueSpec:
							for _, n := range s.Names {
								if resumeBuilders[n.Name] {
									c.defined[n.Name] = true
								}
							}
						}
					}
				}
			}
		}
		if _, admitted := resumeAdmittedDirs[dir]; admitted {
			continue
		}
		local := map[string]bool{} // this file's names for the wire package
		dot := false
		for _, imp := range f.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil || p != wirePath {
				continue
			}
			switch {
			case imp.Name == nil:
				local[path.Base(wirePath)] = true
			case imp.Name.Name == ".":
				dot = true
			case imp.Name.Name != "_":
				local[imp.Name.Name] = true
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				if id, ok := x.X.(*ast.Ident); ok && local[id.Name] && resumeBuilders[x.Sel.Name] {
					hits[rel+": names updaterwire."+x.Sel.Name] = true
				}
			case *ast.Ident:
				if dot && resumeBuilders[x.Name] {
					hits[rel+": names updaterwire."+x.Name] = true
				}
			case *ast.BasicLit:
				if x.Kind == token.STRING {
					if s, err := strconv.Unquote(x.Value); err == nil && resumeFrameRe.MatchString(s) {
						hits[rel+": spells a resume frame"] = true
					}
				}
			}
			return true
		})
	}
	for h := range hits {
		c.offenders = append(c.offenders, h)
	}
	sort.Strings(c.offenders)
	return c, nil
}

// PIN (M4 3b-B U2, dispatch §0/§3): over the REAL tree, nothing but the wire
// package and cmd/nofx-updater names a resume builder or spells a resume
// frame. At this head cmd/nofx-updater does not exist yet, so this proves no
// OTHER package references them; the admitted set is that exact directory.
//
// The name carries "Census" so the standard gate
// (go test -run 'Census|Guard|Walk|Link') runs this real-tree scan; it keeps
// the brief's name as its prefix, so -run TestOnlyTheUpdaterCLIBuildsAResume
// (the name wire.go cites) still selects it.
func TestOnlyTheUpdaterCLIBuildsAResumeCensus(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	c, err := resumeBuilderCensus(root)
	if err != nil {
		t.Fatal(err)
	}
	if c.scanned < 100 {
		t.Fatalf("census scanned only %d files", c.scanned)
	}
	for name := range resumeBuilders {
		if !c.defined[name] {
			t.Fatalf("%s is not declared in %s — the census would judge a name that no longer exists (update resumeBuilders with the rename)", name, resumeWireDir)
		}
	}
	if len(c.offenders) > 0 {
		t.Fatalf("a resume is built only by the attended updater CLI (cmd/nofx-updater); the app never sends one:\n%s", strings.Join(c.offenders, "\n"))
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "nofx-updater")); err != nil {
		t.Logf("cmd/nofx-updater absent at this head (%v): no package in the tree builds a resume", err)
	}
}

// PIN: the census sees every way to name a builder, and admits only the two
// exact directories. Each case plants ONE file in a synthetic module
// (t.TempDir, never the real tree) and is judged by the PRODUCTION census
// function; each has its expected single offender, and the positive controls
// have none.
func TestResumeBuilderCensusSeesEveryForm(t *testing.T) {
	const wireGo = "package updaterwire\n\ntype Verb string\n\nconst VerbResume Verb = \"resume\"\n\n" +
		"type ResumePayload struct{ JobID string }\n\ntype Request struct {\n\tVerb   Verb\n\tResume *ResumePayload\n}\n\n" +
		"func NewResume(j string) Request { return Request{Verb: VerbResume, Resume: &ResumePayload{JobID: j}} }\n"
	census := func(t *testing.T, rel, body string) resumeCensus {
		t.Helper()
		root := t.TempDir()
		files := map[string]string{
			"go.mod":                       "module nofx\n\ngo 1.25\n",
			"internal/updaterwire/wire.go": wireGo,
		}
		if rel != "" {
			files[rel] = body
		}
		for r, b := range files {
			p := filepath.Join(root, filepath.FromSlash(r))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(b), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		c, err := resumeBuilderCensus(root)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	imp := func(pkg, name, use string) string {
		return "package " + pkg + "\n\nimport " + name + " \"nofx/internal/updaterwire\"\n\n" + use + "\n"
	}
	ctor := `var _ = updaterwire.NewResume("job-0001abcd")`

	// the wire package alone: no offender, all three builders seen as defined
	if c := census(t, "", ""); len(c.offenders) != 0 || len(c.defined) != 3 {
		t.Fatalf("definer only: offenders=%v defined=%v", c.offenders, c.defined)
	}

	// positive controls: the admitted directories, and a handler that
	// dispatches on the payload pointer
	for _, ok := range []struct{ name, rel, body string }{
		{"the CLI", "cmd/nofx-updater/main.go", imp("main", "", "func main() {\n\t_ = updaterwire.NewResume(\"job-0001abcd\")\n\t_ = updaterwire.Request{Verb: updaterwire.VerbResume, Resume: &updaterwire.ResumePayload{}}\n}")},
		{"the CLI, a second file", "cmd/nofx-updater/resume.go", imp("main", "uw", "const frame = `{\"v\":1,\"verb\":\"resume\",\"payload\":{}}`\n\nvar _ = uw.VerbResume")},
		{"the wire package itself", "internal/updaterwire/more.go", "package updaterwire\n\nvar _ = NewResume(\"job-0001abcd\")\n\nconst f = `{\"verb\":\"resume\"}`\n"},
		{"a handler reading the payload pointer", "internal/updaterworker/socket.go", imp("updaterworker", "", "func isResume(r updaterwire.Request) bool { return r.Resume != nil }")},
		{"an unrelated resume word", "agent/x.go", "package agent\n\nvar words = []string{\"resume\", \"continue\"}\n"},
	} {
		t.Run("admitted "+ok.name, func(t *testing.T) {
			if c := census(t, ok.rel, ok.body); len(c.offenders) != 0 {
				t.Fatalf("%s must be admitted, census offenders = %v", ok.rel, c.offenders)
			}
		})
	}

	cases := []struct{ name, rel, body, want string }{
		{"the app, constructor", "api/resume.go", imp("api", "", ctor), "names updaterwire.NewResume"},
		{"the app, verb const", "api/resume.go", imp("api", "", "var _ = updaterwire.Request{Verb: updaterwire.VerbResume}"), "names updaterwire.VerbResume"},
		{"the app, payload type", "api/resume.go", imp("api", "", "var _ = new(updaterwire.ResumePayload)"), "names updaterwire.ResumePayload"},
		{"a read of the verb", "internal/updaterworker/socket.go", imp("updaterworker", "", "func isResume(r updaterwire.Request) bool { return r.Verb == updaterwire.VerbResume }"), "names updaterwire.VerbResume"},
		{"renamed import", "api/resume.go", imp("api", "uw", `var _ = uw.NewResume("job-0001abcd")`), "names updaterwire.NewResume"},
		{"dot import", "api/resume.go", imp("api", ".", `var _ = NewResume("job-0001abcd")`), "names updaterwire.NewResume"},
		{"hand-spelled frame", "api/resume.go", "package api\n\nconst f = `{\"v\":1, \"verb\" : \"resume\",\"payload\":{\"job_id\":\"job-0001abcd\"}}`\n", "spells a resume frame"},
		{"escaped frame", "api/resume.go", "package api\n\nconst f = \"{\\\"verb\\\":\\\"resume\\\"}\"\n", "spells a resume frame"},
		{"name-prefix sibling of the CLI", "cmd/nofx-updaterx/main.go", imp("main", "", ctor), "names updaterwire.NewResume"},
		{"subdirectory of the CLI", "cmd/nofx-updater/sub/x.go", imp("sub", "", ctor), "names updaterwire.NewResume"},
		{"the worker-side wire subpackage", "internal/updaterwire/wireserver/x.go", imp("wireserver", "", ctor), "names updaterwire.NewResume"},
		{"the trading app", "trader/x.go", imp("trader", "", ctor), "names updaterwire.NewResume"},
	}
	for _, dir := range censuswalk.NestedProbeDirs() {
		cases = append(cases, struct{ name, rel, body, want string }{
			"nested " + dir, dir + "/resume.go", imp(censuswalk.PackageName(dir), "", ctor), "names updaterwire.NewResume",
		})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := census(t, tc.rel, tc.body)
			if want := tc.rel + ": " + tc.want; len(c.offenders) != 1 || c.offenders[0] != want {
				t.Fatalf("census offenders = %v, want exactly [%s]", c.offenders, want)
			}
		})
	}
}
