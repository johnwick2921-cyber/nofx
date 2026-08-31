package trader

// WEEKLY-BIAS WAVE (2026-08-30) — W2/W4/W5 trader layer: the Sunday weekly
// read (planner seat), the mid-week invalidation watch, and the WARN/shadow
// annotators. ALL of it rides the EXISTING cycle — no new loop, no new gate.
//
// LAW (W5.4): nothing here changes seating, grades, gates or sizes. Every
// shadow artifact is a LOG or a telemetry counter — the real system is frozen.
//
// W2 storage contract: one plans-table row per governed week with
// session='WEEKLY' and trade_date = the week's Monday; the doc JSON carries
// facts_hash (sha256 of the rendered facts sections) — no new DB columns.

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"nofx/kernel"
	"nofx/market"
	"nofx/store"
	"nofx/telemetry"
)

const weeklySystemPrompt = "You are a disciplined CME index-futures weekly-bias reasoner. Output ONLY the single JSON object requested — no prose outside the JSON."

// weeklyState is the per-trader weekly machinery state (all touched from the
// run-loop goroutine except the cache read from the monitor goroutine — the
// mutex keeps the two honest).
type weeklyState struct {
	mu             sync.Mutex
	doc            *kernel.WeeklyDoc // cached weekly doc
	docKey         string            // governing Monday of the cached doc
	loaded         bool
	skipLogDate    string          // dedupes the "skip-fresh" line per week
	shadowLog      map[string]bool // "tradeDate|session|roundedPx" → logged (W5.1 dedupe)
	shadowSeatLog  map[string]bool // "tradeDate|session" → reorder line logged
	invalidatedLog map[string]bool // "monday" → invalidation line logged
}

func (w *weeklyState) resetWeek(key string) {
	if w.docKey != key {
		w.docKey, w.doc, w.loaded = key, nil, false
	}
	if w.shadowLog == nil {
		w.shadowLog = map[string]bool{}
	}
	if w.shadowSeatLog == nil {
		w.shadowSeatLog = map[string]bool{}
	}
	if w.invalidatedLog == nil {
		w.invalidatedLog = map[string]bool{}
	}
}

// weeklyReadClaim dedupes the async Sunday read across traders/restarts the
// same way the planner-read claim does (one AI call per week per trader).
var weeklyReadClaim sync.Map // "weekly:<traderID>:<monday>" → struct{}

func claimWeeklyRead(key string) bool {
	_, loaded := weeklyReadClaim.LoadOrStore(key, struct{}{})
	return !loaded
}
func releaseWeeklyRead(key string) { weeklyReadClaim.Delete(key) }

// weeklyBars1m loads the STORED 1m bars for the trader's futures symbol (the
// same table the W1 computations were built for). Empty on a cold store —
// the read then fails loud, never fakes.
func (at *AutoTrader) weeklyBars1m(now time.Time) []market.Kline {
	if at.store == nil {
		return nil
	}
	rows, err := at.store.BarHistory().BarsBetween(at.futuresSymbol(), "1m", 0, now.UnixMilli())
	if err != nil {
		at.logErrorf("📅 WEEKLY READ: stored 1m bars load failed: %v", err)
		return nil
	}
	bars := make([]market.Kline, 0, len(rows))
	for _, r := range rows {
		bars = append(bars, market.Kline{
			OpenTime: r.OpenTimeMs, Open: r.O, High: r.H, Low: r.L, Close: r.C, Volume: r.V,
		})
	}
	return bars
}

// weeklyReadVerdict is the pure W2 decision (fixture-tested): before the
// week's read deadline → "wait"; at/after it with no stored doc → "read"
// (covers the boot-backfill case — a Monday boot is past the deadline);
// with a stored doc → "skip" (idempotent — never re-run).
func weeklyReadVerdict(now, deadline time.Time, hasDoc bool) string {
	if now.Before(deadline) {
		return "wait"
	}
	if hasDoc {
		return "skip"
	}
	return "read"
}

