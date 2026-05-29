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

// futuresPointValues maps a CME root to the USD value of a 1.00-point move
// for ONE contract (the contract multiplier). Used to size futures positions
// in contracts: contracts = notional / (price × pointValue).
var futuresPointValues = map[string]float64{
	"NQ":  20.0,   // E-mini Nasdaq-100   ($20/pt)
	"MNQ": 2.0,    // Micro E-mini Nasdaq ($2/pt)
	"ES":  50.0,   // E-mini S&P 500      ($50/pt)
	"MES": 5.0,    // Micro E-mini S&P    ($5/pt)
	"RTY": 50.0,   // E-mini Russell 2000 ($50/pt)
	"M2K": 5.0,    // Micro E-mini Russell($5/pt)
	"YM":  5.0,    // E-mini Dow          ($5/pt)
	"MYM": 0.5,    // Micro E-mini Dow    ($0.50/pt)
	"CL":  1000.0, // Crude oil           ($1000 per $1)
	"GC":  100.0,  // Gold                ($100 per $1)
}

// FuturesPointValue returns the USD value of a 1.00-point move for one
// contract of the given CME futures symbol (e.g. MNQ=2, NQ=20). Accepts any
// symbol form (continuous "NQ.c.0", contract "NQM6", bare "MNQ", qualified
// "MNQ 06-26"). Returns 0 for non-futures / unknown roots — callers must
// treat 0 as "unknown" and NOT divide by it.
func FuturesPointValue(symbol string) float64 {
	if root := futuresRoot(symbol); root != "" {
		return futuresPointValues[root]
	}
	return 0
}

// futuresRoot extracts the CME root from any symbol form. Longest-root-first
// so "MNQ" wins over "NQ". Returns "" if no known root matches.
func futuresRoot(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	best := ""
	for root := range cmeFuturesRoots {
		matched := s == root ||
			strings.HasPrefix(s, root+".") ||
			strings.HasPrefix(s, root+" ")
		if !matched && strings.HasPrefix(s, root) && len(s) == len(root)+2 {
			tail := s[len(root):]
			matched = isQuarterlyMonth(tail[0]) && tail[1] >= '0' && tail[1] <= '9'
		}
		if matched && len(root) > len(best) {
			best = root
		}
	}
	return best
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
