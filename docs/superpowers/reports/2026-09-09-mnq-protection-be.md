# MNQ: position 601 protection and break-even audit

Read-only · 9 September 2026 · all event times CT

This is the frozen audit completed on 9 September, not a live position-status display. Published from an isolated worktree; no trading state was changed.

Evidence: [exit-logs.json](2026-09-09-mnq-protection-be-data/exit-logs.json), [fresh-book.json](2026-09-09-mnq-protection-be-data/fresh-book.json), [last-live-bracket.json](2026-09-09-mnq-protection-be-data/last-live-bracket.json), [nt8-stop-events.json](2026-09-09-mnq-protection-be-data/nt8-stop-events.json), [order-update-frames.json](2026-09-09-mnq-protection-be-data/order-update-frames.json), [position-601.json](2026-09-09-mnq-protection-be-data/position-601.json), [probe.py](2026-09-09-mnq-protection-be-data/probe.py), [runtime.json](2026-09-09-mnq-protection-be-data/runtime.json), [store.json](2026-09-09-mnq-protection-be-data/store.json).


**Position 601 is now CLOSED**, manual exit **29426.5 at 2026-09-09T22:09:00.269000-05:00**. The freshest broker snapshot 15497 at 2026-09-09T22:10:36.435000-05:00 contains 0 working orders. This is a closed position, not a naked open position. [A: position row 601; fresh-book.json.]

**Before the close, PROTECTED at the broker**, as of snapshot **15482**, received **2026-09-09T22:08:36.407000-05:00**. Position **601** was LONG 1 @ **29372.50**. The broker reports a sell stop below entry and sell limit above entry, quantity 1 each, sharing one exit OCO. [A: last-live-bracket.json; position-601.json.]

| Order ID | Name | Action | Type | Price | State | OCO |
|---|---|---|---|---:|---|---|
| dac48bcb0b3647159ba69e31bd02f733 | 87acc3bb-a989-473f-b07b-c77c04bfb419-sl | sell | stop | 29342 | Accepted | 87acc3bb-a989-473f-b07b-c77c04bfb419-exit |
| 9f5a73d0b91445e199a2734c8c3c631a | 87acc3bb-a989-473f-b07b-c77c04bfb419-tp | sell | limit | 29435.5 | Working | 87acc3bb-a989-473f-b07b-c77c04bfb419-exit |

The premise “not in the ledger” needs correction: **parent arm 141 exists**, state `filled`, signal `87acc3bb-a989-473f-b07b-c77c04bfb419`, fill **29372.50 × 1**. Its protective children are broker orders, not separate live-entry rows. The open position row has **`source=reconcile`, `entry_order_id=""`**. At **22:05:31 CT**, Desk explicitly returned: **“the open position carries no usable signal id to join on”**. The book line compares broker 2 with ledger 0 even though the filled parent is retained. [A: store.json arm_141; runtime.json.]

Two additional implementation facts explain why the screen cannot establish protection: `trader/ninjatrader/tcp_trader.go:1024` returns a position map without signal_id/entry_order_id; `trader/desk_facts.go:387` requires one of those keys before reading accepted risk. Even a repaired identity join alone would still find **accepted_risk rows 35/36 with accepted_stop_px=NULL**. Those immutable acceptance rows preceded the child acceptance; the later broker book supplies the stop. `trader/armed_executor.go:1777` returns on child order names before the accepted-risk recorder, and `trader/accepted_risk_hook.go:47` only uses the book available at recording time. The ledger stop **29342.038471896165** is not the broker’s tick-aligned **29342.00**. [A]

## Did break-even fire?

**The trigger was met and logged on both dates; an automatic broker stop move did not fire.** No stop-price amendment toward entry appears in the retained broker histories examined. Do not read “trigger MET” as “modify sent.”

| Position | Trigger time CT | Favorable move logged | Broker stop stayed | Would-be BE stop, NOT sent | log_events.id |
|---|---|---:|---:|---:|---:|
| 595 | 09-08 08:49:55.071 | +44.8 points | 29435.25 | 29478.50 | 34875 |
| 601 | 09-09 21:39:36.051 | +41.0 points | 29342.00 | 29372.50 | 38567 |

**Store log 34875, 2026-09-08T08:49:55.071000-05:00:**