// maybeRunWeeklyRead is the W2 scheduler, wired into the EXISTING cycle above
// the session gate (same hoist as maybeFetchCalendar): the Sunday read needs
// no market, no NT8, no account. One AI read per week, idempotent — a stored
// WEEKLY doc for the governed week means never re-run. Boot-backfill: a boot
// AFTER this week's read time with no doc runs exactly once.
func (at *AutoTrader) maybeRunWeeklyRead(now time.Time) {
	if !at.dayPlanEnabled() || at.store == nil || at.exchange != "ninjatrader" {
		return
	}
	monday := kernel.WeekGoverningMonday(now).Format("2006-01-02")
	deadline := kernel.WeeklyReadDeadline(now)
	existing, err := at.store.Plan().GetLatestPlanForTraderSession(monday, "WEEKLY", at.id)
	if err != nil {
		at.logErrorf("📅 WEEKLY READ: doc lookup failed for week %s: %v", monday, err)
		return
	}
	switch weeklyReadVerdict(now, deadline, existing != nil) {
	case "wait":
		return // before this week's read time — nothing to do
	case "skip":
		at.weeklyState.mu.Lock()
		first := at.weeklyState.skipLogDate != monday
		if first {
			at.weeklyState.skipLogDate = monday
		}
		at.weeklyState.mu.Unlock()
		if first {
			at.logInfof("📅 WEEKLY READ skip-fresh — week %s doc already stored (v%d), idempotent.", monday, existing.Version)
		}
		return
	}
	bootBackfill := now.Sub(deadline) > 2*time.Hour // read time long past → this boot caught up
	key := fmt.Sprintf("weekly:%s:%s", at.id, monday)
	if !claimWeeklyRead(key) {
		at.logInfof("📅 WEEKLY READ already in flight for week %s — skipping duplicate call.", monday)
		return
	}
	at.logInfof("📅 WEEKLY READ starting for week %s (boot_backfill=%v)", monday, bootBackfill)
	go func() {
		defer releaseWeeklyRead(key)
		at.runWeeklyRead(now, monday, bootBackfill)
	}()
}

// runWeeklyRead performs ONE read attempt pair (initial + one retry with the
// reject reason appended) and persists the validated doc. The read must never
// crash the trader loop: panic recovery + loud failure logs.
func (at *AutoTrader) runWeeklyRead(now time.Time, monday string, bootBackfill bool) {
	defer func() {
		if r := recover(); r != nil {
			at.logErrorf("⚠️ WEEKLY READ panic recovered for %s: %v", monday, r)
		}
	}()
	bars := at.weeklyBars1m(now)
	if len(bars) == 0 {
		at.logErrorf("⚠️ WEEKLY READ FAILED for %s: no stored 1m bars (thin/cold store)", monday)
		return
	}
	price := bars[len(bars)-1].Close
	facts := kernel.ComputeWeeklyFacts(bars, now, price)
	prompt := kernel.BuildWeeklyPrompt(facts)

	client, modelID := at.resolvePlannerClient()
	if client == nil {
		at.logErrorf("⚠️ WEEKLY READ FAILED for %s: no AI client resolved", monday)
		return
	}
	var lastErr string
	for attempt := 1; attempt <= 2; attempt++ { // initial + ONE retry (spec W2)
		raw, err := client.CallWithMessages(weeklySystemPrompt, prompt)
		if err != nil {
			lastErr = err.Error()
			at.logWarnf("📅 WEEKLY READ attempt %d/2 failed for %s: %v", attempt, monday, err)
			continue
		}
		doc, perr := kernel.ParseWeeklyDoc(raw)
		if perr != nil {
			lastErr = perr.Error()
			at.logWarnf("📅 WEEKLY READ attempt %d/2 parse rejected for %s: %v", attempt, monday, perr)
			continue
		}
		if reason := kernel.ValidateWeeklyDoc(doc, kernel.WeeklyRefSet(facts), facts.ThinHistory); reason != "" {
			lastErr = reason
			at.logWarnf("📅 WEEKLY READ attempt %d/2 validator rejected for %s: %s — retrying once with the reason.", attempt, monday, reason)
			prompt = prompt + "\n\nREJECTED by the validator — fix ONLY these and answer again:\n" + reason
			continue
		}
		// Accepted — stamp the audit fields and write the plan row.
		doc.FactsHash = facts.FactsHash
		doc.ThinHistory = facts.ThinHistory
		// F5 DOA guard (2026-08-30): if the bias's own invalidation basis is
		// ALREADY crossed at write, stamp neutral now — never write a stillborn
		// bias the watch kills seconds later (the 17:07:15 bear lived 3s and
		// the invalidated bear was RIGHT by 250pt).
		var doaBars []market.Kline
		if market.FuturesBarsProvider != nil {
			tf := kernel.WeeklyInvalidationBasisTF(doc.Invalidation.Basis)
			if tf == "" {
				tf = kernel.WeeklyInvalidationTFDefault()
			}
			doaBars = market.FuturesBarsProvider(at.futuresSymbol(), tf, kernel.AISVPBarCount)
		}
		if kernel.ApplyWeeklyDOA(doc, doaBars, time.Now()) {
			at.logWarnf("📅 WEEKLY READ %s stamped NEUTRAL AT WRITE (F5 DOA) — invalidation %.2f already crossed by a closed %s bar", monday, doc.Invalidation.Px, kernel.WeeklyInvalidationBasisTF(doc.Invalidation.Basis))
		}
		docJSON, _ := json.Marshal(doc)
		trigger := "sunday_weekly_read"
		if bootBackfill {
			trigger = "weekly_boot_backfill"
		}
		version, werr := at.store.Plan().AppendPlan(&store.PlanDB{
			PlanID:        at.store.Plan().ResolvePlanID(monday, "WEEKLY", at.id),
			StrategyID:    at.id,
			TradeDate:     monday,
			Session:       "WEEKLY",
			TriggerReason: trigger,
			Lifecycle:     "active",
			ModelID:       modelID,
			PromptHash:    facts.FactsHash,
			Doc:           string(docJSON),
		})
		if werr != nil {
			at.logErrorf("⚠️ WEEKLY READ FAILED for %s: plan row write: %v", monday, werr)
			return
		}
		at.logInfof("📅 WEEKLY READ written %s v%d bias=%s conviction=%s draw=%s@%.2f invalid=%.2f thin=%v facts_hash=%s…",
			monday, version, doc.Bias, doc.Conviction, doc.Draw.Name, doc.Draw.Px, doc.Invalidation.Px, doc.ThinHistory, facts.FactsHash[:12])
		at.weeklyState.mu.Lock()
		at.weeklyState.resetWeek(monday)
		at.weeklyState.doc, at.weeklyState.loaded = doc, true
		at.weeklyState.mu.Unlock()
		return
	}
	// Second reject → no doc this week (fail-open downstream) + loud line.
	at.logErrorf("⚠️ WEEKLY READ FAILED for %s: %s — no weekly doc this week (sessions render WEEKLY: none; nothing else changes).", monday, lastErr)
	telemetry.IncWeeklyReadFailed(at.id)
}

