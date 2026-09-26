# RESEARCH-PLAN-DEATH — census of why planner reads fail

- **Dispatch:** 2026-09-26 04:44Z (owner GO 23:4x CT "send idle agents out"). Lane DS-101.
- **Branch:** `research/plan-death-census` (claim `092e2379`), base `origin/dev` `04ae1c2f` (`git log -1 --format=%h` = `04ae1c2f`).
- **Read-only:** no code changes, no API calls, nothing merged. DB COPY only (`ab.db` from 09-25 22:39 CT). Live logs read-only under `/home/hoang/nofx/data/`.
- **Specs I built from (L3):** this dispatch message; `trader/auto_trader_planner.go` (`runPlannerReadCoreObserved` @ :2052 at `04ae1c2f`), `store/planner_rejected.go` (:71), `trader/plan_liveness.go` (:30), `kernel/planner_prompt.go` (:836, :866), `kernel/plan_authored_invalidation.go` (:130-151), `kernel/validator_hints.go` (:72), `trader/structural_geometry.go`, `kernel/plan_doc.go` (:1231-1234).

## 1. The corpus and the buckets

`planner_rejected_prompts` since 2026-09-10, ALL attempts: **200 rows** (the store cap of 200 — trimmed, so pre-09-16 rejects are under-represented [A]). Session split: ASIA 85, LONDON 51, NY 64. Attempt split: 1×90, 2×82, 3×28 [A].

Normalized class table (per-row CSV: `docs/superpowers/reports/2026-09-26-plan-death-census.csv`, every row carries its id, session, attempt and class):

| class | n | share | attempt 1/2/3 | dominant session |
|---|---|---|---|---|
| write-time feasibility — no_provenance (`entry_zone_edges_or_provenance_unusable`, `identity_not_valid_in_frozen_map`, `scenario_level_id_missing`, `entry_zone_ambiguous`) | 41 | 20.5% | 18/23/0 | ASIA 19 |
| write-time feasibility — rr (gain/risk < min 2.0) | 31 | 15.5% | 16/15/0 | ASIA 15 |
| facts — gap-up/gap-down trigger must reference a level ≥ current price | 28 | 14.0% | 23/2/3 | NY 10 |
| entry-policy shape (`planned_order` on sweep_reclaim, arm on non-armable condition, `entry_mode=pullback`, `wait_confirm`, confirm rule grammar) | 25 | 12.5% | 10/10/5 | NY 9 |
| born-dead authored scenario | 20 | 10.0% | 10/5/5 | mixed |
| schema/field (`scenario economics`, `breakdown{} facts`, `needs exact entry/stop/target > 0`, `entry_zone requires positive [low, high]`) | 18 | 9.0% | 6/7/5 | ASIA 9 |
| obstacle chain omission | 10 | 5.0% | 2/5/3 | ASIA 7 |
| other feasibility (net_nonpositive 4, invalid_geometry 1, other 5) | 10 | 5.0% | — | — |
| breakdown/breakup tape verification | 3 | 1.5% | — | — |
| flip/bias mandatory re-check | 3 | 1.5% | — | — |
| API 402 (insufficient balance — not a plan defect) | 3 | 1.5% | — | — |
| JSON unmarshal | 3 | 1.5% | — | — |
| identity unresolved | 2 | 1.0% | — | — |

**The hidden class — invalidation grammar, publish-time (not in this table).** The grammar refusal lives in the liveness write gate (`trader/plan_liveness.go:30`, W2 2026-09-23), recorded via `RecordPlanLivenessEventWithCheck`, NOT in `planner_rejected_prompts` [A]. Live-log counts of the refusal phrase (both renderings: "outside the supported explicit 1/2 × 5m close grammar" pre-W2, "GRAMMAR REFUSAL" post-W2), read from `/home/hoang/nofx/data/nofx_2026-09-*.log` [A]:

