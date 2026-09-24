package trader

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"nofx/kernel"
	"nofx/market"
	"nofx/store"
)

// ── W1b E1 + E2 — a WORKING arm under a plan that re-priced it ─────────────
//
// E1 (the dead churn guard). The arm loop built its row from the NEW leg and
// copied only the id from the ledger, then asked whether the row was working
// and whether its stop differed from the new stop — it never was, and it never
// did. A re-spec'd plan therefore never moved a resting order's bracket. The
// guard's ModifyBracket call could not have helped either: the AddOn's
// modify_bracket acts only on a FILLED entry's bracket (placedBrackets), and a
// working row is by definition an unfilled entry, so the frame is refused
// ("no live bracket") and nothing in Go reads the reply.
//
// E2 (the moved zone). A working market_in_zone limit is keyed by (plan,
// scenario, leg), which survives a re-plan, but the zone that justified its
// price does not. Nothing compared the resting limit with the CURRENT zone, so
// only the 30-minute rest cap ever ended it.
//
// The one fix for both: CANCEL, then let the normal authoring re-arm. The
// cancel goes through the filled-arm guard (cancelSafetyFor) — a refusal sends
// nothing, WARNs once, and is retried on the next pass. The row moves to
// cancel_pending (a send is not a settlement); once the broker's persisted
// book confirms it, the next pass finds no live row, UpsertArm mints the next
// placement (seq+1) under the new version — the D15 pin blocks the mint only
// within the version the old row was placed under — and every existing gate
// judges the new placement.
//
// AT MOST ONE CANCEL PER ROW: E2 is asked first, E1 only when E2 is silent.

// armRespecCancel is one decided re-spec: the ledger reason, the counter it is
// recorded under, and the short class the WARN names.
type armRespecCancel struct {
	reason  string
	counter string
	class   string
}

// armRespecFor decides whether a working planner row must be replaced. PURE.
//
//   - E2: a market_in_zone row whose resting limit lies outside the zone the
//     current plan authorizes. NOT version-gated: within one zone the far edge
//     is always contained, so this is silent when nothing moved — and an
//     overlay that moves the zone inside the SAME version also cancels (the
//     D15 pin then keeps that version from re-arming: fail-closed, named).
//   - E1: a row whose placed stop or target differs by ≥ 2 ticks from the
//     current leg's, under a NEWER plan version. Version-gated: the composed
//     stop drifts with live ATR inside one version, and that drift is not a
//     re-spec (it would cancel and re-place the order every few minutes).
//
// Only a working row with a signal is ever judged; a Picture row (Source set)
// belongs to its own pin and deadline (pictureDeadlineCancel).
func armRespecFor(planVersion int, prior store.ArmedOrderDB, leg kernel.PlanArmLeg, zl zoneLeg, tick float64) (armRespecCancel, bool) {
	if prior.State != store.StateWorking || strings.TrimSpace(prior.SignalID) == "" || prior.Source != "" {
		return armRespecCancel{}, false
	}
	if tick <= 0 {
		tick = 0.25
	}
	const eps = 1e-9
	if zl.on && prior.Policy == kernel.EntryPolicyMarketInZone &&
		(prior.EntryPx < zl.v.Lo-eps || prior.EntryPx > zl.v.Hi+eps) {
		return armRespecCancel{
			reason:  fmt.Sprintf("zone moved by v%d: limit %.2f outside %.2f–%.2f", planVersion, prior.EntryPx, zl.v.Lo, zl.v.Hi),
			counter: "market_in_zone:zone_moved",
			class:   "zone moved",
		}, true
	}
	if planVersion > prior.Version && churnNeedsModify(prior.StopPx, prior.TargetPx, leg.Stop, leg.Target, tick) {
		return armRespecCancel{
			reason:  fmt.Sprintf("bracket re-spec by v%d: SL %.2f→%.2f TP %.2f→%.2f", planVersion, prior.StopPx, leg.Stop, prior.TargetPx, leg.Target),
			counter: "arm:respec_cancel",
			class:   "bracket re-spec",
		}, true
	}
	return armRespecCancel{}, false
}

// respecWorkingArm is the arm loop's call for a leg that already has a live
// ledger row (prior) and just passed every authoring gate under the current
// plan. true = a cancel was REQUESTED this pass (the row is cancel_pending).
func (at *AutoTrader) respecWorkingArm(ledger *store.ArmedOrderStore, plan *kernel.ActivePlan, sc kernel.PlanScenario, li int, prior store.ArmedOrderDB, leg kernel.PlanArmLeg, zl zoneLeg, now time.Time) bool {
	if ledger == nil || plan == nil {
		return false
	}
	c, ok := armRespecFor(plan.Version, prior, leg, zl, market.FuturesTickSize(at.futuresSymbol()))
	if !ok {
		return false
	}
	nt := at.armedTrader()
	if nt == nil {
		return false
	}
	key := plan.PlanID + ":" + strconv.Itoa(plan.Version) + ":" + sc.ID + ":leg" + strconv.Itoa(li+1) + ":respec"
	if v := at.cancelSafetyFor(prior, now); !v.Allow {
		if armRefusalChanged(&at.armRefusalLast, key, "refused: "+v.Why) {
			at.logWarnf("🛟 armed cancel REFUSED (%s): %s %s leg %d signal=%s — %s; retried next pass",
				c.class, plan.Session, sc.ID, li+1, shortID(prior.SignalID), v.Why)
		}
		return false
	}
	if cerr := nt.CancelOrder(prior.SignalID); cerr != nil {
		at.logWarnf("✕ armed cancel SEND failed (%s): %s %s leg %d signal=%s: %v", c.class, plan.Session, sc.ID, li+1, shortID(prior.SignalID), cerr)
	}
	if err := ledger.RequestCancel(prior.ID, c.reason, now.UnixMilli()); err != nil {
		at.logWarnf("✕ armed cancel (%s): ledger write failed for %s leg %d: %v", c.class, sc.ID, li+1, err)
		return false
	}
	shown := ""
	if at.store != nil {
		if n, err := store.IncSystemCounter(at.store, c.counter); err == nil {
			shown = fmt.Sprintf(" · %s recorded: %d", c.counter, n)
		}
	}
	if armRefusalChanged(&at.armRefusalLast, key, "requested") {
		at.logWarnf("✕ armed cancel REQUESTED (%s): %s %s leg %d signal=%s — pending broker confirmation; re-arms under v%d once the book confirms%s",
			c.reason, plan.Session, sc.ID, li+1, shortID(prior.SignalID), plan.Version, shown)
	}
	return true
}
