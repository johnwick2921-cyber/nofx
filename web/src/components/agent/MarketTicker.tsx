import { useState } from 'react'
import { httpClient, type ApiResponse } from '../../lib/httpClient'
import { useAutoRefresh, REFRESH_TICKER_MS } from '../../lib/autoRefresh'
import { t, type Language } from '../../i18n/translations'

// GET /api/agent/tickers?symbols=… — one row per requested symbol, read from
// NinjaTrader bars for CME symbols. A value the server could not compute is
// null (rendered "n/a", never 0), and `unavailable` names why.
export interface TickerRow {
  symbol: string
  price: number | null
  change_1h_pct: number | null
  change_4h_pct: number | null
  source: string
  unavailable?: string
}

const SYMBOLS = ['MNQ']

const SYMBOL_ICONS: Record<string, string> = {
  MNQ: 'N',
  NQ: 'N',
}

const SOURCE_LABELS: Record<string, string> = {
  nt8: 'NinjaTrader',
}

// The absent-value token. Not translated: it is the same "n/a" the boot lines
// and the guide use for a value that is not known.
const NA = 'n/a'

// A field the contract types as number|null is taken only when it is a finite
// number; anything else is ABSENT (null), never coerced to 0.
const num = (v: unknown): number | null =>
  typeof v === 'number' && Number.isFinite(v) ? v : null

const parseRows = (data: unknown): Record<string, TickerRow> => {
  const map: Record<string, TickerRow> = {}
  if (!Array.isArray(data)) return map
  for (const r of data as Array<Record<string, unknown> | null>) {
    if (!r || typeof r.symbol !== 'string' || r.symbol === '') continue
    const symbol = r.symbol.toUpperCase()
    map[symbol] = {
      symbol,
      price: num(r.price),
      change_1h_pct: num(r.change_1h_pct),
      change_4h_pct: num(r.change_4h_pct),
      source: typeof r.source === 'string' ? r.source : '',
      unavailable:
        typeof r.unavailable === 'string' && r.unavailable !== ''
          ? r.unavailable
          : undefined,
    }
  }
  return map
}

