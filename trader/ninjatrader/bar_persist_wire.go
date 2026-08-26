package ninjatrader

import (
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
			if len(closed) == 0 {
				return
			}
			rows := make([]store.BarHistoryDB, 0, len(closed))
			for _, b := range closed {
				rows = append(rows, store.BarHistoryDB{
					Symbol: symbol, TF: tf, OpenTimeMs: b.T,
					O: b.O, H: b.H, L: b.L, C: b.C, V: b.V,
				})
			}
			if err := bh.InsertBars(rows); err != nil {
				logger.Warnf("bars: persist %s %s failed: %v (never blocks the loop)", symbol, tf, err)
			}
		})
		// Boot backfill + prune loop: the singleton server starts lazily on the
		// first trader load; poll for it briefly, then flush the cache.
		go func() {
			for i := 0; i < 90; i++ {
				server, err := getOrStartTCPServer()
				if err == nil && server != nil && server.BarCache() != nil {
					backfillBars(bh, server)
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
// INSERT OR IGNORE) and logs the spec boot line.
func backfillBars(bh *store.BarHistoryStore, server *ntwire.TCPServer) {
	now := time.Now().UnixMilli()
	total := 0
	for _, pair := range server.BarCache().AllPairs() {
		closed := ntwire.ClosedBarsOnly(server.BarCache().Get(pair[0], pair[1]), pair[1], now)
		rows := make([]store.BarHistoryDB, 0, len(closed))
		for _, b := range closed {
			rows = append(rows, store.BarHistoryDB{Symbol: pair[0], TF: pair[1], OpenTimeMs: b.T,
				O: b.O, H: b.H, L: b.L, C: b.C, V: b.V})
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
}

// pruneLoop runs the retention prune at boot and then daily.
func pruneLoop(bh *store.BarHistoryStore) {
	pruneOnce := func() {
		cutoff := store.RetentionCutoffMs(time.Now())
		if cutoff <= 0 {
			return
		}
		if n, err := bh.PruneOlderThan(cutoff); err != nil {
			logger.Warnf("bars: prune failed: %v", err)
		} else if n > 0 {
			logger.Infof("🧹 bars: pruned %d rows older than %s", n, time.UnixMilli(cutoff).UTC().Format("2006-01-02"))
		}
	}
	pruneOnce()
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	for range t.C {
		pruneOnce()
	}
}
