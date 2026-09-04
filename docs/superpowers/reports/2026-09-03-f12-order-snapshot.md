# F12 — the NT8 order-snapshot frame + AddOn build_id

**Branch** `fix/f12-order-snapshot` @ `c84bd247` (off `origin/dev` 955d4ac8)
**Status** BUILT AND GREEN. **NOT DEPLOYED, AND NOT PROVEN.**
Two halves remain, both the owner's: F1 the Go boot, F2 the NT8 AddOn reload.
Per A20/class 6 this wave is proven only by a RECEIVED frame, and none has been.

---

## 1 — WHAT WAS WRONG

The flat gate has five legs. Four of them ask NinjaTrader. Leg 4 — *are there
working orders resting?* — asked `armed_orders`, **our own ledger**, and said so
in its own source string:

> `armed_orders ledger (no NT8 order frame — F12 open)`

A leg answered by our own bookkeeping cannot detect the one failure a flat gate
exists to detect: the ledger and the broker disagreeing. It had passed on that
basis at every cutover since 35.

The consequence was not theoretical. On **2026-09-02** the 0B cutover waived flat
with **position 588 open**, and the resting stop **could not be verified** —
nothing on the wire could say whether it was there. The rule that followed ("no
override with a position open") was the right answer to an unanswerable question:
a blanket refusal standing in for a check with no data.

Separately, `VL_BUILD_ID` is bumped in our source but NT8 keeps executing whatever
DLL it last compiled, so **no frame could say which build was running**.

---

## 2 — THE FRAME

`order_snapshot`, AddOn → Go, every 30 s (on the heartbeat loop) **and**
immediately after any order state change.

```json
{"type":"order_snapshot","payload":{
  "account":"Sim101","build_id":"2026-09-03-f12",
  "emitted_at_ms":1788480000000,"reason":"periodic|state_change",
  "orders":[{"order_id":"NT-1","name":"VL-S1-entry","action":"buy|sell",
             "type":"limit|stop|stop_limit|market","limit_price":29450.25,
             "stop_price":0,"quantity":1,"filled":0,"state":"Working",
             "oco":"oco-1","symbol":"MNQ"}]}}
```

**Account-scoped, not symbol-scoped**, and that was a correction (see §5). An
empty book is `orders: []` — an **explicit empty, never silence**. "No orders" and
"no answer" are different claims and leg 4 must tell them apart.

`hello` now carries `build_id` too (`omitempty`, so an older AddOn's wire stays
byte-identical).

---

## 3 — FILE:LINES

| What | Where |
|---|---|
| Frame type | `provider/ninjatrader/tcp_framing.go:131` |
| Parse (refuses an unaddressable frame) | `provider/ninjatrader/order_snapshot.go:105` |
| Symbol filter | `order_snapshot.go:87 WorkingOrdersFor` |
| Cache: put / age / ever-received | `order_snapshot.go:157` · `:186` · `:168` |
| Build line + expected constant | `order_snapshot.go:218` · `:214` |
| Snapshot line | `order_snapshot.go:236` |
| Server case (feeds the EXISTING far-side build field) | `provider/ninjatrader/tcp_server.go:1763` |
| Leg 4 | `trader/f12_leg4.go:33 Leg4FromBrokerAt` |
| Override guard | `trader/f12_leg4.go:129 OverrideAllowedAt` |
| Leg-4 source label (kills the boot-line literal) | `trader/f12_leg4.go:214` |
| Shared staleness constant | `trader/f12_leg4.go:209 DefaultOrderSnapshotSecs = 30` |
| Store table | `store/nt8_order_snapshot.go` |
| C# emitter | `ninjascript/VLTraderTCPClient.cs:1478 SendOrderSnapshot` |
| C# constants | `:42 ORDER_SNAPSHOT_INTERVAL_MS` · `:55 VL_BUILD_ID = "2026-09-03-f12"` |
| Protocol | `ninjascript/vltrader_tcp_PROTOCOL.md` (new section) |
| Guide | `web/src/guide/content/guards.ts` §6 |
| Checklist | class **67** |

---

## 4 — E1 RED → GREEN

**RED**, before any implementation:

```
# nofx/provider/ninjatrader [nofx/provider/ninjatrader.test]
order_snapshot_test.go:38:12: undefined: ParseOrderSnapshot
order_snapshot_test.go:101:7:  undefined: NewOrderSnapshotCache
...
# nofx/trader
f12_leg4_test.go:31:9:  undefined: Leg4FromBrokerAt
f12_leg4_test.go:141:16: undefined: OverrideAllowedAt
```

**GREEN:** provider **14** cases, trader **17** cases. Full Go suite **0
failures**; vitest **42 files / 325 tests**; `tsc` clean; goldens pass.

Leg 4's five failure modes are each pinned: a working order (named with state,
type and price), a stale book (age quoted), a live link with no book for the
account, a broker/ledger mismatch (both counts and both id lists), and the
loud-fallback state. The override guard is pinned allowed / no-stop / wrong-price
/ stale / absent.

