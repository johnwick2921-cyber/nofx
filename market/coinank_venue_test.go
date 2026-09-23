package market

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"nofx/provider/coinank/coinank_enum"
)

// ── One CoinAnk venue per exchange; no default venue, no cross-venue read ───
//
// Every crypto kline read (GetWithExchange, GetWithTimeframesVenue,
// GetBoxData) resolves the trader's exchange through CoinAnkVenue. An exchange
// with no CoinAnk book of its own is REFUSED by name before any request; a
// failed or empty answer on the right venue is an error, never a second
// request to another venue. Crypto open interest and funding are ABSENT (no
// source in this build). Driven through the production read functions with
// http.DefaultTransport (what the CoinAnk client uses) replaced.

type venueTrap struct {
	mu      sync.Mutex
	fail    bool
	hosts   []string
	venues  []string
	symbols []string
}

func (v *venueTrap) RoundTrip(r *http.Request) (*http.Response, error) {
	v.mu.Lock()
	v.hosts = append(v.hosts, r.URL.Host)
	v.venues = append(v.venues, r.URL.Query().Get("exchange"))
	v.symbols = append(v.symbols, r.URL.Query().Get("symbol"))
	fail := v.fail
	v.mu.Unlock()
	if fail {
		return nil, errors.New("offline: this venue does not answer")
	}
	rows := make([][]float64, 300)
	for i := range rows {
		start := float64(int64(i) * 180000)
		px := 60000 + float64(i%11)*9.25
		rows[i] = []float64{start, start + 179999, px - 3, px, px + 15, px - 15, 7, 7 * px, 30}
	}
	body, _ := json.Marshal(map[string]any{"success": true, "code": "1", "data": rows})
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(bytes.NewReader(body)), Request: r}, nil
}

func installVenueTrap(t *testing.T, fail bool) *venueTrap {
	t.Helper()
	v := &venueTrap{fail: fail}
	prev := http.DefaultTransport
	http.DefaultTransport = v
	t.Cleanup(func() { http.DefaultTransport = prev })
	return v
}

func TestCoinAnkVenueMapsEachExchangeToItsOwnBookOnly(t *testing.T) {
	want := map[string]coinank_enum.Exchange{
		"bybit": coinank_enum.Bybit, "okx": coinank_enum.Okex, "bitget": coinank_enum.Bitget,
		"gate": coinank_enum.Gate, "hyperliquid": coinank_enum.Hyperliquid, "aster": coinank_enum.Aster,
		" Bybit ": coinank_enum.Bybit,
	}
	for ex, venue := range want {
		got, err := CoinAnkVenue(ex)
		if err != nil || got != venue {
			t.Errorf("CoinAnkVenue(%q) = %q, %v — want %q", ex, got, err, venue)
		}
	}
	for _, ex := range []string{"kucoin", "lighter", "indodax", "ninjatrader", "", "frobnicator"} {
		got, err := CoinAnkVenue(ex)
		if err == nil {
			t.Errorf("CoinAnkVenue(%q) = %q — an exchange with no CoinAnk book must be refused", ex, got)
			continue
		}
		if !strings.Contains(err.Error(), `"`+ex+`"`) {
			t.Errorf("the refusal must name the exchange %q: %v", ex, err)
		}
	}
	if CoinAnkAPISymbol("BTCUSDT", coinank_enum.Okex) != "BTC-USDT-SWAP" || CoinAnkAPISymbol("BTCUSDT", coinank_enum.Bybit) != "BTCUSDT" {
		t.Error("OKX lists perpetuals as BASE-USDT-SWAP; every other venue as BASEUSDT")
	}
}

func TestCryptoReadOnAVenueWithNoBookIsRefusedWithZeroRequests(t *testing.T) {
	trap := installVenueTrap(t, false)
	for _, ex := range []string{"kucoin", "lighter", "indodax", ""} {
		if _, err := GetWithExchange("BTCUSDT", ex); err == nil || !strings.Contains(err.Error(), "no CoinAnk venue") {
			t.Errorf("GetWithExchange(BTCUSDT, %q) must be refused by name, got %v", ex, err)
		}
		if _, err := GetWithTimeframesVenue("BTCUSDT", ex, []string{"5m"}, "5m", 10); err == nil || !strings.Contains(err.Error(), "no CoinAnk venue") {
			t.Errorf("GetWithTimeframesVenue(BTCUSDT, %q) must be refused by name, got %v", ex, err)
		}
		if _, err := GetBoxData("BTCUSDT", ex); err == nil || !strings.Contains(err.Error(), "no CoinAnk venue") {
			t.Errorf("GetBoxData(BTCUSDT, %q) must be refused by name, got %v", ex, err)
		}
	}
	// The venue-blind reads have no venue: a crypto symbol is refused too.
	if _, err := Get("BTCUSDT"); err == nil || !strings.Contains(err.Error(), "no CoinAnk venue") {
		t.Errorf("Get(BTCUSDT) has no venue and must be refused by name, got %v", err)
	}
	if _, err := GetWithTimeframes("BTCUSDT", []string{"5m"}, "5m", 10); err == nil || !strings.Contains(err.Error(), "no CoinAnk venue") {
		t.Errorf("GetWithTimeframes(BTCUSDT) has no venue and must be refused by name, got %v", err)
	}
	trap.mu.Lock()
	defer trap.mu.Unlock()
	if len(trap.hosts) != 0 {
		t.Fatalf("a refused venue must make ZERO requests (never another venue's book), got %v venues %v", trap.hosts, trap.venues)
	}
}

