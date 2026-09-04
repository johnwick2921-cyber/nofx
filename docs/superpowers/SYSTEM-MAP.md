# SYSTEM MAP — every subsystem, its knobs, gates, windows, and boot lines

**Maintained file on dev.** Built from code @ `492d2067` (boot 8) + the settings registry (`web/src/guide/content/settings.ts`) + two censuses (`docs/superpowers/reports/2026-08-30-knob-census.md`, `2026-09-02-belief-census.md`). Every fact carries `file:line` at that revision; a moved line means the map section is stale — see the contract below.

## Maintenance contract (the update rule)

> **Every wave that changes a rule — a knob value, a gate leg, a window, a threshold, a refusal string, or a boot-line text — updates this map's section in the SAME commit.** A contract test greps this map for the boot-line text and fails if a code boot line has no matching line here (the text is the join key). Labels `[R]/[X]/[T]/[I]/[O]` change with evidence, per the legend.

**Research label legend** (from `docs/superpowers/reports/2026-09-02-belief-census.md:8-16`):

| Label | Meaning |
|---|---|
| `[R]` | researched-supported (cite source) |
| `[X]` | researched-CONTRADICTED (cite) |
| `[T]` | measured on own tape (cite report, n) |
| `[I]` | invented / doctrine, untested |
| `[O]` | owner-ruled |

Pending rebrand: `fix/rebrand-phase-1-2` (NOFX→VL Intelligent) is NOT merged at this revision — every `nofx`-named file/boot line below changes the day it merges; that wave MUST update this map per the contract.

---

## 1 · BARS — NT8 TCP → BarCache → kernel

**What it does:** NinjaTrader 8 (Tradovate) is the single live data source. The C# AddOn streams bars/account/positions/orders over TCP (`127.0.0.1:36974`, `provider/ninjatrader/tcp_server.go:34`); Go caches them in a ring `BarCache` and hands them to the kernel through `market.FuturesBarsProvider`. No NT8 → stale cache → no decisions.

- Flow: frames (`bars_historical`, `bar_update`) → `readLoop` → `barIngestCh` → drain goroutine → `BarCache` — `tcp_server.go:392-407,492`; cache keyed `"SYMBOL|TIMEFRAME"`, ring per key — `bar_cache.go:30-37`.
- Provider wiring: `wireFuturesBarsProvider` → `market.FuturesBarsProvider = server.BarCache().Get(...)` — `trader/ninjatrader/bars_market_bridge.go:17-33`; var declaration `market/futures_data.go:14`.
- Kernel reads the SVP snapshot via `FuturesBarsProvider(activeSymbol, AISVPBarInterval, AISVPBarCount)` — `kernel/engine_analysis.go:297`.

**Knobs (resolved default · source):**

| Knob | Value | Source | Label |
|---|---|---|---|
| `AISVPBarInterval` | `"1m"` | kernel/svp.go:46 | — |
| `AISVPBarCount` | 2000 | kernel/svp.go:47 | — |
| `DefaultBarCacheMaxBars` | 2500 | bar_cache.go:24 | — |
| auto-subscribe symbol | `"MNQ"` | tcp_server.go:431 | — |
| auto-subscribe TFs | 1m 3m 5m 15m 30m 1h 2h 4h 6h 8h 12h 1d 3d 1w | tcp_server.go:435 | — |
| auto-subscribe back | 2000 | tcp_server.go:443 | — |
| stale-bar grace | 15 s (`STALE_BAR_GRACE_S`) | kernel/stale_data.go:46 | [R] C2 feed-stamp fix |
| clock-drift tolerance | 60 s | kernel/clock_drift.go:29 | [R] |
| clock warn | 30 s (`CLOCK_WARN_MS`) | kernel/clock_health.go:38-46 | [R] |

**Gates:** stale-data (runtime) refuses with `"stale-data: freshest %s bar %dms old (>%dms) hint=%s"` and converts entries to `wait` — `stale_data.go:129-175`. C2 clock drift is WARN-only (signals are feed-stamped) — `clock_drift.go:58-83`.

**Boot lines:** `"tcp_server: listening"` `tcp_server.go:981` · `"wire_liveness"` `:1024` · `"tcp_server: hello handshake OK"` `:1713` · `"tcp_server: sent bars_subscribe"` `:1478` · `"tcp_server: feed status"` `:2028` · `"⚠️ %s %s: CME futures symbol but no NT8 bar provider wired; skipping"` `market/data.go:247`.

## 2 · LEVELS — detection, scoring, seating, roles

**What it does:** detects ~30 level kinds off the 1m bar tape, scores them (evidence type × HTF zone tier × size × freshness × anchors), seats up to 8 per side into the plan, and assigns roles (entry-trigger vs target-only).

