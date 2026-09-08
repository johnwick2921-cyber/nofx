# Scenario economics — corrected premises and implementation evidence

**Owner ruling received after the initial STOP:** approved new-authoring-only completeness; unchanged legacy UNKNOWN acceptance; role differences WARN + counter; corrected C5 arithmetic; C6 NOT ESTABLISHED and dropped. D4's three contradiction refusals apply to new authoring. The original STOP/evidence below remains historical evidence, not the current instruction to stop. Implementation is in progress; no economics boot has occurred.

2026-09-08. Dispatch: `fix/scenario-economics`. Session: `scenario-economics-83f741b2/root[unlisted]`.

**[A] STOP under A23 and E6. No production code, configuration, database, binary, Guide, or serving assets changed. No build or boot was attempted.** This report is the required measurement before implementation, not a declaration that the economics contract shipped. The existing combined confirmation/liveness boot remains `f8bc7044`.

## Decisions needed before building

1. **C5 transcription correction:** `E[net R] = p·b − (1−p) − c`, with a gross losing outcome of −1R and cost cR. The dispatch instead multiplies the losing probability by c. Its break-even table is correct; its expectation equation is not. Research §03 already states the correct equation.
2. **E2 versus E6/UNKNOWN:** all **799/799 retained scenario documents** lack `first_obstacle`; all **280/280** in the frozen sample lack it too. Requiring the field when replaying unchanged legacy documents would refuse 100%, beyond E6's 5% STOP threshold. This is absent evidence, not a measured contradiction. A proposed resolution, **not implemented**, is an explicit contract version: require the complete contract only for newly authored candidates, accept legacy documents unchanged with missing economics UNKNOWN, and never fabricate obstacle/response/R values. The owner must resolve E2's missing-field refusal against D4-only/UNKNOWN and E6 before coding.
3. **C6 correction:** London v2's authored widths are 36.25, 36.00, 36.25 and 36.50 points, but its stored 5m ATR14 is 23.8431; 1.5× is **35.76465**, smaller than every authored width. These documents do not establish that the ATR floor bound all four composed stops. Exact historical composition inputs/results are not reconstructed; retain that result as UNKNOWN, not an invented frequency.
4. **C4 warning semantics:** the two literal London role/use differences reproduce. Free-text map instructions have no exhaustive role compatibility table. A narrow, explicit two-family census finds 172 level/scenario diagnostics affecting 150/799 scenarios; that is not 150 proven economic contradictions or independent episodes. Proposed disposition remains WARN. It must not become a refusal merely because a lexical count exceeds ten.

The scenario contract still has a measured reason to exist: C1–C3 reproduce on the frozen members, and 13 target/path mismatches exist among all retained scenarios. No target family or risk-floor change is proposed. Stage A retains ownership of recorder changes.

## Runtime, scope, and fresh evidence

[A] Accept base `72bb9de08f7fee83f2b45a5219ec14bd17a30340`, current `origin/dev` at accept. The isolated, locked worktree was `/tmp/nofx-scenario-economics`. Required claim helper created `fix/scenario-economics`; the subsequent **ls-remote SHA** was `7262e50253cdb4f45d529b0463ccd914b37cb7a2`.

[A] At 15:36:55 CT, `/api/health` returned `revision=f8bc7044cc44`, `status=ok`; `/proc/3566770/exe` reported full revision `f8bc7044cc44d58e84904a0a7761e78b420404af`, `vcs.modified=false`. Only nonsensitive floor overrides were inspected: neither `MIN_SL_ATR_MULT` nor `ARM_STOP_ANCHOR_MAX_ATR` was set. The code defaults are 1.5 and 3.0. [Runtime receipt](2026-09-08-scenario-economics-data/runtime.json).

[A] The independent `/api/plan/today` read at **15:35:59 CT** returned HTTP 200, `found=false`, `is_active=false`, `active_session=""`, `session=""`, scenario IDs `[]`. London v2 is a retained historical document at this observation, not the active plan. Its `plans.rowid=270`, creation **2026-09-08 02:08:48 CT**. Calling it “live London v2” now would be wrong.

