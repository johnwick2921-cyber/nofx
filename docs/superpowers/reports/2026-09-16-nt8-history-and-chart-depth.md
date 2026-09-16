# 101 — 5m CHART DEPTH + NT8 HISTORY AT SUBSCRIBE + CHART ACROSS THE ROLL

**Branch** `fix/nt8-history-and-chart-depth` · base dev `8f5f26ee` · running rev `3ce4281a4b6b`
(`/api/health` and `/proc/1834811/exe` agree). **STATUS: STOPPED at Section C** (A23) — the
dispatch's central premise is refuted by the AddOn's own log, and the real mechanism is on
the Go side. Reported to nofx-9e and the owner before any edit.

---

## THE FINDING — NT8 delivered the history; Go destroyed it eleven hours later

**Premise as dispatched:** "NT8 returns ~0 history at subscribe (historical=139)."

**Measured:** at 22:15:08 CT on 09-15 the AddOn emitted, per its own
`emitted bars_historical` lines:

```
MNQ|1M bars=2000   MNQ|5M bars=2000   MNQ|15M bars=2000   MNQ|30M bars=2000
MNQ|1H bars=1536   MNQ|2H bars=802    MNQ|4H (see below)  MNQ|1D bars=69
MNQ total: 13,054 historical bars
```

**NT8 honoured `bars_back`.** `📼 bar source: … historical=139` is a **store** census
(`bars.source='historical'`, live=57395 cannot be a 2,500-cap ring) — it counts persisted
replay rows, which the replay-hold deliberately keeps out of the store. It was misread as
"delivered".

**The 5m ring was 12 days deep until 09:22 on 09-16:**

| when (09-16) | MNQ 5m horizon |
|---|---|
| before 09:22 | `asked=5000 served=2131 span=285h55m` |
| after 09:22 | `asked=220 served=134 span=11h5m` |

**What happened at 09:22:11.** The data feed flapped five times that morning (07:20,
07:22, 07:34, 08:33, 09:22 — NT8 log: "BarsRequests will be recreated on recovery"). Each
reconnect re-subscribed and the AddOn emitted **`MNQ|5M bars=0`** — a zero-bar replay,
because the feed was still down when the BarsRequest ran. Then Go logged:

```
🚨 P0 — REPLAY AND LIVE ARE ON DIFFERENT PRICE SCALES for MNQ 5m at 09:22:11 CT:
last replay close 29320.50, first live close 29467.25, delta 146.75 pts (> 0.50%).
1999 historical bars DROPPED from the ring and refilled from the store's live rows
```

**`29320.50` is the 09-15 22:10 CT bar** — the LAST bar of the *boot-time* replay, eleven
hours earlier (store: `09-15 22:10 c=29320.5 live`). The detector compared it to the first
live bar after 09:22 and read the overnight move as a scale break.

**Mechanism, `provider/ninjatrader/bar_cache.go` `SeedHistorical` (:230):** the scale-check
re-arm `delete(c.liveSeen, key)` at :246 runs **before** the `len(bars)==0` return at :263.
A zero-bar replay re-arms the check without updating the reference bar, so the next live bar
is judged against whatever the newest historical bar in the ring happens to be — however old.
Then the refill (`rehydrateRingFromStoreWith`, `bar_persist_wire.go:178`) rehydrates **1m
only** (`pairsToRehydrate`, :478 `p[1] == rehydrateTimeframe`), so 5m/15m/1h/4h stay live-only
for the life of the process. Same event on MNQ 1m: 1832 dropped.

**Every hypothesis in D1 is refuted, for the record:**

| H | claim | verdict |
|---|---|---|
| H1 | NT8 local DB shallow for Dec | **No** — `db\minute\MNQ 12-26\` has 83 day-files, 06-12→09-16, 5–10 KB each |
| H2 | TradingHours template starves | **No** — `CME US Index Futures ETH`, and 2000/2000 arrived |
| H3 | rolling literal reaches BarsRequest | **No** — `resolved … => MNQ 12-26`; `Subscribe(symbol, tf, barsBack, instrument)` per tf; reconnect resolves the same |
| H4 | Tradovate caps depth | **No** — 2000 returned for every tf ≤ 30m |
| — | **zero-bar reconnect replay re-arms the scale check; false positive drops the seed; refill is 1m-only** | **Yes** — proven above |

## C · EVIDENCE, each reproduced at `3ce4281a`

- **C1** ✓ (after 09:22) `served=154 span=12h45m oldest=process start`; **✗ before 09:22** — `served=2131 span=285h55m`. The premise held only after the destruction.
- **C2** ✗ as stated — `historical=139` is a store count; NT8 emitted 13,054.
- **C3** ✓ at 22:15:1x — but `1H bars=1536` was emitted at 22:15:08; the horizon line likely preceded the drain (ordering, not emptiness). Unverified; noted.
- **C4** ✓ `26 pair(s) SKIPPED as tf!=1m`, `bar_persist_wire.go:478`.
- **C5** ✓ `bar_history.go:483,540,568 Where("contract = ?")`; `handler_klines.go:385`.
- **C6** ✓ MNQ 12-26 = 581 5m rows (live from 09-14); 09-26 = 20,197; store 284,051 rows, 19 contracts, back to 2022-04-11.
- **C7** ✓ served bundle `index-COvgwytr.js` has `limit=1500`, no 5000; dev-tip source has `limit=${limit}`. L5's.
- **C8** ✗ as stated — the AddOn **does** log the count (`emitted bars_historical <key> bars=<n>`) and the resolve (`resolved MNQ -> MNQ ##-## => MNQ 12-26 (rolling->MNQ 12-26)`). D3(a) is unnecessary.

Position 606 reads CLOSED, not OPEN as dispatched. The running binary's log paths carry
`nofx-deploy-r4/main.go` — built in a directory not named `nofx` (class 75).

## WHAT THIS MEANS FOR THE FIX (proposal, not started)

- **D1 as dispatched is moot.** No AddOn change is needed for depth; the AddOn is correct.
- **The real D1 is Go-side, three lines of intent:** (1) a zero-bar replay must not re-arm the
  scale check; (2) the check must compare *adjacent* bars — last replay bar and first live bar
  within one interval — never a reference hours old; (3) after a true drop, refill every
  timeframe from the store (or re-request from NT8), not 1m only.
- **D2 (chart across the roll) stands** — owner ruling, independent of the above.
- **D3(b) stands** — 8,258 horizon WARNs since boot; my own `fade_facts.go:73` is a caller.
- **D3(a) is unnecessary**; the count is already logged.

Awaiting a ruling on the revised scope before any edit.
