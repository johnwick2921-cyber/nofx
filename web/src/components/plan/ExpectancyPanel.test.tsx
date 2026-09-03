// Wave 1D — the expectancy table's honesty properties, pinned.
//
// Every assertion here is about a way the panel could LIE: showing a verdict on
// a sample too small to carry one, rendering an unmeasured statistic as 0,
// letting a counterfactual row sit in the realized table, hiding the sample, or
// timestamping stale data with the moment the page was opened.

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'

let payload: unknown = null
vi.mock('../../lib/api/plan', () => ({
  planApi: { getExpectancy: () => Promise.resolve(payload) },
}))

const renderPanel = async () => {
  const { ExpectancyPanel } = await import('./ExpectancyPanel')
  return render(<ExpectancyPanel />)
}

const row = (over: Record<string, unknown> = {}) => ({
  key: { condition: 'reject', session: 'NY' },
  n: 20,
  wins: 12,
  losses: 8,
  flats: 0,
  sum_pnl_corrected: 40,
  mean: 2,
  sd: 10.052493799,
  win_rate: 0.6,
  wilson_lo: 0.3865779423,
  wilson_hi: 0.7811960326,
  mean_lo: -2.4,
  mean_hi: 6.4,
  t_stat: 0.8897565,
  avg_realized_r: null,
  avg_planned_rr: null,
  median_mae: null,
  median_mfe: null,
  stop_hit_share: null,
  target_hit_share: null,
  excluded_unresolved: 3,
  row_ids: [584, 586, 591],
  descriptive_only: true,
  status: 'NOT ENOUGH DATA',
  ...over,
})

const base = (over: Record<string, unknown> = {}) => ({
  by: 'condition',
  rows: [row()],
  counterfactual_e8: [],
  excluded: {
    unresolved_pnl: 3,
    unresolvable: 1,
    test_seam: 2,
    no_condition: 0,
    crypto_era: 0,
  },
  min_n: 30,
  as_of_ms: 1788000000000,
  as_of_utc: '2026-09-01T18:41:14Z',
  built_at_ms: 1788200000000,
  era_0b_start: '2026-09-02T12:49:06Z',
  promotion_rule:
    'n >= min_n AND mean > 0 AND the 95% interval on the mean excludes 0',
  ...over,
})

describe('ExpectancyPanel', () => {
  beforeEach(() => {
    payload = null
  })

  it('shows n before any statistic and marks a sub-floor row DESCRIPTIVE ONLY', async () => {
    payload = base()
    await renderPanel()
    await waitFor(() =>
      expect(screen.getByTestId('expectancy-panel')).toBeTruthy()
    )

    const cells = screen.getByTestId('expectancy-row-0').textContent ?? ''
    // n is the first number the eye meets — the floor is a property of n.
    expect(cells.indexOf('20')).toBeLessThan(cells.indexOf('2.00'))
    expect(screen.getByTestId('expectancy-status-0').textContent).toMatch(
      /DESCRIPTIVE ONLY/i
    )
    // and NO verdict is rendered from a sample below the floor
    expect(cells).not.toMatch(/PASSES|FAILS/)
  })

  it('renders a verdict only at or above the floor that the payload carries', async () => {
    payload = base({
      rows: [
        row({
          n: 30,
          mean: 16.6666667,
          mean_lo: 9.8037171,
          descriptive_only: false,
          status: 'PASSES',
        }),
      ],
    })
    await renderPanel()
    await waitFor(() =>
      expect(screen.getByTestId('expectancy-panel')).toBeTruthy()
    )
    expect(screen.getByTestId('expectancy-status-0').textContent).toMatch(
      /PASSES/
    )
  })

  it('renders an unmeasured statistic blank, never as zero', async () => {
    payload = base()
    await renderPanel()
    await waitFor(() =>
      expect(screen.getByTestId('expectancy-panel')).toBeTruthy()
    )
    const mae = screen.getByTestId('expectancy-mae-0').textContent ?? ''
    expect(mae.trim()).toBe('—')
    expect(mae).not.toMatch(/0/)
  })

  it('exposes the row ids so any figure can be recomputed', async () => {
    payload = base()
    await renderPanel()
    await waitFor(() =>
      expect(screen.getByTestId('expectancy-panel')).toBeTruthy()
    )
    fireEvent.click(screen.getByTestId('expectancy-ids-toggle-0'))
    expect(screen.getByTestId('expectancy-ids-0').textContent).toMatch(
      /584.*586.*591/
    )
  })

  it('dates the table by the last closed trade, not by page load', async () => {
    payload = base()
    await renderPanel()
    await waitFor(() =>
      expect(screen.getByTestId('expectancy-panel')).toBeTruthy()
    )
    const asOf = screen.getByTestId('expectancy-as-of').textContent ?? ''
    expect(asOf).toMatch(/2026-09-01/)
    expect(asOf).not.toMatch(/2026-09-03/)
  })

  it('states what was excluded and why, beside the table', async () => {
    payload = base()
    await renderPanel()
    await waitFor(() =>
      expect(screen.getByTestId('expectancy-panel')).toBeTruthy()
    )
    const ex = screen.getByTestId('expectancy-excluded').textContent ?? ''
    expect(ex).toMatch(/3/)
    expect(ex).toMatch(/unresolved/i)
    expect(ex).toMatch(/test.seam/i)
  })

  it('keeps counterfactual rows in their own table, flagged', async () => {
    payload = base({
      counterfactual_e8: [
        {
          key: { condition: 'sweep_reclaim', session: 'NY' },
          rule: 'touch',
          n: 9,
          wins: 4,
          losses: 5,
          sum_net_pnl: -12,
          mean: -1.3333,
          counterfactual: true,
          short_suspect: true,
          note: 'counterfactual (E8) · SHORT ROWS SUSPECT (E8 sign bug)',
        },
      ],
    })
    await renderPanel()
    await waitFor(() =>
      expect(screen.getByTestId('expectancy-panel')).toBeTruthy()
    )
    const cf = screen.getByTestId('expectancy-counterfactual')
    expect(cf.textContent).toMatch(/counterfactual/i)
    expect(cf.textContent).toMatch(/SUSPECT/i)
    // the realized table must not have absorbed it
    expect(screen.getByTestId('expectancy-row-0').textContent).not.toMatch(
      /sweep_reclaim/
    )
  })

  it('says so when there is nothing to show, instead of rendering an empty table', async () => {
    payload = base({ rows: [] })
    await renderPanel()
    await waitFor(() =>
      expect(screen.getByTestId('expectancy-empty')).toBeTruthy()
    )
  })
})
