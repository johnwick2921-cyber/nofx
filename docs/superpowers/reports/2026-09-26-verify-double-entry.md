# VERIFY-0926 — DOUBLE ENTRY AFTER RECONNECT (DS-102)

- **Lane:** DS-102 · **Branch:** `verify/0926-double-entry` · claim `14492a85`
- **Base:** `04ae1c2f` — `git log -1`: `04ae1c2f RELEASE 6cac1b896bdb364642227853d27814a63a55be39 — boot 3 …`
- **Claim under test (DS-107 NEW, CTO P1-candidate [B]):** `provider/ninjatrader/tcp_server.go:2574-2606` marks a signal `attempted` when a write fails mid-flush, re-queues it, and re-sends it after reconnect; `ninjascript/VLTraderTCPClient.cs` has no signal_id-already-handled guard, so the same entry can reach NT8 twice.
- **Mode:** READ-ONLY (L1–L17). DB copy read from `../nofx-ds-102-verify0926/v.db` (same 04ae1c2f base, `-readonly`); live logs read-only.
- **Verdict: PROVEN — P1, latent, never observed on this box.** Both ends lack the dedupe; the trigger window is narrow but real; the outcome is a doubled position.

---

## 1. Does ANY layer drop the duplicate?

| Layer | Guard? | Evidence @04ae1c2f |
|---|---|---|
| Go pending/flush path | **NO** | `flushPendingReportFor` re-queues `toSend[i:]` with `attempted=true` on a mid-flush write error (`tcp_server.go:2591`), then `closeConn()`. **`attempted` is NEVER READ anywhere on the resend path** — grep shows only the field def (`:677`), the comment (`:672-676`), and the set at `:2591`. On the next accept, `acceptLoop` calls `_ = s.flushPending()` (`:1574`) and the attempted frame is written again, unguarded. |
| Go stale-age lease | **YES — the only Go-side gate** | `checkSignalAge` (`:2514-2526`) drops frames whose age exceeds `TCPStaleSignalAge = 60s` (`:46`) vs the payload's ORIGINAL timestamp — *"Neither enqueue nor retry is allowed to renew its lease"* (`:2576`). A reconnect re-flush within 60s of the original send therefore PASSES. The C# AddOn auto-redials in seconds, so the re-send is normally inside the window. |
| Go entry latch / one_open_position | **NOT RE-RUN** | They gate PLACEMENT (`runArmedPlacementAt` / `PlaceLimitEntry` / `PlaceStopEntry` call `SendSignal`), not the flush. The resend is queue-level frame replay; no placement gate executes again. The ledger row is already `place_pending` with its signal id, so no second placement pass re-arms it either — the duplicate comes ONLY from the frame replay. |
| AddOn `signalIdentity` / dicts | **NO — overwritten, no skip** | `HandleSignal` stamps `signalIdentity[signalId] = new SignalIdentity{…}` unconditionally (~`VLTraderTCPClient.cs:852`) and later overwrites `pendingBrackets[signalId]` / `workingEntries[signalId]` (~`:1104-1107`). No `ContainsKey(signalId) → skip` anywhere before `CreateOrder`. The stale check the AddOn runs (60s, same original timestamp) also PASSES on the resend. |
| AddOn order naming | **NAME = signal_id** | `submitAccount.CreateOrder(instrument, …, string.Empty /* oco */, signalId /* name */, …)` (~`:1097-1099`) — every entry order is NAMED with the signal id. |
| NT8 name uniqueness | **NO (tier [B])** | NT8 `Account.CreateOrder` + `Submit` does not dedupe by order NAME; each `CreateOrder` is a new Order object and `Submit` sends it. Two submissions with the same name produce two orders (names are client-side labels). Cannot run NT8 here, hence [B]; no name-collision rejection exists in the NT8 API model the AddOn uses. |
| NT8 SIM rail | **irrelevant** | The SIM-only guard blocks live-account damage, not the double itself. |

**Chain:** mid-flush write error with the full frame already delivered (TCP: data delivered, then RST before the write returns) → Go marks attempted + re-queues + closes → AddOn redials within seconds → accept-loop re-flush (age < 60s, original ts) → AddOn re-handles (no skip, identity/bracket overwritten) → second `CreateOrder(name=signal_id)` + `Submit` → second working order in NT8. **No layer between Go's re-flush and NT8's order book drops it.**

