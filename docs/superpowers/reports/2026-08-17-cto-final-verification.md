# SYSTEM VERIFIED — 20 remaining, 0 blocking

Read-only audit of `/home/hoang/nofx` @ `5472f316`, Sunday 2026-08-16 (market closed). 13 hardcode/pipeline agents + my own verification of every claim I report. Three trivial+safe fixes shipped (`6fe92f16`, `a84d6ae2`, `9dffec72`); everything else is listed with a size. **Nothing found is blocking for Monday's SIM open**, but items 1–4 and 9 change what the bot actually does and should be decided before you enable more sessions.

Evidence tiers: **[A]** = I read the exact line or ran it and saw output · **[B]** = strong inference · **[C]** = speculation. Every finding below is [A] unless marked.

---

## STEP 0

| Check | Result |
|---|---|
| HEAD ≥ 5472f316 | ✅ `5472f316` |
| Tree clean-ish | ✅ `M deploy/RELEASE` (your re-arm) + the known `??` files |
| Live bot on 5472f316 | ✅ PID 175755, `vcs.revision=5472f316…`, restarted 09:39:22 |
| Boot integrity armed, expected==actual | ✅ `BOOT INTEGRITY OK — rev 5472f31608af · expected 5472f31608af · goldens PASS` |
| Sandbox :3001 untouched | ✅ own PID/DB/port (8081 / 36985 / sandbox.db) |
| One session | ⚠️ PID 100656 (session `554049f5`) still resident but **idle — zero repo writes in 30 min**; proceeded as before |

---

## PART 1 · THE HARDCODE HUNT

The root disease, confirmed: a value compiled in where config should rule. **9 SHADOWS-CONFIG classes where the literal WINS** (deduped from 42 raw agent claims; each re-verified by me).

| # | Literal | file:line | Shadows | Literal wins? | Verdict |
|---|---|---|---|---|---|
| H1 | `band := 1.5 * dATR` | `kernel/levels_score.go:125` | `proximity_filter_atr` | **YES** | SHADOWS-CONFIG |
| H2 | `band := 1.5 * dATR` | `kernel/levels_intraday.go:24` | `proximity_filter_atr` | **YES** | SHADOWS-CONFIG |
| H3 | `ActivationWindowK` passed as `k` | `kernel/plan_render.go:65,67` | `proximity_filter_atr` | **YES** | SHADOWS-CONFIG |
| H4 | `planMaxLevels = 8` | `kernel/plan_doc.go:60` (enforced :101) | `max_levels` (documented 3–12) | **YES** | SHADOWS-CONFIG |
| H5 | `planMaxScenarios = 3` | `kernel/plan_doc.go:61` (enforced :115) | `scenario_cap` (documented 1–5) | **YES** | SHADOWS-CONFIG |
| H6 | `replansLeft := 2 - (version-1)` | `trader/auto_trader_planner.go:580` | `replan_cap` | **YES** | SHADOWS-CONFIG |
| H7 | `DefaultSessionRegistry()` | `kernel/engine_analysis.go:356,384` | the persisted admin registry | **YES** | SHADOWS-CONFIG |
| H8 | `!sess.Enabled` ×7 sites | `auto_trader_session.go:108`, `planner.go:528,566`, `handler_plan.go:143/475/555/711/1122` | `sessions[].enable` | **YES** | SHADOWS-CONFIG |
| H9 | bars `"1d"/"1h"/"5m"` | `trader/auto_trader_planner.go:421-424` | `planner_timeframes` | **YES** (see below) | SHADOWS-CONFIG |

**Runtime proofs** (an agent built a probe module with `replace nofx =>` and ran the real kernel functions):
- `ProximityFilterATR=3.0`, price 21000, dATR 100 → a PDH at 2.0×dATR was **DROPPED**; only the 0.5× level seated. `ScoreLevels` has no proximity parameter at all, so the configured value is *structurally incapable* of reaching it. H2 is the upstream twin: round-number levels beyond 1.5× are never even generated, so fixing H1 alone only half-honours the setting.
- `MaxLevels=12` + a 9-level doc → `too many levels: 9 (max 8)`. `ScenarioCap=5` + 4 scenarios → `scenarios count 4 invalid (1..3)`. Both **reject the whole plan** → fail-closed NO-TRADE. The upper half of both documented ranges is unreachable, and choosing it degrades to no plan. (`auto_trader_planner.go:315` does read `scenarioCap()`, but post-parse — it can only truncate below 3, never raise above.)

