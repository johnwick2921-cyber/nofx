// Open interest and funding rate have no data source any more (their only
// feed, a crypto-perpetual one, was removed). On a crypto strategy both
// toggles stay — a saved strategy keeps its setting — but each card says there
// is no source and the AI prompt shows n/a. On CME futures both stay hidden.
// Asserted on the production component with only the capability fetch stubbed.
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { IndicatorConfig } from '../../types'

vi.mock('../../lib/api', () => ({
  api: {
    getSupportedTimeframes: vi.fn().mockResolvedValue({
      timeframes: ['1m', '5m', '15m', '1h'],
      soft_warn_above: 6,
    }),
  },
}))

import { IndicatorEditor } from './IndicatorEditor'

const config = {
  klines: {
    primary_timeframe: '5m',
    primary_count: 100,
    enable_multi_timeframe: false,
    selected_timeframes: ['5m'],
  },
  enable_raw_klines: true,
  enable_ema: false,
  enable_macd: false,
  enable_rsi: false,
  enable_atr: false,
  enable_boll: false,
  enable_volume: true,
  enable_oi: true,
  enable_funding_rate: false,
} as unknown as IndicatorConfig

const renderEditor = (isFutures: boolean) =>
  render(
    <IndicatorEditor
      config={config}
      onChange={() => {}}
      language="en"
      isFutures={isFutures}
    />
  )

describe('IndicatorEditor — open interest / funding have no data source', () => {
  afterEach(() => cleanup())

  it('crypto: both toggles stay and each says there is no data source', () => {
    renderEditor(false)
    for (const key of ['enable_oi', 'enable_funding_rate']) {
      expect(
        screen.getByTestId(`indicator-no-source-${key}`)
      ).toHaveTextContent('No data source — the AI prompt shows n/a')
    }
    expect(screen.queryByTestId('indicator-no-source-enable_volume')).toBeNull()
  })

  it('futures: both toggles stay hidden', () => {
    renderEditor(true)
    expect(screen.queryByTestId('indicator-no-source-enable_oi')).toBeNull()
    expect(
      screen.queryByTestId('indicator-no-source-enable_funding_rate')
    ).toBeNull()
  })
})
