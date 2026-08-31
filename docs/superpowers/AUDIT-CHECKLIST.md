# AUDIT-CHECKLIST — the permanent audit playbook

Codified 2026-08-28 from the campaign's 18 proven bug classes. **Every audit
dispatch MUST reference this file instead of re-deriving the probe list. Every
NEW bug class found gets appended here in the SAME PR that fixes it** (canon law
in CLAUDE.md).

---

## PART 1 — THE 25 BUG CLASSES (name · root cause · probe · law)

1. **Self-imposed caps.** Root cause: an AI/HTTP/token cap chosen without
   measuring the provider ceiling or the observed need (the 32768-token
   truncation ceiling; provider ceiling 393216 probed 2026-08-19).
   **Probe:** `grep finish_reason=length` in the journal + compare every cap
   against the probed provider ceiling. **Law:** never self-cap below observed
   need; a cap change ships with the measured ceiling in the commit.

2. **Schema-gate vs validator.** Root cause: the per-attempt validator
   rejections in the 3-attempt loop logged NOTHING (silent `continue`) — the
   gate existed but forensics were blind (P0 side-quota lesson).
   **Probe:** a fixture AT THE GATE BOUNDARY — the validator and its golden test
   share one pure function ("the test IS the write path", T2 stamping).
   **Law:** every gate ships a gate-level fixture in the same PR.

3. **Connection starvation.** Root cause: NT8 TCP link gaps (01:25, 15:22-15:26
   CT) + the 60s ack timeout that closes the conn on a stalled heartbeat.
   **Probe:** the watchdog — `wire_liveness` lines (last_frame_age,
   frames_per_min) + dead-man reconnect verified from the far side.
   **Law:** a stalled heartbeat is a bug until the reconnect cycle is proven
   (e2e dead-man check).

4. **sync.Once / LoadOrStore wrong-owner.** Root cause: `nt.OrderUpdates()`
   evaluated as the LoadOrStore ARGUMENT every cycle → the map held a closed
   channel → 310,808 zero-value drains in 15s.
   **Probe:** grep LoadOrStore call sites — the key argument must be a VALUE,
   never a function call; channels subscribe on miss only + self-heal on close.
   **Law:** never call a function inside LoadOrStore.

5. **Deaf consumer / far-side frames.** Root cause: the C# AddOn filtered
   order_update to Filled/Rejected only, or ran a PRE-shipped binary (md5
   mismatch) — the Go consumer listened to frames the far side never emitted.
   **Probe:** byte-compare the deployed AddOn (md5 vs repo) + capture a RAW far
   -side frame before blaming the consumer. **Law:** prove the far side emits
   before you debug the listener.

6. **Go-side theater.** Root cause: a confident Go log line ("⚔️ armed S1 …")
   for a placement path that was not actually shipped (Phase 2 not live) — the
   first E2 was GO-OPTIMISTIC. **Probe:** demand a FAR-SIDE frame quote
   (dispatcher log / NT8 order state) for every wire claim.
   **Law:** a Go log line is not a wire frame; theater dies in the frame quote.

7. **Timestamp convention.** Root cause: close-stamped frames persisted raw →
   every replayed row landed at T+1m (the 2499/2500 mismatch class).
   **Probe:** `open_time_ms % 60000 == 0` on the whole table + fill-containment
   (own fills must sit inside `[L,H]` of the floor-minute bar; 0 may fit the
   T+1m bar). **Law:** ONE canonical open-stamp conversion shared by every
   reader; the residue check alone cannot catch T+1m (both divide 60000).

8. **Clamp-vs-knob.** Root cause: FE persists values above the engine ceiling
   (leverage 20 vs system 10) — saved but inert; proximity clamp 0.1-3.0.
   **Probe:** ONE shared resolver (file:line) read by UI + gate + prompt; the
   card renders the RESOLVED value. **Law:** a knob without a shared resolver
   is decoration.

9. **is_active-vs-binding.** Root cause: audits read `strategies.is_active=1
   LIMIT 1` and got a strategy NO trader binds (the false-strict-alarm: read
   4104ca0a advisory while the bound a5b7662e was strict×3).
   **Probe:** resolve strategy via `traders.strategy_id` (the TRADER BINDING),
   never is_active. **Law:** audits/sweeps query by binding; `is_active` is a
   legacy flag.

