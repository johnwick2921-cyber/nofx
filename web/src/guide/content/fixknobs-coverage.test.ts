// FIX-KNOBS P2-4 (DS-105, 2026-09-26) — guide coverage: the four knobs the
// 2026-09-26 audit found ABSENT from the guide must each have an entry
// (planner_contract, fade_or_wide_k, last_entry_offset_min, eod_flat_offset_min).
// A presence pin: the guide is prose, but its coverage of a shipped knob is
// checkable, and a knob documented nowhere is a knob that surprises.
import { describe, expect, it } from 'vitest'
import { settings } from './settings'

describe('guide covers the FIX-KNOBS P2-4 knobs', () => {
  const required = [
    'Planner contract (A3)',
    'Fade-or-wide k (W2 LABEL)',
    'Last-entry offset (minutes before session end)',
    'EOD-flat offset (minutes before session end)',
  ]

  it('has one entry per required knob', () => {
    const labels = settings.blocks
      .filter((b) => b.kind === 'knobs')
      .flatMap((b) => b.knobs.map((k) => k.label))
    for (const want of required) {
      expect(labels, `guide entry missing: ${want}`).toContain(want)
    }
  })
})
