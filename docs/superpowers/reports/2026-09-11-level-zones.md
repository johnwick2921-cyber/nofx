# Level zones — Section C audit; implementation STOP

## C1 — the owner's four lines, traced first

[A] Frozen live planner read: **2026-09-11 10:30:27.269319419 CT**, snapshot
`7f9db41a-815f-4be8-a9e2-ea124db26b67`, research input row **23094445**.
Candidate rows **23093602–23094444**, n=843, writer revision
`802fb00b09e51f9801e8d4fbd1bf156c86865d95`. Price=29485,
ATR5m=42.11359506516323, dATR=465.75, final seats=12, intermediate pool=24.
This is one observed read, not a claim about every read today. Full whitelisted
candidate evidence: [snapshot.json](2026-09-11-level-zones-evidence/snapshot.json).

For the owner's approximate 29700 and 29000 lines, this audit explicitly uses
±15 points, matching half the widest supplied 30-point band. This is an audit
matching convention, not a detector parameter. Other bands use inclusive exact
endpoints. Matching an anchor does not establish that it is the owner's original
Sep 4 feature. A wide historical zone overlapping a band is also not that proof.

| Owner band | Detected anchors | Recorded collapse / selection | Model map |
|---|---:|---|---|
| 29685–29715 | 8 | RN 29700, EQH 4h/1h/15m, EQL 4h, Demand/IFVG/OB 4h. None in the recorded 24-row pool or final seats; exact later cut cause is NULL. No identified Sep 4 high with the requested provenance. | None of these anchors in `Levels`; none in capped `HTFZones`. |
| 29570–29590 | 11 | RN 29575 → VWAP+2σ 29573.0987; EQH·15m 29587 and EQL·15m 29583.75 → EQH·1h 29585. OB·1h 29585.875 is in Pool but absent from Levels. Other exact later cut causes are not all recorded. | No anchor in final map or capped HTF section. |
| 29470–29500 | 33 | SWG-H·15m 29475 **already collapses into SWG-H·5m 29475** (rows 23094168 → 23094164). The 5m survivor reaches Pool at score 1.36, then is absent from final Levels. | The shared swing loses its final seat. ONH 29500.75 is outside the strict band; do not silently count it as inside. |
| 28985–29015 | 1 | Demand·1d 29006.625, row 23094181: `proximity distance=478.375 exceeds band=465.75`. Never reaches collapse or seating. | Absent. **No 1d/4h/1h SWG-L trio detected.** |

[A] There are only 12 SWG records in the entire captured universe, rows
23094161–23094172, all 5m or 15m. `SwingPointLevels` explicitly runs 5m/15m;
the HTF detector registry runs EqualHighsLows, SupplyDemandZones, FairValueGaps,
OrderBlocks, not SwingPointLevels. W-TF's report explicitly says base swings
remain 5m/15m by owner ruling. E1's mandatory three higher-timeframe swing lows
cannot be constructed from this tape without inventing references or expanding
detection scope. Both violate this dispatch.

[B] For 29475, pool membership plus final absence proves final selection loss,
but the archive does not identify the exact overriding seat operation. A NULL
exclusion is not evidence for a specific cut. Research pointers are shared across
subsequent scoring passes; raw recorded final scores/reasons must not be treated
as an immutable transcript of each intermediate pass.

## Operative basis and revision

Owner clarification: Round 21 ran in chat; the dispatch's RESEARCH LAW governs.
The full report is **not yet committed** at
`docs/superpowers/research/2026-09-11-level-zones/` in the accepted tree. No full
Round 21 report or SHA has been fabricated. Pin it when it lands; until then
this limitation remains explicit. The dispatch is the user attachment
`fdf096e7-6f1b-4371-b360-fa3b31fbf123/pasted-text.txt`.

[A] `/api/health` returned `{"revision":"802fb00b09e5","status":"ok","time":null}`.
`go version -m /proc/3366586/exe` returned
`vcs.revision=802fb00b09e51f9801e8d4fbd1bf156c86865d95`,
`vcs.time=2026-09-11T13:41:30Z`, `vcs.modified=false`.
Audited code paths have an empty diff against that running SHA.
Branch `fix/level-zones`, accepted dev tip
`616b52a9def4042ce40308deba46529723ba01d7`; isolated locked worktree
`/tmp/nofx-level-zones`. Claim session `level-zones-fdf096e7/root[unlisted]`.

Playbook: `docs/superpowers/AUDIT-CHECKLIST.md`; apply provenance, sample IDs,
no inferred zeros, binding, call-site evidence, and merged-head validation.
No cutover was attempted. Every cited code/report file's `git log -1` receipt is
in [source-freshness.txt](2026-09-11-level-zones-evidence/source-freshness.txt).

## C2 — bands versus points

[A] **511 real bands, 332 points, 843 total** (not approximately 350).
Real band means `0 < lo < hi`; point means equal positive lo/hi. The struct
already has serialized `Price`, `Lo`, `Hi`, with Price documented as a zone's
midpoint. OB, FVG/IFVG and supply/demand retain their real bounds in the captured
raw origin. A claim that these detectors discarded their bounds is NOT
REPRODUCED. The presentation and merging can still fail to use those bounds.
New display bounds must not overwrite serialized legacy Lo/Hi if E6 holds.

## C3 — fixed tolerance and measured pairs

[A] `clusterToleranceFor` returns `LevelClusterTicks * 0.25`, with ticks=12:
3.00 points. Production calls at this revision:
`levels_score.go:590`; `map_candidates.go:138`, `:227`, `:539`.
The first changes pre-seat scoring output; the other three build/merge map or
projection views. They are not equivalent mutation surfaces.

Unordered pairs from all 843 raw references (354903 possible pairs), strictly
more than 3 points apart, inclusive upper bound:

