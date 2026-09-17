
### Population — session-days with a plan, last 30 trade dates (range = 1m high−low over the DefaultSessionRegistry window, own contract)

- session-days: 70 · measured (1m tape present): 68 · unmeasured: ['2026-08-15:NY', '2026-08-16:NY']
- plan rows: 397 = model reads 310 + marker rows {'planner_fail_closed': 33, 'demo_seed': 2, 'replans_exhausted': 5, 'owner_reset': 28, 'dormant': 12, 'rearmed': 5, 'e7_incident_kill': 2} (markers excluded from a/b/c, counted in d)
- range ≥ 250 pt: **23** session-days · < 250: 45 · median range (measured): 210.8
| session-day | range | 1m bars | reads | open→close |
| --- | --- | --- | --- | --- |
| 2026-09-14:NY | 562.00 | 102 | 8 | 28905.75→29439.50 |
| 2026-09-16:NY | 499.75 | 375 | 5 | 29386.75→29173.75 |
| 2026-09-01:LONDON | 473.75 | 390 | 6 | 29515.50→29111.50 |
| 2026-09-10:LONDON | 435.75 | 390 | 6 | 29477.25→29089.25 |
| 2026-08-20:LONDON | 432.50 | 390 | 2 | 29672.50→29350.50 |
| 2026-09-03:NY | 385.75 | 375 | 7 | 29249.50→29524.00 |
| 2026-08-28:NY | 374.75 | 375 | 7 | 29628.00→29503.00 |
| 2026-08-19:NY | 366.25 | 375 | 3 | 29715.00→29530.75 |
| 2026-08-23:ASIA | 355.75 | 540 | 4 | 29392.25→29164.00 |
| 2026-08-17:ASIA | 338.75 | 539 | 6 | 30078.75→29894.50 |
| 2026-09-01:NY | 315.50 | 375 | 5 | 29111.75→29117.75 |
| 2026-08-24:NY | 296.00 | 375 | 8 | 29235.50→29131.00 |
| 2026-08-24:ASIA | 271.25 | 540 | 5 | 29140.00→29262.75 |
| 2026-08-25:NY | 270.50 | 375 | 2 | 29295.25→29238.75 |
| 2026-08-30:ASIA | 270.00 | 540 | 3 | 29535.00→29516.75 |
| 2026-08-21:NY | 268.25 | 375 | 1 | 29471.00→29410.75 |
| 2026-08-20:NY | 267.75 | 375 | 2 | 29350.75→29311.25 |
| 2026-09-08:NY | 262.25 | 375 | 5 | 29644.50→29549.00 |
| 2026-08-18:NY | 256.00 | 375 | 1 | 29692.50→29597.25 |
| 2026-08-19:LONDON | 254.75 | 390 | 3 | 29557.00→29713.75 |
| 2026-08-27:NY | 254.25 | 375 | 5 | 29523.75→29633.00 |
| 2026-08-26:ASIA | 253.50 | 540 | 10 | 29500.00→29498.00 |
| 2026-09-15:LONDON | 253.00 | 390 | 2 | 29354.00→29420.75 |

### (a) A level AHEAD of price (eventual-move side, within ±k×DATR) — range ≥ 250

- reads measured: 73 (recorded input snapshots: 21); median excursion after read: 155.2 pt; dir(excursion)==dir(close): 67/73
| table | n reads | ahead-in-band YES | fraction | ran out (excursion > farthest ahead) | median overshoot | median table size |
| --- | --- | --- | --- | --- | --- | --- |
| seated 12-seat input (recorded) | 21 | 21 | 1.000 | 12 (0.571) | 221.7 | 12 |
| published doc.levels (all reads) | 73 | 72 | 0.986 | 41 (0.562) | 102.0 | 9 |
- seated per session-day (yes/reads): {'2026-09-10:LONDON': '6/6', '2026-09-14:NY': '8/8', '2026-09-15:LONDON': '2/2', '2026-09-16:NY': '5/5'}
- seated: reads with ≥1 HTF (1h/4h/D) level ahead: 15/21; band probe (seated |distance| > k×DATR): 0 reads
- doc per session-day (yes/reads): {'2026-08-17:ASIA': '2/2', '2026-08-18:NY': '1/1', '2026-08-19:LONDON': '3/3', '2026-08-19:NY': '3/3', '2026-08-20:LONDON': '2/2', '2026-08-20:NY': '2/2', '2026-08-21:NY': '1/1', '2026-08-24:ASIA': '3/3', '2026-08-24:NY': '3/3', '2026-08-25:NY': '2/2', '2026-08-26:ASIA': '7/7', '2026-08-27:NY': '2/2', '2026-08-28:NY': '5/5', '2026-09-01:LONDON': '3/4', '2026-09-01:NY': '2/2', '2026-09-03:NY': '5/5', '2026-09-08:NY': '5/5', '2026-09-10:LONDON': '6/6', '2026-09-14:NY': '8/8', '2026-09-15:LONDON': '2/2', '2026-09-16:NY': '5/5'}
- by trigger (n · doc ahead-yes · doc ran-out · seated n · seated ahead-yes · seated ran-out): scheduled: 30 · 30 · 16 · 4 · 4 · 3; level_event: 40 · 39 · 23 · 16 · 16 · 8; other: 3 · 3 · 2 · 1 · 1 · 1

