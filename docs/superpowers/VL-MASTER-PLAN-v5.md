# VL INTELLIGENT — MASTER PLAN v5
### 2026-09-10 · supersedes v4 · the complete plan: strategy, waves, evidence, verification

> **Import and governance amendment (2026-09-10).** This is the owner's supplied
> Master Plan v5, with the canonical-rulebook contract in §6 amended for PR #99.
> Other strategy, wave, runtime and measurement statements retain the original
> author's snapshot at `8941ec68`; this documentation import does not reverify them
> or assert that they describe the latest deployment. Original source: 20,229 bytes,
> SHA256 `9c14019acc78c7205a210c11f0acfdaef8459dc490e8059c1a51caa4cf454a3a`.

Running rev at writing: `8941ec68` (verified on dev — RELEASE and GUIDE_BUILT_REV agree).
Every basis document cited below returns HTTP 200 on dev at the time of writing.

---

## 0. THE STRATEGY, IN ONE PARAGRAPH

**The system is a level-fade book.** Before each session it marks ~12 price references
(prior-day H/L/C, overnight H/L, opening range, initial balance, VWAP bands, volume
profile, swings, supply/demand, order blocks, round numbers, prior-week extremes) and
writes 3–5 conditional scenarios. When price reaches a level and the confirmation
holds, it fades — sells the top of the range, buys the bottom — with a resting limit
at the level, a stop beyond it or 1.5×ATR5m whichever is wider, and a target at the
next opposing reference. One contract, one position, flat at 14:45 CT.

It is a **Market Profile map with ICT names and classic S/R execution** — three schools
stitched together. The audit (`2026-09-08-the-strategy.md`) found the play vocabulary
identical every day and the mix different every day (similarity 0.208–1.0 across 49
session-days). It has no day-type filter, no single-best-trade selection, no read of
the forming candle, and no named edge.

**What the record says:** 58 trades · 12 CME days · −$466 · 32% wins · expectancy
−$8.04, CI [−$35.55, +$22.04] · levels hold 48.8% of first touches, n=423,
[44.1%, 53.6%] · zero trades under the rules running today.

**What the research says** (six rounds, ~90 sources): no number you asked for exists
in the literature; the fade ranks weaker-tier among intraday index-futures strategies;
nothing intraday is proven at one-contract retail scale in this regime; the only
answer is **selection** — when to fade, which level, which entry, what geometry —
measured on your own record.

**So the plan is: build the record, then run the tests, then rule.** Not a new
strategy. The instrument first.

---

## 1. WHAT IS LIVE — verified on dev at `8941ec68`

| Shipped | What it does | Verified by |
|---|---|---|
| Wave A — the record | touch recorder formation-aware; excursions backfilled; exit cause from the broker event; accepted bracket recorded immutably | boot line `📐 record`, 4 proof lines quoted |
| Wave B — the long path | stop-entry trigger in the stop slot; stop-side guard un-inverted; class 77 (a long would have been submitted as a SELL) | NT8 log `Stop price=<trigger>`, hello frame h1 |
| cancel-confirmation | a cancel is confirmed by absence from a fresh broker book | boot line `🧾 cancels`, snapshot ids |
| one-contract invariant | working entries count as exposure per account | refusal line quoted |
| bracket separation | `HandleCancelOrder` cancels the entry only | E1 replay of the 09-06 sequence |
| placement truth | `place_pending` until a frame confirms; rejection in the broker's words | boot line `📤 place-confirm` |
| signal clock | all four outgoing commands stamped with the wall clock | payload age 5–9 ms vs NT8's 60 s guard |
| session calendar | shortened sessions trade; full closures don't; unknown dates closed and named | boot line `🗓 session calendar` |
| confirmation truth | a `5m close` requires a closed bucket; sequences check order; `1m_displacement` separate | 176 changed confirmations, 23+1 REJECT→PASS quoted |
| plan liveness | scenario death keyed by version + anchor; born-dead refusal (n=1); exhaustion WARN | boot line `🧭 plan liveness` |
| scenario economics | every new scenario states obstacle, response, target, both R; self-contradiction refused; sub-1R marked not refused; legacy reads UNKNOWN | boot line `📐 scenario economics`, 0/799 legacy refusals |
| Stage A snapshot | five research objects recorded with NULL semantics; `schema=1` | boot line `🗄 research snapshot` |
| bars-horizon | history tables declare held-vs-claimed; partial candles marked; the ring rehydrates 1m from the store on boot; regime window 7→9 days | boot line dates the change |
| W3 candidates | map kept whole; overlapping refs merge; entry candidacy needs a plausible target; shortlist by reachability; projections beyond the range; PWH/PWL from daily bars | boot line `🗺 map`, E7 golden byte-identical |
| W4 session risk | flat means flat (position + arms + pending, book-confirmed); breaker on the arm path; no-trade band on BOTH paths; boot line reads the BOUND strategy | boot line `🛑 session risk` |
| W5 truthful description | 14 false sentences corrected; 4 lived in a DB row; Guide clock pin fails on drift | E1 lunch-window pin RED→GREEN; live bundle quoted |
| desk strip | 12 lines, every one dated, UNKNOWN where unverifiable | first live dump quoted |
| dashboard v6 | market + equity together; Planner full-width; mobile columns | 5 Chromium groups |
| lock keeper | `acquire` spawns a bounded keeper; process-group kill; expiry enforced | 75/0, attacker's repro flips |

