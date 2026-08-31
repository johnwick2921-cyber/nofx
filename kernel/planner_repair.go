package kernel

import (
	"strings"
)

// BuildPlannerRepairPrompt (planner-speed wave 3, 2026-08-31) composes the
// attempt-≥2 EDIT call: a compact instruction header + the rejected output
// verbatim + ALL validator errors verbatim + minimal law excerpts for the
// violated rules only. NO candle tables, NO level map, NO full playbook —
// the repair is expected to cost a fraction of a full re-author's tokens.
func BuildPlannerRepairPrompt(rejectedOutput string, errors string) string {
	var b strings.Builder
	b.WriteString("You are repairing a rejected plan. Fix ONLY the named defects. Return the COMPLETE corrected plan JSON. Change nothing else.\n\n")
	b.WriteString("## Validator errors (verbatim)\n")
	b.WriteString(errors)
	b.WriteString("\n\n## Rejected plan output (verbatim)\n")
	b.WriteString(rejectedOutput)
	b.WriteString("\n\n## Applicable law (excerpts for the violated rules only)\n")
	b.WriteString(lawExcerptsFor(errors))
	return b.String()
}

// lawExcerptsFor maps a validator error string to the minimal law excerpt for
// the violated rule. Unknown errors get a generic keep-the-structure excerpt.
func lawExcerptsFor(errors string) string {
	var out []string
	switch {
	case strings.Contains(errors, "EXACTLY 2 legs") || strings.Contains(errors, "split requires confirm=touch") || strings.Contains(errors, "arm legs on"):
		out = append(out, "ARM-SPLIT LAW: a scenario with a split contract arm needs EXACTLY 2 legs — leg 1 rests AT the sweep ref with confirm=touch at the sweep ref; leg 2 is 1m_mss (1x5m_close accepted as the leg-2 alternative). Only sweep_reclaim conditions arm split; every other condition arms a SINGLE leg.")
	case strings.Contains(errors, "breakdown is void") || strings.Contains(errors, "came back across"):
		out = append(out, "BREAKDOWN-CONTINUE LAW: the breakdown is void once a close comes back across the breakdown level — author a reject/retest play instead of breakdown_continue.")
	case strings.Contains(errors, "not allowed for"):
		out = append(out, "ENTRY-LAW CONFIRM LAW: breakdown_continue takes 1 confirming close + displacement >= BD_MIN_DISP_ATR x ATR5m OR stop-entry (E7); 2x5m is legal ONLY there. confirm2 mirrors confirm1 unless the law above allows it.")
	default:
		out = append(out, "Copy the machine table's labels and prices; collapse duplicate seats; targets must sit within the proximity band of price.")
	}
	return strings.Join(out, "\n")
}
