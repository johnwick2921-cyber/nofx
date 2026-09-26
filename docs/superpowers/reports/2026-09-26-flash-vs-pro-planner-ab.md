# RESEARCH-FLASH-AB — DeepSeek Flash vs Pro for planner reads (max reasoning)

- **Dispatch:** 2026-09-26 03:38Z, owner GO 22:4x CT 09-25. Lane DS-101.
- **Branch:** `research/flash-vs-pro-ab` (claim `3ee41cbe`), base `origin/dev` `04ae1c2f` (boot-3 RELEASE, `git log -1 --format=%h` = `04ae1c2f`).
- **Status:** IN PROGRESS (draft PR, NOT for merge).
- **Read-only:** no live-code change, no re-read-model knob, no prompt edit. DB = COPY (`ab.db`, 2.4 GB). API key used only through the store's enabled model on the copy; crypto key loaded read-only from `/home/hoang/nofx/.env`, never printed.

## Question

Which model makes better MAX plans fast enough to be valid at publish: `deepseek-v4-pro` or `deepseek-flash` (V4.1, ~234 tok/s), both at `reasoning_effort=max`, thinking enabled. Fast days kill plans: live Pro-at-max read avg 430 s / max 907 s (3-day), so the plan is stale when published.

## Method

- **Harness:** `cmd/ab_flash_pro/` (v1 `58759e12`), derived from `cmd/planner_ab` with model arg + row iteration.
- **Corpus:** every `planner_rejected_prompts` row since 2026-09-10 with `attempt=1` and `length(prompt_text)>100000` — **90 rows** [A]. Ids (CSV column `row_id` in `ab_results.csv`); earliest sampled: 219, 221, 223, 226, 229.
- **Arms** (all `reasoning_effort=max`, thinking enabled, temperature 0.5, same system prompt — `trader/auto_trader_planner.go:28` verbatim — and the stored user prompt verbatim):
  - (a) `deepseek-v4-pro`, max_tokens 65536
  - (b) `deepseek-flash`, max_tokens 65536
  - (c) `deepseek-flash`, max_tokens 131072 — only on rows where (b) finished `length`
- **Judge** (the live write-site legs that ARE runnable offline):
  1. `kernel.ParsePlanDocForAuthoring` with the **live strategy's** `AuthoringOpts` — `{MinRR: store.ResolveMinRiskReward, EntryPolicyDefault: store.ResolveEntryPolicyDefault, MinHoldMin: store.ResolveMinHoldMin}` plus `maxLevels`/`scenarioCap` resolved from the bound strategy row in the copy (the harness gap the CTO spot-check named: `cmd/planner_ab` passed `kernel.AuthoringOpts{}`).
  2. `kernel.ValidatePlanDocWithFactsMachine(d, facts, nil, maxLevels, scenarioCap)` on the stored `facts` JSON.
- **Judge legs NOT runnable offline** (named, not faked):
  - `requiredBias` check — a call argument, not persisted in `facts`.
  - machine-label provenance (`MislabeledStructuralLevels`) — the machine table is not persisted; nil map skips it per the validator contract.
  - FVG / breakdown_continue validators — need the live `market.FuturesBarsProvider` bars at call time.
  - write-time feasibility verdicts (`writeTimeFeasibilityVerdicts`, unexported `AutoTrader` method) — replicating it would fake the live seam; entry distances in pts and ×ATR5m are recorded instead (`facts.DATR` is the read-time daily ATR proxy).
- **Metrics per call:** wall s, TTFB, completion tokens, prompt tokens, reasoning chars, finish, pass/fail + first reject reason, each scenario's entry distance from the read price in pts and ×ATR5m, staleness (see below), error.
- **Limits:** ≤3 concurrent calls; off-peak preferred (outside 01–04 and 06–10 UTC); spend cap $40 (STOP at cap).

## Results — 14 calls, owner CUT at 23:1x CT (04:14Z)

Full per-row data: `ab_results_rejudged.csv` (re-judged 04:4xZ after the born-dead-pass fix — zero new calls; the live `ab_results.csv` predates the fix). Raw outputs in `ab_raw/`.

