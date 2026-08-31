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

  it('has exactly 42 knob cards (Section 7 census = live-page control count; W7 +6 weekly knobs, min-side card removed 2026-08-31, +2 planner-speed 2026-08-31)', () => {
    expect(allKnobs).toHaveLength(42)
  })

  it('every knob card fills all ten mandatory fields', () => {
    for (const k of allKnobs) {
      for (const f of fields) {
        expect(k[f], `knob "${k.label}" field ${f}`).toBeTruthy()
      }
    }
  })

  it('every recommended field states its reason', () => {
    for (const k of allKnobs) {
      expect(k.recommended.length).toBeGreaterThan(10)
    }
  })
})

describe('GuidePage', () => {
  beforeEach(() => {
    // never-resolving default: only the drift tests stub a revision
    vi.stubGlobal('fetch', vi.fn().mockReturnValue(new Promise(() => {})))
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders all 13 sections with deep-link ids', () => {
    render(<GuidePage />)
    expect(screen.getByTestId('guide-page')).toBeTruthy()
    expect(GUIDE_SECTIONS).toHaveLength(13)
    for (const s of GUIDE_SECTIONS) {
      const el = document.getElementById(s.id)
      expect(el, `section id #${s.id}`).toBeTruthy()
      expect(el!.textContent).toContain(String(s.num))
    }
  })

  it('renders the live-component examples (guide-example testids)', () => {
    render(<GuidePage />)
    // mock card renders inside its Example wrapper
    expect(
      screen.getAllByTestId('guide-example').length
    ).toBeGreaterThanOrEqual(1)
    expect(screen.getByTestId('mock-plan-card')).toBeTruthy()
    // real chips inside the mock
    expect(
      document.querySelector('[data-testid^="confirm-chip-"]')
    ).toBeTruthy()
    expect(document.querySelector('[data-testid^="fvg-chip-"]')).toBeTruthy()
  })

  it('search filters hits and links to sections', () => {
    render(<GuidePage />)
    const input = screen.getByTestId('guide-search')
    fireEvent.change(input, { target: { value: 'thin side' } })
    // the planCard callout + glossary should both surface hits
    const links = screen
      .getAllByRole('link')
      .filter((a) => a.getAttribute('href')?.startsWith('#'))
    expect(links.length).toBeGreaterThan(0)
    expect(links.some((a) => a.getAttribute('href') === '#plan-card')).toBe(
      true
    )
  })

  it('search with no matches says so', () => {
    render(<GuidePage />)
    fireEvent.change(screen.getByTestId('guide-search'), {
      target: { value: 'zzzzqqqq' },
    })
    expect(screen.getByText(/no matches/i)).toBeTruthy()
  })

  it('shows the rev-drift banner when the bot revision differs', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        json: () =>
          Promise.resolve({ status: 'ok', revision: '0ldcafe12345678' }),
      })
    )
    render(<GuidePage />)
    expect(await screen.findByTestId('guide-rev-drift')).toBeTruthy()
  })

  it('hides the drift banner when the bot revision matches', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        json: () =>
          Promise.resolve({ status: 'ok', revision: GUIDE_BUILT_REV }),
      })
    )
    render(<GuidePage />)
    // flush the fetch microtask inside act so the revision update is wrapped
    await act(async () => {})
    expect(screen.queryByTestId('guide-rev-drift')).toBeNull()
  })

  it('renders knob cards with all ten fields', () => {
    render(<GuidePage />)
    const knobs = screen.getAllByTestId('guide-knob')
    expect(knobs.length).toBeGreaterThanOrEqual(20)
    const first = within(knobs[0])
    for (const label of [
      'where',
      'range',
      'default',
      'per-session',
      'engine',
      'touch it when',
    ]) {
      expect(first.getByText(new RegExp(label, 'i'))).toBeTruthy()
    }
  })
})