10. **Missing wire identity on materialized positions.** Root cause: the
    2026-08-25 reconcile incident — an NT8-held position with no DB row
    (untracked) materialized without full identity; sub-60s round-trips stayed
    invisible. **Probe:** `SELECT ... WHERE source='reconcile' AND
    (account='' OR entry_price<=0)` + the materialization regression test.
    **Law:** every materialized row carries source + account + entry; priced
    closes are consumed immediately (TestReconcileMaterializesUntrackedNT8Position).

11. **Enum spelling drift.** Root cause: the model wrote flip.rule "2x5m_close"
    vs the canonical "2x5m" → 2 silent rejections (armed wave 09:47/09:53).
    **Probe:** journal grep for the REJECTED spellings + a canonicalization
    function shared by parser and validator. **Law:** enum parsing normalizes
    known spellings and logs the unknown ones by name.

12. **Log-flood retention.** Root cause: per-frame INFO floods — bar_update
    (7.5M lines/day), backpressure WARN, order_update (1.48GB in ONE hour,
    25k lines/s) — each ate the journald 2G cap (retention < 1 day).
    **Probe:** after ANY logging change, re-project: measure bytes/hour of the
    busiest hour vs the 2G cap; target ≥7 days; per-frame logs are DEBUG +
    sampled (1/500) + a 1-line/min INFO summary. **Law:** every logging change
    ships with a retention re-projection.

13. **Concurrent-terminal.** Root cause: two dispatches in the main tree — one
    reset the other's uncommitted work out from under it (armed-orders vs
    level-truth; 6 dirty worktrees lost). **Probe:** porcelain gate
    (`git status --porcelain` empty) + the `~/nofx-main.lock` marker
    (owner/PID/expiry) before ANY main-tree work. **Law:** WORKTREE LAW — the
    main checkout belongs to exactly ONE dispatch; secondary work runs in
    `git worktree add ../nofx-<task>` + `git worktree lock`.

14. **Unattended deploys.** Root cause: timers/schedules performing cutovers
    (0-for-2 history — both failed). **Probe:** grep crontab/systemd timers for
    deploy commands — none may exist. **Law:** NO-UNATTENDED-DEPLOYS — owner
    ack within minutes OR a TESTED auto-rollback; timers are banned outright.

15. **Fantasy-R.** Root cause: planned R:R > ~6 on an arm — mathematically a
    fantasy target; the arm is refused every cycle and learns nothing.
    **Probe:** grep plans for `R:R = |target−entry| / |stop−entry|` > 6 + the
    arm feasibility contract (R:R ≥ 2.0 AND stop ≥ 1.0×ATR5m or the arm is
    refused). **Law:** the contract says what the gate enforces; fantasy
    targets get WARN-flagged at write.

16. **Small-n crowns.** Root cause: verdicts like "reject 75% win +665" quoted
    without n (a 3-trade week crowned as a rule). **Probe:** every verdict in a
    report MUST carry its n (e.g. "A-touch react 79% (n=52)"); n < 10 is
    explicitly labeled anecdote. **Law:** every verdict carries n; no crowns
    on small n.

17. **Secrets in baks.** Root cause: `.env.bak.*` and other backup files left
    in the tree with live keys; they also flip the binary's `+dirty` vcs flag.
    **Probe:** the dirty-flag account in every cutover report lists each
    untracked file by name + a secret-scan on every `.env*` and `*.bak`.
    **Law:** baks live OUTSIDE the repo (or are gitignored + scanned); the
    boot `+dirty` must be exactly the accounted list.

18. **Canon self-compliance.** Root cause: a wave changes a knob/play/gate but
    the guide still describes the old behavior — a guide that lies about the
    running binary is worse than no guide. **Probe:** in the SAME PR as any
    knob/play/chip/gate change: update `web/src/guide/content/*` + bump
    `GUIDE_BUILT_REV` to the shipped rev; the drift banner is a failsafe, not a
    maintenance strategy. **Law:** GUIDE CONTENT LAW.