| day | grammar refusals | born-dead refusals |
|---|---|---|
| 09-10 | 21 | 0 |
| 09-11 | 28 | 0 |
| 09-13 | 43 | 0 |
| 09-14 | 21 | 0 |
| 09-15 | 41 | 0 |
| 09-16 | 23 | 0 |
| 09-17 | 16 | 1 |
| 09-18 | 15 | 0 |
| 09-19 | 91 | 1 |
| 09-22 | 46 | 1 |
| 09-23 | 48 | 8 |
| 09-24 | 0 | 1 |
| 09-25 | 0 | 11 |
| **total** | **393** | 23 |

09-24/25 grammar=0 matches boot days (boot 2 on 09-24, boot 3 on 09-25) with few/no full reads [B]. **393 grammar refusals in 12 active days = ~33/day**: the single largest read-killer, and each one lands AFTER a completed read (7–15 min of model time wasted per read chain — see §4).

## 2. Top-3 classes: root cause, prompt vs validator

### (1) Invalidation grammar — 393 publish-time refusals [A]

- **(a) What the prompt tells the model:** `kernel/planner_prompt.go:836` renders `AuthoredInvalidationGrammarLine()`: `// "invalid" GRAMMAR (machine-checked at write; anything else is REFUSED, never accepted as UNKNOWN): EXACTLY one of "5m close above <price>" | "5m close below <price>" | "2x5m close above <price>" | "2x5m close below <price>" — <price> is ONE plain number (no commas), nothing before or after it: no "any", "back", "then", no second clause, no other timeframe, no MSS.` (grammar at `kernel/plan_authored_invalidation.go:133-135`) [A].
- **(b) What the validator demands:** the same four forms, machine-parsed; anything else → `GRAMMAR REFUSAL` + the read's remaining attempts carry the repair (`trader/plan_liveness.go:30`; refusal repeats the grammar verbatim) [A].
- **(c) Where they disagree:** the model writes prose invalidation ("S1 invalid `5m close below 29993.50 (PDH shelf lost); protective stop 29987.00.`", "…`after the touch — shelf failed, stand down.`", "`Invalid if a 1x5m close below 30063.25 after the touch.`") — extra clauses, timeframes ("1x5m"), and prose. The repair prompt DOES carry the law: `kernel/validator_hints.go:72` `RepairInvalidationGrammarLaw` spells the four forms and "one plain number and nothing else" [A]. **Yet attempts 2 and 3 still fail on it** (the 09-25 log's chains at 07:49→07:51, 08:28→08:31 all rejected born-dead on grammar; the flash-AB 14-call corpus shows 5/5 core-passing rows born-dead on the same grammar [A]). Conclusion: telling the model the rule twice is not enough — the fix is deterministic (normalize or machine-rewrite), not more prompt.
- **Does the repair prompt carry what is needed to fix it?** Yes (the exact four forms), but the model cannot convert its own prose into them reliably at max — this class is repair-resistant by construction [B].

### (2) write-time feasibility `no_provenance` — 41 rows [A]

- **(a) Prompt:** `kernel/planner_prompt.go:866-867` requires `economics:{entry_zone:[low,high], … path_levels:[{price, level, level_id:<map id>, role…}]}` on EVERY scenario, with `level_id:<map id>`.
- **(b) Validator:** `trader/structural_geometry.go` refuses `entry_zone_edges_or_provenance_unusable`, `identity_not_valid_in_frozen_map`, `scenario_level_id_missing`, `entry_zone_ambiguous` — the zone edges must be derivable from a level the frozen identity map holds.
- **(c) Disagreement:** the prompt asks for `<map id>` provenance but does not enumerate WHICH ids exist in the frozen map at write time; the model invents or omits ids (missing / not-valid / ambiguous) and the feasibility gate refuses the arm. The repair (`planner_repair.go:108-110` covers identity≠price / unresolved id) reaches the model only after the read.

### (3) write-time feasibility `rr` — 31 rows [A]

- **(a) Prompt:** the strategy's resolved MinRR (2.0) is stated; `kernel/planner_prompt.go:870` already promises an auto-correction precedent: "an arm-target R whose machine-computed value meets the minimum R:R floor is auto-corrected and accepted instead — never round R values".
- **(b) Validator:** `write_time_feasibility` refuses `geometry: rr (gain=… risk=… rr=0.78 min=2.00) — widen …` — the geometry from the authored entry/stop/target fails the floor.
- **(c) Disagreement:** the auto-correct precedent covers arm-TARGET R, not the entry-zone-edge case; the model authors entry/stop pairs whose gap yields rr < 2 and the gate refuses instead of widening (which the refusal text itself instructs: "widen the stop past …").

### (4, context) facts gap-up trigger — 28 rows [A]

- Validator `kernel/plan_doc.go:1231-1234`: on gap-up, "the long scenario's trigger must reference a level ≥ current price (breakout/retest), not a dip back below". The prompt does not state this rule at all [A] — a pure prompt-contract gap, the cheapest of the four to close.

## 3. Plan outcomes per read

Plans table (`plans`, the copy), 2026-09-10+ [A]: sessions mostly publish one `active` plan + `superseded:flip` / `dormant` versions; `no_trade` rows are rare (09-24 LONDON 1, NY 3; 09-23 NY 1; 09-22 NY 1; 09-21 NY 1). Fail-closed reads do NOT appear as `no_trade` rows — the read dies before a plan row is written, which is why the rejected-prompts table and the liveness events are the only census of them [B].

## 4. Time cost of fail-closed reads (09-25 live log, read-only) [A]

From `/home/hoang/nofx/data/nofx_2026-09-25.log`:

- ASIA 05:33:35 → 05:36:55: attempt 1 `ga…` (gap facts), attempt 2 born-dead, attempt 3 born-dead — **3 attempts, ~3.4 min, fail-closed**.
- LONDON 01:53:24 → 02:04:04: attempt 1 parse/schema, repair, attempt 2 parse/schema, reauthor, attempt 3 parse/schema — **~10.7 min, fail-closed**.
- LONDON 06:19:39 → 06:24:15: parse/schema ×2 then gap — **~4.6 min, fail-closed**.
- Flip-hysteresis hold 30 min (`kernel/regime_ledger.go:13`, `FLIP_MIN_HOLD_MIN`) then applies after a failed death re-read, so the session runs with the stale plan/no plan for that window on top of the failed read's own minutes [A/B].

A fail-closed read therefore costs **~3.5–11 min of wall time** and the session then rides the 30-min hold with no plan — on a fast day that is the whole move.

## 5. Recommendations, ranked by reads saved (NO CODE, owner decision)

1. **Deterministic invalidation normalizer at the liveness gate** (393 refusals, the #1 killer): machine-rewrite the model's prose into the nearest grammar form by extracting the direction + price the sentence already contains (e.g. "5m close below 29993.50 (PDH shelf lost); protective stop 29987.00" → `"5m close below 29993.50"`), refuse only when no price/side is extractable. Smallest change, no trading semantics touched; risk LOW (the gate only gets stricter today; the normalizer accepts sentences whose number the model already wrote). Fits the pending fast-market planner wave (born-dead kills whole reads at publish).
2. **Prompt: enumerate the legal provenance ids at write** (41 no_provenance): render the frozen identity map's ids into the prompt next to `level_id:<map id>`. Smallest change (one render block); risk LOW; the validator stays untouched.
3. **Extend the existing arm-target-R auto-correction to entry-zone-edge rr** (31 rows): the refusal text already tells the model what to do ("widen the stop past …"); do it deterministically like the prompt's own auto-correct precedent (`planner_prompt.go:870`). Risk MEDIUM (moves stops — but only where the refusal instructs it); needs the write-feasibility wave's sign-off.
4. **One prompt line for gap-up/gap-down trigger rule** (28 rows): state the `plan_doc.go:1231` rule in the prompt. Cheapest of all; risk NONE.

**What I did NOT do:** no code, no API calls, no merge, no DB writes (copy only), no `.env` reads beyond the earlier dispatch, no live-system interaction.