**Level kinds — definitions with windows** (all CT; `kernel/levels*.go`):

| Kind | Definition | Source |
|---|---|---|
| PDH/PDL/PDC | prior calendar-day high/low/close (needs ≥900 closed 1m bars) | levels_multiday.go:147-160,225-229 |
| RTH-H/RTH-L | prior day's bars whose active session == NY | levels_multiday.go:95-103,163-169 |
| AS-H/AS-L | overnight Asia (session ASIA) of current session-day | levels_multiday.go:105-121,176-181 |
| LDN-H/LDN-L | overnight London of current session-day | levels_multiday.go:105-121,182-187 |
| ONH/ONL | composite overnight high/low = max/min(AS, LDN) | levels_multiday.go:188-196 |
| PWH/PWL | prior week Mon–Sun (needs ≥4320 bars) | levels_multiday.go:131-141,207-215 |
| PMH/PML | prior calendar month (needs ≥10080 bars) | levels_multiday.go:142-150,216-223 |
| RN (round) | multiples of 100/50/25 within ±proximityK×dATR | levels_intraday.go:17-48 |
| GAP | unfilled gap ≥ 1.0×ATR | levels_intraday.go:56-105 |
| **OR-H/OR-L** | high/low of the **first 5 minutes** after RTH open (08:30–08:35 CT) | levels_intraday.go:111-139 |
| **IB-H/IB-L** (+1.5×/2× ext) | high/low of the **first 60 minutes** (08:30–09:30 CT) | levels_intraday.go:140,171-183 |
| nPOC | prior-session POC not retraded (retires on bracket beyond ±0.25 tick), ≤10 sessions | levels_volume.go:249-305 |
| VWAP (±1σ, ±2σ) | session VWAP anchored at 17:00 CT roll; ±2σ emitted at 0.85 | levels_volume.go:33-66 |
| eVWAP | extended VWAP anchored 15:00 CT cash close | levels_volume.go:93-118 |
| POC/VAH/VAL | prior session-day 120-bin profile; POC=max-vol bin, 70% VA | levels_volume.go:129-232 |
| pdVWAP | prior session-day's session VWAP | levels_volume.go:313-335 |
| SETT | prior settlement ≈ prior session final 1m close (16:00 CT) | levels_volume.go:339-360 |
| MID-O | overnight midpoint (ONH+ONL)/2, 17:00→08:30 CT | levels_volume.go:365-389 |
| EQH/EQL | k=2 strict pivots clustered within 3×tick | levels_zones.go:31-70, levels_assemble.go:77-79 |
| SUPPLY/DEMAND | base ≤6 candles, bodies ≤0.5×ATR, departure ≥1.5×ATR | levels_zones.go:110-130 |
| FVG | 3-candle imbalance, unfilled; floor max(2×tick, 2.0 pt) | levels_zones.go:165-190 |
| IFVG / OB | inverse FVG; order blocks | levels_zones.go, levels_assemble.go:84-85 |
| SWG-H/SWG-L | recent 5m/15m fractal swings (k=2, min-move 0.25×ATR), ≤3 per TF/side, lookback 144/96 | levels_swing.go:22-34,57-170 |
| OWNER | sticky owner-set level | levels.go:65 |
| NWOG (weekly) | weekend gap Friday ≤16:00 → Sunday first print | weekly_bias.go:47-58 |
| IPDA (weekly) | trailing 20/40/60-day HH/LL | weekly_bias.go:58-63 |

Assembly order: MultiDay → Round → OR/IB → Gap → EQH/EQL → S/D → FVG → OB → Volume → Swing → nPOC, then same-kind dedupe within 1 tick — `levels_assemble.go:81-96`.

**Scoring weights** (`levels_score.go`): kind weights :87-122 (structural 1.0 · VWAP/POC 0.90 · ON/nPOC/SWG/VWAP±2σ/eVWAP/pdVWAP 0.85 · VAH/VAL/SETT 0.80 · AS/LDN/OR/IB/EQ 0.70 · MID-O 0.60 · Round/Gap 0.55 · zones 0.30 confluence-only · default 0.50) `[I]` · zone TF tiers 1.0/1.1/1.2/1.3 (:148-161) `[I]` · zone reversal bonus ×1.1 `[I]` · ConfluenceCap 3 (:192-203) `[I]` · zoneSizeMult ladder ≤0.3×ATR ×1.25 … >2.5 ×0.50 (:205-222) `[I]` · freshness ladders (anchor 1/.8/.6/.5 :359-372; zone 1/.6/.3/.15 :378-390) `[I]` · proximity band = proximityK×dATR (:414), default 1.5, owner retune 0.3 per-trader config (plan_lifecycle.go:16,23-29) `[O]` · confluence band 0.10×dATR (:415-418) · cluster tolerance 12 ticks = 3.0 pt (:678-685) `[I]` · tier-1 proximity 12 ticks (:256) `[I]` · DefaultMaxLevels 8 (:54) · MinSideLevels 3 (:753). **`[T]`-positive** (conformance 2026-09-04 corrected a census misread): swing seats DO improve turn capture — missed-turns 80.0/75.0/79.2% → 65.0/60.0/66.7% (grand-audit.md:74, PROVEN) — seats kept `[T]`.

