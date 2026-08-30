// Command vfverify is the SANDBOXED pre-live-fire verification harness
// (P2/P3/P4, 2026-08-30). It never writes to the LIVE DB: every live access
// opens file:...?mode=ro at the SQLite driver level. Writes (P4 test-trader
// row) go ONLY to the sandbox DB copy at /tmp/nofx-vf-db/data.db.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/joho/godotenv"

	"nofx/calendar"
	"nofx/kernel"
	"nofx/market"
	"nofx/store"
)

const (
	liveDBPath    = "/home/hoang/nofx/data/data.db"
	sandboxDBPath = "/tmp/nofx-vf-db/data.db"
	traderID      = "8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265"
	strategyID    = "a5b7662e-7bf7-49bb-9f09-7efa48f95ac8"
	symbol        = "MNQ"
	promptPath    = "/tmp/nofx-vf-p3-prompt.txt"
)

// ── DB layer ────────────────────────────────────────────────────────────────

func openRO(path string) *sql.DB {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	must(err)
	if _, err := db.Exec("PRAGMA query_only=ON"); err != nil {
		fmt.Fprintf(os.Stderr, "query_only pragma failed (continuing on mode=ro): %v\n", err)
	}
	return db
}

func openRW(path string) *sql.DB {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=rw")
	must(err)
	return db
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "FATAL:", err)
		os.Exit(1)
	}
}

func qstr(db *sql.DB, q string, args ...any) string {
	var s string
	if err := db.QueryRow(q, args...).Scan(&s); err != nil {
		return fmt.Sprintf("ERR:%v", err)
	}
	return s
}

func qint(db *sql.DB, q string, args ...any) int64 {
	var n int64
	if err := db.QueryRow(q, args...).Scan(&n); err != nil {
		fmt.Fprintln(os.Stderr, "qint:", err)
		return -1
	}
	return n
}

// ── bars: read 1m rows and aggregate like the live provider ────────────────

type barStore struct {
	db      *sql.DB
	bars1m  []market.Kline
	cutoff  int64 // 0 = no cutoff
	agg     map[string][]market.Kline
	dailyOK bool
}

func newBarStore(db *sql.DB, cutoffMs int64) *barStore {
	bs := &barStore{db: db, cutoff: cutoffMs, agg: map[string][]market.Kline{}}
	rows, err := db.Query(
		`SELECT open_time_ms,o,h,l,c,v FROM bars WHERE symbol=? AND tf='1m' AND (?=0 OR open_time_ms<=?) ORDER BY open_time_ms ASC`,
		symbol, cutoffMs, cutoffMs)
	must(err)
	defer rows.Close()
	for rows.Next() {
		var ms int64
		var k market.Kline
		if err := rows.Scan(&ms, &k.Open, &k.High, &k.Low, &k.Close, &k.Volume); err != nil {
			must(err)
		}
		k.OpenTime = ms
		k.CloseTime = ms + 59_999
		bs.bars1m = append(bs.bars1m, k)
	}
	return bs
}

func (bs *barStore) fetch(tf string, count int) []market.Kline {
	tf = strings.ToLower(strings.TrimSpace(tf))
	var out []market.Kline
	switch tf {
	case "1m":
		out = bs.bars1m
	case "1d", "d":
		out = kernel.DailySessionBars(bs.bars1m)
	default:
		if cached, ok := bs.agg[tf]; ok {
			out = cached
		} else {
			step := map[string]int64{"3m": 3, "5m": 5, "15m": 15, "30m": 30, "1h": 60, "2h": 120, "4h": 240, "6h": 360, "8h": 480, "12h": 720}[tf]
			if step == 0 {
				return nil
			}
			out = kernel.AggregateBars(bs.bars1m, step*60_000)
			bs.agg[tf] = out
		}
	}
	if count > 0 && len(out) > count {
		out = out[len(out)-count:]
	}
	return out
}

// ── live-config replicas (read-only) ───────────────────────────────────────

func loadRegistry(db *sql.DB) kernel.SessionRegistry {
	raw := qstr(db, `SELECT value FROM system_config WHERE key='session_registry'`)
	reg, err := kernel.LoadSessionRegistry(raw)
	must(err)
	return reg
}

func loadStrategy(db *sql.DB) *store.StrategyConfig {
	raw := qstr(db, `SELECT config FROM strategies WHERE id=?`, strategyID)
	var sc store.StrategyConfig
	must(json.Unmarshal([]byte(raw), &sc))
	return &sc
}

