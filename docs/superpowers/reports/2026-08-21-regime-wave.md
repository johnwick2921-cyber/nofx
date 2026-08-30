# REGIME WAVE — G7→G8 (THE SHIFT-DAY FIX) — build report

Date: 2026-08-21 · Branch: `feat/regime-wave` · Base: 17cd52e2 (deployed)
Status: G1–G8 shipped · Cutover 1 (G7) shipped 14:45 CT · Cutover 2 (full wave) shipped 19:01 CT Friday · E4 Sunday soak ARMED · E2 day-2 (08-22) pending data.

## 1. Per-G ledger

| G | Commit | Scope | Tests |
|---|--------|-------|-------|
| G7 flip-eval freshness | `5c320ec2` | `kernel/flip_freshness.go` + fresh-gated `PlanDeathOrFlipSinceFresh` + prompt `staleTFLabel` + `FLIP_EVAL_MAX_STALE_S` (90s) | fresh fires / stale skips / no-closed-bars / replay Δ≈10 min |
| G2 structure detectors | `e8c17f97` | `kernel/structure.go` (swings → HH/HL/LH/LL → TREND, BOS/CHoCH/MSS/SWEEP), `decision_records.structure_json`, executor STRUCTURE line, planner real lines | goldens per event type + shift-day replay |
| G1 HTF veto | `f3f2d2cd` | `kernel/htf_veto.go`, `regime.htf_veto` (Studio, default ON), gate after min-conf | matrix + replay (0/12 blocked — honest) |
| G4+G4.6 transition stand-down | `48969caf` | `kernel/transition.go`, trader state machine, card chip, MSS = 4th planner wake-up | open/close ×4 paths + replay fail-open guard |
| G3 flip hysteresis | `75fe8456` | hold on the flip leg (death first) | hold / past-hold / death-wins |
| G5 consumed-level policy | `edcaf5b6` | demote + badge + planner "Consumed levels" section | demotion matrix + prompt |
| G6 loss-streak pause | `4888e647` | master-independent N-strike pause + banner + log_events | 4 cases |
| G8 watcher hooks | `480cdc8e` | structure line + `structure_conflict` question + dot + hysteresis | prompt/parse/replay |
| Cutover 2 boot ledger | `50ef497c` (HEAD) | regime boot ledger + `docs/regime-wave/gate-order.md` | — |

Full regression green at both cutovers: `go test ./...` · vitest 263/263 · npm build · C# diff `17cd52e2..HEAD` empty (0 files).

## 2. G7 staleness root cause + delta

Root cause (shift-day forensics §4): the flip/death evaluator and the executor prompt read the same NT8 BarCache via `market.FuturesBarsProvider`; on 2026-08-21 London both lagged 1.5–2 h (stored prompts show 15m latest close 01:30 at 03:54 CT; clock-drift WARNs 156–190 s). No freshness veto existed, so flips fired late and deaths never fired.
Fix: the evaluator now SKIPS a stale transition (log `flip_eval_skipped`, never guesses) and the prompt labels stale TFs with the same cap.
Replay Δ: London v3 flip line ("2x5m close below 29470.25", qualifies on the 08:05 close) fires at the first fresh eval ≈08:10 CT vs the actual 08:20:01 death log → **~10 min earlier**; the stale path at 08:20 skips instead of guessing.
G7 soak (62 min, 62 samples): 0 `flip_eval_skipped`, 0 `data stale`, 15m table fresh (`current 15m bar: FORMING (closes 15:00 CT)`); 0 crashes; all WARN classes pre-existing.

## 3. E1 — Cutover 2 boot block (quoted)

```
08-21 19:01:48 [INFO] nofx/main.go:227 🔐 BOOT INTEGRITY OK — rev 50ef497c5353 +dirty · built 2026-08-22T00:00:55Z · expected 50ef497c5353 · goldens PASS
08-21 19:01:48 [INFO] kernel/regime_ledger.go:11 🛡️ regime ledger: htf_veto=ON (Studio regime.htf_veto, default ON) · htf_veto_tf=1h (env HTF_VETO_TF)
08-21 19:01:48 [INFO] kernel/regime_ledger.go:12 🛡️ regime ledger: transition_standdown=ON (Studio regime.transition_standdown, default ON) · cap=45min (env TRANSITION_MAX_MIN)
08-21 19:01:48 [INFO] kernel/regime_ledger.go:13 🛡️ regime ledger: flip hysteresis hold=30min (env FLIP_MIN_HOLD_MIN)
08-21 19:01:48 [INFO] kernel/regime_ledger.go:14 🛡️ regime ledger: loss_streak=4 consecutive (Studio regime.loss_streak_n, default 4; 0=off) · pause=60min (env LOSS_STREAK_PAUSE_MIN)
08-21 19:01:48 [INFO] kernel/regime_ledger.go:15 🛡️ regime ledger: structure engine TFs=[5m 15m 1h] (5m/15m/1h, swing k=2, min-swing 0.25×ATR, MSS body 1.5×ATR)
08-21 19:01:48 [INFO] kernel/regime_ledger.go:16 🛡️ regime ledger: flip-eval freshness cap=90000ms (env FLIP_EVAL_MAX_STALE_S, default 90s)
```

