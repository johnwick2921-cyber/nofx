package kernel

import (
	"fmt"
	"strings"

	"nofx/market"
)

// ENTRY-MECHANICS E2 (2026-08-30) — the per-condition ENTRY LAW.
//
// ONE table, ONE chokepoint, enum-keyed: condition → allowed confirm rules +
// required entry style. ValidateEntryLaw runs inside ValidatePlanDocWithCaps
// (the schema chokepoint), so a plan whose confirm structure violates its
// condition's law is REJECTED AT WRITE with a NAMED message — never a silent
// rewrite. Legacy stored docs that predate the law still load via the base
// plan path (the overlay re-validation armor), so nothing breaks at read.
//
// Owner ruling (2026-08-30): no 15m confirms ever (E1); default 1x5m_close;
// 2x5m_close ONLY for breakdown/breakup continuations; fades enter on touch.

type conditionEntryLaw struct {
	Allowed  map[string]bool // confirm rules legal for Confirm / Confirm2
	Style    string          // the required entry style (human, cited in rejections)
	FadeTouch bool           // reject/fvg_entry: close-confirms on a fade are illegal
}

var entryLaw = map[string]conditionEntryLaw{
	"reject": {
		Allowed: map[string]bool{"touch": true},
		Style:   "touch-entry at the level (limit), stop behind structure by ≥2 ticks",
		FadeTouch: true,
	},
	"fvg_entry": {
		Allowed: map[string]bool{"touch": true},
		Style:   "touch-entry inside the FVG (edge..CE band, checked vs the FRESH FVG list)",
		FadeTouch: true,
	},
	"sweep_reclaim": {
		Allowed: map[string]bool{"touch": true, "1x5m_close": true, "1m_mss": true},
		Style:   "split contract (E4): leg-1 touch at the sweep ref, leg-2 1m_mss (1x5m_close accepted as the leg-2 alternative)",
	},
	"reclaim": {
		Allowed: map[string]bool{"1x5m_close": true, "1m_mss": true},
		Style:   "reclaim-close discipline — 1x5m or 1m_mss, never 2x",
	},
	"breakout_retest": {
		Allowed: map[string]bool{"touch": true, "1x5m_close": true},
		Style:   "touch at the retest limit + stop-entry fallback (E7); 1x5m legal for the break leg",
	},
	"acceptance": {
		Allowed: map[string]bool{"time_hold": true, "1x5m_close": true},
		Style:   "time_hold rule (E6) with 1x5m as the legal fallback",
	},
	"hold": {
		Allowed: map[string]bool{"time_hold": true, "1x5m_close": true},
		Style:   "time_hold rule (E6) with 1x5m as the legal fallback",
	},
	"breakdown_continue": {
		Allowed: map[string]bool{"1x5m_close": true, "2x5m_close": true},
		Style:   "1 confirming close + displacement ≥ BD_MIN_DISP_ATR×ATR5m OR stop-entry (E7); 2x5m legal ONLY here",
	},
	"breakup_continue": {
		Allowed: map[string]bool{"1x5m_close": true, "2x5m_close": true},
		Style:   "1 confirming close + displacement ≥ BD_MIN_DISP_ATR×ATR5m OR stop-entry (E7); 2x5m legal ONLY here",
	},
}

// EntryLawFor returns the law for a condition (ok=false when unknown).
func EntryLawFor(condition string) (conditionEntryLaw, bool) {
	law, ok := entryLaw[strings.ToLower(strings.TrimSpace(condition))]
	return law, ok
}

// 2x5m is RESERVED for the waterfall class — every other condition that
// authors it gets the named rejection "2x5m_reserved".
func twoX5mReserved(condition string) bool {
	return !IsBreakdownCondition(condition)
}

