# Is the Planner preparing like a professional NQ trading desk?

**Read-only review · 7 September 2026 · source and evidence frozen as specified below**

## 1. My verdict

I would use this Planner as a structured draft that needs review. I would not yet treat its output as a consistently coherent, independently validated pre-trade preparation process. It identifies recognizable market references and often builds intelligible scenarios. The principal deficiencies are the timing and consistency of those scenarios, the loss of information while filtering levels, and the inability to reconstruct exactly why a particular version was authored.

The three largest problems are:

1. **A plan can be old when first published.** Tonight, one request took 11.4 minutes; a rejected request plus repair took approximately 18.4 minutes. Market input is not fully rebuilt during repair, and there is no general final age/price-drift acceptance check. A fresh creation timestamp therefore does not mean fresh analysis.
2. **The confirmation contract is broken in reproducible cases.** The current evaluator can count a still-incomplete five-minute bucket as a five-minute close, and can accept a reclaim that occurred before the subsequent sweep. I reproduced both directions using the actual functions with synthetic bars and no broker connection.
3. **The ranking cannot currently justify itself.** Rejected candidates lose their scores; all 600 records have empty score components; the reported exclusion cause is inferred rather than recorded at the actual decision. The valid touch sample contains no unselected comparison group.

The three largest opportunities are:

1. Keep a complete market map, with each level's role and validity separate from whether it is selected as an entry setup.
2. Require every scenario to express an ordered, timestamped sequence and reconcile its entry, thesis invalidation, protective stop, first obstacle and final target.
3. Turn Planner history into an auditable sequence of input snapshots, attempts, transformations and accepted versions. Then compare competing preparation styles without hindsight.

**No production code, configuration, database rows, live prompts, orders, or deployments were changed.** All recommendations remain proposals. A pre-session review automation was created at the owner's explicit request; it only performs read-only review and reports actionable changes. The audit payload remains local pending publication approval.

## 2. What was actually examined

The running source revision is `5457ac5accd97c3519bf6d16ead147a0db2ab0d0`. It was checked through local health and reviewed in a separate, locked checkout. The report branch began at dev `d79cc7acb76b79e7564b8c541091216c3122311a`; its intervening changes in the audited source directories affect Desk facts and tests. Planner findings refer to the running revision, not whatever dev may subsequently contain.

Checklist provenance: `7ddf73cd99b9cd7fece7791c33a15bc56e33192e`, “docs(report+checklist): session calendar report and class 88 — a liveness signal that is a side effect of activity.” I applied the fresh-evidence, independent-math, long/short, source-reference and read-only requirements. Existing reports were background, not proof that a finding remains present.

| Evidence | Coverage | What this establishes |
|---|---:|---|
| Planner documents | 102 versions, 31 August–7 September | Stored plans, including 43 ASIA, 20 LONDON, 37 NY and 2 weekly versions |
| Date/session groups | 18 | Replans are repeated observations within a group, not 102 independent trading sessions |
| Scenario documents | 280 | Full inventory and structural comparison in the companion CSV |
| Complete arm geometry | 111 | Entry, stop and target arithmetic; the other 169 scenarios are not assigned invented arm prices |
| Planner read-facts | 42 | Retained floor, void and regime fields; none is bound to an accepted plan ID/version |
| Candidate records | 600 over 25 reads | A limited intermediate candidate pool, not every raw generated level |
| Touch records | 1,766 total; 124 valid | Validity-filtered evidence, with 87 resolved outcomes and 37 ambiguous outcomes |
| Current-session case studies | All three versions captured on 7 September | Document, input-time context where recoverable, publication-time price and scenario changes |
| Historical bars | 10,724 retained rows read across selected timeframes | Independent price-context checks; the report includes the current-session minute-bar subset |

Planner-history capture is **22:17 Chicago on 7 September**. Later changes are outside this frozen audit. Logical IDs are retained as `date:session:vN`, with row IDs where available; account/trader identifiers are removed from the deliverable.

`[T]` labels measurements from this run's bot data; `[R]` labels primary research or official sources; `[I]` labels analytical judgment requiring validation here. The professional-trader lens is an analytical standard, not a claim of personal trading experience. **PROVEN** refers to inspected behavior, **BROKEN** to a demonstrated contract defect, and **UNVERIFIED** to an event or benefit not established by this audit.