### (a) A level AHEAD of price (eventual-move side, within ±k×DATR) — range < 250 (control)

- reads measured: 231 (recorded input snapshots: 88); median excursion after read: 92.0 pt; dir(excursion)==dir(close): 181/231
| table | n reads | ahead-in-band YES | fraction | ran out (excursion > farthest ahead) | median overshoot | median table size |
| --- | --- | --- | --- | --- | --- | --- |
| seated 12-seat input (recorded) | 88 | 88 | 1.000 | 19 (0.216) | 49.2 | 12.0 |
| published doc.levels (all reads) | 231 | 230 | 0.996 | 53 (0.229) | 66.0 | 10 |
- seated per session-day (yes/reads): {'2026-09-09:ASIA': '1/1', '2026-09-09:NY': '3/3', '2026-09-10:ASIA': '9/9', '2026-09-11:LONDON': '5/5', '2026-09-11:NY': '6/6', '2026-09-13:ASIA': '15/15', '2026-09-14:ASIA': '8/8', '2026-09-14:LONDON': '4/4', '2026-09-15:ASIA': '9/9', '2026-09-15:NY': '2/2', '2026-09-16:ASIA': '13/13', '2026-09-16:LONDON': '11/11', '2026-09-17:LONDON': '2/2'}
- seated: reads with ≥1 HTF (1h/4h/D) level ahead: 71/88; band probe (seated |distance| > k×DATR): 0 reads
- doc per session-day (yes/reads): {'2026-08-16:ASIA': '5/5', '2026-08-17:LONDON': '1/1', '2026-08-17:NY': '2/2', '2026-08-18:ASIA': '2/2', '2026-08-18:LONDON': '1/1', '2026-08-19:ASIA': '3/3', '2026-08-20:ASIA': '3/3', '2026-08-21:LONDON': '4/4', '2026-08-25:ASIA': '8/8', '2026-08-25:LONDON': '2/2', '2026-08-26:LONDON': '12/12', '2026-08-26:NY': '7/7', '2026-08-27:ASIA': '10/10', '2026-08-27:LONDON': '6/7', '2026-08-28:LONDON': '6/6', '2026-08-31:NY': '3/3', '2026-09-01:ASIA': '4/4', '2026-09-02:ASIA': '12/12', '2026-09-02:LONDON': '4/4', '2026-09-02:NY': '11/11', '2026-09-03:ASIA': '7/7', '2026-09-03:LONDON': '1/1', '2026-09-04:LONDON': '3/3', '2026-09-04:NY': '4/4', '2026-09-06:ASIA': '5/5', '2026-09-07:ASIA': '4/4', '2026-09-08:ASIA': '7/7', '2026-09-08:LONDON': '5/5', '2026-09-09:ASIA': '1/1', '2026-09-09:NY': '3/3', '2026-09-10:ASIA': '9/9', '2026-09-11:LONDON': '5/5', '2026-09-11:NY': '6/6', '2026-09-13:ASIA': '15/15', '2026-09-14:ASIA': '8/8', '2026-09-14:LONDON': '4/4', '2026-09-15:ASIA': '9/9', '2026-09-15:NY': '2/2', '2026-09-16:ASIA': '13/13', '2026-09-16:LONDON': '11/11', '2026-09-17:LONDON': '2/2'}
- by trigger (n · doc ahead-yes · doc ran-out · seated n · seated ahead-yes · seated ran-out): scheduled: 47 · 47 · 16 · 11 · 11 · 3; level_event: 183 · 182 · 37 · 76 · 76 · 16; other: 1 · 1 · 0 · 1 · 1 · 0

