# Planner tape is NT8-only — dispatch 101 follow-up (2026-09-16)

**Ruling:** CTO nofx-c5 under the owner's delegation ("you are cto just approve as my guideline"):
`historical_import` rows are excluded from EVERY planner door; the chart keeps them, labelled. Follow-up
to dispatch 101 (`9e200002`, booted 14:33 CT), where guard (iii) closed the rehydrate door and A15 §8 named
the second door with counts read from the store.

## What was true before this wave [A]

`data/data.db` read-only, 2026-09-16 12:2x CT: `MNQ 12-26` 1m = 2,873 live + 25 replay + **426
`historical_import`** = 3,324 rows; `plannerCandleTapeBars = weeklyFactsTapeBars = 12,000`; the shared
reader `LastNBarsOn` filters only mixed + off-scale. So all 426 imported bars were inside the planner's
and the weekly reader's 1m tape. On any `09-26` resolve: 75,492 import rows.

## The doors (A29 — every production call site named)

| door | file:line | before | after |
|---|---|---|---|
| planner + weekly 1m tape | `trader/bars_store_depth.go` `storeBarReader` | `LastNBarsOn` | `LastNBarsFromNT8On` |
| exported planner seam | `trader/bars_store_depth.go` `BarsWithStoreDepth` | `LastNBarsOn` | `LastNBarsFromNT8On` |
| weekly bars from epoch | `trader/auto_trader_weekly.go:84` | `BarsBetweenOn` | `BarsBetweenFromNT8On` |
| weekly `Own1m` window | `trader/auto_trader_weekly.go:116` | `BarsBetweenOn` | `BarsBetweenFromNT8On` |
| POC-touch historical leg (level retirement) | `trader/auto_trader_dayplan.go:227` | `BarsBetweenOn` | `BarsBetweenFromNT8On` — **found by the lint, not by the ruling's list** |

Kept on the shared readers (imports included), by design: `BarsWithStoreDepthDisplay` (:108/:123, the
chart [O]); `bar_persist_wire.go:456` (the rehydrate read — its door refuses and COUNTS imports).

**LEFT by CTO ruling (they MEASURE, they do not decide):** `level_stats_wire.go:121` (nightly level
stats), `trade_excursion_hook.go:168` and `trade_excursion_backfill.go:120` (MAE/MFE over a trade's
window), `follow_plan_wiring.go:217` (recorded-only follow plan, [T]). A measurement over imported
minutes is a separate question from a decision over them.

