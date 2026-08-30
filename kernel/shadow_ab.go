package kernel

import (
	"math"
	"strings"

	"nofx/market"
)

// SHADOW A/B COUNTERFACTUALS (E8, entry-mechanics 2026-08-30) — Sep-9's
// courtroom table. Per armed/authored setup, compute 4 counterfactual fills
// (touch · 1x5m_close · 2x5m_close · 1m_mss) by 1m replay: fill px, MFE, MAE,
// target/stop outcome, time-to-fill. PURE — computes only; the logger writes.
// ZERO effect on real paths (nothing here feeds a gate, a prompt, or a wire).

// ShadowABRow is one counterfactual entry's replay verdict.
type ShadowABRow struct {
	Rule         string  // touch | 1x5m_close | 2x5m_close | 1m_mss
	FillPx       float64 // the counterfactual fill (0 = never filled)
	MFE          float64 // max favorable excursion in pts
	MAE          float64 // max adverse excursion in pts
	Outcome      string  // target | stop | open
	TimeToFillMs int64   // fill bar open − sinceMs
}

// ShadowABForScenario computes the counterfactual rows for ONE armed scenario.
// Needs the arm bracket (entry ref from Confirm.RefPrice; stop/target from the
// arm) — non-armed scenarios have no bracket to replay and return nil.
func ShadowABForScenario(sc PlanScenario, bars []market.Kline, sinceMs, nowMs int64) []ShadowABRow {
	if sc.Arm == nil || !sc.Arm.Enabled || sc.Confirm == nil || sc.Confirm.RefPrice <= 0 {
		return nil
	}
	ref := sc.Confirm.RefPrice
	stop, target := sc.Arm.Stop, sc.Arm.Target
	if stop <= 0 || target <= 0 {
		return nil
	}
	above := strings.EqualFold(sc.Confirm.Side, "above")
	long := strings.EqualFold(strings.TrimSpace(sc.Direction), "long")
	if !long {
		// mirror signs so MFE is always favorable-positive in the replay
		// (short: price DOWN is favorable).
		stop, target = -stop, -target
		ref = -ref
		above = !above
	}
	w := BarsSince(bars, sinceMs)
	if len(w) == 0 {
		return nil
	}
	type fillAt struct {
		rule   string
		barIdx int
		px     float64
	}
	var fills []fillAt
	// touch — the first bar whose range straddles the level fills AT the level.
	for i := range w {
		b := w[i]
		if b.CloseTime >= nowMs {
			continue
		}
		if b.Low <= ref && b.High >= ref {
			fills = append(fills, fillAt{"touch", i, ref})
			break
		}
	}
	// close rules — the qualifying close beyond ref in side; 1x/2x count
	// CONSECUTIVE 5m closes. barIdx maps back to the first 1m bar of the
	// qualifying bucket so the replay starts at the true fill instant.
	closeFill := func(rule string, need int) fillAt {
		five := AggregateBars(w, 5*60_000)
		run := 0
		for i := range five {
			b := five[i]
			if b.CloseTime >= nowMs {
				continue
			}
			beyond := (above && b.Close > ref) || (!above && b.Close < ref)
			if beyond {
				run++
				if run >= need {
					return fillAt{rule, i * 5, b.Close}
				}
			} else {
				run = 0
			}
		}
		return fillAt{}
	}
	if f := closeFill("1x5m_close", 1); f.rule != "" {
		fills = append(fills, f)
	}
	if f := closeFill("2x5m_close", 2); f.rule != "" {
		fills = append(fills, f)
	}
	// 1m_mss — the break bar close (index into w so the replay starts there).
	if m := EvaluateMSS(w, sc.Confirm.Side, nowMs); m.Met {
		idx := -1
		for i := range w {
			if w[i].OpenTime == m.BreakTimeMs {
				idx = i
				break
			}
		}
		fills = append(fills, fillAt{"1m_mss", idx, m.BreakClose})
	}
	if len(fills) == 0 {
		return nil
	}
	out := make([]ShadowABRow, 0, len(fills))
	for _, f := range fills {
		row := ShadowABRow{Rule: f.rule, FillPx: f.px}
		if row.FillPx <= 0 {
			continue
		}
		// Replay from the fill bar forward: MFE/MAE vs stop/target. Intrabar
		// stop+target ambiguity resolves AGAINST the trade (stop first, R9).
		start := 0
		if f.barIdx >= 0 {
			start = f.barIdx
			row.TimeToFillMs = w[f.barIdx].OpenTime - sinceMs
		}
		mfe, mae, outcome := 0.0, 0.0, "open"
		for _, b := range w[start:] {
			if b.CloseTime >= nowMs {
				continue
			}
			hi, lo := b.High, b.Low
			if !long {
				hi, lo = -lo, -hi
			}
			// adverse first (R9): a bar that spans both stop and target counts
			// as a stop-out.
			if lo <= stop {
				mae = math.Max(mae, row.FillPx-stop)
				outcome = "stop"
				break
			}
			if hi >= target {
				mfe = math.Max(mfe, target-row.FillPx)
				outcome = "target"
				break
			}
			mfe = math.Max(mfe, hi-row.FillPx)
			mae = math.Max(mae, row.FillPx-lo)
		}
		row.MFE, row.MAE, row.Outcome = mfe, mae, outcome
		out = append(out, row)
	}
	return out
}
