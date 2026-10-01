package branding

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"vl/internal/censuswalk"
)

// TestSweepRegexPinsSentinels proves the exported literal does what plan v10
// line 108 says: BOTH-boundary aster/lighter kill master/disaster/faster/
// easter/highlighter; quant\b kills "quantity"; "mixed" is DOUBLE-quoted
// only (the single-quoted 'mixed' sites are ContentAssertSites, asserted by
// content in the gate, never by the regex).
func TestSweepRegexPinsSentinels(t *testing.T) {
	re, err := regexp.Compile("(?i)" + SweepRegexLiteral)
	if err != nil {
		t.Fatalf("SweepRegexLiteral does not compile: %v", err)
	}
	mustMatch := []string{
		"aster exchange", "**/aster*.go", "bg-vl-neo-bg-lighter",
		`case "mixed":`, `bracketOCO = "MIXED"`, "币安",
		"use_oi_top", "netflow ranking", "a quant run", "price ranking 5",
	}
	for _, s := range mustMatch {
		if !re.MatchString(s) {
			t.Errorf("sweep regex must match %q", s)
		}
	}
	mustNot := []string{
		"master branch", "disaster recovery", "faster code", "easter egg",
		"highlighter pen", "a quantity of 3",
		`arm.state === 'mixed'`, `source_type === 'mixed'`,
	}
	for _, s := range mustNot {
		if re.MatchString(s) {
			t.Errorf("sweep regex must NOT match %q", s)
		}
	}
}

// TestHostCensusLiteralByteExact pins the C8 host list byte-for-byte (plan
// v10 C8; GO item 2 declined — no alpaca/sina additions).
func TestHostCensusLiteralByteExact(t *testing.T) {
	want := `binance\.com|binance\.vision|bybit\.com|okx\.com|bitget\.com|kucoin\.com|gateio|gate\.io|indodax\.com|hyperliquid\.xyz|asterdex|lighter\.xyz|coinank|claw402\.ai|blockrun`
	if HostCensusLiteral != want {
		t.Fatalf("HostCensusLiteral drifted from the plan literal\n got: %s\nwant: %s", HostCensusLiteral, want)
	}
}

// TestSweepRegexMatchesThePlanLiteral re-derives the plan literal from the
// local plan checkout the same way DS-104's extraction does and asserts the
// guard's export is byte-identical — the guard cannot drift from the plan.
func TestSweepRegexMatchesThePlanLiteral(t *testing.T) {
	planPath := "/home/hoang/crypto-removal-plan/2026-09-30-crypto-removal-plan-v10.md"
	if _, err := os.Stat(planPath); err != nil {
		t.Skipf("plan file not on this box (%v) — literal pin not evaluable here", err)
	}
	b, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	var lit string
	for _, line := range strings.Split(string(b), "\n") {
		if strings.Contains(line, "THE assembled regex") && strings.Contains(line, "R6-102-P1-1") {
			start := strings.Index(line, ": `")
			if start < 0 {
				t.Fatal("plan literal line shape changed")
			}
			rest := line[start+3:]
			end := strings.Index(rest, "`")
			if end < 0 {
				t.Fatal("plan literal line shape changed (no closing backtick)")
			}
			lit = rest[:end]
		}
	}
	if lit == "" {
		t.Fatal("plan literal not found in the v10 line")
	}
	if SweepRegexLiteral != lit {
		t.Fatalf("SweepRegexLiteral != plan v10 literal\n guard: %s\n plan:  %s", SweepRegexLiteral, lit)
	}
}

func importTargetsCrypto(source []byte) (map[string]bool, error) {
	f, err := parser.ParseFile(token.NewFileSet(), "source.go", source, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	targets := map[string]bool{}
	for _, spec := range f.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, err
		}
		targets[path] = true
	}
	return targets, nil
}

// moduleRoot returns the module path the census sees (vl; the rename made
// it vl — never hardcode a second copy).
func moduleRoot(t *testing.T) string {
	if m, err := censuswalk.ModulePath(".."); err == nil {
		return m
	}
	return "vl"
}

// TestCensusEnumeratesFailsLoud is the C8 enumeration invariant: git
// ls-files -z, t.Fatal on git failure (never a skip), and a 2,000-file
// floor so an empty enumeration cannot pass.
func TestCensusEnumeratesFailsLoud(t *testing.T) {
	out, err := exec.Command("git", "-C", "..", "ls-files", "-z").Output()
	if err != nil {
		t.Fatalf("git ls-files failed (never skip): %v", err)
	}
	files := strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00")
	if len(files) < 2000 {
		t.Fatalf("census enumerated only %d tracked files (< 2000 floor) — it is going vacuous", len(files))
	}
}

// TestNoCryptoImportsSDKsHosts is the C8 zero-state census: no import of a
// deleted package prefix, no deleted SDK module in go.mod/go.sum, no
// crypto-venue host in non-test Go files or web/src.
//
// BY DESIGN this test is RED at every pre-cut head and goes GREEN only at
// the integrated head (plan v10 C8: "Build gate: the same grep returns 0 at
// the integrated head"). The parts cut the sites it names.
func TestNoCryptoImportsSDKsHosts(t *testing.T) {
	root := ".."
	out, err := exec.Command("git", "-C", root, "ls-files", "-z").Output()
	if err != nil {
		t.Fatalf("git ls-files failed (never skip): %v", err)
	}
	files := strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00")
	hostRe := regexp.MustCompile("(?i)" + HostCensusLiteral)
	var problems []string

	prefixHit := func(target string) bool {
		target = strings.TrimPrefix(target, moduleRoot(t)+"/")
		for _, p := range DeletedImportPrefixes {
			if target == p || strings.HasPrefix(target, p+"/") {
				return true
			}
		}
		return false
	}
	sdkHit := func(line string) bool {
		for _, m := range DeletedSDKModules {
			if strings.Contains(line, m) {
				return true
			}
		}
		return false
	}

	for _, f := range files {
		b, rerr := os.ReadFile(filepath.Join(root, f))
		if rerr != nil {
			continue
		}
		base := filepath.Base(f)
		if strings.HasSuffix(f, ".go") {
			if targets, perr := importTargetsCrypto(b); perr == nil {
				for tg := range targets {
					if prefixHit(tg) {
						problems = append(problems, f+": imports deleted package prefix "+tg)
					}
				}
			}
			// host census: non-test Go files only
			if !strings.HasSuffix(base, "_test.go") && hostRe.Match(b) {
				problems = append(problems, f+": crypto-venue host in a non-test Go file")
			}
		}
		if strings.HasPrefix(f, "web/src/") && hostRe.Match(b) {
			problems = append(problems, f+": crypto-venue host under web/src")
		}
		if f == "go.mod" || f == "go.sum" {
			for _, line := range strings.Split(string(b), "\n") {
				if sdkHit(line) {
					problems = append(problems, f+": deleted SDK module present ("+line+")")
				}
			}
		}
	}
	if len(problems) > 0 {
		for _, p := range problems[:12] {
			t.Errorf("%s", p)
		}
		t.Fatalf("%d crypto census violations (0 expected at the integrated head)", len(problems))
	}
}
