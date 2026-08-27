package trader

import (
	"fmt"
	"strings"
	"time"

	"nofx/kernel"
	"nofx/market"
	"nofx/store"
)

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
		if n := at.cancelArmedOrders(reason); n > 0 {
			at.logWarnf("🔒 armed cancel: %s — %d order(s) disarmed", reason, n)
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
		// gates AT ARM TIME — a resting order is a pre-passed entry; each gate
		// input that changes materially later triggers a cancel (1.3).
		if verdict := at.armGateVerdict(sc, doc.Bias.Direction, snap, atr5m, minQuality, cfg); verdict != "" {
			at.logWarnf("⚔️ arm REFUSED %s %s: %s", plan.Session, sc.ID, verdict)
			continue
		}
		side := strings.ToLower(strings.TrimSpace(sc.Direction))
		row := &store.ArmedOrderDB{
			TraderID: at.id, PlanID: plan.PlanID, Version: plan.Version, Session: plan.Session,
			Scenario: sc.ID, Side: side, EntryPx: sc.Arm.Entry, StopPx: sc.Arm.Stop, TargetPx: sc.Arm.Target,
			State: "armed", EntryClass: "armed_fill", CreatedAt: now, UpdatedAt: now,
		}
		existing, err := ledger.ListNonTerminal()
		if err == nil {
			for i := range existing {
				if existing[i].TraderID == at.id && existing[i].PlanID == row.PlanID && existing[i].Scenario == sc.ID {
					row.ID = existing[i].ID // already in the ledger — leave state (churn guard applies to placement)
					break
				}
			}
		}
		if row.ID == 0 {
			if err := ledger.UpsertArm(row); err != nil {
				at.logWarnf("⚔️ arm write failed %s %s: %v", plan.Session, sc.ID, err)
				continue
			}
			at.logInfof("⚔️ armed %s %s %s limit %.2f SL %.2f TP %.2f (tick-managed placement is Phase 2)", plan.Session, sc.ID, side, sc.Arm.Entry, sc.Arm.Stop, sc.Arm.Target)
		}
	}
}

// armGateVerdict runs the arm-time gate chain. Empty string = pass. The
// min-confidence gate is N/A for arms — the AI's authorization IS the
// confidence signal (no per-scenario confidence exists to check).
func (at *AutoTrader) armGateVerdict(sc kernel.PlanScenario, biasDirection string, snap map[string]kernel.StructureState, atr5m float64, minQuality string, cfg *store.StrategyConfig) string {
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
	// R:R gate — the same floor the entry path enforces.
	rr := 0.0
	if side == "long" && a.Entry > a.Stop && a.Stop > 0 {
		rr = (a.Target - a.Entry) / (a.Entry - a.Stop)
	} else if side == "short" && a.Stop > a.Entry && a.Entry > 0 {
		rr = (a.Entry - a.Target) / (a.Stop - a.Entry)
	}
	if cfg.RiskControl.MinRiskRewardRatio > 0 && rr+1e-9 < cfg.RiskControl.MinRiskRewardRatio {
		return fmt.Sprintf("R:R %.2f below min %.2f", rr, cfg.RiskControl.MinRiskRewardRatio)
	}
	// min-SL — the same floor (×ATR5m) the entry path enforces.
	if atr5m > 0 {
		dist := a.Entry - a.Stop
		if side == "short" {
			dist = a.Stop - a.Entry
		}
		if dist+1e-9 < kernel.MinSLATRMult()*atr5m {
			return fmt.Sprintf("stop %.2f too close (%.2f < %.2f = %.1f×ATR5m)", a.Stop, dist, kernel.MinSLATRMult()*atr5m, kernel.MinSLATRMult())
		}
	}
	// HTF veto — the same veto the entry path enforces.
	if blocked, vetoReason := kernel.HTFVetoVerdict(snap, "open_"+side, kernel.HTFVetoTF()); blocked {
		return "HTF veto: " + vetoReason
	}
	return ""
}

// cancelArmedOrders moves non-terminal rows for THIS trader to cancelled with a
// reason. Returns the count.
func (at *AutoTrader) cancelArmedOrders(reason string) int {
	rows, err := at.store.ArmedOrders().ListNonTerminal()
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
