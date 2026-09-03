# AUDIT-CHECKLIST — the permanent audit playbook

Codified 2026-08-28 from the campaign's 18 proven bug classes. **Every audit
dispatch MUST reference this file instead of re-deriving the probe list. Every
NEW bug class found gets appended here in the SAME PR that fixes it** (canon law
in CLAUDE.md).

---

## PART 1 — THE 26 BUG CLASSES (name · root cause · probe · law)

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

21. **Committed binaries / embedded secrets.** Root cause: `git add` of build
    artifacts — 14 tracked `nofx-bin.old*` binaries embedded a live-era
    DeepSeek `sk-` key in a PUBLIC repo (caught 2026-08-29 by T14's binary
    scan; every text-only secret scan missed it). **Probe:** `git ls-files`
    for binary artifacts + `strings`-scan EVERY tracked binary for `sk-`/
    key patterns; confirm any embedded key's hash ≠ every live key's hash
    (leak inert). **Law:** binaries are NEVER tracked; `.gitignore` covers
    every binary glob; a history rewrite requires the owner's explicit
    force-push ack and every clone/partner repo must re-clone.

22. **Log-lie counters.** Root cause: a success log reports a WRITE DELTA or
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

23. **Unprobed supply chain.** Root cause: dependency audits were never
    automated — the F1 scan (2026-08-29) found 8 reachable Go
    vulnerabilities (x/text, quic-go ×2, pgx, go-ethereum ×3, jwt/v5) and
    14 npm findings incl. lodash + react-router HIGHs, all fixable by
    patch/minor, none ever bumped. **Probe:** `govulncheck ./...`
    (symbol-level reachability) + `npm audit --omit=dev` on EVERY wave;
    weekly CI + Dependabot (security-only) keep it that way.
    **Law:** CRITICAL/HIGH fixable by patch/minor bump in the SAME wave as
    the finding; major-version upgrades are owner-ruled, never auto-merged
    before a live-fire window.

24. **Report-only path panicked the trading loop.** Root cause
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

25. **Armed re-place loop (manual cancel did NOT win).** Root cause:
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

26. **Far-side capability mismatch.** Root cause: the Go side sent a
    `stop_entry` frame to a pre-E7 AddOn, which executed the UNKNOWN frame
    type as MARKET — the 2026-08-30 test filled at 29346.25 instead of
    resting at 28700 (the far-side proof exists precisely to catch this).
    **Probe:** the AddOn reports a `build_id` on every heartbeat; the Go
    side refuses any frame type the far side hasn't proven
    (`FarSideProven` vs `FarSideBuildE7`) + a loopback fixture proving BOTH
    the refusal (no/old build) and the release (new build). **Law:** NO
    additive wire frame is sent before the far-side build proves it; a
    capability gate ships with its own negative fixture.

30. **GORM alias scan silently reads 0.** Root cause: a `Scan` target struct
    field whose default GORM snake-casing disagrees with the SELECT alias —
    `TotalPnL` maps to `total_pn_l` (misses `total_pnl`), `NetPnL` maps to
    `net_pn_l` (misses `net_pnl`); the field silently scans zero while the
    query succeeds (found live 0A-2 in `GetPositionStats` total_pnl, again
    0C in `ab_confirm_log.net_pnl`). **Probe:** grep `Scan(&` targets for
    PnL-shaped fields without explicit `gorm:"column:…"` tags; ALWAYS assert
    a NONZERO expectation in the fixture (a zero assert passes the broken
    scan). **Law:** every Scan target ships an explicit column tag + a
    fixture whose expected value is nonzero.

31. **Inert-but-visible ledger state.** Root cause: a new non-terminal state
    name is invisible to `ListNonTerminal`'s `state IN ('armed','working')`
    filter, so assertion and cleanup paths silently skip it (0C's "shadowed"
    rows). **Probe:** when a new state ships, quote every `state IN` /
    `state =` filter site and add a fixture reading the new state through the
    SAME query the runtime uses. **Law:** a state is only "inert" if the
    queries that must skip it and the queries that must see it are BOTH
    fixture-pinned in the same PR.

32. **Time-scheduled action riding a data-gated cycle.** Root cause: the
    session read (registry 16:30/01:30/08:00) was invoked INSIDE `runCycle`,
    which `tickOnce` never enters when the bar-close gate or the no-new-data
    dedup idles the tick. 2026-08-31 evidence: CME is halted 16:00-17:00 →
    every cycle 16:26→16:38 logged `cycle_skip=no_new_data` → the 16:30 ASIA
    read fired at ~17:00:03 with the reopen tick — 30 minutes late, no error,
    no alarm, no plan at the open. Registry resolution, Sunday-defer (count=0)
    and trader liveness were all correct. **Probe:** for every time-scheduled
    action (reads, flats, rolls, weekly, digests), quote its invocation site
    and ask "does anything the market's calendar does delay this?" — a halted
    or quiet tape must never move wall-clock work. **Law:** scheduled work is
    evaluated on wall-clock BEFORE the data-gated skips; data gates may skip
    DATA work only. Halt-fired reads author from last stored bars and log
    `🗓 session read fired during halt … (newest <tf> <ts>, age <n>m)`.

33. **A cutover rite that checks exposure but not in-flight work, trusts a
    gate that cannot fail, and leaves the previous process's orders alive.**
    Three measured defects, one class. **(a) No in-flight leg.** PART 3
    steps 1-7 checked positions and orders, never running work: 2026-08-31
    17:34 CT a `kill -9` landed while a planner chain was on attempt 3/3 —
    the chain died silently, no v2, no fail-closed line, nothing re-claimed
    it. Four later cutovers held on this by agent discipline alone.
    **(b) Leg 4 could not fail.** `TCPTrader.GetOpenOrders` was
    `return []types.OpenOrder{}, nil` (tcp_trader.go:1149): the open-orders
    leg passed VACUOUSLY at cutovers 35, 36, 37, 38, 39, 40 and 41, and the
    full-system audit quoted "→ []" as evidence of flatness. NT8 emits no
    working-order frame (audit F12), so the `armed_orders` ledger is the only
    real source — it is what actually held the 09-01 swap (arm 29 WORKING).
    **(c) Pre-boot arms were orphaned.** 2026-09-02 00:16 CT a cutover ran on
    "just go" with S1 @29044 and S3 @29068.05 resting: the old process died,
    its broker orders did not, and they sat with NO listener for 15 minutes
    until the stale-window reconcile cancelled them at 00:31:48 — while the
    new binary re-armed its own S1/S3 and opened position 587 at 00:17:44,
    so for minutes TWO S3 orders existed at the broker. A fill on the dead
    process's order would have been a position no stop was attached to
    (class 27 again). It resolved by luck. **Probe:** for every gate leg, ask
    what input would make it FAIL — a leg with no such input is not a leg.
    For every restart, ask what the dying process leaves running at the
    broker. **Fix:** leg 4 reads the ledger (`AutoTrader.ledgerOpenOrders`,
    wired into `TCPTrader.SetOpenOrdersSource` at construction) and every row
    carries `source: "ledger (no NT8 order frame — F12 open)"`; an unwired or
    erroring source FAILS the leg instead of answering empty. Leg 5 is
    in-flight planner work (`AnyPlannerReadInFlight`, any date/session).
    `GET /api/cutover-gate` returns all five legs in ONE payload so an agent
    cannot quote four and skip the fifth; a leg that cannot be evaluated
    fails. The boot sweep (`sweepPreBootArms`, at the HEAD of
    `maybeManageArmedOrders`, before any authoring or placement) cancels every
    non-terminal row stamped with a different `boot_id` — cancel frame first,
    then `state=cancelled` with reason `boot_sweep: pre-boot order, process
    restarted`; a FAILED cancel leaves the row non-terminal and does not latch
    (never hide a live order behind a clean ledger); an authorized-but-never-
    placed row is left alone (nothing exists at the broker). Counter
    `arms_boot_swept_class33` in system_config. Boot line `🛡 cutover safety
    (class 33)`. **Law:** a gate that cannot fail is not a gate; a cutover
    checks running work as well as exposure; and no process may leave orders
    alive at the broker for a successor that never placed them.

