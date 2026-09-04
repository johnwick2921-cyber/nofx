# DEMOTION-QUEUE RULINGS — owner, 2026-09-04

Nine queue items, ruled one by one. Code changes ride `fix/demotion-rulings-0904`
(same wave, next boot). **KEEP = no code change, label recorded.**

| # | Item | Ruling | Code change |
|---|---|---|---|
| 1 | min-SL 1.5× | **KEEP REJECT.** Grounded now. Revisit only when winners' MAE reaches n≥30 (today p80 22.5 vs floor 33, n=18) | none |
| 2 | swing seats | **KEEP SEATED.** 80→65% is the seats working, `[T]`-positive | none |
| 3 | level-event wakes | **KEEP full replan.** The change-based predicate (stage 4) is the fix, not WARN-first | none |
| 4 | 4a BD_MAX_PULLBACK · 4b confirm closes | 4a **DELETE the dead code.** 4b **KEEP ≥1 confirming close**; the displacement feed-forward covers the zero-displacement case | 4a removed |
| 5 | stale-confirm 2.0×ATR | **KEEP the arm gate** (`[T]`, n=2908) | none |
| 6 | grader ladders | **KEEP unchanged**; the sweep waits for n=200 per kind from touch_outcomes. No retune from a replay that said "nothing changes" | none |
| 7 | ASIA/LONDON | **KEEP ENABLED in SIM**; the instrument needs those sessions to count; rule at n≥30 per session in the 1D table | none |
| 8 | consumed/3rd-touch → target_only | **KEEP**; measure from touch_outcomes per ordinal; first real verdict at n=200 per ordinal | none |
| 9 | killzone / Mon-Thu conviction | **DELETE the prompt lines AND the outside-killzone adherence step-down.** No premium on 567 fills; killzone becomes a label the card shows, never a rule or a grade. Counter kept | both removed |

**Second-order:** the sub-floor authoring is the 1.0-vs-1.5 prompt drift — boxed
for this same code lane (the planner prompt still instructed `≥1.0×ATR5m` while
the gate is 1.5). Fixed in the same wave.

## What changed in code (`fix/demotion-rulings-0904`)

- `kernel/breakdown_continue.go` — deleted dead `bdMaxPullbackFrac()` +
  `BD_MAX_PULLBACK` env knob and the unread `PlanBreakdownContinue.Pullback`
  field (zero callers/readers, `[A]`); 4b unchanged (`BD_MIN_CLOSES=1` stays).
- `kernel/planner_prompt.go` — deleted the "Killzone weighting (advisory)"
  block (NY-AM primary / premium FVG / Monday-Thursday conviction); fixed the
  drift: `≥1.0×ATR5m` → `≥1.5×ATR5m` in the arm feasibility contract and the
  waterfall-play gate-chain text (matches `MinSLATRMultDefault=1.5`).
- `kernel/adherence.go` — removed the outside-killzone grade step-down;
  `InKillzone` stays as an input and the card keeps showing the killzone label.
- Tests: adherence matrix re-pinned (killzone no longer grades), playbook
  prompt-contract assertions dropped for the deleted lines; full suite green;
  boot goldens unchanged (the changed prompt is not part of the embedded
  futures fixtures, `[A]` verified).
- SYSTEM-MAP (§3, §4) updated in the same commit (class 75 contract).
