# VERIFY-0926 — STOP-ENTRY SEAM (DS-106, second pair of eyes)

Follow-up to the knobs cross-check (`2026-09-26-verify-knobs.md`, which CONFIRMED P1-1 as a contradiction of the 2026-09-05 ruling). CTO's question: `.env:40-43` says the seam "Stays off until a cancel-confirmation wave lands", then sets `STOP_ENTRY_SEAM=on`; boot-3's cancel path reads `confirm=broker-snapshot` — so has the precondition been met, and is the seam ON by ruling or by drift?

READ-ONLY (L17). Base `04ae1c2f`, worktree `/home/hoang/nofx-ds-106-verify0926` @ branch `verify/0926-stop-entry`, claim `6526bc44`. Logs `/home/hoang/nofx/data/nofx_2026-09-{18..25}.log` read-only. No code changes, no tests.

## 1. Answer in one paragraph

The seam is ON by a **second owner ruling, 2026-09-18 07:52 CT** — not by drift. The ruling is recorded in the stop-entry guide clause (`web/src/guide/content/guards.ts`, merged in `99349544`, 2026-09-18 13:23:23 CT): *"Two owner rulings set it: 2026-09-05 OFF — no stop entry is placed at all until a wave lands that confirms a broker cancel instead of assuming one — and 2026-09-18 07:52 CT ON, once the far side proved its stop-price frames."* [A]. At the flip-moment boot both 09-05 preconditions held on the wire [A]: the cancel-confirmation wave was live (`cancels: confirm=broker-snapshot · slot-guard=on(refuse-on-live|stale)`) and the far side was proven (`addon build_id=2026-09-07-h1 expected=2026-09-07-h1 match=yes`). The residual problem is **stale prose, not a live-config contradiction**: the `.env` comment predates the second ruling, and my knobs addendum's "no later re-enable ruling exists in docs/" was wrong in scope — the ruling lives in the guide, not in `docs/` reports.

## 2. The two rulings

| date | ruling | recorded where |
|---|---|---|
| 2026-09-05 | OFF — "Until a cancel-confirmation wave lands, the AddOn build floor is not a brake the owner will rely on." Reason: `nt.CancelOrder` reports success on a SEND, not on a confirmation; broker once held **nine** working stops for one slot (snapshot id 1664) on a 2-contract cap. | `docs/superpowers/reports/2026-09-05-wave-b-stop-entry.md:5,247-249,541`; `.env:40-42` comment |
| 2026-09-18 07:52 CT | ON — "once the far side proved its stop-price frames" | `web/src/guide/content/guards.ts` (merged `99349544`) [A] |

## 3. Precondition ledger at the flip moment (09-18 07:52:20 boot, `nofx` rev 249ff3a5eea6)

- **Cancel-confirmation wave: LANDED.** Merged to dev 2026-09-06 00:24–06:48 CT (`871083cc` adjudicator → `22af989e` settlement/reconciliation/boot line → `f61b23a3` per-slot invariant → `ea3f33fe`), first-parent of `04ae1c2f` [A]. Rule A20: *"a cancel is proven by the ORDER'S ABSENCE FROM A FRESH BROKER SNAPSHOT"* (`trader/cancel_confirm.go:21-24`). Live at the flip boot: `🧾 cancels: confirm=broker-snapshot · pending=0 · unconfirmed=0 · slot-guard=on(refuse-on-live|stale) · timeout=1m30s · stale-bound=1m0s · rerequest-cap=5` (log line 23499) [A]. The slot-guard (`refuse-on-live|stale`) is the exact broker-side-stacking brake the 09-05 ruling named.
- **Far-side proof: HELD.** `🎯 stop-entry: seam=on · slots=stop_price · guard=stop-side · unknown=no-op · addon build_id=2026-09-07-h1 expected=2026-09-07-h1 match=yes` (log line 23563) [A]. `PlaceStopEntry` refuses any build below `MinAddonBuildStopSlot` — the build floor the ruling required before relying on the seam.
- Same-boot entry-law line: `stop_entry_seam=ON` (23506); arms line `stop-entry=on(reclaim)` (23503) [A].

## 4. Seam-state timeline (log-verified)

| boot (CT) | seam | build / match | confirm line |
|---|---|---|---|
| 09-18 00:09:58 | OFF (09-05 literal) | 2026-09-07-h1 **match=yes** | broker-snapshot ✓ |
| 09-18 01:32:39 | OFF | match=yes | — |
| 09-18 02:25:22 | OFF | match=yes | — |
| **09-18 07:52:20** | **ON** ← second ruling | match=yes | broker-snapshot ✓ |
| 09-18 13:30:41 | ON | match=yes | ✓ |
| 09-19 01:55:04 | ON | match=yes | ✓ |
| 09-22 00:48:11 | ON | 2026-09-20-p1 match=yes | ✓ |
| 09-23 00:46:33 | ON | build=2026-09-20-p1 expected=2026-09-22-m2 **match=NO** | ✓ |
| 09-24 10:19:48 | ON | 2026-09-23-m21 match=yes | ✓ |
| 09-25 01:40:13 | ON | build_id=none → `slots=unproven(addon build)` | ✓ |
| 09-25 08:21:42 | ON | 2026-09-23-m21 match=yes | ✓ |
| 09-25 20:14:26 (boot 3) | ON | — (`stop-entry=on(reclaim)`, entry law ON) | ✓ |

