# TOTAL END-TO-END INVESTIGATION — WHY ZERO TRADES (2026-08-19)
**Dominant cause: the bot HAS been proposing valid trades, but a clock-drift guard deployed 2026-08-13 converted every entry to `wait` because the WSL clock runs 2.5–7.5 minutes behind the NT8 feed — FIXED (signals now stamped with the feed clock, rev 1a6dcf74d0d7, boot 14:27:35 CT, goldens PASS).**

## Timeline of zero
- Fills daily Jun 2 → Aug 13. **Last fill ever: 2026-08-13 14:49:31 CT.** Zero fills Aug 14–18.
- The C2 clock-drift guard (`e2f2561e`) was deployed **the same day as the last fill**.
- Today: 464 decision cycles. Raw responses contained **7 real entry proposals** (5 long, 2 short) — ALL converted to `wait` by the drift guard (172s–452s behind). Stored decisions therefore showed "0 entries" — the prior census measured the post-guard record, not the model's intent.
- Compounding overnight: **139 cycles died 02:00–07:00 CT today (122 on Aug 12) with DeepSeek 402 "Insufficient Balance"** — the account drains every night.

## What the model actually said today (10 latest waits, verbatim-categorized)
- 9/10: "no active S1–S3 trigger" — the plan's only short required a rally to 29853/29919, ~280 pts away.
- 4/10: "poor R:R / chasing at the lows"; 3/10: "no trade inside sideways zone" (owner's line); 2/10: 12:00–13:30 lunch window; 5/10: "oversold, needs a reclaim that hasn't printed".
- 08:47:28 CT the model PROPOSED S1 long (stop 29630.75, TP 29919, 4.08:1, conf 62) — R:R PASS logged, then drift-guarded to wait. 11:08:31 an open_short met the same fate.

## A/B on the owner's 3 lines (10-cycle-equivalent, today's 3:1 cycles)
<PENDING — lean replay running>

## Pipeline verdicts (hop-by-hop, production data)
- bars→detectors→scorer→key levels→prompt: **COMPLETE** (S/D drop at levels_score.go:167/168 + max-8 cap verified).
- bars→regime→planner→plan→plan block→prompt: **COMPLETE** except decision_records.plan_id/version never written in production (positions do get the citation).
- calendar→slice→blackout→gate: **BREAK (fail-open)** — a missing/undecodable calendar slice yields NO blackout windows, entries allowed.
- owner edit→overlay→resolved plan→prompt→realign: **COMPLETE** (single mutation door; executor receives overlay-resolved plan).
- config→codec→row→hot-read→gate: **COMPLETE** (min conf 60, R:R 3.0, 2-contract cap verified live).
- model call→parse→decision→gates→order→fill: **COMPLETE** (the guard was the only converter to wait).
- fill→bracket/breakeven→exit→MAE/MFE→adherence→digest: **BREAK at digest** — MAE/MFE/adherence never reach the digest.
- registry→scheduler→entry gate→session flat→night mode; alerts→feed→banner→ack: **COMPLETE**.
- Wire post-fix: 33/34 calls finish_reason=stop since 11:39, 0 length; median completion 1320 tokens; median latency 67s, max 149s.

## Today's 3 best missed setups (MNQ range ≈ 29514–29850)
- 08:47 S1 sweep-reclaim long (4.08:1) — drift-blocked; honest note: would have stopped out (market fell).
- 09:28 pullback short ~29654 (declined, citing the owner's sideways-zone line) — would have run ~140 pts to 29514.
- 11:08 open_short (drift-blocked) — would have run ~90 pts to the session low.

## Is the plan itself any good? (judged vs VL-DAYPLAN-FULL-SPEC.md)
- Levels: planner picks from the ranked table, but today ALL 8 sat above price (ONL 29680.75 → EQ cluster 30092) on a day that opened below PDL; zero downside levels; ALL graded A (a 4-level EQ cluster within 3 pts, all A = grade inflation). Same pattern in Aug 15/16/17 plans.
- Scenarios: always 2 longs + 1 short, short always rally-rejection. On the actual breakdown day no scenario could fire; the planner's own death text ("15m below ONL kills all longs") fired ~09:00 and **nothing replanned** — Go's death check is all-levels-consumed only; death/flip text is display-only.
- Verdict: the planner writes balance-day plans every day; on trend days they describe a market nobody could trade.

## Partner differences (ranked, no partner data available)
1) Our day-plan scenario gating (rally-only shorts, no replan-on-flip). 2) The C2 drift guard (now fixed). 3) DeepSeek 402 nightly balance drain. 4) Owner's "no trade inside sideway zone" + sideways-oscillation lines. 5) WSL clock drift. The one artifact that would settle it: his full assembled prompt + gate log for a single shared minute.

## Verdict & actions
- Fixed and deployed: feed-stamped signals + guard now warn-only + 402 loud log (1a6dcf74d0d7).
- Owner: top up DeepSeek balance / enable auto-recharge; fix the WSL clock (wsl --shutdown or NTP) as belt-and-suspenders.
- Next code fix (sized): replan when the planner's flip/death text fires, and require a breakdown scenario when price opens below PDL.
