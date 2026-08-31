import { GUIDE_BUILT_REV, type GuideSection, type KnobSpec } from '../types'

// Knob inventory = the live Strategy page controls (census 2026-08-27).
// "trader" is the one-sentence trader explanation; "consumer" is the engine
// consumer (file:line, verified). Every field is mandatory and linted by test.

const dayPlan: KnobSpec[] = [
  {
    label: 'Plan mode',
    where: 'Strategy → Day Plan → top',
    what: 'How the plan constrains entries: ADVISORY informs · DIRECTION blocks against-bias entries · STRICT blocks anything not citing an armed scenario (and ALL entries with no plan).',
    trader:
      'Strict means no-plan = flat day — the plan is the law, not advice.',
    consumer:
      'trader/auto_trader_planconfig.go:158 (planModeFor) · direction block trader/auto_trader_planconfig.go:206-249',
    range: 'advisory | direction | strict',
    systemDefault: 'advisory',
    recommended:
      '⭐ STRICT ×3 — the owner deliberately runs strict on ASIA/LONDON/NY (the plan is the law). ADVISORY remains the code fallback for anyone without a trusted plan writer.',
    whenToTouch: 'When you want the planner to have teeth (or to remove them).',
    perSession:
      'Yes — tri-state per session: inherit global / advisory / direction / strict. Session override wins; inherit (blank) = the global row above.',
  },
  {
    label: 'Proximity filter',
    where: 'Strategy → Day Plan → slider 0.1–3.0',
    what: 'The day-trade band around price that seats levels in the card (±K × the daily-range proxy, ~±300pt × K on MNQ).',
    trader:
      'Higher = tighter card; lower = wider card. Far levels still feed the bias-tree anchors even when unseated.',
    consumer:
      'kernel/levels_score.go:389 (ScoreLevels proximityK) · trader/auto_trader_planconfig.go:47',
    range: '0.1 – 3.0 (× daily-range proxy, clamped)',
    systemDefault:
      '1.5 (the CODE default — the 0.3 retune is a saved-value change in Strategy → Day Plan, never a code default; GAR-F2 archaeology: nothing ever reverted it)',
    recommended:
      '⭐ 0.3 — LIVE since 2026-08-28 11:59 (owner save); band = K × DailyRangeProxy (≈±85pt at 0.3) on BOTH the bot gate and the engine path.',
    whenToTouch:
      'If the card looks too crowded or too empty for the daily range.',
    perSession:
      'Yes — session override wins; inherit (blank) = the strategy-level value.',
  },
  {
    label: 'Max levels',
    where: 'Strategy → Day Plan → Max levels 3–12',
    what: "Cap on rows in the card's levels table (the planner is asked to write within the resolved cap).",
    trader:
      '8 = a table the AI can copy without hallucinating; 12 = wider map but more copy risk.',
    consumer:
      'kernel/levels_score.go:54 DefaultMaxLevels · trader/auto_trader_planner.go:592 (maxLevels param)',
    range: '3 – 12',
    systemDefault: '8 (owner)',
    recommended: '⭐ 8 — shipped default, verified copy fidelity.',
    whenToTouch:
      'Leave alone; raise only if you find the card missing actionable levels.',
    perSession: 'Yes.',
  },
  {
    label: 'Max scenarios',
    where: 'Strategy → Day Plan → Max scenarios 1–5',
    what: 'Cap on S# rows the planner may write.',
    trader:
      '3 = focused plays; 5 = kitchen sink, but every S# still needs a condition + invalid line.',
    consumer:
      'trader/auto_trader_planconfig.go:162 (scenarioCap) · kernel/planner_prompt.go:66 (resolved caps)',
    range: '1 – 5',
    systemDefault: '5 (owner, live)',
    recommended:
      '⭐ 5 — the live scenario cap; 3 stays a fine default elsewhere.',
    whenToTouch: 'Raise if a fast market needs more play variants.',
    perSession: 'Yes.',
  },
  {
    label: 'Max re-plans',
    where: 'Strategy → Day Plan → Max re-plans 0–4',
    what: 'Re-read budget per session — deaths re-plan on-chain; budget exhausted = NO-TRADE terminal marker (⛔).',
    trader:
      'The v6-after-cap-4 confusion: the last chip IS the no-trade marker, not a real plan.',
    consumer:
      'store/strategy.go:929 (ReplanCap) · trader/auto_trader_planner.go:241 (re-plan on death)',
    range: '0 – 4 per session',
    systemDefault: '2 (owner)',
    recommended: '⭐ 2 — one re-read after an early death, then sit out.',
    whenToTouch:
      "Raise for violent trend days where one death shouldn't end the session.",
    perSession:
      'Yes — session override wins; inherit (blank) = the strategy-level value (ReplanCapFor).',
  },
  {
    label: 'Acceptance window',
    where: 'Strategy → Day Plan → Acceptance rule',
    what: 'The confirm clock for acceptance-type plays.',
    trader: '5m_close = tight confirm; 15m = patient.',
    consumer: 'trader/auto_trader_planconfig.go:168 (acceptanceFor)',
    range: '5m_close | 15m',
    systemDefault: '5m_close',
    recommended: '⭐ 5m_close — current config.',
    whenToTouch: 'If acceptance plays keep getting cut by the confirm clock.',
    perSession:
      'Yes — session override wins; inherit (blank) = the strategy-level value.',
  },
  {
    label: 'Require approval',
    where: 'Strategy → Day Plan → toggle',
    what: 'ON = entries are HELD until the owner taps Approve for this CME session-day.',
    trader: 'You are the gate: no approve, no entries, even on-plan.',
    consumer:
      'trader/auto_trader_orders.go:297 (approval gate) · api/handler_plan.go:129 (approvalRequired)',
    range: 'ON | OFF',
    systemDefault: 'OFF (fully automatic)',
    recommended: '⭐ OFF — SIM phase; ON is the rehearsal for live.',
    whenToTouch:
      'Turn ON to practice the live approval muscle before going live.',
    perSession: 'Yes.',
  },
  {
    label: 'Evening digest',
    where: 'Strategy → Day Plan → toggle',
    what: "A session-close summary of the day's plan, fills and gate-blocks.",
    trader: 'The daily post-mortem in chat form.',
    consumer: 'trader/auto_trader_planconfig.go:13 (evening_digest)',
    range: 'ON | OFF',
    systemDefault: 'OFF',
    recommended: '⭐ ON if you want the 14:45 wrap-up in the chat.',
    whenToTouch: 'Whenever you want more or less noise.',
    perSession: 'No — strategy-level.',
  },
  {
    label: 'Re-align cap',
    where: 'Strategy → Day Plan → Re-align cap 0–10',
    what: 'Budget of planner re-alignments per session after owner level edits.',
    trader: 'Each Apply merge costs one; decline costs nothing.',
    consumer:
      'api/handler_plan.go:1906 (realign endpoint) · store/strategy.go (RealignCap)',
    range: '0 – 10',
    systemDefault: '5 (owner)',
    recommended: '⭐ 5 — enough for a hands-on day.',
    whenToTouch: 'If you edit levels often, keep it ≥3.',
    perSession: 'Yes.',
  },
  {
    label: 'Min scenario quality',
    where: 'Strategy → Day Plan → A/B/C',
    what: 'Lowest grade the planner may write (INFORMATIONAL — nothing gates on it).',
    trader: 'C = full palette; B/A = the planner filters its own plays.',
    consumer:
      'trader/auto_trader_planner.go:592 (AssembleScoredLevelsMinGrade)',
    range: 'A | B | C',
    systemDefault: 'C',
    recommended: "⭐ C — grade is advisory; don't hide plays with it.",
    whenToTouch: 'Only if the card gets cluttered with junk scenarios.',
    perSession:
      'Yes — session override wins; inherit (blank) = the strategy-level row above.',
  },
  {
    label: '1h anchor seat',
    where: 'Strategy → Day Plan → toggle',
    what: 'Seat 1h/4h anchor levels in the card (the HTF context rows).',
    trader: 'ON = the plan sees the bigger map; the 1h/4h floors (B) apply.',
    consumer: 'trader/auto_trader_dayplan.go:53 (seat_1h_zone)',
    range: 'ON | OFF',
    systemDefault: 'ON (owner)',
    recommended: '⭐ ON — the HTF floors only exist when seated.',
    whenToTouch: 'Turn OFF only for pure 5m/15m scalping studies.',
    perSession: 'Yes.',
  },
  {
    label: 'Wake triggers (5 toggles)',
    where: 'Strategy → Day Plan → Wake triggers',
    what: 'The five event classes that wake the planner mid-session: fresh S/D zones, HTF events, 15m events, invalidation, level-touch waves (the W6 wake wave).',
    trader:
      'More ON = the plan reacts to structure as it forms; wakes are advisory refreshes that can never dark a session.',
    consumer:
      'trader/auto_trader_wake_levels.go:17 (wake wave) · maybeRunSessionReadsAt',
    range: '5 × ON | OFF',
    systemDefault: 'ON (the event-diff wave, 2026-08-25)',
    recommended: '⭐ leave ON — deaths still re-plan; wakes only refine.',
    whenToTouch:
      'Disable a class only if it fires too often for a quiet session.',
    perSession: 'Yes.',
  },
  {
    label: 'Min wake interval',
    where: 'Strategy → Day Plan → Min wake interval',
    what: 'The false-alarm/detection-delay knob between wake reads (minutes).',
    trader: 'Lower = jumpier plan, higher = misses fast structure.',
    consumer: 'trader/auto_trader_wake_levels.go:23 (wake_min_interval_min)',
    range: '5 – 120 minutes',
    systemDefault: '30',
    recommended: '⭐ 30 — current config.',
    whenToTouch: 'Lower for news-heavy days, raise for grind days.',
    perSession: 'Yes.',
  },
]

