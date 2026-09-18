// W-ARM-STATE-UI (2026-09-18) — executor verdict pins at the production call
// sites: executorLinesFor (the derivation the card renders) and the badge.

import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { PlanArmView, StructuralGeometryView } from '../../lib/api/plan'
import { executorLinesFor, ExecutorVerdict } from './ExecutorVerdict'
import { StrategyTradingBadge } from '../strategy/StrategyTradingBadge'

const geometryRow = (
  o: Partial<StructuralGeometryView>
): StructuralGeometryView => ({
  scenario: 'S1',
  leg: 1,
  entry: 30000,
  stop_source: 'zone_edge',
  reason: '',
  detail: '',
  quantity: 0,
  ...o,
})

describe('executorLinesFor', () => {
  it('renders nothing when no record exists for the version', () => {
    // armedMapFor always serves a per-scenario view; UNKNOWN = no ledger row.
    const arm: PlanArmView = { state: 'UNKNOWN', reason: 'no arm recorded' }
    expect(executorLinesFor('S1', arm, [])).toEqual([])
    expect(executorLinesFor('S1', undefined, null)).toEqual([])
    const { container } = render(
      <ExecutorVerdict scenario="S1" arm={arm} geometry={[]} />
    )
    expect(
      container.querySelector('[data-testid="executor-verdict-S1"]')
    ).toBeNull()
  })

  it('shows a geometry refusal with reason, detail and time', () => {
    const geometry = [
      geometryRow({
        scenario: 'S1',
        reason: 'no_provenance',
        detail: 'invalid_direction_entry_or_contract_spec',
        quantity: 0,
        time_ms: 1760000000000,
      }),
    ]
    const lines = executorLinesFor('S1', { state: 'UNKNOWN' }, geometry)
    expect(lines).toEqual([
      expect.objectContaining({
        state: 'refused',
        label:
          'refused: no_provenance (invalid_direction_entry_or_contract_spec)',
      }),
    ])
    render(
      <ExecutorVerdict
        scenario="S1"
        arm={{ state: 'UNKNOWN' }}
        geometry={geometry}
      />
    )
    const el = screen.getByTestId('executor-verdict-S1')
    expect(el.textContent).toContain('refused: no_provenance')
    expect(el.firstElementChild?.getAttribute('title')).toContain(
      'invalid_direction_entry_or_contract_spec'
    )
  })

  it('shows the armed_orders verdict — filled #id', () => {
    const arm: PlanArmView = {
      state: 'filled',
      legs: [{ state: 'filled', row_id: 12, leg_index: 0, placement_seq: 1 }],
    }
    const lines = executorLinesFor('S1', arm, [])
    expect(lines).toEqual([
      expect.objectContaining({ state: 'filled', label: 'filled #12' }),
    ])
  })

  it('maps armed, cancelled and not-attempted states', () => {
    expect(
      executorLinesFor(
        'S1',
        {
          state: 'armed',
          legs: [{ state: 'armed', row_id: 7, leg_index: 0, placement_seq: 1 }],
        },
        []
      )
    ).toEqual([expect.objectContaining({ label: 'armed #7' })])
    expect(
      executorLinesFor(
        'S1',
        {
          state: 'cancelled',
          reason: 'boot_sweep …',
          legs: [
            {
              state: 'cancelled',
              reason: 'boot_sweep …',
              leg_index: 0,
              placement_seq: 1,
            },
          ],
        },
        []
      )
    ).toEqual([expect.objectContaining({ state: 'cancelled' })])
    expect(
      executorLinesFor('S1', { state: 'UNKNOWN' }, [
        geometryRow({
          scenario: 'S1',
          reason: 'pending_gates',
          quantity: 0,
          time_ms: 1760000000001,
        }),
      ])
    ).toEqual([
      expect.objectContaining({
        state: 'not_attempted',
        label: 'not attempted',
      }),
    ])
  })
})

describe('StrategyTradingBadge', () => {
  const tr = (k: string, p?: Record<string, string>) =>
    p ? `${k}:${p.names}` : k

  it('shows the binding truth and display truth separately', () => {
    render(<StrategyTradingBadge isActive traderNames={['主账户']} tr={tr} />)
    expect(screen.getByTestId('strategy-trading-badge').textContent).toBe(
      'tradingBoundTo:主账户'
    )
    expect(screen.getByTestId('strategy-display-badge').textContent).toBe(
      'displayActive'
    )
  })

  it('shows honest unbound + inactive texts', () => {
    render(<StrategyTradingBadge isActive={false} traderNames={[]} tr={tr} />)
    expect(screen.getByTestId('strategy-trading-badge').textContent).toBe(
      'tradingNoBound'
    )
    expect(screen.getByTestId('strategy-display-badge').textContent).toBe(
      'displayInactive'
    )
  })
})
