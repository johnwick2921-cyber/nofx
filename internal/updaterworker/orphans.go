package updaterworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"nofx/internal/updaterwire"
)

// ── D8b (CTO 1790279155144 (4), 1790280128238): the interrupted fetch ──────
//
// The attended fetch renames its staging dir to <release_root>/<source_sha>
// and only THEN links <data>/updater/verdicts/<release_id>.json (the verdict
// is written last, on purpose). A kill between those two steps leaves a
// materialized release dir that no verdict names. Such a dir is never
// activated — every path to activation starts from a verdict (the install
// verb refuses "release not verified"; downloaded/verified re-read it) — but
// it blocks every later fetch of that sha at "the release directory already
// exists (it may be the running release)".
//
// QuarantineInterruptedFetches is the recovery the fetch path runs FIRST: every
// <release_root>/<40-hex> dir that no verdict names is renamed (never
// deleted: the bytes stay for the operator) to
// <release_root>/.orphan-<sha>-<unix>, freeing the name for a clean re-fetch.
// It holds <release_root>/.fetch.lock (non-blocking: a fetch in flight makes
// it refuse — that fetch may be exactly between rename and link), and it
// refuses outright when any verdict cannot be read (it cannot then prove a
// dir is unreferenced).

// ErrFetchInFlight: another fetch holds the release root's lock.
var ErrFetchInFlight = errors.New("updaterworker: a fetch is in flight (the release root lock is held)")

const (
	fetchLockName  = ".fetch.lock"
	orphanPrefix   = ".orphan-"
	verdictsSubdir = "verdicts"
	maxVerdictRead = 64 << 10
)

// LockReleaseRoot takes <release_root>/.fetch.lock exclusively without
// waiting. The fetch holds it from before its rename until after its verdict
// link, so the recovery below never races a live fetch.
func LockReleaseRoot(releaseRoot string) (func(), error) {
	if releaseRoot == "" || !filepath.IsAbs(releaseRoot) {
		return nil, errors.New("updaterworker: the release root must be an absolute path")
	}
	f, err := os.OpenFile(filepath.Join(releaseRoot, fetchLockName), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrFetchInFlight
		}
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}

// QuarantineInterruptedFetches moves every release dir no verdict names out
// of the way (see above) and returns the new names.
func QuarantineInterruptedFetches(releaseRoot, dataDir string, now time.Time) ([]string, error) {
	unlock, err := LockReleaseRoot(releaseRoot)
	if err != nil {
		return nil, err
	}
	defer unlock()
	return quarantineLocked(releaseRoot, dataDir, now)
}

// quarantineLocked is the recovery for a caller that ALREADY holds the
// release-root lock — the fetch itself (at the U3 fold: FetchRelease takes
// the lock, runs this, and keeps the lock through its rename and verdict
// link, so no recovery can ever see its own in-between state).
func quarantineLocked(releaseRoot, dataDir string, now time.Time) ([]string, error) {
	named, err := verdictReleaseDirs(dataDir)
	if err != nil {
		return nil, err
	}
	ents, err := os.ReadDir(releaseRoot)
	if err != nil {
		return nil, err
	}
	var moved []string
	for _, e := range ents {
		name := e.Name()
		if !isSHA40(name) || e.Type()&os.ModeSymlink != 0 || !e.IsDir() {
			continue
		}
		dir := filepath.Join(releaseRoot, name)
		if named[dir] {
			continue
		}
		to := filepath.Join(releaseRoot, orphanPrefix+name+"-"+strconv.FormatInt(now.Unix(), 10))
		if err := os.Rename(dir, to); err != nil {
			return moved, fmt.Errorf("updaterworker: quarantine %s: %w", dir, err)
		}
		moved = append(moved, to)
	}
	if len(moved) > 0 {
		if d, err := os.Open(releaseRoot); err == nil {
			_ = d.Sync()
			d.Close()
		}
	}
	return moved, nil
}

// verdictReleaseDirs is the set of release_dir values every verdict names.
// Any verdict that cannot be read or parsed refuses the whole set.
func verdictReleaseDirs(dataDir string) (map[string]bool, error) {
	if dataDir == "" || !filepath.IsAbs(dataDir) {
		return nil, errors.New("updaterworker: the data dir must be an absolute path")
	}
	dir := filepath.Join(dataDir, updaterwire.UpdaterDirName, verdictsSubdir)
	ents, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	var names []string
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, n := range names {
		p := filepath.Join(dir, n)
		f, err := os.OpenFile(p, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
		if err != nil {
			return nil, fmt.Errorf("updaterworker: verdict %s: %w", n, err)
		}
		b, err := io.ReadAll(io.LimitReader(f, maxVerdictRead))
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("updaterworker: verdict %s: %w", n, err)
		}
		var v struct {
			ReleaseDir *string `json:"release_dir"`
		}
		if err := json.Unmarshal(b, &v); err != nil || v.ReleaseDir == nil || !filepath.IsAbs(*v.ReleaseDir) {
			return nil, fmt.Errorf("updaterworker: verdict %s names no absolute release_dir — refusing to judge any release dir unreferenced", n)
		}
		out[filepath.Clean(*v.ReleaseDir)] = true
	}
	return out, nil
}
