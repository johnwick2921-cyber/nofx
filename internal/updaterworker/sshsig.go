package updaterworker

// sshsig.go — W-ONE-BUTTON M4 (3b-B, unit U3): verify the OpenSSH SSHSIG
// signature that 3a's release workflow puts beside every manifest
// (`ssh-keygen -Y sign -f <key> -n release manifest.json`, release.yml), with
// the standard library only.
//
// POLICY (brief C6, CTO ruling 1790259689740 — stricter than ssh-keygen, never
// looser). A signature is accepted only when ALL of these hold:
//
//   - the armor is "-----BEGIN SSH SIGNATURE-----" … "-----END SSH SIGNATURE-----";
//   - the blob is magic "SSHSIG", version 1, with nothing after the signature;
//   - the signing key is a plain ssh-ed25519 key (no certificate, no sk-, no
//     RSA/ECDSA);
//   - the namespace is "release" (checked, AND bound into the signed data);
//   - the reserved field is empty (what ssh-keygen writes);
//   - the hash algorithm is sha512 (ssh-keygen also accepts sha256; we do not);
//   - the signing key is BYTE-EQUAL to an ssh-ed25519 key the allowed-signers
//     file lists for the principal "release" — by exact principal, never by
//     pattern: a wildcard line ssh-keygen would honour admits nothing here,
//     and a line whose principal list carries ANY negation ("!…") admits
//     nothing either (ssh-keygen vetoes the line only when the negated
//     pattern matches; not implementing patterns, we veto it always);
//   - ed25519 verifies over "SSHSIG" ‖ string(namespace) ‖ string("") ‖
//     string("sha512") ‖ string(SHA-512(message)) (PROTOCOL.sshsig).
//
// The allowed-signers file is READ AT RUN TIME from the path the caller passes
// (the worker passes <installDir>/deploy/release_allowed_signers). Absent ⇒
// refused: no release can be installed until the owner commits the file
// (deploy/release/README.md, "Owner action, once").
//
// Cross-checked against the real tool: TestSSHSIGVerifiesAnSshKeygenSignature
// and every TestSSHSIGRefuses* case also run `ssh-keygen -Y verify`, so each
// refusal is shown to be the one its name claims.

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ReleaseSignaturePrincipal is the ONLY principal whose keys may sign a
	// release (release.yml verifies with -I release).
	ReleaseSignaturePrincipal = "release"
	// ReleaseSignatureNamespace is the ONLY namespace a release signature may
	// carry (release.yml signs with -n release).
	ReleaseSignatureNamespace = "release"
	// ReleaseSignatureHashAlg is the ONLY hash algorithm accepted
	// (ssh-keygen's default; the sha256 it would also accept is refused).
	ReleaseSignatureHashAlg = "sha512"

	sshsigMagic   = "SSHSIG"
	sshsigVersion = 1
	sshEd25519    = "ssh-ed25519"
	armorBegin    = "-----BEGIN SSH SIGNATURE-----"
	armorEnd      = "-----END SSH SIGNATURE-----"

	// MaxAllowedSignersBytes / MaxSignatureBytes cap what is read: a trust
	// anchor or a signature larger than this is refused, never truncated.
	MaxAllowedSignersBytes = 64 << 10
	MaxSignatureBytes      = 64 << 10
)

// Each refusal is a sentinel so a test can prove WHICH check refused.
var (
	ErrNoAllowedSigners     = errors.New("sshsig: no allowed-signers file")
	ErrAllowedSignersUnsafe = errors.New("sshsig: allowed-signers file is unsafe")
	ErrAllowedSignersBad    = errors.New("sshsig: allowed-signers file is malformed")
	ErrSigFormat            = errors.New("sshsig: signature is malformed")
	ErrSigKeyType           = errors.New("sshsig: signing key is not a plain ssh-ed25519 key")
	ErrSigNamespace         = errors.New("sshsig: wrong namespace")
	ErrSigHashAlg           = errors.New("sshsig: hash algorithm is not sha512")
	ErrSigPrincipal         = errors.New("sshsig: allowed-signers lists no ssh-ed25519 key for principal release")
	ErrSigForeignKey        = errors.New("sshsig: signing key is not an allowed release key")
	ErrSigInvalid           = errors.New("sshsig: signature does not verify")
)

