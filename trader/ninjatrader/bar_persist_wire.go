package ninjatrader

import (
	"nofx/market"
	"sync"
	"sync/atomic"
	"time"

	"nofx/kernel"
	"nofx/logger"
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// WireBarPersistence (2026-08-26) — installs the closed-bar writer on the TCP
// BarCache, flushes the boot backfill (~33h in-memory history), and arms the
// nightly retention prune. Idempotent (sync.Once): every trader's Start calls
// it, only the first wires.
var wireBarPersistenceOnce sync.Once

// WireBarPersistence attaches the store to the bar feed. st == nil → no-op.
func WireBarPersistence(st *store.Store) {
	if st == nil {
		return
	}
	wireBarPersistenceOnce.Do(func() {
		bh := st.BarHistory()
		if err := bh.Migrate(); err != nil {
			logger.Warnf("bars: migrate failed: %v (persistence disabled)", err)
			return
		}
		ntwire.SetBarPersister(func(historical bool, symbol, tf string, bars []ntwire.Bar) {
			closed := ntwire.ClosedBarsOnly(bars, tf, time.Now().UnixMilli())
			if historical {
				// BAR-TRUTH 2026-08-28: replay frames arrive CLOSE-stamped -
				// apply the cache's open-stamp conversion (the 2499/2500 mismatch root cause).
				closed = ntwire.OpenStampBars(closed, tf)
			}
			if len(closed) == 0 {
				return
			}
			rows := make([]store.BarHistoryDB, 0, len(closed))
			for _, b := range closed {
				rows = append(rows, store.BarHistoryDB{
					Symbol: symbol, TF: tf, OpenTimeMs: b.T,
					O: b.O, H: b.H, L: b.L, C: b.C, V: b.V,
					Convention: market.StampConvention(tf),
				})
			}
			if err := bh.InsertBars(rows); err != nil {
				logger.Warnf("bars: persist %s %s failed: %v (never blocks the loop)", symbol, tf, err)
			}
		})
		// Boot backfill + prune loop: the singleton server starts lazily on the
		// first trader load; poll for it briefly, then flush the cache. The
		// AddOn's bars_historical replay lands a few seconds after our restart,
		// so retry while the flush stays empty.
		go func() {
			for i := 0; i < 90; i++ {
				server, err := getOrStartTCPServer()
				if err == nil && server != nil && server.BarCache() != nil {
					backfilled := backfillBars(bh, server)
					if backfilled == 0 {
						// Cache still empty (replay in flight) — retry a few
						// times; the live persister catches bars regardless.
						for r := 0; r < 20 && backfilled == 0; r++ {
							time.Sleep(15 * time.Second)
							backfilled = backfillBars(bh, server)
						}
					}
					// BARS HORIZON (2026-09-09) — the replay has landed, so the
					// EMPTY arm may speak, and the depth line can report what
					// the ring ACTUALLY holds rather than a cold cache.
					noteBarHorizonBackfillLanded()
					// D3 — REHYDRATE THE RING FROM THE STORE, alongside the
					// AddOn's seed. `now` is taken HERE, at the boot entry
					// point, and handed down (A28).
					//
					// ORDERING (fixed in review, 2026-09-09): this runs BEFORE
					// the afterBackfillHook, because that hook prints the
					// "📊 bars after backfill" line AND the BarResolver behind it
					// picks nt8 vs nt8_agg vs own1m from what the cache can
					// reach. Running the rehydrate afterwards left the
					// best-known boot line describing a ring the bot no longer
					// had — two lines about the same instant, and the older,
					// more-read one wrong.
					rehydrateRingFromStore(bh, server, time.Now())
					// R1 (2026-09-02) — the boot 📊 bars line ran before this
					// replay landed, so it reported own1m for every TF on a
					// cold cache. Now that the pantry is in, say what the
					// resolver can ACTUALLY reach.
					if h := afterBackfillHook.Load(); h != nil {
						if fn, ok := h.(func()); ok && fn != nil {
							fn()
						}
					}
					logger.Infof("%s", barHorizonBootLine(server.BarCache(), time.Now()))
					go pruneLoop(bh)
					return
				}
				time.Sleep(time.Second)
			}
			logger.Warnf("bars: TCP server never came up — boot backfill skipped")
		}()
	})
}

// backfillBars flushes every closed bar the cache already holds (idempotent —
// INSERT OR IGNORE) and logs the spec boot line. Returns the flushed count so
// the caller can retry while the AddOn's bars_historical replay is still
// arriving (the cache is empty for the first seconds after a Go restart).
func backfillBars(bh *store.BarHistoryStore, server *ntwire.TCPServer) int {
	now := time.Now().UnixMilli()
	total := 0
	for _, pair := range server.BarCache().AllPairs() {
		closed := ntwire.ClosedBarsOnly(server.BarCache().Get(pair[0], pair[1]), pair[1], now)
		rows := make([]store.BarHistoryDB, 0, len(closed))
		for _, b := range closed {
			rows = append(rows, store.BarHistoryDB{Symbol: pair[0], TF: pair[1], OpenTimeMs: b.T,
				O: b.O, H: b.H, L: b.L, C: b.C, V: b.V, Convention: market.StampConvention(pair[1])})
		}
		if len(rows) > 0 {
			if err := bh.InsertBars(rows); err != nil {
				logger.Warnf("bars: backfill %s %s failed: %v", pair[0], pair[1], err)
				continue
			}
			total += len(rows)
		}
	}
	pairs, _ := bh.SymbolTFCount()
	count, _ := bh.Count()
	logger.Infof("📦 bars: persisting %d symbol×tf retention=%dd rows=%d (backfilled %d)",
		pairs, store.BarRetentionDays(), count, total)
	return total
}

// pruneLoop runs the retention prune + the NIGHTLY INTEGRITY CHECK (F5,
// 2026-08-27) at boot and then daily: duplicate natural-key groups must be 0
// and only tf='1m' may be stored (aggregates derive on read). WARN on drift.
func pruneLoop(bh *store.BarHistoryStore) {
	integrityCheck := func() {
		dups, tfs, total, err := bh.BarsIntegrity()
		if err != nil {
			logger.Warnf("bars: integrity check failed: %v", err)
			return
		}
		if dups > 0 {
			logger.Warnf("🚨 bars integrity DRIFT: dups=%d tfs=%v total=%d (expected dups=0) — replay/calibration readers must not trust stored aggregates", dups, tfs, total)
			return
		}
		logger.Infof("✅ bars integrity OK: dups=0 tfs=%v total=%d", tfs, total)
	}
	pruneOnce := func() {
		// BAR-SOURCE WAVE 2026-09-02 — retention is PER TF. The old single
		// cutoff was TF-blind and would have deleted the 383 weekly bars back
		// to 2019 on the first nightly prune after they were persisted.
		byTF, err := bh.PruneByTF(time.Now())
		if err != nil {
			logger.Warnf("bars: prune failed: %v", err)
			return
		}
		for tf, n := range byTF {
			logger.Infof("🧹 bars: pruned %d %s rows older than %dd (per-TF retention)", n, tf, store.RetentionDaysFor(tf))
		}
	}
	pruneOnce()
	integrityCheck()
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	for range t.C {
		pruneOnce()
		integrityCheck()
	}
}

// afterBackfillHook lets the trader layer print its post-backfill bar-source
// line without this package importing it. nil = nothing printed.
var afterBackfillHook atomic.Value

// SetAfterBackfillHook installs the callback fired once the first backfill
// completes. Safe to call more than once; the last registration wins.
func SetAfterBackfillHook(fn func()) { afterBackfillHook.Store(fn) }

// ── D3 — THE RING REHYDRATES FROM THE STORE ON BOOT ─────────────────────────
// (owner-authorised expansion, wave BARS HORIZON 2026-09-09)
//
// A Go restart drops the ring to the AddOn's 2000-bar seed
// (defaultAutoBarsBack) while the persisted `bars` table holds 21 days
// (measured 2026-09-09: MNQ 1m 20,043 rows back to 2026-08-19 10:00 CT). The
// ring climbs back to 2500 across a session because SeedHistorical MERGES —
// and then the next restart shortens the horizon again. Nothing ever
// rehydrated it.
//
// This runs ALONGSIDE the AddOn's seed, immediately after the replay lands, and
// it deliberately touches nothing else: not the AddOn, not the subscription,
// not the backfill.
//
// EVERY CONSTRAINT IS ENFORCED BY RehydrateOlder, not here:
//   - it MERGES (mergeBarsByTime, the same discipline SeedHistorical uses);
//   - it never replaces a live bar with a stale one (older-only, existing wins);
//   - it is bounded by the ring's own maxBars;
//   - it is a NO-OP on a cold key, so a dead feed can never be made to look alive.
//
// 1m ONLY — OWNER CONDITION (a), RULING 2026-09-09 18:18 CT.
//
//	"RULING on D3: the regime input MAY change. Rehydrating the ring from the
//	 store changes what RVBaseline is fed — and what it is fed today is 41
//	 hours labelled as 20 days. A regime value computed on a shorter window
//	 than its name is the defect; correcting the window is not a scope
//	 violation, it is the fix. A31 forbids changing the RULE, not correcting
//	 the INPUT the rule was promised."
//
// The first cut rehydrated EVERY (symbol, timeframe) pair the cache held. That
// deepened the 5m ring from the AddOn's ~2000-bar seed to the 2500 cap, and
// min5Long (auto_trader_planner.go) was the ONE production ring read above
// 2000 — so it moved a REGIME input using NT8's own 5m aggregates. MEASURED on
// the live store, MNQ, kernel.RVBaselineFrom5mDays(·, 20, 5):
//
//	stored 5m, newest 2000 rows → 0.876371 over 7 complete session-days
//	stored 5m, newest 2500 rows → 0.884746 over 9 complete session-days
//	1m tape 12000 → agg 5m 2400 → 0.893543 over 9 complete session-days
//
// The last two cover the SAME nine days and disagree by 0.9%: the stored 5m
// rows do not agree with their own 1m constituents (store/bar_history.go's
// migration once deleted every non-1m row for exactly that reason). So the
// rehydrate takes the 1m rows ONLY — the feed's own closed bars, the tape every
// other series is aggregated FROM — and the deeper regime window is served from
// that 1m tail instead (owner condition (b): ResolveRVBaselineTape,
// trader/regime_input_window.go). The non-1m rings keep exactly what the AddOn
// seeded, exactly as before the wave.
//
// THIS WAVE DOES MOVE COMPUTED VALUES, AND THE REPORT NAMES THEM. Claiming
// otherwise while the boot line shows a change is class 82. What moves:
// RVBaseline / RVBaselineDays (7 → 9 complete session-days), and — through the
// D2 1m store splice, not through this rehydrate — CompletedWeekCount (0 → 2)
// and WeeklyShadowRefs (1 → 3). The RULE (kernel.RVBaselineFrom5mDays) is
// byte-for-byte unchanged and pinned so by an E7 golden.
//
// A10: a failed store read WARNs and boot continues. Nothing here gates,
// refuses, blocks or blanks.
func rehydrateRingFromStore(bh *store.BarHistoryStore, server *ntwire.TCPServer, now time.Time) {
	if bh == nil || server == nil || server.BarCache() == nil {
		logger.Warnf("🧯 ring rehydrate SKIPPED: store=%v server=%v — the ring keeps whatever the AddOn seeded",
			bh != nil, server != nil)
		return
	}
	cache := server.BarCache()
	pairs := cache.AllPairs()
	if len(pairs) == 0 {
		logger.Warnf("🧯 ring rehydrate SKIPPED: the cache holds 0 symbol×tf pairs — a COLD ring is never filled from the store (the store deepens a live tape, it never substitutes for one)")
		return
	}
	selected := pairsToRehydrate(pairs)
	skipped := len(pairs) - len(selected)
	totalAdded, deepened, failed := 0, 0, 0
	for _, pair := range selected {
		symbol, tf := pair[0], pair[1]
		before := cache.Count(symbol, tf)
		rows, err := bh.LastNBars(symbol, tf, cache.MaxBars())
		if err != nil {
			failed++
			logger.Warnf("🧯 ring rehydrate %s %s FAILED: %v — boot continues on the AddOn seed alone (%d bars)", symbol, tf, err, before)
			continue
		}
		if len(rows) == 0 {
			continue
		}
		bars := make([]ntwire.Bar, 0, len(rows))
		for _, r := range rows {
			bars = append(bars, ntwire.Bar{T: r.OpenTimeMs, O: r.O, H: r.H, L: r.L, C: r.C, V: r.V})
		}
		added := cache.RehydrateOlder(symbol, tf, bars)
		if added == 0 {
			continue
		}
		totalAdded += added
		deepened++
		after := cache.Get(symbol, tf)
		h := kernel.HorizonOf(barsToKlines(after, tf), tf, cache.MaxBars(), now)
		logger.Infof("🧯 ring rehydrated %s %s: %d → %d bars (+%d older from the store, cap %d) · %s",
			symbol, tf, before, len(after), added, cache.MaxBars(), h.Line())
	}
	logger.Infof("🧯 ring rehydrate done: %d of %d symbol×tf pairs deepened, +%d bars total, %d read failure(s), %d pair(s) SKIPPED as tf!=%s (stored non-1m rows are NT8 aggregates — never fed to a live regime input) · store retention %s=%dd (the ring is the cache; the store is the horizon)",
		deepened, len(pairs), totalAdded, failed, skipped, rehydrateTimeframe, rehydrateTimeframe, store.RetentionDaysFor(rehydrateTimeframe))
}

// rehydrateTimeframe is the ONLY timeframe the boot rehydrate touches. See the
// header above for why it is not every pair the cache holds.
const rehydrateTimeframe = "1m"

// pairsToRehydrate is THE selection (owner condition (a), 2026-09-09),
// extracted so a pin drives IT rather than a copy of it (class 86): only the
// 1m pairs are rehydrated, and the order the cache handed us is preserved so
// the log line's counts are reproducible.
//
// A pin that only checked the constant still exists would pass a mutation that
// disabled the filter, so the filter is a function with its own fixture.
func pairsToRehydrate(pairs [][2]string) [][2]string {
	out := make([][2]string, 0, len(pairs))
	for _, p := range pairs {
		if p[1] == rehydrateTimeframe {
			out = append(out, p)
		}
	}
	return out
}
