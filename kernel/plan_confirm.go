package kernel

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"nofx/logger"
	"nofx/market"
)

// C1 (F3) — machine evaluation of a scenario's structured confirmation, using
// the SAME acceptance machinery as plan death (A2-consistent rule semantics).

// confirmAcceptanceRule maps the confirm vocabulary onto acceptance rules.
func confirmAcceptanceRule(rule string) string {
	switch rule {
	case "15m_close":
		return "15m-close"
	case "1x5m_close":
		return "5m-close" // the A2-fixed one-close rule
	default:
		return "2x5m" // "2x5m_close"
	}
}

// ConfirmVerdict is one scenario's machine-computed confirmation state.
type ConfirmVerdict struct {
	Rule     string           `json:"rule"`
	RefPrice float64          `json:"ref_price"`
	Side     string           `json:"side"`
	Met      bool             `json:"met"`
	Detail   string           `json:"detail"`         // e.g. "last 15m close 29641.00"
	Legs     []ConfirmVerdict `json:"legs,omitempty"` // F2: the per-leg states of a two-leg confirm
}

// ConfirmVerdict.Met on a two-leg scenario is the OVERALL verdict (leg1 &&
// leg2); a partial never reports Met=true.

// EvaluateConfirm computes MET/NOT-MET for one confirm object over the bars
// since the plan's birth (touch-gated like plan death; windowed identically).
func EvaluateConfirm(c PlanConfirm, bars []market.Kline, sinceMs, nowMs int64) ConfirmVerdict {
	v := ConfirmVerdict{Rule: c.Rule, RefPrice: c.RefPrice, Side: c.Side}
	w := BarsSince(bars, sinceMs)
	if len(w) == 0 {
		v.Detail = "no bars yet"
		return v
	}
	if c.Rule == "touch" {
		for i := range w {
			if w[i].Low <= c.RefPrice && w[i].High >= c.RefPrice {
				v.Met = true
				v.Detail = "level touched"
				return v
			}
		}
		v.Detail = "not touched since plan birth"
		return v
	}
	rule := confirmAcceptanceRule(c.Rule)
	above := strings.EqualFold(c.Side, "above")
	// EVER-fired semantics via the sanctioned facts API (the acceptance-interval
	// guard forbids raw counting outside scenario_facts.go).
	best, need, lastClose := AcceptanceRunEver(w, rule, c.RefPrice, above)
	if lastClose == 0 && best == 0 {
		v.Detail = "no closed bars at the rule timeframe yet"
		return v
	}
	v.Met = best >= need
	v.Detail = fmt.Sprintf("last %s close %.2f (best run %d/%d closes %s %.2f since plan birth)",
		strings.TrimSuffix(c.Rule, "_close"), lastClose, best, need, c.Side, c.RefPrice)
	return v
}

// StaleConfirmATR is the staleness threshold in 5m Wilder ATR14 multiples (env
// STALE_CONFIRM_ATR, default 2.0). Citation: register S2 (mega-research
// 2026-08-26) — the shipped 1.0×dATR unit (~350pt daily-range proxy) marked
// only 38/2,908 (1.3%) of the week's MET confirms stale; at 2.0×ATR5m (~40-70
// pt) the rule marks ~37%, matching the empirical stale mass (median |price−ref|
// = 58.75 pt). The old dATR path is DELETED — this is ATR-only.
func StaleConfirmATR() float64 {
	if v := os.Getenv("STALE_CONFIRM_ATR"); v != "" {
		if n, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && n > 0 {
			return n
		}
	}
	return 2.0
}

// staleConfirmATRLogged guards the fail-open log (once per process) so a
// missing 5m ATR doesn't spam the prompt path every cycle.
var staleConfirmATRLogged bool

// StaleConfirmATR5m computes the Wilder ATR14 on 5m buckets of the 1m snapshot
// — the same math the min-SL gate uses via the 5m structure ATR. 0 = unavailable.
func StaleConfirmATR5m(bars []market.Kline) float64 {
	if len(bars) == 0 {
		return 0
	}
	five := AggregateBars(bars, 5*60_000)
	if len(five) < 14 {
		return 0
	}
	return market.ExportCalculateATR(five, 14)
}