// freshness provider — the SAME identity + aging the trader installs
// (trader/auto_trader_dayplan.go installLevelStateProvider).
func makeFreshnessFn(db *sql.DB, now time.Time) func(string, string, kernel.DetectedLevel) string {
	type row struct {
		consumed  bool
		freshness string
		createdAt time.Time
		updatedAt time.Time
	}
	rows, err := db.Query(`SELECT level_key, consumed, freshness, created_at, updated_at FROM level_state`)
	must(err)
	defer rows.Close()
	byKey := map[string]row{}
	for rows.Next() {
		var k string
		var r row
		var ca, ua sql.NullTime
		if err := rows.Scan(&k, &r.consumed, &r.freshness, &ca, &ua); err != nil {
			must(err)
		}
		if ca.Valid {
			r.createdAt = ca.Time
		}
		if ua.Valid {
			r.updatedAt = ua.Time
		}
		byKey[k] = r
	}
	return func(traderID, symbol string, l kernel.DetectedLevel) string {
		key := store.MakeLevelKey(traderID, symbol, kernel.LevelTypeFromLabel(l.Label), "", kernel.LevelBinIndex(l.Price))
		r, ok := byKey[key]
		if !ok {
			return "" // no persisted state → fresh
		}
		return store.AgedFreshness(&store.LevelStateDB{
			Consumed: r.consumed, Freshness: r.freshness,
			CreatedAt: r.createdAt, UpdatedAt: r.updatedAt,
		}, now)
	}
}

// naked POC extras — the SAME store read the trader installs
// (installNakedPOCProvider).
func nakedPOCs(db *sql.DB, bars []market.Kline, now time.Time) []kernel.DetectedLevel {
	rows, err := db.Query(`SELECT session_date, poc FROM session_profiles WHERE symbol=? ORDER BY session_date DESC LIMIT 30`, symbol)
	must(err)
	defer rows.Close()
	var pocs []kernel.PriorPOC
	weeklyBefore := kernel.CMESessionDayKey(now.AddDate(0, 0, -5))
	for rows.Next() {
		var d string
		var p float64
		must(rows.Scan(&d, &p))
		pocs = append(pocs, kernel.PriorPOC{SessionDate: d, POC: p, Weekly: d < weeklyBefore})
	}
	if len(pocs) == 0 {
		return nil
	}
	return kernel.NakedPOCs(pocs, bars, now)
}

// ── shared planner-input assembly (replica of assemblePlannerInputWithCtx) ──

type assembled struct {
	in            kernel.PlannerInput
	machineGrades map[float64]string
	machineLabels map[float64]string
	facts         kernel.PlanFacts
	sideQuota     int
	atr5m         float64
	bars1m        []market.Kline
	price, dATR   float64
	seated        []kernel.ScoredLevel
	pool          []kernel.ScoredLevel
	reg           kernel.SessionRegistry
	dp            *store.DayPlanConfig
	sc            *store.StrategyConfig
	now           time.Time
}

func resolvePlanCfg(dp *store.DayPlanConfig, session string) (int, string, []string) {
	maxLevels := kernel.DefaultMaxLevels
	timeframes := []string{"D", "4h", "1h", "15m"}
	var minGrade string
	if dp == nil {
		return maxLevels, minGrade, timeframes
	}
	if dp.MaxLevels > 0 {
		maxLevels = dp.MaxLevels
	}
	if len(dp.PlannerTimeframes) > 0 {
		timeframes = dp.PlannerTimeframes
	}
	for _, so := range dp.Sessions {
		if so.Session == session && so.MinGrade != nil {
			minGrade = *so.MinGrade
		}
	}
	return maxLevels, minGrade, timeframes
}

