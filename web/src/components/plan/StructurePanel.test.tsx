// S5 — StructurePanel pins: renders NOTHING when the structure block is absent
// (no placeholder rows), and one row per TF with trend + zone TF badges when
// present. Mirrors the S1 contract field names verbatim.

import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StructurePanel, structureZonesToOverlay } from './StructurePanel'
import type { StructureMapView } from '../../lib/api/plan'

const structure: StructureMapView = {
  as_of_ms: 1760000000000,
  contract: 'MNQ 12-26',
  tfs: {
    D: {
      trend: 'up',
      last_swing_high: { price: 29600, time_ms: 1759999000000 },
      last_swing_low: { price: 29200, time_ms: 1759900000000 },
      impulse_lo: 29200,
      impulse_hi: 29600,
      premium_discount: 0.25,
      zones: [{ kind: 'Supply', lo: 29550, hi: 29600, tf: 'D', fresh: 'fresh', score: 3.1 }],
      bars: 60,
    },
    '4h': { trend: 'down', premium_discount: 0.6, zones: [], bars: 120 },
    '1h': { trend: 'range', premium_discount: 0.5, zones: [{ kind: 'OB', lo: 29400, hi: 29440, tf: '1h', fresh: 'tested-1', score: 1.9 }], bars: 40 },
  },
}

describe('StructurePanel', () => {
  it('renders NOTHING when the structure block is absent', () => {
    const { container } = render(<StructurePanel structure={undefined} />)
    expect(container).toBeEmptyDOMElement()
    render(<StructurePanel structure={null} />)
    expect(screen.queryByTestId('structure-panel')).toBeNull()
  })

  it('renders the bias-only header and one row per TF with badges', () => {
    render(<StructurePanel structure={structure} />)
    expect(screen.getByTestId('structure-panel')).toBeTruthy()
    expect(screen.getByText(/STRUCTURE — bias only, not entries/)).toBeTruthy()
    expect(screen.getByTestId('structure-tf-D')).toBeTruthy()
    expect(screen.getByTestId('structure-tf-4h')).toBeTruthy()
    expect(screen.getByTestId('structure-tf-1h')).toBeTruthy()
    expect(screen.getAllByTestId('structure-zone')).toHaveLength(2)
    const badges = screen.getAllByTestId('structure-zone-tf').map((n) => n.textContent)
    expect(badges).toEqual(['D', '1h'])
    expect(screen.getByText(/pd 25%/)).toBeTruthy()
  })

  it('maps zones to chart overlay entries with kind·tf labels and raw bands', () => {
    const overlay = structureZonesToOverlay(structure)
    expect(overlay).toHaveLength(2)
    expect(overlay[0].label).toBe('Supply\u00b7D')
    expect(overlay[0].range).toEqual([29550, 29600])
    expect(overlay[1].label).toBe('OB\u00b71h')
    expect(structureZonesToOverlay(undefined)).toEqual([])
  })
})
