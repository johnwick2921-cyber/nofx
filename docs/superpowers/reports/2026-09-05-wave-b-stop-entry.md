# WAVE B — the stop entry reaches the broker with a trigger, and the guard that judges it is the stop-side one

**Branch:** `fix/wave-b-stop-entry` · **base:** `a45cf551` (= `origin/dev` tip at accept) · **claim:** `cb23d9eb`
**Session:** waveb-stopentry-0905 · **worktree:** `/home/hoang/nofx-waveb`
**Status: PUSHED, GREEN, NOT DEPLOYED.** No release build, no binary swap, no kill, nothing copied into the NinjaTrader Documents AddOns folder, no NT8 restart (A3). Sections F1/F2 are the owner's.

**SPEC FRESHNESS (class 73).** Base `a45cf551` (2026-09-05 18:22:18 CT). `git log -1` for every file built from — nothing moved after the base, no rebase needed:

| file | last commit before base |
|---|---|
| `ninjascript/VLTraderTCPClient.cs` | `c84bd247` 2026-09-03 20:04:23 |
| `trader/armed_executor.go` | `51916172` 2026-09-04 09:48:03 |
| `provider/ninjatrader/tcp_framing.go` | `a1aa1eb6` 2026-09-03 19:41:44 |
| `provider/ninjatrader/order_snapshot.go` | `c84bd247` 2026-09-03 20:04:23 |
| `trader/ninjatrader/tcp_trader.go` | `2c0f005c` 2026-09-02 06:41:43 |
| `trader/class33_boot_sweep.go` | `c84bd247` 2026-09-03 20:04:23 |
| `trader/arms_boot_line.go` | `59d01948` 2026-09-04 08:11:25 |
| `trader/reaper_snapshot.go` | `e54d0ad7` 2026-09-04 09:51:48 |
| `docs/superpowers/SYSTEM-MAP.md` | `a96224dd` 2026-09-04 09:07:37 |
| `docs/superpowers/AUDIT-CHECKLIST.md` | `15340faa` 2026-09-04 13:22:07 |
| `web/src/guide/content/plays.ts` | `59d01948` 2026-09-04 08:11:25 |
| `web/src/guide/content/guards.ts` | `c84bd247` 2026-09-03 20:04:23 |

---

## THE ONE THING TO KNOW BEFORE ANYTHING ELSE

**The two defects masked each other, and shipping D1 without D2 is a REGRESSION, not a partial fix.**

A `StopMarket` whose `StopPrice` is 0 has a trigger of zero, so it rests inert forever. That is the only reason the inverted guard did no damage: on 2026-09-04 it admitted 21 sell stops while the market sat 50.00–102.75 points *below* their trigger — every one already-through — and all 21 sat harmless because the trigger was zero. Correct the C# slot alone and those same 21 placements become 21 orders NinjaTrader acts on the instant they land, roughly 75 points adverse (or 21 `Sell stop or sell stop limit orders can't be placed above the market` rejections — that native error already appears 14 times in the NT8 logs). Either outcome is wrong.

**D1 and D2 are in the same commit and must stay in the same binary.** If the wave is ever split, D2 alone is safe (it converts inert orders into cancels); D1 alone is not.

---

## SECTION G — C1 THROUGH C5

### C1 — CONFIRMED verbatim, at the corrected lines :972-979

The dispatch's original `:972-978` was off by one at each end; `:974-979` as the owner instructed. Read at base:

```
972  bool isLimit    = orderType == "limit" && limitPx > 0;
973  bool isStopEntry = orderType == "stop_entry" && stopPx > 0;
974  OrderType orderT = isLimit ? OrderType.Limit : (isStopEntry ? OrderType.StopMarket : OrderType.Market);
975  double orderPx  = isLimit ? limitPx : (isStopEntry ? stopPx : 0);
976  var entryOrder = submitAccount.CreateOrder(
977      instrument, entryAction, orderT, OrderEntry.Manual,
978      TimeInForce.Day, qty, orderPx, 0, string.Empty, signalId,
979      Core.Globals.MaxDate, null);
```

`Account.CreateOrder`'s positional signature after `quantity` is `(limitPrice, stopPrice, oco, name, gtd, customOrder)`. For a `StopMarket` the trigger sat in `limitPrice` and `stopPrice` received a literal `0`.