> ⏸ auto-breakeven SUSPENDED (0B) — suspended 2026-09-02 pending MFE data (wave 1A); trigger condition MET and NOT sent to the broker: MNQ LONG +44.8 pts, stop would move to entry 29478.50. Exits are stop/target/EOD-flat/invalidation only. Set EXIT_MECHS_SUSPENDED=0 to restore.

**Store log 38567, 2026-09-09T21:39:36.051000-05:00:**

> ⏸ auto-breakeven SUSPENDED (0B) — suspended 2026-09-02 pending MFE data (wave 1A); trigger condition MET and NOT sent to the broker: MNQ LONG +41.0 pts, stop would move to entry 29372.50. Exits are stop/target/EOD-flat/invalidation only. Set EXIT_MECHS_SUSPENDED=0 to restore.

These are two retained suspension notices, not a count of every favorable-price observation: the suspension logger is once per mechanism/process (`trader/exit_mechs_suspend.go:57`).

## Every position opened since 09-08 00:00 CT

The query returns **8 rows, IDs 594–601**, including reconciliation artifacts. There are no earlier positions carried into the interval in the store query. No performance filters are applied: the owner asked for every position. **NULL accepted stop remains unknown; broker callbacks/snapshots are separately identified.** Stored entry times can be reconciliation materialization times, later than the real broker fill.

| Position; stored entry CT; entry | Join evidence | accepted_risk IDs / accepted_stop_px | Broker -sl first Accepted (CT), price; snapshot ID | Later broker evidence |
|---|---|---|---|---|
| 594; 2026-09-08T02:53:37.612000-05:00; LONG 29546.25 | exit_order_id | 15,16 / NULL,NULL | 2026-09-08T02:52:12.020000-05:00; 29510; snapshot 11515 | unchanged 29510; 2026-09-08 03:02:04:824 CT; Filled; trace.20260908.00000.txt:747 |
| 595; 2026-09-08T08:45:01.035000-05:00; LONG 29478.5 | entry_order_id | 17,18 / 29435.25,29435.25 | 2026-09-08T08:44:40.017000-05:00; 29435.25; snapshot 12237 | unchanged 29435.25; 2026-09-08 09:10:25:387 CT; Filled; trace.20260908.00000.txt:1235 |
| 596; 2026-09-08T09:45:01.033000-05:00; SHORT 29534.0 | entry_order_id | 19,20 / NULL,NULL | 2026-09-08T09:44:31.791000-05:00; 29601.75; snapshot 12374 | unchanged 29601.75; 2026-09-08 10:13:20:087 CT; Filled; trace.20260908.00000.txt:1354 |
| 597; 2026-09-08T12:58:17.861000-05:00; LONG 29595.25 | entry_order_id empty; exit_order_id 1370d16b… | No uniquely joined row | UNKNOWN for this position identity | Adjacent dae090d5 bracket is 29552.75, but assigning it uniquely to this sync/reconcile row would be inference. |
| 598; 2026-09-08T12:59:57.702000-05:00; LONG 29595.25 | entry_order_id | 21,22 / NULL,NULL | 2026-09-08T12:56:46.100000-05:00; 29552.75; snapshot 12775 | unchanged 29552.75; 2026-09-08 12:58:19:158 CT; Cancelled; trace.20260908.00000.txt:1618; replacement fdbaffce -sl Accepted 09-08 12:58:21.485 CT at the SAME 29552.75 (snapshot 12801), cancelled 13:00:28.079 CT; identity discontinuity, not BE |
| 599; 2026-09-08T23:45:10.303000-05:00; SHORT 29604.75 | entry_order_id | 25,26 / NULL,NULL | 2026-09-08T23:43:53.421000-05:00; 29629.75; snapshot 14126 | unchanged 29629.75; 2026-09-09 01:14:21:580 CT; Filled; trace.20260909.00000.txt:17 |
| 600; 2026-09-09T19:58:37.635000-05:00; LONG 29417.0 | exit_order_id | 31,32 / NULL,NULL | 2026-09-09T19:57:10.807000-05:00; 29386.5; snapshot 15193 | unchanged 29386.5; 2026-09-09 20:20:50:978 CT; Filled; trace.20260909.00001.txt:8516 |
| 601; 2026-09-09T21:28:37.006000-05:00; LONG 29372.5 | filled arm 141 + live matching MNQ bracket; missing position join | 35,36 / NULL,NULL | 2026-09-09T21:27:13.747000-05:00; 29342; snapshot 15397 | unchanged 29342.00 through live snapshot 15482 at 2026-09-09T22:08:36.407000-05:00; then cancellation/manual close 2026-09-09T22:09:00.269000-05:00 |