Notable: on 09-18 00:09–02:25 both preconditions **already held** (match=yes, broker-snapshot live) yet the seam stayed OFF until the 07:52 ruling — the flip was a decision, not an accident. On 09-23 the seam was ON with `match=NO` (expected constant bumped ahead of the AddOn): placement still refused below the build floor (`MinAddonBuildStopSlot`), and the next day's build closed the gap — a transient, floor-protected, not a violation.

## 5. Verdicts

- **P1-1 as CONFIRMED in the knobs addendum: REFUTED here (corrected).** The seam ON is not a contradiction of the 09-05 ruling; the 09-05 ruling was superseded by the 2026-09-18 07:52 CT owner ruling, which the guide clause records. My knobs addendum statement "No later re-enable ruling exists anywhere in docs/" was over-narrow: I grepped `docs/` reports only; the ruling lives in `web/src/guide/content/guards.ts`.
- **NEW P2 (docs staleness):** `.env:40-42` still reads "…Stays off until a cancel-confirmation wave lands…" directly above `STOP_ENTRY_SEAM=on` (`.env:43`). After the second ruling the comment is wrong and actively misleads — it should cite both rulings (09-05 OFF → 09-18 07:52 ON). The 09-05 report remains the only report-level ruling text; recommend a one-line ruling note in a report or the comment fix. Not a runtime defect — the binary reads the env, not the comment.
- **What the CTO observed, verified:** boot-3 cancel path `confirm=broker-snapshot` — true since 09-18 00:09:58 at least [A]; the wave is live and the seam's remaining precondition has held since the flip.

## 6. What I did NOT verify

The second ruling's out-of-band provenance (the owner's 07:52 instruction channel — bridge mail older than 09-25 is not retained; the guide clause is the in-repo record). The 09-23 `match=NO` day is covered in §9-10 (no stop-entry row was created that day; the build 2026-09-20-p1 was above the floor, so the floor would not itself have refused — `match` compares received against the Go-side expected constant, a different signal from the floor). No code, tests, or checklist changes (L17).

---

## 7. The 2026-09-05 hazards, one by one — closed or still open @04ae1c2f

**Hazard A — `nt.CancelOrder` reports success on a SEND, not on a confirmation. CLOSED.**
- The cancel-confirmation wave merged to dev 2026-09-06 00:24–06:48 CT (`871083cc` → `22af989e` → `f61b23a3` → `ea3f33fe`), all first-parent of `04ae1c2f` [A].
- At base: adjudication `adjudicateSlot` (`trader/cancel_confirm.go:113`) — a cancel is proven by the order's ABSENCE FROM A FRESH BROKER SNAPSHOT (rule A20, `:21-24`); per-cycle settlement `confirmPendingCancels` (`:366`), with `:404` "ConfirmCancel demands (snapID > 0)"; boot sweep leaves rows `cancel_pending` until the book confirms (`class33_boot_sweep.go:109-110`) [A].
- An NT8-relayed `Cancelled` state is broker truth and settles a row too (`armed_executor.go:2267`) — that is the path row 174 took (see §9).