// weeklyDocCached returns the parsed WEEKLY doc for the current governed week
// (nil when absent/failed). Cached per week so per-cycle callers are cheap.
func (at *AutoTrader) weeklyDocCached(now time.Time) *kernel.WeeklyDoc {
	if at.store == nil {
		return nil
	}
	monday := kernel.WeekGoverningMonday(now).Format("2006-01-02")
	at.weeklyState.mu.Lock()
	defer at.weeklyState.mu.Unlock()
	at.weeklyState.resetWeek(monday)
	if at.weeklyState.loaded {
		return at.weeklyState.doc
	}
	at.weeklyState.loaded = true
	row, err := at.store.Plan().GetLatestPlanForTraderSession(monday, "WEEKLY", at.id)
	if err != nil || row == nil {
		return nil
	}
	var doc kernel.WeeklyDoc
	if json.Unmarshal([]byte(row.Doc), &doc) != nil {
		return nil
	}
	at.weeklyState.doc = &doc
	return &doc
}

// maybeCheckWeeklyInvalidation is the W4 watch, called from the EXISTING
// cycle: when a CLOSED bar of the invalidation basis TF crosses the
// invalidation price, the WEEKLY doc flips bias→neutral with invalidated_at
// stamped (a NEW appended version — plans rows are immutable). NEVER
// auto-flips the opposite side. Idempotent: the invalidated_at guard makes it
// once per week max; no re-read until next Sunday.
func (at *AutoTrader) maybeCheckWeeklyInvalidation(now time.Time) {
	if !at.dayPlanEnabled() || at.store == nil || at.exchange != "ninjatrader" {
		return
	}
	if market.FuturesBarsProvider == nil {
		return
	}
	doc := at.weeklyDocCached(now)
	if doc == nil {
		return
	}
	bias := strings.ToLower(strings.TrimSpace(doc.Bias))
	if bias != "bull" && bias != "bear" {
		return
	}
	if strings.TrimSpace(doc.InvalidatedAt) != "" {
		return // guard flag — once per week max
	}
	tf := kernel.WeeklyInvalidationBasisTF(doc.Invalidation.Basis)
	if tf == "" {
		tf = kernel.WeeklyInvalidationTFDefault()
	}
	bars := market.FuturesBarsProvider(at.futuresSymbol(), tf, kernel.AISVPBarCount)
	if !kernel.WeeklyInvalidationCrossed(bias, doc.Invalidation.Px, bars) {
		return
	}
	monday := kernel.WeekGoverningMonday(now).Format("2006-01-02")
	row, err := at.store.Plan().GetLatestPlanForTraderSession(monday, "WEEKLY", at.id)
	if err != nil || row == nil {
		return
	}
	var cur kernel.WeeklyDoc
	if json.Unmarshal([]byte(row.Doc), &cur) != nil || strings.TrimSpace(cur.InvalidatedAt) != "" {
		return // re-check the stored row's guard (cache could race a fresh write)
	}
	oldBias := cur.Bias
	cur.Bias = "neutral"
	cur.InvalidatedAt = kernel.FormatCT(now)
	docJSON, _ := json.Marshal(&cur)
	version, werr := at.store.Plan().AppendPlan(&store.PlanDB{
		PlanID:        row.PlanID,
		StrategyID:    at.id,
		TradeDate:     monday,
		Session:       "WEEKLY",
		TriggerReason: "weekly_invalidated",
		Lifecycle:     "active",
		ModelID:       row.ModelID,
		PromptHash:    row.PromptHash,
		Doc:           string(docJSON),
	})
	if werr != nil {
		at.logErrorf("📅 WEEKLY INVALIDATED write failed for %s: %v", monday, werr)
		return
	}
	at.logInfof("📅 WEEKLY INVALIDATED %s @ %.2f (%s, v%d) — bias→neutral, no auto-flip; no re-read until next Sunday.",
		oldBias, doc.Invalidation.Px, tf, version)
	at.weeklyState.mu.Lock()
	at.weeklyState.doc = &cur
	at.weeklyState.mu.Unlock()
}