**Three independent witnesses, all agreeing, none of them our own logging:**
1. **NinjaTrader's own order log** — `Limit price=29590.5 Stop price=0 … Type='Stop Market'`, and none of the 21 ever reached `Working` or `Filled`.
2. **The broker's book on our wire** — `nt8_order_snapshots` ids 1480/1483/1485 hold `{"type":"stop","limit_price":29590.5,…}` with `stop_price` **absent**. `NT8Order.StopPrice` is `json:"stop_price,omitempty"` (`order_snapshot.go:35`), so absence there means **zero**, not unknown.
3. **The in-file control** — the bracket stop-loss at `:1822-1825` builds a `StopMarket` the other way round (`b.Qty, 0, b.Sl`) and is proven correct by a live fill (2026-09-03, `f2b1eb20-…-sl`, filled at 29355).

**Corrections to the dispatch's framing, both from the investigators and both re-verified here:**
- **The blast radius is wider than one session.** Stop-market *entries* exist on exactly two days — 2026-08-31 (1) and 2026-09-04 (21) = **22 lifetime**, every one with `Stop price=0`, **0 fills, lifetime fill rate 0/22**. The stop-entry path has never once placed a well-formed order.
- **`WHERE kind='stop_entry'` returns 21 and under-reports by one.** Re-verified read-only tonight: `armed_orders` kind census is `''=17, limit=29, stop_entry=21`; the 08-31 twin is **id 16** (`kind=''`, it predates the column, entry 28700, signal `1deb5e23-…`, created 2026-08-31 00:09:35). Any rate quoted off the kind filter has the wrong denominator.
- **The 08-31 order was the E7 capability's own PROOF.** It was accepted as a pass because it rested and cancelled — and a zero-trigger stop rests perfectly, forever. The proof that certified the feature was the bug.
- **No other `CreateOrder` call site is affected.** All four sites checked: `:976` (the defect), `:1137` limit close, `:1822` stop-loss, `:1826` take-profit. The last three are correct.

**Sample ids (A21), re-verified read-only at 21:48 CT 2026-09-05:** the 21 are `38, 62, 65, 67, 70, 73, 75, 77, 79, 81, 83, 85, 87, 89, 91, 93, 95, 97, 99, 101, 102` — all NY / S2 / SHORT / `entry_px` 29591.02, wire trigger 29590.50, **21 cancelled, 0 filled** (`SUM(fill_price>0) = 0`).

### C2 — CONFIRMED, and BOTH sides are inverted, not just the long side

`armed_executor.go:940` called `limitMarketableWrongSide(price, trigger, r.Side)` with the **trigger** in the `entry` argument. The predicate returns `price < entry` for long and `price > entry` for short. For a resting stop that is exactly the VALID condition, so all four cells are backwards:

| # | order | market vs trigger | correct | code did |
|---|---|---|---|---|
| 1 | BUY STOP | price **below** trigger — valid resting | place | **cancelled** |
| 2 | BUY STOP | price **at/above** trigger — already through | cancel | **placed** |
| 3 | SELL STOP | price **above** trigger — valid resting | place | **cancelled** |
| 4 | SELL STOP | price **at/below** trigger — already through | cancel | **placed** |

**Consequence, measured:** cell 4 fired 21 of 21 times on 2026-09-04. Cell 3 has **never been observed** (n=0) — not one tick of that window put price above the trigger — so the "cancels valid ones" half is proven by code reading only, and is stated as such. The stop-side guard has **never fired** in any log (`grep "already traded through"` → 0), while the limit-side guard fires normally (7 times).

**A secondary defect the dispatch did not name:** `throughWord` (`:999-1004`) returns "above" for short and "below" for long — correct for limits, **inverted for stops**. The cancel message at `:942` would have asserted the opposite of what happened. Fixed by giving the stop branch a message that prints the relation the guard actually evaluated (`>=` / `<=`) rather than a shared English word; `throughWord` itself is untouched because the limit branch at `:961` depends on it.

### C3 — CORRECTED. The direction claim holds; both counts do not.

- Reclaim-family **by condition** is **24 of 24 SHORT**, ids `38, 62-102, 103, 104, 105`. The dispatch's "21 of 21" is the **stop-entry subset**.
- **37 of 67 rows (ids 1-37) carry `condition=''`** — the column postdates them — so the family census rests on 30 classifiable rows. Any percentage must name that denominator.
- **`kind='stop_entry'` has never once been authored LONG.** The buy-stop path has **zero production exercise**: every line of the long branch this wave fixes is unexercised by live data, which is why the pins carry both sides and the exact-touch boundary explicitly.
- "The only stop-entry ever placed" is wrong as a *placement* count: **22 distinct broker orders** were placed (one logical arm re-placed 21 times on 09-04, plus id 16).

### C4 — NOT TOUCHED, as ruled

`stopEntryNeedsRetestWindow(r.Condition)` / `stopEntryFallbackDue(...)` and their comment ("a reclaim's buy stop IS the entry — waiting for a no-retest window would miss the reclaim it exists to catch") are byte-identical. `ArmKindFor`, the armable set, `composeArmStop`, the ATR floor, the reaper, prompts, validators, levels, wakes and bars are untouched — confirmed by the scope diff below. The waterfall pullback-limit design is not flipped.

