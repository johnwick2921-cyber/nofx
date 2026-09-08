# Stage A — research snapshot

Status: claimed; premise audit in progress. No Stage A implementation, migration, deployment or live proof is claimed at this checkpoint.

Lane `stage-a-snapshot-96604090/root[unlisted]`; branch `fix/stage-a-snapshot`; accept base `cd5b9a6b9c479eae97eced7fb0abb40594bf1ac5`. The claim's `ls-remote` receipt was `af1ded7e2c0af427b7268ff2fb76cbf20b4997ca`. Source freshness for every cited repository path is in [source-freshness.csv](2026-09-08-stage-a-snapshot-data/source-freshness.csv).

## Governing scope

The governing research is `docs/superpowers/research/2026-09-08-trading-policy/README.md` at **982091d4d908f4a5b8b65022cedf5b8c7c8202d5**, verified **67,959 bytes**. Section **08 Stage A** requires five evidence objects and preservation of actual-contract price scale. Sections **02, 05 and 06** explain the frozen evidence, component/selection provenance and four clocks. Section **09** forbids prescribing target, confirmation, confluence or TTL policy from that evidence. This wave introduces no market belief, trade policy, new refusal or scoring change.

Audit protocol: `AUDIT-CHECKLIST.md`, Part 2 R1–R10; checklist number assigned only at merge. [A] means directly read/executed, [B] inference, [C] unproven. Data reads use SQLite `mode=ro`, `query_only=ON`; the census was frozen within one read transaction. Counts name their retained row IDs; they are not trade or profitability statistics.

## Running revisions and source freshness

[A] At **15:12:33 CT**, `/api/health` returned `33672fdd2cd2`; `/proc/3260027/exe` reported full revision **33672fdd2cd2fee60a2c562a9693e06ab3b13551**, `vcs.modified=false`. A separate locked read-only worktree preserves that running source at `/tmp/nofx-stage-a-running-33672`.

[A] While this audit ran, the plan-liveness deploy owner booted the combined candidate at **15:17:31 CT**, PID **3566770**, revision **f8bc7044cc44d58e84904a0a7761e78b420404af**, `vcs.modified=false`; its observed BOOT INTEGRITY line says goldens PASS. This is not a Stage A boot. The Stage A accept base already contains that code; the refreshed dev tip remained `cd5b9a6b`. The candidate/score/ordinal/store and NT8 source files cited below are byte-identical between the initial running revision and this branch. Planner and confirmation code additionally contains the separately owned combined-wave changes; that difference is named where relevant.

## C1 — empty components and discarded propensity

[A] Current census: **936 candidate rows, IDs 1–936, across 39 reads**. All **936** have `score_components='{}'`. **469 cut rows**, IDs listed individually in [census.json](2026-09-08-stage-a-snapshot-data/census.json), all have `score=0`, empty grade, and the same reason **`max_levels: 12 seated, cap 12`** (cut ID extrema 10–931). The dispatch's 600/181 figures are not today's sample. The pinned planner audit itself describes **600 rows / 301 excluded**, not 181; no historical denominator is silently substituted.

[A] `trader/auto_trader_planner.go:2355–2363` strips `[]ScoredLevel` to each embedded `DetectedLevel` before `recordDetectorOutputs`. `kernel/detector_recorder.go:43–49` retains scores only for the final seated rows; `:70–73` initializes components to `{}`, and the cut branches at `:83–90` infer a cap reason from the final table size. `trader/detector_record.go:108–126` persists that result. The zero cut score is lost evidence, not a computed poor score. The intermediate pool is already capped/collapsed; it does not retain every original detection.

## C2 — scoring inputs that survive

[A] `DetectedLevel` (`kernel/levels.go:70`) carries bounds, origin date, detection TF, zone pattern and formation time. `ScoredLevel` (`kernel/levels_score.go:24`) adds grade, freshness label, score, confluence, distance and role. The actual score uses type/zone evidence, freshness, capped family confluence, size and HTF/TF factors, followed by grade overrides, cluster collapse and seating passes (`kernel/levels_score.go:430–578`). `candidate_pool` retains price/kind/label plus the final seated score/grade and an observed weakest-seated threshold. It does **not** persist bounds, pattern, TF, formation, freshness, family count, raw/capped factors or applied overrides. Its `rank` is nearest-first seated output position, not a pure score ranking. No factors may be reconstructed later and presented as captured inputs.

## C3 — ordinal premise corrected

