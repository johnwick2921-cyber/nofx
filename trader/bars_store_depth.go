package trader

import (
	"time"

	"nofx/kernel"
	"nofx/logger"
	"nofx/market"
	"nofx/store"
)

// ── D2 — THE STORE IS THE HORIZON; THE RING IS THE CACHE ────────────────────
// (owner ruling, wave BARS HORIZON 2026-09-09)
//
// Three call sites asked for depth NO MARKET CONDITION COULD SUPPLY. The ring
// ceiling is DefaultBarCacheMaxBars = 2500 per (symbol, timeframe)
// (provider/ninjatrader/bar_cache.go:24), and a Go restart rebuilds it from a
// 2000-bar seed (defaultAutoBarsBack, tcp_server.go):
//
//	auto_trader_planner.go   1m × 12000 = 200.0 h  vs a 41.7 h ceiling
//	auto_trader_weekly.go    1m × 12000 = 200.0 h  vs a 41.7 h ceiling
//	auto_trader_planner.go   5m ×  3000 = 250.0 h  vs a 208.3 h ceiling
//
// The store held the depth the whole time and was wired to nothing:
// market.FuturesBarsProvider has ONE production assignment
// (trader/ninjatrader/bars_market_bridge.go) and it reads the BarCache only.
//
// CHOICE PER CALLER, recorded at the call site itself:
//   - the two 1m sites take (a) READ THE STORE — a candle table promising 8
//     session-days needs ~11,040 open minutes and the store has 21 days of them;
//   - the 5m site takes (b) CORRECT THE ASK AND STOP PROMISING 20 DAYS — see
//     rvBaseline5mBarsAsk in auto_trader_planner.go for why (a) was refused there.
//
// A10: this never gates, refuses, blocks or blanks. Every failure path returns
// what the ring returned, which is exactly the pre-wave value.

// barsWithStoreDepth serves a bar request the ring alone cannot fill, by
// EXTENDING it backwards from the store. `now` comes from the caller (A28).
func (at *AutoTrader) barsWithStoreDepth(symbol, tf string, n int, now time.Time) []market.Kline {
	var ring []market.Kline
	if market.FuturesBarsProvider != nil {
		ring = market.FuturesBarsProvider(symbol, tf, n)
	}
	return barsWithStoreDepthFrom(ring, at.storeBarReader(symbol, tf), symbol, tf, n, now)
}

// storeBarReader returns the store read as a closure, so barsWithStoreDepthFrom
// — the merge that actually matters — is driven by pins rather than by a copy
// of itself (class 86). A nil store yields a reader that reports UNAVAILABLE
// rather than an empty result, so "no store" can never look like "no history".
func (at *AutoTrader) storeBarReader(symbol, tf string) func(int) ([]market.Kline, error) {
	return func(n int) ([]market.Kline, error) {
		if at == nil || at.store == nil {
			return nil, errStoreUnavailable
		}
		rows, err := at.store.BarHistory().LastNBars(symbol, tf, n)
		if err != nil {
			return nil, err
		}
		return storeRowsToKlines(rows, tf), nil
	}
}

// errStoreUnavailable distinguishes "there is no store" from "the store is
// empty" — an uncomputed answer is UNKNOWN, never zero (A24).
var errStoreUnavailable = storeUnavailableError{}

type storeUnavailableError struct{}

func (storeUnavailableError) Error() string { return "no store attached — depth read skipped" }

// storeRowsToKlines adapts stored rows to the ONE Kline shape every reader uses.
// CloseTime is derived from the timeframe exactly as the NT8 bridge derives it,
// so a store-sourced bar and a ring-sourced bar are indistinguishable downstream.
func storeRowsToKlines(rows []store.BarHistoryDB, tf string) []market.Kline {
	if len(rows) == 0 {
		return nil
	}
	durMs, _ := kernel.TFDurationMs(tf)
	out := make([]market.Kline, 0, len(rows))
	for _, r := range rows {
		k := market.Kline{OpenTime: r.OpenTimeMs, Open: r.O, High: r.H, Low: r.L, Close: r.C, Volume: r.V}
		if durMs > 0 {
			k.CloseTime = r.OpenTimeMs + durMs - 1
		}
		out = append(out, k)
	}
	return out
}

// barsWithStoreDepthFrom splices a store read onto the OLDER end of a ring read.
//
// THE RULES, in the order they are applied:
//
//  1. AN EMPTY RING IS NEVER SUBSTITUTED. An empty ring means the feed is down
//     or the AddOn seed has not landed; handing a planner 12,000 stored bars
//     would let it write a plan on a dead tape, which is the exact failure
//     "no NT8 → no decisions" exists to prevent. The store DEEPENS a live tape;
//     it never stands in for one.
//  2. A RING THAT ALREADY SERVES THE ASK IS NOT SECOND-GUESSED — no store read
//     at all, so the ~29 healthy call sites pay nothing.
//  3. ONLY BARS STRICTLY OLDER THAN THE RING'S OLDEST ARE TAKEN. A live bar is
//     never replaced by a stored one, so the freshest OHLCV always wins and the
//     forming bar survives untouched.
//  4. A FAILED OR EMPTY STORE READ DEGRADES TO THE RING, loudly (A9/A10).
//
// The result is ascending by open time, tail-capped at n, and byte-identical to
// the pre-wave value whenever the store adds nothing.
func barsWithStoreDepthFrom(ring []market.Kline, storeRead func(int) ([]market.Kline, error), symbol, tf string, n int, now time.Time) []market.Kline {
	if len(ring) == 0 || len(ring) >= n || n <= 0 || storeRead == nil {
		return ring
	}
	stored, err := storeRead(n)
	if err != nil {
		logger.Warnf("📚 bars depth: %s %s ask=%d ring=%d — store read FAILED (%v); serving the ring alone (nothing gated)",
			symbol, tf, n, len(ring), err)
		return ring
	}
	oldestRing := ring[0].OpenTime
	older := make([]market.Kline, 0, len(stored))
	for _, k := range stored {
		if k.OpenTime < oldestRing {
			older = append(older, k)
		}
	}
	if len(older) == 0 {
		logger.Infof("📚 bars depth: %s %s ask=%d ring=%d store=%d older=0 — the store holds nothing before the ring's oldest bar (%s CT)",
			symbol, tf, n, len(ring), len(stored), kernel.ClockHHMMCT(time.UnixMilli(oldestRing)))
		return ring
	}
	out := make([]market.Kline, 0, len(older)+len(ring))
	out = append(out, older...)
	out = append(out, ring...)
	if len(out) > n {
		out = out[len(out)-n:]
	}
	before := kernel.HorizonOf(ring, tf, n, now)
	after := kernel.HorizonOf(out, tf, n, now)
	logger.Infof("📚 bars depth: %s %s ask=%d · ring %s → store-deepened %s (+%d older bars; live tail untouched)",
		symbol, tf, n, before.Line(), after.Line(), len(older))
	return out
}
