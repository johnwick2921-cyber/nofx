package branding

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ── W-NB Part B — Binance is gone from the code, entirely ────────────────────
//
// The owner ruled "remove binance totally". No tracked code file — Go,
// TypeScript, or the agent's embedded skills JSON — may carry the name or its
// Chinese spellings, except the files
// below, each of which must spell it to do its job. docs/ is history (reports,
// plans, the checklist) and stays as written. An allowlist row that matches
// nothing fails: the list only shrinks.
var noBinanceAllowlist = map[string]string{
	"branding/no_binance_guard_test.go":      "this guard — it must spell what it forbids",
	"market/no_binance_source_guard_test.go": "the market source guard — its host pattern spells the name",
	"trader/no_binance_futures_test.go":      "the outbound-host trap — it fails on any host containing the name",
	"manager/removed_exchange_test.go":       "migration truth — a legacy row must name the removed exchange to prove it is refused",
}

// No bare "fapi." needle: an R3-protected broker (Aster) legitimately calls
// its own fapi.asterdex.com host; every Binance host contains the name.
var noBinanceNeedle = regexp.MustCompile(`(?i)binance|币安|必安`)

func TestNoBinanceAnywhereInCode(t *testing.T) {
	out, err := gitIn("..", "ls-files", "--", "*.go", "*.ts", "*.tsx", "agent/skills/*.json")
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(files) < 500 {
		t.Fatalf("the guard sees only %d tracked code files — it is going vacuous", len(files))
	}
	hit := map[string]bool{}
	for _, f := range files {
		if f == "" || strings.HasPrefix(f, "docs/") {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join("..", f))
		if rerr != nil {
			continue // deleted in the working tree, not yet in the index
		}
		for i, line := range strings.Split(string(b), "\n") {
			if !noBinanceNeedle.MatchString(line) {
				continue
			}
			hit[f] = true
			if _, ok := noBinanceAllowlist[f]; !ok {
				t.Errorf("%s:%d carries the removed exchange: %s", f, i+1, strings.TrimSpace(line))
			}
		}
	}
	for f := range noBinanceAllowlist {
		if !hit[f] {
			t.Errorf("allowlist row %q matches nothing — delete it (the list only shrinks)", f)
		}
	}
}

// gitIn runs git in dir (the repo root, from this package: "..").
func gitIn(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Output()
}
