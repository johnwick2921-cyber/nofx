package kernel

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"nofx/market"
	"nofx/store"
	"nofx/telemetry"
)

// ── Crypto open interest / funding are ABSENT (no source in this build) ────
//
// The exchange-hosted OI and funding fetchers were removed with their broker.
// A crypto read now carries OpenInterest nil and FundingRateKnown false, so:
//   - the OI liquidity filter REFUSES a crypto candidate it cannot judge
//     (named log line + gate counter "oi_unavailable") instead of letting it
//     through unjudged (fail-closed, never fail-open);
//   - an open position is still fetched and rendered — OI and funding n/a;
//   - the grid prompt prints funding n/a, never a fabricated 0.0000%.
// Driven through the production call site fetchMarketDataWithStrategy, with
// every outbound request answered offline by canned CoinAnk klines.

// coinankCanned answers every request with a CoinAnk kline page and records
// the host and the venue each request asked for.
type coinankCanned struct {
	mu     sync.Mutex
	hosts  []string
	venues []string
}

func (c *coinankCanned) RoundTrip(r *http.Request) (*http.Response, error) {
	c.mu.Lock()
	c.hosts = append(c.hosts, r.URL.Host)
	c.venues = append(c.venues, r.URL.Query().Get("exchange"))
	c.mu.Unlock()
	rows := make([][]float64, 300)
	for i := range rows {
		start := float64(int64(i) * 300000)
		px := 60000 + float64(i%9)*13.5
		rows[i] = []float64{start, start + 299999, px - 5, px, px + 20, px - 20, 12, 12 * px, 40}
	}
	body, _ := json.Marshal(map[string]any{"success": true, "code": "1", "data": rows})
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(bytes.NewReader(body)), Request: r}, nil
}

func installCoinankCanned(t *testing.T) *coinankCanned {
	t.Helper()
	c := &coinankCanned{}
	prev := http.DefaultTransport
	http.DefaultTransport = c
	t.Cleanup(func() { http.DefaultTransport = prev })
	return c
}

func cryptoEngine(venue string) *StrategyEngine {
	cfg := &store.StrategyConfig{Language: "en"}
	cfg.CoinSource.StaticCoins = []string{"BTCUSDT"}
	cfg.Indicators.EnableOI = true
	cfg.Indicators.EnableFundingRate = true
	cfg.Indicators.Klines.PrimaryTimeframe = "5m"
	cfg.Indicators.Klines.SelectedTimeframes = []string{"5m"}
	cfg.Indicators.Klines.PrimaryCount = 60
	e := &StrategyEngine{config: cfg}
	e.SetVenue(venue)
	return e
}

func oiUnavailableCount(trader string) int {
	_, table := telemetry.GateBlockSnapshot()
	return table[trader]["oi_unavailable"]
}

func TestCryptoCandidateWithAbsentOIIsRefusedByTheLiquidityFilter(t *testing.T) {
	canned := installCoinankCanned(t)
	e := cryptoEngine("bybit")
	const trader = "t-oi-absent-candidate"
	before := oiUnavailableCount(trader)

	ctx := &Context{TraderID: trader, CandidateCoins: []CandidateCoin{{Symbol: "BTCUSDT"}}}
	if err := fetchMarketDataWithStrategy(ctx, e); err != nil {
		t.Fatalf("fetchMarketDataWithStrategy: %v", err)
	}
	if d, ok := ctx.MarketDataMap["BTCUSDT"]; ok {
		t.Fatalf("a crypto candidate with ABSENT open interest must be refused by the liquidity filter, got %+v", d)
	}
	if got := oiUnavailableCount(trader) - before; got != 1 {
		t.Fatalf("the refusal must be counted once under gate oi_unavailable, got %d", got)
	}

	canned.mu.Lock()
	defer canned.mu.Unlock()
	if len(canned.hosts) == 0 {
		t.Fatal("vacuous: the candidate's klines were never requested")
	}
	for i, h := range canned.hosts {
		if h != "api.coinank.com" || canned.venues[i] != "Bybit" {
			t.Fatalf("request %d went to host %q venue %q — want only CoinAnk on the trader's own venue (Bybit)", i, h, canned.venues[i])
		}
	}
}