### C5 — SPLIT: one half confirmed exactly, one half REFUTED

- **Arm 33's overshoot is 0.70 points — CONFIRMED exactly.** `29167.50 − 29166.80 = 0.70` (2.8 MNQ ticks), the smallest of the 7 marketable cancels on record. **"1.7" appears nowhere in the record.** The nearest lookalikes are arm 36's 7.70-point overshoot and arm 13's 1.7933 points of R:R headroom — both look like transcriptions that lost their column.
- **"Arm 35 at R:R 2.01 could not absorb a single adverse tick" — REFUTED.** Arm 35 measures **R:R 2.1087** (`140.50 / 66.6285`) with **9.66 ticks** of headroom to the 2.0 floor. "2.01" is not any arm's R:R; it is arm 36's headroom-to-floor **in points** (2.0104).
- **The underlying worry is real and lands elsewhere — REPORTED, NOT FIXED (A31).** The R:R gate (`armed_executor.go:1379-1387`) computes `rr` from the **authored entry**; the trigger is computed ~450 lines away and never re-gated. All 21 stop-entry arms passed the 2.0 floor at their authored entry (2.0058–2.0195) and were **already below it at the trigger the wire actually carried** (1.9775–1.9909) before a single tick of slippage — and a stop-market fills at or beyond its trigger *by construction*, so for this order type the slip is structural. Arm 38 had **0.43 ticks** of headroom. Stated; nothing proposed. No tick cap (separate owner ruling).

---

## THE CHANGE — exact before/after

### D1 · C# — one order, correct slots · `ninjascript/VLTraderTCPClient.cs`

**Before (`:975-979`)** → **After (`:986-991`)**:

```csharp
- double orderPx  = isLimit ? limitPx : (isStopEntry ? stopPx : 0);
- var entryOrder = submitAccount.CreateOrder(
-     instrument, entryAction, orderT, OrderEntry.Manual,
-     TimeInForce.Day, qty, orderPx, 0, string.Empty, signalId,
-     Core.Globals.MaxDate, null);
+ double limitArg = isLimit ? limitPx : 0;
+ double stopArg  = isStopEntry ? stopPx : 0;
+ var entryOrder = submitAccount.CreateOrder(
+     instrument, entryAction, orderT, OrderEntry.Manual,
+     TimeInForce.Day, qty, limitArg, stopArg, string.Empty, signalId,
+     Core.Globals.MaxDate, null);
```

`:972-974` (the type selection) and the parse at `:767-776` are untouched — both were already correct. Market orders keep `0, 0`. **StopMarket only; no stop-limit introduced.**

**A9 — the log now prints what was SENT, not what was parsed** (`:1013-1015`). The old line printed `stop@29590.5` on all 21 malformed submissions, which is precisely why a slot bug survived in a file that logs every placement:

```csharp
- + (isLimit ? (" limit@" + limitPx) : (isStopEntry ? (" stop@" + stopPx) : (" entry≈" + entry)))
+ + " action=" + entryAction + " type=" + orderT
+ + " limitPrice=" + limitArg + " stopPrice=" + stopArg
```

`VL_BUILD_ID` (`:55`) `"2026-09-03-f12"` → **`"2026-09-05-g1"`**. Emission verified, not duplicated — three existing sites, one per frame: hello `:1437`, order_snapshot `:1561`, heartbeat `:2383`.

### D2/D3 · Go — the stop-side guard, chosen BY KIND · `trader/armed_executor.go`

**Before (`:940-943`)** → **After (`:942-961`)**:

```go
- if price > 0 && limitMarketableWrongSide(price, trigger, r.Side) {
-     _ = ledger.SetState(r.ID, "cancelled", "trigger already traded through — never placed")
-     at.logWarnf("✕ armed %s stop-entry cancelled — price %.2f already %s the trigger %.2f (never placed)", …throughWord(r.Side)…)
-     continue
- }
+ verdict, why := stopEntryGuardVerdict(r.Side, trigger, price)
+ switch verdict {
+ case stopGuardThrough:   // cancel, naming the relation it evaluated
+ case stopGuardUnknown:   // D3 — no cancel, no placement: WARN + counter
+ }
```

New, at `:1030-1122`:
- `stopGuardVerdict` — an iota enum whose **zero value is `stopGuardUnknown`**, so a forgotten verdict reads as the safe branch. `String()`'s `default` is `"unknown"`.
- `stopEntryMarketableWrongSide(side, trigger, price)` (`:1077`) — long `price >= trigger`, short `price <= trigger`. **Inclusive**, because a stop AT its trigger fires.
- `stopEntryGuardVerdict(side, trigger, price) (verdict, why)` (`:1093`) — pure, returns the reason in the words the log prints.

