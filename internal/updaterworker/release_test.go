package updaterworker

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ── release materialization, at the production call sites ────────────────────
//
// Every archive here is made the way release.yml makes it: 3a's own
// deploy/release/package.sh stages the allow-list, deploy/release/manifest.sh
// writes the manifest, the REAL `ssh-keygen -Y sign -n release` signs it, and
// GNU `tar -C stage -czf` packs it. The only deviation from release.yml is the
// one routed to 3a/103 (brief C8): the manifest is written OUTSIDE the stage
// and moved in, so it does not list itself — and one test runs release.yml's
// verbatim redirect to prove that today's output is REFUSED for exactly that.
// Hostile archives (entries the scripts would never produce) are written with
// archive/tar over a legitimately staged tree, with a control that the same
// tree WITHOUT the hostile entry is accepted.

const testReleaseID = "v0.0.1-u3"

var (
	testSHA     = strings.Repeat("c0ffee", 6) + "abcd" // 40 lowercase hex
	testNow     = time.Date(2026, 9, 24, 14, 30, 0, 123456789, time.UTC)
	testBuildID = "2026-09-24-u3"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "deploy", "release", "package.sh")); err != nil {
		t.Fatalf("repo root %s has no deploy/release/package.sh: %v", root, err)
	}
	return root
}

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// releaseSource is a repo tree with exactly what package.sh requires, plus a
// STALE deploy/RELEASE that package.sh must not ship.
func releaseSource(t *testing.T, withIndex bool) string {
	t.Helper()
	src := t.TempDir()
	files := map[string]string{
		"nofx-bin":                             "\x7fELF u3 stand-in binary\n",
		"LICENSE":                              "test licence\n",
		"ninjascript/vltrader_tcp_PROTOCOL.md": "protocol_version: 3\n",
		"ninjascript/VLTraderTcp.cs":           "public const string VL_BUILD_ID = \"" + testBuildID + "\";\n",
		"web/dist/assets/app.js":               "console.log('u3')\n",
		"deploy/RELEASE":                       strings.Repeat("a", 40) + "\n",
	}
	if withIndex {
		files["web/dist/index.html"] = "<!doctype html><title>u3</title>\n"
	}
	writeFiles(t, src, files)
	if err := os.Chmod(filepath.Join(src, "nofx-bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	return src
}

func runIn(t *testing.T, dir string, stdoutOnly bool, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = keygenEnv(t)
	var out []byte
	var err error
	if stdoutOnly {
		out, err = cmd.Output()
	} else {
		out, err = cmd.CombinedOutput()
	}
	if err != nil {
		var stderr []byte
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = ee.Stderr
		}
		t.Fatalf("%s %v: %v\n%s%s", name, args, err, out, stderr)
	}
	return out
}

type testRelease struct {
	archive  string
	stage    string
	signer   testSigner
	signers  string // allowed-signers: the signer under principal release
	fp       string
	manifest []byte // the signed bytes
	sig      []byte
}

type releaseOpts struct {
	selfEntry      bool // release.yml:172 verbatim — manifest.sh redirected INTO the stage
	noIndex        bool
	beforeManifest func(t *testing.T, stage string) // add to the stage before manifest.sh lists it
	editManifest   func(b []byte) []byte            // rewrite the manifest BEFORE it is signed
	afterSign      func(t *testing.T, stage string) // tamper after signing, before tar
}

func buildRelease(t *testing.T, o releaseOpts) testRelease {
	t.Helper()
	sshKeygen(t)
	root := repoRoot(t)
	work := t.TempDir()
	stage := filepath.Join(work, "stage")
	runIn(t, root, false, "bash", "deploy/release/package.sh", releaseSource(t, !o.noIndex), stage, testSHA)
	manifestPath := filepath.Join(stage, "manifest.json")
	if o.beforeManifest != nil {
		o.beforeManifest(t, stage)
	}
	if o.selfEntry {
		runIn(t, root, false, "bash", "-c", `bash deploy/release/manifest.sh "$1" "$2" "$3" > "$1/manifest.json"`, "_", stage, testSHA, testReleaseID)
	} else {
		out := runIn(t, root, true, "bash", "deploy/release/manifest.sh", stage, testSHA, testReleaseID)
		if o.editManifest != nil {
			out = o.editManifest(out)
		}
		if err := os.WriteFile(manifestPath, out, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	signer := newTestSigner(t, work, "release-signer")
	sig := signer.sign(t, manifestPath, "release")
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if o.afterSign != nil {
		o.afterSign(t, stage)
	}
	archive := filepath.Join(work, testReleaseID+".tar.gz")
	runIn(t, work, false, "tar", "-C", stage, "-czf", archive, ".")
	return testRelease{
		archive: archive, stage: stage, signer: signer,
		signers: writeAllowedSigners(t, work, "release "+signer.pub),
		fp:      signer.fingerprint(t), manifest: manifest, sig: sig,
	}
}

type fetchEnv struct{ releaseRoot, dataDir string }

func newFetchEnv(t *testing.T) fetchEnv {
	t.Helper()
	base := t.TempDir()
	e := fetchEnv{releaseRoot: filepath.Join(base, "releases"), dataDir: filepath.Join(base, "data")}
	for _, d := range []string{e.releaseRoot, e.dataDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return e
}

func (e fetchEnv) cfg(r testRelease) FetchConfig {
	return FetchConfig{
		Archive: r.archive, ReleaseID: testReleaseID, ReleaseRoot: e.releaseRoot,
		AllowedSigners: r.signers, DataDir: e.dataDir, Now: func() time.Time { return testNow },
	}
}

func (e fetchEnv) verdictPath() string {
	return filepath.Join(e.dataDir, "updater", "verdicts", testReleaseID+".json")
}

func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range ents {
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out
}

// assertNothingLanded: a refused fetch leaves no verdict (not even the
// verdicts dir), no release dir and no staging debris.
func (e fetchEnv) assertNothingLanded(t *testing.T) {
	t.Helper()
	if _, err := os.Lstat(e.verdictPath()); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a refused fetch left a verdict at %s (lstat err %v)", e.verdictPath(), err)
	}
	if got := dirNames(t, e.releaseRoot); len(got) != 0 {
		t.Fatalf("a refused fetch left %v in the release root", got)
	}
}

// releaseDirOf returns the verdict's release dir, refusing anything but an
// absolute dir inside this env's release root — a broken FetchRelease (an
// empty ReleaseDir) must never make a tamper helper write relative to the
// test's cwd, which is the package's SOURCE directory.
func (e fetchEnv) releaseDirOf(t *testing.T, v ReleaseVerdict) string {
	t.Helper()
	if !filepath.IsAbs(v.ReleaseDir) || filepath.Dir(v.ReleaseDir) != e.releaseRoot {
		t.Fatalf("verdict release_dir %q is not a dir inside the test's release root %s — refusing to tamper", v.ReleaseDir, e.releaseRoot)
	}
	return v.ReleaseDir
}

// ── the success path: the layout activation.Resolve reads ────────────────────

func TestFetchMaterializesTheActivationLayout(t *testing.T) {
	r := buildRelease(t, releaseOpts{})
	e := newFetchEnv(t)
	v, err := FetchRelease(e.cfg(r))
	if err != nil {
		t.Fatalf("FetchRelease refused a release made by 3a's own scripts: %v", err)
	}
	final := filepath.Join(e.releaseRoot, testSHA)
	if got := dirNames(t, e.releaseRoot); len(got) != 1 || got[0] != testSHA {
		t.Fatalf("release root = %v, want exactly [%s] (no staging debris)", got, testSHA)
	}
	// activation.Resolve's layout: <dir>/{nofx-bin, web/dist, RELEASE, manifest.json}
	bin, err := os.Lstat(filepath.Join(final, "nofx-bin"))
	if err != nil || !bin.Mode().IsRegular() || bin.Mode().Perm()&0o100 == 0 {
		t.Fatalf("nofx-bin: %v %v (want a regular executable file)", bin, err)
	}
	if b, err := os.ReadFile(filepath.Join(final, "web", "dist", "index.html")); err != nil || !bytes.Contains(b, []byte("u3")) {
		t.Fatalf("web/dist/index.html: %q %v", b, err)
	}
	rel, _ := os.ReadFile(filepath.Join(final, "RELEASE"))
	dep, _ := os.ReadFile(filepath.Join(final, "deploy", "RELEASE"))
	if string(rel) != testSHA+"\n" || !bytes.Equal(rel, dep) {
		t.Fatalf("RELEASE = %q, deploy/RELEASE = %q; want both %q (copied, byte-identical)", rel, dep, testSHA+"\n")
	}
	raw, err := os.ReadFile(filepath.Join(final, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatal(err)
	}
	var names []string
	for k := range keys {
		names = append(names, k)
	}
	sort.Strings(names)
	if strings.Join(names, ",") != "binary_md5,signature_verdict,source_sha" {
		t.Fatalf("activation manifest keys = %v, want exactly activation.Manifest's {source_sha, binary_md5, signature_verdict}", names)
	}
	var am struct {
		SourceSHA string `json:"source_sha"`
		BinaryMD5 string `json:"binary_md5"`
		Signature string `json:"signature_verdict"`
	}
	_ = json.Unmarshal(raw, &am)
	binBytes, _ := os.ReadFile(filepath.Join(final, "nofx-bin"))
	md := md5.Sum(binBytes)
	if am.SourceSHA != testSHA || am.BinaryMD5 != hex.EncodeToString(md[:]) || am.Signature != "sshsig:release:"+r.fp {
		t.Fatalf("activation manifest = %+v; want source_sha %s, binary_md5 %x, signature_verdict sshsig:release:%s", am, testSHA, md, r.fp)
	}
	// the signed pair, byte-identical, and the real tool still verifies it there
	sm, _ := os.ReadFile(filepath.Join(final, "signed", "manifest.json"))
	ss, _ := os.ReadFile(filepath.Join(final, "signed", "manifest.json.sig"))
	if !bytes.Equal(sm, r.manifest) || !bytes.Equal(ss, r.sig) {
		t.Fatalf("the signed pair was not kept byte-identical under signed/")
	}
	if err := keygenVerify(t, r.signers, ss, sm); err != nil {
		t.Fatalf("ssh-keygen refuses the materialized signed pair: %v", err)
	}
	// the verdict: exact key set, private file in private dirs, values READ
	vp := e.verdictPath()
	for p, want := range map[string]fs.FileMode{vp: 0o600, filepath.Dir(vp): 0o700, filepath.Dir(filepath.Dir(vp)): 0o700} {
		fi, err := os.Lstat(p)
		if err != nil || fi.Mode().Perm() != want {
			t.Fatalf("%s: mode %v err %v, want %#o", p, fi.Mode(), err, want)
		}
	}
	vraw, _ := os.ReadFile(vp)
	var vkeys map[string]json.RawMessage
	if err := json.Unmarshal(vraw, &vkeys); err != nil {
		t.Fatal(err)
	}
	names = names[:0]
	for k := range vkeys {
		names = append(names, k)
	}
	sort.Strings(names)
	if got, want := strings.Join(names, ","), "artifacts,hashalg,manifest_sha256,release_dir,release_id,schema,signer,signer_fingerprint,source_sha,verified_at"; got != want {
		t.Fatalf("verdict keys = %s\nwant exactly %s (brief §3.2)", got, want)
	}
	var art struct{ Artifacts []json.RawMessage }
	_ = json.Unmarshal(r.manifest, &art)
	msum := sha256.Sum256(r.manifest)
	want := ReleaseVerdict{
		Schema: 1, ReleaseID: testReleaseID, SourceSHA: testSHA, ReleaseDir: final,
		Signer: "release", SignerFingerprint: r.fp, HashAlg: "sha512",
		ManifestSHA256: hex.EncodeToString(msum[:]), Artifacts: len(art.Artifacts),
		VerifiedAt: testNow.Format(time.RFC3339Nano),
	}
	if v != want {
		t.Fatalf("returned verdict = %+v\nwant %+v", v, want)
	}
	if rv, err := ReadReleaseVerdict(e.dataDir, testReleaseID); err != nil || rv != want {
		t.Fatalf("ReadReleaseVerdict = %+v, %v; want %+v", rv, err, want)
	}
	// the job's re-proofs (states downloaded / verified) accept what fetch made
	if n, err := RehashRelease(final, v.ManifestSHA256); err != nil || n != v.Artifacts {
		t.Fatalf("RehashRelease = %d, %v; want %d, nil", n, err, v.Artifacts)
	}
	m, sv, err := ReverifyRelease(final, r.signers, testReleaseID)
	if err != nil || m.SourceSHA != testSHA || m.ReleaseID != testReleaseID || m.AddonBuildID != testBuildID || sv.Fingerprint != r.fp || m.SHA256 != v.ManifestSHA256 {
		t.Fatalf("ReverifyRelease = %+v, %+v, %v", m, sv, err)
	}
	// NOT proven here (waits for 103's merge — internal/activation is not on
	// dev): activation.Resolve(final) + activation.Stage on a real stamped
	// binary. U4's TestMaterializedReleaseResolvesAndStages owns that.
}

// ── extraction ────────────────────────────────────────────────────────────────

type tarEntry struct {
	hdr  tar.Header
	body string
}

// repack writes the staged tree (./-prefixed, like `tar -C stage .`) plus the
// extra entries with archive/tar.
func repack(t *testing.T, stage string, extra ...tarEntry) string {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	put := func(e tarEntry) {
		t.Helper()
		h := e.hdr
		if h.Typeflag == tar.TypeReg {
			h.Size = int64(len(e.body))
		}
		if h.Mode == 0 {
			h.Mode = 0o644
		}
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatal(err)
		}
		if h.Typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	put(tarEntry{hdr: tar.Header{Typeflag: tar.TypeDir, Name: "./", Mode: 0o755}})
	err := filepath.WalkDir(stage, func(p string, d fs.DirEntry, err error) error {
		if err != nil || p == stage {
			return err
		}
		rel, _ := filepath.Rel(stage, p)
		if d.IsDir() {
			put(tarEntry{hdr: tar.Header{Typeflag: tar.TypeDir, Name: "./" + filepath.ToSlash(rel) + "/", Mode: 0o755}})
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fi, _ := d.Info()
		put(tarEntry{hdr: tar.Header{Typeflag: tar.TypeReg, Name: "./" + filepath.ToSlash(rel), Mode: int64(fi.Mode().Perm())}, body: string(b)})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range extra {
		put(e)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), testReleaseID+".tar.gz")
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func reg(name, body string) tarEntry {
	return tarEntry{hdr: tar.Header{Typeflag: tar.TypeReg, Name: name, Mode: 0o644}, body: body}
}

func TestExtractRefusesTraversalSymlinkDeviceAbsolute(t *testing.T) {
	r := buildRelease(t, releaseOpts{})
	// control: the same tree, repacked by archive/tar with nothing hostile, is ACCEPTED
	{
		e := newFetchEnv(t)
		c := e.cfg(r)
		c.Archive = repack(t, r.stage)
		if _, err := FetchRelease(c); err != nil {
			t.Fatalf("control: the repacked clean tree is refused: %v", err)
		}
	}
	outside := t.TempDir() // where an absolute entry would land
	for name, bad := range map[string]tarEntry{
		"traversal ../":           reg("../escape-dotdot", "x"),
		"traversal inside a path": reg("./web/../../escape-mid", "x"),
		"traversal deep":          reg("./web/dist/../../../../escape-deep", "x"),
		"absolute path":           reg(filepath.Join(outside, "abs-escape"), "x"),
		"home-relative path":      reg("~/escape-home", "x"),
		"backslash path":          reg(`web\..\..\escape-bs`, "x"),
		"symlink out of the tree": {hdr: tar.Header{Typeflag: tar.TypeSymlink, Name: "./evil-link", Linkname: "/etc/passwd"}},
		"symlink inside the tree": {hdr: tar.Header{Typeflag: tar.TypeSymlink, Name: "./web/dist/alias.html", Linkname: "index.html"}},
		"hard link":               {hdr: tar.Header{Typeflag: tar.TypeLink, Name: "./hard", Linkname: "nofx-bin"}},
		"char device":             {hdr: tar.Header{Typeflag: tar.TypeChar, Name: "./dev-null", Devmajor: 1, Devminor: 3}},
		"block device":            {hdr: tar.Header{Typeflag: tar.TypeBlock, Name: "./dev-sda", Devmajor: 8}},
		"fifo":                    {hdr: tar.Header{Typeflag: tar.TypeFifo, Name: "./pipe"}},
		"setuid file":             {hdr: tar.Header{Typeflag: tar.TypeReg, Name: "./suid-bin", Mode: 0o4755}, body: "x"},
		"duplicate entry":         reg("./LICENSE", "a second LICENSE"),
	} {
		t.Run(name, func(t *testing.T) {
			e := newFetchEnv(t)
			c := e.cfg(r)
			c.Archive = repack(t, r.stage, bad)
			v, err := FetchRelease(c)
			if err == nil {
				t.Fatalf("ACCEPTED an archive carrying %s (%q); verdict %+v", name, bad.hdr.Name, v)
			}
			if !errors.Is(err, ErrUnsafeEntry) {
				t.Fatalf("%s: err = %v, want ErrUnsafeEntry", name, err)
			}
			for _, p := range []string{
				filepath.Join(e.releaseRoot, "escape-dotdot"), filepath.Join(e.releaseRoot, "escape-mid"),
				filepath.Join(filepath.Dir(e.releaseRoot), "escape-deep"), filepath.Join(e.releaseRoot, "escape-deep"),
				filepath.Join(outside, "abs-escape"),
			} {
				if _, err := os.Lstat(p); !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("an entry escaped the extraction root: %s exists", p)
				}
			}
			e.assertNothingLanded(t)
		})
	}
}

// ── rehash: every artifacts[] entry, and nothing else ─────────────────────────

func TestRehashRefusesExtraMissingOrChangedArtifact(t *testing.T) {
	flip := func(rel string) func(*testing.T, string) {
		return func(t *testing.T, stage string) {
			p := filepath.Join(stage, filepath.FromSlash(rel))
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			b[len(b)/2] ^= 0x20
			if err := os.WriteFile(p, b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	// (a) at fetch: the archive disagrees with its own signed manifest
	for name, c := range map[string]struct {
		o     releaseOpts
		want  error
		names string
	}{
		"extra file":               {releaseOpts{afterSign: func(t *testing.T, s string) { writeFiles(t, s, map[string]string{"web/dist/extra.js": "x"}) }}, ErrArtifactMismatch, "extra (not in artifacts[]): web/dist/extra.js"},
		"extra empty file":         {releaseOpts{afterSign: func(t *testing.T, s string) { writeFiles(t, s, map[string]string{"notes.txt": ""}) }}, ErrArtifactMismatch, "extra (not in artifacts[]): notes.txt"},
		"missing file":             {releaseOpts{afterSign: func(t *testing.T, s string) { _ = os.Remove(filepath.Join(s, "LICENSE")) }}, ErrArtifactMismatch, "missing: LICENSE"},
		"changed bytes, same size": {releaseOpts{afterSign: flip("web/dist/index.html")}, ErrArtifactMismatch, "changed: web/dist/index.html"},
		"changed size": {releaseOpts{afterSign: func(t *testing.T, s string) {
			f, _ := os.OpenFile(filepath.Join(s, "nofx-bin"), os.O_APPEND|os.O_WRONLY, 0)
			_, _ = f.WriteString("more")
			_ = f.Close()
		}}, ErrArtifactMismatch, "changed: nofx-bin"},
		// C8: release.yml:172 redirects manifest.sh INTO the stage, so the shell
		// has created a 0-byte manifest.json before find runs and artifacts[]
		// lists it — an entry the signed file can never match.
		"the 0-byte manifest self-entry (release.yml verbatim)": {releaseOpts{selfEntry: true}, ErrManifest, "lists itself"},
	} {
		t.Run("fetch/"+name, func(t *testing.T) {
			r := buildRelease(t, c.o)
			if name == "the 0-byte manifest self-entry (release.yml verbatim)" {
				if !bytes.Contains(r.manifest, []byte(`{"path":"manifest.json","sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","bytes":0}`)) {
					t.Fatalf("fixture: release.yml's redirect did not produce the 0-byte self-entry:\n%s", r.manifest)
				}
			}
			e := newFetchEnv(t)
			v, err := FetchRelease(e.cfg(r))
			if err == nil {
				t.Fatalf("ACCEPTED (%s); verdict %+v", name, v)
			}
			if !errors.Is(err, c.want) || !strings.Contains(err.Error(), c.names) {
				t.Fatalf("%s: err = %v; want %v naming %q", name, err, c.want, c.names)
			}
			e.assertNothingLanded(t)
		})
	}
	// (b) at the job's downloaded state: RehashRelease over the materialized dir
	r := buildRelease(t, releaseOpts{})
	for name, c := range map[string]struct {
		tamper func(t *testing.T, dir string)
		want   error
		names  string
	}{
		"extra file":      {func(t *testing.T, d string) { writeFiles(t, d, map[string]string{"web/dist/extra.js": "x"}) }, ErrArtifactMismatch, "extra (not in artifacts[]): web/dist/extra.js"},
		"missing file":    {func(t *testing.T, d string) { _ = os.Remove(filepath.Join(d, "ninjascript", "VLTraderTcp.cs")) }, ErrArtifactMismatch, "missing: ninjascript/VLTraderTcp.cs"},
		"changed binary":  {func(t *testing.T, d string) { flip("nofx-bin")(t, d) }, ErrArtifactMismatch, "changed: nofx-bin"},
		"planted symlink": {func(t *testing.T, d string) { _ = os.Symlink("/etc/passwd", filepath.Join(d, "web", "dist", "x.js")) }, ErrArtifactMismatch, "not a regular file: web/dist/x.js"},
		"RELEASE rewritten": {func(t *testing.T, d string) {
			_ = os.WriteFile(filepath.Join(d, "RELEASE"), []byte(strings.Repeat("b", 40)+"\n"), 0o644)
		}, ErrLayout, "RELEASE"},
		"activation manifest md5 rewritten": {func(t *testing.T, d string) {
			p := filepath.Join(d, "manifest.json")
			b, _ := os.ReadFile(p)
			var m map[string]string
			_ = json.Unmarshal(b, &m)
			m["binary_md5"] = strings.Repeat("0", 32)
			nb, _ := json.Marshal(m)
			_ = os.WriteFile(p, nb, 0o644)
		}, ErrLayout, "binary_md5"},
		"signed manifest swapped": {func(t *testing.T, d string) {
			p := filepath.Join(d, "signed", "manifest.json")
			b, _ := os.ReadFile(p)
			_ = os.WriteFile(p, append(b, ' '), 0o644)
		}, ErrManifest, "manifest_sha256"},
	} {
		t.Run("job/"+name, func(t *testing.T) {
			e := newFetchEnv(t)
			v, err := FetchRelease(e.cfg(r))
			if err != nil {
				t.Fatal(err)
			}
			if n, err := RehashRelease(v.ReleaseDir, v.ManifestSHA256); err != nil || n != v.Artifacts {
				t.Fatalf("control: RehashRelease on the fresh release = %d, %v", n, err)
			}
			c.tamper(t, e.releaseDirOf(t, v))
			n, err := RehashRelease(v.ReleaseDir, v.ManifestSHA256)
			if err == nil {
				t.Fatalf("RehashRelease ACCEPTED a release with %s (%d artifacts)", name, n)
			}
			if !errors.Is(err, c.want) || !strings.Contains(err.Error(), c.names) {
				t.Fatalf("%s: err = %v; want %v naming %q", name, err, c.want, c.names)
			}
		})
	}
	t.Run("job/no expected manifest sha", func(t *testing.T) {
		e := newFetchEnv(t)
		v, err := FetchRelease(e.cfg(r))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := RehashRelease(v.ReleaseDir, ""); !errors.Is(err, ErrManifest) {
			t.Fatalf("RehashRelease with no expected manifest sha: err = %v, want ErrManifest", err)
		}
	})
}

// ── the verdict: written once, and only after every check ────────────────────

func TestVerdictWrittenOnlyAfterEveryCheck(t *testing.T) {
	good := buildRelease(t, releaseOpts{})
	foreign := newTestSigner(t, t.TempDir(), "foreign")
	for name, c := range map[string]struct {
		setup func(t *testing.T, e fetchEnv, cfg *FetchConfig)
		o     *releaseOpts
		want  error
	}{
		"foreign signer": {func(t *testing.T, e fetchEnv, c *FetchConfig) {
			c.AllowedSigners = writeAllowedSigners(t, t.TempDir(), "release "+foreign.pub)
		}, nil, ErrSigForeignKey},
		"no allowed-signers file": {func(t *testing.T, e fetchEnv, c *FetchConfig) {
			c.AllowedSigners = filepath.Join(t.TempDir(), "deploy", "release_allowed_signers")
		}, nil, ErrNoAllowedSigners},
		"release id disagrees with the signed manifest": {func(t *testing.T, e fetchEnv, c *FetchConfig) { c.ReleaseID = "v9.9.9" }, nil, ErrManifest},
		"signed source_sha disagrees with deploy/RELEASE": {nil, &releaseOpts{editManifest: func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"source_sha": "`+testSHA+`"`), []byte(`"source_sha": "`+strings.Repeat("d", 40)+`"`), 1)
		}}, ErrLayout},
		"unsafe entry": {func(t *testing.T, e fetchEnv, c *FetchConfig) {
			c.Archive = repack(t, good.stage, tarEntry{hdr: tar.Header{Typeflag: tar.TypeSymlink, Name: "./l", Linkname: "/"}})
		}, nil, ErrUnsafeEntry},
		"extra artifact":         {nil, &releaseOpts{afterSign: func(t *testing.T, s string) { writeFiles(t, s, map[string]string{"x": "x"}) }}, ErrArtifactMismatch},
		"no web/dist/index.html": {nil, &releaseOpts{noIndex: true}, ErrLayout},
		"archive absent": {func(t *testing.T, e fetchEnv, c *FetchConfig) {
			c.Archive = filepath.Join(t.TempDir(), "absent.tar.gz")
		}, nil, ErrFetchConfig},
		"not a gzip archive": {func(t *testing.T, e fetchEnv, c *FetchConfig) {
			c.Archive = filepath.Join(t.TempDir(), "plain.tar.gz")
			_ = os.WriteFile(c.Archive, []byte("not gzip"), 0o644)
		}, nil, ErrArchive},
		"release root writable by others": {func(t *testing.T, e fetchEnv, c *FetchConfig) { _ = os.Chmod(e.releaseRoot, 0o777) }, nil, ErrFetchConfig},
		"relative release root":           {func(t *testing.T, e fetchEnv, c *FetchConfig) { c.ReleaseRoot = "releases" }, nil, ErrFetchConfig},
		"forged release id":               {func(t *testing.T, e fetchEnv, c *FetchConfig) { c.ReleaseID = "../v0.0.1-u3" }, nil, ErrFetchConfig},
	} {
		t.Run(name, func(t *testing.T) {
			r := good
			if c.o != nil {
				r = buildRelease(t, *c.o)
			}
			e := newFetchEnv(t)
			cfg := e.cfg(r)
			if c.setup != nil {
				c.setup(t, e, &cfg)
			}
			v, err := FetchRelease(cfg)
			if err == nil {
				t.Fatalf("ACCEPTED (%s); verdict %+v", name, v)
			}
			if !errors.Is(err, c.want) {
				t.Fatalf("%s: err = %v, want %v", name, err, c.want)
			}
			// the id is refused where it is first known, before anything is materialized
			if name == "release id disagrees with the signed manifest" && !strings.Contains(err.Error(), `not the "v9.9.9" asked for`) {
				t.Fatalf("the release id was not refused at the manifest step: %v", err)
			}
			_ = os.Chmod(e.releaseRoot, 0o755)
			if _, err := os.Lstat(filepath.Join(e.dataDir, "updater", "verdicts")); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("a refused fetch created the verdicts dir (lstat err %v)", err)
			}
			e.assertNothingLanded(t)
		})
	}
	// the release dir for this sha already exists: refused, and left exactly as it was
	t.Run("release dir already exists", func(t *testing.T) {
		e := newFetchEnv(t)
		keep := filepath.Join(e.releaseRoot, testSHA)
		writeFiles(t, keep, map[string]string{"sentinel": "the running release"})
		if _, err := FetchRelease(e.cfg(good)); !errors.Is(err, ErrReleaseDirExists) {
			t.Fatalf("err = %v, want ErrReleaseDirExists", err)
		}
		if got := dirNames(t, e.releaseRoot); len(got) != 1 || got[0] != testSHA {
			t.Fatalf("release root = %v, want only the pre-existing %s", got, testSHA)
		}
		if got := dirNames(t, keep); len(got) != 1 || got[0] != "sentinel" {
			t.Fatalf("the pre-existing release dir was touched: %v", got)
		}
		if _, err := os.Lstat(e.verdictPath()); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("verdict written for a refused fetch")
		}
	})
	// an EMPTY dir under the sha is refused too (rename(2) would silently replace it)
	t.Run("release dir already exists (empty)", func(t *testing.T) {
		e := newFetchEnv(t)
		if err := os.Mkdir(filepath.Join(e.releaseRoot, testSHA), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := FetchRelease(e.cfg(good)); !errors.Is(err, ErrReleaseDirExists) {
			t.Fatalf("err = %v, want ErrReleaseDirExists", err)
		}
		if got := dirNames(t, filepath.Join(e.releaseRoot, testSHA)); len(got) != 0 {
			t.Fatalf("the pre-existing empty release dir was filled: %v", got)
		}
		if _, err := os.Lstat(e.verdictPath()); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("verdict written for a refused fetch")
		}
	})
	// the verdict write itself fails (a loose updater dir): the verdict is
	// absent AND the just-materialized release dir does not survive without one
	t.Run("verdict write refused", func(t *testing.T) {
		e := newFetchEnv(t)
		if err := os.Mkdir(filepath.Join(e.dataDir, "updater"), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := FetchRelease(e.cfg(good)); err == nil || !strings.Contains(err.Error(), "updater dir") {
			t.Fatalf("err = %v, want the private-dir refusal of <data>/updater", err)
		}
		e.assertNothingLanded(t)
	})
	// and the control: every check passes ⇒ exactly one verdict
	t.Run("every check passes", func(t *testing.T) {
		e := newFetchEnv(t)
		v, err := FetchRelease(e.cfg(good))
		if err != nil {
			t.Fatal(err)
		}
		if got := dirNames(t, filepath.Dir(e.verdictPath())); len(got) != 1 || got[0] != testReleaseID+".json" {
			t.Fatalf("verdicts dir = %v, want exactly [%s.json] (no temp debris)", got, testReleaseID)
		}
		if rv, err := ReadReleaseVerdict(e.dataDir, testReleaseID); err != nil || rv != v {
			t.Fatalf("ReadReleaseVerdict = %+v, %v; want %+v", rv, err, v)
		}
	})
}

func TestFetchRefusesAnExistingVerdict(t *testing.T) {
	r := buildRelease(t, releaseOpts{})
	e := newFetchEnv(t)
	if _, err := FetchRelease(e.cfg(r)); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(e.verdictPath())
	if err != nil {
		t.Fatal(err)
	}
	again := func(t *testing.T, c FetchConfig) {
		t.Helper()
		v, err := FetchRelease(c)
		if !errors.Is(err, ErrVerdictExists) {
			t.Fatalf("second fetch: verdict %+v, err = %v; want ErrVerdictExists", v, err)
		}
		after, _ := os.ReadFile(e.verdictPath())
		if !bytes.Equal(before, after) {
			t.Fatalf("the existing verdict was rewritten:\n%s\n→\n%s", before, after)
		}
		if got := dirNames(t, e.releaseRoot); len(got) != 1 || got[0] != testSHA {
			t.Fatalf("release root = %v after a refused re-fetch, want only [%s]", got, testSHA)
		}
	}
	t.Run("same archive again", func(t *testing.T) { again(t, e.cfg(r)) })
	// the check comes FIRST: with no archive at all, the answer is still "verdict exists"
	t.Run("before any other work", func(t *testing.T) {
		c := e.cfg(r)
		c.Archive = filepath.Join(t.TempDir(), "absent.tar.gz")
		c.AllowedSigners = filepath.Join(t.TempDir(), "absent")
		again(t, c)
	})
	// written once and immutable: a verdict nobody can parse, or a dangling
	// link, still blocks — only an operator removing it by hand re-opens the id
	for name, plant := range map[string]func(p string){
		"a garbage verdict": func(p string) { _ = os.WriteFile(p, []byte("{}"), 0o600) },
		"a dangling link":   func(p string) { _ = os.Symlink(filepath.Join(filepath.Dir(p), "nowhere"), p) },
	} {
		t.Run(name, func(t *testing.T) {
			e2 := newFetchEnv(t)
			if err := os.MkdirAll(filepath.Dir(e2.verdictPath()), 0o700); err != nil {
				t.Fatal(err)
			}
			_ = os.Chmod(filepath.Join(e2.dataDir, "updater"), 0o700)
			plant(e2.verdictPath())
			if _, err := FetchRelease(e2.cfg(r)); !errors.Is(err, ErrVerdictExists) {
				t.Fatalf("err = %v, want ErrVerdictExists", err)
			}
			if got := dirNames(t, e2.releaseRoot); len(got) != 0 {
				t.Fatalf("release root = %v, want empty", got)
			}
			if v, err := ReadReleaseVerdict(e2.dataDir, testReleaseID); !errors.Is(err, ErrVerdict) {
				t.Fatalf("ReadReleaseVerdict on %s = %+v, %v; want ErrVerdict (absent fields are not empty ones)", name, v, err)
			}
		})
	}
	if p, err := VerdictPath(e.dataDir, testReleaseID); err != nil || p != e.verdictPath() {
		t.Fatalf("VerdictPath = %q, %v; want %q", p, err, e.verdictPath())
	}
	for _, bad := range []struct{ dir, id string }{{"relative/data", testReleaseID}, {e.dataDir, "../x"}, {e.dataDir, ""}, {"", testReleaseID}} {
		if p, err := VerdictPath(bad.dir, bad.id); err == nil {
			t.Fatalf("VerdictPath(%q, %q) = %q, want a refusal", bad.dir, bad.id, p)
		}
	}
}

// ── no network, no MAC, no activation: the unit's imports, pinned ────────────

func TestReleaseFetchHasNoNetworkCode(t *testing.T) {
	forbidden := map[string]string{
		"net": "network", "net/http": "network", "net/url": "network", "net/rpc": "network",
		"os/exec":                  "a subprocess (verification is in-process, never ssh-keygen at run time)",
		"crypto/hmac":              "a MAC (only internal/updateauth computes one — census rule 4)",
		"nofx/internal/updateauth": "the enrollment/MAC package",
		"nofx/internal/activation": "the kill library (not on dev; U4 owns the adapter)",
		"golang.org/x/crypto/ssh":  "a dependency the stdlib verifier does not need",
	}
	for _, file := range []string{"sshsig.go", "release.go"} {
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range f.Imports {
			p, _ := strconv.Unquote(im.Path.Value)
			if why, bad := forbidden[p]; bad {
				t.Errorf("%s imports %s — %s", file, p, why)
			}
		}
	}
}

// Names the materialized layout owns (RELEASE, signed/…) may not be artifacts:
// they would collide with what fetch writes. Put there BEFORE manifest.sh runs,
// so they are listed with their true hashes and only this rule can refuse them.
func TestManifestRefusesArtifactsTheLayoutOwns(t *testing.T) {
	for name, extra := range map[string]map[string]string{
		"RELEASE at the root":  {"RELEASE": testSHA + "\n"},
		"a file under signed/": {"signed/manifest.json": "{}"},
		"the signature name":   {"manifest.json.sig": "x"},
	} {
		t.Run(name, func(t *testing.T) {
			r := buildRelease(t, releaseOpts{beforeManifest: func(t *testing.T, s string) { writeFiles(t, s, extra) }})
			e := newFetchEnv(t)
			v, err := FetchRelease(e.cfg(r))
			if err == nil {
				t.Fatalf("ACCEPTED a manifest listing %v; verdict %+v", extra, v)
			}
			if !errors.Is(err, ErrManifest) || !strings.Contains(err.Error(), "the materialized layout owns") {
				t.Fatalf("err = %v, want ErrManifest (a name the materialized layout owns)", err)
			}
			e.assertNothingLanded(t)
		})
	}
}

// Only zero padding may follow the tar end, and the gzip stream is read to
// its end so its CRC is checked.
func TestExtractRefusesDataAfterTheTarEnd(t *testing.T) {
	r := buildRelease(t, releaseOpts{})
	clean := repack(t, r.stage)
	raw, err := os.ReadFile(clean)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	var tarBytes bytes.Buffer
	if _, err := tarBytes.ReadFrom(zr); err != nil {
		t.Fatal(err)
	}
	regz := func(b []byte) string {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, _ = w.Write(b)
		_ = w.Close()
		p := filepath.Join(t.TempDir(), testReleaseID+".tar.gz")
		if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	crc := append([]byte(nil), raw...)
	crc[len(crc)-8] ^= 0xff // the gzip trailer's CRC-32
	crcPath := filepath.Join(t.TempDir(), testReleaseID+".tar.gz")
	if err := os.WriteFile(crcPath, crc, 0o644); err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct {
		archive string
		ok      bool
	}{
		"control: zero padding after the end": {regz(append(tarBytes.Bytes(), make([]byte, 10240)...)), true},
		"a payload after the end":             {regz(append(tarBytes.Bytes(), []byte("#!/bin/sh\nhidden\n")...)), false},
		"a corrupted gzip CRC":                {crcPath, false},
	} {
		t.Run(name, func(t *testing.T) {
			e := newFetchEnv(t)
			cfg := e.cfg(r)
			cfg.Archive = c.archive
			v, err := FetchRelease(cfg)
			if c.ok {
				if err != nil {
					t.Fatalf("control refused: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ACCEPTED %s; verdict %+v", name, v)
			}
			if !errors.Is(err, ErrArchive) {
				t.Fatalf("%s: err = %v, want ErrArchive", name, err)
			}
			e.assertNothingLanded(t)
		})
	}
}

// The job's verified state re-proves the signature NOW, against the current
// trust anchor, and refuses an activation manifest whose verdict it did not
// itself produce.
func TestReverifyRefusesAVerdictItDidNotProve(t *testing.T) {
	r := buildRelease(t, releaseOpts{})
	foreign := newTestSigner(t, t.TempDir(), "foreign")
	for name, c := range map[string]struct {
		tamper func(t *testing.T, dir string) (signers, id string)
		want   error
	}{
		"control": {func(t *testing.T, d string) (string, string) { return r.signers, testReleaseID }, nil},
		"forged signature_verdict": {func(t *testing.T, d string) (string, string) {
			p := filepath.Join(d, "manifest.json")
			b, _ := os.ReadFile(p)
			_ = os.WriteFile(p, bytes.Replace(b, []byte("sshsig:release:SHA256:"), []byte("sshsig:release:SHA256:forged"), 1), 0o644)
			return r.signers, testReleaseID
		}, ErrLayout},
		"the trust anchor changed since fetch": {func(t *testing.T, d string) (string, string) {
			return writeAllowedSigners(t, t.TempDir(), "release "+foreign.pub), testReleaseID
		}, ErrSigForeignKey},
		"another release id": {func(t *testing.T, d string) (string, string) { return r.signers, "v0.0.2-u3" }, ErrManifest},
		"signed manifest edited": {func(t *testing.T, d string) (string, string) {
			p := filepath.Join(d, "signed", "manifest.json")
			b, _ := os.ReadFile(p)
			_ = os.WriteFile(p, append(b, ' '), 0o644)
			return r.signers, testReleaseID
		}, ErrSigInvalid},
	} {
		t.Run(name, func(t *testing.T) {
			e := newFetchEnv(t)
			v, err := FetchRelease(e.cfg(r))
			if err != nil {
				t.Fatal(err)
			}
			signers, id := c.tamper(t, e.releaseDirOf(t, v))
			m, sv, err := ReverifyRelease(v.ReleaseDir, signers, id)
			if c.want == nil {
				if err != nil || m.SourceSHA != testSHA || sv.String() != "sshsig:release:"+r.fp {
					t.Fatalf("control: %+v %+v %v", m, sv, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ReverifyRelease ACCEPTED (%s): %+v %+v", name, m, sv)
			}
			if !errors.Is(err, c.want) {
				t.Fatalf("%s: err = %v, want %v", name, err, c.want)
			}
		})
	}
}