**Roles:** five roles — `levels_role.go:24-29`; consumed / 3rd-touch / far-HTF → **target_only, never entry** (:28,107-118) `[I]`.

## 3 · PLANNER — plan authoring and sessions

**What it does:** the prompt (`kernel/planner_prompt.go`) instructs the LLM; the output parses into a plan document (`kernel/plan_doc.go`) with bias, levels, scenarios, confirms, arms; `ValidatePlanDocWithCaps` chains every write-site validator.

**Rules in the prompt** (resolved, current lines): bias tree — close>PDH→bull HIGH · PDH sweep+close back→bear MEDIUM · inside-day→close vs PDC LOW — `planner_prompt.go:158-160` `[I]` · STOP-DOING: acceptance without prior sweep+displacement = 0% win evidence (:658-660) `[T]` · HTF zones are confluence, never standalone triggers (:532-546) `[O]` · scenario mix follows regime+day_type (:706-707) `[I]` · `entry_mode=ce` default (:626) `[R/O]`. **Deleted by owner ruling 09-04 (item 9):** the NY-AM/premium-FVG killzone weighting lines and the Monday/Thursday conviction line (ex-:653-656) — killzone is a label the card shows, never a prompt rule; the outside-killzone adherence grade step-down went with them.

**Sessions** (`kernel/session_registry.go:87-126`, CT): ASIA 17:00→02:00 (Read 16:30, kz 19:00–23:00, disabled) · LONDON 02:00→08:30 (Read 01:30, kz 02:00–05:00, disabled) · NY 08:30→14:45 (Read 08:00, kz 08:30–11:00 + 13:00–14:45, enabled; **session end == EOD flat**, owner contract) `[O]`.

**Blackouts:** T1 red news ±15 min hard no-trade (`T1BlackoutMinutes=15` `calendar_blackout.go:14`, windows :23-39) `[O]` · T2 caution-only (:21-22) · lunch 12:00–13:30 CT (`no_trade_band.go:42`) `[O]` · first-5-min no-trade (:34-37) `[O]` · session gate `auto_trader_session.go:98-127`.

**Boot lines:** `"📜 prompt/validator contract: %d restrictions, all stated in prompt (class 38 guard)"` `prompt_contract.go:164-172` · `"no-trade band: first_n=%dm lunch=%s–%s …"` `no_trade_band.go:199-203` · `"void scope: session-day window · %s×%d · one resolver for prompt AND validator (parity)"` `void_scope.go:100-104`.

## 4 · VALIDATORS — REJECT-at-write law

All chained through `ValidatePlanDocWithCaps` — `kernel/plan_doc.go:588`. Each refuses authoring at write time:

| Validator | Knobs (resolved) | Refuses | Source · Label |
|---|---|---|---|
| breakdown/continue | `BD_MIN_DISP_ATR=1.0`, `BD_MIN_CLOSES=1`, `BD_MAX_LEVEL_DIST_ATR=5.0`, `BD_MIN_SL_ATR=1.0` | missing breakdown{}, wrong direction, level >5×ATR, **"a close came back across %.2f — the breakdown is void"** (owner entry law), no confirming close, displacement <1.0×ATR5m, pullback-only/wait_confirm/confirm/min-SL arms | breakdown_continue.go:43-93,213-284 · [T]/[I]/[O] |
| entry law | law table :33-88 | `fade_requires_touch`, `2x5m_reserved`, `sweep_leg1_requires_touch` (leg 1 needs a real sweep touch [O]), `sweep_leg2_requires_mss_or_1x5m`, fade stop <2 ticks beyond level | entry_law.go:153-216 · [O] |
| FVG entry | `FVG_ENTRY_MIN_DISP_ATR=1.5`, `FVG_CE_WIDTH_PTS=20` ("NQ gap sweet spot 20–80 pts"), gap floor max(2×tick, 2.0 pt) | displacement <1.5×ATR5m, gap < floor, CE band = max(0.5, 10% width) | fvg_entry.go:26-49,235-362 · [R] |
| min-SL | `MinSLATRMultDefault=1.5` (was 1.0, owner-ruled 2026-09-02 with citation), `MinSLTickClearance=2` | `"sl_too_tight: %.1f < %.1f×ATR (%.1f) — widen or skip"` | min_sl.go:40-68 · [O] |
| confirm staleness | `StaleConfirmATR=2.0×ATR5m` | a MET confirm farther than 2.0×ATR5m from ref_price is stale-MET | plan_confirm.go:118-177 · [I] |
| HTF veto | mode `1h\|cross\|4h`, default 1h | `"htf_veto: %s vs %s %s (%s)"` — 1h (and 4h in cross) blocks counter-trend entries; fail-open WARN | htf_veto.go:17-142 · [O] |
| structure | k=2, min-move 0.25×ATR, MSS body 1.5×ATR, MSS displacement 0.5×ATR5m | swing/MSS confirm conditions not met | structure.go:27-29, mss.go:22-30 · [T]/[I] |
| accepts | 2x5m needs 2 closes, 5m-close needs 1; `AcceptHoldMin=10` min | confirm not MET per rule | plan_confirm.go:52-115, scenario_facts.go:100-119 · [I] |

