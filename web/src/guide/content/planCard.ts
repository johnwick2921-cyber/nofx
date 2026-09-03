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
          body: 'Version chips v1…vN (tap any to read it as it was) + the lifecycle chip (ACTIVE gold / EXPIRED / DIED / SUPERSEDED / NO-TRADE red) + why this version was written (owner_reset, level_event, NY_scheduled_read, death_replan, owner_reread…). Only death_replan and owner_reread spend the re-plan budget (class 35 recorded counter).',
          cite: 'web/src/components/plan/chips.tsx:169,252',
        },
        {
          title: '4 · Level-row anatomy',
          body: 'price · provenance label (PDH, ONH, nPOC·Tue, RN, EQH…) · planner grade A/B/C · m: machine grade (detector-side: type × freshness × confluence × HTF) · ROLE badge (what the level is FOR) · distance (gold when within 12pt) · touch chip ○ approaching ◐ touching ✕ rejected ▲ accepted · fresh dot (fresh / tested / consumed — consumed rows dim).',
          cite: 'web/src/components/plan/ZoneTable.tsx:28-150 · SessionPlanCard.tsx:682',
        },
        {
          title: '5 · Scenario-row anatomy',
          body: 'S# · condition (reclaim/hold/sweep_reclaim/reject/acceptance/breakout_retest/fvg_entry/breakdown_continue/breakup_continue) · direction · quality A+/A/B/C (INFORMATIONAL — nothing gates on it) · confirm{} chip CONFIRM MET / not met (machine, advisory; stale ones say so) · TWO-LEG confirms (breakdown/breakup plays) render leg-by-leg: "leg 1/2 MET · leg 2/2 NOT MET → overall NOT MET" — a partial never reads MET · fvg chip IN-ZONE/ABOVE/BELOW/FILLED_INVALID · chain_after: the S# this play FOLLOWS (e.g. fvg_entry after its sweep_reclaim) · targets a→b→c · invalid line.',
          cite: 'web/src/components/plan/ScenarioList.tsx',
        },
        {
          title: '6 · No-trade windows (the band)',
          body: "The MACHINE's sit-out list, not the plan's prose. Three sources, all of them enforcing: the first-N-minutes and lunch bands come from the same definitions the entry gate refuses on and the adherence grader scores against, and the red-news blackouts are literally the windows the gate will refuse inside (clock widening and the fail-closed static fallback included). Every window is stamped against YOUR clock when you open the card: live ones are the rules, spent and other-session ones collapse behind a count. The model's own no_trade prose is COLLAPSED behind a 'Model notes' toggle (owner ruling 2026-09-03) — it is a note, never a rule, and it does not render until you ask for it. It was showing the machine's own windows back at you: ASIA v14 carried 'first 5m (CT)' and '12:00-13:30 CT lunch' because the prompt's JSON example demonstrated exactly those. The example is a placeholder now ('<your own sit-out conditions, or omit>'), so the field should carry the model's own reasons or nothing — and it renders nowhere by default either way. Before 2026-09-02 the card rendered that prose verbatim, so an ASIA card at 23:00 CT listed three constraints (first 5m spent six hours earlier, NY lunch, a 09:00 blackout) and not one of them could refuse an entry. A plan written before the band has no machine windows and still renders its prose as rules.",
          cite: 'web/src/components/plan/RulesBlock.tsx · kernel/no_trade_band.go · boot line 🗓 no-trade band',
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
          title: '9 · level table rules',
          body: 'Any level count is fine — the per-side minimum is DELETED (owner ruling 2026-08-31, no ⚖ note anymore). The only hard fails left: 0 levels on a side (the 2026-08-18 one-sided-map pathology) and an empty machine map.',
          cite: 'kernel/plan_doc.go ValidatePlanDocWithFactsMachine',
        },
        {
          title: '10 · Armed chips',
          body: '⏳ armed = resting order placed, waiting on its confirm (wait_confirm) · 📌 working = resting at the broker · ⚡ filled = entry taken (the real fill) · ✕ cancelled/expired/refused. The arm is the fast path: it pre-commits the entry so the fill happens at the plan price, not after a 2-minute debate.',
          cite: 'web/src/components/plan/ScenarioList.tsx:218-251',
        },
        {
          title: '11 · 😴 dormant + auto-rearm',
          body: 'Dormant = the plan (or its arm) was parked by a flip/death or no-active-plan — NOT dead. It auto-rearms when price closes back through the mirror buffer (0.5×ATR14, 2 decision-TF closes) and arms re-place on the next cycle.',
          cite: 'kernel/plan_lifecycle.go (dormant + rearm) · trader/armed_executor.go',
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
            'Levels — any consumed (dim) rows?',
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