Seven positions have an identifiable primary stop history using entry/exit signal or filled-arm/broker evidence before the manual close; **597 remains an explicit identity exception**. The eight stop-child identities examined include the additional replacement `fdbaffce…-sl`. Across **35 retained NT8 stop OrderUpdate callbacks**, each child identity has one stop price; no ChangePending/ChangeSubmitted amendment appears. Cancellations and replacement at the same price are not break-even moves. Full file/line quotes are in nt8-stop-events.json and below. The extracted periodic/state-change broker snapshot range contains **4343 snapshots**, enumerated in store.json, with the separate freshest snapshot above. This is evidence of no observed modification, not a guarantee against an unrecorded manual action.

## Actual retained order_update frames

The research archive retains **53 order_update records** from **09-09 13:18:13.239 CT** through **09-09 21:27:14.110 CT** in this read. Only positions 600 and 601 have -sl frames in that retained subset. Older positions use the NT8 broker callbacks and persisted snapshots above; no earlier raw frame is invented. The wire payload does **not** include stop_price: its quantity is filled quantity, so zero in an accepted child frame does not mean a zero-sized protective order. `ninjascript/VLTraderTCPClient.cs:1677` constructs that schema.

`research_facts.id=2786693` · receipt **2026-09-09T19:57:10.647000-05:00**

```json
{"fill_price": 0, "order_name": "6705d2c4-158b-41b4-951d-0e5ba09dfb90-sl", "quantity": 0, "seq": 2, "signal_id": "6705d2c4-158b-41b4-951d-0e5ba09dfb90", "state": "initialized", "symbol": "MNQ"}
```

`research_facts.id=2786700` · receipt **2026-09-09T19:57:10.654000-05:00**

```json
{"fill_price": 0, "order_name": "6705d2c4-158b-41b4-951d-0e5ba09dfb90-sl", "quantity": 0, "seq": 2, "signal_id": "6705d2c4-158b-41b4-951d-0e5ba09dfb90", "state": "submitted", "symbol": "MNQ"}
```

`research_facts.id=2786737` · receipt **2026-09-09T19:57:10.753000-05:00**

```json
{"fill_price": 0, "order_name": "6705d2c4-158b-41b4-951d-0e5ba09dfb90-sl", "quantity": 0, "seq": 2, "signal_id": "6705d2c4-158b-41b4-951d-0e5ba09dfb90", "state": "accepted", "symbol": "MNQ"}
```

`research_facts.id=3094672` · receipt **2026-09-09T20:20:50.938000-05:00**

```json
{"fill_price": 29386.5, "order_name": "6705d2c4-158b-41b4-951d-0e5ba09dfb90-sl", "quantity": 1, "seq": 2, "signal_id": "6705d2c4-158b-41b4-951d-0e5ba09dfb90", "state": "filled", "symbol": "MNQ"}
```

`research_facts.id=3762541` · receipt **2026-09-09T21:27:13.994000-05:00**

```json
{"fill_price": 0, "order_name": "87acc3bb-a989-473f-b07b-c77c04bfb419-sl", "quantity": 0, "seq": 5, "signal_id": "87acc3bb-a989-473f-b07b-c77c04bfb419", "state": "submitted", "symbol": "MNQ"}
```

`research_facts.id=3762550` · receipt **2026-09-09T21:27:14.094000-05:00**

```json
{"fill_price": 0, "order_name": "87acc3bb-a989-473f-b07b-c77c04bfb419-sl", "quantity": 0, "seq": 5, "signal_id": "87acc3bb-a989-473f-b07b-c77c04bfb419", "state": "accepted", "symbol": "MNQ"}
```

Prices come from the broker book and its own OrderUpdateCallback traces, not fabricated fields on those frames. Example 601: archive **3762550** records `state=accepted`, while snapshot **15397** records the corresponding stop price **29342.00**.

