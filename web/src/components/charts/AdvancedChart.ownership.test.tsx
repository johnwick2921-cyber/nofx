import { act, render, screen, fireEvent, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { AdvancedChart } from './AdvancedChart'
const state = vi.hoisted(() => ({
  candles: vi.fn(),
  attach: vi.fn(),
  request: vi.fn(),
  remove: vi.fn(),
}))
vi.mock('../../contexts/LanguageContext', () => ({
  useLanguage: () => ({ language: 'en' }),
}))
vi.mock('../../lib/httpClient', () => ({
  httpClient: { request: state.request },
}))
vi.mock('./primitives/SessionVolumeProfile', () => ({
  SessionVolumeProfile: class {
    setData = vi.fn()
  },
}))
vi.mock('lightweight-charts', () => ({
  CandlestickSeries: 'candles',
  HistogramSeries: 'volume',
  LineSeries: 'line',
  createSeriesMarkers: () => ({ setMarkers: vi.fn() }),
  createChart: () => ({
    addSeries: (kind: string) => ({
      setData: kind === 'candles' ? state.candles : vi.fn(),
      priceScale: () => ({ applyOptions: vi.fn() }),
      attachPrimitive: state.attach,
      detachPrimitive: vi.fn(),
      createPriceLine: vi.fn(),
      removePriceLine: vi.fn(),
    }),
    applyOptions: vi.fn(),
    remove: state.remove,
    removeSeries: vi.fn(),
    subscribeCrosshairMove: vi.fn(),
    timeScale: () => ({ fitContent: vi.fn() }),
  }),
}))
const bars = (price: number) => ({
  success: true,
  data: [
    {
      openTime: 1700000000000,
      open: price,
      high: price + 1,
      low: price - 1,
      close: price,
      volume: 1,
    },
  ],
})
beforeEach(() => {
  state.candles.mockReset()
  state.attach.mockReset()
  state.request.mockReset()
  state.remove.mockReset()
  vi.stubGlobal(
    'ResizeObserver',
    class {
      observe() {}
      disconnect() {}
    }
  )
  vi.spyOn(console, 'log').mockImplementation(() => {})
})
afterEach(() => {
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
})
it('ignores a previous symbol response arriving after the current symbol', async () => {
  let old!: (x: unknown) => void
  state.request.mockImplementation((url: string) =>
    url.includes('symbol=OLD')
      ? new Promise((r) => {
          old = r
        })
      : Promise.resolve(bars(200))
  )
  const { rerender } = render(<AdvancedChart symbol="OLD" exchange="binance" />)
  rerender(<AdvancedChart symbol="NEW" exchange="binance" />)
  await waitFor(() =>
    expect(state.candles.mock.calls.at(-1)?.[0][0]?.close).toBe(200)
  )
  await act(async () => old(bars(100)))
  expect(state.candles.mock.calls.at(-1)?.[0][0]?.close).toBe(200)
})
it('does not paint a removed chart when its pending request completes', async () => {
  let resolve!: (x: unknown) => void
  state.request.mockImplementation(
    () =>
      new Promise((r) => {
        resolve = r
      })
  )
  const { unmount } = render(<AdvancedChart />)
  unmount()
  const count = state.candles.mock.calls.length
  await act(async () => resolve(bars(100)))
  expect(state.candles).toHaveBeenCalledTimes(count)
  expect(state.remove).toHaveBeenCalledOnce()
})
it('does not reattach SVP after it was toggled off while loading', async () => {
  let resolve!: (x: unknown) => void
  state.request.mockImplementation((url: string) =>
    url.includes('/svp?')
      ? new Promise((r) => {
          resolve = r
        })
      : Promise.resolve(bars(200))
  )
  render(<AdvancedChart symbol="MNQ" exchange="ninjatrader" />)
  await waitFor(() =>
    expect(state.candles.mock.calls.at(-1)?.[0][0]?.close).toBe(200)
  )
  fireEvent.click(screen.getByRole('button', { name: 'Indicators' }))
  fireEvent.click(screen.getByText('SVP'))
  await waitFor(() => expect(resolve).toBeDefined())
  fireEvent.click(screen.getByText('SVP'))
  await act(async () => resolve({ success: true, data: { bins: [] } }))
  expect(state.attach).not.toHaveBeenCalled()
})
