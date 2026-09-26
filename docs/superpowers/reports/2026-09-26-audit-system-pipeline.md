# AUDIT-0926 — SYSTEM PIPELINE (DS-107) @ 04ae1c2f

- **Lane:** DS-107 · **Branch:** `audit/0926-system-pipeline` · claim `68911460`
- **Base:** `04ae1c2f` = RELEASE `6cac1b89` — boot 3 (#227), live since 19:33:22 CT 2026-09-25, restarted 20:14:24 CT with `FAST_MARKET_REASONING=max`.
- **`git log -1` of the audited tree:** `04ae1c2f RELEASE 6cac1b896bdb364642227853d27814a63a55be39 — boot 3 …`
- **Mode:** READ-ONLY audit (L17). Zero code changes, zero checklist edits, zero deploys. `go build ./...` + `go vet ./...` green at the claim head; no go tests run (box loaded: DS-101 API job, DS-104 suite; L16).
- **Evidence:** [A] = read the exact line at 04ae1c2f or ran the exact command; [B] inferred; [C] speculation.
- **Live reads:** DB COPY `/tmp/ds107-audit.db` (sqlite3 `-readonly` + online `.backup` of `data/data.db`), `data/nofx_2026-09-25.log`, `GET /api/health` + `GET /api/cutover-gate` (401). `.env`: key NAMES only; no values read or printed.

---

## 1. MAP — system pipeline at 04ae1c2f

### 1.1 Boot (`main.go` + `config/` + `kernel/`)
1. `main.go:40` logger init (stdout + `data/nofx_DATE.log`); `main.go:45` `.env` load (fails open, parse errors redacted `main_dotenv.go:52-68`).
2. `main.go:52` `config.Init()` (`config/config.go:146`); JWT load+default-WARN `:165-173`; `API_SERVER_HOST` off-loopback WARN `:186-190`.
3. `main.go:58` `crypto.NewCryptoService()` (requires `RSA_PRIVATE_KEY`+`DATA_ENCRYPTION_KEY`, boot-FATAL on failure `:59`).
4. `main.go:84` `researchsnapshot.Start(dbPath+".research.db")` — separate DB.
5. `main.go:91` `store.NewWithConfig` → `store/gorm.go:23` InitGorm: pool 4/4/30m (`:46-48`), PRAGMA `foreign_keys` `:51`, WAL `:58`, `synchronous=FULL` `:59`, `busy_timeout=5000` `:60` (one conn — see finding P1-B).
6. `store/store.go:137-249` initTables (23 sub-stores) + `:252` initDefaultData; one-time corrections `main.go:127-145`; `logger.AttachDBSink` `main.go:110-123` (WARN+ → `log_events`).
7. `main.go:260-277` F12 `SetOrderSnapshotSink` registered BEFORE trader load.
8. `main.go:278` `LoadTradersFromStore` (`manager/trader_manager.go:477-583`, fail-closed `SettingsTruthRefusal` `:605-610`) → `NewAutoTrader` (`:739`) → `transport.go:51` `getOrStartTCPServer` (TCP listener opens) → `:753-767` `go trader.Run()` if `IsRunning`.
9. `main.go:316` `kernel.AssertBootIntegrity()` (`boot_integrity.go:132`; revision vs `deploy/RELEASE` `:84-110`; prompt goldens `golden_selfcheck.go:112-149`); refusal → `tradingRefused` + ERROR.
10. `main.go:326-666` boot lines (READ values; deliberate literal n/a args documented); `:668` API server; `:689` telegram; `:693-701` signal wait → Shutdown → StopAll.

### 1.2 NT8 bridge (`provider/ninjatrader/`)
- Lifecycle: `Start :1184`, `Stop :1256`, `acceptLoop :1525` (single concurrent client `:1551-1557`), `readLoop :1894`, `closeConn :2638`.
- Handshake/build proof: `FrameHello :1978-2014`, `recordHello :2003`, `farSideBuild` stores `:1998-2001, :2107-2110, :2131-2134`, `FarSideProven` `tcp_framing.go:353`, `ExpectedAddonBuild="2026-09-23-m21"` `order_snapshot.go:214`, `MinAddonBuildStopSlot` `:322`; consumed at `PlaceStopEntry` `tcp_trader.go:772`.
- Framing: `WriteFrame :739` / `ReadFrame :761` (4-byte BE, 1MB cap `:49`); writes serialize `writeMu` + 5s deadlines.
- Reconnect: the C# AddOn dials; server accepts forever; Go retry = `sendAutoBarsSubscribe :1717`; mid-flush write failure re-queues tail `:2590-2605`.
- Feed: `IsFeedConnected :1427` (`feedFreshWindow=90s :1418`), consumers `ninjaFeedDown` `auto_trader.go:59-76` + dead-man watchdog `:85-100`.
- Snapshots: `OrderSnapshotCache` (`order_snapshot.go:125-197`, `LatestReceived :175`, `AgeAt :186`), positions `:2363-2380`, account `:2277`.
- BarCache: `SeedHistorical :255`, `Upsert :363`, cap 2500 `:25`; ingest queue 4096 drop-oldest `:1830-1858`; persist worker 4096 + 300ms flush + 30s watchdog (`bar_persist.go:26,79-90`); no purge on disconnect (unsubscribe/roll only).
- Heartbeat: 30s interval `:37`, 60s ack timeout `:41` → `closeConn`; fan-out: `ListenOrderUpdates :458`, multi-consumer `runOrderUpdateFanout :494`; advisory drops WARN `:334-342`.
- Ordered execution: `ordered_exec.go:87,127,150,221` (never drops, unbounded queue); echo verify: `echo_verify.go:46-192`, `pendingOpsCap=4096` `tcp_server.go:272`.

### 1.3 Persistence (`store/` + backups)
- Pool/PRAGMAs `store/gorm.go:44-60` (above); GORM default tx per write (`:25-31`).
- BEGIN IMMEDIATE only in `store/immediate_tx.go:74-109`, used by `store/nt8_exit_receipt.go:69,189` (ApplyNT8Exit / SavePendingExit, dedicated conns). All other transactions deferred (`store/store.go:721`, `settings_truth.go:501`, `strategy.go:2441`, `armed_orders.go:823`, `order_settle.go:13` …).
- Lazy AutoMigrate with discarded errors: `watchdog_fire.go:61`, `planner_rejected.go:48`, `candidate_pool.go:56`, `config_diff.go:40`, `touch_outcomes.go:205`, `accepted_risk.go:73`, `planner_read_facts.go:122`.
- Backups: `deploy/nofx-db-backup.sh:17-77` (online `sqlite3.backup`, quick_check, KEEP_DAILY=14 / WEEKLY=8, timer 05:00+17:30 CT). `research.db` NOT backed up.
- Logs: `logger/logger.go:90-95` one `data/nofx_DATE.log` per process, no rotation/retention. DB sink WARN+ drop-on-overload (`db_sink.go:51-60`, `log_event.go:85-99`); retention 30d `:75-84`; **no API route reads log_events**.
- Retention dead code: `decision.go:361 CleanOldRecords`, `equity.go:149`, `nt8_order_snapshot.go:72 PruneBefore`, `level_stats.go:59` — zero production callers.
- Busy retry only for NT8 exit receipts (`close_sync.go:322-329`, `ntExitBusyRetries=4` → park).

### 1.4 Deploy/ops (`deploy/`, `internal/activation/`, `internal/updater*`)
- `deploy/cutover.sh` v7 = thin wrapper printing a manual runbook `:185-196` (no unattended activation); v6 (git history) did install→dist→RELEASE→kill-9→prove relaunch (`:123-128, :186-202`).
- `nofx-activate` (`cmd/nofx-activate/main.go`) + `internal/activation/activation.go:89-234` (`Resolve` builds release-dir layout; `ReleaseFile = dir/RELEASE :97`), `steps.go:97-176, :381-467` (Activate/rollback), `system.go:24-48` (SIGKILL because SIGTERM-exit-0 does not relaunch under `Restart=on-failure`).
- RELEASE: bot reads `internal/installpath/releasedir.go:55-65` (`deploy/RELEASE` when unset) vs `kernel/boot_integrity.go:84-96`; mismatch → `tradingRefused` + read-only.
- `deploy/nofx-lock.sh` 525 lines: acquire `:94` (atomic mkdir + keeper), check rc 0/1/2/3/4 `:383-402`, reclaim rc 3 = inherited `:459-487`, expiry enforced at heartbeat `:283-290`.
- systemd: `deploy/nofx.service:35,42` `StartLimitIntervalSec=0`, `Restart=on-failure`, `RestartSec=5`; user timers for backup/clock-guard.
- Maintenance hold: `store/maintenance_hold.go:62-283` (flock, stat-keyed cache, corrupt ⇒ HELD fail-closed); `trader/maintenance_gate.go:30-75`; AddOn push `provider/ninjatrader/maintenance_wire.go:142-204`.
- Updater auth M3/M4: `internal/updateauth/*` (0600 files, O_NOFOLLOW, HMAC 5-min window `mac.go:34`, single-use Consume with replay/prune watermark `seen.go:230-290`); `api/handler_updates.go:215-664` (18-step `updatesRefusal :376-494`, knob `NOFX_UPDATER=1 :102-137`).
- Cutover-gate: `GET /api/cutover-gate` `api/handler_order.go:398-412` → `trader/class33_cutover_gate.go:60-140` (5 legs, unevaluable = fail). Installation-gate: `trader/installation_gate.go:95-428` (12 legs). Both JWT-protected (`api/server.go:467,476`).

### 1.5 Observability (`telemetry/`, `logger/`, `api/`)
- Gate-blocks: `GET /api/risk/gate-blocks` `api/server.go:512` → `api/handler_gate_blocks.go:30`; `telemetry.IncGateBlock` `gate_blocks.go:38`; ~45 gate names; admission chain refusals all counted (`entry_admission.go:136-140`).
- Health: `GET /api/health` `api/server.go:120,701-708` — liveness only, `time` reads a context key set nowhere → always `null`.
- DB sink WARN+ (`db_sink.go:51`); journald no level filter (`deploy/journald-nofx.conf`, 2G cap).
- INFO-only failure paths (journal/file only, DB-sink-invisible): equity save `auto_trader_decision.go:34`, decision save `:59`, guardrail-skip `auto_trader_loop.go:732`, data-source failures `kernel/engine.go:574-618`, binance sync family `trader/binance/order_sync.go:*`.
- Write-only counters (no production readers): `FarArmCounts`, `WeeklyCounterSnapshot`, `ResearchSnapshotDrops/Rows`, `RepairRegressionCount`, `ShadowedArmRefusalCount`, `LogEventStore.Dropped()` (`log_event.go:96`), `IncBreakdownGapNoted` (zero callers), `FillLatency`, `DatabentoErrorsTotal`.

### 1.6 Security (`auth/`, `api/`, `telegram/`, `config/`)
- JWT: default public secret `config/config.go:20` + WARN `:170`; `JWTSecretUnfitForUpdates :47-60` (refused by `/api/updates*` only); 24h expiry `auth/auth.go:146`; in-memory blacklist `:24-58`; strict parser `:173`.
- Auth: `api/server.go:180` `protected := api.Group("/", authMiddleware, planTraderOwnership)` — every mutating route inside; `authMiddleware :909-988` (blacklist → validate → retirement → machine-denied → context).
- Telegram: `bot.go:46-80, 100-150` (chatID binding + post-bind refusal), `/start` first-come-first-bind no confirmation `:100-116`; machine bot token `agent.go:77-79`; token masked in API `handler_telegram.go:23-31`; plaintext in DB `store/telegram_config.go:19`.
- CORS `*` (`api/server.go:94-107`), Bearer-only (no cookies) — safe; `SetTrustedProxies(nil) :64`.
- Secrets in logs: masked via `MaskSensitiveString` (`api/utils.go:5-13`) at exchange/model update logs (`handler_exchange.go:322-342`, `handler_ai_model.go:258`); no key-value logged at boot (area-1 finding 3).

### 1.7 Resource/timing (`trader/`, `provider/`)
- Per-trader loops: main Run ticker (`auto_trader.go:937,1072`), drawdown monitor (`auto_trader_risk.go:19-39`), armed event pass (`armed_event_pass.go:77-89`), picture-HTF consumer (`picture_htf_broker.go:372-385`), close sync, reconcile 20s (`reconcile.go:94`), level-stats nightly 17:05 CT (`level_stats_wire.go:57-65`), bar prune 24h ticker (`bar_persist_wire.go:356`).
- Server singleton: acceptLoop, drainBarIngest, livenessReporter (60s), maintenanceLoop (`tcp_server.go:1195-1204`).
- Wall-clock production gates: EOD flat 14:45 CT, lunch window, calendar slices, half-days — single-sourced in `session_registry.go:112`; daily report gated on exact 21:00 hour-minute (`agent/scheduler.go:39,49`).
- Clock: clock-guard unit + `StampAligned`; monotonic stamps in heartbeat/maintenance (`maintenance_wire.go:24,33,162`).
- Bounded: BarCache 2500, replayHold 2×2500, pendingOps 4096, barPersist 4096, liveSink 2048, fill chan 32. Unbounded: ordered-exec queue (`ordered_exec.go:139`), exchange account-state cache (no sweep).

---

## 2. CENSUS (live, from the DB copy + log; rows name their source)

| What | Value | Evidence |
|---|---|---|
| DB file | `data/data.db` 2,434,543,616 bytes (~2.27 GiB) | `ls -la` on copy [A] |
| `decision_records` | 46,698 rows · **905 MB** | `dbstat` on copy [A] |
| `bars` | 174 MB | `dbstat` [A] |
| `plans` | 47 MB | `dbstat` [A] |
| `log_events` | 100,513 rows · 27.7 MB | COUNT + `dbstat` [A] |
| `nt8_order_snapshots` | 6.8 MB | `dbstat` [A] |
| `trader_equity_snapshots` | 6.7 MB | `dbstat` [A] |
| `traders` | 1 row | COUNT [A] |
| `armed_orders` states | 111 cancelled · 35 filled · 14 superseded · 14 non-terminal | GROUP BY [A] |
| Log file `nofx_2026-09-25.log` | 7,982,222 bytes · 28,648 INFO · 4,002 WARN · 0 ERROR | grep counts [A] |
| `/api/health` live | `{"revision":"6cac1b896bdb","status":"ok","time":null}` | curl GET [A] |
| `/api/cutover-gate` live | 401 without token | curl GET [A] |
| `.env` (names only) | 30 keys present; secret-bearing names all `<set>` | key-NAME extraction [A] |

---

## 3. FINDINGS by severity

### P0 — none held.
Candidates examined and refuted: default JWT secret (loopback bind + WARN + updates-refusal chain, area 6); Telegram bind (bounded by secret token + post-bind refusal); bar-persist drops (loud + counted + watchdog); overlay carry failure (P1 alert + fail-closed). None can lose money, trade wrong, or trade when it must not.

### P1 (wrong behaviour / structural hazard, no direct loss proven today)

**P1-A — Far-side AddOn build proof survives a reconnect; a downgraded AddOn keeps the old proof.**
`provider/ninjatrader/tcp_server.go:1998-2001, 2107-2110, 2131-2134` store `farSideBuild` only on non-empty build ids; nothing clears it on `closeConn` (`:2638-2646`) or when a new connection's hello/heartbeat carries no build id. `FarSideProven` (`tcp_framing.go:353`) then answers for a build that is no longer connected, and `PlaceStopEntry`'s gate (`trader/ninjatrader/tcp_trader.go:772`) passes → a reverted/pre-m21 AddOn mis-executes `stop_entry` as MARKET (the 2026-08-30 class incident). *Refute attempt:* read every `farSideBuild` reference — only the three non-empty-guarded stores exist; no clear path. **Survives.** Checklist nearest: class 96 (latch with no automatic release). [A]

**P1-B — `busy_timeout=5000` reaches only 1 of 4 pool connections; every other write fails SQLITE_BUSY immediately.**
`store/gorm.go:46` sets pool 4; `:60` sets busy_timeout via a single `db.Exec` (one connection). GORM's implicit per-write tx upgrades to a write lock at its first write statement; on the other 3 conns the busy handler is 0ms. The repo's own checklist already records the probe ("prove which pooled connections actually carry it", `AUDIT-CHECKLIST.md:7685`) and the 2026-09-25 08:55 CT incident (row 618) is the live instance. Mitigation exists ONLY on the exit-receipt path (`store/nt8_exit_receipt.go:69,189` dedicated conns + `close_sync.go:322-329` retry). *Refute attempt:* WAL windows are short — refuted by the production incident itself; "writes stay on one conn" — refuted (database/sql hands any free conn). **Survives.** Class 105 also at `store/gorm.go:53-55` ("preserved by SetMaxOpenConns(1) above" — the line above is `SetMaxOpenConns(4)`). [A]

**P1-C — Legacy-layout mismatch: the v7 runbook's activation paths cannot install the flat layout.**
`deploy/cutover.sh:185-196` prints `nofx-activate activate -release $RELEASES/$NEW_SHA -prev $RELEASES/$OLD_SHA`; `internal/activation/activation.go:97` builds `ReleaseFile = dir/RELEASE` (release-dir layout); the bot reads `internal/installpath/releasedir.go:55-65` (`deploy/RELEASE` when unset). Followed literally on this flat box: if release dirs exist, `Activate` clobbers the ONLY rollback copy with the new files while the flat binary stays old (null cutover, class 74 shape); even with `prev=$INSTALL`, the marker lands at `$INSTALL/RELEASE` while the bot reads `$INSTALL/deploy/RELEASE` → next boot new-binary-vs-old-marker → `boot_integrity.go:135-147` refuses trading. *Refute attempt:* with no release dirs `Resolve` refuses (no manifest) — harmless today; the script itself says the worker (3b-B) is "not built yet". **Survives as a latent landmine** the CTO asked to confirm and describe. [A]

**P1-D — Decision-record save failure: INFO + error ignored at 12 of 13 call sites.**
`trader/auto_trader_decision.go:59` logs `logger.Infof("⚠️ Failed to save decision record")` and returns; callers discard the error at `trader/auto_trader_loop.go:394,493,513,666,754,782,794,801` and `trader/auto_trader_watcher.go:262,371,381,433` (only `:926` warns; grid path warns `auto_trader_grid.go:614-615`). Under P1-B, a busy DB loses decision rows with only an INFO line — journald-prunable at the 2G cap and invisible to the WARN+ DB sink. `decision_records` is the audit/forensics substrate (905MB live). *Refute:* not fully silent (INFO exists); no trading impact. **Survives as P1 for observability, the slice this audit owns.** [A]

**P1-E — Armed-order lifecycle writes discarded (`_ =`) with no log.**
`store/armed_orders.go:598` (`SetState`) and `:777` (`RequestCancel`) errors discarded at ~20 sites in `trader/armed_executor.go` (`:488,491,521,582,705,776,822,1232,1430,1765,2128,2254-2256,2267,2273,2678,2705,2780,2975`) + `trader/cancel_confirm.go:641`. A failed persist leaves the row in its old lifecycle state; withdraw views and settlement can miss it. *Refute:* reconcile nets exist (order_snapshot reconcile, settlement pass, idempotent RequestCancel re-request `store/armed_orders.go:782-784`) → eventual convergence, not per-event forensics. **Survives as P1/P2 borderline; held at P1 within a system-pipeline audit.** [A]

**P1-F — Main loop and drawdown monitor goroutines have no panic recovery.**
`trader/auto_trader.go:937` (`Run`) and `trader/auto_trader_risk.go:19-39` (`startDrawdownMonitor`) run bare `go func` with no `recover()`; a panic in `tickOnce`/`monitorTick` kills the WHOLE process (all traders), not just the trader. `Restart=on-failure` relaunches, but one trader's panic takes every trader down. *Refute:* local recover guards exist inside some hooks, but none at the loop frame (grep `defer func` in `auto_trader.go` → none). **Survives.** [A]

### P2

- **P2-1 Unbounded tables, no retention.** `decision_records` (905MB live [A]), `trader_equity_snapshots`, `nt8_order_snapshots`, `level_stats`: prune functions exist with zero production callers (`store/decision.go:361`, `store/equity.go:149`, `store/nt8_order_snapshot.go:72`, `store/level_stats.go:59`). [A]
- **P2-2 No log rotation/retention.** `logger/logger.go:90-95` one file per process lifetime, no cap, no prune (7.98MB/day today). [A]
- **P2-3 CLASS 105 comment/code drift ×2.** `store/gorm.go:53-55` ("preserved by SetMaxOpenConns(1) above" vs `:46` = 4); `store/log_event.go:26` ("SQLite pool is capped at 1 conn"). [A]
- **P2-4 Read→write upgrade inside deferred txs.** `store/armed_orders.go:823` (ConfirmCancel) and `store/strategy.go:2464` (Duplicate): `tx.First` then write → SQLITE_BUSY on upgrade under WAL; callers bounded/retryable (API, cancel sweep). [A]
- **P2-5 `research.db` not backed up.** `deploy/nofx-db-backup.sh:17` only `data.db`. [A]
- **P2-6 `/api/health` liveness-only, `time` always null.** `api/server.go:701-708` reads a context key set nowhere; LIVE output `{"status":"ok","time":null}` [A]. Dead-DB/dead-NT8 bot answers 200.
- **P2-7 Guardrail HOLDs bypass the gate-block table.** `executor_plan_gate` (`kernel/engine_analysis.go:604`) and `schema_parse_failed` (`:610`) WARN/stamp but never `IncGateBlock` → `/api/risk/gate-blocks` under-reports cycle-holding gates. [A]
- **P2-8 Write-only counters.** `telemetry/far_arms.go:33`, `weekly.go:63`, `research_snapshot.go:24-27`, `planner_wave.go:17`, `shadow_conditions.go:13`, `gate_blocks.go:57` (zero callers), `metrics.go:41,52`; `store/log_event.go:96` `.Dropped()` zero readers — WARN+ shipping can silently drop events. [A]
- **P2-9 GA4 telemetry swallows all failures.** `telemetry/experience.go:112,160,187-189,231-233` — non-2xx counts success; can be 100% dead with zero indication. [A]
- **P2-10 Logout blacklist in-memory only.** `auth/auth.go:24-58` — a restart revives logged-out tokens for the rest of 24h. [A]
- **P2-11 No rate limit on `/login`.** `api/handler_user.go:166-211` (unknown email fast-401, known email bcrypt); only password-change has the 1s delay. Bounded by loopback default bind + off-loopback WARN. [A]
- **P2-12 Telegram first-`/start` bind has no confirmation gate.** `telegram/bot.go:100-116` + `store/telegram_config.go:104-116` — while unbound, the first chat that reaches the bot becomes the bound account. Bounded: token secret, post-bind all other chats refused. H1 fixed route scope, not bind races. [A]
- **P2-13 `🔑 JWT secret configured` fires unconditionally.** `main.go:150-152` even under the insecure default (real WARN is `config/config.go:170`). Class 45/49 shape; already flagged in `AGENTS.md`. [A]
- **P2-14 Ordered-exec queue unbounded.** `provider/ninjatrader/ordered_exec.go:139` append-never-drop, no cap; worker handlers can block on DB. [A]
- **P2-15 Picture-HTF consumer goroutine has no stop path.** `trader/picture_htf_broker.go:372-385` leaks on trader restart while the server subscription lives. [A]
- **P2-16 Daily report / hourly cleanup gated on exact wall-clock minute.** `agent/scheduler.go:39,49` (`Hour()==21`, `Minute()==0`) — a missed minute skips the report for the day; no boundary-aligned ticker. [A]
- **P2-17 Exchange account-state cache entries never swept.** `api/exchange_account_state.go:58-72` TTL-checked on read, never deleted. [A]
- **P2-18 `NT_TRANSPORT` unset = silent legacy-CSV fallback.** `transport.go:114-123` `case "", "csv"`; live `.env` has it `<set>`. Config-dependent silent fallback. [A]

### NOTE

- Boot integrity assert runs AFTER traders auto-start (`main.go:278` vs `:316`); admission chain reads `TradingRefused()` (`entry_admission.go:224`), assert latches in ms — ordering gap only. [A]
- One-time DB corrections run before the binary's identity is verified (`main.go:127-145` vs `:316`) — idempotent/WHERE-scoped + backup-first guards. [A]
- Goldens cover only the futures decision prompt (`golden_selfcheck.go:119-127`). [A]
- Boot logs print account NAMES (never keys) (`trader_manager.go:461,750`). [A]
- `.env` parse-error leak closed (`main_dotenv.go:52-68`). [A]
- Advisory channels drop with WARN-only; ordered worker + snapshots + reconcile recover durable state. [A]
- Pending-signal queue unbounded while disconnected (`tcp_server.go:1277-1283`); production entries blocked by dead-man watchdog → growth requires watchdog bypass. [A]
- `livenessReporter` hardcodes MNQ (`tcp_server.go:1235-1238`). [A]
- `ExpectedAddonBuild` hardcoded (`order_snapshot.go:214`) — per-AddOn-deploy bump, class 105 family. [A]
- Mid-flush re-queue keeps the `attempted` entry in the tail (`tcp_server.go:2576-2605`) — duplicate window bounded by `TCPStaleSignalAge` 60s and B3 at place-time (not flush-time). [A]
- Level-stats per-trader goroutine has no stop path (idles 24h). `armed_event_pass.go:67-72` leaked 5s timer per Stop. `auto_trader.go:1081` orphaned monitor on grid-init failure. [A]
- Bot token stored plaintext in DB (`store/telegram_config.go:19`) — house posture, same as exchange/model creds. [A]
- `no-sudo` backup/clock-guard timers + journald 2G cap exist but are owner-installed steps (not on the boot path). [A]

---

## 4. REFUTED / WITHDRAWN

- **Default JWT secret = takeover** → refuted: loopback default bind + off-loopback WARN + `/api/updates*` refusal chain; residual only if operator exposes off-loopback AND leaves `JWT_SECRET` unset. [A]
- **SSE `?ticket=` leaks session** → refuted: opaque 32B random, single-use, 30s TTL, minted under JWT (`api/stream_ticket.go:20-49`, `api/server.go:195`). [A]
- **CORS `*`** → refuted: Bearer-header auth, no cookies; `SetTrustedProxies(nil)` prevents XFF spoof. [A]
- **Bar-persist drops lose bars silently** → refuted: WARN rate-limited 1/min + ERROR on close-drop + watchdog + `ClosedCacheTail` self-heal. [A]
- **Overlay carry failure silent** → refuted: total-failure path is ERROR + P1 alert; only the single-bad-overlay case is unlogged. [A]
- **`researchsnapshot` write storms the bot DB** → refuted: separate `.research.db` file, own 1-conn pool. [A]
- **Boot-line literals that could drift** → refuted: count/state lines are READ from the same maps the validators read (`kernel/plan_doc.go:429-445`); deliberate literals are documented n/a args (A11/A24). [A]
- **Heartbeat timeout wall-clock misuse** → refuted: monotonic `time.Since` + ack stamps. [A]

---

## 5. WHAT I COULD NOT VERIFY

- **Which pooled connections actually carry `busy_timeout` at runtime** — needs `PRAGMA busy_timeout` per-connection introspection from inside the process; checklist class's own probe (`AUDIT-CHECKLIST.md:7685`). The structural argument (single `db.Exec` on a 4-conn pool) is [A]; the runtime distribution is [B].
- **Whether the v7 activation runbook has ever been exercised against this flat layout** — no release-dir tree exists on the box today; the mismatch is latent (the script admits the worker path "is not built yet").
- **AddOn build downgrade while connected** — hazard requires an NT8-side revert with a build-id-less hello; cannot be observed from the Go side alone.
- **`/login` brute-force exposure** — depends on `API_SERVER_HOST` remaining loopback; `.env` name is set but its value was not read (L12).
- **Telegram bind race in the wild** — single-user deployment, token secret; no way to prove no third party ever reached the bot pre-bind.
- **Whether `researchsnapshot` archive tables grow unbounded** — `rows=16047` today (boot line [A]) but no prune path found; growth rate over weeks unverified.

---

## 6. CROSS-SLICE NOTES (one line each)

- DS-102 (settings/knobs): `busy_timeout` pool finding (P1-B) and `NT_TRANSPORT` silent fallback (P2-18) are config-adjacent — their slice.
- DS-106 (trading pipeline): decision-record INFO loss (P1-D) and armed-ledger `_ =` discards (P1-E) touch their record substrate.
- CTO: legacy-layout mismatch (P1-C) is confirmed-and-described as asked; remediation belongs to a deploy wave, not this audit.
- CTO: the 7 unchecked-telemetry counters (P2-8) would be a natural fold into the parked knob-prune list.

---

*DS-107 · audit read-only · zero code changes · findings terse here by design; all file:line facts at 04ae1c2f.*
