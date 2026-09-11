# NOFX — New York preparation review, 11 September 2026

> **Historical snapshot, not a live dashboard:** observations end at 08:24 CT on 2026-09-11 at revision `dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7`. Publication is based on dev `616b52a9def4042ce40308deba46529723ba01d7`, which contains later updates. Findings have not been re-audited against those later revisions.

**Observation window: 08:16–08:24 America/Chicago (CDT). Scheduled review: 08:15, before the configured 08:30 NY session. Evidence closes at 08:24:00.**

**Verdict:** the newest NY plan is not yet a coherent executable preparation sheet under the policy currently running. Both enabled scenarios advertise a final target that clears 2R, but the running one-setup policy substitutes the declared first obstacle; that produces **1.682R and 0.522R**, respectively. The map's principal daily anchors are reproducible. The issues are the relationship between the map, the trading hypothesis, freshness and actual entry/target policy, rather than evidence that every level is wrong.

The review was read-only. No trading code, configuration, prompts, database records, orders or deployments were changed. This publication adds documentation only. It does not authorize an entry or a policy change. A preparation critique cannot establish expectancy from one plan.

Evidence labels: **[A]** directly read or independently calculated; **[B]** interpretation of those facts; **[I]** proposed policy/research, not an established trading edge. This is the public evidence edition. Trader/account identifiers are replaced by stable aliases; timestamps, row IDs, market prices and calculation inputs are retained. Full executor system/input prompts and unrelated plan fields are omitted; the original Planner prompt and the cited decision output fields remain available. All source links below are pinned to the running revision **dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7**; the older audit worktree's checked-out files were not treated as current implementation.

## 1. What was actually reviewed

- NY plan **rowid 316, version 1**, created **08:16:32.529 CT**, prompt hash **23fb62c19af0**, trigger `NY_scheduled_read`, lifecycle `active`.
- Plan ID: `2026-09-11:NY:TRADER_A`.
- Five versions of today's London plan, particularly v5, rowid 315, published **08:13:24.196 CT**. London remains the current session before 08:30; seeing London in the executor before then is not evidence of a wrong-plan handoff.
- NY read facts **101**, recorded **08:00:56.482 CT**; London facts **100**, recorded **08:00:51.493 CT**. Both carry ATR5m **38.2487864**, stop floor **57.3731796**, and the same newest 1m open at 08:00. These facts rows have empty plan IDs and version zero, so their linkage is by session/time/read sequence, not a falsely claimed database foreign key.
- NY candidate rows **1969–1992**, 24 total, **12 seated / 12 excluded**. The full exclusion reasons and score components are preserved.
- NY first-attempt rejection **180** includes the actual original prompt. It was rejected at 08:15:47 for mismatched composite labels on RTH-H and PDL. The completed v1 corrects those labels. The first prompt is not misrepresented as a complete recording of every repair message.
- Executor records **40015–40019** link to **London v5** and return `wait`. There is no NY executor-cycle claim in this pre-open report.

Evidence: [plans](evidence/plans.json), [read facts](evidence/facts.json), [NY prompt](evidence/ny-prompt.txt), [first-attempt rejection](evidence/rejected-ny.json), [candidate pool](evidence/candidate-pool.json), [executor history](evidence/decisions.json), [final snapshot](evidence/closeout.json).

## 2. Market and clock context

[A] BLS schedules today's CPI release at **08:30 Eastern / 07:30 Chicago**. Thus this is preparation after the scheduled CPI time, not preparation ahead of an imminent 08:30 CT CPI release. The local NY plan correctly carries 07:15–07:45 news windows. The configured NY opening exclusion is **08:30–08:35 CT**. [BLS September calendar](https://www.bls.gov/schedule/2026/09_sched.htm).