## Running exit policy and the bound MNQ knob

Running revision: **27e062eab5d536b5c42b3d0f1e0843ca9fd7537b**, PID **368964**, health agrees with `/proc/368964/exe` build metadata, vcs.modified=false. Loaded/disk SHA-256 at 22:05:31 CT: `7a6fb2874e5339a1ae95a4dcfb86b8a10d8f9e4bdb706567b5219fcba3a4f03b`.

Boot quote: `nofx_2026-09-09.log:2252087`, **09-09 18:14:25 CT**:

> 09-09 18:14:25 [INFO] nofx/main.go:341 🛑 exits: stop=max(anchor+clr, 1.5×ATR5m) · anchor_max=3.0×ATR5m · BE=off · trail=off · size=1 · re-arm-after-sweep=on (0B)

Strategy is resolved through `traders.strategy_id`, not an arbitrary active strategy. Bound ID **a5b7662e-7bf7-49bb-9f09-7efa48f95ac8** stores **breakeven_enabled=true, breakeven_trigger_points=40**. The resolver at `trader/auto_trader.go:204` uses 40 because it is positive; 50 is only the fallback. `EXIT_MECHS_SUSPENDED` is unset in dotenv and initial process environment; `trader/exit_mechs_suspend.go:35` therefore returns true. **Effective BE is OFF.** The boot and the two live refusal logs independently agree.

**Which surface is lying about live behavior? The ON presentation in Strategy → Risk Control is misleading, not the BE=off boot line.** `web/src/components/strategy/RiskControlEditor.tsx:230` renders `on={config.breakeven_enabled === true}` without resolving the global suspension. It truthfully shows the saved preference but does not truthfully communicate whether the mechanism can send. The Guide separately says SUSPENDED (`web/src/guide/content/settings.ts:401`). No knob was changed.

**Code path:** `trader/auto_trader_risk.go:109` → `maybeMoveStopToBreakeven` → `breakevenTrigger` → **exitMechSuspendedRefuse at auto_trader.go:166 → return**. The unsent path would continue through `moveStopWire` (`auto_trader.go:185`, `exit_mechs_suspend.go:49`) → `TCPTrader.MoveStopToBreakeven` (`trader/ninjatrader/tcp_trader.go:655`) → `SendMoveStop` (`:673`) → C# `HandleMoveStop`, `StopPriceChanged=newStop; ba.Change(new[]{pb.SlOrder});` (`ninjascript/VLTraderTCPClient.cs:2026`). **No observed modification has a sending code path to attribute: both identified BE attempts stop before this wire path.**

## Reproduction and source freshness

SQLite connections used mode=ro. No code, configuration, database, order or process was changed; source inspected in an isolated worktree. Applied docs/superpowers/AUDIT-CHECKLIST.md evidence/identity/knob checks. Query statements are in probe.py, and the additional exact projections are recorded below. Private account identifiers were removed from artifacts.

```sql
SELECT id,side,entry_price,entry_order_id,exit_order_id,entry_time,exit_time,status,source
FROM trader_positions WHERE entry_time>=1788843600000 ORDER BY id;
-- Output: 594,595,596,597,598,599,600,601 (8 rows).
SELECT id,signal_id,accepted_entry_px,accepted_stop_px,ledger_stop_px,accepted_at_ms
FROM accepted_risk WHERE accepted_at_ms>=1788843600000 ORDER BY id;
-- Output: IDs 13 through 36 (24 rows); values in store.json.
SELECT id,received_at_ms,orders_json FROM nt8_order_snapshots
WHERE received_at_ms>=1788843600000 ORDER BY id;
SELECT id,receipt_ms,fields_json FROM research_facts
WHERE object='exec' AND event='order_update' ORDER BY id;
-- Research DB output: 53 rows, full IDs/frames in order-update-frames.json.
SELECT t.strategy_id,s.config FROM traders t JOIN strategies s ON s.id=t.strategy_id;
-- Safe projection: bound strategy and only BE fields are in store.json.
```

`git log -1 27e062eab5d5 -- "docs/superpowers/AUDIT-CHECKLIST.md"`

```text
27e062eab5d536b5c42b3d0f1e0843ca9fd7537b 2026-09-09T14:55:48-05:00 merge dev (W5) + correct the three sentences this boot makes false
```