const formatPrice = (price: number | null) => {
  if (price === null) return NA
  if (Math.abs(price) >= 1000)
    return price.toLocaleString('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })
  if (Math.abs(price) >= 1) return price.toFixed(2)
  return price.toFixed(4)
}

const formatPct = (pct: number | null) => {
  if (pct === null) return NA
  return `${pct > 0 ? '+' : ''}${pct.toFixed(2)}%`
}

const pctColor = (pct: number | null) =>
  pct === null || pct === 0 ? '#6c6c82' : pct > 0 ? '#00e5a0' : '#F6465D'

export function MarketTicker({ language }: { language: Language }) {
  const [rows, setRows] = useState<Record<string, TickerRow>>({})
  const [fetchError, setFetchError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchTickers = async () => {
    let res: ApiResponse<TickerRow[]>
    try {
      // Batch fetch: single API call for all symbols
      res = await httpClient.request<TickerRow[]>(
        `/api/agent/tickers?symbols=${SYMBOLS.join(',')}`,
        { silent: true }
      )
    } catch (err) {
      // Network / 404 / 5xx: keep no stale rows — every symbol renders n/a
      // with the reason. Rethrown so the shared auto-refresh backs off.
      setRows({})
      setFetchError(
        (err instanceof Error && err.message) ||
          t('agentTicker.requestFailed', language)
      )
      setLoading(false)
      throw err
    }
    setLoading(false)
    if (!res.success) {
      const reason = res.message || t('agentTicker.requestFailed', language)
      setRows({})
      setFetchError(reason)
      throw new Error(reason)
    }
    setRows(parseRows(res.data))
    setFetchError(null)
  }

  // Shared auto-refresh: skip-if-in-flight, pause on hidden tab, error
  // backoff (replaces the bare setInterval).
  useAutoRefresh(fetchTickers, REFRESH_TICKER_MS, { runOnMount: true })

  if (loading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        {SYMBOLS.map((sym) => (
          <div
            key={sym}
            style={{
              padding: '12px',
              background: 'rgba(255,255,255,0.02)',
              borderRadius: 10,
              border: '1px solid rgba(255,255,255,0.04)',
              height: 56,
            }}
          >
            <div
              style={{
                width: '60%',
                height: 10,
                background: 'rgba(255,255,255,0.04)',
                borderRadius: 4,
                animation: 'pulse 1.5s infinite',
              }}
            />
          </div>
        ))}
        <style>{`
          @keyframes pulse {
            0%, 100% { opacity: 0.4; }
            50% { opacity: 0.8; }
          }
        `}</style>
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      {SYMBOLS.map((sym) => {
        const row = rows[sym]
        const price = row ? row.price : null
        const ch1h = row ? row.change_1h_pct : null
        const ch4h = row ? row.change_4h_pct : null
        // No row for a requested symbol is itself a named state: the request
        // failed (its reason) or the server returned nothing for it.
        const reason = row
          ? row.unavailable
          : fetchError || t('agentTicker.noRow', language)
        const sourceLabel = row
          ? SOURCE_LABELS[row.source] || row.source || NA
          : NA
        const color = pctColor(ch1h)
        const bgColor =
          ch1h === null || ch1h === 0
            ? 'rgba(108,108,130,0.06)'
            : ch1h > 0
              ? 'rgba(0,229,160,0.06)'
              : 'rgba(246,70,93,0.06)'
        const icon = SYMBOL_ICONS[sym] || sym[0]

        return (
          <div
            key={sym}
            data-testid={`ticker-row-${sym}`}
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: 6,
              padding: '10px 11px',
              background: 'rgba(255,255,255,0.02)',
              borderRadius: 10,
              border: '1px solid rgba(255,255,255,0.04)',
              transition: 'all 0.15s ease',
              cursor: 'default',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'rgba(255,255,255,0.04)'
              e.currentTarget.style.borderColor = 'rgba(255,255,255,0.08)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'rgba(255,255,255,0.02)'
              e.currentTarget.style.borderColor = 'rgba(255,255,255,0.04)'
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <div
                  style={{
                    width: 28,
                    height: 28,
                    borderRadius: 8,
                    background: bgColor,
                    display: 'grid',
                    placeItems: 'center',
                    fontSize: 13,
                    fontWeight: 700,
                    color: color,
                    fontFamily: 'system-ui',
                  }}
                >
                  {icon}
                </div>
                <div>
                  <div
                    style={{
                      fontSize: 12.5,
                      fontWeight: 600,
                      color: '#e0e0ec',
                      letterSpacing: '-0.01em',
                    }}
                  >
                    {sym}
                  </div>
                  <div
                    data-testid={`ticker-source-${sym}`}
                    style={{ fontSize: 10, color: '#4c4c62' }}
                  >
                    {t('agentTicker.source', language, {
                      source: sourceLabel,
                    })}
                  </div>
                </div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div
                  data-testid={`ticker-price-${sym}`}
                  style={{
                    fontSize: 12.5,
                    fontWeight: 600,
                    color: '#e0e0ec',
                    fontFamily: '"IBM Plex Mono", monospace',
                    letterSpacing: '-0.02em',
                  }}
                >
                  {formatPrice(price)}
                </div>
                <div
                  style={{
                    fontSize: 10.5,
                    fontWeight: 600,
                    fontFamily: '"IBM Plex Mono", monospace',
                    color: '#6c6c82',
                  }}
                >
                  {t('agentTicker.change1h', language)}{' '}
                  <span
                    data-testid={`ticker-1h-${sym}`}
                    style={{ color: pctColor(ch1h) }}
                  >
                    {formatPct(ch1h)}
                  </span>
                  {' · '}
                  {t('agentTicker.change4h', language)}{' '}
                  <span
                    data-testid={`ticker-4h-${sym}`}
                    style={{ color: pctColor(ch4h) }}
                  >
                    {formatPct(ch4h)}
                  </span>
                </div>
              </div>
            </div>
            {reason && (
              <div
                data-testid={`ticker-unavailable-${sym}`}
                style={{ fontSize: 10.5, color: '#d4a84b', lineHeight: 1.4 }}
              >
                {t('agentTicker.unavailable', language, { reason })}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
