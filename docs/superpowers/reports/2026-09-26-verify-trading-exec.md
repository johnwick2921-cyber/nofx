# VERIFY-0926 — Trading Pipeline stages 4–6, independent adversarial cross-check (DS-107)

- **Lane:** DS-107 · **Branch:** `verify/0926-trading-exec` · claim `da880c39`
- **Base:** `04ae1c2f` (live boot 3, `git log -1` = `04ae1c2f RELEASE 6cac1b89 …`)
- **Input under check:** DS-106 `docs/superpowers/reports/2026-09-26-audit-trading-pipeline.md` @ `c3ee15ca` (read via `git show origin/audit/0926-trading-pipeline:<path>`), draft PR #229.
- **Method:** worktree `/home/hoang/nofx-ds-107-verify0926` @ 04ae1c2f, READ-ONLY (L17). DB copy `v.db` (`sqlite3 -readonly .backup`), live log read-only. No tests (box loaded, L16); `go build ./...` + `go vet ./...` green at my head.
- **Verdict vocabulary:** CONFIRMED / REFUTED / PARTIAL — per claim; NEW for findings DS-106 does not have. Every check names file:line at 04ae1c2f and tier.

## 1. Admission chain — re-derived from code (task item 1)

**CONFIRMED — the chain DS-106 lists is the chain the code runs, in that order.** Re-derived from `trader/entry_admission.go` `admitChain :169`:

1. picture-only: trader running + day-plan on `:172-184`
2. `feed_down` ALL `:196` (default-ALLOW until a `feed_status` frame — `:190-195` comment)
3. dead-man `:206`
4. A4 freeze `:215`
5. boot integrity `:224` (outranks everything — `TradingRefused()`)
6. `stop_until` pause `:238`
7. maintenance hold `:249`
8. contract-roll `:260`
9. consecutive-loss: decision+agent `:269`; arm+picture `sessionRiskGateAt` `:281`
10. last-entry cutoff ALL `:296`
11. session gate + force-flat: decision+agent `:305/:315`; arm+picture `CMEClosedReason` `:324`
12. plan-mode: decision+agent `:335`
13. approval ALL `:345`
14. re-entry cooldown arm+picture `:357`
15. EntryGate switch `:370-408`: decision `entryGateForDecisionAt :376` · agent bracket-fail-closed `agentBracketRefusal :394` + same gate `:399` · picture `pictureEntryGate :408` · arm at authoring (`armed_executor.go:799`, G1 same-pass placement).

EntryGate leg order re-verified in `trader/entry_gate.go`: leg D daily loss `:165` → 0 plan-mode `:195` → 1 bias `:213` → 2 scenario-direction `:222` → 3 invalidation (arm) `:237` → 4 shadow `:253` → 5 R:R `:280` → 6 min-SL `:300` → 7 one_open_position `:321` (any open refuses any open, exit legs exempt) → NO-CHASE last `:333`.

Call sites re-verified: decision `auto_trader_orders.go:182` · arm `arm_admission.go:68` · agent `entry_admission.go:457` · picture `picture_htf_evaluator.go:719`. **No path skips admission.** Closes are not admissions (close half of feed gate only, `auto_trader_orders.go:160-176`); Emergency Flat / drawdown monitor bypass admission BY DESIGN as exits (`auto_trader_loop.go:870-875`). [A]

**NEW check DS-106's table does not name explicitly — `max_trades` coverage:** decision+agent enforce it inside the session gate (`sessionEntryBlockedT1At` → `sessionTradeCapBlocked` `trader/auto_trader_session.go:60,88-110`; count scoped to the active session INSTANCE and the active NT account, cap 0 honored); arm+picture get it via `sessionRiskGateAt` (per-session trade cap, `entry_admission.go:38,279` comments). All four paths capped. One boundary: when the Day-Plan master is OFF, `sessionEntryBlockedT1At` returns early (`:47-49`) — the cap is dormant by design (the plan is the gate). PARTIAL/NOTE-grade; recorded for completeness. [A]