**Standing owner rulings:** guardrails master and daily_loss_enabled stay OFF by choice
(the $450 limit is decorative and the boot line says so) · BE/trail suspended · no
trade cap · one contract · no live money · findings are recorded and batched, never
dispatched one by one.

---

## 2. THE RESEARCH — six rounds, what each settled

| Round | Question | Verdict | What it changed |
|---|---|---|---|
| 10 | limit vs stop entry, adverse selection | **TESTED**: ~66% of passive NQ fills precede an adverse move (Lalor & Swishchuk 2024); fill-on-touch backtests biased (Lo–MacKinlay–Zhang) | keep the fade, model fills honestly, no stop-entry continuation |
| policy | targets, confirmation, ranking, staleness | **none prescribed**; break-even table (0.5R needs 67% wins; you run 32%); confirmation costs R; exclusion ≠ invalidation; four clocks | scenario economics; Stage A; the four experiments |
| 11 | what a fade book requires | seven requirements, you have one partially; a reaction ≠ a trade, a touch ≠ a fill, 50% ≠ the null; Wilson corrected to [44.1, 53.6] | W1 (episode), W2 (permission), W3 (candidates), W4 (risk) |
| 12–15 | higher-TF structure · forming candle · targets beyond the map · staleness | **every specific number UNTESTED on NQ**; `touch_episodes` field names mislead; the 1.2× multiplier has no foundation; no plan half-life exists | W-TF ships detection not weight; W-LIVE waits for 13f; 103's projections stay [I] |
| 16 | strategy landscape, 11 candidates ranked | **the fade is weaker-tier**; nothing intraday proven at one-contract scale; momentum with long holds is the best-supported alternative and it's mixed/decaying | keep it in SIM, filter it, measure fills, replace only if filtered still reads zero |
| NT8 | order lifecycle, OCO, cancel semantics | 13 states; `Accepted` is live; `CancelPending` can still fill; OCO is "any member cancels → all cancel"; local not exchange-side | Wave B, cancel-confirmation, bracket separation |

**The four `[T]` tests the research wrote for your own record** — 13f (the forming
candle, on `touch_episodes` with field corrections), 14b (continuation after 1.5× the
median range), 15e (plan drift), and 12e's frozen-candidate comparison — are the
experiments in section 4, made concrete.

---

## 3. THE WAVES — code, split, with detail and verification

Every wave: full hard-rules block inline · claim with composite id + expected files ·
measure first (eleven dispatches this week were refuted on their central premise) ·
RED before GREEN with real mutations · own fresh five-leg gate · RELEASE → mv → VERIFY
→ owner runs the kill · marker before lock release · report pinned by sha with byte
count · BACK UP → CHANGE → VERIFY.

### IN FLIGHT

