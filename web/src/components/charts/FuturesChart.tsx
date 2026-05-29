import { useEffect, useRef, useState } from 'react'
import {
  createChart,
  IChartApi,
  ISeriesApi,
  UTCTimestamp,
  CandlestickSeries,
  HistogramSeries,
} from 'lightweight-charts'

// Raw bar off the SSE wire (provider/ninjatrader Bar — epoch-ms time).
interface WireBar {
  t: number // Unix epoch ms, UTC
  o: number
  h: number
  l: number
  c: number
  v: number
}

interface FuturesChartProps {
  symbol?: string
  interval?: string // timeframe token, e.g. "5m"
  traderID?: string
  height?: number
}

const UP = '#0ECB81'
const DOWN = '#F6465D'

// Wire bar (ms) → lightweight-charts candle (seconds).
function toCandle(b: WireBar) {
  return {
    time: Math.floor(b.t / 1000) as UTCTimestamp,
    open: b.o,
    high: b.h,
    low: b.l,
    close: b.c,
  }
}

function toVolume(b: WireBar) {
  return {
    time: Math.floor(b.t / 1000) as UTCTimestamp,
    value: b.v,
    color: b.c >= b.o ? 'rgba(14, 203, 129, 0.5)' : 'rgba(246, 70, 93, 0.5)',
  }
}

/**
 * FuturesChart renders live NT8 candles streamed over SSE from the SAME
 * BarCache the kernel reads (Plan 4.4 Stage 4), so the chart and AI decisions
 * see identical data. EventSource cannot set an Authorization header, so the
 * JWT rides in the ?token= query param.
 */
