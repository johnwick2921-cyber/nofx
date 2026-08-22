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

## 5. E3 / E4 / E5

- E3 (live stored prompt with STRUCTURE line + gate-order doc): gate-order doc shipped (`docs/regime-wave/gate-order.md`, matches observed order). Live-prompt quote is pending a live market cycle — weekend cycles carry no market data (expected); the Sunday soak captures it.
- E4 (90-min soak across ≥1 session boundary): **armed** — collector starts Sunday 16:55 CT, samples through the 17:00 CT ASIA open (structure_json presence, STRUCTURE line, gate WARN classes, crashes) → `~/soak-g7/e4.log`.
- E5 (off-switch proof): each blocker has its own off-switch restoring pre-wave behavior, pinned in tests — `htf_veto=false` (store round trip), `transition_standdown=false` (trader test), `loss_streak_n=0` (trader test). G7/G3 are evaluation policies with env defaults (rollback = config, not redeploy).

## 6. #51 402-alert verdict

**FIRED + PERSISTED — no gap.** `day_plan_alerts` id 224: P0 `ai-payment`, event `ai402:2026-08-21T19:55`, created 14:55:33 CT, still unacked at check time → the persistent dashboard banner was live through the whole 31–33-call outage (auto-acks on first success). 0 `alert EMIT failed`. The missing "AI-402 OUTAGE START" journal line is journald flood suppression (35,094 messages suppressed in that window; the owner's `install-journald.sh` run landed after), not a code gap — the durable artifact is the row.

## 7. Found-not-fixed

- `docs/research/plan-card/*` (7 research files) and `docs/superpowers/reports/2026-08-21-shift-day-loss-forensics.md` are NOT in this repo — built from the dispatch text + `docs/market-regime-classification-en.md` §6.4 vocabulary; calibration constants are env-tunable (`STRUCTURE_SWING_K`, `STRUCTURE_MIN_SWING_ATR`, `STRUCTURE_MSS_BODY_ATR`).
- journald flood suppression can still hide latch log lines (rate limit raised, but NT8 floods ~58k lines/min) — the durable alert rows are the load-bearing surface.
- Pre-existing FE test exclusions + guardrail-master OFF (owner's dated choice) unchanged.

## 8. Owner decision queue

1. **G1 calibration**: on 08-21 the 3-swing standard vetoed 0/12. Loosening (`STRUCTURE_SWING_K=1`, lower `STRUCTURE_MIN_SWING_ATR`) would have let the 1h confirm earlier — the dispatch's "expected: most" was based on research values that are absent from this repo. Pick the calibration or ship as-is.
2. **LOSS_STREAK_N value**: shipped 4; the dispatch left the exact N to the research (absent) — confirm 4.
3. **Veto TF choice**: `HTF_VETO_TF=1h` shipped; 15m veto would catch more but overlaps G4's stand-down.
4. **08-22 replay + classification** pending Sunday's data; will complete E2 then.

## 9. Residual risk

- G4's stand-down opens only on real close-throughs — wick flushes (like 08-21's 10:45) don't open it. If the research's MSS definition includes FVG confirmation, it's absent from the repo and not implemented.
- The regime toggles ride the same Studio marshal/merge seam as day_plan (pinned by tests), but the FE Strategy Studio has no UI controls for `regime.*` yet — toggles are API/config-level until the Studio panel ships.

PR: see `gh pr create` output.
