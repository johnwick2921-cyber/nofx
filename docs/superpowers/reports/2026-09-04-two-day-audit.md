# Two-day audit — every loss, every refused opportunity, and why a fast tape produced one trade

**Window** 2026-09-02 00:00 CT → 2026-09-03 23:31 CT (with 08-26…09-01 as baseline)
**Dispatch** owner hoang, 2026-09-04 · READ-ONLY · SIM-only (Sim101), MNQ only
**Branch** `docs/two-day-audit-0904` · base `dfbfa660` (dev tip at accept) · worktree `~/nofx-2day04`
**Running rev at audit time** boot line `530009ff`; PID 594377 started 2026-09-03 23:12:54 CT
**Evidence classes** **[A]** directly verified · **[B]** inferred from strong evidence · **[C]** speculation

---

## 0. DANGEROUS RIGHT NOW — nothing

**[A] There is no open position and no resting arm.** Nothing needs the owner's hand tonight.

```sql
sqlite3 "file:/home/hoang/nofx/data/data.db?mode=ro" \
 "SELECT * FROM trader_positions WHERE status='OPEN';"                        -- 0 rows
sqlite3 "file:/home/hoang/nofx/data/data.db?mode=ro" \
 "SELECT * FROM armed_orders
   WHERE state NOT IN ('filled','cancelled','canceled','expired','done');"    -- 0 rows
```

The last position (id 591) closed 2026-09-03 09:20:45 CT. The last arm (id 37) was cancelled
2026-09-03 12:15:01 CT. Nothing has been working at the broker since.

---

## 1. VERDICT FIRST

The two days were **not** a gate problem. In proportion, from the tables below:

| # | cause | share | the number that says so |
|---|---|---|---|
| **3** | **PLANNER / ENTRY SHAPE** | **~55%** | 09-03: **0 `open_long` proposed in 575 decisions** during a +483-pt rally; long scenarios were arm-enabled **4.3%** of the time vs **44.4%** for shorts |
| **5** | **NOT-A-RULE, NOT-THE-TAPE — host outage + post-boot blindness** | **~30%** | **the WSL2 host rebooted at 14:18** (`who -b`); silence 12:24:33→14:18:24 CT (**113 min 51 s**), then **~50 min blind** on FEED DOWN / `NT8 TCP link DOWN — NEW entries BLOCKED`. Together: the NY afternoon containing the day's high (29585) |
| **2** | NO SETUPS | ~10% | the day's only arm-enabled long (29481.05) **missed by 8.2 pts**; 09-02's four losses each stopped at their own authored stop |
| **1** | GATES TOO TIGHT | **~5%** | the entry gate's eight legs refused **zero** entries across both days (counter-proven); the two gates that did fire net **negative** in their counterfactual ledgers |
| **4** | CADENCE | **~0%** | class-47 cutoff/cooldown suppressed **nothing**: 09-02 ran WARN-first, 09-03 ran enforce and its single firing was fast-market-exempted |

**Can this system, as configured today, take a trade on a trending fast tape?**
**[A] Only by accident.** On 09-03 it correctly identified the trend — the machine regime read
`up/NORMAL` on all 17 planner reads, and from NY v3 (09:15 CT) the plan bias was `long` with
`day_type: trend` — and it still took exactly one trade: a **short**. The rule that prevents it is
not a gate. It is that **the planner grants resting arms almost exclusively to short scenarios**,
while long scenarios are left on the AI decision path, which proposed `open_long` zero times all
day. This is a **planner-shape answer**, not a gate answer.

---

## 2. C1 — THE TRADES

`pnl_corrected` only. **[A] Zero exclusions on both days**: 0 NULL-corrected, 0 `UNRESOLVABLE`,
0 `source='e7_farside_test'`, 0 pre-era (`DayPlanEraStart` = 2026-08-15 00:00 CT,
`store/attribution.go:131-158`).

| day (CT, by exit) | n | Σ pnl_corrected | excluded |
|---|---|---|---|
| 2026-09-01 (context) | 6 | **+212.00** | 0 / 0 / 0 |
| **2026-09-02** | **4** | **−381.50** | 0 / 0 / 0 |
| **2026-09-03** | **1** | **−140.00** | 0 / 0 / 0 |

### D1 — every closed position in the window

| id | sess | v | scen | side | entry (CT) | exit (CT) | entry→exit | pnl | authored SL | exit vs SL | path |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 587 | ASIA | 7 | S3 | LONG | 09-02 00:17:44 @29079.25 | 01:03:41 @29048.00 | −31.25 | **−62.50** | 29048.00 | **exact** | system |
| 588 | LONDON | 3 | S2 | LONG | 09-02 07:41:05 @29082.50 | 07:51:38 @29050.00 | −32.50 | **−65.00** | 29050.00 | **exact** | system |
| 589 | NY | 3 | S3 | LONG | 09-02 09:41:04 @29192.50 | 09:59:27 @29115.00 | −77.50 | **−155.00** | 29115.00 | **exact** | system |
| 590 | NY | 5 | S4 | LONG | 09-02 10:37:17 @29193.25 | 10:49:30 @29143.75 | −49.50 | **−99.00** | 29143.75 | **exact** | system |
| 591 | NY | 2 | S1 | SHORT | 09-03 09:05:14 @29285.00 | 09:20:45 @29355.00 | −70.00 | **−140.00** | 29351.63 | +3.37 slip | armed→reconcile |

**[A] All five were clean stop-outs at the position's own authored stop** — four to the tick, one
with 3.37 pts of slippage. There is **no premature-exit defect**: `close_reason='sync'` is the
broker-state sync recording a stop fill, not an early close. Multiplier confirmed empirically:
591 moved 70 pts for −140.00 → **$2/point** (MNQ).

```sql
-- the authored stop for each decision-path loss
SELECT json_extract(j.value,'$.action'), json_extract(j.value,'$.stop_loss'),
       json_extract(j.value,'$.confidence'), datetime(d.timestamp,'-5 hours')
FROM decision_records d, json_each(d.decision_json) j
WHERE json_valid(d.decision_json) AND json_extract(j.value,'$.action') LIKE 'open%';
-- 587 SL=29048 conf=62 · 588 SL=29050 conf=63 · 589 SL=29115 conf=65 · 590 SL=29143.75 conf=68
```

### The exits were right; the stops were where price turned

30 minutes after each exit, from 1m bars:

