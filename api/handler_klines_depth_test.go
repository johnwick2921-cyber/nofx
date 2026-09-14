package api

import (
	"path/filepath"
	"testing"
	"time"

	"nofx/market"
	"nofx/store"
)

// F1 (2026-09-14) — the dashboard klines path (ninjatrader) deepens from the
// store on the CURRENT contract only. The retired contract's rows exist but must
// never be served, and a ring (live) bar is never replaced by a stored one.
func TestKlinesNinjaTraderStoreDepthContractFiltered(t *testing.T) {
	orig := market.FuturesBarsProvider
	defer func() { market.FuturesBarsProvider = orig }()

	st, err := store.New(filepath.Join(t.TempDir(), "klines.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	if err := st.BarHistory().Migrate(); err != nil {
		t.Fatal(err)
	}

	// Ring: 2 live bars at 10:00 / 10:01, current contract.
	ringStart := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	market.FuturesBarsProvider = func(symbol, tf string, count int) []market.Kline {
		return []market.Kline{
			{OpenTime: ringStart.UnixMilli(), Open: 100, High: 101, Low: 99, Close: 100.5, Volume: 1},
			{OpenTime: ringStart.Add(time.Minute).UnixMilli(), Open: 101, High: 102, Low: 100, Close: 101.5, Volume: 1},
		}
	}

	// Store: three older bars on the CURRENT contract, and three OLDER-STILL
	// bars on a retired contract whose Close carries a sentinel (-1) so a leak
	// is visible at a glance.
	older := func(i int) int64 { return ringStart.Add(-time.Duration(i) * time.Minute).UnixMilli() }
	var rows []store.BarHistoryDB
	for i := 1; i <= 3; i++ {
		rows = append(rows, store.BarHistoryDB{Symbol: "MNQ", TF: "1m", OpenTimeMs: older(i), O: 90, H: 91, L: 89, C: 90.5, V: 1, Contract: "MNQ 12-26", Source: store.BarSourceLive})
	}
	for i := 4; i <= 6; i++ {
		rows = append(rows, store.BarHistoryDB{Symbol: "MNQ", TF: "1m", OpenTimeMs: older(i), O: -1, H: -1, L: -1, C: -1, V: 1, Contract: "MNQ 09-26", Source: store.BarSourceLive})
	}
	if err := st.BarHistory().InsertBars(rows); err != nil {
		t.Fatal(err)
	}

	s := &Server{store: st}
	out := s.getKlinesFromNinjaTrader("MNQ", "1m", 6)
	if len(out) != 5 {
		t.Fatalf("served %d bars, want 5 (3 stored current-contract + 2 ring); the retired contract must be filtered", len(out))
	}
	if out[0].OpenTime != older(3) {
		t.Fatalf("oldest served %d, want the store's oldest CURRENT-contract bar %d", out[0].OpenTime, older(3))
	}
	if got := out[len(out)-1].OpenTime; got != ringStart.Add(time.Minute).UnixMilli() {
		t.Fatalf("newest served %d — the ring's live tail must survive untouched", got)
	}
	for _, k := range out {
		if k.Close < 0 {
			t.Fatalf("a retired-contract (09-26) bar was served: %+v", k)
		}
	}
}

// F1 — no store attached: the handler serves the ring exactly as before.
func TestKlinesNinjaTraderNoStoreServesRing(t *testing.T) {
	orig := market.FuturesBarsProvider
	defer func() { market.FuturesBarsProvider = orig }()
	market.FuturesBarsProvider = func(symbol, tf string, count int) []market.Kline {
		return []market.Kline{{OpenTime: 1, Close: 5}}
	}
	s := &Server{}
	out := s.getKlinesFromNinjaTrader("MNQ", "1m", 100)
	if len(out) != 1 || out[0].Close != 5 {
		t.Fatalf("without a store the ring must pass through unchanged: %+v", out)
	}
}

// F1 (2026-09-14) — the dashboard's 5,000-bar ask must SURVIVE the handler's
// limit parsing. The global 1500 clamp was a Coinank constraint that was
// silently also capping the ninjatrader path, so the store splice could never
// fire on a warm ring.
func TestResolveKlinesLimitPerExchange(t *testing.T) {
	cases := []struct {
		exchange string
		limit    string
		want     int
	}{
		{"ninjatrader", "5000", 5000},    // F1 dashboard ask survives
		{"ninjatrader", "999999", 20000}, // ninjatrader ceiling, not Coinank's
		{"NinjaTrader", "5000", 5000},    // case-insensitive
		{"binance", "5000", 1500},        // Coinank cap unchanged
		{"", "5000", 1500},               // default exchange inherits Coinank cap
		{"binance", "abc", 1000},         // bad value → default
		{"ninjatrader", "0", 1000},       // non-positive → default
		{"", "", 1000},                   // missing → default
	}
	for _, tc := range cases {
		if got := resolveKlinesLimit(tc.exchange, tc.limit); got != tc.want {
			t.Errorf("resolveKlinesLimit(%q, %q) = %d, want %d", tc.exchange, tc.limit, got, tc.want)
		}
	}
}
