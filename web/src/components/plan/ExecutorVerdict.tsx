// W-ARM-STATE-UI (2026-09-18) — the EXECUTOR's verdict per scenario, rendered
// beside the evaluator's. Sourced ONLY from armed_orders rows (plan.armed,
// served by armedMapFor) and the executor geometry records
// (plan.structural_geometry, served by planStructuralGeometry). No record for
// the displayed plan version → the column renders NOTHING (no dash, no "ok").

import type { PlanArmView, StructuralGeometryView } from '../../lib/api/plan'

export type ExecutorVerdictState =
  | 'armed'
  | 'filled'
  | 'cancelled'
  | 'refused'
  | 'not_attempted'

export interface ExecutorLine {
  state: ExecutorVerdictState
  orderId?: number
  reason?: string
  detail?: string
  timeMs?: number
  label: string
}

function idSuffix(id?: number): string {
  return typeof id === 'number' ? ` #${id}` : ''
}

function fmtTime(ms: number): string {
  const d = new Date(ms)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// The refusal vocabulary the executor writes (trader/structural_geometry.go
// composeGeometry refuse() reasons + the atr_fallback refusal class).
const REFUSAL_REASONS = new Set([
  'no_provenance',
  'scenario_level_id_missing',
  'no_target',
  'invalid_geometry',
  'net_nonpositive',
  'rr',
  'atr_fallback',
])

function lineForState(
  state: string,
  rowId: number | undefined,
  reason: string | undefined
): ExecutorLine | null {
  if (state === 'filled')
    return {
      state: 'filled',
      orderId: rowId,
      reason,
      label: `filled${idSuffix(rowId)}`,
    }
  if (state === 'armed' || state === 'place_pending' || state === 'working')
    return {
      state: 'armed',
      orderId: rowId,
      reason,
      label: `armed${idSuffix(rowId)}`,
    }
  if (state === 'cancelled' || state === 'rejected' || state === 'expired')
    return {
      state: 'cancelled',
      orderId: rowId,
      reason,
      label: `cancelled: ${reason || state}`,
    }
  return null
}

/**
 * Pure derivation — exported so the vitest pins exercise the production call
 * site. armed_orders speaks first (the executor's final word for this plan
 * version); state UNKNOWN means the ledger has no row for this version — it is
 * NOT a verdict, so the geometry records speak next.
 */
export function executorLinesFor(
  scenario: string,
  arm: PlanArmView | undefined,
  geometry: StructuralGeometryView[] | null | undefined
): ExecutorLine[] {
  if (arm && arm.state !== 'UNKNOWN') {
    const legs = (arm.legs ?? []).filter((l) => l.state !== 'UNKNOWN')
    if (arm.state === 'mixed') {
      const lines = legs
        .map((l) => lineForState(l.state, l.row_id, l.reason))
        .filter((l): l is ExecutorLine => l !== null)
      if (lines.length > 0) return lines
    } else {
      const line = lineForState(arm.state, legs[0]?.row_id, arm.reason)
      if (line) return [line]
    }
  }
  const rows = (geometry ?? []).filter((r) => r.scenario === scenario)
  if (rows.length === 0) return []
  const latest = rows.reduce((a, b) =>
    (b.time_ms ?? 0) > (a.time_ms ?? 0) ? b : a
  )
  if (REFUSAL_REASONS.has(latest.reason))
    return [
      {
        state: 'refused',
        reason: latest.reason,
        detail: latest.detail,
        timeMs: latest.time_ms,
        label: `refused: ${latest.reason}${latest.detail ? ` (${latest.detail})` : ''}`,
      },
    ]
  if (
    latest.reason === 'pending_gates' ||
    (latest.reason === '' && latest.quantity === 0)
  )
    return [
      {
        state: 'not_attempted',
        reason: latest.reason || 'pending_gates',
        detail: latest.detail,
        timeMs: latest.time_ms,
        label: 'not attempted',
      },
    ]
  if (latest.quantity > 0)
    // Admitted geometry whose order row is not on this version — truth: an
    // admitted arm, no order id to cite.
    return [
      {
        state: 'armed',
        reason: 'admitted',
        detail: latest.detail,
        timeMs: latest.time_ms,
        label: 'armed',
      },
    ]
  return []
}

const lineColor = (line: ExecutorLine) =>
  line.state === 'filled'
    ? 'var(--vl-long)'
    : line.state === 'armed'
      ? 'var(--vl-gold)'
      : line.state === 'refused' || line.state === 'cancelled'
        ? 'var(--vl-short)'
        : 'var(--vl-faint)'

export function ExecutorVerdict({
  scenario,
  arm,
  geometry,
}: {
  scenario: string
  arm?: PlanArmView
  geometry?: StructuralGeometryView[] | null
}) {
  const lines = executorLinesFor(scenario, arm, geometry)
  if (lines.length === 0) return null
  return (
    <span
      data-testid={`executor-verdict-${scenario}`}
      className="text-[9px] font-bold px-1.5 py-0.5 rounded"
      style={{
        border: '1px solid var(--vl-hair)',
        background: 'transparent',
        color: lineColor(lines[0]),
        display: 'inline-flex',
        gap: 4,
        alignItems: 'center',
      }}
    >
      {lines.map((line, i) => (
        <span
          key={i}
          title={[
            line.reason,
            line.detail,
            line.timeMs ? fmtTime(line.timeMs) : '',
          ]
            .filter(Boolean)
            .join(' · ')}
        >
          {line.label}
        </span>
      ))}
    </span>
  )
}
