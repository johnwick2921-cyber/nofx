# Repository control-boundary repairs — in progress

Owner requested detailed traced review and repairs until complete. Source base:
`63968be62e44db2fb07a92883e02127b9064b0be`; branch
`fix/repo-audit-control-boundaries-20260913`; session
`reporepair-8d38e6ac/Codex[unlisted]`. Isolated worktree
`/tmp/nofx-audit-repairs-20260913`. No production database, settings, account,
order, deployment or runtime changes. Source review evidence is on
`docs/repo-understanding-20260913` under its separately claimed report scope.

Spec freshness: `dfda15e1425ba38671deb7ffa7a4f211a27846e2 test: isolate session clock fixtures from weekly backfill workers`
was the last AUDIT-CHECKLIST change at acceptance. Current tracked CLAUDE-canon
is read; main stays deploy-only. Branch names identify provenance, not author.

## Confirmed and repaired so far

- Authenticated object ownership: production router regression reproduces a
  foreign trader DELETE returning 200 and removing its equity child row. The
  middleware now validates all query/body trader IDs and trader path IDs. The
  delete handler additionally checks ownership; the store authorizes before
  touching children and performs both deletes transactionally. Injected child
  deletion failure verifies rollback. Own-trader reads remain available.
- Q&A: no plan plus available bars reproduces a nil row panic. Creation time
  is absent (zero) when no row exists. Historical fallback uses the existing
  trader-scoped query, tested with a foreign plan and an empty owned trader.
- Strategy updates: oversized prompt reproduces HTTP rejection after persisted
  changes. Token validation now runs before write/log/reload. Regression compares
  stored name and configuration bytes after rejection.

## Verification to date

Go 1.25.13. Focused production-route/transaction/context/rejected-save tests fail
on the preceding implementation and pass on repaired code. Existing plan/risk
ownership and strategy create/edit preservation tests included. Tests use
throwaway SQLite files and synthetic bars, never the owner's database or wire.
A first sandbox run failed because the external Go cache was read-only; it was
rerun with approved cache access. That setup failure is not a product test result.

Full suite, race coverage where relevant, independent repair review, guide rev
finalization, publication and final artifact verification are **pending**.
No runtime incident, exploitation, real fill, or deployment success is asserted.

## Execution boundary repairs

- Boot refusal: six adapter-side regression cases failed to see the boot latch;
  a ready loopback peer additionally received a limit entry with refusal true.
  Checks now precede market/limit/stop entry registration/send. Boot assertion
  runs before saved-trader autostart. Existing capability and unbound-account
  tests pass alongside the new tests. The loopback is a synthetic peer, not NT8.
- Broker-terminal retirement: same-version cancelled and filled rows each minted
  a fresh authorization in a temporary ledger. Retirement now precedes successor
  creation. Focused store tests confirm new-version and boot-sweep behavior.

These changes have not been deployed. Full combined review remains pending.

## Session mutation repair

Ask-Planner apply now resolves a session before dereferencing it; session gaps
return the existing stale-plan refusal. Apply and realign use the wrap-aware
chain date already used by plan reads, so after-midnight Asia mutations address
the previous calendar date's chain. Focused tests cover the session gap and
overnight date plus existing ask/realign/context tests. This defect was established
by source inspection; the new helper tests were not run against the old code.
Atomic proposal application/version binding remain separate open findings.

## Broker-book evidence and cancellation

Regressions reproduced missing/null `orders` parsing as an empty book, stale
book evidence authorizing cancellation, and a failed cancel send returning true.
The parser now rejects an uncomputed list (explicit `[]` remains valid). Live
book data and receipt time are read under one cache lock. Signal/row cancellation
uses the existing snapshot age limit and reports send failure as failure.
The existing independently established flat-position exception is preserved.
Focused parser, cancellation and desync tests pass. No real cancel was sent.

## History stream teardown

A race-detector test drives actual TCP history-frame dispatch concurrently with
subscription teardown/replacement. The preceding implementation panicked with
`send on closed channel` in `TCPServer.readLoop`. The read lock now covers the
nonblocking send; teardown closes only under the corresponding write lock.
The same test passes under `-race` and delivers a sentinel after the churn.
This proves the exercised loopback interleaving; it is not NT8 integration proof.

## Boot sweep settlement

A temporary-ledger regression reproduced successful cancel transmission becoming
`cancelled` with snapshot ID zero. The sweep now persists cancellation intent,
uses the guarded sender, and leaves the row pending. A received cancellation
update does not bypass pending snapshot confirmation. Existing same-version
boot-sweep reauthorization occurs only after confirmation. The recorded completion
counter increments transactionally with that transition, not at send time.
The boot line now says requests, not completed cancellations. Focused boot-sweep,
reauthorization and cancel-lifecycle tests pass. Old tests that equated send with
settlement were updated to inject persisted broker evidence. Historical counter
values are not retrospectively corrected by this code change.

## First complete Go-suite result

`go test ./...` at repair checkpoint `234b0262` completed: every package passed
except store. Three tests failed: the new terminal-state fixture duplicated
canonical state names, and two older append-only fixtures expected replacement
without advancing authorization version. The fixture now enumerates canonical
terminal states; append-only scenarios explicitly advance the plan version.
Their original record-retention and no-row-per-cycle assertions remain. All
three plus the expanded same-version regression pass in a focused rerun.
This is not yet the final combined-suite result.

## Account isolation and reset synchronization

Synthetic regressions reproduced: a bound account with no balance snapshot (and
an unbound adapter) returned another account's equity; another account's open
MNQ row suppressed materialization of the bound account's held position; reset
replaced pending maps without acquiring their mutex. Balance now returns
unavailable without its own snapshot. Reconciliation's legacy fallback queries
only the same trader's unassigned-account row. Reset takes pendingMu before mu,
matching reconciliation's lock order. Focused tests pass under `-race`, including
the existing legacy-empty-account deduplication case. An older test explicitly
expecting foreign-balance fallback was corrected. No actual account was selected
or mutated. The lock-ownership regression is controlled synchronization evidence,
not proof of every possible interleaving.
