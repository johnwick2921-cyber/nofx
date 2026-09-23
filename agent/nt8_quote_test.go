package agent

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"nofx/market"
	"nofx/mcp"
)

// The chat's two market reads — GET /api/agent/tickers (WebHandler.HandleTickers)
// and the get_market_snapshot tool (handleToolCall) — read NinjaTrader bars
// only. These pins drive the production entry points: the HTTP handler and the
// tool dispatcher. market.FuturesBarsProvider is the real production seam (the
// NT8 bridge binds it at TCP-server start), so a CME read below runs the real
// market.GetWithExchange.

// outboundHTTPCounter replaces http.DefaultTransport and fails every request,
// so a test can prove the read made ZERO outbound HTTP calls.
type outboundHTTPCounter struct{ calls atomic.Int64 }

func (c *outboundHTTPCounter) RoundTrip(req *http.Request) (*http.Response, error) {
	c.calls.Add(1)
	return nil, errors.New("outbound HTTP is forbidden in this test: " + req.URL.String())
}

func forbidOutboundHTTP(t *testing.T) *outboundHTTPCounter {
	t.Helper()
	counter := &outboundHTTPCounter{}
	prev := http.DefaultTransport
	http.DefaultTransport = counter
	t.Cleanup(func() { http.DefaultTransport = prev })
	return counter
}

// stubNT8Bars installs a FuturesBarsProvider that serves `closes5m` as 5m bars
// and `closes1h` as 1h bars for exactly the symbol "MNQ", counting calls.
func stubNT8Bars(t *testing.T, closes5m, closes1h []float64) *atomic.Int64 {
	t.Helper()
	var calls atomic.Int64
	prev := market.FuturesBarsProvider
	market.FuturesBarsProvider = func(symbol, timeframe string, count int) []market.Kline {
		calls.Add(1)
		if symbol != "MNQ" {
			return nil
		}
		var closes []float64
		var step int64
		switch timeframe {
		case "5m":
			closes, step = closes5m, 300_000
		case "1h":
			closes, step = closes1h, 3_600_000
		default:
			return nil
		}
		out := make([]market.Kline, 0, len(closes))
		base := int64(1_790_000_000_000)
		for i, c := range closes {
			open := base + int64(i)*step
			out = append(out, market.Kline{OpenTime: open, Open: c, High: c + 1, Low: c - 1, Close: c, Volume: 100, CloseTime: open + step - 1})
		}
		return out
	}
	t.Cleanup(func() { market.FuturesBarsProvider = prev })
	return &calls
}

func rising5mCloses() []float64 {
	closes := make([]float64, 30)
	for i := range closes {
		closes[i] = 20000 + float64(i) // 20000 … 20029
	}
	return closes
}

func getTickers(t *testing.T, query string) (int, []map[string]any, string) {
	t.Helper()
	h := NewWebHandler(nil, slog.Default())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/agent/tickers"+query, nil)
	h.HandleTickers(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		return rec.Code, nil, body
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(body), &rows); err != nil {
		t.Fatalf("tickers body is not a JSON array: %v — %s", err, body)
	}
	return rec.Code, rows, body
}

// assertUnavailableRow pins one contract row that carries no read: every value
// key is PRESENT and null (absent ≠ 0), source is "nt8", and the named reason
// is set.
func assertUnavailableRow(t *testing.T, row map[string]any, wantSymbol, wantReasonContains string) {
	t.Helper()
	if row["symbol"] != wantSymbol {
		t.Fatalf("symbol = %v, want %q (as requested)", row["symbol"], wantSymbol)
	}
	for _, key := range []string{"price", "change_1h_pct", "change_4h_pct"} {
		v, present := row[key]
		if !present {
			t.Fatalf("%s: key %q missing — the contract carries it as null", wantSymbol, key)
		}
		if v != nil {
			t.Fatalf("%s: %s = %v, want null (no fabricated value)", wantSymbol, key, v)
		}
	}
	if row["source"] != "nt8" {
		t.Fatalf("%s: source = %v, want nt8", wantSymbol, row["source"])
	}
	reason, _ := row["unavailable"].(string)
	if reason == "" || !strings.Contains(reason, wantReasonContains) {
		t.Fatalf("%s: unavailable = %q, want it to contain %q", wantSymbol, reason, wantReasonContains)
	}
}