const risk: KnobSpec[] = [
  {
    label: 'Min confidence',
    where: 'Strategy → Risk Control',
    what: "Floor on the AI's confidence integer; below = refused.",
    trader: 'The simplest honesty gate: 60 means "be at least 60% sure".',
    consumer: 'kernel/engine_position.go:188 (confidence gate)',
    range: '50 – 100',
    systemDefault: '60 (owner)',
    recommended:
      '⭐ 60 — live config; Sep-9 ruling: the 65 raise is DEFERRED, not dead — the 60–64 band gets judged at full n (protection lives in strict + R:R + min-SL + armed meanwhile).',
    whenToTouch: 'Raise if the AI enters low-conviction junk too often.',
    perSession: 'No.',
  },
  {
    label: 'Max positions',
    where: 'Strategy → Risk Control',
    what: 'Max simultaneous open positions.',
    trader: '1 = single position; 3 = diversified.',
    consumer: 'kernel/engine_analysis.go:125 (max_positions)',
    range: '1 – 3',
    systemDefault: '3 (owner)',
    recommended: '⭐ 3 — matches config; MNQ SIM never needs the extra legs.',
    whenToTouch: 'Set 1 for single-position discipline.',
    perSession: 'No.',
  },
  {
    label: 'Leverage BTC/ETH / alt',
    where: 'Strategy → Risk Control',
    what: 'Leverage multiplier per coin class for futures sizing.',
    trader:
      'Code-enforced ceiling is 10/5 (system) even if the page shows up to 20/20.',
    consumer: 'kernel/engine_analysis.go (btcEthLeverage/altcoinLeverage)',
    range: '1 – 20 (page) · code-enforced ≤10 BTC/ETH, ≤5 alt',
    systemDefault: '5 / 5 (owner) · system duality 10/5',
    recommended: '⭐ 5/5 — current config.',
    whenToTouch: 'Lower to derisk; page values above 10/5 are inert.',
    perSession: 'No.',
  },
  {
    label: 'Min risk:reward',
    where: 'Strategy → Risk Control',
    what: 'The R:R floor every entry must clear (computed on real stop/target).',
    trader: 'Raising this is the single strongest filter on bad entries.',
    consumer: 'kernel/engine_position.go:122 (validateDecisions minRiskReward)',
    range: '1 – 10 (step 0.5)',
    systemDefault: '3 (owner)',
    recommended: '⭐ 3 — current config; 4+ measurably cuts entry count.',
    whenToTouch: 'Raise to 4+ if wins are too small to cover losers.',
    perSession: 'No.',
  },
  {
    label: 'Max margin',
    where: 'Strategy → Risk Control',
    what: 'Margin ceiling per position (AI-guided).',
    trader: '90 keeps a single position from eating the account.',
    consumer: 'kernel/engine_analysis.go:529 (riskConfig.MaxMargin)',
    range: 'AI-guided (page) · default 90',
    systemDefault: '90 (owner)',
    recommended: '⭐ 90 — current config.',
    whenToTouch: 'Lower in high-vol regimes.',
    perSession: 'No.',
  },
  {
    label: 'Min position size',
    where: 'Strategy → Risk Control',
    what: 'Smallest position notional/contract count allowed.',
    trader: 'Below 12 the economics of the trade stop making sense.',
    consumer: 'kernel/engine_analysis.go:530 (riskConfig.MinPosition)',
    range: 'page numeric · default 12',
    systemDefault: '12 (owner)',
    recommended: '⭐ 12 — current config.',
    whenToTouch: 'Leave alone in SIM.',
    perSession: 'No.',
  },
  {
    label: 'Hold lock',
    where: 'Strategy → Risk Control',
    what: 'Lock positions against early exit until the hold condition clears.',
    trader: 'Stops you (and the AI) from cutting winners early.',
    consumer: 'kernel/engine_position.go (hold lock path)',
    range: 'ON | OFF',
    systemDefault: 'OFF',
    recommended: '⭐ OFF — the plan already manages exit timing.',
    whenToTouch: "ON if exits keep firing before the plan's own criteria.",
    perSession: 'No.',
  },
  {
    label: 'Breakeven trigger',
    where: 'Strategy → Risk Control',
    what: 'Move the stop to entry after the position gains this much (ticks).',
    trader: '50 ticks = the free-trade tripwire.',
    consumer: 'kernel/engine_position.go (breakeven path)',
    range: 'ticks · default 50',
    systemDefault: '50',
    recommended: '⭐ 50 — current config.',
    whenToTouch: 'Tighten in chop; loosen in trends.',
    perSession: 'No.',
  },
  {
    label: 'Trailing stop',
    where: 'Strategy → Risk Control',
    what: 'ATR-multiplier trail: mult (default 2.0) × ATR(period, default 14), arms after breakeven / N points / immediately.',
    trader: 'Lets runners run while protecting realized gain.',
    consumer: 'kernel/engine_position.go (trailing path)',
    range:
      'mult 0.5–5 · period 7–28 · arm: after_breakeven | N-points | immediately',
    systemDefault: '2.0 / 14 / after_breakeven',
    recommended: '⭐ 2.0/14/after_breakeven — current config.',
    whenToTouch: 'Raise mult for wider runners on trend days.',
    perSession: 'No.',
  },
  {
    label: 'Guardrails master',
    where: 'Strategy → Risk Control → Guardrails',
    what: 'Master switch for the daily guardrails stack (loss/profit caps, max trades, consecutive-loss halt, reentry cooldown, consistency, blackout windows).',
    trader:
      'Currently OFF by owner ruling — the would-have-tripped counters still display.',
    consumer: 'kernel/engine_position.go (guardrail evaluation)',
    range: 'ON | OFF',
    systemDefault: 'ON',
    recommended:
      '⭐ OFF for now (owner ruling) — re-armed after the risk audit is reviewed.',
    whenToTouch: 'ON when you want the daily circuit breakers live.',
    perSession: 'No.',
  },
  {
    label: 'Max contracts (always-on)',
    where: 'Strategy → Risk Control → always-on row',
    what: 'Contract cap per order — on with or without the master.',
    trader: 'The one guardrail you cannot switch off.',
    consumer: 'kernel/engine_position.go (max contracts path)',
    range: 'page value · default 2',
    systemDefault: '2 (owner)',
    recommended: '⭐ 2 — current config.',
    whenToTouch: 'Raise only for deliberate sizing studies.',
    perSession: 'No.',
  },
  {
    label: 'Notional cap (always-on)',
    where: 'Strategy → Risk Control → always-on row',
    what: 'Max notional per position — on with or without the master.',
    trader: 'The second unswitchable guardrail.',
    consumer: 'kernel/engine_position.go (notional cap path)',
    range: 'page value · default 20',
    systemDefault: '20',
    recommended: '⭐ 20 — current config.',
    whenToTouch: 'Raise only for deliberate sizing studies.',
    perSession: 'No.',
  },
  {
    label: 'Position value ratio (BTC/ETH / alt)',
    where: 'Strategy → Risk Control',
    what: 'position_value ≤ equity × ratio — the CODE-ENFORCED sizing ceiling (page values cannot bypass it).',
    trader:
      '5x BTC/ETH and 1x alt = the bot can size up to 5× equity on majors, 1× on alts.',
    consumer:
      'trader/auto_trader_risk.go:229 (enforcePositionValueRatio) · kernel/engine_analysis.go:527',
    range: 'page 1–20 · code-enforced 5 / 1',
    systemDefault: '5 / 1',
    recommended: '⭐ 5/1 — current config.',
    whenToTouch: 'Lower to derisk; values above 5/1 are inert (code ceiling).',
    perSession: 'No.',
  },
  {
    label: 'Daily loss limit',
    where: 'Strategy → Risk Control → Guardrails',
    what: 'Realized PnL ≤ −limit trips the daily-loss halt (force-flat class).',
    trader: 'The circuit breaker that ends a bad day at a known dollar number.',
    consumer:
      'kernel/risk_limits.go:184 (DailyLossLimitUSD) · engine_analysis.go:145',
    range: 'USD · env RISK_MAX_DAILY_LOSS_USD fallback',
    systemDefault: 'ON (with master) · value in Risk Control',
    recommended: '⭐ set it to a loss you can absorb once a week.',
    whenToTouch: 'Set at the start of the week; review after every trip.',
    perSession: 'No.',
  },
  {
    label: 'Daily profit cap',
    where: 'Strategy → Risk Control → Guardrails',
    what: 'Realized PnL ≥ cap stops new entries for the day (lock-in, not close-out).',
    trader: 'Takes the win and stops overtrading a good day.',
    consumer: 'kernel/risk_limits.go:185 (DailyProfitEnabled)',
    range: 'USD · enabled with master',
    systemDefault: 'ON (with master)',
    recommended: '⭐ ON — one of the cheapest edge protections in the stack.',
    whenToTouch: 'Disable only if you deliberately want unlimited upside days.',
    perSession: 'No.',
  },
  {
    label: 'Max daily trades',
    where: 'Strategy → Risk Control → Guardrails',
    what: 'Entry count cap per session-day.',
    trader: "Stops revenge-trading after the day's quota is spent.",
    consumer: 'kernel/risk_limits.go:187 (MaxDailyTradesEnabled)',
    range: 'count · enabled with master',
    systemDefault: 'ON (with master)',
    recommended: "⭐ ON with a number that fits the strategy's hit rate.",
    whenToTouch: 'Review weekly against the win rate.',
    perSession: 'No.',
  },
  {
    label: 'Consecutive-loss halt',
    where: 'Strategy → Risk Control → Guardrails',
    what: 'N consecutive losing closes halt entries until the next session.',
    trader:
      'The streak-breaker: three losers in a row is the market telling you something.',
    consumer:
      'store/position_query.go:57 (CountConsecutiveLossesSince) · telemetry gate-block consecutive_loss',
    range: 'count · enabled with master',
    systemDefault: 'ON (with master)',
    recommended: '⭐ ON, threshold 2–3.',
    whenToTouch: 'Leave ON — this is the cheapest guardrail in the stack.',
    perSession: 'No.',
  },
  {
    label: 'Re-entry cooldown',
    where: 'Strategy → Risk Control → Guardrails',
    what: 'Minimum minutes between a close and the next entry.',
    trader: 'Prevents immediately re-entering after being stopped.',
    consumer: 'kernel/risk_limits.go (guardrail soft set)',
    range: 'minutes · enabled with master',
    systemDefault: 'ON (with master)',
    recommended: '⭐ ON — 5–15 minutes.',
    whenToTouch: "Tune to the strategy's average re-arm time.",
    perSession: 'No.',
  },
  {
    label: 'Consistency',
    where: 'Strategy → Risk Control → Guardrails',
    what: 'Max daily PnL percentage swing before the consistency rule fires.',
    trader: 'Bounds how much one day may deviate from the norm.',
    consumer: 'kernel/risk_limits.go:195 (ConsistencyMaxDayPct)',
    range: 'percent · enabled with master',
    systemDefault: 'ON (with master)',
    recommended: '⭐ ON for a smooth equity curve.',
    whenToTouch: 'Loosen only after a verified regime change.',
    perSession: 'No.',
  },
  {
    label: 'Blackout windows',
    where: 'Strategy → Risk Control → Guardrails',
    what: 'Configured CT time windows (start+end) with zero entries.',
    trader: 'Hard blackout — the bot simply will not trade inside it.',
    consumer: 'kernel/risk_limits.go:193 (BlackoutConfigured / InBlackoutNow)',
    range: 'start+end CT · enabled with master',
    systemDefault: 'OFF',
    recommended:
      '⭐ ON with 12:00–13:30 CT (matches the lunch gate) or your worst hours.',
    whenToTouch: 'Set for your known-bad hours from the journal.',
    perSession: 'No.',
  },
]

