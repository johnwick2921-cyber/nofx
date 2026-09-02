package ninjatrader

import (
	"nofx/market"
	"sync"
	"time"

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