// weeklyConfluenceShadow (W5.1) is the SHADOW scorer: seated levels within
// WEEKLY_CONFLUENCE_BAND_ATR × ATR5m of a weekly-class reference log their
// shadow grade ONCE per level per session, plus one per-session line counting
// how many of the top-8 seats the shadow ordering would change. View only —
// the real seating is already written and never touched.
func (at *AutoTrader) weeklyConfluenceShadow(tradeDate, session string, levels []kernel.PlanLevel) {
	if !at.dayPlanEnabled() || at.store == nil || len(levels) == 0 {
		return
	}
	if market.FuturesBarsProvider == nil {
		return
	}
	now := time.Now()
	bars1m := market.FuturesBarsProvider(at.futuresSymbol(), "1m", 12000)
	bars5m := market.FuturesBarsProvider(at.futuresSymbol(), "5m", kernel.AISVPBarCount)
	price := 0.0
	if len(bars1m) > 0 {
		price = bars1m[len(bars1m)-1].Close
	}
	atr5m := kernel.StaleConfirmATR5m(bars5m)
	if atr5m <= 0 {
		atr5m = market.ExportCalculateATR(bars5m, 14)
	}
	refs := kernel.WeeklyShadowRefs(bars1m, now, price)
	band := kernel.WeeklyConfluenceBandATR()
	mult := kernel.WeeklyShadowMult()

	shadowLevels := make([]kernel.WeeklyShadowLevel, 0, len(levels))
	for _, l := range levels {
		shadowLevels = append(shadowLevels, kernel.WeeklyShadowLevel{Price: l.Price, Label: l.Label, Grade: l.Grade})
	}
	confluent, reorder := kernel.WeeklyShadowReorder(shadowLevels, refs, band, atr5m, mult)

	at.weeklyState.mu.Lock()
	if at.weeklyState.shadowLog == nil {
		at.weeklyState.shadowLog = map[string]bool{}
	}
	if at.weeklyState.shadowSeatLog == nil {
		at.weeklyState.shadowSeatLog = map[string]bool{}
	}
	seatKey := tradeDate + "|" + session
	for _, lv := range shadowLevels {
		if !kernel.WeeklyConfluent(lv, refs, band, atr5m) {
			continue
		}
		key := fmt.Sprintf("%s|%.0f", seatKey, lv.Price*100)
		if at.weeklyState.shadowLog[key] {
			continue
		}
		at.weeklyState.shadowLog[key] = true
		at.logInfof("🌗 SHADOW wk-confl %s@%.2f (%s) real=%s shadow=%.1f — view only, real seating unchanged.",
			lv.Label, lv.Price, session, lv.Grade, float64(kernel.GradeRank(lv.Grade))*mult)
	}
	if !at.weeklyState.shadowSeatLog[seatKey] {
		at.weeklyState.shadowSeatLog[seatKey] = true
		at.logInfof("🌗 SHADOW wk-seating %s %s: %d confluent level(s); shadow top-8 would differ from real on %d seat(s) — counter for the Sep-9 promotion table.",
			tradeDate, session, confluent, reorder)
	}
	at.weeklyState.mu.Unlock()
}

