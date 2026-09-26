# FIX-PLANNER item 5 — offline re-judge of the 14 killed chains (zero API)

- **Lane:** DS-101 · branch `fix/planner-killers` · base `origin/dev` `04ae1c2f`
- **Specs built from (L3):** CTO dispatch 2026-09-26 04:59Z (id 1790398758580-4857-000001, scope-amended 05:06Z); census `docs/superpowers/reports/2026-09-26-plan-death-census.md` (6bbb3a6c, accepted in principle).
- **Method:** read-only on the DB COPY (`/home/hoang/nofx-ds-101-plandeath/ab.db`, 09-25 22:39 CT). No API calls, no writes, no live system. The salvage rule is judged from the PERSISTED `born_check` records the refusal events themselves wrote (`system_config plan_liveness_event:born_dead_refusal:*`), the same verdicts `validateAuthoredScenariosAt` stored at each refusal — never reconstructed.

## Born-dead kills (3) — would the salvage publish them?

Salvage rule applied: publish iff the final refusal had ≥1 scenario verdict `alive`/`tape_unknown`, zero grammar refusals, and `death`/`flip` not met.

| chain ids | day/session | final attempt event (born_check) | verdicts | salvage |
|---|---|---|---|---|
| 363-365 | 09-24 NY | 1790261940785248219 @09:59:00.785 CT (row 365 @14:59:00Z) | S1 alive · S2 alive · S3 invalidated · death not_met · flip not_met | **WOULD PUBLISH** (S1+S2 kept) |
| 397-399 | 09-25 LONDON | 1790332615980638386 @05:36:55.980 CT (row 399 @10:36:55Z) | S1 alive · S2 invalidated · death not_met · flip not_met | **WOULD PUBLISH** (S1 kept) |
| 409-411 | 09-25 NY | 1790345154017862744 @09:05:54.018 CT (row 411 @14:05:54Z) | S1 alive · S2 invalidated · death not_met · flip not_met | **WOULD PUBLISH** (S1 kept) |

**3 of 14 killed chains would now publish.** Each also saves its last attempt's
wall via the one-fresh-tape-repair cap (no attempt 3 burned).

## The other 11 kills — not measurable zero-call (stated, not papered over)

| class | kills | why the re-judge cannot credit them |
|---|---|---|
| obstacle_chain | 349-351, 352-354, 385-387 | prompt-side (repair listing, item 2). The unchanged validator still refuses the SAVED text; the fix changes what the model is next told. A zero-call replay cannot run the model. |
| entry_policy_shape | 360-362, 394-396, 412-414 | prompt-side (shape table, item 3). Same reasoning. |
| flip/bias mandatory | 339-341, 376-378 | unchanged by this wave — not in scope. |
| tape verify (breakdown) | 391-393, 357-359 | unchanged by this wave — not in scope. |
| gap trigger | 400-402 | unchanged by this wave (the gap-trigger prompt line was in the queued scope but the CTO's amended scope dropped it) — not in scope. |

## Downstream-gate caveat

A salvaged candidate then runs the SAME downstream validators (write truth,
feasibility). Those were never reached by these saved reads, and the saved
`facts` (400 B) do not carry the frozen identity map / zones, so a fuller
offline replay is not possible from the row alone (only row 411 has a saved
`response_text`, 5,308 B). The claim above is therefore scoped to the born
gate exactly as persisted.

**Net for the owner:** 3/14 killed chains publish under item 1; items 2+3 are
prompt-side and their effect can only be measured on live reads (or with an
approved API spend re-running the saved outputs through the new prompts).