## 2. Worst case per entry type

- **Market:** the first order fills ~instantly; the duplicate fills too → position doubles to 2 contracts, and `SubmitBracketOnEntryFill` places a SECOND SL/TP pair (the first pair already exists). Worst case: doubled size AND doubled brackets.
- **Armed limit:** two resting limit orders at the same price, same name → both can fill on a touch → doubled position. (Both tracked under one `workingEntries[signalId]` key, so the AddOn's own book sees only the second — a cancel would cancel only the tracked one.)
- **Stop-entry:** two StopMarket orders at the same trigger → both fire on the breakout → doubled position.
- Common to all three: the duplicate fill echoes the SAME `signal_id`, so Go-side ledgers/echo-verify see two fills for one signal id — accounting divergence in addition to size.

## 3. Has it ever happened? — NO

- Live logs: `grep -h 'flush signal failed' data/nofx_*.log` = **0 occurrences** (the mid-flush failure path has never fired on this box).
- DB copy: `armed_orders` — no `signal_id` with >1 row in `filled/working/accepted` (query returned empty). `trader_fills` (460 rows) — no duplicate `exchange_order_id` except `"Close"`×9 (the close-position name, not entries) and NULL×391. `nt8_exit_receipts` — no duplicate `signal_id`.
- AddOn-side "submitted entry signal_id" lines live in the NT8 log, not Go logs — 0 in Go logs is expected and proves nothing by itself; the Go-side evidence above is the negative record.

## 4. VERDICT — P1 (latent structural)

- **PROVEN mechanically** at both ends with file:line reads [A]; the only uncertain step (NT8 name dedupe) is [B].
- **Never observed** [A] — 0 mid-flush failures in the box's entire log set, so the trigger (partial delivery + reconnect + re-flush < 60s) has never fired here.
- Upgraded from NEW to **P1**, agreeing with the CTO's candidate: outcome is a doubled position with doubled brackets, and no layer stops it. It stays P1 rather than P0 because the trigger needs a precise TCP failure shape AND the 60s window, and the box's history shows zero occurrences.

## 5. SMALLEST FIX OPTIONS

- **(a) AddOn dedupe by `signalIdentity` (primary, recommended).** At the top of `HandleSignal` (after the parse/stale guards, before `CreateOrder`): `lock (signalMapLock) { if (signalIdentity.ContainsKey(signalId)) { …echo the CURRENT state for this signal (ack/fill/reject) and return… } }`. `signalIdentity` is only populated when a frame ARRIVED, so a first delivery that never reached the AddOn is unaffected; the entry is removed on the position lifecycle's end (`:1441`), so a legitimate re-send of a completed signal is not blocked forever. Risk: LOW — the only in-window re-send of an identical signal_id IS the attempted-frame replay (Go generates a fresh uuid per placement); echoing state instead of resubmitting is exactly the desired behaviour. Do NOT key the skip on `pendingBrackets`/`workingEntries` alone — both are removed on fill (`:2049, :1507, :2337`), which would let a post-fill duplicate through.
- **(b) Go reconcile-before-resend of `attempted` frames (defense-in-depth).** Before flushing a frame with `attempted=true`, consult the fresh `order_snapshot` for an order named `signal_id`; if present, drop the frame and settle the ledger row from the book instead. Risk: snapshot freshness/latency — a miss while the AddOn is reconnecting would either suppress the only delivery (missed entry) or still resend (no change). Needs a fail-closed decision (suppress-and-reconcile) and the snapshot to be REQUIRED-fresh before the attempted frame is released.
- **(c) Go settle-don't-resend (simplest, trades a double for a miss).** On a mid-flush failure, drop the attempted frame and let the ledger row settle via reconcile (absent-from-book → cancelled; present → working). Risk: if the bytes never arrived, the entry is LOST for that plan pass (armed legs can re-arm next pass; market legs are one-shot). Safer for money than (b)'s double, worse for entry completeness than today.
- **Recommendation:** ship (a) now (one guarded block in the AddOn, no wire-schema change), and (b) as the Go-side follow-up in the next deploy wave.

## 6. WHAT I DID NOT DO
- No code changes, no tests, no build (L17 read-only).
- Did not run NT8 (name-dedupe step is [B] by necessity).
- Did not review the W2 re-queue's other consequences beyond the duplicate-entry claim.
