// W7 (weekly-bias wave, 2026-08-30) — the WEEKLY chip on the day-plan card.
// Four states: active bias (arrow + conviction dot + draw px) · grey "none" ·
// strikethrough invalidated · thin history. Tooltip = the doc's ≤3-line
// narrative. Pure view — the weekly doc never gates the session plan.
import type { PlanWeekly } from '../../lib/api/plan'

const CONV_COLOR: Record<string, string> = {
  high: 'var(--vl-long)',
  med: 'var(--vl-gold)',
  low: 'var(--vl-muted)',
}

export function WeeklyChip({ weekly }: { weekly?: PlanWeekly | null }) {
  if (!weekly) {
    return (
      <span
        data-testid="weekly-chip"
        title="WEEKLY: none — no Sunday weekly read stored for this week (plans render the none form; nothing else changes)."
        style={{
          fontSize: 9,
          fontWeight: 700,
          letterSpacing: '.08em',
          color: 'var(--vl-faint)',
          border: '1px solid var(--vl-hair)',
          borderRadius: 5,
          padding: '2px 6px',
          fontFamily: 'var(--vl-font-ui)',
        }}
      >
        WEEKLY none
      </span>
    )
  }
  const bias = (weekly.bias ?? '').toLowerCase()
  const conv = (weekly.conviction ?? '').toLowerCase()
  const invalidated = Boolean(weekly.invalidated_at)
  const arrow = bias === 'bear' ? '▼' : bias === 'bull' ? '▲' : '•'
  const color =
    bias === 'bull'
      ? 'var(--vl-long)'
      : bias === 'bear'
        ? 'var(--vl-short)'
        : 'var(--vl-muted)'
  const tooltip = [
    weekly.thin_history
      ? `WEEKLY: thin history — low conviction`
      : invalidated
        ? `WEEKLY: neutral (invalidated ${weekly.invalidated_at})`
        : `WEEKLY: ${bias}/${conv} · draw ${weekly.draw_name} ${weekly.draw_px} · invalid ${weekly.invalidation_px} (${weekly.invalidation_basis})`,
    ...(weekly.narrative ? weekly.narrative.split('\n').slice(0, 3) : []),
  ].join('\n')

  return (
    <span
      data-testid="weekly-chip"
      title={tooltip}
      style={{
        fontSize: 9,
        fontWeight: 700,
        letterSpacing: '.08em',
        color: invalidated ? 'var(--vl-faint)' : color,
        border: `1px solid ${invalidated ? 'var(--vl-hair)' : color}`,
        borderRadius: 5,
        padding: '2px 6px',
        fontFamily: 'var(--vl-font-ui)',
        textDecoration: invalidated ? 'line-through' : 'none',
        opacity: invalidated ? 0.7 : 1,
      }}
    >
      {arrow} WEEKLY {weekly.thin_history ? `thin · ${conv}` : bias}
      {!weekly.thin_history && !invalidated && (
        <span style={{ color: CONV_COLOR[conv] ?? 'var(--vl-muted)' }}>
          {' '}
          ·{conv} {weekly.draw_px > 0 ? weekly.draw_px : ''}
        </span>
      )}
    </span>
  )
}