func TestCryptoPositionIsStillReadWithOIAndFundingNA(t *testing.T) {
	installCoinankCanned(t)
	e := cryptoEngine("bybit")
	const trader = "t-oi-absent-position"
	before := oiUnavailableCount(trader)

	ctx := &Context{TraderID: trader, Positions: []PositionInfo{{Symbol: "BTCUSDT", Side: "long"}},
		CandidateCoins: []CandidateCoin{{Symbol: "BTCUSDT"}}}
	if err := fetchMarketDataWithStrategy(ctx, e); err != nil {
		t.Fatalf("fetchMarketDataWithStrategy: %v", err)
	}
	d := ctx.MarketDataMap["BTCUSDT"]
	if d == nil {
		t.Fatal("an open position must still be read — the liquidity filter judges new candidates only")
	}
	if d.OpenInterest != nil || d.FundingRateKnown {
		t.Fatalf("crypto OI/funding must be ABSENT (no source in this build), got OI %+v funding known %v", d.OpenInterest, d.FundingRateKnown)
	}
	if got := oiUnavailableCount(trader) - before; got != 0 {
		t.Fatalf("a held position is not a refused candidate, counter moved by %d", got)
	}
	md := e.formatMarketData(d)
	if !strings.Contains(md, "Open Interest: n/a") || !strings.Contains(md, "Funding Rate: n/a") {
		t.Fatalf("absent OI and funding render n/a:\n%s", md)
	}
	if strings.Contains(md, "Latest: 0.00") || strings.Contains(md, "0.00e+00") {
		t.Fatalf("absent OI/funding must never render a fabricated 0:\n%s", md)
	}
}

func TestCryptoReadOnAVenueWithoutCoinAnkBookMakesNoRequest(t *testing.T) {
	canned := installCoinankCanned(t)
	for _, venue := range []string{"kucoin", "lighter", "indodax", ""} {
		e := cryptoEngine(venue)
		ctx := &Context{TraderID: "t-no-venue", Positions: []PositionInfo{{Symbol: "BTCUSDT", Side: "long"}}}
		_ = fetchMarketDataWithStrategy(ctx, e)
		if d, ok := ctx.MarketDataMap["BTCUSDT"]; ok {
			t.Fatalf("venue %q has no CoinAnk book: the read must be refused, got %+v", venue, d)
		}
	}
	canned.mu.Lock()
	defer canned.mu.Unlock()
	if len(canned.hosts) != 0 {
		t.Fatalf("a venue with no CoinAnk book must be refused BEFORE any request (no other venue's book), got %v %v", canned.hosts, canned.venues)
	}
	if _, err := market.GetWithTimeframesVenue("BTCUSDT", "kucoin", []string{"5m"}, "5m", 10); err == nil || !strings.Contains(err.Error(), `"kucoin"`) {
		t.Fatalf("the refusal must name the venue, got %v", err)
	}
}

// The grid prompt reads its funding from the same market data: absent funding
// is n/a in both languages, never a fabricated 0.0000%.
func TestGridPromptRendersAbsentFundingAsNA(t *testing.T) {
	installCoinankCanned(t)
	d, err := market.GetWithTimeframesVenue("BTCUSDT", "bybit", []string{"5m", "4h"}, "5m", 50)
	if err != nil {
		t.Fatalf("fixture: crypto read on bybit: %v", err)
	}
	cfg := &store.GridStrategyConfig{Symbol: "BTCUSDT", GridCount: 10, TotalInvestment: 1000, Leverage: 2}
	gctx := BuildGridContextFromMarketData(d, cfg)
	if gctx.FundingRateKnown {
		t.Fatal("crypto funding is absent in this build; FundingRateKnown must be false")
	}
	for _, lang := range []string{"zh", "en"} {
		p := BuildGridUserPrompt(gctx, lang)
		if strings.Contains(p, "0.0000%") {
			t.Fatalf("%s grid prompt printed a fabricated funding 0.0000%%:\n%s", lang, p)
		}
		if !strings.Contains(p, "Funding Rate: n/a") && !strings.Contains(p, "资金费率: n/a") {
			t.Fatalf("%s grid prompt must say funding n/a:\n%s", lang, p)
		}
	}
}