**Boot lines:** `"entry law: bd_min_closes=%d bd_min_disp_atr=%.2f mss_min_disp_atr=%.2f …"` `entry_law.go:93-96` · confirm-rule ledger at main.go:327.

## 5 · ARMS — resting orders at plan levels

**What it does:** turns plan scenarios into resting limit orders ("arms") at levels, gates them at arm time, places inside a tick band, manages their lifecycle (working → filled/rejected/cancelled), and re-arms after plan-version changes.

**Chain per tick** (`maybeManageArmedOrders` `armed_executor.go:188`): stop-anchor composition (0B) ~:390-420 → wait_confirm (:410-428, `"⚔️ arm %s leg %d wait_confirm MET (%s) — arming"` :426) → `armGateVerdictFor` :430 → `oneLiveArmGuard` :483 → `entryGateForArm` :510 → kind validation :530 → far-arm warn :535-542 → `UpsertArm` :560-620 → shadow-AB :630 → split-sibling stop-out cancel :639.

**Gate legs + refusals** (`armGateVerdictFor` :1316): invalid ArmSpec → err as-is (:1322) · direction not armable `"direction %q not armable"` (:1325) · plan bias `"against plan bias %q (plan_mode=direction)"` (:1335) · quality `"quality %s below min_scenario_quality %s"` (:1341) · R:R `"R:R %.2f below arm min %.2f (studio min_risk_reward_ratio)"` (:1353, one floor = `min_risk_reward_ratio`, default 3.0 `store/strategy.go:76`; ARM_MIN_RR env DELETED `[O]`) · min-SL `"stop %.2f too close (%.2f < %.2f = %.1f×ATR5m)"` (:1362) · HTF veto `"HTF veto: " + reason` (:1375). `oneLiveArmGuard` (:640, class-27 FIX 4): `"one_open_position: %s arm %s refused — position %d open …; no adds, no flips (owner ruling 2026-09-03)"` (:657-658). Marketable-side limits never placed: `"level accepted through — marketable, never placed"` (:958) `[R]` 08-30 incident.

**Re-arm rules** (`store/armed_orders.go` UpsertArm :171): working row → refuse rewrite (:194-198) · terminal + signal ≠ → re-authorize ONLY on plan-version change (MANUAL-CANCEL-WINS :244-250) · boot-sweep rows re-arm under same version (:254-265) · canonical side UPPER at write (class 28, :181).

**Knobs:** `ARM_PLACE_TICKS=100` (:34-42) `[T]` · `ARM_WORKING_STALE_MIN=15` (:122-130) `[T]` → cancel `"no order_update within stale window (reconnect/reconcile)"` (:1058) · `ARM_STOP_ANCHOR_MAX_ATR=3.0` (arm_stop_anchor.go:38) `[I]` provisional · `ARM_FAR_ATR_MULT=3.0` (arm_far_counter.go:36) `[T]` warn-first · `ARMED_CANCEL_ACK_TIMEOUT_MS=2000` (:1424) `[T]`.

**Boot lines:** `"⚔️ armed_orders=on place_band=%dt stale_working=%dm test_seam=%s arm_rr=%.1f (gate-at-arm only; market-entry floor %.1f unchanged) (resting limits fill at the authorized price; stale_reeval NOT applied)"` `auto_trader_dayplan.go:64` · `"🎯 arms: bias-coherent=warn · stop-entry=… · far-arm counter=on(%.1f×ATR5m) · ledger append-only=on"` `arms_boot_line.go:14/26`, main.go:430.

## 6 · ENTRYGATE — the one canonical gate, both seams

