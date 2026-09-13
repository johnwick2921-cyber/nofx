# NT8 account routing and cancellation lifecycle repair

Branch `fix/repo-audit-nt8-lifecycle-20260913`; isolated worktree `/tmp/nofx-audit-nt8-repair-20260913`; source baseline `63968be62e44db2fb07a92883e02127b9064b0be`; pushed claim `50b32f84`. No deployment, NT8 process control, AddOns-directory writes, runtime tests, account/settings changes or network trading. Parent owns integration, guide/checklist updates and release coordination.

Rules consulted: `git log -1 --oneline -- docs/superpowers/CLAUDE-canon.md` → `07b53e65 cleanup batch 1 — B5 the mutation harness that cannot fake a verdict, B6 the worktree-add check`; audit checklist → `dfda15e1 test: isolate session clock fixtures from weekly backfill workers`. The dispatch explicitly chose immutable baseline63968be. No added mandatory per-trade risk cap.

## Result

[A: source and offline tests] Explicit account names now determine both close and emergency protective-stop routing. An unknown name refuses, never falling back to active account. The selected target must pass existing SIM, connected and declared-session allowlist checks. Absent account uses the active account for legacy callers, with the same guards; the lossy root-to-account hint no longer chooses an execution target.

`TryResolveHeldPosition` reads the target account's actual positive, non-flat holdings. Bare root with two held expiries refuses ambiguity. A qualified contract selects that exact holding; closing an old expiry uses its actual Instrument instead of resolving today's front month. Limit close quantity must be positive and no greater than held quantity; side must match when supplied. Protective-stop side/quantity must agree with the actual position; fractional/oversized quantity refuses. Limit Submit exceptions never trigger an additional market close: accepted-then-thrown submission is ambiguous, not proof of failure.

[A] Entry cancellation retains working-entry and deferred bracket state while Cancel is pending or throws. Error ACK is returned when Cancel throws; successful ACK means request receipt, not confirmed cancellation. On terminal Cancelled/Rejected, cumulative filled quantity is protected before deferred state retires. Delayed smaller partial events never shrink a bracket. Filled entries also retire pending bookkeeping. An entry Submit exception after the call began retains correlation and emits `signal_submit_unknown` rather than fabricating a rejected fill.

[A] Bracket submission reserves the signal's pair identity before CreateOrder/Submit. Synchronous reentrant fills increase desired cumulative quantity; after submission the existing pair catches up via Change. Ambiguous Submit retains the pair/order identities, so later fills cannot create a second pair. A separate amendment-in-progress guard prevents synchronous Change callbacks from recursively requesting the same amendment. Deferred quantity adjustment waits for both legs to be Working/Accepted; no cancel-and-replace or second OCO pair is introduced.

[A] Bracket cancellation retains tracked legs until both have terminal Filled/Cancelled/Rejected evidence. Cancel throwing or returning does not retire tracking. Flat cleanup now scopes by exact contract and account rather than root, so flat MNQ June does not cancel September protection in the same account. A completed limit order does not prove the entire position is flat; immediate cleanup checks actual remaining holding, and a later PositionUpdate performs the exact-contract flat sweep. Rejected/partial exits preserve protection. Account selection retains the old account's OrderUpdate subscription because its outstanding order maps remain active; termination still removes all subscribed handlers.

## Verification

[A] `ninjascript/tests/build_lifecycle_harness.py` extracts 16 actual production method bodies and the actual PendingBracket/PlacedBracket types from `VLTraderTCPClient.cs`. `lifecycle_harness.cs.in` supplies inert account/order/position fakes plus assertions. Tests execute production handlers and callbacks, including recording CreateOrder/Submit/Change/Cancel/Flatten. No NT8 assembly is loaded by this behavioral harness and no socket/account is touched.

28 offline C# assertions passed, including:

- Explicit/unknown account routing, held old expiry, two-expiry refusal and qualified selection.
- SIM/disconnected/not-allowlisted refusals, protective side/size and GTC shape.
- Ambiguous limit Submit without a second market close.
- Cancel failure/request retention; terminal cancellation with partial fill; fill winning cancel; terminal rejection with filled quantity; unfilled retirement; out-of-order smaller partial.
- Accepted-then-thrown bracket Submit, synchronous fill during Submit, synchronous live callback during Change.
- Rejected/partial/smaller completed exit preserving remaining protection.
- Failed/requested bracket cancellation retaining tracking; one-leg then both-leg terminal cleanup.
- Exact-expiry cleanup isolation in one account.

[A] All five AddOn source files compiled successfully using installed Windows .NET SDK9.0.307 Roslyn against the installed NT8 Core/Gui and .NET Framework reference assemblies. Input response file `/tmp/nofx-csharp-verification/repair.rsp` derives from root's baseline.rsp, substituting only source worktree and output DLL. Output `/tmp/nofx-csharp-verification/nofx-addon-repair.dll` is verification-only, never deployed. Harness source/exe/response file and exact output are in the same temporary directory (`lifecycle-harness.cs`, `harness.rsp`, `lifecycle-harness.exe`, `lifecycle-results.log`). Compiler returned0 with no diagnostics.

[A] Existing Go AddOn source checks were run offline with `GOCACHE=/tmp/nofx-nt8-go-cache GOPROXY=off`; default cache was read-only on first attempt, corrected with temporary cache. These are supplementary textual pins, not C# execution. All 10 tests in bracket_cancel_addon_test.go and wave_b_addon_slot_test.go passed (ok nofx/trader 0.011s). A first expanded run failed a textual pin demanding literal local names b.Sl/exitOco; those existing local names were restored without changing order slots, and the final expanded run passed. `git diff --check` passed.

## Limits and remaining risks

[B] Fakes verify deterministic event permutations, not NT8's real event scheduler, concurrent broker mutation, connectivity transitions, routing delays, simulation fills or OCO behavior. Actual-assembly compilation validates API/type compatibility, not live integration. A terminal zero-filled cancel followed by a contradictory later positive fill is not inferred as impossible; the harness covers terminal cumulative-positive ordering and cancel-request/fill races, not arbitrary contradictory broker evidence.

[A] An ambiguous bracket submit preserves identity even when creation/submission fails before any live legs appear. This intentionally avoids unsafe duplicate pairs; it does not prove protection exists. Unknown/non-live legs retain desired quantity and require later broker evidence or existing reconciliation/operator handling. Quantity Change acceptance is not terminal broker confirmation; pb.Qty denotes last submitted quantity change, not independently confirmed coverage. Tracking may remain if terminal notifications never arrive, rather than infer cancellation from absence. Standalone emergency-stop duplicate retries and general entry signal replay idempotency are outside this bounded repair.

[A] The legacy allowlist behavior before authoritative declaration remains unchanged, as instructed. The root-only ownership hint remains for historical bookkeeping but is no longer an execution routing authority. Other legacy wire/cache details, close identity propagation, broker position-read atomicity and general account-selection policy are not redesigned here. Guard/account checks can become stale after inspection; broker errors remain observable rather than silently rerouted.

[A] Source build marker remains the baseline marker; this is an unshipped repair, and integration must coordinate AddOn/Go marker and guide revision where required. Root owns guide/checklist integration and merged-head/full-suite checks. No running revision or deployment success is claimed.
