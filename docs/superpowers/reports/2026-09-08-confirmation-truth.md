# Confirmation truth — pre-change consumer census and stored-scenario blast radius

**Status: owner approved the blast radius; implementation and verification in progress. No deployment.** Sections 1–7 preserve the pre-change audit, whose pinned publication is `ba833a9d4113c792432cb08cf2ca238c8a07b01b`. The owner subsequently authorized both confirmation paths, validation changes, feeds-forward changes and a separate `1m_displacement` rule. Section 8 records implementation after that ruling; the historical statements about unchanged code in sections 1–7 describe the audit phase.

Lane: `confirmation-truth-96604090/root[unlisted]`, branch `fix/confirmation-truth`. Audit source HEAD: `5308c2bbcc010add0c8f7ba15a7d78e8da046507`, incorporating dev `63d902ac9c345e6e51cfd237b035d4062a59acf0`. At 2026-09-08 08:44:46 CT, `/api/health` reported `33672fdd2cd2`; systemd PID was `3260027`. The earlier executable read at 08:34:14 CT showed `33672fdd2cd2fee60a2c562a9693e06ab3b13551`, `vcs.modified=false`. The five core source files examined here are byte-identical between that running revision and audit HEAD.

Evidence grades: **[A]** directly read or executed; **[B]** inference from those facts; **[C]** unproven. Source `git log -1` results for every cited repository file are in [source-freshness.csv](2026-09-08-confirmation-truth-data/source-freshness.csv). The original basis is `docs/superpowers/reports/2026-09-07-planner-preparation-audit/README.md` at `6095ca58fe5901ba398be374e4f9d3488d0bed6b`, verified **77,345 bytes**. SYSTEM-MAP last change at audit: `6310eaf8`; checklist last change: `33672fdd`.

## 1. Answer for the ruling

[A] `BreakdownContinueState` has **four direct production call sites**, plus **six direct test call sites**. Its production consumers change confirmation, validation rejection, and planner input facts. There is **no direct arm-placement or exit call to this function**. A separate route through `EvaluateConfirm` **does** gate `wait_confirm` arm eligibility and therefore belongs in the same closed-bucket correction.

[A] Frozen corpus: **789 scenarios in 272 stored plan rows/versions**. A scenario identity is `(plans.rowid, scenario.id)`, with the full plan ID and version included in the CSV. There are **four** stored waterfall scenarios: three `breakdown_continue`, one `breakup_continue`.

[A] Input-only replay at **4,246 recorded decision snapshot instants** evaluated **627 unique scenarios**, producing **13,160 scenario observations**:

- **176/627 scenarios changed overall MET at least once**, across **299 observations**: **298 MET→NOT MET**, **1 NOT MET→MET**.
- Of these, **168 scenarios / 285 changed observations** have a reconstructed original verdict that matches the saved prompt's original MET/NOT MET at the changed observation.
- **162/789 scenarios were not replayed**. This is not a claim that those scenarios would be unchanged.
- Generic `wait_confirm` checks changed for **11 stored scenarios / 25 observations** in the audited single-arm shape. These are eligibility changes, **not 25 placements or trades**.
- Among the **four** stored waterfall scenarios, **one** changed overall confirmation; **two** changed validation PASS/REJECT when re-evaluated at those decision instants. All **286** waterfall observations' reconstructed overall verdicts match their saved prompt verdicts.

[B] This is a counterfactual of **closed-bucket semantics**, using retained completed 1m tape and unchanged production functions. It is not a count of historical bad trades, not a claim about all 789 scenarios, and not a historical write-time rejection census. The ordered-sequence fix is deliberately **not included**; combining it here would hide which changes come from the shared bucket predicate.

## 2. Every direct production consumer of the shared breakdown evaluator

