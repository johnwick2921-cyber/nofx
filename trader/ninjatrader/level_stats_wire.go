package ninjatrader

import (
	"encoding/json"
	"fmt"
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
// Level-truth wave (2026-08-27) — the job used to evaluate 0 rows on EVERY
// production run because (a) boot-time and 17:05-roll runs race the bot's own
// single-connection DB storm ("database is locked" after busy_timeout), and
// (b) every skip reason was swallowed by a bare `continue`. Skip reasons are
// now logged, and transient errors retry with backoff so the nightly evaluation
// actually lands.
func runLevelStatsDay(st *store.Store, ls *store.LevelStatsStore, traderID string) {
	now := time.Now()
	cur := kernel.CMESessionDayStart(now)
	dayStart := cur.AddDate(0, 0, -1)
	dayKey := dayStart.In(kernel.CTLocation()).Format("2006-01-02")

	// Retry transient store contention (single-connection + boot/roll DB storm).
	// 4 attempts × 15s backoff ≈ 1 min — long enough to ride out a query burst,
	// short enough to never stall the trade loop (this runs in its own goroutine).
	for attempt := 1; ; attempt++ {
		n, err := runLevelStatsDayOnce(st, ls, traderID, dayKey, dayStart.UnixMilli(), cur.UnixMilli(), now.UnixMilli())
		if err == nil {
			logger.Infof("📊 level_stats: %s evaluated %d seated level(s) (total rows %d) — forward validation accumulating", dayKey, n, mustCount(ls))
			return
		}
		logger.Warnf("📊 level_stats: %s attempt %d failed: %v", dayKey, attempt, err)
		if attempt >= 4 {
			logger.Warnf("📊 level_stats: %s giving up after %d attempts — next run at the next session roll", dayKey, attempt)
			return
		}
		time.Sleep(15 * time.Second)
	}
}

// runLevelStatsDayOnce is one evaluation attempt. It returns the number of
// evaluated rows and a descriptive error for every skip reason (no more silent
// continues).
func runLevelStatsDayOnce(st *store.Store, ls *store.LevelStatsStore, traderID, dayKey string, dayStartMs, dayEndMs, nowMs int64) (int, error) {
	barsDB, err := st.BarHistory().BarsBetween("MNQ", "1m", dayStartMs, dayEndMs)
	if err != nil {
		return 0, fmt.Errorf("bars read: %w", err)
	}
	if len(barsDB) == 0 {
		return 0, fmt.Errorf("no persisted 1m bars for the window (forward validation needs the persisted 1m)")
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
		if err != nil {
			return 0, fmt.Errorf("plan list %s/%s: %w", dayKey, sess, err)
		}
		if len(vers) == 0 {
			logger.Infof("📊 level_stats: %s/%s no plan versions — skipped", dayKey, sess)
			continue
		}
		last := vers[len(vers)-1]
		doc := kernel.PlanDoc{}
		if err := json.Unmarshal([]byte(last.Doc), &doc); err != nil {
			return 0, fmt.Errorf("plan doc unmarshal %s/%s v%d: %w", dayKey, sess, last.Version, err)
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
		return 0, fmt.Errorf("upsert: %w", err)
	}
	// T1 (2026-08-26) — touch telemetry feeds level_stats: episode counts per
	// level join for the 2-week verdict (rejections vs acceptances per level).
	if eps, err := st.TouchEpisodes().EpisodeCountByLevel(traderID, dayKey); err == nil && len(eps) > 0 {
		logger.Infof("📊 level_stats: %s fed %d touch episode(s) across %d level(s) (join: touch_episodes)", dayKey, sumEpisodes(eps), len(eps))
	}
	return len(rows), nil
}

// BackfillLevelStats (level-truth wave, 2026-08-27) re-evaluates EVERY day
// that has persisted bars AND plans — the reconstruction the 0-row outage
// missed. Idempotent (Upsert). Returns per-day row counts. DST-safe: each day
// is computed via the same CME session-day arithmetic as the nightly job.
func BackfillLevelStats(st *store.Store, ls *store.LevelStatsStore, traderID string, days []string) (map[string]int, error) {
	if st == nil || ls == nil || traderID == "" {
		return nil, fmt.Errorf("store/trader required")
	}
	out := map[string]int{}
	for _, dayKey := range days {
		d, err := time.ParseInLocation("2006-01-02", dayKey, kernel.CTLocation())
		if err != nil {
			out[dayKey] = -1
			continue
		}
		// Session-day `dayKey` = [dayKey 17:00 CT, dayKey+1 17:00 CT).
		start := d
		end := d.AddDate(0, 0, 1)
		n, err := runLevelStatsDayOnce(st, ls, traderID, dayKey, start.UnixMilli(), end.UnixMilli(), time.Now().UnixMilli())
		if err != nil {
			logger.Warnf("📊 level_stats backfill: %s: %v", dayKey, err)
			out[dayKey] = -1
			continue
		}
		out[dayKey] = n
	}
	return out, nil
}

func mustCount(ls *store.LevelStatsStore) int64 {
	n, _ := ls.Count()
	return n
}

func sumEpisodes(rows []store.EpisodeCountRow) int64 {
	var n int64
	for _, r := range rows {
		n += r.Count
	}
	return n
}
