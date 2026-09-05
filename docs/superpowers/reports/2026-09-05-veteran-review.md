# VETERAN REVIEW — LEAD REPORT (INTERIM)

**Owner:** hoang · **Date:** 2026-09-05 · **Lead agent**
**Status: INTERIM.** Sub-agent parts A–D were dispatched and had not delivered when this
was pushed. Sections 1–8 are therefore RESERVED. What follows is the one-page summary and
the lead's own sections 9–10, resting **only** on numbers I computed myself from data
committed in this repository. Nothing here is quoted from a sub-agent.

---

## EVIDENCE BASIS — read this before any number

This review ran in a **fresh cloud clone of the repository**, not on the owner's machine.
I verified, before dispatching anything:

- **No running engine.** `curl http://localhost:8080/api/health` → connection refused;
  nothing listening on any port. `/api/health`, `/api/expectancy` and
  `/api/config/resolved` were **not reachable** and were not called.
- **No SQLite store anywhere in the container.** Every store query named in the dispatch
  (`touch_outcomes`, `trader_positions`, `plans`, `bars`, `trade_excursions`,
  `planner_rejected_prompts`, …) was **unexecutable**.
- **No tape and no `~/nofx-analysis/`.** **No replay was performed.** Any statement in this
  report about what "would have been armed or filled" is reasoning over code and committed
  data, never a simulation.
- `docs/superpowers/plans/VL-MASTER-PLAN-v2.md` **does not exist** in the tree.
  `docs/superpowers/research/` contains only `INDEX.md`.

What rescued the review is that prior waves **committed their data**. Everything numeric
below was recomputed by me from files in this repo:

| Source | Path |
|---|---|
| the 1E rig + its input | `docs/superpowers/reports/2026-09-03-mc-drawdown-data/{mc_drawdown.py,trade_sample.csv,day_sim.csv,drawdown_paths.csv}` |
| level-touch outcomes | `docs/superpowers/reports/2026-09-04-research-conformance-data/D5b-touch_outcomes-by-{kind,ordinal,session}.csv` |
| the arms funnel | `docs/superpowers/reports/2026-09-04-two-day-audit-data/arms.csv` |
| planner shape | `docs/superpowers/reports/2026-09-04-two-day-audit.md:279-292` (SQL shown in the report) |

**Research labels** follow the house legend (`SYSTEM-MAP.md`, from `2026-09-02-belief-census.md:8-16`):
`[R]` paper/study named · `[T]` own-tape number with n · `[I]` my experience, untested here.

**Two dispatch requirements could not be met and are not claimed:** the report is on
`claude/vl-veteran-review-0905-qnhnt0` (this session is pinned to that branch), **not**
`docs/veteran-review-0905`; and there is no reachable `dev` host, so **no `curl 200`
verification was performed**.

---

## ONE-PAGE SUMMARY

**Verdict on the system as it stands.** The engineering is better than the trading. The
audit discipline here — every knob carrying a citation or a suspension, every belief
labelled `[R]/[X]/[T]/[I]/[O]` — is stronger than most professional desks I have worked on,
and I mean that. But it is scaffolding around a strategy whose central premise is
**unmeasured on its own tape, and absent where it has been measured.** The system fades
levels. Its own `touch_outcomes` say levels hold **48.8%** of the time on first touch. That
is a coin flip, and you cannot build a fade book on a coin flip. Meanwhile the machinery
that would express any edge is leaking badly: **only 20% of arms ever fill**, and a third
of them die because price reached the level *too fast*. This is not ready for live money,
and — importantly — **the system's own Stage-A rule already says so.** I would not change
that rule. I would change what it is measuring.

**The three biggest problems.**

1. **The fade premise has no measured edge `[T]`.** First-touch hold rate **118/242 =
   48.76%, Wilson 95% [42.5%, 55.0%]** — 50% sits comfortably inside. Of 18 level kinds,
   exactly one excludes 50%, and it excludes it *on the wrong side for a fade*: RTH-L holds
   **20/63 = 31.75%, Wilson [21.6%, 44.0%]**, i.e. the prior-day RTH low **breaks 68% of
   the time**. The most-seated kind, DEMAND, is **44/85 = 51.8%, Wilson [41.3%, 62.1%]** —
   a coin flip with a wide interval. The one kind that looks tradeable, VWAP
   (**43/70 = 61.4%, Wilson [49.7%, 72.0%]**), barely clips 50% *and* **38.6% of its
   touches were unclassifiable**, so what survives is a selected sample, not a random one.
