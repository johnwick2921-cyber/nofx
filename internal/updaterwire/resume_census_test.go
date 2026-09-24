package updaterwire

import (
	"encoding/json"
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

// resumeFrameRe matches a hand-spelled resume frame (or a fragment of one)
// inside a string literal. Case-insensitive, because json.Unmarshal into a
// Request matches "Verb" as readily as "verb" (U2 verifier defect 2).
var resumeFrameRe = regexp.MustCompile(`(?i)"verb"\s*:\s*"resume`)

// resumeValueRe is the same rule for a decoded JSON "verb" value.
var resumeValueRe = regexp.MustCompile(`(?i)^resume`)

// spellsResume reports whether a string literal's (Go-unquoted) value spells a
// resume frame: the raw fragment rule, or — when the value is JSON — a "verb"
// key whose value DECODES to resume anywhere in it, or a "resume" key (the
// payload json.Unmarshal would put in Request.Resume). JSON escapes (a
// unicode escape inside the key or the value) are what DecodeRequest reads,
// not what a regex sees (U2 verifier defect 1).
func spellsResume(s string) bool {
	if resumeFrameRe.MatchString(s) {
		return true
	}
	var v any
	if json.Unmarshal([]byte(s), &v) != nil {
		return false
	}
	return jsonSpellsResume(v)
}

// jsonSpellsResume walks a decoded JSON value; a string inside it is judged
// again, so a frame encoded inside a JSON string is seen too.
func jsonSpellsResume(v any) bool {
	switch v := v.(type) {
	case map[string]any:
		for k, e := range v {
			if s, ok := e.(string); ok && strings.EqualFold(k, "verb") && resumeValueRe.MatchString(s) {
				return true
			}
			if strings.EqualFold(k, "resume") {
				return true
			}
			if jsonSpellsResume(e) {
				return true
			}
		}
	case []any:
		for _, e := range v {
			if jsonSpellsResume(e) {
				return true
			}
		}
	case string:
		return spellsResume(v)
	}
	return false
}

const (
	resumeWriteReason   = "writes a Resume field"
	resumePointerReason = "takes a pointer to an updaterwire.Request"
)

// throughResume reports whether e reaches a Resume field (X.Resume, or
// anything below it such as X.Resume.JobID) — the shape a write to it has.
func throughResume(e ast.Expr) bool {
	for {
		switch x := e.(type) {
		case *ast.SelectorExpr:
			if x.Sel.Name == "Resume" {
				return true
			}
			e = x.X
		case *ast.ParenExpr:
			e = x.X
		case *ast.StarExpr:
			e = x.X
		case *ast.IndexExpr:
			e = x.X
		default:
			return false
		}
	}
}

// resumeWrites reports how f fills a Resume field without naming a builder
// (U2 verifier defect 2); local and dot are f's names for the wire package.
//
//   - Every file: a composite key Resume, &X.Resume, an assignment to
//     X.Resume, or anything below it. Fail closed: Resume is a field of the
//     wire Request alone today, so an unrelated one is renamed, not admitted.
//   - Files importing the wire package: an unkeyed Request literal (its last
//     element IS Resume), and any pointer to a Request — *Request, &Request{},
//     new(Request), or &x for a name that holds one (declared with the type,
//     or assigned a Request literal or any call into the wire package) —
//     because a decoder needs a pointer to fill Resume, and nothing in the
//     wire API takes one (Client.Do and the worker's Handler take values).
//
// READS are not judged: r.Resume != nil and r.Resume.JobID stay admitted.
func resumeWrites(f *ast.File, local map[string]bool, dot bool) []string {
	isRequest := func(e ast.Expr) bool {
		switch x := e.(type) {
		case *ast.SelectorExpr:
			id, ok := x.X.(*ast.Ident)
			return ok && local[id.Name] && x.Sel.Name == "Request"
		case *ast.Ident:
			return dot && x.Name == "Request"
		}
		return false
	}
	fromWire := func(e ast.Expr) bool {
		switch x := e.(type) {
		case *ast.CompositeLit:
			return isRequest(x.Type)
		case *ast.CallExpr:
			switch fn := x.Fun.(type) {
			case *ast.SelectorExpr:
				id, ok := fn.X.(*ast.Ident)
				return ok && local[id.Name]
			case *ast.Ident: // a dot import makes the wire package's calls bare
				return dot && ast.IsExported(fn.Name)
			}
		}
		return false
	}
	holds := map[string]bool{} // names that hold a wire Request (scope-blind, fail closed)
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Field:
			if isRequest(x.Type) {
				for _, nm := range x.Names {
					holds[nm.Name] = true
				}
			}
		case *ast.ValueSpec:
			for i, nm := range x.Names {
				if (x.Type != nil && isRequest(x.Type)) || (i < len(x.Values) && fromWire(x.Values[i])) ||
					(i == 0 && len(x.Values) == 1 && fromWire(x.Values[0])) {
					holds[nm.Name] = true
				}
			}
		case *ast.AssignStmt:
			for i, l := range x.Lhs {
				nm, ok := l.(*ast.Ident)
				if ok && ((len(x.Rhs) == len(x.Lhs) && fromWire(x.Rhs[i])) || (i == 0 && len(x.Rhs) == 1 && fromWire(x.Rhs[0]))) {
					holds[nm.Name] = true
				}
			}
		}
		return true
	})
	found := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.KeyValueExpr:
			if id, ok := x.Key.(*ast.Ident); ok && id.Name == "Resume" {
				found[resumeWriteReason] = true
			}
		case *ast.AssignStmt:
			for _, l := range x.Lhs {
				if throughResume(l) {
					found[resumeWriteReason] = true
				}
			}
		case *ast.UnaryExpr:
			if x.Op != token.AND {
				break
			}
			if throughResume(x.X) {
				found[resumeWriteReason] = true
			}
			switch y := x.X.(type) {
			case *ast.Ident:
				if holds[y.Name] {
					found[resumePointerReason] = true
				}
			case *ast.CompositeLit:
				if isRequest(y.Type) {
					found[resumePointerReason] = true
				}
			}
		case *ast.StarExpr:
			if isRequest(x.X) {
				found[resumePointerReason] = true
			}
		case *ast.CallExpr:
			if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "new" && len(x.Args) == 1 && isRequest(x.Args[0]) {
				found[resumePointerReason] = true
			}
		case *ast.CompositeLit:
			if isRequest(x.Type) && len(x.Elts) > 0 {
				if _, keyed := x.Elts[0].(*ast.KeyValueExpr); !keyed {
					found[resumeWriteReason] = true
				}
			}
		}
		return true
	})
	out := make([]string, 0, len(found))
	for r := range found {
		out = append(out, r)
	}
	return out
}

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
					if s, err := strconv.Unquote(x.Value); err == nil && spellsResume(s) {
						hits[rel+": spells a resume frame"] = true
					}
				}
			}
			return true
		})
		for _, why := range resumeWrites(f, local, dot) {
			hits[rel+": "+why] = true
		}
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

