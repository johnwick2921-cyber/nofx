import { GUIDE_BUILT_REV, type GuideSection } from '../types'

export const planCard: GuideSection = {
  id: 'plan-card',
  num: 3,
  title: 'Reading the Plan Card',
  tagline: 'The centerpiece — every element, decoded.',
  asBuiltRev: GUIDE_BUILT_REV,
  blocks: [
    {
      kind: 'p',
      text: 'This is the card the planner writes and the executor trades against. Below is a mock built from the REAL card components (dashed border = example). Every callout maps to a piece of it.',
    },
    { kind: 'mockCard' },
    {
      kind: 'callout',
      title: 'tap to expand — every element',
      items: [
        {
          title: '1 · Bias + conviction',
          body: "The big word (LONG/SHORT/NEUTRAL) + conviction (high/medium/low). Direction and conviction are the AI's read — nothing gates on them.",
          cite: 'web/src/components/plan/BiasBlock.tsx:8',
        },
        {
          title: '2 · The bias-tree line',
          body: 'The planner must NAME the machine branch it took. Branches: 1 close>PDH → bull-continuation HIGH · 2 PDH sweep+reclaim → bear MEDIUM (mirror PDL) · 3 inside the day → close vs PDC, LOW · 4 closed outside but back inside → NO bias · 5 premium/discount vs the dealing-range midpoint · 6 runner = the draw (nearest opposing pool). Out-of-range prices render "BEYOND range (extended)" — never a >100% figure.',
          cite: 'kernel/planner_prompt.go RenderBiasTree',
        },
        {
          title: '3 · v# + trigger pill',
          body: 'Version chips v1…vN (tap any to read it as it was) + the lifecycle chip (ACTIVE gold / EXPIRED / DIED / SUPERSEDED / NO-TRADE red) + why this version was written (owner_reset, level_event, NY_scheduled_read…).',
          cite: 'web/src/components/plan/chips.tsx:169,252',
        },
        {
          title: '4 · Level-row anatomy',
          body: 'price · provenance label (PDH, ONH, nPOC·Tue, RN, EQH…) · planner grade A/B/C · m: machine grade (detector-side: type × freshness × confluence × HTF) · ROLE badge (what the level is FOR) · distance (gold when within 12pt) · touch chip ○ approaching ◐ touching ✕ rejected ▲ accepted · fresh dot (fresh / tested / consumed — consumed rows dim) · ⚖ thin-side note when the machine map itself was short on a side.',
          cite: 'web/src/components/plan/ZoneTable.tsx:28-150 · SessionPlanCard.tsx:682',
        },
        {
          title: '5 · Scenario-row anatomy',
          body: 'S# · condition (reclaim/hold/sweep_reclaim/reject/acceptance/breakout_retest/fvg_entry) · direction · quality A+/A/B/C (INFORMATIONAL — nothing gates on it) · confirm{} chip CONFIRM MET / not met (machine, advisory; stale ones say so) · fvg chip IN-ZONE/ABOVE/BELOW/FILLED_INVALID · chain_after: the S# this play FOLLOWS (e.g. fvg_entry after its sweep_reclaim) · targets a→b→c · invalid line.',
          cite: 'web/src/components/plan/ScenarioList.tsx',
        },
        {
          title: '6 · No-trade windows',
          body: "The plan's own sit-out list: first 5m, lunch 12:00–13:30 CT, T1 red-news blackouts. The card shows them; the executor honors them as plan discipline.",
          cite: 'web/src/components/plan/RulesBlock.tsx:38',
        },
        {
          title: '7 · Death line + flip line',
          body: 'Plan dies if … (structured death{} object, machine-evaluated every cycle) and Flips … (flip_to direction). A prose-only death gets a "PROSE-ONLY" warn at write.',
          cite: 'web/src/components/plan/BiasBlock.tsx · kernel/plan_doc.go PlanCondition',
        },
        {
          title: '8 · NO-TRADE banners — the two variants',
          body: 'Variant A (fail-closed): "⛔ Plan read failed — sitting out — {reason}" — the read failed after 3 attempts; safe, never stale. Variant B (AI skip-day): the AI\'s own no_trade declaration ("balance day — no A/B zone in reach, skip") — a decision, not a failure.',
          cite: 'web/src/components/plan/SessionPlanCard.tsx:474,510',
        },
        {
          title: '9 · ⚖ thin-side note',
          body: 'Written when the assembled in-band map ITSELF had fewer than min_side_levels on a side — machine-caused, so the plan writes with a warn instead of failing. Not an error, not an AI mistake.',
          cite: 'kernel/plan_doc.go SideQuotaNote · SessionPlanCard.tsx:682',
        },
      ],
    },
    { kind: 'h', text: 'Read it in 30 seconds' },
    {
      kind: 'checklists',
      items: [
        {
          title: 'The 6-glance read',
          steps: [
            'Bias word + conviction — what is it trying to do?',
            'Bias-tree line — which branch, and does the reasoning agree?',
            'NO-TRADE banner? — fail-closed vs skip-day changes everything.',
            'Levels — any consumed (dim) or thin-side ⚖ rows?',
            'Scenarios — which one is armed, and does confirm say MET?',
            'Death/flip — what kills the plan, and at what price?',
          ],
        },
      ],
    },
    { kind: 'h', text: 'The four buttons' },
    {
      kind: 'buttons',
      items: [
        {
          label: '↺ Reset planner',
          api: 'POST /api/plan/reset (api/handler_plan.go:1049)',
          sideEffects:
            'Abandons the chain (history + death reasons preserved), re-arms the full re-plan budget, clears NO-TRADE, reads a fresh plan now. Positions and brackets never touched.',
          budget: 'New chain starts at v1 with the full cap (default 4)',
          undo: 'None — but history stays readable, and nothing is deleted.',
          useWhen:
            'The plan is wrong, stale, or a fail-closed NO-TRADE sits where a tradeable day exists. Confirm text: "Abandon this plan chain and start fresh?"',
        },
        {
          label: '⟳ Re-read',
          api: 'POST /api/plan/reread (api/handler_plan.go:1001)',
          sideEffects:
            'One more planner call → a new version on the SAME chain.',
          budget:
            'Costs 1 re-read from the session budget (shows "spend one of N?")',
          undo: 'Old versions stay tappable via the version chips.',
          useWhen: 'You want a second opinion without abandoning the chain.',
        },
        {
          label: '⟳ Re-align plan',
          api: 'POST /api/plan/realign (api/handler_plan.go:1906)',
          sideEffects:
            'Planner reviews your edit and proposes a merged plan change ("would become v{n}") — you Apply merge or Keep as-is.',
          budget: 'Consumes the re-align budget (realign_cap, default 5)',
          undo: 'Keep as-is declines; applied merges are versions (tappable).',
          useWhen:
            'After an owner level edit you want the planner to re-anchor around.',
        },
        {
          label: 'Approve',
          api: 'POST /api/plan/approve (api/handler_plan.go:1799)',
          sideEffects:
            'Grants entries for this CME session-day when approval_required is ON. One click, no modal (by design).',
          budget: 'None — one grant per session-day.',
          undo: 'None needed — it only unlocks the gate the strategy asked for.',
          useWhen:
            'The strategy has approval_required ON and you accept the plan as-is.',
        },
      ],
    },
  ],
}
