# W-STRUCTURE-ZONE-SEATS — structure-map zones as seat candidates (knob OFF)

**Branch** `feat/structure-zone-seats` · **base** `origin/dev` @ `de869d38` · owner order 2026-09-17 ("no do it now"), given after the CTO stated Round 24 found zone-at-entry is not a filter. Ships behind `day_plan.structure_zone_seats`, default OFF, seating byte-identical when off.

**Owner decision — not gated by the knob.** Fix 1 (one-setup seeds the plan's own seated levels) is NOT gated by `structure_zone_seats`: it changes live arm decisions for every trader at boot (plan levels that were never in one-setup's map become armable); the `one_setup:*` counters record allowed vs declined so the delta is visible after the boot.

## What was true before (evidence tier [A], line refs at base `de869d38`)

- The S1 structure map (`kernel/structure_map.go:39-48` `StructureZone{Kind,Lo,Hi,TF,Fresh,Score}`; `:186-199` `structureZonesFor`) draws its zones FROM the read's uncapped in-band HTF-zone universe (`trader/auto_trader_planner.go:2531,2555` `htfZonesFull`, `:2801` the map call) and renders them as context only (`structure_map.go:208-209`: "A structure zone is context — never an entry").
- Every HTF zone already enters the seat race as its DETECTOR row — `DetectHTFLevels` output rides `extra` into `AssembleResearchLevels` (`auto_trader_planner.go:2518-2520`; `kernel/levels_assemble.go:247`) — but that row is priced at the zone **midpoint** (`kernel/levels.go:74`, `:123-128` `zoneLevel`). A 40-pt 4h demand zone whose top edge sits 8 pt below price is a row 28 pt below price. Its edge — the thing the owner wants tradable — is not a reference anywhere.
- The seat pipeline (`kernel/levels_score.go`): band `:465`; zone grader `:540-590` (evidence × size × freshness × confluence × TF, then the TF floor/cap and the Tier-1 proximity cap B2 `:587-589`); cluster collapse `:779-842` — **zones are exempt** (`:803-807`); priority sort `:616-629` (today-priority kinds first, then score); `seatHTF` `:1048`; the 2×cap scorer + **max-nearest final cut** `:695-702` (`ScoreLevelsMinGradeFullSeats`: the pool is the top-2×max by priority/score, the table is the max NEAREST of those).
- The validator's structural-label rule `kernel/plan_doc.go:868-880` (`structuralLabels`) and `:907-929` (`MislabeledStructuralLevels`, called at `trader/auto_trader_planner.go:1965`) flags a plan label only when EITHER side is a structural anchor and they disagree. A `ZONE-*` label is not in the set on either side, so it is a free label.
- One-setup (`kernel/one_setup.go:89-153`, `:158-170` same-level by id or price within the cluster width, `:176-206` best = grade first, distance second) reads the merged map (`kernel/map_candidates.go:132-197` `BuildMapCandidates`, names carry `CollapsedNames`). Its production seam `trader/one_setup_wiring.go:131` assembled its map with `AssembleScoredLevels(...)` and NO HTF extras — no HTF seat (zone or otherwise) was in one-setup's live candidate map. FIXED in this PR's review round (fix 1 below).
- The card's level rows print `doc.Levels[i].Label` (`api/handler_plan.go:553-585`) and enrich by price from `planW3Map` (`:2454-2461`, again no HTF extras). A ZONE-* seat the model copies renders like any level — no new UI.
- Knob plumbing pattern: CLASS 141 `0d54518c` (`store/strategy.go` field + `FlipRereadEnabled`, `store/knob_registry_table.go` row, `trader/flip_reread_boot.go` line printed at `manager/trader_manager.go:597`, `DayPlanEditor.tsx` toggle + test, `plan-translations.ts`, `guide/content/settings.ts` 10-field entry, `GuidePage.test.tsx` count 52).

## The seating rule (five lines, as implemented)

