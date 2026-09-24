package updaterworker

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ── the boot proof Watch does not give (C13 as ruled) ──────────────────────
//
// activation.Watch is GREEN on any log token that prefixes the sha — and the
// REFUSED boot line carries the same "rev <sha12>" token as the OK one
// (kernel/boot_integrity.go Line(): "🔐 BOOT INTEGRITY %s — rev %s…"; main.go
// prints the REFUSED form at ERROR and the process keeps serving /api/health
// "ok"). So a boot that REFUSED trading passes Watch. boot_verified therefore
// reads the log itself, from the offset recorded BEFORE the kill:
//
//   - the literal "BOOT INTEGRITY OK — rev <sha12> ·" must be there (the " ·"
//     after the rev excludes the " +dirty" form, which is never a release);
//   - NO "BOOT INTEGRITY REFUSED" line may follow the offset, for any rev
//     (stricter than "for it": a refused boot after our kill is ours to own).

const (
	bootOKPrefix  = "BOOT INTEGRITY OK — rev "
	bootRefused   = "BOOT INTEGRITY REFUSED"
	maxBootScan   = 64 << 20
	bootLineShort = 12 // kernel/boot_integrity.go prints rev[:12]
)

// verifyBootLine scans path from off. ev gets the READ facts.
func verifyBootLine(path string, off int64, sha string, ev map[string]string) error {
	if len(sha) < bootLineShort || !isSHA40(sha) {
		return fmt.Errorf("boot check: %q is not a release sha", sha)
	}
	if off < 0 {
		return errors.New("boot check: negative log offset")
	}
	want := bootOKPrefix + sha[:bootLineShort] + " ·"
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("boot check: the boot log: %w", err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("boot check: %w", err)
	}
	if fi.Size() < off {
		return fmt.Errorf("boot check: %s is %d bytes, shorter than the offset %d recorded before the kill (truncated or replaced)", path, fi.Size(), off)
	}
	if _, err := f.Seek(off, io.SeekStart); err != nil {
		return fmt.Errorf("boot check: %w", err)
	}
	ev["log_offset"] = strconv.FormatInt(off, 10)
	sc := bufio.NewScanner(io.LimitReader(f, maxBootScan))
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	ok := 0
	for sc.Scan() {
		ln := sc.Text()
		if strings.Contains(ln, bootRefused) {
			return fmt.Errorf("boot check: a REFUSED boot line follows the kill: %q", clipText(strings.TrimSpace(ln)))
		}
		if strings.Contains(ln, want) {
			ok++
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("boot check: reading %s: %w", path, err)
	}
	if ok == 0 {
		return fmt.Errorf("boot check: no %q line in %s after offset %d", strings.TrimSuffix(want, " ·"), path, off)
	}
	ev["boot_line"] = "OK rev " + sha[:bootLineShort]
	return nil
}