// SignatureVerdict is what a verified signature proves.
type SignatureVerdict struct {
	Principal   string // always ReleaseSignaturePrincipal
	Namespace   string // always ReleaseSignatureNamespace
	HashAlg     string // always ReleaseSignatureHashAlg
	Fingerprint string // "SHA256:<unpadded base64>" of the signer's key blob — the `ssh-keygen -l` form
}

// String is the manifest's signature_verdict: "sshsig:release:SHA256:…".
func (v SignatureVerdict) String() string {
	return "sshsig:" + v.Principal + ":" + v.Fingerprint
}

// ReleaseAllowedSignersPath is where an installation keeps the release trust
// anchor: <installDir>/deploy/release_allowed_signers (release.yml verifies
// against the same committed file).
func ReleaseAllowedSignersPath(installDir string) string {
	return filepath.Join(installDir, "deploy", "release_allowed_signers")
}

// VerifySSHSIG verifies an armored SSHSIG signature over message against the
// allowed-signers file at allowedSignersPath, under the policy above. The
// allowed-signers file is read first: absent or unsafe refuses before the
// signature is even parsed.
func VerifySSHSIG(message, armored []byte, allowedSignersPath string) (SignatureVerdict, error) {
	keys, err := releaseSignerKeys(allowedSignersPath)
	if err != nil {
		return SignatureVerdict{}, err
	}
	blob, err := dearmorSSHSIG(armored)
	if err != nil {
		return SignatureVerdict{}, err
	}
	r := &sshReader{b: blob}
	magic, ok := r.raw(len(sshsigMagic))
	if !ok || string(magic) != sshsigMagic {
		return SignatureVerdict{}, fmt.Errorf("%w: magic is not %q", ErrSigFormat, sshsigMagic)
	}
	version, ok := r.u32()
	if !ok || version != sshsigVersion {
		return SignatureVerdict{}, fmt.Errorf("%w: version %d, want %d", ErrSigFormat, version, sshsigVersion)
	}
	pubBlob, ok1 := r.str()
	namespace, ok2 := r.str()
	reserved, ok3 := r.str()
	hashAlg, ok4 := r.str()
	sigBlob, ok5 := r.str()
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return SignatureVerdict{}, fmt.Errorf("%w: truncated signature object", ErrSigFormat)
	}
	if len(r.b) != 0 {
		return SignatureVerdict{}, fmt.Errorf("%w: %d byte(s) of trailing data", ErrSigFormat, len(r.b))
	}
	pub, err := parseEd25519Blob(pubBlob)
	if err != nil {
		return SignatureVerdict{}, err
	}
	if string(namespace) != ReleaseSignatureNamespace {
		return SignatureVerdict{}, fmt.Errorf("%w: %q, want %q", ErrSigNamespace, namespace, ReleaseSignatureNamespace)
	}
	if len(reserved) != 0 {
		return SignatureVerdict{}, fmt.Errorf("%w: reserved field is not empty", ErrSigFormat)
	}
	if string(hashAlg) != ReleaseSignatureHashAlg {
		return SignatureVerdict{}, fmt.Errorf("%w: %q", ErrSigHashAlg, hashAlg)
	}
	sr := &sshReader{b: sigBlob}
	sigType, okT := sr.str()
	rawSig, okS := sr.str()
	if !okT || !okS || len(sr.b) != 0 || string(sigType) != sshEd25519 || len(rawSig) != ed25519.SignatureSize {
		return SignatureVerdict{}, fmt.Errorf("%w: signature is not one ssh-ed25519 signature", ErrSigFormat)
	}
	if len(keys) == 0 {
		return SignatureVerdict{}, fmt.Errorf("%w (%s)", ErrSigPrincipal, allowedSignersPath)
	}
	listed := false
	for _, k := range keys {
		listed = listed || bytes.Equal(k, pubBlob)
	}
	if !listed {
		return SignatureVerdict{}, fmt.Errorf("%w: %s is not listed for principal %q in %s",
			ErrSigForeignKey, keyFingerprint(pubBlob), ReleaseSignaturePrincipal, allowedSignersPath)
	}
	// PROTOCOL.sshsig: the signed data binds OUR namespace and hash algorithm,
	// never the blob's — a check that is skipped still cannot be satisfied by
	// a signature made for another namespace.
	digest := sha512.Sum512(message)
	var signed []byte
	signed = append(signed, sshsigMagic...)
	signed = appendSSHString(signed, []byte(ReleaseSignatureNamespace))
	signed = appendSSHString(signed, nil)
	signed = appendSSHString(signed, []byte(ReleaseSignatureHashAlg))
	signed = appendSSHString(signed, digest[:])
	if !ed25519.Verify(pub, signed, rawSig) {
		return SignatureVerdict{}, ErrSigInvalid
	}
	return SignatureVerdict{
		Principal:   ReleaseSignaturePrincipal,
		Namespace:   ReleaseSignatureNamespace,
		HashAlg:     ReleaseSignatureHashAlg,
		Fingerprint: keyFingerprint(pubBlob),
	}, nil
}