**`one_setup_boot.go:174` — read the code, LEFT:** the call sits inside `BackfillOneSetupVerdicts`
(header :87–95: "D9 for the verdicts, three-state. For every episode opened at or after sinceMs with no
verdict: the LEVEL leg is judged against the last candidate-pool read before the episode's open … price
from the store's bars through the contract filter"), reading the six minutes around a PAST episode's open
(`r.OpenedAtMs-5*60_000, r.OpenedAtMs+60_000`) and writing
`store.OneSetupStamp{… Backfill: "recomputed"}` via `ts.StampOneSetup(r.ID, stamp)` at :199–200. It stamps
a verdict onto an episode that already happened; nothing at boot arms, retires or authorizes from it. The
live one-setup verdict at an open is `one_setup_wiring.go` and reads the ring, not the store.

`LastNBarsOn` / `BarsBetweenOn` are byte-identical. The new readers are theirs plus
`AND COALESCE(source,'') <> 'historical_import'`; `ImportRowsOn` is the census the boot line prints.

## Pins — all RED first (quotes)

- store `TestNT8OnlyReaderExcludesImportsTheSharedReaderReturns`: 500 live + 5 import rows NEWER than the
  live ones (a reader that merely trims the oldest could not pass) → shared 505 (premise) / NT8-only 500 /
  census 5. RED: `NT8-only reader: want 500 rows, got 505`.
- store `TestNT8OnlyRangeReaderExcludesImports`: RED `want 100, got 103`.
- trader `TestPlannerStoreReaderServesNoImportRows` — the REAL door (`&AutoTrader{store}` → `storeBarReader`,
  `currentContract` from the store's latest stamp). RED: `planner reader served 505 rows, want 500`.
  A8 mutation (shared reader put back in `storeBarReader`, build rc=0): RED again; restored, green.
- trader `TestNoPlannerDoorReadsTheSharedBarReader` — text tripwire (class-113 caveat stated) on the four
  planner files + the two door BODIES in `bars_store_depth.go`. Its first form cut the file at the display
  seam and missed `storeBarReader` (which sits after it) — the mutation showed it did not bite; rewritten by
  function body, mutation-proved: `storeBarReader( reads the shared reader`. Its first run is what found the
  POC-touch door.
- trader `TestPlannerTapeBootLineNamesTheMovedValues` (class 82: import count, tape length and baseline both
  ways, 8 vs 9 session-days measured; UNKNOWN never 0) and `TestPlannerTapeAccountingIsZeroWhenTheInputDoesNotChange`
  (E7-style: same 9-session tape both ways → `Δ+0` rows, `Δ+0.000000` baseline — the rule did not move).
- `TestEveryClaimedProductionPathHasACallSite`: `LastNBarsFromNT8On`, `BarsBetweenFromNT8On`, `ImportRowsOn`,
  `PlannerTapeBootLine`, `logPlannerTapeAccounting` registered.

Suite: store, trader, trader/ninjatrader, api ok. Guide vitest 23/23, tsc clean.

## The boot line (A11, every field read)

`🧮 planner tape [NT8-only, CTO ruling 2026-09-16] @<t>: MNQ 1m contract=<c> · import rows on contract=<n>
(excluded from every planner door) · tape NT8-only=<a> rows vs with imports=<b> rows (Δ<a−b>) · regime
baseline NT8-only=<x> vs with imports=<y> (Δ<x−y>) · chart keeps imports, labelled` — emitted beside the
📈 regime line in the dayplan boot hook; the with-imports half is measured through the shared reader in
`planner_tape_nt8_only.go` ONLY (a census, never a planner input — the lint names that file as the exception).

## A15 — what the owner will see / what moved

- Expected at the boot on `12-26`: `import rows on contract=426`, tape Δ **−426** rows — UNLESS the 426 are
  older than the newest 12,000 NT8 rows by then (they were inside on 09-16). The regime Δ is the number
  that says whether the exclusion mattered; it is printed, not predicted.
- Levels: the POC-touch leg no longer sees imported minutes, so a POC an import "touched" may un-retire
  at the next planner read. Named, not measured — no fixture can measure today's levels.
- The chart is unchanged.
- CLASS 129 appended to the checklist (the door shape, the by-body lint, the newer-than-feed fixture, the census line).

## A12 · Docs in this commit
RULEBOOK §A (the planner's tape is NT8's own), SYSTEM-MAP §1 (doors + the kept readers), Guide `status.ts`
(paragraph + the 🧮 boot line). `GUIDE_BUILT_REV` is bumped at the cutover to the merged rev (A27).

## Rollback
Go binary swap only; no migration, no flag. The readers are additive; the previous binary reads the shared readers.

## F3 · THE BOOT — c6579347, 2026-09-16 15:31:59 CT (owner-run runbook; F3 PARTIAL)

PID 2748996. `🔐 BOOT INTEGRITY OK — rev c6579347580a · built 2026-09-16T19:57:23Z · expected c6579347580a
· goldens PASS`; `/api/health` c6579347; `deploy/RELEASE=c6579347`; `nofx-bin.old.9e200002` HOLDS 9e200002
(`go version -m`). `🧯 nt8 history at subscribe: MNQ 1m..30m 2000/2000, HTF n/a`; `🧯 ring rehydrated MNQ 1m
[O]: nt8=2001 store_live=2500 store_hist=0 import=0 (refused at the door — guard iii) total=2500/2500`;
`🧯 ring rehydrate done: 9 of 17 pairs deepened, +1104`; `📼 … mismatches this process: none`; zero
`scale check SKIPPED` / `DIFFERENT PRICE SCALES` lines; 4h horizon 406/500 at 14:5x on the prior boot,
HTF `n/a` at this subscribe (lands late, as before).

**F3 PARTIAL — the `🧮 planner tape` and `📈 regime input window` lines did NOT print this boot.** Read
from timing [A]: both are emitted by the trader's `afterBackfillHook` (`auto_trader_dayplan.go:87` →
`SetAfterBackfillHook`). This boot the replay landed fast — `📦 bars: persisting` 15:32:02, the hook check
and the `🕳 bar horizon:` boot line 15:32:03 — and the trader installed the hook at 15:32:05 (its `📊 bars`
line): `afterBackfillHook.Load()` was nil, the block skipped it silently, and the R1 "📊 bars after
backfill" line is missing with them. At 14:33 the backfill waited until 14:33:11 and the trader had loaded
at 14:33:11 — the hook won by chance. **Class candidate: a hook that loses when its event beats its
installer** (silent by construction — absent, not wrong; A24). The exclusion itself is live (the planner's
reads go through `LastNBarsFromNT8On`; static pins + lint); the LIVE accounting is owed. CTO ruling (b):
no third boot for a log line — the fix (hook fires immediately if the backfill already landed; 🖥 compares
bundle rev to binary rev, not timestamps) ships as the next PR and the `🧮` proof is taken at the next
scheduled boot.

**Correction (CTO, [A]):** my first read said "dist f53f4e94 installed by the owner". Wrong. `web/dist`
serves `assets/index-C0CWYHKw.js` — the **bcd70c0d** dist (bundle `GUIDE_BUILT_REV=9e200002`, mtime
14:40:15), installed after the 14:33 boot (which served the 09-13 bundle). The **f53f4e94** dist
(`index-GYuDKImr.js`, rev c6579347, parked at `scratchpad/cc101d/web-dist-f53f4e94/`) is NOT installed
and is owed — the owner ran the binary runbook only. So the `🖥 ui … STALE` line this boot is REAL (bundle
rev 9e200002 ≠ binary c6579347) and the Guide banner shows drift until that cp runs; the timestamp-vs-rev
point is a separate small fix, not this boot's explanation.

Marker: this commit, from `~/nofx`, `deploy/RELEASE=c6579347` (written by the runbook before the kill, A19).