// resumeCensusWireGo is the synthetic module's wire package: the three
// builders, and the Request field a write can fill.
const resumeCensusWireGo = "package updaterwire\n\ntype Verb string\n\nconst VerbResume Verb = \"resume\"\n\n" +
	"type ResumePayload struct{ JobID string }\n\ntype Request struct {\n\tVerb   Verb\n\tResume *ResumePayload\n}\n\n" +
	"func NewResume(j string) Request { return Request{Verb: VerbResume, Resume: &ResumePayload{JobID: j}} }\n"

// resumeCensusOf plants files (module-relative path → body) in a synthetic
// module (t.TempDir, never the real tree) beside resumeCensusWireGo, and runs
// the PRODUCTION census on it. A planted internal/updaterwire/wire.go
// replaces the synthetic one.
func resumeCensusOf(t *testing.T, files map[string]string) resumeCensus {
	t.Helper()
	root := t.TempDir()
	all := map[string]string{
		"go.mod":                       "module nofx\n\ngo 1.25\n",
		"internal/updaterwire/wire.go": resumeCensusWireGo,
	}
	for r, b := range files {
		all[r] = b
	}
	for r, b := range all {
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

// resumeCensusImp is a one-file body importing the wire package as name
// ("" = its own name, "." = dot import).
func resumeCensusImp(pkg, name, use string) string {
	return "package " + pkg + "\n\nimport " + name + " \"nofx/internal/updaterwire\"\n\n" + use + "\n"
}

// resumeJSONU spells a JSON unicode escape (backslash, u, four hex) at run
// time, so no editor or tool that decodes escapes can quietly turn a case
// back into the raw form the regex already sees — the cases guard it.
func resumeJSONU(hex string) string { return `\` + "u" + hex }

// wantResumeOffenders fails unless the census judged exactly want (reasons,
// each prefixed with rel) — no more, no fewer.
func wantResumeOffenders(t *testing.T, c resumeCensus, rel string, want ...string) {
	t.Helper()
	full := make([]string, 0, len(want))
	for _, w := range want {
		full = append(full, rel+": "+w)
	}
	sort.Strings(full)
	if strings.Join(c.offenders, "\n") != strings.Join(full, "\n") {
		t.Fatalf("census offenders = %v, want exactly %v", c.offenders, full)
	}
}

// PIN: the census sees every way to name a builder, and admits only the two
// exact directories. Each case plants ONE file in a synthetic module
// (t.TempDir, never the real tree) and is judged by the PRODUCTION census
// function; each has its expected single offender, and the positive controls
// have none.
func TestResumeBuilderCensusSeesEveryForm(t *testing.T) {
	census := func(t *testing.T, rel, body string) resumeCensus {
		t.Helper()
		files := map[string]string{}
		if rel != "" {
			files[rel] = body
		}
		return resumeCensusOf(t, files)
	}
	imp := resumeCensusImp
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

	// U2 verifier defect 1: frames with NO raw `"verb":"resume"` in them that
	// the PRODUCTION codec still decodes to exactly NewResume(job) — JSON
	// escapes are what DecodeRequest reads, not what a regex sees. The codec
	// proof first, so no case here is a strawman.
	wantFrame, err := EncodeRequest(NewResume("job-0001abcd"))
	if err != nil {
		t.Fatal(err)
	}
	jsonU := resumeJSONU
	for _, f := range []struct{ name, frame string }{
		{"JSON-escaped verb value", `{"v":1,"verb":"` + jsonU("0072") + `esume","payload":{"job_id":"job-0001abcd"}}`},
		{"JSON-escaped verb key", `{"v":1,"v` + jsonU("0065") + `rb":"resume","payload":{"job_id":"job-0001abcd"}}`},
	} {
		if strings.Contains(f.frame, `"verb":"resume"`) {
			t.Fatalf("%s is not escaped, the case proves nothing: %s", f.name, f.frame)
		}
		r, err := DecodeRequest([]byte(f.frame))
		got, eerr := EncodeRequest(r)
		if err != nil || eerr != nil || string(got) != string(wantFrame) {
			t.Fatalf("%s: the codec must read %s as a resume (decode %v, encode %v, frame %q)", f.name, f.frame, err, eerr, got)
		}
		cases = append(cases, struct{ name, rel, body, want string }{
			f.name, "api/resume.go", "package api\n\nconst f = `" + f.frame + "`\n", "spells a resume frame",
		})
	}
	inner, err := json.Marshal(`{"v` + jsonU("0065") + `rb":"resume"}`)
	if err != nil {
		t.Fatal(err)
	}
	cases = append(cases, struct{ name, rel, body, want string }{
		"escaped frame inside a JSON string", "api/resume.go", "package api\n\nconst f = `{\"frames\":[" + string(inner) + "]}`\n", "spells a resume frame",
	})
	for _, ok := range []struct{ name, rel, body string }{
		{"JSON naming another verb", "agent/x.go", "package agent\n\nconst f = `{\"v\":1,\"verb\":\"status\",\"payload\":{}}`\n"},
		{"a receipt step called resume", "internal/updaterworker/receipt.go", "package updaterworker\n\nconst f = `{\"step\":\"resume\",\"resumed_at\":\"2026-09-24T10:00:00Z\"}`\n"},
	} {
		t.Run("admitted "+ok.name, func(t *testing.T) {
			if c := census(t, ok.rel, ok.body); len(c.offenders) != 0 {
				t.Fatalf("%s must be admitted, census offenders = %v", ok.rel, c.offenders)
			}
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

// PIN (U2 verifier defect 2, probe.out H2/H3/H4): a resume Request is valid
// only with a non-nil Resume, so WRITING that field builds one without naming
// any builder. The census flags every write form outside the admitted set —
// a composite key Resume, an unkeyed wire Request literal, &X.Resume and an
// assignment to X.Resume (or below it) — and every pointer to a wire Request,
// because a decoder (json.Unmarshal, Decoder.Decode, any helper) needs one to
// fill Resume; and the frame rules are case-insensitive, as json.Unmarshal
// into a Request is. READS stay admitted: a worker-side handler dispatches on
// r.Resume != nil and reads r.Resume.JobID.
func TestResumeCensusSeesEveryResumeFieldWrite(t *testing.T) {
	const (
		write = "writes a Resume field"
		ptr   = "takes a pointer to an updaterwire.Request"
		frame = "spells a resume frame"
	)
	file := func(pkg, body string) string {
		return "package " + pkg + "\n\nimport (\n\t\"bytes\"\n\t\"encoding/json\"\n\n\t\"nofx/internal/updaterwire\"\n)\n\n" +
			"var _, _ = bytes.NewReader, json.Unmarshal\n\n" + body + "\n"
	}
	for _, tc := range []struct {
		name, rel, body string
		want            []string
	}{
		// the verifier's probes
		{"H2 json into &r.Resume", "api/h2.go", file("api", "func f() ([]byte, error) {\n\tr := updaterwire.Request{Verb: \"resume\"}\n\t_ = json.Unmarshal([]byte(`{\"job_id\":\"job-0001abcd\"}`), &r.Resume)\n\treturn updaterwire.EncodeRequest(r)\n}"), []string{write}},
		{"H3 Verbs()[3] and &r.Resume", "api/h3.go", file("api", "func f() ([]byte, error) {\n\tr := updaterwire.Request{Verb: updaterwire.Verbs()[3]}\n\t_ = json.Unmarshal([]byte(`{\"job_id\":\"job-0001abcd\"}`), &r.Resume)\n\treturn updaterwire.EncodeRequest(r)\n}"), []string{write}},
		{"H4 Go-struct JSON into a whole Request", "api/h4.go", file("api", "func f() ([]byte, error) {\n\tvar r updaterwire.Request\n\t_ = json.Unmarshal([]byte(`{\"Verb\":\"resume\",\"Resume\":{\"job_id\":\"job-0001abcd\"}}`), &r)\n\treturn updaterwire.EncodeRequest(r)\n}"), []string{frame, ptr}},
		// every other write form
		{"composite key Resume", "api/w.go", file("api", "func f(r0 updaterwire.Request) updaterwire.Request {\n\treturn updaterwire.Request{Verb: r0.Verb, Resume: r0.Resume}\n}"), []string{write}},
		{"assignment to X.Resume", "api/w.go", file("api", "func f(r0 updaterwire.Request) (r updaterwire.Request) {\n\tr.Verb = r0.Verb\n\tr.Resume = r0.Resume\n\treturn r\n}"), []string{write}},
		{"assignment below X.Resume", "api/w.go", file("api", "func f(r updaterwire.Request) updaterwire.Request {\n\tr.Resume.JobID = \"job-0002abcd\"\n\treturn r\n}"), []string{write}},
		{"address below X.Resume", "api/w.go", file("api", "func f(b []byte, r updaterwire.Request) error {\n\treturn json.Unmarshal(b, &r.Resume.JobID)\n}"), []string{write}},
		{"unkeyed Request literal", "api/w.go", file("api", "func f(r0 updaterwire.Request) updaterwire.Request {\n\treturn updaterwire.Request{r0.Verb, nil, nil, nil, r0.Resume}\n}"), []string{write}},
		{"an unrelated Resume field (fail closed)", "agent/x.go", "package agent\n\ntype state struct{ Resume bool }\n\nfunc f(s *state) { s.Resume = true }\n", []string{write}},
		// pointers to a wire Request
		{"a decoder into a runtime frame", "api/p.go", file("api", "func f(b []byte) (updaterwire.Request, error) {\n\tvar r updaterwire.Request\n\terr := json.NewDecoder(bytes.NewReader(b)).Decode(&r)\n\treturn r, err\n}"), []string{ptr}},
		{"a Request parameter's address", "api/p.go", file("api", "func f(b []byte, r updaterwire.Request) (updaterwire.Request, error) {\n\treturn r, json.Unmarshal(b, &r)\n}"), []string{ptr}},
		{"a wire constructor's result, addressed", "api/p.go", file("api", "func f(b []byte) (updaterwire.Request, error) {\n\tr := updaterwire.NewStatus(\"\")\n\treturn r, json.Unmarshal(b, &r)\n}"), []string{ptr}},
		{"a pointer through any()", "api/p.go", file("api", "func f(b []byte) (updaterwire.Request, error) {\n\tvar r updaterwire.Request\n\treturn r, json.Unmarshal(b, any(&r))\n}"), []string{ptr}},
		{"new(Request)", "api/p.go", file("api", "func f(b []byte) (updaterwire.Request, error) {\n\tp := new(updaterwire.Request)\n\treturn *p, json.Unmarshal(b, p)\n}"), []string{ptr}},
		{"&Request{}", "api/p.go", file("api", "func f(b []byte) error {\n\treturn json.Unmarshal(b, &updaterwire.Request{})\n}"), []string{ptr}},
		{"a *Request parameter", "api/p.go", file("api", "func f(b []byte, p *updaterwire.Request) error {\n\treturn json.Unmarshal(b, p)\n}"), []string{ptr}},
		{"renamed import", "api/p.go", resumeCensusImp("api", "uw", "func f(r uw.Request) *uw.Request { return &r }"), []string{ptr}},
		{"dot import", "api/p.go", resumeCensusImp("api", ".", "func f(r Request) *Request { return &r }"), []string{ptr}},
		// case-insensitive frame rules
		{"case-insensitive fragment", "api/f.go", "package api\n\nconst f = `, \"Verb\" : \"RESUME`\n", []string{frame}},
		{"an escaped, re-cased verb key", "api/f.go", "package api\n\nconst f = `{\"V" + resumeJSONU("0065") + "rb\":\"resume\"}`\n", []string{frame}},
		{"a resume payload key in JSON", "api/f.go", "package api\n\nconst f = `{\"Resume\":{\"job_id\":\"job-0001abcd\"}}`\n", []string{frame}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wantResumeOffenders(t, resumeCensusOf(t, map[string]string{tc.rel: tc.body}), tc.rel, tc.want...)
		})
	}

	// READS stay admitted (the U4 socket dispatches on r.Resume != nil, never
	// on case VerbResume); so do the app's by-value send, a relay that only
	// decodes and re-encodes (outside any name census by design: the socket's
	// peer uid and the worker's job file are the gates), decoding something
	// else beside a Request, and the CLI, which may build freely.
	for _, ok := range []struct{ name, rel, body string }{
		{"reads of the payload", "internal/updaterworker/socket.go", file("updaterworker", "func jobOf(r updaterwire.Request) (string, bool) {\n\tif r.Resume != nil {\n\t\treturn r.Resume.JobID, true\n\t}\n\treturn \"\", false\n}")},
		{"the app sends by value", "api/updates.go", file("api", "func send(c *updaterwire.Client) (updaterwire.Response, error) {\n\treturn c.Do(updaterwire.NewInstall(\"rel\", \"job-0001abcd\"))\n}")},
		{"a relay", "api/relay.go", file("api", "func relay(b []byte) ([]byte, error) {\n\tr, err := updaterwire.DecodeRequest(b)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn updaterwire.EncodeRequest(r)\n}")},
		{"decoding something else", "api/other.go", file("api", "func f(b []byte, r updaterwire.Request) (int, error) {\n\tvar n int\n\terr := json.Unmarshal(b, &n)\n\t_ = r.Verb\n\treturn n, err\n}")},
		{"the CLI", "cmd/nofx-updater/main.go", file("main", "func main() {\n\tvar r updaterwire.Request\n\t_ = json.Unmarshal(nil, &r.Resume)\n\t_ = json.Unmarshal(nil, &r)\n\t_ = updaterwire.Request{Resume: r.Resume}\n}")},
	} {
		t.Run("admitted "+ok.name, func(t *testing.T) {
			wantResumeOffenders(t, resumeCensusOf(t, map[string]string{ok.rel: ok.body}), ok.rel)
		})
	}
}
