package kernel

import (
	"fmt"
	"sort"
	"strings"
)

// CLASS 34 (owner ruling 2026-08-31) — validator hints must name only legal
// conditions. Tonight both ASIA chains failed: the breakdown-void reject said
// "author a reject/retest play instead", the model authored condition
// "reject_retest", and parse/schema rejected it — the model complied with the
// hint and was punished for it. A hint is an instruction; instructions must be
// checkable. This registry pins every hint's condition tokens against the enum
// and the default shadow map (the table test IS the guard; the boot line
// re-runs it at startup).

// BreakdownReclaimedHint is the remediation phrase for a void breakdown (a
// close came back across the breakdown level).
const BreakdownReclaimedHint = "author a `reject` play instead (do NOT combine condition names; `reject_retest` is not a valid condition)"

// BreakdownDisplacementHint is the remediation phrase for a sub-BD_MIN_DISP_ATR
// waterfall authoring.
const BreakdownDisplacementHint = "author a normal `reject` play instead (do NOT combine condition names; `reject_retest` is not a valid condition)"

// ArmLegsSplitContract is the arm-legs condition contract fragment
// (sweep_reclaim only splits).
const ArmLegsSplitContract = "(the split entry is the sweep_reclaim contract; other conditions arm single)"

// RepairBreakdownLaw is the BREAKDOWN-CONTINUE law excerpt for repair prompts.
const RepairBreakdownLaw = "BREAKDOWN-CONTINUE LAW: the breakdown is void once a close comes back across the breakdown level — author a `reject` play instead of breakdown_continue (do NOT combine condition names; `reject_retest` is not a valid condition)."

// RepairArmSplitLaw is the ARM-SPLIT law excerpt for repair prompts.
const RepairArmSplitLaw = "ARM-SPLIT LAW: a scenario with a split contract arm needs EXACTLY 2 legs — leg 1 rests AT the sweep ref with confirm=touch at the sweep ref; leg 2 is 1m_mss (1x5m_close accepted as the leg-2 alternative). Only sweep_reclaim conditions arm split; every other condition arms a SINGLE leg."

// RepairEntryConfirmLaw is the ENTRY-LAW CONFIRM excerpt for repair prompts.
const RepairEntryConfirmLaw = "ENTRY-LAW CONFIRM LAW: breakdown_continue takes 1 confirming close + displacement >= BD_MIN_DISP_ATR x ATR5m OR stop-entry (E7); 2x5m is legal ONLY there. confirm2 mirrors confirm1 unless the law above allows it."

// ValidatorHint pairs a validator message/hint site with the condition tokens
// its text names.
type ValidatorHint struct {
	Site       string
	Text       string
	Conditions []string // every LEGAL condition name the text mentions
}

// ValidatorHints is the class-34 registry. Every hint shipped in a validator
// message or a repair excerpt MUST be listed here; the guard test asserts each
// token is in the enum and is not shadowed by default.
func ValidatorHints() []ValidatorHint {
	return []ValidatorHint{
		{Site: "breakdown_continue.go reclaimed", Text: BreakdownReclaimedHint, Conditions: []string{"reject"}},
		{Site: "breakdown_continue.go displacement", Text: BreakdownDisplacementHint, Conditions: []string{"reject"}},
		{Site: "plan_doc.go arm-legs contract", Text: ArmLegsSplitContract, Conditions: []string{"sweep_reclaim"}},
		{Site: "planner_repair.go breakdown law", Text: RepairBreakdownLaw, Conditions: []string{"reject", "breakdown_continue"}},
		{Site: "planner_repair.go arm-split law", Text: RepairArmSplitLaw, Conditions: []string{"sweep_reclaim"}},
		{Site: "planner_repair.go entry-law confirm", Text: RepairEntryConfirmLaw, Conditions: []string{"breakdown_continue"}},
	}
}

// ValidateValidatorHints is the class-34 guard: every condition token a hint
// names must exist in the enum and must not be shadowed by default. The table
// test is the hard build gate; the boot line re-runs this at startup.
func ValidateValidatorHints() error {
	known := make(map[string]bool, len(KnownConditions()))
	for _, c := range KnownConditions() {
		known[c] = true
	}
	resolved := ResolvedConditionStatuses(nil, nil, "") // defaults — the static contract
	for _, h := range ValidatorHints() {
		for _, c := range h.Conditions {
			if !known[c] {
				return fmt.Errorf("validator hint %q names unknown condition %q (enum: %v)", h.Site, c, KnownConditions())
			}
			if resolved[c] == ConditionShadow {
				return fmt.Errorf("validator hint %q names shadowed condition %q", h.Site, c)
			}
		}
	}
	return nil
}

// ResolvedLiveConditions returns the sorted list of conditions that resolve
// LIVE for the given base/session maps + env — the vocabulary appended to the
// planner reject block (class 34, fix 5).
func ResolvedLiveConditions(base, session map[string]string, env string) []string {
	resolved := ResolvedConditionStatuses(base, session, env)
	out := make([]string, 0, len(resolved))
	for c, st := range resolved {
		if st != ConditionShadow {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}

// LiveConditionsLine renders the "Valid conditions: [...]" reject-block suffix.
func LiveConditionsLine(live []string) string {
	if len(live) == 0 {
		return ""
	}
	return fmt.Sprintf("\nValid conditions: [%s] (use exactly ONE token from this list; do NOT combine condition names).", strings.Join(live, ", "))
}
