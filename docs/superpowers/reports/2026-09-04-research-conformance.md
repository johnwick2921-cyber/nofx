# RESEARCH-CONFORMANCE RE-CHECK — every live rule against the research verdicts

**Window** boot 8, rev `70af663d`, booted 2026-09-04 08:30:11 CT (PID 878451) · SIM-only, MNQ
**Branch** `docs/research-conformance-0904` · claim `fb50903f` · base `492d2067` (dev tip, boot-8 marker)
**Method** READ-ONLY. Resolved values are taken from the running process's own boot lines or the
resolver path — never a file default. `/api/config/resolved` and `/api/risk/gate-blocks` both
return `{"error":"Missing Authorization header"}` from this session (A11 stated, not guessed).
**Evidence classes** **[A]** verified · **[B]** inferred · **[C]** speculation

---

## 0. THE HONEST LINE AT THE TOP

**[A] The dispatch conflates two different populations, and the headline is only meaningful once
they are separated.**

| population | size | what it is |
|---|---|---|
| **live knobs** | **144** | `KnobLive` entries in `store/knob_registry_table.go` (167 total: 144 live · 16 candidate · 7 ineffective) — schema leaves with a proven consumer outside `store/` |
| **labelled beliefs** | **49** | the rules labelled [R]/[X]/[T]/[I]/[O] in the 2026-09-02 belief census (A1-A11, B1-B10, C1-C8, D1-D10, E1-E5, F1-F5) |

A knob is a *number with a reader*; a belief is a *rule with a claim about the market*. The census
never labelled the 144, and the registry never labelled the 49. **"How many of the 144 live rules
are [R]/[T] vs [I]" cannot be answered as posed** — the 144 have no labels. What can be answered,
and is, below: the label distribution of the 49 beliefs then vs now, and the demotion-queue status.

```bash
grep -c 'Status: KnobLive'        store/knob_registry_table.go   # 144
grep -c 'Status: KnobCandidate'   store/knob_registry_table.go   # 16
grep -c 'Status: KnobIneffective' store/knob_registry_table.go   # 7
git log -1 -- docs/superpowers/reports/2026-09-02-belief-census.md
# ee64a494 2026-09-02 08:50:38 -0500
```

---

## 1. PREMISE CORRECTIONS (A17 — measured before anything was concluded)

| # | dispatch said | measured | consequence |
|---|---|---|---|
| P1 | "trade_excursions rows now exist — quote MAE/MFE p50/p80" | **0 rows** (`SELECT COUNT(*) FROM trade_excursions` → 0) | D3 is answered from `trader_positions.mae/mfe` instead, and that substitution is stated wherever a number rests on it |
| P2 | "touch_outcomes=135+" | **359 rows** (candidate_pool 192), first written 2026-09-04 05:30:54 CT; the boot line's 293/168 was the count at 08:30:11 | D5 has more data than the dispatch assumed |
| P3 | "research rounds 1–9 (docs/superpowers/research/…)" | that directory holds **only `INDEX.md`**; every report lives in `docs/superpowers/reports/` | claims attributed to "round N" are cited to a named report or marked *source could not be located* |
| P4 | "the findings ledger §1 of VL-MASTER-PLAN-v2.md if present" | **not present on dev** | — |
| P5 | "the arm rules (D1–D5 of arms-follow-bias)" | **no report by that name exists on dev** | the arm rules are enumerated from code and the missing source is named |
| P6 | "settings registry: live=144" | **correct** — 144 `KnobLive` of 167 entries | the only dispatch premise about counts that held exactly |

---

## 2. D3 — EXITS: the stop floor against the measured pullback

### The floor, resolved

```
09-04 08:30:11  🛑 exits: stop=max(anchor+clr, 1.5×ATR5m) · anchor_max=3.0×ATR5m
                          · BE=off · trail=off · size=1 · re-arm-after-sweep=on (0B)
09-04 08:30:11  🛑 min-sl guard: atr_mult=1.5 level_clearance=2tick(s)
```

`kernel/min_sl.go:34 MinSLATRMultDefault = 1.5` · `MIN_SL_ATR_MULT` is unset in `.env`, so the
resolved value **is** the code default — stated, not assumed.

### The measured distribution

