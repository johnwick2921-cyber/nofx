# W-PROOF-AND-REPLAY — DS-104 report (2026-09-18)

**Lane:** DS-104 · branch `docs/proof-and-replay-0918` · claim `12e096d044cc`
**Scope:** read-only research and verification. No code, no bot changes.
**Basis:** worktree `/home/hoang/nofx-104` off `origin/dev` b70fc6ca (2026-09-18 02:28 CT).
**Data:** backup copy `~/nofx-backups/pre-bars-key-20260918-022516.db` (consistent copy, 02:25 CT) for everything historical; live DB `data/data.db` read-only for rows after 02:25 CT only.
**Laws:** L1–L14 applied. Evidence tiers [A]/[B]/[C] on every claim.

## A. STOP-ENTRY LIVE PROOF — in progress

- Watcher: `/tmp/ds104-stopentry-watch.sh` (PID live, 60s poll, journal since 07:52 CT,
  cursor patched to seconds after an early repeat bug). `pgrep` = alive [A].
- **Boot line @ 07:52:20 CT (READ from journal) [A]:**
  `🎯 stop-entry: seam=on · slots=stop_price · guard=stop-side · unknown=no-op · addon build_id=2026-09-07-h1 expected=2026-09-07-h1 match=yes`
- **NY v1 plan written 08:06:20 CT** (planner call 359.8s, reasoning=max) [A].
  Arm specs: S1 reject entry 29799.64 stop 29840 · S2 reclaim entry 29766.75 stop 29796.00 · S3 reject entry 29889.52 stop 29920.25 (all `wait_confirm:true`).
- **S2 min-SL prediction CONFIRMED at authoring [A]:** journal 08:06:20
  `⚔️ arm feasibility: S2 arm stop 29796.00 too close (29.25 < 29.29 = 1.5×ATR5m) — min-SL gate will refuse it (WARN — write proceeds; the gate-at-arm chain enforces)`.
  So the candidate's stop is 0.04 pt inside the 1.5×ATR5m floor, exactly as the dispatch predicted. No `kind=stop_entry` row for S2 exists in armed_orders as of 08:35 CT (only id 166, ASIA, cancelled pre-seam-on).
- **No stop_entry placement has reached the wire yet** — no `signal_id`, no NT8 frame for any stop_entry row today. Watching.
- Observed on the same write [A]: `🪪 map 29647.50 id=NULL: missing formed_close_ms` and
  `🪪 map 29967.25 id=NULL: missing formed_close_ms` — NULL-id map entries in the NY v1 machine map (the DS-102 id-assignment surface, live today).
- Also observed [A]: `⏱ wake SKIPPED` ×4 (08:08–08:14, "22→16 min to flat (cutoff 25m) — seated Demand·1h invalidated") — no re-read between 08:06 and the last-entry gate.
- **NY v2 written 08:49:48 CT [A]:** `⚔️ arm feasibility: S1 arm stop 29795.25 too close (40.50 < 42.27 = 1.5×ATR5m)` and `S2 arm stop 29922.50 too close (40.88 < 42.27)` — the 1.5×ATR5m floor widened to 42.27 (ATR rose) and BOTH arms are again min-SL-refused at authoring. **Every authored arm today has been min-SL-refused at write; zero stop_entry rows exist to place (live check @ ~10:20 CT: only id 166).** The stop-entry seam is ON, the AddOn is proven (build_id match), and nothing reaches the wire because the gate-at-arm chain refuses each authored arm at write — the live proof of "the evaluator armed it, the executor never attempted it".

### A queries (re-runnable)

```sql
-- today's stop_entry rows (live DB, read-only)
SELECT id, kind, state, session, scenario, side, printf('%.2f',entry_px),
       printf('%.2f',stop_px), created_at
FROM armed_orders WHERE kind='stop_entry' AND date(created_at)='2026-09-18' ORDER BY id;
-- result @ ~08:35 CT: id=166|cancelled|ASIA|S2|LONG|29894.00|29849.00|2026-09-18 01:36:39
```

