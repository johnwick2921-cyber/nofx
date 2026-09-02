import { GUIDE_BUILT_REV, type GuideSection } from '../types'

export const guards: GuideSection = {
  id: 'guards',
  num: 6,
  title: 'Guards & Safety',
  tagline: 'What can hard-block a trade vs what only informs.',
  asBuiltRev: GUIDE_BUILT_REV,
  blocks: [
    { kind: 'h', text: 'CAN-HARD-BLOCK vs ADVISORY-ONLY (the truth table)' },
    {
      kind: 'table',
      head: ['Gate', 'Kind', 'What it does'],
      rows: [
        [
          'SIM lock (isAccountTradeable)',
          'HARD',
          'Refuses non-SIM accounts at account routing — the bot cannot go live.',
        ],
        [
          'Feed down / dead-man / freeze / boot integrity',
          'HARD',
          'No bars, no bridge, no trading — cycle skipped or refused.',
        ],
        [
          'Boot sweep (class 33)',
          'HARD',
          "At boot, before anything is armed: every resting order left behind by the PREVIOUS process is cancelled at NinjaTrader and marked cancelled in the ledger (reason boot_sweep). A cancel that FAILS leaves the row live and retries — the ledger never goes clean while an order might still be at the broker. On 2026-09-02 00:16 CT, before this existed, two arms outlived their process for 15 minutes and briefly double-ordered S3.",
        ],
        [
          'plan_mode direction/strict',
          'HARD',
          'Refuses entries against plan bias (direction) or without a cited scenario (strict); no plan + direction/strict = no trades.',
        ],
        [
          'min_confidence',
          'HARD',
          'Confidence below the floor (default 60) → entry refused.',
        ],
        [
          'MIN-SL (env MIN_SL_ATR_MULT, 1.0)',
          'HARD',
          'Stop closer than the floor (×ATR + 2-tick clearance) → refused.',
        ],
        [
          'HTF veto',
          'HARD',
          "Entry against the HTF regime at a veto anchor → refused. MODE (HTF_VETO_MODE): 1h | cross | 4h — LIVE = cross: vetoes only when 1h AND 4h both agree (the 2026-08-28 autopsy: 1h-only blocked 3 would-have-won arms = +$352, 4h was RANGING at all 7 → cross blocks nothing the evidence doesn't support).",
        ],
        [
          'ARM floors (ARM_MIN_RR 2.0)',
          'HARD',
          'The resting-order gate: R:R ≥ 2.0 AND stop ≥ 1.0×ATR5m or the arm is REFUSED every cycle. The 3.0 floor stays the FULL market-entry gate — a resting order pre-commits at a lower bar because the entry IS the plan.',
        ],
        [
          'T1 red news blackout',
          'HARD',
          'No entries in the ±15m window around T1 events (calendar).',
        ],
        [
          'Lunch / session windows / EOD flat',
          'HARD',
          'Clock gates: no entries 12:00–13:30 CT; flat at session end.',
        ],
        [
          'Consecutive-loss halt (guardrails ON)',
          'HARD',
          'N losers in a row → halt (guardrails master must be ON).',
        ],
        [
          'Side-quota (0-on-a-side / empty map)',
          'HARD',
          'A one-sided plan or an empty machine map fail-closes the read.',
        ],
        [
          'Confirm MET / stale-MET',
          'ADVISORY',
          'Informs the AI + card. Never blocks.',
        ],
        ['Touch chips ○◐✕▲', 'ADVISORY', 'Telemetry only.'],
        [
          'fvg IN-ZONE/ABOVE/BELOW/FILLED_INVALID',
          'ADVISORY',
          'Informs. Never blocks.',
        ],
        [
          'quality A+/A/B/C + m: machine grade',
          'ADVISORY',
          'Informational (D3 ruling) — no gate consumes them.',
        ],
        ['scenario status dots', 'ADVISORY', 'Read-only backend state.'],
        [
          'chain warnings / role mismatches',
          'ADVISORY',
          'Warn at write, never a fail.',
        ],
      ],
    },
    { kind: 'h', text: 'The refusal decoder' },
    {
      kind: 'callout',
      title: 'every refusal is a named string — here is the translation',
      items: [
        {
          title: 'confidence too low (N), must be ≥M',
          body: "The AI's confidence was under min_confidence. Not a bug — the bar.",
          cite: 'kernel/engine_position.go:188',
        },
        {
          title: 'no matched scenario cited (strict mode)',
          body: "plan_mode=strict and the action didn't cite an armed S#. The plan is the law.",
          cite: 'trader/auto_trader_planconfig.go:206-249',
        },
        {
          title: 'against the plan (direction mode)',
          body: 'The bias is long and the entry is short. Advisory says fine; direction says no.',
        },
        {
          title: '|entry−SL| below MIN_SL_ATR_MULT × ATR',
          body: 'The stop is too tight for the volatility floor.',
          cite: 'kernel/engine_position.go:196',
        },
        {
          title: 'past last-entry time · outside session window · lunch',
          body: 'Clock gates — see Section 2 timeline.',
        },
        {
          title: 'only N levels above price … must carry ≥Q on EACH side',
          body: 'AI-caused omission (the map had them) → the read retries; machine-caused → now a ⚖ WARN and the plan writes.',
          cite: 'kernel/plan_doc.go ValidatePlanDocWithFactsMachine',
        },
        {
          title: 'awaiting approval',
          body: 'approval_required is ON and nobody approved this session-day. Tap Approve.',
        },
        {
          title: 'gate-block counters',
          body: '"Refused this session" panel shows every label + count; reset at the 17:00 roll and on restart.',
          cite: 'web/src/components/plan/GateBlocksPanel.tsx',
        },
      ],
    },
    { kind: 'h', text: 'plan_mode — the three levels' },
    {
      kind: 'table',
      head: ['Mode', 'Blocks', 'Allows'],
      rows: [
        [
          'advisory (default)',
          'Nothing',
          'Everything — the plan informs, the AI decides.',
        ],
        [
          'direction',
          'Entries against the plan bias',
          'Entries with the bias; anything not direction-conflicting.',
        ],
        [
          'strict',
          'Entries not citing an armed scenario; ANY entry with no active plan',
          'Only on-plan, scenario-cited entries.',
        ],
      ],
    },
    {
      kind: 'p',
      text: 'Strict\'s warning, plain: "no plan = no trades" — a fail-closed day in strict mode is a flat day, by design. Strict is the optional NY experiment. Per-session overrides exist (Strategy → Day Plan → Sessions).',
    },
    { kind: 'h', text: 'Guardrails + SIM lock' },
    {
      kind: 'p',
      text: 'Risk guardrails: the master switch (default ON, currently OFF by owner ruling — the boot log says "master OFF") arms daily loss/profit/trade limits, consecutive-loss halt, re-entry cooldown, blackout windows, max-contracts and notional caps. The always-on pair (max contracts/order, notional cap) needs no toggle. Would-have-tripped counters are visible in the dashboard. SIM lock: every account list is filtered to SIM; the bot cannot route to a live NT account — do not try.',
    },
    { kind: 'h', text: 'THE FIVE-LEG CUTOVER GATE (class 33)' },
    {
      kind: 'p',
      text: "Before any restart of the bot, GET /api/cutover-gate answers all five legs in one payload: (1) open positions in the database, (2) positions from the API, (3) the NinjaTrader positions snapshot for the bound account, (4) working orders — read from the armed_orders ledger, because NinjaTrader sends no working-order frame, and (5) in-flight planner work. ready:false means HOLD. Legs 4 and 5 are new on 2026-09-02: leg 4 used to be a stub that always answered empty, so it passed at every cutover from 35 to 41 including one with two orders resting; leg 5 did not exist, so a kill on 2026-08-31 17:34 CT landed mid-read and the planner chain died silently. A leg that cannot be evaluated counts as failed.",
    },
  ],
}