Era = `DayPlanEraStart` 2026-08-15 00:00 CT (`store/attribution.go:131-158`). Closed positions in
era: **n=71**, excluded 4 NULL `pnl_corrected` · 9 `UNRESOLVABLE` · 3 `e7_farside_test`.
Clean set with MAE/MFE present: **n=58**.

| percentile | MAE (pts) | MFE (pts) |
|---|---|---|
| p50 | **32.75** | 25.75 |
| p80 | **49.25** | 69.25 |

**[A] But this MAE distribution CANNOT answer the question the dispatch asks, and reporting it as
if it could would be the error.** MAE on a stopped-out trade is bounded by the stop we placed, so
it measures *our stop placement*, not *the market's pullback*:

| outcome | n | mean MAE | mean realized move | MAE within 2pt of the move |
|---|---|---|---|---|
| LOSER | 38 | 42.01 | 35.85 | **20 of 38 (52.6%)** |
| WINNER | 20 | 15.35 | 56.46 | 1 of 20 |

Over half of losing trades have MAE equal to their own realized move to within 2 points — they
went straight to the stop, and MAE *is* the stop distance. **The MAE p50 of 32.75 sitting almost
exactly on the recorded median `stop_floor_pts` of 33.2 (n=25 planner reads) is what censoring
predicts, not evidence of calibration.**

### The uncensored number, and what it says

The decision-relevant quantity is how much adverse room a **winning** trade actually needed:

| MAE on winners | value |
|---|---|
| n | **18** |
| p50 | **11.5 pts** |
| p80 | **22.5 pts** |
| max | 61.5 pts |

**n=18 is below the pre-registered floor of 30, so NO VERDICT is rendered (A24).** The numbers are
reported as-is. For scale only: the recorded median stop floor is 33.2 pts — **2.9× the winners'
p50 and 1.5× their p80** — and this morning the floor resolved to 44.59 and 50.73 pts against
planner-authored stops of 35.00–38.50.

**Answer to D3 as posed:** the floor sits *outside* the pullback that winning trades experience,
not inside it. The dispatch's framing assumed the risk was a floor too tight; the measured
direction is the opposite. This is stated as a measurement, not a ruling.

### The 1A wave has still not delivered its instrument

`docs/superpowers/reports/2026-09-03-trade-excursions.md` (`git log -1` → `0c1a808c 2026-09-03
00:05:11 -0500`) specified `trade_excursions` precisely so this question could be answered from
path data rather than endpoint data. **The table is still empty**, so this section rests on the
censored substitute. That is the finding, not a footnote.

---

## 3. D1 — the demotion queue, item by item

`docs/superpowers/reports/2026-09-02-belief-census.md:127-137` (`ee64a494`) ranked nine [X]/[I]
beliefs with live teeth. Status at boot 8:

| # | queue item | asked for | status at boot 8 | verdict |
|---|---|---|---|---|
| **1** | min-SL 1.0×ATR5m + 2t, [I], REJECT at write | demote to WARN-first; sweep the multiplier | **still a hard REJECT** (`kernel/engine_position.go:227-231`, `return fmt.Errorf`); the multiplier went **1.0 → 1.5** (`kernel/min_sl.go:34`) | **NOT DONE — moved the wrong way** |
| 2 | swing seats improve turns, **[X]** | re-test; likely unseat | seats still live (boot: `seats=8`) | see §4 |
| 3 | level-event wakes deserve a re-read, [T]-weak | WARN-first N=25 | boot line: cutoffs `enforce` govern LEVEL_EVENT — but that is the *cadence* cutoff, not the E5 demotion | see agent unit `cadence` |
| 4 | BD_MAX_PULLBACK / BD_CONFIRM_CLOSES, [I] | WARN-first | — | see agent unit `validator` |
| 5 | stale-confirm 2.0×ATR5m, [I] | WARN-first | boot: `stale_confirm=2.0×ATR5m` — **unchanged** | **NOT DONE** |
| 6 | zoneTFMult + ladders, [I] | sensitivity sweep first | boot: `zone-ladder=1.0/0.6/0.3/0.15` — unchanged | see agent unit `levels` |
| 7 | pre-NY sessions carry edge, **[X]**-candidate | owner decision on session enablement | ASIA/LONDON still run (registry `enabled:false` is only the default; the resolver overrides — proven by positions 587 ASIA / 588 LONDON) | **NOT DONE** |
| 8 | consumed/3rd-touch → never entry, [I] | measure react rate | touch telemetry advisory; `touch_outcomes` now n=359 — see §5 | data now exists |
| 9 | killzone/premium-window + Monday/Thursday conviction, [I] | keep advisory, add n counter or delete | prompt lines still present | advisory (no teeth) |