## 2. P0 hunt (task item 2)

Hunted, in order, from code + live DB + log:

- **Trade when it must not** (outside session / no-trade band / paused / over max_trades / one_open_position / non-SIM): every candidate path re-traced above; each refusal passes `admitRefuse :120` (always logged + `IncGateBlock`). SIM-only: `isAccountTradeable` unchanged (not re-touched, L11). **No P0.** [A]
- **Trade twice** (latch races / duplicate arm / reconnect replay): all three entry sends take the entry latch (`tcp_trader.go:568,697,794`; wired `entry_latch_wiring.go:27`; stale/absent book = refusal). Duplicate arm: terminal-row laws + `StateCancelPending` refusal in `armed_orders.go` (verified in my system audit). Reconnect replay: see NEW-P2 below. [A/B]
- **Naked position**: `reconcileProtectionAt` `trader/protection_reconciler.go:273`, P0 alert `🚨 UNPROTECTED POSITION` `:346`, standalone `PlaceProtectiveStop` `:355-368`; bracket-on-fill is C#-side. **No P0.** [A]
- **Lost/mispriced close**: #226 ordered worker (`ordered_exec.go:87,127,150,221`; never drops) + #227 busy retry (`close_sync.go:322-329`, 4 tries → `putPricedClose`+`SavePendingExit`) + `RetryPendingNT8Exits` replay; `pnl_corrected` from receipts (corrected-column law). Live proof: the single receipt row is applied and row 619 carries non-NULL `pnl_corrected −32.0` (see §4). **No P0.** [A]
- **Flatten without a ledger explanation**: row 618 IS the scar — flattened by NT8 with no close frame and no netting fill; reconcile wrote it `unresolved` with a named reason (see §4), never a fabricated price. 619 flattened by EOD with a receipt row. **No P0.** [A]

