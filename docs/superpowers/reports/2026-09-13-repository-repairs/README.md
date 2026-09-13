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

## Successor provenance

The production-style fresh successor lost BootID and ArmedUnderVersion because
its early Create bypassed initialization. Regression across canonical terminal
states reproduced both blank values. Successor creation now stamps this process
and the new authorization version. Focused provenance/append-only tests pass.

## Authenticated chat memory

Both HTTP handlers accepted a caller-selected numeric conversation ID. A
synthetic authenticated `/clear` request reproduced deleting a different owner's
history while leaving its own history intact. HTTP now always derives that
identity from authenticated middleware. Normal and SSE handler regressions pass;
no model/network call or real conversation was used. Telegram's separate identity
flow is unchanged. Shared mutable AI-client selection remains an open finding.

## Cancellation evidence timing and restart budgets

[A] A regression reproduced a fresh but pre-request empty snapshot settling a
later cancellation. Confirmation now requires a valid explicit book whose receipt
is at or after the persisted cancel request; missing and future receipt times
remain unavailable. The test follows older empty, newer working and newer empty
evidence. A second regression reproduced a previous process exhausting the new
process's retry cap. The cap check now applies the process identity before
counting attempts. Both regressions and existing settlement/boot-sweep tests pass
(`/tmp/nofx-cancel-evidence-after.log`). No broker or runtime settings were touched.

## Stop-entry refusal and other scenarios

[A] The actual placement-loop regression reproduced S2 being cancelled with
`one_live_entry: S1 placed` when S1 was already through, had unknown price, or
was refused by the AddOn build gate. The helper now returns whether placement
was registered; only that result closes the pass and retires other scenarios.
All three cases pass. A separate dispatch regression confirms registration
still commits the account after an ambiguous send failure. Existing receipt,
slot and fast-rejection tests pass (`/tmp/nofx-stop-refusal-after.log`).
The limit-path behavior after an ambiguous send remains a separate review item.

## Structural prompt and advisory parity

[A] After the owner reported the usage limit reset, the previously blocked
regressions ran and reproduced a universal ATR-floor prompt and false legacy
R:R/stop warnings for a reject arm. Prompt contract and facts now separate
structural reject fades from non-reject legacy floors; omitted arms do not grant
an AI-route bypass. Reject feasibility is left to composed-geometry admission,
not authored-price legacy warnings. Focused prompt, warning and class45 checks
pass (`/tmp/nofx-structural-prompt-after.log`). The existing heading and historical
reject-warning assertions were updated; legacy floor arithmetic remains tested.

## Second full Go suite and protection shape

[A] The full suite at710ea1c8 completed with two failing tests: the historical
no-one-setup-reference scan and source coordinates in SYSTEM-MAP. The former now
exempts only the AST-bounded advisory ArmFeasibilityWarnings function; the map
golden stays byte-identical. Updated coordinates and focused guard tests pass.

[A] Four protection adjudicator regressions reproduced an `-sl` name overriding
wrong side, wrong type, missing action and missing type. Known protection now
requires the order shape; incomplete named live stops remain UNKNOWN. Focused
protection, short-side and map-reference tests pass in
`/tmp/nofx-protection-shape-after.log`; map golden/prompt tests pass in
`/tmp/nofx-map-guard-after.log`. These are synthetic book tests, not a live
unprotected-position incident. Final combined suite remains due.

## Missing permission and inherited authorizations

[A] A synthetic permission-facts panic reproduced an old authorization reaching
the loopback wire while the cycle logged fail closed. Missing verdicts now retire
unplaced authorizations. Retirement query/write/panic errors propagate to the
cycle, stopping new placement. The same production-cycle test with an injected
SQLite retirement failure sends nothing; the existing allowed-versus-declined
scenario test still passes (`/tmp/nofx-unknown-arm-after.log`). Named system-map
coordinates were synchronized and the map reference check passes.

## Current admission covers the placement pass

[A] A second production-cycle regression reproduced an inherited row reaching
the wire after current scenario quality was refused. Production now passes the
IDs of successfully admitted/saved rows into placement. Old rows cannot inherit
permission merely from their stored armed state. A failed refresh does not enter
that set. Focused missing-permission, quality-refusal, valid-placement and split
fixtures pass (`/tmp/nofx-arm-current-gate-after.log`). Direct offline/debug
placement helpers retain their existing explicitly invoked test semantics.

## Frozen tape and session retirement

[A] A tickOnce regression reproduced an unplaced NY authorization surviving
16:05 CT on unchanged bars. Session/news cutoff enforcement now precedes data
cadence skips, using the tick clock. The regression and existing class32
wall-clock scheduling/EOD/T1 tests pass (`/tmp/nofx-wallclock-retire-after.log`).
The test uses an unconnected synthetic SIM adapter and temporary ledger; it
does not claim a real resting broker order was cancelled.

## Order-fill ownership

[A] Source inspection found the authenticated order-fill handler calling a
shared-store query by order ID alone. The repaired path verifies order ownership
and filters fill ownership, and reads history without a running engine. The
production-router fixture covers own order, foreign order and an inconsistent
foreign fill attached to an owned order; all pass. Before repair the fixture
failed because the stopped owned trader was unavailable, so that before run
does not itself reproduce a foreign-data disclosure. No real records were read.
Logs: `/tmp/nofx-order-fills-before.log`, `/tmp/nofx-order-fills-after.log`.
