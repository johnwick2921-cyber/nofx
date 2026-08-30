package trader

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"nofx/kernel"
	"nofx/logger"
	"nofx/market"
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
	ntTrader "nofx/trader/ninjatrader"
)

// armedPlaceTicks is the placement band (ARM_PLACE_TICKS, default 100): the
// resting limit is placed once price comes within this many ticks of entry.
func armedPlaceTicks() int {
	if v := os.Getenv("ARM_PLACE_TICKS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return n
		}
	}
	return 100
}

// armMinRR is the R:R floor for the GATE-AT-ARM chain only
// (ARM_MIN_RR, default 2.0). Autopsy-response wave (2026-08-27): armed limits
// fill AT the level (better entry by construction, no stale risk) and the one
// refused arm replayed +$108 — the global entry floor (3.0) is NOT lowered;
// AI-proposed market entries keep their own gate unchanged.
func armMinRR() float64 {
	if v := os.Getenv("ARM_MIN_RR"); v != "" {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && f > 0 {
			return f
		}
	}
	return 2.0
}

// armedWorkingStaleMin is the reconnect/reconcile safety net
// (ARM_WORKING_STALE_MIN, default 15): a working row with no order_update for
// this long is cancelled with an honest reason.
// E7 (entry-mechanics 2026-08-30) — stop-entry knobs.
// STOP_ENTRY_SEAM (default OFF): the stop_entry order path is NEVER sent on
// the wire until the far-side AddOn has proven the frame (D-rule — an unproven
// C# path must not ship at night). STOP_ENTRY_OFFSET_TICKS (default 2): the
// stop trigger sits N ticks beyond the break candle for a stop-market entry.
func stopEntrySeamOn() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("STOP_ENTRY_SEAM")), "on")
}

