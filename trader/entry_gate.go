package trader

import (
	"fmt"
	"strings"
	"time"

	"nofx/kernel"
	"nofx/market"
	"nofx/store"
	"nofx/telemetry"
)

// ── CLASS 48 — ONE ENTRY GATE FOR BOTH ORDER PATHS ──────────────────────────
//
// The five entry protections (R:R floor, shadow map 0C, scenario-direction
// consistency, min-SL ×ATR5m, one-live-arm) historically lived ONLY at the arm
// seam (armed_executor.go) while the decision path (auto_trader_orders.go →
// OpenLong market entry) ran a different, thinner chain and the agent chat
// path ran almost none. Live proof 2026-09-02: positions 587/589 filled BELOW
// the owner's 2.0 R:R floor at the real fill (the floor had been evaluated
// against the prompt-time snapshot), and 589/590 traded the SHADOWED condition
// breakout_retest, and 590 opened long citing a scenario that other versions
// authored short — all three admitted by the decision path.
//
// EntryGate is THE one canonical check. Both seams call it before any order
// leaves. The callers resolve every input (the function itself is pure and
// testable): the arm seam feeds arm-leg prices, the decision path feeds the
// LIVE market price at execution time (which is what actually fixes the
// snapshot mismatch — the floor is now judged at the price the order will
// transact at).
//
// Fail-open contract (mirrors validateDecision): a leg whose input is missing
// (no market price, no ATR, no plan, no shadow resolver) SKIPS. The gate
// refuses only on POSITIVE evidence, never on absence of evidence.

// EntryIntent is the resolved input to EntryGate.
type EntryIntent struct {
	Path   string // "arm" | "decision" — recorded per path on refusal
	Action string // "open_long" | "open_short"
	Symbol string

	Entry  float64 // the price the entry will actually transact at (arm leg entry / live market price)
	Stop   float64
	Target float64

	ATR5m     float64
	MinRR     float64 // resolved min_risk_reward_ratio (decision path) or ARM_MIN_RR (arm seam)
	MinSLMult float64 // resolved MIN_SL_ATR_MULT (0 = leg off)

	// Plan context (all optional — legs skip on absence).
	PlanBias      string // resolved plan bias ("long"/"short"/"")
	PlanMode      string // "advisory" | "direction" | "strict"
	CitedScenario string // the scenario the intent cites ("" = none)
	ScenarioDir   string // the cited scenario's direction ("long"/"short"/"")
	ScenarioCond  string // the cited scenario's condition ("" = none)

	// One-live-arm input: the currently OPEN position side ("" = flat).
	OpenPositionSide string

	// Shadow resolver (nil = shadow leg off, fail-open).
	ConditionShadowed func(condition string) bool
}

