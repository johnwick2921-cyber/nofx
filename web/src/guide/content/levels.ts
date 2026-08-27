import { GUIDE_BUILT_REV, type GuideSection } from '../types'

export const levels: GuideSection = {
  id: 'levels',
  num: 4,
  title: 'The Level System',
  tagline: 'levels = WHERE · roles = WHAT-FOR · grades = HOW-STRONG.',
  asBuiltRev: GUIDE_BUILT_REV,
  blocks: [
    {
      kind: 'p',
      text: 'Every plan and every decision hangs on levels. A level is a price the machine detected (structure, volume, imbalance); a role says what to DO with it; a grade says how trustworthy it is.',
    },
    { kind: 'h', text: 'The kinds — Anchors / Volume / Zones' },
    {
      kind: 'cards',
      cards: [
        {
          title: 'PDH / PDL / PDC',
          body: "Prior-day high/low/close — the dealing-range anchors the bias-tree branches on. Why care: the tree's branches 1–3 are defined on them. Role: react zone. Grade: A (coverage-guarded — a truncated day is omitted, never fabricated).",
          tag: 'Anchors',
        },
        {
          title: 'ONH / ONL',
          body: 'Overnight (Asia+London composite) high/low. Why care: backtested 94.2% broken eventually (2,827-day NQ) — treat as liquidity, fade only on a confirmed sweep-reclaim. Role: liquidity/breakout.',
          tag: 'Anchors · 94.2%',
        },
        {
          title: 'RTH-H / RTH-L · AS-H/L · LDN-H/L',
          body: 'Session-scoped highs/lows. Why care: they delimit where each shift traded. Role: react zone.',
          tag: 'Anchors',
        },
        {
          title: 'OR-H / OR-L · IB-H / IB-L',
          body: "Opening-range (5m) and initial-balance (60m) extremes + extensions. Why care: the first-hour range defines the day's frame. Role: react zone.",
          tag: 'Anchors',
        },
        {
          title: 'PWH/PWL · PMH/PML',
          body: 'Prior week / prior month extremes. Why care: the HTF context the 1h/4h sections quote. Role: target_only for far-HTF continuation zones.',
          tag: 'Anchors',
        },
        {
          title: 'VWAP ±1σ · eVWAP',
          body: 'Session volume-weighted average + its band; eVWAP anchored at 15:00 CT. Why care: the institutional magnet — mean-revert entries look here. Role: magnet/mean-revert.',
          tag: 'Volume',
        },
        {
          title: 'POC / VAH / VAL · nPOC · SETT · MID-O',
          body: 'Value-area profile (point of control, area high/low), naked POC, settlement, mid-of-overnight. Why care: where the market did business — magnets and targets. Tier-1 since R-A13.',
          tag: 'Volume',
        },
        {
          title: 'S/D · FVG · iFVG · OB',
          body: "Supply/demand zones, fair-value gaps (inverted when filled), order blocks. Why care: the playbook's entry surface (sweep → displacement → FVG retrace). FVG gap floor 2pt/8t; session-break guard on.",
          tag: 'Zones',
        },
        {
          title: 'EQH / EQL',
          body: "Equal highs/lows (3-tick tolerance) — resting liquidity. Why care: sweeps of equal highs are the A-setup's opening move.",
          tag: 'Zones',
        },
        {
          title: 'RN (round numbers)',
          body: '100/50/25-point steps generated inside the band. Why care: self-fulfilling pauses; the card labels them RN.',
          tag: 'Zones',
        },
      ],
    },
    { kind: 'h', text: 'Grading pipeline' },
    {
      kind: 'code',
      title: 'evidence × freshness × confluence × TF, with floors/caps',
      lines: [
        'zones: zoneEvidence(kind, TF, reversal×1.1) × zoneSizeMult(0.5–1.25)',
        '       × freshness × (1 + 0.20×confluence) × zoneTFMult(1.0/1.1/1.2/1.3)',
        'lines: typeEvidence(kind) × freshness × (1 + 0.20×conf) × htf(×1.2)',
        '',
        'freshness ladder (zones): 1.0 / 0.6 / 0.3 / 0.15',
        'confluence: distinct families only, cap 3 → ×1.6 max',
        '',
        'floors/caps: 1m zones forced C · 15m forced B (both ways) ·',
        '  1h/4h floor B (may reach A) · above-C only within 12 ticks',
        '  of a Tier-1 anchor (B2 gate)',
        '',
        '3 FVGs of the same family = 1 family entry (one seat, not three)',
      ],
    },
    { kind: 'h', text: 'The 5 roles — what a level is FOR' },
    {
      kind: 'cards',
      cards: [
        {
          title: 'MAGNET / MEAN-REVERT',
          body: 'VWAP family, POC. Invited play: fade extremes back to it, target it. Forbidden play: breakout through it without a sweep story.',
          tag: 'ROLE',
        },
        {
          title: 'LIQUIDITY / BREAKOUT',
          body: 'ONH/ONL, EQH/EQL. Invited play: wait for the sweep, then the reclaim — never fade the first poke. The ONH story: fade-the-poke was the −131 week; sweep-reclaim is the house rule.',
          tag: 'ROLE',
        },
        {
          title: 'REACT ZONE',
          body: 'PDH/PDL, session highs/lows, OR/IB. Invited play: first-touch reactions with confirmation. Forbidden: chasing a close beyond without a retest.',
          tag: 'ROLE',
        },
        {
          title: 'TARGET ONLY',
          body: 'Far-HTF continuation zones. Invited play: take profit / trail. Forbidden: entry against the HTF trend at these.',
          tag: 'ROLE',
        },
        {
          title: 'PIVOT',
          body: "The bias-flip reference (env LEVEL_ROLE_MAP). The plan's flip line anchors here.",
          tag: 'ROLE',
        },
      ],
    },
    { kind: 'h', text: 'Seats & band' },
    {
      kind: 'p',
      text: 'Why 8 (max_levels): a tight table the AI can actually copy without hallucinating. The ±band (proximity_filter_atr × daily-range, retuned 0.3 → ≈±100pt): day-trade relevance — far levels still exist in the detector universe (they now feed the bias-tree anchors even unseated). min_side_levels (default 2): the per-side floor — a machine-thin side writes with a ⚖ note instead of failing the whole session (that exact failure sat ASIA out on 08-26).',
    },
  ],
}