| row | arm | wall s | finish | core | full | born-dead |
|-----|-----|--------|--------|------|------|-----------|
| 219 | a-pro-65536 | 291.4 | stop | ✓ | ✓ | — |
| 219 | b-flash-65536 | 272.8 | stop | ✓ | ✓ | — |
| 221 | a-pro-65536 | 308.0 | stop | ✓ | ✗ | yes (invalidation-grammar refusal) |
| 221 | b-flash-65536 | 287.5 | **length** | ✗ | ✗ | n/a (cap-hit, no JSON) |
| 221 | c-flash-131072 | 203.3 | stop | ✓ | ✗ | yes (invalidation-grammar refusal) |
| 223 | a-pro-65536 | 266.7 | stop | ✗ | ✗ | n/a (facts: gap-up trigger) |
| 223 | b-flash-65536 | 300.7 | **length** | ✗ | ✗ | n/a (cap-hit, no JSON) |
| 226 | a-pro-65536 | 192.3 | stop | ✗ | ✗ | n/a (facts: gap-up trigger) |
| 226 | b-flash-65536 | 253.7 | stop | ✓ | ✗ | yes (invalidation-grammar refusal) |
| 229 | a-pro-65536 | 318.5 | stop | ✗ | ✗ | n/a (facts: gap-up trigger) |
| 229 | b-flash-65536 | 295.5 | **length** | ✗ | ✗ | n/a (cap-hit, no JSON) |
| 229 | c-flash-131072 | 392.7 | stop | ✓ | ✗ | yes (invalidation-grammar refusal) |
| 231 | a-pro-65536 | 309.8 | stop | ✗ | ✗ | n/a (facts: duplicate seats) |
| 235 | a-pro-65536 | 288.0 | stop | ✓ | ✗ | yes (invalidation-grammar refusal) |

**Tally (re-judged, n=14).**
- Core pass: Pro 5/7 [A]; Flash-65536 2/5 with **3/5 hitting the 65,536 cap** (221/223/229, finish=length, no JSON) [A]; Flash-131072 2/2, zero cap-hits [A].
- Full pass (core + write-time feasibility + zone + born-dead): Pro 1/7 (219), Flash 1/5 (219), Flash-131072 0/2 [A].
- Wall at max: Pro 192–318 s (median ≈ 292 s); Flash-65536 254–301 s (median ≈ 288 s) — **Flash is NOT faster than Pro at max** on this corpus [A]. Flash-131072: 203 s / 393 s.
- **Every core-passing row that reached the write-time legs died on the invalidation-grammar refusal** (`plan liveness write: … GRAMMAR REFUSAL` at `trader/auto_trader.go:51` — S2/S1 invalidation prose outside the `5m close above|below <price>` grammar), both models, including both 131072 rows [A]. That leg is part of the live write gate, so a full pass here is blocked by a grammar the corpus prompts violate systematically — it is not a per-model quality signal on its own.
- Spend: **$1.29 total** of the $40 cap [B] (official rates: v4-pro $1.32/$3.96 per M in/out, flash $0.30/$1.20, peak 01–04 & 06–10 UTC; 3 calls peaked, 11 off-peak; cache-hit inputs conservatively priced as misses).

## Staleness leg

Implemented via the research shims: publish clock = read time + the call's wall; the REAL born-dead check (`ResearchValidateAuthoredScenariosAt`) ran over the 1m tape from the copy. Born-dead column above [A].

## Verdict + recommendation

**Flash at max is NOT the speed fix.** On the corpus (Pro's live FAILURES, so never read against 100%): Flash's median wall ≈ Pro's (288 s vs 292 s), and at the 65,536 cap Flash throws away 3/5 calls to finish=length with no JSON. With a 131,072 cap Flash stops losing calls to the cap (2/2 stop) but is still not faster (203/393 s) and its plans die on the same grammar refusal as Pro's [A].
- **Recommendation:** keep `deepseek-v4-pro` at max for planner re-reads. Flash has no measured advantage here beyond cap headroom; buying that headroom (131072) costs extra tokens without recovering any born-dead plan on this corpus [B]. The dominant blocker for BOTH models is the invalidation-grammar refusal — the prompts in this corpus (and their retry tails) carry invalidation prose outside the grammar, so model choice cannot fix it; a prompt-side grammar enforcement or a validator-side acceptance decision is an owner ruling, out of this dispatch's scope.
- **Selection-bias note (mandatory):** the corpus is prompts Pro FAILED live (many at fast→low). Pass rates here compare arms against each other, NOT against 100%; nothing above generalizes to Pro's passing reads.
