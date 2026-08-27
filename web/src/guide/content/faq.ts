import { GUIDE_BUILT_REV, type GuideSection } from '../types'

export const faq: GuideSection = {
  id: 'faq',
  num: 12,
  title: 'FAQ',
  tagline: 'The twelve questions everyone asks.',
  asBuiltRev: GUIDE_BUILT_REV,
  blocks: [
    {
      kind: 'faq',
      items: [
        {
          q: 'Why did the bot sit out all session?',
          a: 'Look at the card: a ⛔ fail-closed banner (read failed — safe, never stale) or a no_trade declaration (the AI decided to skip — balance day, no A/B zone). Then check the gate-block ledger for the refusing label.',
          mechanism:
            'Read attempts WARN in the log: "📐 planner attempt N/3 …".',
          link: '#plan-card',
        },
        {
          q: 'Why was my trade refused?',
          a: 'Every refusal is a named string — confidence too low, against the plan, stop below MIN_SL, past last-entry, awaiting approval… The refusal decoder table maps each to its meaning and knob.',
          mechanism:
            'kernel/engine_position.go · trader/auto_trader_planconfig.go',
          link: '#guards',
        },
        {
          q: 'Is any of this real money?',
          a: 'No. Every path is NinjaTrader SIM. isAccountTradeable blocks non-SIM accounts at routing, and the SIM lock must not be weakened.',
          link: '#guards',
        },
        {
          q: 'What does the ⛔ NO-TRADE chip mean?',
          a: 'The re-plan budget is exhausted: the last version row is the terminal marker, not a real plan. Reset (↺) re-arms the budget and reads fresh.',
          mechanism:
            'replan_cap=4 legitimately ends at a row labelled v6 (marker).',
          link: '#plan-card',
        },
        {
          q: 'What do the level grades mean?',
          a: "A/B/C = the planner's confidence in a level (evidence × freshness × confluence × TF). m: is the machine grade. Both are INFORMATIONAL — nothing gates on them.",
          link: '#levels',
        },
        {
          q: 'Do I need to approve plans?',
          a: 'Only if approval_required is ON. Then entries are HELD until you tap Approve for that CME session-day. Default is OFF (fully automatic).',
          link: '#plan-card',
        },
        {
          q: 'Why does the card say ⚖ thin side?',
          a: 'The assembled in-band map itself had fewer than min_side_levels on a side — a machine-caused shortage, so the plan writes with a warn instead of failing. Zero sides still fail-closed.',
          link: '#plan-card',
        },
        {
          q: 'What is the A-setup?',
          a: 'sweep (liquidity taken) → displacement → FVG retrace, chained as S1 sweep_reclaim → S2 fvg_entry. A bare FVG with no sweep precursor gets a WARN at write — it has no standalone edge.',
          link: '#plays',
        },
        {
          q: 'Why did the bot flatten at 14:45?',
          a: 'EOD-flat ladder: the session ends flat by design (limit ticks, then market after the grace seconds). The session registry enforces it for all three sessions.',
          link: '#trading-day',
        },
        {
          q: 'NT8 is down — what do I do?',
          a: 'Nothing on the bot first: no bars means no decisions, and the gates (feed down / dead-man) say so. Fix NT8 + the AddOn connection (copy → F5 compile → full restart), then verify the TCP reconnect.',
          link: '#routines',
        },
        {
          q: "A knob change didn't do anything — why?",
          a: "Either it wasn't saved (toast + saved-chip), the value is above a code ceiling (leverage 20 → inert above 10/5), the guardrail master is OFF, or the config is cached at trader-load and needs the trader reload/restart to go hot.",
          link: '#settings',
        },
        {
          q: 'Why does the guide banner warn about revision drift?',
          a: "This guide was built against one code revision. The banner compares it to the running bot's revision — amber means the guide is older than the bot; verify before trusting a cite.",
          link: '#welcome',
        },
      ],
    },
  ],
}
