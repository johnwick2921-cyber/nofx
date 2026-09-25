package updaterworker

// reverifier.go — W-ONE-BUTTON M4 (3b-B, unit U4N item A): the production
// Reverifier, the release re-proof the job's downloaded and verified states
// (and the nt8 rule, the resume and the boot check, through facts) call. It is
// a thin adapter over U3's own functions and nothing else:
//
//	Verdict(id)  updaterjob.ReadVerdict(<data dir>, id)          (the app-side reader),
//	             refused unless its release_dir is <ReleaseRoot(install)>/<source_sha> NOW
//	Rehash(v)    RehashRelease(<v as the verdict file>)
//	Reverify(v)  ReverifyRelease(<v as the verdict file>, <install>/deploy/release_allowed_signers)
//
// Reverify is bound to the verdict it is GIVEN (deps.go: the signed manifest's
// sha256, release_id and source_sha and the signing key must be the ones that
// verdict records — U3's manifestIsTheVerdicts refuses otherwise); it never
// re-reads a verdict of its own. The allowed-signers file is read NOW, at
// every re-proof, from the installation tree (C6/C7): absent ⇒ refused.
//
// The mirror ⇄ verdict-file mapping lives HERE and nowhere else (mirrorVerdict,
// Verdict.file); TestVerdictMirrorIsTheVerdictFile pins its field parity.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"nofx/internal/updaterjob"
)

// NewReleaseReverifier is the production Reverifier for the installation t
// (cmd/nofx-updater's newReverifier). It touches no file: the verdict and the
// allowed-signers file are read at each call.
func NewReleaseReverifier(t Target) (Reverifier, error) {
	if !filepath.IsAbs(t.DataDir) || !filepath.IsAbs(t.InstallDir) {
		return nil, fmt.Errorf("updaterworker: re-proof: data dir %q and install dir %q must be absolute", t.DataDir, t.InstallDir)
	}
	return releaseReverifier{installDir: t.InstallDir, dataDir: t.DataDir, allowedSigners: ReleaseAllowedSignersPath(t.InstallDir)}, nil
}

// releaseReverifier is the production Reverifier (see the file comment).
type releaseReverifier struct {
	installDir     string // the installation (the release root must be outside it)
	dataDir        string // the installation's data dir (the verdicts live under it)
	allowedSigners string // <install>/deploy/release_allowed_signers
}

// Verdict is updaterjob.ReadVerdict's file for releaseID, as the mirror —
// and only while it is a verdict for the CURRENT release root (U4F defect 6,
// a fail-closed default named for the CTO): the release root is re-checked
// NOW (ReleaseRoot: set, absolute, its own resolved path, outside the
// install) and the verdict's release_dir must be exactly <that root>/<its
// source sha>. A release fetched under a root the operator no longer names
// is refused, so no step resolves it (brief row 3: Resolve(<RELEASE_DIR>/<sha>)).
func (r releaseReverifier) Verdict(releaseID string) (Verdict, error) {
	v, err := updaterjob.ReadVerdict(r.dataDir, releaseID)
	if err != nil {
		return Verdict{}, err
	}
	root, err := ReleaseRoot(r.installDir)
	if err != nil {
		return Verdict{}, fmt.Errorf("release %s: %w", releaseID, err)
	}
	if want := filepath.Join(root, v.SourceSHA); v.ReleaseDir != want {
		head := fmt.Sprintf("the verdict for %s names the release dir %s, not %s under the current NOFX_RELEASE_DIR", releaseID, v.ReleaseDir, want)
		vpath, perr := updaterjob.VerdictPath(r.dataDir, releaseID)
		if perr != nil {
			return Verdict{}, fmt.Errorf("%w: %s (and its path: %w)", ErrReleaseRoot, head, perr)
		}
		return Verdict{}, fmt.Errorf("%w: %s — %s", ErrReleaseRoot, head, movedRootGuidance(root, want, v.ReleaseDir, vpath, r.installDir, releaseID))
	}
	return mirrorVerdict(v), nil
}

