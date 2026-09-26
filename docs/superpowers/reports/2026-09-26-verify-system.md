# VERIFY-0926 — SYSTEM PIPELINE CROSS-CHECK (DS-102)

- **Lane:** DS-102 · **Branch:** `verify/0926-system` · claim `052509c5`
- **Base:** `04ae1c2f` — `git log -1 --format='%H %s'`: `04ae1c2f RELEASE 6cac1b896bdb364642227853d27814a63a55be39 — boot 3 …`
- **Subject:** DS-107's `docs/superpowers/reports/2026-09-26-audit-system-pipeline.md` (branch `audit/0926-system-pipeline` @ `fdc52252`, read via `git show origin/<branch>:<path>`).
- **Mode:** READ-ONLY (L17). Zero code changes. DB = copy `v.db` (2,434,543,616 bytes, online `.backup` of `data/data.db`, `-readonly`). Live `.env` read for key NAMES only. No tests run; `go build ./...` and `go vet ./...` not run (L17 verification, no code changed).
- **Tiers:** [A] = read the exact line at 04ae1c2f / ran the exact command · [B] inferred · [C] speculation.

## TOTALS

| Verdict | Count |
|---|---|
| CONFIRMED | 6/6 P1 · 17/18 P2 · 8/8 §4 refuted-stand |
| PARTIAL | 1 (P2-15) |
| REFUTED | 0 |
| NEW findings | 3 (N-1, N-2, N-3) |

DS-107's report is **solid**: every P1 and P2 checked out at the cited lines, the §4 refuted/withdrawn section stands, and the live census numbers reproduced exactly.

---

## 1. P1 CONFIRM/REFUTE — every one, my own read

| Finding | Verdict | My evidence @04ae1c2f |
|---|---|---|
| **P1-A** far-side build proof survives reconnect | **CONFIRMED [A]** | All three stores guard only `p.BuildID != ""` (`tcp_server.go:1999-2001, 2108-2110, 2132-2134`); `closeConn` (`:2636-2646`) nils `s.conn` only — `farSideBuild` (`:92`) is never cleared. `FarSideProven("")=false` (`tcp_framing.go:353-355`) does not help: a build-id-less hello/heartbeat from a downgraded AddOn leaves the STALE non-empty id proving. `PlaceStopEntry` gate reads it (`tcp_trader.go:772`). Grep of every `farSideBuild` reference (7 sites) — no clear path exists. |
| **P1-B** busy_timeout reaches 1 of 4 pool conns | **CONFIRMED [A]** | `gorm.go:46-48` `SetMaxOpenConns(4)`; `:60` single `db.Exec("PRAGMA busy_timeout = 5000")` on one conn. The repo's OWN comment admits it: `immediate_tx.go:14` *"busy_timeout with ONE db.Exec on a pool of 4 — that pragma reaches exactly …"* (line continues past the cut). Exit receipts carry dedicated-conn busy budgets (`nt8_exit_receipt.go:67,183` + `withImmediateWriteTxBusy` `immediate_tx.go:57,74`) — mitigation ONLY there. `close_sync` retry lives at `trader/ninjatrader/close_sync.go:240,319` (DS-107 cited the package-relative path). |
| **P1-C** v7 legacy-layout mismatch | **CONFIRMED [A]** | `cutover.sh:188-196` prints the v7 runbook verbatim incl. *"the worker (3b-B) … is not built yet"*. `activation.go:89-97` `Resolve` builds release-dir layout (`ReleaseFile = dir/RELEASE`, requires `manifest.json`). Bot reads `deploy/RELEASE` when unset (`releasedir.go:55-65`, `boot_integrity.go:88-100`). Live: `deploy/RELEASE` exists (41 B, 20:11), no release-dir tree → today `Resolve` REFUSES (no manifest); the mismatch is a latent landmine, exactly as DS-107 described. The clobber scenario requires release dirs to exist AND the runbook followed literally; marker divergence (`$RELEASES/$NEW_SHA/RELEASE` vs read `deploy/RELEASE`) is mechanically certain in that case. |
| **P1-D** decision-save failure INFO + ignored | **CONFIRMED [A]** | `auto_trader_decision.go:59` `logger.Infof("⚠️ Failed to save decision record…")` + `return err`. Callers discard at `auto_trader_loop.go:394,493,513,666,754,782,794,801` and `auto_trader_watcher.go:262,371,381,433`; only `:926` checks. DB sink is WARN+ → INFO is journal-only. |
| **P1-E** armed lifecycle writes discarded `_ =` | **CONFIRMED [A]** | `SetState`/`RequestCancel` return errors (`store/armed_orders.go:598,777`); 16 discard lines in `armed_executor.go` (grep count) + `cancel_confirm.go:641`. Reconcile nets exist (idempotent re-request `armed_orders.go:782-784`) → convergence yes, per-event forensics no. |
| **P1-F** loop + drawdown goroutines no recover | **CONFIRMED [A]** | `grep -c 'recover()' trader/auto_trader.go` = 0; `Run` (`:937+`) and `startDrawdownMonitor` (`auto_trader_risk.go:19-39`, only `monitorWg.Done()` defer) run bare. The CALLER `manager.StartAll` (`:100-107`) has no recover either — a panic kills the whole process. |