**W1 — THE EPISODE CONTRACT** (101, `fix/episode-contract`, boots 14:45 today)
*What:* the unit an experiment needs. Extends `touch_outcomes` with a labelled
scenario link (`ScenarioNearest` + `ScenarioLinkBasis`: price_proximity with distance,
or NULL with the reason — never a column named `scenario` that reads as fact), the
five-rung opportunity outcome (never_reached / reached_declined / confirmed_not_armed /
armed_not_filled / filled — later rungs imply earlier, so a fill claiming never-reached
is inexpressible), the attainable entry, the session-end closer, and a three-state
backfill.
*Refuted twice:* the join I said didn't exist partly does (touch_outcomes already
carries ordinal, k, Δ, band, horizon, verdict); then the "1,205-row join" was a cross
product (481 × 9 × 7). C1 stood: `plan_id` is the only common column.
*The finding:* the backfill recomputes **zero** — 96.2% of in-era rows have no
formation time, and the scenario link is new. The historical record cannot answer the
per-opportunity question because the inputs were never written down. That is the
research's central claim, measured.
*Verify:* boot line `🎫 episodes: … backfill recomputed=0 unrecomputable=4860 · k=3[I]
Δ=resolved-per-read (kernel.MeanAbsIncrement) H=12[I]` — the resolver named, never
the value; the boot-line pin moves the resolver and requires the line to move; first
episode opened, first transition, first close with cause.

### NEXT — three lanes free