// releaseSignerKeys reads the allowed-signers file and returns the key blob
// of every ssh-ed25519 key it lists for the principal "release". The file is
// a trust anchor: it must be a regular file (never a symlink) that neither
// group nor other can write. Any malformed line refuses the WHOLE file — a
// parser that skips what it cannot read decides trust on a file nobody wrote.
// A line with options (namespaces=, valid-before=, cert-authority, …) that
// names release is refused: those are restrictions ssh-keygen enforces and
// this verifier does not implement, so ignoring them would widen trust.
func releaseSignerKeys(path string) ([][]byte, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrNoAllowedSigners, path, err)
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s is not a regular file (%s)", ErrAllowedSignersUnsafe, path, fi.Mode().Type())
	}
	if fi.Mode().Perm()&0o022 != 0 {
		return nil, fmt.Errorf("%w: %s is writable by group or other (%#o)", ErrAllowedSignersUnsafe, path, fi.Mode().Perm())
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrNoAllowedSigners, path, err)
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, MaxAllowedSignersBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrAllowedSignersBad, path, err)
	}
	if len(raw) > MaxAllowedSignersBytes {
		return nil, fmt.Errorf("%w: %s is larger than %d bytes", ErrAllowedSignersBad, path, MaxAllowedSignersBytes)
	}
	var keys [][]byte
	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		forRelease, negated := false, false
		for _, p := range strings.Split(fields[0], ",") {
			forRelease = forRelease || p == ReleaseSignaturePrincipal
			negated = negated || strings.HasPrefix(p, "!")
		}
		// ssh-keygen lets a matching negation ("release,!rel*") veto the line;
		// patterns are not implemented here, so ANY negation vetoes it.
		if !forRelease || negated {
			continue
		}
		if len(fields) < 3 {
			return nil, fmt.Errorf("%w: %s:%d: a release line needs <principals> <keytype> <base64>", ErrAllowedSignersBad, path, i+1)
		}
		if !looksLikeKeyType(fields[1]) {
			return nil, fmt.Errorf("%w: %s:%d: options on a release line are not supported (%q)", ErrAllowedSignersBad, path, i+1, fields[1])
		}
		if fields[1] != sshEd25519 {
			continue // can never verify under this policy; admits nothing
		}
		blob, err := base64.StdEncoding.DecodeString(fields[2])
		if err != nil {
			return nil, fmt.Errorf("%w: %s:%d: key is not base64: %w", ErrAllowedSignersBad, path, i+1, err)
		}
		if _, err := parseEd25519Blob(blob); err != nil {
			return nil, fmt.Errorf("%w: %s:%d: %w", ErrAllowedSignersBad, path, i+1, err)
		}
		keys = append(keys, blob)
	}
	return keys, nil
}