[A] `measure.py` used a single read transaction through SQLite `mode=ro`. It read authored JSON and the immutable indicator block, independently computed Decimal arithmetic, and did not call production geometry/validation functions. No trades or P&L rates were computed. The date floor follows `store.DayPlanEraStart`'s **2026-08-15 CT** date; no pre-era rows were included. `pnl_corrected`/NULL/exclusion rules remain applicable to any later outcome study, but no outcome columns were queried here.

The retained database had **276 plan rows, 799 scenarios**. Two weekly documents (rows **223, 257**) have no scenarios, hence 274 scenario-bearing plans. The frozen CSV's 280 members map to 100 scenario-bearing plans, plus those two weekly plans gives the cited 102 versions. All 280 logical scenario identities matched the current database; all 111 complete frozen entry/stop/target triples matched their raw CSV fields exactly. Membership, row IDs, versions and scenario indexes are saved for every row; repeated versions are not independent trades.

- [Reproduction script](2026-09-08-scenario-economics-data/measure.py)
- [Census and complete plan-row denominator](2026-09-08-scenario-economics-data/census.json)
- [Every scenario and derived geometry](2026-09-08-scenario-economics-data/scenarios.json)
- [Every literal role diagnostic](2026-09-08-scenario-economics-data/role-diagnostics.json)

## Pinned research and source freshness

Both requested documents are evidence at their specified immutable revisions. Neither path exists on dev in the requested form. They were not substituted with a branch URL or treated as a new dispatch.

