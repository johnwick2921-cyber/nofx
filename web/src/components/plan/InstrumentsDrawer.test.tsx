// The instruments drawer — three descriptive instruments, collapsed by default.
//
// Each of these pins a way the drawer could mislead: opening loud by default,
// showing a number without saying where it came from, reviving the legacy
// mae/mfe columns (default-0, class 40 shape) as if they were measurements, or
// printing a retired instrument's rate without the word RETIRED next to it.

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'

let tradesPayload: unknown = null
let statsPayload: unknown = null
vi.mock('../../lib/api/plan', () => ({
  planApi: {
    getPlanTrades: () => Promise.resolve(tradesPayload),
    getPlanStats: () => Promise.resolve(statsPayload),
  },
}))

const renderDrawer = async (rows: unknown[] = []) => {
  const { InstrumentsDrawer } = await import('./InstrumentsDrawer')
  return render(
    <InstrumentsDrawer traderId="t1" expectancyRows={rows as never} />
  )
}

const baseTrades = (over: Record<string, unknown> = {}) => ({
  trades: [
    // legacy mae/mfe present and NON-ZERO: the drawer must still refuse them
    { mae: 12.5, mfe: 30, adherence_grade: 'A' },
    { mae: 0, mfe: 0, adherence_grade: 'D' },
  ],
  summary: { counts: { A: 12, B: 4, D: 3 }, total: 19, gpa: 3.21 },
  ...over,
})

const baseStats = (over: Record<string, unknown> = {}) => ({
  weekly: null,
  progress: [
    { level_type: 'PDH', n: 140, target_n: 1565, reactions: 61, warming: true },
    { level_type: 'ONL', n: 88, target_n: 1565, reactions: 40, warming: true },
  ],
  target_n: 1565,
  alpha: 0.00625,
  ...over,
})

