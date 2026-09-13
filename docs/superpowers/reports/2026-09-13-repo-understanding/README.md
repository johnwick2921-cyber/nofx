# Repository understanding and verification — work in progress

Source base: `63968be62e44db2fb07a92883e02127b9064b0be`. Audit branch: `docs/repo-understanding-20260913`; claim: `8c7bc6be8f1a425067135b612ba62933e1fc3da1`.

The owner requested a detailed, accurately traced, approximately 30-agent repository review. This directory preserves the evidence and coverage rather than equating an index with understanding. The later instruction to repair confirmed defects is recorded; any repairs require a separately scoped branch, independent review and meaningful verification. This audit does not authorize deployment or account/settings changes.

## Scope and current status

- 3,726 tracked files inventoried by path, SHA-256, byte and line count.
- 1,049 first-party source files (250,582 lines) assigned once across 28 source reviews, followed by two independent cross-boundary reviews.
- Tests, fixtures and historical non-code artifacts are inventoried; relevant ones are followed. This is **not** a claim that every test or historical artifact was manually read.
- Go AST inventory: 1,292 files parsed, 10,269 functions/literals and 78,157 syntactic calls; zero parser errors. These are syntax inventories, **not type-resolved call edges or human reading**.
- `orientation/` holds initial subsystem reviews with explicit read ledgers and unresolved concerns. Their findings require root/cross-review validation.
- **30/30 scoped reviews reported**, covering all 1,049 assigned source files / 250,582 lines. The two independent cross-boundary reviews examined repair revision `99a06543`; subsequent repairs require their own verification.
- `review-plan.json` tracks assignments; `coverage-validation.json` checks source hashes, full line ranges, and named Go/TypeScript/JavaScript function notes. It does not certify semantic understanding or runtime behavior.

## Historical maps

The recovered Understand Anything graph is dated July 10, 2026 at `7a8adce0043729950f3d7cbfaf30810c8304709a`: 3,121 nodes, 9,588 edges, 11 layers and 15 tour steps. Its SHA-256 is `23aa686864d6e1af4e6b43ae52856175356d07e8baf066fe36d0e90418a2c43f`. CGC responds with 781 indexed files and 5,095 functions, but exposes no index revision/date. Both require comparison with current source. Neither establishes current completeness.

## Evidence contract

[A] directly read/run/observed; [B] inference from cited evidence; [C] hypothesis. Static concerns, offline reproductions and runtime incidents are separate categories. Each final trace will name source revision and file/function/line, actual test results, and remaining uncertainty. A passing suite is not proof of every behavior. No fabricated coverage, caller resolution, runtime observations, profitability or universal safety claims.

## Repair lane — still in progress

`fix/repo-audit-control-boundaries-20260913` holds isolated code repairs. Current
checkpoint `7b2eb894` includes authenticated object ownership/deletion, Q&A
empty/trader-scoped context, validation-before-save, boot refusal on all NT8 entry
methods, same-version terminal-arm retirement, session mutation chain dates,
broker-book absence/freshness/send truth, history-channel teardown locking, and
boot cancellation settlement. The separate repair report names reproductions
and test limits. These changes are not deployed and do not alter owner settings.
The first full Go suite completed with three store-test failures; all other packages passed. Corrected authorization fixtures pass focused reruns. Later account and chat repairs have focused tests; final combined verification remains pending.

The latest execution limitation and exact resume queue are recorded in [CHECKPOINT.md](CHECKPOINT.md).

Resumed after owner reported usage reset: structural prompt regression reproduced and fixed at repair710ea1c8; focused kernel checks passed. Second full Go run is in progress.
