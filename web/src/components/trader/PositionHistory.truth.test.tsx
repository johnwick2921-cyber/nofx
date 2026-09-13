import { render, screen, waitFor, act } from '@testing-library/react'
import { beforeEach, expect, it, vi } from 'vitest'
import { PositionHistory } from './PositionHistory'
const history = vi.hoisted(() => vi.fn())
vi.mock('../../lib/api', () => ({ api: { getPositionHistory: history } }))
vi.mock('../../contexts/LanguageContext', () => ({
  useLanguage: () => ({ language: 'en' }),
}))
vi.mock('../../lib/autoRefresh', () => ({
  useAutoRefresh: vi.fn(),
  REFRESH_HISTORY_MS: 30000,
}))
const payload = (id: number, symbol: string) => ({
  positions: [
    {
      id,
      symbol,
      side: 'LONG',
      entry_price: 100,
      exit_price: 101,
      quantity: 1,
      exit_time: new Date().toISOString(),
      close_reason: 'sync',
      pnl_corrected: null,
      realized_pnl: 9876.54,
    },
  ],
  stats: null,
  symbol_stats: [],
  direction_stats: [],
})
beforeEach(() => history.mockReset())
it('renders unresolved correction and exclusion count without raw PNL', async () => {
  history.mockResolvedValue(payload(9001, 'MNQ'))
  render(<PositionHistory traderId="fixture" />)
  await screen.findByText('P&L unresolved')
  expect(screen.getByTestId('unresolved-pnl-count')).toHaveTextContent(
    '1 unresolved excluded'
  )
  expect(document.body.textContent).not.toContain('9876.54')
  expect(document.body.textContent).not.toContain('9,876.54')
})
it('ignores previous trader history arriving after new trader history', async () => {
  let resolve!: (x: unknown) => void
  history
    .mockImplementationOnce(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )
    .mockResolvedValueOnce(payload(9002, 'NEW'))
  const { rerender } = render(<PositionHistory traderId="old" />)
  rerender(<PositionHistory traderId="new" />)
  await waitFor(() =>
    expect(screen.getAllByText('NEW').length).toBeGreaterThan(0)
  )
  await act(async () => resolve(payload(9001, 'OLD')))
  expect(screen.queryByText('OLD')).toBeNull()
})

it('shows the server aggregate unresolved count separately from loaded-row counts', async () => {
  history.mockResolvedValue({
    ...payload(9001, 'MNQ'),
    stats: {
      total_trades: 2,
      resolved_trades: 2,
      unresolved_excluded: 7,
      total_pnl: 10,
    },
  })
  render(<PositionHistory traderId="fixture" />)
  expect(await screen.findByTestId('aggregate-pnl-counts')).toHaveTextContent(
    '2 resolved · 7 unresolved excluded'
  )
  expect(screen.getByTestId('unresolved-pnl-count')).toHaveTextContent(
    '1 unresolved excluded'
  )
})
