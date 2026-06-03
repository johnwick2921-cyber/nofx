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
	if err := validateDecision(d, eq, btcEthLev, altLev, btcEthRatio, altRatio, 0); err != nil {
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
	if err := validateDecision(d, eq, btcEthLev, altLev, btcEthRatio, altRatio, 0); err == nil {
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
	if err := validateDecision(d, eq, btcEthLev, altLev, btcEthRatio, altRatio, 0); err == nil {
		t.Fatal("expected crypto SOLUSDT $60k open to STILL be rejected by the $50k cap")
	}
}

// TestFuturesGate_WaitAlwaysValid — a wait decision (the common futures
// output) validates regardless of symbol.
func TestFuturesGate_WaitAlwaysValid(t *testing.T) {
	eq, btcEthLev, altLev, btcEthRatio, altRatio := gateArgs()
	d := &Decision{Symbol: "MNQ", Action: "wait"}
	if err := validateDecision(d, eq, btcEthLev, altLev, btcEthRatio, altRatio, 0); err != nil {
		t.Fatalf("wait should always validate, got: %v", err)
	}
}

// TestRRGate_SyntheticEntryIsAlways4to1 — premise refutation for BE-2.
// validateDecision derives R/R from a SYNTHETIC entry placed 20% from SL toward
// TP, so the measured R/R is a constant 4.0:1 for ANY structurally-valid open
// (long OR short, regardless of how wide SL/TP are). Therefore the historical
// hardcoded 3.0:1 floor NEVER rejects a valid decision — the "silently rejects
// 1.5-R/R futures trades" concern in the change-map report is refuted by this
// entry model. A tight 0.5%-wide bracket (which an operator would read as ~1:1)
// still passes the 3.0 floor, proving the gate's R/R is not the AI's real R/R.
func TestRRGate_SyntheticEntryIsAlways4to1(t *testing.T) {
	eq, btcEthLev, altLev, btcEthRatio, altRatio := gateArgs()
	// MNQ long, SL 21500 / TP 21600 (a narrow bracket); synthetic entry = 21520,
	// risk 20 / reward 80 => measured R/R 4.0, passes the 3.0 floor.
	long := &Decision{Symbol: "MNQ", Action: "open_long", Leverage: 1, PositionSizeUSD: 60000, StopLoss: 21500, TakeProfit: 21600}
	if err := validateDecision(long, eq, btcEthLev, altLev, btcEthRatio, altRatio, 0); err != nil {
		t.Fatalf("synthetic-entry R/R is always 4.0; a valid MNQ long must pass the 3.0 floor, got: %v", err)
	}
	short := &Decision{Symbol: "MNQ", Action: "open_short", Leverage: 1, PositionSizeUSD: 60000, StopLoss: 21600, TakeProfit: 21500}
	if err := validateDecision(short, eq, btcEthLev, altLev, btcEthRatio, altRatio, 0); err != nil {
		t.Fatalf("synthetic-entry R/R is always 4.0; a valid MNQ short must pass the 3.0 floor, got: %v", err)
	}
}

// TestRRGate_FuturesThreadsConfig_CryptoUnchanged — BE-2 hygiene: the futures
// branch now reads the strategy's configured Min R/R (the arg) instead of the
// hardcoded 3.0, so the editor's field is authoritative for futures; crypto
// still uses 3.0 byte-identical (the futures arg is ignored on the crypto path).
// Behaviorally inert today (see TestRRGate_SyntheticEntryIsAlways4to1) but
// correct: it future-proofs the gate if the entry model ever uses the real entry.
func TestRRGate_FuturesThreadsConfig_CryptoUnchanged(t *testing.T) {
	eq, btcEthLev, altLev, btcEthRatio, altRatio := gateArgs()
	mnq := &Decision{Symbol: "MNQ", Action: "open_long", Leverage: 1, PositionSizeUSD: 60000, StopLoss: 21480, TakeProfit: 21560}
	if err := validateDecision(mnq, eq, btcEthLev, altLev, btcEthRatio, altRatio, 1.5); err != nil {
		t.Fatalf("futures open at configured 1.5 R/R floor should PASS, got: %v", err)
	}
	// Crypto ignores the futures R/R arg and keeps the 3.0 floor; a within-cap
	// BTC/ETH open (ratio 5x * 50k = 250k cap) still passes unchanged.
	btc := &Decision{Symbol: "BTCUSDT", Action: "open_long", Leverage: 5, PositionSizeUSD: 200000, StopLoss: 60000, TakeProfit: 66000}
	if err := validateDecision(btc, eq, btcEthLev, altLev, btcEthRatio, altRatio, 1.5); err != nil {
		t.Fatalf("crypto open must be byte-identical regardless of the futures R/R arg, got: %v", err)
	}
}
