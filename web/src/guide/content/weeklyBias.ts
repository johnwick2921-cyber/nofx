import { GUIDE_BUILT_REV, type GuideSection } from '../types'

// W7 (weekly-bias wave, 2026-08-30) — the Weekly Bias + Planner candles guide
// cards. The wave is WARN/shadow ONLY: nothing here gates, resizes or re-grades
// anything. Promotion to hard rules is a SEP-9 data decision (W8 table).

export const weeklyBias: GuideSection = {
  id: 'weekly-bias',
  num: 13,
  title: 'Weekly Bias & Planner Candles',
  tagline: 'The Sunday read, the weekly chip, and the candle ground-truth law.',
  asBuiltRev: GUIDE_BUILT_REV,
  blocks: [
    { kind: 'h', text: 'What the Sunday read does' },
    {
      kind: 'cards',
      cards: [
        {
          title: 'One read per week (Sunday 16:30 CT)',
          body: 'At WEEKLY_READ_CT (default "sun 16:30" CT) the bot runs ONE AI read over the STORED 1m bars: 12 completed weekly candles, weekly references (weekly_open · PWH/PWL/PWC), the last 5 weekend gaps (NWOG), the 20/40/60-day IPDA ranges, and a prior-week recap. The doc is stored on a plans row with session=WEEKLY; a stored doc means never re-run (idempotent). A Monday boot backfills exactly once.',
        },
        {
          title: 'Tier-A evidence only — closed bars only',
          body: 'The bias may cite ONLY: (a) price vs weekly_open with acceptance, (b) PWH/PWL break-AND-HOLD vs sweep-and-reject, (c) the 3-week structure tags (HH/outside/LL/LH/inside). NWOG and IPDA are draw/target material only — citing them as bias evidence is REJECTED by the validator. The validator also rejects: bad enums, a missing invalidation, a draw that matches no computed reference (±1 tick), a >3-line narrative, day-of-week tokens, and non-low conviction on thin history.',
        },
        {
          title: 'Day-of-week reasoning is BANNED',
          body: 'The folklore law is hard inside the weekly read: any weekday token in the narrative rejects the doc (validator r5). "Monday seasonality" can never be a bias reason.',
        },
        {
          title: 'Invalidation semantics',
          body: 'Every weekly doc carries a MANDATORY invalidation price + basis ("1h close beyond <px>"). When a CLOSED bar of the basis TF crosses it mid-week, the doc flips bias→neutral with invalidated_at stamped — NEVER an auto-flip to the opposite side, and no re-read until next Sunday. The chip renders "WEEKLY neutral" (tooltip carries the invalidated date) — an invalidated weekly is a VALID neutral state, never struck through (owner ruling 2026-08-31).',
        },
        {
          title: 'WARN mode annotates — it never blocks',
          body: 'The whole wave is SHADOW: counter-trend entries log ⚖️ WEEKLY-COUNTER clauses (would-halve-size · would-require-A-grade · would-need-RR≥4.0), weekly-confluent levels log 🌗 SHADOW grades, decision rows get a draw-alignment tag. Seating, grades, gates and sizes change ZERO (THE LAW — the fixtures prove it). Promotion to hard rules is a SEP-9 data decision, not this wave.',
        },
      ],
    },
    { kind: 'h', text: 'The six knobs (env only)' },
    {
      kind: 'knobs',
      knobs: [
        {
          label: 'WEEKLY_READ_CT',
          where: 'environment',
          what: 'When the Sunday weekly read fires ("<dow> HH:MM" CT).',
          trader: 'Trader: sets the read time of the week.',
          consumer: 'kernel/weekly_knobs.go WeeklyReadSpec',
          range: '"sun 16:30" (day 3-letter + HH:MM CT)',
          systemDefault: 'sun 16:30',
          recommended:
            '⭐ keep the default — 16:30 CT sits before the 17:00 CT Sunday open.',
          whenToTouch: 'Never in normal operation.',
          perSession: 'no — process-wide env',
        },
        {
          label: 'WEEKLY_CONFLUENCE_BAND_ATR',
          where: 'environment',
          what: 'The W5.1 shadow confluence band width in ATR5m units.',
          trader: 'Trader: wider band = more 🌗 SHADOW confluence logs.',
          consumer: 'kernel/weekly_knobs.go WeeklyConfluenceBandATR',
          range: '> 0 (float)',
          systemDefault: '0.25',
          recommended:
            '⭐ 0.25 — the researched default; widen only for the Sep-9 study.',
          whenToTouch: 'Only for the promotion study.',
          perSession: 'no',
        },
        {
          label: 'WEEKLY_SHADOW_MULT',
          where: 'environment',
          what: 'The shadow grade multiplier for weekly-confluent levels.',
          trader: 'Trader: changes only the SHADOW reorder counter.',
          consumer: 'kernel/weekly_knobs.go WeeklyShadowMult',
          range: '> 0 (float)',
          systemDefault: '1.5',
          recommended: '⭐ 1.5 — shadow only; the real grade is untouched.',
          whenToTouch: 'Only for the promotion study.',
          perSession: 'no',
        },
        {
          label: 'WEEKLY_COUNTER_MODE',
          where: 'environment',
          what: 'warn (default) logs ⚖️ WEEKLY-COUNTER annotations; off silences them.',
          trader:
            'Trader: warn keeps the journal honest; off hides the counter lines.',
          consumer: 'kernel/weekly_knobs.go WeeklyCounterMode',
          range: 'warn | off',
          systemDefault: 'warn',
          recommended:
            '⭐ warn — the annotations are the Sep-9 counter evidence.',
          whenToTouch: 'Set off only to quiet the journal.',
          perSession: 'no',
        },
        {
          label: 'WEEKLY_INVALIDATION_TF_DEFAULT',
          where: 'environment',
          what: "The invalidation watch TF when a stored doc's basis has no parseable timeframe.",
          trader: 'Trader: which closed-bar series the W4 watch evaluates.',
          consumer: 'kernel/weekly_knobs.go WeeklyInvalidationTFDefault',
          range: 'any bar interval string ("1h", "15m"…)',
          systemDefault: '1h',
          recommended: '⭐ 1h — matches the spec basis contract.',
          whenToTouch: 'Never in normal operation.',
          perSession: 'no',
        },
        {
          label: 'PLANNER_CANDLES',
          where: 'environment',
          what: 'on (default) renders the ## Candles block (12×15m · 12×1h · 8×4h · 8×daily) in every session planner prompt.',
          trader: 'Trader: the planner gets raw candle eyes.',
          consumer: 'kernel/weekly_knobs.go PlannerCandlesEnabled',
          range: 'on | off',
          systemDefault: 'on',
          recommended: '⭐ on — candles are ground truth for structure.',
          whenToTouch: 'Off only if the token budget ever matters.',
          perSession: 'no',
        },
      ],
    },
    { kind: 'h', text: 'Planner candles' },
    {
      kind: 'cards',
      cards: [
        {
          title: 'Candles are ground truth',
          body: 'The session planner prompt now carries the raw candle tables (last 12×15m, 12×1h, 8×4h, 8×daily) rendered by the SAME formatter the executor uses. The playbook line is explicit: "Candles are ground truth for structure; ranked levels and tags are summaries. On conflict, trust the candles and say so in the scenario rationale." Scenario lines citing "per candles" are counted into the planner_candle_citations telemetry counter — the P3 promotion evidence.',
        },
        {
          title: 'Weekly context (soft law)',
          body: 'Every session prompt also carries a ≤3-line ## Weekly Context block — "WEEKLY: bull/high · draw PWH 30500.25 · invalid 30300.00 (1h close beyond 30300.00)" or "WEEKLY: none" when no doc exists (fail-open: a missing doc changes nothing else). Counter-weekly scenarios are allowed but must state their justification (HTF level / sweep-reclaim of the draw); target chains toward the draw are preferred. The executor prompt gets one matching context line. The weekly bias never gates your plan.',
        },
      ],
    },
  ],
}
