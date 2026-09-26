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

## Results

PENDING (tables + CSV after the run).

## Staleness leg

PENDING (born-dead vs the 1m bars from the copy — implementation note).

## Recommendation

PENDING — with evidence tiers and the mandatory selection-bias note: the corpus is prompts Pro FAILED live (many at fast→low), so arms compare against each other, not against 100%.
