// Package ninjatrader — Plan 4.4 Stage 2 bar cache.
//
// Holds the latest N OHLCV bars per (symbol, timeframe) received over the
// TCP wire from the C# AddOn (frames bars_historical + bar_update). Thread-
// safe and bounded; read by the kernel (Stage 3 decision feed) and the
// chart relay (Stage 4) via Get().
//
// Backpressure invariant: writes from the socket read loop go through a
// bounded channel into a dedicated drain goroutine, which then calls into
// this cache. The cache itself uses a brief Lock() per write and an RLock
// + copy-on-read for snapshots, so reads never block writes for long.
package ninjatrader

import "sync"

// DefaultBarCacheMaxBars is the per-(symbol, timeframe) ring-buffer
// capacity. 1024 covers EMA200 + ATR14 + 4-hour intraday history at 1m
// granularity with comfortable headroom for the chart's lookback window.
const DefaultBarCacheMaxBars = 1024

// BarCache stores the most recent bars per (symbol, timeframe). Writes
// from bars_historical SEED the slice; writes from bar_update UPSERT,
// replacing the bar at the same t (in-progress bar) or appending new bars
// (capped at MaxBars — oldest bars are dropped FIFO once the ring fills).
type BarCache struct {
	mu      sync.RWMutex
	bars    map[string][]Bar // key: "SYMBOL|TIMEFRAME"
	maxBars int
}

// NewBarCache constructs an empty cache. maxBars <= 0 uses
// DefaultBarCacheMaxBars.
func NewBarCache(maxBars int) *BarCache {
	if maxBars <= 0 {
		maxBars = DefaultBarCacheMaxBars
	}
	return &BarCache{
		bars:    make(map[string][]Bar),
		maxBars: maxBars,
	}
}

// SeedHistorical replaces all cached bars for (symbol, timeframe) with the
// provided slice. Called once per (symbol, timeframe) when the C# AddOn
// emits its initial bars_historical frame after a fresh BarsRequest.
//
// If bars exceeds maxBars, only the tail (most recent maxBars bars) is
// kept. The input is assumed to be ascending by t, matching the protocol
// contract (vltrader_tcp_PROTOCOL.md §6).
func (c *BarCache) SeedHistorical(symbol, timeframe string, bars []Bar) {
	if symbol == "" || timeframe == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(bars) > c.maxBars {
		bars = bars[len(bars)-c.maxBars:]
	}
	// Copy into a new slice to detach from the caller's backing array.
	stored := make([]Bar, len(bars))
	copy(stored, bars)
	c.bars[barKey(symbol, timeframe)] = stored
}

// Upsert merges streaming bar_update bars into the cache. For each input
// bar:
//   - If the t matches the last cached bar's t: REPLACE (in-progress bar
//     update — the current minute's bar is still forming).
//   - If the t is strictly greater than the last cached bar's t: APPEND
//     (a new bar boundary was crossed).
//   - If the t is strictly less than the last cached bar's t: IGNORE
//     (out-of-order — the protocol contract is ascending, but defend
//     against partial misalignment after reconnect).
//
// The ring is bounded: when the slice exceeds maxBars, the oldest bar is
// dropped (FIFO).
//
// The input slice may contain MULTIPLE bars (the NT8 multi-bar gotcha:
// a single tick can update MinIndex..MaxIndex). Each is applied in input
// order via the rules above.
func (c *BarCache) Upsert(symbol, timeframe string, bars []Bar) {
	if symbol == "" || timeframe == "" || len(bars) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	key := barKey(symbol, timeframe)
	existing := c.bars[key]
	for _, b := range bars {
		if len(existing) == 0 {
			existing = append(existing, b)
			continue
		}
		last := existing[len(existing)-1]
		switch {
		case b.T == last.T:
			existing[len(existing)-1] = b
		case b.T > last.T:
			existing = append(existing, b)
			if len(existing) > c.maxBars {
				// Drop the oldest bar; copy the tail forward in place
				// to keep the slice's backing array bounded.
				copy(existing, existing[1:])
				existing = existing[:c.maxBars]
			}
		default:
			// Out-of-order bar — defensive: ignore.
		}
	}
	c.bars[key] = existing
}

// Get returns a SNAPSHOT (defensive copy) of the cached bars for
// (symbol, timeframe). The returned slice is safe for the caller to hold
// or mutate without affecting the cache. Returns nil if no bars are
// cached yet.
//
// The RLock is held only for the duration of the slice copy, which is
// microseconds. Readers never block writers for long.
func (c *BarCache) Get(symbol, timeframe string) []Bar {
	c.mu.RLock()
	defer c.mu.RUnlock()
	src, ok := c.bars[barKey(symbol, timeframe)]
	if !ok || len(src) == 0 {
		return nil
	}
	out := make([]Bar, len(src))
	copy(out, src)
	return out
}

// Count returns the number of cached bars for (symbol, timeframe). Useful
// for tests and metrics.
func (c *BarCache) Count(symbol, timeframe string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.bars[barKey(symbol, timeframe)])
}

// Keys returns the list of currently-populated (symbol, timeframe) pairs.
// Returned in unspecified order. Useful for diagnostics + the Stage 4
// chart relay enumerating its outbound subscriptions.
func (c *BarCache) Keys() [][2]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([][2]string, 0, len(c.bars))
	for k := range c.bars {
		s, t, ok := splitBarKey(k)
		if !ok {
			continue
		}
		out = append(out, [2]string{s, t})
	}
	return out
}

// barKey encodes (symbol, timeframe) as a single map key. We deliberately
// keep the encoding simple (single delimiter) and document it so the
// splitBarKey reverse is unambiguous. Symbols + timeframes are bounded
// strings without "|" so the delimiter is safe.
func barKey(symbol, timeframe string) string {
	return symbol + "|" + timeframe
}

// splitBarKey reverses barKey. Returns ok=false if the format is
// unexpected (defensive — should only be reachable if external code
// mutated the cache directly).
func splitBarKey(k string) (symbol, timeframe string, ok bool) {
	for i := 0; i < len(k); i++ {
		if k[i] == '|' {
			return k[:i], k[i+1:], true
		}
	}
	return "", "", false
}