| Basis | Exact revision | HTTP / downloaded bytes = Git blob bytes | `git log -1 -- <file>` at that revision |
| --- | --- | --- | --- |
| [Trading-policy research](https://raw.githubusercontent.com/johnwick2921-cyber/nofx/982091d4d908f4a5b8b65022cedf5b8c7c8202d5/docs/superpowers/research/2026-09-08-trading-policy/README.md) | `982091d4d908f4a5b8b65022cedf5b8c7c8202d5` | 200 / 67,959 = 67,959 | `0ec5bd2c 2026-09-08T08:30:15-05:00 docs(research): publish four-policy trading study with 28 cited sources` |
| [Planner-preparation audit](https://raw.githubusercontent.com/johnwick2921-cyber/nofx/6095ca58fe5901ba398be374e4f9d3488d0bed6b/docs/superpowers/reports/2026-09-07-planner-preparation-audit/README.md) | `6095ca58fe5901ba398be374e4f9d3488d0bed6b` | 200 / 77,345 = 77,345 | `6095ca58 2026-09-07T23:09:02-05:00 docs: record owner approval for public audit evidence publication` |

Research §02 separates authored geometry from actual admission/execution and corrects the six under-2R records. §03 requires explicit obstacle response and explains why a favorable final R is not evidence of profitable behavior at a nearby obstacle. §07 connects causal preparation, permission, geometry and management, and prohibits inventing partial exits. §09 says **“Do not prescribe now: a mandatory 1R first target”**. Its priority two is consistency between obstacle, response, arm target and risk geometry. §05 preserves a level's possible target/obstacle/invalidation usefulness even when it is unsuitable for entry.

The dispatch's governing RESEARCH LAW is: **“MAY: require that a scenario state its first opposing obstacle, its planned response there, its arm target and the implied R”** and refuse contradictory numbers; it forbids selecting a target family or declaring a placement superior. [O] This report chooses **no structural, fixed-R, ATR, partial, trailing or mandatory-1R target policy**. Numerical illustrations below are arithmetic, not market beliefs or estimated NOFX rates. No new [R]/[I] profitability claim is introduced.

[Source-freshness receipt](2026-09-08-scenario-economics-data/source-freshness.txt) gives exact `git log -1` for **every production/source file cited here** and verifies each is byte-identical to running `f8bc7044`. `SYSTEM-MAP.md`: `e020885b 2026-09-08T14:21:34-05:00`; `AUDIT-CHECKLIST.md`: `78eed09b 2026-09-08T14:32:14-05:00`. The audit follows its PART 2 R1–R10. No implementation spec was taken from a stale worktree base.

## C1–C3: independent recomputation

[T/A] These are authored geometries before execution rounding, costs or arm-time stop composition. “First listed target” does **not** mean that the author has identified it as the first opposing obstacle; that missing declaration is part of this wave's proposed contract.

| Measurement | Frozen membership | All retained scenario documents |
| --- | --- | --- |
| Scenarios | 280 | 799 |
| Complete positive entry/stop/arm target, nonzero risk | 111 | 177 |
| Positive directional first-listed target distance below 1R | 45/111 = 40.54% | 69/177 = 38.98% |
| Arm target R below 2 | 6/111 | 17/177 |
| Correct-side stop AND target | 111/111 | 177/177 |
| Arm target absent from path, half-tick tolerance 0.125 | 4/111 | 13/177 |
| Arm target absent from path, D2 one-tick tolerance 0.25 | 4/111 | 13/177 |

All contributing IDs and exclusions are explicitly enumerated in `scenarios.json`: `complete`, `sub1`, `arm_under2`, `absent_half_tick`, `absent_tick`, and `frozen`. The remaining **622/799** lack a complete stored arm geometry, including **169/280** frozen scenarios. Their geometry remains UNKNOWN; no entry/stop is supplied from prose. The displayed arm is optional (`kernel/plan_doc.go:53`, `:90`); D1 must not accidentally mandate an enabled arm or expand the armable set.

### C2 — all six dispatched rows reproduce

| Plan row / scenario | Date/session/version | Entry | Stop | Arm target | Arm R |
| --- | --- | --- | --- | --- | --- |
| 170/S1 | 2026-08-31:NY:v3 | 29351.47 | 29408.52 | 29280.88 | 1.237336 |
| 180/S1 | 2026-09-01:LONDON:v4 | 29182 | 29212.5 | 29154.38 | 0.905574 |
| 182/S1 | 2026-09-01:LONDON:v6 | 29123.25 | 29147.25 | 29085 | 1.593750 |
| 185/S2 | 2026-09-01:NY:v3 | 29085 | 29125 | 29062.75 | 0.556250 |
| 187/S1 | 2026-09-01:NY:v5 | 29100.5 | 29130 | 29082.75 | 0.601695 |
| 252/S1 | 2026-09-04:NY:v2 | 29611.25 | 29481.5 | 29720 | 0.838150 |

Minimum in the frozen group is **0.556250**, row **185/S2**. The formula uses unrounded stored prices: `abs(target-entry)/abs(entry-stop)`. These observations do not reconstruct the contemporaneous R:R setting or prove that an under-floor order was admitted. For completeness, the other eleven retained under-2R observations are:

| Plan row / scenario | Date/session/version | Arm R |
| --- | --- | --- |
| 141/S1 | 2026-08-27:ASIA:v8 | 1.340291 |
| 142/S3 | 2026-08-27:ASIA:v9 | 1.916667 |
| 144/S3 | 2026-08-27:ASIA:v11 | 1.600000 |
| 145/S2 | 2026-08-27:ASIA:v12 | 1.766181 |
| 145/S3 | 2026-08-27:ASIA:v12 | 1.417023 |
| 146/S1 | 2026-08-27:ASIA:v13 | 1.350814 |
| 147/S2 | 2026-08-28:LONDON:v1 | 0.931507 |
| 147/S3 | 2026-08-28:LONDON:v1 | 1.388889 |
| 149/S2 | 2026-08-28:LONDON:v3 | 1.934524 |
| 272/S1 | 2026-09-08:LONDON:v4 | 0.527696 |
| 272/S3 | 2026-09-08:LONDON:v4 | 0.205964 |

### C3 — all four dispatched mismatches reproduce

| Plan row / scenario | Date/session/version | Arm target | Path |
| --- | --- | --- | --- |
| 178/S1 | 2026-09-01:LONDON:v2 | 29418.62 | 29447.5, 29435.75, 29422.12 |
| 194/S3 | 2026-09-01:ASIA:v7 | 29131.66 | 29099.85, 29113.23 |
| 230/S4 | 2026-09-02:ASIA:v14 | 29214.5 | 29235, 29228.75, 29218.25, 29212.5, 29207.5 |
| 259/S1 | 2026-09-06:ASIA:v2 | 29575.48 | 29545, 29566.02, 29587.75 |

Each remains absent at one tick, so changing the tolerance from half a tick to D2's one tick does not remove these four. The running code validates positive chain prices (`kernel/plan_doc.go:646`), while the arm gate computes R from `leg.Target` (`trader/armed_executor.go:1903`). The plan card renders `target_chain` at `web/src/components/plan/ScenarioList.tsx:216`. The prompt calls the chain guidance at `kernel/planner_prompt.go:722`. These different meanings are confirmed, not a claim that all displayed target prices are executable orders.

## C4 — literal role/use evidence; no new role rejection

[T/A] Row **270**, London v2:

| Scenario | Map index (zero-based) / level | Map role | Scenario use |
| --- | --- | --- | --- |
| S2 | 9 / Supply·1h, 29657.38, grade C | confluence | arm target and second target-chain value |
| S3 | 10 / SWG-L·5m, 29675.75, grade A | target | explicit `5m close above 29675.75` invalidation |

The census intentionally tests only two narrow diagnostic families: a pure confluence instruction referenced as an arm/path target, and a pure target instruction referenced numerically as invalidation. It normalizes whitespace/underscores and explicitly lists recognized synonyms in the saved script. Price matching is within 0.25 point. It counts a level/scenario once even when both arm target and path reference it. It does not invent entry references from unrelated numbers or treat arbitrary free text as an enum.

**Frozen: 81 level/scenario diagnostics, affecting 73/280 scenarios. All retained: 172 diagnostics, affecting 150/799 scenarios.** Every row, instruction, price and use is in `role-diagnostics.json`. This is a measured lower bound for these two literal families, **not an exhaustive semantic role-disagreement rate**. Other uses/prose remain unevaluated, not silently compatible or incompatible. No historical structured exception field exists. It would require a policy decision to interpret every role as exclusive; the research explicitly permits several uses. The lexical n exceeds ten but is not a basis here for choosing a role-refusal policy. WARN remains the proposed disposition pending the owner's correction ruling.

## C5 — corrected arithmetic and contract-size boundary

[A] For gross outcomes +bR and −1R, and cost cR on the trade:

`E[net R] = p*b - (1-p) - c = p*(b+1) - 1 - c`

Setting this to zero gives `p = (1+c)/(1+b)`. The dispatched expression `p*b - (1-p)*c` instead gives `p=c/(b+c)` and cannot yield the provided table. Research §03 at `982091d4...` already uses the correct equation.

| b | Break-even, c=0 | Break-even, c=0.04 |
| --- | --- | --- |
| 0.5R | 66.67% | 69.33% |
| 1R | 50.00% | 52.00% |
| 2R | 33.33% | 34.67% |
| 3R | 25.00% | 26.00% |

`0.5*0.5R + 0.5*3R = 1.75R`. These four examples and the weighted example are arithmetic, **n=0 observed trade outcomes**, not estimates of rates or fees.

[A] The production limit-placement call uses quantity **1** (`trader/armed_executor.go:1067`, also the shared placement seam `:2161`). A single contract cannot be split into half a contract. However, “one contract” must not be misread as a universal configured one-leg ceiling: the one running strategy row **a5b7662e-7bf7-49bb-9f09-7efa48f95ac8** stores `max_contracts_per_order=2`, `max_contracts_enabled=false`, `plan_mode=strict`, `min_risk_reward_ratio=2`. `armLegCapacity`/`splitLegCapacity` (`:731`, `:741`) read the positive capacity directly; they do not consult that enabled switch. London v2 S1 itself authors two one-contract legs. This is not an observed two-contract position and is not authorization to add size for partial exits. [Whitelisted configuration evidence](2026-09-08-scenario-economics-data/risk-settings.json). No account name or credential is retained.

## C6 — widths reproduced; composed-floor claim NOT REPRODUCED

[T/A] Row **270**, same stored ATR14(5m)=23.8431 and 1.5 multiplier:

| Scenario | Entry | Authored stop | Width | Stored-ATR floor |
| --- | --- | --- | --- | --- |
| S1 | 29546.25 | 29510 | 36.25 | 35.76465 |
| S2 | 29579.75 | 29543.75 | 36.00 | 35.76465 |
| S3 | 29646 | 29682.25 | 36.25 | 35.76465 |
| S4 | 29736.75 | 29773.25 | 36.50 | 35.76465 |

The differences above that floor are S1 **0.48535**, S2 **0.23535**, S3 **0.48535**, S4 **0.73535** points. **0/4 equal it exactly; 1/4 is within one tick** (S2). All four may have been authored with the floor in mind, but that causal claim is [C], not measured proof.

Across the complete geometries, every one has a retained indicator-block 5m ATR: **111/111 frozen; 177/177 all retained**. Authored risk equals 1.5×that stored ATR in **0/111 and 0/177**; within a tick, **13/111 and 22/177**. Matching IDs are the `authored_within_tick_stored_floor=true` rows in the scenario evidence. These are explicitly **authoring-snapshot comparisons**, not actual composed-stop counts. The stored ATR is printed to four decimals, so exact equality is additionally sensitive to that precision.

Actual composition reads fresh arm-seam ATR and picks the widest of authored stop, nearest seated risk-side anchor with clearance, and ATR floor (`trader/armed_executor.go:292`, `:429`; `trader/arm_stop_anchor.go:60`, `:132`). It never tightens the authored stop. The `armed_orders` schema does not persist that complete input/result tuple. The log emits composition only when widened or unanchored (`armed_executor.go:431`); silence is not an observed authored/floor result. No complete per-scenario composition replay was established here: **177/177 historical composed-stop outcomes remain unreconstructed in this census**, not zero floor-bound stops. No floor, arm-composition or executor edit is warranted by this report.

## E6 — projected refusal census and STOP

[A] Applying E2 literally to unchanged retained documents yields **799/799** missing-obstacle refusals, including **280/280** frozen members. Each exact identity is listed in `scenarios.json` with `first_obstacle_present=false`. The reason for each is the new missing required field; that reason is not an economic contradiction. This is a prospective structural projection, **not execution of a completed validator**, and it is enough to trigger E6's STOP before building it.

Separately, the D2-only numeric comparison identifies these **13/799** potential path contradictions (1.63% of all scenarios; **13/177=7.34%** of complete geometries). The original four are **4/280=1.43%** or **4/111=3.60%**. The chosen denominator must be stated; none licenses refusing UNKNOWN geometry. An eventual explicit exception contract would still need to be evaluated before calling them actual write refusals.

| Plan row / scenario | Date/session/version | Arm target | Path | Projected reason |
| --- | --- | --- | --- | --- |
| 142/S2 | 2026-08-27:ASIA:v9 | 29645 | 29677.86, 29652.97, 29643.5, 29591.5 | arm target absent within 0.25 |
| 142/S3 | 2026-08-27:ASIA:v9 | 29615 | 29623.48, 29619.5, 29591.5, 29576.5 | arm target absent within 0.25 |
| 143/S1 | 2026-08-27:ASIA:v10 | 29625 | 29624.65, 29619.5, 29591.5 | arm target absent within 0.25 |
| 143/S2 | 2026-08-27:ASIA:v10 | 29625 | 29624.65, 29591.5 | arm target absent within 0.25 |
| 143/S3 | 2026-08-27:ASIA:v10 | 29592 | 29591.5, 29581.88, 29576.5 | arm target absent within 0.25 |
| 144/S2 | 2026-08-27:ASIA:v11 | 29591 | 29591.5, 29576.5 | arm target absent within 0.25 |
| 144/S3 | 2026-08-27:ASIA:v11 | 29570 | 29576.5, 29432.25 | arm target absent within 0.25 |
| 152/S2 | 2026-08-28:LONDON:v6 | 29620 | 29644.38, 29577.75 | arm target absent within 0.25 |
| 156/S1 | 2026-08-28:NY:v4 | 29654 | 29707.5, 29657.39, 29642, 29577.75 | arm target absent within 0.25 |
| 178/S1 | 2026-09-01:LONDON:v2 | 29418.62 | 29447.5, 29435.75, 29422.12 | arm target absent within 0.25 |
| 194/S3 | 2026-09-01:ASIA:v7 | 29131.66 | 29099.85, 29113.23 | arm target absent within 0.25 |
| 230/S4 | 2026-09-02:ASIA:v14 | 29214.5 | 29235, 29228.75, 29218.25, 29212.5, 29207.5 | arm target absent within 0.25 |
| 259/S1 | 2026-09-06:ASIA:v2 | 29575.48 | 29545, 29566.02, 29587.75 | arm target absent within 0.25 |

There are no authored first-obstacle/R declarations to test for obstacle-beyond-target or stated-R contradictions. Those counts are **unevaluable**, not zero validated contradictions. No synthetic case is represented as a measured live instance. The schema's missing-field requirement and the unmeasured contradiction halves need a clear warn/refuse ruling before implementation.

The E3 example independently reproduces at **row 265/S2**, 2026-09-07 ASIA v2: entry 29664.50, stop 29640.00, arm target 29721.25, risk 24.50, arm R **2.3163265306**, first listed target 29671.42, R **0.2824489796**. It is a prospective WARN fixture, not a declared historical obstacle or an executed exit.

## Source footprint, tests, and recorder coordination

[A] Production before/after is identical. No E1–E9 implementation fixtures, RED/GREEN or mutations are claimed. The census script ran successfully and its frozen triple/membership cross-check passed. Full Go, goldens, Vitest, tsc and clean-clone binary builds are **not run for this evidence-only STOP**; there is no candidate binary to certify. No new production function/call site exists. Guide/SYSTEM-MAP/boot behavior stays unchanged because no behavior changed.

[A] At accept, `fix/stage-a-snapshot` was `250546854b249a89591befb4b05cd64ffd482b08`, with claim `af1ded7e` naming `stage-a-snapshot-96604090/root[unlisted]`. Its claim owns record-only schema/writers/export, record hooks, boot and related docs; it explicitly excludes the scenario contract. Its current diff contains its own evidence report/data only. This wave has changed none of its files or record hooks. This is branch/claim evidence of ownership, not attribution from Git author identity and not a claim of direct lane acknowledgement. On resume, fetch/rebase on whichever lands first and re-read the recorder footprint; do not edit `store/plan.go` persistence as if no other lane owned it.

[A] Checklist census used both numbering formats, numeric sort and **`uniq -c`**: highest occupied **91**, duplicates **75/76/77 each have count 2**. [Full census](2026-09-08-scenario-economics-data/checklist-census.txt). No new class is assigned for this report; assignment belongs to the behavior merge and must be rechecked then. No other lane's entry is renumbered.

## Owner-visible limits, rollback, and publication

A15: the card still renders its existing target path, without a declared obstacle/response or both economics R values. The desk has no new economics line. There is no scenario-economics boot line or counter. No first-new-contract scenario, contradiction refusal, sub-1R WARN/counter, or live card proof exists. Confirmation/liveness remain the previously booted combined revision. This wave has not fixed the surface yet.

Rollback: this wave changes documentation and evidence only; no runtime rollback, DB backup/migration or binary swap is needed. A later implementation must retain A31, version legacy handling explicitly, run production call-site pins RED→GREEN and actual mutations, update Guide/SYSTEM-MAP together, and obtain this wave's own GO before the normal merged-head/clean-clone/five-leg/RELEASE/swap/VERIFY/owner-kill procedure. The earlier combined-wave GO is not authorization for a scenario-economics boot.

The correction report is to be fast-forwarded to dev under the main-tree lock, with automatic heartbeat started at acquisition, and its raw URL verified at that exact committed SHA before lock release. That documentation-only merge does not require a boot and does not change RELEASE or GUIDE_BUILT_REV. The final chat receipt supplies the resulting immutable report URL, HTTP status and Git/download byte equality; this file cannot name its own future commit without a self-reference.
