package kernel

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"
)

// TestChecklistNoUnnumberedClassHeadings pins the checklist's own numbering
// law: every class heading carries a number, and the header's "Highest
// occupied class" equals the largest heading number. A heading that still
// reads "(assigned at merge)" or "(pending)" fails the pin
// (docs/checklist-numbers, PASS 1, DS-102 2026-09-24).
func TestChecklistNoUnnumberedClassHeadings(t *testing.T) {
	mb, err := os.ReadFile("../docs/superpowers/AUDIT-CHECKLIST.md")
	if err != nil {
		t.Fatalf("cannot read AUDIT-CHECKLIST.md: %v", err)
	}
	if off := checklistNumberingOffences(string(mb)); len(off) > 0 {
		t.Fatal(off[0])
	}
}

// checklistNumberingOffences returns why src breaks the numbering law ("" =
// none): an unnumbered class heading, or a header that does not equal the
// largest heading number.
func checklistNumberingOffences(src string) []string {
	// ALLOW pattern: a class heading is `## CLASS <digits> —` and nothing
	// else. Every other line opening "## CLASS " is unnumbered, whatever the
	// placeholder spells ("NN (assigned at merge)", "M3-07", "??", ...).
	var m []string
	numbered := regexp.MustCompile(`^## CLASS \d+ — `)
	for _, line := range regexp.MustCompile(`(?m)^## CLASS .*$`).FindAllString(src, -1) {
		if !numbered.MatchString(line) {
			m = append(m, line)
		}
	}
	m = append(m, regexp.MustCompile(`(?m)^### \(pending\) `).FindAllString(src, -1)...)
	if len(m) > 0 {
		shown := m
		if len(shown) > 5 {
			shown = shown[:5]
		}
		return []string{fmt.Sprintf("%d unnumbered class heading(s) remain: %v", len(m), shown)}
	}
	max := 0
	for _, m := range regexp.MustCompile(`(?m)^## CLASS (\d+) —`).FindAllStringSubmatch(src, -1) {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return []string{fmt.Sprintf("bad class number %q: %v", m[1], err)}
		}
		if n > max {
			max = n
		}
	}
	header := regexp.MustCompile(`Highest occupied class: \*\*(\d+)\*\*`).FindStringSubmatch(src)
	if header == nil {
		return []string{"no 'Highest occupied class' header found"}
	}
	hn, err := strconv.Atoi(header[1])
	if err != nil {
		return []string{fmt.Sprintf("bad header number %q: %v", header[1], err)}
	}
	if hn != max {
		return []string{fmt.Sprintf("header says highest occupied class is %d but the max heading number is %d", hn, max)}
	}
	return nil
}

// The unnumbered-heading rule is an ALLOW pattern, not a list of the
// placeholders someone has thought of: a deny-list that recognises only
// "NN (assigned at merge)" passed a stray "## CLASS M3-07 —" silently (M3
// class drafts, CTO 1790245281578 — the class-168 shape inside the guard).
func TestChecklistNumberingRefusesEveryNonNumberedClassHeading(t *testing.T) {
	const header = "*Highest occupied class: **7** (2026-09-24).\n\n## CLASS 7 — fixture\n\n"
	for _, h := range []string{
		"## CLASS NN (assigned at merge) — x",
		"## CLASS M3-07 — x",
		"## CLASS NN — x",
		"## CLASS 12a — x",
		"## CLASS  8 — x",
		"## CLASS ?? — x",
	} {
		if off := checklistNumberingOffences(header + h + "\n"); len(off) == 0 {
			t.Errorf("heading %q is not numbered and must be refused", h)
		}
	}
	if off := checklistNumberingOffences(header); len(off) != 0 {
		t.Errorf("control: a numbered checklist must pass: %v", off)
	}
	if off := checklistNumberingOffences(header + "### (pending) x\n"); len(off) == 0 {
		t.Errorf("a '### (pending)' heading must still be refused")
	}
}
