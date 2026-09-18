package trader

import (
	"fmt"
	"strings"

	"nofx/kernel"
	"nofx/market"
	"nofx/store"
)

// W-WRITE-TIME-FEASIBILITY (2026-09-18, owner ruling "fix all" 08:3x CT).
//
// The planner wrote scenarios the executor's gate-at-arm chain would refuse and
// only WARNed (kernel.ArmFeasibilityWarnings). The model never learned why its
// plan did not trade and re-authored the same shape. This wave judges the SAME
// predicates the executor runs at arm — via the executor's own functions (canon
// 53, no re-implementation) — BEFORE the plan is written:
//
//   - attempts 1..N-1: unarmable scenarios become a restriction-with-hint error
//     that feeds the existing repair prompt (class 38 machinery);
//   - the last attempt: the scenarios are written with arm.enabled=false +
//     arm_disabled_reason + ONE WARN each + a system counter.
//
// Knob `day_plan.write_time_feasibility` — nil/unset = ON (owner default);
// explicit false = today's WARN-only behaviour byte-identical (the verdict
// function returns nil and the prompt sentence is absent).

// writeTimeFeasibilityOn resolves the knob the way the trading path does
// (store.DayPlanConfig.WriteTimeFeasibilityEnabled): nil config → ON.
func (at *AutoTrader) writeTimeFeasibilityOn() bool {
	if at == nil || at.config.StrategyConfig == nil || at.config.StrategyConfig.DayPlan == nil {
		return true
	}
	return at.config.StrategyConfig.DayPlan.WriteTimeFeasibilityEnabled()
}

// writeFeasLabel renders the boot-line word from the RESOLVED knob (read,
// never typed).
func writeFeasLabel(dp *store.DayPlanConfig) string {
	if dp.WriteTimeFeasibilityEnabled() {
		return "on"
	}
	return "off"
}

type writeTimeFeasibilityIssue struct {
	Scenario string
	Reason   string
	// Kind is "gate" (armGateVerdictFor / geometry) or "stop_side" (the
	// executor's stop-side placement guard, CTO amendment 2026-09-18).
	Kind    string
	Cond    string  // scenario condition, for the stop-side hint words
	Trigger float64 // stop-side only: the tick-rounded trigger the wire would carry
	Price   float64 // stop-side only: the read-time price the guard judged
	Side    string  // stop-side only: canonical lowercase side
}

// writeTimeFeasibilityVerdicts runs the executor's own gate-at-arm predicates
// for every enabled single arm in the doc, at write time. The session-risk band
// is time-based and deliberately NOT judged here (spec). price is the read-time
// price the planner already has (facts.Price) — the stop-side guard judges the
// SAME value the placement branch will judge at arm time. Returns nil when the
// knob is OFF — byte-identical to today.
func (at *AutoTrader) writeTimeFeasibilityVerdicts(d *kernel.PlanDoc, atr5m float64, cfg *store.StrategyConfig, session string, price float64) []writeTimeFeasibilityIssue {
	if !at.writeTimeFeasibilityOn() || d == nil {
		return nil
	}
	minQuality := ""
	if cfg != nil && cfg.DayPlan != nil {
		minQuality = cfg.DayPlan.MinGradeFor(session)
	}
	bias := biasDirectionFor(d.Bias.Direction)
	var out []writeTimeFeasibilityIssue
	for i := range d.Scenarios {
		sc := &d.Scenarios[i]
		if sc.Arm == nil || !sc.Arm.Enabled {
			continue
		}
		leg := kernel.PlanArmLeg{Entry: sc.Arm.Entry, Stop: sc.Arm.Stop, Target: sc.Arm.Target}
		// The executor applies the structural-geometry path only to the
		// level-fade play and skips the raw-leg min-SL check there (the stop is
		// recomposed at arm). Mirror both, exactly as armed_executor.go does.
		structuralFade := sc.Condition == kernel.OneSetupPlay && !strings.EqualFold(leg.Kind, "exit")
		if v := at.armGateVerdictFor(*sc, leg, bias, nil, atr5m, minQuality, cfg, session, structuralFade); v != "" {
			out = append(out, writeTimeFeasibilityIssue{Scenario: sc.ID, Reason: v, Kind: "gate"})
			continue
		}
		if structuralFade {
			if _, reason := ResolveEntryGeometryZone(d, *sc); reason != "" {
				out = append(out, writeTimeFeasibilityIssue{Scenario: sc.ID, Reason: "geometry: " + reason, Kind: "gate"})
			}
		}
		// STOP-SIDE (CTO amendment 2026-09-18) — the executor's OWN placement
		// guard, adjudicated here with the same function and the same inputs the
		// placement branch uses (armed_executor.go decideStopEntry: entry ±
		// STOP_ENTRY_OFFSET_TICKS, tick-rounded). A stop-entry trigger already
		// through the read-time price cancels at placement (`✕ armed …
		// stop-entry CANCELLED [guard=stop-side …]`) — the model must be told at
		// write instead of authoring a dead trigger 21× in a row.
		if kind, _ := armLegKindFor(*sc, kernel.PlanArmLeg{}); kind == kernel.ArmKindStopEntry {
			tick := market.FuturesTickSize(at.futuresSymbol())
			dec := decideStopEntry(sc.Direction, leg.Entry, float64(stopEntryOffsetTicks())*tick, tick, price)
			if dec.Verdict == stopGuardThrough {
				out = append(out, writeTimeFeasibilityIssue{
					Scenario: sc.ID, Reason: "stop_side_wrong", Kind: "stop_side",
					Cond: sc.Condition, Trigger: dec.Trigger, Price: price, Side: dec.Side,
				})
			}
		}
	}
	return out
}