34. **Validator hint naming a nonexistent condition.** Root cause: the
    breakdown-void reject said "author a reject/retest play instead" — the
    model authored condition `reject_retest`, and parse/schema rejected it:
    the model complied with the hint and was punished for it. 2026-08-31
    evidence: identical in BOTH ASIA chains, both fail-closed (v1 no_trade +
    the in-flight reset chain killed by the 0C cutover). Compounded by 0C:
    `breakout_retest` is shadowed, so the old hint steered toward either a
    nonexistent or a demoted condition. **Probe:** grep every validator
    message / repair-law excerpt for condition-shaped tokens; each must be in
    the enum AND resolve live. The registry + table test IS the guard
    (`kernel.ValidatorHints()` + `ValidateValidatorHints()`), re-run at boot,
    and the planner reject block now appends `Valid conditions: [<resolved
    live list>]`. **Law:** a hint is an instruction — instructions must be
    checkable; never name a composite or shadowed token as an authoring
    target.

35. **Counter inferred from row count (replan budget arithmetic).** Root
    cause: `ReplansUsedFrom = version − baseline` counted EVERY appended plan
    row as a spent re-plan, trigger-agnostic. 2026-09-01 LONDON: chain
    [planner_fail_closed, level_event, dormant:flip, level_event ×3] — zero
    death re-plans, zero owner re-reads — read as 5 of 4 spent, replans_left
    0; the next scenario death would have fail-closed a budget never touched.
    Compounded: death re-plan rows landed as `<S>_scheduled_read` (no class
    label), `trigger_reason` is overwritten in place by dormant/rearm
    transitions, and the card carried a THIRD formula
    (`noTradeVersion−2 : version−1`). Fourth silent-counter defect in one
    week: replan budget (this class) · guardrail ENTRIES count includes
    test-seam rows (open) · P&L summed realized_pnl not pnl_corrected (fixed)
    · GORM alias scan returned a plausible zero (class 30, fixed). **Probe:**
    for every cap/budget/quota, find the increment site — if the "used" value
    is derived (rows, versions, ids, timestamps) rather than written by the
    consuming path, it is inferred; fixture the live chain shape and assert
    the resolved value at the gate (`TestClass35PinTodayChain`). **Fix:**
    `store.GetReplanBudget` / `SpendReplan` — a recorded counter in
    system_config keyed `dayplan_replans_used:<trader>:<date>:<session>:b<baseline>`,
    incremented when a `death_replan` / `owner_reread` row lands; the card
    reads `replans_left` from the API. **Law:** counters record events; they
    do not infer them.

36. **Scheduled work inheriting the market's calendar — layer 2, the
    preflight (sibling of class 32).** Root cause: `plannerPreflight`
    (trader/auto_trader_feedwatch.go) compared the newest 1m bar's age to
    FEED_ALERT_S (600s) for EVERY planner trigger class. Class 32 fixed the
    TRIGGER (wall-clock evaluation), but the read then met a freshness check
    that is UNSATISFIABLE inside the 16:00–17:00 halt or on a weekend by
    definition. 2026-09-01: the 16:30 ASIA trigger fired; fifteen
    `stale_bars_1865s … 3545s` refusals 16:31→16:59; the read launched
    17:01:05 on the reopen tick, three attempts died on the same
    breakdown_continue-void class, fail-closed 17:23:14 (ASIA v1
    planner_fail_closed) — the halt refusal ate the 31 minutes that would
    have absorbed a retry BEFORE the open. The Sunday weekly read had the
    same shape (31 minutes late 2026-08-30) and ALSO still lived inside
    runCycle behind the data gate. Two contracts contradicted: "author from
    last stored bars" (class 32) vs the preflight's freshness requirement.
    **Probe:** for every scheduled action, walk the WHOLE path after the
    trigger — every gate it crosses must be satisfiable at the scheduled
    time by construction (a check that needs a live tape can never pass
    during a halt). **Fix:** the freshness check is SCOPED by trigger class
    (`preflightScheduledBypass`): scheduled session reads + the weekly
    bypass it only while `!IsCMEOpen(now)`; death_replan / owner_reread /
    level_event / structure_mss keep it; a scheduled read into a silent OPEN
    tape still refuses (the 08-19 outage class). The weekly read moved onto
    the wall-clock evaluator (`evaluateWallClockWeeklyRead`, before the
    session reads; `sundayAsiaDeferred` unchanged). Both outcomes are loud:
    `🗓 preflight bypass (class 36) …` (WARN) and `⛔ planner preflight refused
    <class>: <reason>` (ERROR). Executor halt block untouched
    (`cmeSessionClosedSkip` / `IsCMEOpen`). **Law:** never trading during a
    halt is the executor's rule; never AUTHORING during a halt defeats the
    open−30 design — a preflight may not refuse scheduled work because the
    market is closed when the work exists to run while the market is closed.

