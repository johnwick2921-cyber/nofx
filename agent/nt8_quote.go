package agent

import (
	"strings"

	"nofx/market"
)

// nt8QuoteSource names the ONE market-data source the chat assistant reads:
// NinjaTrader 8 bars (the NT8 BarCache behind market.FuturesBarsProvider).
const nt8QuoteSource = "nt8"

// nt8QuoteNonCMEReason is the named refusal for any symbol that is not a CME
// futures symbol. There is no crypto, stock or other external fallback — the
// chat's market reads are NinjaTrader-only.
const nt8QuoteNonCMEReason = "only CME futures symbols are available (NinjaTrader market data)"

// nt8Quote is one symbol's NinjaTrader read, shared by GET /api/agent/tickers
// and the get_market_snapshot chat tool. It is the contract the web ticker
// reads:
//
//	{"symbol": <as requested>, "price": number|null, "change_1h_pct": number|null,
//	 "change_4h_pct": number|null, "source": "nt8", "unavailable": "<reason>"}
//
// A value that was not read is null — never a fabricated 0 — and "unavailable"
// is present only when the read did not happen or failed, carrying the reason.
type nt8Quote struct {
	Symbol      string   `json:"symbol"`
	Price       *float64 `json:"price"`
	Change1hPct *float64 `json:"change_1h_pct"`
	Change4hPct *float64 `json:"change_4h_pct"`
	Source      string   `json:"source"`
	Unavailable string   `json:"unavailable,omitempty"`
}

// readNT8Quote reads one symbol. A CME futures symbol goes through
// market.GetWithExchange(sym, "ninjatrader") — the NT8 bar path, which makes no
// external market-data call; its error is passed through verbatim as the
// unavailable reason. Anything else is refused by name without any read.
func readNT8Quote(requested string) nt8Quote {
	q := nt8Quote{Symbol: requested, Source: nt8QuoteSource}
	sym := strings.TrimSpace(requested)
	if sym == "" || !market.IsCMEFuturesSymbol(market.Normalize(sym)) {
		q.Unavailable = nt8QuoteNonCMEReason
		return q
	}
	data, err := market.GetWithExchange(sym, "ninjatrader")
	if err != nil {
		q.Unavailable = err.Error()
		return q
	}
	if data == nil {
		q.Unavailable = sym + ": the NinjaTrader market read returned no data"
		return q
	}
	price, change1h, change4h := data.CurrentPrice, data.PriceChange1h, data.PriceChange4h
	q.Price = &price
	q.Change1hPct = &change1h
	q.Change4hPct = &change4h
	return q
}