const sessions: KnobSpec[] = [
  {
    label: 'Session overrides (ASIA / LONDON / NY)',
    where: 'Strategy → Day Plan → Sessions accordion',
    what: 'Per-session override rows: min grade, min scenario quality, max trades, plan mode, max re-plans, acceptance window. Min grade, quality, max trades and plan mode are tri-state: inherit (blank) = the strategy-level row; an explicit value wins. Stored values that EQUAL the strategy level are auto-migrated to inherit. (min side levels REMOVED — owner ruling 2026-08-31: the per-side count concept is deleted.)',
    trader:
      'The current rows: min_grade B · min_scenario_quality C · max_trades 7/10/10 (ASIA/LONDON/NY) · plan_mode strict ×3 · max re-plans 4 · acceptance 5m_close.',
    consumer:
      'store/strategy.go:921-975 (per-session resolvers) · trader/auto_trader_planconfig.go:158-168',
    range:
      'per-session rows; the four tri-state knobs inherit (blank) = strategy value, explicit = override',
    systemDefault:
      'ASIA 16:55 read 17:00→02:00 · LONDON 01:55 02:00→08:30 · NY 08:25 08:30→14:45 (all EOD-flat)',
    recommended:
      '⭐ keep the current rows — they ARE the deployed session map.',
    whenToTouch: 'Only with a deliberate session-thesis change.',
    perSession: 'N/A (they define it).',
  },
]

