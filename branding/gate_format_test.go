package branding

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCryptoUnionGateCanonicalFormat drives scripts/crypto-union-gate.sh in a
// synthetic repo and pins the CTO-ruled canonical table format:
//   | path | line | token | DELETE|CUT|KEEP | OWNER | reason |
// plus the single-ownership invariant: a line claimed by TWO tables is a
// DOUBLE-CLAIM FAIL, a prose/range line is an unparseable FAIL, and a clean
// canonical union exits 0. (The production call site is the script itself.)
func TestCryptoUnionGateCanonicalFormat(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	tmp := t.TempDir()
	tblDir := t.TempDir() // tables live OUTSIDE the swept repo

	// minimal guard: the script extracts the literal from this exact line
	write(t, filepath.Join(tmp, "branding/no_crypto.go"),
		"package branding\n\nconst SweepRegexLiteral = `bybit`\n")

	// the real gate script
	scriptSrc, err := os.ReadFile(filepath.Join("..", "scripts", "crypto-union-gate.sh"))
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(tmp, "scripts", "crypto-union-gate.sh")
	write(t, script, string(scriptSrc))
	if err := os.Chmod(script, 0o755); err != nil {
		t.Fatal(err)
	}

	write(t, filepath.Join(tmp, "src/kept.go"), "package src\nvar K = \"bybit api\" // line 2, covered by a KEEP row\n")
	write(t, filepath.Join(tmp, "web/src/components/plan/ExecutorVerdict.tsx"), "arm.state === 'mixed'\n")
	write(t, filepath.Join(tmp, "web/src/components/trader/TraderConfigModal.tsx"), "// cut complete\n")
	// meet the 2,000-file enumeration floor
	for i := 0; i < 2000; i++ {
		write(t, filepath.Join(tmp, "fill", fmt.Sprintf("f%05d.txt", i)), "")
	}

	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmp
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	git("init", "-q")
	git("add", "-A")
	git("commit", "-qm", "init")
	head := strings.TrimSpace(git("rev-parse", "HEAD"))

	const guardLine = 3 // branding/no_crypto.go line holding the literal
	const keptLine = 2  // src/kept.go line holding "bybit api"

	tableA := fmt.Sprintf(`branch-point: %s
integrator-tip: %s
paths: .
regex: bybit
| branding/no_crypto.go | %d | bybit | KEEP | CR-B | guard's own exported literal |
| src/kept.go | %d | bybit | KEEP | CR-A | deliberate keep for the proof |
`, head, head, guardLine, keptLine)
	tableAPath := filepath.Join(tblDir, "tblA.md")
	write(t, tableAPath, tableA)

	runGate := func(args ...string) (string, int) {
		cmd := exec.Command("bash", append([]string{script}, args...)...)
		cmd.Dir = tmp
		out, err := cmd.CombinedOutput()
		rc := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				rc = ee.ExitCode()
			} else {
				t.Fatalf("gate run: %v\n%s", err, out)
			}
		}
		return string(out), rc
	}

	// clean canonical union -> exit 0
	out, rc := runGate(tableAPath)
	if rc != 0 {
		t.Fatalf("clean canonical union must exit 0, got %d:\n%s", rc, out)
	}
	for _, want := range []string{"sweep complete", "DOUBLE-CLAIM"} {
		if want == "DOUBLE-CLAIM" && strings.Contains(out, want) {
			t.Fatalf("clean union must not report DOUBLE-CLAIM:\n%s", out)
		}
	}

	// double-claim: a second table claims the SAME kept.go line as its own
	tableB := fmt.Sprintf(`branch-point: %s
integrator-tip: %s
paths: .
regex: bybit
| src/kept.go | %d | bybit | KEEP | CR-B | conflicting claim |
`, head, head, keptLine)
	tableBPath := filepath.Join(tblDir, "tblB.md")
	write(t, tableBPath, tableB)
	out, rc = runGate(tableAPath, tableBPath)
	if rc == 0 {
		t.Fatalf("double-claim union must FAIL, got exit 0:\n%s", out)
	}
	if !strings.Contains(out, "DOUBLE-CLAIM") {
		t.Fatalf("expected DOUBLE-CLAIM in output:\n%s", out)
	}

	// non-canonical line: prose ranges must be refused, not silently parsed
	tableC := fmt.Sprintf(`branch-point: %s
integrator-tip: %s
paths: .
regex: bybit
CUT src/kept.go [2] -- a prose range, not a canonical row
`, head, head)
	tableCPath := filepath.Join(tblDir, "tblC.md")
	write(t, tableCPath, tableC)
	out, rc = runGate(tableCPath)
	if rc == 0 {
		t.Fatalf("prose-range table must FAIL, got exit 0:\n%s", out)
	}
	if !strings.Contains(out, "unparseable") {
		t.Fatalf("expected 'unparseable' in output:\n%s", out)
	}

	// file-level exclusivity (Finding 3): the same PATH claimed by two tables
	// on DIFFERENT lines — per-line check passes, the file-level check must FAIL
	tableD := fmt.Sprintf(`branch-point: %s
integrator-tip: %s
paths: .
regex: bybit
| src/kept.go | %d | bybit | KEEP | CR-B | same path, different line, different owner |
`, head, head, keptLine+1)
	tableDPath := filepath.Join(tblDir, "tblD.md")
	write(t, tableDPath, tableD)
	out, rc = runGate(tableAPath, tableDPath)
	if rc == 0 {
		t.Fatalf("file-level double-claim must FAIL, got exit 0:\n%s", out)
	}
	if !strings.Contains(out, "FILE DOUBLE-CLAIM") || !strings.Contains(out, "src/kept.go") {
		t.Fatalf("expected FILE DOUBLE-CLAIM on src/kept.go:\n%s", out)
	}

	// glob exclusion (Finding 4): a dated-export file with a hit is excluded
	// by an explicit glob KEEP row carrying its reason — never a silent skip
	write(t, filepath.Join(tmp, "src/gen/report.tsv"), "bybit\thit\n")
	git("add", "-A")
	git("commit", "-qm", "gen-export")
	head = strings.TrimSpace(git("rev-parse", "HEAD"))
	tableE := fmt.Sprintf(`branch-point: %s
integrator-tip: %s
paths: .
regex: bybit
| branding/no_crypto.go | %d | bybit | KEEP | CR-B | guard's own exported literal |
| src/kept.go | %d | bybit | KEEP | CR-A | deliberate keep for the proof |
| src/gen/*.tsv | * | * | KEEP | CR-C | dated research export — historical record, not shipped code, KEEP byte-identical |
`, head, head, guardLine, keptLine)
	tableEPath := filepath.Join(tblDir, "tblE.md")
	write(t, tableEPath, tableE)
	out, rc = runGate(tableEPath)
	if rc != 0 {
		t.Fatalf("clean union with a glob exclusion must exit 0, got %d:\n%s", rc, out)
	}
	if !strings.Contains(out, "excluded 1 file(s)") {
		t.Fatalf("expected the glob exclusion to be REPORTED:\n%s", out)
	}

	// a glob row matching nothing is a stale/typo exclusion -> FAIL
	tableF := fmt.Sprintf(`branch-point: %s
integrator-tip: %s
paths: .
regex: bybit
| src/gen/*.tsv | * | * | KEEP | CR-C | dated research export |
| src/nope/*.tsv | * | * | KEEP | CR-C | typo glob, matches nothing |
`, head, head)
	tableFPath := filepath.Join(tblDir, "tblF.md")
	write(t, tableFPath, tableF)
	out, rc = runGate(tableFPath)
	if rc == 0 {
		t.Fatalf("a glob row matching nothing must FAIL, got exit 0:\n%s", out)
	}
	if !strings.Contains(out, "matches NO tracked file") {
		t.Fatalf("expected the stale-glob FAIL:\n%s", out)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
