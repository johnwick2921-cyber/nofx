import { GUIDE_BUILT_REV, type GuideSection } from '../types'

export const welcome: GuideSection = {
  id: 'welcome',
  num: 1,
  title: 'Welcome',
  tagline: 'What this thing is, in one screen.',
  asBuiltRev: GUIDE_BUILT_REV,
  blocks: [
    {
      kind: 'cards',
      cards: [
        {
          title: 'NOFX / VL',
          body: 'One structure-permissioned, level-anchored MNQ SIM trader — advisory-first. The Go engine computes the map; the AI (deepseek-v4-pro) writes one day-plan per session, then decides every cycle against it. You hold the knobs.',
          tag: 'SIM',
        },
        {
          title: 'Three principles',
          body: 'Advisory — most machine output informs, it does not block. Evidence — stats are cited or marked [UNVERIFIED]. Self-scoring — the system grades its own discipline (GPA), levels (level_stats) and refusals (gate-blocks).',
        },
        {
          title: 'Your role',
          body: 'Session windows, plan_mode, budgets and guardrails are yours. Destructive buttons ask first. Everything else runs itself. See Section 9 for the daily routines.',
        },
      ],
    },
    { kind: 'h', text: 'The stack' },
    {
      kind: 'code',
      title: 'tap a layer to see its job',
      lines: [
        'NT8 (Tradovate MNQ bars, Sim101)   ← SIM only, never live',
        '  │ TCP (bars / account / positions / orders)',
        'C# AddOn (NinjaTrader AddOns folder)  ← compile + NT8 restart to change',
        '  │ framed messages (ninjascript/vltrader_tcp_PROTOCOL.md)',
        'Go bot (nofx-bin) ─── BarCache + bars table (SQLite data/data.db = "memory")',
        '  ├─ PLANNER: DeepSeek → 1 day-plan per session (advisory JSON)',
        '  ├─ EXECUTOR: every ~2 min → DeepSeek decision → risk gates → Sim101',
        '  └─ WATCHER: in-position advisory only (zero order authority)',
      ],
    },
    {
      kind: 'p',
      text: 'Boot truth: the bot refuses to trade unless its revision and prompt goldens pass — `🔐 BOOT INTEGRITY OK — rev … · goldens PASS` (kernel/boot_integrity.go). The SIM lock is hard: non-SIM accounts are refused at the account-routing layer (provider/ninjatrader isAccountTradeable).',
    },
    { kind: 'h', text: 'Quick start' },
    {
      kind: 'p',
      text: 'Three jumps: read the plan card (Section 3) · learn what levels are (Section 4) · understand why it sits out (Section 6 + FAQ "ASIA sat out all night").',
    },
  ],
}