### The one thing that DID move on item 1 — and it is not the demotion

A WARN-first pre-check exists and fired four times this morning:

```
09-04 08:09:03 ⚔️ arm feasibility: S1 arm stop 29520.00 too close (38.50 < 50.73 = 1.5×ATR5m)
                 — min-SL gate will refuse it (WARN — write proceeds; the gate-at-arm chain enforces)
09-04 08:09:03 ⚔️ arm feasibility: S2 arm stop 29620.00 too close (35.00 < 50.73 = 1.5×ATR5m) …
09-04 08:09:03 ⚔️ arm feasibility: S4 arm stop 29240.00 too close (38.00 < 50.73 = 1.5×ATR5m) …
09-04 08:25:34 ⚔️ arm feasibility: S2 arm stop 29518.00 too close (36.50 < 44.59 = 1.5×ATR5m) …
```

**[A] This is NOT the demotion the queue asked for.** It is `trader/auto_trader_planner.go:1671-1682`,
shipped by the **london-forensics F4 wave on 2026-08-28** — five days *before* the census ranked
min-SL #1. Its own text says the gate still enforces. The census knew about it and asked for the
gate itself to become WARN-first; that did not happen.

**And it exposes a second-order finding [A]:** the planner is systematically authoring stops
*below* its own floor — 35.00, 36.50, 38.00, 38.50 pts against floors of 44.59 and 50.73. Four
scenarios in one morning were written knowing the gate would refuse them.

---

## 4. D1 (second half) — every [X] belief still enforced

| belief | label source | still live? | teeth |
|---|---|---|---|
| **D9 — swing seats improve turn capture** | census D9: [X] on own tape, `grand-audit.md:74` (missed-turns 80.0/75.0/79.2% → 65.0/60.0/66.7%, Δ −15pts) | **YES** — boot 8 `volume wave: … seats=8` | seated **weight** on every level grade |
| **queue #7 — pre-NY sessions carry edge** | census: [X]-candidate, own tape 0/6 −$353.5 pre-NY vs NY 3/3 +$177 (weekend P2 audit) | **YES** — ASIA and LONDON both run | session **enablement**; today's only two arm-enabled scenarios were both LONDON |