func stopEntryOffsetTicks() int {
	if v := strings.TrimSpace(os.Getenv("STOP_ENTRY_OFFSET_TICKS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 2
}

func retestWaitBars() int {
	if v := strings.TrimSpace(os.Getenv("RETEST_WAIT_BARS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 6
}

// stopEntryFallbackDue (E7, pure) — the breakout-retest fallback window: a
// stop-entry leg is due when NO bar has touched its entry level within the
// last RETEST_WAIT_BARS 1m bars AND at least RETEST_WAIT_BARS bars elapsed
// since the plan's birth (no retest came → chase with a stop beyond the
// break candle instead of waiting forever).
func stopEntryFallbackDue(bars []market.Kline, entryPx, sinceMs, nowMs int64) bool {
	need := retestWaitBars()
	if len(bars) < need {
		return false
	}
	var closedSinceBirth int
	for i := len(bars) - 1; i >= 0 && closedSinceBirth < need; i-- {
		b := bars[i]
		if b.CloseTime >= nowMs {
			continue
		}
		if b.OpenTime < sinceMs {
			break
		}
		closedSinceBirth++
		if b.Low <= float64(entryPx) && b.High >= float64(entryPx) {
			return false // a retest touch came in-window — the limit logic owns it
		}
	}
	return closedSinceBirth >= need
}

func armedWorkingStaleMin() int {	if v := os.Getenv("ARM_WORKING_STALE_MIN"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return n
		}
	}
	return 15
}

// armedSubs holds each trader's order_update stream (created once per trader).
var armedSubs sync.Map

// armedTrader type-asserts the bound trader to the TCPTrader (nil when the
// trader is crypto or unwired — the engine then stays dormant).
func (at *AutoTrader) armedTrader() *ntTrader.TCPTrader {
	nt, _ := at.trader.(*ntTrader.TCPTrader)
	return nt
}

// ARMED-ORDER EXECUTOR — Wave 2 (2026-08-27).
//
// PHASE 1 (this file): the ARMING CONTRACT. The AI authorizes arming per
// scenario (plan.scenarios[].arm); Go evaluates gates AT ARM TIME and keeps the
// durable armed_orders ledger. Placement as a working NT8 order is PHASE 2 —
// Phase 1 rows stay state=armed. Everything here is dormant until a plan
// actually carries arm specs (no behavior change to today's trading).

// maybeManageArmedOrders runs every cycle (called from runCycle). It is a no-op
// unless day_plan is on and armed specs exist. snap is the structure snapshot
// for the HTF veto gate.

// declineHadFreshMet (S5, autopsy-response wave) — true when the active plan
// has a scenario whose confirm{} is machine-MET and NOT stale at this instant:
// the honest-wait leak the autopsy quantified (declines while a FRESH confirm
// was live). Mirrors RenderConfirmLines' staleness rule.
func (at *AutoTrader) declineHadFreshMet() bool {
	plan := kernel.ActivePlanFor(at.id, at.futuresSymbol())
	if plan == nil {
		return false
	}
	var bars []market.Kline
	if market.FuturesBarsProvider != nil {
		bars = market.FuturesBarsProvider(at.futuresSymbol(), kernel.AISVPBarInterval, kernel.AISVPBarCount)
	}
	if len(bars) == 0 {
		return false
	}
	nowMs := time.Now().UnixMilli()
	nowPrice := bars[len(bars)-1].Close
	atr5m := market.ExportCalculateATR(kernel.AcceptanceBars(bars, "2x5m"), 14)
	for _, s := range plan.Doc.Scenarios {
		if s.Confirm == nil {
			continue
		}
		v := kernel.EvaluateConfirm(*s.Confirm, bars, plan.BirthMs, nowMs)
		if !v.Met {
			continue
		}
		if atr5m > 0 && math.Abs(nowPrice-s.Confirm.RefPrice) > kernel.StaleConfirmATR()*atr5m {
			continue // stale-MET is NOT the leak class
		}
		return true
	}
	return false
}

func (at *AutoTrader) maybeManageArmedOrders(snap map[string]kernel.StructureState) {
	if !at.dayPlanEnabled() || at.store == nil || at.exchange != "ninjatrader" {
		return
	}
	ledger := at.store.ArmedOrders()
	if ledger == nil {
		return
	}
	now := time.Now()

	// 1.4 — plan → dormant/no_trade/absent = ALL its armed orders cancelled
	// instantly. Re-arm does NOT auto-re-arm (fresh AI authorization required).
	// 2.4 — no active session (EOD flat) cancels everything too.
	plan := kernel.ActivePlanFor(at.id, at.futuresSymbol())
	_, sessOK := at.sessionRegistry(now).ActiveSession(now)
	reason := ""
	if plan == nil {
		if !sessOK {
			reason = "session ended (EOD flat)"
		} else {
			reason = "no active plan"
		}
	} else {
		row, err := at.store.Plan().GetLatestPlanForTraderSession(kernel.PlanTradeDateFor(plan), plan.Session, at.id)
		if err != nil || row == nil {
			reason = "plan row unavailable"
		} else if row.Lifecycle != "active" {
			reason = fmt.Sprintf("plan lifecycle %q", row.Lifecycle)
		}
	}
	if reason != "" {
		// S-list closer: synchronous (ack-waited) cancel on session end and
		// dormancy too — the resting limit must be dead BEFORE any flatten or
		// the next cycle, not up to 2m later.
		if n, unacked := at.cancelArmedOrdersSync(reason); n > 0 {
			at.logWarnf("🔒 armed cancel: %s — %d order(s) disarmed", reason, n)
			if unacked > 0 {
				at.logWarnf("⚠️ armed cancel: %d unacked after retry (ledger cancelled; wire reconciles next cycle)", unacked)
			}
		}
		return
	}

	// 2. arm evaluation for the ACTIVE plan's arm specs.
	doc := plan.Doc
	cfg := at.GetStrategyConfig()
	if cfg == nil {
		return
	}
	var bars []market.Kline
	if market.FuturesBarsProvider != nil {
		bars = market.FuturesBarsProvider(at.futuresSymbol(), kernel.AISVPBarInterval, kernel.AISVPBarCount)
	}
	atr5m := 0.0
	if len(bars) > 0 {
		atr5m = market.ExportCalculateATR(kernel.AcceptanceBars(bars, "2x5m"), 14)
	}
	minQuality := ""
	if dp := at.dayPlanCfg(); dp != nil {
		minQuality = dp.MinScenarioQualityFor(plan.Session)
	}
	for _, sc := range doc.Scenarios {
		if sc.Arm == nil || !sc.Arm.Enabled {
			continue
		}
		// E4 (2026-08-30) — split-entry legs: a two-leg arm writes one ledger
		// row PER leg (LegIndex 0/1, LegCount 2). A single arm is leg 0 of a
		// one-row pair (LegCount 0 = legacy shape).
		legs := sc.Arm.Legs
		if len(legs) == 0 {
			legs = []kernel.PlanArmLeg{{Entry: sc.Arm.Entry, Stop: sc.Arm.Stop, Target: sc.Arm.Target,
				WaitConfirm: sc.Arm.WaitConfirm, Rule: "touch", Kind: "limit"}}
		}
		legCount := 0
		if len(sc.Arm.Legs) == 2 {
			legCount = 2
		}
		for li, leg := range legs {
			// S2b chained arm (autopsy-response wave): wait_confirm legs stay
			// DORMANT until the chain confirm is machine-MET. E4: leg 1 chains
			// on confirm2 (1m_mss|1x5m_close); a legacy single arm chains on
			// its own confirm{}.
			if leg.WaitConfirm {
				chain := sc.Confirm2
				if len(sc.Arm.Legs) == 0 {
					chain = sc.Confirm
				}
				if chain == nil {
					continue
				}
				v := kernel.EvaluateConfirm(*chain, bars, plan.BirthMs, now.UnixMilli())
				if !v.Met {
					continue
				}
				at.logInfof("⚔️ arm %s leg %d wait_confirm MET (%s) — arming", sc.ID, li+1, leg.Rule)
			}
			// gates AT ARM TIME — a resting order is a pre-passed entry; each gate
			// input that changes materially later triggers a cancel (1.3).
			if verdict := at.armGateVerdictFor(sc, leg, biasDirectionFor(doc.Bias.Direction), snap, atr5m, minQuality, cfg); verdict != "" {
				// F4 (LONDON-FORENSICS 2026-08-28) — log the REFUSED verdict ONCE
				// per arm-spec (the same infeasible arm re-refused every cycle
				// printed ~120 lines/session); silent until the spec changes.
				// GAR-F6 (2026-08-28): the comparison VALUE is the verdict CLASS,
				// not the ATR-bearing string — live ATR drift re-logged the same
				// refusal every few minutes (LONDON S4 min-SL 18.29→18.67).
				// PRE-REOPEN F3 (2026-08-28) — the missing 1.3 clause: a gate input
				// that changes materially while the arm is WORKING cancels it the
				// same cycle (the LONDON S4 class stayed resting through repeated
				// re-refusals until the 08:30 sweep).
				if rows, lerr := ledger.ListNonTerminal(at.id); lerr == nil {
					for _, r := range rows {
						if r.TraderID == at.id && r.PlanID == plan.PlanID && r.Scenario == sc.ID &&
							r.State == "working" && r.SignalID != "" {
							if nt := at.armedTrader(); nt != nil {
								if cerr := nt.CancelOrder(r.SignalID); cerr == nil {
									_ = ledger.SetState(r.ID, "cancelled", "gate changed: "+armRefusalClass(verdict))
									at.logWarnf("✕ armed cancel (gate changed %s): %s %s", armRefusalClass(verdict), plan.Session, sc.ID)
								}
							}
						}
					}
				}
				key := plan.PlanID + ":" + strconv.Itoa(plan.Version) + ":" + sc.ID + ":leg" + strconv.Itoa(li+1)
				if armRefusalChanged(&at.armRefusalLast, key, armRefusalClass(verdict)) {
					at.logWarnf("⚔️ arm REFUSED %s %s leg %d: %s", plan.Session, sc.ID, li+1, verdict)
				}
				continue
			}
			side := strings.ToLower(strings.TrimSpace(sc.Direction))
			row := &store.ArmedOrderDB{
				TraderID: at.id, PlanID: plan.PlanID, Version: plan.Version, Session: plan.Session,
				Scenario: sc.ID, Side: side, EntryPx: leg.Entry, StopPx: leg.Stop, TargetPx: leg.Target,
				State: "armed", EntryClass: "armed_fill", CreatedAt: now, UpdatedAt: now,
				LegIndex: li, LegCount: legCount, Kind: leg.Kind,
			}
			existing, err := ledger.ListNonTerminal(at.id)
			if err == nil {
				for i := range existing {
					if existing[i].TraderID == at.id && existing[i].PlanID == row.PlanID && existing[i].Scenario == sc.ID && existing[i].LegIndex == row.LegIndex {
						row.ID = existing[i].ID // already in the ledger — leave state (churn guard applies to placement)
						break
					}
				}
			}
			if row.ID == 0 {
				if err := ledger.UpsertArm(row); err != nil {
					at.logWarnf("⚔️ arm write failed %s %s leg %d: %v", plan.Session, sc.ID, li+1, err)
					continue
				}
				// PRE-REOPEN F3 (2026-08-28) — the authored log fires ONCE per
				// spec (dedup by plan:version:scenario + prices); the dead-row
				// re-log spam (69+ lines/day) came from logging every cycle.
				akey := plan.PlanID + ":" + strconv.Itoa(plan.Version) + ":" + sc.ID + ":leg" + strconv.Itoa(li+1)
				aval := fmt.Sprintf("%s %.2f/%.2f/%.2f", side, leg.Entry, leg.Stop, leg.Target)
				if armRefusalChanged(&at.armAuthoredLast, akey, aval) {
					at.logInfof("⚔️ armed %s %s leg %d %s limit %.2f SL %.2f TP %.2f (tick-managed placement is Phase 2)", plan.Session, sc.ID, li+1, side, leg.Entry, leg.Stop, leg.Target)
				}
			} else {
				// CHURN GUARD (2.1): re-spec a working arm's bracket only when the
				// plan moved SL or TP by ≥ 2 ticks (cancel+re-place on modify).
				tick := market.FuturesTickSize(at.futuresSymbol())
				if tick <= 0 {
					tick = 0.25
				}
				if row.State == "working" && churnNeedsModify(row.StopPx, row.TargetPx, leg.Stop, leg.Target, tick) {
					if nt := at.armedTrader(); nt != nil {
						_ = nt.ModifyBracket(row.SignalID, leg.Stop, leg.Target)
						at.logInfof("📌 armed %s leg %d bracket modify (churn guard) SL %.2f→%.2f TP %.2f→%.2f",
							sc.ID, li+1, row.StopPx, leg.Stop, row.TargetPx, leg.Target)
					}
				}
				row.EntryPx, row.StopPx, row.TargetPx = leg.Entry, leg.Stop, leg.Target
				row.Version = plan.Version
				_ = ledger.UpsertArm(row)
			}
		}
	}

	// E8 (2026-08-30) — shadow A/B counterfactual logger (Sep-9's courtroom):
	// per armed scenario, log the 4 rule counterfactuals once per plan version.
	// ZERO effect on real paths — writes ONLY the ab_confirm_log table.
	for _, sc := range doc.Scenarios {
		at.logShadowAB(plan, sc, bars, now.UnixMilli())
	}

	// E4 (2026-08-30) — split-sibling law: EITHER leg's STOP-OUT cancels the
	// sibling's unfilled order (no doubling into a failed level). Runs on the
	// existing cancel machinery; session-end/news/dormant cancel paths already
	// cover BOTH legs (cancel-all by trader).
	at.cancelSplitSiblingOnStopOut(ledger)

	// PHASE 2 — placement engine (armed → working within the tick band), wire
	// cancel/modify, and the order_update event machine.
	at.runArmedPlacement(bars, plan.BirthMs)
}

// biasDirectionFor normalizes the plan bias direction ("" → empty).
func biasDirectionFor(dir string) string {
	return strings.ToLower(strings.TrimSpace(dir))
}

// logShadowAB (E8) writes the 4 counterfactual confirm-fill rows for one armed
// scenario — once per (plan, version, scenario, rule). Advisory/report-only:
// nothing here feeds a gate or a prompt.
func (at *AutoTrader) logShadowAB(plan *kernel.ActivePlan, sc kernel.PlanScenario, bars []market.Kline, nowMs int64) {
	if at.store == nil || plan == nil || len(bars) == 0 {
		return
	}
	rows := kernel.ShadowABForScenario(sc, bars, plan.BirthMs, nowMs)
	if len(rows) == 0 {
		return
	}
	ac := at.store.AbConfirm()
	now := time.Now()
	for _, r := range rows {
		if ac.Has(plan.PlanID, plan.Version, sc.ID, r.Rule) {
			continue
		}
		if err := ac.Upsert(&store.AbConfirmLogDB{
			TraderID: at.id, PlanID: plan.PlanID, Version: plan.Version, Session: plan.Session,
			Scenario: sc.ID, Rule: r.Rule, FillPx: r.FillPx, MFE: r.MFE, MAE: r.MAE,
			Outcome: r.Outcome, TimeToFillMs: r.TimeToFillMs, CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			at.logWarnf("ab-confirm shadow write failed %s %s: %v", sc.ID, r.Rule, err)
		}
	}
}

// splitSiblingCancelDecision (E4, pure) — given one split pair (legs of the
// same plan:scenario, LegCount=2) and the CLOSED positions of that trader,
// decide which sibling rows to cancel: a FILLED leg whose position CLOSED at
// its STOP (exit within 2 ticks of the leg's stop price) voids the level, so
// the sibling's unfilled order must die — no doubling into a failed level.
// A target-out or an open position cancels NOTHING.
func splitSiblingCancelDecision(pair []store.ArmedOrderDB, closed []store.TraderPosition, tick float64) []store.ArmedOrderDB {
	if tick <= 0 {
		tick = 0.25
	}
	var out []store.ArmedOrderDB
	stopHit := false
	for _, leg := range pair {
		if leg.State != "filled" || leg.SignalID == "" {
			continue
		}
		for _, p := range closed {
			if p.EntryOrderID != leg.SignalID || p.TraderID != leg.TraderID {
				continue
			}
			if math.Abs(p.ExitPrice-leg.StopPx) <= 2*tick && math.Abs(p.ExitPrice-leg.TargetPx) > 2*tick {
				stopHit = true // the filled leg stopped out
			}
		}
	}
	if !stopHit {
		return nil
	}
	for _, leg := range pair {
		if leg.State == "armed" || leg.State == "working" {
			out = append(out, leg)
		}
	}
	return out
}

// cancelSplitSiblingOnStopOut (E4) — the wire half of the split-sibling law,
// riding the existing cancel machinery (cancel + ledger state + ack seam).
func (at *AutoTrader) cancelSplitSiblingOnStopOut(ledger *store.ArmedOrderStore) {
	if ledger == nil {
		return
	}
	rows, err := ledger.ListNonTerminal(at.id)
	if err != nil || len(rows) == 0 {
		return
	}
	// Group split pairs.
	pairs := map[string][]store.ArmedOrderDB{}
	for _, r := range rows {
		if r.TraderID != at.id || r.LegCount != 2 {
			continue
		}
		key := r.PlanID + ":" + r.Scenario
		pairs[key] = append(pairs[key], r)
	}
	if len(pairs) == 0 {
		return
	}
	// Collect the filled legs' signal ids.
	var sigs []string
	for _, p := range pairs {
		for _, r := range p {
			if r.State == "filled" && r.SignalID != "" {
				sigs = append(sigs, r.SignalID)
			}
		}
	}
	if len(sigs) == 0 {
		return
	}
	closedPtrs, err := at.store.Position().ListClosedByEntryOrderIDs(at.id, sigs)
	if err != nil || len(closedPtrs) == 0 {
		return
	}
	closed := make([]store.TraderPosition, 0, len(closedPtrs))
	for _, cp := range closedPtrs {
		closed = append(closed, *cp)
	}
	tick := market.FuturesTickSize(at.futuresSymbol())
	for _, p := range pairs {
		for _, sibling := range splitSiblingCancelDecision(p, closed, tick) {
			reason := "sibling stopped out — split contract (E4)"
			nt := at.armedTrader()
			if nt != nil && sibling.SignalID != "" {
				if cerr := nt.CancelOrder(sibling.SignalID); cerr == nil {
					_ = ledger.SetState(sibling.ID, "cancelled", reason)
					at.logWarnf("✕ armed cancel %s %s leg %d: %s", sibling.Session, sibling.Scenario, sibling.LegIndex+1, reason)
					continue
				}
			}
			if sibling.SignalID == "" {
				// Still just an authorization — kill it in the ledger; placement
				// will never fire for it.
				_ = ledger.SetState(sibling.ID, "cancelled", reason)
				at.logWarnf("✕ armed cancel %s %s leg %d: %s", sibling.Session, sibling.Scenario, sibling.LegIndex+1, reason)
			}
		}
	}
}

// armedLines renders the per-cycle ARMED: lines for the executor prompt.
func (at *AutoTrader) armedLines() string {
	if at.store == nil {
		return ""
	}
	rows, err := at.store.ArmedOrders().ListNonTerminal(at.id)
	if err != nil || len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	for _, r := range rows {
		if r.TraderID != at.id {
			continue
		}
		glyph := map[string]string{"armed": "⏳ armed", "working": "📌 working"}[r.State]
		if glyph == "" {
			continue
		}
		kind := "limit"
		if r.Kind == "stop_entry" {
			kind = "stop"
		}
		fmt.Fprintf(&b, "ARMED: %s %s %s %s %.2f SL %.2f TP %.2f (%s)\n", r.Scenario, r.Side, r.State, kind, r.EntryPx, r.StopPx, r.TargetPx, glyph)
	}
	if b.Len() == 0 {
		return ""
	}
	return b.String()
}

// runArmedPlacement drives the armed→working transition, the churn guard, and
// the order_update event machine. No-op unless a TCPTrader is bound.
// sinceMs = the plan's birth (the E7 stop-entry fallback window is measured
// from it).
func (at *AutoTrader) runArmedPlacement(bars []market.Kline, sinceMs int64) {
	nt := at.armedTrader()
	if nt == nil {
		return
	}
	ledger := at.store.ArmedOrders()
	if ledger == nil {
		return
	}
	now := time.Now()
	price := 0.0
	if len(bars) > 0 {
		price = bars[len(bars)-1].Close
	}
	tick := market.FuturesTickSize(at.futuresSymbol())
	if tick <= 0 {
		tick = 0.25
	}
	band := float64(armedPlaceTicks()) * tick

	rows, err := ledger.ListNonTerminal(at.id)
	if err != nil {
		return
	}
	for _, r := range rows {
		if r.TraderID != at.id {
			continue
		}
		switch r.State {
		case "armed":
			// E7 — stop-entry legs (breakout-retest fallback / breakdown
			// immediate alternative): never placed until the far-side AddOn
			// proves the frame (STOP_ENTRY_SEAM) AND the no-retest window has
			// elapsed (RETEST_WAIT_BARS). The stop trigger sits
			// STOP_ENTRY_OFFSET_TICKS beyond the level.
			if r.Kind == "stop_entry" {
				if !stopEntrySeamOn() {
					continue // seam off → the leg stays armed (never on the wire)
				}
				if !stopEntryFallbackDue(bars, int64(r.EntryPx), sinceMs, now.UnixMilli()) {
					continue // still inside the retest window
				}
				offset := float64(stopEntryOffsetTicks()) * tick
				trigger := r.EntryPx - offset
				if r.Side == "long" {
					trigger = r.EntryPx + offset
				}
				sid, perr := nt.PlaceStopEntry(at.futuresSymbol(), r.Side, 1, trigger, r.StopPx, r.TargetPx)
				if perr != nil {
					at.logWarnf("📌 stop-entry place failed %s: %v", r.Scenario, perr)
					continue
				}
				_ = ledger.SetSignal(r.ID, sid)
				_ = ledger.SetState(r.ID, "working", "")
				at.logInfof("📌 armed %s → WORKING stop-entry %.2f signal=%s (no retest in %d bars, offset %dt)", r.Scenario, trigger, sid, retestWaitBars(), stopEntryOffsetTicks())
				continue
			}
			if price > 0 && math.Abs(price-r.EntryPx) <= band {
				sid, perr := nt.PlaceLimitEntry(at.futuresSymbol(), r.Side, 1, r.EntryPx, r.StopPx, r.TargetPx)
				if perr != nil {
					at.logWarnf("📌 armed place failed %s: %v", r.Scenario, perr)
					continue
				}
				_ = ledger.SetSignal(r.ID, sid)
				_ = ledger.SetState(r.ID, "working", "")
				at.logInfof("📌 armed %s → WORKING limit %.2f signal=%s (band ±%.0ft)", r.Scenario, r.EntryPx, sid, band/tick)
			}
		}
	}
	// reconnect/reconcile safety net (separate pass — cancelFn is the wire seam).
	at.reconcileStaleWorking(ledger, rows, now, armedWorkingStaleMin(), func(sid string) { _ = nt.CancelOrder(sid) })
	at.consumeArmedOrderUpdates(nt, ledger)
}

// churnNeedsModify — the churn guard predicate: the plan re-spec'd a working
// arm's SL or TP by ≥ 2 ticks (2.1). Pure for tests.
func churnNeedsModify(oldStop, oldTarget, newStop, newTarget, tick float64) bool {
	return math.Abs(oldStop-newStop) >= 2*tick || math.Abs(oldTarget-newTarget) >= 2*tick
}

// armRefusalClass (GAR-F6) — the refusal's verdict CLASS, stripped of live
// ATR numbers. The dedup key uses the class so an ATR drift (e.g. min-SL
// "18.29×ATR5m" → "18.67×ATR5m") does not re-log the same refusal.
func armRefusalClass(verdict string) string {
	v := strings.ToLower(verdict)
	switch {
	case strings.HasPrefix(v, "r:r"):
		return "rr"
	case strings.Contains(v, "too close"):
		return "min_sl"
	case strings.Contains(v, "veto"):
		return "veto"
	case strings.Contains(v, "not armable") || strings.Contains(v, "non-armable"):
		return "not_armable"
	default:
		return "other"
	}
}

// armRefusalChanged (F4) — true when this arm-spec's refusal verdict is new or
// changed (the caller logs); false when the identical verdict was already
// logged for the same spec (the caller stays silent).
func armRefusalChanged(last *map[string]string, key, verdict string) bool {
	if *last == nil {
		*last = map[string]string{}
	}
	if prev, ok := (*last)[key]; ok && prev == verdict {
		return false
	}
	(*last)[key] = verdict
	return true
}

// workingStale — the reconnect predicate: no order_update for the stale window.
func workingStale(updatedAt, now time.Time, staleMin int) bool {
	return now.Sub(updatedAt) > time.Duration(staleMin)*time.Minute
}

// reconcileStaleWorking cancels working rows that have seen no order_update for
// the stale window (reconnect safety net). cancelFn issues the NT8 cancel — the
// ledger flips to cancelled with the reason regardless.
func (at *AutoTrader) reconcileStaleWorking(ledger *store.ArmedOrderStore, rows []store.ArmedOrderDB, now time.Time, staleMin int, cancelFn func(signalID string)) {
	for _, r := range rows {
		if r.TraderID != at.id || r.State != "working" {
			continue
		}
		if !workingStale(r.UpdatedAt, now, staleMin) {
			continue
		}
		if r.SignalID != "" && cancelFn != nil {
			cancelFn(r.SignalID)
		}
		_ = ledger.SetState(r.ID, "cancelled", "no order_update within stale window (reconnect/reconcile)")
		at.logWarnf("✕ armed %s cancelled — no order_update for %dm (reconnect/reconcile)", r.Scenario, staleMin)
	}
}

// consumeArmedOrderUpdates subscribes ONCE to the trader's order_update
// stream and drains pending events into the ledger.
//
// 2026-08-27 bug: the old code called nt.OrderUpdates() as the LoadOrStore
// argument on EVERY cycle — the argument is evaluated first, and
// SubscribeOrderUpdatesFor CLOSES+replaces the channel on each subscribe.
// The map then held a closed channel forever: the drain read 310,808
// zero-value payloads in 15s at 13:34:48 and the consumer was permanently
// dead. Now: subscribe only on the miss path, and self-heal (delete the map
// entry) if the channel is ever closed.
func (at *AutoTrader) consumeArmedOrderUpdates(nt *ntTrader.TCPTrader, ledger *store.ArmedOrderStore) {
	// S-list closer: the subscription is now created via armedUpdateStream —
	// subscribe exactly once on the miss path, never re-subscribing per cycle.
	ch := at.armedUpdateStream(nt)
	if ch == nil {
		return
	}
	for {
		select {
		case u, open := <-ch:
			if !open {
				armedSubs.Delete(at.id)
				at.logWarnf("📡 armed order_update channel closed — re-subscribing next cycle")
				return
			}
			at.onArmedOrderUpdate(u, ledger)
		default:
			return
		}
	}
}

// Armed order_update frame receipt — FORENSICS HYGIENE (2026-08-28): the
// per-frame INFO log at this site ate 1.48GB of journal in one hour
// (25k lines/s during an NT8 order_update storm at 13:35-14:35 CT, the
// 3.6h-retention S-finding). The frame is now logged at DEBUG sampled 1-in-N
// (T8 pattern) and the receive path stays provable via a 1-line/min INFO
// summary with per-state counts.
var (
	armedOUFrames  atomic.Int64 // frames since the last summary
	armedOUSample  atomic.Int64 // per-frame sample counter
	armedOUByState sync.Map     // state string -> *atomic.Int64
	armedOULastSum atomic.Int64 // unix secs of the last summary
)

func armedOrderUpdateLogSample() int64 {
	if v := os.Getenv("ARMED_ORDER_UPDATE_LOG_SAMPLE"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return 500
}

func logArmedOrderUpdateSummary() {
	now := time.Now().Unix()
	if now-armedOULastSum.Load() < 60 {
		return
	}
	if !armedOULastSum.CompareAndSwap(armedOULastSum.Load(), now) {
		return
	}
	n := armedOUFrames.Swap(0)
	var b strings.Builder
	fmt.Fprintf(&b, "📡 armed order_update summary (1-line/min): frames=%d", n)
	armedOUByState.Range(func(k, v any) bool {
		if c := v.(*atomic.Int64).Swap(0); c > 0 {
			fmt.Fprintf(&b, " %s=%d", k, c)
		}
		return true
	})
	logger.Infof("%s", b.String())
}

// onArmedOrderUpdate applies one NT8 order state change to the armed ledger.
func (at *AutoTrader) onArmedOrderUpdate(u ntwire.OrderUpdatePayload, ledger *store.ArmedOrderStore) {
	// Frame-receipt proof (cutover confirmation wave): the C# dispatcher's
	// receive path stays provable from the journal via the 1-line/min summary;
	// the per-frame content is DEBUG + 1-in-N sampled (FORENSICS HYGIENE —
	// this was the 1.48GB/hour journal flood).
	armedOUFrames.Add(1)
	if c, _ := armedOUByState.LoadOrStore(u.State, &atomic.Int64{}); c.(*atomic.Int64).Add(1) == 0 {
		_ = c // kept for clarity: the value-add already happened
	}
	if armedOUSample.Add(1)%armedOrderUpdateLogSample() == 0 {
		logger.Debugf("📡 armed order_update frame: state=%s signal=%s acct=%s fill=%.2f",
			u.State, u.SignalID, u.Account, u.FillPrice)
	}
	logArmedOrderUpdateSummary()
	rows, err := ledger.ListNonTerminal(at.id)
	if err != nil {
		return
	}
	for _, r := range rows {
		if r.TraderID != at.id || r.SignalID != u.SignalID {
			continue
		}
		switch strings.ToLower(u.State) {
		case "filled", "partfilled":
			_ = ledger.SetState(r.ID, "filled", "fill@"+strconv.FormatFloat(u.FillPrice, 'f', 2, 64))
			_ = ledger.SetFillPrice(r.ID, u.FillPrice)
			_ = ledger.Touch(r.ID)
			at.stampArmedFillLineage(r, u.FillPrice)
			at.logInfof("⚡ armed fill %s @ %.2f (entry_class=armed_fill — stale_reeval NOT applied)", r.Scenario, u.FillPrice)
		case "rejected":
			_ = ledger.SetState(r.ID, "cancelled", "NT8 reject")
			at.logWarnf("✕ armed %s NT8-REJECTED — disarmed", r.Scenario)
		case "cancelled":
			_ = ledger.SetState(r.ID, "cancelled", "cancelled in NT8")
			at.logInfof("✕ armed %s cancelled in NT8", r.Scenario)
		}
		return
	}
}

// stampArmedFillLineage links the freshly-filled position row to the plan the
// arm cited — the same fields AI entries carry (S3 SetPlanLinkFull).
func (at *AutoTrader) stampArmedFillLineage(r store.ArmedOrderDB, fillPrice float64) {
	pos, err := at.store.Position().GetOpenPositionBySymbol(at.id, at.futuresSymbol(), r.Side)
	if err != nil || pos == nil {
		// PRE-REOPEN F4 (2026-08-28) — the fill frame precedes position
		// materialization (all 4 live fills hit this race). The LEDGER row
		// carries the pending marker; the reconcile materialization path
		// (StampArmedLineageIfMatched) completes the stamp and clears it.
		if e2 := at.store.ArmedOrders().SetState(r.ID, "filled", fmt.Sprintf("%s;stamp_pending", r.StateReason)); e2 != nil {
			at.logWarnf("⚡ armed fill %s: pending-marker write failed: %v", r.Scenario, e2)
			return
		}
		at.logInfof("⚡ armed fill %s @ %.2f: position row not materialized yet — stamp pending (reconcile completes it)", r.Scenario, fillPrice)
		return
	}
	tradeDate := r.PlanID
	if i := strings.Index(r.PlanID, ":"); i > 0 {
		tradeDate = r.PlanID[:i]
	}
	if err := at.store.Position().SetPlanLinkFull(pos.ID, r.Version, r.Scenario, true, "armed_fill", r.PlanID, tradeDate, r.Session); err != nil {
		at.logWarnf("⚡ armed fill lineage stamp failed: %v", err)
	}
}

// armGateVerdict runs the arm-time gate chain for a SINGLE arm (legacy shape).
// Empty string = pass.
func (at *AutoTrader) armGateVerdict(sc kernel.PlanScenario, biasDirection string, snap map[string]kernel.StructureState, atr5m float64, minQuality string, cfg *store.StrategyConfig) string {
	if sc.Arm == nil {
		return "no arm"
	}
	return at.armGateVerdictFor(sc, kernel.PlanArmLeg{Entry: sc.Arm.Entry, Stop: sc.Arm.Stop, Target: sc.Arm.Target}, biasDirection, snap, atr5m, minQuality, cfg)
}

// armGateVerdictFor runs the arm-time gate chain for ONE LEG's prices (E4: the
// split legs gate independently — each leg is a pre-passed entry of its own).
// The min-confidence gate is N/A for arms — the AI's authorization IS the
// confidence signal (no per-scenario confidence exists to check).
func (at *AutoTrader) armGateVerdictFor(sc kernel.PlanScenario, leg kernel.PlanArmLeg, biasDirection string, snap map[string]kernel.StructureState, atr5m float64, minQuality string, cfg *store.StrategyConfig) string {
	a := sc.Arm
	if err := kernel.ArmSpecValid(sc); err != nil {
		return err.Error()
	}
	side := strings.ToLower(strings.TrimSpace(sc.Direction))
	if side != "long" && side != "short" {
		return fmt.Sprintf("direction %q not armable", sc.Direction)
	}
	// plan_mode direction — the plan is the law, same as the entry path.
	if at.planModeFor("") == "direction" {
		bias := strings.ToLower(strings.TrimSpace(biasDirection))
		if bias != "" && bias != side {
			return fmt.Sprintf("against plan bias %q (plan_mode=direction)", bias)
		}
	}
	// quality floor (min_scenario_quality).
	if minQuality != "" {
		if kernel.ScenarioQualityRank(sc.Quality) < kernel.ScenarioQualityRank(minQuality) {
			return fmt.Sprintf("quality %s below min_scenario_quality %s", sc.Quality, minQuality)
		}
	}
	// R:R gate — ARM_MIN_RR (default 2.0), the gate-at-arm floor. Autopsy
	// response (2026-08-27): resting limits fill AT the level (better entry by
	// construction) — the global 3.0 floor for AI market entries is untouched.
	rr := 0.0
	if side == "long" && leg.Entry > leg.Stop && leg.Stop > 0 {
		rr = (leg.Target - leg.Entry) / (leg.Entry - leg.Stop)
	} else if side == "short" && leg.Stop > leg.Entry && leg.Entry > 0 {
		rr = (leg.Entry - leg.Target) / (leg.Stop - leg.Entry)
	}
	if rr+1e-9 < armMinRR() {
		return fmt.Sprintf("R:R %.2f below arm min %.2f", rr, armMinRR())
	}
	// min-SL — the same floor (×ATR5m) the entry path enforces.
	if atr5m > 0 {
		dist := leg.Entry - leg.Stop
		if side == "short" {
			dist = leg.Stop - leg.Entry
		}
		if dist+1e-9 < kernel.MinSLATRMult()*atr5m {
			return fmt.Sprintf("stop %.2f too close (%.2f < %.2f = %.1f×ATR5m)", leg.Stop, dist, kernel.MinSLATRMult()*atr5m, kernel.MinSLATRMult())
		}
	}
	// HTF veto — the same veto the entry path enforces.
	if blocked, vetoReason := kernel.HTFVetoVerdict(snap, "open_"+side, kernel.HTFVetoTF()); blocked {
		return "HTF veto: " + vetoReason
	}
	_ = a
	return ""
}

// cancelArmedOrders moves non-terminal rows for THIS trader to cancelled with a
// reason. Returns the count.
func (at *AutoTrader) cancelArmedOrders(reason string) int {
	rows, err := at.store.ArmedOrders().ListNonTerminal(at.id)
	if err != nil {
		return 0
	}
	n := 0
	for _, r := range rows {
		if r.TraderID != at.id {
			continue
		}
		if err := at.store.ArmedOrders().SetState(r.ID, "cancelled", reason); err == nil {
			n++
			at.logInfof("✕ armed cancel %s %s: %s", r.Scenario, r.PlanID, reason)
		}
	}
	return n
}

// ── S-LIST CLOSER (2026-08-27) — synchronous armed cancel ────────────────────
// The EOD race (deep-verify hole 11): enforceEODFlatAt flattened POSITIONS
// only, and the armed cancel ran on the NEXT cycle — a working limit could
// fill up to one 2m cycle AFTER the flat. Every lifecycle path that flattens
// or disarms around a session boundary (EOD flat, session end, dormancy, T1
// force-flat) now cancels working arms FIRST, SYNCHRONOUSLY: the NT8 cancel
// frame is sent, then the shared order_update stream is drained until the ack
// lands or the deadline passes (one retry). Acked or not the ledger flips to
// cancelled and the flatten proceeds — the cancel-before-flatten WIRE ORDER is
// what kills the window, and the flatten is never held hostage by a stuck ack.

// armedCancelAckTimeout is the per-order ack wait (ARMED_CANCEL_ACK_TIMEOUT_MS,
// default 2000). One retry ⇒ a stuck ack costs ≤ 2× this before the flatten
// proceeds.
func armedCancelAckTimeout() time.Duration {
	if v := os.Getenv("ARMED_CANCEL_ACK_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return time.Duration(n) * time.Millisecond
		}
	}
	return 2000 * time.Millisecond
}

// armedSyncSeam is the fixture seam for the synchronous cancel (nil = prod TCP).
type armedSyncSeam struct {
	Cancel  func(signalID string) error
	Stream  func() <-chan ntwire.OrderUpdatePayload
	Timeout time.Duration // 0 = armedCancelAckTimeout()
}

// armedUpdateStream returns THIS trader's shared order_update subscription,
// creating it exactly once on the miss path. NEVER LoadOrStore with
// nt.OrderUpdates() as the eager argument — the argument evaluates FIRST and
// SubscribeOrderUpdatesFor closes+replaces the consumer's channel on every
// cycle (the 2026-08-27 consumer-death bug).
func (at *AutoTrader) armedUpdateStream(nt *ntTrader.TCPTrader) <-chan ntwire.OrderUpdatePayload {
	if v, ok := armedSubs.Load(at.id); ok {
		ch, _ := v.(<-chan ntwire.OrderUpdatePayload)
		return ch
	}
	ch := nt.OrderUpdates()
	v, _ := armedSubs.LoadOrStore(at.id, ch)
	stored, _ := v.(<-chan ntwire.OrderUpdatePayload)
	return stored
}

// cancelArmedOrdersSync cancels every non-terminal armed row for THIS trader
// with ack-waited wire cancels (one retry per order). Returns the rows
// cancelled and the rows whose ack never arrived (ledger flipped anyway).
func (at *AutoTrader) cancelArmedOrdersSync(reason string) (n, unacked int) {
	if at.store == nil {
		return 0, 0
	}
	if s := at.armedSyncSeam; s != nil {
		timeout := s.Timeout
		if timeout <= 0 {
			timeout = armedCancelAckTimeout()
		}
		return at.cancelArmedOrdersSyncWith(reason, timeout, s.Cancel, s.Stream)
	}
	nt := at.armedTrader()
	if nt == nil {
		return at.cancelArmedOrders(reason), 0
	}
	return at.cancelArmedOrdersSyncWith(reason, armedCancelAckTimeout(), nt.CancelOrder,
		func() <-chan ntwire.OrderUpdatePayload { return at.armedUpdateStream(nt) })
}

// cancelArmedOrdersSyncWith is the pure body: per-row cancel + ack drain. Every
// frame drained is applied through the SAME onArmedOrderUpdate the cycle
// consumer uses, so no ledger state is lost and no second subscription is ever
// made (a second subscribe would close the consumer's channel).
func (at *AutoTrader) cancelArmedOrdersSyncWith(reason string, timeout time.Duration, cancelFn func(string) error, src func() <-chan ntwire.OrderUpdatePayload) (n, unacked int) {
	ledger := at.store.ArmedOrders()
	if ledger == nil {
		return 0, 0
	}
	rows, err := ledger.ListNonTerminal(at.id)
	if err != nil {
		return 0, 0
	}
	var mine []store.ArmedOrderDB
	for _, r := range rows {
		if r.TraderID == at.id {
			mine = append(mine, r)
		}
	}
	for _, r := range mine {
		if r.State != "working" || r.SignalID == "" || cancelFn == nil || src == nil {
			_ = ledger.SetState(r.ID, "cancelled", reason)
			n++
			continue
		}
		acked := false
		for attempt := 1; attempt <= 2 && !acked; attempt++ {
			if err := cancelFn(r.SignalID); err != nil {
				at.logWarnf("⚠️ armed sync cancel send %s signal=%s failed: %v", r.Scenario, r.SignalID, err)
			}
			ch := src()
			if ch == nil {
				break
			}
			deadline := time.Now().Add(timeout)
			for !acked && ch != nil {
				remaining := time.Until(deadline)
				if remaining <= 0 {
					break
				}
				timer := time.NewTimer(remaining)
				select {
				case u, open := <-ch:
					timer.Stop()
					if !open {
						ch = nil // stream closed — the consumer self-heals next cycle
						break
					}
					at.onArmedOrderUpdate(u, ledger)
					acked = !at.armedRowStillActive(ledger, r.ID)
				case <-timer.C:
				}
			}
		}
		if acked {
			n++
		} else {
			_ = ledger.SetState(r.ID, "cancelled", reason+" (ack timeout — flatten proceeds)")
			unacked++
			at.logWarnf("⚠️ armed sync cancel UNACKED %s signal=%s after retry — ledger cancelled, flatten proceeds", r.Scenario, r.SignalID)
		}
	}
	return n, unacked
}

// armedRowStillActive reports whether the ledger row is still non-terminal.
func (at *AutoTrader) armedRowStillActive(ledger *store.ArmedOrderStore, id int64) bool {
	rows, err := ledger.ListNonTerminal(at.id)
	if err != nil {
		return true // unknown → keep waiting until the deadline
	}
	for _, r := range rows {
		if r.ID == id {
			return true
		}
	}
	return false
}

// ── E2 DEBUG SEAM (2026-08-27, level-truth wave ruling "a") ─────────────────
// POST /api/armed/test-arm drives the REAL placement path with a TEST-E2 row so
// the armed-orders cutover can be proven end-to-end (place → 📌 working in the
// NT8 Orders tab → cancel → ✕ chain) even when the planner can't produce an
// active plan. Gated by env ARMED_TEST_SEAM=on (default OFF) AND the bound
// account being SIM — a debug endpoint that places orders must not exist
// unarmed.

// armedTestSeamOn reads the env gate.
func armedTestSeamOn() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("ARMED_TEST_SEAM")))
	return v == "on" || v == "1" || v == "true"
}

// armedSeamStateLabel is the boot-line spelling of the seam state.
func armedSeamStateLabel() string {
	if armedTestSeamOn() {
		return "ON"
	}
	return "off"
}

// armedSeamDenied returns the blocker when the seam is gated off ("" = allowed).
func (at *AutoTrader) armedSeamDenied() string {
	if !armedTestSeamOn() {
		return "ARMED_TEST_SEAM is off"
	}
	if !strings.EqualFold(at.currentAccountName(), "Sim101") {
		return "seam is SIM-only (bound account " + at.currentAccountName() + ")"
	}
	return ""
}

// TestArmPlace places a resting limit on the REAL wire path (TCPTrader
// PlaceLimitEntry — the same call runArmedPlacement makes) with a ledger row
// tagged TEST-E2, skipping the price band (the tester pins the price).
func (at *AutoTrader) TestArmPlace(side string, entry, stop, target float64) (store.ArmedOrderDB, error) {
	var out store.ArmedOrderDB
	if reason := at.armedSeamDenied(); reason != "" {
		return out, fmt.Errorf("test-arm denied: %s", reason)
	}
	nt := at.armedTrader()
	if nt == nil {
		return out, fmt.Errorf("no TCPTrader bound")
	}
	ledger := at.store.ArmedOrders()
	if ledger == nil {
		return out, fmt.Errorf("no armed ledger")
	}
	side = strings.ToLower(strings.TrimSpace(side))
	if side != "long" && side != "short" {
		return out, fmt.Errorf("side must be long|short")
	}
	if entry <= 0 || stop <= 0 || target <= 0 {
		return out, fmt.Errorf("entry/stop/target must be > 0")
	}
	sid, perr := nt.PlaceLimitEntry(at.futuresSymbol(), side, 1, entry, stop, target)
	if perr != nil {
		return out, perr
	}
	row := &store.ArmedOrderDB{
		TraderID: at.id,
		PlanID:   "TEST-E2:" + sid,
		Session:  "TEST-E2",
		Scenario: "TEST-E2",
		Side:     side,
		EntryPx:  entry,
		StopPx:   stop,
		TargetPx: target,
	}
	if err := ledger.UpsertArm(row); err != nil {
		return out, fmt.Errorf("ledger upsert: %w", err)
	}
	_ = ledger.SetSignal(row.ID, sid)
	_ = ledger.SetState(row.ID, "working", "")
	row.SignalID = sid
	row.State = "working"
	at.logInfof("🧪 TEST-E2 arm → WORKING limit %.2f signal=%s (seam)", entry, sid)
	return *row, nil
}

// TestArmCancel cancels a seam row's NT8 order on the real wire and flips the
// row to cancelled with an honest reason.
func (at *AutoTrader) TestArmCancel(signalID string) error {
	if reason := at.armedSeamDenied(); reason != "" {
		return fmt.Errorf("test-arm denied: %s", reason)
	}
	nt := at.armedTrader()
	if nt == nil {
		return fmt.Errorf("no TCPTrader bound")
	}
	ledger := at.store.ArmedOrders()
	if ledger == nil {
		return fmt.Errorf("no armed ledger")
	}
	signalID = strings.TrimSpace(signalID)
	if signalID == "" {
		return fmt.Errorf("signal_id required")
	}
	if err := nt.CancelOrder(signalID); err != nil {
		return fmt.Errorf("cancel on wire: %w", err)
	}
	rows, _ := ledger.ListNonTerminal(at.id)
	for _, r := range rows {
		if r.TraderID == at.id && r.SignalID == signalID {
			_ = ledger.SetState(r.ID, "cancelled", "test seam cancel")
			at.logInfof("🧪 TEST-E2 cancel sent signal=%s (row %d → cancelled)", signalID, r.ID)
		}
	}
	return nil
}