**What it does:** one gate (`trader/entry_gate.go` `EntryGate` :140) called by BOTH the arm seam (`entryGateForArm` :328) and the decision path (`entryGateForDecision` :388). Fail-open contract: a leg with missing inputs SKIPS (header :27-30).

**Legs in order + exact refusals:**

| # | Leg | Refusal string | Line |
|---|---|---|---|
| 0 | plan_mode STRICT (R4) | `"entry_gate: refused: strict — …"` (4 variants) | :160-172 `[O]` |
| 1 | direction vs plan bias | `"entry_gate: %s entry against plan bias %q (plan_mode=direction)"` | :179-182 |
| 2 | scenario-direction consistency (class 48) | `"entry_gate: %s entry cites scenario %s authored %s — direction mismatch (class 48)"` | :190-194 |
| 3 | INVALIDATION (arm path, ruling 2026-09-03) | `"entry_gate: scenario %s invalidated at %s (accepted through %.2f) — <reason>"`; unavailable → PASSES (no verdict is not a refusal) | :205-226 `[O]` |
| 4 | shadow map (0C) | `"entry_gate: scenario %s condition %s is SHADOW (0C) — authored + E8-scored, never placed on any path"` | :233-234 `[O]` |
| 5 | R:R at REAL execution price | `"entry_gate: R:R %.2f below floor %.2f at execution price %.4f (SL %.4f TP %.4f)"`; floor = `min_risk_reward_ratio` (default 3.0) | :237-257 `[O]` R1 single floor (bound MNQ strategy carries 2.0 — drift D-21, conformance 2026-09-04) |
| 6 | min-SL ×ATR5m | `"entry_gate: stop %.2f too close (%.2f < %.2f = %.1f×ATR5m)"`; mult = `MinSLATRMult` 1.5 | :262-268 `[O]` owner-ruled 2026-09-02 |
| 7 | one open position per instrument | `"entry_gate: %s entry refused: %s (one_open_position, owner ruling 2026-09-03); no adds, no flips"` | :277-280 `[O]` |
| — | NO-CHASE (WARN-first, refuses nothing, A24) | — | :288 |

**ATR5m source:** `armSeamATR5m(d.Symbol)` :425 → `market.ExportCalculateATR(kernel.AcceptanceBars(bars, "2x5m"), 14)` :300-317 — **never `kernel.PlanDATRFor`** (that stores the DAILY ATR; "one gate, two ATRs" bug, no-trade-rider 2026-09-03) `[O]`.

**Telemetry:** decision-path refusals → `entryGateDecisionTelemetry` :475 → `telemetry.IncGateBlock(at.id, "entry_gate")` :478; arm path → `store.IncArmRefusal` :469.

## 7 · EXECUTOR — order placement on NT8

**What it does:** places market/limit/stop entries over the TCP wire to the C# AddOn (NT8 SIM), sizes futures contracts, cancels/modifies, and processes order updates.

- Placement: `placeEntry` `tcp_trader.go:314` · `PlaceLimitEntry` :419 · `PlaceStopEntry` :482 → `tcp_server.go:1047 SendSignal`. `CancelOrder` :554 · `ModifyBracket` :562 · `MoveStopToBreakeven` via `SendMoveStop` :1078.
- **SIM-only:** `isAccountTradeable` `tcp_trader.go:292` — SIM (`Account.Simulation`) AND on `NT_ALLOWED_ACCOUNTS` if set; "The LIVE/funded account is never tradeable" (:307,324); enforced at entry :333, armed :438, stop-entry :501. `assertBoundAccount` pre-submit identity invariant (A1) :399,456,519.
- **Sizing:** `futuresOrderQuantity` `auto_trader_orders.go:31` — `round(notional / (price × pointValue))`, floor 1, cap `maxFuturesContracts=2.0` (:25, "researched 2" `[R]`); notional ceiling 20×equity (`futuresMaxNotionalLeverage=20.0`, auto_trader_risk.go:14 + engine_position.go:66-73) `[X]` per census.
- `telemetry.IncGateBlock(at.id, "boot_integrity")` blocks opens when `kernel.TradingRefused()` — auto_trader_orders.go:198-210.

**Boot lines:** `"🚀 AI-driven automatic trading system started"` `auto_trader.go:833` · `"⚙️  Scan interval: %v"` :837 · `"🔧 NinjaTrader position-reconcile started (anchors entry_price to NT8 avg + clears orphan rows)"` `reconcile.go:101`.

## 8 · EXITS — stop management, suspensions, EOD flat

**What it does:** breakeven and trailing moves, EOD/T1 flattening, dormant/re-arm lifecycle.

