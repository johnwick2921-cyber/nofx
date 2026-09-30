# RESEARCH-PLAN-DEATH — census of why planner reads fail

- **Dispatch:** 2026-09-26 04:44Z (owner GO 23:4x CT "send idle agents out"). Lane DS-101.
- **Branch:** `research/plan-death-census` (claim `092e2379`), base `origin/dev` `04ae1c2f` (`git log -1 --format=%h` = `04ae1c2f`).
- **Read-only:** no code changes, no API calls, nothing merged. DB COPY only (`ab.db` from 09-25 22:39 CT). Live logs read-only under `/home/hoang/nofx/data/`.
- **Specs I built from (L3):** this dispatch message; `trader/auto_trader_planner.go` (`runPlannerReadCoreObserved` @ :2052 at `04ae1c2f`), `store/planner_rejected.go` (:71), `trader/plan_liveness.go` (:30), `kernel/planner_prompt.go` (:836, :866), `kernel/plan_authored_invalidation.go` (:130-151), `kernel/validator_hints.go` (:72), `trader/structural_geometry.go`, `kernel/plan_doc.go` (:1231-1234).

## 1. Scope correction (CTO refutation, 04:55Z — accepted)

My first draft miscounted. The 393 "grammar refusals" were NOT refusals [A]:
(a) pre-W2 (before boot 662c79bd, 09-23 18:50 CT) the line read "plan liveness write: … UNKNOWN: authored invalidation is outside the supported … grammar; ACCEPTED for this check" — a per-scenario WARN on an ACCEPTED plan (data/nofx_2026-09-23.log 01:40:12). (b) It logged once PER SCENARIO (4 lines per write on 09-23 01:51:47, 13:56:57, 14:01:18). (c) Since W2 every boot line reads "invalidation: enforced (grammar) · grammar refusals=0" (09-24 10:19:46 → 09-25 20:14:26). **Live grammar refusals post-W2 = 0** [A]. The pre-W2 warn counts (393 log lines, ~33/day) stay in this report as HISTORY ONLY — they were acceptance warnings, not kills.

## 2. The post-W2 census (09-23 18:50 CT onward) — counted per READ, not per line

`planner_rejected_prompts` post-W2: **83 rows = 38 reads** (a read = one attempt-1 chain; ids named). 14 reads were KILLED (3 rejected attempts, no plan); the other 24 recorded rejects then either published or stopped early.

**Killed reads (ids → final reject class), wall minutes [A]:**

| chain ids | day/session | final killer | wall |
|---|---|---|---|
| 339-341 | 09-23 ASIA | flip/bias mandatory | 9.6 |
| 349-351 | 09-23 ASIA | obstacle chain | 17.3 |
| 352-354 | 09-23 ASIA | obstacle chain | 2.9 |
| 376-378 | 09-24 ASIA | flip/bias mandatory | 6.2 |
| 385-387 | 09-24 ASIA | obstacle chain | 7.7 |
| 391-393 | 09-24 ASIA | tape verify (breakdown) | 2.3 |
| 357-359 | 09-24 NY | tape verify (breakdown) | 3.2 |
| 360-362 | 09-24 NY | entry-policy shape | 7.4 |
| 363-365 | 09-24 NY | born-dead | 14.9 |
| 394-396 | 09-25 LONDON | entry-policy shape | 10.7 |
| 397-399 | 09-25 LONDON | born-dead | 3.3 |
| 400-402 | 09-25 LONDON | gap trigger | 4.6 |
| 409-411 | 09-25 NY | born-dead | 12.0 |
| 412-414 | 09-25 NY | entry-policy shape | 2.8 |

Mean **7.5 min** per killed read (max 17.3). The 24 non-killed reads still burned reject attempts (born_dead 6, obstacle_chain 6, entry_policy_shape 3, identity_unresolved 2, feas_no_provenance 2, feas_other 2, feas_net 1 as their last recorded rejects) before publishing or stopping.

**Read-final reject class over all 38 reads:** born_dead 9, obstacle_chain 9, entry_policy_shape 6, feas_no_provenance 4, flip_bias 2, identity_unresolved 2, feas_other 2, tape_verify 2, feas_net 1, gap_trigger 1.

## 3. Post-W2 killers, re-ranked

1. **born-dead** — 9/38 read-finals, 3/14 kills. Every post-W2 born-dead reason is a REAL market breach ("authored condition \"5m close below …\" breached between read and publish") [A]. This is the staleness killer: the read is slower than the move. The flash-AB (same day) measured reads at 192–318 s wall, and on fast days that window is where the flip/breach lands.
2. **obstacle-chain omission** — 9 read-finals, 3 kills (ids 349-351, 352-354, 385-387). Prompt (`kernel/planner_prompt.go:866-870`) demands `first_obstacle{price,level,family,response}` + `path_levels[]`; the validator refuses omissions ("S3 obstacle chain: omits SWG-H·15m 30644.25 (273.5 pts …)"). The prompt cannot enumerate WHICH obstacles the frozen map holds, so the model omits ones it never saw.
3. **entry-policy shape** — 6 read-finals, 3 kills (ids 360-362, 394-396, 412-414): `planned_order` illegal on sweep_reclaim legs, split-contract leg counts, arm on non-armable conditions. A prompt↔schema contract gap, deterministic and cheap to close.
4. **write-time feasibility no_provenance / rr** — big in the FULL 200-row pre-W2 set (41/31) but post-W2 only 4 read-finals and 0 kills: the W3 entry-policy/authoring waves already shrank it [A]. It is no longer a top killer.

## 4. Time cost (post-W2, per read) [A]

- Killed reads: mean 7.5 min, max 17.3 min (ids above).
- All 38 reads: first-to-last attempt span median 3.2 min.
- After a failed death re-read the flip hysteresis holds 30 min (`kernel/regime_ledger.go:13`, `FLIP_MIN_HOLD_MIN`): the session runs with the stale plan/no plan through the hold. On 09-25 NY the reads at 08:07-08:10 published, but the LONDON chains 394-402 and NY 409-414 killed — the session rode the hold through the fast sell-off.

## 5. Recommendations, re-ranked on post-W2 data (NO CODE; owner decision)

1. **Faster reads — the born-dead killer IS staleness.** 3 of 14 killed reads + 9 of 38 read-finals died to market breach during the read window. This is the same conclusion as flash-AB: read wall must come down (192–318 s today). Options, smallest first: (a) the re-read-model knob on flash-day policy (owner's separate GO; flash-AB found Flash NOT faster at max on this corpus, so the knob alone is insufficient — cap and prompt matter); (b) shrink what the read re-derives (reuse the frozen map/seats instead of re-authoring everything). Risk: LOW for (a) behind a knob; MEDIUM for (b).
2. **Obstacle-chain: enumerate the frozen map's obstacles in the prompt** (like the provenance fix below) — the model cannot cite obstacles it never saw. Smallest change, LOW risk. Fits the pending fast-market planner wave.
3. **Entry-policy shape: state the legal arm shapes in the prompt** (planned_order illegal on sweep_reclaim, split=EXACTLY 2 legs, armable conditions list). The validator already names each refusal verbatim; the prompt lacks the list. LOW risk.
4. **Provenance ids in the prompt** (post-W2 4 read-finals): render the frozen identity map ids beside `level_id:<map id>`. LOW risk.
5. **The grammar normalizer from the first draft is DROPPED**: post-W2 live grammar refusals = 0 and no attempt-level grammar rejects exist in the rejected table. Nothing to fix there.

**What I did NOT do:** no code, no API calls, no merge, no DB writes (copy only), no live-system interaction.
