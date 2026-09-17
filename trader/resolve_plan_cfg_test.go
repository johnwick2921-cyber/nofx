package trader

import (
	"reflect"
	"testing"

	"nofx/kernel"
	"nofx/store"
)

// Config-truth (settings fix) — the planner assembler HONORS the day_plan config:
// max_levels, per-session min_grade, planner_timeframes. A nil/default config
// reproduces the prior behavior (8 / no-filter / D,4h,1h,15m).

func TestResolveSessionPlanCfgDefaults(t *testing.T) {
	maxLevels, htfSeats, minGrade, tfs := resolveSessionPlanCfg(nil, "NY")
	if maxLevels != kernel.DefaultMaxLevels {
		t.Fatalf("nil config → default max_levels %d, got %d", kernel.DefaultMaxLevels, maxLevels)
	}
	if htfSeats != kernel.LegacyHtfSeats {
		t.Fatalf("nil config → legacy htf_seats %d, got %d", kernel.LegacyHtfSeats, htfSeats)
	}
	if minGrade != "" {
		t.Fatalf("nil config → no min_grade, got %q", minGrade)
	}
	if !reflect.DeepEqual(tfs, []string{"D", "4h", "1h", "15m"}) {
		t.Fatalf("nil config → default timeframes, got %v", tfs)
	}
}

func TestResolveSessionPlanCfgHonorsConfig(t *testing.T) {
	ny := "A"
	dp := &store.DayPlanConfig{
		MaxLevels:         6,
		PlannerTimeframes: []string{"D", "1h"},
		Sessions: []store.DayPlanSessionOverride{
			{Session: "NY", MinGrade: &ny},
			{Session: "ASIA"}, // no min_grade override
		},
	}
	maxLevels, htfSeats, minGrade, tfs := resolveSessionPlanCfg(dp, "NY")
	if maxLevels != 6 {
		t.Fatalf("config max_levels 6 not honored: got %d", maxLevels)
	}
	if htfSeats != 2 {
		t.Fatalf("config htf_seats unset → legacy 2, got %d", htfSeats)
	}
	if minGrade != "A" {
		t.Fatalf("NY session min_grade A not honored: got %q", minGrade)
	}
	if !reflect.DeepEqual(tfs, []string{"D", "1h"}) {
		t.Fatalf("config timeframes not honored: got %v", tfs)
	}

	// S3 — a legal 0 must stay 0 (no HTF seating), not fall back to the default.
	zero := 0
	dp.HtfSeats = &zero
	if ml, hs, _, _ := resolveSessionPlanCfg(dp, "NY"); hs != 0 {
		t.Fatalf("htf_seats=0 must resolve 0, got %d", hs)
	} else if ml != 6 {
		t.Fatalf("maxLevels drifted: %d", ml)
	}
	// S3 — clamp above the range.
	nine := 9
	dp.HtfSeats = &nine
	if _, hs, _, _ := resolveSessionPlanCfg(dp, "NY"); hs != 6 {
		t.Fatalf("htf_seats=9 must clamp to 6, got %d", hs)
	}

	// a session with no override inherits (empty min_grade).
	_, _, asiaGrade, _ := resolveSessionPlanCfg(dp, "ASIA")
	if asiaGrade != "" {
		t.Fatalf("ASIA has no min_grade override → inherit (empty), got %q", asiaGrade)
	}
}