2. **The planner will not author the trade the tape is offering `[T]`.** On 09-03, during a
   **+483-point rally** with the plan's own bias `long` and `day_type: trend`, the planner
   proposed **0 `open_long` in 575 decisions**; long scenarios were arm-enabled **1 of 23 =
   4.3%** against shorts at **8 of 18 = 44.4%** (`2026-09-04-two-day-audit.md:279-292`,
   arithmetic verified by me). The prior day: long 14.1%, short 62.8%. This is a structural
   short bias, not a bad day.
3. **The arms funnel loses almost everything, mechanically `[T]`.** Of 15 arms in the
   two-day window, **3 filled (20%)**. The single largest killer is the marketable guard:
   **5 of 15 (33%)** died as *"level accepted through — marketable, never placed"*. Read
   that against the audit's D6 — **71% of arm-enabled scenarios (34 of 48) were authored at
   a price the tape never reached during that version's own life**
   (`2026-09-04-two-day-audit.md:877`) — and the system in one sentence is: **it authors
   levels the market does not visit, and refuses the ones it visits too fast.**

**The three biggest opportunities.**

1. **Stop fading and start trading the break, at the one place the tape gives you an edge
   `[T]`.** RTH-L breaking 68% of the time (n=63) is the only statistically separated
   result in the entire level census. It is currently traded backwards.
2. **Fix the funnel and you multiply whatever edge exists by ~5×, for free `[T]`.** A 20%
   fill rate means four of every five correct opinions never reach the market. The
   marketable-guard rejections in particular are not risk management — they are a systematic
   filter that keeps you out of exactly the fast moves that carry follow-through.
3. **Turn the $450 daily loss limit ON — it is a free option `[T]`.** On the 11 committed
   session days it trips on **1 day (9.09%)**, the −$492.00 day of 08-20, and it forfeits
   **$0.00** — because that day's limit-crossing trade was the day's *last* trade
   (`day_sim.csv`; `mc_drawdown.py:127-128` prints exactly this condition). Measured cost:
   nothing. It is currently OFF.

---

## 1–8. RESERVED — sub-agent parts A–D

Not delivered at push time. Sections 1–3 (way it trades / levels / decisions), 4–6
(monitoring / execution / risk), 7 (prompts / laws / settings) and 8 (the three-day stretch)
will be merged in on delivery, each verified against the store-derived CSVs before use.

---

## 9. MY TOP TEN

Ordered. Each carries **what · why (evidence) · what it takes · the number I would watch**.
Every item was cross-checked against `AUDIT-CHECKLIST.md` so that nothing already fixed is
recommended again; the things I found already fixed are listed at the end of this section
and I am **not** re-recommending them.

**1. Stop seating levels as fade triggers until the kind has a hold rate whose Wilson
interval excludes 50%.**
*Why:* first touch is **48.76%, n=242, Wilson [42.5%, 55.0%]** `[T]`. Of 18 kinds, 17 have
50% inside their interval. Seating a level as an entry trigger is asserting a market belief;
on this tape that belief is unsupported for every kind but one.
*What it takes:* an owner ruling plus a seating-eligibility predicate — data already exists.
*Watch:* count of seated entry-trigger kinds whose Wilson lower bound > 0.50. Today: **0**.

**2. Trade RTH-L as a break, not a hold.**
*Why:* **20/63 = 31.75% hold, Wilson [21.6%, 44.0%]** `[T]` — the only kind in the census
that separates from a coin flip, and it separates toward the break. n=63 is thin, so I would
size it as a probe, not a program.
*What it takes:* prompt + scenario vocabulary (the `breakdown` scenario already exists).
*Watch:* RTH-L break-side expectancy, target n≥100 before it earns a real allocation.

**3. Fix the marketable guard — do not let it silently eat a third of the book.**
*Why:* **5 of 15 arms (33%)** died *"level accepted through — marketable, never placed"*,
and **9 of 15** had `price_traded_through=YES` `[T]`. `[I]` In my experience a limit that
refuses to become marketable is not protecting you from slippage; it is selecting for slow,
mean-reverting approaches and discarding fast ones — that is adverse selection you are doing
to yourself. The guard should convert to a controlled stop-entry or a bounded marketable
order with a slippage cap, not cancel.
*What it takes:* code, in `armed_executor.go`, plus an owner ruling on the slippage cap.
*Watch:* arms filled ÷ arms whose level was reached. Today **3 of 9 = 33%**.