**`limitMarketableWrongSide` is byte-identical and still called at `:983` with `r.EntryPx`.** It is correct for limits, has fired correctly 7 times in production, and `TestLimitMarketableWrongSide` is unedited. The boundary difference is the point of the split: **a limit at its price rests (strict `<`/`>`); a stop at its trigger fires (inclusive `>=`/`<=`).**

Refusal string, in the dispatch's exact shape — rendered live, not quoted from the spec:

```
accepted through (stop side): price 29515.25 <= trigger 29590.50
```

**D3 unknown path (`:951-959`):** no cancel, no placement, arm left exactly as it is for the next cycle, `⚠️ armed %s stop-entry NOT adjudicated [guard=stop-side] …` + `store.IncArmRefusal` class `stop_entry:guard_unknown` + `telemetry.IncGateBlock`, deduped by arm-spec via `armRefusalChanged`. Rendered live:

```
stop-side guard not evaluated: no price (trigger 29610.00) — nothing cancelled
```

**A9 on the success path too (`:977-979`):** the placement line now names the order type, the trigger, the price, the side and which guard cleared it — `📌 armed S2 → WORKING stop-entry [guard=stop-side] SHORT stop-market trigger=… price=… signal=… (rests (stop side): …)`. Previously it named neither the price nor the guard, which is why the 21 admissions left no trace and the investigators had to reconstruct price from a `📏 arm far` line in a different function.

### D5 · the AddOn build floor · `provider/ninjatrader/tcp_framing.go`, `trader/ninjatrader/tcp_trader.go`

No new machinery — `FarSideProven` is reused unchanged. One new constant plus a sentinel:

```go
const MinAddonBuildStopSlot = "2026-09-05-g1"          // tcp_framing.go:241
var   ErrAddonBuildTooOld = errors.New("addon build predates the stop-slot fix")  // :246
```

`PlaceStopEntry`'s first check (`tcp_trader.go:502`) now gates on `MinAddonBuildStopSlot`:

```
ninjatrader/tcp: refusing stop-entry short MNQ trigger=29590.50 qty=1 [guard=far_side_build] —
addon build predates the stop-slot fix (build_id=2026-09-03-f12, need ≥ 2026-09-05-g1):
does not prove stop_entry support; F5-compile + restart the new AddOn
```

The old substring `does not prove stop_entry support` is retained so the existing refusal fixture stays byte-identical. `%q` on the build id is gone — an unheard-from AddOn now renders as `none` via a single shared `BuildIDForLog` (`order_snapshot.go`), used by both the refusal and `AddonBuildLine`, so the two cannot disagree about what "we have not heard from NT8" looks like (A24).

**`ExpectedAddonBuild` bumped in lockstep** to `"2026-09-05-g1"`, and a test reads `VL_BUILD_ID` **out of the .cs file** rather than restating it, so a half-bump cannot go green.

**THE LEXICAL TRAP, and why the date moved.** `FarSideProven` is a bytewise string compare and suffixes are not zero-padded: `"2026-09-03-f9" >= "2026-09-03-f12"` is **TRUE**. A same-date suffix bump would silently open the gate on the next two-digit build. The new floor advances the **ISO date**, and the pin asserts `MinAddonBuildStopSlot[:10] > "2026-09-03"` so a future suffix-only bump fails the suite.

**`FarSideBuildE7` is RETAINED and now gates nothing** — a deliberate, recorded decision, not an oversight. Deleting it would strand the prose comment at `provider/ninjatrader/tcp_server.go:80`, and that file is outside this wave's footprint (A31). Its doc comment now says plainly that it proved the AddOn *parsed* a stop_entry frame and nothing more, and it earns a real ongoing role as the **negative fixture** in `TestStopEntryRefusedOnPreStopSlotBuild`. Flagged here rather than hidden: **this is one constant with zero production call sites**, and the next lane to touch `tcp_server.go` should delete it and the comment together.

### D4 · the boot line · `StopEntryBootLine` (`armed_executor.go:1160`)

Every field READ from the enforcing code, none a literal (A11). Rendered live, all three states:

```
🎯 stop-entry: slots=stop_price · guard=stop-side · unknown=no-op · addon build_id=2026-09-05-g1 expected=2026-09-05-g1 match=yes
🎯 stop-entry: slots=unproven(addon build) · guard=stop-side · unknown=no-op · addon build_id=2026-09-03-f12 expected=2026-09-05-g1 match=NO
🎯 stop-entry: slots=unproven(addon build) · guard=stop-side · unknown=no-op · addon build_id=none expected=2026-09-05-g1 match=NO
```

