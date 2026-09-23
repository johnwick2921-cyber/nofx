// The chat sidebar's market ticker reads GET /api/agent/tickers, whose rows
// come from NinjaTrader bars: {symbol, price, change_1h_pct, change_4h_pct,
// source: "nt8", unavailable?}. A null value renders "n/a" — never 0 — and a
// named `unavailable` reason is shown as given. Asserted on the production
// component with only the HTTP client stubbed.
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { MarketTicker } from './MarketTicker'

const mocks = vi.hoisted(() => ({ request: vi.fn() }))

vi.mock('../../lib/httpClient', () => ({
  httpClient: { request: mocks.request },
}))

const renderTicker = () => render(<MarketTicker language="en" />)

describe('MarketTicker (NinjaTrader rows)', () => {
  beforeEach(() => {
    mocks.request.mockReset()
  })

  afterEach(() => {
    cleanup()
  })

  it('asks the agent tickers endpoint for MNQ only', async () => {
    mocks.request.mockResolvedValue({ success: true, data: [] })
    renderTicker()
    await screen.findByTestId('ticker-row-MNQ')
    expect(mocks.request).toHaveBeenCalledWith(
      '/api/agent/tickers?symbols=MNQ',
      { silent: true }
    )
  })

  it('renders price and the 1h/4h changes with the NinjaTrader source', async () => {
    mocks.request.mockResolvedValue({
      success: true,
      data: [
        {
          symbol: 'MNQ',
          price: 30323.75,
          change_1h_pct: 0.12,
          change_4h_pct: -0.3,
          source: 'nt8',
        },
      ],
    })
    renderTicker()
    expect(await screen.findByTestId('ticker-price-MNQ')).toHaveTextContent(
      '30,323.75'
    )
    expect(screen.getByTestId('ticker-1h-MNQ')).toHaveTextContent('+0.12%')
    expect(screen.getByTestId('ticker-4h-MNQ')).toHaveTextContent('-0.30%')
    expect(screen.getByTestId('ticker-source-MNQ')).toHaveTextContent(
      'NinjaTrader'
    )
    expect(screen.queryByTestId('ticker-unavailable-MNQ')).toBeNull()
  })

  it('renders null values as n/a (never 0) and shows the unavailable reason', async () => {
    mocks.request.mockResolvedValue({
      success: true,
      data: [
        {
          symbol: 'MNQ',
          price: null,
          change_1h_pct: null,
          change_4h_pct: null,
          source: 'nt8',
          unavailable: 'no NinjaTrader bars for MNQ yet',
        },
      ],
    })
    renderTicker()
    expect(await screen.findByTestId('ticker-price-MNQ')).toHaveTextContent(
      /^n\/a$/
    )
    expect(screen.getByTestId('ticker-1h-MNQ')).toHaveTextContent(/^n\/a$/)
    expect(screen.getByTestId('ticker-4h-MNQ')).toHaveTextContent(/^n\/a$/)
    expect(screen.getByTestId('ticker-unavailable-MNQ')).toHaveTextContent(
      'no NinjaTrader bars for MNQ yet'
    )
    expect(screen.getByTestId('ticker-row-MNQ').textContent).not.toMatch(
      /\b0\.00\b/
    )
  })

  it('keeps a real zero change distinct from an absent one', async () => {
    mocks.request.mockResolvedValue({
      success: true,
      data: [
        {
          symbol: 'MNQ',
          price: 30000,
          change_1h_pct: 0,
          change_4h_pct: null,
          source: 'nt8',
        },
      ],
    })
    renderTicker()
    expect(await screen.findByTestId('ticker-1h-MNQ')).toHaveTextContent(
      '0.00%'
    )
    expect(screen.getByTestId('ticker-4h-MNQ')).toHaveTextContent(/^n\/a$/)
  })

  it('names a refused request instead of going blank', async () => {
    mocks.request.mockResolvedValue({
      success: false,
      message: 'tickers unavailable',
    })
    renderTicker()
    expect(
      await screen.findByTestId('ticker-unavailable-MNQ')
    ).toHaveTextContent('tickers unavailable')
    expect(screen.getByTestId('ticker-price-MNQ')).toHaveTextContent(/^n\/a$/)
  })

  it('names a symbol the server returned no row for', async () => {
    mocks.request.mockResolvedValue({ success: true, data: [] })
    renderTicker()
    expect(
      await screen.findByTestId('ticker-unavailable-MNQ')
    ).toHaveTextContent('the server returned no data for this symbol')
    expect(screen.getByTestId('ticker-price-MNQ')).toHaveTextContent(/^n\/a$/)
    expect(screen.getByTestId('ticker-source-MNQ')).toHaveTextContent('n/a')
  })
})
