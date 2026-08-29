import { GUIDE_BUILT_REV, type GuideSection } from '../types'

export const plays: GuideSection = {
  id: 'plays',
  num: 5,
  title: 'The Plays',
  tagline: 'Eight condition templates + the A-setup chain.',
  asBuiltRev: GUIDE_BUILT_REV,
  blocks: [
    {
      kind: 'cards',
      cards: [
        {
          title: 'reclaim',
          body: 'Setup: level touched → price closes back on the original side → confirm: 1×5m close beyond (rule 1x5m_close). The planner writes it when a level was lost and re-taken. Expectancy: the bread-and-butter retest play.',
          tag: 'play 1',
        },
        {
          title: 'hold',
          body: 'Setup: price holds a level (no close beyond) → confirm: the hold itself. Written when the plan expects support/resistance to keep working.',
          tag: 'play 2',
        },
        {
          title: 'sweep_reclaim',
          body: 'Setup: level swept (liquidity taken) → close back inside → confirm: the reclaim CLOSE (the prompt says it: "sweep_reclaim requires the reclaim CLOSE"). The house play at ONH/ONL. Expectancy: week 0% / −192 when misplayed as fade-the-poke; the sweep first is what makes it valid.',
          tag: 'play 3',
        },
        {
          title: 'reject',
          body: 'Setup: first touch rejected at a zone → confirm: 2×5m closes beyond or a 15m close. NY stat: 75% win / +665 (the best measured condition×session pair).',
          tag: 'play 4 · 75%',
        },
        {
          title: 'acceptance',
          body: 'Setup: closes through a level and holds → confirm: acceptance rule. HOUSE RULE: acceptance requires a PRIOR sweep + displacement or the plan skips it (0% win evidence for bare acceptance).',
          tag: 'play 5',
        },
        {
          title: 'breakout_retest',
          body: 'Setup: break above/below → first retest holds → confirm: 1×5m close back on the breakout side. The ONH breakout play when price is at the top of the stack.',
          tag: 'play 6',
        },
        {
          title: 'fvg_entry',
          body: "Setup: the A-setup's finish — displacement leaves a gap → first retrace INTO the FVG → confirm: touch at the gap band. entry_mode ce by default (edge only for A-grade HTF-confluent origins); stop beyond the sweep extreme; T1 = first opposing pool; runner = the draw. chain_after links it to its sweep_reclaim. FVG DEMAND (A2c): a fresh machine gap matching your bias expects an fvg_entry — or a one-line reason why not.",
          tag: 'play 7',
        },
        {
          title: 'breakdown_continue',
          body: 'Setup: price closes BELOW a broken level with displacement ≥ 1.0×ATR5m (BD_MIN_DISP_ATR) and no reclaim close → the waterfall SHORT. Two-leg confirm: leg 1 = the breakdown close, leg 2 = the retest that FAILS to reclaim (pullback < 0.4× the break leg). entry_mode pullback rests at the broken level (ARM path); immediate is now PLAN-AUTHORABLE on the AI path (market entry on the 2nd confirming close through the FULL gate chain — no arm; arms stay pullback-only). CHOOSE immediate for no-retest waterfalls (displacement expanding), pullback otherwise. Born 2026-08-28: the −347pt NY crash — bias right, but NO plan-legal continuation short existed (v4, S2 declined on the retest rule).',
          tag: 'play 8 · −347',
        },
        {
          title: 'breakup_continue',
          body: 'The LONG mirror: closes ABOVE a broken level with displacement ≥ 1.0×ATR5m, no reclaim → continuation long; same two-leg confirm. entry_mode pullback = arm at the broken level; immediate = AI-path market entry on the 2nd confirming close (full gate chain, no arm) — choose immediate for no-retest breakouts, pullback otherwise.',
          tag: 'play 8b',
        },
      ],
    },
    { kind: 'h', text: 'THE A-SETUP (the chained play)' },
    {
      kind: 'code',
      title: 'sweep → displacement → FVG retrace, at a fresh Tier-1 origin',
      lines: [
        'S1 sweep_reclaim : liquidity pool swept (ONH/EQH) + reclaim close',
        '        │  chain_after: S1',
        '        ▼',
        'S2 fvg_entry    : the displacement gap retraces into the FVG band',
        '        entry_mode = ce (edge only for A-grade HTF-confluent origins)',
        '        stop beyond the sweep extreme · T1 = first opposing pool',
        '        runner = the draw (nearest opposing liquidity beyond T1)',
        '',
        'a bare fvg_entry with NO sweep precursor at a non-A/B origin',
        'gets a WARN at write — the raw-FVG null (40k sample) says it',
        'has no standalone edge.',
      ],
    },
    { kind: 'h', text: 'No-trade is a trade' },
    {
      kind: 'p',
      text: 'Sitting out is a position. The gates the planner is taught: balance day → edges only or skip · opening gap >1.2×ATR or outside-range open → never fade · no A/B zone in reach AND no pool swept by 10:30 ET → skip the day · lunch 11:30–13:30 ET → no entries · Tier-1 news → stand aside. A real skip looks like: "no_trade: balance day — no A/B zone in reach by 10:30 ET, skip" — that is a decision, not a failure.',
    },
  ],
}
