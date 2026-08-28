// deep-verify-22 kernel fixtures (package kernel — unexported API access).
package kernel

import (
	"strings"
	"testing"
)

// G1 1.1 — the GLOBAL market-entry R:R floor (validateDecision, config-driven)
// must refuse a 2.5-R entry at minRiskReward=3.0 and pass 3.2-R. Twin long/short.
func TestDeepG1MarketEntryRRFloorIsolation(t *testing.T) {
	mk := func(action string, rr float64) Decision {
		d := Decision{Symbol: "MNQ", Action: action, Leverage: 1, Price: 100, PositionSizeUSD: 60000}
		if action == "open_long" {
			d.StopLoss = 100 - 5
			d.TakeProfit = 100 + 5*rr
		} else {
			d.StopLoss = 100 + 5
			d.TakeProfit = 100 - 5*rr
		}
		return d
	}
	ctx := &Context{}
	for _, action := range []string{"open_long", "open_short"} {
		d25 := mk(action, 2.5)
		if err := validateDecision(&d25, 100000, 1, 0, 0, 0, 3.0, 50, 0, ctx); err == nil || !strings.Contains(err.Error(), "too low") {
			t.Fatalf("%s R=2.5 at floor 3.0 must be refused, got %v", action, err)
		}
		d32 := mk(action, 3.2)
		if err := validateDecision(&d32, 100000, 1, 0, 0, 0, 3.0, 50, 0, ctx); err != nil {
			t.Fatalf("%s R=3.2 must pass, got %v", action, err)
		}
	}
}