// weeklyCounterShadow (W5.2) annotates every entry opposing the weekly bias
// (conviction med|high) with the clauses that WOULD have changed the trade
// under the hypothetical Sep-9 hard rules. Aligned entries: silent. Shadow
// only — this call can never block or resize anything.
func (at *AutoTrader) weeklyCounterShadow(decision *kernel.Decision) {
	if !at.dayPlanEnabled() || decision == nil {
		return
	}
	if decision.Action != "open_long" && decision.Action != "open_short" {
		return
	}
	if kernel.WeeklyCounterMode() == "off" {
		return
	}
	doc := at.weeklyDocCached(time.Now())
	if doc == nil {
		return
	}
	side := "long"
	if decision.Action == "open_short" {
		side = "short"
	}
	rr := 0.0
	entry := 0.0
	if bars := market.FuturesBarsProvider(at.futuresSymbol(), "1m", 1); len(bars) > 0 {
		entry = bars[len(bars)-1].Close
	}
	if entry > 0 && decision.StopLoss > 0 && decision.TakeProfit > 0 {
		risk, reward := entry-decision.StopLoss, decision.TakeProfit-entry
		if side == "short" {
			risk, reward = decision.StopLoss-entry, entry-decision.TakeProfit
		}
		if risk > 0 {
			rr = reward / risk
		}
	}
	grade := at.weeklyScenarioGrade(decision.CitedScenario)
	clauses := kernel.WeeklyCounterClauses(doc.Bias, doc.Conviction, side, grade, rr)
	if len(clauses) == 0 {
		return // aligned or low-conviction → silent
	}
	telemetry.IncWeeklyCounter(at.id)
	for _, c := range clauses {
		switch c {
		case "would-require-A-grade":
			telemetry.IncWeeklyCounterBlock(at.id)
		case "would-halve-size":
			telemetry.IncWeeklyCounterResize(at.id)
		}
	}
	at.logInfof("⚖️ WEEKLY-COUNTER %s vs weekly %s/%s (%s) — %s (SHADOW: annotates, never blocks)",
		side, doc.Bias, doc.Conviction, decision.CitedScenario, strings.Join(clauses, " · "))
}

// weeklyScenarioGrade resolves the cited scenario's quality grade from the
// active session plan ("" when unknown — treated as non-A by the clauses).
func (at *AutoTrader) weeklyScenarioGrade(cited string) string {
	if at.store == nil || strings.TrimSpace(cited) == "" || cited == "off-plan" {
		return ""
	}
	now := time.Now()
	reg := at.sessionRegistry(now)
	sess, ok := reg.ActiveSession(now)
	if !ok {
		return ""
	}
	tradeDate, okDate := kernel.PlanChainTradeDate(sess, now)
	if !okDate {
		tradeDate = plannerTradeDateCT(now)
	}
	row, err := at.store.Plan().GetLatestPlanForTraderSession(tradeDate, sess.Name, at.id)
	if err != nil || row == nil {
		return ""
	}
	var doc kernel.PlanDoc
	if json.Unmarshal([]byte(row.Doc), &doc) != nil {
		return ""
	}
	for _, s := range doc.Scenarios {
		if strings.EqualFold(strings.TrimSpace(s.ID), strings.TrimSpace(cited)) {
			return s.Quality
		}
	}
	return ""
}

// weeklyDrawAlignTag (W5.3) computes the draw-alignment tag for a decision
// row: toward_draw | away | neutral. Called from saveDecision so EVERY
// decision row carries it.
func (at *AutoTrader) weeklyDrawAlignTag(record *store.DecisionRecord) string {
	doc := at.weeklyDocCached(time.Now())
	if doc == nil {
		return "neutral"
	}
	side := ""
	entry := 0.0
	for _, d := range record.Decisions {
		if d.Action == "open_long" {
			side = "long"
			break
		}
		if d.Action == "open_short" {
			side = "short"
			break
		}
	}
	if side == "" {
		return "neutral"
	}
	if bars := market.FuturesBarsProvider(at.futuresSymbol(), "1m", 1); len(bars) > 0 {
		entry = bars[len(bars)-1].Close
	}
	return kernel.WeeklyDrawAlignTag(doc.Bias, side, doc.Draw.Px, entry)
}

// applyWeeklyDecisionShadow is the single per-entry hook: counter annotation
// (W5.2, log-only). The draw-align tag rides saveDecision.
func (at *AutoTrader) applyWeeklyDecisionShadow(decision *kernel.Decision) {
	defer func() {
		if r := recover(); r != nil {
			at.logWarnf("weekly shadow annotation recovered: %v (shadow never affects the trade)", r)
		}
	}()
	at.weeklyCounterShadow(decision)
}