// EntryGate runs the single canonical entry gate chain. Empty reason = allow.
func EntryGate(in EntryIntent) (reason string, refused bool) {
	side := strings.ToLower(strings.TrimSpace(in.Action))
	if side == "open_long" {
		side = "long"
	} else if side == "open_short" {
		side = "short"
	}
	if side != "long" && side != "short" {
		return "", false // not an entry intent — not this gate's job
	}

	// Leg 1 — plan bias (direction mode only; the arm seam's legacy chain and
	// the decision path's planModeBlocked both already enforce this in
	// direction mode — kept here so the ONE gate is complete on its own).
	if in.PlanMode == "direction" {
		bias := strings.ToLower(strings.TrimSpace(in.PlanBias))
		if bias != "" && bias != side {
			return fmt.Sprintf("entry_gate: %s entry against plan bias %q (plan_mode=direction)", side, bias), true
		}
	}

	// Leg 2 — scenario-direction consistency (CLASS 48 core): an entry must
	// match the direction of the scenario it CITES. The decision path's
	// recordPlanCitation only ever logged this ("advisory, never gates") —
	// position 590-class refusals were impossible there.
	if in.CitedScenario != "" {
		dir := strings.ToLower(strings.TrimSpace(in.ScenarioDir))
		if dir == "long" || dir == "short" {
			if dir != side {
				return fmt.Sprintf("entry_gate: %s entry cites scenario %s authored %s — direction mismatch (class 48)", side, in.CitedScenario, dir), true
			}
		}
	}

	// Leg 3 — shadow map (0C owner ruling 2026-08-31): a shadowed condition is
	// authored + E8-scored but NEVER placed, on ANY path. The decision path had
	// no copy of this check — 589 and 590 both traded breakout_retest.
	if in.ScenarioCond != "" && in.ConditionShadowed != nil && in.ConditionShadowed(in.ScenarioCond) {
		return fmt.Sprintf("entry_gate: scenario %s condition %s is SHADOW (0C) — authored + E8-scored, never placed on any path", in.CitedScenario, in.ScenarioCond), true
	}

	// Leg 4 — R:R at the REAL entry price (the fix for 587/589): the floor is
	// judged at the price the order will transact at, not the prompt snapshot.
	// entry<=0 or stop<=0 → skip (fail-open; validateDecision owns wrong-side).
	rr := 0.0
	ok := false
	if in.Entry > 0 && in.Stop > 0 && in.Target > 0 && in.Entry != in.Stop {
		if side == "long" && in.Entry > in.Stop {
			rr = (in.Target - in.Entry) / (in.Entry - in.Stop)
			ok = true
		} else if side == "short" && in.Stop > in.Entry {
			rr = (in.Entry - in.Target) / (in.Stop - in.Entry)
			ok = true
		}
	}
	if ok {
		floor := in.MinRR
		if floor <= 0 {
			floor = 3.0 // same fallback validateDecision uses for an unset knob
		}
		if rr+1e-9 < floor {
			return fmt.Sprintf("entry_gate: R:R %.2f below floor %.2f at execution price %.4f (SL %.4f TP %.4f)", rr, floor, in.Entry, in.Stop, in.Target), true
		}
	}

	// Leg 5 — min-SL ×ATR5m (same floor as both legacy chains).
	if in.ATR5m > 0 && in.MinSLMult > 0 && in.Entry > 0 && in.Stop > 0 {
		dist := in.Entry - in.Stop
		if side == "short" {
			dist = in.Stop - in.Entry
		}
		if dist+1e-9 < in.MinSLMult*in.ATR5m {
			return fmt.Sprintf("entry_gate: stop %.2f too close (%.2f < %.2f = %.1f×ATR5m)", in.Stop, dist, in.MinSLMult*in.ATR5m, in.MinSLMult), true
		}
	}

	// Leg 6 — one-live-arm (class 27 FIX 4 semantics, applied to BOTH paths):
	// an opposite-side entry while a position is open NETS it on a netting
	// account — refuse. Same-side add is outside this guard's scope.
	if in.OpenPositionSide != "" {
		open := strings.ToLower(strings.TrimSpace(in.OpenPositionSide))
		if (open == "long" || open == "short") && open != side {
			return fmt.Sprintf("entry_gate: %s entry would net the open %s position (one_live_arm_guard, class 27)", side, open), true
		}
	}

	return "", false
}

// ── Arm-seam builder ────────────────────────────────────────────────────────

// entryGateForArm builds the intent for an arm leg and runs EntryGate. The arm
// chain's own gates (armGateVerdictFor, oneLiveArmGuard) run before this —
// EntryGate is the SAME function the decision path runs, so an arm can never
// be held to a weaker standard than a market entry.
func (at *AutoTrader) entryGateForArm(plan *kernel.ActivePlan, sc kernel.PlanScenario, leg kernel.PlanArmLeg, side, biasDir string, atr5m float64) (string, bool) {
	openSide := ""
	// class-27 escape, mirrored from oneLiveArmGuard: a leg explicitly authored
	// as an exit/flip leg for the open position is allowed to be opposite-side.
	if !strings.EqualFold(strings.TrimSpace(leg.Kind), "exit") && at.store != nil {
		opens, err := at.store.Position().GetOpenPositions(at.id)
		if err == nil {
			sym := market.Normalize(at.futuresSymbol())
			for _, p := range opens {
				if strings.EqualFold(p.Symbol, sym) && (p.Side == "long" || p.Side == "short") {
					openSide = strings.ToLower(p.Side)
					break
				}
			}
		}
	}
	session := plan.Session
	return EntryGate(EntryIntent{
		Path:              "arm",
		Action:            "open_" + side,
		Symbol:            at.futuresSymbol(),
		Entry:             leg.Entry,
		Stop:              leg.Stop,
		Target:            leg.Target,
		ATR5m:             atr5m,
		MinRR:             armMinRR(),
		MinSLMult:         kernel.MinSLATRMult(),
		PlanBias:          biasDir,
		PlanMode:          at.planModeFor(session),
		CitedScenario:     sc.ID,
		ScenarioDir:       sc.Direction,
		ScenarioCond:      sc.Condition,
		OpenPositionSide:  openSide,
		ConditionShadowed: func(cond string) bool { return at.conditionShadowedFor(cond, session) },
	})
}

