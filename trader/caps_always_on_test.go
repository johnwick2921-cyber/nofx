package trader

import (
	"testing"

	"nofx/kernel"
	"nofx/store"
)

// 6.4 (ruling B) + FIX-KNOBS A (2026-09-26) — the always-on clamps. The dead
// toggles (max_contracts_enabled / notional_cap_enabled) are REMOVED, so the
// venue-safety clamps are now always-on BY CONSTRUCTION. Resolvers still honour
// the live VALUE knobs (max_contracts_per_order / max_notional_leverage) and the
// Stage-A ceiling — asserted with REAL values (review M-A): the Stage-A cap is
// raised to 3 via env, so an explicit 1 and the default 3 are DISTINCT and an
// explicit 5 is clamped.
func TestSizeCapsAlwaysOnAfterToggleRemoval(t *testing.T) {
	t.Setenv("STAGE_A_CONTRACT_CAP", "3")
	rc := store.RiskControlConfig{}
	// no toggles exist to store — the clamps resolve regardless. With the
	// REAL Stage-A ceiling 3, the venue default 2 is now UNCLAMPED (under
	// the shipped cap of 1 it would clamp to 1 — the env raise is visible).
	if got := kernel.ResolveMaxContracts(rc.MaxContractsPerOrder, 2); got != 2 {
		t.Errorf("contracts clamp = %d, want 2 (venue default, unclamped under the Stage-A ceiling 3)", got)
	}
	if got := kernel.ResolveNotionalLeverage(rc.MaxNotionalLeverage, 20); got != 20 {
		t.Errorf("notional cap = %.0f, want 20 (always-on)", got)
	}
	// An explicit 1 is REAL: distinct from the default 3.
	rc.MaxContractsPerOrder = 1
	rc.MaxNotionalLeverage = 10
	if got := kernel.ResolveMaxContracts(rc.MaxContractsPerOrder, 2); got != 1 {
		t.Errorf("explicit contracts value ignored (got %d)", got)
	}
	if got := kernel.ResolveNotionalLeverage(rc.MaxNotionalLeverage, 20); got != 10 {
		t.Errorf("explicit notional value ignored (got %.0f)", got)
	}
	// An explicit 5 is clamped to the Stage-A ceiling 3.
	rc.MaxContractsPerOrder = 5
	if got := kernel.ResolveMaxContracts(rc.MaxContractsPerOrder, 2); got != 3 {
		t.Errorf("explicit 5 must clamp to the Stage-A ceiling 3 (got %d)", got)
	}
}
