// W7 (weekly-bias wave) — WEEKLY chip: 4 render states (bull · bear · none ·
// invalidated strikethrough) + the card-level mount. Advisory view only.
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { WeeklyChip } from './WeeklyChip'
import { SessionPlanCard } from './SessionPlanCard'
import type { PlanToday, PlanWeekly } from '../../lib/api/plan'

const bull: PlanWeekly = {
  bias: 'bull',
  conviction: 'high',
  draw_name: 'PWH',
  draw_px: 30500.25,
  invalidation_px: 30300,
  invalidation_basis: '1h close beyond 30300.00',
  invalidated_at: '',
  narrative: 'accepted above the prior high\nholding above weekly open',
  weekly_levels: [{ name: 'PWH', px: 30500.25 }],
  thin_history: false,
}
const bear: PlanWeekly = { ...bull, bias: 'bear', conviction: 'low' }
const invalidated: PlanWeekly = {
  ...bull,
  invalidated_at: '2026-08-28 10:15 CT',
}

describe('WeeklyChip', () => {
  it('bull state — arrow + conviction + draw px', () => {
    render(<WeeklyChip weekly={bull} />)
    const chip = screen.getByTestId('weekly-chip')
    expect(chip.textContent).toContain('▲')
    expect(chip.textContent).toContain('bull')
    expect(chip.textContent).toContain('high')
    expect(chip.textContent).toContain('30500.25')
    expect(chip.getAttribute('title')).toContain('WEEKLY: bull/high')
  })

  it('bear state — down arrow', () => {
    render(<WeeklyChip weekly={bear} />)
    const chip = screen.getByTestId('weekly-chip')
    expect(chip.textContent).toContain('▼')
    expect(chip.textContent).toContain('bear')
  })

  it('none state — grey WEEKLY none', () => {
    render(<WeeklyChip weekly={null} />)
    const chip = screen.getByTestId('weekly-chip')
    expect(chip.textContent).toContain('WEEKLY none')
    expect(chip.getAttribute('title')).toContain('WEEKLY: none')
  })

  it('invalidated state — strikethrough + neutral tooltip', () => {
    render(<WeeklyChip weekly={invalidated} />)
    const chip = screen.getByTestId('weekly-chip')
    expect(chip.style.textDecoration).toBe('line-through')
    expect(chip.getAttribute('title')).toContain('WEEKLY: neutral (invalidated')
  })
})

describe('SessionPlanCard — WEEKLY chip mounts on the card', () => {
  const plan = (weekly: PlanWeekly | null): PlanToday => ({
    found: true,
    trade_date: '2026-08-31',
    session: 'NY',
    night: false,
    mode: 'advisory',
    version: 1,
    lifecycle: 'active',
    model_id: 'deepseek-v4-pro',
    replans_left: 2,
    is_active: true,
    weekly,
    doc: {
      reasoning: 'n/a',
      bias: { direction: 'long', conviction: 'medium', flip_condition: 'n/a' },
      levels: [],
      scenarios: [],
      no_trade: [],
      death_condition: 'n/a',
      day_type: 'range',
    },
    level_facts: [],
  })

  it('renders the chip when a weekly doc exists', () => {
    render(
      <SessionPlanCard
        plan={plan(bull)}
        traderId="t1"
        symbol="MNQ"
        exchange="ninjatrader"
        language="en"
      />
    )
    expect(screen.getByTestId('weekly-chip').textContent).toContain('bull')
  })

  it('renders the grey none chip when no weekly doc exists', () => {
    render(
      <SessionPlanCard
        plan={plan(null)}
        traderId="t1"
        symbol="MNQ"
        exchange="ninjatrader"
        language="en"
      />
    )
    expect(screen.getByTestId('weekly-chip').textContent).toContain(
      'WEEKLY none'
    )
  })
})