### (b) Scenarios whose target_chain was EXHAUSTED (price beyond the last target) before the session flat

| pool | scenarios | measured | malformed | exhausted | fraction | first target hit | median min→last | median last-target dist | median #targets |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| range ≥ 250 | 203 | 182 | {'last_target_behind_price': 21} | 88 | 0.484 | 140 | 33 | 69.8 | 3.0 |
| control | 684 | 596 | {'last_target_behind_price': 77} | 247 | 0.414 | 448 | 45 | 69.6 | 3.0 |
- big per session-day (exhausted/measured): {'2026-08-17:ASIA': '2/4', '2026-08-18:NY': '1/3', '2026-08-19:LONDON': '1/9', '2026-08-19:NY': '4/9', '2026-08-20:LONDON': '5/6', '2026-08-20:NY': '5/6', '2026-08-21:NY': '1/3', '2026-08-24:ASIA': '3/10', '2026-08-24:NY': '5/9', '2026-08-25:NY': '1/6', '2026-08-26:ASIA': '21/21', '2026-08-27:NY': '3/7', '2026-08-28:NY': '3/11', '2026-09-01:LONDON': '5/10', '2026-09-01:NY': '6/7', '2026-09-03:NY': '4/12', '2026-09-08:NY': '4/12', '2026-09-10:LONDON': '5/14', '2026-09-14:NY': '1/11', '2026-09-15:LONDON': '2/4', '2026-09-16:NY': '6/8'}

### (c) Candidate rule — nearest BEYOND-band HTF (1h/4h/D) level in the trend direction — range ≥ 250

- reads: 73; HTF universe source: {'DetectHTFLevels': 73}; median beyond-band HTF levels per read: 192
| trend def | trend counts | trended reads | level existed | fraction | reached before flat | hit rate | median min→hit | median dist (pt) | tf of level | role (recorded only) |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| s1_1h | {'range': 17, 'down': 12, 'up': 34, 'n/a': 10} | 46 | 46 | 1.000 | 9 | 0.196 | 1 | 298.5 | {'4h': 29, '1h': 14, '1d': 3} | {'(unassigned)': 46} |
| s1_4h | {'up': 22, 'range': 17, 'down': 21, 'n/a': 13} | 43 | 43 | 1.000 | 7 | 0.163 | 208 | 294.9 | {'4h': 24, '1h': 17, '1d': 2} | {'(unassigned)': 43} |
| s1_D | {'range': 46, 'up': 14, 'n/a': 13} | 14 | 14 | 1.000 | 5 | 0.357 | 1 | 295.9 | {'4h': 10, '1h': 4} | {'(unassigned)': 14} |
| plan_bias | {'up': 19, 'range': 17, 'down': 37} | 56 | 54 | 0.964 | 9 | 0.167 | 38 | 307.2 | {'4h': 37, '1d': 4, '1h': 13} | {'(unassigned)': 54} |

### (c) Candidate rule — nearest BEYOND-band HTF (1h/4h/D) level in the trend direction — control

- reads: 231; HTF universe source: {'DetectHTFLevels': 231}; median beyond-band HTF levels per read: 198
| trend def | trend counts | trended reads | level existed | fraction | reached before flat | hit rate | median min→hit | median dist (pt) | tf of level | role (recorded only) |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| s1_1h | {'range': 106, 'down': 49, 'up': 41, 'n/a': 35} | 90 | 90 | 1.000 | 8 | 0.089 | 87 | 320.8 | {'1h': 26, '4h': 60, '1d': 4} | {'(unassigned)': 90} |
| s1_4h | {'up': 29, 'down': 104, 'range': 48, 'n/a': 50} | 133 | 133 | 1.000 | 14 | 0.105 | 1 | 354.8 | {'4h': 102, '1d': 13, '1h': 18} | {'(unassigned)': 133} |
| s1_D | {'range': 133, 'up': 48, 'n/a': 50} | 48 | 48 | 1.000 | 5 | 0.104 | 45 | 308.1 | {'1h': 19, '4h': 29} | {'(unassigned)': 48} |
| plan_bias | {'up': 94, 'range': 38, 'down': 99} | 193 | 177 | 0.917 | 28 | 0.158 | 1 | 327.5 | {'4h': 96, '1h': 61, '1d': 20} | {'(unassigned)': 177} |

### (c′) Alternative rule — FARTHEST in-band HTF (1h/4h/D) level in the trend direction — range ≥ 250