func TestFailedVenueReadIsAnErrorNeverASecondVenue(t *testing.T) {
	trap := installVenueTrap(t, true)
	if _, err := GetWithExchange("BTCUSDT", "bybit"); err == nil {
		t.Fatal("a failed CoinAnk read on the trader's venue must be an error")
	}
	trap.mu.Lock()
	defer trap.mu.Unlock()
	if len(trap.venues) != 1 || trap.venues[0] != string(coinank_enum.Bybit) {
		t.Fatalf("exactly ONE request, on Bybit — no cross-venue fallback; got venues %v", trap.venues)
	}
}

func TestCryptoReadCarriesOIAndFundingAbsent(t *testing.T) {
	trap := installVenueTrap(t, false)
	d, err := GetWithExchange("BTCUSDT", "okx")
	if err != nil {
		t.Fatalf("crypto read on okx: %v", err)
	}
	if d.OpenInterest != nil || d.FundingRateKnown || d.FundingRate != 0 {
		t.Fatalf("crypto OI/funding must be ABSENT (no source in this build): OI %+v funding %v known %v", d.OpenInterest, d.FundingRate, d.FundingRateKnown)
	}
	tf, err := GetWithTimeframesVenue("BTCUSDT", "okx", []string{"5m"}, "5m", 20)
	if err != nil {
		t.Fatalf("multi-timeframe crypto read on okx: %v", err)
	}
	if tf.OpenInterest != nil || tf.FundingRateKnown {
		t.Fatalf("GetWithTimeframesVenue: OI/funding must be ABSENT, got OI %+v known %v", tf.OpenInterest, tf.FundingRateKnown)
	}
	out := Format(d)
	if !strings.Contains(out, "Open Interest: n/a") || !strings.Contains(out, "Funding Rate: n/a") || strings.Contains(out, "0.00e+00") {
		t.Fatalf("Format must render absent OI/funding as n/a, never a fabricated 0:\n%s", out)
	}
	trap.mu.Lock()
	defer trap.mu.Unlock()
	if len(trap.hosts) == 0 {
		t.Fatal("vacuous: no request was made")
	}
	for i, h := range trap.hosts {
		if h != "api.coinank.com" || trap.venues[i] != string(coinank_enum.Okex) || trap.symbols[i] != "BTC-USDT-SWAP" {
			t.Fatalf("request %d: host %q venue %q symbol %q — want only CoinAnk, the trader's own venue (Okex), its own symbol spelling", i, h, trap.venues[i], trap.symbols[i])
		}
	}
}

func TestBoxDataIsVenueAware(t *testing.T) {
	trap := installVenueTrap(t, false)
	if _, err := GetBoxData("MNQ", "ninjatrader"); err == nil {
		t.Error("a CME futures symbol has no box source: refused")
	}
	if _, err := GetBoxData("BTCUSDT", "ninjatrader"); err == nil || !strings.Contains(err.Error(), "NinjaTrader venue") {
		t.Errorf("a non-CME symbol on the NinjaTrader venue is refused, got %v", err)
	}
	trap.mu.Lock()
	n := len(trap.hosts)
	trap.mu.Unlock()
	if n != 0 {
		t.Fatalf("refused box reads must make zero requests, got %d", n)
	}
	if _, err := GetBoxData("BTCUSDT", "gate"); err != nil {
		t.Fatalf("a crypto box read on gate: %v", err)
	}
	trap.mu.Lock()
	defer trap.mu.Unlock()
	if len(trap.venues) != 1 || trap.venues[0] != string(coinank_enum.Gate) {
		t.Fatalf("the box read must use the trader's own venue (Gate), got %v", trap.venues)
	}
}