37. **Whole-request ceiling on a LIVE stream (split deadlines that did not
    split).** Root cause:
    the planner-speed wave (2026-08-31) moved the planner onto SSE with a 30s
    idle watchdog and DOCUMENTED "the whole-request ceiling stays
    http.Client.Timeout (600s) — a live-but-slow stream is never killed"
    (`kernel/planner_speed.go`, `mcp/client.go`, guide settings.ts). Both
    halves cannot be true: `http.Client.Timeout` bounds the body read, so every
    max-reasoning attempt still streaming at 600.0s died with
    `stream interrupted: context deadline exceeded (Client.Timeout …)`.
    2026-08-30 17:00 → 09-01 17:30 CT: 11 of 80 max full/re-author attempts
    (13.8%) killed at 600000-600001 ms with 71k-140k reasoning chars already
    received (ttfb 474-578 ms — the provider had answered); 0 of 22 repair and
    0 of 42 fast-reasoning attempts. Successful max reads p50 448s · p95 581s ·
    max 599.5s (right-censored). 2 of 9 fail-closed sessions had a kill consume
    an attempt (08-30 ASIA v3, 08-31 NY v2); 3 wake re-plans landed 15-30 min
    late. Compounded: the kill's transport text was fed to attempt 2 as a
    "validator reason" (planner_rejected_prompts 71-72), the `ai_call` line
    carried no failure class, and a failed call inherited the previous call's
    ttfb/reasoning numbers — so the owner saw "the API keeps failing".
    **Probe:** for every deadline claim in a comment/guide, find the
    transport-level timeout that still applies (`http.Client.Timeout`,
    `ResponseHeaderTimeout`, dialer) and fixture the live-but-slow case
    ACROSS that timeout (the 4.4 fixture used `Timeout: 10s` against a 0.5s
    stream — it never crossed the ceiling); grep `ai_call … ok=false` and
    demand a `class=` token on every line. **Fix:** planner stream rides
    `CallWithRequestStreamRetryDeadlines(idle, total)` — total =
    `AI_PLAN_TOTAL_DEADLINE_SECS` (default 1200, from the distribution) on a
    per-call http.Client copy with Timeout=0; the 600s ceiling stays on
    non-stream paths; `classifyAIError` + `context.Cause` stamp
    `class=total_deadline|idle_deadline|client_timeout|transport|http_status`,
    `http_status=`, `request_id=` on every ai_call failure; per-call telemetry
    is reset at call start; the planner logs `provider_row=` on failure; boot
    lines print idle/total/ceiling/retries/row. **Law:** a deadline a design
    says "never fires" is asserted by a fixture that crosses it, and every
    failed provider call carries a failure class — "the API keeps failing" is
    not a log line.

38. **Prompt/validator contract mismatch — the prompt offers what the
    validator refuses.** Root cause: the authoring contract lives in TWO
    places that drift. 2026-09-01 ASIA, ONE read, three attempts, three
    DIFFERENT rejects (`planner_rejected_prompts` 78/79/80): (a) the entry-law
    `Style` string — quoted VERBATIM into the rejection the model reads — said
    "2x5m legal ONLY here" while `confirm.rule`'s enum is
    touch|1x5m_close|2x5m_close|1m_mss|time_hold; "2x5m" belongs to the
    SEPARATE death/flip enum; (b) the model copied that token into
    confirm2.rule and was rejected for complying, and its repair came back
    unparseable; (c) the schema line offered `"legs":[…]` on EVERY scenario
    with no condition qualifier while its siblings fvg{}/breakdown{} on the
    same line DO carry "REQUIRED iff …", and the sweep_reclaim-only rule lived
    ONLY in plan_doc.go. Rendered-prompt counts before the fix: "arm_legs" 0 ·
    "split entry" 0 · "arm single" 0 · "EXACTLY 2 legs" 0 · "split contract" 0.
    Scale: 35 of 121 validator rejects in 72h were legs on a non-sweep
    condition (breakdown_continue 24, reject 11); 7 landed on attempt 3/3 and
    two sessions fail-closed on it. The class-34 guard stayed green throughout
    because it checked CONDITION tokens only. **Probe:** for every validator
    branch keyed by condition/field, grep the RENDERED prompt for the sentence
    that states it — absent = the defect; and scan every hint string for
    enum-valued tokens, checking each against the enum of ITS OWN field (the
    same spelling can be legal in one field and illegal in another).
    **Fix:** `kernel/prompt_contract.go` — a registry of the 17 condition-keyed
    restrictions with the sentence each requires, `ValidatePromptContracts`
    failing the build (table test) and shouting at boot
    (`📜 prompt/validator contract: N restrictions, all stated in prompt`);
    `ValidateValidatorHints` extended with `HintRuleField` so every rule token
    is checked against its own field enum, and the entry-law `Style` strings
    registered as hints; one token vocabulary everywhere, with the death/flip
    enum declared beside its own schema line. **Law:** the prompt states every
    restriction its validator enforces — a rule the author cannot read is a
    trap, and a hint is an instruction, so it must be legal in the field it
    describes.

39. **Reject-where-normalize-was-deterministic (legs on a non-sweep
    condition).** Root cause: the arm contract says every non-sweep_reclaim
    condition arms SINGLE, and the validator REJECTED any `legs[]` on such a
    scenario (`arm_legs_sweep_reclaim_only`, plan_doc.go ArmSpecValid) —
    burning one of three attempts (~450 s of max-reasoning each) on a shape
    whose correct form was already fully determined: the top-level arm. 72 h to
    2026-09-01: 35 of 121 validator rejects (breakdown_continue 24, reject 11);
    7 landed on attempt 3/3; two sessions fail-closed directly on it
    (`planner_rejected_prompts` rows 69, 80). The only retained instance (row
    69 S1, breakdown_continue, ONE leg, rule=touch) already carried a valid
    top-level arm that mirrored the leg — dropping the array alone made it
    pass; the two sweep_reclaim one-leg instances (rows 69 S2, 85 S1) must
    keep rejecting because there the fix would be authoring. **Probe:** for
    every validator reject, ask whether the correct shape is uniquely
    determined by the contract with NO invented value; if yes, it is a
    normalization (WARN) not a reject; if any value must be synthesized, it
    stays a reject. Precedent in the same file: level auto-collapse. **Fix
    (owner ruling 2026-09-01, verbatim):** on a non-sweep_reclaim condition
    with ANY legs array — drop the array, re-run the full arm validation on
    what remains; valid → proceed with a ⚖ WARN naming every dropped leg;
    still invalid → REJECT UNCHANGED with the original reason, no second pass;
    never synthesize a leg; never normalize the reverse. Implemented in
    `NormalizePlanDocRules` (runs before `validateArmSpecs`), recorded on the
    plan doc (`arm_normalizations`), stamped on the E8 row
    (`normalized`, `dropped_legs`), counted in system_config
    (`arms_normalized_class39`); `plannerRejectedCap` 20 → 200 so the next
    class has a sample. **Law:** the validator's job is to refuse what is
    ambiguous or unsafe, not what is merely misspelled — when the contract
    fully determines the answer, normalize and WARN; when it does not, reject
    with the original reason.

---

