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
    target. (Class 33 is unoccupied — this wave shipped as 34 per dispatch.)

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