---

## 5 — SURPRISES (A23), both mine

**5.1 — I invented two C# APIs.** The first emitter used `ToUnixMs()` and an
`instrument` field. **Neither exists** in `VLTraderTCPClient.cs`. I cannot run a
C# compiler here, so this would have failed on the owner's machine at F5 — the
worst place to find it. Replaced with a local epoch helper mirroring
`VLBarsSubscriptionManager.NowUtcMs`, and the symbol read from each order's own
`Instrument`. **Every C# symbol in this wave was then re-checked against the file
rather than against my memory of it.**

**5.2 — Symbol-keying the frame is what forced the invention.** `Account.Orders`
is an account collection and the AddOn holds no persistent per-instrument handle,
so a per-symbol frame had nothing legitimate to key on. Making the frame
account-scoped removed the invention *and* fixed a real modelling error: a
symbol-keyed empty book cannot be expressed, because with no orders there is no
symbol to file it under. The Go cache, leg 4, the override guard and their tests
were all re-keyed.

**5.3 — a boot line that would have started lying.** The class-33 line printed
`leg4=ledger` as a **literal**. This wave makes that false at every boot, so it
now takes the resolved source as an argument, pinned over all four states. Its
existing test was migrated with the reason rather than weakened.

**5.4 — a second build-id source, avoided.** My cache initially kept its own
`build_id`. The server has owned that value since the E7 handshake
(`farSideBuild` / `FarSideBuildID`). Two stores of the same received value are
free to disagree and the boot line would then have to pick one — which is how a
"received" value quietly becomes a guess. The snapshot and the hello now feed the
existing field; the cache keeps none.

---

## 6 — F1 vs D3, resolved toward F1

They disagree about the never-received case. D3 says a missing snapshot FAILS
leg 4; F1 says the Go side boots *before* the AddOn is reloaded and must fall
back to the ledger with the source printed.

Resolved toward **F1**, because a hard fail would make the wave undeployable — the
Go half must boot first by construction. So:

- **never received** → ledger answer, source names it, every read. *Never a
  SILENT fallback* is the rule; *never a fallback* is not.
- **received before, now absent for this account** → **FAIL**. That is a
  regression, not a cold start, and `EverReceived()` is what separates them.
- **stale (> 2N)** → **FAIL** with the age.
- **mismatch** → **FAIL** with both counts.

---

## 7 — WHAT IS NOT PROVEN

**No `order_snapshot` frame has been received.** The build line will read:

```
🔌 nt8 addon: build_id=none expected=2026-09-03-f12 match=NO (no frame carrying a build_id received yet)
🔌 order_snapshot: last=none · leg4 source=ledger (no snapshot yet)
```

That is the correct output for this moment, and it is also the measure of what is
left. Until the owner recompiles the AddOn and NT8 sends a frame, **leg 4 still
reads the ledger and the override guard still refuses** — the pre-F12 behaviour,
now stated out loud instead of assumed.

---

## 8 — CUTOVER, both halves

**F1 — the Go boot (owner GO).** Five-leg gate `ready:true` → clean-clone build
(`vcs.modified=false`) → `deploy/RELEASE` from the main tree BEFORE the kill →
swap → owner runs `kill -9` → boot within 90 s → marker after the boot, same
tree, pushed, five-reference check. Expected: `match=NO`, `source=ledger`.

**F2 — the AddOn (owner's hands).** `cp ninjascript/VLTraderTCPClient.cs` to
`C:\Users\hoang\Documents\NinjaTrader 8\bin\Custom\AddOns\` → **F5** → **full NT8
restart**, in a flat window. Editing the repo file alone does nothing.

**PROOF to quote (class 6):** the first received frame, from the Go log with its
timestamp —
`tcp_server: far-side AddOn build_id=2026-09-03-f12 (from order_snapshot)` — and
the health line flipping to `match=yes` with `leg4 source=broker`.

**Rollback, both halves.** Go: the prior binary, by content, plus `RELEASE`. NT8:
**keep the previous DLL with its build id in the name** (A13) — the AddOn is the
half with no `git revert`, and the old build is identifiable only by what its
frames say.

---

## 9 — NOTE FOR ANOTHER LANE

Checklist **66 is taken on dev** (the two-price-spaces class, renumbered at
combined boot 4). My own UI-serving wave (`fix/ui-serving-path-clean`) still
claims 66 on an unmerged branch and **needs renumbering before it lands**. This
wave took **67** rather than the next free number on its own branch, for the same
reason.