**4. Make the planner author longs, and prove it with a per-side arm-enablement floor.**
*Why:* **4.3% long vs 44.4% short arm-enablement on a +483-pt trend day with a `long` plan
bias; 0 `open_long` in 575 decisions** `[T]`. A system that cannot express its own stated
bias does not have a bias.
*What it takes:* prompt work first (the checklist already records the mechanism at
`:848-853` — an unconditional `MUST` naming a continuation short, which the validator then
voided, so the model was punished for obeying). Then a measured floor.
*Watch:* long-side arm-enablement rate on days the plan bias is `long`. Today **4.3%**.

**5. Turn the $450 daily loss limit ON.**
*Why:* trips **1 of 11 days (9.09%)**, forfeits **$0.00/day** `[T]`. Zero measured cost, and
it caps the −$492.00 tail. `[I]` A daily stop's real job is not P&L, it is stopping the
operator from re-engaging on a day his read is demonstrably wrong.
*What it takes:* a knob, already built and currently off.
*Watch:* days tripped, and the forfeited-P&L figure — if forfeited/day turns materially
negative over a larger sample, revisit.

**6. Do NOT add a daily trade cap — and say so in the record.**
*Why:* `max_daily_trades_3` trips **81.84% of days** and costs **−$24.54/day** net `[T]`
(`day_sim.csv`). It is the intuitive risk control that this tape says is wrong. I am listing
it because it is the kind of thing that gets added after a bad week.
*What it takes:* an owner ruling, written down.
*Watch:* nothing — the point is to not build it.

**7. Report the effective sample as days, not trades.**
*Why:* n=64 trades are **11 session days**, and the P&L is dominated by two of them
(+$469.00, −$492.00) `[T]`. Trades inside a day share one plan, one bias, one regime — they
are not independent. The MC rig bootstraps by trade with a mean block of 5
(`mc_drawdown.py:35-38`), which approximates a day by accident of scale (64/11 = 5.8
trades/day) rather than by design.
*What it takes:* data/reporting change; key the block bootstrap to `session_day_ct`.
*Watch:* distinct session days. Today **11**. I would not take any strategy conclusion
seriously below ~60.

**8. Publish the ambiguity rate next to every hold rate.**
*Why:* SWG-L and eVWAP are **50% unclassifiable**, SWG-H **48.3%**, OB **44.4%**, VWAP
**38.6%** `[T]`. A rate computed on the classifiable half of a sample is not that sample's
rate. This is quietly the biggest measurement defect in the level census and it flatters
exactly the kind (VWAP) that looks most tradeable.
*What it takes:* reporting change in the detector telemetry; the field already exists.
*Watch:* ambiguous ÷ rows_all per kind; suppress any kind above ~25% from decision use.

**9. REMOVAL — retire `eVWAP`, `VWAP±2σ`, `EQL`, `ONH`, `ONL`, `PDL` as decision inputs.**
*Why:* n_den of **5, 4, 2, 3, 3, 3** respectively `[T]`. `ONH`, `PDL` and `EQL` show
100% hold on **n=3, n=3, n=2** — those are not edges, they are rounding. Carrying them
implies a knowledge the system does not have, and every one of them widens the surface the
planner must reason over.
*What it takes:* code + owner ruling.
*Watch:* seated kinds count; and no decision citing a kind with n < 30.

**10. REMOVAL — delete the 105-minute structural wake blackout, or state it as a rule.**
*Why:* NY `WindowEndCT`=14:45 and ASIA `ReadCT`=16:30 make `inSessionReadWindow` false for
all of **14:45–16:30 every weekday** (`2026-09-04-two-day-audit.md:880`) `[T]`. Either it is
a deliberate no-trade window — in which case it belongs in the plan, on the card, where the
owner can see it — or it is an artefact of two constants that were never diffed. `[I]` Dead
time you did not choose is how you end up flat through the move that pays for the month.
*What it takes:* code, or an owner ruling that promotes it to a stated rule.
*Watch:* decisions per hour across 14:45–16:30. Today: none, by construction.

### Cross-check — already fixed, NOT re-recommended
- **BE and the ATR trail suspended** behind `EXIT_MECHS_SUSPENDED`
  (`AUDIT-CHECKLIST.md:632-644`). This was the right call and the evidence is clean: 2 BE
  moves and 8 trail ratchets on 09-01, **$719.50 of giveback, and zero trail EXITS ever**. A
  mechanism that moves a live stop and has never once been the reason a trade ended is
  giving away money for nothing. Leave it suspended.
