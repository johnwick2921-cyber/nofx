package trader

import (
	"nofx/kernel"
	"nofx/market"
	"testing"
	"time"
)

// Exercise the reader's production call site, not two calls to the same resolver.
// Distinct Friday extremes reveal a seven-day epoch bucket crossing CME weeks.
func TestWeeklyReaderPreservesDailyCalendarEvidence(t *testing.T) {
	old := market.FuturesBarsProvider
	t.Cleanup(func() { market.FuturesBarsProvider = old })
	ct := kernel.CTLocation()
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, ct)
	var daily []market.Kline
	for week := 0; week < 2; week++ {
		for day := 0; day < 5; day++ {
			stamp := time.Date(2026, 8, 31+week*7+day, 0, 0, 0, 0, ct)
			p := 100 + float64(week*100+day)
			daily = append(daily, market.Kline{OpenTime: stamp.UnixMilli(), Open: p, High: p + 10, Low: p - 10, Close: p + 1, Volume: 1})
		}
	}
	market.FuturesBarsProvider = func(_, tf string, _ int) []market.Kline {
		if tf == "1d" {
			return daily
		}
		return nil
	}
	at := &AutoTrader{}
	bars, _ := at.weeklyDailyBars(now)
	if len(bars) != 10 {
		t.Fatalf("weekly reader must preserve ten daily observations, got %d", len(bars))
	}
	weeks := kernel.CompletedWeekCandles(bars, now, 0)
	if len(weeks) != 2 {
		t.Fatalf("want two CME weeks, got %+v", weeks)
	}
	for i, w := range weeks {
		p := 100 + float64(i*100)
		if w.Open != p || w.High != p+14 || w.Low != p-10 || w.Close != p+5 || w.Volume != 5 {
			t.Fatalf("week %d crossed calendar boundary: %+v", i, w)
		}
	}
	facts := kernel.ComputeWeeklyFacts(bars, now, 205)
	if facts.Refs.PWH != 214 || facts.Refs.PWL != 190 || facts.Refs.PWC != 205 {
		t.Fatalf("prior-week references lost daily evidence: %+v", facts.Refs)
	}
}