## 4. E2 — replay through the finished stack (08-21, the 12 trades 533–544)

Honest machine truth, pinned by tests (`TestG1Replay_ShiftDayVetoes`, `TestG4Replay_NoFalseStanddownOnRealBars`, `TestG3FlipHold`, G7 replay):

| id | side | pnl | G1 (1h veto) | G4 (15m stand-down) | G6 (streak) | verdict |
|----|------|-----|--------------|---------------------|-------------|---------|
| 533–539 | SHORT | −315.5 | pass (1h unconfirmed: 2–3 swings, no 4-swing pair) | pass (no 15m close-through) | pass (streak < 4 at each) | happen |
| 540–541 | LONG | −181.0 | pass (no confirmed down) | pass | pass | happen |
| 542–544 | SHORT | −150.0 | pass (1h RANGING) | pass (10:45 was a wick, not a close-through) | pass (pnl 0 reset at 542) | happen |

Σ pnl_corrected delta vs the real −492.00: **0.00**. No entry is blocked on 08-21 — the day's 1h never confirmed a trend by the 3-swing standard and no 15m close-through ever occurred. The wave's day-1 value is the mechanism + honest detection (flip fires ~10 min earlier on a fresh feed), not retroactive saves.
**08-22 (addendum)**: no rows yet (last close 08-21 10:51 CT). Will run after Sunday's sessions and classify its losers by the forensics taxonomy (faithful-but-wrong / transition / post-flip).

### 4.1 G6 SEQUENCING — explicit replay with real close timestamps

Real close order (exit_time, CT) with EffectivePnL:

| # | close CT | pnl | streak (shipped rule: pnl < 0) |
|---|----------|-----|-------------------------------|
| 533 | 08-20 19:12:43 | −54.50 | 1 |
| 534 | 08-20 20:49:55 | −66.00 | 2 |
| 535 | 08-20 21:20:57 | −31.00 | 3 |
| 536 | 08-20 23:27:34 | +124.00 | reset |
| 537 | 08-21 01:35:40 | −79.50 | 1 |
| 538 | 08-21 01:45:01 | +30.50 | reset |
| 539 | 08-21 05:00:11 | −84.50 | 1 |
| 540 | 08-21 06:01:16 | −98.00 | 2 |
| 541 | 08-21 08:13:42 | −83.00 | **3** |
| 542 | 08-21 09:27:42 | **0.00** | reset (0 is NOT < 0) |
| 543 | 08-21 10:34:30 | −88.50 | 1 |
| 544 | 08-21 10:51:49 | −61.50 | 2 |

- **The streak never reaches 4**: max = 3 (539,540,541 by 08:13:42). 542's zero PnL resets it, so 543 = 1 and 544 = 2.
- **Session scope**: all 12 closes fall inside ONE CME session-day (08-20 17:00 → 08-21 17:00 CT). G6 is session-day-scoped by design — the LONDON→NY boundary (08:30) does NOT reset the streak, and it did not matter here: the run was broken by 542's 0.0, not by a boundary.
- **Pause windows that WOULD have opened** (hypothetical rules): (a) if pnl ≤ 0 counted: streak 4 at 542 close 09:27:42 → pause until 10:27:42 CT → would have refused 543 (entry 10:12:43 CT) → −88.50 saved; (b) if 0.00 were neutral: streak 4 at 543 close 10:34:30 → pause until 11:34:30 CT → would have refused 544 → −61.50 saved. Under the shipped rule: NO pause opens on 08-21.
- **E2 restated with ALL gates active incl. G6**: 0 blocked; Σ pnl_corrected delta vs −492.00 stays **0.00**.

### 4.2 NEAR-MISS TABLE (08-21, real 15m/1h bars — measured, not felt)

| Metric | Near-miss value | Binding constraint |
|---|---|---|
| CHoCH-up close-through | **71.0 pts short** (max post-flush 15m close 29417.50 @11:30 vs swing high 29488.50 @10:45) | no close ever broke the 10:45 swing high |
| Flip line 29470.25 (upside close after the crash) | **18.5 pts short** (max close 29451.75 @10:30; pre-crash 06:00–06:15 closes were above the line) | — |
| 15m 3-swing trend grade | **Δ=0.00 pts — the two lows were EQUAL (29220.25 twice)**, so the low-pair can never confirm UP or DOWN | equal-low refused TRENDING_DOWN all day; full-day pairs mixed → RANGING at every probe (05:00/08:45/11:30/15:00) |
| 1h 3-swing trend grade | max 3 confirmed swings at any entry instant (H 29399.75, L 29321.25, H 29539.75); the 4th (L 29220.25 @09:00) confirms ~11:00+ and the pairs are MIXED (up-pair highs, down-pair lows) | RANGING at every probe (03:54/08:49/10:47/14:45) |
| MSS displacement | the 10:30 up-bar body was **83.75 pts ≈ 2.9×ATR (≥ 1.5× threshold)** — displacement present, but no trend grade to attach the CHoCH to (the RANGING branch never emits CHoCH/MSS) | the trend grade, not the displacement, was the blocker |
| Intrabar beyond the flip line without a close beyond | **60 min** across 4 bars (06:30, 07:00, 10:30, 10:45) | — |