// collectMachineGrades — replica of trader/auto_trader_planner.go
// collectMachineGrades (S1-wave A3): seated + pool + HTFZones + HTFZonesFull.
func collectMachineGrades(in kernel.PlannerInput, grades, labels map[float64]string) {
	record := func(price float64, grade string) {
		if grade == "" {
			return
		}
		k := math.Round(price*100) / 100
		if old, ok := grades[k]; !ok || kernel.GradeRank(grade) > kernel.GradeRank(old) {
			grades[k] = grade
		}
	}
	recordLabel := func(price float64, label string) {
		if label == "" {
			return
		}
		k := math.Round(price*100) / 100
		if _, ok := labels[k]; !ok {
			labels[k] = label
		}
	}
	for _, l := range in.Levels {
		record(l.Price, l.Grade)
		recordLabel(l.Price, l.Label)
	}
	for _, pl := range in.Pool {
		record(pl.Price, pl.Grade)
		recordLabel(pl.Price, pl.Label)
	}
	for _, z := range in.HTFZones {
		record(z.Price, z.Grade)
		recordLabel(z.Price, z.Label)
	}
	for _, z := range in.HTFZonesFull {
		record(z.Price, z.Grade)
		recordLabel(z.Price, z.Label)
	}
}

func assemble(db *sql.DB, session, tradeDate string, now time.Time, cutoffMs int64) *assembled {
	sc := loadStrategy(db)
	dp := sc.DayPlan
	maxLevels, minGrade, timeframes := resolvePlanCfg(dp, session)
	proximity := kernel.ResolveProximityK(dp.ProximityFilterATR)

	// cutoffMs=0 → all stored 1m bars (table ends Fri 2026-08-28 15:59 CT).
	// cutoffMs>0 → the P2 write-time replay (bars ≤ plan write instant).
	bs := newBarStore(db, cutoffMs)
	bars1m := bs.bars1m
	reg := loadRegistry(db)

	fetch := func(tf string, count int) []market.Kline { return bs.fetch(tf, count) }
	htfLevels := kernel.DetectHTFLevels(fetch, timeframes, symbol, now)

	extra := append([]kernel.DetectedLevel{}, nakedPOCs(db, bars1m, now)...)
	extra = append(extra, htfLevels...)

	market.FuturesBarsProvider = func(sym, tf string, count int) []market.Kline {
		return bs.fetch(tf, count)
	}
	// kernel.NakedPOCProvider left nil — naked POCs are folded into extra above.

	fresh := makeFreshnessFn(db, now)
	kernel.LevelStateProvider = fresh

	seated, pool, price, dATR := kernel.AssembleScoredLevelsFullMinGrade(
		traderID, bars1m, reg, symbol, maxLevels, now, proximity, minGrade, extra...)
	if dp != nil && dp.Seat1HZoneEnabled() {
		seated = kernel.Seat1HZone(seated, maxLevels)
	}

	var htfZoneScored, htfZonesFull []kernel.ScoredLevel
	var zones []kernel.DetectedLevel
	for _, l := range htfLevels {
		switch l.Kind {
		case kernel.KindSupply, kernel.KindDemand, kernel.KindFVG, kernel.KindOB:
			zones = append(zones, l)
		}
	}
	if len(zones) > 0 && price > 0 && dATR > 0 {
		htfZoneScored = kernel.ScoreLevels(zones, price, dATR, nil, 4, proximity)
		if dp != nil && dp.Seat1HZoneEnabled() {
			htfZoneScored = kernel.Seat1HZone(htfZoneScored, 4)
		}
		htfZonesFull = kernel.ScoreLevels(zones, price, dATR, nil, len(zones), proximity)
	}

	// owner levels (sticky) — none in the live DB today (verified count 0).
	scored := seated
	scored = kernel.FilterLevelsByMinGrade(scored, minGrade)

	daily := fetch("1d", 300)
	hour1 := fetch("1h", 300)
	min5 := fetch("5m", 300)
	min5Long := fetch("5m", 3000)
	rvBaseline, _ := kernel.RVBaselineFrom5m(min5Long, 20, 5)
	priorClose, sessionOpen := kernel.PriorCloseSessionOpen(daily)
	regime := kernel.ComputeRegime(kernel.RegimeInputs{
		Price: price, DailyBars: daily, Hour1Bars: hour1, Min5Bars: min5,
		RVBaseline20d: rvBaseline, PriorClose: priorClose, SessionOpen: sessionOpen,
	})

	var calEvents []kernel.PlannerCalendarEvent
	if raw := qstr(db, `SELECT events_json FROM calendar_slices WHERE trade_date=?`, tradeDate); raw != "" && !strings.HasPrefix(raw, "ERR") {
		var evs []calendar.Event
		if json.Unmarshal([]byte(raw), &evs) == nil {
			loc := kernel.CTLocation()
			if loc == nil {
				loc = time.UTC
			}
			for _, e := range calendar.EventsForSession(evs, session) {
				calEvents = append(calEvents, kernel.PlannerCalendarEvent{
					TimeCT:   e.Time.In(loc).Format("15:04"),
					Currency: e.Currency,
					Title:    e.Title,
					Impact:   string(e.Impact),
				})
			}
		}
	}

	warming := ""
	if n := qint(db, `SELECT COUNT(*) FROM session_profiles WHERE symbol=?`, symbol); n < 10 {
		warming = fmt.Sprintf("session-profile store warming (%d/10)", n)
	}

	// digest chain (P3.6-A): current-date session digests + last-7 dailies.
	var sessionDigests, dailies []string
	if r, err := db.Query(`SELECT text FROM day_plan_digests WHERE trader_id=? AND symbol=? AND trade_date=? AND kind='session' ORDER BY id ASC`, traderID, symbol, tradeDate); err == nil {
		for r.Next() {
			var t string
			r.Scan(&t)
			sessionDigests = append(sessionDigests, t)
		}
		r.Close()
	}
	if r, err := db.Query(`SELECT text FROM day_plan_digests WHERE trader_id=? AND symbol=? AND kind='daily' ORDER BY trade_date DESC LIMIT 7`, traderID, symbol); err == nil {
		for r.Next() {
			var t string
			r.Scan(&t)
			dailies = append(dailies, t)
		}
		r.Close()
	}
	digestChain := kernel.BuildDigestChain(sessionDigests, dailies)

	// structure summary (H9 / G2) — same detector + one line per configured TF.
	structure := structureLines(bars1m, timeframes, now)

	// W11 indicator mirror.
	indicatorsBlock, aiConfigHash := renderMirror(sc.Indicators)

	// G5 consumed levels.
	rule := "2x5m"
	if dp != nil {
		rule = dp.AcceptanceRuleFor(session)
	}
	var consumedLines []string
	if len(bars1m) > 0 {
		lvls := make([]kernel.PlanLevel, len(scored))
		for i, s := range scored {
			lvls[i] = kernel.PlanLevel{Price: s.Price, Label: s.Label}
		}
		for _, s := range scored {
			if !kernel.EvaluateLevelFacts(bars1m, s.Price, kernel.DirAbove, rule, 3, now.UnixMilli()).StillValid {
				consumedLines = append(consumedLines, fmt.Sprintf("%.2f %s", s.Price, s.Label))
			}
		}
	}

	bcFacts := kernel.ComputeBiasContext(bars1m, scored, now)
	kernel.ApplyUniverseDayAnchors(&bcFacts, kernel.ExtractMultiDayLevels(bars1m, reg, now))

	candleTables := ""
	if kernel.PlannerCandlesEnabled() {
		candleTables = kernel.BuildPlannerCandleTables(bars1m)
	}
	weeklyCtx := kernel.WeeklyContextLine(nil, 0) // no WEEKLY doc until Sun 16:30 CT

	freshFVGs := kernel.FreshFvgCandidates(bars1m, symbol, now)

	sideQuota := kernel.SideQuotaFromEnv()
	if dp != nil {
		if q := dp.MinSideLevelsFor(session); q > 0 {
			sideQuota = q
		}
	}

	in := kernel.PlannerInput{
		TradeDate:        tradeDate,
		Session:          session,
		Now:              now,
		ReadKind:         session + " scheduled read (stored+cached data)",
		Price:            price,
		DATR:             dATR,
		Regime:           regime,
		Levels:           scored,
		Pool:             pool,
		HTFZones:         htfZoneScored,
		HTFZonesFull:     htfZonesFull,
		StructureSummary: structure,
		ConsumedLevels:   consumedLines,
		FreshFVGs:        freshFVGs,
		Calendar:         calEvents,
		DigestChain:      digestChain,
		Warming:          warming,
		IndicatorsBlock:  indicatorsBlock,
		AIConfigHash:     aiConfigHash,
		MaxLevels:        maxLevels,
		ScenarioCap:      scDayPlanScenarioCap(dp),
		BiasCtx:          bcFacts.Line(),
		BiasCtxFacts:     bcFacts,
		CandleTables:     candleTables,
		WeeklyCtx:        weeklyCtx,
	}

	machineGrades := map[float64]string{}
	machineLabels := map[float64]string{}
	collectMachineGrades(in, machineGrades, machineLabels)

	facts := kernel.PlanFacts{Price: in.Price, DATR: in.DATR}
	for _, l := range in.Levels {
		switch l.Kind {
		case kernel.KindPDH:
			facts.PDH = l.Price
		case kernel.KindPDL:
			facts.PDL = l.Price
		}
	}
	if facts.PDH <= 0 {
		facts.PDH = in.BiasCtxFacts.PDH
	}
	if facts.PDL <= 0 {
		facts.PDL = in.BiasCtxFacts.PDL
	}

	return &assembled{
		in: in, machineGrades: machineGrades, machineLabels: machineLabels,
		facts: facts, sideQuota: sideQuota, atr5m: kernel.StaleConfirmATR5m(bars1m),
		bars1m: bars1m, price: price, dATR: dATR, seated: seated, pool: pool,
		reg: reg, dp: dp, sc: sc, now: now,
	}
}

