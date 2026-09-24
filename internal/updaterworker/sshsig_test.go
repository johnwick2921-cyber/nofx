package updaterworker

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ── SSHSIG, cross-checked against the REAL ssh-keygen ─────────────────────────
//
// Every signature here is made by `ssh-keygen -Y sign` (the tool release.yml
// signs with), and every case also asks `ssh-keygen -Y verify -I release -n
// release` (release.yml's own verification form) for its opinion, so a
// refusal is shown to be the one the test name claims and not a fixture that
// was broken some other way. Where our policy is deliberately STRICTER than
// the tool (sha256 hashalg, a wildcard principal), the test proves the tool
// accepts the fixture and we refuse it.

// sshKeygen returns the real tool or skips the test with the reason.
func sshKeygen(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("ssh-keygen")
	if err != nil {
		t.Skip("ssh-keygen is absent on this box — the SSHSIG tests sign and cross-verify with the REAL tool (OpenSSH >= 8.1), never a Go re-implementation")
	}
	return p
}

// keygenEnv keeps an ssh-agent and the user's ~/.ssh out of every run.
func keygenEnv(t *testing.T) []string {
	return []string{"PATH=" + os.Getenv("PATH"), "HOME=" + t.TempDir(), "LC_ALL=C"}
}

func runKeygen(t *testing.T, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(sshKeygen(t), args...)
	cmd.Env = keygenEnv(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ssh-keygen %v: %v\n%s", args, err, out)
	}
	return out
}

type testSigner struct {
	priv string // the private half's file (a temp dir; never the repo)
	pub  string // "ssh-ed25519 AAAA… comment" — the .pub line
}

func newTestSigner(t *testing.T, dir, name string) testSigner {
	t.Helper()
	priv := filepath.Join(dir, name)
	runKeygen(t, "-q", "-t", "ed25519", "-N", "", "-C", "nofx-test-"+name, "-f", priv)
	pub, err := os.ReadFile(priv + ".pub")
	if err != nil {
		t.Fatal(err)
	}
	return testSigner{priv: priv, pub: strings.TrimSpace(string(pub))}
}

// fingerprint is `ssh-keygen -l -E sha256 -f <pub>`'s second field.
func (s testSigner) fingerprint(t *testing.T) string {
	t.Helper()
	f := strings.Fields(string(runKeygen(t, "-l", "-E", "sha256", "-f", s.priv+".pub")))
	if len(f) < 2 || !strings.HasPrefix(f[1], "SHA256:") {
		t.Fatalf("ssh-keygen -l printed %q", f)
	}
	return f[1]
}

// sign signs the file at msgPath with `ssh-keygen -Y sign` and returns the
// armored signature it wrote to msgPath.sig.
func (s testSigner) sign(t *testing.T, msgPath, namespace string, extra ...string) []byte {
	t.Helper()
	_ = os.Remove(msgPath + ".sig")
	args := append([]string{"-Y", "sign", "-f", s.priv, "-n", namespace}, extra...)
	runKeygen(t, append(args, msgPath)...)
	sig, err := os.ReadFile(msgPath + ".sig")
	if err != nil {
		t.Fatal(err)
	}
	return sig
}

