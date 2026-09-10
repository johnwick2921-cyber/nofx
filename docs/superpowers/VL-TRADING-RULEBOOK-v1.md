# VL INTELLIGENT — THE TRADING RULEBOOK
### What the system does every day, in a trader's words · 2026-09-10 · v1

This is not the plan and not the checklist. It is the book: the rules the system
trades by, written so a trader can read them in ten minutes and know what will
happen and what will not. Every rule carries its evidence label — **[R]** researched,
**[T]** measured on our own tape, **[O]** owner-ruled, **[I]** invented and unproven.
Where a rule is [I], the system still follows it — but it says so, and it is being
measured.

The system today, in one sentence: **it marks levels on every timeframe, waits for
price to reach one, fades it if the level holds, and stops itself out beyond it —
one contract, one position, flat at the close.** Everything below is the detail.

---

## PART 1 — WHAT IT IS

**A level trader.** Not a momentum trader, not a scalper, not a trend-follower. It
believes price reacts at prices that mattered before: yesterday's high, the
overnight low, the opening range, the volume shelf, a swing that held. It marks
those, and it trades the reaction.

**A fade book, by weight.** 116 reject scenarios, 80 sweep-reclaims, 34 reclaims,
29 retests in the record. When it has a choice, it sells the top of the range and
buys the bottom. It carries continuation plays in its vocabulary; it rarely
arms them. [T]

**With a Market Profile map and ICT names.** The map is Dalton's — value area, POC,
initial balance, opening range, VWAP. The play names are ICT's — sweep, reclaim,
order block, FVG, displacement. The execution is classic support/resistance: a
resting limit at the level, a stop beyond it. Three schools; the rulebook below is
what survived when we asked which parts have evidence.

**What it is NOT allowed to be.** It does not scale in. It does not average down.
It does not move stops to breakeven or trail (suspended by 0B until measured). It
does not hold overnight. It does not trade a second contract. It does not enter
without a plan scenario that names the trade — `plan_mode=strict`, the decision
loop's own entries are refused. [O]

---

## PART 2 — THE DAY, HOUR BY HOUR (all times CT)

```
16:30  ASIA read       — the planner writes the ASIA plan
17:00  Globex opens    — arms may rest; the fade book is live
01:30  LONDON read     — new plan; ASIA's arms are superseded
08:00  NY read         — new plan
08:30  RTH opens       — OR forms (first 5 min), IB forms (first 60)
08:30–11:00  morning   — the highest-weighted window [I]
12:00–13:30  lunch     — NO NEW ENTRIES, both paths [O, enforced 09-09]
13:30–14:45  afternoon — entries permitted
14:20  last entry      — 25 min before the flat [O]
14:45  FLAT            — position closed, arms cancelled, pending
                         placements dropped, confirmed by the broker book [O]
14:45–16:30            — the maintenance window; no reads, no arms
```

**Shortened sessions** (Labor Day, Thanksgiving Friday, Christmas Eve, New Year's
Eve): the session calendar says so at boot, the close is the stated early close,
and the book is flat at it. Full closures: no reads, no arms. A date the calendar
cannot classify is CLOSED and named on the boot line. [O]

**What wakes a re-plan between reads:** a seated level touched, broken, or
invalidated · a new level seated · a structure event (MSS/BOS) · a fast-market
move ≥ 1.5×ATR5m · the plan's own flip or death condition · the last tradeable
scenario dying (exhaustion, WARN + counter for now). Cooldown 30 min and cutoff
25 min before the flat, both enforced; fast-market exempts the cooldown only. [O]

---

## PART 3 — THE MAP

