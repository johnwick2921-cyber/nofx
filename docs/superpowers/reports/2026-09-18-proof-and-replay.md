# W-PROOF-AND-REPLAY — DS-104 report (2026-09-18)

**Lane:** DS-104 · branch `docs/proof-and-replay-0918` · claim `12e096d044cc`
**Scope:** read-only research and verification. No code, no bot changes.
**Basis:** worktree `/home/hoang/nofx-104` off `origin/dev` b70fc6ca (2026-09-18 02:28 CT).
**Data:** backup copy `~/nofx-backups/pre-bars-key-20260918-022516.db` (consistent copy, 02:25 CT) for everything historical; live DB `data/data.db` read-only for rows after 02:25 CT only.
**Laws:** L1–L14 applied. Evidence tiers [A]/[B]/[C] on every claim.

## A. STOP-ENTRY LIVE PROOF — in progress

- Watcher running since 08:20 CT: `/tmp/ds104-stopentry-watch.sh` → log `/tmp/ds104-stopentry-watch.log` (60s poll, journal since 07:52 CT, patterns `stop_entry|stop-entry|⚔️ arm|armed_fill|fill@|🧾 cancels|nt8 addon|order_snapshot|cancel`, trader_name=hoang). `pgrep -f ds104-stopentry-watch` = alive [A].
- Live state as of ~08:25 CT: today's only `kind=stop_entry` row in the LIVE DB is **id 166** (ASIA S2 LONG, entry 29894.00, stop 29849.00, created 01:36:39 CT, state=cancelled, pre-seam-on) [A, query below].
- The dispatch's candidate (NY v1 S2 reclaim, entry 29766.75, stop 29796.00) had NOT materialized in `armed_orders` as of 08:25 CT. Watching.
- Timelines will be appended here as frames arrive, with ids and journal quotes (L7: read, never typed).

### A queries (re-runnable)

```sql
-- today's stop_entry rows (live DB, read-only)
SELECT id, kind, state, session, scenario, side, printf('%.2f',entry_px),
       printf('%.2f',stop_px), created_at
FROM armed_orders WHERE kind='stop_entry' AND date(created_at)='2026-09-18' ORDER BY id;
-- result @ ~08:25 CT: id=166|cancelled|ASIA|S2|LONG|29894.00|29849.00|2026-09-18 01:36:39
```

## B. REPLAY TABLE 1 — what the switch cost

**Method.** Every `kind=stop_entry` row since 2026-09-04 (n=30, ids 38..166).
For each row, in order:
1. **Gate simulation** — the real placement gate voids an arm whose entry is on
   the wrong side of the market (`limitMarketableWrongSide` strict boundary:
   LONG entry ≥ market, SHORT entry ≤ market; equality marketable) at placement.
   Market proxy = close of the newest 1m bar at/before `created_at` in the
   contract series the bot was receiving.
2. **Trigger search** — for resting-valid arms, first 1m bar with h≥entry
   (LONG) / l≤entry (SHORT) from max(created_at, session_start) until
   last-entry deadline (session end − 15 min).
3. **Outcome** — stop-first on same-bar, target, else flat-at-close; slippage 0 (stated).

**Windows (stated):** ASIA 17:00→02:00 · LONDON 02:00→08:30 · NY 08:30→14:45 CT,
last-entry 15 min before flat. **Contract rule (stated):** MNQ 09-26 through
09-14 15:33 CT, MNQ 12-26 after; for the 09-11 overnight the cache carried both
(see finding B2) and the gate's market proxy is the newest bar of either.

**B0 — classification of all 30 rows at creation [A]:**

