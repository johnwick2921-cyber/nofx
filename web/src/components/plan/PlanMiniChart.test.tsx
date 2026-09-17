// W-CHART-ZONE-WALL — the component: a 40-zone structure block draws ≤ 12+6
// overlay entries by default, the "Show all zones" control is OFF by default,
// reports "6 of 30" (duplicates merged), flips to all on a click, and the
// choice is remembered per viewer in localStorage.
//
// jsdom has no canvas, so the chart degrades to its placeholder; the zone
// control renders on the placeholder too, which is what this pins.
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { PlanMiniChart, factsToOverlay } from './PlanMiniChart'
import { SHOW_ALL_ZONES_KEY, selectChartOverlay } from './chartOverlaySelect'
import type { OverlayLevel } from './LevelOverlayPrimitive'
import type { PlanLevelFact } from '../../lib/api/plan'
import { httpClient } from '../../lib/httpClient'

vi.spyOn(httpClient, 'request').mockResolvedValue({
  success: true,
  data: [],
} as never)

const facts: PlanLevelFact[] = Array.from({ length: 12 }, (_, i) => ({
  price: 29000 + i * 25,
  label: `L${i}`,
  grade: 'A',
  instruction: 'watch',
  distance: 0,
  sweep: false,
  closes_beyond: 0,
  accept_have: 0,
  accept_need: 0,
  still_valid: true,
}))

function fortyZones(): OverlayLevel[] {
  const distinct: OverlayLevel[] = Array.from({ length: 30 }, (_, i) => {
    const lo = 28000 + i * 100
    return {
      price: lo + 40,
      label: i % 2 ? 'SUPPLY·1h' : 'DEMAND·4h',
      grade: 'A',
      range: [lo, lo + 40] as [number, number],
    }
  })
  return [...distinct, ...distinct.slice(0, 10).map((z) => ({ ...z }))]
}

describe('PlanMiniChart · zone wall', () => {
  beforeEach(() => {
    window.localStorage.clear()
  })

  it('draws ≤ 12 seated + 6 zones from a 40-zone block; control OFF by default', () => {
    render(
      <PlanMiniChart
        symbol="MNQ"
        exchange="ninjatrader"
        facts={facts}
        language="en"
        structureZones={fortyZones()}
      />
    )
    const box = screen.getByTestId('chart-show-all-zones') as HTMLInputElement
    expect(box.checked).toBe(false)
    expect(screen.getByTestId('chart-zone-count').textContent).toContain(
      '6 of 30'
    )
    // the SAME selection the component hands the overlay, on the same inputs
    const sel = selectChartOverlay(
      factsToOverlay(facts),
      fortyZones(),
      undefined,
      {
        showAll: false,
      }
    )
    expect(sel.levels.length).toBeLessThanOrEqual(12 + 6)
    expect(sel.levels.filter((l) => l.range)).toHaveLength(6)
  })

  it('"Show all zones" flips to every distinct zone and is remembered', () => {
    render(
      <PlanMiniChart
        symbol="MNQ"
        exchange="ninjatrader"
        facts={facts}
        language="en"
        structureZones={fortyZones()}
      />
    )
    fireEvent.click(screen.getByTestId('chart-show-all-zones'))
    expect(screen.getByTestId('chart-zone-count').textContent).toContain(
      '30 of 30'
    )
    expect(window.localStorage.getItem(SHOW_ALL_ZONES_KEY)).toBe('1')
    // a fresh mount reads the preference back
    render(
      <PlanMiniChart
        symbol="MNQ"
        exchange="ninjatrader"
        facts={facts}
        language="en"
        structureZones={fortyZones()}
      />
    )
    const boxes = screen.getAllByTestId(
      'chart-show-all-zones'
    ) as HTMLInputElement[]
    expect(boxes[boxes.length - 1].checked).toBe(true)
  })

  it('no zones → no control (nothing to cap)', () => {
    render(
      <PlanMiniChart
        symbol="MNQ"
        exchange="ninjatrader"
        facts={facts}
        language="en"
        structureZones={[]}
      />
    )
    expect(screen.queryByTestId('chart-zone-control')).toBeNull()
  })
})
