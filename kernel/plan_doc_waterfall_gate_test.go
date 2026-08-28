package kernel

import (
	"strings"
	"testing"
)

// PRE-REOPEN F1 (2026-08-28) — GATE-LEVEL test for the 8th/9th conditions.
// The weekend audit found scenarioConds had 7 entries while the prompt
// mandates breakdown_continue/breakup_continue: ValidatePlanDocWithCaps
// hard-rejected the waterfall play at parse BEFORE the waterfall validator
// could run. The unit tests called the validator directly and missed it.
// This test drives the FULL schema gate end-to-end.

func waterfallDocJSON(cond, direction string) string {
	return `{
		"reasoning": "waterfall continuation after the gap-and-go",
		"bias": {"direction": "short", "conviction": "high"},
		"death_condition": "2x5m close back above 29657.39 cancels the plan",
		"levels": [
			{"price": 29657.39, "label": "VWAP", "grade": "B"}
		],
		"scenarios": [
			{
				"id": "S1",
				"trigger": "price closes below 29657.39 with displacement and no reclaim",
				"condition": "` + cond + `",
				"direction": "` + direction + `",
				"quality": "B",
				"target_chain": [29580.00],
				"invalid": "2x5m close back above 29657.39 cancels the short"
			}
		]
	}`
}

func TestPlanDocSchemaGateAcceptsWaterfallConditions(t *testing.T) {
	for _, tc := range []struct {
		cond, dir string
	}{
		{"breakdown_continue", "short"},
		{"breakup_continue", "long"},
	} {
		doc, err := ParsePlanDocCapped(waterfallDocJSON(tc.cond, tc.dir), 12, 5)
		if err != nil {
			t.Fatalf("%s: schema gate rejected the waterfall play: %v", tc.cond, err)
		}
		if len(doc.Scenarios) != 1 || doc.Scenarios[0].Condition != tc.cond {
			t.Fatalf("%s: parsed scenario wrong: %+v", tc.cond, doc.Scenarios)
		}
		// The direct caps validator must also accept it (the path a re-validated
		// doc takes on read).
		if err := ValidatePlanDocWithCaps(doc, 12, 5); err != nil {
			t.Fatalf("%s: caps re-validation rejected: %v", tc.cond, err)
		}
	}
}

func TestPlanDocSchemaGateStillRejectsUnknownCondition(t *testing.T) {
	doc, err := ParsePlanDocCapped(waterfallDocJSON("ninth_condition", "short"), 12, 5)
	if err == nil {
		t.Fatalf("an unknown condition must be rejected, got doc %+v", doc)
	}
	if !strings.Contains(err.Error(), "condition") {
		t.Fatalf("rejection must name the condition, got: %v", err)
	}
}