**[A] Two [X]-labelled beliefs still carry live teeth.** Neither is a REJECT or a MUST — one is a
weight, one is an enablement — so the strict reading of D1 ("every [X] still enforced as a REJECT
or MUST") is: **none**. The honest reading is: **two**, with weight/enablement teeth.

---

## 5. D5 — levels and the D1′ detector

### D1′ live rates, with n

`touch_outcomes` n=**359** (first row 2026-09-04 05:30:54 CT — the hook began firing the morning
after the two-day audit predicted it would).

| outcome | n | share |
|---|---|---|
| hold | 137 | 38.2% |
| break | 129 | 35.9% |
| **ambiguous_horizon** | **93** | **25.9%** |

Per level kind, **suppressing every cell below n=30** (A24):

| level_kind | n | hold | break | ambiguous |
|---|---|---|---|---|
| VWAP | 108 | 42 (38.9%) | 23 (21.3%) | **43 (39.8%)** |
| DEMAND | 72 | 36 (50.0%) | 32 (44.4%) | 4 |
| RTH-L | 55 | 16 (29.1%) | 35 (63.6%) | 4 |

15 further kinds hold 124 rows between them, all below n=30 — **no rate is quoted for any of them.**

**Surprise (A23, included not acted on):** VWAP resolves *ambiguous* 39.8% of the time at horizon
H=12. Two of every five VWAP touches never resolve inside the detector's window.

### Is D1′ still DESCRIPTIVE ONLY? — yes [A]

Every production reader of `touch_outcomes` / `candidate_pool` is a **writer**, a **boot counter**,
or a **reporting CLI**:

| reader | file:line | kind |
|---|---|---|
| episode + pool writer | `trader/detector_record.go:50,83,101` | WRITE |
| boot-line counter | `main.go:404` → `store/touch_outcomes.go:233` | COUNT |
| offline report | `cmd/detector-report/main.go:42,44` | CLI |

```bash
grep -rn "TouchOutcomes\|CandidatePool" --include=*.go kernel/ trader/ api/ | grep -v _test
# → detector_record.go (writer), detector_d1prime.go + detector_recorder.go (producers). No gate.
```

**Zero gates, zero weights, zero scorers read either table.** D1′ conforms to "descriptive only".

### A defect the new data confirms

**[A] `touch_outcomes.candidate_seated` is degenerate — all 359 rows carry `1`, none carry `0`.**

```sql
SELECT candidate_seated, COUNT(*) FROM touch_outcomes GROUP BY candidate_seated;  -- 1 | 359
```

The column cannot distinguish a seated candidate from an unseated one, so every future analysis
keyed on it is vacuous. This confirms the two-day audit's C6-F2 finding
(`docs/superpowers/reports/2026-09-04-two-day-audit.md`, `git log -1` → `f3c640c3 2026-09-04
07:26:52 -0500`) against 359 live rows rather than against the code alone.

**And `candidate_pool` records exactly one cut reason:**

| seated | cut_reason | n |
|---|---|---|
| 0 | `max_levels: 12 seated, cap 12` | 97 |
| 1 | *(empty)* | 95 |

Every cut in the window is the seat cap. Any other reason a candidate might be dropped is invisible
in the table.

---

## 6. D10 — expectancy: the data forbids both a promotion and a demotion

**The pre-registered criterion, verbatim** (`2026-09-03-expectancy-1d.md:25-28`, `git log -1` →
`38a63a9b 2026-09-03 15:26:02 -0500`):

> **Floor.** `n < 30` → `DESCRIPTIVE ONLY`, no verdict rendered. At or above it the status is
> computed from the pre-registered rule: **PASSES ⟺ n ≥ MinN ∧ mean > 0 ∧ mean interval excludes 0**,
> else FAILS.

`MinN = 30` — `expectancy/model.go:27`.

**The live boot line, today, both boots:**

```
09-04 07:38:40  📊 expectancy: cells=41 with_n>=30=0 judged_rollups=2 unresolved=3 excluded_test=3
09-04 08:30:11  📊 expectancy: cells=41 with_n>=30=0 judged_rollups=2 unresolved=3 excluded_test=3
```

**[A] D10(a): no cell at n≥30 has a status differing from the criterion, because ZERO of the 41
cells reach n≥30.**

**[A] D10(b): no promotion or demotion is justified — the data forbids both.** Every cell is below
the floor; the criterion renders `DESCRIPTIVE ONLY` and no verdict, and this audit renders none.

### Era expectancy, rebuilt from the store

`pnl_corrected` only; era ≥ 2026-08-15 00:00 CT; `UNRESOLVABLE` and `e7_farside_test` excluded.

| | value |
|---|---|
| n | **58** |
| total | **−$466.43** |
| mean | **−$8.04** |
| sd | $103.98 |
| **95% CI** | **[−$34.80, +$18.72]** |
| win rate | 31.0% |

Against the MC rig (`2026-09-03-mc-drawdown.md`, `git log -1` → `77e1cdfc 2026-09-03 00:39:25
-0500`), whose verdict line reads *"expectancy indistinguishable from zero (CI −$31…+$18, n=64)"*:

- **The CI still straddles zero** — unchanged conclusion.
- **n did not grow; it is LOWER (58 vs 64).** The rig's population and this rebuild's differ by 6
  rows. I could not reconcile the two exclusion sets from the report text, so **the discrepancy is
  reported, not resolved** — one of the two sets is wrong and it needs an owner or a follow-up to
  say which. It does not change the verdict: both straddle zero.

By condition roll-up:

| plan_band | n | Σ | mean | W/L |
|---|---|---|---|---|
| *(none)* | 49 | −$508.93 | −$10.39 | 13/34 |
| `armed_fill` | 9 | +$42.50 | +$4.72 | 5/4 |

Both below n=30 → **DESCRIPTIVE ONLY, no verdict.**

---

## 7. D7 + D11 — arms, and whether a long can trade today

### D7 — the arm-enablement asymmetry, re-measured

The two-day audit (`f3c640c3`) found the planner grants resting arms almost only to shorts.
Re-measured on the live table today:

| trade_date | side | arm-enabled | scenarios | rate |
|---|---|---|---|---|
| 2026-09-02 | long | 9 | 64 | **14.1%** |
| 2026-09-02 | short | 27 | 43 | **62.8%** |
| 2026-09-03 | long | 2 | 28 | **7.1%** |
| 2026-09-03 | short | 8 | 20 | **40.0%** |
| **2026-09-04** | **long** | **1** | **5** | **20.0%** |
| **2026-09-04** | **short** | **1** | **5** | **20.0%** |

**[A] The asymmetry is absent today — 20.0% vs 20.0%, exact parity.** But **n=5 per side**, far
below any verdict floor, so this is a reading, not a finding: one more plan version either way
flips it. **No verdict (A24).**

**A correction to my own prior number [A].** The two-day audit reported 09-03 as long 1/23 and
short 8/18. The live table now reads long 2/28 and short 8/20, because `plans` kept growing after
that snapshot — 09-03 now holds **16 versions**, latest written 2026-09-04 01:20:35 CT (the ASIA
session runs to 02:00). The audit's figure was correct at its snapshot and is superseded here.

### D11 — can a long plan trade today?

**Both routes traced.**

Plans written on 2026-09-04:

| session | v | created (CT) | bias | day_type | lifecycle | long | long armed | short | short armed |
|---|---|---|---|---|---|---|---|---|---|
| LONDON | 1 | 01:41:15 | **long** | **trend** | active | 3 | **1** | 1 | 0 |
| LONDON | 2 | 02:04:14 | short | trend | active | 0 | 0 | 2 | **1** |
| LONDON | 3 | 07:56:17 | short | balance | active | 1 | 0 | 2 | 0 |
| NY | 1 | 08:11:45 | neutral | **no-trade** | `no_trade` | 1 | 0 | 0 | 0 |

**[A] Yes — a long plan CAN arm, and one did.** LONDON v1 at 01:41:15 CT was long-biased,
`day_type: trend`, and carried an arm-enabled long scenario. So the arm route is open to longs;
the two-day audit's asymmetry was a planner *tendency*, not a structural block.

**[A] But nothing traded. Zero `armed_orders` rows and zero positions exist for 2026-09-04.**

```sql
SELECT COUNT(*) FROM armed_orders   WHERE created_at CT >= '2026-09-04 00:00';  -- 0
SELECT COUNT(*) FROM trader_positions WHERE entry_time >= 2026-09-04 05:00 CT;  -- 0
```

Two arm-enabled scenarios were authored and **neither reached the ledger**. The log says why:

```
09-04 07:40:44  ⚔️ arm S1 leg 1 wait_confirm MET (touch) — arming
09-04 07:40:44  ⚔️ arm REFUSED LONDON S1 leg 1: R:R 1.07 below arm min 2.00
                  (studio min_risk_reward_ratio) · rr refusals this session: 1
```

then the confirm re-fired seven more times (07:42:39, 07:45:01, 07:46:39, 07:48:39, 07:50:38,
07:52:38, 07:55:01) without ever producing an arm.

**[A] The binding constraint on the arm path today was the R:R-at-arm floor of 2.00, at R:R 1.07** —
and the two-day audit established that `composeArmStop` widens the stop *before* R:R is scored, so
the floor is applied to a stop the planner did not author. The min-SL floor and the R:R floor
compound: the stop is widened to clear min-SL, which lowers R:R, which then fails the R:R gate.

**Counts since boot 8 (08:30:11 CT) — the window the dispatch asks about is ~13 minutes long:**
0 plans, 0 arms, 0 positions. The D11 answer therefore rests on the full 09-04 day, and says so.

### Log-line counts, 2026-09-04 (whole day so far)

| line | count |
|---|---|
| `⚔️ armed` (a real arm placed) | **0** |
| `arm REFUSED` | 1 |
| `🚦 entry-gate` (arm-path refusal) | 0 |
| `≈armed` (display-only estimate) | 7 |
| `wait_confirm … arming` | 8 |
| `MIN-SL REJECT` (decision path) | 0 |
| `⚔️ arm feasibility` WARN | 4 |