| Call site | What it reads and changes | Downstream use | Effect of requiring closed 5m buckets |
|---|---|---|---|
| `kernel/plan_confirm.go:191`, `EvaluateScenarioConfirm` | `Leg1Met`, `Leg2Met`, `Reclaimed`, `BreakLegPts` become a confirmation verdict. At `:192` the leg is labeled `<n>x5m_close` despite the shared evaluator currently counting raw 1m bars. | Executor advisory through `RenderConfirmLines` (`:344`) → `kernel/engine_analysis.go:507`; plan-card display through `trader/auto_trader_levelstate.go:250`. | **Confirmation + display; indirect AI behavior.** A forming bucket cannot supply a confirming close. A 1m reclaim that disappears inside the 5m bucket can cease to void the play. No direct order/exit call here. |
| `kernel/breakdown_continue.go:252`, `ValidateBreakdownContinueScenarios` | `Reclaimed` rejects at `:257`; missing `Leg1Met` rejects pullback mode at `:260`; `BreakLegPts` fails the displacement floor at `:264`. | Live planner write validator at `trader/auto_trader_planner.go:1736`; shadow validator at `trader/rootfix_shadow_ab.go:159`. | **Validation reject: behavior changes in both directions.** A premature qualifying minute can no longer authorize a 5m condition. Conversely, a minute-only reclaim can stop causing rejection. Live validation drives repair/retry and possible failure to publish; shadow validation changes comparison results only. |
| `kernel/class45_feeds_forward.go:57`, `BreakdownLevelReclaimed` | Reads `Reclaimed` to declare a level void; then obtains its displayed time separately. | `ComputeVoidBreakdownLevels` at `:98` → `trader/auto_trader_planner.go:2329` → `kernel/planner_prompt.go:459` (`RenderVoidBreakdownLevels`). | **Planner fact/display; indirect validation/authoring effects.** The list must continue to agree with the corrected validator. It can lose minute-only voids. AI scenario selection can change; the specific downstream orders cannot be inferred. |
| `kernel/displacement_feeds_forward.go:49`, `displacementProbe` | Returns shared state; `ComputeLevelDisplacements` at `:65–69` selects a side only when `Leg1Met` and uses `BreakLegPts`. | `trader/auto_trader_planner.go:2334` → `kernel/planner_prompt.go:464` (`RenderDisplacementLines`). | **Planner fact/display; indirect authoring effects.** “Broken,” direction and measured displacement can change when only completed 5m buckets qualify. A value that reaches the validator's displacement threshold can therefore differ. |

[A] Exhaustiveness command: `git grep -n 'BreakdownContinueState(' -- '*.go' ':!docs/**'`. The declaration is `kernel/breakdown_continue.go:135`; the four production call sites above are the entire production result. No function-value alias or additional production reference was found by the broader name census.

[A] The six direct test calls are `kernel/breakdown_continue_test.go:133`, `:143`, `:207`; `kernel/class45_pin_test.go:64`; `kernel/displacement_feeds_forward_test.go:117`; `kernel/entry_law_test.go:163`. Their expected results must be reviewed during the implementation wave; no tests were rewritten in this audit.

### Complete downstream map, including the other confirmation route