`git log -1 27e062eab5d5 -- "trader/auto_trader.go"`

```text
6310eaf8941f53194fa2c5e7368552eed1ed8d64 2026-09-07T23:51:51-05:00 merge: reconcile placement confirmation with pre-send identity and owner ruling
```

`git log -1 27e062eab5d5 -- "trader/auto_trader_risk.go"`

```text
8e6cf957efb14996acf46fd7396aee9af99a78e2 2026-09-07T10:37:50-05:00 feat(protection): D5/D6/D7 — a position without a stop is found, named, and given one
```

`git log -1 27e062eab5d5 -- "trader/exit_mechs_suspend.go"`

```text
4657560bbaf616fcde7457816da5b8aff22431dc 2026-09-02T07:33:39-05:00 fix(0B): exit sanity — stop floor 1.5xATR, stop anchored to seated structure, BE+trail suspended, size 1, re-arm after boot sweep
```

`git log -1 27e062eab5d5 -- "trader/ninjatrader/tcp_trader.go"`

```text
1fb3c21ecd3f7adb144cc9c40e108cc097349b09 2026-09-08T01:30:26-05:00 fix(signal-clock): the last bar-clock on an outgoing command, and class 89
```

`git log -1 27e062eab5d5 -- "trader/desk_facts.go"`

```text
14e1cbcba0bf3ab51ca42fe15a81e348f9e90257 2026-09-09T14:10:38-05:00 fix(risk): D4(c) the CME roll lifts the daily trip; desk strip needs BOTH toggles
```

`git log -1 27e062eab5d5 -- "trader/armed_executor.go"`

```text
2bb9c03eebc12af4b1ba4b439c402cd9c435a651 2026-09-09T14:20:06-05:00 fix(flat): a live row is never retired by our own plumbing being absent; T1 stops reading a failed query as flat
```

`git log -1 27e062eab5d5 -- "trader/accepted_risk_hook.go"`

```text
9ad647de084ffef4e67ba5529c5f453a5172856c 2026-09-07T10:17:27-05:00 fix(bracket): D1/D2/D3/D6 — a cancel targets the entry, the bracket is sized from the fill, one state vocabulary, protective orders GTC
```

`git log -1 27e062eab5d5 -- "ninjascript/VLTraderTCPClient.cs"`

```text
b4195e6f877032090812214b8ae4b6acae777a4f 2026-09-07T10:53:37-05:00 fix(exit): class 80 — a REJECTED limit exit no longer cancels the position's bracket
```

`git log -1 27e062eab5d5 -- "web/src/components/strategy/RiskControlEditor.tsx"`

```text
9470a79d6895fa6be87c4c550220a55d07ed3aa7 2026-08-20T01:04:14-05:00 fix(E5): v1 strays — the chat-path OHLCV table renders CT like every trading prompt (the last Time(UTC) site), and the Studio min-confidence display fallback mirrors the real shared default 60 (was a stale 75)
```

`git log -1 27e062eab5d5 -- "web/src/guide/content/settings.ts"`

```text
486cec4646c433d190fd1ab794e7d6f6eac57f69 2026-09-09T14:15:47-05:00 docs(W5): the system describes itself truthfully — 12 false claims corrected to what the code does
```

Cited BE/desk/bridge implementations were compared between the running revision and inspected dev checkout 8941ec68; these cited files are unchanged. All claims about current execution are anchored to running 27e062eab5d5, not the later unbooted dev tip.

## Complete retained NT8 stop callback quotes

`trace.20260908.00000.txt:722` (CT)

```text
2026-09-08 02:52:11:901 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Submitted orderId='031639e73d714ab2b9d3255ffc9f2f89' account='[redacted]' name='f159e573-9c38-4d24-be8f-24586c214d16-sl' orderState=Submitted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29510 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 02:52:11' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:729` (CT)

```text
2026-09-08 02:52:12:020 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Accepted orderId='031639e73d714ab2b9d3255ffc9f2f89' account='[redacted]' name='f159e573-9c38-4d24-be8f-24586c214d16-sl' orderState=Accepted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29510 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 02:52:12' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:744` (CT)

```text
2026-09-08 03:02:04:823 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Working orderId='031639e73d714ab2b9d3255ffc9f2f89' account='[redacted]' name='f159e573-9c38-4d24-be8f-24586c214d16-sl' orderState=Working instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29510 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 03:02:04' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:747` (CT)