**NEW-P1 (borderline P2) — reconnect replay of an `attempted` entry frame.** `provider/ninjatrader/tcp_server.go:2576-2605`: on a mid-write conn death the entry that STARTED writing (`attempted=true`, bytes may have reached the AddOn) stays in the re-queued tail and is re-written on the next accept's `flushPending`. The B3 order-guard runs at placeEntry time (`tcp_trader.go:526-565`), NOT at flush time; the re-flush sends the SAME seq (assigned once at enqueue, `:1277`). If the first frame delivered and placed, the second identical frame is a duplicate the Go side cannot see (echo verify checks INBOUND echoes, not outbound replays; the C# AddOn captures (trader_id, seq) for echo `VLTraderTCPClient.cs:836-841` but no inbound-dedupe-before-place was found in the source). Window: mid-write conn death + reconnect < `TCPStaleSignalAge` (60s) + same order key. Mitigations: dead-man blocks NEW entries after a gap (queued ones still flush by design), reconcile + one_open_position nets eventually converge. DS-106's Stage 6 names fill/close dedupe but not this entry-replay window. **NEW finding; held at P2** because the live incident history shows no instance and the seq identity is the designed dedupe signal. [B: Go side read; C# side source-read but binary not inspected]

## 3. DS-106 §4 "refuted" list — re-checked (task item 3)

1. **EOD/T1 sync cancel vs a row that just FILLED** — CONFIRMED (their refutation is sound). `cancelArmedOrdersSyncWith` (`armed_executor.go`) counts filled rows separately (`filled := 0` ~`:2652`), cancels by signal id, writes `cancel_pending` (never `cancelled`) when the wire is missing (`:2700-2705`, class 81 comment), and drains through the same `onArmedOrderUpdate`. [A]
2. **Agent door without EntryGate** — CONFIRMED. `admitAgent` runs the full `admitChain` + `agentBracketRefusal` (fail-closed stop/target/side/ATR) + `entryGateForDecisionAt` (`entry_admission.go:385-406`), above `policy.CanExecuteTrade` at `agent/tools.go:2745`. [A]
3. **Silent refusal** — CONFIRMED. Every refusal passes `admitRefuse :120` (always logs + `IncGateBlock`; arm/picture deduped on (path, key, CLASS), never on the moving reason). [A]
4. **Double-open latch** — CONFIRMED. Latch on all three sends + wiring + stale/absent book refusal (`tcp_trader.go:568,697,794`; `entry_latch_wiring.go:27`). [A]
5. **Row 619 close misprice** — CONFIRMED. Live receipt row: `nt8-exit-v2-b31f… Sim101 MNQ SHORT limit 1.0 @ 30909.25 applied=1 position_id=619`; DB row 619 `exit_price=30909.25, close_reason=sync, realized_pnl=−32.0, pnl_corrected=−32.0` (receipts-sum note). [A]

## 4. Rows 618 and 619 — my own trace (task item 4)

From the DB copy `v.db` + `data/nofx_2026-09-25.log` (timestamps in the table below are epoch-ms → CT):

| row | side | entry | exit | status | close_reason | pnl | pnl_corrected | note | entry CT | exit CT |
|---|---|---|---|---|---|---|---|---|---|---|
| **618** | SHORT | 30923.5 | **0.0** | CLOSED | `unresolved` | 0.0 | **NULL** | "class-27: NT8 flat, no close frame, no netting fill — exit price UNKNOWN" | 08:55:21 | 08:58:38 |
| **619** | SHORT | 30893.25 | 30909.25 | CLOSED | `sync` | −32.0 | **−32.0** | "NT8 execution receipts: sum(price delta × actual contracts × futures point value)" | 14:45:22 EOD | 14:45:22 |

- 618 = the pre-#227 lost-close (08:55 CT incident window); exit 0.0 + NULL pnl is the corrected-column law applied (unresolved excluded, COUNT shown) — **matches DS-106's census exactly; their P2 about the invisible loss is accurate**: the AI-context line at 08:50:13 says "0 unresolved, rendered without P&L" [A: log line 29562].
- 619 = the EOD flat whose close rode the #227 receipt path; −32 = (30893.25−30909.25) × 1 contract × $2/point, receipt applied, non-NULL — **matches their §4 item 5**. [A]
- `nt8_exit_receipts`: exactly 1 row, applied=1, 0 pending — "never parked since boot 3" CONFIRMED. [A]
- Minor drift: DS-106 cites "log line 29808" for the unresolved-context line; the same text sits at line 29562 in the file as read now (log grows; the cited line number is stale). NOTE-grade, content CONFIRMED.

## 5. NEW findings (not in DS-106's report)

- **NEW-P2** — reconnect replay of an `attempted` entry frame (full detail in §2).
- **NEW-NOTE** — `max_trades` enforcement path asymmetry worth recording: decision/agent cap lives INSIDE the session gate (dormant when Day-Plan master is OFF), arm/picture cap inside `sessionRiskGateAt`; no single cap site. Design-consistent today (the plan IS the gate when enabled); a future no-plan mode would need its own cap.
- **NEW-NOTE** — DS-106's log-line citations are content-correct but line-stale (29808 vs 29562): the live log is append-only; line-number citations in audit reports should quote the matching text, not the number.

## 6. What I could NOT verify

- C# AddOn duplicate-frame behavior (binary not inspected; only source read) — the NEW-P2 hinges on it and is graded P2/[B] for that reason.
- `-race` claims: no race run (box loaded, L16) — DS-106's latch/worker race-free claims rest on code reading + prior CI evidence.
- `.env` values (L12): key names only.

## Counts

CONFIRMED 12 · REFUTED 0 · PARTIAL 1 (max_trades boundary) · NEW 3 (1×P2, 2×NOTE).

*DS-107 · read-only verification · zero code changes · addendum for the CTO's consolidated report.*