// staleConfirmAnnotation builds the "(stale — …)" parenthetical for a MET
// confirm whose ref context has drifted > STALE_CONFIRM_ATR × ATR5m from the
// current price since plan write (register S2: ATR-only, the dATR path is gone).
// The context is the confirm's own RefPrice (validated into the prose at write
// time), overridden by the FVG distal anchor when the scenario is an fvg_entry
// play. FAIL-OPEN: a missing 5m ATR skips the annotation (logged once), never
// gates. Advisory text only — the AI stays the judge.
func staleConfirmAnnotation(s PlanScenario, v ConfirmVerdict, nowPrice, atr5m float64) string {
	if !v.Met || nowPrice <= 0 {
		return ""
	}
	if atr5m <= 0 {
		if !staleConfirmATRLogged {
			staleConfirmATRLogged = true
			logger.Warnf("⚠️ stale-confirm annotation skipped this cycle: 5m ATR unavailable (fail-open)")
		}
		return ""
	}
	ctx := v.RefPrice
	if s.Fvg != nil && (s.Fvg.Lo > 0 || s.Fvg.Hi > 0) {
		if a, ok := ScenarioAnchor(s, nil); ok {
			ctx = a
		}
	}
	if ctx <= 0 {
		return ""
	}
	if math.Abs(nowPrice-ctx) > StaleConfirmATR()*atr5m {
		return fmt.Sprintf("(stale — written %.2f context, price now %.2f; treat as expired)", ctx, nowPrice)
	}
	return ""
}

// EvaluateScenarioConfirm computes the OVERALL confirm state of one scenario.
// Two-leg scenarios (Confirm2, or the waterfall-class breakdown plays whose
// legs are machine-derived) report each leg and an aggregate: leg 1 MET + leg 2
// NOT MET renders as overall NOT MET — a partial never prints as a bare "MET"
// (F2, the S2 10:54 artifact). Leg 2 is windowed from leg 1's FIRST fire time
// (a retest touch that happened BEFORE the breakdown cannot count).
func EvaluateScenarioConfirm(s PlanScenario, bars []market.Kline, sinceMs, nowMs int64) ConfirmVerdict {
	if IsBreakdownCondition(s.Condition) && s.Breakdown != nil {
		st := BreakdownContinueState(s, bars, sinceMs, nowMs)
		leg1 := ConfirmVerdict{Rule: fmt.Sprintf("%dx5m_close", bdConfirmCloses()), RefPrice: s.Breakdown.Level,
			Side: map[bool]string{true: "below", false: "above"}[breakdownShort(s.Condition)], Met: st.Leg1Met,
			Detail: fmt.Sprintf("best run %d closes beyond %.2f", 0, s.Breakdown.Level)}
		if st.Leg1Met {
			leg1.Detail = fmt.Sprintf("displacement %.2f pts, no reclaim", st.BreakLegPts)
		}
		v := ConfirmVerdict{Met: st.Leg1Met && st.Leg2Met, Legs: []ConfirmVerdict{leg1, {
			Rule: "retest_fail", RefPrice: s.Breakdown.Level, Side: leg1.Side,
			Met: st.Leg2Met, Detail: retestLegDetail(s, st),
		}}}
		v.Rule, v.RefPrice, v.Side = leg1.Rule, leg1.RefPrice, leg1.Side
		v.Detail = retestLegDetail(s, st)
		return v
	}
	if s.Confirm == nil {
		return ConfirmVerdict{}
	}
	v1 := EvaluateConfirm(*s.Confirm, bars, sinceMs, nowMs)
	if s.Confirm2 == nil {
		return v1
	}
	// Leg 2 is windowed from leg 1's first fire (the retest leg cannot be
	// satisfied by touches/closes that happened before the breakdown).
	since2 := firstConfirmFireMs(*s.Confirm, bars, sinceMs, nowMs)
	if since2 <= 0 {
		since2 = sinceMs
	}
	v2 := EvaluateConfirm(*s.Confirm2, bars, since2, nowMs)
	v2.Met = v1.Met && v2.Met // leg 2 only counts once leg 1 is satisfied (ordered legs)
	return ConfirmVerdict{Rule: v1.Rule, RefPrice: v1.RefPrice, Side: v1.Side,
		Met: v1.Met && v2.Met, Detail: v2.Detail, Legs: []ConfirmVerdict{v1, v2}}
}

// firstConfirmFireMs returns the open time of the bar where the close-rule
// confirm FIRST fired (the run reached the required count), 0 when not met.
func firstConfirmFireMs(c PlanConfirm, bars []market.Kline, sinceMs, nowMs int64) int64 {
	if c.Rule == "touch" {
		return 0
	}
	w := BarsSince(bars, sinceMs)
	if len(w) == 0 {
		return 0
	}
	rule := confirmAcceptanceRule(c.Rule)
	var dur, need int64
	switch rule {
	case "5m-close":
		dur, need = 5*60_000, 1
	case "2x5m":
		dur, need = 5*60_000, 2
	case "15m-close":
		dur, need = 15*60_000, 1
	default:
		return 0
	}
	buckets := AggregateBars(w, dur)
	above := strings.EqualFold(c.Side, "above")
	run := int64(0)
	for _, b := range buckets {
		if b.CloseTime > nowMs {
			continue
		}
		beyond := (above && b.Close > c.RefPrice) || (!above && b.Close < c.RefPrice)
		if beyond {
			run++
			if run >= need {
				return b.OpenTime
			}
		} else {
			run = 0
		}
	}
	return 0
}