| id | side | exit px | best 30m **for** the trade | worst 30m **against** | reading |
|---|---|---|---|---|---|
| 587 | LONG | 29048.00 | +10.25 | −31.50 | exit correct |
| 588 | LONG | 29050.00 | +30.75 | −17.00 | marginal |
| 589 | LONG | 29115.00 | **+93.00** | −8.75 | **stopped at the low** |
| 590 | LONG | 29143.75 | **+54.25** | −8.00 | **stopped at the low** |
| 591 | SHORT | 29355.00 | **+82.50** | −9.75 | **stopped at the high** |

**[A] Three of five stopped within ~9 pts of the exact turning point** and then ran 54–93 pts in the
trade's favour. That is a stop-*placement* finding, not an exit-path finding: the stops were
honoured exactly as authored, and each sat where the market reversed from. 589 and 590 are
substantially the same idea 56 minutes apart (long 29192.50 then long 29193.25) — one thesis,
re-entered, stopped twice, for −254.00 of the day's −381.50.

---

## 3. C4 — THE TAPE

Session windows, read from the running process's own ledger boot line (not assumed):

```
🧾 ledger boot: sessions[ASIA 17:00→02:00 CT (last-entry 01:45, flat 02:00)
                       | LONDON 02:00→08:30 CT (last-entry 08:15, flat 08:30)
                       | NY    08:30→14:45 CT (last-entry 14:30, flat 14:45)]
```

| day | low | high | range | shape |
|---|---|---|---|---|
| 2026-09-02 | 28927.25 | 29255.00 | **327.75** | **chop / balance** — 40–90 pt swings, no sustained leg; open 29059 → close 29207.50 |
| 2026-09-03 | 29075.00 | 29585.00 | **510.00** | **strong uptrend** — 29101.75 (06:00 low) → 29585.00 (13:00 high) = **+483.25 in 7 hours** |

09-03 hourly (CT), the two impulse hours in bold:

| hour | O | H | L | C | Δ |
|---|---|---|---|---|---|
| 06:00 | 29129.00 | 29187.25 | **29101.75** | 29105.25 | −23.75 |
| **07:00** | 29105.25 | 29260.00 | 29104.00 | 29244.50 | **+139.25** |
| 08:00 | 29244.75 | 29375.25 | 29194.50 | 29332.25 | +87.50 |
| 09:00 | 29332.25 | 29367.75 | 29241.50 | 29363.25 | +31.00 |
| **10:00** | 29363.50 | 29543.75 | 29362.00 | 29538.25 | **+174.75** |
| 13:00 | 29543.50 | **29585.00** | 29543.00 | 29568.50 | +25.00 |

The owner's "price moving crazy" is 09-03, and it is real: a 510-point day against a 327-point
chop day before it.

---

## 4. THE CENTRAL FINDING — the planner arms shorts and not longs

**[A] The planner correctly read the trend.** Machine regime was `up/NORMAL` on **all 17**
`planner_read_facts` rows for 09-03. From NY v3 (09:15 CT) the plan's own bias was `long`,
`day_type: trend`, and stayed so through v7 (11:27 CT). v4 authored **three long scenarios and
zero shorts**.

| 09-03 NY version | created (CT) | bias | conviction | day_type | longs | shorts |
|---|---|---|---|---|---|---|
| v1 | 08:05:19 | short | medium | balance | 1 | 2 |
| v2 | 08:45:05 | short | medium | balance | 1 | 2 |
| **v3** | **09:15:31** | **long** | medium | **trend** | 2 | 2 |
| **v4** | **09:44:33** | **long** | medium | **trend** | **3** | **0** |
| v5 | 10:14:45 | long | medium | trend | 1 | 1 |
| v6 | 10:49:57 | long | medium | trend | 2 | 1 |
| v7 | 11:27:53 | long | medium | trend | 2 | 1 |

**And it took no long.** The mechanism is arm-enablement:

| day | side | scenarios authored | `arm.enabled` | **rate** |
|---|---|---|---|---|
| 2026-09-02 | long | 64 | 9 | **14.1%** |
| 2026-09-02 | short | 43 | 27 | **62.8%** |
| **2026-09-03** | **long** | **23** | **1** | **4.3%** |
| **2026-09-03** | **short** | **18** | **8** | **44.4%** |

```sql
SELECT p.trade_date, json_extract(s.value,'$.direction') dir,
  SUM(CASE WHEN json_extract(s.value,'$.arm.enabled')=1 THEN 1 ELSE 0 END) armed_enabled,
  COUNT(*) total
FROM plans p, json_each(json_extract(p.doc,'$.scenarios')) s
WHERE p.trade_date IN ('2026-09-02','2026-09-03') GROUP BY p.trade_date, dir;
```

A scenario without `arm.enabled` has no resting order; it can only become a trade if the AI
decision loop proposes an open. **[A] On 09-03 the decision loop proposed `open_long` zero times
in 575 decisions:**

| day | wait | open_long | open_short | decisions | open-intent |
|---|---|---|---|---|---|
| 2026-08-27 | 585 | 0 | 4 | 589 | 0.68% |
| 2026-08-28 | 403 | 0 | 0 | 405 | 0.00% |
| 2026-08-30 | 155 | 0 | 0 | 155 | 0.00% |
| 2026-08-31 | 463 | 0 | 0 | 463 | 0.00% |
| 2026-09-01 | 640 | 1 | 2 | 643 | 0.47% |
| **2026-09-02** | 623 | 5 | 5 | 633 | **1.58%** |
| **2026-09-03** | 561 | **0** | **14** | 575 | **2.43%** |

```sql
SELECT date(d.timestamp,'-5 hours') day, json_extract(j.value,'$.action') action, COUNT(*)
FROM decision_records d, json_each(d.decision_json) j
WHERE json_valid(d.decision_json) AND d.decision_json <> '' GROUP BY day, action;
```

**Two corrections to the dispatch's premises, both material:**

1. **The drought is not new and 09-03 was not the quiet day.** At 2.43%, 09-03 had the **highest**
   open-intent rate of the eight days measured. Four of the seven baseline days produced **zero**
   open intents. "Wait" at 91–100% is this system's normal state, not a new tightening. **[A]**
2. **The system placed one trade "since last night" is understated.** The last fill is position 591,
   entered **09-03 09:05:14 CT** — one trade in the whole of 09-03, from the NY morning, and
   nothing in the ~14 hours since. **[A]**

### The one arm-enabled long of 09-03 missed by 8.2 points

NY v6 S2, `long` @ **29481.05** (`arm.enabled: 1`), live 10:49:57 → 11:27:53 CT when v7 superseded it.

