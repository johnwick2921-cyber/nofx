package kernel

import "strings"

// Wave 2 armed orders (2026-08-27) — pure classification + price derivation
// for arming scenarios as resting orders. The arming DECISION stays with the
// AI (plan.scenarios[].arm.authorized); this file only says WHAT an armed
// scenario means in machine terms.

// ArmableCondition reports whether a condition supports a price-deterministic
// resting entry. acceptance and raw sweep_reclaim need close-confirmation
// first and stay on the AI path. breakout_retest was EXCLUDED by the
// grand-audit response wave (F4, 2026-08-28): its replay expectancy is negative
// at every R-floor, so it stays a normal AI play and is never armed.
func ArmableCondition(condition string) bool {
	switch strings.ToLower(strings.TrimSpace(condition)) {
	case "fvg_entry", "reject":
		return true
	}
	return false
}

// ArmedEntryPx derives the resting LIMIT price for an armed scenario.
//   - fvg_entry: CE when entry_mode == "ce", else the gap EDGE (long → fvg_hi,
//     short → fvg_lo — the first touch of the gap in the trade's direction).
//   - breakout_retest: the retest level = the scenario's snapped anchor.
//   - reject: the anchor, offset one tick INTO the trade's favor
//     (long → anchor − tick, short → anchor + tick).
// Returns 0 when underivable (no fvg object / no anchor / bad direction).
func ArmedEntryPx(sc PlanScenario, anchor float64, tick float64) float64 {
	switch strings.ToLower(strings.TrimSpace(sc.Condition)) {
	case "fvg_entry":
		if sc.Fvg == nil || sc.Fvg.Lo <= 0 || sc.Fvg.Hi <= 0 {
			return 0
		}
		if strings.EqualFold(strings.TrimSpace(sc.Fvg.EntryMode), "ce") {
			return sc.Fvg.CE
		}
		if strings.EqualFold(strings.TrimSpace(sc.Direction), "short") {
			return sc.Fvg.Lo
		}
		return sc.Fvg.Hi
	case "breakout_retest":
		return anchor
	case "reject":
		if tick <= 0 {
			tick = 0.25
		}
		if strings.EqualFold(strings.TrimSpace(sc.Direction), "short") {
			return anchor + tick
		}
		return anchor - tick
	}
	return 0
}

// ScenarioQualityRank ranks A+ > A > B > C (unknown → 0).
func ScenarioQualityRank(q string) int {
	switch strings.ToUpper(strings.TrimSpace(q)) {
	case "A+":
		return 4
	case "A":
		return 3
	case "B":
		return 2
	case "C":
		return 1
	}
	return 0
}
