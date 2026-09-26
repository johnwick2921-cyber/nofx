package trader

import "testing"

// TestFastMarketWireDefaults — F3 knobs resolve to the shipped defaults.
// FIX-KNOBS A2 (2026-09-26): REASONING=MAX everywhere — the shipped default
// is max, not fast(low).
func TestFastMarketWireDefaults(t *testing.T) {
	if fastMarketATR() != 1.5 {
		t.Fatalf("FAST_MARKET_ATR default want 1.5, got %.2f", fastMarketATR())
	}
	m, e := fastMarketReasoningWire()
	if m != "enabled" || e != "max" {
		t.Fatalf("FAST_MARKET_REASONING default want enabled/max (A2), got %s/%s", m, e)
	}
	if fastMarketReasoningLabel() != "max" {
		t.Fatalf("label want max, got %s", fastMarketReasoningLabel())
	}
}