| Caller / consumer | Classification | Current consequence and change |
|---|---|---|
| `kernel/plan_confirm.go:107` → `AcceptanceRunEver` (`kernel/scenario_facts.go:440`) | Confirmation | Sole production call to this ever-fired counter. It aggregates to the named TF but receives no evaluation clock, so it counts a forming final bucket. Fixing this path changes generic close confirmations. |
| `kernel/plan_confirm.go:209`, `:219` → `EvaluateConfirm` | Confirmation / sequence | Computes each scenario leg. `firstConfirmFireMs` has one production caller at `:215`; the zero fallback is at `:216–217`. The future ordered fix changes this route separately from the closure-only replay here. |
| `kernel/plan_confirm.go:344` → `EvaluateScenarioConfirm`; `kernel/engine_analysis.go:507` → `RenderConfirmLines` | Confirmation fed to executor AI | Changes advisory MET/NOT MET, stale annotations and the opposing-confirmations warning. AI can choose different entries or management; these facts are not a deterministic exit command. |
| `trader/auto_trader_levelstate.go:250`, persisted at `:254` | Display record | Stores the same confirmation object under `ScenarioMetaKey`. No placement is performed by this write. |
| `api/handler_plan.go:2279` → `scenarioMeta`; `web/src/components/plan/ScenarioList.tsx:404–405` → `ConfirmChip` at `:75` | Display | Reads and renders the stored confirmation object. Presently the explanatory detail is primarily a tooltip; the requested bucket/close evidence still needs implementation. |
| `trader/armed_executor.go:454` → `EvaluateConfirm` | **Arm eligibility** | A `wait_confirm` leg remains dormant when false; true continues to arm gates and row creation. This is a real behavior change even with no edit to this caller. Multi-leg arms select `Confirm2`; single arms select `Confirm`. The caller currently evaluates that selected leg independently from plan birth, not the overall ordered scenario verdict. |
| `trader/armed_executor.go:177` → `EvaluateConfirm`, in `declineHadFreshMet` | Telemetry | Caller `trader/auto_trader_loop.go:791–792` increments `decline_fresh_met` for a superseded wait. A corrected result changes the counter; it does not turn that discarded wait into an order or exit. |
| `trader/auto_trader_planner.go:1736` → shared validator | **Live validation rejection** | Error triggers rejection bookkeeping, repair/retry and a `continue` at `:1745`; it can change which plan is accepted or whether the read exhausts its attempts. |
| `trader/rootfix_shadow_ab.go:159` → shared validator | Shadow validation result | Appends failure reasons for the shadow comparison. Does not publish the shadow as the live plan or submit an order. |
| `trader/auto_trader_planner.go:2329`, `:2334` → void/displacement producers | Planner inputs | These facts are rendered at `kernel/planner_prompt.go:459`, `:464`. Correcting the common predicate changes their contents, not level scoring or seating algorithms. |
| `trader/desk_facts.go:139–149` | Display, currently **no confirmation consumer** | The current strip has 12 rows and no confirmation field. The requested confirmation row would be a new reader of the recorded verdict, not another evaluator. |

[A] **Direct exits: zero in these call graphs.** Existing plan death/flip and level facts use other functions. `AcceptanceBars` is an aggregation utility, **not** the shared yes/no closure predicate: it also supplies ATR and structure consumers (`trader/armed_executor.go:172`, `trader/entry_gate.go:336`, `trader/auto_trader_planner.go:2826`, `kernel/structure.go:396` through `StructureAggregateToMinutes`). Globally turning that utility into a closed-only data source would also alter these measurements. This audit did not do that. The new closure evaluator must be invoked where closure is being judged; an aggregation-only caller must not acquire an unrelated behavior change by accident.

[B] **Indirect AI-managed exits are not provably unchanged.** The executor sees different advisory facts. No replay of hypothetical AI responses or claim of a specific resulting exit is made.

## 3. Four stored waterfall scenarios: row IDs and observed counterfactual changes

Each row is one stored scenario, not a count of fills. Plan rows are identified by SQLite `rowid`; [scenario-census.csv](2026-09-08-confirmation-truth-data/scenario-census.csv) includes full plan IDs and versions.

| Plan row / session / version / scenario | Decision observations | Overall confirmation flips | Validation PASS/REJECT flips |
|---|---:|---:|---:|
| **163**, 2026-08-30 ASIA v2 **S1**, short immediate | **128**, IDs **34786–34926** (not every integer exists) | **0** | **22 REJECT→PASS** |
| **181**, 2026-09-01 LONDON v5 **S1**, short immediate | **24**, IDs **35640–35663** | **0** | **0** |
| **183**, 2026-09-01 NY v1 **S1**, short immediate | **11**, IDs **35681–35692** | **1 NOT MET→MET** | **1 REJECT→PASS** |
| **273**, 2026-09-08 LONDON v5 **S2**, long pullback, arm enabled | **123**, IDs **38212–38334** | **0** | **0** |

[A] Thus **1/4** changes its overall confirmation and **2/4** change a validation pass/reject result in this replay. **23 validator flips** all remove a rejection. These are **revalidations at recorded decision instants**, not observed original planner writes. The original accepted-request tape and exact validation instant are not preserved for all four plans, so an exact historical write-rejection count is **UNKNOWN**.

