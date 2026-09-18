# PR #117 Slicing Review — D101-1 (DS-101)

- **Dispatch:** D101-1 (CTO, 2026-09-16 22:12Z). Lane DS-101.
- **PR:** `fix/repo-audit-control-boundaries-20260913` → dev. Head `8d860b77`, 58 commits, **163 files, +7485 / −1753**.
- **Bases:** worktree cut at `054e97e5`; dev tip at test time `f6465143`; `git merge-base origin/dev pr117` = `4b56e9dc` (PR is ~3 days behind dev; dev has since gained ~60 dev-only files).
- **Method:** all experiments in worktree `/home/hoang/nofx-ds101`, throwaway merges aborted after, zero writes outside the worktree. `go test ./...` was run at the clean merged head.

## VERDICT (short)

**The PR is large but NOT too big to merge as a whole — it merges into today's dev with only 2 trivial text conflicts and the merged head is fully green (build rc=0, full `go test ./...` rc=0).** Slicing is optional; if sliced, 2 of 7 slices (`toolchain`, `docs`) can land first alone. Do NOT split the exec-order slice — it is one atomic safety change across Go + C#.

---

## 1. Slice table

| Slice | Files | +/− | Standalone on dev? | Cross-slice deps |
|---|---|---|---|---|
| (a) toolchain | 4 (go.mod, go.sum, docker/Dockerfile.backend, web/package-lock.json) | +77/−43 | **[A] YES — builds rc=0** (go1.26.8 auto-downloaded) | none |
| (b) exec-order | 77 (trader/ 33, trader/ninjatrader/ 15, store/ 14, provider/ninjatrader/ 10, ninjascript/ 5) | +4490/−909 | **[B] Not testable via dir-checkout** (see §2 caveat); full-PR and merged-head both green [A] | internal: store.`NT8ExitReceipt` ↔ trader/ninjatrader; wire ↔ C# (lockstep) |
| (c) auth-owner | 18 (api/ 12, agent/ 6) | +591/−104 | merged-head green [A] | `kernel.PlanChainTradeDate`, `SessionRunnable` (already on dev) |
| (d) planner | 9 (kernel/) | +109/−11 | merged-head green [A] | none new — `TradingRefused` wiring lives in exec-order, the fn itself is dev-side `kernel/boot_integrity.go:173` [A] |
| (e) frontend | 47 (web/src) + web top | +1892/−664 | **not locally compiled** (no node_modules in worktree [A]) | `guide/types.ts` conflicts with dev (§4) |
| (f) docs | 2 (AUDIT-CHECKLIST.md, SYSTEM-MAP.md) | +257/−7 | n/a | AUDIT-CHECKLIST conflicts with dev (§4) |
| (g) ci-misc | 6 (main.go, main_boot_order_test.go, 3×workflows, branding) | +69/−15 | merged-head green [A] | main.go reorder is self-contained |

---

## 2. Compile / merge experiments (all in worktree, reset after)

- Baseline dev tip build: **rc=0** [A].
- Full PR117 head build: **rc=0** [A]. (`go.mod` says `go 1.26.8`, no `toolchain` directive; `GOTOOLCHAIN=auto` downloaded go1.26.8 — local `go version` inside the module printed `go1.26.8` [A].)
- Toolchain slice alone (pr117 `go.mod`+`go.sum`): **rc=0** [A]. `go.mod` without `go.sum` fails on missing entries — the two must travel together [A].
- **Slice-by-directory-checkout is unreliable and I retract it**: checking out `provider/ninjatrader` from pr117 dropped dev-side additions (`history_at_subscribe.go`, `histAtSub` field) and produced a false `s.histAtSub undefined`. The real three-way merge keeps both sides: merged `tcp_server.go` has dev's `histAtSub`/`histReplay` (lines 63–64) AND pr117's `entryReceipts` (line 190) [A]. One earlier "merged build green" and one "manager build failed" were both from this tainted method; both are invalidated.
- **Clean merged head** (pr117 → origin/dev f6465143): `go build ./...` **rc=0** [A]; **full `go test ./...` rc=0** [A] — every package green (store 157.8s, trader 319.2s, trader/ninjatrader 62.6s, provider/ninjatrader 20.2s).

## 3. Risk per slice (SIM-only live process)

**(b) exec-order — the order path. HIGH attention, but the changes are safety-positive:**