func TestHandleTickers_NonCMESymbolsAreNamedUnavailableWithNoReadAndNoOutboundHTTP(t *testing.T) {
	outbound := forbidOutboundHTTP(t)
	reads := stubNT8Bars(t, rising5mCloses(), []float64{19900, 20000})

	code, rows, body := getTickers(t, "?symbols=BTCUSDT,AAPL,eth")
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — %s", code, body)
	}
	if len(rows) != 3 {
		t.Fatalf("want one row per requested symbol (3), got %d — %s", len(rows), body)
	}
	for i, sym := range []string{"BTCUSDT", "AAPL", "eth"} {
		assertUnavailableRow(t, rows[i], sym, nt8QuoteNonCMEReason)
	}
	if n := reads.Load(); n != 0 {
		t.Fatalf("a non-CME symbol must not reach the market read; FuturesBarsProvider called %d times", n)
	}
	if n := outbound.calls.Load(); n != 0 {
		t.Fatalf("outbound HTTP calls = %d, want 0", n)
	}
}

func TestHandleTickers_CMESymbolReadsNT8Bars(t *testing.T) {
	outbound := forbidOutboundHTTP(t)
	stubNT8Bars(t, rising5mCloses(), []float64{19900, 20000})

	code, rows, body := getTickers(t, "?symbols=MNQ,BTCUSDT")
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — %s", code, body)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d — %s", len(rows), body)
	}
	mnq := rows[0]
	if mnq["symbol"] != "MNQ" || mnq["source"] != "nt8" {
		t.Fatalf("MNQ row = %v", mnq)
	}
	if _, has := mnq["unavailable"]; has {
		t.Fatalf("a successful read must not carry unavailable: %v", mnq)
	}
	// price = the last 5m close (market.GetWithExchange). The 1h change is
	// measured by bar TIME on the 5m series: the bar that closed 60 min before
	// the last one is bar 17 (12 bars × 5 min back), close 20017. The 30-bar
	// series spans 150 min, so the 4h window has no reference bar → null,
	// never 0 and never the prior 1h close.
	want := map[string]float64{
		"price":         20029,
		"change_1h_pct": (20029.0 - 20017.0) / 20017.0 * 100,
	}
	if v, has := mnq["change_4h_pct"]; !has || v != nil {
		t.Fatalf("MNQ change_4h_pct = %v (present=%v), want null — the series does not reach back 4h", v, has)
	}
	for key, w := range want {
		got, ok := mnq[key].(float64)
		if !ok {
			t.Fatalf("MNQ %s = %v, want a number", key, mnq[key])
		}
		if math.Abs(got-w) > 1e-9 {
			t.Fatalf("MNQ %s = %v, want %v", key, got, w)
		}
	}
	assertUnavailableRow(t, rows[1], "BTCUSDT", nt8QuoteNonCMEReason)
	if n := outbound.calls.Load(); n != 0 {
		t.Fatalf("outbound HTTP calls = %d, want 0", n)
	}
}

func TestHandleTickers_NT8ReadFailurePassesTheReasonThrough(t *testing.T) {
	forbidOutboundHTTP(t)

	// No bars cached for the symbol → the read's own error, verbatim.
	stubNT8Bars(t, nil, nil)
	_, rows, body := getTickers(t, "?symbols=MNQ")
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %s", body)
	}
	assertUnavailableRow(t, rows[0], "MNQ", "no NT8 5m bars cached")

	// No provider wired (no NT8 trader loaded) → named, not a price.
	prev := market.FuturesBarsProvider
	market.FuturesBarsProvider = nil
	t.Cleanup(func() { market.FuturesBarsProvider = prev })
	_, rows, body = getTickers(t, "?symbols=MNQ")
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %s", body)
	}
	assertUnavailableRow(t, rows[0], "MNQ", "no NT8 bar provider wired")
}