## 2. P2 spot-verify (17/18 CONFIRMED, 1 PARTIAL)

CONFIRMED [A] at the cited lines: **P2-1** (4 prune funcs exist — `decision.go:361`, `equity.go:149`, `nt8_order_snapshot.go:72`, `level_stats.go:59` — zero non-test callers) · **P2-2** (`logger.go:88-96` one daily file, append, no cap) · **P2-3** (`gorm.go:53-55` "preserved by SetMaxOpenConns(1) above" vs `:46`=4; `log_event.go:26` same class) · **P2-4** (`armed_orders.go:823` ConfirmCancel `tx.First` then write; `strategy.go:2464` Duplicate same shape) · **P2-5** (`nofx-db-backup.sh:17` only `data.db`) · **P2-6** (`server.go:701-708` health returns `time` from a context key set nowhere) · **P2-7** (`engine_analysis.go:604,610` WARN/stamp, no `IncGateBlock` adjacent) · **P2-8** (`IncBreakdownGapNoted` `gate_blocks.go:57` definition-only, zero callers; `.Dropped()` zero non-test readers) · **P2-9** (`experience.go` async `_ = sendTradeEvent`, json/req errors swallowed) · **P2-10** (`auth.go:24-58` in-memory map) · **P2-11** (`handler_user.go:166-180` no rate limit; unknown-email fast 401) · **P2-12** (`bot.go:100-116` first-`/start` binds while `allowedChatID==0`, no confirmation) · **P2-13** (`main.go:150-152` unconditional `🔑 JWT secret configured`) · **P2-14** (`ordered_exec.go:139` append-never-drop) · **P2-16** (`scheduler.go:39,49` `Hour()==21` + `Minute()==0` exact-minute) · **P2-17** (`exchange_account_state.go:58-72` TTL-checked on read, never deleted) · **P2-18** (`transport.go:114-123` `case "", "csv"` silent fallback).

**P2-15 PARTIAL [A]** — `picture_htf_broker.go:372-385`: the goroutine has no explicit stop, BUT it self-heals (`pictureHtfBrokerConsumers.Delete` + return when the subscription dies). "Leak on trader restart" holds only while the server-side subscription lives; the goroutine is not permanent. Severity stands, wording slightly softened.

## 3. §4 refuted/withdrawn — re-check, 8/8 stand

- **Default JWT = takeover** → stands [A]: `config.go:20` insecure default + `:168-173` WARN + `JWTSecretUnfitForUpdates :47`; **live `.env` `JWT_SECRET` = `<set>`** (1 key, `.env:4`) — the live box is not on the default.
- **SSE `?ticket=`** → stands [A]: `stream_ticket.go:20-49` 32B crypto-random, single-use, 30s TTL.
- **CORS `*`** → stands [A]: `SetTrustedProxies(nil)` (`server.go:64`), no cookie auth anywhere (grep `SetCookie|http.Cookie` = 0), Bearer-only.
- **Bar-persist drops** → stands [A]: `bar_persist.go:20-53` counted + 1-line/min WARN + F2 watchdog + self-healing cache tail.
- **Overlay carry failure** → stands [A]: P1 alert naming the count (`auto_trader_planner.go:422`, `auto_trader_levelstate.go:31`).
- **researchsnapshot storm** → stands [A]: separate `.research.db` (`main.go:84`).
- **Boot-line literals** → stands [A]: READ values (map-driven validators).
- **Heartbeat wall-clock** → stands [A]: `tcp_server.go:2459` `time.Since(lastAckTime)`.

## 4. P0 HUNT (CTO's five questions)

