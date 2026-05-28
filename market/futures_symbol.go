// Task 12 / Cluster D — CME futures symbol detection.
//
// Single source of truth for "is this a CME futures symbol?" used by
// market.Normalize (case-preservation bypass) and market.GetWithTimeframes
// (route to Databento instead of CoinAnk). store/strategy.go has a small
// duplicate of this helper because store deliberately does not import market;
// keep the two in sync.

package market

import "strings"

// cmeFuturesRoots is the deny-list of crypto-collision-safe CME root symbols.
// Extend as new CME products are needed; keep conservative so this never
// matches a crypto ticker (BTC, ETH, SOL, etc.).
var cmeFuturesRoots = map[string]struct{}{
	"NQ":  {}, // E-mini Nasdaq-100
	"MNQ": {}, // Micro E-mini Nasdaq-100
	"ES":  {}, // E-mini S&P 500
	"MES": {}, // Micro E-mini S&P 500
	"RTY": {}, // E-mini Russell 2000
	"M2K": {}, // Micro E-mini Russell 2000
	"YM":  {}, // E-mini Dow
	"MYM": {}, // Micro E-mini Dow
	"CL":  {}, // Crude oil
	"GC":  {}, // Gold
}

// IsCMEFuturesSymbol reports whether a symbol is a CME futures symbol that
// must bypass crypto normalization (no ToUpper, no USDT append) and route to
// Databento.
//
// Matches:
//   - continuous form `<ROOT>.c.<N>` (e.g. NQ.c.0, MNQ.c.0) — case-sensitive
//     on the lowercase `.c.` segment per Databento's continuous symbology
//     convention.
//   - known CME roots, optionally followed by a contract suffix
//     (NQ, NQM6, MNQ, MNQU6, etc.) — uppercased for matching.
//
// Never matches crypto tickers (BTC, ETH, SOL, …) because the root set is
// conservative.
func IsCMEFuturesSymbol(symbol string) bool {
	s := strings.TrimSpace(symbol)
	if s == "" {
		return false
	}
	// Continuous form: any symbol containing ".c." (lowercase c) is a
	// Databento continuous symbol.
	if strings.Contains(s, ".c.") {
		return true
	}
	// Known root, optionally followed by ".something" or "<month><year>".
	upper := strings.ToUpper(s)
	if _, ok := cmeFuturesRoots[upper]; ok {
		return true
	}
	for root := range cmeFuturesRoots {
		if strings.HasPrefix(upper, root+".") {
			return true
		}
		// Contract code form: <ROOT><month-letter HMUZ><year-digit>
		// (NQM6, MNQU6, etc.). Require exact root prefix + 2 more chars
		// where the first is a CME quarterly month code and the second
		// is a single digit.
		if strings.HasPrefix(upper, root) && len(upper) == len(root)+2 {
			tail := upper[len(root):]
			if isQuarterlyMonth(tail[0]) && tail[1] >= '0' && tail[1] <= '9' {
				return true
			}
		}
	}
	return false
}

// isQuarterlyMonth reports whether b is one of the CME quarterly month
// codes (H=Mar, M=Jun, U=Sep, Z=Dec). Index futures roll quarterly so
// only these four appear in NQ/ES/RTY/YM contract codes.
func isQuarterlyMonth(b byte) bool {
	switch b {
	case 'H', 'M', 'U', 'Z':
		return true
	}
	return false
}
