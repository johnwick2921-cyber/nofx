package kernel

import "testing"

// gateArgs mirrors validateDecision's risk params for a typical $50k SIM
// account (btcEthLev=10, altLev=5, btcEthRatio=5, altRatio=1).
func gateArgs() (equity float64, btcEthLev, altLev int, btcEthRatio, altRatio float64) {
	return 50000, 10, 5, 5.0, 1.0
}

// TestFuturesGate_AcceptsRealisticMNQOpen proves the futures notional
// exemption: a 1-contract-ish MNQ open (~$60k notional) now PASSES the gate,
// where the crypto equity×ratio cap ($50k) previously rejected it.
func TestFuturesGate_AcceptsRealisticMNQOpen(t *testing.T) {
	eq, btcEthLev, altLev, btcEthRatio, altRatio := gateArgs()
	d := &Decision{
		Symbol:          "MNQ",
		Action:          "open_long",
		Leverage:        1,
		PositionSizeUSD: 60000, // ~1 MNQ contract notional (> the old $50k cap)
		StopLoss:        21480.00,
		TakeProfit:      21560.00, // SL<TP, 0.2-entry placement => R/R 4:1
	}
	if err := validateDecision(d, eq, btcEthLev, altLev, btcEthRatio, altRatio); err != nil {
		t.Fatalf("expected MNQ $60k open to PASS the futures gate, got: %v", err)
	}
}

// TestFuturesGate_RejectsAbsurdMNQNotional confirms the cap is REAL, not
// accept-everything: a notional above equity×futuresMaxNotionalLeverage
// ($50k×20 = $1M) is rejected.
func TestFuturesGate_RejectsAbsurdMNQNotional(t *testing.T) {
	eq, btcEthLev, altLev, btcEthRatio, altRatio := gateArgs()
	d := &Decision{
		Symbol:          "MNQ",
		Action:          "open_long",
		Leverage:        1,
		PositionSizeUSD: 2_000_000, // > $1M ceiling
		StopLoss:        21480.00,
		TakeProfit:      21560.00,
	}
	if err := validateDecision(d, eq, btcEthLev, altLev, btcEthRatio, altRatio); err == nil {
		t.Fatal("expected absurd $2M MNQ notional to be REJECTED, but it passed")
	}
}

// TestFuturesGate_CryptoCapUnchanged is the regression guard: a crypto open
// above equity×ratio ($50k) is STILL rejected (the futures branch must not
// loosen crypto).
func TestFuturesGate_CryptoCapUnchanged(t *testing.T) {
	eq, btcEthLev, altLev, btcEthRatio, altRatio := gateArgs()
	d := &Decision{
		Symbol:          "SOLUSDT", // altcoin, ratio 1x => $50k cap
		Action:          "open_long",
		Leverage:        5,
		PositionSizeUSD: 60000, // > $50k crypto cap
		StopLoss:        100.0,
		TakeProfit:      130.0,
	}
	if err := validateDecision(d, eq, btcEthLev, altLev, btcEthRatio, altRatio); err == nil {
		t.Fatal("expected crypto SOLUSDT $60k open to STILL be rejected by the $50k cap")
	}
}

// TestFuturesGate_WaitAlwaysValid — a wait decision (the common futures
// output) validates regardless of symbol.
func TestFuturesGate_WaitAlwaysValid(t *testing.T) {
	eq, btcEthLev, altLev, btcEthRatio, altRatio := gateArgs()
	d := &Decision{Symbol: "MNQ", Action: "wait"}
	if err := validateDecision(d, eq, btcEthLev, altLev, btcEthRatio, altRatio); err != nil {
		t.Fatalf("wait should always validate, got: %v", err)
	}
}
