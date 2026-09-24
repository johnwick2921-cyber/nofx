package kernel

import (
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
	src := string(mb)
	unnumbered := regexp.MustCompile(`(?m)^## CLASS NN \(assigned at merge\)|^### \(pending\) `)
	if m := unnumbered.FindAllString(src, -1); len(m) > 0 {
		shown := m
		if len(shown) > 5 {
			shown = shown[:5]
		}
		t.Fatalf("%d unnumbered class heading(s) remain: %v", len(m), shown)
	}
	max := 0
	for _, m := range regexp.MustCompile(`(?m)^## CLASS (\d+) —`).FindAllStringSubmatch(src, -1) {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			t.Fatalf("bad class number %q: %v", m[1], err)
		}
		if n > max {
			max = n
		}
	}
	header := regexp.MustCompile(`Highest occupied class: \*\*(\d+)\*\*`).FindStringSubmatch(src)
	if header == nil {
		t.Fatal("no 'Highest occupied class' header found")
	}
	hn, err := strconv.Atoi(header[1])
	if err != nil {
		t.Fatalf("bad header number %q: %v", header[1], err)
	}
	if hn != max {
		t.Fatalf("header says highest occupied class is %d but the max heading number is %d", hn, max)
	}
}
