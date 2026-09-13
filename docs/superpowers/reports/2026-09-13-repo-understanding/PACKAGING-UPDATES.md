# Packaging-time repair updates

This additive disposition notice supersedes the specific older open items named below. It preserves the original reports and their revision-scoped evidence. Commit existence and the corresponding source/report changes were inspected during packaging; this packaging step did not rerun repair tests. The root reviewer still owns final combined verification and revision stamps.

| Later committed repair | Superseded historical open item | Scope and remaining limit |
| --- | --- | --- |
| `c20d0a82` — unresolved exits and account truth | First-ever missing position snapshot treated as flat; valid exit quantity greater than currently materialized entry discarded | Missing/stale account state refuses admission; excess exit receipt remains pending for later resolution. Pending evidence does not itself prove replay or causal execution ordering. |
| `d3e4638e` — observer replacement handoff | Replaced reconciliation observer continues indefinitely | Synchronous subscriptions and receipt-channel drain retire the replaced instance's periodic worker. Ordinary Stop retains protection observation; an already running reconciliation pass may finish. This does not prove live NT8 lifecycle scheduling or all shutdown paths. |
| `3f21431a` — terminal cancelled/rejected EXIT fills | Positive-filled terminal EXIT outside the Filled-only wire path | The bounded C# repair includes positive terminal cumulative exits and retains bracket siblings for residual exposure. Its report records an extracted 89-assertion harness and five-source reference compilation; these are reported offline evidence, not an installed AddOn or real fill. |
| `dd11670b` — open-order display uncertainty | HTTP failures displayed as fresh empty orders; editable existing account name silently ignored | Same-view prior snapshot retained as UNKNOWN/stale; selected-account changes invalidate requests, while the banner discloses trader-bound endpoint scope. Existing binding names are read-only; account creation remains editable. Backend selected-account filtering and binding migration remain separate work. |

**Still active at packaging:** the unified ordered-execution repair and its final verification. Do not interpret the preceding patches as resolving every interleaving between transport receipt, entry accounting, exit accounting and position reconciliation. The root reviewer will update the final disposition after that work and merged-head tests finish.

**Still unverified or open:** broker-side atomic expected-position fencing; crash-safe delivery of process-local hooks; eventual pending-receipt replay; ambiguous multi-row attribution; selected-account completeness in history/chart endpoints; other explicitly listed UI/persistence/operations findings without a named repair. No source audit establishes installed NT8 behavior, backup restoration, production safety, or out-of-sample profitability.

Source receipts: [first snapshot/pending exit](https://github.com/johnwick2921-cyber/nofx/commit/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d), [observer replacement](https://github.com/johnwick2921-cyber/nofx/commit/d3e4638e495425874aedba34a2fab0585acdae85), [terminal EXIT](https://github.com/johnwick2921-cyber/nofx/commit/3f21431aff0d7830b0dc0457417365a7a555aee7), [order display](https://github.com/johnwick2921-cyber/nofx/commit/dd11670be159760942d886b5b8c1aef2573e0d02). These identify source commits, not final release stamps.

## Receive-order and subsequent boundary checks

9b379c8c installs one account/symbol execution owner before its own entries,
processing OrderUpdate, Fill and PositionClose in TCP receive order. Real TCP
tests distinguish entry1→entry2→exit1 from entry1→exit1→entry2 and preserve
actual residuals and cost basis. Independent review and focused race tests
cover owner replacement, raw/advisory replay, cache resurrection and outbound
callback progress. This is receive-order accounting, not reconstructed exchange
timestamps or a durable transport journal. Source commit9b379c8c is followed by
additional positive-rejection/entry-snapshot checks; final combined results
remain root-owned and must replace this qualification before release readiness.

The AddOn candidate identifies itself as2026-09-13-execution-evidence; Go's
expected source marker matches. Received installed-runtime identity was not
changed or verified by this source audit. The final reference compile and
regenerated89-assertion harness pass at this marker.
