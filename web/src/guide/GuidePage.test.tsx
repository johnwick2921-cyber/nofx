// GuidePage content + render tests. FE-only, no backend.
import { act, fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { GuidePage, GUIDE_SECTIONS } from './GuidePage'
import { GUIDE_BUILT_REV } from './types'

// ── per-knob required-fields lint: EVERY KnobSpec field must be non-empty ──
describe('knob spec completeness', () => {
  const fields = [
    'label',
    'where',
    'what',
    'trader',
    'consumer',
    'range',
    'systemDefault',
    'recommended',
    'whenToTouch',
    'perSession',
  ] as const

  const allKnobs = GUIDE_SECTIONS.flatMap((s) =>
    s.blocks.flatMap((b) => (b.kind === 'knobs' ? b.knobs : []))
  )

  it('has exactly 48 knob cards (Section 7 census = live-page control count; W7 +6 weekly knobs, min-side card removed 2026-08-31, +2 planner-speed 2026-08-31, +1 planner stream total deadline class 37 2026-09-01, +1 planner stream retry tries+backoff class 41 2026-09-02, +1 fast-mode shadow A/B root-fix 2026-09-02, +1 stop floor + structure anchor 0B 2026-09-02), +1 wake cadence class 47 2026-09-02, −2 weekly knobs retired class 50 (WEEKLY_INVALIDATION_TF_DEFAULT, WEEKLY_COUNTER_MODE — refs-only weekly has no invalidation and no counter), +2 one-setup knobs (switch + min grade) dispatch 102 2026-09-11, +1 structure table (S1, day_plan.structure_map, default OFF) 2026-09-16, +1 by-TF freshness knob (S2, day_plan.levels_fresh_by_tf, default OFF) 2026-09-16, +2 HTF seating knobs (S3: day_plan.htf_seats, day_plan.htf_score_multiplier, both default unchanged) 2026-09-16, +1 flip re-read knob (W-FLIP-REREAD: day_plan.flip_reread, default OFF) 2026-09-17, +1 red-news hard-block currencies (W-T1-CURRENCIES: day_plan.t1_currencies, default USD) 2026-09-18, −9 W-KNOB-PRUNE 2026-09-18 (structure table, HTF score multiplier, max scenarios, acceptance window, evening digest, re-align cap, 1h anchor seat, HTF freshness by TF, min wake interval removed/folded off the page; the 5-toggle wake card became the 1-switch wake card), +1 write-time feasibility knob (W-WRITE-TIME-FEASIBILITY: day_plan.write_time_feasibility, default ON) 2026-09-18, +1 geometry reference levels (W-GEOMETRY-REFUSAL: day_plan.geometry_reference_levels, default ON) 2026-09-18, +1 death re-read (W-DEATH-REREAD: day_plan.death_reread, default ON) 2026-09-18, +1 picture HTF (W-PICTURE-HTF: day_plan.picture_htf, default OFF) 2026-09-20, +4 entry-policy knobs (W-EXEC-TRUTH W3: day_plan.entry_policy_default market_in_zone, zone_max_pts 10, zone_rest_max_min 30, min_hold_min 3) 2026-09-23, +1 planner fresh tape (PLANNER A6: day_plan.planner_fresh_tape, default ON) 2026-09-25 +1 zone placement reach (PLANNER B1: day_plan.zone_place_within_pts, default 25) 2026-09-25', () => {