func scDayPlanScenarioCap(dp *store.DayPlanConfig) int {
	if dp != nil && dp.ScenarioCap >= 1 && dp.ScenarioCap <= 5 {
		return dp.ScenarioCap
	}
	return 3
}

// structureLines — replica of trader structureSummaryLines (G2, honest H9).
func structureLines(bars1m []market.Kline, timeframes []string, now time.Time) []string {
	lines := make([]string, 0, len(timeframes)+1)
	snap := kernel.StructureSnapshot(bars1m, now.UnixMilli())
	for _, tf := range timeframes {
		st, ok := snap[tf]
		if !ok {
			lines = append(lines, tf+": unavailable")
			continue
		}
		label := st.Trend
		if st.Swing != nil {
			label += fmt.Sprintf(" (%s %.2f @%s)", st.Swing.Kind, st.Swing.Price, kernel.ClockCT(time.UnixMilli(st.Swing.TimeMs)))
		}
		lines = append(lines, tf+": "+label)
	}
	var lastEv *kernel.StructureEvent
	var lastTF string
	for _, tf := range kernel.StructureTFs {
		st, ok := snap[tf]
		if !ok {
			continue
		}
		for i := range st.LastEvents {
			if lastEv == nil || st.LastEvents[i].TimeMs > lastEv.TimeMs {
				e := st.LastEvents[i]
				lastEv = &e
				lastTF = tf
			}
		}
	}
	if lastEv != nil {
		lines = append(lines, fmt.Sprintf("last event: %s-%s %s @%s", lastEv.Type, lastEv.Dir, lastTF, kernel.ClockCT(time.UnixMilli(lastEv.TimeMs))))
	}
	return lines
}