- **Stop floor 1.0→1.5×ATR5m, composed stop, level never invented** (`:640-644`).
- **"One gate, two ATRs"** — the decision path read the DAILY ATR (300.4) and refused
  against a phantom `1.5×ATR5m = 450.56` while the arm seam used ATR5m 12.78–14.12; all
  three refused targets printed within 28 minutes. Fixed; both seams now read
  `armSeamATR5m` (`:777-789`).
- **Stage-A size ceiling: 1 contract until n≥30 with a positive lower-CI expectancy**
  (`:645-647`). **This is the best rule in the checklist and I would not touch it.** Note
  where it currently stands: n=64 clears n≥30, but the lower CI is **−$31.27**. *The
  system's own rule already says do not size up.*

---

## 10. IDEAS

**The honest go/no-go for live money, in numbers `[T]`.** n=64, mean **−$6.6239**, sd
**$100.5891**, SE **$12.5736**, **t = −0.527**, 95% CI **[−$31.27, +$18.02]**. Zero is
inside the interval; so is −$30 and +$18. Win rate **32.8%**, average win **$114.67**,
average loss **−$65.86**, payoff **1.74** — which needs **36.5%** to break even, so the
whole deficit is **3.7 percentage points, about 2.4 trades out of 64**. At this variance you
need roughly **886 trades** to resolve a mean of this size at 95%. Against that, the median
50-trade drawdown is **$866** (iid; $847 block), p90 **$1,477**, p99 **$2,065**, worst
**$3,029.50** (B=10,000). So the expected 50-trade P&L is about **−$331** while the *median*
drawdown you must sit through is **$866**. **The signal is an order of magnitude smaller
than the noise.** No-go, and it is not close — but note that "no-go" here means *not yet
measurable*, not *proven unprofitable*. Those are different findings and the system deserves
the more precise one.

**`[I]` What I would do with the 2-minute executor loop.** I would stop asking the LLM to
decide *whether* to trade and ask it only to decide *where the level is and whether the
level is still valid*. Entry, sizing and exit are arithmetic; they belong in code where they
can be tested. In my own systems the LLM earns its keep as a **state classifier** — is this
still a trend day, has this level been reclaimed, is the news window live — and loses money
the moment it is allowed to author the trade. This system currently lets it author the trade
and then spends a large gate chain undoing that. Untested here.

**`[I]` The R:R 2.0 target is fighting the 32.8% win rate.** A fade book at a coin-flip
level with a 2R target is asking for a 36.5%+ hit rate at a location where the tape gives
you 50/50 minus costs. Either the entry has to get better or the target has to come in. I
would test 1R-to-scratch-runner before I would test anything else, because it converts the
payoff problem into a win-rate problem, and win rate is the thing this system could actually
measure at n=64.

**`[I]` What I have never seen work, and this system does.** Authoring resting orders at
prices derived from a model's *narrative* rather than from live book pressure — the 71%
never-reached figure is what that looks like from the outside, and it is not a prompt bug,
it is what happens when the thing choosing the price cannot see the depth. Also: a
one-open-position constraint combined with a fade book. The two fight each other — a fade
book's whole risk profile assumes you can be wrong at one level and right at the next, and a
single slot means your first bad level costs you the session.

**`[I]` One thing I would steal from a discretionary desk.** A written "what would have to be
true" line on every plan, and an end-of-day check of whether it was. Not for the model — for
the owner. The reason this project's audit discipline is so strong and its trading is not is
that the audits measure the machine and nobody is measuring the thesis.

---

## SURPRISES — recorded, not acted on

- **The 1E rig is committed to the repository** (`mc_drawdown.py` + `trade_sample.csv`), so
  the drawdown numbers are fully reproducible by anyone who clones. That is unusually good
  practice and it is the only reason this review has numbers at all.
- **`day_sim.csv` shows the $450 limit forfeiting exactly $0.00**, which looks like a bug
  until you read `mc_drawdown.py:127-128` and find the script anticipated it: *"forfeited 0:
  every trip landed on the day's LAST trade — nothing followed it."* The author of that
  script was being careful. Worth saying out loud.
- **`docs/superpowers/plans/VL-MASTER-PLAN-v2.md` is referenced by the dispatch but is not
  in the tree.** Either it was never committed or it moved. Not investigated.