**leg-4 working-vs-armed + one terminal predicate** (102, PR #97 pending its sweep fix)
*What:* the cutover gate compares WORKING rows (signal_id set) to the broker book;
ARMED rows (no signal_id) are counted on their own line and never fail the leg;
`place_pending` counts as working. One exported SQL fragment derived from
`isTerminalArmState`; a source guard fails any retyped state list anywhere.
*Why:* ruled 09-06, blocked four cutovers on 09-09 — two lanes retyped the state list
in opposite directions the same hour (one over-reported 11 dead arms, one
under-reported a live `place_pending` to zero).
*The regression it must not ship:* the boot sweep's three-state list is DELIBERATE —
`ConfirmCancel` is the only path to `cancelled` and needs a snapshot id; sweeping
`cancel_pending` bypasses it. A/B proven. Fix: `SweepableArmStateSQL` = non-terminal
MINUS cancel_pending, named, with the reason in a comment. *A single source, never a
single predicate* (class 107).
*Verify:* 2 armed + broker 0 → PASS · place_pending + broker 0 → FAIL · 0 working +
broker 1 → FAIL · a cancel_pending row at boot is NOT swept and IS picked up by
confirmPendingCancels · a retyped list anywhere → grep fails.

**settlement-and-flat-truth** (104, Section C already pinned by sha)
*What:* the five findings 104 filed on 09-09, verbatim: (1) `ConfirmCancel` has never
fired — 0 of 4 cancelled rows via the confirmed path; the ack-timeout branch promotes
straight to `cancelled`; (2) every flatten reads "no open position" from the local
store, documented as ~80 s behind the broker, while `liveBook`/`brokerBook` already
exist in-process; (3) the no-link fallback writes `cancelled` with zero broker contact;
(4) "acked" means left-the-set, so a fill during the drain is logged as a cancel;
(5) the flatten is the last unguarded cancel site. Plus: the cancel attempt budget
survives the restart that invalidates it.
*The thesis, in 104's words:* four of the five are a word in our ledger written from
something other than the broker's answer.
*Verify:* the first cancel that settles through `ConfirmCancel` with its snapshot id;
a flatten that reads the broker; a fill during the drain recorded as a fill; the
attempt counter resetting on boot; 0 unguarded cancel senders in the census.

**W-TF — EVERY DETECTOR, EVERY TIMEFRAME** (103)
*What:* the owner's standing rule the system never honoured. Swing, supply/demand,
OB, FVG, equal-highs detectors run on every timeframe the store holds (1m–1w); every
level carries its timeframe as identity; a 1d swing and a 5m swing at one price MERGE
(W3's D2) into one candidate with both names; the map stays whole. **No grading
change** — the HTF 1.2× stays [I] because round 12 found no foundation for it.
*Research law:* round 12's 12e — hold the family definition constant, distinguish
bar-count parameters from time/volatility parameters, test a timeframe hierarchy
rather than presume one. Detection is [O]; weighting waits for E4.
*Verify:* a daily swing appears on the map with `tf=1d`; the same price on 5m and 1d
merges to one candidate, two names; the score golden is byte-identical; the model's
level table shows the timeframe column; boot line `🗺 map: … tfs=[1m 5m 15m 1h 4h 1d
1w]`.

### AFTER W1 LANDS

**W2 — FADE PERMISSION** (the wave the research endorses first)
*What:* is the book allowed to fade right now? Not a day classifier — round 11 says
none is established for MNQ. The benchmark it recommends: **all qualifying fades vs a
few pre-declared exclusions**, and a classifier only if it beats that on held-out
days. Ships as a **label and a counter** on the desk strip and the plan card
(`fade_permitted=yes/no` with the reason), never a gate until E3 shows the exclusions
earn their keep. Exclusions are pre-declared, labelled [I], and recorded per episode —
which is why it needs W1.
*Candidate exclusions (all [I] until measured):* opening range > k× the 20-day median ·
IB broken and held by 09:30 · price beyond every mapped level in its direction (103's
projection state) · a T1 news window · the first N minutes.
*Verify:* the label renders on every scenario; the counter splits episodes by
permitted/not; nothing is refused; E3 reads it.

**THE IDENTITY WAVE** (after W1; needs 103's merged candidates)
*What:* `PlanScenario` names its level by candidate id, not a price. Turns W1's
labelled heuristic into identity. Must solve formation capture alongside naming —
`researchCandidateID` already exists but aliases where `formed_at_ms` is absent (96%).
*Verify:* a scenario's level resolves by id; two scenarios on one merged candidate
share it; formation present on new rows approaches 100%.

**W-LIVE — REACT TO THE FORMING CANDLE** (after 13f's test)
*What:* the trader watching the bar form at the level. The executor reads
`touch_episodes`' live fields — with round 12's corrections: `wick_pen_pts` is max
penetration not a wick; `vol_ratio` grows with duration; `approach_atr` is
dimensionless. Ships only after 13f says which field predicts, at what horizon,
against what null.
*Verify:* 13f's eight-step test run on the 1,688 rows first; a shadow signal recorded
per episode before any rule acts.

### CLEANUP — one batch, one lane, between trading waves

Batch 1 (recorded): sweep predicate by name · cancel attempt budget resets on boot ·
`handler_svp.go:50` comment ("cache cap" = 2000; cap is 2500; 2000 is AISVPBarCount
and correct — the comment is wrong) · `boot_sweep.go:38` comment (claims four states,
SQL lists three) · CLAUDE-canon mirror contract test (done) · classes 88–91 (void —
they exist in the second heading format).
Batch 2 (recorded): the unapplied-mutation harness pin (three "survivors" were
`sed`s that never landed) · `git worktree add` exit-code check before `cd` (a lane
committed onto another lane's branch) · class 99's Law corrected to point at 107.
**Rule:** new findings go into the next batch. No box the same hour.

### HELD — until E3 rules

Rebrand phases 3–5 · BE/trail · sizing above one contract · live money.

---

## 4. THE EXPERIMENTS — read-only, on the record

Run after Stage A + W1 have ~20 sessions. Each: paired on the same episodes,
chronologically split, pre-declared minimum effect, Reality Check (Sullivan–
Timmermann–White: a best-rule p of 0.042 became 0.908 under one). Each returns
SUPPORTED / CONTRADICTED / **INCONCLUSIVE**. None touches the bot. Lanes 01/02.

| # | Question | Held fixed | Varied | Answer decides |
|---|---|---|---|---|
| E1 | entry type | opportunity set, stop, target | resting limit · rejection candle · failed break · reclaim; queue-aware fills; never-confirmed = zero-trade outcome | whether the resting limit survives adverse selection |
| E2 | stop, target, horizon — jointly | entries | stop beyond level / wick / ATR / time × target structural / fixed-R / none × horizon | the 1.5× floor (one arm, not the baseline); target policy |
| E3 | fade permission | everything else | W2's exclusions vs no exclusions, held-out days | whether W2 becomes a gate; whether the fade has selected expectancy > 0 |
| E4 | ranking vs distance | universe, downstream policy | score terms, seat rule, vs nearest-N | whether the ladder beats distance alone; the 1.2× and the confluence bonus |
| E5 | session risk | — | worst-session contribution, runs of losses, time under water, slippage beyond stops | W4's numbers instead of guesses |

Plus the research's own `[T]` tests: **13f** (forming candle) → W-LIVE; **14b**
(continuation after 1.5× range) → 103's projections; **15e** (plan drift) → the
staleness question.

---

## 5. THE ORDER, AND WHY

```
today      101 W1 boots 14:45
now        102 leg-4 (once PR #97's sweep fix lands) · 103 W-TF · 104 settlement
after W1   W2 fade permission (label) · identity wave
cleanup    batch 1 between waves, one lane
~20 sess.  E1 · E2 · E3 · E4 · E5 · 13f · 14b · 15e — parallel, read-only
then       W2 → gate IF E3 · stop floor moves IF E2 · ranking changes IF E4 ·
           W-LIVE IF 13f · out-of-map targets IF 14b — each a separate ruling
```

**Nothing in the trading waves changes what trades**, except leg-4 (which changes
what BOOTS) and W-TF (which adds levels to a map the model already sees). Everything
that would change a trade waits for its experiment. That's the research's law and
it's yours.

---

## 6. VERIFICATION — how every claim in this plan is checked

- **"Live" means:** the boot line quoted with its rev, all five references agreeing
  (RELEASE · binary vcs.revision · HEAD:deploy/RELEASE · GUIDE_BUILT_REV ·
  /api/health), the marker on dev, and dev read **pinned by commit sha, never a branch
  path** (a branch URL served a superseded revision for five minutes on 09-06).
- **"Proven" means:** a live line, not a green suite. Every wave names what is still
  unproven at closeout.
- **"Measured" means:** n and row ids on every figure; `pnl_corrected` only; NULL is
  UNRESOLVED and counted; a zero and an unknown never read alike.
- **"Refuted" means:** the agent measured the premise at the running rev and it was
  wrong. Eleven this week. Each was cheaper than the build it prevented.
- **A pin is a pin only if** it was RED first, and its mutation is quoted with the
  line it changed — a passing mutation, or one whose `sed` never applied, proves
  nothing.
- **The checklist ceiling** is read with a `uniq -c` census across both heading
  formats; the number is assigned at merge, never reserved at accept.
- **The canonical rulebook** is [VL-TRADING-RULEBOOK-v1.md](VL-TRADING-RULEBOOK-v1.md),
  the corrected document in [PR #99](https://github.com/johnwick2921-cyber/nofx/pull/99).
  On merge, it supersedes the owner's supplied v1; the retained filename does not
  make the supplied v1 a second authority. Every wave must agree with this rulebook.
  Its **five sections** are **A. WHAT RUNS**, **B. WHAT IS WANTED**,
  **C. WHAT IS STILL TO BUILD OR PROVE**, **D. HISTORICAL RECORD**, and
  **E. TRADER'S RECOMMENDATIONS**. The sources appendix is not a sixth policy section.
- **A wave touching a trading rule** updates **A. WHAT RUNS** in the **same commit**
  as the rule change, with the affected code-file and line references and the
  implementation revision. It **never updates B. WHAT IS WANTED to represent what
  runs**: B records owner intent, C records unfinished or unproven work, D records
  population-specific evidence, and E records labelled recommendations. A change
  committed but not yet verified live must be identified as such; updating the
  rulebook does not substitute for the live-proof requirements above.

---

## 7. THE LIVE-MONEY GATE — unchanged

n ≥ 100 closed trades across ≥ 40 active CME days under ONE unchanged policy · both
95% lower bounds > 0 after documented costs · max drawdown ≤ $900 · worst day ≤ $450
· fee schedule verified · stop/cancel/reconnect/kill drills passed. Every line fails
today. E3 is the first experiment that can move any of them.

---

## 8. THE OWNER'S ITEMS

```
TODAY    run the kill 101 prints at 14:45
DAILY    NT8 up before each session
RULINGS  send 103 (W-TF) and 104 (settlement) now, or hold for W1's boot
         — my call: send both; they don't touch W1's files
AFTER    W2 when W1 lands · cleanup batch 1 when a lane frees
E1–E5    after ~20 recorded sessions — not before
```

---

## 9. THE SHAPE OF IT

v3 said stop shipping rules and start recording. v4 said the strategy is a level fade,
real, nameable, and unselected. v5 says: **the record exists as of today's boot, the
strategy is written down in a trader's words, six research rounds have said what is
and isn't known, and the only remaining work is to build the last three recording
pieces and let the record answer.**

Seven waves, five experiments, one rulebook. Then you rule.