19. **Half-shipped guard.** Root cause: a wave declares the guard's state
    (atomics, knob resolver, comments) but never wires the call sites — the
    pre-reopen F2 persist watchdog shipped `persistLastFlushAt`/
    `persistAlarmAt` + `persistWatchdogSeconds()` with ZERO `.Store`/`.Load`
    usages, so the alarm could never fire (caught 2026-08-29 by the
    pre-live-fire sweep). **Probe:** grep every guard atomic/knob for
    `.Store`/`.Load` call sites AND ship a BEHAVIOR fixture that makes the
    alarm FIRE (simulated stall → the ERROR line fires exactly once, quoted;

20. **OS-side fix that silently regresses.** Root cause: an OS-level remediation
    installs "successfully" but is handcuffed by a wrapper — chrony on WSL2 was
    started by chronyd-starter.sh which detected a container + missing
    CAP_SYS_TIME and appended `-x` ("Disabled control of system clock"), so
    `makestep 1 -1` never stepped, and the cron fallback called a binary that
    does not exist in the rootfs (`hwclock`). Result: 0.12s at 09:xx → −41s at
    17:01, NTPSynchronized=no. **Probe:** after ANY host-clock remediation,
    verify the fix's own mechanism actually fired (chronyc tracking shows a
    step-capable daemon; `journalctl -u chrony` free of "Disabled control of
    system clock") AND ship a machine-side escalation: at tolerance breach
    defer NEW plan authoring (negative drift = feed-in-future = provably
    broken) + widen T1 news windows by the measured drift. **Law:** the bot
    must never trade on a clock it knows is broken.
    resumed flush → stamp advances, no repeat) — existence tests cannot catch
    this class. **Law:** a guard without a firing fixture is decoration.

20. **Committed binaries / embedded secrets.** Root cause: `git add` of build
    artifacts — 14 tracked `nofx-bin.old*` binaries embedded a live-era
    DeepSeek `sk-` key in a PUBLIC repo (caught 2026-08-29 by T14's binary
    scan; every text-only secret scan missed it). **Probe:** `git ls-files`
    for binary artifacts + `strings`-scan EVERY tracked binary for `sk-`/
    key patterns; confirm any embedded key's hash ≠ every live key's hash
    (leak inert). **Law:** binaries are NEVER tracked; `.gitignore` covers
    every binary glob; a history rewrite requires the owner's explicit
    force-push ack and every clone/partner repo must re-clone.

21. **Log-lie counters.** Root cause: a success log reports a WRITE DELTA or
    a derived counter instead of the thing it claims — `maybeFetchCalendar`
    logged "fetched 0 events" on every healthy fetch because frozen
    `forexfactory` slices return `wrote=false` (`store/calendar.go:92-93`)
    and the line counted stored rows, not fetched events (23 false zeros
    08-27→08-29, caught by the 2026-08-29 news-feed forensics). A lying
    success line is worse than silence: it masks healthy state as broken
    and trains operators to ignore the signal. **Probe:** every "fetched N"
    / "processed N" log line must count the OBJECT of the verb, with write
    deltas logged separately; grep log lines against their counter
    definitions. **Law:** counters log what they name. Also: independent
    prompt re-renders (audit re-implementations) MUST enumerate EVERY input
    source — T9's renderer omitted `calendar_slices` and printed a false
    "(no filtered events)" for a populated prompt (same forensics).

22. **Unprobed supply chain.** Root cause: dependency audits were never
    automated — the F1 scan (2026-08-29) found 8 reachable Go
    vulnerabilities (x/text, quic-go ×2, pgx, go-ethereum ×3, jwt/v5) and
    14 npm findings incl. lodash + react-router HIGHs, all fixable by
    patch/minor, none ever bumped. **Probe:** `govulncheck ./...`
    (symbol-level reachability) + `npm audit --omit=dev` on EVERY wave;
    weekly CI + Dependabot (security-only) keep it that way.
    **Law:** CRITICAL/HIGH fixable by patch/minor bump in the SAME wave as
    the finding; major-version upgrades are owner-ruled, never auto-merged
    before a live-fire window.

23. **Report-only path panicked the trading loop.** Root cause
    (2026-08-30 entry-mechanics cutover): the E8 shadow A/B logger computed
    the counterfactual fill bar as `bucket_index × 5` — WRONG when the plan
    window starts mid-5m-bucket or spans <5 bars, so `w[5]` on a 4-bar
    window panicked `maybeManageArmedOrders` → `runCycle` → the whole bot,
    2 min after a clean boot. The shadow logger had zero gates downstream,
    yet one bad index took trading down. **Probe:** every advisory/report
    path MUST be panic-hardened — (a) boundary-safe index math proven by a
    fixture that reproduces the exact window shape (crossing a 5m boundary
    with 4 bars), (b) `recover()` at the call-site seam with a WARN, never
    a silent swallow. **Law:** a report-only path may degrade to a warning,
    never to a panic; the trading loop owns no `recover` blanket — each
    advisory seam carries its own.

