// FIX-KNOBS B2 (DS-105, 2026-09-26) — FAST_MARKET_REASONING conflict: the
// owner's .env carries 4 duplicate keys (all max) and dotenv is last-wins, so
// the effective value is invisible. Fix: the boot line prints the effective
// value + its source, and duplicate .env keys are detected with a WARN.
package trader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFastMarketReasoningBootSource(t *testing.T) {
	t.Setenv("FAST_MARKET_REASONING", "max")
	_, _, label, src := fastMarketReasoningWireWithSource()
	if label != "max" || !strings.Contains(src, "FAST_MARKET_REASONING") {
		t.Fatalf("label=%q src=%q — the boot line must name the env source", label, src)
	}

	t.Setenv("FAST_MARKET_REASONING", "")
	_, _, label, src = fastMarketReasoningWireWithSource()
	if label != "fast→low" || !strings.Contains(src, "code default") {
		t.Fatalf("label=%q src=%q — unset must read the code default and say so", label, src)
	}
}

func TestEnvDupKeyWarning(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	os.WriteFile(p, []byte("FAST_MARKET_REASONING=max\nOTHER=1\nFAST_MARKET_REASONING=low\nFAST_MARKET_REASONING=max\n"), 0o600)

	warn, n, last := envDupKeyWarning(p, "FAST_MARKET_REASONING")
	if n != 3 || last != "max" {
		t.Fatalf("n=%d last=%q — want 3 duplicates, last-wins max", n, last)
	}
	if !strings.Contains(warn, "3") || !strings.Contains(warn, "FAST_MARKET_REASONING") {
		t.Fatalf("warn must name the count and the key: %q", warn)
	}

	// No duplicates → no warning.
	os.WriteFile(p, []byte("FAST_MARKET_REASONING=max\nOTHER=1\n"), 0o600)
	if warn, n, _ := envDupKeyWarning(p, "FAST_MARKET_REASONING"); warn != "" || n != 1 {
		t.Fatalf("single key must not warn, got warn=%q n=%d", warn, n)
	}
}