40. **Coerced aggregator inside the model's context window (corrected-column
    law on prompt-facing P&L).** Root cause: `EffectivePnL()` — corrected
    value if present, else raw `realized_pnl` — was the accessor every P&L
    aggregator summed (`GetFullStats`, `GetSymbolStats`, `GetRecentTrades`,
    `GetHoldingTimeStats`, `GetDirectionStats`, `GetHistorySummary`); two
    more kept a dead `COALESCE(pnl_corrected, realized_pnl)` fallback in SQL;
    the AgentBeta trade tool read the raw column outright. 2026-09-01
    evidence: decision record 36090 (23:07:13 CT, Sim101) told the executor
    `Total PnL: -203.68 USDT` over 220 trades, where the strict truth is
    **+304.32 over 105 resolved trades, 115 unresolved excluded** (rows
    237–586; row 526 alone: raw −1,458.00 vs corrected −69.43, the ×21
    lot-math artifact riding straight into the prompt); an unresolved short
    with exit 0 rendered as `Profit +0.00 USDT (+100.00%)`. Sign, magnitude
    and count were all wrong — every executor decision was made against a
    fabricated track record. The dashboard header showed the NT8-native
    total (0.00) beside a +212.00 ledger day total. Fourth silent
    counter/aggregator defect in a week (35 replan budget · guardrail ENTRIES
    · GORM alias zero · this) and the first found INSIDE the model's context.
    **Probe:** for every figure a prompt or tool renders, trace the column to
    the row: any accessor with an `else raw` branch, any COALESCE onto the raw
    column, any sum that does not return its exclusion count is coercion.
    **Fix:** `CorrectedPnL() (float64, bool)`; every aggregator strict, NULL
    rows counted as `UnresolvedExcluded` and excluded from sums/averages/win
    rates/streaks; prompt line `Track record: +X over N resolved trades (K
    unresolved trades excluded — see note)` and `#id side entry→? UNRESOLVED`
    rows; `/api/account` ledger day total (footer rule); build-time lint
    (`store/pnl_surface_guard_test.go`) fails on any raw aggregation outside
    the allow-list; boot line `🧾 P&L surfaces: N aggregators strict-corrected,
    0 raw`. **Law:** the model never reads a fabricated track record — a
    figure travels with its resolved n and its unresolved count, and an
    unknown is UNRESOLVED, never a coerced number and never a plausible zero.

41. **Provider mid-stream cut treated as a validator reject (transport
    resets).** Root cause: on 2026-09-01 four of 81 planner SSE calls (4.9 per
    100; 0 of 31 on 08-31) died mid-body — 01:46 `connection reset by peer`
    (RST), 23:47:13 / 23:47:33 / 23:52:41 `stream interrupted: unexpected
    EOF` after 250 s / 18 s / 308 s with 55k / 3k / 70k reasoning chars, all
    `http_status=200 request_id=""`. WHO closed the socket: reproduced
    in-process (`mcp/transport_cut_probe_test.go`) — a peer FIN mid-body
    yields exactly `stream interrupted: unexpected EOF` class=transport, a
    peer RST yields exactly the 01:46 string, and the idle watchdog yields
    `stream idle deadline exceeded … class=idle_deadline` (context.Cause is
    checked BEFORE the reader error, so a watchdog kill can never be
    mislabelled). Verdict: **THEM** — the peer (DeepSeek edge, a CloudFront
    distribution at `api.deepseek.com` → `d3bbv8sr76az5s.cloudfront.net`)
    or its origin closed a live HTTP/1.1 chunked response; our Go code is
    excluded [A]; a middlebox on the WSL2-mirrored / Windows path cannot be
    excluded from strings alone (passive socket-state watcher armed). Our
    side then made it worse twice: (1) the client retry waited a FIXED 2 s
    and call 2 died 18 s later on the same flap; (2) the planner loop
    treated the exhausted transport error as a VALIDATOR reject — attempt 2
    re-authored with `still failed after 2 retries: stream interrupted:
    unexpected EOF` appended as its "validator reason" (owner ruling class
    37 M4 had said: identical prompt, no reject block). **Probe:** for every
    failure path, ask whether the model ever answered; if not, there is
    nothing to repair and no reason to append — resend. For every kill
    switch (watchdog, deadline, ceiling), ask whether it LOGS when it fires;
    an unlogged switch makes "0 kills" an absence of evidence. **Fix:**
    `mcp.IsProviderFailure` (transport / idle_deadline / total_deadline /
    client_timeout / context) → the planner attempt loop re-sends the
    byte-identical prompt (`resend-identical`, no reject block, no
    rejected-prompt row); stream retries count CALLS via
    `AI_PLAN_STREAM_TRIES` (default 3) with the exponential schedule
    `AI_PLAN_STREAM_BACKOFF` (default 2s→15s→45s); the idle watchdog logs a
    `⏱ stream idle watchdog FIRED: Ns since last SSE line` WARN when it fires;
    dialer keepalive 30 s confirmed in effect (`ss -o` timer), unchanged;
    executor serialization NOT added (all 71 overlapped streams: 4 cuts; 6
    non-overlapped: 0 — no power, no effect shown). Boot line `🔁 planner
    stream policy (class 41)`. **Law:** a provider failure is retried, a
    validator reject is repaired — never append a transport error to a
    prompt; and every kill switch logs its own fire.

42. **Optimising the 4% (a wave aimed at the wrong term).** Root cause: the
    planner call was slow (p50 448 s, p95 581 s) and the plan JSON was assumed
    to be the output. It is not. Measured 2026-09-02 over n=67 full-author
    calls (2026-08-31 → 09-02, `prompt>9000`): **p50 completion 23,769 tokens,
    mean 22,376, mean wall 349 s, mean reasoning 72,477 chars**; the stored
    plan docs (n=61 since 2026-08-28) average **3,088 bytes ≈ 920 tokens**.
    The plan JSON is therefore **~4% of the output and reasoning is ~96%** —
    deleting the entire schema could not deliver the 40-50% cut the wave was
    scoped for, and the field-by-field audit found **no removable field**
    (every one of the 9 top-level fields has a reader: levels ~402 tok,
    scenarios ~237, reasoning ~161, no_trade ~42, bias ~33, death_condition
    ~18, flip ~14, death ~10, day_type ~3). **Probe:** before optimising a
    cost, measure its COMPOSITION and quote the share of the term you intend
    to cut; if the term is under ~20% of the total, the optimisation cannot
    reach a headline target no matter how well it is executed. Ask what the
    other 80% is made of. **Fix:** no schema cut shipped (it would have been
    risk without reward); the finding is pinned by two tests
    (`TestRootFixEveryPlanFieldHasAReader`, which fails if any field ever
    loses its last reader, and `TestRootFixPlanJSONIsASmallFractionOfOutput`,
    which fails if the JSON share ever exceeds 15%) and by the boot line
    `✂ planner schema`. The real lever — reasoning MODE — ships as a
    measurement instrument, not a change: a fast-mode shadow A/B
    (`SHADOW_AB_ENABLED`, default OFF) that re-runs the identical prompt at
    reasoning=fast AFTER the live read, validates it through the full chain
    offline, writes nothing, and is judged against a criterion registered
    BEFORE the data (legal-rate ≥ max AND median wall ≤50% of max at n≥10).
    **Law:** measure the composition before you optimise a total, and state
    the share you can actually reach; a pre-registered criterion beats a
    narrative when the result arrives.

