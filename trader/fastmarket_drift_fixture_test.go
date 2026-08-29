package trader

import (
	"math"
	"testing"

	"nofx/kernel"
	"nofx/market"
)

// TestFastMarketDriftBeyondATRFires is the P2/C3 drift>1.5×ATR fixture for the
// F3 fast-market wake path (trader/auto_trader_planner.go:846 fastMarketDrift
// at :1104). It feeds 100 synthetic 1m bars (constant 10-pt range → Wilder
// ATR14(5m) == 10.0 exactly) through the REAL fastMarketDrift and asserts the
// wake fires with the correct (driftPts, driftAtr) pair, plus the sub-threshold
// and unset-price fail-closed cases.
func TestFastMarketDriftBeyondATRFires(t *testing.T) {
	prev := market.FuturesBarsProvider
	defer func() { market.FuturesBarsProvider = prev }()

	base := int64(1_783_000_000_000) // arbitrary epoch-ms origin
	const n = 100
	market.FuturesBarsProvider = func(symbol, tf string, count int) []market.Kline {
		if symbol != "MNQ" || tf != "1m" {
			t.Fatalf("unexpected provider args %s/%s (want MNQ/1m)", symbol, tf)
		}
		if count != kernel.AISVPBarCount {
			t.Fatalf("provider count %d (want %d)", count, kernel.AISVPBarCount)
		}
		bars := make([]market.Kline, 0, n)
		for i := 0; i < n; i++ {
			bars = append(bars, market.Kline{
				OpenTime: base + int64(i)*60_000,
				Open:     30000.0,
				High:     30005.0, // 10-pt range every bar
				Low:      29995.0,
				Close:     30000.0,
				Volume:   1.0,
			})
		}
		return bars
	}

	at := &AutoTrader{}
	at.lastPlanWritePrice.Store(math.Float64bits(30000.0))

	// ATR5m == 10.0 → threshold = fastMarketATR() × 10. Drift 45 pts fires.
	d, a := at.fastMarketDrift(30045.0)
	wantThr := fastMarketATR() * 10.0
	if d != 45.0 || a != 4.5 {
		t.Fatalf("fast-market wake: drift=%.2f atr=%.2f (want 45.00 / 4.50, threshold %.2f)", d, a, wantThr)
	}
	if d <= wantThr {
		t.Fatalf("drift %.2f not above threshold %.2f", d, wantThr)
	}

	// Sub-threshold: 12 pts ≤ 1.5×10 → (0,0), no wake.
	if d2, a2 := at.fastMarketDrift(30012.0); d2 != 0 || a2 != 0 {
		t.Fatalf("sub-threshold drift must not fire: got (%.2f, %.2f)", d2, a2)
	}

	// Unset write price → fail-closed (0,0).
	at.lastPlanWritePrice.Store(0)
	if d3, a3 := at.fastMarketDrift(30090.0); d3 != 0 || a3 != 0 {
		t.Fatalf("unset lastPlanWritePrice must not fire: got (%.2f, %.2f)", d3, a3)
	}
}