```text
2026-09-08 03:02:04:824 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Filled orderId='031639e73d714ab2b9d3255ffc9f2f89' account='[redacted]' name='f159e573-9c38-4d24-be8f-24586c214d16-sl' orderState=Filled instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29510 quantity=1 orderType='Stop Market' filled=1 averageFillPrice=29510 time='2026-09-08 03:02:04' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1196` (CT)

```text
2026-09-08 08:44:39:907 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Submitted orderId='fdbc976f8c734ad590588604b3f803db' account='[redacted]' name='f68d8dc6-8b5c-4a2f-be80-b30de90f6ca8-sl' orderState=Submitted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29435.25 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 08:44:39' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1203` (CT)

```text
2026-09-08 08:44:40:017 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Accepted orderId='fdbc976f8c734ad590588604b3f803db' account='[redacted]' name='f68d8dc6-8b5c-4a2f-be80-b30de90f6ca8-sl' orderState=Accepted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29435.25 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 08:44:40' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1232` (CT)

```text
2026-09-08 09:10:25:387 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Working orderId='fdbc976f8c734ad590588604b3f803db' account='[redacted]' name='f68d8dc6-8b5c-4a2f-be80-b30de90f6ca8-sl' orderState=Working instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29435.25 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 09:10:25' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1235` (CT)

```text
2026-09-08 09:10:25:387 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Filled orderId='fdbc976f8c734ad590588604b3f803db' account='[redacted]' name='f68d8dc6-8b5c-4a2f-be80-b30de90f6ca8-sl' orderState=Filled instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29435.25 quantity=1 orderType='Stop Market' filled=1 averageFillPrice=29435.25 time='2026-09-08 09:10:25' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1312` (CT)

```text
2026-09-08 09:44:31:678 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Submitted orderId='2bb152392e7e4203baff9ab48119299c' account='[redacted]' name='55a2f48f-0cd4-46a9-a6a7-fa29a461dcdb-sl' orderState=Submitted instrument='MNQ 09-26' orderAction=BuyToCover limitPrice=0 stopPrice=29601.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 09:44:31' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1319` (CT)

```text
2026-09-08 09:44:31:791 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Accepted orderId='2bb152392e7e4203baff9ab48119299c' account='[redacted]' name='55a2f48f-0cd4-46a9-a6a7-fa29a461dcdb-sl' orderState=Accepted instrument='MNQ 09-26' orderAction=BuyToCover limitPrice=0 stopPrice=29601.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 09:44:31' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1351` (CT)

```text
2026-09-08 10:13:20:087 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Working orderId='2bb152392e7e4203baff9ab48119299c' account='[redacted]' name='55a2f48f-0cd4-46a9-a6a7-fa29a461dcdb-sl' orderState=Working instrument='MNQ 09-26' orderAction=BuyToCover limitPrice=0 stopPrice=29601.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 10:13:20' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1354` (CT)

```text
2026-09-08 10:13:20:087 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Filled orderId='2bb152392e7e4203baff9ab48119299c' account='[redacted]' name='55a2f48f-0cd4-46a9-a6a7-fa29a461dcdb-sl' orderState=Filled instrument='MNQ 09-26' orderAction=BuyToCover limitPrice=0 stopPrice=29601.75 quantity=1 orderType='Stop Market' filled=1 averageFillPrice=29601.75 time='2026-09-08 10:13:20' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1580` (CT)

```text
2026-09-08 12:56:45:981 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Submitted orderId='72e28c0224064353b9881b7f85e4611d' account='[redacted]' name='dae090d5-680f-42d6-93c0-197920abc97d-sl' orderState=Submitted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 12:56:45' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1587` (CT)

```text
2026-09-08 12:56:46:100 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Accepted orderId='72e28c0224064353b9881b7f85e4611d' account='[redacted]' name='dae090d5-680f-42d6-93c0-197920abc97d-sl' orderState=Accepted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 12:56:46' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1603` (CT)

```text
2026-09-08 12:58:19:053 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=CancelPending orderId='72e28c0224064353b9881b7f85e4611d' account='[redacted]' name='dae090d5-680f-42d6-93c0-197920abc97d-sl' orderState=CancelPending instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 12:58:19' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1613` (CT)