The source review follows the decision-relevant functions end to end. It is not a certification that every line in the entire repository is correct. Every scenario is inventoried; the three current-session versions receive detailed contextual review. Exact historical live-cache reconstruction for all 102 versions is unavailable, so I do not claim it.

## 3. The preparation pipeline and where information changes

| Stage | Current behavior / review point | Source |
|---|---|---|
| Market input | NT8 bars feed features, references and session context; persisted bars permit a separate arithmetic check | [store/bar_history.go:22](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/bar_history.go#L22) |
| Candidate selection | A broader pool is clustered, prioritized, balanced and capped before final grade/seating selection | [kernel/levels_score.go:535](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/levels_score.go#L535), [kernel/levels_score.go:601](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/levels_score.go#L601) |
| Candidate evidence | Input conversion loses scoring fields; selected scores are restored, excluded scores are not | [trader/auto_trader_planner.go:2341](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L2341), [kernel/detector_recorder.go:67](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/detector_recorder.go#L67) |
| Planner request | Input and rendered prompt are assembled before the model call | [trader/auto_trader_planner.go:895](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L895) |
| Retry / repair | Repair uses prior output, errors and instruction excerpts; reauthor uses the original full prompt | [trader/auto_trader_planner.go:1508](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L1508) |
| General validation | Validates against original facts; specialized checks can read newer bars | [trader/auto_trader_planner.go:1660](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L1660), [trader/auto_trader_planner.go:1698](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L1698) |
| Output normalization | Normalizes vocabulary and may collapse levels or remove inappropriate arm legs | [kernel/plan_doc.go:393](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_doc.go#L393), [kernel/plan_doc.go:472](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_doc.go#L472), [kernel/plan_doc.go:1210](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_doc.go#L1210) |
| Authoring provenance | Input facts use the AI configuration hash and omit accepted-plan linkage | [trader/auto_trader_planner.go:2447](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L2447) |
| Accepted version | Appends a new version and stores normalized document, indicator mirror and hashes | [trader/auto_trader_planner.go:1938](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L1938), [store/plan.go:249](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/plan.go#L249) |
| Scenario confirmation | Evaluates conditions since plan publication/birth, with the timing defects demonstrated below | [kernel/plan_confirm.go:49](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_confirm.go#L49), [kernel/plan_confirm.go:209](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_confirm.go#L209) |
| Scenario state | Derives status using inferred anchor and a generic danger-direction ladder | [kernel/scenario_state.go:96](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/scenario_state.go#L96), [kernel/scenario_state.go:194](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/scenario_state.go#L194) |
| Lifecycle | Writes dormant/active changes separately from the original authoring trigger | [store/plan.go:287](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/plan.go#L287) |
| Executor context | Resolves an active plan and overlays; separate armed-order text can describe retained older authorization | [trader/auto_trader_planner.go:2627](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L2627), [kernel/plan_render.go:156](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_render.go#L156), [trader/armed_executor.go:921](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/armed_executor.go#L921) |
| History display | Summarizes bias, level membership and scenario count; does not fully compare scenario semantics | [api/handler_plan.go:715](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/api/handler_plan.go#L715) |

The key separation is **what the model authored**, **what normalization retained**, **what the machine says has happened**, and **what the executor received**. A number visible in one stage cannot automatically be attributed to another.

## 4. Trading style and market context

### Is the style itself unreasonable?

No. The captured documents contain 116 `reject`, 80 `sweep_reclaim`, 34 `reclaim`, 29 `breakout_retest`, 17 `hold`, 2 `acceptance` and 2 `breakdown_continue` scenarios. This is principally a level-based, conditional preparation style with considerable fading and reversal logic. These are **authored scenario counts**, not trade frequencies or independent opportunities. [T: all-scenarios.csv]

A countertrend short can be coherent when a well-defined upward auction fails at resistance; a long bias does not prohibit every short fade at the opposite edge of a balanced range. What is unacceptable is calling the same observation both continuation and failed breakout without a stated precedence rule and different confirming evidence. Likewise, being above a moving midpoint alone is not proof that upside is exhausted. Those are conditional hypotheses to test, not universal market laws. [I]

### Current-session history, independently compared

| Version | Published, Chicago | Stated working context | Latest fully closed minute-bar close at publication |
|---|---|---|---:|
| 7 Sep ASIA v1 | 20:47:57 | Short/medium; failure around ONH 29687.50, price described at 29678.00 | 29672.00 |
| v2 | 22:03:44 | Long/medium; pullback longs plus upper-edge shorts; narrative describes roughly 29687 | 29665.25 |
| v3 | 22:16:50 | Short/low; balance and premium, narrative price 29665.75 | 29689.00 |

The input-related minute-bar checks corroborate v1's session high **29687.50** and low **29521.75**. Later data corroborate high **29721.25**, the same low and the **199.50-point** session range. This is positive evidence: these core price levels were not simply invented. [T: case-bars.json, planner-history.json]

Changing short → long → short is not, by itself, an error. The market and inputs changed. The problem is that the timing of those changes must be explicit. v3's narrative price is **23.25 points below** the last completed minute close by publication. Its short entry at 29675.22 is already behind that close, and the narrative invalidation is a five-minute close above that entry. This is a concrete need to re-evaluate context before calling the plan ready, not proof that a broker order was placed incorrectly. [T]

### Detailed scenario assessment

- **v1 S1, ONH sweep/reclaim short:** The idea is intelligible if an actual sweep precedes a completed close back under the level. It has no stored arm, so I cannot claim a precise proposed trade risk or call it executable under strict mode. It is a conditional idea until an admissible entry/stop policy is attached.
- **v1 S2, EQH rejection short:** The market can reject equal highs, but the system's own role warning says a liquidity fade needs sweep/reclaim, while the stored condition is plain rejection. Its 30-point stop lies before the textual 29750.62 close-based invalidation. An intentionally tighter protective stop is possible, but that distinction should be explicit. The first target offers only 0.667R versus 2.362R at the arm target.
- **v1 S3, break-and-hold short:** Waiting for sustained acceptance is coherent in principle. The narrative, ten-minute hold evaluator and invalidation need to agree on duration and on which completed bars count. It also lacks an arm in the captured document.
- **v2 S1/S2, pullback longs:** These are coherent hypotheses only if the market still supports continuation when the plan becomes available. S2's first target is only 6.92 points away against 24.50 points of risk. Calling the final target 2.316R is arithmetically correct but does not solve that near-obstacle problem.
- **v2 S3/S4, ONH rejection and sweep/reclaim shorts:** They share entry, stop and target but require different information. They must not become two interchangeable confirmations of the same event. S4 stores both sides as `below`; normalization does not explain that away. However, touch evaluation ignores the side field entirely, so this field mismatch alone does not prove reversed execution. The stronger proven defect is the missing temporal ordering.
- **v3 S1/S2, upper/lower value fades:** The two-sided balance concept is understandable. Their first targets offer roughly 0.513R and 0.609R while final targets exceed 2R. They require an explicit plan for the first obstacle and a fresh-context check; a distant mean target is not sufficient justification.
- **v3 S3, ONH sweep/reclaim short:** It has the clearest planned distance to the first target in this captured version, about 1.918R before costs. That makes its geometry comparatively less constrained, not more likely to win. It still depends on a real sweep, a later completed reclaim and an admissible retest.

I would retain these as hypotheses while correcting their contracts. I would not infer that fading is superior, or replace it with trend following, from this small and repeated version sample. [I]

## 5. Levels, ranking and information lost through filtering

<!-- LEVEL_INVENTORY_PENDING -->

### The grading history does not establish a calibrated ranking

All 600 candidate rows have `score_components = {}`. There are 299 selected rows and 301 excluded rows; every excluded row carries the same cap-related reason. Source inspection establishes that excluded scores/grades are not retained and that the cap reason is inferred from the final table being full. Therefore a zero is missing evidence, not a verified poor score. The reported threshold is the minimum selected score; the rank is a final table position, normally nearest-first—not a pure merit ranking. [T; [kernel/detector_recorder.go:50](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/detector_recorder.go#L50), [kernel/detector_recorder.go:67](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/detector_recorder.go#L67)]

The saved candidate pool begins **after** clustering and other selection, and is capped at twice the final capacity. It cannot reveal every raw level lost earlier. On the latest captured read, EQL·4h **29666.00** (row 577) and EQL·1h **29717.50** (row 584) are excluded; ONH **29721.25** (row 586) is retained. The EQL/ONH separation is 3.75 points, exceeding the three-point direct cluster tolerance; claiming that this pair was directly merged would be wrong. The erased scoring evidence prevents proving the exact cause of exclusion. [T; [kernel/levels_score.go:535](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/levels_score.go#L535), [kernel/levels_score.go:678](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/levels_score.go#L678)]

This answers the owner's central concern: **a level disappearing from the selected table does not establish that its market role has disappeared.** An excluded level may still matter as an obstacle, a target or invalidation reference. Conversely, its mere presence on a chart is not proof of predictive value. The proposed design keeps these two questions separate. [I]

### Output can discard information that input filtering preserved

Input clustering protects zones from some line-level merges. Output `CollapsePlanLevels` merges by price distance and model grade, without preserving canonical constituent identities or zone bounds. Scenario references are not rebuilt as a semantic graph during this collapse. A role-preserving input design can therefore lose information again after model generation. [PROVEN code behavior: [kernel/levels_score.go:715](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/levels_score.go#L715), [kernel/plan_doc.go:25](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_doc.go#L25), [kernel/plan_doc.go:393](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_doc.go#L393)]

### What the tape can and cannot say

The 1,766 touch rows include 423 legacy-unverified, 965 without formation provenance, 254 marked duplicate and 124 marked valid. Among the valid rows, 46 hold, 41 break and 37 are ambiguous. The valid sample spans two date groups, covers only selected levels and is confined to demand, supply and order-block families. There is no valid unselected control group. [T: touch-evidence.json, metrics.json]

The current rate implementation **does** filter to valid rows and excludes ambiguous observations from its hold/break denominator. I do not repeat the older accusation that current rates necessarily use every duplicate row. The problem now is inadequate comparative evidence. [[store/touch_outcomes.go:237](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/touch_outcomes.go#L237), [store/touch_outcomes.go:281](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/touch_outcomes.go#L281)]

No level family should be retired, promoted or reversed on these observations alone. A proper comparison records each candidate before selection, evaluates the same future horizon, preserves ambiguity, avoids duplicate events and groups uncertainty by session/day. [I]

## 6. Entry, stop, target and the first obstacle

All 111 complete stored arms have their stop and final target on the correct side of entry. That is a useful basic invariant, and the validator checks it. It does **not** establish that the trade makes economic sense. [T; [kernel/plan_doc.go:206](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_doc.go#L206)]

**45 of 111** complete arms put their first listed target at a positive distance below 1R: 40.5%, nominal Wilson 95% interval **31.9%–49.8%**. Versions share market history, so this nominal interval is descriptive and is not an independent-sample estimate of an edge. Four arm targets are absent from the scenario target chain even allowing half a tick. Fifty-four arms contain at least one non-tick-aligned analytical price; later transport rounding is a separate stage, so these are not proof of invalid broker orders. [T: all-scenarios.csv, metrics.json; [trader/ninjatrader/tcp_trader.go:447](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/ninjatrader/tcp_trader.go#L447)]

| Captured scenario | Entry / stop | Final target | Risk, points | First-target R | Final-target R |
|---|---|---:|---:|---:|---:|
| 7 Sep v1 S2 short | 29707.50 / 29737.50 | 29636.64 | 30.00 | 0.667 | 2.362 |
| v2 S2 long | 29664.50 / 29640.00 | 29721.25 | 24.50 | 0.282 | 2.316 |
| v2 S3 short | 29721.25 / 29746.50 | 29664.50 | 25.25 | 0.515 | 2.248 |
| v3 S1 short | 29675.22 / 29700.00 | 29621.97 | 24.78 | 0.513 | 2.149 |
| v3 S2 long | 29568.73 / 29544.75 | 29621.97 | 23.98 | 0.609 | 2.220 |
| v3 S3 short | 29721.25 / 29745.25 | 29646.00 | 24.00 | 1.918 | 3.135 |

These are authored values before tick rounding, fees, slippage or later stop composition. For one MNQ, $ risk before costs is points × $2; the v2 S2 plan therefore represents $49 before costs. CME specifies MNQ at $2 per index point and a 0.25-point minimum tick. [R: [CME MNQ specifications](https://www.cmegroup.com/markets/equities/nasdaq/micro-e-mini-nasdaq-100.contractSpecs.html)]

`target_chain` is checked for positive prices and proximity to original market price, but not for complete semantic ordering, directional agreement with entry or consistency with the arm target. The arm R:R check uses the separate arm target. [[kernel/plan_doc.go:646](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_doc.go#L646), [kernel/plan_doc.go:913](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_doc.go#L913), [trader/armed_executor.go:1905](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/armed_executor.go#L1905)]

**My proposed rule is not “every first target below 1R is forbidden.”** The first target might be a reference rather than a mandatory exit. The plan must say which it is and justify carrying through it. With one contract, “take partial profit at the first target” is not a feasible default. If realistic reward is inadequate, reject the setup or require a better entry mechanism; do not pull a structural stop inside invalidation merely to improve the displayed ratio. [I]

A complete scenario should distinguish:

- Entry location and evidence that activates it.
- Thesis invalidation, with price, side, timeframe and completed-bar requirement.
- Protective stop, including any deliberate difference from close-based thesis invalidation.
- First obstacle and the required response there.
- Final target and why the path beyond intervening levels remains plausible.
- Executable tick-rounded prices and net economics under stated cost assumptions.

## 7. Confirmation and scenario-state defects

### P1 — Five-minute confirmation before five minutes have closed: BROKEN, reproduced

`EvaluateConfirm` calls `AcceptanceRunEver`, which aggregates bars but does not filter the final bucket using the supplied current time. I supplied one completed minute bar inside a new five-minute interval. Both above and below variants returned `Met=true` for `1x5m_close`, although no five-minute candle had closed. [[kernel/plan_confirm.go:103](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_confirm.go#L103), [kernel/scenario_facts.go:440](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/scenario_facts.go#L440)]

### P2 — Reclaim before sweep accepted as sweep then reclaim: BROKEN, reproduced

For a two-part scenario, the second condition should begin when the first condition becomes true. `firstConfirmFireMs(touch)` returns zero; the caller falls back to plan birth. I supplied five minutes closed on the reclaim side, followed by the first touch/sweep, with the last bar still on the wrong side for a subsequent reclaim. Both long and short variants returned `Met=true`. [[kernel/plan_confirm.go:209](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_confirm.go#L209), [kernel/plan_confirm.go:238](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_confirm.go#L238)]

The [four probe results](confirmation-probe.json) and [inert source file](confirmation-probe.go.txt) are included. The probe calls the actual running-revision functions in an isolated process, not a duplicate implementation. It does not place orders or use live broker state. These counterexamples prove the evaluator defect; they do not measure how often it caused a historical fill.

### P3 — “Armed” and “invalidated” are not complete representations of the authored setup

The scenario status ladder calls a plan near its anchor “armed.” It evaluates acceptance in the direction dangerous to the trade before checking the condition-specific trigger. `acceptance` and `breakout_retest` themselves use `Accepted` for their trigger mapping, making that trigger branch unreachable in this ladder. Several other conditions have no mapping there. This is not a full representation of each authored setup. [[kernel/scenario_state.go:137](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/scenario_state.go#L137), [kernel/scenario_state.go:194](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/scenario_state.go#L194)]

Anchor selection scans the trigger text before invalidation text and takes the first matching level. It does not parse the complete authored invalidation rule. This heuristic also reaches a refusal path, so it is more than a cosmetic label concern. [[kernel/scenario_state.go:96](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/scenario_state.go#L96), [trader/invalidation_resolver.go:45](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/invalidation_resolver.go#L45), [trader/entry_gate.go:223](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/entry_gate.go#L223)]

Counterevidence matters: absent states can display unknown; inactive plans project expired. A level accepted through one direction can legitimately remain relevant in a different role. Therefore a `still_valid=false` level next to an “armed” scenario is a diagnostic signal, not standalone proof that a trade bypassed protection. [[web/src/components/plan/ScenarioList.tsx:350](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/web/src/components/plan/ScenarioList.tsx#L350), [api/handler_plan.go:2275](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/api/handler_plan.go#L2275)]

### P4 — Death reactivation reads obsolete provenance: PROVEN code mismatch; occurrence unverified

Lifecycle updates now preserve the original authoring trigger and append the dormancy reason to a separate log. The reactivation path still looks for `dormant:death:` inside the original trigger; otherwise it chooses the flip rule. A missing selected condition can clear immediately before the minimum-hold check. This can test the wrong predicate. An independent death gate remains, so this is not proof that a still-dead plan can place a trade. [[store/plan.go:287](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/plan.go#L287), [trader/auto_trader_planner.go:604](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L604), [trader/auto_trader_planner.go:571](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L571)]

Tonight's captured lifecycle events for v1 are **flip** dormancy at 21:19:35 and reactivation at 22:01:28; they do not demonstrate the death-only defect. No historical occurrence is invented.

## 8. Planner history, prompts and executor consistency

### Old inputs can be presented with a new publication time

The initial request and original facts survive retries. General validation uses those facts, while some specialized validators fetch newer bars. Thus the accepted document can combine old general context with newer specialized checks. The logged calls were 378.0 seconds for v1, 683.7 seconds plus 423.3 seconds for the rejected/repair sequence producing v2, and 683.4 seconds for v3. [T: planning-log.txt; [trader/auto_trader_planner.go:1508](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L1508), [trader/auto_trader_planner.go:1660](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L1660)]

Plan publication becomes the confirmation birth time. Events occurring during generation are therefore outside that subsequent confirmation window, although they may already have invalidated the underlying idea. The solution needs separate input time, publication time, scenario validity and a fresh-context decision; merely increasing model speed or resetting birth time is insufficient. [[store/plan.go:59](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/plan.go#L59), [trader/auto_trader_planner.go:2651](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L2651), [kernel/plan_lifecycle.go:156](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_lifecycle.go#L156)]

### The saved facts do not identify the accepted request

All 42 read-fact rows have empty plan IDs and version zero. The field called prompt hash is assigned the AI configuration hash. Repeated hashes therefore do not prove identical requests. Accepted versions preserve normalized documents and indicator mirrors, not a full, linked request/response/transformation chain. A repair also differs from the initial prompt whose hash is retained. [T; [trader/auto_trader_planner.go:2447](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L2447), [trader/auto_trader_planner.go:907](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L907), [trader/auto_trader_planner.go:1938](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L1938)]

Candidate versions name the latest stored version observed during assembly; the new version is allocated afterward. The latest candidate rows stamped v2 before v3 are therefore consistent with current implementation, not evidence that the database is one version accidentally behind. Blindly adding one to historical stamps would be wrong. [[trader/auto_trader_planner.go:2335](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L2335), [store/plan.go:249](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/plan.go#L249), [store/plan.go:396](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/plan.go#L396)]

### What the history screen omits

The current difference summary covers bias, conviction, day type, levels and scenario count. Same-count changes to direction, entry, stop, target, confirmation and invalidation are omitted. The full versions exist, but the summary is inadequate for the owner's requested comparison. [[api/handler_plan.go:715](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/api/handler_plan.go#L715)]

### Prompt obligations are not consistently prioritized

The prompt contains directional/conviction restrictions and later says the labels impose no obligation. It instructs a bias-aligned arm while also suggesting an AI-entry fallback for an infeasible arm, even though strict mode refuses that route. Current warning-only policies may be deliberate owner decisions; the report does not silently turn them into hard gates. The prompt should state the actual mode and resolved capabilities clearly. [[kernel/planner_prompt.go:157](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/planner_prompt.go#L157), [kernel/planner_prompt.go:612](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/planner_prompt.go#L612), [kernel/planner_prompt.go:730](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/planner_prompt.go#L730), [trader/entry_gate.go:184](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/entry_gate.go#L184), [trader/auto_trader_planner.go:1691](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_planner.go#L1691)]

The FVG template says its midpoint is computed rather than authored; the validator compares the supplied midpoint without first assigning the omitted value. Existing positive examples supply it. This is an additional contract inconsistency, not a reason to switch on FVG trading. Historical frequency is unverified. [[kernel/planner_prompt.go:697](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/planner_prompt.go#L697), [kernel/fvg_entry.go:268](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/fvg_entry.go#L268)]

### Executor linkage

The executor resolves the active plan and overlays, but its plan narrative and armed-order geometry are separately rendered. Working arms deliberately preserve original authorization and may belong to an older version. That is defensible only if the prompt and interface identify both versions explicitly. [[kernel/plan_render.go:156](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/plan_render.go#L156), [trader/armed_executor.go:921](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/armed_executor.go#L921), [store/armed_orders.go:247](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/store/armed_orders.go#L247)]

There is also a source-level race opportunity: prompt construction resolves a plan before the AI call, while decision attribution resolves latest again afterward. A replan between these stages can stamp a different version from the one read. This audit establishes the opportunity, not a specific historical occurrence. [[kernel/engine_analysis.go:468](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/kernel/engine_analysis.go#L468), [trader/auto_trader_loop.go:595](https://github.com/johnwick2921-cyber/nofx/blob/5457ac5accd97c3519bf6d16ead147a0db2ab0d0/trader/auto_trader_loop.go#L595)]

## 9. Research: what supports the review, and what it does not prove

- **[R] CME's trade-plan framework** includes instrument specification, planned entry/stop/target, risk, trading times and important economic events. It supports evaluating preparation as a complete decision process. It does not endorse this bot's ranking weights, a mandatory 2R target or a particular confirmation rule. [CME trade-plan worksheet](https://www.cmegroup.com/education/courses/files/download-trade-plan.pdf)
- **[R] Osler (2000)** studied published support/resistance forecasts in foreign exchange. Some levels predicted intraday interruptions, with meaningful differences across providers and currencies; relative strength rankings were not reliably established. This motivates testing level identification separately from level ranking. It is not evidence that an ONH fade or a particular MNQ grading term works. [Federal Reserve Bank of New York paper](https://www.newyorkfed.org/medialibrary/media/research/epr/00v06n2/0007osle.pdf)
- **[R] Lo, Mamaysky and Wang (2000)** found some technical patterns added information in historical daily US stock data. The available NBER abstract supports conditional, testable pattern definitions, not transferring results directly to intraday MNQ. Full-text retrieval was unavailable in this run. [NBER paper record](https://www.nber.org/papers/w7613)
- **[R] Bailey and coauthors** explain why selecting strategies from many backtests can produce overfit winners. A ranking comparison needs a held-out evaluation and an accounting of variants tried; repeated replans on the same session cannot masquerade as independent validation. [Author-hosted paper](https://www.davidhbailey.com/dhbpapers/backtest-prob.pdf)
- **[R] CQG documentation** distinguishes time-at-price/TPO profile information from volume displays. A bar-derived approximation needs its own label and error checks before it is treated as precise traded volume at each price. [CQG Market Profile documentation](https://help.cqg.com/cqgic/25/Documents/marketprofilemp.htm)

The captured plans include static fallback blackout entries for multiple macro releases. They are not evidence those releases occur on this session's date. The official September FOMC meeting is 15–16 September; BLS lists the August CPI release for 11 September. A preparation report must distinguish an actual dated event from a conservative fallback window and report calendar freshness. [Federal Reserve calendar](https://www.federalreserve.gov/monetarypolicy/fomccalendars.htm), [BLS CPI](https://www.bls.gov/cpi/)

I found no basis here for calling any style or level family a proven MNQ edge. That is a limitation of the evidence, not a conclusion that every level is useless.

## 10. Proposed changes, ordered by importance — none applied

| Priority | Proposal | Why / evidence | Work needed after approval | Acceptance evidence and metric |
|---|---|---|---|---|
| 1 | Correct completed-bar and ordered confirmation semantics | BROKEN in all four synthetic probes; section 7 | Code and shared scenario contract | No incomplete five-minute confirmation; no reclaim-before-sweep acceptance, both directions |
| 2 | Separate input age from publication and revalidate before a plan becomes usable | Old input plus long calls; sections 4 and 8 | Code, explicit age/drift policy, data | Delayed author/repair across a break, invalidation and session boundary cannot silently activate obsolete prep |
| 3 | Preserve complete level map and roles through every filter | Exclusion is not invalidation; input/output collapse differ | Data model, selection ledger, prompt presentation | Every generated level has stage history, identity, role and validity; excluded entry levels remain available as applicable obstacles |
| 4 | Retain real scores and true exclusion reasons | 600 empty component records; inferred cap reasons | Recorder/schema instrumentation | Independent score reconstruction; grade/distance/cluster/cap decisions separately reproducible |
| 5 | Make scenarios economically explicit | 45/111 first targets below 1R; distinct arm/chain fields | Scenario schema, prompt and validation | First obstacle, final target, cost assumptions, stop and thesis invalidation reconcile for each admitted setup |
| 6 | Define style-specific states and structured invalidation | Generic status ladder contradicts some conditions | State model and shared consumers | Every supported condition tested in both directions; version changes cannot retain unrelated S-number states |
| 7 | Link every read/attempt to its accepted version | 42 unbound fact rows; candidate stamp ambiguity | Provenance records | One accepted version resolves its input, attempts, transformations and exact executor prompt version |
| 8 | Remove contradictory prompt fallbacks and obligations | Strict-mode fallback and bias precedence conflict | Prompt proposal plus mode/capability rendering | No unavailable action is offered; balance/trend and bias constraints have explicit precedence |
| 9 | Make Planner history compare actual scenario changes | Current diff omits same-count semantic changes | Read-only history presentation | Direction, trigger, confirmation, stop, target and invalidation changes are visible with input timestamps |
| 10 | Repair death-reactivation provenance and test the FVG authoring contract | Source defects, historical occurrence unverified | Narrow, separate fixes only after review | Correct predicate clears dormancy; advertised FVG payload passes derivation/validation in mirrored fixtures |

The two removals I recommend are **removing a fallback that the current mode forbids**, and **removing misleading claims that a missing score is zero quality or that a full table proves capacity caused every exclusion**. I do not recommend deleting level families, tightening stops, raising contract size or enabling shadow setups based on this audit. [I]

### How I would structure pre-trade preparation

1. Establish the exact instrument, session, input time, completeness and dated events. A unavailable or outdated input is visible as such.
2. Build the complete higher-timeframe and session map. Mark each level's role, validity and source; prioritize display separately.
3. Describe the auction: directional acceptance, failed expansion, balance or uncertainty. Record the evidence that would change that interpretation.
4. Produce mutually distinguishable conditional scenarios, including a valid stand-aside alternative. Avoid mandatory trading direction when no feasible setup exists.
5. For each scenario, define ordered evidence, entry location, thesis invalidation, protective stop, first obstacle, final target and expiry.
6. Recheck the plan against fresh context before use and bind the executor to the exact version and resolved capabilities.
7. Review subsequent versions against what was known at their creation, not against the eventual winning move.

This is a proposed preparation process, not a new live prompt or an instruction to trade. [I]

### Validation before any strategy change

First correct the contract failures with narrow fixtures and replayed event sequences. Then collect immutable input/output records. Compare alternative selection policies using the same candidate universe and time windows, with events grouped by session/day. Predefine treatment of ambiguous intrabar ordering, missing bars and regime labels. Separate level relevance, scenario feasibility, confirmation correctness and eventual outcome. Use costs and simulation limitations only when evaluating executable strategy economics.

Do not tune against the entire observed history and then report that history as an independent test. Do not add one to candidate version stamps to manufacture lineage. Do not reconstruct missing original prompts with today's renderer and call them historical originals. [I; R: Bailey paper above]

### Scheduled review

The saved heartbeat runs at **01:45 London**, **08:15 New York**, and **16:45 ASIA**, America/Chicago. London and New York apply Monday–Friday; ASIA applies Sunday–Thursday. The single heartbeat skips nonapplicable weekend occurrences and checks market holidays/reopenings. Its instructions prioritize Planner history and full preparation quality, prohibit automatic fixes or orders, and notify on meaningful changes or failures. Saving the schedule is verified; successful future unattended runs and data access have not yet been observed.

### Remaining limitations

- No complete input/response/transformation archive exists for every accepted plan. Exact line-by-line historical request reconstruction is therefore unverified.
- Candidate telemetry omits earlier generation/filter stages and useful rejected scores. The best ranking policy cannot be determined from the present store.
- Valid touch observations span only two date groups and omit unselected controls; no profitability conclusion or live-money approval is made.
- Persisted bars are not necessarily the exact historical live-cache revisions; no visual inspection of the Windows chart is claimed.
- No fixes were implemented. The additional proposed tests have not been run unless explicitly identified as the four executed confirmation probes.
- “codex 101” was located in VS Code metadata, but the desktop messaging channel could not access it. No message was falsely reported as delivered to that task.
