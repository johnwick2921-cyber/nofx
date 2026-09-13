# Audit and repair checkpoint

## Completed scope

All 30 scoped assignments are reported. At baseline 63968be62e44db2fb07a92883e02127b9064b0be, 28 primary reviews cover 1,049 unique first-party files / 250,582 lines. Two bounded independent reviews examine repair 99a06543 and add no primary files. Coverage and publication validators establish artifact/hash/range/function-note consistency, not a second semantic reading or runtime correctness. The 30 reports and their 120 standard artifacts are preserved.

Current candidate: `cd2978b77da54e2fceddfb19e1d3d148bd2bfb62` on `fix/repo-audit-control-boundaries-20260913`. The disposition and CTO assessment include the ordered-execution and positive-entry/replacement follow-ups through this revision. Those implementation tasks are committed; they are not awaiting an unspecified future repair. CORE-TRACE.md pins selected declarations to this candidate.

Historical checkpoints remain revision-scoped: Go full suite/build and selected race checks at 99a06543; frontend 71 files/451 tests plus build at 28a6f32e; later chart follow-up 9 focused tests/build; final-marker C# reference compilation and 89 extracted assertions. No earlier green result is relabelled as a pass at the final candidate.

## Substantive remaining limits

Receive-order processing does not reconstruct exchange chronology or pre-owner queued events. No durable inbound journal or crash-atomic process-local hook delivery exists; storage failures still need later evidence/replay. Synchronous callbacks backpressure TCP; continuing cumulative exposure does not rebuild all already-emitted terminal analytics. Broker-side atomic expected-position fencing remains absent. Ambiguous multiple-row attribution refuses. Replacement handoff is repaired, while installed NT8 scheduling/OCO, reconnect and final shutdown still need controlled verification.

UI limits include backend selected-account completeness in history/chart/orders, forming-candle marker association, bulk model payload replay/extra-model thinking knobs, submit lifecycle and DayPlanEditor draft/default/inheritance/translation issues. Existing account names are now read-only; backend binding migration is separate. Corrected-PNL fallback, stale response identity and open-order error-to-empty have named repairs and are not wholly open.

The audit did not establish backup restoration, a first live structural composition/refusal, calibrated stop buffers, target optimality or causal out-of-sample net expectancy. Historical exploratory research is qualified separately from current measurement dependencies. Main runtime, owner settings, accounts and live records were not changed by this source audit; SIM and the owner's daily-loss policy remain intact.

## Final verification

Root-owned ledger. Frozen verification tree: `e333de41bfffec2ea2cce67ab4296ce8bbd12fc0`; backend source candidate and guide stamp: `cd2978b77da54e2fceddfb19e1d3d148bd2bfb62`. The intervening freeze commit changes protected hashes/guide/report metadata. The guide stamp identifies this source candidate, not a running or shipped binary. Results below are supplied by the root runner; this documentation pass did not rerun them.

| Required record | Result |
| --- | --- |
| Final combined Go suite | Pending root runner result |
| Final Go build | PASS, root runner at frozen tree |
| Final relevant race verification | Pending root runner results |
| Final frontend suite and production build | PASS: 71 files / 454 tests; production build passes with existing chunk-size warning |
| Offline smoke checks | PASS, root runner |
| Final C# source marker/reference compile/harness receipt | Pending root consolidation of revision-scoped receipts |
| Final guide/source identity and publication hashes | Pending root stamp |
| Final evidence bundle, PDF and downloadable Markdown links | Pending root publication |
| Deployment / installed NT8 / live SIM lifecycle | Not performed by this audit; no approval inferred |

After completing this ledger, regenerate assembled reports with tools/build-reports.py and update index sizes from report-package.json. The earlier progress bundle was incremental and required baseline 63968be; it was not a standalone restoration backup. Final publication must retain that distinction.
