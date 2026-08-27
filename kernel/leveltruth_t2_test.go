package kernel

import (
	"testing"
)

// T2 golden — the EQL·15m (HTF) grade-A "mystery" reproduced from first
// principles: typeEvidence(EQL)=0.70 × freshMult(1.0) × confluence (≥1 family
// in band → ×1.2) × HTF (×1.2) = 1.008 → gradeFromScore → A. Without
// confluence: 0.70×1.2 = 0.84 → B. Every seated grade must be recomputable
// from the published tables — this test pins both branches.
func TestT2EQL15mHTFGradeAReproducible(t *testing.T) {
	mk := func(confluent bool) []ScoredLevel {
		levels := []DetectedLevel{
			{Kind: KindEQL, Price: 15550, Label: "EQL·15m (HTF)", HTF: true, TF: "15m"},
		}
		if confluent {
			// Different family within confBand (0.10 × dATR = 20pt).
			levels = append(levels, DetectedLevel{Kind: KindPDC, Price: 15560, Label: "PDC"})
		}
		fresh := func(DetectedLevel) string { return "" }
		return ScoreLevels(levels, 15550, 200, fresh, 8, 1.5)
	}
	if got := mk(true)[0].Grade; got != "A" {
		t.Fatalf("EQL·15m (HTF) with confluence grade = %s, want A (0.70×1.2×1.2=1.008)", got)
	}
	if got := mk(false)[0].Grade; got != "B" {
		t.Fatalf("EQL·15m (HTF) without confluence grade = %s, want B (0.70×1.2=0.84)", got)
	}
}

// T2 golden — no-trade plans are machine-authored: every level's grade IS the
// machine grade (the first unstamped population, 256/795 regression).
func TestT2NoTradeDocStampsAll(t *testing.T) {
	lv := []PlanLevel{
		{Price: 100, Label: "PDC", Grade: "B"},
		{Price: 101, Label: "ONH", Grade: "A"},
	}
	doc := NoTradePlanDocWithLevels("fail-closed", lv)
	if len(doc.Levels) != 2 {
		t.Fatalf("levels = %d", len(doc.Levels))
	}
	for i, l := range doc.Levels {
		if l.MachineGrade == "" || l.MachineGrade != l.Grade {
			t.Fatalf("no-trade level %d unstamped: %+v", i, l)
		}
	}
}

// T2 golden — StampMachineGrades covers the FULL pool (prices that lost the
// seat race must still get stamped).
func TestT2StampMachineGradesFromMap(t *testing.T) {
	doc := &PlanDoc{Levels: []PlanLevel{
		{Price: 29499.75, Grade: "B"},
		{Price: 123.456, Grade: "A"}, // rounded to 2dp in the map
	}}
	n := StampMachineGrades(doc, map[float64]string{29499.75: "B", 123.46: "A"})
	if n != 2 {
		t.Fatalf("stamped %d, want 2", n)
	}
	for _, l := range doc.Levels {
		if l.MachineGrade != l.Grade {
			t.Fatalf("level %v machine_grade=%q", l.Price, l.MachineGrade)
		}
	}
}