43. **Uncited code-canon governing money (the [C] knob that was never
    researched).** Root cause: `MIN_SL_ATR_MULT = 1.0×ATR5m` shipped as the
    stop floor with no citation — the knob census labelled it **[C]
    code-canon**, and three gates read it (arm-time, AI-entry, planner
    authoring WARN). Round-7 research tests the day-trade range at
    **1.5–2.5×ATR** and finds stop-out rates above 60% on noise alone below
    1.0×; our own tape has **6 of 8 losers with MAE beyond the stop** and
    **15 of 27 losers stopped-too-tight**. Worse, width alone was never the
    defect: on the five biggest losers **0 of 5 stops sat ON a seated level**
    and **2 of 5 sat in dead zones 40+ points away** — a wider stop in a dead
    zone is still a stop in a dead zone. Two live exit mechanisms (BE+40, the
    ATR trail) were moving those stops **with zero measurement** (2 BE moves
    and 8 trail ratchets on 09-01; $719.50 of giveback with zero trail EXITS
    ever), and the size resolvers disagreed in production — arm-leg capacity
    resolved 1 while order sizing resolved 2 and the boot line said
    capacity=1. **Probe:** for every knob that decides money, demand its
    citation. A default with no research behind it is a guess with a
    confidence interval of the whole real line; a mechanism that moves a live
    stop without a measurement is worse than one that does nothing.
    **Fix (0B):** floor 1.0→1.5 (the BOTTOM of the researched range, not the
    middle); the stop is COMPOSED — beyond the nearest seated level on the
    risk side + clearance, floored at the ATR multiple, widest wins, never
    tighter than authored, `stop_unanchored` + ATR floor in a dead zone and a
    level is NEVER invented; BE and the trail suspended behind
    `EXIT_MECHS_SUSPENDED` with a single wire seam so a fixture proves zero
    move_stop frames; Stage-A size ceiling of 1 contract until n≥30 with a
    positive lower-CI expectancy (the floor raises dollar risk ~50% at
    constant size — which is why size does not move with it). **Law:** a knob
    that decides money carries a citation or a suspension — never a number
    someone once typed.

44. **The repair prompt judged the model against a vocabulary it never
    showed it.** Root cause: attempt ≥2 defaults to a REPAIR call
    (`BuildPlannerRepairPrompt`). It carried the rejected output, the validator
    errors and a law excerpt — but never `LiveConditionsLine`, which only the
    RE-AUTHOR tail appended (`plannerRejectBlock`). So from the moment class 34
    shipped the condition vocabulary, the DEFAULT retry path ran without it.
    Worse, `lawExcerptsFor` was a first-match `switch` whose cases matched
    neither `fade_requires_touch` nor `invalid (` — the two commonest confirm
    defects — so those fell through to a GENERIC excerpt about level labels and
    targets. **Measured** (all repair attempts, 2026-09-01 → 09-02, n=28): 18
    rejected at the parse/schema step (64%), 8 accepted, 2 rejected later.
    Of the 18: **1** packaging failure (`cannot unmarshal number 0.5 into …
    PlanArmLeg…size of type int` — a fractional contract size, 04:24:17) and
    **17 that PARSED CLEANLY** and were rejected on field values — 10 of them
    confirm-rule vocabulary errors (`"2x5m"` and `"displacement"` written into
    `confirm2.rule`, `1x5m_close` on a fade). **11 of 17 received an irrelevant
    law excerpt.** **Probe:** when a retry keeps failing, read what the retry
    prompt actually CONTAINS, not what the author path contains — a default
    path and a fallback path drift apart silently. And check whether an error
    router is first-match when errors can be plural. **Fix:** the repair prompt
    carries `LiveConditionsLine`; `lawExcerptsFor` collects EVERY applicable
    excerpt; a new `RepairConfirmVocabLaw` names the confirm enum and states
    that the death/flip enum is a DIFFERENT vocabulary (class-38 rule); the
    return contract is restated at head AND tail (lost-in-the-middle); a
    fragment gets its own reason instead of a confusing schema error; outcomes
    are classified (`ok|content|packaging|fragment|no_outcome`) and RECORDED in
    system_config, replacing a log line that called all 18 "UNPARSEABLE".
    **Note the dispatch's premise was wrong and the audit's was right about the
    RATE only:** extraction already tolerated fences and prose
    (`extractJSONObject` scans to the first `{`), so an extractor rewrite would
    have fixed 0 of 18. **Law:** every retry path shows the model the same
    vocabulary the validator will judge it by; and a diagnosis label must name
    what actually happened, or it hides the defect it was added to expose.

45. **A pantry nobody could reach (two bar layers, no resolver) — and a
    provider calendar that is not ours.** Root cause: two unconnected bar
    layers existed with no single door between them. The NT8 BarCache held
    native per-TF series (measured live on 0465a10b, 2026-09-02: **1w 383
    bars back to 2019-05-03 · 1d 1500 back to 2020-11-11 · 4h 1500 back to
    2025-09-11 · 1h 1500**), all memory-only; the persisted `bars` table held
    **1m only**, because `InsertBars` carried the line `if r.TF != "1m" {
    continue }` — so every restart discarded the entire coarse pantry. The
    weekly reader read the 1m table directly (`BarsBetween(symbol,"1m",…)`),
    which starts 2026-08-19, saw **2 completed weeks** against a ≥4 guard, and
    rendered "WEEKLY thin · low" while 383 native weekly bars sat unread. Four
    hand-rolled 1m→TF aggregators each bucketed on their own convention and no
    caller could answer where a daily bar came from. **The second, sharper
    defect:** NT8's native weekly bars run **Friday 00:00 → Thursday 23:59**,
    while every weekly concept in this system is Monday-governed
    (`weekStartMonday`; "Sunday 17:00 CT first print"; PWH/PWL from the prior
    Monday week). Pointing the weekly reader at native 1w — the obvious
    "nt8-first" fix — would have shifted every week by three days and replaced
    an honestly-labelled *thin* with silently WRONG data. **Probe:** before
    consuming a provider's aggregate, print its first and last bar timestamps
    and ask which calendar they are on; never infer the convention from the
    TF's name. And for any two-layer store, ask what single function answers
    "give me completed bars for X" — if none exists, every consumer has its
    own answer. **Fix:** one resolver (`market.CompletedBars` /
    `CompletedBar`) with the fallback ladder as DATA (`barLadder`), the
    repaint law applied at one chokepoint (`dropForming`), and the source
    (`nt8|nt8_agg|own1m`) travelling with the bars. Weekly's ladder is
    `1d → 1m`: native 1w is EXCLUDED with its reason recorded in
    `ladderExclusions`, and `StampAligned()` catches the same mismatch class
    generically. Every cached TF is now persisted; retention became PER-TF
    (`tfRetentionDays`) because the old TF-blind 90-day cutoff would have
    deleted the deep weekly history on the first nightly prune after it was
    stored — 1m 90d, 5m 180d, 15m 365d, 1h and coarser forever, ≈31 MB steady
    state. Native 1w is persisted labelled `convention=fri_thu` for research
    only. **Law:** one resolver, one stamp convention, and a provider's
    calendar is a measurement — not an assumption you inherit from a
    timeframe's name.
