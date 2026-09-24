package updaterjob

// verdict.go — W-ONE-BUTTON M4 (3b-B, U3 repair D3): the release verdict file
// <dataDir>/updater/verdicts/<release_id>.json (brief §3.2), READ here.
//
// Why here: the worker's `fetch` (internal/updaterworker.FetchRelease) writes
// the verdict once, after every check has passed, but the API's install gate
// (U5b, brief §3.8: updaterjob.ReadVerdict(trader.MaintenanceDataDir(), id))
// must read it, and the trading app may not link nofx/internal/updaterworker
// (store/maintenance_hold_writers_test.go, forbiddenWorkerPackages). This
// package is app-linkable and imports only the standard library and
// internal/updaterwire.
//
// READ-ONLY ON PURPOSE. Nothing in this file creates, links, renames, chmods
// or removes a file (TestVerdictFileHasNoWriter pins it). The ONE writer stays
// in the worker package, which the app cannot link: a writer here would let
// app code mint the evidence the install gate trusts. (Brief §3.1 listed a
// WriteVerdict here; it is deliberately kept worker-side — fail-closed.)

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"nofx/internal/updaterwire"
)

const (
	// VerdictSchema is the verdict file's "schema". A reader refuses any other value.
	VerdictSchema = 1
	// VerdictSigner and VerdictHashAlg are the only values a verdict may carry
	// (the worker's SSHSIG policy: principal "release", hash sha512).
	VerdictSigner  = "release"
	VerdictHashAlg = "sha512"
	// MaxVerdictBytes caps a verdict file; a larger one is refused, never truncated.
	MaxVerdictBytes = 64 << 10

	verdictsDirName = "verdicts"
)

var (
	// ErrVerdictPath: the data dir or the release id cannot name a verdict.
	ErrVerdictPath = errors.New("updaterjob: verdict path refused")
	// ErrVerdict: the verdict file is absent, unsafe, unreadable or malformed.
	// An absent file also wraps fs.ErrNotExist.
	ErrVerdict = errors.New("updaterjob: verdict file is absent, unreadable or malformed")
)

var (
	verdictSHA40Re  = regexp.MustCompile(`^[0-9a-f]{40}$`)
	verdictSHA256Re = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Verdict is the verdict file (brief §3.2): exactly these keys.
type Verdict struct {
	Schema            int    `json:"schema"`
	ReleaseID         string `json:"release_id"`
	SourceSHA         string `json:"source_sha"`
	ReleaseDir        string `json:"release_dir"`
	Signer            string `json:"signer"`
	SignerFingerprint string `json:"signer_fingerprint"`
	HashAlg           string `json:"hashalg"`
	ManifestSHA256    string `json:"manifest_sha256"`
	Artifacts         int    `json:"artifacts"`
	VerifiedAt        string `json:"verified_at"` // RFC3339Nano, UTC
}

// VerdictPath is <dataDir>/updater/verdicts/<releaseID>.json. It touches no
// filesystem; it refuses a relative data dir and any id the wire refuses.
func VerdictPath(dataDir, releaseID string) (string, error) {
	if dataDir == "" || !filepath.IsAbs(dataDir) {
		return "", fmt.Errorf("%w: data dir %q is not absolute", ErrVerdictPath, dataDir)
	}
	if !updaterwire.ValidReleaseID(releaseID) {
		return "", fmt.Errorf("%w: release id %q is not a valid release id", ErrVerdictPath, releaseID)
	}
	return filepath.Join(filepath.Clean(dataDir), updaterwire.UpdaterDirName, verdictsDirName, releaseID+".json"), nil
}

// Check reports whether every field is computed and in range (never an empty
// stand-in). The worker checks a verdict before writing it; ReadVerdict
// checks it after reading.
func (v Verdict) Check() error {
	if v.Schema != VerdictSchema || !updaterwire.ValidReleaseID(v.ReleaseID) || !verdictSHA40Re.MatchString(v.SourceSHA) ||
		!filepath.IsAbs(v.ReleaseDir) || v.Signer != VerdictSigner || !strings.HasPrefix(v.SignerFingerprint, "SHA256:") ||
		v.HashAlg != VerdictHashAlg || !verdictSHA256Re.MatchString(v.ManifestSHA256) || v.Artifacts <= 0 {
		return fmt.Errorf("%w: a field is absent or out of range: %+v", ErrVerdict, v)
	}
	if _, err := time.Parse(time.RFC3339Nano, v.VerifiedAt); err != nil {
		return fmt.Errorf("%w: verified_at: %w", ErrVerdict, err)
	}
	return nil
}

// ReadVerdict reads and validates the verdict for releaseID: a private
// regular file, exactly the schema's keys, the id asked for, and every field
// computed.
func ReadVerdict(dataDir, releaseID string) (Verdict, error) {
	p, err := VerdictPath(dataDir, releaseID)
	if err != nil {
		return Verdict{}, err
	}
	fi, err := os.Lstat(p)
	if err != nil {
		return Verdict{}, fmt.Errorf("%w: %w", ErrVerdict, err)
	}
	if !fi.Mode().IsRegular() || fi.Mode().Perm()&0o077 != 0 {
		return Verdict{}, fmt.Errorf("%w: %s is not a private regular file (%v)", ErrVerdict, p, fi.Mode())
	}
	f, err := os.Open(p)
	if err != nil {
		return Verdict{}, fmt.Errorf("%w: %w", ErrVerdict, err)
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, MaxVerdictBytes+1))
	if err != nil || len(b) > MaxVerdictBytes {
		return Verdict{}, fmt.Errorf("%w: %s unreadable or oversized (%v)", ErrVerdict, p, err)
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var v Verdict
	if err := dec.Decode(&v); err != nil {
		return Verdict{}, fmt.Errorf("%w: %s: %w", ErrVerdict, p, err)
	}
	if v.ReleaseID != releaseID {
		return Verdict{}, fmt.Errorf("%w: %s names release %q", ErrVerdict, p, v.ReleaseID)
	}
	if err := v.Check(); err != nil {
		return Verdict{}, fmt.Errorf("%s: %w", p, err)
	}
	return v, nil
}
