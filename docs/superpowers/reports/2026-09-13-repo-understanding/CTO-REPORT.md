# CTO repository and trading-process report

Generated from preserved Markdown sources by `tools/build-reports.py`. This is an editable assembled report; make durable corrections in the linked originals and rebuild. Baseline source scope is **1,049 files / 250,582 lines at 63968be62e44db2fb07a92883e02127b9064b0be**, across 28 primary reviews plus two bounded independent reviews. Source review is not runtime verification, deployment approval, or profitability evidence.

**Final source revision, combined test results and release/guide stamps remain root-owned and pending unless explicitly recorded in the underlying final disposition.** Packaging is not an additional audit or test pass. The packaging-time updates below supersede only their named historical open items; unified ordered execution remains active at this snapshot.

Relative links have been rebased to the original artifacts; fragment links point to the original document to avoid duplicate-heading ambiguity. Code fences and original source files are preserved. Machine-readable baseline consistency evidence remains in [coverage-validation.json](coverage-validation.json) and [publication-validation.json](publication-validation.json); the checkpoint describes its limits.

## Contents

- [PACKAGING-UPDATES.md](#section-1)
- [CTO-TRADING-LOGIC.md](#section-2)
- [REPAIR-STATUS.md](#section-3)
- [CORE-TRACE.md](#section-4)
- [CHECKPOINT.md](#section-5)

---

<a id="section-1"></a>

> Original: [PACKAGING-UPDATES.md](PACKAGING-UPDATES.md). Preserved source document; its own revision/checkpoint statements govern. Read Packaging-time repair updates before interpreting older open lists.

## Packaging-time repair updates

This additive disposition notice supersedes the specific older open items named below. It preserves the original reports and their revision-scoped evidence. Commit existence and the corresponding source/report changes were inspected during packaging; this packaging step did not rerun repair tests. The root reviewer still owns final combined verification and revision stamps.

| Later committed repair | Superseded historical open item | Scope and remaining limit |
| --- | --- | --- |
| `c20d0a82` — unresolved exits and account truth | First-ever missing position snapshot treated as flat; valid exit quantity greater than currently materialized entry discarded | Missing/stale account state refuses admission; excess exit receipt remains pending for later resolution. Pending evidence does not itself prove replay or causal execution ordering. |
| `d3e4638e` — observer replacement handoff | Replaced reconciliation observer continues indefinitely | Synchronous subscriptions and receipt-channel drain retire the replaced instance's periodic worker. Ordinary Stop retains protection observation; an already running reconciliation pass may finish. This does not prove live NT8 lifecycle scheduling or all shutdown paths. |
| `3f21431a` — terminal cancelled/rejected EXIT fills | Positive-filled terminal EXIT outside the Filled-only wire path | The bounded C# repair includes positive terminal cumulative exits and retains bracket siblings for residual exposure. Its report records an extracted 89-assertion harness and five-source reference compilation; these are reported offline evidence, not an installed AddOn or real fill. |
| `dd11670b` — open-order display uncertainty | HTTP failures displayed as fresh empty orders; editable existing account name silently ignored | Same-view prior snapshot retained as UNKNOWN/stale; selected-account changes invalidate requests, while the banner discloses trader-bound endpoint scope. Existing binding names are read-only; account creation remains editable. Backend selected-account filtering and binding migration remain separate work. |

**Still active at packaging:** the unified ordered-execution repair and its final verification. Do not interpret the preceding patches as resolving every interleaving between transport receipt, entry accounting, exit accounting and position reconciliation. The root reviewer will update the final disposition after that work and merged-head tests finish.

**Still unverified or open:** broker-side atomic expected-position fencing; crash-safe delivery of process-local hooks; eventual pending-receipt replay; ambiguous multi-row attribution; selected-account completeness in history/chart endpoints; other explicitly listed UI/persistence/operations findings without a named repair. No source audit establishes installed NT8 behavior, backup restoration, production safety, or out-of-sample profitability.

Source receipts: [first snapshot/pending exit](https://github.com/johnwick2921-cyber/nofx/commit/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d), [observer replacement](https://github.com/johnwick2921-cyber/nofx/commit/d3e4638e495425874aedba34a2fab0585acdae85), [terminal EXIT](https://github.com/johnwick2921-cyber/nofx/commit/3f21431aff0d7830b0dc0457417365a7a555aee7), [order display](https://github.com/johnwick2921-cyber/nofx/commit/dd11670be159760942d886b5b8c1aef2573e0d02). These identify source commits, not final release stamps.

---

<a id="section-2"></a>

> Original: [CTO-TRADING-LOGIC.md](CTO-TRADING-LOGIC.md). Preserved source document; its own revision/checkpoint statements govern. Read Packaging-time repair updates before interpreting older open lists.

## CTO assessment: does the system behave like a disciplined level trader?

Status: SOURCE REVIEW COMPLETE; repairs and final combined verification in progress. Not a deployment or profitability approval.

Source references below are repository-relative. Strategy source was inspected at
`63968be62e44db2fb07a92883e02127b9064b0be`; repair-specific behavior is at
the repair commits named in [REPAIR-STATUS.md](REPAIR-STATUS.md). Named
function boundaries below avoid carrying baseline line numbers onto changed source. The repair branch is undeployed.
[A] means source inspected or an explicitly named test run; [B] means inference.
Owner-supplied backtest numbers are not independently recomputed here. This is a
repository engineering assessment, not a new literature review or evidence that
a discretionary trading doctrine is profitable.

### Executive judgment

The system has many useful controls, but their number does not establish a
coherent trading process. I cannot sign off on complete trading correctness.
All 30 scoped review reports are complete:28 source slices cover1,049 files and
250,582 lines; two cross-boundary reviews examined repair 99a06543. Subsequent
repairs receive separate focused review/tests. Final combined verification is
still pending. The number of reviews is not the number of simultaneous agents.

The core standard is consistency: the same setup identity, contract, account,
entry, stop, target, permission and lifecycle must survive from market data to
planner to risk admission to the broker and finally into performance records.
A convincing prompt cannot compensate for a different execution rule. A green
unit test cannot establish profitability. A successful socket write cannot
establish an accepted order, a fill, a cancellation or protection.

### 1. The job is to select a worthwhile trade, including selecting no trade

[B] “Find the best trade” must mean best among currently eligible candidates
under an explicit, validated objective. It cannot mean always produce an order,
choose the highest model confidence, or choose the largest displayed R multiple.
The repository review has not established that the ranking identifies the best
future outcome or a positive-expectancy subset.

A reviewable selection record needs: observation time; concrete contract and
account; session; frozen level identity; scenario/version; evidence for the
setup; entry condition; invalidation; opposing obstacles; composed geometry;
all admission results; and why the chosen candidate outranked alternatives.
This is an engineering acceptance requirement, not a new enabled trading rule.

[A] `trader/armed_executor.go` separates placement from authoring. Repairs
2dc94a19 and 456b38d4 exercise the actual cycle and loopback transport: missing
permission or current quality refusal cannot leave an inherited authorization
eligible. Placement consumes IDs admitted in the current cycle. Retirement
write failure prevents placement. These are enforcement tests, not merely
assertions about the model's prose.

### 2. Stop location: invalidate the identified setup, then report its exposure

[A] `ResolveEntryGeometryZone` in `trader/structural_geometry.go` resolves the entry zone from the scenario's
frozen identity and source provenance. It refuses missing or ambiguous identity.
This is stronger than attaching the stop to whatever unrelated level is nearest
when the order is submitted.

[A] For a valid reject-fade zone, `composeGeometry` in that file sets the long stop
below the lower zone edge by the configured buffer, rounding outward to the
contract tick. A short stop is above the upper edge. Invalid/missing buffer is a
refusal. The ATR fallback recorded when provenance is missing also returns
`no_provenance`: it is diagnostic geometry, not permission to trade.

[A] The production arm caller at `trader/armed_executor.go` (the `structuralFade` branch) selects this
structural branch for the reject play, excluding explicit exit legs. Other
plays retain their legacy stop construction. Therefore “ATR has been removed
from all stops” would be false.

Unresolved: a configured/calibrated buffer is not automatically a validated
noise allowance. Its sample must be causal, contract/session appropriate and
out of sample. An overshoot distribution conditioned only on levels that later
held excludes breakdowns; it cannot by itself establish total stop-out risk.
The research review records that limitation without claiming a new calibration.

### 3. The planner and execution contract now agree on reject-fade geometry

[A] Repair 710ea1c8 updates the actual planner prompt builder, legacy stop-floor
facts and feasibility warning consumer. Reject fades use frozen structural
geometry; the legacy ATR floor is explicitly scoped to other plays. The prompt
no longer suggests switching execution routes to escape a refused arm. Authored
reject stops are not evaluated as if they were the later composed trade. Focused
production builder/warning tests pass and the map golden remains unchanged.
This is a consistency repair; it does not validate the buffer or trading edge.

### 4. Target selection: the next level is precise, but not proved optimal

[A] `FirstGeometryTarget`, `trader/structural_geometry.go`, chooses the nearest
complete sourced zone strictly beyond the entry zone's profit-side edge. It does
not rank by source count, grade, timeframe, minimum reward distance or measured
quality. The long target uses the target zone's near lower edge; shorts use its
near upper edge, with conservative tick rounding.

This answers WHICH level the code selects. It does not establish that all zones
represent equally material obstacles. A very dense map can still leave little
room. Conversely, skipping a nearby real obstacle merely to display 2R does not
prove the farther target is reachable.

[B] Keep “first obstacle” and “trade objective” conceptually distinct in research.
If a different target policy is proposed, compare it prospectively with the
current nearest-zone rule using identical entry events and execution assumptions.
Do not silently reinterpret historical target results or install an untested
quality ranking. A refusal because there is insufficient room is a valid output;
persistent refusals are a diagnostic to investigate, not permission to weaken
the gate until trades appear.

### 5. Entry, stop and target must describe one trade

[A] `composeGeometry` freezes tick-normalized entry, stop and target before
admission. It requires positive directional risk/reward, known costs, positive
reward after those costs, and the configured gross reward/risk threshold. It
reports planned one-contract dollar loss, then leaves daily risk and other
admission checks to their existing gates. Quantity is not authorized by this
geometry function alone.

Important distinction: the current R gate uses gross gain divided by stop
 distance. Costs are checked separately; this is not a net-R gate and not an
expected-value calculation. A 2R objective can lose money if it is reached too
rarely. A lower-R trade can have positive expectation with a sufficiently high
realized win rate. Neither ratio establishes an edge by itself.

For owner-supplied average win W, average loss L, win probability p and constant
round-trip cost c, the simplified expectancy is pW-(1-p)L-c. With W=4.12,
L=24.57, p=0.899 and c=2 points, this is approximately -0.778 points/trade.
The break-even win rate is (L+c)/(W+L), approximately92.6%. Thus “no win rate
saves it” is too absolute for those averaged outcomes; the supplied observed
win rate does not save it. This arithmetic is not a fresh backtest and does not
explain all fill-assumption differences.

### 6. Daily loss is an owner control; it is not a stop-placement algorithm

[A] The owner clarified DAILY loss, not a mandatory additional per-trade cap.
The structural core explicitly retains that distinction at
`composeGeometry` in `trader/structural_geometry.go`. Do not restore the removed cap or silently
set a new value. Contract value converts distance to planned exposure; it does
not decide where the setup is invalidated.

[A] `bootRiskFacts` in `trader/session_risk.go` reads the guardrails master and
individual daily-loss enable switch for boot reporting. This is a reporting
boundary, not evidence by itself that admission enforced the configured amount.
`entryGateForArm` and `entryGateForDecision` supply `DailyForceFlatReason` to
`EntryGate`; its daily-force-flat leg refuses when that resolver reports a trip.
`SessionRiskBootLine` labels a configured but disabled limit decorative. A displayed dollar value alone is
not proof of enforcement. Earlier observed settings are historical snapshots;
this assessment did not read or change current settings.

Required verification: define and test the precise daily boundary, realized and
unrealized components, commissions, account/trader scope, loss-trigger response,
resting-entry cancellation, and protection/flattening behavior. Whether to reserve
remaining daily budget for a prospective stop is a separate explicit owner
policy, not an assumption to introduce during a repair.

### 7. The broker lifecycle matters as much as the setup

[A] Committed repairs now distinguish authorization, placement pending, working,
fill, cancel pending and confirmed cancellation. Tests reproduce and repair:
startup checks occurring after autostart; entry paths escaping the boot latch;
same-version terminal arms being reauthorized; cancellation send being called
settlement; pre-request empty snapshots being used as cancellation evidence;
and previous-process retries exhausting a new process's retry allowance.

[A] A production placement-loop test also reproduced a refused stop entry
cancelling other scenarios as if it had placed. Commit 7b2eb894 makes successful
pre-send registration the commitment boundary. A separate test retains that
commitment after an ambiguous send error. Repair b63747ea covers the analogous limit path through the actual TCP adapter
and registration callback. Both commit on durable registration; neither calls
an ambiguous send a fill or confirmed cancellation.

All these are offline fixture results. They do not prove the installed AddOn
accepted an order or that every asynchronous interleaving is correct.

### 8. NT8 protection and account routing: repaired source, bounded evidence

[A] C# source repairs 9140f6c9/f1b7cc10 resolve explicit accounts without fallback,
select the actual held expiry, refuse ambiguous bare roots, and preserve SIM,
connection and session-account restrictions. Entry cancellation retains deferred
protection until terminal broker evidence. Submit ambiguity retains bracket
identity; cumulative fills amend one pair. Actual leg quantities confirm an
amendment, and synchronous terminal receipts survive submission reentrancy.

[A] Independent review reproduced two gaps in the first repair, then verified
both fixes. All33 extracted-production-method harness assertions pass and all
five AddOn sources compile against installed NT8 assemblies. These checks do
not recreate NT8 scheduling, the broker's OCO implementation or real fills.
The source AddOn has not been copied, compiled in the live NT8 installation or
restarted. No runtime protection claim follows from the temporary DLL.

[A/report] Additional cross-boundary tests reproduced completed partial exits
being recorded as whole-position closes and positive cumulative ENTRY fills on
terminal cancellation being omitted. Repair a982cc74 now records actual exit
quantity, receipt identity, fill and residual cost basis atomically, and handles
cumulative entry growth without overwriting partial-exit accounting. The
completed-exit frame remains Filled-only: positive-filled terminal-cancelled
EXIT orders are still a concrete wire gap. This is not the same as repaired
terminal ENTRY materialization. Current exit wire also lacks commission data;
zero additional recorded fee is unreported commission, not measured zero cost.

### 9. One contract and management rules must remain executable

One contract cannot be partially reduced. “Take some off and leave a runner” is
not an executable MNQ instruction for this account size. Every planned response
must map to hold, full exit, or another expressly supported action. Do not add
contracts to imitate a book's scaling example.

The historical task description says no stop movement, no re-entry and a session
cutoff. Those premises must be reconciled with current implemented stop repair,
management, same-version retirement and new-version authorization rules. A
protective-stop restoration is different from discretionary tightening, but the
UI and logs must identify which occurred. New plan versions must not become an
accidental loophole around the owner's intended re-entry policy.

Delayed flatten repair 94e08cf0 checks immutable position/entry lineage, invalidates
timers on Stop and preserves protection after close refusal. A broker-side atomic
position fence is still absent: a stale local row cannot prove that no unseen
replacement exists. Broker observers intentionally outlive ordinary Stop while
positions may remain. Final removal needs an explicit safe handoff design.

### 10. What establishes success, and the order of work

1. Close account-routing, protection, asynchronous settlement and stale
   authorization defects. Preserve SIM restrictions and real owner settings.
2. Align planner, geometry, admission, execution and displayed explanations.
   Every refusal must name the actual reason and retire incompatible permissions.
3. Baseline frontend/runtime-source and independent cross-boundary reviews are
   complete. Run the full suite on the final combined repaired source, relevant
   race tests, frontend verification and controlled broker lifecycle tests. Mark external/runtime
   checks unavailable until actually observed.
4. Validate the strategy separately: causal detector inputs and frozen levels;
   realistic limit fills, gaps and ambiguity; all costs; one-account chronological
   order/position constraints; actual session exits; and untouched evaluation data.
5. Compare policies with all candidates and refusals retained, not only executed
   winners. Report net expectancy, uncertainty with session dependence, tail loss,
   drawdown from the initial equity baseline, exposure and execution sensitivity.
   A hold rate alone is not trade win rate or expectancy.
6. If credible out-of-sample results remain nonpositive across realistic fills,
   do not label more code or a prettier R ratio an edge. Keep the strategy
   unapproved for promotion while recording the failed hypothesis honestly.

No precise buffer, optimum number of trades, universal target rule or profitable
regime classifier has been established by this source audit. The engineering job
is to make those hypotheses measurable and execution faithful. The trading job
is then to demonstrate that the selected opportunities pay after losses and costs.

### Evidence and completion limits

See README.md, CHECKPOINT.md, reviews/01 through reviews/30, coverage-validation.json
and [REPAIR-STATUS.md](REPAIR-STATUS.md), which separates repaired baseline
findings from concrete remaining source/runtime limitations. Frontend 28a6f32e
passed451 tests/build and was integrated as 6c4092bf. Full Go/build/focused race checks
passed at 99a06543, before later changes. Final combined checks are still due.
No deployment, owner-setting change, live database write or real order occurred.
Historical runtime snapshots from the earlier daily-loss dispatch are not new
observations. The earlier usage-blocked report is archived under interim/.

The final report must distinguish fixed/reproduced defects, static concerns,
intentional policies, dormant legacy paths and checks requiring real runtime.
Neither complete source coverage nor a green suite demonstrates profitable
trade selection. Owner-supplied backtest statistics remain supplied context,
not an independently repeated experiment in this engineering dispatch.

---

<a id="section-3"></a>

> Original: [REPAIR-STATUS.md](REPAIR-STATUS.md). Preserved source document; its own revision/checkpoint statements govern. Read Packaging-time repair updates before interpreting older open lists.

## Repair disposition and publication status

This is the bridge between the **baseline review** and **later repairs**. It prevents a fixed baseline finding from being presented as still current, or a focused repair from being presented as runtime proof. Snapshot: audit branch 65e19141; control repair branch observed at 939e21db; frontend branch 28a6f32e pushed to origin. Root will stamp final integrated hashes and test results after the remaining work. No final combined revision is asserted here.

### How to read the evidence

- **Baseline source review complete:**30 reports/120 standard artifacts are indexed.28 primary slices account for1,049 unique assigned files/250,582 lines at 63968be. Reviews 29/30 examine bounded repairs at 99a06543 and add no primary files. Publication-validation.json independently finds no missing artifacts/index links, duplicate assignments, ledger hash/range gaps or errors in the supplied named-function validators. This is an artifact consistency check, not a second reading of every function.
- **Repair committed:** source was changed on a named repair branch. A stated focused test covers only its actual fixture/call site. Neither branch-green nor an extracted C# harness is final combined or live NT8 evidence.
- **Runtime unverified:** no new broker order, account/settings mutation, live trade-row inspection, deployment/restart, or profitability experiment was authorized or performed by this source audit. Historical receipts remain dated historical evidence.

### Disposition table

| Boundary / review evidence | Disposition and commit scope | Verification and remaining limit |
| --- | --- | --- |
| Ownership, boot entry latch, terminal arms, causal cancellation, cancel retry identity ([01](reviews/01/report.md), [20](reviews/20/report.md), [29](reviews/29/report.md)) | Repaired on control branch; consolidated Go checkpoint 99a06543 | Full Go suite/build and selected race checks at 99a06543. Later integration is separate. Not all-package race or broker runtime proof. |
| Structural planner contract, current-cycle admission, missing permission retirement ([08](reviews/08/report.md), [11](reviews/11/report.md), [30](reviews/30/report.md)) | Prompt 710ea1c8; admission 2dc94a19/456b38d4 and associated control repairs | Production builder/admission fixtures. No buffer calibration, ranking quality or expectancy validation. |
| Stop/limit registration ambiguity ([20](reviews/20/report.md), [29](reviews/29/report.md), [30](reviews/30/report.md)) | Stop 7b2eb894; limit b63747ea | Both commit admission on durable registration. Actual adapter/loop tests; no claim that ambiguous transmission is accepted/filled. The limit gap recorded at 99a06543 is subsequently repaired. |
| C# explicit account/expiry, entry cancel/protection and bracket amendments ([14](reviews/14/report.md), [29](reviews/29/report.md)) | Original 9140f6c9/f1b7cc10 integrated as e8d2243f/cc766e1c |33 extracted production-method assertions and five-source compilation against installed references reported. Live AddOn not installed/restarted; NT8 scheduling/OCO remains unverified. |
| Delayed flatten and Stop observer lifetime ([22](reviews/22/report.md), [23](reviews/23/report.md)) |94e08cf0 binds fallback to position lineage, invalidates timers on Stop, preserves protection on close refusal | Focused lifecycle/race checks. No broker-side atomic expected-position fence; final observer disposal still needs safe handoff. Ordinary Stop cannot blindly remove protection observers. |
| Completed partial exits and cumulative **entry** terminal receipts | a982cc74 adds atomic receipt/fill/residual accounting and cumulative entry growth | Focused store/adapter/trader/race and C# reference/harness evidence reported. Positive-filled cancelled **exit** orders remain outside the Filled-only wire path; process-local hooks can be lost after commit/crash; missing replay/delivery and ambiguous multi-row attribution remain limits. |
| Agent HTTP identity / per-request model choice ([03](reviews/03/report.md), [04](reviews/04/report.md), [26](reviews/26/report.md)) | Control ownership fixes and c8323d09 request-local clients | Synthetic authenticated ownership/model tests; selected race checks. Not a claim that every agent trade-confirmation/background lifecycle is fully isolated. |
| Browser chat, SSE, local history and SWR cache ([25](reviews/25/report.md), [26](reviews/26/report.md)) | Frontend ed85a80a; integrated source 03090352. Data-truth follow-up 28a6f32e uses provider-local dashboard mutate and opaque cache key | Production stream/cache fixtures, arbitrary chunk splits, late-user/same-user completions. Browser abort is not backend tool cancellation. Unowned guest history retained separately, not assigned to later users. |
| Indexed plan edits ([24](reviews/24/report.md), [30](reviews/30/report.md)) | Control API expected-revision checks plus frontend opening-snapshot save/delete tuple in ed85a80a | API temporary-store conflicts and frontend draft/poll fixture. Historical Ask/Q&A proposals lack equivalent authored-version identity; separate JSON Patch test operations are not the same guarantee. |
| Corrected P&L UI and chart ownership ([26](reviews/26/report.md), [28](reviews/28/report.md), [30](reviews/30/report.md)) | Frontend 28a6f32e integrated as 6c4092bf | Missing/nonfinite corrections excluded and counted; server aggregate counts separated from loaded filtered counts. Late chart/history results and SVP toggle races tested. Selected-account completeness, forming-candle marker association and empty-on-error open-order wrapper remain open. |
| Modal drafts, NT edit defaults, config/errors, market selector and unsafe FAQ ([25](reviews/25/report.md), [28](reviews/28/report.md)) | Frontend ed85a80a / integration 03090352 | Core targeted fixtures and type/build checks. Not every wallet/model form path reproduced. Bulk model replay/extra-instance knobs, submit busy lifecycle and account-name persistence remain open. |
| Breaker zero/default display ([27](reviews/27/report.md)) | Frontend 28a6f32e removes false Off in **RiskControlEditor**, not DayPlanEditor |0 means server threshold(default 8 unless overridden); env 0 can disable. No daily-loss policy, saved values, quantities or mandatory per-trade cap changed. |
| Swing wick provenance / aggregate volume / weekly daily-input reader ([10](reviews/10/report.md), [12](reviews/12/report.md)) |05a1775c and df4af389 | Synthetic detector/aggregate and actual weekly-reader regressions. Generic epoch aggregation elsewhere and daily-data limits on intraday gap timing remain open. Corrected input measurement, not proof of improved trading returns or recalibrated distributions. |
| Five dependency advisories | gnark-crypto0.19.2 in ffbcf3e4; four npm transitives in e26f63a8, integrated f5409132 | Targeted compatible versions; npm audit 0and web suite/build at private updated install. Advisory affected versions confirmed by paginated GitHub read. Default-branch alert closure awaits merge/scanning; exploitation/reachability not established. |
| Maps/guide discrepancies ([24](reviews/24/report.md), [30](reviews/30/report.md)) | Existing stale guide paragraphs corrected 6058d9fe; actual-fill guide 939e21db; final GUIDE_BUILT_REV/root map synchronization pending | Historical UA/CGC stay historical. Narrow map line guards do not verify all prose. Final published source revision must be stamped after integration. |
| Backup, runtime composition and strategy profitability ([17](reviews/17/report.md), [18](reviews/18/report.md), [19](reviews/19/report.md), [30](reviews/30/report.md)) | **Not certified by source review** | Prior receipt is not a restore rehearsal. Partial backup cleanup/weekly copy risks remain reported. First live composition/refusal, NT8 lifecycle and causal out-of-sample net expectancy remain unverified. |

The table tracks principal reviewed repair boundaries, not a declaration that every finding in all 30 reports was repaired. Findings without an explicit repair disposition retain their baseline status and reachability qualifications. In particular, crypto broker/client contracts ([06](reviews/06/report.md), [07](reviews/07/report.md), [13](reviews/13/report.md)), persistence/query concerns ([15](reviews/15/report.md), [16](reviews/16/report.md)), and operations/tooling concerns ([19](reviews/19/report.md)) were reviewed but not globally rewritten. They must not disappear behind the phrase “source review complete.”

### Verified checkpoints, not a manufactured final green

| Checkpoint | Evidence available | Does not establish |
| --- | --- | --- |
| Go 99a06543 | Full go test ./..., go build ./..., selected focused race tests | Later merged C#/Go/frontend correctness or every race |
| Frontend 28a6f32e |71 Vitest files/451 tests; TypeScript/Vite build; private dependency tree, npm audit 0 after compatible updates | Final control-branch suite, real browser E2E, broker execution or security completeness |
| C# lifecycle source pair |33 extracted-method harness assertions; five AddOn sources compile with NT8 references | Installed AddOn version, real callback timing, account/broker OCO or live fills |
| Coverage publication check |30 indexed reviews,120 artifacts,1,049 unique primary files,250,582 lines; errors[] | Every tracked test/document read, exact function correctness, runtime safety or edge |

Final integration revision: **pending root stamp**. Final combined Go/web/race/build outputs: **pending root stamp**. Final guide revision: **pending root stamp**. Final bundle/manifest and publication links: **pending root stamp**. These fields are intentionally unresolved, not inferred from preceding checkpoints.

### Open work that must survive publication

**Execution correctness:** positive-filled terminal-cancelled EXIT orders still lack the completed-Filled exit wire receipt; broker-side atomic expected-position fencing is absent; observer replacement/final removal needs a safe handoff; process-local hooks are not transactionally delivered across a crash after receipt commit; durable pending receipts still require later replay/delivery; ambiguous same-account/root/side multi-row attribution refuses rather than infers. Completed partial exits and cumulative ENTRY terminal receipts were repaired in a982cc74 and must not be conflated with the remaining cancelled-EXIT gap. Controlled NT8 lifecycle confirmation remains outstanding.

**Active UI/data truth:** account-qualified history/chart completeness; loaded-window versus complete aggregate distinctions; forming-candle marker matching; open-order failures represented as empty; bulk model payload replay and missing extra-model thinking knobs; submit request lifecycle; editable account name without backend persistence; DayPlanEditor polled-draft reset/default/inheritance/translation issues. The breaker display, ordinary corrected-PNL fallback and late-response identity bugs listed as baseline findings have specific later repairs above.

**Research and historical analysis:** many scripts are archived exploratory tools rather than production admission. Flaws there do not by themselves prove the live measurements are wrong. Equally, an in-sample held-level overshoot distribution does not establish unconditional stop risk, trade win rate or profitable expectancy. Retain chronological/account constraints, all candidates and refusals, fill ambiguity, costs and untouched evaluation samples for any later strategy experiment.

**Dormant paths:** legacy crypto/CSV/competition/chart/reset-password prose or interfaces are not automatically active execution or reachable exploits. Preserve the reachability qualifications in individual reports. The current NT8 SIM source path and its active issues take priority.

### Concrete corrections found during publication review

1. CHECKPOINT still called all frontend state work active although its source/dependency/core-data batches were committed and the last branch pushed. Now separate lane completion from integration/final verification.
2. CTO section 10 still instructed completing runtime/frontend source reviews after its header said all 30 complete. Now source review is complete; controlled runtime verification is a different outstanding activity.
3. CTO daily-risk citation used `session_risk.go:279`, which is **boot reporting facts**, as if it established admission. Now explicitly identify the reporting boundary; runtime enforcement needs its own tests/call sites.
4. CTO structural function coordinate 28 was a comment/start vicinity, not the resolver declaration 30; named function anchors now identify the boundary. Review-specific source lines must not be silently restamped as final-repair locations.
5. Repair report on the separately owned control branch still described frontend propagation and npm work as pending at the snapshot. This table records the committed lanes without editing that branch; root must update its final repair README/hash stamps.
6. Thirty completed assignments are not thirty simultaneous agents, thirty new full source inventories, or a final merged test pass. Primary and cross-review counts remain distinct throughout.

The detailed control repair report lives on `fix/repo-audit-control-boundaries-20260913` at `docs/superpowers/reports/2026-09-13-repository-repairs/README.md`; frontend evidence lives on `fix/repo-audit-web-state-20260913` at `docs/superpowers/reports/2026-09-13-web-state-repairs/README.md`. Include those branches/reports in the final publication bundle. Neither is rewritten as part of this docs-only publication pass.

---

<a id="section-4"></a>

> Original: [CORE-TRACE.md](CORE-TRACE.md). Preserved source document; its own revision/checkpoint statements govern. Read Packaging-time repair updates before interpreting older open lists.

## Core trading source trace

Source snapshot: `c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d`. Each link names an exact committed declaration. This is a selected source locator, not evidence that every branch ran. Detailed baseline function notes and connections are in the 30 review folders. Execution-order and integration tests must be read beside these source links.

| Boundary | Exact source | Responsibility / transfer limit |
| --- | --- | --- |
| Transport | [readLoop](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/provider/ninjatrader/tcp_server.go#L1725) | Decodes broker frames; execution processing order must be verified separately. |
| Market data | [GetWithTimeframes](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/market/data.go#L190) | Builds requested market context; futures provider path differs from legacy crypto. |
| Canonical symbol | [Normalize](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/market/data.go#L670) | Preserves the CME normalization boundary. |
| Level evidence | [AssembleResearchLevels](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/kernel/levels_assemble.go#L212) | Assembles raw, pool and seated candidates; heuristic scores are not probabilities. |
| Weekly evidence | [weeklyDailyBars](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/auto_trader_weekly.go#L138) | Preserves daily input for CME-week aggregation. |
| Weekly facts | [CompletedWeekCandles](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/kernel/weekly_bias.go#L93) | Groups observations into completed Monday-governed weeks. |
| Planner invocation | [runPlannerReadCoreWithFacts](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/auto_trader_planner.go#L1121) | Machine facts and model response enter planner persistence/validation. |
| Frozen setup identity | [ResolveEntryGeometryZone](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/structural_geometry.go#L30) | Rejects missing or ambiguous frozen source identity. |
| First structural obstacle | [FirstGeometryTarget](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/structural_geometry.go#L78) | Chooses nearest complete sourced zone beyond the entry zone. |
| Geometry | [ComposeLevelFadeGeometry](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/structural_geometry.go#L106) | Production wrapper freezes structural stop/target before admission. |
| Geometry arithmetic | [composeGeometry](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/structural_geometry.go#L123) | Zone-edge buffer, outward rounding, costs and gross-R refusal; no ranking proof. |
| Arm orchestration | [maybeManageArmedOrdersAt](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/armed_executor.go#L199) | Current-cycle authorization and gate results precede placement. |
| Placement | [runArmedPlacementAt](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/armed_executor.go#L1184) | Consumes currently eligible arm identities and broker/account evidence. |
| One-contract guard | [oneContractGuard](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/one_contract.go#L159) | Account exposure and entry-order admission boundary. |
| Session controls | [sessionRiskGateAt](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/session_risk.go#L120) | Session breaker/band verdict; does not alone establish daily-loss implementation. |
| Daily reporting | [bootRiskFacts](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/session_risk.go#L260) | Reporting facts only; do not cite as executable daily-loss gate. |
| Decision risk | [GetFullDecisionWithStrategy](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/kernel/engine_analysis.go#L57) | Strategy decision/control pipeline; distinct from resting-arm placement. |
| Position admission | [ntHeldPosition](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/auto_trader_orders.go#L377) | Broker position errors must remain unknown instead of flat. |
| Decision long entry | [executeOpenLongWithRecord](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/auto_trader_orders.go#L464) | Actual decision entry call site; admission failure must prevent wire submission. |
| Decision short entry | [executeOpenShortWithRecord](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/auto_trader_orders.go#L612) | Short counterpart requires the same ownership and exposure discipline. |
| Resting limit | [PlaceLimitEntry](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/ninjatrader/tcp_trader.go#L446) | Registers identity before transmission; transmission is not broker acceptance. |
| Stop entry | [PlaceStopEntry](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/ninjatrader/tcp_trader.go#L510) | Kind-specific stop entry adapter; distinct from protective stop placement. |
| Cumulative entry | [onArmedOrderUpdate](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/armed_executor.go#L1952) | Consumes actual entry state/quantity including positive terminal cancellations. |
| Entry accounting | [materializeArmedEntry](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/armed_executor.go#L2068) | Preserves cumulative entry quantity/notional and residual position accounting. |
| Broker exit | [recordClose](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/ninjatrader/close_sync.go#L88) | Builds actual exit receipt with account and broker-order identity. |
| Atomic exit | [ApplyNT8Exit](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/store/nt8_exit_receipt.go#L44) | Receipt, actual fill and residual/P&L update share one transaction. |
| Reconciliation | [reconcilePositions](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/ninjatrader/reconcile.go#L118) | Reconciles observations; a database row is not broker-flat proof. |
| Positions truth | [GetPositions](https://github.com/johnwick2921-cyber/nofx/blob/c20d0a829ca8cc0a9975ab78eed3b6a700f3b11d/trader/ninjatrader/tcp_trader.go#L921) | Selected bound-account position snapshot and freshness admission. |

The ordered execution boundary is under repair until final verification is stamped. Transport receipt, broker acceptance, execution, position reconciliation and permission to enter are separate facts. The root report records their verification status.

---

<a id="section-5"></a>

> Original: [CHECKPOINT.md](CHECKPOINT.md). Preserved source document; its own revision/checkpoint statements govern. Read Packaging-time repair updates before interpreting older open lists.

## Current audit, repair and publication checkpoint

The usage-blocked checkpoint under `interim/` is historical. Start with
[REPAIR-STATUS.md](REPAIR-STATUS.md) for the current repair/open-limit table;
[CTO-TRADING-LOGIC.md](CTO-TRADING-LOGIC.md) explains the trading-process judgment.

### Completed baseline coverage

All 30 scoped review assignments are reported:28 primary slices cover1,049 unique
first-party source files /250,582 lines at **63968be62e44db2fb07a92883e02127b9064b0be**.
Two independent cross-reviews use repair 99a06543 and add no primary coverage.
The independent publication check finds all 30 index links and120 standard artifacts,
no duplicate assignments, missing reports, unread items, hash/range gaps or
recorded named-function validator errors. These are baseline consistency facts,
not full-source coverage of later repairs, every test/document read, or runtime proof.

### Repair checkpoints

Control branch: `fix/repo-audit-control-boundaries-20260913`, observed at 939e21db.
Frontend source branch: `fix/repo-audit-web-state-20260913`, pushed at 28a6f32e.
Root is integrating and testing; do not use these snapshot hashes as final publication stamps.

- Go full suite/build and selected race checks passed at 99a06543, before later repairs.
- C# lifecycle source 9140f6c9/f1b7cc10 integrated as e8d2243f/cc766e1c; extracted-method harness and NT8-reference compile evidence are scoped in the repair report. No installed AddOn change occurred.
- Frontend source/dependency/state commits ed85a80a/e26f63a8/28a6f32e integrated as 03090352/f5409132/6c4092bf. Private updated dependencies passed71 files/451 tests and production build at 28a6f32e; npm audit reported0. Default-branch GitHub alert closure is not asserted.
- a982cc74 repairs completed partial exits and cumulative ENTRY terminal receipt accounting. Positive-filled terminal-cancelled EXIT wire handling remains open. A receipt transaction cannot guarantee process-local hooks across crash; pending receipts still need later delivery/reconciliation.
- df4af389 fixes the weekly reader's production daily-input call site. Other generic epoch aggregation and daily-data intraday-gap limits remain separate.
- Guide 6058d9fe/939e21db removes old ATR/permission claims and explains residual exposure. Final GUIDE_BUILT_REV must follow final source integration.

### Remaining source and runtime limitations

Concrete execution gaps: positive-filled cancelled EXIT orders outside the Filled-only
exit wire path; no atomic broker-side expected-position fence; safe observer replacement
and final disposal/handoff; crash-time process-local hook delivery; unresolved replay and
ambiguous multiple-row attribution. These are not dismissed as minor cosmetic debt.

Active UI/data gaps: selected-account history/chart completeness; bounded loaded history
versus complete aggregates; current-forming-candle marker association; open-order errors
still converted to empty data; bulk model replay/extra-instance thinking knobs; model
submit busy lifecycle; account-name persistence; DayPlanEditor polled-draft/default/
inheritance/translation problems. Specific corrected-PNL, chart identity, cache refresh
and breaker display defects have later repairs and should not be relisted as wholly open.

Runtime and strategy evidence: installed NT8 callback/OCO/fill behavior, first live
structural composition/refusal, backup restore rehearsal, and causal out-of-sample net
expectancy remain unverified by this audit. Source understanding is not runtime or
profitability approval. Archived exploratory script flaws are qualified separately
from currently relied-on measurements in their source reviews.

### Root-owned final publication work

Final integrated source revision: **pending**. Combined Go/web/race/build log results:
**pending** (root started final combined 01; no result inferred). Final guide revision:
**pending**. Final source/evidence bundle, manifest hashes and publication links:
**pending**. Root updates these once actually verified; preceding branch-green results
must not be relabelled as final merged results.

Verified earlier progress bundle: `/tmp/nofx-audit-progress-30-reviews.bundle`,
containing four branch refs and requiring baseline 63968be. It is incremental,
not a standalone full-repository restore. Preserve that limitation when shipping
its replacement and include the separate repair reports in the final evidence set.

No source audit action deployed/restarted the bot or NT8, changed owner settings,
read/wrote live trade records or submitted an order. DAILY loss remains the owner's
control; no additional mandatory per-trade cap or quantity policy was introduced.
Ordinary Stop must retain protection/close observers for held positions until a
safe handoff exists.