- `slots` — resolved from the **same gate** `PlaceStopEntry` uses. The slot order lives in the C# and this process cannot read it, so the only honest Go-side claim is "the AddOn that answered proves the fix". An unproven build reads `unproven`, never `stop_price`.
- `guard` — resolved by asking `stopEntryGuardVerdict` on the canonical already-through and resting cases. An inverted guard renders `MISROUTED`.
- `unknown` — resolved from the verdict enum on an unevaluable input. A guard that cancelled on ignorance would render `CANCELS`.
- `build_id` — `at.farSideBuildID()`, **the last received frame**, never the source constant. The build half is rendered by `AddonBuildLine`, the one renderer that decides what `match` means.

**Emission site, and why not the obvious one.** It is emitted once per trader on the first armed cycle with a bound NT8 trader (`logStopEntryBootLineOnce`, `:1191`, called at `:896`) — **not** hung off `class33_boot_sweep.go:121`. That path latches and is skipped entirely when the sweep defers, when the ledger read fails, or when any cancel failed — three ways for the line to silently not exist (3 of 8 observed 🔌 lines already read `build_id=none` for exactly that reason). The first armed cycle is the moment the stop-entry path becomes live.

---

## THE TESTS — every pin RED before GREEN (A8)

Baseline at `cb23d9eb`: `go build ./...` rc=0, full suite green. Any red below is this wave's.

### E3 — the four cells · `trader/wave_b_stop_guard_test.go`

RED, run against the production expression as it stood (`limitMarketableWrongSide(price, trigger, side)`) — **all 8 cases inverted, zero exceptions:**

```
--- FAIL: TestStopEntryMarketableWrongSide (0.00s)
    buy stop rests below: side="long" trigger=29610.00 price=29590.25: got through=true want false
    buy stop through above: side="long" trigger=29610.00 price=29612.25: got through=false want true
    buy stop exact touch: side="long" trigger=29610.00 price=29610.00: got through=false want true
    sell stop rests above: side="short" trigger=29590.50 price=29650.00: got through=true want false
    sell stop through below (09-04 id 38): side="short" trigger=29590.50 price=29515.25: got through=false want true
    sell stop exact touch: side="short" trigger=29590.50 price=29590.50: got through=false want true
    case folded long / case folded short: … got through=false want true
```
**GREEN** after D2. The two exact-touch cases are the ones a strict mirror of the limit predicate would still get wrong.

### E5 — UNKNOWN never cancels

Compile-red before D3 (`undefined: stopEntryGuardVerdict`), **GREEN** after: five unevaluable inputs (no price, no trigger, neither, unknown side, empty side) all return `stopGuardUnknown` with a non-empty reason, the enum's zero value is UNKNOWN, and `String()` says "unknown".

### E2 — the C# slot · `trader/wave_b_addon_slot_test.go`

**There is no C# harness in this repo** — no `.csproj`, no `.sln`, no `*Test*.cs`; NinjaScript compiles only inside NT8. **I chose the source-grep pin and say so plainly:** it proves the shipped `.cs` text, and D5's build gate proves the DLL NT8 actually loaded. Neither alone is sufficient; together they are.

RED:
```
--- FAIL: TestAddonEntryOrderPassesTheTriggerInTheStopSlot (0.00s)
    entry CreateOrder passes a literal 0 into the stopPrice slot (limit="orderPx" stop="0")
    — a StopMarket built this way has a ZERO trigger and rests inert forever
--- FAIL: TestAddonSubmissionLogNamesTheFourValuesItSent (0.00s)
    the submission log does not name action= / type= / limitPrice= / stopPrice= …
```
**GREEN** after D1. The pin also asserts each slot is selected by its own branch (`isLimit` / `isStopEntry`), not merely that the two names differ, and that the in-file bracket-SL control still has its shape.

### E6 — the build floor · `trader/ninjatrader/stop_entry_wire_test.go`

RED:
```
--- FAIL: TestStopEntryRefusedOnPreStopSlotBuild (0.17s)
    a stop entry was sent to an AddOn that predates the stop-slot fix — it would go out with a ZERO trigger
```
**GREEN** after D5. The existing `TestPlaceStopEntryFrameOnLoopback` was reseeded from `FarSideBuildE7` to `MinAddonBuildStopSlot` **by import** (A24 — a fixture must not hold its own copy of a constant).

### E1 — the payload · **honestly not a red→green pin**

