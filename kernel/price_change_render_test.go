package kernel

import (
	"strings"
	"testing"
	"time"

	"vl/market"
	"vl/store"
)

// W1 (presence-aware values) — the 1h/4h price change reaches the prompts as
// a measured value or as n/a, never a fabricated 0 and never a shorter window
// under the longer window's name. Every test drives the production chain: the
// real market read for CME futures (the NT8 bar provider, wired below), then
// the real renderer.

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
// 150-minute NT8 tape → 1h measured, 4h n/a.
func TestGridPromptRendersAnUnmeasurableWindowAsNA(t *testing.T) {
	prev := market.FuturesBarsProvider
	market.FuturesBarsProvider = func(symbol, tf string, count int) []market.Kline {
		if symbol != "MNQ" || tf != "5m" {
			return nil
		}
		return fixedBars(5*time.Minute, 30, 60000)
	}
	t.Cleanup(func() { market.FuturesBarsProvider = prev })

	d, err := market.GetWithTimeframes("MNQ", []string{"5m", "4h"}, "5m", 50)
	if err != nil {
		t.Fatalf("fixture: the grid read failed: %v", err)
	}
	gctx := BuildGridContextFromMarketData(d, &store.GridStrategyConfig{Symbol: "MNQ", GridCount: 10, TotalInvestment: 1000, Leverage: 1})
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