### (a) Restart/cutover: lose a close / double-send / two bot processes
- **Lose a close — NO, durable [A].** `SavePendingExit` persists the park (`nt8_exit_receipt.go:185-195`, BEGIN IMMEDIATE dedicated conn); `RetryPendingNT8Exits` (`close_sync.go:390-399`) re-applies parked receipts idempotently on the next cumulative-entry update and from the close path (`:315`); exhaustion logs **ERROR** (`:349`). A restart cannot silently drop a close — worst case is park-until-next-update latency (the parked row survives). This is the #227 exit path working as designed.
- **Double-send — NO [A/B].** Pending signal frames are in-memory only (not persisted) → restart drops them, no re-send. Pre-boot non-terminal ledger rows are swept at boot (class-33 `boot_sweep.go:37-39`) → no place_pending re-placement. Far-side seq + echo-verify bound duplicates on the wire.
- **Two bot processes — transient window, NOTE [A/B].** `Restart=on-failure + RestartSec=5 + StartLimitIntervalSec=0` (`deploy/nofx.service`) relaunches forever on failure. A manually launched second instance: `LoadTradersFromStore` auto-starts traders (`main.go:278`) BEFORE the API bind; bind failure → `logger.Fatalf` (`main.go:684-688`) exits the process — so the window is the seconds between trader auto-start and bind detection, not sustained dual trading. No lock prevents the window; it self-closes via Fatal.

### (b) DB lock/BUSY silent drop beyond the #227 exit path
- **YES — three silent families [A], all downstream of P1-B's pool reality:** (i) 16 armed `SetState`/`RequestCancel` discards with no log (P1-E); (ii) decision-record saves INFO-only (P1-D, DB-sink-invisible); (iii) **NEW sharpening: `_ = db.AutoMigrate(...)` at 4 sites** (`watchdog_fire.go:61`, `planner_rejected.go:48`, `candidate_pool.go:56`, `config_diff.go:40`) — a failed migration leaves the table ABSENT and every subsequent write to it fails silently inside those same `_ =`-style paths. Not a new hazard class, but the AutoMigrate-`_ =` instance is unlisted in DS-107's P2-8/P1-E.

### (c) Does anything log a secret
- **No [A].** Grep of `logger.*f` against APIKey/Secret/Passphrase/PrivateKey/PasswordHash: only assignments, no log calls. Exchange/model update logs mask (`handler_exchange.go:333`, `handler_ai_model.go:258` `MaskSensitiveString`). Boot prints account NAMES only. `.env` parse errors redacted (`main_dotenv.go:52-68`). Telegram bot token plaintext in DB is house posture (DS-107 NOTE).

### (d) Unauthenticated / non-owner mutation of traders/accounts/keys/.env
- **NEW finding N-1 (P2): any AUTHENTICATED user can write the wallet secret into `.env` and the process env.** `POST /api/onboarding/beginner` (`server.go:196`) sits in `protected` (JWT only); `planTraderOwnership` passes with no `trader_id`; `handleBeginnerOnboarding` then runs `os.Setenv("CLAW402_WALLET_KEY"/"CLAW402_WALLET_ADDRESS"/"CLAW402_DEFAULT_MODEL", …)` (`handler_onboarding.go:72-75`) and `persistBeginnerWalletEnv` upserts `.env` / `./.env` / `/app/.env` (`:241-266`). Process-wide, cross-user mutation with no ownership/role gate. Bounded by loopback bind + single-user deployment + need for a valid account; still a non-owner `.env` writer the CTO asked about. DS-107 did NOT flag it.
- Everything else mutating sits inside `protected` (grep of non-protected POST/PUT/DELETE in `setupRoutes` = none); `/metrics` is the only non-protected GET.

### (e) JWT default secret
- **`JWT_SECRET` = `<set>`** in `.env (1 key, line 4).

## 5. (4) nofx-activate v7 legacy-layout mismatch — CONFIRMED
Covered under P1-C above: runbook quoted verbatim, `Resolve` release-dir layout, bot reads `deploy/RELEASE` unset, live box flat with `deploy/RELEASE` (41 B). Latent, not active — `Resolve` refuses today (no `manifest.json` anywhere in a release dir).

## 6. NEW FINDINGS (severity, file:line @04ae1c2f, refute attempt)

- **N-1 (P2)** — onboarding env/.env write by any authenticated user — §4(d). Refute attempt: the route IS JWT-protected and the box is single-user/loopback → downgraded from P1. Stands as P2.
- **N-2 (NOTE)** — transient double-trader window before bind-Fatal on a manually launched second instance — §4(a). Refute: Fatal closes it in seconds; no persistent dual trading.
- **N-3 (NOTE)** — `_ = db.AutoMigrate` at 4 stores: a failed table migration makes subsequent writes fail silently inside already-silent paths — §4(b). Refute: migrations are lazy-create and these tables are non-critical (watchdog fires, planner rejections, candidate pool, config diffs) → NOTE.

## 7. WHAT I DID NOT DO
- No code changes, no tests, no `go build` (L17 read-only verification).
- Did not re-verify DS-107's §5 open items (pool introspection, bind-exposure from host, telegram wild, researchsnapshot growth).
- Did not review the trading-pipeline or knobs reports beyond the cross-references the CTO asked for.
- Did not check the partner repo (outside this dispatch's (1)-(4) scope).
