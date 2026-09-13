# Current audit, repair and publication checkpoint

The usage-blocked checkpoint under `interim/` is historical. Start with
[REPAIR-STATUS.md](REPAIR-STATUS.md) for the current repair/open-limit table;
[CTO-TRADING-LOGIC.md](CTO-TRADING-LOGIC.md) explains the trading-process judgment.

## Completed baseline coverage

All 30 scoped review assignments are reported:28 primary slices cover1,049 unique
first-party source files /250,582 lines at **63968be62e44db2fb07a92883e02127b9064b0be**.
Two independent cross-reviews use repair 99a06543 and add no primary coverage.
The independent publication check finds all 30 index links and120 standard artifacts,
no duplicate assignments, missing reports, unread items, hash/range gaps or
recorded named-function validator errors. These are baseline consistency facts,
not full-source coverage of later repairs, every test/document read, or runtime proof.

## Repair checkpoints

Control branch: `fix/repo-audit-control-boundaries-20260913`, observed at 939e21db.
Frontend source branch: `fix/repo-audit-web-state-20260913`, pushed at 28a6f32e.
Root is integrating and testing; do not use these snapshot hashes as final publication stamps.

- Go full suite/build and selected race checks passed at 99a06543, before later repairs.
- C# lifecycle source 9140f6c9/f1b7cc10 integrated as e8d2243f/cc766e1c; extracted-method harness and NT8-reference compile evidence are scoped in the repair report. No installed AddOn change occurred.
- Frontend source/dependency/state commits ed85a80a/e26f63a8/28a6f32e integrated as 03090352/f5409132/6c4092bf. Private updated dependencies passed71 files/451 tests and production build at 28a6f32e; npm audit reported0. Default-branch GitHub alert closure is not asserted.
- a982cc74 repairs completed partial exits and cumulative ENTRY terminal receipt accounting. Positive-filled terminal-cancelled EXIT wire handling remains open. A receipt transaction cannot guarantee process-local hooks across crash; pending receipts still need later delivery/reconciliation.
- df4af389 fixes the weekly reader's production daily-input call site. Other generic epoch aggregation and daily-data intraday-gap limits remain separate.
- Guide 6058d9fe/939e21db removes old ATR/permission claims and explains residual exposure. Final GUIDE_BUILT_REV must follow final source integration.

## Remaining source and runtime limitations

Concrete execution gaps: positive-filled cancelled EXIT orders outside the Filled-only
exit wire path; no atomic broker-side expected-position fence; safe observer replacement
and final disposal/handoff; crash-time process-local hook delivery; unresolved replay and
ambiguous multiple-row attribution. These are not dismissed as minor cosmetic debt.

Active UI/data gaps: selected-account history/chart completeness; bounded loaded history
versus complete aggregates; current-forming-candle marker association; open-order errors
still converted to empty data; bulk model replay/extra-instance thinking knobs; model
submit busy lifecycle; account-name persistence; DayPlanEditor polled-draft/default/
inheritance/translation problems. Specific corrected-PNL, chart identity, cache refresh
and breaker display defects have later repairs and should not be relisted as wholly open.

Runtime and strategy evidence: installed NT8 callback/OCO/fill behavior, first live
structural composition/refusal, backup restore rehearsal, and causal out-of-sample net
expectancy remain unverified by this audit. Source understanding is not runtime or
profitability approval. Archived exploratory script flaws are qualified separately
from currently relied-on measurements in their source reviews.

## Root-owned final publication work

Final integrated source revision: **pending**. Combined Go/web/race/build log results:
**pending** (root started final combined 01; no result inferred). Final guide revision:
**pending**. Final source/evidence bundle, manifest hashes and publication links:
**pending**. Root updates these once actually verified; preceding branch-green results
must not be relabelled as final merged results.

Verified earlier progress bundle: `/tmp/nofx-audit-progress-30-reviews.bundle`,
containing four branch refs and requiring baseline 63968be. It is incremental,
not a standalone full-repository restore. Preserve that limitation when shipping
its replacement and include the separate repair reports in the final evidence set.

No source audit action deployed/restarted the bot or NT8, changed owner settings,
read/wrote live trade records or submitted an order. DAILY loss remains the owner's
control; no additional mandatory per-trade cap or quantity policy was introduced.
Ordinary Stop must retain protection/close observers for held positions until a
safe handoff exists.