- `main.go` — boot integrity assertion **moved earlier**, before `LoadTradersFromStore` [A]; and `trader/ninjatrader/tcp_trader.go` adds `kernel.TradingRefused()` refusal at the top of **all three entry call sites** `placeEntry`, `PlaceLimitEntry`, `PlaceStopEntry` [A]. New entries are blocked on boot-integrity mismatch. Defense-in-depth.
- `isAccountTradeable` (SIM-only gate) is **NOT modified** — appears only as diff context [A]. SIM-only law intact.
- `store/armed_orders.go` — fixes the manual-cancel-wins guard: old condition `existing.Version == row.Version && !IsBootSweepReason` would re-arm any same-version row; new code requires `IsTerminalArmState(existing.State)` before the version check, plus a loud journal for boot-sweep re-arms [A]. This closes the E7 re-place loop.
- `store/nt8_exit_receipt.go` (NEW) + `trader/ninjatrader/close_sync.go` — completed-exit receipts deduped by account/symbol/side/`exit_order_id`; refusal paths for symbol mismatch, missing identity, uncommitted receipts; snapshot fence before flat is admitted [A].
- `trader/auto_trader_clock.go` — limit-flatten lifecycle: copies caller-owned state before timer capture (closure-capture bugfix), `CancelStopOrders` only after `closeMarket()` succeeds, `limitFlattenStopped` guard [A].
- **Wire (lockstep):** `provider/ninjatrader/tcp_framing.go` (45 lines) + `ninjascript/VLTraderTCPClient.cs` (632 lines) + protocol doc. **Additive only** — protocol stays v3, C# emits `position_close.exit_order_id`, Go adds internal `json:"-"` flags. Legacy frames remain accepted [A, protocol text]. **Deploy coupling:** Go is backward-compatible without the C#; exit-receipt identity requires the C# copy→F5→full-restart dance. **C# changed ⇒ must follow the NT8 deploy procedure, never Go-only in a cutover that relies on the new dedup.**
- Flat gate: `GetOpenOrders` wiring comment unchanged; no flat-gate weakening found in the diff [B].

**(c) auth-owner:** `api/plan_mutation_session.go` — wrap-aware chain-date + `SessionRunnable` gate before plan mutation; gaps are refusals, never nil-session fallthrough [A]. `agent/request_runtime.go` — `stateOwner()` prevents shallow-copying locks [A]. Ownership tests added across api/agent. Low risk.

**(d) planner:** small — `levels_swing.go` 9 lines, `plan_doc.go` +5, `planner_prompt.go` 2, pin/regression tests. No ECONOMICS changes to R4 auto-correct (`scenario_economics.go` is dev-side, not in PR) [A]. Low risk.

**(e) frontend:** largest rewrite — `AgentChatPage.tsx` −460, new `lib/agentStream.ts` +424 (streaming), `AuthContext` +89, `AdvancedChart` +209, PositionHistory +87. Not compiled here (no node_modules in worktree [A]); CI workflows touched (3 files, +2/−2 each). Medium risk, **must pass web build in CI** before cutover; no live-order impact (read-only UI + chat).

**(g) ci-misc:** `main.go` reorder is boot-path; covered by `main_boot_order_test.go` (+47). Low risk, but it IS the boot line area — release ordering must be re-verified at cutover.

## 4. Conflicts (merge-tree + real throwaway merge agree) [A]

1. `docs/superpowers/AUDIT-CHECKLIST.md` — both sides appended (~160 dev vs ~233 PR lines). Resolve = union of both.
2. `web/src/guide/types.ts` — `GUIDE_BUILT_REV` stamp. Dev stamped a newer release rev; PR stamped its internal candidate `40ed5d95`. Resolve = re-stamp after merge (GUIDE CONTENT LAW: the merged binary's rev).

Both trivial, neither touches code.

## 5. Toolchain slice

- Installed: `go1.25.13`. Dev `go.mod`: `go 1.25.13`. Running binary `nofx-bin`: built with go1.25.13, mod rev `7e87a375acf7` [A].
- PR: `go 1.26.8` + x/crypto 0.53.0→0.56.0, gnark-crypto 0.19.0→0.19.2 [A].
- **Can the toolchain slice merge first and alone? YES, with a caveat.** It builds green here [A] because `GOTOOLCHAIN=auto` auto-downloaded go1.26.8. Caveat: every subsequent local build then requires go1.26.8 (network on first use; offline boxes break). Deploy boxes must pre-install go1.26.8 or set `GOTOOLCHAIN=local` with it present before this lands.

## 6. Recommendation

1. **Merge order if slicing:** (a) toolchain first (alone, green) → (f) docs → then **all of (b)+(c)+(d)+(g) as ONE merge** (they are one safety wave; splitting exec-order internally buys nothing and breaks compile standalone) → (e) frontend last, gated on CI web build.
2. **Simpler and defensible: merge the whole PR as one.** Merged head is build+test green [A], conflicts are two trivial stamp/doc unions. The size is mitigated by the green merged-head evidence above.
3. **Do not cut over this PR without:** (i) resolving the 2 conflicts (stamp + checklist union), (ii) re-running the full suite at the final merged HEAD (canon: green branches ≠ green merged — my run predates any new dev commits), (iii) NT8 C# deploy procedure for `VLTraderTCPClient.cs` (copy → F5 → full restart), (iv) a flat/safe window per deploy law.
4. **Close/drop nothing** — no slice is obsolete or duplicated by dev.

## Evidence base

`git log -1` bases: worktree `054e97e5`; dev tip at test `f6465143`; PR head `8d860b77`; merge-base `4b56e9dc`.
All [A] claims above were run in `/home/hoang/nofx-ds101` on 2026-09-16: baseline build, full-PR build, toolchain-slice build, clean merged build, clean merged `go test ./...` (rc=0), merge-tree + throwaway merge, object greps. [B]/[C] are labeled inline.