- Breakeven: `maybeMoveStopToBreakeven` `auto_trader.go:148`, trigger `breakeven_trigger_points` default **50 pts** when unset (:201, opt-in OFF; owner-ruled ON at +40 pt per census E1 — the saved strategy carries `breakeven_enabled:true`, drift D-3, conformance 2026-09-04) `[O]`. WARN `"🎯 auto-breakeven: %s %s +%.1f pts in profit → stop moved to breakeven (entry %.2f)"` :189.
- Trailing: `defaultTrailingATRMult=2.0` × ATR(14,5m) after breakeven — auto_trader_trailing.go:22-30, rails R-A ratchet / R-B never-below-entry :94-112 `[O]` (census E2 owner-ruled 2.0×ATR14).
- **SUSPENDED by default:** `exitMechsSuspended()` returns true — `exit_mechs_suspend.go:33-41` ("suspended 2026-09-02 pending MFE data (wave 1A)"; Round-7: worst exit family of 15, 567k backtests `[R]`); env `EXIT_MECHS_SUSPENDED=0` re-enables. Note: suspension contradicts the census's owner-ruled ON position (drift D-3 — which ruling stands is open; conformance 2026-09-04).
- EOD flat: `enforceEODFlatAt` `auto_trader_clock.go:476` — flat = session end − offset, half-day pull-in :484-493, **cancels ALL arms (ack-waited) BEFORE flattening** :517-524, log `"🕒 EOD-FLAT (%s): session close — flattening %d open position(s) via the trader close path."` :527 · `"🔒 EOD-FLAT: %d armed order(s) cancelled before flattening"` :520. T1 force-flat lead 2 min (:562, "research v5 C.5" `[R]`).
- NT8 OCO close → `position_close` frame → close-sync records realized P&L — `close_sync.go:156-166`, `"📕 NT position closed: %s %s qty=%.2f exit=%.2f reason=%s pnl=%.2f (owner=%s)"` :196.
- Flip/death → dormant + auto re-arm: `maybeRunSessionReadsAt` auto_trader_planner.go:190,315-330 (`"😴 plan %s %s v%d DORMANT — %s (entries blocked; auto re-arms when price closes back; replan budget untouched)"` :328); re-arm `"⚡ plan %s %s v%d REARMED — %s"` :287 `[O]`.

## 9 · RECONCILE — NT8 truth vs DB

**What it does:** every 20 s (`reconcileInterval` `reconcile.go:32`) reconciles DB positions/orders against NT8 TCP snapshots; fixes, freezes, or materializes.

- Entry-truth: stale `entry_price` → anchored to NT8 avg (:101 boot line) · orphan-clear: NT8 flat → close row (`orphanGraceMs=120s` :35, `flatGraceMs=60s` :41, `entryConfirmGraceMs=45s` :63) · qty divergence → A4 FREEZE after `reconcileDivergenceGraceMs=60s` :207-210 · untracked NT8 position → materialize OPEN row after 60 s :351-446.
- Netting fills (class 27): latest OPPOSITE-side fill within `nettingFillWindowMs=25s` is the real exit price, else `CloseReasonUnresolved` — netting_fills.go:31-56. Unknown-P&L reasons: `ReconcileFlat / Unresolved / TestSeam` — store/position.go:102-122.
- P&L integrity counter: `pnl_integrity_mismatch` gate block when recomputed-vs-recorded Δ > $0.50 — auto_trader_clock.go:810-818.

## 10 · P&L — columns, corrected law, expectancy

**What it does:** records realized P&L at close with the correction column; the expectancy package is a read model (never gates — expectancy/model.go:4-6).

- Columns (`store/position.go:128-200`): `entry_price`, `exit_price`, `exit_time`, `realized_pnl` :152, **`pnl_corrected` (nullable) + `pnl_correction_note`** :157-158 (corrections additive, original never edited), `fee`, `close_reason`, `source`, `mae/mfe` (wave 1A), `plan_id/plan_version/cited_scenario_id/plan_matched/plan_band/adherence_grade` :186-197.
- **Corrected-column law:** `pnl_corrected` only; NULL = UNRESOLVED, excluded AND counted — model.go:14-17 · sample-id law: every row claim names ids — `row_ids` :117-134 · `MinN=30` :37 · z=1.96 :41 · eras pre/post-0B (`Era0BStart` = 2026-09-02 07:49:06 CT) :83-90.
- Realized formula `(ExitPrice−EntryPrice)×qty×pointValue` — close_sync.go:156-166.

## 11 · WAKES — what re-plans the plan

