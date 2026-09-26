// FIX-KNOBS B1 (DS-105, 2026-09-26) — ONE truth for session enablement.
// Today the per-session sessions[].enable wins while the stored
// sessions_enabled list still gates the fallback — so the live owner's rows
// (list ["NY"] + ASIA/LONDON enable=true) run all three sessions while the
// list says [NY] everywhere (planner, Studio, boot line). The fix: the
// per-session enable is THE truth; the stored list is parse-only and the
// EFFECTIVE list is DERIVED from the per-session enables (registry default
// when no override).
package trader

import (
	"testing"

	"nofx/kernel"
	"nofx/store"
)

func TestB1PerSessionEnableIsTheOnlyTruth(t *testing.T) {
	cfg := store.GetDefaultStrategyConfig("en")
	tv := true
	cfg.DayPlan = &store.DayPlanConfig{
		SessionsEnabled: []string{"NY"}, // the lying stored list
		Sessions:        []store.DayPlanSessionOverride{{Session: "ASIA", Enable: &tv}},
	}
	at, _ := resetTrader(t, cfg)

	// ASIA is explicitly ON for this strategy — the stored [NY] list must not
	// veto it.
	if !at.sessionEnabledForStrategy("ASIA") {
		t.Fatal("ASIA must be enabled: the per-session enable is the truth; the stored [NY] list must not veto it")
	}
	// NY: no override, registry-enabled → ON.
	if !at.sessionEnabledForStrategy("NY") {
		t.Fatal("NY must be enabled: registry default with no override")
	}
	// LONDON: no override, registry-disabled → OFF (the stored list may not
	// change either direction any more).
	if at.sessionEnabledForStrategy("LONDON") {
		t.Fatal("LONDON must be off: no override and the registry disables it")
	}

	// The derived effective list reports what actually runs.
	got := at.derivedSessionsEnabled()
	for _, want := range []string{"NY", "ASIA"} {
		if !containsString(got, want) {
			t.Fatalf("derived sessions_enabled %v must contain %s", got, want)
		}
	}
	if containsString(got, "LONDON") {
		t.Fatalf("derived sessions_enabled %v must not contain LONDON", got)
	}

	// And the sessionRunnable gate agrees with the derived list — ASIA is
	// runnable despite the stored list.
	reg := kernel.DefaultSessionRegistry()
	for _, s := range reg.Sessions {
		if s.Name != "ASIA" {
			continue
		}
		if runnable, why := at.sessionRunnable(&s); !runnable {
			t.Fatalf("ASIA must be runnable, got %q", why)
		}
	}
}

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