export function FuturesChart({
  symbol = 'MNQ',
  interval = '5m',
  traderID,
  height = 550,
}: FuturesChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candleSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null)
  const didFitRef = useRef(false)

  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [lastPrice, setLastPrice] = useState<number | null>(null)
  const [barCount, setBarCount] = useState(0)

  // Create chart once.
  useEffect(() => {
    if (!chartContainerRef.current) return

    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth || 800,
      height: chartContainerRef.current.clientHeight || height,
      layout: {
        background: { color: '#0B0E11' },
        textColor: '#B7BDC6',
        fontSize: 12,
      },
      grid: {
        vertLines: { color: 'rgba(43, 49, 57, 0.2)', style: 1, visible: true },
        horzLines: { color: 'rgba(43, 49, 57, 0.2)', style: 1, visible: true },
      },
      crosshair: { mode: 1 },
      rightPriceScale: {
        borderColor: '#2B3139',
        scaleMargins: { top: 0.1, bottom: 0.25 },
      },
      timeScale: {
        borderColor: '#2B3139',
        timeVisible: true,
        secondsVisible: false,
        rightOffset: 5,
        barSpacing: 8,
      },
    })
    chartRef.current = chart

    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: UP,
      downColor: DOWN,
      borderUpColor: UP,
      borderDownColor: DOWN,
      wickUpColor: UP,
      wickDownColor: DOWN,
    })
    candleSeriesRef.current = candleSeries as any

    const volumeSeries = chart.addSeries(HistogramSeries, {
      color: '#26a69a',
      priceFormat: { type: 'volume' },
      priceScaleId: '',
      lastValueVisible: false,
      priceLineVisible: false,
    })
    volumeSeries.priceScale().applyOptions({
      scaleMargins: { top: 0.8, bottom: 0 },
    })
    volumeSeriesRef.current = volumeSeries as any

    const resizeObserver = new ResizeObserver((entries) => {
      if (entries.length === 0 || !entries[0].contentRect) return
      const { width, height } = entries[0].contentRect
      chart.applyOptions({ width, height })
    })
    resizeObserver.observe(chartContainerRef.current)

    return () => {
      resizeObserver.disconnect()
      chart.remove()
      chartRef.current = null
      candleSeriesRef.current = null
      volumeSeriesRef.current = null
    }
  }, [])

  // Open the SSE stream whenever the target changes.
  useEffect(() => {
    didFitRef.current = false
    setError(null)
    setConnected(false)
    setBarCount(0)

    const token = localStorage.getItem('auth_token')
    if (!token) {
      setError('Not authenticated')
      return
    }

    const params = new URLSearchParams({ symbol, tf: interval, token })
    if (traderID) params.set('trader_id', traderID)
    const url = `/api/v1/bars/stream?${params.toString()}`
    const es = new EventSource(url)

    es.addEventListener('open', () => setConnected(true))

    es.addEventListener('snapshot', (e: MessageEvent) => {
      setConnected(true)
      try {
        const payload = JSON.parse(e.data) as { bars: WireBar[] | null }
        const bars = payload.bars || []
        // lightweight-charts needs ascending, unique times.
        const sorted = [...bars].sort((a, b) => a.t - b.t)
        const deduped = sorted.filter(
          (b, i, arr) => i === 0 || b.t !== arr[i - 1].t
        )
        candleSeriesRef.current?.setData(deduped.map(toCandle))
        volumeSeriesRef.current?.setData(deduped.map(toVolume))
        setBarCount(deduped.length)
        if (deduped.length > 0) {
          setLastPrice(deduped[deduped.length - 1].c)
        }
        if (!didFitRef.current && deduped.length > 0) {
          chartRef.current?.timeScale().fitContent()
          didFitRef.current = true
        }
      } catch (err) {
        console.error('[FuturesChart] bad snapshot:', err)
      }
    })

    es.addEventListener('bar', (e: MessageEvent) => {
      try {
        const b = JSON.parse(e.data) as WireBar
        // update() updates the matching time or appends a newer one.
        candleSeriesRef.current?.update(toCandle(b))
        volumeSeriesRef.current?.update(toVolume(b))
        setLastPrice(b.c)
      } catch (err) {
        console.error('[FuturesChart] bad bar:', err)
      }
    })

    es.onerror = () => {
      // EventSource auto-reconnects; surface the gap in the UI meanwhile.
      setConnected(false)
    }

    return () => es.close()
  }, [symbol, interval, traderID])

  return (
    <div
      className="relative shadow-xl"
      style={{
        background: 'linear-gradient(180deg, #0F1215 0%, #0B0E11 100%)',
        borderRadius: '12px',
        overflow: 'hidden',
        border: '1px solid rgba(43, 49, 57, 0.5)',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <div
        className="flex items-center justify-between px-4 py-2"
        style={{
          borderBottom: '1px solid rgba(43, 49, 57, 0.6)',
          background: '#0D1117',
          flexShrink: 0,
        }}
      >
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <span className="text-sm font-bold text-white">{symbol}</span>
            <span className="text-[10px] px-1.5 py-0.5 rounded bg-[#1F2937] text-gray-400">
              {interval}
            </span>
            <span
              className="text-[10px] px-1.5 py-0.5 rounded font-medium uppercase"
              style={{
                background: 'rgba(80, 227, 194, 0.1)',
                color: '#50E3C2',
              }}
            >
              NT8 LIVE
            </span>
          </div>
          {lastPrice !== null && (
            <span className="text-base font-bold tabular-nums text-white pl-3 border-l border-[#2B3139]">
              {lastPrice.toLocaleString(undefined, {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2,
              })}
            </span>
          )}
        </div>

        <div className="flex items-center gap-1.5">
          <span
            className="inline-block w-2 h-2 rounded-full"
            style={{
              background: connected ? UP : '#6B7280',
              boxShadow: connected ? `0 0 6px ${UP}` : 'none',
            }}
          />
          <span
            className="text-[11px]"
            style={{ color: connected ? UP : '#6B7280' }}
          >
            {connected ? 'Live' : 'Connecting…'}
          </span>
        </div>
      </div>

      <div style={{ position: 'relative', flex: 1, minHeight: 0 }}>
        <div
          ref={chartContainerRef}
          style={{ height: '100%', width: '100%' }}
        />

        {/* Empty-state overlay: connected but BarCache has no bars yet. */}
        {connected && barCount === 0 && !error && (
          <div
            className="absolute inset-0 flex items-center justify-center pointer-events-none"
            style={{ color: '#6B7280' }}
          >
            <div className="text-center">
              <div className="text-sm">Waiting for {symbol} bars…</div>
              <div className="text-[11px] mt-1 opacity-70">
                NT8 will seed the cache on the next bar
              </div>
            </div>
          </div>
        )}
      </div>

      {error && (
        <div
          className="absolute inset-0 flex items-center justify-center"
          style={{ background: 'rgba(11, 14, 17, 0.9)' }}
        >
          <div className="text-center">
            <div className="text-2xl mb-2">⚠️</div>
            <div style={{ color: DOWN }}>{error}</div>
          </div>
        </div>
      )}
    </div>
  )
}
