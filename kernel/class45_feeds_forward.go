package kernel

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"nofx/market"
)

// ── CLASS 45 (2026-09-02) — THE PROMPT FEEDS FORWARD WHAT THE VALIDATOR KNOWS ─
//
// LONDON 2026-09-02 (planner_rejected_prompts 92/93/94), measured:
//   · 01:32:54 attempt 1 — S1 breakdown_continue at 29021.25 → "a close came
//     back across 29021.25 — the breakdown is void"
//   · 01:35:04 attempt 2 — the repair obeyed the hint, killed by fade_requires_touch
//   · 01:37:44 attempt 3 — a FULL re-author put breakdown_continue back at the
//     SAME 29021.25 → void again → fail-closed, session lost
//
// The standing order sat at 70% depth; the correction was the last 239 chars
// (59 of 6,602 tokens, under 1%, at 99% depth) and described attempt 2's fade
// defect, not attempt 1's void; the facts block named 29021.25 six times and
// never once said it was dead. The model did what it was told.
//
// Everything here CLOSES that gap by carrying the validator's own knowledge
// forward into the prompt. The void verdict is not re-derived: it is the
// validator's own predicate (BreakdownContinueState), reached through a
// level-oriented entry point, so prompt and validator cannot hold two opinions.

// VoidBreakdownLevel is one level the write-site validator WOULD refuse a
// waterfall play at, because a close came back across it.
type VoidBreakdownLevel struct {
	Price         float64
	Short         bool   // true = breakdown (short side); false = breakup (long side)
	ReclaimedAtCT string // when the reclaiming close printed ("" = unknown)
}

// BreakdownLevelReclaimed runs THE VALIDATOR'S OWN predicate for one level by
// building the minimal scenario BreakdownContinueState expects. This is the
// single source of the void verdict — never a second implementation.
//
// The reclaim TIMESTAMP is derived separately (the state struct does not carry
// one); the VERDICT always comes from the predicate.
func BreakdownLevelReclaimed(level float64, short bool, bars []market.Kline, sinceMs, nowMs int64) (bool, string) {
	if level <= 0 || len(bars) == 0 {
		return false, ""
	}
	cond := "breakup_continue"
	if short {
		cond = "breakdown_continue"
	}
	sc := PlanScenario{
		ID: "probe", Condition: cond, Direction: map[bool]string{true: "short", false: "long"}[short],
		Breakdown: &PlanBreakdownContinue{Level: level, EntryMode: "pullback"},
	}
	st := BreakdownContinueState(sc, bars, sinceMs, nowMs)
	if !st.Reclaimed {
		return false, ""
	}
	return true, reclaimStampCT(level, short, bars, sinceMs, nowMs)
}

// reclaimStampCT finds the first close back ACROSS the level after the level was
// broken — the label only. Voidness is the predicate's call, never this.
func reclaimStampCT(level float64, short bool, bars []market.Kline, sinceMs, nowMs int64) string {
	broken := false
	for _, b := range bars {
		if b.OpenTime < sinceMs || b.CloseTime > nowMs {
			continue
		}
		beyond := (short && b.Close < level) || (!short && b.Close > level)
		if beyond {
			broken = true
			continue
		}
		if broken {
			return FormatCT(time.UnixMilli(b.OpenTime))
		}
	}
	return ""
}

// ComputeVoidBreakdownLevels runs the predicate over every ranked level, both
// sides, and returns those the validator would refuse. Deterministic order
// (price ascending) so the prompt text is stable across cycles.
// VOID PARITY (2026-09-02): the scope is resolved, not chosen by the caller —
// the same VoidScope the write-site validator reads.
func ComputeVoidBreakdownLevels(levels []ScoredLevel, scope VoidScope, nowMs int64) []VoidBreakdownLevel {
	bars, sinceMs := scope.Bars, scope.SinceMs
	var out []VoidBreakdownLevel
	seen := map[string]bool{}
	for _, l := range levels {
		if l.Price <= 0 {
			continue
		}
		for _, short := range []bool{true, false} {
			void, at := BreakdownLevelReclaimed(l.Price, short, bars, sinceMs, nowMs)
			if !void {
				continue
			}
			key := fmt.Sprintf("%.2f:%v", l.Price, short)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, VoidBreakdownLevel{Price: l.Price, Short: short, ReclaimedAtCT: at})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Price != out[j].Price {
			return out[i].Price < out[j].Price
		}
		return out[i].Short && !out[j].Short
	})
	return out
}

// RenderVoidBreakdownLevels is the facts-block line (E2). Empty when none —
// silence means "nothing is void", never an empty header.
func RenderVoidBreakdownLevels(v []VoidBreakdownLevel) string {
	if len(v) == 0 {
		return ""
	}
	var parts []string
	for _, x := range v {
		side := "breakup"
		if x.Short {
			side = "breakdown"
		}
		when := ""
		if x.ReclaimedAtCT != "" {
			when = fmt.Sprintf(" (reclaimed %s)", x.ReclaimedAtCT)
		}
		parts = append(parts, fmt.Sprintf("%.2f %s%s", x.Price, side, when))
	}
	return "## VOID breakdown levels (a close came back across since the break — the write-site validator REFUSES a waterfall play at these)\n" +
		strings.Join(parts, " · ") +
		"\n- do NOT author breakdown_continue or breakup_continue at these prices. Any other condition is legal there.\n\n"
}

// RenderStopFloorLine is the facts-block line (E3): the floor the composer will
// enforce, from the SAME resolver it uses. On 2026-09-02 all 13 armed stops were
// widened because the planner was never told this number, and every widening
// silently cut the planned R:R toward the 2.0 arm gate.
func RenderStopFloorLine(atr5m, mult float64) string {
	if atr5m <= 0 || mult <= 0 {
		return ""
	}
	return fmt.Sprintf("## Minimum stop distance this cycle\n%.1f pts (%.1f×ATR5m %.2f, resolved). Stops tighter than this are WIDENED by the executor before the R:R gate sees them — author stops AND targets consistent with it, or your R:R will not survive the widening.\n\n",
		mult*atr5m, mult, atr5m)
}

// PromptFeedsForwardBootLine (E6) — every field READ, never a literal.
func PromptFeedsForwardBootLine(voidLevels int, atr5m, mult float64) string {
	// HONESTY (checklist class 49): at boot there are no bars, so the void list
	// is NOT "zero levels are void" — it is "not computed yet". A negative count
	// means uncomputed and prints as n/a; 0 means measured-and-empty. The floor
	// MULTIPLIER is known at boot even when ATR is not, so it is always stated.
	floor := fmt.Sprintf("%.1f×ATR5m (n/a — no ATR yet)", mult)
	if atr5m > 0 && mult > 0 {
		floor = fmt.Sprintf("%.1f×ATR5m=%.1fpts", mult, mult*atr5m)
	}
	levels := fmt.Sprintf("void-levels=%d", voidLevels)
	if voidLevels < 0 {
		levels = "void-levels=n/a (computed per read)"
	}
	return fmt.Sprintf("prompt feeds forward: %s · stop-floor=%s · reject-block=top+tail (class 45)", levels, floor)
}

// planDocGapDownMessage is the gap-down refusal text, exported here so the pin
// can assert the MESSAGE matches the RULE (which tests direction only).
func planDocGapDownMessage() string { return gapDownDirectionMessage }