// ValidateEntryLaw enforces the per-condition confirm law over a whole doc.
func ValidateEntryLaw(d *PlanDoc) error {
	if d == nil {
		return nil
	}
	for i, s := range d.Scenarios {
		law, ok := EntryLawFor(s.Condition)
		if !ok {
			continue // unknown conditions are caught by scenarioConds upstream
		}
		for _, c := range []struct {
			label string
			c     *PlanConfirm
		}{{"confirm", s.Confirm}, {"confirm2", s.Confirm2}} {
			if c.c == nil {
				continue
			}
			// Fades are touch-only — the fade law wins over the 2x5m rule so a
			// close-confirm on reject/fvg_entry is ALWAYS fade_requires_touch.
			if law.FadeTouch && c.c.Rule != "touch" {
				return fmt.Errorf("scenario[%d].%s.rule %q — fade_requires_touch (a %s fade enters on the touch at the level, never on a close-confirm: %s)",
					i, c.label, c.c.Rule, s.Condition, law.Style)
			}
			if c.c.Rule == "2x5m_close" && twoX5mReserved(s.Condition) {
				return fmt.Errorf("scenario[%d].%s.rule 2x5m_close — 2x5m_reserved (double-close confirms are legal ONLY on breakdown_continue|breakup_continue; %s uses: %s)",
					i, c.label, s.Condition, law.Style)
			}
			if !law.Allowed[c.c.Rule] {
				return fmt.Errorf("scenario[%d].%s.rule %q not allowed for %s — entry law: %s",
					i, c.label, c.c.Rule, s.Condition, law.Style)
			}
		}
		// sweep_reclaim split-contract (E4): leg 1 must be the sweep touch; the
		// leg-2 alternative set is {1m_mss, 1x5m_close}. A single-confirm
		// sweep_reclaim may carry the reclaim-close discipline on its own
		// (1x5m_close | 1m_mss — the owner's standing discipline).
		if strings.EqualFold(strings.TrimSpace(s.Condition), "sweep_reclaim") && s.Confirm2 != nil {
			if s.Confirm != nil && s.Confirm.Rule != "touch" {
				return fmt.Errorf("scenario[%d].confirm.rule %q — sweep_leg1_requires_touch (the sweep leg of a two-leg sweep_reclaim enters on the touch at the sweep ref)", i, s.Confirm.Rule)
			}
			if s.Confirm2.Rule != "1m_mss" && s.Confirm2.Rule != "1x5m_close" {
				return fmt.Errorf("scenario[%d].confirm2.rule %q — sweep_leg2_requires_mss_or_1x5m (leg 2 chains on 1m_mss, 1x5m_close accepted as the alternative)", i, s.Confirm2.Rule)
			}
		}
		// fade stop law: an ARMED fade (reject) must carry a structure stop
		// BEYOND the level by ≥2 ticks (the validator checks what it can see —
		// the arm's bracket stop; AI-path prose stops stay AI-judged).
		if law.FadeTouch && s.Arm != nil && s.Arm.Enabled && s.Confirm != nil {
			tick := market.FuturesTickSize("MNQ")
			if tick <= 0 {
				tick = 0.25
			}
			dir := strings.ToLower(strings.TrimSpace(s.Direction))
			if dir == "long" && s.Arm.Stop > s.Confirm.RefPrice-2*tick {
				return fmt.Errorf("scenario[%d].arm.stop %.2f must sit ≥2 ticks BEYOND the fade level %.2f for a long reject (stop %.2f < level−2tick %.2f) — structure stop required",
					i, s.Arm.Stop, s.Confirm.RefPrice, s.Arm.Stop, s.Confirm.RefPrice-2*tick)
			}
			if dir == "short" && s.Arm.Stop < s.Confirm.RefPrice+2*tick {
				return fmt.Errorf("scenario[%d].arm.stop %.2f must sit ≥2 ticks BEYOND the fade level %.2f for a short reject (stop %.2f > level+2tick %.2f) — structure stop required",
					i, s.Arm.Stop, s.Confirm.RefPrice, s.Arm.Stop, s.Confirm.RefPrice+2*tick)
			}
		}
	}
	return nil
}