func TestHandleTickers_RefusesNoSymbolsAndNamesAMalformedOne(t *testing.T) {
	forbidOutboundHTTP(t)
	stubNT8Bars(t, rising5mCloses(), []float64{19900, 20000})

	for _, q := range []string{"", "?symbols=", "?symbols=%20,%20"} {
		code, _, body := getTickers(t, q)
		if code != http.StatusBadRequest || !strings.Contains(body, "symbols required") {
			t.Fatalf("query %q: status %d body %s — want 400 'symbols required' (no default list)", q, code, body)
		}
	}

	code, rows, body := getTickers(t, "?symbols=MNQ,bad%2Fsym")
	if code != http.StatusOK || len(rows) != 2 {
		t.Fatalf("status %d body %s", code, body)
	}
	assertUnavailableRow(t, rows[1], "bad/sym", "invalid symbol")

	many := strings.TrimSuffix(strings.Repeat("MNQ,", maxTickerSymbols+1), ",")
	if code, _, body := getTickers(t, "?symbols="+many); code != http.StatusBadRequest {
		t.Fatalf("%d symbols: status %d body %s, want 400", maxTickerSymbols+1, code, body)
	}
}

func callMarketSnapshotTool(t *testing.T, args string) map[string]any {
	t.Helper()
	a := New(nil, nil, DefaultConfig(), nil)
	raw := a.handleToolCall(context.Background(), "default", 1, "en", mcp.ToolCall{
		Type:     "function",
		Function: mcp.ToolCallFunction{Name: "get_market_snapshot", Arguments: args},
	})
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("tool result is not JSON: %v — %s", err, raw)
	}
	return out
}

func TestMarketSnapshotTool_NonCMESymbolIsNamedUnavailableWithNoOutboundHTTP(t *testing.T) {
	outbound := forbidOutboundHTTP(t)
	reads := stubNT8Bars(t, rising5mCloses(), []float64{19900, 20000})

	for _, sym := range []string{"BTC", "ETHUSDT", "AAPL"} {
		out := callMarketSnapshotTool(t, `{"symbol":"`+sym+`"}`)
		assertUnavailableRow(t, out, sym, nt8QuoteNonCMEReason)
	}
	if n := reads.Load(); n != 0 {
		t.Fatalf("FuturesBarsProvider called %d times for non-CME symbols, want 0", n)
	}
	if n := outbound.calls.Load(); n != 0 {
		t.Fatalf("outbound HTTP calls = %d, want 0", n)
	}
}

func TestMarketSnapshotTool_CMESymbolReadsNT8Bars(t *testing.T) {
	outbound := forbidOutboundHTTP(t)
	stubNT8Bars(t, rising5mCloses(), []float64{19900, 20000})

	out := callMarketSnapshotTool(t, `{"symbol":"MNQ"}`)
	if out["symbol"] != "MNQ" || out["source"] != "nt8" {
		t.Fatalf("snapshot = %v", out)
	}
	if _, has := out["unavailable"]; has {
		t.Fatalf("a successful read must not carry unavailable: %v", out)
	}
	if price, _ := out["price"].(float64); price != 20029 {
		t.Fatalf("price = %v, want 20029 (the last NT8 5m close)", out["price"])
	}
	if n := outbound.calls.Load(); n != 0 {
		t.Fatalf("outbound HTTP calls = %d, want 0", n)
	}

	if raw := callMarketSnapshotTool(t, `{"symbol":"  "}`); raw["error"] != "symbol is required" {
		t.Fatalf("blank symbol = %v, want the named 'symbol is required'", raw)
	}
}

func TestGetKlineToolIsGone(t *testing.T) {
	for _, tool := range agentTools() {
		if tool.Function.Name == "get_kline" {
			t.Fatal("get_kline is still offered to the model")
		}
	}
	for _, domain := range []string{"market", "general", "account", "trader", "diagnosis"} {
		for _, name := range plannerToolNamesForDomain(domain) {
			if name == "get_kline" {
				t.Fatalf("get_kline still listed in the %q planner domain", domain)
			}
		}
	}
	a := New(nil, nil, DefaultConfig(), nil)
	raw := a.handleToolCall(context.Background(), "default", 1, "en", mcp.ToolCall{
		Function: mcp.ToolCallFunction{Name: "get_kline", Arguments: `{"symbol":"BTC"}`},
	})
	if !strings.Contains(raw, "unknown tool: get_kline") {
		t.Fatalf("get_kline dispatch = %s, want unknown tool", raw)
	}
}