// writeTimeFeasibilityHint renders the restriction-with-hint text the repair
// prompt carries: the refusal, the numbers, and the vocabulary to fix it.
func writeTimeFeasibilityHint(issues []writeTimeFeasibilityIssue) string {
	var b strings.Builder
	b.WriteString("write-time feasibility: ")
	gateIssues := 0
	for i, is := range issues {
		if i > 0 {
			b.WriteString("; ")
		}
		if is.Kind == "stop_side" {
			// The amendment's own vocabulary (CTO 2026-09-18): name the
			// trigger, its relation to price, and the two fixes.
			aboveBelow := "above"
			if strings.EqualFold(is.Side, "long") {
				aboveBelow = "below"
			}
			b.WriteString(fmt.Sprintf("%s %s trigger %.2f is already %s price %.2f: a stop entry there fills at market on placement — author the trigger ahead of price, or author a reject/limit at the level", is.Scenario, is.Cond, is.Trigger, aboveBelow, is.Price))
			continue
		}
		gateIssues++
		b.WriteString(fmt.Sprintf("%s would be refused at arm — %s", is.Scenario, is.Reason))
	}
	if gateIssues > 0 {
		b.WriteString(" — widen the stop past the min-SL floor / raise the arm R:R / pick a mapped level with an id / choose a different condition")
	}
	return b.String()
}

// applyWriteTimeArmDisable is the LAST-attempt path: scenarios that are still
// unarmable are written with arm.enabled=false + arm_disabled_reason, ONE WARN
// each, and a counter recorded (counters record, never infer).
func (at *AutoTrader) applyWriteTimeArmDisable(d *kernel.PlanDoc, issues []writeTimeFeasibilityIssue, tradeDate, session string) {
	if d == nil || len(issues) == 0 {
		return
	}
	byID := make(map[string]string, len(issues))
	for _, is := range issues {
		byID[is.Scenario] = is.Reason
	}
	for i := range d.Scenarios {
		sc := &d.Scenarios[i]
		if sc.Arm == nil || !sc.Arm.Enabled {
			continue
		}
		reason, ok := byID[sc.ID]
		if !ok {
			continue
		}
		sc.Arm.Enabled = false
		sc.Arm.DisabledReason = reason
		at.logWarnf("⚔️ arm disabled at write: %s %s %s", session, sc.ID, reason)
		if at.store == nil {
			continue
		}
		key := fmt.Sprintf("arm_disabled_at_write:%s:%s:%s:%s", at.id, tradeDate, session, reason)
		if _, err := store.IncSystemCounter(at.store, key); err != nil {
			at.logWarnf("⚔️ arm-disabled counter write failed: %v", err)
		}
	}
}