// retestLegDetail renders the leg-2 state of a waterfall-class play.
func retestLegDetail(s PlanScenario, st BreakdownState) string {
	if st.Reclaimed {
		return "reclaimed — the breakdown is void"
	}
	if !st.Leg1Met {
		return "waiting for the breakdown leg"
	}
	if strings.EqualFold(strings.TrimSpace(s.Breakdown.EntryMode), "immediate") {
		return "immediate mode — entry signal is the 2nd confirming close"
	}
	if st.Leg2Met {
		return "pullback failed to reclaim — entry live"
	}
	return "retest not yet failed"
}

// RenderConfirmLines renders the per-scenario advisory lines for the executor
// prompt: machine truth the AI reasons FROM — never a gate.
//
// ADDENDUM S (2026-08-26, quiet-day audit) — two prompt-text annotations on
// top, still zero gates:
//  1. STALE: a MET confirm whose ref context is now > STALE_CONFIRM_ATR × dATR
//     from price is annotated "MET (stale — written X context, price now Y;
//     treat as expired)".
//  2. CONFLICT: opposite-direction confirms MET in the same cycle get one
//     trailing line "CONFLICT: opposing confirms MET — structural ambiguity,
//     default WAIT unless fresh trigger".
func RenderConfirmLines(doc PlanDoc, bars []market.Kline, sinceMs, nowMs int64, nowPrice, atr5m float64) string {
	var b strings.Builder
	metLong, metShort := false, false
	for _, s := range doc.Scenarios {
		if s.Confirm == nil && !(IsBreakdownCondition(s.Condition) && s.Breakdown != nil) {
			continue
		}
		v := EvaluateScenarioConfirm(s, bars, sinceMs, nowMs)
		if len(v.Legs) == 0 && v.Rule == "" {
			continue
		}
		if len(v.Legs) == 0 {
			// Single-leg scenarios keep the legacy byte-identical line.
			met := "NOT MET"
			if v.Met {
				met = "MET"
				dir := strings.ToLower(strings.TrimSpace(s.Direction))
				if dir != "long" && dir != "short" {
					if strings.EqualFold(v.Side, "above") {
						dir = "long"
					} else if strings.EqualFold(v.Side, "below") {
						dir = "short"
					}
				}
				switch dir {
				case "long":
					metLong = true
				case "short":
					metShort = true
				}
				if a := staleConfirmAnnotation(s, v, nowPrice, atr5m); a != "" {
					met += " " + a
				}
			}
			b.WriteString(fmt.Sprintf("  %s confirm: %s %s %.2f — %s (%s)\n",
				s.ID, strings.ReplaceAll(v.Rule, "_", " "), v.Side, v.RefPrice, met, v.Detail))
			continue
		}
		// F2 — two-leg rendering: every leg prints, a partial NEVER prints as
		// a bare "MET".
		var legs strings.Builder
		for i, l := range v.Legs {
			lm := "NOT MET"
			if l.Met {
				lm = "MET"
			}
			legs.WriteString(fmt.Sprintf("leg %d/%d %s — %s", i+1, len(v.Legs),
				strings.ReplaceAll(l.Rule, "_", " "), lm))
			if i < len(v.Legs)-1 {
				legs.WriteString(" · ")
			}
		}
		overall := "NOT MET"
		if v.Met {
			overall = "MET"
			dir := strings.ToLower(strings.TrimSpace(s.Direction))
			if dir == "long" {
				metLong = true
			} else if dir == "short" {
				metShort = true
			}
			if a := staleConfirmAnnotation(s, v, nowPrice, atr5m); a != "" {
				overall += " " + a
			}
		}
		b.WriteString(fmt.Sprintf("  %s confirm: %s → overall %s (%s)\n", s.ID, legs.String(), overall, v.Detail))
	}
	if metLong && metShort {
		b.WriteString("CONFLICT: opposing confirms MET — structural ambiguity, default WAIT unless fresh trigger\n")
	}
	if b.Len() == 0 {
		return ""
	}
	return "Machine-computed confirmations (advisory — you remain the judge):\n" + b.String()
}
