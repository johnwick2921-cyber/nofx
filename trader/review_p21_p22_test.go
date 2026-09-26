// FIX-KNOBS review P2-1/P2-2 (DS-101 review of #259) — pins for the runtime
// registry source and the 12-USDT admission floor.
package trader

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/store"
)

// P2-1 — derivedSessionsEnabled must read the RUNTIME admin registry (the same
// source the planner schedule uses: sessionRegistry → system_config), not the
// compile-time default. A modified admin registry must change the derived list.
func TestP21DerivedSessionsReadTheRuntimeRegistry(t *testing.T) {
	cfg := store.GetDefaultStrategyConfig("en")
	at, st := resetTrader(t, cfg)

	reg := kernel.DefaultSessionRegistry()
	for i := range reg.Sessions {
		reg.Sessions[i].Enabled = true // the admin turned everything on
	}
	blob, err := json.Marshal(reg)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetSystemConfig(kernel.SessionRegistryConfigKey, string(blob)); err != nil {
		t.Fatalf("store the admin registry: %v", err)
	}

	now := time.Date(2026, 8, 18, 14, 0, 0, 0, time.UTC)
	got := at.derivedSessionsEnabledAt(now)
	for _, want := range []string{"NY", "ASIA", "LONDON"} {
		if !containsString(got, want) {
			t.Fatalf("derived sessions %v must contain %s — the admin registry (all enabled) is the runtime source", got, want)
		}
	}
}

// P2-2 — the replacement 12-USDT floor (for the removed min_position_size) must
// refuse below 12 at the production admission gate (enforceMinPositionSize).
// Mutation: remove the floor → this test FAILS (a 5-USDT order would pass).
func TestMinPositionFloorRefusesBelowTwelve(t *testing.T) {
	cfg := store.GetDefaultStrategyConfig("en")
	at, _ := resetTrader(t, cfg)

	if minPositionSizeFloorUSDT != 12.0 {
		t.Fatalf("the floor constant must stay 12.0, got %v", minPositionSizeFloorUSDT)
	}
	if err := at.enforceMinPositionSize(5); err == nil {
		t.Fatal("5 USDT must be refused at the admission gate")
	} else if !strings.Contains(err.Error(), "12") {
		t.Fatalf("the refusal must name the floor, got: %v", err)
	}
	if err := at.enforceMinPositionSize(12); err != nil {
		t.Fatalf("exactly the floor must pass, got: %v", err)
	}
	if err := at.enforceMinPositionSize(60); err != nil {
		t.Fatalf("above the floor must pass, got: %v", err)
	}
}
