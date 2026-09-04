package trader

// CLASS 60 LINT — adding a time-dependent rule without a clock seam fails here.
//
// The list of rules lives in ONE place, ../clock-seams.list, and this test
// reads it. It does not carry its own copy of the rule names (A24: a fixture
// with its own copy of a constant is not a check, it is a second thing to keep
// in sync).
//
// For each row it asserts the two halves of the seam:
//   1. the …At variant exists in that file
//   2. the entry point is a ONE-LINE DELEGATE to it — nothing else in the body
//
// (2) is the half that matters. An entry point that reads the clock AND does
// work is exactly class 60: the test pins a fixture, the code reads the wall,
// and the suite is green until the hour changes.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type seamRule struct{ file, entry, at string }

func loadSeamRules(t *testing.T) []seamRule {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "clock-seams.list"))
	if err != nil {
		t.Fatalf("read clock-seams.list: %v", err)
	}
	var out []seamRule
	for _, ln := range strings.Split(string(raw), "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		p := strings.Split(ln, ":")
		if len(p) != 3 {
			t.Fatalf("malformed row %q — want <file>:<entry>:<At>", ln)
		}
		out = append(out, seamRule{p[0], p[1], p[2]})
	}
	if len(out) == 0 {
		t.Fatal("clock-seams.list has no rules — an empty list passes vacuously")
	}
	return out
}

// bodyOf returns the lines between the opening and closing brace of the first
// top-level func whose declaration contains "<name>(".
func bodyOf(src, name string) ([]string, bool) {
	lines := strings.Split(src, "\n")
	for i, l := range lines {
		if !strings.HasPrefix(l, "func ") || !strings.Contains(l, name+"(") {
			continue
		}
		var body []string
		for j := i + 1; j < len(lines); j++ {
			if lines[j] == "}" {
				return body, true
			}
			body = append(body, lines[j])
		}
	}
	return nil, false
}

// statements strips blank lines and comments — a delegate may be commented.
func statements(body []string) []string {
	var out []string
	for _, l := range body {
		s := strings.TrimSpace(l)
		if s == "" || strings.HasPrefix(s, "//") {
			continue
		}
		out = append(out, s)
	}
	return out
}

func TestEveryTimeDependentRuleHasAClockSeam(t *testing.T) {
	for _, r := range loadSeamRules(t) {
		raw, err := os.ReadFile(filepath.Join("..", r.file))
		if err != nil {
			t.Errorf("%s: %v", r.file, err)
			continue
		}
		src := string(raw)

		if _, ok := bodyOf(src, r.at); !ok {
			t.Errorf("%s: rule %q has no %q variant — the clock cannot be stated by a test",
				r.file, r.entry, r.at)
			continue
		}
		body, ok := bodyOf(src, r.entry)
		if !ok {
			t.Errorf("%s: entry point %q not found (renamed? update clock-seams.list)", r.file, r.entry)
			continue
		}
		st := statements(body)
		if len(st) != 1 {
			t.Errorf("%s: %q must be a ONE-LINE delegate to %q, has %d statements: %v",
				r.file, r.entry, r.at, len(st), st)
			continue
		}
		if !strings.Contains(st[0], r.at+"(") {
			t.Errorf("%s: %q does not delegate to %q — body is %q", r.file, r.entry, r.at, st[0])
		}
	}
}

// E2 — the detector must FAIL on an unseamed rule. Pinned against synthetic
// source rather than by breaking a real file, so it keeps working forever.
func TestSeamLintRejectsAnUnseamedRule(t *testing.T) {
	unseamed := `package x

func (at *AutoTrader) somethingTimed() bool {
	now := time.Now()
	return now.Hour() > 12
}
`
	if _, ok := bodyOf(unseamed, "somethingTimedAt"); ok {
		t.Fatal("detector found an At variant that does not exist")
	}
	body, ok := bodyOf(unseamed, "somethingTimed")
	if !ok {
		t.Fatal("detector failed to find the entry point at all")
	}
	if n := len(statements(body)); n == 1 {
		t.Fatal("detector called a 2-statement wall-clock body a one-line delegate")
	}

	seamed := `package x

func (at *AutoTrader) somethingTimed() bool {
	// the clock lives here and nowhere below
	return at.somethingTimedAt(time.Now())
}

func (at *AutoTrader) somethingTimedAt(now time.Time) bool {
	return now.Hour() > 12
}
`
	if _, ok := bodyOf(seamed, "somethingTimedAt"); !ok {
		t.Fatal("detector missed a real At variant")
	}
	body, _ = bodyOf(seamed, "somethingTimed")
	st := statements(body)
	if len(st) != 1 || !strings.Contains(st[0], "somethingTimedAt(") {
		t.Fatalf("detector rejected a correct seam: %v", st)
	}
}
