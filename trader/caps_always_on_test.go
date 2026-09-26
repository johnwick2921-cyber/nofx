package trader

import (
	"testing"

	"nofx/kernel"
	"nofx/store"
)

// 6.4 (ruling B) + FIX-KNOBS A (2026-09-26) — the always-on clamps. The dead
// toggles (max_contracts_enabled / notional_cap_enabled) are REMOVED, so the
// venue-safety clamps are now always-on BY CONSTRUCTION: no stored value can
// disable them, because no field carries a disable. Resolvers still honour the
// live VALUE knobs (max_contracts_per_order / max_notional_leverage) and the
// Stage-A ceiling.
func TestSizeCapsAlwaysOnAfterToggleRemoval(t *testing.T) {
        rc := store.RiskControlConfig{}
        // no toggles exist to store — the clamps resolve regardless.
        if got := kernel.ResolveMaxContracts(rc.MaxContractsPerOrder, 2); got != 1 {
                t.Errorf("contracts clamp = %d, want 1 (always-on, Stage-A ceiling)", got)
        }
        if got := kernel.ResolveNotionalLeverage(rc.MaxNotionalLeverage, 20); got != 20 {
                t.Errorf("notional cap = %.0f, want 20 (always-on)", got)
        }
        // Explicit per-strategy values still win (the VALUE is live).
	rc.MaxContractsPerOrder = 1
	rc.MaxNotionalLeverage = 10
	if got := kernel.ResolveMaxContracts(rc.MaxContractsPerOrder, 2); got != 1 {
		t.Errorf("explicit contracts value ignored (got %d)", got)
	}
	if got := kernel.ResolveNotionalLeverage(rc.MaxNotionalLeverage, 20); got != 10 {
		t.Errorf("explicit notional value ignored (got %.0f)", got)
	}
}