// looksLikeKeyType reports whether an allowed-signers field is a key type
// rather than an option list (OpenSSH key type names).
func looksLikeKeyType(s string) bool {
	for _, p := range []string{"ssh-", "ecdsa-", "sk-", "rsa-", "x509v3-"} {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// parseEd25519Blob returns the 32-byte key of a plain ssh-ed25519 public key
// blob: string("ssh-ed25519") ‖ string(key), nothing after.
func parseEd25519Blob(blob []byte) (ed25519.PublicKey, error) {
	r := &sshReader{b: blob}
	kt, ok1 := r.str()
	k, ok2 := r.str()
	if !ok1 || !ok2 || len(r.b) != 0 {
		return nil, fmt.Errorf("%w: public key blob is malformed", ErrSigFormat)
	}
	if string(kt) != sshEd25519 {
		return nil, fmt.Errorf("%w: %q", ErrSigKeyType, kt)
	}
	if len(k) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("%w: ed25519 key is %d bytes", ErrSigFormat, len(k))
	}
	return ed25519.PublicKey(k), nil
}

// keyFingerprint is OpenSSH's SHA256 fingerprint of a key blob.
func keyFingerprint(blob []byte) string {
	sum := sha256.Sum256(blob)
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(sum[:])
}

// dearmorSSHSIG strips the armor ssh-keygen -Y sign writes: the BEGIN line
// first, base64 lines, the END line, then nothing but whitespace.
func dearmorSSHSIG(armored []byte) ([]byte, error) {
	if len(armored) > MaxSignatureBytes {
		return nil, fmt.Errorf("%w: larger than %d bytes", ErrSigFormat, MaxSignatureBytes)
	}
	s := string(armored)
	if !strings.HasPrefix(s, armorBegin+"\n") {
		return nil, fmt.Errorf("%w: does not begin with %q", ErrSigFormat, armorBegin)
	}
	s = s[len(armorBegin)+1:]
	end := strings.Index(s, armorEnd)
	if end < 0 {
		return nil, fmt.Errorf("%w: no %q line", ErrSigFormat, armorEnd)
	}
	if strings.TrimSpace(s[end+len(armorEnd):]) != "" {
		return nil, fmt.Errorf("%w: text after %q", ErrSigFormat, armorEnd)
	}
	body := strings.ReplaceAll(s[:end], "\n", "")
	blob, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, fmt.Errorf("%w: armor body: %w", ErrSigFormat, err)
	}
	return blob, nil
}

// sshReader reads the SSH wire encoding (RFC 4251 §5): uint32 big-endian,
// and string = uint32 length ‖ bytes.
type sshReader struct{ b []byte }

func (r *sshReader) raw(n int) ([]byte, bool) {
	if n < 0 || len(r.b) < n {
		return nil, false
	}
	v := r.b[:n]
	r.b = r.b[n:]
	return v, true
}

func (r *sshReader) u32() (uint32, bool) {
	v, ok := r.raw(4)
	if !ok {
		return 0, false
	}
	return binary.BigEndian.Uint32(v), true
}

func (r *sshReader) str() ([]byte, bool) {
	n, ok := r.u32()
	if !ok || uint64(n) > uint64(len(r.b)) {
		return nil, false
	}
	return r.raw(int(n))
}

func appendSSHString(dst, v []byte) []byte {
	dst = binary.BigEndian.AppendUint32(dst, uint32(len(v)))
	return append(dst, v...)
}
