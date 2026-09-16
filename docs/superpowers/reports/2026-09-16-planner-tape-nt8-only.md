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