describe('InstrumentsDrawer', () => {
  beforeEach(() => {
    tradesPayload = baseTrades()
    statsPayload = baseStats()
  })

  it('is COLLAPSED on load and shows only its label', async () => {
    await renderDrawer()
    await waitFor(() =>
      expect(screen.getByTestId('instruments-drawer')).toBeTruthy()
    )
    expect(screen.getByTestId('instruments-drawer-toggle').textContent).toMatch(
      /Instruments · discipline · MAE\/MFE · level gate \(descriptive\)/
    )
    // none of the three sections are in the document until expanded
    expect(screen.queryByTestId('instrument-discipline')).toBeNull()
    expect(screen.queryByTestId('instrument-maemfe')).toBeNull()
    expect(screen.queryByTestId('instrument-randomgate')).toBeNull()
  })

  it('expands to exactly the three sections, each naming its source and n', async () => {
    await renderDrawer()
    await waitFor(() =>
      expect(screen.getByTestId('instruments-drawer')).toBeTruthy()
    )
    fireEvent.click(screen.getByTestId('instruments-drawer-toggle'))

    const disc = screen.getByTestId('instrument-discipline').textContent ?? ''
    expect(disc).toMatch(/GPA/)
    expect(disc).toMatch(/3\.21/)
    expect(disc).toMatch(/n=19/)
    expect(disc).toMatch(/source:/i)

    const mm = screen.getByTestId('instrument-maemfe').textContent ?? ''
    expect(mm).toMatch(/source:/i)

    const rg = screen.getByTestId('instrument-randomgate').textContent ?? ''
    expect(rg).toMatch(/source:/i)
  })

  // THE ONE THAT MATTERS MOST: the legacy columns are default-0 and were being
  // averaged over whichever rows happened to be non-zero. With no excursion
  // rows the answer is "not measured", never a number derived from them.
  it('says there are no excursion rows rather than averaging the legacy columns', async () => {
    await renderDrawer([
      {
        key: { condition: 'reject' },
        n: 31,
        median_mae: null,
        median_mfe: null,
      },
    ])
    await waitFor(() =>
      expect(screen.getByTestId('instruments-drawer')).toBeTruthy()
    )
    fireEvent.click(screen.getByTestId('instruments-drawer-toggle'))

    const mm = screen.getByTestId('instrument-maemfe').textContent ?? ''
    expect(mm).toMatch(/no excursion rows yet/i)
    // the legacy values from the trades payload must appear NOWHERE
    expect(mm).not.toMatch(/12\.5/)
    expect(mm).not.toMatch(/30/)
    expect(mm).toMatch(/trade_excursions/)
  })

  it('renders MAE/MFE from trade_excursions once rows exist, with n', async () => {
    await renderDrawer([
      {
        key: { condition: 'reject' },
        n: 31,
        median_mae: 8.25,
        median_mfe: 19.5,
      },
      {
        key: { condition: 'reclaim' },
        n: 5,
        median_mae: null,
        median_mfe: null,
      },
    ])
    await waitFor(() =>
      expect(screen.getByTestId('instruments-drawer')).toBeTruthy()
    )
    fireEvent.click(screen.getByTestId('instruments-drawer-toggle'))
    const mm = screen.getByTestId('instrument-maemfe').textContent ?? ''
    expect(mm).toMatch(/8\.2|8\.3/)
    expect(mm).toMatch(/19\.5/)
    expect(mm).toMatch(/n=31/)
    // the condition with no excursion data must not be silently folded in
    expect(mm).not.toMatch(/n=36/)
  })

  // 1B has not booted: touch_outcomes is empty and nothing reads it. Until it
  // does, the level-gate chips are the BIASED legacy instrument and may not be
  // shown as a live rate without saying so.
  it('labels the legacy level-gate chips RETIRED and never shows a bare rate', async () => {
    await renderDrawer()
    await waitFor(() =>
      expect(screen.getByTestId('instruments-drawer')).toBeTruthy()
    )
    fireEvent.click(screen.getByTestId('instruments-drawer-toggle'))
    const rg = screen.getByTestId('instrument-randomgate')
    const txt = rg.textContent ?? ''
    expect(txt).toMatch(/legacy — retired, pending D1′/)
    expect(txt).toMatch(/level_stats/)
    expect(txt).toMatch(/PDH/)
    expect(txt).toMatch(/140/)

    // A rate must never render outside an element that also carries RETIRED.
    const rateNodes = Array.from(rg.querySelectorAll('[data-rate]'))
    for (const node of rateNodes) {
      const scope = node.closest('[data-retired="true"]')
      expect(
        scope,
        'a rate rendered outside a RETIRED-labelled scope'
      ).toBeTruthy()
    }
  })

  it('shows the discipline seam caveat honestly while the seam wave is not in the build', async () => {
    await renderDrawer()
    await waitFor(() =>
      expect(screen.getByTestId('instruments-drawer')).toBeTruthy()
    )
    fireEvent.click(screen.getByTestId('instruments-drawer-toggle'))
    const disc = screen.getByTestId('instrument-discipline').textContent ?? ''
    // It must not CLAIM seam exclusion it did not perform.
    expect(disc).toMatch(/seam/i)
    expect(disc).not.toMatch(/seam-excluded:\s*yes/i)
  })

  // Ported from the retired DisciplinePanel: the frozen weekly verdict was a
  // real surface and must not be lost in the fold.
  it('renders the frozen weekly verdict, still inside the RETIRED scope', async () => {
    statsPayload = baseStats({
      weekly: {
        iso_week: '2026-W36',
        verdicts: [
          { level_type: 'PDH', status: 'NO-EDGE' },
          { level_type: 'ONL', status: 'BEATS-RANDOM' },
        ],
      },
    })
    await renderDrawer()
    await waitFor(() =>
      expect(screen.getByTestId('instruments-drawer')).toBeTruthy()
    )
    fireEvent.click(screen.getByTestId('instruments-drawer-toggle'))
    const rg = screen.getByTestId('instrument-randomgate')
    const txt = rg.textContent ?? ''
    expect(txt).toMatch(/2026-W36/)
    expect(txt).toMatch(/PDH NO-EDGE/)
    expect(txt).toMatch(/ONL BEATS-RANDOM/)
    // a verdict is a claim about an edge — it stays inside the retired scope
    expect(rg.getAttribute('data-retired')).toBe('true')
  })

  it('says first eval Sunday when no weekly verdict is frozen yet', async () => {
    await renderDrawer()
    await waitFor(() =>
      expect(screen.getByTestId('instruments-drawer')).toBeTruthy()
    )
    fireEvent.click(screen.getByTestId('instruments-drawer-toggle'))
    expect(screen.getByTestId('instrument-randomgate').textContent).toMatch(
      /first eval Sunday/i
    )
  })

  it('says so when an instrument has nothing to show', async () => {
    tradesPayload = baseTrades({ summary: { counts: {}, total: 0, gpa: 0 } })
    statsPayload = baseStats({ progress: [] })
    await renderDrawer()
    await waitFor(() =>
      expect(screen.getByTestId('instruments-drawer')).toBeTruthy()
    )
    fireEvent.click(screen.getByTestId('instruments-drawer-toggle'))
    expect(screen.getByTestId('instrument-discipline').textContent).toMatch(
      /no graded trades yet/i
    )
    expect(screen.getByTestId('instrument-randomgate').textContent).toMatch(
      /no level rows yet/i
    )
  })
})
