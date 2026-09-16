# 101 — 5m CHART DEPTH + NT8 HISTORY AT SUBSCRIBE + CHART ACROSS THE ROLL

**Branch** `fix/nt8-history-and-chart-depth` · base dev `8f5f26ee` · running rev `3ce4281a4b6b`
(`/api/health` and `/proc/1834811/exe` agree). **STATUS: STAGED-AND-GREEN at the cutover gate** — stopped at Section C (A23), the
premise refuted, scope revised by the CTO after verifying the refutation, built under the
revised scope. (3a)/(3b) — the AddOn re-request and the all-timeframe store rehydrate —
are HELD pending the owner's direct word in his chat.

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



---

## D · WHAT SHIPPED (revised scope, CTO rulings 11:35 / 11:50 CT)

| | files | pinned by |
|---|---|---|
| **D1'(1)** an empty replay cannot re-arm the scale check | `bar_cache.go` | `TestZeroBarReplayDoesNotReArmTheScaleCheck` (5m+1m), `TestZeroBarReseedDoesNotRejudgeAVerifiedSeed` |
| **D1'(2)** the check judges ADJACENT bars only (≤2 intervals); older → SKIP + WARN with both ages, stay armed | `bar_source.go`, `bar_persist_wire.go` | `TestStaleReferenceIsSkippedNotDropped`, `TestPastGapBackfillDoesNotDrop`; rule 3 `TestAdjacentRealBreakStillDrops` |
| **D1'(3) narrow** a confirmed break is COUNTED and the P0 says "LIVE-ONLY UNTIL NT8's NEXT FULL REPLAY" for non-1m | `telemetry/scale_break.go`, `bar_persist_wire.go` | `TestScaleBreakDropIsCountedWithItsBars` |
| **D2** chart across the roll, flag ON [O] | `store/bar_history_across_roll.go`, `api/handler_klines.go`, `market.Kline.Contract` (additive, omitempty) | `TestPriorContractsFillBehindTheCurrentContract`, `TestKlinesAcrossRollLabelsEveryBarAndKeepsTheStep`, `TestDecisionReadersStayCurrentContractOnly` |
| **D3(b)** one horizon WARN per (symbol,tf,why) per 5 min, callers aggregated | `bar_horizon_warn.go` | `TestFourCallersInOneWindowEmitOneLineWithCallersAggregated`, `TestHorizonWindowIsFiveMinutesAndReArmsWithCallers` |
| **E1** `🧯 nt8 history at subscribe:` per-tf received/asked, n/a for absent | `history_at_subscribe.go`, `tcp_server.go` (enqueue count) | `TestHistoryAtSubscribeLineRendersResolvedValues`, `…AbsentIsNotZero` |
| **A12** class 127, SYSTEM-MAP bars section, RULEBOOK §A | docs | same commit |

**No AddOn change. No `.cs` copy. `bar_history.go` byte-identical. `pairsToRehydrate` untouched.**
A31 grep against the do-not-touch list: none. Class 100 deletions vs dev: none.

### D1'(4) — the three facets from batch 2, decided or named LEFT

- **(c) drop all replay bars or only the offending seed — DECIDED: drop all, and here is why.**
  The ring stamps source per bar, not per seed; after a merge (`mergeSeedKeepingLive`) two
  replays' bars are indistinguishable. Dropping "only the offending seed" would require a
  seed id on every bar — a schema change beyond this wave. With rules 1–2 in place a drop
  only happens on an ADJACENT real break, so the destructive response now has the
  precondition it lacked. Documented in `bar_source.go`.
- **(a) the 20×-median-body rule blinds higher timeframes — LEFT**, named: on 15m/1h the
  median body is large enough that a real shift can pass the range check. It is 104's OWED
  item from batch 2 and touching it changes what a break IS, which is a market-belief
  change this wave must not make (research law). Rules 1–2 reduce its exposure (fewer
  spurious checks) without changing its threshold.
- **(b) the re-seed clears `seedOffScale` before the persister's replay hold reads it — LEFT**,
  named: rule 1 now means an EMPTY re-seed no longer clears it (the early return precedes
  the delete), which closes the 09-16 path; a NON-empty re-seed still clears it, which is
  correct (a new replay is a new verdict) but races the hold. Ordering the hold's read
  before the clear is 104's.

## E · MUTATIONS — line quoted, sed confirmed, build GREEN, RED quoted

| # | file:line | mutation | died on |
|---|---|---|---|
| M1 | `bar_cache.go:254` | `if len(bars)==0 && len(existing)>0` → `if false` | rule 1's own fixture: "an empty re-seed re-opened a verdict already given: historical 30 → 0" — **after the 09:22 fixture alone SURVIVED it** (rule 2 also covers 09:22) |
| M2 | `bar_source.go:186` | drop the adjacency guard | stale-reference and gap-backfill |
| M3 | `bar_source.go` | adjacency = 1000 intervals | stale-reference |
| M4 | `bar_horizon_warn.go:193` | caller back in the key | "four callers … got 4" — **first attempt's sed did not apply (delimiter clash) and passed vacuously; re-run correctly, recorded in a08e0625** |
| M5 | `handler_klines.go` | +292 on prior rows (a back-adjust) | "the basis step is 0.00, want 292.00" |
| M6 | `history_at_subscribe.go` | absent rendered as 0 | "absent must be n/a" |