| Wake | Trigger | Source · Label |
|---|---|---|
| scheduled session read | inside ReadCT window per session (ASIA 16:30 / LONDON 01:30 / NY 08:00 CT) | auto_trader_clock.go:98,126 → auto_trader_planner.go:190 `[O]` |
| death re-plan | plan death → capped replan (`replan_cap`, only deaths spend it) | auto_trader_planner.go:315-339 |
| MSS wake | 15m MSS event after plan birth; one per event | auto_trader_transition.go:146-197 |
| level-event wake | fresh zones/FVG/OB/iFVG/invalidation candidates | auto_trader_wake_levels.go:86-250 |
| fast-market | `\|Δprice\| > 1.5×ATR5m` since plan write (`FAST_MARKET_ATR=1.5` auto_trader_loop.go:80) | `[R]` |
| stale_reeval | superseded entry re-validated; drift ≥0.25×ATR14 discards (`discard_burn.go:38`) | `[T]` |
| kick channel | stale-dodge, post-exit one-shot | auto_trader_clock.go:890-929 |

Cadence governance (class 47): `WakeCutoffMinDefault=25` (:52), `WakeCooldownMinDefault=30` (:59) — `[R]` 7-day tape + live observation. **"wake-predicate" cutover: NOT FOUND** in production code (only a comment in `trader/wiring_gate_test.go:20`).

**Boot line:** `"⏱ wake cadence …"` main.go:340 (the full string lives in the cadence boot helper).

## 12 · CADENCE — clock and session rhythm

- Scan interval: `scan_interval_minutes` default **3**, min 3 — store/trader.go:28-29, api/handler_trader.go:451-453, agent/tools.go:2487-2488 `[X]` per census (never tape-tested).
- Cadence modes: `CadenceInterval` / `CadenceBarClose` — auto_trader_clock.go:42-50; main loop ticker `auto_trader.go:936`; bar-close gate :905; stale-dodge :921.
- Sessions (CT): ASIA 17:00→02:00 · LONDON 02:00→08:30 · NY 08:30→14:45 (see §3) — session_registry.go:83-117.
- Half-days: `half_days.json` (Labor Day 2026-09-07 12:00 CT etc.) — auto_trader_halfdays.go, halfDayCutoffMin auto_trader_clock.go:452 `[R]` CME sources cited.
- Calendar: live ForexFactory JSON + static T1 fallback `calendar_static_t1.json` — calendar/calendar.go:33-92.
- Closed-market backoff: 3 min in 10 s slices — auto_trader_loop.go:937,980-992.
- Daily report 21:00 local, risk check 4 h — agent/scheduler.go:37-52.

## 13 · SETTINGS — registry and resolved values

- **Guide knob registry:** `web/src/guide/content/settings.ts` — dayPlan 18 knobs :7, risk 22 :277, sessions 1 :575; `KnobSpec` fields `label, where, what, trader, consumer, range, systemDefault, recommended, whenToTouch, perSession` — guide/types.ts:56-68; "9 env knobs" callout :620-669 (ARM_MIN_RR=2.0, HTF_VETO_MODE=cross, HTF_VETO_TF=1h, FAST_MARKET_ATR=1.5, BD_MIN_DISP_ATR=1.0, FVG_ENTRY_MIN_DISP_ATR=1.5, INGEST_QUEUE_CAP=1024, AI_PLAN_MAX_TOKENS=65536, PERSIST_STALL_WATCHDOG_S=60).
- **Agent-side field catalog:** `agent/entity_field_catalog.go:3-113` (trader/model/exchange fields, editability, keywords).
- **Safe defaults + hard limits** (`store/strategy.go`): `SafeDefaultMinRiskReward=3.0` :76 `[R/O]` · `SafeDefaultMinConfidence=60` :83 `[O]` · `MinRiskReward=1.0` :54 · `MinConfidence=50` :60 · MaxRR 10.0 · `ClampLimits` applies them :196-224.
- Trader table defaults: `ScanIntervalMinutes default:3` :28 · `IsCrossMargin default:true` :48 · `CadenceMode ''→interval` :30-34 · `PositionMode ''→ai_watch` :37-46 · 8 deprecated leverage fields DEAD at runtime :59-67.
- **Resolved endpoint:** `GET /api/config/resolved` — api/server.go:152 → api/config_resolved.go (same resolvers the engine calls).
- **Boot line:** `"⚙ knob registry …"` main.go:429.

## 14 · UI — pages, guide, endpoints