**Every level, every timeframe, always kept.** Nothing is deleted from the map
because it lost a seat. A level not good enough to enter on is still the target,
the obstacle, or the invalidation. **Exclusion is not invalidation.** [R — the
research's one structural ruling]

**What is marked** (and what it is called on the card):

| Family | Definition | Evidence |
|---|---|---|
| PDH / PDL / PDC | prior session-day high/low/close (calendar-day bucketing is a known defect — fix pending) | [R Osler, FX] |
| ONH / ONL | overnight session high/low, current session-day | [I] |
| OR-H / OR-L | first 5 minutes after 08:30 | [R Zarattini, ETF only] |
| IB-H / IB-L | first 60 minutes | [I doctrine — Dalton] |
| VWAP, ±1σ, ±2σ | session VWAP anchored 17:00 | [I] |
| POC / VAH / VAL | prior-session volume profile (close-bin proxy, not true volume-at-price — known) | [I doctrine] |
| Swings | 5m / 15m fractal pivots | [T positive — the one seat rule with a measured effect] |
| Supply / Demand zones | 6-candle base, 1.5×ATR departure | [I] |
| OB / FVG | ICT order block, fair value gap | [I — no evidence anywhere] |
| Round numbers | 100 / 50 / 25 multiples | [R Osler, FX clustering] |
| PWH / PWL | prior-week high/low, from daily bars | [I — untested on NQ] |
| Projections | measured move, ATR-projected extreme, round numbers beyond the range | [I — targets and obstacles only, never entries] |

**How the shortlist is built.** Overlapping references within 3 points merge into
one candidate carrying all its names — three names on one price count once for the
model, though the score still counts them (known, E4 will judge). A level is an
ENTRY candidate only when an opposing reference exists at least a stop-floor away.
The entry shortlist orders by reachability — nearest first — up to the 12-seat cap.
Tier-1 anchors (PDH/PDL/PDC, ONH/ONL, VWAP) always seat. [O, labelled I]

**The grade is a label, not a probability.** Kind × freshness × confluence × HTF
1.2 — every term is uncalibrated. The research says so; the card says so. [R]

---

## PART 4 — THE TRADE

**A scenario must state, before it is accepted** [O, since 09-08]:
entry zone · trigger · confirmation · structural invalidation · protective stop ·
**the first opposing obstacle** between entry and target, with its provenance ·
**what it will do there** (pass / reduce / exit / decline) · the arm target · the
implied R to the obstacle and to the target.

A scenario that contradicts itself — target off its own path, obstacle beyond
the target, R that disagrees with its geometry — is **refused at write**. A first
obstacle under 1R is **not refused, but marked** — a fact the owner sees, not a
rule the system enforces. [O — the research forbids prescribing a target policy]

**Confirmation** [O, since 09-09]:
- "5m close beyond X" means a **closed** 5-minute bucket. A forming bucket is NOT
  MET. Not "≈ met." The verdict names the bucket and its close time.
- "Sweep then reclaim" checks the **order**. Part two must occur after part one.
  If part one's instant cannot be established, the verdict is UNKNOWN — not met,
  never plan-birth.
- Immediate-mode displacement keeps its own named rule, `1m_displacement`, and
  never borrows the 5m function.

**Entry.** A resting limit at the level for fades; a stop-entry beyond the trigger
for reclaim. The entry order carries its own OCO; the stop and target are created
on the fill with a *different* shared OCO — so cancelling an unfilled entry can
never touch a bracket. [O, since 09-08]

**Known cost of the resting limit:** ~66% of passive NQ fills immediately precede
an adverse move. The fade is selling adverse selection back to the market; its
edge, if any, must exceed that. [R Lalor & Swishchuk 2024]

**The stop.** Beyond the nearest seated level, or 1.5×ATR5m, whichever is wider.
The floor is owner-ruled and unvalidated; the first number pointing at it — winners'
MAE p80 22.5 vs a 33-point floor, n=18 — says it may be wide. [O, labelled I]

**The target.** Whatever the scenario named — and the scenario now has to name the
first obstacle on the way. 79% of first targets historically sat past a nearer
level; planned 2.55R paid 1.66R. No target family is prescribed; E2 decides. [T]

**The gate, in the order it runs** — every leg must pass or no order goes out:

```
one open position     one contract, one position; working entries count as
                      exposure per ACCOUNT, checked against the broker's book [O]
daily limit           $450 — DECORATIVE by owner choice (both switches off) [O]
strict                only a plan scenario may trade [O]
scenario direction    the scenario's side matches the order [O]
shadow                shadowed conditions cannot place [O]
invalidation          a scenario already invalidated cannot arm [O]
R:R at fill           ≥ 2.0 against the arm's target [O]
min-SL                ≥ 1.5×ATR5m [O, I]
no-trade band         lunch 12:00–13:30 and first-N, on BOTH paths [O]
breaker               8 consecutive losers halts the day; 5 warns [I]
marketable guard      a limit already through its level is cancelled, not chased [O]
```

**What the gate has proven so far:** the 44 refusals since 09-02, taken at their
authored geometry on the real tape, would have **lost $861**. The gates are not the
problem. [T]

---

## PART 5 — THE POSITION

**Once filled:** the bracket rests at the broker — stop-market, GTC, target limit,
one OCO. The system reads the broker's book as truth, never its own ledger. A
position whose bracket the ledger cannot see is reported PROTECTION UNKNOWN on the
desk strip, not assumed protected. [O]

**Nothing moves the stop.** Breakeven and trail are suspended — the knob in Studio
says ON, the binary says OFF, the Guide now says so. Until E2 measures the exit
geometry, the initial stop is the stop. [O — 0B]

**Exits:** the stop, the target, or the 14:45 flat. The exit cause is recorded from
the broker's own event, never inferred from price. [O, since 09-06]

**A cancel is confirmed by the broker's book** — never by the call returning. A
placement is `place_pending` until a received frame names it. A rejection goes
terminal in the broker's own words. [O, classes 81/83]

---

## PART 6 — RISK

```
size            1 contract, always                                 [O]
positions       1 at a time                                        [O]
daily limit     $450 — set, NOT enforced (owner's choice)          [O]
breaker         8 losers → halt; 5 → warn                          [I]
flat            14:45, or the early close — position AND arms      [O]
overnight       never                                              [O]
scaling         never                                              [O]
live money      NO — the gate is n≥100 over ≥40 active days under
                one unchanged policy, both 95% lower bounds > 0,
                max DD ≤ $900, worst day ≤ $450. Every line fails. [O]
```

---

## PART 7 — WHAT THE RECORD SAYS ABOUT THE BOOK

```
58 trades · 12 CME days · −$466 · 32% wins · payoff 1.75
expectancy −$8.04, 95% CI [−$35.55, +$22.04] — indistinguishable from zero
longs −$808 (19) · shorts +$342 (39)
ASIA −$552 (16) · LONDON +$24 · NY +$62
reject fades +$586 (31, the only positive cell) · every follow play negative
levels hold 48.8% of first touches, n=423, [44.1%, 53.6%] — a coin flip
zero trades under the rules running today
```

**What that means, as a trader:** the book has not shown an edge. The fade is the
only cell that pays; longs and follows lose; ASIA loses. None of it is enough
trades to call. And the research says no intraday strategy at one-contract retail
scale on NQ is proven either — so "switch strategies" is not an answer the evidence
supports. **The answer is: select.** When to fade, which level, which entry, what
geometry. Every one of those is now a recorded question.

---

## PART 8 — WHAT A TRADER WOULD DO THAT IT DOES NOT — YET

| The trader | The system today | The wave |
|---|---|---|
| Reads the first hour and decides range or trend; stands aside on trend | Fades every day the same way — sold into +483 pts | W2 fade permission (label first, gate after E3) |
| Picks the one level that matters | Writes 3–5, arms whichever fires first | W3 shipped the shortlist; "one trade" waits on E4 |
| Watches the candle form at the level | Waits for a closed 5m bar, checked every 2 min | W-LIVE, after 13f's test |
| Draws every timeframe | Detects swings on 5m/15m, zones on 1h; daily/weekly never | W-TF, detection now, weight after round 12 |
| Knows what's beyond the last level on a trend day | Had no target past the map until 09-09 | 103 shipped projections as targets only |
| Takes the first obstacle or trails through it | Aimed three levels away, paid 2R for 0.6R | scenario economics shipped; E2 decides the policy |
| Manages the winner | Initial stop is the stop | after E2 |
| Sizes by conviction | One contract, always | after live-money gate |

---

## PART 9 — THE RULES THE SYSTEM MUST NEVER BREAK

1. **The broker's book is the truth.** Never the ledger's word.
2. **A send is not a settlement.** Placed means a frame came back; cancelled means
   the book no longer lists it; flat means the book is empty.
3. **UNKNOWN never takes the destructive branch.** A stale book refuses a
   placement; it never confirms a cancel; it never reads as flat.
4. **A forming bar is not a closed bar.** A confirmation on a bucket that has not
   closed is not met.
5. **A scenario may not contradict itself.** Target on its path; obstacle before
   the target; R that matches its geometry.
6. **A level cut from the shortlist stays on the map.** Exclusion is not
   invalidation.
7. **One contract, one position, per account, counting working entries.**
8. **Flat means flat** — position, arms, and pending placements, book-confirmed.
9. **No market belief ships without a research verdict, and every rule carries its
   label.** A grade is a label. A floor is a parameter. A [I] is a guess we are
   measuring.
10. **The system says what it does.** Every boot line reads from the code that
    enforces it; every Guide sentence matches the binary; a number nobody can
    compute reads UNKNOWN, never zero.

---

## PART 10 — WHAT CHANGES THIS BOOK, AND WHAT DOES NOT

**Changes it:** an experiment on the record — E1 entry type, E2 exit geometry,
E3 fade permission, E4 ranking vs distance, E5 session risk — each paired,
chronologically split, with a Reality Check, returning SUPPORTED / CONTRADICTED /
INCONCLUSIVE. Or an owner ruling, labelled [O], recorded with its date.

**Does not change it:** a good week. A bad week. A trader's instinct without a
number. A research round that says "untested." A green suite. A passing check.

That is the book. It is a level fade, it is honest about what it knows, and as of
this week it can be measured. The first honest answer — does the fade pay on range
days — is E3, after twenty recorded sessions.