[A] All **four** stored waterfall scenarios have a changed numeric state field: `BreakLegPts` differs in **31 observations**, and `LastClose` in **162**. The plan-birth-scoped `Leg1Met`, `Leg2Met`, and `Reclaimed` flags each differ in the single row-183 observation. The validator uses its separately resolved session scope, which explains why its changes also include row 163. These numeric-state counts do **not** measure changes to all historical seated-level void/displacement prompt lists: those producers probe both sides of every ranked level, beyond the four authored scenario levels. Their complete historical output-change count remains **UNKNOWN**.

[A] Concrete loosening example: plan row **183**, NY v1 S1, decision **35691**, snapshot **2026-09-01 08:49:08 CT**, level **29122.75**:

```text
unchanged evaluator on completed 1m tape:
  Leg1Met=false Leg2Met=false Reclaimed=true BreakLegPts=82.75
same evaluator on completed 5m buckets of the same scoped tape:
  Leg1Met=true Leg2Met=true Reclaimed=false BreakLegPts=82.75
saved prompt overall verdict: NOT MET (matches the original replay)
validator original: "S1 breakdown_continue: a close came back across 29122.75 — the breakdown is void; ..."
validator counterfactual: PASS
```

The not-yet-closed 08:45–08:50 bucket cannot provide a completed 5m reclaim at 08:49:08. The original minute-level reclaim can therefore stop voiding an already confirmed break. This is why the change is not merely “fewer early confirmations.” See [all waterfall observations](2026-09-08-confirmation-truth-data/breakdown-observations.jsonl).

[A] The generic arm-check changes occur at **11 scenario keys**: `141/S1`, `142/S2`, `143/S1`, `143/S2`, `143/S3`, `145/S3`, `153/S4`, `253/S2`, `258/S2`, `259/S2`, `273/S1`. These are stored single-arm configurations with `arm.enabled`, `arm.wait_confirm` and `confirm` present. The **25** changed checks are published in [changed-observations.jsonl](2026-09-08-confirmation-truth-data/changed-observations.jsonl). Other entry gates were not replayed, and no historical placement count is inferred.

## 4. Replay method, exclusions and limits

[A] The DB was opened with `mode=ro`, `PRAGMA query_only=ON`, and one read transaction for a frozen corpus. [summary.json](2026-09-08-confirmation-truth-data/summary.json) records capture time and counts. The corpus contains **19,011 retained MNQ 1m bars**. All source/current configuration reads were non-mutating; no credentials or account names are included.

