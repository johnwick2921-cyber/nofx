import { useEffect, useState, type ComponentProps, type ReactNode } from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  render,
  screen,
  fireEvent,
  waitFor,
  cleanup,
} from '@testing-library/react'
import useSWR, { SWRConfig } from 'swr'
import { TraderDashboardPage } from './TraderDashboardPage'
const lifecycle = vi.hoisted(() => ({ mounts: 0, unmounts: 0 }))
vi.mock('../contexts/LanguageContext', () => ({
  useLanguage: () => ({ language: 'en' }),
}))
const closePosition = vi.hoisted(() =>
  vi.fn().mockResolvedValue({ message: 'sent' })
)
vi.mock('../lib/api', () => ({ api: { closePosition } }))
vi.mock('../lib/notify', () => ({
  confirmToast: vi.fn().mockResolvedValue(true),
  notify: { success: vi.fn(), error: vi.fn() },
}))
vi.mock('../lib/api/plan', () => ({
  planApi: { getRiskErrors: async () => ({ rows: [] }) },
}))
vi.mock('../components/charts/AdvancedChart', () => ({
  AdvancedChart: ({ symbol }: { symbol: string }) => (
    <div data-testid="market-chart">{symbol}</div>
  ),
}))
vi.mock('../components/charts/EquityChart', () => ({
  EquityChart: () => <div data-testid="equity-chart" />,
}))
vi.mock('../components/plan/PlanCard', () => ({
  PlanCard: ({ symbol }: { symbol: string }) => {
    const [draft, setDraft] = useState('')
    useEffect(() => {
      lifecycle.mounts++
      return () => {
        lifecycle.unmounts++
      }
    }, [])
    return (
      <div data-testid="plan-card">
        <span>{symbol}</span>
        <input
          aria-label="Planner draft"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
        />
      </div>
    )
  },
}))
vi.mock('../components/trader/AccountSelector', () => ({
  AccountSelector: () => null,
}))
vi.mock('../components/trader/EmergencyFlatButton', () => ({
  EmergencyFlatButton: () => null,
}))
vi.mock('../components/trader/PauseButton', () => ({
  PauseButton: () => null,
  isPauseActive: () => false,
  pauseUntilCT: () => '',
}))
vi.mock('../components/trader/PositionHistory', () => ({
  PositionHistory: () => null,
}))
vi.mock('../components/trader/DecisionAudit', () => ({
  DecisionAudit: () => <div>Decision audit</div>,
}))
vi.mock('../components/common/DeepVoidBackground', () => ({
  DeepVoidBackground: ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  ),
}))
vi.mock('../components/common/PunkAvatar', () => ({
  PunkAvatar: () => null,
  getTraderAvatar: () => 1,
}))
function props(futures = true): ComponentProps<typeof TraderDashboardPage> {
  return {
    selectedTrader: {
      trader_id: 'test-123456789',
      trader_name: 'Test trader',
      ai_model: 'deepseek',
      exchange_id: 'ex',
    },
    selectedTraderId: 'test-123456789',
    exchanges: [
      {
        id: 'ex',
        exchange_type: futures ? 'ninjatrader' : 'binance',
        name: 'Test',
        account_name: 'Test SIM',
        enabled: true,
        type: 'cex',
      },
    ],
    positions: [
      {
        symbol: 'MNQ',
        side: 'long',
        entry_price: 100,
        mark_price: 101,
        quantity: 1,
        leverage: 1,
        unrealized_pnl: 2,
        unrealized_pnl_pct: 1,
        liquidation_price: 0,
        margin_used: 100,
      },
    ],
    decisions: [],
    decisionsLimit: 5,
    onDecisionsLimitChange: vi.fn(),
    onTraderSelect: vi.fn(),
    onNavigateToTraders: vi.fn(),
    lastUpdate: '',
    language: 'en',
  }
}
beforeEach(() => {
  lifecycle.mounts = 0
  lifecycle.unmounts = 0
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => ({ json: async () => ({ status: 'ok', symbols: [] }) }))
  )
})
afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})
describe('Dashboard production composition', () => {
  it('mounts exactly one equity and one market chart before the guarded Planner', async () => {
    render(<TraderDashboardPage {...props()} />)
    await screen.findByTestId('market-chart')
    expect(screen.getAllByTestId('equity-chart')).toHaveLength(1)
    const market = screen.getByTestId('market-chart'),
      equity = screen.getByTestId('equity-chart'),
      plan = screen.getByTestId('plan-card')
    expect(
      market.compareDocumentPosition(equity) & Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(
      equity.compareDocumentPosition(plan) & Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(screen.getByText(/ID: test-123/)).toBeInTheDocument()
  })
  it('preserves the same PlanCard instance and draft while Overview is CSS-hidden', async () => {
    render(<TraderDashboardPage {...props()} />)
    fireEvent.change(screen.getByLabelText('Planner draft'), {
      target: { value: 'retain this draft' },
    })
    const plan = screen.getByTestId('plan-card')
    fireEvent.click(screen.getByTestId('decisions-tab'))
    expect(screen.getByTestId('plan-card')).toBe(plan)
    expect(plan.closest('.hidden')).not.toBeNull()
    expect(lifecycle.unmounts).toBe(0)
    fireEvent.click(screen.getByTestId('overview-tab'))
    expect(screen.getByLabelText('Planner draft')).toHaveValue(
      'retain this draft'
    )
    expect(lifecycle.mounts).toBe(1)
    await waitFor(() =>
      expect(screen.getByText(/RESPONDING/)).toBeInTheDocument()
    )
  })
  it('does not render Planner for crypto and retains its crypto-only hidden columns', () => {
    render(<TraderDashboardPage {...props(false)} />)
    expect(screen.queryByTestId('plan-card')).toBeNull()
    expect(document.querySelector('#planner')).toBeNull()
    const table = screen.getByRole('table')
    expect(table.querySelectorAll('th.hidden')).toHaveLength(2)
    expect(table.querySelectorAll('td.hidden')).toHaveLength(2)
  })
  it('keeps all six futures Entry/Mark/Value header and cell surfaces available on mobile', () => {
    render(<TraderDashboardPage {...props()} />)
    const table = screen.getByRole('table')
    expect(table.querySelectorAll('th.hidden,td.hidden')).toHaveLength(0)
    expect(table.closest('.overflow-x-auto')).not.toBeNull()
  })
})

it('refreshes the active provider account-scoped cache after closing a position', async () => {
  const fetcher = vi.fn(async (key: string) => key)
  const p = { ...props(), selectedAccount: 'SimFixture' }
  function Cached() {
    useSWR(`positions-${p.selectedTraderId}-SimFixture`, fetcher)
    useSWR(`account-${p.selectedTraderId}-SimFixture`, fetcher)
    return null
  }
  render(
    <SWRConfig value={{ provider: () => new Map(), dedupingInterval: 0 }}>
      <Cached />
      <TraderDashboardPage {...p} />
    </SWRConfig>
  )
  await waitFor(() => expect(fetcher).toHaveBeenCalledTimes(2))
  fireEvent.click(screen.getByRole('button', { name: /^close$/i }))
  await waitFor(() => expect(closePosition).toHaveBeenCalled())
  await waitFor(() => expect(fetcher).toHaveBeenCalledTimes(4))
})
