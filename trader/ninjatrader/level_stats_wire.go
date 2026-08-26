package ninjatrader

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"nofx/kernel"
	"nofx/logger"
	"nofx/market"
	"nofx/store"
)

// B4 — LEVEL_STATS nightly evaluation (forward-validation table, Pack B owner
// override 2026-08-26). At each CME session roll (17:00 CT, + boot) this job
// evaluates EVERY level the PREVIOUS session-day's plans seated against the
// day's persisted 1m bars: TOUCHED (±4pts) / REACTED (≥8pt in 3 bars) /
// BROKE-CLEAN / CHOPPED. The accumulated rows are the forward validation the
// waived backward replay was replaced with — the 2-week verdict on the volume
// family's weights reads store.LevelStats().AggregateByGrade/Family.

var levelStatsOnce sync.Once

// WireLevelStatsNightly starts the per-session-day evaluation loop. Idempotent
// (once); nil-safe; never blocks the trade loop (own goroutine, own errors).
// traderID is the owning AutoTrader's id (the TCP trader's own field is empty
// until StartCloseSync runs later).
func WireLevelStatsNightly(st *store.Store, traderID string) {
	if st == nil || traderID == "" {
		return
	}
	levelStatsOnce.Do(func() {
		ls := st.LevelStats()
		if err := ls.Migrate(); err != nil {
			logger.Warnf("level_stats: migrate failed: %v", err)
			return
		}
		go func() {
			runLevelStatsDay(st, ls, traderID)
			for {
				// Next 17:05 CT boundary (the daily roll + 5m settling time).
				next := kernel.NextSessionRollCT(time.Now()).Add(5 * time.Minute)
				time.Sleep(time.Until(next))
				runLevelStatsDay(st, ls, traderID)
			}
		}()
	})
}

// runLevelStatsDay evaluates the PREVIOUS CME session-day (17:00→17:00 CT).
func runLevelStatsDay(st *store.Store, ls *store.LevelStatsStore, traderID string) {
	now := time.Now()
	cur := kernel.CMESessionDayStart(now)
	dayStart := cur.AddDate(0, 0, -1)
	dayKey := dayStart.In(kernel.CTLocation()).Format("2006-01-02")
	dayStartMs, dayEndMs := dayStart.UnixMilli(), cur.UnixMilli()

	barsDB, err := st.BarHistory().BarsBetween("MNQ", "1m", dayStartMs, dayEndMs)
	if err != nil || len(barsDB) == 0 {
		logger.Infof("📊 level_stats: %s no bars — skipped (forward validation needs the persisted 1m)", dayKey)
		return
	}
	klines := make([]market.Kline, 0, len(barsDB))
	for _, b := range barsDB {
		klines = append(klines, market.Kline{
			OpenTime: b.OpenTimeMs, Open: b.O, High: b.H, Low: b.L, Close: b.C, Volume: b.V,
		})
	}

	// Latest plan per session for the previous day (trader-scoped).
	var rows []store.LevelStatsDB
	seen := map[string]bool{}
	for _, sess := range []string{"NY", "ASIA", "LONDON"} {
		vers, err := st.Plan().ListVersionsForTrader(dayKey, sess, traderID)
		if err != nil || len(vers) == 0 {
			continue
		}
		last := vers[len(vers)-1]
		doc := kernel.PlanDoc{}
		if err := json.Unmarshal([]byte(last.Doc), &doc); err != nil {
			continue
		}
		for _, l := range doc.Levels {
			key := dayKey + "|" + strconv.FormatFloat(l.Price, 'f', 2, 64) + "|" + l.Label
			if seen[key] {
				continue
			}
			seen[key] = true
			kind := kernel.KindForLabel(l.Label)
			out := kernel.EvaluateLevelOutcome(klines, l.Price, 0, 0)
			rows = append(rows, store.LevelStatsDB{
				TraderID: traderID, SessionDay: dayKey,
				Price: l.Price, Label: l.Label,
				Kind: string(kind), Grade: l.Grade,
				Role: string(kernel.RoleForLabel(l.Label)), Family: kernel.FamilyFor(kind),
				Touched: out.Touched, Reacted: out.Reacted, BrokeClean: out.BrokeClean, Chopped: out.Chopped,
			})
		}
	}
	if err := ls.UpsertStats(rows); err != nil {
		logger.Warnf("level_stats: %s upsert failed: %v", dayKey, err)
		return
	}
	n, _ := ls.Count()
	logger.Infof("📊 level_stats: %s evaluated %d seated level(s) (total rows %d) — forward validation accumulating", dayKey, len(rows), n)
}