[A] The claimed per-read reset is **not reproduced as the current call path**. `trader/detector_record.go:83–87` already passes the episode's own CME session day into `store/touch_outcomes.go:125`, which reads the stored maximum; formation-aware de-duplication precedes writes. Current table: **3,093 rows, IDs 1–3093**, of which **743** have ordinal 1. The old 677 rows remain classified **254 invalid:duplicate + 423 legacy:unverified**; their ordinal-1 counts are 254 + 217 = the historical 471. Newer rows include **2,287 unverified:no_formation** and **129 valid** rows, with 251 and 21 ordinal-1 rows respectively. These validity classes must not be pooled into a certified ordinal sample.

[A] Counterexamples to “every read resets to one”: valid row IDs **733–737** carry ordinals **2, 3, 4, 5, 6** for the same 29539.375 level and formation stamp. Row **739** carries ordinal 2 in a different session day. All IDs and event stamps are retained in the census. This is evidence of the corrected caller, not certification of every historical ordinal. Stage A does not modify the detector or ordinal rule.

## C4 — authoring provenance

[A] **56 `planner_read_facts` rows, IDs 1–56**: all have `version=0`, `tokens_in=0`, and empty `bias_ai` / `bias_tree`. The “regime word and little else” description is incomplete: void list/count, ATR, stop floor/multiplier, scope start/interval/bar count and a hash also survive. `persistReadFacts` (`trader/auto_trader_planner.go:2446–2481`) populates these, but its `PromptHash` is assigned `in.AIConfigHash`, not a hash of the exact submitted prompt. It has no accepted-version update or per-attempt timeline.

[A] The author/repair/resend paths exist in `runPlannerReadCoreWithFactsGradesClock` (`trader/auto_trader_planner.go:1470`). Validator rejects have attempt, reason and prompt rows; provider failures intentionally skip that store. Accepted versions have model/prompt/config identifiers and normalized plan output. These stores do not form a complete exact-input → attempts → accepted-version chain.

[A] The four cited durations are directly present in the pinned audit's [retained planning log](2026-09-08-stage-a-snapshot-data/frozen-planning-log.txt): **378.0 s at 20:47:57 CT**, **683.7 s at 21:56:43**, **423.3 s at 22:03:44**, and **683.4 s at 22:16:50**, all September 7. A fresh query of the current journal for that period returned no matching retained lines; therefore this is verification of the frozen artifact, not a claim of newly observed calls or exact recoverable request-start clocks. The failed/repair fixture will retain both durations and the reason without inferring an unobserved time from rounded elapsed text.

## C5 — four clocks

[A] `bars` has only symbol, timeframe, open-time stamp, OHLCV and convention. It cannot answer receipt/availability time; overwrites do not preserve prior versions or their correction times. Candidate `read_at_ms` identifies a read instant, while `created_at` is the DB writer's clock, not a source-event clock. Planner fact rows have only their writer's creation time and scope start; scope start is not receipt time. Plan `created_at` and confirmation's window reference previously stood in for market-history boundaries. The combined wave now adds explicit confirmation evaluation/reference evidence and authored validation; Stage A must record those existing results rather than reinterpret them.

[A] A concrete conflation is the candidate path using a read stamp while dropping `DetectedLevel.FormedAtMs`; a new DB creation stamp cannot establish when the historical bar/level became available. Another is windowing level history with `cur.CreatedAt` at `trader/auto_trader_levelstate.go:120–124`. These are different clocks, even if their values sometimes coincide. No missing source timestamp will be filled from another clock.

## C6 — actual-contract scale and remaining unknowns

[A] The live subscription receipt at **15:19:04 CT** reports **MNQ 09-26** (and separately ES 09-26) in [nt-symbols.json](2026-09-08-stage-a-snapshot-data/nt-symbols.json). `VLContractResolver.cs:51–55,129–150` uses quarterly expiries, third Friday minus eight days, and the condition `rollDate >= today`; `.c.*` aliases are stripped to a root before resolving a request contract. `VLBarsSubscriptionManager.cs:177` resolves the instrument, emits its full name, and constructs a `BarsRequest` at `:356`.

[A] The stored `bars` key is root symbol/timeframe/open time, with no per-bar contract or adjustment-policy field. The inspected AddOn request does not explicitly set a merge/adjustment policy, and the wire does not supply one. Therefore **historical per-bar contract and back-adjustment status are UNKNOWN**. A current subscription contract cannot certify every historical bar's actual contract or orderable price scale. Stage A will retain the received contract and its basis where known, and separately expose these unsupported historical fields as NULL with reasons. It will not change roll handling or C# behavior.

## Implementation and proof status

Pending: record-only schema/capture/writers/export, red/green production pins, fault and latency bounds, mutation evidence, Guide/SYSTEM-MAP, merge-time checklist, suite/build, owner-controlled cutover and live receipts. No trading decision, score, contract, order, DB row or runtime setting has been changed in this audit.