| class | n | ids |
|---|---|---|
| wrong-side at creation → would be VOIDED at placement, never sent | 29 | 38,62,65,67,70,73,75,77,79,81,83,85,87,89,91,93,95,97,99,101,102 (NY S2 SHORT, entry 29591.02, market 29485–29536), 125 (LONDON S1 LONG, entry 29553.00 < 29563.75), 140 (ASIA S4 LONG, 29383.92 < 29401.25), 148 (NY S2 LONG, 29125.00 < 29154.75), 159 (ASIA S1 LONG, 29316.75 < 29354.50), 165 (NY S3 LONG, 29643.25 < 29671.75), 166 (ASIA S2 LONG, 29894.00 < 29902.50), 150, 153 (ASIA LONG, entries 29200.25/29156.25 vs newest cache bar 29519.50 — mixed-contract window, see B2) |
| resting-valid | 1 | 127 (NY S2 LONG, entry 29636.00, market at creation below) |

**B1 — outcomes of resting-valid rows [A]:**

| id | session/scenario | entry | stop | target | window bars | outcome |
|---|---|---|---|---|---|---|
| 127 | NY S2 LONG | 29636.00 | 29588.00 | 29736.75 | 130 (09-26) | **NO_TRIGGER** — max high 29632.00 < 29636.00 (missed by 4.00 pt) |

**Result: would-have fills = 0; would-have net = 0 pt.** The switch being off
did not forego any fill under the gates as coded: the authoring side kept
writing arms that were already through the trigger when written.

**B2 — the underlying defect [A]:** 21 of the 29 wrong-side rows are
re-authorizations of ONE spec — NY S2 SHORT entry 29591.02 across 21 plan
versions on 2026-09-04 (ids 38..102), all authored while the market traded
29485–29536, i.e. 55–106 points BELOW the sell-stop. Independent confirmation of
a mixed-contract feed in that era: at 09-11 00:28 CT `MNQ 09-26` closes
29138.25 and one minute later `MNQ 12-26` closes 29423.00 — an impossible
+285 pt one-minute "spread" (verified at identical wall-clock minutes on
09-08/09/10: only one series present per minute; the split series ran
09-07→09-11 with 12-26 offset ≈ +290 pt). This is the class-149 mixed-contract
corruption surface the bars-by-contract migration fixed today.

**Baseline (same period) [A]:** limit arms filled n=16 (`kind='limit' AND
state='filled' AND date(created_at)>='2026-09-04'`); `trade_excursions` rows
n=594 (table populated by the 1A backfill).

### B queries (re-runnable)

```sql
SELECT COUNT(*) FROM armed_orders WHERE kind='stop_entry' AND date(created_at)>='2026-09-04';          -- 30
SELECT COUNT(*) FROM armed_orders WHERE kind='limit' AND state='filled' AND date(created_at)>='2026-09-04'; -- 16
SELECT COUNT(*) FROM trade_excursions;                                                                      -- 594
SELECT contract, o, h, l, c FROM bars WHERE symbol='MNQ' AND tf='1m'
  AND open_time_ms IN (strftime('%s','2026-09-11 00:28:00 -05:00')*1000, strftime('%s','2026-09-11 00:29:00 -05:00')*1000);
-- 09-11 00:28 MNQ 09-26 c=29138.25 · 09-11 00:29 MNQ 12-26 c=29423.00
```

## C. REPLAY TABLE 2 — reject plays under the geometry gate

NOT YET RUN. Planned (due 11:00 CT):
- geometry records: 93 in backup + 2 newer in live DB = 95 total
  (`SELECT COUNT(*) FROM system_config WHERE key LIKE 'structural_geometry%'`) [A].
- join every `reject` scenario in plans since 09-13 to its verdict and DS-102's
  resolution rules (second column [B] until DS-102's branch exists).

## D. VERIFY OTHER LANES' CLAIMS

Nothing bridged yet; rolling.

## E. UNKNOWNS / NOT MEASURED

- Live-DB rows after 02:25 CT (beyond id 166) — none as of 08:25 CT.
- Whether the bot's cache at 09-11 01:01 held 12-26 or 09-26 as newest — [B]
  inference from bars table; the gate's price source is the cache's newest bars
  (armed placement reads current price from the futures bars provider) [B, code
  path `trader/armed_executor.go` placement + `market.FuturesBarsProvider`].
- NY v1 S2 candidate outcome — not yet authored.