47. **A scheduled mechanism paced by its own throttle, not by events.** Root
    cause: the level-event wake had exactly one limiter — `wake_min_interval_min`
    (30 min) — and no notion of whether a wake could still produce a TRADEABLE
    plan. Measured 2026-09-02: 60 wake re-plans in 7 days against 33 arm rows, 23
    ever placed and 9 ever working/filled; today's wakes fired at 08:42:30 ·
    09:12:30 · 09:42:30 · 10:14:30 · 10:44:30 · 11:15 · 11:45 · 12:16:30 ·
    12:48:29 · 13:18:29 · 13:48:29 · 14:20:29 — a clean ~30-minute drumbeat,
    which is the signature of a condition that is CONTINUOUSLY true being paced
    by the throttle rather than by events. NY bought 12 plan versions that way,
    including a max-reasoning read at 14:20:29 that sat 10 minutes from the
    last-entry cutoff and 25 from the flat: a plan that could never be entered.
    Two adjacent defects fell out of the same audit: (a) the planner in-flight
    claim is keyed per (trader, trade_date, session), so a LONDON read and an NY
    read hold different claims and stream concurrently — 08:01:06 today opened a
    second max-reasoning stream while 07:51:06 was still running; and (b) a
    NEVER-PLACED arm row survives its own plan version indefinitely — NY row 32
    (v5, S3, no signal id, 10:30:30) stayed non-terminal until the 14:45 EOD
    flat, ~4h15m across v5→v12, holding the class-33 cutover gate's leg 4 shut
    the whole time. **Probe:** for any periodic mechanism, plot its firing times.
    Even spacing at exactly the throttle interval means the throttle is the
    scheduler and the trigger is noise; then ask what the work it produces is
    still USED for. **Fix:** cutoff and cooldown land WARN-first — they log
    `would_skip` with a recorded per-session counter and the wake still runs, so
    the suppression ruling is made on a week of counts rather than impressions;
    a wake (never a scheduled read) defers while any planner stream is open; a
    never-placed arm from a superseded version goes terminal `superseded`, with
    placed rows untouched. **Law:** a mechanism whose only limiter is its own
    throttle is not scheduled, it is idling at rate — and before suppressing it,
    measure what suppression would have cost.

48. **The decision path bypassed the arm-seam gates.** Root cause: the five
    entry protections lived ONLY at the arm seam (`armed_executor.go`): the
    R:R floor (`armMinRR`), the 0C shadow map (`conditionShadowedFor`),
    scenario-direction consistency, stop composition (`composeArmStop`) and
    one-live-arm (`oneLiveArmGuard`). The AI market-entry path
    (`auto_trader_orders.go` → `executeOpenLongWithRecord` → `trader.OpenLong`)
    ran a different, thinner chain: `validateDecision` enforced R:R + min-SL +
    HTF veto at the PROMPT-TIME SNAPSHOT price while the fill is a MARKET
    order ~10 points away, and nothing checked shadow / scenario-direction /
    one-live-arm / stop composition; the agent chat `execute_trade` ran almost
    none. **Measured 2026-09-02:** 587 R:R eval `2.03 → PASS` @ snapshot
    29069.50, filled 29079.25 → real R:R **1.09** (below the owner's 2.0
    floor); 589 and 590 traded the SHADOWED condition `breakout_retest`; the
    08:13 R:R 3→2 save had no persisted row (the class-44 `config_changes`
    table wires only future saves). **Probe:** for every protection, list the
    call sites — a gate whose only callers live in one file is absent from the
    other path; and a floor judged on a stale reference is not a floor.
    **Fix:** ONE `EntryGate` (`trader/entry_gate.go`) — legs direction →
    shadow → R:R-at-live-price → min-SL → one-live-arm — called by BOTH seams
    before any order leaves; refusals recorded per path
    (`decision_records.Error` + gate-block counter; arm-refusal counters).
    **Law:** a protection that exists on one order path and not the other is
    not a protection — it is a suggestion.

49. **Instrument theatre: a boot line that could not be wrong, a label that
    usually was, and a watchdog that never fired.** Root cause: the
    instruments reporting on the transport wave were themselves unmeasured.
    **(a) The class-41 boot line printed `watchdog_log=on`,
    `serialize_executor=off`, `resend_identical=on` and `keepalive=30s` as
    STRING LITERALS, and its fixture asserted the same literals** — the pair
    could only ever agree with each other, never with reality, and reality
    differed: keepalive on the wire ran 14-20 s. Class 6 (Go-side theatre) on
    the very wave shipped to make transport honest. **(b) `timeout_source`
    DEFAULTED to `"transport"`** and was overridden for four sentinels only,
    so it tagged 5xx bodies, parse failures and empty 200s as transport: right
    on 5 of 50 audited failures, wrong on 23. Two labelling systems
    (`timeout_source` and `class=`) disagreed on the same line. **(c)
    `IsProviderFailure` returned false for `class=other`**, so an empty 200, a
    parse failure or an over-long answer was appended to the next prompt as
    the MODEL's defect ("fix this: unexpected EOF") — the class-34/37 poisoned-
    feedback disease, still open. **(d) The idle watchdog reset on every
    scanned SSE LINE, including DeepSeek's `: keep-alive` comments**, so a
    stalled generation that was still heartbeating ran to the 1200 s ceiling.
    It had never fired once — and its close was indistinguishable from a peer
    EOF, so "0 idle kills" could not be told from "0 idle kills we can see".
    **(e) A 503 burst produced 3 planner attempts × 3 client tries = 9
    provider calls in ~7 s** at an edge already shedding load. **(f) The
    sockwatch bash loop wrote 12,947 lines and caught ZERO FIN/CLOSE-WAIT
    states** — blind to the single thing it was built for. **Probe:** for
    every field an instrument prints, ask which function ENFORCES it and
    whether the fixture calls that function; a fixture that asserts the same
    literal as the code tests nothing. For every default label, ask what
    fraction of cases it is right about. For every kill switch, ask what input
    would make it fire, and whether its output is distinguishable from the
    thing it is meant to detect. **Fix:** `mcp.PlannerClientPolicyLine()` —
    every field read from its enforcer, keepalive from the one place that sets
    it, and `observed=n/a` rather than implying the set value is the seen one;
    `timeout_source` DELETED and `ClassifyFailure(err, httpStatus)` the only
    classifier (it sees the status, so a 503 body is `http_5xx` even when its
    text says EOF, and `http_status` splits 5xx/4xx so retry policy can tell
    "provider overloaded" from "our request is wrong");
    `FailureIsProviderSide` decides resend-vs-repair, with `parse` deliberately
    MODEL-side (resending an unparseable document identically would loop
    forever — the pre-existing class-41 fixture caught that over-generalisation
    during this wave); a two-timer watchdog (pre-token, heartbeats allowed;
    post-token, reset ONLY by content/reasoning deltas) closing with its own
    `ErrWatchdogIdle`; a per-READ storm cap (`AI_PLAN_STORM_CAP`, default 5);
    and `httptrace` reporting `closed_by=peer_fin|local_close|clean` from
    inside the process, labelled INFERRED because httptrace sees no TCP flags.
    Rider (owner ruling): the confirm enum attaches to every repair whose
    DOCUMENT carries a confirm object, not only when the incoming error names
    one. **Law:** an instrument that cannot disagree with the code is not
    evidence — every printed field reads from its enforcer, every default
    label earns its default, and every kill switch must have an input that
    fires it and an output you can tell apart.