```sql
SELECT round(MIN(l),2) FROM bars WHERE symbol='MNQ' AND tf='1m'
 AND open_time_ms BETWEEN strftime('%s','2026-09-03 15:49:57')*1000
                      AND strftime('%s','2026-09-03 16:27:53')*1000;   -- 29489.25
```

**[A] Minimum low during its life: 29489.25 — 8.2 points above the entry.** It never became an
`armed_orders` row; the log only ever shows the display-only estimate
(`🎯 scenario S2 → ≈armed @ 29481.05 … never execution-wired`). Price did reach 29476.00 at
11:28 CT — one minute *after* v6 was superseded, and v7's two long scenarios had no arms.
Cause: **`never_reached`**, by 8.2 points and one minute.

---

## 5. THE TRADING DAY WAS LOST TO A HOST OUTAGE, THEN TO POST-BOOT BLINDNESS

**[A] The WSL2 host itself went down. This is not deploy churn and not a defect of the trading
system** — an earlier reading of this audit blamed the restart cadence and was wrong:

```
$ who -b          →  system boot  2026-09-03 14:18
$ last reboot     →  reboot  system boot  6.6.87.2-microso  Thu Sep  3 14:18  still running
```

The Go process, `nofx-web.service` and vite all came up together at 14:18:2x, systemd-started.
The 548 bytes of NUL at offset 779709 of `nofx_2026-09-03.log` are the classic unclean-shutdown
signature (dirty page cache lost), which is why the log's last *readable* line is 12:23:33 while
the process's last verified act is `log_events` id 25754 at **12:24:33 CT**.

**Corrected silence window: 2026-09-03 12:24:33 → 14:18:24 CT = 113 min 51 s.**

What the tape did while nothing was running:

| | value |
|---|---|
| price at down / at up | 29484.50 → 29536.50 |
| low / high in the window | 29479.75 / **29585.00** ← the day's high |
| range traversed | **105.25 pts** |
| NY entry time consumed | 2h of the 08:30–14:30 last-entry window |

**[A] MNQ 1m bars are COMPLETE across the outage** (backfilled on reconnect), so an audit that
looked only at bar coverage would see no outage at all. The tape reconstruction in this report is
therefore sound; only the *live* decision path was absent.

### The genuine defect is what happened AFTER the boot

The process came up at 14:18:24 into a 27-minute remainder of the NY session and **could not
trade for any of it**:

```
09-03 14:18:24→15:10 CT:   31 × 🚨 FEED DOWN (no NT8 bar, age reaching 40m0s, while CME is OPEN)
                            3 × 🚨 NT8 TCP link DOWN — NEW entries BLOCKED  (dead-man watchdog)
                            3 × no_balance_frame                            (equity 0, no AddOn)
```

The next decision cycle after 12:22:44 is **15:08:51** — a 166-minute decision drought, of which
**52 minutes had a fully alive, loudly-logging process that was simply blind**. The NY flat is
14:45. **[A] The whole NY afternoon — 12:24:33 to 14:45, 2h20m containing the day's high — ran
with either no host or no market data, and nothing alerted on the second half.**

**[B] The restart cadence is a separate finding, not the cause of this gap.** `nofx_2026-09-03.log`
alone carries **11 boot banners from at least 6 different build trees** (`cleanclone`, `cc2`,
`cc3`, `cleanbuild`, `cc6`, `nofx`) — audit and fix lanes cutting over repeatedly on a live
trading day, one of them at 11:10:33 CT landing *between* NY v6 (10:49:57) and v7 (11:27:53), so
the running code changed mid-session. Twice in the window the rev-guard caught a stale binary and
refused outright:

```
09-02 00:10:21 [ERRO] 🔐 TRADING REFUSED — binary is revision "23f56f49a536"
                       but the intended release is "aeb11179df5a" — a stale binary is running
09-03 23:11:59 [ERRO] 🔐 TRADING REFUSED — binary is revision "89673ccc5984"
                       but the intended release is "530009ff" — a stale binary is running
```

Both were boot-time and cleared on the next boot, so neither blocked trading for long — but each
is a cutover that shipped the wrong binary onto a live trading host.

---

## 6. D3 + D4 — the arms, and why arm 37 died