The Go payload was never the defect. `PlaceStopEntry` already rounds the trigger and sends `OrderType:"stop_entry", StopPrice: entry`, and `tcp_trader.go` / `SignalPayload` needed **no change** — I looked for a reason to override the owner's ruling and found none. Rather than write a pin that cannot fail, I **strengthened the existing loopback test** with the assertion it lacked: on both sides, a stop entry's `limit_price` must be **0**, so a trigger leaking into the limit slot would rebuild the 2026-09-04 defect from the Go end. It is a regression guard and is **not counted as a red→green pin**.

### E7 — the replay · `trader/wave_b_replay_test.go` · **CORRECTS THE DISPATCH**

The dispatch expects "ONE well-formed submission". Replayed against the **investigators' measured** 21 cycle prices (their number, not the dispatch's framing of one arm) the fixed path yields **ZERO submissions**: at every one of the 21 cycles the market was 50.00–102.75 points below a sell-stop trigger, so the correct answer is 21 cancels and no order at all. A well-formed submission appears only in the valid-side replay (price 29650.00 above the trigger → rests → one placement). Arm 35's LIMIT case replays byte-identically, including that a limit **at** its own price still rests while a stop **at** its trigger does not.

Proven to bite by restoring the inverted semantics — 21 real failures, first three:
```
--- FAIL: TestReplayNY0904S2ThroughTheFixedGuard (0.00s)
    cycle 0: price 29515.25 is 75.25 pts BELOW a sell-stop trigger 29590.50 and must never be placed
    cycle 1: price 29524.75 is 65.75 pts BELOW a sell-stop trigger 29590.50 and must never be placed
    cycle 2: price 29540.50 is 50.00 pts BELOW a sell-stop trigger 29590.50 and must never be placed
```

### E8 — A29, built ≠ wired

`TestStopEntryGuardHasAProductionCallSite` reads `armed_executor.go` and requires the stop branch to call `stopEntryGuardVerdict`, to **not** call the limit predicate with the trigger, to keep `limitMarketableWrongSide(price, r.EntryPx, r.Side)` on the limit branch, and to distinguish `ErrAddonBuildTooOld`. Proven to bite by reverting the call site:
```
--- FAIL: TestStopEntryGuardHasAProductionCallSite (0.00s)
    the stop-entry branch does not call stopEntryGuardVerdict — the guard is built but not wired
    the stop-entry branch still calls the LIMIT predicate with the trigger — the inversion is back
```

### D4 boot-line pin — proven to bite by replacing the resolved `slots` with a literal:
```
--- FAIL: TestStopEntryBootLineIsRead (0.00s)
    a pre-stop-slot build must not claim slots=stop_price: 🎯 stop-entry: slots=stop_price … build_id=2026-09-03-f12 … match=NO
    an unheard-from AddOn cannot prove the slot: 🎯 stop-entry: slots=stop_price … build_id=none … match=NO
```

### E4 — limits unchanged
`TestLimitMarketableWrongSide` is **unedited and passing**. `limitMarketableWrongSide` is byte-identical. **There is no placement golden in `trader/` or `provider/`** — every golden in this repo is a kernel *prompt* golden — so E4's "golden diff must be EMPTY" is satisfied as: `git diff origin/dev...HEAD -- kernel/testdata` returns **0 files**, and `go test ./kernel/` is ok. No golden was regenerated; this wave touches no prompt.

### E9 — the full suite, at this head

| gate | command | result |
|---|---|---|
| build | `go build ./...` | rc=0 |
| Go suite | `go test ./...` | **28 packages ok, 0 FAIL** |
| goldens | `go test ./kernel/` + `git diff -- kernel/testdata` | ok · **0 files changed** |
| tsc + vite | `cd web && npm run build` | ✓ built in 4.14s |
| vitest | `cd web && npm test` | **44 files, 345 tests, all passed** |

(`web/node_modules` was absent in this worktree; `npm ci` was run **in the worktree only** — the main tree was never touched.)

---

## F2 — THE FOUR LIVE PROOF LINES: **NOT YET RECEIVED — awaiting the owner's NT8 recompile**

None of these can exist until the owner copies `ninjascript/VLTraderTCPClient.cs` to the Documents AddOns folder, F5-compiles, and **fully restarts NT8** — and then a CME session opens. All four are owed:

1. **NOT YET RECEIVED** — the boot line reading `🎯 stop-entry: slots=stop_price · guard=stop-side · unknown=no-op · addon build_id=2026-09-05-g1 expected=2026-09-05-g1 match=yes`. Until the recompile it will correctly read `slots=unproven(addon build) … build_id=2026-09-03-f12 … match=NO`.
2. **NOT YET RECEIVED** — an NT8 order-log line for a stop ENTRY reading `Limit price=0 Stop price=<trigger> … Type='Stop Market'`. **This is the acceptance criterion, and it must be the price slot — not "it rested and cancelled".** That was the 2026-08-31 criterion and the order that passed it carried `Stop price=0`.
3. **NOT YET RECEIVED** — that order reaching NT8 state **`Working`**. No stop entry has ever reached `Working` in the system's history (0 of 22).
4. **NOT YET RECEIVED** — an `order_snapshot` frame carrying `build_id=2026-09-05-g1` with the entry's **non-zero `stop_price`** parsed back off the wire. Assert the parsed struct field or `limit_price == 0` beside it: `stop_price` is `omitempty`, so a zero **vanishes** from the persisted JSON and an assertion written as "stop_price is absent" would pass on the bug.