50. **The prompt withheld what the validator enforces — and the correction
    remembered only the last mistake.** (Dispatch "class 45"; checklist slot 45
    was already the pantry class, hence 50.) Root cause: the planner prompt and
    the plan validator are two statements of the same law, written in different
    places, and nobody diffed them. **(a) The prompt ORDERED A PLAY, the
    validator VOIDED it.** Line 589 read "If price sits BELOW PDL you MUST write
    a continuation short" — an unconditional MUST naming one condition. But
    `BreakdownContinueState` voids a breakdown continuation once a close comes
    back across the level, so on any reclaimed level the only compliant answer
    was a rejected one. The model obeyed and was punished; the 2026-09-02 LONDON
    01:32 read burned attempt 1 exactly this way. **(b) The stop floor was
    enforced but never stated.** 0B raised the arm-time floor to 1.5xATR5m and
    the planner was told no number at all, so it authored stops that were
    silently widened at arm time — and the R:R gate then judged the WIDER stop,
    refusing arms the model believed it had sized correctly. **(c) The
    correction block carried only the LAST defect.** Same read: attempt 1
    rejected for the voided breakdown, attempt 2 for a fade with no touch,
    attempt 3 told only about the fade — it fixed the fade and walked back into
    the void, spending the whole chain on two defects it had already been told
    about separately. The block also sat at the TAIL of a ~6,600-token prompt,
    the position most likely to be skimmed. **Probe:** enumerate every MUST in
    the prompt and name the validator function that can reject a document
    obeying it — any MUST that names a CONDITION rather than a DIRECTION is a
    contradiction waiting for the right market. For every gate that can rewrite
    or refuse the model's output, ask whether its threshold appears in the
    prompt as a number. For every retry, ask whether it states the defects of
    ALL prior attempts or only the most recent. **Fix:** the MUST now orders a
    DIRECTION and names the legal conditions; `ComputeVoidBreakdownLevels`
    lists every already-reclaimed breakdown level in the prompt by CALLING
    `BreakdownContinueState` itself (a parity fixture pins 40/40 checks across
    20 tapes, so the list cannot drift from the verdict); `RenderStopFloorLine`
    prints the floor with the live ATR reading; `addDistinctReject` accumulates
    the chain's DISTINCT defects and `plannerRejectHeader`/`plannerRejectTail`
    state them at the TOP and the TAIL (~240 tokens), recorded at all eight
    reject sites. **Law:** the prompt must state every rule the validator will
    enforce, in the validator's own words and by calling the validator's own
    code where a verdict is involved; a correction is cumulative or it teaches
    the model to trade one mistake for another.
51. **A direction was shipped on evidence that never existed — the weekly bias
    was anti-predictive.** (Dispatch "weekly refs only"; class 50 wave.) Root
    cause: the weekly-bias design (2026-08-30) assumed the Sunday doc's
    directional call carried signal, and every consumer — the W4 invalidation
    watch, the F5 write-time DOA stamp, the draw-alignment tag, the
    WEEKLY-COUNTER shadow — read `WeeklyDoc.Bias` as a direction. The first
    out-of-sample test of that direction (bias calibration 2026-09-02,
    pre-registered; report `docs/superpowers/reports/2026-09-02-bias-calibration.md`)
    found the reconstructed rule ANTI-predictive on holdout (raw hit 25–28%,
    called-only 45–51%; net-of-friction t ≈ −14) — a label was governing prompts
    with negative signal, and no reader had ever asked for its evidence.
    **Probe:** for every advisory/soft-law label in a prompt, ask what
    out-of-sample test earned it the label and what it would take to strip it;
    a "shadow" annotation that READS a direction is a consumer, not a shadow.
    **Fix:** the weekly doc is REFS ONLY — weekly_levels (PWH/PWL/IPDA/NWOG)
    plus a facts-only narrative; the validator now REJECTS directional tokens
    (r4) instead of demanding them; the chip and both prompt lines render
    "WEEKLY: refs only — PWH x · PWL y"; the invalidation watch, DOA stamp,
    counter shadow and draw-align tag are retired (nothing reads bias as a
    direction anymore); the deterministic rule survives as `shadow_bias` on the
    doc + a log line so the anti-prediction keeps being measured — never read,
    never inverted. **Law:** a directional label is evidence or it is noise; a
    label with measured negative signal must be demoted to shadow, and nothing
    — not even a "shadow" annotation — may consume it as a direction.

52. **A rule rendered from the clock it was WRITTEN by, not the clock it is
    READ by.** The plan card's no-trade list was the model's prose, stored at
    authoring time and shown verbatim for the rest of the session. On
    2026-09-02 the ASIA card at 23:00 CT listed three constraints and not one
    of them could refuse an entry: the first-5m band had closed six hours
    earlier, the lunch window belongs to NY and cannot apply to an ASIA
    session, and the red-news blackout had fired fourteen hours before. The
    same defect had a second face: the windows were defined THREE times — the
    entry gate's `cur < start+N` and its `InBlackoutWindow(t,"12:00","13:30")`,
    the adherence grader's own copies, and whatever prose the model happened to
    write — so the gate refused one window, the grader scored another and the
    card claimed a third, with nothing that could fail if they drifted apart. A
    fourth copy sat in the prompt, teaching the model windows nobody enforced.
    And the clock-drift widening appended "+1m (clock drift)" for ANY nonzero
    offset, so a healthy 108 ms NTP reading put "the clock is drifting" on the
    card for the whole trading day. **Probe:** for every rule a surface
    displays, ask WHEN it was evaluated and against WHOSE clock; a rule
    rendered from stored text is a rule frozen at write time. For every window,
    threshold or window name that appears on more than one surface, find the
    single function all of them call — if there isn't one, count the copies.
    For every advisory suffix, ask which measurement makes it TRUE and whether
    that measurement can be zero. **Fix:** `kernel/no_trade_band.go` holds one
    definition each (`FirstNoTradeMinutes`, `LunchWindowCT`) read by the gate,
    the grader and the card; the machine writes `no_trade_windows` onto the doc
    at plan time, taking red news from `t1WindowsFor` — literally the windows
    the gate will refuse inside, widening and fail-closed fallback included, so
    the card cannot compute a second answer; `EvaluateNoTradeWindows` stamps
    each one live / elapsed / other_session against the READER's clock, asking
    session geometry first (does this window touch the session at all) and
    elapsed second; the model's prose is untouched and renders as notes; the
    prompt's example and rule sentence are generated from the same functions,
    with a literal scan allowing exactly one copy of each bound at the
    definition site; and the drift claim is made only when the offset alone can
    move an event by a whole minute, while the minute of boundary protection is
    kept for every input. **Law:** a surface renders a rule's STATUS, never its
    text — and a rule with more than one definition has none.

