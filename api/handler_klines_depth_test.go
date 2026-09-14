package api

import (
	"path/filepath"
	"testing"
	"time"

	"nofx/market"
	"nofx/store"
)

// F1.1 (2026-09-14) — a thin coarse-TF ring gets older CLOSED buckets
// aggregated from a deep finer rung, capped at the ask, and a ring the finer
// rung cannot extend passes through unchanged.
func TestKlinesAggregatedDepthPrependsOlderClosedBuckets(t *testing.T) {
	now := time.Date(2026, 9, 14, 18, 0, 0, 0, time.UTC)
	base := []market.Kline{
		{OpenTime: now.Add(-6 * time.Hour).UnixMilli(), Open: 10, High: 11, Low: 9, Close: 10.5},
		{OpenTime: now.Add(-2 * time.Hour).UnixMilli(), Open: 11, High: 12, Low: 10, Close: 11.5},
	}
	one := func(h int) market.Kline {
		return market.Kline{OpenTime: now.Add(-time.Duration(h) * time.Hour).UnixMilli(), Open: float64(h), High: float64(h) + 1, Low: float64(h) - 1, Close: float64(h) + 0.5, Volume: 1}
	}
	ring1h := []market.Kline{one(10), one(9), one(8), one(7), one(6), one(5)}
	provider := func(symbol, tf string, count int) []market.Kline {
		if tf == "1h" {
			return ring1h
		}
		return base
	}
	out := klinesWithAggregatedDepth(base, provider, "MNQ", "4h", 10, now)
	if len(out) != 3 {
		t.Fatalf("served %d bars, want 3 (1 aggregated closed 4h bucket + 2 base)", len(out))
	}
	wantOldest := now.Add(-10 * time.Hour).Truncate(4 * time.Hour).UnixMilli()
	if out[0].OpenTime != wantOldest {
		t.Fatalf("oldest served %d, want the 08:00-12:00 aggregated bucket %d", out[0].OpenTime, wantOldest)
	}
	// The AGGREGATED prefix must be closed (the base's own tail may hold the
	// forming bucket — the chart shows it, the splice never ADDS one).
	if out[0].OpenTime+4*3600_000 > now.UnixMilli() {
		t.Fatalf("a forming 4h bucket was served: %+v", out[0])
	}
}

func TestKlinesAggregatedDepthFallsBackWhenNoOlder(t *testing.T) {
	now := time.Date(2026, 9, 14, 18, 0, 0, 0, time.UTC)
	base := []market.Kline{
		{OpenTime: now.Add(-6 * time.Hour).UnixMilli(), Open: 10, High: 11, Low: 9, Close: 10.5},
	}
	// The finer rung holds NOTHING older than the base's oldest bar.
	provider := func(symbol, tf string, count int) []market.Kline {
		if tf == "1h" {
			return []market.Kline{{OpenTime: now.Add(-1 * time.Hour).UnixMilli(), Close: 5}}
		}
		return base
	}
	out := klinesWithAggregatedDepth(base, provider, "MNQ", "4h", 10, now)
	if len(out) != 1 {
		t.Fatalf("a rung with no older bars must not change the series: %d bars", len(out))
	}
}

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
	// A wave-101 bulk-import snapshot OLDER than everything else, sentinel
	// Close=50: honest store history, but the DISPLAY chart must skip it.
	// (Imports ride ImportBars — InsertBars refuses historical_import.)
	snapshot := store.BarHistoryDB{Symbol: "MNQ", TF: "1m", OpenTimeMs: older(7), O: 50, H: 50, L: 50, C: 50, V: 1, Contract: "MNQ 12-26", Source: store.BarSourceHistoricalImport}
	if err := st.BarHistory().InsertBars(rows); err != nil {
		t.Fatal(err)
	}
	if _, _, err := st.BarHistory().ImportBars([]store.BarHistoryDB{snapshot}); err != nil {
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
		if k.Close == 50 {
			t.Fatalf("a historical_import snapshot rendered on the display chart: %+v", k)
		}
	}
}

// F1.2 (2026-09-14) — an EMPTY native ring for the requested TF is aggregated
// from the finer live rung (closed buckets only). Measured live: after a
// re-subscribe NT8 returned zero 2h/4h bars while the 1h ring held 1,500.
func TestKlinesAggregatedDepthFillsEmptyBase(t *testing.T) {
	now := time.Date(2026, 9, 14, 18, 0, 0, 0, time.UTC)
	one := func(h int) market.Kline {
		return market.Kline{OpenTime: now.Add(-time.Duration(h) * time.Hour).UnixMilli(), Open: float64(h), High: float64(h) + 1, Low: float64(h) - 1, Close: float64(h) + 0.5, Volume: 1}
	}
	// 1h rung: 09:00,10:00,11:00,12:00 → one closed 4h bucket (08:00-12:00) and
	// one bucket (12:00-16:00, closed too). At 18:00 both are closed.
	ring1h := []market.Kline{one(9), one(8), one(7), one(6)}
	provider := func(symbol, tf string, count int) []market.Kline {
		if tf == "1h" {
			return ring1h
		}
		return nil // the requested TF's own ring is EMPTY
	}
	out := klinesWithAggregatedDepth(nil, provider, "MNQ", "4h", 10, now)
	if len(out) != 2 {
		t.Fatalf("empty base + deep finer rung must aggregate: got %d 4h bars, want 2", len(out))
	}
	if out[0].OpenTime != now.Add(-10*time.Hour).Truncate(4*time.Hour).UnixMilli() {
		t.Fatalf("oldest aggregated bucket %d, want 08:00", out[0].OpenTime)
	}
	for _, k := range out {
		if k.OpenTime+4*3600_000 > now.UnixMilli() {
			t.Fatalf("a forming 4h bucket was served from an empty base: %+v", k)
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
