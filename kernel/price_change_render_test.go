package kernel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"vl/hook"
	"vl/market"
	"vl/store"
)

// W1 (presence-aware values) — the 1h/4h price change reaches the prompts as
// a measured value or as n/a, never a fabricated 0 and never a shorter window
// under the longer window's name. Every test drives the production chain: the
// real market read over CoinAnk's kline endpoint (served offline here), then
// the real renderer.

// coinankTape answers CoinAnk's kline endpoint from fixed series keyed by the
// request's interval, and fails every other host (the crypto OI/funding
// fetchers get an error, as they would offline).
type coinankTape struct {
	mu     sync.Mutex
	series map[string][]market.Kline
	hosts  []string
}

func (c *coinankTape) RoundTrip(r *http.Request) (*http.Response, error) {
	c.mu.Lock()
	c.hosts = append(c.hosts, r.URL.Host)
	c.mu.Unlock()
	if r.URL.Host != "api.coinank.com" || r.URL.Path != "/api/kline/list/open" {
		return nil, fmt.Errorf("no network in this test (%s)", r.URL.Host)
	}
	s := c.series[r.URL.Query().Get("interval")]
	if n, err := strconv.Atoi(r.URL.Query().Get("size")); err == nil && n > 0 && n < len(s) {
		s = s[len(s)-n:]
	}
	rows := make([][]float64, len(s))
	for i, k := range s {
		rows[i] = []float64{float64(k.OpenTime), float64(k.CloseTime), k.Open, k.Close, k.High, k.Low, k.Volume, k.Volume, 1}
	}
	body, _ := json.Marshal(map[string]any{"success": true, "data": rows})
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}},
		Body: io.NopCloser(strings.NewReader(string(body))), Request: r}, nil
}

func installCoinankTape(t *testing.T, series map[string][]market.Kline) *coinankTape {
	t.Helper()
	c := &coinankTape{series: series}
	prev, had := hook.Hooks[hook.SET_HTTP_CLIENT]
	prevEnabled := hook.EnableHooks
	hook.EnableHooks = true
	hook.RegisterHook(hook.SET_HTTP_CLIENT, func(args ...any) any {
		return &hook.SetHttpClientResult{Client: &http.Client{Transport: c, Timeout: time.Second}}
	})
	prevDT := http.DefaultTransport
	http.DefaultTransport = c
	t.Cleanup(func() {
		http.DefaultTransport = prevDT
		if had {
			hook.Hooks[hook.SET_HTTP_CLIENT] = prev
		} else {
			delete(hook.Hooks, hook.SET_HTTP_CLIENT)
		}
		hook.EnableHooks = prevEnabled
	})
	return c
}

// fixedBars: n contiguous bars of `step` ending at a FIXED instant (goldens
// must not move with the clock); closes rise by 1.5 per bar from base.
func fixedBars(step time.Duration, n int, base float64) []market.Kline {
	end := time.Date(2026, 9, 23, 14, 0, 0, 0, time.UTC)
	start := end.Add(-time.Duration(n) * step)
	out := make([]market.Kline, n)
	for i := range out {
		ot := start.Add(time.Duration(i) * step).UnixMilli()
		c := base + float64(i)*1.5
		out[i] = market.Kline{OpenTime: ot, Open: c - 0.5, High: c + 1, Low: c - 1, Close: c, Volume: 10, CloseTime: ot + step.Milliseconds() - 1}
	}
	return out
}

// The grid prompt (en and zh): the grid's real read (5m primary) over a
// 150-minute tape → 1h measured, 4h n/a.
func TestGridPromptRendersAnUnmeasurableWindowAsNA(t *testing.T) {
	installCoinankTape(t, map[string][]market.Kline{
		"5m": fixedBars(5*time.Minute, 30, 60000),
		"4h": fixedBars(4*time.Hour, 30, 58000),
	})
	d, err := market.GetWithTimeframes("BTCUSDT", []string{"5m", "4h"}, "5m", 50)
	if err != nil {
		t.Fatalf("fixture: the grid read failed: %v", err)
	}
	gctx := BuildGridContextFromMarketData(d, &store.GridStrategyConfig{Symbol: "BTCUSDT", GridCount: 10, TotalInvestment: 1000, Leverage: 1})
	gctx.CurrentTime = "2026-09-23 09:00:00"
	en := BuildGridUserPrompt(gctx, "en")
	if !strings.Contains(en, "- 4h Change: n/a\n") || strings.Contains(en, "- 1h Change: n/a") {
		t.Fatalf("grid en: 1h measured, 4h n/a:\n%s", en)
	}
	if zh := BuildGridUserPrompt(gctx, "zh"); !strings.Contains(zh, "- 4小时涨跌: n/a\n") {
		t.Fatalf("grid zh: 4h n/a:\n%s", zh)
	}
	assertGolden(t, "grid_user_en_change.txt", en)
}