1. Enumerate every stored scenario in all **272** plans, retaining `(rowid, plan_id, version, scenario.id)`.
2. Link decision rows by the stored plan ID/version. Parse the actual `Snapshot: HH:MM:SS CT` and the input's dated CT clock. **Do not replace a missing snapshot with the later decision-save time.**
3. At each snapshot, take the preceding completed 1m bars, capped by the running source's `kernel.AISVPBarCount`. Historical partial-minute OHLC is not reconstructed from that minute's later final value.
4. Baseline: call the unchanged `kernel.EvaluateScenarioConfirm`. Counterfactual generic close leg: window first, aggregate with `kernel.AcceptanceBars` for its named rule, retain only buckets whose true end (`CloseTime+1`, repo's inclusive-last-millisecond convention) is `<= snapshot`, then call unchanged `EvaluateConfirm` on that pre-windowed tape.
5. Counterfactual waterfall: apply the same original plan/session window **before** aggregating to completed 5m buckets, then call the unchanged breakdown/validation functions. The second window pass is disabled for already-windowed inputs. Otherwise canonical 5m open times preceding a mid-bucket plan birth would accidentally add a new whole-bucket publication restriction. This correction was made before final counts.
6. Keep the original sequence-reference behavior in this **closure-only** comparison. The audit source copies the old unexported reference helper verbatim into the isolated runner for that purpose; no production helper is replaced. The ordered-reference repair and its additional behavior changes remain outside this count.
7. For validation comparisons, hold price, ATR and all other inputs fixed between arms of the comparison. Change only the scoped tape delivered to the existing validator. This holds the displacement floor constant; any different `BreakLegPts` comes from the changed eligible buckets.
8. Compare reconstructed baseline overall verdicts to saved prompt verdicts when present. **12,260 of 12,550 match**; **290 do not**. Causes can include missing historical partial-minute values, retained-tape revisions and different historical code/configuration. They are not silently declared faithful historical replays. Of the changed scenario set, **168** have at least one changed observation whose original result matches the stored prompt; the headline **176** is the broader computed retained-tape result.

[A] Of **4,554** linked decisions, **308** were excluded: **138** missing recorded snapshot clocks; **95** plan births preceding retained tape; **73** snapshots preceding the bound plan's publication; **2** with the newest completed minute more than 120 seconds old. **4,246** decisions remain. The **162 un-replayed scenarios** include versions with no usable linked observations, not merely the sum of these decision exclusions. Every scenario appears in the CSV with its coverage label.

[A] Audit correction: a preliminary run counted a third waterfall scenario as validation-changing. Its only such observation, decision **35639**, was stamped **07:07:28 CT**, while its bound plan row **181** was published **07:07:32.457 CT**. It cannot establish a verdict for a plan that did not yet exist. Excluding all **73** such misaligned records reduced the final validation scenario count to **two**. It did not change the 176 overall-confirmation count.

[B] Final minute bars still do not reconstruct historical partial-minute values or prove an identical old cache. These counts describe the stated counterfactual and coverage, not an exact causal reconstruction of all live decisions. The corpus uses current default resolvers (the allowlisted env overrides were absent), not a fabricated history of environment values.

Reproduce from audit source HEAD `5308c2bbcc010add0c8f7ba15a7d78e8da046507`, with the evidence files available, without editing any production file:

```bash
cp docs/superpowers/reports/2026-09-08-confirmation-truth-data/replay.go.txt /tmp/confirmation-truth-replay.go
go run /tmp/confirmation-truth-replay.go docs/superpowers/reports/2026-09-08-confirmation-truth-data/replay-corpus.json > /tmp/confirmation-truth-replay-results.json
```

The committed corpus keeps evaluator inputs and confirmation lines; unrelated plan prose is omitted. [manifest.json](2026-09-08-confirmation-truth-data/manifest.json) gives bytes and SHA-256 for every evidence file.

[A] The minimized public corpus was re-run through the isolated replay. Every observation and every result field other than the deliberately minimized plan documents matched the private extraction's replay exactly. All evidence byte counts and SHA-256 values were then checked against the manifest. The CSV independently reproduces 789 stored / 627 evaluated / 176 changed / 168 record-supported scenarios, 11 arm-check-changing scenarios and two validation-changing scenarios.

## 5. Original premises and the surprise that prompted this ruling

[A] **C1 reproduced.** All four pinned audit probes were re-run against the unchanged current kernel source. The incomplete-5m above/below probes both returned `actual_met=true` when the expected result was false. The counter path is `EvaluateConfirm` (`kernel/plan_confirm.go:107`) → `AcceptanceRunEver` (`kernel/scenario_facts.go:440`), which does not receive `now` and counts all aggregated buckets.

[A] **Live C1/C2 instance:** decision **38329**, LONDON v5 S3, snapshot **08:16:59 CT**; saved **08:17:50.474 CT**:

```text
S3 confirm: leg 1/2 touch — MET · leg 2/2 1x5m close — MET → overall MET
(last 1x5m close 29656.75 (best run 48/1 closes below 29661.25 since plan birth))
current 5m bar: FORMING (closes 08:20 CT) — prior bars closed
```

The executor returned `wait`. The stored scenario's first touch is at reference **29661.25**; the completed **08:15** 1m bar reaches **29662.00**. The later completed **08:19** bar closes **29667.00**, above the reclaim reference, so the replay must not assume a valid reclaim at 08:20. The live episode exhibits both hazards: the current forming bucket is presented as a close, and earlier pre-touch closes can satisfy the second leg. A closure-only fix is not evidence that the sequence defect has disappeared. [Saved snapshot and decision](2026-09-08-confirmation-truth-data/decision-38329-snapshot.json).

[A] **C2/C3 reproduced.** The long and short “reclaim before sweep” probes both returned MET. `firstConfirmFireMs` (`kernel/plan_confirm.go:238–240`) returns zero for `touch`; its sole caller at `:215–217` treats nonpositive as missing and substitutes `sinceMs`. Zero is a missing-result sentinel here, **not a found epoch event**. The fallback falsely gives the second condition the whole plan-birth window. The first-close helper also uses `AggregateBars`, whose implementation (`kernel/fvg_entry.go:290`) does not populate `CloseTime`; its apparent closure check cannot establish the bucket's real end. All of these need the one resolved reference/closure contract in the eventual fix.

[A] **C4 verified, separate lane.** Active plan `BirthMs` is `row.CreatedAt.UnixMilli()` at `trader/auto_trader_planner.go:2651`; confirmation receives it at `trader/auto_trader_levelstate.go:250` and `kernel/engine_analysis.go:507`. The window helper keeps bars whose open is at/after that birth (`kernel/plan_lifecycle.go:156`). It is publication-based, not the facts' authoring timestamp. No change to this policy or the plan-liveness lane's implementation is made here.

[A] **The extra shared-path probe:** at **08:16 CT**, one completed 1m bar from **08:15** produced a `1x5m_close` breakdown MET verdict even though the named bucket closes **08:20**. `EvaluateScenarioConfirm` calls the shared raw-bar evaluator at `:191` and labels it 5m at `:192`. [Probe source](2026-09-08-confirmation-truth-data/breakdown-probe.go.txt), [result](2026-09-08-confirmation-truth-data/breakdown-probe.json), [original four probe results](2026-09-08-confirmation-truth-data/running-rev-probe.json).

## 6. Contract points requiring the owner's blast-radius ruling

1. **Both stricter and looser outcomes:** correcting the named timeframe suppresses premature confirmations, but also removes minute-only reclaims from 5m invalidation/void judgments. The 23 replayed validator REJECT→PASS observations quantify the latter for two stored scenarios.
2. **Immediate authoring has an explicit pre-confirm intent.** `kernel/breakdown_continue.go:242–250` says immediate-mode authoring is legal as soon as displacement exists; the confirming close is the entry trigger. If that fact genuinely needs completed 1m evidence before a 5m bucket closes, it must become an explicitly named 1m displacement/observation rule, separately from the 5m confirmation. It must not make a `5m-close` function return minute semantics for this caller. No such exception is implemented or presumed approved here.
3. **Reclaim timestamps must follow the same event.** `reclaimStampCT` (`kernel/class45_feeds_forward.go:67–80`) currently scans raw minutes separately after the authoritative shared verdict. Leaving it unchanged would pair a corrected 5m verdict with an earlier minute timestamp. The future reference-instant function must supply the event actually judged.
4. **Arm chaining is execution-related.** The generic `EvaluateConfirm` caller at `trader/armed_executor.go:454` is a real eligibility seam. Its independent selection of `Confirm2` also needs an explicit decision in the ordered-sequence scope; fixing the displayed aggregate alone would not establish ordered arm authorization. The caller was not edited during this census.
5. **No direct exit consumer was found**, and aggregation-only ATR/structure readers must remain distinct from the named closure predicate. Any behavior change beyond the audited shared confirmation/validation/fact paths must be reported before implementation.

## 7. Current live-surface truth and remaining work

[A] The owner can still see a forming bucket labeled MET, missing sequence references silently falling back to plan birth, an apparent 5m breakdown driven by minute bars, and bare confirmation chips with evidence hidden in a tooltip. The desk strip has no confirmation row. None is reported fixed.

Implementation, red/green gate-level fixtures, actual mutation failures, closed-case golden comparison, Guide/SYSTEM-MAP/boot line, checklist census at merge, merged-head Go/vitest/tsc, build, deployment gate and live verdict proof are **pending the owner's ruling**. The raw probes reproduce defects; they are not a green suite or a shipped fix. No new checklist number is claimed before the implementation merge.

Rollback for this phase: documentation/evidence only; no binary swap, RELEASE change, DB migration or runtime rollback is required. The engineering branch is retained for the subsequent authorized fix.

## 8. Owner ruling and implementation (2026-09-08)

[A] Owner ruling: apply the fix to every listed confirmation consumer; accept removal of minute-only reclaim rejections; preserve immediate-mode pre-confirmation displacement with its own named function and tests; quote the changed feeds-forward facts; retain the unevaluated population. The **162 are scenarios**, not an inferred collection of usable observations. The 168 record-supported scenarios do not fill their missing evidence.

Implementation base after integrating plan-liveness: `c3fd6f6f013b130d9651066407512d6386cfa393`. This merge incorporates dev `b0f95bc6`; the current SYSTEM-MAP last-change receipt was `94f0d7df`, checklist `393712c1`. Core changes use one explicit-clock `EvaluateBucketClose` predicate, one `(instant, ok)` reference lookup, and a separately named `Evaluate1mDisplacement`. Waiting arms and decline telemetry read the ordered scenario verdict. The stored versioned verdict, card and desk carry bucket/reference evidence. Immediate displacement is shown separately in planner facts; directions already void on completed 5m closes cannot be presented as immediate preparation candidates.

[A] **The approved carve-out adds one more validation PASS.** In plan row **163 / S1 / decision 34790**, the original minute evaluator rejected a reclaim. The closure-only audit removed that reclaim but still rejected its **23.75-point closed-5m displacement**, below the resolved floor rendered as **23.8 points**. The explicitly separate completed-1m displacement meets the floor. Final approved semantics produce **24 REJECT→PASS observations**, versus **23** in the frozen closure-only audit, across the same two scenario identities. The boot receipt keeps **23** labeled as the closure-only audit and separately reports **24 with 1m_displacement (+1)**; neither number is a live refusal counter.

[A] RED before implementation: decision **38329** returned MET at **08:16:59**, with best run **48/1**; each of `1x5m_close`, `2x5m_close`, and legacy `15m_close` returned MET one second before its bucket ended; reclaim-before-touch returned MET; missing touch lacked UNKNOWN; a single completed minute satisfied the waterfall's 5m leg. On the pinned baseline, with only clock injection added, the actual arm manager wrote **one row** before the 5m close and also armed the out-of-order sequence. The UI rendered only `1×5m not met` and `touch not met`, hiding the bucket reason and UNKNOWN. The desk-row pin failed because the row was absent.

[A] Initial corrected kernel/trader pins pass, as do the first full Go suite and frontend **51 files / 368 tests**, plus TypeScript. These are preliminary source checks; final merged-HEAD verification, mutations, build receipt and live proof are not yet claimed.

Existing fixture review: five tests initially failed semantically—`TestBreakdownContinueValidatorRealTape`, `TestBreakupContinueMirror`, `TestClass45PinLondon0132`, `TestRehearsalS4CaseStillRejects`, `TestLegacy15mConfirmStillEvaluates`. The real waterfall tape is unchanged; only its synthetic retest extension now completes its 5m bucket. Synthetic mirror/London stages now contain complete 5m buckets. The rehearsal keeps its original minute tape and correctly refuses a missing 5m break without inventing a 5m reclaim. The legacy 15m test now evaluates a completed 15m interval. Two additional weak-displacement/reclaim tests could falsely pass on the new missing-close error; their synthetic tapes now close the required buckets and their assertions name the intended rejection. The desk shape test now includes the new thirteenth recorded-confirmation row. A complete fixture/change and mutation receipt follows after final verification.

[B] Corrected feeds-forward facts intentionally change what the planner knows; subsequent authoring may differ. The original 31 displacement and 162 last-close changes are **plan-birth-scoped scenario states**, whereas planner facts use their existing resolved session scope. Those populations must not be mislabeled as identical.