Read-out: the binding constraints on 08-21 were (1) the equal-low pair — a 0.00-pt margin — and (2) mixed swing pairs on both TFs. No calibration of swing window/min-move changes an equal low into a strictly-lower low; the day was genuinely two-sided on closes. The stand-down's only real trigger path that day was a 15m close ≥ 29488.50 (71 pts away).

## 4.3 Research files — cherry-pick CONFIRMED landed

The 7 research files existed on branch `docs/research-import-shift-forensics` (commit `f573b38f`, "docs: import plan-card design research (7)") but had never been cherry-picked onto the wave branch — the report's decision queue was correct. **Cherry-picked onto `feat/regime-wave` now**: `docs/research/plan-card/` holds the Final Build Plan v5, the Implementation Plan, FULL-SPEC, PLAN-CARD-DESIGN-SYSTEM, Strategy-Studio-Complete-Plan, Build Plan v3 and the config mockup (md + docx + html). The specs cite the smart-money-concepts library vocabulary (BOS/CHoCH/SWEEP) but do not hardcode different NQ swing constants than shipped — the near-miss table above is the calibration evidence either way.

## 5. E3 / E4 / E5

- E3 (live stored prompt with STRUCTURE line + gate-order doc): gate-order doc shipped (`docs/regime-wave/gate-order.md`, matches observed order). Live-prompt quote is pending a live market cycle — weekend cycles carry no market data (expected); the Sunday soak captures it.
- E4 (90-min soak across ≥1 session boundary): **armed** — collector starts Sunday 16:55 CT, samples through the 17:00 CT ASIA open (structure_json presence, STRUCTURE line, gate WARN classes, crashes) → `~/soak-g7/e4.log`.
- E5 (off-switch proof): each blocker has its own off-switch restoring pre-wave behavior, pinned in tests — `htf_veto=false` (store round trip), `transition_standdown=false` (trader test), `loss_streak_n=0` (trader test). G7/G3 are evaluation policies with env defaults (rollback = config, not redeploy).

## 6. #51 402-alert verdict

**FIRED + PERSISTED — no gap.** `day_plan_alerts` id 224: P0 `ai-payment`, event `ai402:2026-08-21T19:55`, created 14:55:33 CT, still unacked at check time → the persistent dashboard banner was live through the whole 31–33-call outage (auto-acks on first success). 0 `alert EMIT failed`. The missing "AI-402 OUTAGE START" journal line is journald flood suppression (35,094 messages suppressed in that window; the owner's `install-journald.sh` run landed after), not a code gap — the durable artifact is the row.

## 7. Found-not-fixed

- ~~`docs/research/plan-card/*` missing~~ — RESOLVED 2026-08-21 (cherry-picked `f573b38f` onto the wave branch; see §4.3).
- `docs/superpowers/reports/2026-08-21-shift-day-loss-forensics.md` is not in this repo — the E2 day-1 numbers come from the position store + stored prompts directly.
- journald flood suppression can still hide latch log lines (rate limit raised, but NT8 floods ~58k lines/min) — the durable alert rows are the load-bearing surface.
- Pre-existing FE test exclusions + guardrail-master OFF (owner's dated choice) unchanged.

## 8. Owner decision queue

1. **G1/G4 calibration — now with numbers**: the 08-21 binding constraint was the equal-low pair (Δ 0.00) + mixed pairs on both TFs, not the swing window or min-move. The stand-down's real trigger path was a 15m close ≥ 29488.50 (71 pts away); the 10:30 up-bar had MSS displacement (2.9×ATR) but no trend grade to attach it to. Calibration options (if wanted): a "close ≥ swing high" CHoCH variant in RANGING, or an extra pair-lookback — both are behavior changes to decide with the near-miss table in hand.
2. **G6 zero-PnL classification**: 542's 0.00 is the pivot of the day — counting it as a loser would have paused 543 (−88.50); neutrality would have paused 544 (−61.50). The shipped rule (pnl < 0) treats 0 as a reset. Confirm or change.
3. **LOSS_STREAK_N value**: shipped 4; the dispatch left the exact N to the research (now present, no hardcoded N found) — confirm 4.
4. **Veto TF choice**: `HTF_VETO_TF=1h` shipped; 15m veto would overlap G4's stand-down.
5. **08-22 replay + classification** pending Sunday's data; will complete E2 then (same table + near-miss columns).

## 9. Residual risk

- G4's stand-down opens only on real close-throughs — wick flushes (like 08-21's 10:45) don't open it. If the research's MSS definition includes FVG confirmation, it's absent from the repo and not implemented.
- The regime toggles ride the same Studio marshal/merge seam as day_plan (pinned by tests), but the FE Strategy Studio has no UI controls for `regime.*` yet — toggles are API/config-level until the Studio panel ships.

PR: see `gh pr create` output.