All `armed_orders` rows in the window. **Timestamp warning (defect #4 below): `created_at` is
stored with a CT offset and `updated_at` with `+00:00` UTC.** Both are normalised to CT here.

| id | day | sess | scen | side | entry | state | alive | reason | cause class |
|---|---|---|---|---|---|---|---|---|---|
| 29 | 09-01 | ASIA | S1 | long | 29044.00 | cancelled | 70.8m | no order_update within stale window (reconnect) | **defect (restart)** |
| 30 | 09-01 | ASIA | S1 | long | 29035.25 | cancelled | 18.1m | " | **defect (restart)** |
| 31 | 09-02 | ASIA | S3 | long | 29068.05 | cancelled | 15.3m | " | **defect (restart)** |
| 32 | 09-02 | NY | S3 | long | 29070.00 | cancelled | 254.5m | session ended (EOD flat) | never_reached |
| 33 | 09-02 | NY | S3 | short | 29166.80 | cancelled | 0.0m | level accepted through — marketable, never placed | marketable_guard |
| 34 | 09-02 | ASIA | S1 | short | 29199.50 | cancelled | 0.0m | " | marketable_guard |
| **35** | **09-03** | **NY** | **S1** | **short** | **29285.00** | **filled** | 85.6m | → position 591 | **the one trade** |
| 36 | 09-03 | NY | S2 | short | 29351.05 | cancelled | 0.0m | " (price 29358.75 already above 29351.05) | marketable_guard |
| 37 | 09-03 | NY | S3 | short | 29543.75 | cancelled | 16.5m | cancelled in NT8 | **parked(death)** |

**Arms 29/30/31** were cancelled by the stale-order-update reconciler at 00:07–00:31 CT on 09-02 —
inside the window of three boots in eleven minutes (00:01:06, 00:10:20, 00:11:47). **[B]** The
restart resets the NT8 order-update stream; the reconciler then sees no update inside the stale
window and cancels. Three long arms in ASIA died to deploy churn, not to the market.

### Arm 37 — the case that looks like a broker defect and is not

Arm 37 was a **real working order**, not a display estimate — `nt.PlaceLimitEntry` returned a
signal id (`armed_executor.go:927-935`):

```
09-03 11:58:33 ⚔️ armed NY S3 leg 1 short limit 29543.75 SL 29592.50 TP 29397.11
09-03 11:58:33 📌 armed S3 → WORKING limit 29543.75 signal=43512e1d-… (band ±100t)
```

Price later traded at or above 29543.75 on **83 one-minute bars**, up to 29585.00. It never filled.
That looks like a broker defect. **It is not** — the order was already gone:

```
09-03 12:15:00 😴 plan 2026-09-03 NY v7 DORMANT — death-condition: 2x5m close below 29502.25
09-03 12:15:01 📡 armed order_update summary: frames=4 accepted=1 working=1 cancelpending=1 submitted=1
09-03 12:15:01 ✕ armed S3 cancelled in NT8
09-03 12:15:01 🔒 armed cancel: no active plan — 1 order(s) disarmed
```

**[A] The plan's death condition fired at 12:15:00, the disarm-on-no-active-plan rule cancelled the
arm, and NT8 acked it.** "cancelled in NT8" is our own cancel being acknowledged, not an external
one. During the arm's actual 16.5-minute life the maximum high was **29534.25 — 9.5 points short
of the entry**. Price first reached 29543.75 at **12:51 CT, 36 minutes after the arm was gone.**

**Counterfactual, had the arm survived** (filled 12:51 @29543.75, stop 29592.50, target 29397.11):

- stop **never hit** — session max was 29585.00, **7.5 pts short of the stop**
- target **never hit** — session min after fill was 29481.50
- outcome: still open at the NY flat, closed ≈29505 → **≈ +38.75 pts ≈ +$77.50**

One trade's worth of profit, forgone to a death condition that fired 36 minutes early. Note the
death rule and the arm disagreed by construction: v7's bias was **long**, its death was a *close
below* 29502.25, yet the arm it carried was a **short** at 29543.75.

### After the death, nothing replaced it

**[A] NY v7 (11:27:53 CT) is the last NY plan version of the day.** It went dormant at 12:15:00.
The next plan of any session is ASIA v1 at **16:37:36 CT**. Between them the process was down
113 min 51 s (host reboot) and the NY session ended (flat 14:45). No NY plan was on the desk for the final
2h30m of the session, which contained the day's high.

---

## 7. D2 — every gate refusal, and A29 call sites

### The entry gate refused nothing

`trader/entry_gate.go` holds eight legs — strict, plan-bias direction, class-48 direction
mismatch, invalidation-wired, shadow(0C), R:R at execution price, min-SL, one_open_position.
Both production call sites exist:

| leg family | refusal branch | production call site | callers |
|---|---|---|---|
| all 8 legs, decision path | `trader/entry_gate.go:162-284` | `trader/auto_trader_orders.go:333` | 1 |
| all 8 legs, arm path | `trader/entry_gate.go:162-284` | `trader/armed_executor.go:495` | 1 |

**[A] Across 09-02 and 09-03 these legs produced ZERO refusals.** This is a *recorded-counter*
result, not an inferred zero. Refusals increment `arm_refusals_0b:<trader>:<date>:<session>:<class>`
where class is `entry_gate:`+leg (`armed_executor.go:467`). Every counter that exists in the store:

```sql
SELECT key, value FROM system_config WHERE key LIKE '%refus%';
arm_refusals_0b:…:2026-09-02:ASIA:rr    2
arm_refusals_0b:…:2026-09-02:NY:rr      3
arm_refusals_0b:…:2026-09-03:LONDON:rr  1
arm_refusals_0b:…:2026-09-03:NY:rr      1
```

Four keys, all class `rr`, summing to **7**. **No `entry_gate:*` counter exists at all**, and
`grep -c "entry_gate:"` over both log files returns **0 / 0**. The named new legs — strict,
one-open-position, invalidation-wired, shadow map, direction — refused nothing.

### The gates that did fire

| gate | where | n (09-02+09-03) | governing knob (live value) |
|---|---|---|---|
| **min-SL (decision path)** | `kernel/engine_position.go:229` → `kernel/min_sl.go:56` | **34** | `MIN_SL_ATR_MULT` = **1.5** (1.0 on 09-02 before 07:30) |
| **R:R at arm** | `trader/armed_executor.go:455` | **7** | arm min R:R = **2.00** |
| marketable guard | `armed_executor.go:918-924` | 3 | — (wrong-side guard, no knob) |
| stale order_update | `armed_executor.go:940` | 3 arms (ids 29/30/31) | `armedWorkingStaleMin` = 15m |
| plan-death disarm | 12:15:01 09-03 | 1 arm (id 37) | death: 2×5m close, buffer 0.5×ATR14 |
| no-chase | `auto_trader.go` (warn) | 27 lines | **mode=warn — refuses nothing** |
| lunch no-trade | `kernel/no_trade_band.go:58` → `trader/auto_trader_session.go:120` | 0 recorded | 12:00–13:30 CT |

**A11 caveat.** `GET /api/config/resolved` and `/api/risk/gate-blocks` both return
`{"error":"Missing Authorization header"}` from this session, so live knob values are taken from
the process's own boot lines and from the refusal messages themselves (which print the multiplier
they applied) — not from file defaults. Knobs I could **not** establish live are named as such
rather than guessed.

### A reporting defect that invites a false read of the min-SL gate

`kernel/min_sl.go:56-66` formats the raw ATR where a reader expects the threshold:

```go
if dist < mult*atr {
    return true, fmt.Sprintf("sl_too_tight: %.1f < %.1f×ATR (%.1f) — widen or skip", dist, mult, atr)
}
```

So `sl_too_tight: 40.2 < 1.5×ATR (27.3)` reads as "40.2 is less than 27.3", which is false. The
threshold is 1.5 × 27.3 = 40.95, and 40.2 < 40.95 is correct. **[A] All 34 refusals are
arithmetically correct** (`gate_correct=YES` for every row of
`refusals_minsl_counterfactual.csv`). The gate is sound; the line is misleading, and this auditor
initially misread it as a defect.

---

## 8. SECTION F — THE COUNTERFACTUAL LEDGER

The direct test of "gates too tight". Multiplier $2/pt, verified from position 591.

### min-SL leg — n=34 events, 11 distinct opportunity clusters

The refusals cluster: eight of them (09-02 18:48:13→18:59:36) are the *same short* re-proposed
in eleven minutes. Counting events and clusters separately matters and is done here.