// movedRootGuidance is the refusal's way out when a verdict names a release
// dir that is not <root>/<sha> under the current release root. Every case it
// can be printed in is correct when followed as printed (U4G verify defect 1):
//
//   - <root>/<sha> absent (the operator moved only NOFX_RELEASE_DIR): BOTH
//     steps (U4F verify note 2) — the verdict is written once, so it goes
//     first, by hand, then the re-fetch;
//   - <root>/<sha> present (the operator MOVED THE DIRECTORY, mv A B): a
//     re-fetch would refuse ("never overwritten"), so removing the verdict
//     first would leave nothing usable — the text says so and moves that dir
//     aside BEFORE the rm (bytes kept, never deleted), and names pointing
//     NOFX_RELEASE_DIR back when the verdict's dir still exists;
//   - <root>/<sha> is the CURRENT release (<root>/current names it), or that
//     cannot be checked, or <root>/<sha> itself cannot be checked: NO rm and
//     NO mv is printed — only pointing NOFX_RELEASE_DIR back (after moving the
//     root back when the verdict's dir is gone). Fail-closed default.
//
// Commands are not shell-quoted (the RecoveryText convention; a named limit).
func movedRootGuidance(root, want, verdictDir, vpath, installDir, releaseID string) string {
	fetch := fmt.Sprintf("nofx-updater --install-dir %s fetch %s", installDir, releaseID)
	oldRoot := filepath.Dir(verdictDir)
	_, oerr := os.Lstat(verdictDir)
	oldThere := oerr == nil
	_, werr := os.Lstat(want)
	switch {
	case errors.Is(werr, fs.ErrNotExist):
		return fmt.Sprintf("to use this release there: (1) remove the old verdict by hand: rm %s (2) then re-fetch it: %s", vpath, fetch)
	case werr != nil:
		return pointBack(fmt.Sprintf("%s cannot be checked (%v), so do not move anything and do NOT remove the verdict", want, werr), verdictDir, oldRoot, oldThere)
	}
	cur := filepath.Join(root, "current")
	isCur, cerr := namesDir(cur, want)
	switch {
	case cerr != nil:
		return pointBack(fmt.Sprintf("%s already exists and a fetch never overwrites it, and whether it is the CURRENT release cannot be checked (%s: %v): do not move it and do NOT remove the verdict", want, cur, cerr), verdictDir, oldRoot, oldThere)
	case isCur:
		return pointBack(fmt.Sprintf("%s already exists and a fetch never overwrites it, and it is the CURRENT release (%s names it): do not move it and do NOT remove the verdict", want, cur), verdictDir, oldRoot, oldThere)
	}
	alt := ""
	if oldThere {
		alt = fmt.Sprintf("or point NOFX_RELEASE_DIR back at %s, where the verdict's release still is; ", oldRoot)
	}
	return fmt.Sprintf("%s already exists and a fetch never overwrites it, so do NOT remove the verdict first; %s"+
		"to use this release there: (1) move that directory aside first: mv -T %s %s (2) then remove the old verdict by hand: rm %s (3) then re-fetch it: %s",
		want, alt, want, filepath.Join(root, ".aside-"+filepath.Base(want)), vpath, fetch)
}

// pointBack is the no-rm, no-mv way out: point NOFX_RELEASE_DIR back at the
// root the verdict names, after moving that root back when its dir is gone.
func pointBack(why, verdictDir, oldRoot string, oldThere bool) string {
	if oldThere {
		return fmt.Sprintf("%s; point NOFX_RELEASE_DIR back at %s", why, oldRoot)
	}
	return fmt.Sprintf("%s; move the release root back so %s exists again, then point NOFX_RELEASE_DIR back at %s", why, verdictDir, oldRoot)
}

// namesDir reports whether the link at cur resolves to dir. An absent cur is
// false; any other error (a dangling link, an unreadable element) is returned.
func namesDir(cur, dir string) (bool, error) {
	if _, err := os.Lstat(cur); errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	rc, err := filepath.EvalSymlinks(cur)
	if err != nil {
		return false, err
	}
	rd, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return false, err
	}
	return rc == rd, nil
}

// Rehash is RehashRelease over the verdict it is given.
func (r releaseReverifier) Rehash(v Verdict) (int, error) {
	return RehashRelease(v.file())
}

// Reverify is ReverifyRelease over the verdict it is given, against the
// installation's allowed-signers file read NOW, and the facts the signed
// manifest proves.
func (r releaseReverifier) Reverify(v Verdict) (ReleaseFacts, error) {
	m, sv, err := ReverifyRelease(v.file(), r.allowedSigners)
	if err != nil {
		return ReleaseFacts{}, err
	}
	arts := make(map[string]string, len(m.Artifacts))
	for _, a := range m.Artifacts {
		arts[a.Path] = a.SHA256
	}
	return ReleaseFacts{
		ReleaseID:         m.ReleaseID,
		SourceSHA:         m.SourceSHA,
		ManifestSHA256:    m.SHA256,
		SignerFingerprint: sv.Fingerprint,
		AddonBuildID:      m.AddonBuildID,
		Artifacts:         arts,
	}, nil
}

// mirrorVerdict is the verdict file as the worker's mirror (the ONE mapping).
func mirrorVerdict(v updaterjob.Verdict) Verdict { return Verdict(v) }

// file is the mirror as the verdict file U3's re-proofs take (the ONE mapping).
func (v Verdict) file() updaterjob.Verdict { return updaterjob.Verdict(v) }