// renderMirror — replica of trader renderIndicatorMirror (W11).
func renderMirror(ic store.IndicatorConfig) (string, string) {
	hash := ic.AIConfigFingerprint()
	tfs := append([]string{}, ic.Klines.SelectedTimeframes...)
	primary := ic.Klines.PrimaryTimeframe
	if len(tfs) == 0 {
		if primary != "" {
			tfs = append(tfs, primary)
		}
		if ic.Klines.LongerTimeframe != "" {
			tfs = append(tfs, ic.Klines.LongerTimeframe)
		}
	}
	if len(tfs) == 0 {
		return "", hash
	}
	if primary == "" {
		primary = tfs[0]
	}
	count := ic.Klines.PrimaryCount
	if count <= 0 {
		count = 30
	}
	indPeriods := market.IndicatorPeriods{
		EMA: ic.EMAPeriods, RSI: ic.RSIPeriods, ATR: ic.ATRPeriods, BOLL: ic.BOLLPeriods,
	}
	mkt, err := market.GetWithTimeframes(symbol, tfs, primary, count, indPeriods)
	if err != nil || mkt == nil {
		return "", hash
	}
	return kernel.RenderPlannerIndicatorBlock(mkt, ic, tfs), hash
}

func main() {
	_ = godotenv.Load("/home/hoang/nofx/.env") // live env knobs (DATA_ENCRYPTION_KEY etc.) — read-only
	sub := "p3"
	if len(os.Args) > 1 {
		sub = os.Args[1]
	}
	switch sub {
	case "p2":
		runP2()
	case "p3":
		runP3()
	case "p4":
		runP4()
	default:
		fmt.Fprintln(os.Stderr, "usage: vfverify <p2|p3|p4>")
		os.Exit(2)
	}
}