Because a min-SL-rejected decision is **not persisted with its prices** (the retry overwrites the
record with the final `wait` — defect #7), an exact stop/target counterfactual is impossible. The
honest substitute: each refusal's own too-tight stop distance against the realised excursion.

| outcome | n | points |
|---|---|---|
| would have hit its **own** too-tight stop within 30 min | **18** | **−431.2 pts = −$862** |
| still alive at 30 min | 16 | ΣMFE30 = +715.5 pts = +$1,431 |
| **upper-bound net** (every survivor exits at its exact 30-min high — unattainable) | 34 | **+284.3 pts = +$569** |

Per side: short n=25 (12 stopped, −248.9 pts) · long n=9 (6 stopped, −182.3 pts).

**Verdict: min_sl is NOT a leg to review.** At n=34 its refusal set is positive only under an
assumption of perfect exits at the 30-minute high; under any realistic exit rule the ledger is
around break-even or negative. Over half of the refused trades would have been stopped out on
their own stop within half an hour.

### R:R-at-arm leg (min 2.00) — n=7

Full stop/target counterfactual (prices recovered from the plan doc scenario in force):

| time (CT) | sess | scen | side | entry | stop | target | R:R logged | outcome | pts |
|---|---|---|---|---|---|---|---|---|---|
| 09-02 09:55:01 | NY | S1 | long | 29135.65 | 29099.00 | 29209.25 | 0.95 | **TARGET** 15:06 | **+73.60** |
| 09-02 10:25:01 | NY | S3 | long | 29058.75 | 29020.00 | 29143.97 | 1.74 | never_reached | 0.00 |
| 09-02 13:06:29 | NY | S1 | long | 29163.99 | 29138.75 | 29246.25 | 1.32 | STOP 13:12 | −25.24 |
| 09-02 18:47:17 | ASIA | S2 | short | 29209.25 | 29223.00 | 29181.64 | 1.30 | **TARGET** 19:07 | **+27.61** |
| 09-02 20:07:16 | ASIA | S3 | short | 29175.66 | 29194.30 | 29136.25 | 1.75 | STOP 20:16 | −18.64 |
| 09-03 08:18:54 | LONDON | S1 | short | 29219.97 | 29255.00 | 29144.50 | 1.68 | STOP 08:30 | −35.03 |
| 09-03 08:30:54 | NY | S1 | short | 29259.67 | 29289.00 | 29178.73 | 1.68 | STOP 08:38 | −29.33 |
| | | | | | | | | **net** | **−7.03 pts = −$14** |

**n=7 is below the n=10 floor, so no verdict is stated.** The number is reported plainly: the
refused set was approximately break-even. Four of seven were stopped; the gate did not cost money.

### entry_gate legs — n=0

No counterfactual is possible and none is invented. **A leg that refused nothing over two days
cannot be the reason the system did not trade.**

---

## 9. C5 + D5 — CADENCE suppressed nothing

**C3 — the rule in force matters, and it changed mid-window.** The wake mode is printed at every
boot, and it is not the same on the two days:

| boots | wake mode printed |
|---|---|
| 09-02 21:19:21, 21:32:52, 22:37:38, 22:41:58, 23:24:56 | `cutoff=25m cooldown=30m` — **WARN-first** |
| 09-03 10:28:29 onward (10 boots) | `cutoff=25m(enforce) cooldown=30m(enforce, fast-market≥1.5×ATR exempt)` |

So class-47 was advisory for all of 09-02 and enforcing for 09-03. **[A] It suppressed nothing on
either day.**

| event | 09-02 file (09-02 00:01→09-03 10:28) | 09-03 file (10:28→23:31) |
|---|---|---|
| wakes **fired** (`waking the planner`) | 43 | 8 |
| skipped on `wake_min_interval_min` | 446 | 47 |
| class-47 cutoff/cooldown lines | 8 | 1 |
| of those, actually **suppressed** | **0** | **0** |

All eight class-47 lines on 09-02 read `would_skip … WARN-first: the wake PROCEEDS`. The single
09-03 line reads:

```
09-03 19:50:05 ⏱ wake cooldown bypassed: fast market 2.4×ATR (cooldown 21 min …, cooldown 30m)
```

— the fast-market exemption working as designed. **The 446/47 `wake_min_interval_min` skips are
the designed cadence ceiling, not a class-47 refusal**, and 43 and 8 wakes still fired through it.

**D5 — the big moves and the plan in force.** The two impulse hours on 09-03 (07:00 +139.25,
10:00 +174.75) both occurred with the process **up** and a plan on the desk (NY v3–v6 were
authored at 09:15, 09:44, 10:14, 10:49 — four re-plans inside the rally). Cadence delivered
plans during the move. What did not follow was an entry, for the arm-enablement reason in §4.

The one genuine cadence hole is **not** class-47: after NY v7 died at 12:15:00 no NY re-plan was
authored, and 9 minutes later the host went down for 113 min 51 s. **[B]** Whether a re-plan
would have fired in that window cannot be determined — the process was not running to fire one.

---

## 10. C6 — touch_outcomes and candidate_pool: zero rows, and it is NOT a defect

```sql
SELECT COUNT(*) FROM touch_outcomes;   -- 0
SELECT COUNT(*) FROM candidate_pool;   -- 0
```

**Premise corrected.** The dispatch framed this as a defect candidate; it is not one **[A]**. The
1B recorder's **production call site landed 2026-09-03 21:31:00 CT** (commit `f1d7cf51`) and
**first ran in a live process at 21:48:12 CT**. The last planner read of the day completed at
**20:24:03 CT**. The hook has therefore never had an opportunity to fire. The tables themselves
were created by AutoMigrate at ~19:06:08 CT on 09-03.

The zero is **(a) the code only just entered the running binary**, not (b) a flag or (c) a wired
hook writing nothing. It settles on the LONDON read scheduled 01:30 CT 2026-09-04 — look for
`🔬 detector recorded: N new episode(s)`.

**But a real defect was found in the new code [A]:** `touch_outcomes.candidate_seated` is
degenerate — the only production writer stamps the literal `true`
(`trader/detector_record.go:66`), so the column can never distinguish a seated candidate from an
unseated one, and every analysis keyed on it will be vacuous from the first row written.

**Related: `trade_excursions` has ZERO rows all-time [A]**, and all 11 positions in the window
lack an excursion row. **This too is expected, not a runtime failure** — the excursion writer
merged to dev 09-03 10:22:16 CT and first booted at 10:28:29 CT, **67 minutes after position 591,
the last trade, had already closed**. `excursionOnClose()` deliberately early-returns for
pre-wave positions.

**Correction to a stronger claim this audit nearly made:** MAE/MFE are **not** missing. They are
present for all 11 positions in `trader_positions.mae`/`mfe` (0 NULLs) — e.g. 589 `mae=80.5
mfe=10.25`, 591 `mae=75.0 mfe=43.5` — written by the same `kernel.ComputePathExcursion` path the
excursion table uses. Only the excursion-table *copy* is absent. Expectancy degrades gracefully:
`loadExcursions()` returns an empty map and `BuildAt` still built `cells=41` at every boot.

The 30-minute after-exit excursions in §2 were computed by this audit from 1m bars and are
labelled as such; they answer a different question (what happened *after* the stop) than the
stored MAE/MFE (what happened *during* the trade).

---

## 11. E — ATTRIBUTION, and the baseline it is judged against

Cause vocabulary per the dispatch. Counts are **opportunities**, not log lines.

### 2026-09-02

| cause | n | detail |
|---|---|---|
| planner_shape (no arm authored on the bias side) | 55 | long scenarios authored without `arm.enabled` |
| gate_min_sl | 6 clusters (16 events) | all arithmetically correct; ledger negative |
| gate_rr_at_arm | 5 | net −36.3 pts across the five |
| marketable_guard | 2 | ids 33, 34 |
| defect(restart→stale arm) | 1 | id 31 (+ ids 29/30 authored 09-01) |
| never_reached | 1 | id 32, EOD flat after 254.5 min |
| cadence_suppressed | **0** | class-47 in WARN mode |

### 2026-09-03

| cause | n | detail |
|---|---|---|
| **planner_shape** | **22** | long scenarios authored without `arm.enabled` during a `trend`/`long` plan |
| **host outage + post-boot blindness** | **1 window** | 113m51s down (12:24:33→14:18:24) + ~50 min blind after the boot; covers the day's high and the last 2h of NY entry time |
| parked(death) | 1 | arm 37; counterfactual **+38.75 pts** |
| never_reached | 1 | the only arm-enabled long, missed by **8.2 pts** |
| marketable_guard | 1 | id 36 |
| gate_min_sl | 5 clusters (18 events) | ledger negative |
| gate_rr_at_arm | 2 | net −64.4 pts |
| cadence_suppressed | **0** | one cooldown, fast-market-exempted |

### Baseline — the 7 days before 09-02

| day | positions | Σ pnl_corrected | arms authored | filled | fill rate | open-intent |
|---|---|---|---|---|---|---|
| 08-27 | 3 | (1 UNRESOLVABLE excluded) | 5 | 1 | 20% | 0.68% |
| 08-28 | 4 | — | 6 | 3 | 50% | 0.00% |
| 08-30 | 4 | (3 seam + 3 unresolvable excluded) | 4 | 1 | 25% | 0.00% |
| 08-31 | 6 | (3 NULL-corrected excluded) | 7 | 2 | 29% | 0.00% |
| 09-01 | 6 | +212.00 | 8 | 2 | 25% | 0.47% |
| **09-02** | **4** | **−381.50** | **4** | **0** | **0%** | 1.58% |
| **09-03** | **1** | **−140.00** | **3** | **1** | **33%** | 2.43% |

**[A] The fill rate on the two audited days (1 of 7 = 14%) is low but inside the historical band
(20–50%). What actually collapsed is arm SUPPLY: 5–8 arms/day across the baseline, 4 and 3 on the
audited days.** Combined with the arm-enablement asymmetry in §4, the constraint is upstream of
every gate — the planner is authoring fewer resting orders, and granting them almost only to
shorts.

---

## 12. THE DEFECT REGISTER

Every defect found, with the code path. None was acted on — this audit is read-only.

### Blocking the two days

| # | defect | evidence | path |
|---|---|---|---|
| **D1** | **HOST outage, 113 min 51 s** — *not a system defect; recorded for attribution* | `who -b` / `last reboot` → system boot 2026-09-03 14:18; NUL hole at log offset 779709; last verified act `log_events` 25754 @12:24:33 CT | WSL2 host |
| **D2** | **NT8 feed did not reconnect after the boot** | 46 `🚨 FEED DOWN` lines on 09-03; episode 1 runs 14:28:39→14:58:28, newest-bar age reaching **40m0s** | `trader/auto_trader.go:53` |
| **D3** | **~50 min of post-boot blindness with no alert** — 3 × `NT8 TCP link DOWN — NEW entries BLOCKED`, 3 × `no_balance_frame` (equity 0), next decision cycle not until **15:08:51** vs 12:22:44 | the real defect in the outage story; the NY flat is 14:45 | dead-man watchdog + balance-frame wait |
| **D4** | **Planner arms shorts, not longs** — 4.3% vs 44.4% arm-enablement on 09-03 while the plan bias was `long`/`trend` | §4 | planner output; `plans.doc.scenarios[].arm.enabled` |
| **D5** | **09-03 NY v7 authored a LONG and a SHORT on the identical level 29543.75; both confirms became true at 11:58 CT; only the short carried an arm** | agent D4 section | the same |
| **D6** | **71% of arm-enabled scenarios (34 of 48) were authored at a price the tape never reached during that version's own life** | agent D4 section | the same |
| **D7** | **On 09-02 NY, six versions each carried a short arm at 29317.25 while the whole calendar day's high was 29255.00 — 62.25 pts below.** Unreachable by construction | agent D4 section | the same |
| **D8** | **Wake storm** — on 09-02 NY, 11 of 12 plan versions were woken by the identical never-clearing condition (`seated OB(bull)·4h invalidated: close X below 29246.25`) | agent C5 section | `auto_trader_wake_levels.go` |
| **D9** | **Structural 105-min wake blackout every weekday** — NY `WindowEndCT`=14:45 and ASIA `ReadCT`=16:30, so `inSessionReadWindow` is false for all of 14:45–16:30 | agent C5 section | `auto_trader_planner.go` |
| **D10** | **The 15-min stale-working reaper is a hard time-stop on every placed arm** regardless of link health — it killed arms 29/30/31 across a restart | `armed_executor.go:1021-1032` | — |

### Stop and R:R construction

| # | defect | evidence |
|---|---|---|
| **D11** | **591's stop was 8.3× its own invalidation distance** — stop 29351.63 (66.63 pts from entry) while the scenario's own invalid is "a 5m close above 29293.00 ONH" (8.00 pts). The ATR floor overrode the structure the scenario was built on |
| **D12** | **591's arm was placed 15 pts wider than the plan authored** — plan `09-03 NY v2 S1 arm.stop = 29340` (55 pts); ledger `stop_px = 29351.63`; the SL is recomputed from the live ATR at placement, not taken from the plan |
| **D13** | **The R:R gate evaluates against `current_price_snapshot`, not the fill** (`kernel/engine_position.go:140,187`). Positions 587 and 589 filled 9.75 and 10.x pts away from the price the gate scored, so the R:R that passed is not the R:R that traded |

### Data integrity and observability

| # | defect | evidence |
|---|---|---|
| **D13b** | **`BackfillExcursions` has no automatic trigger** — it is reachable only from `cmd/excursions/main.go` (manual CLI) and one test, never from boot. Eleven boots since 09-03 10:28 all printed `backfilled=0`, so the 11 pre-wave positions stay unrecorded until an operator runs it | coverage gap |
| **D14** | `armed_orders.created_at` is stored with a **CT offset** and `updated_at` with **`+00:00` UTC** — same table, same subsystem, two clocks. Reading the raw column makes every arm look ~5h longer-lived than it was |
| **D15** | `plans.created_at` is **UTC** while `plan_lifecycle_log.at` is **CT** — sibling tables of the same subsystem disagree on the wall clock |
| **D16** | `planner_read_facts.bias_ai` and `bias_tree` are **empty in all 17 rows**, and `plan_id=''`/`version=0` in all 17 (`persistReadFacts`, `auto_trader_planner.go:2420-2453`). The AI and tree bias labels the dispatch asked for **do not exist in the store** |
| **D17** | `planner_read_facts` has **zero rows for 2026-09-02** — no `bias_regime` figure of any kind exists for that day (n=0), so 09-02's regime cannot be audited |
| **D18** | `plan_lifecycle_log` holds **2 rows all-time**; the D3 fix that creates it landed 2026-09-03 17:40:18 CT. 09-03 LONDON v1 went dormant **twice** and neither death nor the re-arm in between is in the database |
| **D19** | **38 of 52 plan versions** record their trigger as the bare token `level_event` — *which* level woke the planner is never persisted |
| **D20** | Until commit `4e901261` (09-03 17:40:18 CT), `UpdatePlanLifecycle` **overwrote `plans.trigger_reason`** with the dormant/flip marker, destroying the authoring trigger. This is why §6's version table shows death markers in the trigger column |
| **D21** | **`data/nofx_<date>.log` rotates on PROCESS START, not on date.** Position 591's close (09-03 09:20:45 CT) is at line 85591 of **`nofx_2026-09-02.log`**. Any grep scoped by filename silently misses events |
| **D22** | **`nofx_2026-08-27.log` is 2,049,637,081 bytes** (2.0 GB) against 0.6–13 MB for every other day — 180,139 lines in the first 40 MB are empty-payload `📡 armed order_update frame` spam |
| **D23** | **The min-SL refusal line prints the raw ATR where the threshold belongs** (`kernel/min_sl.go:62`), so a correct refusal reads as an arithmetic error. All 34 were correct |
| **D24** | **A min-SL-rejected decision is not persisted with its prices** — the retry overwrites the record with the final `wait`, so the refused set's exact counterfactual is unrecoverable from the store |
| **D25** | `armed_orders` rows for positions **582 and 585 were overwritten in place** by later re-arms (`UNIQUE INDEX idx_armed_orders_plan_scenario`), so their brackets are gone |
| **D26** | `source='reconcile'` rows (584, 586, **591**) carry `entry_time` = **materialization time, not fill time** (`trader/ninjatrader/reconcile.go:412` sets `EntryTime=nowMs` after a 60s grace) |
| **D27** | Entry fills exist in `trader_fills` **only for `source='system'`** positions; `armed_entry` and `reconcile` positions have exit fills only |
| **D28** | `WakeCadenceBootLine` (`trader/class47_wake_cadence.go:222`) claims the cutoffs govern `LEVEL_EVENT`/`structure_mss` wakes only, but `maybeWakePlannerOnMSSAt` does not honour that — the boot line misdescribes the running rule |
| **D29** | `collectLevelWakeCandidates` emits **no log line** when it returns zero candidates (`auto_trader_wake_levels.go:263-265`) and none when the bars provider is nil — a silent wake path |
| **D30** | Boot-time rev-mismatch `🔐 TRADING REFUSED` fired **twice** in the window (09-02 00:10:21, 09-03 23:11:59) — two cutovers shipped the wrong binary |
| **D31** | **8 of 21 big-move windows drew no wake line at all**, 6 of them sitting on a plan **190–370 minutes stale** |

---

## 13. D6 — THE ONE TRADE, in full

**Arm 35 → position 591.** NY S1, short, entry 29285.00, stop 29351.63, target 29144.50.
Armed 09-03 09:02:54.66 CT, filled 09:05:14, exited 09:20:45 @29355.00, **−140.00**, grade A.

**Why it was authored.** Plan `2026-09-03:NY` v2 (08:45:05 CT), bias `short`/medium,
`day_type: "balance"`. Scenario S1 in the plan's own words:

> "A retest of 29285.00 OR-H stalls after the 08:30 failure; short the touch."
> `invalid`: "A 5m close above 29293.00 ONH voids the fade."
> `arm`: `{enabled: true, entry: 29285, stop: 29340, target: 29144.5, wait_confirm: true}`

The plan's reasoning names the premise: *"I lean short fades unless 29285 is reclaimed… S1 carries
a resting short at 29285 because R:R clears the min-stop gate."* The machine regime on that same
read was **`up/NORMAL`**, and the day was called **"balance"** on a day that ran 510 points.

**Why it filled.** Price genuinely traded 29285.00 at 09:05 CT. The fill is correct.

**What killed it.** A stop-out at the arm's own stop with 3.37 pts of slippage
(29351.63 → filled 29355.00). Not a market close, not a sync artefact.

**Where the stop came from.** Not the plan. The plan authored `stop: 29340` (55 pts). The ledger
carries **29351.6284728996** — the ATR floor recomputed at placement
(`stop_floor_pts` on the 09:06:54 read was **67.65** against `atr5m` 45.1). **[A] The stop was
widened 15 pts beyond the author's intent by the floor, and simultaneously sat 8.3× further out
than the scenario's own invalidation level (29293.00, 8 pts away).** The scenario said the idea
was dead at 29293; the stop let it run to 29351.63 before admitting it.

**What refused everything else that session.**

| time (CT) | event |
|---|---|
| 09:20:47 | arm 36 (short 29351.05) cancelled — price 29358.75 already above entry → **marketable_guard** |
| 09:15:31 | plan flips **long** (v3), `day_type: trend` — and authors long scenarios **without arms** |
| 10:49:57 | the day's only arm-enabled long (29481.05) goes live; **misses by 8.2 pts**; superseded 11:27:53 |
| 11:58:33 | arm 37 (short 29543.75) placed — the *short* half of v7's two scenarios on that level |
| 12:15:00 | v7 dies (2×5m close below 29502.25) → arm 37 disarmed, 36 min before price reached it |
| 12:24:33 | **HOST down for 113 min 51 s** (`who -b` → system boot 14:18) |
| 14:18:24 | host+process boot; **FEED DOWN ×31**, `NT8 TCP link DOWN — NEW entries BLOCKED`, `no_balance_frame` — blind for the remaining 27 min of NY |
| 14:45 | NY flat |

**The tape after the exit.** From 29355.00 the market went to **29585.00** (+230). Had 591 been
held it would have been −460 at the extreme; the stop-out was correct. Had the system flipped long
at 29355.00 it would have had **+230 points (+$460)** available on the day's continuation. It
proposed `open_long` zero times.

---

## 14. G — VERDICT

**[A]** Two days that closed −521.50 combined (09-02 −381.50 on n=4, 09-03 −140.00 on n=1, zero
exclusions) did not lose because the new gates were too tight. The eight-leg entry gate — strict,
one-open-position, invalidation-wired, R:R at execution, min-SL, no-chase, shadow map, direction —
**refused nothing at all across both days**, proven by the absence of any `entry_gate:*` counter in
`system_config` where four `rr`-class counters do exist. The two gates that did fire have negative
counterfactual ledgers: min-SL at **n=34 events / 11 clusters** would have been stopped out on its
own too-tight stops 18 times for **−431.2 pts**, positive only under an unattainable
perfect-exit assumption; R:R-at-arm at **n=7** nets **−7.03 pts** and is below the n=10 floor, so no
verdict is stated for it. **Gates account for roughly 5%.** Cadence accounts for **~0%**: class-47
ran WARN-first for all of 09-02 and enforcing on 09-03, and across both days it suppressed **zero**
wakes — its one enforcing firing was fast-market-exempted, and 43 and 8 wakes fired through the
`wake_min_interval_min` ceiling. What is left is planner shape and a defect. **Planner shape is the
majority, ~55%**: on 09-03 the system read the trend correctly — `up/NORMAL` on all 17 planner
reads, plan bias `long` with `day_type: trend` from 09:15 CT — and then authored its long scenarios
**without resting arms**, 1 of 23 (4.3%) against 8 of 18 shorts (44.4%), leaving the longs to a
decision loop that proposed `open_long` **zero times in 575 decisions**; v7 put a long and a short
on the *identical* level 29543.75 with both confirms true at 11:58 CT and armed only the short;
71% of all arm-enabled scenarios were authored at prices their own version never saw, and on 09-02
six versions carried a short arm 62.25 points above the day's high. **Something that is neither a rule nor the tape accounts for
~30%**, and it is two things, not one: the **WSL2 host rebooted at 14:18 CT** (`who -b`,
`last reboot`), taking the process down from 12:24:33 for 113 min 51 s — an infrastructure event,
not a defect of this system — and then the boot came up **blind**, with 31 `FEED DOWN` lines, the
dead-man watchdog logging `NT8 TCP link DOWN — NEW entries BLOCKED`, and `no_balance_frame` at
equity 0, so the next decision cycle after 12:22:44 was **15:08:51**; the NY flat is 14:45, so the
entire afternoon containing the day's high of 29585 had either no host or no data, and **the
second half of that — ~50 minutes of a live process unable to enter — raised no alert.** The
restart cadence (11 boot banners from 6 build trees on 09-03 alone, one landing mid-session
between NY v6 and v7; two rev-guard `TRADING REFUSED` cutovers) is a real hazard but is **not** the
cause of this gap. **No-setup accounts for ~10%**: the single
arm-enabled long of 09-03 missed by 8.2 points and one minute, and the five losses each stopped at
their own authored stop, three of them within ~9 points of the exact turning point before the
market ran 54–93 points their way.

**Gates whose ledger says "review": none.** min_sl at n=34 is doing its job; rr_at_arm is at n=7,
below the floor. **Every other leg has n=0 and cannot be assessed.**

**Can the system take a trade on a trending fast tape, as configured today? [A] No — not
reliably.** The rule that prevents it is **not a gate**. It is that the planner grants resting arms
almost exclusively to short scenarios, and the decision path — the only route a long has — did not
propose a single long during a 483-point rally. A fade-only *arm* set inside a correctly-identified
*long* plan is a planner-shape answer, and it is the answer here.

---

## 15. CSVs

All under `docs/superpowers/reports/2026-09-04-two-day-audit-data/`:

| file | contents |
|---|---|
| `trades_window.csv` | every closed position 09-01→now with plan lineage |
| `refusals_minsl_counterfactual.csv` | all 34 min-SL refusals with threshold check + 30/60-min excursions |
| `refusals_rr_counterfactual.csv` | all 7 R:R-at-arm refusals with full stop/target outcome |
| `minsl_rejects_raw.txt` | the raw refusal lines |
| `arms.csv` | every `armed_orders` row with cause class |
| `plans.csv` / `d6_plan_0903NY_v2.json` | plan versions, scenarios, condition-truth |
| `cadence.csv` | wake event ledger |
| `tape.csv`, `tape_moves_0903.csv`, `d6_bars_1m_0903.csv` | the tape |
| `baseline.csv`, `decision_intent_by_day.csv` | the 7-day baseline |
| `c6-*.csv` | 1B writer inventory and merge-vs-boot timeline |

## 16. Limits of this audit

- **[A]** `GET /api/config/resolved` and `/api/risk/gate-blocks` require an Authorization header
  this session does not have. Live knob values are taken from the process's own boot lines and
  from refusal messages that print the multiplier they applied — never from a file default
  presented as live. Knobs not established that way are named, not guessed.
- **[A]** `bias_ai` and `bias_tree` are empty in every `planner_read_facts` row (D16), and the
  table has no 09-02 rows at all (D17). The AI/tree bias-accuracy table the dispatch asked for
  **cannot be built** — only `bias_regime` and the plan doc's own `bias.direction` exist.
- **[A]** `trade_excursions` is empty (0 rows all-time) because the writer shipped 67 minutes
  after the last trade closed. MAE/MFE **do** exist for all 11 positions in
  `trader_positions.mae`/`mfe`; the 30-min after-exit excursions in §2 are a different measurement,
  computed by this audit from 1m bars and labelled as such.
- **[A]** The exact entry/stop/target of a min-SL-refused decision is unrecoverable (D24); §8's
  min-SL ledger therefore uses each refusal's own stop distance against realised excursion, which
  is stated in place rather than dressed up as a stop/target counterfactual.
- **[B]** Whether a re-plan would have fired between 12:15 and 14:18 CT on 09-03 cannot be
  determined — the process was not running to fire one.