**Pages** (web/src/pages/): AgentChatPage (chat: ticker, positions, trader status, messages) · BeginnerOnboardingPage · FAQPage · SettingsPage (exchange/telegram/model modals + ResolvedKnobPanel) · StrategyStudioPage (plan/risk/indicator editor) · TraderDashboardPage (charts, DecisionCard, DecisionAudit, pause, EmergencyFlatButton, PositionHistory) · PageNotFound.
**HeaderBar tabs:** Agent (Beta) · Config · Dashboard · Strategy · 📖 Guide — HeaderBar.tsx:112-140 (mobile duplicates ~:450-510).
**Guide:** 14 sections (welcome…expectancy) — GuidePage.tsx:24-40; drift check vs `/api/health` revision, 12-char prefix compare :455-477; `GUIDE_BUILT_REV` stamped by `web/scripts/stamp-guide-rev.sh` (never hand-typed) — types.ts:6.
**i18n:** en/zh/id — translations.ts:1; agent Go side zh/en — agent/i18n.go:3-83.
**Design system:** `--nofx-*` app chrome vars (index.css:18-66) + `--vl-*` Plan-Card tokens (theme/vl-tokens.css) + tailwind palette (tailwind.config.js:10-31). The rebrand branch collapses these to `vl-*` — update this line when it merges.
**Endpoints:** `GET /api/health` `{status,time,revision}` — api/server.go:625-631 · `GET /api/cutover-gate` — :430-434 → class33_cutover_gate.go:58 (5 legs: db_open_positions · api_positions · nt8_positions_snapshot · working_orders · planner_in_flight; `ready` only if ALL pass).
**Boot lines:** `"🖥 ui: served-by=go-static build=…"` (or STALE warning) — api/ui_serving.go:110-146, main.go:303-308.

## 15 · DEPLOY — units, lock, cutover, boot integrity

**Units** (deploy/): `nofx.service` (template, `ExecStart=__NOFX_DIR__/nofx-bin`, Restart=on-failure, journal-only — nofx.service:42-47) · `nofx-web.service` (Vite :3000) · user units `nofx-backup.service/.timer` (OnCalendar 05:00 + 17:30, Persistent) · `nofx-clock-guard.service/.timer` (OnCalendar `*:0/15`). Backup retention `KEEP_DAILY=14`, `KEEP_WEEKLY=8`, layout `~/nofx-backups/auto/{daily,weekly}` — nofx-db-backup.sh:17-23. journald dropin `SystemMaxUse=2G` — journald-nofx.conf:16-19.

**Main-tree lock** (`deploy/nofx-lock.sh`): `LOCK_DIR=~/nofx-main.lock.d` (:43) · atomic `mkdir` acquire (:73) · owner-scoped heartbeat every 120 s (:46) · STALE after 300 s, never "dead" (:45) · `check` rc 0 free / 1 held / 2 stale (:133) · `reclaim` refuses fresh heartbeats, appends history, rc 3 (:173-200) · `release` holder-only (:203).

**Cutover** (`leveltruth-cutover.sh`): FLAT-GATE (DB OPEN=0 + NT8 count=0 snapshot in last 5 min) → RELEASE marker = build sha → binary swap + `kill -9` (SIGTERM exits 0, no relaunch) → poll ≤90 s for `BOOT INTEGRITY OK`.

**Boot integrity** (`kernel/boot_integrity.go`): `NOFX_EXPECTED_REVISION` env wins, else first line of `deploy/RELEASE` (:86-98); prefix match vs embedded `vcs.revision` (:102-118); 3 prompt goldens re-rendered (:92-127 golden_selfcheck.go); latches `tradingRefused` (:122-158). **Boot lines:** `"🔐 BOOT INTEGRITY %s — rev %s%s · built %s · expected %s · goldens %s"` :73 · `"🔐 TRADING REFUSED — %s"` main.go:288 · `"🔐 No new positions will be opened until this is fixed and the bot is restarted."` :289 · banner `"║           🚀 NOFX - AI-Powered Trading System              ║"` :44 · `"🔑 JWT secret configured"` :141 (fires unconditionally; verify override via `grep JWT_SECRET .env`) · `"🛡 clock-guard [boot] rtc_vs_go=%s timer=%s last_check=%s%s warn_ms=%d tolerance_ms=%d resync=unavailable-no-root …"` clock_health.go:164 · `"🕰 clock-health [<tag>] go=… nt8_last_bar=… drift_ms=… timesync{…} tolerance_ms=…"` :80.

---

*Map generated 2026-09-04 from code @ `492d2067` + settings registry + the 2026-08-30 knob census and 2026-09-02 belief census, then aligned with the 2026-09-04 research-conformance corrections (D9 swing seats [T]-positive, min-SL [O], breakeven/trailing [O]-ruled-but-suspended, R:R 2.0-vs-3.0 drift). Drift found and recorded: `BD_MIN_CLOSES` 1 (was 2), `MinSLATRMultDefault` 1.5 (was 1.0), code breakeven default 50 when unset (owner ruling 40), OR = first 5 min (IB = first 60 min), no wake-predicate cutover in production code.*