Read-only tonight: **only `2026-09-03-f12` has ever been received** — 5,887 `nt8_order_snapshots` rows, 100% of them, most recent 2026-09-05 21:48:23 CT. The AddOn currently deployed is byte-identical to the repo at base, so NT8 genuinely holds the defect right now, and from this commit the repo and the deployed DLL diverge until the recompile.

---

## A15 — LIVE-SURFACE TRUTH

- **Nothing is deployed.** The running binary is rev `36648655`, booted 2026-09-04 13:25:47 CT. It has **none** of this. Today is Saturday 2026-09-05, CME closed, engine idle.
- **`STOP_ENTRY_SEAM=on` in `/home/hoang/nofx/.env`.** This is a LIVE path, not a dormant one — that is how 21 malformed orders reached a broker.
- **The cutover state is clean.** No working stop entry exists anywhere. The only non-terminal rows are ids 92/94/96/98 (`superseded`) and **104, 105** (`armed`), all `kind='limit'`, none carrying a `signal_id` — so nothing is resting at the broker.
- **After a Go-side deploy but before the NT8 recompile:** every stop entry is REFUSED and counted (`stop_entry:addon_build`), the boot line reads `match=NO`, and **limits behave exactly as today**. That is fail-closed and is the intended intermediate state.
- **`GUIDE_BUILT_REV` is NOT bumped** (`web/src/guide/types.ts:6`, still `36648655…`). It must equal the **deployed** rev, which does not exist yet; setting it to a sha that never ships would make the drift banner lie in the other direction. **Owner-side at cutover: bump it and rebuild `web/dist` BEFORE the boot** (boot-5 rule), not in the marker commit. The Guide's *content* is updated in this PR, satisfying the substantive half of the GUIDE CONTENT LAW.

---

## WHAT IS NOW POSSIBLE THAT WAS NOT

1. **A stop entry can actually work.** Lifetime fill rate was 0/22 — the feature has never once placed an order NinjaTrader could trigger. After the recompile it can.
2. **A valid stop entry can be placed at all.** The inverted guard cancelled 100% of the valid cases; cell 3 never occurred live only because the market never offered one.
3. **An already-through stop is refused instead of admitted.** The 21 admissions of 2026-09-04 would now be 21 cancels naming the relation: `accepted through (stop side): price 29515.25 <= trigger 29590.50`.
4. **A guard that cannot answer does nothing, out loud, and is counted.** Previously an unevaluable state fell through to placement.
5. **A stale AddOn is refused rather than sent something it will mis-execute.** The old floor could not refuse the malformed build — `"2026-09-03-f12" >= "2026-08-30-e7"` is true.
6. **The AddOn's own log can now witness the defect class.** It prints the arguments handed to `CreateOrder`, so the AddOn log and NT8's order log must agree or disagree visibly.
7. **The posture is stated at boot from resolved values**, so an inverted guard, an unproven build or a cancel-on-unknown each change the line rather than leaving it lying.

---

## ROLLBACK — both halves, independently

**Go half (safe to revert alone).** `git revert` the code commit, or roll back the binary and `deploy/RELEASE`. Reverting Go alone with the NEW AddOn compiled in NT8 restores the inverted guard and the old build floor — the AddOn would then build correct orders that the inverted guard admits at the wrong moments. **Do not do this while a plan can arm a reclaim; disable `STOP_ENTRY_SEAM` first** (`STOP_ENTRY_SEAM=off` in `.env` + restart) — limits are unaffected.

**C# half (safe to revert alone).** Restore the previous `VLTraderTCPClient.cs`, F5-compile, restart NT8. The reverted AddOn reports `2026-09-03-f12`, which the new Go floor **refuses** — so with the Go half deployed, rolling back the AddOn fails CLOSED: no stop entries at all, loudly, counted, with limits untouched. That is the designed behaviour and needs no coordination.

**The dangerous combination is exactly one: the NEW AddOn with the OLD Go binary.** The build floor cannot stop it (the old floor is permissive), and the old inverted guard would then hand correctly-formed orders to the market at already-through prices. **Deploy Go FIRST, recompile the AddOn SECOND** — Go-first is fail-closed at every intermediate moment. If a rollback must cross both halves, roll the AddOn back first, then the binary.