53. **One question, two answers: a predicate shared by two callers that fed it
    different inputs.** (Numbered 51 at merge against a tree that did not
    yet carry class 50's entry; renumbered to 53 by owner ruling 2026-09-02 —
    class 50 keeps 51, the no-trade band keeps 52. Class 46 is deliberately
    free — see class 50.) Root cause: the prompt's VOID list and the write-site
    validator both ask "has a close come back across this level?" and both call
    `BreakdownContinueState`, but the render passed
    `sinceMs = CMESessionDayStart` over a 12,000-bar slice while the validator
    passed `sinceMs = 0` over 2,000. The predicate filters on
    `b.OpenTime < sinceMs`, so the window is load-bearing: a level broken and
    reclaimed before the 17:00 CT boundary was void to the validator and
    invisible to the prompt. On 2026-09-02 20:58 CT the prompt listed eight
    seated levels VOID, omitted ONL 29141.25, and the read was rejected on
    exactly that level. **The class-45 parity fixture passed 40/40 across 20
    tapes while this was live**, because it fed BOTH sides the same `sinceMs`:
    it pinned the two FUNCTIONS and never the CALL SITES. **A second-order
    trap:** the first fix made both sides read the whole slice, which on the
    real tape voids 20 entries across 12 levels — a list saying "author no
    waterfall play anywhere", noise by construction. The deleted render-side
    comment ("a level broken and reclaimed days ago is not today's news") was
    RIGHT; its error was being applied to one caller only. **Probe:** for every
    predicate with more than one caller, diff the ARGUMENTS at each call site,
    not the function; then ask which caller's answer the user sees and which
    one enforces. Any test that constructs both sides' inputs itself proves only
    self-consistency. **Fix:** `kernel/void_scope.go` — `VoidScope` +
    `ResolveVoidScope`; neither caller chooses a window or a slice, and the
    scope VALUE is owner-ruled as the CME session day, so the VALIDATOR narrowed
    too (a rule change in the permitted direction: strictly fewer rejects).
    `voidWindowStartMs` deleted. Parity fixtures pin the CALL SITES in both
    directions: a pre-session reclaim is void for NEITHER, an in-session reclaim
    for BOTH. Plus `planner_read_facts` — one row per read, ACCEPTED OR
    REJECTED, carrying the rendered void list, floor, ATR and resolved scope,
    because until now a rendered prompt survived only when a read FAILED and a
    working fix erased its own evidence. **Law:** a predicate with two callers
    has one resolver for its inputs; a parity test that builds both sides'
    arguments tests nothing; and the instrument that proves a fix works must
    not fire only when the fix fails.

54. **A default of 0 on a column that means "how far did it go against us".**
    (Highest occupied at merge: 53.) `trader_positions.mae REAL DEFAULT 0`
    and `mfe` the same. A trade whose excursion was never computed and a trade
    that never went against us are then the same bit pattern, and no reader can
    tell them apart. Measured 2026-09-02: of 586 closed positions **517 carried
    mae=0 AND mfe=0** — the never-computed signature, since price always moves —
    while 9 carried a single genuine zero beside a real number. Round 7 ruled
    that exit rules, stop sizes and targets come from MAE/MFE distributions, so
    every one of those rulings was waiting on a column that was 88% unreadable.
    Two more faces of the same defect: `kernel.LearningLine` guarded its average
    with `if t.MAE > 0 || t.MFE > 0`, which silently drops a genuine zero AND
    counts an uncomputed row as merely absent, then printed an average with no n
    at all; and `ComputeExcursion` filtered bars with `b.OpenTime < entryMs →
    skip`, so unless a fill landed exactly on a bar boundary the bar CONTAINING
    the entry — the one holding the first adverse move — was excluded. **Probe:**
    for every numeric column, ask what value means "not measured" and count the
    rows holding it; if that value is also a legal measurement, the column
    cannot answer the question it exists for. For every average, ask what its n
    is and whether the reader is told. For every window over bars, ask whether
    the boundary bar is in or out and whether the answer is the same at both
    ends. **Fix:** `trade_excursions` — one row per position, every numeric
    NULLABLE and NULL until computed, written by the machine at open, recomputed
    each tick while the position lives, closed with the exit half, and carrying
    what a distribution needs: extremes with their timestamps and bar offsets,
    bars held, bars that reached BOTH the stop and the target, the resolution
    the path was built from, and pnl_corrected. A hold the tape does not reach
    is marked `resolution="none"` and keeps its NULLs — never a guessed number,
    and the count of those rows rides the boot line. `mae`/`mfe` on
    trader_positions became `*float64` with a migration that nulls only the
    517-row pair; `LearningTrade` carries `Measured` and the line prints
    "(n=1 of 2)"; `ComputePathExcursion` counts every bar whose window
    intersects the hold. **Law:** a column that cannot say "unknown" cannot be
    the input to a ruling — and an aggregate that hides its n is not evidence.

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

## PART 3 — PRE-CUTOVER (standing 7-step protocol; flat gate = 5 legs, class 33)

1. **Tree gate:** porcelain-clean + `~/nofx-main.lock` acquired (owner/PID/
   expiry) + HEAD is the single allowed branch for this dispatch.
2. **Build:** from the MAIN checkout at the deploy commit (worktree builds lose
   vcs stamping → `<no-vcs>` → INTEGRITY REFUSED). `go build -o nofx-bin.next`.
3. **Marker:** `deploy/RELEASE` = the 8-char build rev, committed (marker AFTER
   build; RELEASE must equal the BUILD sha).
4. **Flat gate — FIVE legs (class 33), all quoted:** `GET /api/cutover-gate`
   returns them in one payload; quote it, do not assemble them by hand.
   (1) DB `trader_positions` OPEN = 0 · (2) API positions `[]` · (3) NT8
   positions snapshot count = 0 · (4) **working orders = the `armed_orders`
   ledger's non-terminal rows** (NT8 emits no working-order frame, audit F12 —
   before 2026-09-02 this leg was a stub returning empty and passed vacuously
   at cutovers 35→41) · (5) **no in-flight planner work** — `replan_in_flight`
   false AND no planner read claimed for this trader on any date/session (the
   2026-08-31 17:34 defect: a kill landed on attempt 3/3 and the chain died
   silently). A leg that cannot be EVALUATED fails. `ready:false` = HOLD.
5. **Owner ack:** explicit "go" — reachable and acking the boot line within
   minutes, OR a TESTED auto-rollback. Timers banned. **Override rule
   (class 33):** the owner MAY override leg 4/5 and swap with arms resting —
   the override is permitted, leaving orders alive is not. Such a cutover
   REQUIRES the boot sweep to run and its result to be quoted in the report
   (`🛡 boot sweep CANCELLED pre-boot arm …` per row, or `cancelled 0`).
6. **Swap:** `mv nofx-bin nofx-bin.old.<tag>` → `mv nofx-bin.next nofx-bin` →
   `kill -9 <PID>` (SIGKILL — SIGTERM exits 0 and systemd does NOT relaunch).
   The classifier denies the kill to the agent: print the command and have the
   OWNER run it.
7. **Boot checklist (within 90s):**
   `🔐 BOOT INTEGRITY OK — rev <8char> +dirty · built <ts> · expected <8char> ·
   goldens PASS` + exactly ONE PID + feed warmed (bars_historical replay ~30s
   before decisions). **Rollback rule:** no boot line within 90s OR goldens
   fail → restore the prior binary + RELEASE, kill -9, restart, alert the owner.
