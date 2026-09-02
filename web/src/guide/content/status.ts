import { GUIDE_BUILT_REV, type GuideSection } from '../types'

export const status: GuideSection = {
  id: 'status',
  num: 10,
  title: 'Status & Signals',
  tagline: 'Every indicator strip, banner, and log line — decoded.',
  asBuiltRev: GUIDE_BUILT_REV,
  blocks: [
    { kind: 'h', text: 'The boot ledger, line by line' },
    {
      kind: 'code',
      title: 'the lines printed at startup, in order',
      lines: [
        '🔐 BOOT INTEGRITY OK — rev <sha> [+dirty] · built <ts>',
        '    · expected <sha> · goldens PASS      ← code matches deploy record',
        '📜 planner playbook: playbook=v2 bias_tree=on …',
        '🛡 plan facts guards: 0-side + empty map fail-closed …',
        '🚀 planner speed wave: retry=repair stream=on stream_idle=30s stream_total=1200s …',
        '🛰 planner client: provider_row=<ai_models id> stream_idle=30s stream_total=1200s http_ceiling=600s …  ← class 37 (per trader)',
        '🧪 validator hints: N sites — condition tokens legal+live, rule tokens in-field',
        '📜 prompt/validator contract: N restrictions, all stated in prompt  ← class 38',
        '🎛 volume wave …   ← wave detector knobs',
        '🎯 touch telemetry …',
        '📐 fvg_entry …',
        '🔧 S-wave …',
        '',
        '+dirty usually = an untracked file (.env.bak…) — Go vcs.modified',
        'counts untracked files. NOT a code change.',
      ],
    },
    { kind: 'h', text: 'SYSTEM_STATUS strip (dashboard)' },
    {
      kind: 'table',
      head: ['Item', 'Meaning'],
      rows: [
        [
          'NT8 feed',
          'TCP bridge alive + bars flowing — the single source of truth.',
        ],
        ['Boot integrity', 'Running binary == deploy record (goldens PASS).'],
        ['Dead-man watchdog', 'Kernel heartbeats ok.'],
        ['Trader frozen', 'The trader loop is stuck — investigate.'],
        ['Clock drift', 'Host clock vs NT8 clock mismatch.'],
        [
          '402 banner',
          'An upstream model API returned HTTP 402 (billing) — the model is down for payment, not code.',
        ],
      ],
    },
    { kind: 'h', text: 'Gate-block labels — the full list' },
    {
      kind: 'code',
      title: 'every label the gate panel can show',
      lines: [
        'NT8 feed down · Dead-man watchdog · Trader frozen · Boot integrity',
        'Consecutive-loss halt · Past last-entry time · Outside session window',
        'Against the plan · Awaiting approval · Clock drift',
        'Duplicate order dropped · Order rate breaker · Burned level re-touched',
        'Night/day transition',
        '',
        'Reset: at the 17:00 session roll, and on bot restart.',
      ],
    },
    { kind: 'h', text: 'Traffic light — one glance' },
    {
      kind: 'p',
      text: "GREEN = bot running, gates quiet, plan armed (or flat by plan). AMBER = gates firing repeatedly — read the ledger, don't override. RED = feed down / frozen / boot mismatch — use the emergency checklist (Section 9). The card's NO-TRADE banner is not a light: it is a state.",
    },
  ],
}
