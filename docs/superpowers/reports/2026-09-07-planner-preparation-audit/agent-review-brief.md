# Review brief for codex 101 — proposal only, not dispatched

## Purpose and boundary

The owner requested a detailed, read-only assessment of whether Planner prepares coherent NQ/MNQ session plans. This brief is a review artifact for the agent working on the dashboard. It does not authorize implementation, edits to the live prompt, settings, order management, deployment, or database repair. No message has been delivered to codex 101 through the desktop messaging tool.

## Reference baseline and artifacts

Source under audit: `5457ac5accd97c3519bf6d16ead147a0db2ab0d0`. Consult the adjacent README, scenario inventory, candidate-selection records, publication-context evidence and confirmation probe. Compare against the currently proposed UI revision before treating a source finding as still present; do not attribute another lane's changes without branch/worktree evidence.

## Questions for review

1. Does the dashboard distinguish market-input time, plan publication time and latest evaluation time? A plan published at 22:16 used context around 29665.75, while the last completed minute bar by publication closed at 29689.00. The source reuses original input across repairs. A creation-time badge alone is insufficient evidence of freshness.
2. Does the history diff show same-count scenario changes: direction, trigger, confirmation rule and side, entry, stop, target, first obstacle, invalidation and expiry? The current coarse diff omits most of them.
3. Does a level's chart role remain visible when it is excluded from setup selection? Candidate score zero and the cap-reason string are not reliable explanations of the actual exclusion process. Do not turn them into a confident “poor level” UI label.
4. Does each scenario distinguish a market reference, a confirmation event, an authorization and a broker-accepted order? The owner's present audit prioritizes preparation quality; execution state must not substitute for that assessment.
5. Is the exact version supplied to the executor preserved beside its response? Current resolution before and after an AI call creates an attribution race opportunity; no historical occurrence is claimed.

## Additional level and context review points

The full inventory in report section 5 adds six concrete questions. Each proposed UI label must be backed by the actual producer:
- Does the range disclose whether it is a prior calendar-day range, CME session range or rolling value-area fallback?
- Are actual window coverage and developing-bar inputs visible, including the main 2,000-bar and production 2,500-bar limits?
- Are the two profile estimators distinguished, and are date-only cache identity and frozen partial-profile limitations acknowledged?
- Are evening MID-O absence and developing OR/IB represented truthfully?
- Does the final output really retain the claimed HTF/1h reserved reference after every cap and sort?
- Does RV specify its actual baseline rather than implying volume, remaining opportunity or 20 completed days?

These are source findings with linked references in section 5. Historical impact and benefit of alternative policy remain separate questions.

## Reproducible high-priority proof

The inert `confirmation-probe.go.txt` calls the actual source revision's confirmation functions. It constructs two cases, each mirrored long/short:

- One fully closed minute inside a new five-minute interval returns a five-minute confirmation even though the five-minute interval has not ended.
- A completed reclaim-side close before a later sweep can satisfy a nominal sweep-then-reclaim sequence, even without a subsequent reclaim.

All four expected results are false; all four recorded actual results are true. This is a synthetic logic proof, not a claim that the four events were real trades. A future correction must retain the probe as an independent regression target and add boundary/ordering cases rather than changing the expected result to match existing behavior.

## Proposed acceptance conditions, subject to owner approval

- One canonical scenario event sequence, with event time and known-at time, is consumed by the plan card, executor prompt and eligibility checks.
- Confirmation requires fully completed bars at the stated timeframe; a later event cannot be satisfied by an earlier event outside the required order.
- Raw analytical level prices remain distinguishable from executable tick-rounded order prices.
- Candidate lineage records the original score and components plus the actual exclusion stage. Null means unavailable; zero is a measured value only when actually calculated.
- Full market-map roles survive display prioritization, with explicit zone bounds and merged constituent identities.
- A plan becoming usable has an explicit fresh-context assessment after authoring or repair, with the original snapshot retained for audit.
- Version history includes meaningful scenario diffs, and executor attribution binds immutably to the prompt version actually read.
- No claim of strategy profitability is made from repeated versions, two days of valid touch observations or a selected-only sample.

## Required review response

For each finding, report agreement/disagreement, exact source reference at the reviewed revision, a concrete counterexample or corroborating test, UI/backend impact, and the narrowest proposed change with its acceptance criterion. Separate directly proven behavior from historical occurrence and from proposed trading policy. Return the review for owner approval before any fix is implemented.
