package branding

// no_crypto_test.go — the crypto-removal census guard (plan v10 FINAL, C8 +
// C13). The literal lists below are the SINGLE home of the sweep and host
// regexes: the disposition tables and the union gate EXPORT them from this
// file, never re-typing them (C13 invariant 5, DS-102 audit8 P2-1).
//
// The sweep regex is the BASE literal for GO item 1 (crypto only): under GO
// item 2 the CTO broadcasts the appended |alpaca|twelvedata|sina form and the
// tables regenerate. This file is a *_test.go, so the C13 sweep (non-test
// lines only) never scans its own literals.

import "testing"

const (
	// CryptoSweepRegex is THE assembled sweep regex (plan v10 FINAL, C13).
	// Case-insensitive, Go regexp (\b is not POSIX ERE).
	CryptoSweepRegex = `binance|bybit|okx|bitget|kucoin|gate\.io|gateio|indodax|hyperliquid|\baster\b|asterdex|\blighter\b|coinank|usdc|usdt|x402|claw402|blockrun|wallet|ai500|hyper_all|hyper_main|oi_top|oi_low|netflow|quant\b|price ranking|"mixed"|币安|欧易|火币|U本位|永续`
	// CryptoHostRegex is the C8 host-census literal (plan v10 FINAL, C8).
	// Dots escaped; grep -E with EXACTLY this list.
	CryptoHostRegex = `binance\.com|binance\.vision|bybit\.com|okx\.com|bitget\.com|kucoin\.com|gateio|gate\.io|indodax\.com|hyperliquid\.xyz|asterdex|lighter\.xyz|coinank|claw402\.ai|blockrun`
)

// TestCryptoGuardLiteralsArePinned pins both literals byte-for-byte so a drift
// in either list cannot pass silently.
func TestCryptoGuardLiteralsArePinned(t *testing.T) {
	const wantSweep = `binance|bybit|okx|bitget|kucoin|gate\.io|gateio|indodax|hyperliquid|\baster\b|asterdex|\blighter\b|coinank|usdc|usdt|x402|claw402|blockrun|wallet|ai500|hyper_all|hyper_main|oi_top|oi_low|netflow|quant\b|price ranking|"mixed"|币安|欧易|火币|U本位|永续`
	const wantHosts = `binance\.com|binance\.vision|bybit\.com|okx\.com|bitget\.com|kucoin\.com|gateio|gate\.io|indodax\.com|hyperliquid\.xyz|asterdex|lighter\.xyz|coinank|claw402\.ai|blockrun`
	if CryptoSweepRegex != wantSweep {
		t.Fatal("CryptoSweepRegex drifted from the v10 FINAL literal")
	}
	if CryptoHostRegex != wantHosts {
		t.Fatal("CryptoHostRegex drifted from the v10 FINAL literal")
	}
}