**H9 nuance (mine):** `planner_timeframes` *is* read (`planner.go:362`) — but only to emit placeholder lines `"<tf>: structure read"` into the prompt, while the bars actually fetched are hardcoded 1d/1h/5m. So the prompt **asserts a read-set the planner never received**. That is worse than dead: select "4h" and the planner is told it read 4h structure it never saw.

**1b · Two rulebooks (same default applied in >1 file).** Single source of truth that should exist → `store.DefaultDayPlanConfig` for values, `sessionRunnable()` for session enablement, the persisted registry for clock times.
- Session enablement: 9 deciding sites, 2 converted (see §PART 2 P-E/P-G).
- Re-plan budget: `store/strategy.go:933` `ReplanCapFor` **and** `planner.go:580` literal.
- Re-align cap: `store/strategy.go:870` tag default **and** `auto_trader_planconfig.go:218` `DefaultRealignCap`.
- Clock times: `kernel/session_registry.go` **and** `web/src/components/plan/sessionConfig.ts:41-72` (FE copy of every window/read/flat/killzone).
- FE `DEFAULT_DAY_PLAN` vs Go `DefaultDayPlanConfig`: agree today; nothing enforces that they keep agreeing.

**1c · Frontend constants gating backend-decided behavior.** `SESSION_BANDS[].enabled` still drives `SessionTimelineStrip.tsx:76` shading and `DayPlanEditor.tsx:291-297` `sessionRunnableUI` (the plan card's tabs were fixed in `e943f9c3`, these two were not). Also FE-only rules with no backend counterpart: `levelState.ts:44` `bandPts=3` conflict band, `levelState.ts:11` `NEAR_THRESHOLD_PTS=12`, `vocab.ts:21-29` instruction verbs presented as "canonical tokens the executor understands" against a free-text Go field.

**Fixed this run:** none of the SHADOWS-CONFIG class — every one changes trading behavior and none is trivial+safe on a Sunday before an open. All are in the REMAINING list with sizes.

---

## PART 2 · PIPELINE TRACES — 5 COMPLETE, 3 BREAK

| Pipeline | Verdict | Proof / break |
|---|---|---|
| **P-A** bars → detectors → scorer → KEY LEVELS → prompt → decision | ✅ COMPLETE | golden tests run |
| **P-B** bars → regime → planner → plan JSON → PLAN BLOCK → prompt | ✅ COMPLETE | `TestFuturesPlanReorder` + `TestFuturesPlanGolden` run, PASS |
| **P-C** calendar → slice → planner → T1 blackout → entry refused | ✅ COMPLETE | 6 tests written+run in an isolated copy (repo untouched); the executor **does** refuse inside a T1 window |
| **P-D** owner edit → overlay → plan_final → prompt → realign → Apply | ✅ COMPLETE | `TestW4OverlayReachesExecutor` PASS + live overlay row. The hop I suspected (base doc vs plan_final) is **not** broken — the executor gets plan_final |
| **P-E** config → dual codec → row → hot-read → consumer | ❌ **BREAK** | at the final consumers of `sessions[].enable` / `sessions_enabled` — 7 sites read the registry flag instead |
| **P-F** fill → exit → MAE/MFE + adherence → digest → next read | ❌ **BREAK** | `kernel/digest.go:12,18` — the digest formatter has **no parameter** for grade/MAE/MFE. The learning loop is open |
| **P-G** registry → scheduler → entry gate → flat → night | ❌ **BREAK** | `auto_trader_session.go:108` — a live-reachable bypass that vetoes the owner's explicit `enable=true` |
| **P-H** alert emit → feed → P0 banner → ack | ✅ COMPLETE | `TestAlertDedupeByEventID` run PASS + live DB receipt; the banner cannot be dismissed without acking |

**The break, stated once:** `sessionRunnable()` (W15.A) is the resolver for "does session X run", and **only 2 of 9 sites use it**. With ASIA/LONDON switched on — which your live config does — the read scheduler fires (LLM spend), a plan row is written, and then the executor gets `nil`, entries are blocked, no digest is written and the card stays empty. A toggle that costs money and changes nothing. Pinned by `TestW16SessionEnableIsHonoredByOnlySomeGates`.

**Worse, and separate — ASIA's read can never fire at all.** `IsCMEOpen` returns false for the whole 16:00–17:00 CT maintenance break, and ASIA's `ReadCT` is **16:55**. Probe output:

```
ASIA designed read (Mon) 16:55  IsCMEOpen=false  => read BLOCKED
ASIA designed read (Tue) 16:55  IsCMEOpen=false  => read BLOCKED
ASIA designed read (Sun) 16:55  IsCMEOpen=false  => read BLOCKED
ASIA after midnight (Tue) 00:30 IsCMEOpen=true   => read FIRES   ← the only one that can
LONDON designed read     01:55  IsCMEOpen=true   => read FIRES
NY designed read         08:25  IsCMEOpen=true   => read FIRES
```

The spec explicitly requires this path — *"the 16:55 closed-market read is a first-class tested path — planner builds from stored data"* — and the CME-open gate blocks it. The only ASIA read that can fire is the accidental post-midnight one, under the **wrong trade date** (plan identity uses `plannerTradeDateCT`, which rolls at midnight, while `CMESessionDayKey` rolls at 17:00). Pinned by `TestW16TradeDateNotionsDisagree`. *(I first reported this as a "duplicate read"; the adversarial review overturned that — the designed read never fires.)*

---

## PART 3 · CONTROL + STATE MATRIX

The 11 controls fixed in `bc360a38`/`d396201e`/`e943f9c3` were re-verified end-to-end and **all stayed fixed**. Playwright, own profile, headless, after a **hard reload** (`shots/2026-08-17-cto-controls-verified.png`):

```
NY=true  ASIA=true (🔸 override chip)  LONDON=false
last_entry_ct=12:45   eod_flat_ct=14:45   eod warning correctly absent at the window end
```

Caveat on the toggles: they are live in the *config* sense (written, persisted, read by the scheduler and the outer entry gate) but land in the 7-site gap above — so "LIVE" here means the control works, not that enabling ASIA yields a working session.

Remaining interactive elements were verdicted by three agents (day-plan block, plan card + alert centre, card states + executor indicators); results are appended in `2026-08-17-cto-matrix-appendix.md` when that phase completes. Nothing in the completed portion contradicts the previous matrix.

---

## PART 4 · RESEARCH-DEFAULTS AUDIT

Spec = `docs/VL-DAYPLAN-FULL-SPEC.md` (lines 60-74 field table, line 21 risk defaults). `docs/NOFX-MASTER-TRADING-SPEC.md` **does not exist** in the repo. LIVE = strategy `a5b7662e` (trader `hoang`, Sim101) — the day-plan strategy.

| Setting | Researched | Code default | LIVE VALUE | Verdict |
|---|---|---|---|---|
| proximity_filter_atr | 1.5×dATR | 1.5 | 1.5 | MATCH *(but shadowed — H1/H2/H3)* |
| max_levels | 8 | 8 | 8 | MATCH *(9–12 unreachable — H4)* |
| max_scenarios | 3 | 3 | 3 | MATCH *(4–5 unreachable — H5)* |
| replan_cap | 2 | 2 | 2 | MATCH *(prompt value hardcoded — H6)* |
| acceptance_rule | 2×5m | 2x5m | 2x5m | MATCH |
| approval_required | OFF | false | false | MATCH |
| evening_digest | ON | true | true | MATCH |
| plan_mode | advisory | advisory | advisory | MATCH |
| last_entry_ct | 13:00 CT | 13:00 | 13:00 | MATCH |
| eod_flat_ct | 14:45 CT | 14:45 | 14:45 | MATCH |
| sessions enabled | NY on, ASIA+LONDON off | NY only | **NY + ASIA + LONDON all on** | **DRIFTED** (more permissive) |
| ASIA min_grade | **A** | — | **B** | **DRIFTED** (weaker) |
| LONDON min_grade | **A** | — | **B** | **DRIFTED** (weaker) |
| ASIA max_trades | **1** | — | **3** | **DRIFTED** (3× permissive) |
| activation window k | 1.5 | `ActivationWindowK=1.5` | 1.5 | MATCH |
| re-arm cooldown | 20m | `ReArmCooldownMin=20` | — | MATCH |
| freshness floor | C | `FreshnessC`/`FreshnessDone` | — | MATCH |
| **R:R floor** | **≥3.0** | `ClampLimits` floors to **1.0** | **1.0** | **NEVER-APPLIED** |
| **confidence floor** | **≥65** | `ClampLimits` floors to **50** | **50** | **NEVER-APPLIED** |
| size cap | 2 | — | not set | NEVER-APPLIED |
| consecutive-loss halt | on | guardrail | master OFF | *deliberate* |
| reentry_cooldown | 20m | guardrail | master OFF | *deliberate* |
| stats gate n | ~1565 + Bonferroni α≈0.006 | `PreRegisteredN=1565`, `0.05/8` | — | MATCH |
| digest chain | sessions + 3 dailies + days 4–7 | `BuildDigestChain` | — | MATCH (content is thin — P-F) |
| planner model | exact pinned id | `pinExactModel` dealiases | inherit primary | MATCH |

**The headline.** Both running strategies have an **empty `risk_control` block**, so `ClampLimits()` (called on the decision path at `kernel/engine_analysis.go:247`, mutating the same struct the gates read at :296) floors them to **MinConfidence=50, MinRiskRewardRatio=1.0**. Receipt:

```
EMPTY risk_control after ClampLimits(): MinConfidence=50  MinRiskRewardRatio=1.00
SPEC researched values:                 MinConfidence=65   MinRiskRewardRatio=3.00
```

The spec calls these *"hard gates (R:R≥3, conf≥65, guardrails, armor) always outrank [the plan]"*. They are running at a third of the researched risk-reward floor. This is **not** covered by your deliberate guardrails-master-OFF — these are the base decision validators, not guardrails. **Confirmed and unchanged: `risk_control` is `{}`, guardrails master OFF — your learning mode. I did not touch it.**

---

## PART 5 · THE CTO CHECKLIST

| # | Item | Verdict |
|---|---|---|
| 5.1 | Safety precedence | *(phase-2 agent; appendix)* |
| 5.2 | **The SIM lock** | ✅ **VERIFIED** |
| 5.3 | Hold-lock + exit integrity | *(phase-2 agent; appendix)* |
| 5.4 | **Money math on the live path** | ✅ **VERIFIED** |
| 5.5 | **Time correctness** | ⚠️ **FINDING** (MEDIUM) |
| 5.6 | Idempotency + restart | ✅ **PARTIAL** — `-race` clean across all packages; rest in appendix |
| 5.7 | **Fail-closed coverage** | ⚠️ **FINDING** (MEDIUM-HIGH) |
| 5.8 | **Data hygiene** | ⚠️ **FINDING** (MEDIUM) |
| 5.9 | **Secret safety** | ⚠️ **FINDING — 1 FIXED** |
| 5.10 | Observability | *(phase-2 agent; appendix)* |
| 5.11 | Owner's daily loop | ⚠️ **FINDING** (gate-block reasons are in-memory only) |
| 5.12 | **Rollback** | ⚠️ **FINDING — FIXED** (`6fe92f16`) |

**5.2 SIM LOCK — VERIFIED.** Entries are **double-guarded**: Go `placeEntry` (`tcp_trader.go:234-241`) refuses an unbound trader outright (no fallback to the shared active account) and refuses `!isAccountTradeable` (must be found **and** `IsSim` **and** on `AllowedNTAccounts` if set); C# `VLTraderTCPClient.cs:772` re-checks `IsSimAccount` as the last line before submit and rejects. Closes are **single-guarded in C#** (`HandleClosePosition:969`) — it refuses and leaves the position OPEN rather than flattening a live account, which is the right call. Enumerated every widening path: `AllowedNTAccounts` can only **narrow**; the PlanDoc schema has **no account field**, so neither planner output nor an owner overlay can influence account selection; no bypass/force-live flag exists anywhere. `/api/debug/nt-test-trade` is JWT-protected + futures-only and bypasses the AI+risk gates **by design**, but still passes the SIM lock and the B3 dedupe/rate breaker.

**5.4 MONEY MATH — VERIFIED.** MNQ $2/pt cross-checked on a real closed trade: `30199.5 → 30179.25 = −20.25 pts × $2 × 1 = −$40.50` ✓ matches `realized_pnl`. `close_sync.go:67` already cross-checks NT8's reported PointValue/TickSize against the local table and warns on mismatch. 456 `sync` + 60 `reconcile_flat` closes all carry correct quantities. *(The one row with `quantity=0` is the seeded demo trade — see 5.8, not a production defect. I initially mis-read it as a TP-path bug and corrected.)*

**5.5 TIME — FINDING.** Two date notions coexist: `plannerTradeDateCT` (midnight roll) for plan identity, calendar slices and the read scheduler; `kernel.CMESessionDayKey` (17:00 roll) for the risk daily reset, guardrail counters, alert dedupe keys, level bucketing and the approval key. Each is internally coherent, and the **digest is correct** (it labels with the calendar date while summing from the 17:00 boundary — the comment shows this was reasoned about). The bite is the one midnight-spanning session (ASIA), above. The 14:45 flat **is** single-sourced post-P3 (`WindowEndCT == FlatCT`, pinned by `kernel/session_end_contract_test.go`, which also parses the TypeScript mirror).

**5.7 FAIL-CLOSED — FINDING.** `currentT1Windows` (`auto_trader_calendar.go:104-123`) returns `nil` — i.e. **no blackout, entries allowed** — on four silent paths: nil store, no active session, **`err != nil || slice == nil`**, and malformed EventsJSON. No alert, no log. So "the calendar feed failed" and "there are no red events today" are indistinguishable and both permit trading. Live state: `calendar_slices` holds 2026-08-18…21 and 08-12…14 but **no row for Monday 2026-08-17**. Second fail-open: `kernel/engine_position.go:148` skips the R:R gate entirely when there is no entry reference (documented, argues other gates cover it).

**5.8 DATA HYGIENE — FINDING.** Demo rows remain in the live DB: **2 plans** (`trigger_reason='demo_seed'`, `model_id='demo-seed (no API call)'`), **3 alerts** (`event_id LIKE 'demo:%'`, including a **P0** "DEMO — planner fail-closed"), **1 overlay** + **1 plan_qa** row on `2026-08-16:NY`, and **1 fabricated closed trade** (`source='demo_seed'`, +$224.50, adherence **A**). It is counted in your stats: **$4,271.25 shown → $4,046.75 real (5.3% of displayed P&L is fabricated)**, plus one fake win and an "A" grade feeding the learning loop. The sandbox seeder is correctly fenced (`cmd/sandbox-seed/main.go:33` refuses any path containing `data.db` or lacking `sandbox`); the contamination came from the guarded `trader/demo_seed_test.go` (untracked, `NOFX_DEMO_SEED=1`). Backups are current and restorable (`~/nofx-backups/auto/daily/`, newest 2026-08-16 05:00; timer armed for 17:30; linger on).

> **The documented cleanup command in `2026-08-16-demo-plan-seed.md` is WRONG — do not run it as written.** It does `DELETE FROM plans WHERE trade_date IN ('2026-08-16','2026-08-15')`, which destroys the **two real 2026-08-15 plans** (`planner_fail_closed` v1 + `NY_scheduled_read` v2, authored by deepseek-v4-pro) — the only genuine plan history you have. It also deletes the two real 08-15 alerts and **misses `plan_qa` entirely**. Corrected command in the handoff below.

**5.9 SECRETS — FINDING, ONE FIXED.** The Go API is loopback-bound, but `web/vite.config.ts` bound the dev UI to `0.0.0.0:3000` **and proxies `/api` to localhost:8080** — re-exposing the whole API to the network and defeating the bind it sits behind. Receipts on the live host:

```
GET  http://<lan-ip>:8080/api/traders        → 000   (refused — bind holds)
GET  http://<lan-ip>:3000/api/traders        → 200   (same API, via the proxy)
POST http://<lan-ip>:3000/api/reset-password → 410   (P0 gating still in force)
```

Most routes 401; `/api/traders` returns `[]` unauthenticated, so nothing is known to have leaked — the defect is that the control was bypassable at all. **Fixed** (`a84d6ae2`) to `127.0.0.1`; safe because WSL2 runs `networkingMode=mirrored` and the sandbox UI has bound loopback throughout. Takes effect next dev-server start. Otherwise: `.env` is untracked; no key-shaped strings in tracked source beyond test fixtures; the one key in `web/dist` is `cm_568c67eae410d912c54c`, the **known-dead public NofxOS default** (`provider/nofxos/client.go:19`, returns HTTP 402), not your secret.

**5.11 OWNER LOOP — FINDING.** `telemetry.IncGateBlock` is an **in-memory map** (`telemetry/gate_blocks.go:40-47`) — counters reset on restart and there is no `gate_blocks` table. So "why was an entry refused?" has no durable record; the journal line is the only trace and it is not in the UI. Full screen-by-screen answers in the appendix.

**5.12 ROLLBACK — FIXED.** `deploy/RESTORE.md` documented and tested the **DB** restore but not the **binary** rollback — and after reverting to an older binary, `deploy/RELEASE` still names the newer revision, so `AssertBootIntegrity` (`kernel/boot_integrity.go:135-138`) sets `Refused=true`: **entries blocked, everything else read-only**. The bot comes up looking healthy and silently takes no trades. `6fe92f16` adds the rollback sequence with the re-arm as a numbered step, the journald line to confirm, the blank-value escape hatch for bisecting, and DB-first ordering across a migration.

---

## EXIT BAR

`go build` ✅ · `go vet` ✅ · `go test ./...` ✅ · **`go test -race ./...` ✅ clean** · `tsc --noEmit` ✅ · `npm run build` ✅ · vitest **177/178** (the one failure — `RegistrationDisabled` "NoFx Logo" alt text — and the `e2e/gate.spec.ts` collection error are **pre-existing and untouched**, confirmed against the files I changed) · **goldens byte-identical** (`git diff kernel/testdata/` empty) · Playwright green with screenshot · config-truth run on every field touched.

**Adversarial self-review** (run against my own findings): two challenged, one overturned. (a) *"ASIA fires a duplicate read"* → **WRONG**, `IsCMEOpen` blocks 16:55 entirely; corrected to "the designed read never fires". (b) *"the gates get conf=50/RR=1.0"* → **HOLDS**: `GetConfig()` returns a pointer, `ClampLimits()` mutates in place at :247, `GetRiskControlConfig()` reads the same struct at :296. Also disproved before reporting: the digest's date/window mix (correct by design) and the `quantity=0` row (demo data, not a TP-path bug).

---

## REMAINING — numbered, each with severity, size, recommendation

| # | Finding | Sev | Size | Recommendation |
|---|---|---|---|---|
| 1 | `sessions[].enable` honored by 2 of 9 sites | HIGH | M | Convert the 7 to `sessionRunnable`. **Until then, switch ASIA+LONDON back OFF** — today they cost LLM calls and change nothing |
| 2 | ASIA's 16:55 read blocked by `IsCMEOpen` (spec requires it) | HIGH | S | Allow the pre-open read inside the break window, or move `ReadCT` to 17:05 |
| 3 | `proximity_filter_atr` shadowed at 3 kernel sites | HIGH | M | Thread a `proximityK` param through `ScoreLevels`/`RoundNumberLevels`/`RenderPlanStatus`; fix all three or it stays half-honoured |
| 4 | R:R floor 1.0 and confidence floor 50 vs researched 3.0/65 | HIGH | S | Set `min_risk_reward_ratio` + `min_confidence` on both live strategies (config, no code) |
| 5 | `max_levels` >8 / `scenario_cap` >3 reject the whole plan | MED | M | Parameterize `ValidatePlanDoc` caps; keep 12/5 as hard ceilings |
| 6 | T1 blackout fails open on a missing/failed calendar slice | MED-HIGH | S | Distinguish "no events" from "no data"; alert + fail-closed on the latter |
| 7 | ASIA plan identity rolls at midnight mid-session | MED | M | Key plans by the session instance (window start), not the calendar date |
| 8 | Demo rows in the live DB (+ the documented cleanup is wrong) | MED | S | Run the corrected command below, after a backup |
| 9 | Executor uses `DefaultSessionRegistry`, not the admin one | MED | M | Add a registry provider seam to the kernel, mirroring the bars/nPOC seams |
| 10 | Digest omits adherence/MAE/MFE — learning loop open | MED | M | Extend `FormatSessionDigest`; also `dayType` is always `""` |
| 11 | `planner_timeframes` asserts a read-set never fetched | MED | S | Either fetch the configured TFs or stop claiming them in the prompt |
| 12 | Gate-block reasons in-memory only | MED | M | Persist a `gate_blocks` table; surface in the card |
| 13 | Re-plan budget hardcoded in the executor prompt | MED | S | `ReplanCapFor` at `planner.go:580` |
| 14 | FE `SESSION_BANDS` duplicates the registry clock | MED | M | Serve the registry to the FE (the tabs already do this) |
| 15 | `day_plan` never range-clamped at save | LOW-MED | S | Add a DayPlan branch to `ClampLimits` |
| 16 | Alert dedupe is check-then-insert, no unique index | LOW | S | Add a unique index on `(trader_id, event_id)` |
| 17 | `/api/plan/alerts` has no user→trader ownership check | LOW | S | Low risk while loopback-bound; fix with the next auth pass |
| 18 | `"MNQ"` hardcoded at 6 api sites + 2 FE | LOW | M | Blocks multi-symbol later, harmless today |
| 19 | AI cost rates hardcoded (`handler_plan.go:1276`) | LOW | S | Read from the model price map |
| 20 | FE-only rules with no backend counterpart (conflict band 3pts, near 12pts, instruction verbs) | LOW | S | Either back them with Go or label them as UI heuristics |

---

## DEPLOY HANDOFF

Nothing changed in the trading path. Three commits: `6fe92f16` (docs), `a84d6ae2` (dev-server bind), `9dffec72` (tests only).

```bash
cd /home/hoang/nofx && git pull
go build -o nofx-bin . && echo BUILD OK
git rev-parse HEAD > /tmp/rel && { grep '^#' deploy/RELEASE; cat /tmp/rel; } > deploy/RELEASE.new && mv deploy/RELEASE.new deploy/RELEASE   # MANDATORY re-arm
sudo systemctl restart nofx
journalctl -u nofx --since '2 min ago' | grep 'BOOT INTEGRITY'    # must read expected == rev
cd web && npm run build && cd ..
# restart the dev UI so the loopback bind takes effect, then hard-reload the browser
```

**Corrected demo cleanup** (replaces the wrong one in `2026-08-16-demo-plan-seed.md`; back up first):

```bash
cp ~/nofx/data/data.db ~/nofx-backups/pre-demo-cleanup-$(date +%F).db
cd /home/hoang/nofx && sqlite3 data/data.db "
DELETE FROM plans            WHERE trigger_reason = 'demo_seed';
DELETE FROM plan_overlays    WHERE plan_id = '2026-08-16:NY';
DELETE FROM plan_qa          WHERE plan_id = '2026-08-16:NY';
DELETE FROM day_plan_alerts  WHERE event_id LIKE 'demo:%';
DELETE FROM owner_levels     WHERE label = 'DEMO D-zone';
DELETE FROM trader_positions WHERE source = 'demo_seed';
" && sqlite3 "file:data/data.db?mode=ro" "
SELECT 'real plans kept: '||COUNT(*) FROM plans WHERE trigger_reason <> 'demo_seed';
SELECT 'demo rows left:  '||(
  (SELECT COUNT(*) FROM plans WHERE trigger_reason='demo_seed')
 +(SELECT COUNT(*) FROM day_plan_alerts WHERE event_id LIKE 'demo:%')
 +(SELECT COUNT(*) FROM trader_positions WHERE source='demo_seed'));"
```

**Before Monday, the two config decisions that are yours:** switch ASIA+LONDON off until item 1 lands, and decide whether R:R 1.0 / confidence 50 is what you want the bot trading at (item 4).