| trend def | trended reads | level existed | fraction | reached | hit rate | median min→hit | median dist | tf | farther than doc table & trend==move | …of which reached |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| s1_1h | 46 | 46 | 1.000 | 12 | 0.261 | 11 | 281.6 | {'4h': 22, '1h': 21, '1d': 3} | 17 | 6 |
| s1_4h | 43 | 43 | 1.000 | 8 | 0.186 | 127 | 281.6 | {'4h': 24, '1h': 17, '1d': 2} | 17 | 5 |
| s1_D | 14 | 14 | 1.000 | 1 | 0.071 | 2 | 282.8 | {'4h': 9, '1h': 4, '1d': 1} | 6 | 0 |
| plan_bias | 56 | 55 | 0.982 | 18 | 0.327 | 18 | 270.5 | {'4h': 32, '1h': 18, '1d': 5} | 24 | 11 |
- ORACLE vs the recorded 12-SEAT table (recorded reads only): 21 reads, table ran out on 12; reads with ≥1 in-band HTF level on the move side beyond the seated farthest: 15; reads where ≥1 such level was REACHED: 12 ['2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@1', '2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@2', '2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@3', '2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@4', '2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@5']
- ORACLE (move direction known — lookahead, for sizing the miss only): reads with ≥1 in-band HTF level on the move side beyond the doc table's farthest: 61/73 (0.836); such levels total 881, reached 539; reads where ≥1 was reached: 52; e.g. ["2026-08-17:ASIA@1 ['EQH·1d@30076.75✓', 'EQH·4h@30076.75✓', 'SUPPLY·4h@30147.75✓']", "2026-08-17:ASIA@2 ['EQH·4h@29887.00✓', 'EQH·4h@29985.00✓', 'EQH·4h@30010.00✓']", "2026-08-18:NY:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@1 ['EQL·4h@29518.25✓', 'EQL·4h@29577.25✓', 'EQL·4h@29666.00✓']"]

### (c′) Alternative rule — FARTHEST in-band HTF (1h/4h/D) level in the trend direction — control

| trend def | trended reads | level existed | fraction | reached | hit rate | median min→hit | median dist | tf | farther than doc table & trend==move | …of which reached |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| s1_1h | 90 | 90 | 1.000 | 14 | 0.156 | 7 | 301.0 | {'4h': 57, '1h': 28, '1d': 5} | 47 | 9 |
| s1_4h | 133 | 133 | 1.000 | 40 | 0.301 | 18 | 293.2 | {'4h': 106, '1h': 25, '1d': 2} | 71 | 28 |
| s1_D | 48 | 48 | 1.000 | 4 | 0.083 | 128 | 292.8 | {'1d': 1, '4h': 24, '1h': 23} | 18 | 1 |
| plan_bias | 193 | 189 | 0.979 | 35 | 0.185 | 53 | 284.2 | {'4h': 105, '1h': 79, '1d': 5} | 98 | 22 |
- ORACLE vs the recorded 12-SEAT table (recorded reads only): 88 reads, table ran out on 19; reads with ≥1 in-band HTF level on the move side beyond the seated farthest: 52; reads where ≥1 such level was REACHED: 32 ['2026-09-09:NY:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@1', '2026-09-09:NY:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@3', '2026-09-09:ASIA:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@2', '2026-09-10:ASIA:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@2', '2026-09-10:ASIA:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@3']
- ORACLE (move direction known — lookahead, for sizing the miss only): reads with ≥1 in-band HTF level on the move side beyond the doc table's farthest: 193/231 (0.835); such levels total 2732, reached 868; reads where ≥1 was reached: 143; e.g. ["2026-08-16:ASIA@1 ['EQH·4h@30275.00✓', 'EQL·4h@30336.75', 'DEMAND·4h@30273.50✓']", "2026-08-16:ASIA@2 ['EQH·4h@30275.00✓', 'EQL·4h@30336.75', 'DEMAND·4h@30273.50✓']", "2026-08-16:ASIA@3 ['EQH·4h@30275.00✓', 'EQL·4h@30336.75', 'DEMAND·4h@30273.50✓']"]

### (c2) The crux — reads whose level table RAN OUT (excursion beyond the farthest ahead level): would the candidate rule's level have been there and been reached?