```bash
journalctl -u nofx --since '2026-09-18 07:52' --no-pager | grep -E 'stop-entry|arm feasibility|arm REFUSED'
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

**Method.** Live DB read-only (superset). Plans with `trade_date >= '2026-09-13'`
and `doc LIKE '%reject%'` → parse `doc.scenarios[]`, keep `condition=='reject'`,
dedupe by (plan_id, version, scenario id). Join verdicts from `system_config`
keys `structural_geometry:*` parsed as JSON (plan_id/version/scenario in the
value). Re-runnable queries below.

**C1 — totals [A]:** unique reject scenarios n=113 (ASIA 58 · LONDON 27 · NY 28);
geometry verdict records n=95 (93 in the 02:25 backup + 2 live-only); scenarios
WITHOUT a verdict record n=18.

**C2 — verdict distribution (n=95) [A]:**

| reason | n | detail breakdown |
|---|---|---|
| `no_provenance` | 65 | `entry_source_not_in_frozen_zones` 46 · `scenario_level_id_missing` 17 · `entry_zone_edges_or_provenance_unusable` 2 |
| `rr` | 28 | all R:R < 2.0 floor: min 0.568, max 1.468 (gain/risk read from the detail string) |
| `pending_gates` | 1 | `geometry_pass risk=10.5 gain=22.0 net=20.0` — passed geometry, refused later |
| `entry_gate` | 1 | entry-gate refusal |

All 95 records carry `quantity=0` (nothing was authorized).

**C3 — the no-verdict 18 (n, with named ids)** [A]: scenarios whose
(plan_id,version,S#) has no `structural_geometry` record — e.g.
2026-09-13 ASIA v7 S3, v14 S3/S4/S5, 2026-09-14 ASIA v1 S1, LONDON v1 S1/S3, v4
S3, NY v5 S1, … (full list in the archive file). Two plausible causes,
unmeasured which: arm authored before the gate wrote records, or key-shape
mismatch (leg). NOT MEASURED individually.

**C4 — entry-level kinds from trigger prose** (regex over the 113 triggers;
multiple tokens per scenario possible) [A]: ONH 30 · VWAP 30 · SWG 22 ·
VWAP+1σ 15 · ONL 14 · PDC 10 · PDH 10 · PDL 4 · nPOC 4. The reject anchors are
overwhelmingly the id-less kinds DS-102's fix targets (ONH/ONL/VWAP family).

**C5 — would the level resolve under DS-102's rules (second column)?**
**NOT MEASURED — [B] pending.** Their branch is not on dev yet; the gate's
current matching (pre-DS-102, `trader/structural_geometry.go:23-77`
`ResolveEntryGeometryZone`) fails closed on `scenario_level_id_missing`,
identity price/label/TF not matching a frozen zone source
(`entry_source_not_in_frozen_zones`), and unusable zone edges. Per the
dispatch I will re-run each of the 113 against their exported resolver from a
`_test.go` under `docs/` in this worktree the moment their PR lands (Part D),
and report every disagreement.

### C queries (re-runnable)

```sql
SELECT COUNT(*) FROM system_config WHERE key LIKE 'structural_geometry%';              -- 95
SELECT COUNT(*) FROM plans WHERE trade_date >= '2026-09-13' AND doc LIKE '%reject%';   -- 79 backup; live is superset
```

```bash
python3 /tmp/ds104_replay2.py   # produces /tmp/ds104-replay2.txt (per-scenario lines)
```

## D. VERIFY OTHER LANES' CLAIMS

**Item 1 — PR #174 (DS-103, feat/arm-state-ui).** Verified twice, in detached
worktrees (removed after): at the CTO-named HEAD `89e4de3c` and re-run at the
fixed head `4c923a76` after F1–F3 review fixes. **All PASS at 4c923a76** [A]:

1. `git diff --stat b70fc6ca..4c923a76` = 11 files — FE (`web/src/**`), guide
   content, i18n, AUDIT-CHECKLIST; zero Go/server files.
2. Executor column sources ONLY: `plan.armed` served by `api/handler_plan.go:448`
   via `armedMapFor` (impl `api/handler_plan_order_truth.go:63`, ledger rows,
   row_id) and `plan.structural_geometry` served by `api/handler_plan.go:505`
   via `planStructuralGeometry` (impl `api/handler_plan_geometry.go:6`, time_ms).
   `ExecutorVerdict.tsx` imports only `PlanArmView` + `StructuralGeometryView`.
3. `npx vitest run src/components/plan/ExecutorVerdict.test.tsx` — 9 passed (9), 517ms.
4. `npx tsc --noEmit` rc=0; `npx vitest run src/components/plan src/components/strategy src/guide` — 53 files, 307 tests passed.
5. Review pins present and green: `{reason:'entry_gate', quantity:0}` → refused
   (test `refuses entry_gate and one_setup via the store rule`); superseded+admitted
   → `superseded: …` (`a ledger row wins over an admitted geometry record`);
   `traderNames` undefined → no badge (`renders no trading claim when the binding
   is unloaded`; component gate `StrategyTradingBadge.tsx:17`).

Spec-freshness note: the branch moved past the first named HEAD while I verified
(89e4de3c → 4c923a76); both verified, verdicts reported per head.

**Item 2 — PR #175 (DS-102, fix/geometry-refusal, HEAD fa9bad01).** Verified in a
detached worktree (removed after). **Mixed verdict** [A]:

1. **FLAG** — `trader/auto_trader_planner.go` IS touched (9 lines) although
   DS-102's DONE message says "no auto_trader_planner.go write-time hint
   (DS-101's)". Changed: (a) `runPlannerReadWithTriggerClaimedCtx` ~:1287 —
   identityMap hoisted + `kernel.EnsureReferenceLevelIDs(identityMap)` gated on
   `GeometryRefIDsEnabled()`; (b) `assemblePlannerInputWithCtx` ~:2919 —
   `GeometryRefIDs` added to the planner input. Full diff = 18 files (listed in
   the CTO message).
2. PASS — all five named tests run and passed (`TestGeometryRefusalWarnOncePerKeyChange`,
   `TestEnsureReferenceLevelIDsStableAndScoped`, `TestGeometryTFWildcardExecutorOnly`,
   `TestGeometryReferenceLevelsKnobResolution`, `TestGeometryRefBootLineReadsResolvedKnob`),
   `ok nofx/trader 1.137s`.
3. PASS — `go build ./...` rc=0; `go vet ./...` rc=0.
4. PASS with named gap — my independent replay via their exported
   `trader.ArmGeometryVerdict` (scratch test under docs/, removed with the
   worktree): backup copy n=109 reject scenarios; live DB adds today's
   LONDON v2 S1/S2, NY v1 S1/S3, NY v2 S2 (live n=114 at query; their 113 = an
   earlier snapshot). OFF split ARMED 35 / source_not_frozen 52 + zone_edges 3 =
   55 / missing 19 (their 20). ON split ARMED 33 / zone_edges_unusable 55 /
   missing 19 / ambiguous 2 — consistent with their 33/58/20/2 modulo live rows
   and the missing bucket ±1.
5. PASS — both named flips reproduced exactly: 09-13 ASIA v10 S2 long
   OFF=ARMED → ON=`entry_zone_ambiguous`; 09-17 ASIA v2 S1 short same. No OTHER
   ARMED→X flips. Classification-change count on the 109 backup rows = 54 (their
   57 presumably includes the live rows).

**Item 3 — PR #176 (DS-101, fix/write-time-feasibility, HEAD 3e3c97d1, base 088020cf).**
Verified in a detached worktree (removed after). **Mixed verdict** [A]:

1. **FLAG** — `web/src/guide/types.ts` GUIDE_BUILT_REV changed IN the PR:
   `249ff3a5eea6…` → `26087401ffbc…` (their branch build) — the CTO's rule says it
   must not. Diff = 21 files (Go: kernel/plan_doc, planner_prompt,
   prompt_contract + parity test; trader/write_time_feasibility(+test),
   auto_trader_dayplan, auto_trader_planner; store/strategy, knob_registry_table;
   scripts/replay_write_time_feasibility.py; web guide ×3; AUDIT-CHECKLIST; report).
2. PASS — five named tests run and passed (tails `ok nofx/trader 4.062s`,
   `ok nofx/kernel 0.027s`).
3. PASS — `go build` rc=0, `go vet` rc=0, `npx tsc` rc=0, **FULL** `npx vitest run`
   72 files / 468 tests passed.
4. **FAIL with root cause proven** — their replay script
   (`scripts/replay_write_time_feasibility.py:143`) reads the LIVE bars table
   with NO contract filter AFTER the class-149 migration (live PK =
   `(symbol,tf,contract,open_time_ms)`; both `MNQ 09-26` AND `MNQ 12-26` rows
   exist at the same minutes in the 09-07..09-14 overlap — 25+25 on 09-13).
   Their report's `atr5m=280.51` for 09-13 ASIA v1 is garbage; the same script
   on the single-contract backup gives `32.80`. Their min-SL bucket (39) rides
   the inflated ATR; my backup run of THEIR script gives min-SL=25. Their
   "96 plans" = my 93 backup + 3 live-only plans (consistent). Python replay,
   not the Go call site (their own flag stands).
5. **False-negative risk = 0 (measured)** — of their 67 geometry-bucket rows
   (backup run) plus the 5 live-only rows, **zero** are ADMITTED by DS-102's
   knob-ON `ArmGeometryVerdict` (my per-row [A] map: 109 backup + 5 live rows;
   the 33 knob-ON admits are all rows legacy also admitted — none sit in their
   geometry bucket). #176 staying on the legacy geometry function cannot admit
   rows the fix would refuse.

**Item 3b — PR #176 FIXED head 2f4168ee (CTO-REJECT rework).** Re-verified in a
detached worktree (removed after). **All PASS + one merge finding** [A]:

(a) PASS — full `go test ./... -count=1`: 35 ok, 0 FAIL (clean rerun; the first
    run's 5 FAILs were MY contamination — I edited `store/strategy.go` mid-run
    during the merged-head test; named for honesty). `gofmt -l` empty.
(b) PASS — four named tests ran and passed (`ok nofx/trader 0.966s`; kernel
    parity test green).
(c) PASS — GUIDE_BUILT_REV == `249ff3a5…` (reverted in the PR).
(d) PASS with named delta — reworked script HAS the contract filter
    (`contract_at` tie-break; history from the pre-149 copy). Backup-only run:
    min-SL 25 / geometry 67 / stop_side 4 / none 70 / no_arm 33 (93 plans) vs
    their re-issued 27/72/4/70/33 — the delta is exactly the live-only rows
    (93 vs 96 plans).
(e) PASS + finding — dry-merge 2f4168ee+146baf1d CONFLICTS in 4 files
    (store/strategy.go, guide settings, GuidePage.test, AUDIT-CHECKLIST).
    Union-resolved throwaway merged head: build rc=0, vet rc=0. **Admit answer:
    YES** — the write-time geometry leg calls
    `ArmGeometryVerdict(doc, sc, GeometryRefIDsEnabled())`
    (`trader/write_time_feasibility.go:31-34`) and ADMITS a ref| ONH reject play
    with the knob ON at the merged head (throwaway `ds104_merged_admit_test.go`:
    ON→ADMITTED, OFF→`identity_not_valid_in_frozen_map`).

**Item 2b — PR #175 FIXED head 146baf1d (F1–F6).** Re-verified in a detached
worktree (removed after). **All PASS + measured drift** [A]:

1. PASS — `go test ./... -count=1` FULL: 35 ok, 0 FAIL;
   `TestEveryClaimedProductionPathHasACallSite` passes explicitly
   (`wiring_gate_test.go:284`).
2. PASS — `gofmt -l` over changed .go files: empty.
3. PASS — five named tests ran and passed (tails `ok nofx/trader 2.876s`,
   `ok nofx/kernel 0.026s`).
4. MEASURED — Table 2 [A] re-run at 146baf1d (109 backup rows):
   **ARMED 58 / entry_zone_edges 20 / missing 19 / ambiguous 12** vs 33/55/19/2
   at fa9bad01. 37 rows changed, every one named in the CTO message + report
   archive: 25 edges_unusable→ADMIT (zero-width admission), 10→ambiguous (F4
   wildcard), 2 ADMIT→ambiguous (09-13 ASIA v9 S1, v14 S1 — regression direction
   for those two: previously admitted, now refused). The two earlier-named flips
   stay ambiguous.
5. PASS — adversarial (throwaway `docs/ds104_adversarial_test.go`): a
   reference-KIND zone with explicit lo==hi non-nil + VALID strict id → REFUSED
   `entry_zone_edges_or_provenance_unusable`; a nil-width line with a
   non-reference-kind source → REFUSED same. Both refuse via the zero-width
   guard itself (ids computed with `levelidentity.ID`, so `LevelByID` succeeds
   and the guard is what fires).

**Item 2c — PR #175 final head 6adedf07 (DS-102 F8: local-band admission,
local zones copy).** Verified in detached worktree (removed after). **All PASS**
[A], sent 17:29:44Z:

1. PASS — `go test ./... -count=1` FULL = 35 ok / 0 FAIL; `gofmt -l` empty over
   the changed .go files. (My throwaway six-order test lives under docs/ and is
   green; removed with the worktree.)
2. Named tails: `TestGeometryRefAdmissionDoesNotMutateSharedDoc` PASS,
   `TestGeometryRefWildcardDoesNotFlagMergedNames` PASS,
   `TestGeometryRefIDsOnhRejectPlayAdmitted` PASS,
   `TestOneSetupRefIDRejectPlayArms` **SKIP — inside the lunch no-trade window
   12:00–13:30 CT (geometry_refusal_wave_test.go:416)**; reported as SKIP, not
   PASS.
3. PASS — six-order adversarial (throwaway `docs/ds104_sixorder_test.go`): doc
   with three reject scenarios (ONH ref| long, Supply short, Demand long)
   composed in all 6 orders → every scenario's (stop,target) identical across
   orders and the marshalled `doc.Zones` byte-identical after each.
   (onh 95.50/109.50, sup 140.00/0, dem 60.00/0 — the 0 targets are my synthetic
   map having no complete zone on that side, not a mutation.)
4. PASS — Replay Table 2 verdict maps at 6adedf07 vs 99768755 on the BACKUP:
   **IDENTICAL, 0 rows moved** — 64 ADMIT / 24
   `entry_zone_edges_or_provenance_unusable` / 19 `scenario_level_id_missing` /
   2 ambiguous at BOTH heads. DS-102's 68/24/20/2 is the live DB (114 rows); the
   backup is the 02:25 CT snapshot (109 rows), so my absolute counts are −4
   ADMIT / −1 missing — population, not a head delta. No regressions.

**Item 3c — PR #176 round-2 head 9018ca85 (DS-101 GeometryRefIDs in the
write-site context).** Verified in detached worktree (removed after). **All PASS
with one mandatory merge resolution** [A], sent 17:29:51Z:

1. PASS — `go test ./... -count=1` FULL = 35 ok / 0 FAIL; `gofmt -l` empty over
   the changed .go files.
2. PASS — named tails: `TestWriteTimeFeasibilityNeverMutatesZoneMap`,
   `TestWriteTimeFeasibilityHintRedToGreen`,
   `TestWriteTimeFeasibilityLastAttemptDisablesArm`,
   `TestGeometryRefIDsKnobResolution` all PASS.
3. **MERGED HEAD FAILS TO COMPILE AS MERGED** — dry-merge 9018ca85 +
   fix/geometry-refusal@6adedf07: `store/strategy.go` unions to ONE
   `GeometryReferenceLevels` field + ONE `GeometryRefIDsEnabled()` resolver
   (DS-101 named their copy identically on purpose — no duplicate there), BUT
   `store/knob_registry_table.go:187` git-AUTO-MERGES with NO conflict marker
   (both branches appended a `"geometry_reference_levels"` registry row at
   different lines) and `go build` FAILS: `duplicate key
   "geometry_reference_levels" in map literal`. DS-101's carried row claims
   "identical to DS-102's row" — **FALSE**, the Consumers lists differ. This is
   the silent-compile-break hazard class the CTO asked about, in a file outside
   the 5-file conflict set. Resolution: delete the carried row, keep DS-102's
   canonical row → `go build` rc=0, `go vet` rc=0.
   Throwaway at the resolved merged tree: ONH ref| fixture knob ON → write site
   issues EMPTY + executor ADMITS idx=0; knob OFF → write site
   `geometry_no_provenance (identity_not_valid_in_frozen_map)` + executor
   refusal (the SAME knob threads both); sweep_reclaim leg with raw stop inside
   the 1.5×ATR5m floor → write-site geometry issues EMPTY (only an unrelated
   wait_confirm class). One nuance recorded: the authored arm must still carry
   exact stop/target (`PlanArmSpec.Validate`), geometry composes at gate time —
   fixture authored 29892/29950.
4. PASS-with-note — their re-issued replay numbers are TRUE on the LIVE DB
   (read-only): R:R 3 / geometry 72 / stop_side 6 / none 95 / no_arm 33, exact.
   On the BACKUP copy I get R:R 3 / geometry 67 / stop_side 6 / none 93 /
   no_arm 33 — the −5 geometry / −2 none is population: 7 plans written after
   the 02:25 CT snapshot. Recommendation for the real merge: delete the carried
   registry row; nothing else blocks.

## E. UNKNOWNS / NOT MEASURED

- Whether the bot's cache at 09-11 01:01 held 12-26 or 09-26 as newest — [B]
  inference from bars table; the gate's price source is the cache's newest bars
  (armed placement reads current price from the futures bars provider) [B, code
  path `trader/armed_executor.go` placement + `market.FuturesBarsProvider`).
- The 4 live-only reject rows at 08:55 CT were later 5; today's rows keep growing
  during RTH — the C tables name the count at query time.
- DS-102's replay "57 change" vs my 54 flips on the 109 backup rows — the delta
  presumably their live-row inclusion; NOT re-derived from their artifact.

## F. PROPOSED CLASS (per L6 — text proposed, checklist edit left to the coder lane)

**CLASS NN (proposed): "the evaluator says triggered, the executor never
attempts — and nothing on the card said so."** Born 2026-09-04 (first stop_entry
arm written) and found 2026-09-18 by the owner ("why no trade"). Probe: a
scenario that the evaluator/trigger surface renders ≈triggered or ≈armed while
EVERY executor-side gate at the arm seam refused it — geometry provenance
(65/95 verdict records), min-SL at write (every arm authored today, 29.25/40.50/40.88
vs the 1.5×ATR5m floor), or wrong-side placement (29 of 30 stop_entry rows since
09-04) — and the UI showed the evaluator's state, not the executor's, so the
owner read "armed" while zero orders existed. The breadth: three distinct
refusal sites, none surfaced on the card before DS-103's ExecutorVerdict column.

## G. VERDICT — DONE_WITH_CONCERNS

Report: `docs/superpowers/reports/2026-09-18-proof-and-replay.md` on branch
`docs/proof-and-replay-0918` (report sha at the time of this line:
`a5f67d35`; pushes follow every update). Concerns: (1) both coder lanes shipped
review items that contradict their own DONE messages (DS-102's
`auto_trader_planner.go` edit; DS-101's GUIDE_BUILT_REV bump + contract-unfiltered
replay ATR) — both flagged, neither silent; (2) Part A is still open by nature
(zero placements so far; watching until 14:45 CT).