```text
2026-09-08 12:58:19:053 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=CancelSubmitted orderId='72e28c0224064353b9881b7f85e4611d' account='[redacted]' name='dae090d5-680f-42d6-93c0-197920abc97d-sl' orderState=CancelSubmitted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 12:58:19' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1618` (CT)

```text
2026-09-08 12:58:19:158 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Cancelled orderId='72e28c0224064353b9881b7f85e4611d' account='[redacted]' name='dae090d5-680f-42d6-93c0-197920abc97d-sl' orderState=Cancelled instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 12:58:19' statementDate='2026-09-08' error=NoError comment='' nr=5
```

`trace.20260908.00000.txt:1638` (CT)

```text
2026-09-08 12:58:21:373 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Submitted orderId='d99065217a17455cb34e2568f7dc0182' account='[redacted]' name='fdbaffce-40fb-458d-a8a1-965854b1cde1-sl' orderState=Submitted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 12:58:21' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1645` (CT)

```text
2026-09-08 12:58:21:485 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Accepted orderId='d99065217a17455cb34e2568f7dc0182' account='[redacted]' name='fdbaffce-40fb-458d-a8a1-965854b1cde1-sl' orderState=Accepted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 12:58:21' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00000.txt:1662` (CT)

```text
2026-09-08 13:00:27:967 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=CancelPending orderId='d99065217a17455cb34e2568f7dc0182' account='[redacted]' name='fdbaffce-40fb-458d-a8a1-965854b1cde1-sl' orderState=CancelPending instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 13:00:27' statementDate='2026-09-08' error=NoError comment='' nr=3
```

`trace.20260908.00000.txt:1674` (CT)

```text
2026-09-08 13:00:27:968 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=CancelSubmitted orderId='d99065217a17455cb34e2568f7dc0182' account='[redacted]' name='fdbaffce-40fb-458d-a8a1-965854b1cde1-sl' orderState=CancelSubmitted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 13:00:27' statementDate='2026-09-08' error=NoError comment='' nr=4
```

`trace.20260908.00000.txt:1678` (CT)

```text
2026-09-08 13:00:28:079 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Cancelled orderId='d99065217a17455cb34e2568f7dc0182' account='[redacted]' name='fdbaffce-40fb-458d-a8a1-965854b1cde1-sl' orderState=Cancelled instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29552.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 13:00:28' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00001.txt:601` (CT)

```text
2026-09-08 23:43:53:305 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Submitted orderId='c77d4bfe0f804cb8a6650637570b3b8b' account='[redacted]' name='c5424fe7-6fa3-446a-8e94-de6d73620b0f-sl' orderState=Submitted instrument='MNQ 09-26' orderAction=BuyToCover limitPrice=0 stopPrice=29629.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 23:43:53' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260908.00001.txt:608` (CT)

```text
2026-09-08 23:43:53:421 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Accepted orderId='c77d4bfe0f804cb8a6650637570b3b8b' account='[redacted]' name='c5424fe7-6fa3-446a-8e94-de6d73620b0f-sl' orderState=Accepted instrument='MNQ 09-26' orderAction=BuyToCover limitPrice=0 stopPrice=29629.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-08 23:43:53' statementDate='2026-09-08' error=NoError comment='' nr=-1
```

`trace.20260909.00000.txt:14` (CT)

```text
2026-09-09 01:14:21:580 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Working orderId='c77d4bfe0f804cb8a6650637570b3b8b' account='[redacted]' name='c5424fe7-6fa3-446a-8e94-de6d73620b0f-sl' orderState=Working instrument='MNQ 09-26' orderAction=BuyToCover limitPrice=0 stopPrice=29629.75 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-09 01:14:21' statementDate='2026-09-09' error=NoError comment='' nr=-1
```

`trace.20260909.00000.txt:17` (CT)

```text
2026-09-09 01:14:21:580 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Filled orderId='c77d4bfe0f804cb8a6650637570b3b8b' account='[redacted]' name='c5424fe7-6fa3-446a-8e94-de6d73620b0f-sl' orderState=Filled instrument='MNQ 09-26' orderAction=BuyToCover limitPrice=0 stopPrice=29629.75 quantity=1 orderType='Stop Market' filled=1 averageFillPrice=29629.75 time='2026-09-09 01:14:21' statementDate='2026-09-09' error=NoError comment='' nr=-1
```

