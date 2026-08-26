package store

import (
	"encoding/json"
	"testing"
)

// P0-relax (2026-08-27) — MinSideLevelsFor resolution seam: session override →
// strategy value → 0 (unset → caller falls back to MIN_SIDE_LEVELS env →
// kernel.DefaultSideQuota).
func TestDayPlanMinSideLevelsFor(t *testing.T) {
	base := func(v int) *DayPlanConfig {
		c := DefaultDayPlanConfig()
		c.MinSideLevels = intPtr(v)
		return c
	}
	if got := DefaultDayPlanConfig().MinSideLevelsFor("NY"); got != 2 {
		t.Fatalf("shipped default = %d, want 2", got)
	}
	if got := base(3).MinSideLevelsFor("NY"); got != 3 {
		t.Fatalf("strategy 3 = %d, want 3 (old hard rule reachable)", got)
	}
	c := base(2)
	ov3 := 3
	c.Sessions = []DayPlanSessionOverride{{Session: "NY", MinSideLevels: &ov3}}
	if got := c.MinSideLevelsFor("NY"); got != 3 {
		t.Fatalf("session override 3 = %d, want 3", got)
	}
	if got := c.MinSideLevelsFor("LONDON"); got != 2 {
		t.Fatalf("non-overridden session = %d, want 2", got)
	}
	ovZero := 0
	c.Sessions[0].MinSideLevels = &ovZero
	if got := c.MinSideLevelsFor("NY"); got != 0 {
		t.Fatalf("override 0 = %d, want 0 (unset → env fallback)", got)
	}
	var nilCfg *DayPlanConfig
	if got := nilCfg.MinSideLevelsFor("NY"); got != 0 {
		t.Fatalf("nil config = %d, want 0", got)
	}
}

// P0-relax (2026-08-27) — JSON round-trip: the knob + its session override
// survive the config PUT paths (both the full strategy-config save and the
// day-plan update go through this JSON envelope).
func TestDayPlanMinSideLevelsJSONRoundTrip(t *testing.T) {
	c := DefaultDayPlanConfig()
	c.MinSideLevels = intPtr(3)
	ov4 := 4
	c.Sessions = []DayPlanSessionOverride{{Session: "NY", MinSideLevels: &ov4}}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back DayPlanConfig
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := back.MinSideLevelsFor("NY"); got != 4 {
		t.Fatalf("override round-trip = %d, want 4", got)
	}
	if got := back.MinSideLevelsFor("LONDON"); got != 3 {
		t.Fatalf("base round-trip = %d, want 3", got)
	}
}
