package agent

import (
	"strings"
	"time"

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
	price := data.CurrentPrice
	q.Price = &price
	// The changes are NOT data.PriceChange1h/4h: on the futures route those are
	// bar-count windows over 5m/1h bars (20 bars = 100 min; the prior 1h close),
	// and 0 when the bars run short. Here each window is measured by bar time
	// on the NT8 5m series, and is null when the series does not reach back.
	bars := market.FuturesBarsProvider(sym, "5m", nt8QuoteBars)
	q.Change1hPct = changeOverWindow(bars, time.Hour)
	q.Change4hPct = changeOverWindow(bars, 4*time.Hour)
	return q
}

// nt8QuoteBars is how many 5m bars the change windows read: 200 bars reach
// back 16h40m, well past the 4h window.
const nt8QuoteBars = 200

// changeOverWindow is the percent change from the close of the latest bar that
// closed at least `window` before the series' last bar, to that last close.
// It returns nil — absent, never a fabricated 0 — when no bar reaches back
// that far or the reference close is not positive.
func changeOverWindow(bars []market.Kline, window time.Duration) *float64 {
	if len(bars) == 0 {
		return nil
	}
	last := bars[len(bars)-1]
	target := last.CloseTime - window.Milliseconds()
	for i := len(bars) - 2; i >= 0; i-- {
		if bars[i].CloseTime > target {
			continue
		}
		if bars[i].Close <= 0 {
			return nil
		}
		pct := (last.Close - bars[i].Close) / bars[i].Close * 100
		return &pct
	}
	return nil
}