| pool · table | ran-out reads | trend def | trend aligned with move | level existed | reached | reached / ran-out reads | median min→hit |
| --- | --- | --- | --- | --- | --- | --- | --- |
| range ≥ 250 · doc | 41 | s1_1h | 10 | 10 | 4 | 0.098 | 39 |
| range ≥ 250 · doc | 41 | s1_4h | 9 | 9 | 5 | 0.122 | 249 |
| range ≥ 250 · doc | 41 | s1_D | 3 | 3 | 2 | 0.049 | 20 |
| range ≥ 250 · doc | 41 | plan_bias | 16 | 15 | 5 | 0.122 | 40 |
| range ≥ 250 · seated | 12 | s1_1h | 0 | 0 | 0 | 0.000 | n/a |
| range ≥ 250 · seated | 12 | s1_4h | 6 | 6 | 5 | 0.417 | 249 |
| range ≥ 250 · seated | 12 | s1_D | 0 | 0 | 0 | 0.000 | n/a |
| range ≥ 250 · seated | 12 | plan_bias | 1 | 0 | 0 | 0.000 | n/a |
| control · doc | 53 | s1_1h | 14 | 14 | 5 | 0.094 | 103 |
| control · doc | 53 | s1_4h | 21 | 21 | 6 | 0.113 | 49 |
| control · doc | 53 | s1_D | 5 | 5 | 2 | 0.038 | 23 |
| control · doc | 53 | plan_bias | 24 | 24 | 11 | 0.208 | 1 |

### (d) Replan budget on those days (class 35: only death_replan/owner_reread SPEND; level_event is FREE)

| pool | session-days | level_event re-reads (total/median/max) | spending reads | recorded counter used (sum) | cap | exhausted (counter ≥ cap) | would exhaust if level_events counted |
| --- | --- | --- | --- | --- | --- | --- | --- |
| range ≥ 250 | 23 | 40/0/7 | 1 | 1 | 4 | 0 [] | 8 ['2026-08-26:ASIA', '2026-08-28:NY', '2026-09-01:LONDON', '2026-09-03:NY', '2026-09-08:NY'] |
| control | 45 | 184/4/15 | 0 | 0 | 4 | 0 [] | 23 ['2026-08-25:ASIA', '2026-08-26:LONDON', '2026-08-26:NY', '2026-08-27:ASIA', '2026-08-27:LONDON'] |
- range ≥ 250 — era split: pre-class-35 (< 2026-09-01): 15 session-days, exhausted-marker rows on 0 [], level_event re-reads 10; post (recorded counter): 8 session-days, counter-exhausted 0, counter used total 1, level_event re-reads 30, exhausted markers 0
- control — era split: pre-class-35 (< 2026-09-01): 19 session-days, exhausted-marker rows on 4 ['2026-08-16:ASIA', '2026-08-25:ASIA', '2026-08-26:LONDON', '2026-08-26:NY'], level_event re-reads 47; post (recorded counter): 26 session-days, counter-exhausted 0, counter used total 0, level_event re-reads 137, exhausted markers 0
- big per session-day: {"2026-08-17:ASIA": {"reads": 6, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-18:NY": {"reads": 1, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-19:LONDON": {"reads": 3, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-19:NY": {"reads": 3, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-20:LONDON": {"reads": 2, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-20:NY": {"reads": 2, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-21:NY": {"reads": 1, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-23:ASIA": {"reads": 4, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-24:ASIA": {"reads": 5, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-24:NY": {"reads": 8, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-25:NY": {"reads": 2, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-26:ASIA": {"reads": 10, "level_event": 4, "spending": 0, "counter": 0}, "2026-08-27:NY": {"reads": 5, "level_event": 2, "spending": 0, "counter": 0}, "2026-08-28:NY": {"reads": 7, "level_event": 4, "spending": 0, "counter": 0}, "2026-08-30:ASIA": {"reads": 3, "level_event": 0, "spending": 0, "counter": 0}, "2026-09-01:LONDON": {"reads": 6, "level_event": 4, "spending": 0, "counter": 0}, "2026-09-01:NY": {"reads": 5, "level_event": 1, "spending": 0, "counter": 0}, "2026-09-03:NY": {"reads": 7, "level_event": 5, "spending": 0, "counter": 0}, "2026-09-08:NY": {"reads": 5, "level_event": 4, "spending": 0, "counter": 0}, "2026-09-10:LONDON": {"reads": 6, "level_event": 5, "spending": 0, "counter": 0}, "2026-09-14:NY": {"reads": 8, "level_event": 7, "spending": 0, "counter": 0}, "2026-09-15:LONDON": {"reads": 2, "level_event": 1, "spending": 0, "counter": 0}, "2026-09-16:NY": {"reads": 5, "level_event": 3, "spending": 1, "counter": 1}}