func writeAllowedSigners(t *testing.T, dir string, lines ...string) string {
	t.Helper()
	p := filepath.Join(dir, fmt.Sprintf("allowed_signers_%d", len(lines)))
	for i := 0; ; i++ {
		if _, err := os.Lstat(p); errors.Is(err, fs.ErrNotExist) {
			break
		}
		p = filepath.Join(dir, fmt.Sprintf("allowed_signers_%d_%d", len(lines), i))
	}
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// keygenVerify is release.yml's verification, verbatim in form.
func keygenVerify(t *testing.T, signers string, sig, msg []byte) error {
	t.Helper()
	sigPath := filepath.Join(t.TempDir(), "m.sig")
	if err := os.WriteFile(sigPath, sig, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(sshKeygen(t), "-Y", "verify", "-f", signers, "-I", "release", "-n", "release", "-s", sigPath)
	cmd.Stdin = bytes.NewReader(msg)
	cmd.Env = keygenEnv(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v: %s", err, bytes.TrimSpace(out))
	}
	return nil
}

type sigFixture struct {
	dir     string
	signer  testSigner
	msg     []byte
	msgPath string
	signers string // lists signer under principal "release" (README form, comment included)
}

func newSigFixture(t *testing.T) sigFixture {
	t.Helper()
	sshKeygen(t)
	dir := t.TempDir()
	s := newTestSigner(t, dir, "signer")
	msg := []byte(`{"release_id":"v0.0.1-test","source_sha":"` + strings.Repeat("ab", 20) + `","artifacts":[]}` + "\n")
	msgPath := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(msgPath, msg, 0o644); err != nil {
		t.Fatal(err)
	}
	return sigFixture{dir: dir, signer: s, msg: msg, msgPath: msgPath, signers: writeAllowedSigners(t, dir, "release "+s.pub)}
}

func TestSSHSIGVerifiesAnSshKeygenSignature(t *testing.T) {
	f := newSigFixture(t)
	sig := f.signer.sign(t, f.msgPath, "release")
	if err := keygenVerify(t, f.signers, sig, f.msg); err != nil {
		t.Fatalf("fixture: the real ssh-keygen refuses its own signature: %v", err)
	}
	fp := f.signer.fingerprint(t)
	// README form ("release <keytype> <base64> <comment>"), plus the shapes a
	// hand-kept file has: comments, blank lines, another principal's line,
	// and a principal LIST that contains release.
	other := newTestSigner(t, f.dir, "other")
	busy := writeAllowedSigners(t, f.dir,
		"# nofx release trust anchor",
		"",
		"ci "+other.pub,
		"ci,release "+f.signer.pub,
	)
	for name, signers := range map[string]string{"readme form": f.signers, "busy file": busy} {
		t.Run(name, func(t *testing.T) {
			if err := keygenVerify(t, signers, sig, f.msg); err != nil {
				t.Fatalf("fixture: ssh-keygen refuses %s: %v", name, err)
			}
			v, err := VerifySSHSIG(f.msg, sig, signers)
			if err != nil {
				t.Fatalf("VerifySSHSIG refused a signature ssh-keygen accepts: %v", err)
			}
			if v.Fingerprint != fp {
				t.Fatalf("fingerprint = %q, ssh-keygen -l says %q", v.Fingerprint, fp)
			}
			if v.Principal != "release" || v.Namespace != "release" || v.HashAlg != "sha512" {
				t.Fatalf("verdict = %+v", v)
			}
			if got, want := v.String(), "sshsig:release:"+fp; got != want {
				t.Fatalf("signature_verdict = %q, want %q", got, want)
			}
		})
	}
}

// refuseBoth asserts ssh-keygen's opinion (toolAccepts) and that ours refuses
// with want.
func refuseBoth(t *testing.T, signers string, sig, msg []byte, toolAccepts bool, want error) {
	t.Helper()
	kerr := keygenVerify(t, signers, sig, msg)
	if toolAccepts && kerr != nil {
		t.Fatalf("fixture: ssh-keygen was expected to ACCEPT (our policy is the stricter one) but refused: %v", kerr)
	}
	if !toolAccepts && kerr == nil {
		t.Fatalf("fixture: ssh-keygen ACCEPTS this signature — the fixture is not broken the way the test claims")
	}
	v, err := VerifySSHSIG(msg, sig, signers)
	if err == nil {
		t.Fatalf("VerifySSHSIG ACCEPTED (verdict %q); want %v", v.String(), want)
	}
	if !errors.Is(err, want) {
		t.Fatalf("VerifySSHSIG refused with %v; want %v", err, want)
	}
}

func TestSSHSIGRefusesWrongNamespace(t *testing.T) {
	f := newSigFixture(t)
	sig := f.signer.sign(t, f.msgPath, "file") // a real key, the wrong namespace
	refuseBoth(t, f.signers, sig, f.msg, false, ErrSigNamespace)
}

func TestSSHSIGRefusesWrongPrincipal(t *testing.T) {
	f := newSigFixture(t)
	sig := f.signer.sign(t, f.msgPath, "release")
	t.Run("the key listed under another principal", func(t *testing.T) {
		refuseBoth(t, writeAllowedSigners(t, f.dir, "ci "+f.signer.pub), sig, f.msg, false, ErrSigPrincipal)
	})
	t.Run("a negated principal", func(t *testing.T) {
		refuseBoth(t, writeAllowedSigners(t, f.dir, "!release "+f.signer.pub), sig, f.msg, false, ErrSigPrincipal)
	})
	// A list that names release AND negates a pattern matching it: ssh-keygen
	// refuses (a matching negation wins); so must we — never looser (verifier D1).
	for _, list := range []string{"release,!release", "release,!*", "release,!rel*"} {
		t.Run("a list that names and negates release: "+list, func(t *testing.T) {
			refuseBoth(t, writeAllowedSigners(t, f.dir, list+" "+f.signer.pub), sig, f.msg, false, ErrSigPrincipal)
		})
	}
	// Any negation on the line admits nothing here, even one that cannot match
	// release (ssh-keygen honours "!foo,release"; we do not implement patterns).
	t.Run("a list with a negation ssh-keygen honours", func(t *testing.T) {
		refuseBoth(t, writeAllowedSigners(t, f.dir, "!foo,release "+f.signer.pub), sig, f.msg, true, ErrSigPrincipal)
	})
	// ssh-keygen matches principals as PATTERNS; we match the exact name only.
	t.Run("a wildcard principal ssh-keygen honours", func(t *testing.T) {
		refuseBoth(t, writeAllowedSigners(t, f.dir, "rel* "+f.signer.pub), sig, f.msg, true, ErrSigPrincipal)
	})
}

func TestSSHSIGRefusesForeignKey(t *testing.T) {
	f := newSigFixture(t)
	foreign := newTestSigner(t, f.dir, "foreign")
	sig := foreign.sign(t, f.msgPath, "release") // right namespace, right hash, a key nobody listed
	refuseBoth(t, f.signers, sig, f.msg, false, ErrSigForeignKey)
}

func TestSSHSIGRefusesSha256Hashalg(t *testing.T) {
	f := newSigFixture(t)
	sig := f.signer.sign(t, f.msgPath, "release", "-O", "hashalg=sha256")
	refuseBoth(t, f.signers, sig, f.msg, true, ErrSigHashAlg)
}

func TestSSHSIGRefusesTamperedManifest(t *testing.T) {
	f := newSigFixture(t)
	sig := f.signer.sign(t, f.msgPath, "release")
	for name, msg := range map[string][]byte{
		"one byte changed":  bytes.Replace(f.msg, []byte("v0.0.1-test"), []byte("v0.0.2-test"), 1),
		"a byte appended":   append(append([]byte(nil), f.msg...), ' '),
		"the newline gone":  bytes.TrimSuffix(f.msg, []byte("\n")),
		"an empty manifest": {},
	} {
		t.Run(name, func(t *testing.T) {
			refuseBoth(t, f.signers, sig, msg, false, ErrSigInvalid)
		})
	}
}

// Absent ⇒ refuse (brief C6): no release installs until the owner commits
// deploy/release_allowed_signers.
func TestSSHSIGRefusesAnAbsentOrUnsafeAllowedSignersFile(t *testing.T) {
	f := newSigFixture(t)
	sig := f.signer.sign(t, f.msgPath, "release")
	absent := filepath.Join(f.dir, "deploy", "release_allowed_signers")
	if _, err := VerifySSHSIG(f.msg, sig, absent); !errors.Is(err, ErrNoAllowedSigners) || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("absent allowed-signers: err = %v, want ErrNoAllowedSigners wrapping fs.ErrNotExist", err)
	}
	if got := ReleaseAllowedSignersPath("/srv/nofx"); got != "/srv/nofx/deploy/release_allowed_signers" {
		t.Fatalf("ReleaseAllowedSignersPath = %q", got)
	}
	link := filepath.Join(f.dir, "signers-link")
	if err := os.Symlink(f.signers, link); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySSHSIG(f.msg, sig, link); !errors.Is(err, ErrAllowedSignersUnsafe) {
		t.Fatalf("symlinked allowed-signers: err = %v, want ErrAllowedSignersUnsafe", err)
	}
	loose := writeAllowedSigners(t, f.dir, "release "+f.signer.pub, "# loose")
	if err := os.Chmod(loose, 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySSHSIG(f.msg, sig, loose); !errors.Is(err, ErrAllowedSignersUnsafe) {
		t.Fatalf("group/world-writable allowed-signers: err = %v, want ErrAllowedSignersUnsafe", err)
	}
	// An option on the release line (namespaces=, valid-before=, cert-authority)
	// is a restriction ssh-keygen would enforce; one we do not implement is
	// refused rather than ignored.
	opt := writeAllowedSigners(t, f.dir, `release namespaces="release" `+f.signer.pub)
	if err := keygenVerify(t, opt, sig, f.msg); err != nil {
		t.Fatalf("fixture: ssh-keygen refuses the optioned line: %v", err)
	}
	if _, err := VerifySSHSIG(f.msg, sig, opt); !errors.Is(err, ErrAllowedSignersBad) {
		t.Fatalf("optioned release line: err = %v, want ErrAllowedSignersBad", err)
	}
	bad := writeAllowedSigners(t, f.dir, "release ssh-ed25519 not-base64!!", "#")
	if _, err := VerifySSHSIG(f.msg, sig, bad); !errors.Is(err, ErrAllowedSignersBad) {
		t.Fatalf("undecodable release key: err = %v, want ErrAllowedSignersBad", err)
	}
}

// ── envelope mutations of a REAL signature ───────────────────────────────────

func sshString(b []byte) []byte {
	out := binary.BigEndian.AppendUint32(nil, uint32(len(b)))
	return append(out, b...)
}

type sigFields struct {
	magic     []byte
	version   uint32
	pub, ns   []byte
	reserved  []byte
	hash, sig []byte
	trailing  []byte
}

func decodeRealSig(t *testing.T, armored []byte) sigFields {
	t.Helper()
	s := strings.TrimSpace(string(armored))
	s = strings.TrimPrefix(s, "-----BEGIN SSH SIGNATURE-----")
	s = strings.TrimSuffix(s, "-----END SSH SIGNATURE-----")
	blob, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(s, "\n", ""))
	if err != nil {
		t.Fatal(err)
	}
	var f sigFields
	f.magic, blob = blob[:6], blob[6:]
	f.version, blob = binary.BigEndian.Uint32(blob), blob[4:]
	next := func() []byte {
		n := binary.BigEndian.Uint32(blob)
		v := blob[4 : 4+n]
		blob = blob[4+n:]
		return v
	}
	f.pub, f.ns, f.reserved, f.hash, f.sig = next(), next(), next(), next(), next()
	f.trailing = blob
	return f
}

func (f sigFields) armor() []byte {
	var blob []byte
	blob = append(blob, f.magic...)
	blob = binary.BigEndian.AppendUint32(blob, f.version)
	for _, v := range [][]byte{f.pub, f.ns, f.reserved, f.hash, f.sig} {
		blob = append(blob, sshString(v)...)
	}
	blob = append(blob, f.trailing...)
	b64 := base64.StdEncoding.EncodeToString(blob)
	var sb strings.Builder
	sb.WriteString("-----BEGIN SSH SIGNATURE-----\n")
	for len(b64) > 70 {
		sb.WriteString(b64[:70] + "\n")
		b64 = b64[70:]
	}
	sb.WriteString(b64 + "\n-----END SSH SIGNATURE-----\n")
	return []byte(sb.String())
}

func TestSSHSIGRefusesMalformedEnvelopes(t *testing.T) {
	f := newSigFixture(t)
	real := f.signer.sign(t, f.msgPath, "release")
	base := decodeRealSig(t, real)
	// control: the re-armoring is faithful, so each mutant differs ONLY by its mutation
	if _, err := VerifySSHSIG(f.msg, base.armor(), f.signers); err != nil {
		t.Fatalf("control: the faithfully re-armored real signature is refused: %v", err)
	}
	certType := sshString([]byte("ssh-ed25519-cert-v01@openssh.com"))
	for name, c := range map[string]struct {
		armored []byte
		want    error
	}{
		"bad magic":          {func() []byte { m := base; m.magic = []byte("SSHSIH"); return m.armor() }(), ErrSigFormat},
		"version 2":          {func() []byte { m := base; m.version = 2; return m.armor() }(), ErrSigFormat},
		"trailing data":      {func() []byte { m := base; m.trailing = []byte{0}; return m.armor() }(), ErrSigFormat},
		"non-empty reserved": {func() []byte { m := base; m.reserved = []byte("x"); return m.armor() }(), ErrSigFormat},
		"certificate key type": {func() []byte {
			m := base
			m.pub = append(certType, base.pub[len(sshString([]byte("ssh-ed25519"))):]...)
			return m.armor()
		}(), ErrSigKeyType},
		"no armor":          {[]byte(strings.SplitN(string(real), "\n", 2)[1]), ErrSigFormat},
		"text before armor": {append([]byte("x\n"), real...), ErrSigFormat},
		"no end line":       {[]byte(strings.Replace(string(real), "-----END SSH SIGNATURE-----", "", 1)), ErrSigFormat},
		"text after armor":  {append(append([]byte(nil), real...), []byte("junk\n")...), ErrSigFormat},
		"empty":             {nil, ErrSigFormat},
		"oversized":         {bytes.Repeat([]byte("A"), MaxSignatureBytes+1), ErrSigFormat},
	} {
		t.Run(name, func(t *testing.T) {
			v, err := VerifySSHSIG(f.msg, c.armored, f.signers)
			if err == nil {
				t.Fatalf("ACCEPTED a %s envelope (verdict %q)", name, v.String())
			}
			if !errors.Is(err, c.want) {
				t.Fatalf("%s: err = %v, want %v", name, err, c.want)
			}
		})
	}
}