export const settings: GuideSection = {
  id: 'settings',
  num: 7,
  title: 'Settings & Knobs',
  tagline:
    'Every knob on the Strategy page, what it really does, and who reads it.',
  asBuiltRev: GUIDE_BUILT_REV,
  blocks: [
    {
      kind: 'p',
      text: 'Every knob card below names the engine consumer (file:line) that reads it — so you always know whether a slider is real or decorative. FE persists but NO production code reads: nothing here is in that category; the three that used to be (plan_mode, proximity_filter_atr, …) are wired now.',
    },
    { kind: 'h', text: 'Day Plan knobs' },
    { kind: 'knobs', knobs: dayPlan },
    { kind: 'h', text: 'Risk Control knobs' },
    { kind: 'knobs', knobs: risk },
    { kind: 'h', text: 'Session map' },
    { kind: 'knobs', knobs: sessions },
    { kind: 'h', text: 'Env-only knobs (not Studio)' },
    {
      kind: 'callout',
      title: 'The 9 env knobs — .env only, never a Studio slider',
      items: [
        {
          title: 'ARM_MIN_RR = 2.0',
          body: 'The gate-at-arm R:R floor for resting orders (the market-entry floor stays 3.0).',
        },
        {
          title: 'HTF_VETO_MODE = cross',
          body: 'Veto mode: 1h | cross | 4h — LIVE = cross (1h AND 4h must agree; the $352/0 autopsy).',
        },
        {
          title: 'HTF_VETO_TF = 1h',
          body: 'The veto timeframe when mode is 1h.',
        },
        {
          title: 'FAST_MARKET_ATR = 1.5',
          body: 'Wake-read fast threshold: |price drift| since the last write > K×ATR5m → fast re-plan.',
        },
        {
          title: 'FAST_MARKET_REASONING = fast',
          body: 'The reasoning wire for fast-market wake reads (FAST TAPE).',
        },
        {
          title: 'BD_MIN_DISP_ATR = 1.0',
          body: 'Breakdown/breakup displacement floor in ATR5m multiples.',
        },
        {
          title: 'FVG_ENTRY_MIN_DISP_ATR = 1.5',
          body: 'FVG displacement floor in ATR5m multiples.',
        },
        {
          title: 'INGEST_QUEUE_CAP = 1024',
          body: 'Bar-ingest queue depth (peak_depth is logged; 0 drops is the invariant).',
        },
        {
          title: 'AI_PLAN_MAX_TOKENS = 65536',
          body: 'Planner completion budget — truncation is a 🚨 WARN, never silent.',
        },
        {
          title: 'PERSIST_STALL_WATCHDOG_S = 60',
          body: 'Bar-persist silence alarm: no successful flush for N seconds while live bar frames are FLOWING → loud ERROR (the Friday ~2h GORM stall can never go silent again). Frame-aware: an idle wire (weekend, the daily break, NT8 closed) stays silent — no cry-wolf.',
        },
      ],
    },
    { kind: 'h', text: 'The save ritual' },
    {
      kind: 'p',
      text: 'Every Strategy-page change must be SAVED to take effect. Ritual: make the change → press Save → "Strategy saved" toast → the `saved {MM/DD, HH:MM} CT` chip updates. Unsaved changes are inert — and the knob-vs-code truth is: a page value above a code ceiling (e.g. leverage 20 vs system 10) saves but does nothing.',
    },
    {
      kind: 'callout',
      title: 'knob-vs-code — the four patterns',
      items: [
        {
          title: 'Wired + clamped',
          body: 'Page value used, code clamps to the system ceiling (leverage 20 → 10/5).',
          cite: 'kernel/engine_analysis.go:125',
        },
        {
          title: 'Wired + per-session',
          body: 'Session override wins over strategy value (plan_mode, proximity, caps).',
          cite: 'store/strategy.go:921-975',
        },
        {
          title: 'Inert without master',
          body: 'Guardrail rows do nothing while the master is OFF — the counters still show would-have-tripped.',
        },
        {
          title: 'Always-on',
          body: 'Max contracts + notional cap ignore the master entirely.',
        },
      ],
    },
  ],
}