[A] CME's normal Micro E-mini schedule covers this Friday morning; the listed September Labor Day holiday period is earlier in the week. This review did not reinterpret a maintenance break as today's NY session being closed. MNQ is **$2 per index point**, tick **0.25 points / $0.50**. The risk figures below use one MNQ, not one NQ. [CME hours](https://www.cmegroup.com/trading-hours.html), [CME Micro E-mini contract specifications](https://www.cmegroup.com/trading/equity-index/files/cme-micro-e-mini-futures-fact-card.pdf).

[A] Strategy session overrides explicitly enable London and Asia even though their registry flags are false. NY is enabled. The effective NY strategy uses strict plan mode, minimum scenario grade B, last entry 13:00 CT and flat 14:45 CT. The current resolver checks explicit per-session enablement before the registry. [Enablement resolver](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/trader/auto_trader_planconfig.go#L104), [saved day-plan settings](evidence/day-plan-config.json), [registry](evidence/session-registry.json).

[A] The stored 07:30 five-minute bar spans **29204.50–29397.50**, a **193-point** range, with volume **38,890**. Between the 08:00 read and publication, the stored 1m series reached **29500.75** and later printed **29438.00** at 08:15. This is substantial post-news movement. It supports reassessing the read before activation; it does not prove that the remainder of the day must trend or reverse. These are NT8/Tradovate bar observations, not an independent exchange tick replay. [Bars](evidence/bars-cpi.json).

## 3. Each NY scenario: hypothesis, risk and executable target

| NY v1 | S1 — long pullback | S2 — short fade |
|---|---:|---:|
| Condition / confirmation | reject / touch | reject / touch |
| Entry | 29326.09 | 29508.25 |
| Authored stop | 29262.00 | 29566.00 |
| Authored risk, points | 64.09 | 57.75 |
| One MNQ gross stop risk | $128.18 | $115.50 |
| Declared first obstacle | 29433.88 | 29478.12 |
| Distance to obstacle | 107.79 | 30.13 |
| R to first obstacle | **1.68185** | **0.52173** |
| Authored final arm target | 29476.25 | 29326.09 |
| R to final arm target | 2.34296 | 3.15429 |
| First-obstacle response in plan | pass_through | pass_through |
| Meets current 2R floor at first obstacle? | **No** | **No** |

[A] These are authored/policy calculations for **n=2 scenarios**, not realized performance. Fees, slippage, partial fills and broker tick normalization are not included. `enabled:true` is a proposal field, not evidence that an order is working. At the closing snapshot there were no working broker orders and no open recorded position.

### 3.1 The final-target story and running policy disagree

[A] The **07:46:52 boot line** explicitly says `one setup: ON`, `target=first-obstacle`, `permission-required=yes`. The bound strategy has no override disabling this mode. Its min-R:R value is **2**, saved under **ai_config.risk_control.min_risk_reward_ratio**. The null top-level `risk_control` is not the loaded value. [Boot evidence](evidence/boot-policy-lines.txt), [selected risk fields](evidence/risk-fields.json), [one-setup resolver](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/store/resolve_source.go#L63).

[A] In the running arm path, the stop is composed first, then an allowed one-setup scenario's target is replaced by `first_obstacle.price`, before the gate examines R:R. The helper checks whether the obstacle is a valid price on the profit side; it **does not inspect `response:pass_through` or `target_path_exception`**. A narrative exception therefore does not preserve the final target in this path. [Target replacement](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/trader/armed_executor.go#L508), [obstacle helper](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/trader/one_setup_wiring.go#L178), [R:R check](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/trader/armed_executor.go#L2055).

[A] At the saved read ATR, neither stop needs widening: S1's authored stop is wider than both its nearest seated risk-side anchor and the 1.5×ATR floor; S2's authored distance exceeds the ATR floor. Later widening can only lower these R values because the composer does not tighten an authored stop. **If selection and confirmation allow either unchanged scenario to reach this gate, it fails the 2R test.** This is a deterministic conditional result from the current code and saved plan, **not a claim that an actual NY refusal or attempted fill was observed before NY activation**. [Stop composition](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/trader/arm_stop_anchor.go#L71).

[B] S1 is a comprehensible post-rally pullback idea; S2 is a counter-rally first-touch fade. They can be separate economic hypotheses. But neither can be sold as an executable 2R setup under the current first-obstacle contract. Keeping a distant target in the prose cannot make the nearer obstacle disappear.

[I] Recommendation for the preparation document: display authored final-target R and the policy-selected target R together; mark feasibility under the current policy explicitly. If an obstacle is believed consumed, identify the evidence and distinguish a proposed pass-through policy from the first-obstacle policy actually running. Do not compress the stop or move the target merely to make a ratio pass. No threshold or bot changes are proposed in this report.

### 3.2 S1 also conflicts with the prompt's own range rule

[A] The prompt says longs only below the dealing-range midpoint. It provides range **29038.00–29508.25**; midpoint is **29273.125**. S1's long entry **29326.09** is **52.965 points above that midpoint**, at **61.263%** of the range. Waiting for a pullback from the current 92%-of-range price does not bring this entry into the prescribed lower half. See prompt lines 244–251 and the S1 arm in the plan.

[B] This is internal rule disagreement, not proof that buying above the midpoint loses money. The 50% rule itself has no demonstrated expectancy in this review. A momentum pullback can have different premises from a balance-day discount buy; the document must say which is being proposed and why that class is permitted.

[I] Recommendation: make each scenario declare its playbook, applicable context and explicit exception if any. Compare the intended premium/discount policy against the **entry price**, not merely the price when the prompt was written. Research any proposed exception separately; do not infer validity from today's eventual P&L.

## 4. Freshness and preparation quality

[A] NY read facts were written at 08:00:56; v1 was published at 08:16:32 — about **15m36s** later. The selected ONH at read time was **29476.25**, while bars during authoring reached **29500.75**, **24.50 points higher**, before publication. S1 still targets the old 29476.25 value. The new high does not automatically invalidate that older price as a historical reference; it means it cannot be called the current overnight high without a time label.

[A] The 29503.38 four-hour OB mentioned in both London and NY is present in the prompt's HTF zone section (line 198). It is **not an invented number simply because it is absent from the 12-row ranked table**. The shortlist and HTF confluence inventory are distinct inputs.

[A] London v5 and NY v1 were authored from nearly simultaneous read facts, yet one labels the bias short/low and the other long/low. Their common fade is at PDH, but they disagree about what the post-CPI move represents. Session-specific hypotheses can differ legitimately; the report does not call different session plans a wrong-plan execution bug.

[B] The practical preparation gap is an explicit transition: what evidence establishes post-news continuation, what establishes failed continuation, and what permits the opposite-side first-touch fade? NY calls the move a trend day up, yet uses an undated historical win-rate claim to support the counter-short. London uses overbought RSI and supply to favor fading. Neither label should substitute for a testable context definition.

[I] Recommendation: assess validity at publication and again at activation, using changed anchors, completed bars, volatility and scheduled-event status. Keep the original read timestamp and old references visible; record whether changed conditions preserve, amend or invalidate each hypothesis. A particular re-read interval is not proven here.

## 5. Confirmations, invalidation and what the executor actually knows

[A] NY improves on the older ambiguous wording: its two reject triggers explicitly say **first touch**, and both carry `confirm.rule=touch`. The current entry-law table permits touch only for reject. A touch-entry and a confirmed rejection are different trades; this report does not quietly substitute one for the other. [Entry law](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/kernel/entry_law.go#L38), [touch evaluation](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/kernel/confirmation_evidence.go#L27).

[A] London v5 S1 still says a tag followed by a 1m/5m close back below, while its machine confirmation is touch. Executor decisions 40015–40019 repeat the close-back interpretation and wait. Before the PDH touch, both rules can agree on waiting; that agreement **does not prove** they would agree after a touch. An RSI clause in prose is likewise not shown to be part of the touch predicate.

[A] At NY publication, **both scenario invalidations (2/2)** were logged as **UNKNOWN and accepted for that check**, events **44460–44461**. S1 adds an RTH-H annotation and explanatory suffix; S2 uses prose outside the narrow accepted grammar. The exact parser is deliberately constrained, and the write-time validator explicitly accepts UNKNOWN. This is a proven limit on the authoring liveness check, **not proof that broker protective stops are missing or that every other gate passes**. [Parser](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/kernel/plan_authored_invalidation.go#L16), [write-time handling](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/trader/plan_liveness.go#L21), [events](evidence/logs.json).

[A] The separate running scenario evaluator uses an anchor/acceptance approximation; it is not the same as parsing the authored invalidation sentence. Its current status must not be reported as verification of every word of that sentence. [Scenario evaluator](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/kernel/scenario_state.go#L169).

[I] Recommendation: preparation should show the machine-supported invalidation separately from narrative explanation, and display UNKNOWN as UNKNOWN. Validate touch and confirmed-entry variants in separate samples with their own fills, stops, costs and missed trades; do not proclaim close confirmation superior without that comparison.

## 6. Level construction and ranking — verified positives and unresolved questions

[A] Independent queries against current stored 1m bars reproduce five named anchors exactly under the source's declared windows:

| Reference | Plan | Independent result | n / window |
|---|---:|---:|---|
| PDH | 29508.25 | 29508.25 | 1,380 bars, prior CT calendar day |
| PDL | 29038.00 | 29038.00 | same 1,380 |
| PDC | 29124.25 | 29124.25 | final close in that calendar day |
| RTH-H | 29275.25 | 29275.25 | 375 bars, configured NY 08:30–14:45 |
| RTH-L | 29057.75 | 29057.75 | same 375 |

The prior **Globex** session instead gives low 29042.50 and final 1m close 29142.25. That difference is explained by the current code defining PDH/PDL/PDC by **calendar day**, not by the CME trade date. The configured RTH window also ends at 14:45, not the full cash-close boundary. These definitions deserve visible labels; they are not silently interchangeable. [Extractor](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/kernel/levels_multiday.go#L41), [reproduction](evidence/anchor-reproduction.json).

[A] Session HLC3-volume VWAP+1σ recomputed from **900 overnight bars** is **29326.0752**, versus selected **29326.0892**: difference **0.0140 points**, below one MNQ tick. This is close agreement, **not byte-exact reconstruction of the original in-memory snapshot**. The current persisted bars and recorded prompt are separate observations. The code uses volume-weighted dispersion of bar typical prices, not exchange tick-volume-at-price. [VWAP formula](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/kernel/levels_volume.go#L67), [input bars](evidence/bars-overnight-read.json).

[A] This read preserves score components for kept and excluded levels — an improvement over earlier records without those components. However, **one 24-row candidate read cannot establish that the score predicts reaction or trading profit**. Grade is a selection output, not a probability. The filter excludes both nearby 29312.625 demand/OB rows due to a Tier-1 proximity grade cap, while far lower levels consume seats. Those omitted rows may still help explain structural stop placement for a long at 29326.09; their exclusion is not proof they have ceased to exist or are profitable entries.

[I] Recommendation: retain distinct purposes for entry eligibility, structural invalidation, first opposition and background context. Compare full candidate populations at matched decision times and regimes, keeping rejected candidates and outcomes with no hindsight selection. Do not promote the confluence multipliers to an established edge because they are now logged.

### Complete retained/excluded candidate inventory

| Row | Price | Label | Grade | Seated? / reason |
|---|---:|---|---|---|
| 1969 | 29476.25000 | ONH | A | Yes, rank 1 |
| 1970 | 29478.12500 | OB(bear)·1h | A | Yes, rank 2 |
| 1971 | 29439.62500 | OB(bear)·4h | A | Yes, rank 3 |
| 1972 | 29433.87500 | Supply·4h | A | Yes, rank 4 |
| 1973 | 29508.25000 | PDH | A | Yes, rank 5 |
| 1974 | 29585.87500 | OB(bear)·1h | C | OB 29585.88 [OB(bear)·1h]: minimum grade: C below B |
| 1975 | 29601.50000 | OB(bear)·1h | C | OB 29601.50 [OB(bear)·1h]: minimum grade: C below B |
| 1976 | 29326.08922 | VWAP+1σ | A | Yes, rank 6 |
| 1977 | 29312.62500 | OB(bull)·1h | C | OB 29312.62 [OB(bull)·1h]: minimum grade: C below B |
| 1978 | 29312.62500 | Demand·1h | C | DEMAND 29312.62 [Demand·1h]: minimum grade: C below B |
| 1979 | 29275.25000 | RTH-H | B | Yes, rank 7 |
| 1980 | 29216.52576 | VWAP | A | max_levels: 12 seated, cap 12 |
| 1981 | 29161.12500 | OB(bull)·1h | A | max_levels: 12 seated, cap 12 |
| 1982 | 29124.25000 | PDC | A | Yes, rank 8 |
| 1983 | 29106.96231 | VWAP−1σ | A | max_levels: 12 seated, cap 12 |
| 1984 | 29852.37500 | Demand·4h | C | DEMAND 29852.38 [Demand·4h]: minimum grade: C below B |
| 1985 | 29863.50000 | OB(bull)·4h | C | OB 29863.50 [OB(bull)·4h]: minimum grade: C below B |
| 1986 | 29870.62500 | OB(bull)·4h | C | OB 29870.62 [OB(bull)·4h]: minimum grade: C below B |
| 1987 | 29875.12500 | OB(bull)·1h | C | OB 29875.12 [OB(bull)·1h]: minimum grade: C below B |
| 1988 | 29057.75000 | RTH-L | A | Yes, rank 9 |
| 1989 | 29038.00000 | PDL | A | Yes, rank 10 |
| 1990 | 29029.87500 | Supply·4h | A | Yes, rank 11 |
| 1991 | 29922.37500 | OB(bear)·4h | C | OB 29922.38 [OB(bear)·4h]: minimum grade: C below B |
| 1992 | 29004.50000 | iFVG(bear)·4h | A | Yes, rank 12 |

The published NY document retains ten level rows and a separate eleven-row identity map, while the machine shortlist seated twelve. The HTF row at 29503.38 comes from the separate zone block. These are three representations; their counts should not be treated as one unexplained missing-data number.

## 7. The historical claim still biases today's reasoning

[A] The actual NY prompt repeats **75% win and +665 “this week”** for reject, without a sample count or a dated population. The source still embeds that text, and today's NY reasoning explicitly appeals to the “75%-win class” to justify the short. The audit checklist already names this exact small-sample failure class. This review does not allege the original arithmetic was fabricated; it establishes that a dated historical sample is being presented as current guidance without its denominator or transfer limits. [Prompt source](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/kernel/planner_prompt.go#L789), [actual prompt line 283](evidence/ny-prompt.txt), [audit checklist](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/docs/superpowers/AUDIT-CHECKLIST.md#L123).

[I] Recommendation: historical statistics used in preparation need their date range, n, population definition, realized-versus-authored distinction, costs and uncertainty. A small old reject sample cannot by itself justify fading a post-CPI trend. Use it as a research hypothesis, with separate evaluation by regime and entry mechanism.

## 8. History and handoff, without confusing sessions

| Session/version | Published CT | Main bias | Scenario families | Observation |
|---|---|---|---|---|
| London v1 | 01:36:07 | long / low | 2 reject longs | opening read |
| London v2 | 02:09:08 | long / low | reject + breakout-retest | one enabled reject arm |
| London v3 | 02:49:59 | long / low | reclaim + reject | updated long anchors |
| London v4 | 03:29:27 | long / low | 2 reject longs + reject short | became dormant at 08:05:01 |
| London v5 | 08:13:24 | short / low | reject short + unarmed reclaim long | current pre-08:30 executor plan |
| NY v1 | 08:16:32 | long / low | reject long + reject short | next-session plan reviewed here |

[A] Lifecycle [event 32](evidence/lifecycle.json) marks London v4 dormant on a buffered flip condition at 08:05:01. Executor records 40012–40014 waited with no active plan during the transition; 40015 onward identify London v5. This is evidence of version-specific handling, not a claim that the full NY handoff has already been tested. NY activation occurs after this review's cutoff.

## 9. Operational boundaries and next verification

[A] Closing health returned revision `dd1e2f0f0ae1`, unchanged through the observation window. The last stored MNQ 1m bar opened **08:23**, source `live`, contract `MNQ 09-26`. The last stored 5m bar opened 08:15, a completed bucket; the 08:20 bucket was not yet complete at 08:24. No 1m timestamp gaps were found in the inspected sequence since 07:30. Broker snapshot **19688**, received about 29 seconds before the closing observation, reports **zero working orders** for SIM_ACCOUNT_A. Open-position rows are zero; all 116 arm rows are terminal (83 cancelled, 23 filled, 10 superseded). There is no outstanding arm demonstrated in this snapshot. This is not a screenshot of NinjaTrader or a separately retrieved broker position frame.

[A] These observations support “data is arriving and no working orders were reported,” not a blanket guarantee of system health, full bar accuracy or risk protection during a future position. Dark-regime count remains 2 in the plan. No performance return, fill quality, bracket behavior or edge was inferred from this flat snapshot.

[I] Before treating the NY plan as trade-ready, the next checks are:

1. At/after the NY boundary, verify executor records identify NY v1 (or a specifically documented newer version), rather than infer the handoff from a visible card.
2. Resolve the preparation contradiction between `pass_through` prose and the first-obstacle execution contract. Preserve the existing refusal while that contract is unchanged; do not lower the R gate just to obtain a trade.
3. Revalidate the entry hypotheses using the new overnight high and current post-news conditions. If a level is retained from the read snapshot, mark it historical.
4. Make the advertised premium/discount exception and invalidation support explicit before calling the plan internally consistent.
5. In a future authorized replay/SIM study, compare touch versus confirmed-entry with separate fills and costs, and compare target policies using the same eligible events. Today's n=2 geometry is not that study.

**PR #99 follow-up (historical, at review time):** checked before the trading review. The exact authorized head and dev base are unchanged; nine individual CI jobs still fail. No merge or bypass was performed. This records the check performed during the review, not the PR status at publication. Its conditional merge remains a separate task.

## 10. Reproducibility and limits

The audit checklist at the running revision was consulted for timestamp conventions, bound-strategy selection, sample IDs, risk ratios, small samples and version attribution. This is a scheduled preparation review of the named plans, not a new line-by-line audit of every subsystem. The successful plan record, original rejected Planner prompt, selected decision outputs and logs, candidates and arithmetic inputs are bundled alongside this Markdown report. Code is referenced at its published commit rather than copied into the documentation tree.

- Code claims are pinned to dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7 and available in the [published arm-path source](https://github.com/johnwick2921-cyber/nofx/blob/dd1e2f0f0ae11834ca3b1c6ff79fe8055a2284f7/trader/armed_executor.go).
- Market/calendars are primary external sources; bot facts are local evidence; trading recommendations are labelled [I].
- The selected market data were read after authoring; where an immutable original payload is unavailable, exact replay is not claimed.
- No open NY trade, NY broker acceptance, future outcome or statistical superiority has been observed in this report.
- Both formula checks and evidence-link checks were run on the report. The manifest records byte counts and SHA-256 hashes for the preserved evidence.