1. For each structure-map zone (D, 4h, 1h — the map's own top-6 per TF), the candidate price is the zone **edge nearest to price** (inside → still the nearer edge; tie → the low). `kernel/zone_seats.go zoneNearestEdge`.
2. An edge farther than the proximity band (`proximity_filter_atr × dATR`) from price is dropped — nothing changes for that zone (`ZoneSeatCandidates`, recorded as `out_of_band`).
3. The candidate enters the pool BEFORE scoring as a `DetectedLevel` cloning the zone's own detector row (kind, TF, Lo/Hi, HTF, `ZonePattern`, origin) with `Price = edge`, `Label = ZONE-<TF>-<KIND>` — graded by the SAME zone grader as every zone on that TF (`AssembleResearchLevelsZoneSeats`, `levels_assemble.go`).
4. In the cluster collapse (12 ticks = 3.00 pt) a `ZONE-*` row within the tolerance of ANY other survivor merges INTO it — never the other way round (it sorts after every non-ZONE row as a keeper, `levels_score.go collapseLevelClusters`): the detector level keeps its seat and label, the ZONE name always rides on `CollapsedNames` (merged map, card names). Otherwise the ZONE row stands and competes under the SAME priority rule, the SAME max-nearest cut and the SAME cap — no reserved seat, no multiplier.
5. `ZONE-*` is a free label for the structural-label rule; a scenario authored on a ZONE seat passes `ValidatePlanDocWithCaps`, `MislabeledStructuralLevels` and `ValidatePlanDocWithFactsMachine` (pinned).

Why rule 4 lives in the collapse rather than a pre-scoring alias: an alias decided against the raw pool is lost when the host row is later deduped or cut (seen in the first version of the coincident test); the collapse runs on survivors, so the alias can never vanish.

## What changed

| file | change |
|---|---|
| `store/strategy.go` | `DayPlanConfig.StructureZoneSeats bool` (`structure_zone_seats`) + `StructureZoneSeatsEnabled()` |
| `store/knob_registry_table.go` | `structure_zone_seats` row, KnobLive |
| `kernel/zone_seats.go` (new) | `ZoneSeatLabel`, `IsZoneSeatLabel`, `ZoneSeatCandidates` (rules 1–3), `ZoneSeatOutcome` (recorded counters) |
| `kernel/levels_assemble.go` | `AssembleResearchLevels` → shared `assembleResearchLevels` with a pool hook; `AssembleResearchLevelsZoneSeats`; `PlannerPriceAndRange` (the same price/dATR derivation, exported for the pre-pass) |
| `kernel/levels_score.go` | `collapseLevelClusters`: ZONE-* rows are not zone-exempt, never a keeper over a non-ZONE row, name always carried — inert with no ZONE rows |
| `trader/zone_seats_wire.go` (new) | `zoneSeatsEnabled`, `ZoneSeatsBootLine`, `zoneSeatCandidatesForRead` (the ONE production pass), `zoneSeatsReadLine` |
| `trader/structure_map_wire.go` | `structureMapCompute` (ungated) under the gated `structureMapForRead` |
| `trader/auto_trader_planner.go` | knob OFF → the untouched `AssembleResearchLevels` call; ON → the pass + `AssembleResearchLevelsZoneSeats` + read line; the S1 section reuses the pass's map when its own knob is on |
| `manager/trader_manager.go` | boot line beside the S3 htf line |
| web | `types/strategy.ts`, `DayPlanEditor.tsx` toggle (under Structure map) + test, `plan-translations.ts` en/zh/id, `guide/content/settings.ts` entry, `GuidePage.test.tsx` 52 → 53 |

Boot line: `🗺 zone-seats=off(default) (W-STRUCTURE-ZONE-SEATS)` / `🗺 zone-seats=on(saved) (W-STRUCTURE-ZONE-SEATS)`.
Read line (ON only): `🗺 zone-seats @NY: zones=3 in_band=2 out_of_band=1 duplicates=0 no_source=0 → merged injected=1 aliased=1 [ZONE-4H-DEMAND@19980.00 ZONE-1H-SUPPLY@20040.00]` — the first counters are the candidate pass, the `merged` pair is what the scorer's output records.

The zone-seat pass computes the structure map whether or not `day_plan.structure_map` is on (the seating must not depend on whether a prompt section renders); the section, its log and the doc stamp stay gated by their own knob.

## Proofs (`kernel/zone_seats_test.go`, `trader/zone_seats_wire_test.go`)

- **OFF byte-identical** — `TestZoneSeatsOffIsByteIdentical`: on the identity-golden tape (`identity_output_parity_test.go`, 12 seats filled), `AssembleResearchLevels` (origin/dev's function, signature untouched) and `AssembleResearchLevelsZoneSeats(nil)` give `reflect.DeepEqual` seated/pool/raw and an identical `RenderKeyLevelsBlock`. The identity golden itself (`testdata/identity_legacy_output.json`) still passes, so the refactor of the shared body moved nothing.
- **ON, in band** — `TestZoneSeatsOnSeatsInBandZoneAtEdgeAndValidates`: a 40-pt 4h demand zone whose top edge is the nearest 3-pt-clear spot below price seats as `ZONE-4H-DEMAND` at the edge, with the SAME grade as its detector row; a plan with a `reject` scenario on it passes all three validators; it is a merged-map candidate with its grade and one-setup compares it as a resolved level (on that tape OR-L, grade A at 0.6 pt, legitimately outranks it).
- **One-setup picks it when best** — `TestZoneSeatOneSetupPicksZoneSeatWhenBest`: production functions (`ScoreLevelsMinGradeFullSeats` → `BuildMapCandidates` → `OneSetupAllowsAt`) with the anchors outside the band → `Allowed`, `BestNames` = the ZONE seat.
- **Coincident zone not double-seated** — `TestZoneSeatsCoincidentZoneIsNotDoubleSeated`: edge 1 pt from a seated detector row → `injected=0 aliased=1`, exactly one seat near the edge, the detector's label, ZONE name in its `CollapsedNames` and in the merged map's names.
- **Out of band** — `TestZoneSeatsOutOfBandChangesNothing`: no candidate, seats `DeepEqual` to OFF.
- **Collapse rule** — `TestZoneSeatCollapseKeepsDetectorRowAndFoldsZoneEdges`: a ZONE row scoring above a VWAP within 3 pt merges INTO the VWAP; two ZONE edges fold; a plain `Supply·1h` stays exempt; the no-ZONE path is unchanged.
- **The production call site** — `TestZoneSeatsAtThePlannerReadCallSite` (`trader/zone_seats_wire_test.go`): `assemblePlannerInput` on a stub tape with the log captured. OFF: no zone-seats line, no ZONE row. ON: exactly one read line per read; `Structure` stays nil (the section is gated by its own knob); every ZONE seat inside the band; the line's `merged injected` equals the ZONE rows the read's pool carries; with nothing injected the ON table equals the OFF table. ON + structure_map: the section carries the pass's map. On that periodic tape the map lists one 4h OB band six times — the pass records `duplicates=10` and emits one candidate per distinct band.
- Label rule, nearest-edge cases, no-source and duplicate counting, boot/read lines, nil-safe resolver.

## The Round 24 finding this contradicts

Round 24 HTF-zone-entry (`docs/round-24-htf-zone-entry`, PR #151, 2026-09-17): the density headline — nothing about the zone at entry passes; the one apparent 0.786 (S4d) was a label flip (approach vs hold-trade). Zone-at-entry does not predict a better entry. This wave does not dispute that; it gives the owner the live test he asked for, behind a knob that defaults OFF and whose guide entry says so in plain words.

## The measurement that validates or kills it — Round 25 (seat displacement)

With the knob ON, every read records (a) the candidate pass and the merge outcome (the read line), (b) which seat a ZONE row DISPLACED (the pool is unchanged apart from the injected rows, so the displaced row is the 12th/13th boundary of the OFF table on the same tape — reconstructable from `detector_record` and the seated table), (c) scenarios authored on ZONE seats, their outcomes (`trade_excursions`, `pnl_corrected`) and their adherence. The kill condition: over N ≥ 30 ZONE-authored scenarios, the ZONE-seat trades do not beat the displaced-reference baseline on MAE/MFE-adjusted outcome, OR ZONE seats displace Tier-1/volume references at a rate that lowers the plan's target-reach (Round 24 Q4's 57% "table runs past the farthest seated level" is the baseline). Either → the knob stays OFF and the pass is removed.

A caution the measurement must respect: the max-nearest cut means a zone edge only ever seats when it is among the 12 nearest pooled rows — a far zone changes nothing however strong, so "the knob is on" is not "zones are seated"; count `merged injected` per read before attributing anything.

## Review fixes (PR #159 MERGE-WITH-FIXES, 2026-09-17)

**Fix 1 — one-setup judged arms against a different pool than the planner authored on (CLASS NN, assigned at merge; `docs/superpowers/AUDIT-CHECKLIST.md`).** `trader/one_setup_wiring.go:131` re-assembled the arm-time map with `AssembleScoredLevels` and no HTF extras, so a scenario on a D/4h/1h seat or a ZONE-* seat found no candidate at its price and was declined every cycle (one-setup is ON by default). Fix: `oneSetupSeedPlanLevels` seeds the plan doc's own levels into the live map (zero score, own identity) before `BuildMapCandidates` — seated for the planner == candidate for one-setup by construction. BOUNDED (second re-review): only a level the write site MACHINE-graded stands alone as a candidate; an unstamped level (stale / model-invented — the population one-setup exists to catch) may only alias into a live row within the 3-pt merge width (live grade wins) and is otherwise dropped and counted `one_setup:seed_unstamped_dropped`. On the golden fixture the unstamped PDL is dropped even at the 6-pt band (S1 stays declined); the machine-stamped copy arms. `TestOneSetupUnstampedSoloLevelIsNotACandidate`: unstamped solo → `level_no_candidate`, counter 2. Measured on the golden fixture's REAL arm path: BEFORE, S1/S2/S3 all `level_not_best:ONL@29494(B)`, zero arms; AFTER at the default band (flat tape, dATR 2.00 → 3 pt) UNCHANGED because the plan's levels at ±5 are outside that band; AFTER with the strategy's `proximity_filter_atr` at 3.0 (6 pt) S1 flips to ALLOWED (best = PDL A) and arms. `TestOneSetupArmSiteZoneSeatScenarioArms` / `TestOneSetupArmSiteDailyHTFSeatScenarioArms` drive `maybeManageArmedOrdersAt` end to end: S1 on a ZONE-4H-DEMAND seat / a `Demand·1d` seat → `level=ok`, best = that seat, ledger row armed at 29490. The E2 OFF golden is byte-identical (OFF builds no map).

**Fix 2 — the clone's freshness key.** The W7 level-state key is type-from-label + a 1.25-pt price bin (`trader/auto_trader_dayplan.go` provider); a ZONE-* edge clone sits in a different bin than its midpoint source, so a consumed zone resurfaced FRESH through its edge. Fix: `DetectedLevel.StateKeyLabel/StateKeyPrice` (json:"-"), set on the clone from its source row; the provider keys on them when present. Under `levels_fresh_by_tf` the grader is band-based (`Lo/Hi`) and already identical for the clone. `TestZoneSeatCloneFreshnessKeyedOnSourceRow`: source marked consumed → clone reads `done`, both grade `flipped` through the scorer; a keyless (no-source) clone reads its own state.

**Fix 4 —** the registry row named a function that does not exist; it now names `collapseLevelClusters`.

## Not done / known limits

- **Double-seat capacity cost (Round 25 must count it).** A zone taller than ~6.5 pt keeps its midpoint detector row AND its edge clone as two seats: they sit > 3 pt apart, so neither the collapse (rule 4) nor the validator's duplicate rule (`plan_doc.go:986-992`) folds them. One zone can therefore occupy two of the 12 seats and displace two references. Round 25 records, per read, how many seats ZONE-* rows hold whose source midpoint row is ALSO seated.
- **Pre-existing S1 duplicate listing (candidate class for whoever touches `structure_map` next).** `structureZonesFor` reads the pre-dedupe zone universe, so on a periodic tape the map lists the same (kind, lo, hi) band up to six times per TF (seen live in `TestZoneSeatsAtThePlannerReadCallSite`: `duplicates=10`). The zone-seat pass dedupes and records it; the map, its prompt section and its log line still show the duplicates.

- `GUIDE_BUILT_REV` is bumped at boot time by the deploy lane (the CLASS 141 pattern: a separate `guide:` commit before the boot), not here.
- The card's `planW3Map` (`api/handler_plan.go:2456`) still enriches by a no-extras map; the row itself renders from the doc (label, grade, price), so a ZONE seat shows — only the merged-names/role enrichment can miss. Same class as fix 1, card-side; not changed here.
- CLASS NN (one-setup judged arms against a different pool than the planner authored on) is in `docs/superpowers/AUDIT-CHECKLIST.md`; its number is assigned at merge.