`trace.20260909.00001.txt:8476` (CT)

```text
2026-09-09 19:57:10:702 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Submitted orderId='fab42136783242f9b4eb1ab45e153e3c' account='[redacted]' name='6705d2c4-158b-41b4-951d-0e5ba09dfb90-sl' orderState=Submitted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29386.5 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-09 19:57:10' statementDate='2026-09-09' error=NoError comment='' nr=-1
```

`trace.20260909.00001.txt:8483` (CT)

```text
2026-09-09 19:57:10:807 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Accepted orderId='fab42136783242f9b4eb1ab45e153e3c' account='[redacted]' name='6705d2c4-158b-41b4-951d-0e5ba09dfb90-sl' orderState=Accepted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29386.5 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-09 19:57:10' statementDate='2026-09-09' error=NoError comment='' nr=-1
```

`trace.20260909.00001.txt:8513` (CT)

```text
2026-09-09 20:20:50:977 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Working orderId='fab42136783242f9b4eb1ab45e153e3c' account='[redacted]' name='6705d2c4-158b-41b4-951d-0e5ba09dfb90-sl' orderState=Working instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29386.5 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-09 20:20:50' statementDate='2026-09-09' error=NoError comment='' nr=-1
```

`trace.20260909.00001.txt:8516` (CT)

```text
2026-09-09 20:20:50:978 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Filled orderId='fab42136783242f9b4eb1ab45e153e3c' account='[redacted]' name='6705d2c4-158b-41b4-951d-0e5ba09dfb90-sl' orderState=Filled instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29386.5 quantity=1 orderType='Stop Market' filled=1 averageFillPrice=29386.5 time='2026-09-09 20:20:50' statementDate='2026-09-09' error=NoError comment='' nr=-1
```

`trace.20260909.00001.txt:8638` (CT)

```text
2026-09-09 21:27:13:642 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Submitted orderId='dac48bcb0b3647159ba69e31bd02f733' account='[redacted]' name='87acc3bb-a989-473f-b07b-c77c04bfb419-sl' orderState=Submitted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29342 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-09 21:27:13' statementDate='2026-09-09' error=NoError comment='' nr=1
```

`trace.20260909.00001.txt:8645` (CT)

```text
2026-09-09 21:27:13:747 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Accepted orderId='dac48bcb0b3647159ba69e31bd02f733' account='[redacted]' name='87acc3bb-a989-473f-b07b-c77c04bfb419-sl' orderState=Accepted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29342 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-09 21:27:13' statementDate='2026-09-09' error=NoError comment='' nr=-1
```

`trace.20260909.00001.txt:8708` (CT)

```text
2026-09-09 22:09:00:022 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=CancelPending orderId='dac48bcb0b3647159ba69e31bd02f733' account='[redacted]' name='87acc3bb-a989-473f-b07b-c77c04bfb419-sl' orderState=CancelPending instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29342 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-09 22:09:00' statementDate='2026-09-09' error=NoError comment='' nr=3
```

`trace.20260909.00001.txt:8717` (CT)

```text
2026-09-09 22:09:00:023 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=CancelSubmitted orderId='dac48bcb0b3647159ba69e31bd02f733' account='[redacted]' name='87acc3bb-a989-473f-b07b-c77c04bfb419-sl' orderState=CancelSubmitted instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29342 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-09 22:09:00' statementDate='2026-09-09' error=NoError comment='' nr=-1
```

`trace.20260909.00001.txt:8719` (CT)

```text
2026-09-09 22:09:00:134 (lucid ) Cbi.Account.OrderUpdateCallback: realOrderState=Cancelled orderId='dac48bcb0b3647159ba69e31bd02f733' account='[redacted]' name='87acc3bb-a989-473f-b07b-c77c04bfb419-sl' orderState=Cancelled instrument='MNQ 09-26' orderAction=Sell limitPrice=0 stopPrice=29342 quantity=1 orderType='Stop Market' filled=0 averageFillPrice=0 time='2026-09-09 22:09:00' statementDate='2026-09-09' error=NoError comment='' nr=-1
```
