# VERIFY-0926 — plan-death census cross-check (DS-104)

READ-ONLY. Zero code changes. Scope: DS-101's `2026-09-26-plan-death-census.md`
(`research/plan-death-census` @ `6bbb3a6c`, PR #238).

Base per L3: `git log -1` = `04ae1c2f RELEASE 6cac1b896bdb… — boot 3: #227 …`.
Worktree `/home/hoang/nofx-ds-104-verify0926` @ `04ae1c2f`, branch
`verify/0926-plan-death` (claim `a8da8fe6`). DB copy `v2.db` made 09-26 00:47 CT
via `sqlite3 -readonly "file:…?mode=ro" ".backup"`. Live logs read-only.

## Verdict

The census is **substantially right, with three corrections**. The 14 killed chains,
their ids, wall times and the born-dead ranking all re-derive exactly. The class
LABELS of 4 of the 14 kills are wrong at the final-killer level, the post-W2 window
boundary mixes UTC storage with a CT literal (83→80 rows, 38→36 reads), and the
census misses a whole reject class — `plan JSON unmarshal` — that ties born-dead
as the top killer (3 kills each).

Counts: **CONFIRMED 6 · PARTIAL 1 · REFUTED 3 · NEW 1.**

## 1. Re-derivation (my copy, ids named)

### 1.1 Window boundary — REFUTED (83 rows / 38 reads)

`planner_rejected_prompts.created_at` is stored as UTC (+00:00). The W2 boot was
09-23 18:50 **CT** = 09-23 23:50 **Z**. DS-101's filter
`created_at >= '2026-09-23 18:50:00'` compares UTC storage against a CT literal
and admits 3 PRE-boot rows: ids **336** (09-23 NY, 18:56:57Z = 13:56:57 CT),
**337-338** (09-23 ASIA, 21:32:50Z/21:34:30Z = 16:32/16:34 CT) — all before the
boot. [A: v2.db]

- Post-W2 true count: **80 rows = 36 reads** (chains: 36; 3-attempt chains: 14).
- DS-101's 83 = 80 + those 3 pre-boot rows; their 38 reads = my 36 + the 2
  pre-boot chains. Their killed table itself correctly starts at id 339.

### 1.2 The 14 killed chains — CONFIRMED (ids + wall times)

All 14 three-attempt chains in my copy are exactly:
339-341, 349-351, 352-354, 357-359, 360-362, 363-365, 376-378, 385-387,
391-393, 394-396, 397-399, 400-402, 409-411, 412-414.
Wall spans re-derived per chain: 9.6, 17.3, 2.9, 3.2, 7.4, 14.9, 6.2, 7.7, 2.3,
10.7, 3.3, 4.6, 12.0, 2.8 min — **mean 7.5, max 17.3, identical to the census**. [A]

### 1.3 born-dead kills — CONFIRMED

363-365, 397-399, 409-411 final attempt-3 reasons all read `born-dead authored
scenario … breached between read and publish`. 3/14 kills. The staleness reading
(each is a real market breach during the read window) matches the text. [A]

### 1.4 Flip-hold claim — CONFIRMED

`DefaultFlipMinHoldMinutes = 30` (`kernel/flip_freshness.go:23-27`), env
`FLIP_MIN_HOLD_MIN`; boot line reads it (`kernel/regime_ledger.go:13`). [A]
(Note: this is distinct from `DORMANT_MIN_HOLD_MIN=5` — the dormant flap guard,
not the flip hysteresis.)

### 1.5 "0 live grammar refusals post-W2" — CONFIRMED

Boot lines read `invalidation: enforced (grammar) · grammar refusals=0`:
09-24 10:19:46 (boot-4c05158b) and 09-25 20:12:52 / 20:14:26 (boot 3). [A]

## 2. The class labels of the kills — REFUTED for 4 of 14

Re-bucketing the 14 kills by the FINAL attempt-3 reject reason in my copy:

| chain | DS-101 label | final reason (my copy, verbatim start) | my class |
|---|---|---|---|
| 339-341 | flip/bias | `plan flip{below 30735.25 5m_close → short} met between read…` | flip ✓ |
| 349-351 | obstacle chain | `S1 obstacle chain: first_obstacle 30687.75 is not the nearest seated…` | obstacle ✓ |
| 352-354 | **obstacle chain** | `S1 identity unresolved: level_id "a285e150…"` | **identity_unresolved** |
| 357-359 | **tape verify (breakdown)** | `plan JSON unmarshal: json: cannot unmarshal string into Go struct field PlanBreakdownConti…` | **schema_json** |
| 360-362 | **entry-policy shape** | `S1 breakdown_continue: the tape shows NO confirming close beyond 30439.50 yet…` | **tape_verify** |
| 363-365 | born-dead | `born-dead authored scenario: S3: … breached…` | born_dead ✓ |
| 376-378 | flip/bias | `plan flip{below 30714.00 5m_close → short} met between read…` | flip ✓ |
| 385-387 | obstacle chain | `S1 obstacle chain: omits SWG-L·5m 30771.00…` | obstacle ✓ |
| 391-393 | **tape verify (breakdown)** | `plan JSON unmarshal: json: cannot unmarshal string into Go struct field PlanBreakdownConti…` | **schema_json** |
| 394-396 | entry-policy shape | `plan JSON unmarshal: json: cannot unmarshal number 0.5 into Go struct field PlanArmLeg.sce…` | schema_json (chain is arm-policy dominated: attempts 1-2 = planned_order illegal on leg 2, unknown policy "on_confirm") |
| 397-399 | born-dead | `born-dead authored scenario: S2: … 2x5m close below…` | born_dead ✓ |
| 400-402 | gap trigger | `gap-up at 30975.50 (> PDH 30827.50)…` | gap ✓ |
| 409-411 | born-dead | `born-dead authored scenario: S2: … 5m close below…` | born_dead ✓ |
| 412-414 | entry-policy shape | `scenario[0].confirm.side "" invalid (above|below)` | entry_policy ✓ |

Corrected kill ranking (14): **born_dead 3 · schema_json 3 · obstacle_chain 2 ·
flip 2 · identity_unresolved 1 · tape_verify 1 · gap_trigger 1 ·
entry_policy_shape 1.**

DS-101's "top-3 = born-dead 3, obstacle chain 3, entry-policy shape 3" does not
survive at the final-killer level: obstacle chain is 2 (352-354's final killer is
an identity failure), entry-policy shape is 1 (360-362 ends on a tape-verify,
394-396 ends on a JSON unmarshal), and **schema_json — absent from the census
entirely — ties born-dead at 3**.

### 2.1 Read-final buckets (all reads) — PARTIAL

My 36-chain buckets vs DS-101's 38: born_dead 9 ✓ · obstacle_chain 9 ✓ ·
entry_policy_shape **4** (they say 6) · **schema_json 3 (missing from their
taxonomy)** · feas_other 3 (they say 2) · flip 2 ✓ · identity_unresolved 2 ✓ ·
feas_no_provenance 2 (they say 4) · tape_verify 1 (they say 2) · gap_trigger 1 ✓.
The differences are the same mislabels plus the two pre-boot chains.

## 3. NEW finding — schema_json is a top killer and the repair prompt gives it no law

Three killed reads (357-359, 391-393, 394-396) and 3 further read-finals die on
`plan JSON unmarshal` — the model emits wrong JSON types for
`PlanBreakdownConti…` / `PlanArmLeg.sce…` fields across all three attempts. The
repair prompt (`BuildPlannerRepairPrompt`, `kernel/planner_repair.go`) appends
validator errors verbatim and selects law excerpts by pattern — and
`lawExcerptsFor` has **no `unmarshal`/`json` case** (cases at `planner_repair.go
:64-122`; verified by grep) — so every JSON-shape reject gets the GENERIC
fallback excerpt, not the field's schema. The model then re-emits the same wrong
type. Report-level finding (no code): a schema-shape excerpt (or the failing
field's type contract) in the repair law map is the cheapest un-claimed fix in
this census, and it addresses a kill class the census does not name.

## 4. Recommendations re-check

- Rec 1 (faster reads — born-dead IS staleness): stands (3/14 kills + 9 read-finals).
- Rec 2 (enumerate obstacles in the prompt): stands (2/14 kills).
- Rec 3 (legal arm shapes in the prompt): still valid but its kill share is 1/14,
  not 3/14 — do not rank it above the JSON-shape fix.
- Rec 4 (provenance ids): stands.
- Rec 5 (drop the grammar normalizer): stands (grammar refusals=0 live).

## 5. What I did NOT do

No code changes, no merges, no API calls, no DB writes (copy only), no live-system
interaction. All claims [A] from `v2.db` or the cited files at `04ae1c2f`.
