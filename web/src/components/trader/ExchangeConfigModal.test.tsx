// The exchange picker offers exactly the venues the backend still builds, and
// a stored row whose type is no longer offered (a removed broker, or no type
// at all) renders a NAMED notice instead of an empty modal. Both are asserted
// on the production component (ExchangeConfigModal), not a rebuilt list.
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { Exchange } from '../../types'

// The environment check fetches the server's crypto config on mount; the
// picker's option set does not depend on it.
vi.mock('../common/WebCryptoEnvironmentCheck', () => ({
  WebCryptoEnvironmentCheck: () => null,
}))

import { ExchangeConfigModal } from './ExchangeConfigModal'

const REMAINING = [
  'aster',
  'bitget',
  'bybit',
  'gate',
  'hyperliquid',
  'indodax',
  'kucoin',
  'lighter',
  'ninjatrader',
  'okx',
]

const renderModal = (
  allExchanges: Exchange[],
  editingExchangeId: string | null
) =>
  render(
    <ExchangeConfigModal
      allExchanges={allExchanges}
      editingExchangeId={editingExchangeId}
      onSave={async () => {}}
      onDelete={() => {}}
      onClose={() => {}}
      language="en"
    />
  )

const storedRow = (exchange_type: string): Exchange => ({
  id: 'row-1',
  exchange_type,
  account_name: 'Legacy',
  name: 'Legacy',
  type: 'cex',
  enabled: true,
  states: {},
})

describe('ExchangeConfigModal — the offered exchange set', () => {
  it('offers exactly the remaining exchange types, nothing more', () => {
    renderModal([], null)
    const offered = screen
      .getAllByTestId(/^exchange-option-/)
      .map((el) =>
        (el.getAttribute('data-testid') || '').replace('exchange-option-', '')
      )
      .sort()
    expect(offered).toEqual(REMAINING)
  })
})

describe('ExchangeConfigModal — a stored row of a type no longer offered', () => {
  it('names the stored type instead of rendering an empty body', () => {
    renderModal([storedRow('retired-venue')], 'row-1')
    const notice = screen.getByTestId('exchange-unsupported-notice')
    expect(notice.textContent).toContain('"retired-venue"')
    expect(notice.textContent).toContain('no longer supported')
    // No credential form is offered for a type the backend refuses to load.
    expect(screen.queryByTestId('exchange-submit')).toBeNull()
  })

  it('prints n/a — never a made-up venue — when the row has no type', () => {
    renderModal([storedRow('')], 'row-1')
    expect(
      screen.getByTestId('exchange-unsupported-notice').textContent
    ).toContain('"n/a"')
  })

  it('shows no notice for a type that is still offered', () => {
    renderModal([storedRow('bybit')], 'row-1')
    expect(screen.queryByTestId('exchange-unsupported-notice')).toBeNull()
    expect(screen.getByTestId('exchange-submit')).toBeTruthy()
  })
})