**A29** call sites: `stampFadePermission`-style wrappers not needed here; each new function's
production caller counted: `IncScaleBreakDrop` 1, `ScaleBreakCounts` 1, `OnScaleCheckSkip` 1,
`barHorizonCallersTxt` 1, `HistoryAtSubscribeLineFor` 1, `ChartAcrossRollResolved` 1,
`PriorContractBarsBefore`/`FirstLiveOn` 2 (api/ only; **0 under kernel/, trader/, provider/** — E4).

**E6** at the merged head `8f5f26ee`, 11:42 CDT (outside 12:00–13:30): **SUITE 32 ok / 0 FAIL**;
lock suite 101/0. One pre-existing pin re-pointed: `TestKlinesNinjaTraderStoreDepthContractFiltered`
asserted the 09-14 rule the 09-16 ruling reverses; it now pins the ruling and, with the flag
off, the 09-14 behaviour exactly.

## A15 · WHAT THE OWNER WILL STILL SEE WRONG

1. **The running binary `3ce4281a` has the 09:22 defect.** Until this boots, any feed flap
   can still empty every non-1m ring. The 5m chart on `:8080` stays 11 hours deep.
2. **The served bundle is `index-COvgwytr.js` with `limit=1500`** — L5's dist rebuild is
   needed for the chart to ask past 1,500 at all; the `contract` label on each kline arrives
   with this boot but the FE that renders it is L5's.
3. **After this boot, a TRUE scale break still leaves 5m+ live-only** until NT8's next full
   replay — (3a)/(3b) are what change that, and both wait on the owner's direct word.
4. **The five `bars=0` reconnect replays** (07:20, 07:22, 07:34, 08:33, 09:22 — BarsRequest
   run while the feed was down) are an observation for 104; the AddOn returns nothing and
   says so honestly. With rule 1 they are harmless; they are still wasted requests.
5. **"1H EMPTY at boot while 1536 emitted"** — the horizon line at 22:15:1x likely preceded
   the drain. Noted, unverified; did not fall out of the fixtures.
6. **The running binary's log paths say `nofx-deploy-r4/main.go`** — built in a directory
   not named `nofx` (class 75). This build is from a clean clone named `nofx`.
7. The **1833 vs 1832** in the 1m fixture is the ring cap trimming one bar on the live upsert
   — 2,000 seeded + ~667 live > 2,500 — the same arithmetic as the live "1832 dropped". A
   non-defect that reads like one.

## STORE CENSUS AT THE GATE — nofx-93, read-only, against 10f424b0 semantics

Accepted and reproduced. What matters for this PR and the next:

- **10f424b0's 1m rehydrate on MNQ 12-26 selects 2,500 rows, all `live`, 0 `historical_import`,
  filtered 0** (T2). The import yes/no is structurally open, numerically moot for this boot.
  T4: zero `replay:off-scale` rows on MNQ.
- **D2's boundary is right where it matters.** `FirstLiveOn(MNQ,5m,12-26)` = 09-14 10:00 CT,
  dense at 5-minute gaps from the first row; 93's "7 before cutoff" are the 10:00–10:30 bars
  just ahead of their 10:34 CT cutoff, not July. On 1h: 09-14 02:00 CT, 55 rows, largest gap 2h.
- **93's T3 finding, verified here and named LEFT for 104 (the wire/persist path):** MNQ 12-26
  carries ~9 `live`-stamped rows per tf dated BEFORE the roll — 1d from 09-02 at **rowid 438391**
  (db max ≈2.27M), 4h from 09-11, 3d 18 back to 08-10, 1w 8 back to 07-17 — old rows recently
  UPDATED to `live` + `12-26`. That is a fixed-count closed-bar catch-up delivered as
  `bar_update` right after a (re)subscribe, stamped `live` by the persister because the
  frame type says so, and upserted over real rows because only historical-over-live is
  blocked. **A `bar_update` for a bar that closed before the subscribe is a replay.** The
  mis-stamp is in the wire/persist path, not a separate higher-tf writer. For D2 it is
  harmless (the chart labels contract, not source, and the prices are December's); for
  (3b)'s guard (ii) it means the "older than boot" test cannot be store-side — it must be
  ring-side, against this process's first live frame for the key, whatever the stamp says.
  The CTO's optional in-scope WARN (persist path: a `bar_update` whose close predates the
  subscribe) is **LEFT**: a diagnostic line is not worth a rebuild of a binary already parked
  green at a gate blocked on the main tree; it is one fixture and one line for 104.
- The 09-26 pre-wave `live` rows on every tf (3m 4805 of 5453 before the 09-11 cutoff …)
  are the migration's stamp — the `source` column did not exist before 09-10. Age alone
  cannot distinguish them from real live rows. Same conclusion: ring-side, not store-side.

## OWED (not this PR)

- (3a) AddOn re-request on a confirmed break; (3b) all-tf store rehydrate with the four guards
  — **held for the owner's word**, then as the CTO specified.
- (3c)/(3d) follow (3b): regime fallback arm, `weeklyBias.ts:122` (to L5), the rehydrate
  done-line, `bar_horizon_warn_test.go:243`, both headers — none change until (3b) does.
- The Guide paragraph for `web/src/guide/content/status.ts` — sent to L5 as text.
- Store-count re-run at the gate — nofx-93's offer, accepted.

## ROLLBACK

Go half: `mv nofx-bin nofx-bin.failed.<rev> && mv nofx-bin.old.<prev-rev> nofx-bin && echo <prev> > deploy/RELEASE && kill -9 $(pgrep -x nofx-bin)`. No AddOn half. No migration (Kline.Contract is wire-only; no DB column). `NOFX_CHART_ACROSS_ROLL=off` disables D2 without a rebuild.