24. **Armed re-place loop (manual cancel did NOT win).** Root cause:
    `UpsertArm` re-authorized TERMINAL rows every cycle while the confirm
    stayed MET and the placement band allowed the wrong side — the
    2026-08-30 S2 loop: terminal → armed → marketable fill (limit above
    market) → instant stop-out → terminal → armed… 8 generations in 26 min;
    an owner/NT8 cancel never stuck. **Probe:** (a) same-version
    re-authorization fixture (terminal row + UpsertArm same version → stays
    terminal; version bump → re-authorizes), (b) the wrong-side predicate
    `limitMarketableWrongSide` (long limit below market / short above =
    marketable, never placed), (c) journal has zero `WORKING` lines for a
    terminal row. **Law:** MANUAL-CANCEL-WINS — a terminal row is
    re-authorized ONLY on a plan-version change; a limit whose level the
    price already accepted through is cancelled, never placed.

25. **Far-side capability mismatch.** Root cause: the Go side sent a
    `stop_entry` frame to a pre-E7 AddOn, which executed the UNKNOWN frame
    type as MARKET — the 2026-08-30 test filled at 29346.25 instead of
    resting at 28700 (the far-side proof exists precisely to catch this).
    **Probe:** the AddOn reports a `build_id` on every heartbeat; the Go
    side refuses any frame type the far side hasn't proven
    (`FarSideProven` vs `FarSideBuildE7`) + a loopback fixture proving BOTH
    the refusal (no/old build) and the release (new build). **Law:** NO
    additive wire frame is sent before the far-side build proves it; a
    capability gate ships with its own negative fixture.

---

## PART 2 — PRE-AUDIT (standing hard rules)

- **R1 fresh evidence only** — produced THIS run: CT-timestamped queries,
  quoted journal lines, committed script deltas. Citing any prior report as
  proof = automatic UNVERIFIED.
- **R2 independent math** — recompute from raw stores; never call the function
  under test (the recompute is its own implementation).
- **R3 twin paths** — long/short mirrors both exercised.
- **R4 file:line** — every code claim cites the exact location.
- **R5 grades** — S/A/B/C; S-findings listed first.
- **R6 verdict grammar** — PROVEN / EVENT-WAIT (SHIPPED-UNPROVEN + the exact
  awaited event) / BROKEN / UNVERIFIED; never upgrade a grade without fresh
  evidence.
- **R7 pnl rule** — `pnl_corrected` everywhere + `excluded_null_pnl` for the
  354 legacy NULL rows (`WHERE pnl_corrected IS NOT NULL` in every expectancy
  query; position_query.go).
- **R8 times** — all times CT.
- **R9 isolation** — read-only sweeps run in a worktree at the RUNNING rev;
  zero code/config/DB/env changes; no restarts. Main tree untouched.

---

## PART 3 — PRE-CUTOVER (standing 7-step protocol)

1. **Tree gate:** porcelain-clean + `~/nofx-main.lock` acquired (owner/PID/
   expiry) + HEAD is the single allowed branch for this dispatch.
2. **Build:** from the MAIN checkout at the deploy commit (worktree builds lose
   vcs stamping → `<no-vcs>` → INTEGRITY REFUSED). `go build -o nofx-bin.next`.
3. **Marker:** `deploy/RELEASE` = the 8-char build rev, committed (marker AFTER
   build; RELEASE must equal the BUILD sha).
4. **Flat gate (all-origin):** API positions `[]` + DB OPEN=0 + NT8 open-orders
   empty ×2 + open-orders endpoint — all four quoted.
5. **Owner ack:** explicit "go" — reachable and acking the boot line within
   minutes, OR a TESTED auto-rollback. Timers banned.
6. **Swap:** `mv nofx-bin nofx-bin.old.<tag>` → `mv nofx-bin.next nofx-bin` →
   `kill -9 <PID>` (SIGKILL — SIGTERM exits 0 and systemd does NOT relaunch).
7. **Boot checklist (within 90s):**
   `🔐 BOOT INTEGRITY OK — rev <8char> +dirty · built <ts> · expected <8char> ·
   goldens PASS` + exactly ONE PID + feed warmed (bars_historical replay ~30s
   before decisions). **Rollback rule:** no boot line within 90s OR goldens
   fail → restore the prior binary + RELEASE, kill -9, restart, alert the owner.