**Fastest kill switch, needing neither rollback:** `STOP_ENTRY_SEAM=off` + restart. The stop-entry branch short-circuits before any of this code runs, and the limit path is untouched.

---

## SURPRISES — recorded, not acted on (A23)

1. **The two bugs masked each other.** Not in the dispatch. C1 alone is a regression. This drove the whole sequencing and is why both are in one commit.
2. **The stop-entry path has never worked, ever.** 22 lifetime submissions across two days, `Stop price=0` on every one, 0 ever reaching `Working`. Wider than the dispatch's single-session framing.
3. **NT8 accepted a zero-trigger stop silently.** No reject, no error, no counter — and none on our side either. A silent-refusal path in both directions.
4. **The capability's own acceptance proof was the bug.** The 2026-08-31 E7 far-side proof order went out as `Limit price=28700 Stop price=0` and was recorded as a PASS, because the criterion was "it rests and cancels" and a zero-trigger stop rests perfectly.
5. **Up to NINE stop entries rested concurrently for a SINGLE arm slot** (`nt8_order_snapshots` 1604-1606 and 1664, `order_count=9`, 8 Accepted + 1 Initialized, none cancel-pending). Between 10:22 and 10:40 the re-arm loop placed 8 new orders with no cancels. Harmless only because all nine were inert. **With D1 fixed and this cancel path unchanged, one arm could take nine positions on an account capped at `maxFuturesContracts=2.0`** whose one-open-position rule is enforced Go-side at arm time and cannot reach orders already resting at the broker. The cancel path is outside this wave's footprint (A31) — **this needs an owner ruling before the seam runs again on a live session.**
6. **A row whose `kind` contradicts its `condition`.** Id 38 carries `kind='stop_entry'` with `condition='sweep_reclaim'`, but `ArmKindFor("sweep_reclaim")` returns `ArmKindLimit` — consistent with `UpsertArm` updating `condition` in place while leaving `kind` stale. `kind` and `condition` can disagree on a real row.
7. **`TestStopEntryKnobDefaults` (`trader/split_entry_test.go:267`) reads the AMBIENT environment** with no `t.Setenv`, asserting `STOP_ENTRY_SEAM` defaults off. It passes in a clean shell and would FAIL in one that has sourced `/home/hoang/nofx/.env`. All runs in this report were from a shell that has **not** sourced `.env`. Outside the footprint; not touched.
8. **Guide drift already on dev, unrelated to this wave.** `plays.ts` lists `STOP_ENTRY_SEAM (off)` in the *defaults* line while `.env` has it **on** and every boot line since 09-01 reads `stop_entry_seam=ON`. It is defensible as a default list, but it reads as the live value on the very path this wave fixes. **Left alone — it needs a ruling**, and rewriting a defaults list into resolved values is a different change.
9. **AUDIT-CHECKLIST class 75's contract test does not exist.** `grep -rln 'SYSTEM-MAP|SYSTEM_MAP|system-map' --include=*.go .` returns nothing. Nothing mechanically catches a stale SYSTEM-MAP; the class-75 update in this PR is discipline, not enforcement, and must not be reported as machine-verified.
10. **100 broker rejections on the SIM account** reading `Your maximum position limit has been met… Scope: all Rule #<id>`, across eight dates, plus 4 `User-initiated trading lockout is currently active`. Entirely outside this footprint.

---

## SCOPE (A31)

`git diff --name-only origin/dev...HEAD` — 13 files, every one inside the footprint:

```
docs/superpowers/AUDIT-CHECKLIST.md      docs/superpowers/SYSTEM-MAP.md
ninjascript/VLTraderTCPClient.cs         provider/ninjatrader/order_snapshot.go
provider/ninjatrader/tcp_framing.go      trader/armed_executor.go
trader/ninjatrader/stop_entry_wire_test.go   trader/ninjatrader/tcp_trader.go
trader/wave_b_addon_slot_test.go         trader/wave_b_replay_test.go
trader/wave_b_stop_guard_test.go         web/src/guide/content/guards.ts
web/src/guide/content/plays.ts
```

The forbidden-name grep (`entry_gate|arm_kind|composeArmStop|arm_stop_anchor|reaper|planner_prompt|validator|levels|wake|bars|detector_record|trade_excursions|close_sync`) returns **NONE**. No gate leg, no armable set, no stop composition, no R:R floor, no tick cap, no flip of the waterfall design. `trader/ninjatrader/tcp_trader.go` was edited for **one reason, quoted**: the D5 build gate at `:502`. Its `SignalPayload` construction (`:524-537`) is **untouched** — the owner's ruling that the wire already carries the trigger correctly is confirmed end to end.