// ── Decision-path builder ───────────────────────────────────────────────────

// entryGateForDecision builds the intent for an AI market entry and runs
// EntryGate. livePrice is the execution-time market price (the caller resolves
// it; ≤0 skips the R:R/min-SL legs, fail-open). All plan inputs resolve from
// the trader's ActivePlan; absent plan → those legs skip.
func (at *AutoTrader) entryGateForDecision(d *kernel.Decision, livePrice float64) (string, bool) {
	if d == nil {
		return "", false
	}
	openSide := ""
	if at.store != nil {
		opens, err := at.store.Position().GetOpenPositions(at.id)
		if err == nil {
			sym := market.Normalize(d.Symbol)
			for _, p := range opens {
				if strings.EqualFold(p.Symbol, sym) && (p.Side == "long" || p.Side == "short") {
					openSide = strings.ToLower(p.Side)
					break
				}
			}
		}
	}
	minRR := 3.0
	if at.config.StrategyConfig != nil && at.config.StrategyConfig.RiskControl.MinRiskRewardRatio > 0 {
		minRR = at.config.StrategyConfig.RiskControl.MinRiskRewardRatio
	}
	now := time.Now()
	session := ""
	if s, ok := at.sessionRegistry(now).ActiveSession(now); ok {
		session = s.Name
	}
	intent := EntryIntent{
		Path:             "decision",
		Action:           d.Action,
		Symbol:           d.Symbol,
		Entry:            livePrice,
		Stop:             d.StopLoss,
		Target:           d.TakeProfit,
		ATR5m:            kernel.PlanDATRFor(at.id),
		MinRR:            minRR,
		MinSLMult:        kernel.MinSLATRMult(),
		PlanMode:         at.planModeFor(session),
		CitedScenario:    strings.TrimSpace(d.CitedScenario),
		OpenPositionSide: openSide,
	}
	if kernel.HasTraderPlanProvider(at.id) {
		if ap := kernel.ActivePlanFor(at.id, at.futuresSymbol()); ap != nil {
			intent.PlanBias = strings.ToLower(strings.TrimSpace(ap.Doc.Bias.Direction))
			for _, sc := range ap.Doc.Scenarios {
				if sc.ID == intent.CitedScenario {
					intent.ScenarioDir = strings.ToLower(strings.TrimSpace(sc.Direction))
					intent.ScenarioCond = sc.Condition
					break
				}
			}
		}
	}
	if intent.ScenarioCond != "" {
		intent.ConditionShadowed = func(cond string) bool { return at.conditionShadowedFor(cond, session) }
	}
	return EntryGate(intent)
}

// recordEntryGateRefusal is the per-path refusal RECORD. The arm seam records
// into the arm-refusal counter family (per session-day, per class); the
// decision path records into decision_records via actionRecord + the gate-block
// counter. Both log.
func (at *AutoTrader) recordEntryGateRefusal(path, symbol, action, reason string, plan *kernel.ActivePlan) {
	class := armRefusalClass(reason)
	if class == "" {
		class = "entry_gate"
	}
	if path == "arm" && plan != nil && at.store != nil {
		if n, cerr := store.IncArmRefusal(at.store, at.id, kernel.PlanTradeDateFor(plan), plan.Session, "entry_gate:"+class); cerr == nil {
			at.logWarnf("🚦 entry-gate REFUSED arm %s: %s · refusals this session: %d", plan.Session, reason, n)
			return
		}
	}
	at.logWarnf("🚦 entry-gate REFUSED %s %s %s: %s", path, symbol, action, reason)
}

// entryGateDecisionTelemetry stamps the decision-path refusal (recorded in
// decision_records through actionRecord.Error + execution_log).
func entryGateDecisionTelemetry(at *AutoTrader, actionRecord *store.DecisionAction, reason string) {
	telemetry.IncGateBlock(at.id, "entry_gate")
	actionRecord.Success = false
	actionRecord.Error = "entry_gate: " + reason
}