| m | m × captured ATR5m | Pair count |
|---|---:|---:|
| .25 | 10.528398766290808 | 7165 |
| .50 | 21.056797532581616 | 17077 |
| .75 | 31.585196298872425 | 26506 |

[A] Computed from exact archived anchors, no rounding before comparison. These
are cumulative pair counts, not resulting cluster counts or independent sources;
raw duplicates remain because this is the detected-universe census. ATR5m at
this read is 42.11, not the dispatch's illustrative 20–35.

## C4 — name retention

[A] `collapseLevelClusters` excludes bands from point clustering, records a
loser's label only when its TF differs from the keeper, and propagates prior
CollapsedNames. Same-TF distinct labels are omitted. The 29475 cross-TF pair
already merges; E1's assertion that it must fail today as separate levels is
NOT REPRODUCED. Seat loss still hides both labels because the model map is built
from `in.Levels`, not the raw universe (`planner_prompt.go:552`).

## C5 — ordering, score, daily census

[A] Map order: absolute distance ascending, score descending on equal distance.
Merge keeper order uses score. Scoring/seating also applies today-priority,
score, distance, HTF/volume/both-side seat passes, grade filtering, and final
nearest ordering. There is **no independent 1.2 multiplier in map sorting** to
remove. The 1.2 is inside the non-zone score when `HTF=true`; zones instead use
zoneTFMult. The HTF detection set includes 15m,30m,1h,2h,4h,6h,8h,12h,1d,3d,1w;
actual configured input here emitted 1d/4h/1h/15m references. Daily/weekly zone
tier maps to 4h. PDH/PDL and other anchors can also carry HTF=true.

`zoneSizeMult`: size=(hi-lo)/dATR; thresholds ≤.30→1.25, ≤.60→1.10,
≤1→1, ≤1.5→.85, ≤2.5→.70, otherwise .50; invalid inputs→1.
[A] Daily rows 23094176–23094196: 21 raw; seven in the ±465.75 band including
one same-kind duplicate; six with recorded scores; zero final daily seats.
Highest recorded daily score is **1.27296** (row 23094194), not .862.
The old .862/1.260 figures are **NOT REPRODUCED** on this snapshot. No causal
claim that zoneSizeMult alone explains daily seat loss is established.

## C6 — prior touches are not a free universal count

[A] `touch_outcomes` contains ordinal/outcome and identity fields. Its
`NextOrdinal` scopes by trader, root symbol, exact price and session-day lower
bound, then MAX+1. It does not match formation identity or require the same
contract, and does not impose an upper as-of bound. It is not a safe direct
prior-touch-ranking query for all newly merged references.

The read records 12 `candidate:prior_episodes` facts separately. Row 23093275
(ONH identity) records two detector episodes; row 23093276 (PDH identity) one.
Both say `formation_verified=false`, with basis explicitly limited to the
available detector window. The 843 candidate planner-read records have
`prior_episodes=null`. Do not replace this unknown with zero or interpret ordinal
as a globally independent, formation-correct historical count. Ranking needs a
specified as-of identity and missing-data policy; this audit has not established
a complete count for all 843 references.

## C7 — round numbers

[A] `RoundNumberLevels` uses price ± proximityK×dATR, grids 100/50/25,
deduplicating identical prices. Captured rows 23093613–23093650: 38 references,
29025 through 29950. The measured window is 29019.25–29950.75. Proximity to these
already detected references is available without changing detection. Distance
to an abstract grid outside this window would be a distinct ranking choice.

## Blocking corrections — STOP under A23 / H

1. E1's higher-timeframe swing-low trio is absent by detector design; the
   29475 pair already collapses. Use the actual recorded pair and its final
   seat loss as the live acceptance case. Keep missing higher-timeframe swings
   explicitly outside this wave rather than fabricate them.
2. D4 changes what the existing SCORE term counts, but E6 requires the full
   serialized score output unchanged. Existing code already counts distinct
   families, with a different taxonomy. Altering that input can alter scores
   even when its numerical weight is unchanged. W3's pinned-basis report §2
   records this same conflict and its owner ruling: presentation/order only.
3. D2 says replace every caller, including score-time collapse. That operation
   alters serialized survivor sets and Confluence, while E6 compares Seated
   and Pool JSON. A render-only carrier cannot hide those changes. Do not
   silently weaken E6 or change the scorer to satisfy the new presentation.

Proposed corrected scope for owner ruling: preserve legacy scorer, collapse,
and anchor outputs byte-for-byte; build zones and five-family counts from the
uncut detector universe on a separate render/shortlist path; rank that path
without reading HTF-weighted Score. Apply volatility merge there and explicitly
retain the legacy scorer's tolerance. Use measured available references for E1.
This is a proposal, not an implemented or approved scope change.

No production edits, mutation experiments, changed fixtures, builds, merges,
DB writes, or boots were performed in this audit. No RED/GREEN or golden
preservation claim is made. The score incompatibility is source-proven design
coupling, not a claimed executed mutant. A23 stops implementation here.

## A15 / remaining proof

Full Round 21 SHA outstanding. Exact cut causes are missing for some archive
rows. Full-map prior-touch identity policy remains unresolved. The owner will
still see the existing point-oriented, seated map because nothing deployed.
No backfill was run; proposed zones remain read-time computation. Any future
zone-vs-line or multi-source claim remains UNTESTED; widths, merge factors,
caps and ranking weights remain [I]. The dispatch's proposed experiment and
sample-size/power figures are not independently validated by this audit.

Rollback: documentation-only; running binary and data are untouched. Final
cutover proof, tests at merged HEAD, Guide stamp, boot counts, marker and
SHA-pinned HTTP byte verification remain owed after corrected scope and build.
