// P4.3 — the rules block: no-trade windows + the plan-death line, in a
// short-tinted container (this is where the plan STOPS trading).
//
// NO-TRADE BAND (2026-09-02). This block used to render doc.no_trade verbatim:
// prose the model wrote when the plan was authored, shown as live rules for the
// whole session. An ASIA card at 23:00 CT listed three constraints — the first
// 5m (spent six hours earlier), the NY lunch window, and a 09:00 red-news
// blackout — while none of them could refuse a single entry.
//
// The band is the machine's own windows (same definitions the entry gate and
// the adherence grader use), each stamped by the server with a status for the
// clock the reader is holding. This component decides nothing: it renders the
// live ones as rules, collapses the rest, and demotes the model's prose to
// notes. No times, thresholds or window names are written here.

import { useState } from 'react'
import type { Language } from '../../i18n/translations'
import { tp } from '../../i18n/plan-translations'

export type NoTradeBandWindow = {
  start_min: number
  end_min: number
  start_ct: string
  end_ct: string
  kind: string
  label: string
  source: string
  status: 'live' | 'elapsed' | 'other_session'
}

export function RulesBlock({
  noTrade,
  band,
  deathCondition,
  language,
}: {
  noTrade: string[]
  band?: NoTradeBandWindow[]
  deathCondition: string
  language: Language
}) {
  const [showSpent, setShowSpent] = useState(false)

  const hasBand = Array.isArray(band) && band.length > 0
  const live = hasBand ? band!.filter((w) => w.status === 'live') : []
  const spent = hasBand ? band!.filter((w) => w.status !== 'live') : []
  const prose = noTrade && noTrade.length > 0 ? noTrade : []

  // Pre-band docs carry no machine windows: keep rendering the prose as rules
  // rather than claiming a plan has no constraints.
  const proseIsRules = !hasBand
  if (!hasBand && prose.length === 0 && !deathCondition) return null

  const rowStyle = { fontFamily: 'var(--vl-font-ui)' } as const

  return (
    <div
      className="flex flex-col gap-1.5 p-2.5"
      style={{
        background: 'rgba(224,108,108,0.06)',
        border: '1px solid rgba(224,108,108,0.22)',
        borderRadius: 'var(--vl-radius-inner)',
      }}
    >
      {(hasBand || proseIsRules) && (prose.length > 0 || hasBand) && (
        <div className="text-[11px]" style={rowStyle}>
          <span
            className="uppercase tracking-wide"
            style={{ color: 'var(--vl-short)' }}
          >
            {tp('noTrade', language)}
          </span>
          <span style={{ color: 'var(--vl-muted)' }}>
            {' '}
            ·{' '}
            {proseIsRules
              ? prose.join(' · ')
              : live.length > 0
                ? live
                    .map((w) => `${w.label} (${w.start_ct}–${w.end_ct} CT)`)
                    .join(' · ')
                : tp('noTradeNoneLive', language)}
          </span>
        </div>
      )}

      {spent.length > 0 && (
        <div className="text-[11px]" style={rowStyle}>
          <button
            type="button"
            aria-expanded={showSpent}
            onClick={() => setShowSpent((v) => !v)}
            className="uppercase tracking-wide"
            style={{
              color: 'var(--vl-muted)',
              background: 'none',
              border: 'none',
              padding: 0,
              cursor: 'pointer',
              font: 'inherit',
            }}
          >
            {showSpent ? '▾' : '▸'} {spent.length}{' '}
            {tp('noTradeSpent', language)}
          </button>
          {showSpent && (
            <div
              className="mt-1 flex flex-col gap-0.5"
              style={{ color: 'var(--vl-muted)', opacity: 0.7 }}
            >
              {spent.map((w, i) => (
                <div key={`${w.kind}-${w.start_min}-${i}`}>
                  {w.label} ({w.start_ct}–{w.end_ct} CT) ·{' '}
                  {w.status === 'elapsed'
                    ? tp('bandElapsed', language)
                    : tp('bandOtherSession', language)}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {hasBand && prose.length > 0 && (
        <div className="text-[11px]" style={rowStyle}>
          <span
            className="uppercase tracking-wide"
            style={{ color: 'var(--vl-muted)' }}
          >
            {tp('modelNotes', language)}
          </span>
          <span style={{ color: 'var(--vl-muted)', opacity: 0.7 }}>
            {' '}
            · {prose.join(' · ')}
          </span>
        </div>
      )}

      {deathCondition && (
        <div className="text-[11px]" style={rowStyle}>
          <span
            className="uppercase tracking-wide"
            style={{ color: 'var(--vl-short)' }}
          >
            {tp('planDies', language)}
          </span>
          <span style={{ color: 'var(--vl-muted)' }}> · {deathCondition}</span>
        </div>
      )}
    </div>
  )
}