**Hazard B — NINE working stop orders for one arm slot (snapshot id 1664), on a 2-contract cap. CLOSED + empirically absent since.**
- Per-slot invariant on BOTH placement paths: `armSlotGuard` at `armed_executor.go:1407` (stop-entry path, guard passed into `placeOneStopEntry`) and `:1450` (the other placement path) — refuse on live/stale slot (`cancel_confirm.go:314-338`) [A].
- One-entry serialization on the wire: maintenance permit + `acquireEntryLatch` ("the ONE entry latch", `tcp_trader.go:792-800`) + `b3Reserve` dedupe (`:800`) [A].
- Rest-cap re-arm waits for the broker book's word (`arm_respec.go:237`, `63faf9c0` "fix(B1 P1 fold, CTO #213)" — in base) [A].
- Boot sweep (class 33) never writes `cancelled` at send time (`86922f72`, PR #216; `5c0edb11`, PR #224 — both merged to dev before base) [A].
- Empirical: 3,626 snapshots with ≥1 working order since 09-18 14:46 — **zero** with >1 working order under one name (§9) [A].

**Hazard C — 21 malformed orders (stop trigger passed in `limitPrice` with literal 0 in `stopPrice`; NT8 accepts and never works it). CLOSED.**
- C# fix (Wave B/D1): `limitArg = isLimit ? limitPx : 0; stopArg = isStopEntry ? stopPx : 0` (`ninjascript/VLTraderTCPClient.cs:1064-1065`), with the rejection of a `stop_entry` frame missing `stop_price` (`:824`) [A].
- Build floor on the wire: `MinAddonBuildStopSlot = "2026-09-05-g2"` (`provider/ninjatrader/tcp_framing.go:322`) enforced in `PlaceStopEntry` (`tcp_trader.go:765-775`): "an AddOn that only proves the parse is refused — never sent a frame it will mis-execute" [A].
- Live: AddOn `2026-09-23-m21` ≥ floor, `match=yes` since 09-24 10:19; the one placed stop entry triggered and filled correctly (§9) [A].

## 8. The live stop-entry path end to end @04ae1c2f

```
arm (armed_executor.go:1378 seam gate → :1407 armSlotGuard → guard=stop-side verdict)
  → placeOneStopEntry (one_live_entry / rest-cap / book-wait rules)
    → TCPTrader.PlaceStopEntry (tcp_trader.go:760):
       1. build floor  FarSideProven(MinAddonBuildStopSlot)        (:765-775)
       2. bound account + isAccountTradeable (SIM-only)            (:780-790)
       3. maintenance permit  acquireEntryPermit                    (:787)
       4. ONE entry latch   acquireEntryLatch("stop-entry …")       (:792)
       5. b3 dedupe        b3Reserve                                (:800)
    → SendSignal stop_entry frame (trigger, sl, tp)
  → AddOn (VLTraderTCPClient.cs:789-824 parse stop_price; :1046-1085 StopMarket
     with stopArg in the stopPrice slot) → NT8 book
cancel: RequestCancel → cancel_pending (never 'cancelled' at send)
  → confirmPendingCancels (broker-snapshot absence, snapID > 0) OR
     NT8-relayed Cancelled state (armed_executor.go:2267)
  → slot guard re-admits only slotFree (refuse-on-live|stale)
  → boot sweep: cancel_pending until book confirms; rerequest-cap=5
```

Which guards stop a second working stop for the same slot: the `armSlotGuard` (both placement paths, `:1407`/`:1450`) refuses unless the broker book shows the slot free; the ONE entry latch + b3 reserve serialize the send; and the ledger never re-arms a row whose cancel is still `cancel_pending`. That is three independent brakes on the 09-04 stacking defect.

## 9. Live history since 2026-09-18 (DB copy + retained logs)

`armed_orders` rows `kind='stop_entry'` created since 09-18 — **five, all named**:
- **id 166** (09-18 01:36) `cancelled — no_trade_band: LONDON first-5m no-trade window` — never placed [A].
- **id 167** (09-18 13:56) **`filled` @29836.25**, signal `edc85c52-6978-4705-9296-9b93b950d3c3` — the first stop entry after the 07:52 ruling. Log 42686: `📌 armed S1 placement requested stop-entry [guard=stop-side verdict=rest action=place] LONG stop-market trigger=29836.25 … signal=edc85c52…` — the trigger worked as a real stop-market (+0.5 tick from entry 29835.75) [A].
- **id 170** (09-20 19:30) `cancelled — accepted through (stop side) … never placed` [A].
- **id 174** (09-21 17:43→17:46) `cancelled — cancelled in NT8`, signal `1f58c7bd-0b51-4995-98ae-259bee10be3b`; `cancel_attempts=1`, `cancel_settled_snapshot_id=0`. In snapshots the signal has exactly ONE order id with lifecycle Initialized→Submitted→Accepted→CancelPending→CancelSubmitted — one working stop for its slot, then gone. The settle path was the NT8-relayed `Cancelled` state (`armed_executor.go:2267`), which is broker truth but did not exercise the strict snapshot-id receipt — the one residual nuance (§10) [A]. The 09-21 log is not retained; this is DB + snapshot evidence.
- **id 177** (09-22 00:50) `cancelled — accepted through (stop side) … never placed` [A].

Stacking scan: all 22,252 `nt8_order_snapshots` since 09-18 parsed; 3,626 have ≥1 working order; **zero snapshots where one name has >1 working order** (the 09-04 defect shape). Stop-type orders seen are single `-sl` bracket stops, one per filled entry [A].

Coverage gaps, stated: snapshots exist only from 09-18 14:46 CT (id 167's order filled before snapshot coverage); logs for 09-20 and 09-21 are not retained.

## 10. Verdict (for the owner; nothing changed)

**SAFE-ON.** Residual risk in one line: a stop-entry cancel can settle on NT8's relayed `Cancelled` state without a persisted snapshot id (row 174) — broker truth, but not the strict A20 absence-proof — and no stacking has recurred in 3,626 working snapshots since the flip.

Conditions already holding live: seam ON only ever reached the wire with the build floor proven (09-18/19/22/24/25 `match=yes`; 09-23 `match=NO` but build 2026-09-20-p1 ≥ floor 09-05-g2 — and no stop entry was attempted that day: zero rows); slot guard refuse-on-live|stale; boot sweep cancel_pending-until-book; rerequest-cap 5. The stale `.env:40-42` comment (and the missing report-level record of the 09-18 ruling) is the only open item, P2 docs-staleness — I do not change `.env` (owner's decision).

