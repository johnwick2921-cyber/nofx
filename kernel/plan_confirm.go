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
	Rule     string  `json:"rule"`
	RefPrice float64 `json:"ref_price"`
	Side     string  `json:"side"`
	Met      bool    `json:"met"`
	Detail   string  `json:"detail"` // e.g. "last 15m close 29641.00"
}

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
		if s.Confirm == nil {
			continue
		}
		v := EvaluateConfirm(*s.Confirm, bars, sinceMs, nowMs)
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
	}
	if metLong && metShort {
		b.WriteString("CONFLICT: opposing confirms MET — structural ambiguity, default WAIT unless fresh trigger\n")
	}
	if b.Len() == 0 {
		return ""
	}
	return "Machine-computed confirmations (advisory — you remain the judge):\n" + b.String()
}
