# Full audit and fix report — Saturday 2026-09-26

**For:** the owner. **Written by:** the CTO at 01:00 CT on Saturday 2026-09-26.

**Live system:** boot 3, nofx `6cac1b89`, release marker `04ae1c2f`, booted Friday 19:33 CT and restarted at 20:14 CT with `FAST_MARKET_REASONING=max`. The bot is flat. CME is closed until Sunday 17:00 CT.

Evidence tiers: **[A]** means someone ran it or read the exact line. **[B]** means inferred from strong evidence. **[C]** means a guess.

---

## 1. Bottom line

- **Nothing lost money, and nothing is dangerous in the live bot this weekend.**
  - There are no P0 findings in any audit, after a second, independent lane cross-checked each one.
  - The bot is flat and the market is closed.
- **One real money risk was found, proven and is being fixed now.**
  - An entry order can be **sent twice after a reconnect**.
  - It has **never happened**: there are 0 failed sends in every live log and 0 duplicate order ids.
  - Nothing in the code stops it today.
  - The fix is being built (§5, DS-102).
- **The main reason plans die was found, and the fix is being built.**
  - In the current era, 14 of 36 planner reads died.
  - The top killers are:
    - plans already dead when published (the market moved during the 7–15 minute read);
    - broken plan JSON;
    - a missing level in the obstacle list.
  - It is **not** the model. DeepSeek Flash was tested and was **not faster** at max, and it failed more often. We keep Pro, and reasoning stays at max.
- **Settings.** Every setting was censused, 419 in total:
  - 16 do nothing;
  - 3 conflict with another setting;
  - 13 are overridden by another setting;
  - 110 labels say something the code does not do.

  Four of the dead ones are **risk controls that look on but do nothing**: max contracts, max margin usage, min position size and notional cap. Each is being wired to a real guard or removed. Per your rule, no cosmetic fixes.
- **Row 618 is fixed** (your "yes"): exit 30913.75, **+$19.50**, reason `tp`. I backed up first.
- **Sessions:** all three are ON, as you set them.
- **Database:** nothing is ever deleted; retention is OFF, as you ruled ("we are testing").

---

## 2. How this was checked (the process)

1. **3 audits** of the live code at `04ae1c2f`, all read-only. Each lane read a copy of the database, never the live one.
   - Settings and knobs: DS-102, PR #228, with a 419-row census CSV.
   - Trading pipeline: DS-106, PR #229.
   - System pipeline: DS-107, PR #230.
2. **Every audit was cross-checked by a different lane**, which tried to disprove it:
   - #232 checked trading entries, execution and exits;
   - #235 checked the trading planner and arming;
   - #233 checked the system audit;
   - #234 checked the settings audit (68 of 68 sampled values correct);
   - #239 re-checked the partner repo.
3. **Deep dives on every serious lead:**
   - duplicate entry, #236, proven;
   - stop-entry safety, #237, safe;
   - Flash vs Pro, #231;
   - the planner-failure count, #238, cross-checked by #242;
   - dead code, #241.
4. **The CTO spot-checked every serious finding** by reading the cited lines at the live commit, and caught **3 wrong claims** (§4).

---

## 3. Verified findings and what happens to each

| # | Finding | Severity | Verified | Status |
|---|---|---|---|---|
| 1 | **Duplicate entry after reconnect.** A half-sent entry is re-sent after a reconnect. Nothing checks whether NT8 already has it: not Go, not the AddOn, not NT8 (`tcp_server.go:2574-2606`, `VLTraderTCPClient.cs` ~780-860). Worst case is a double fill with two brackets and double size. | **P1** (never happened) | [A] #232 + #236, CTO read | **FIXING** — DS-102 `fix/double-entry-replay` |
| 2 | **4 dead risk controls** in Studio: max_contracts_enabled, max_margin_usage, min_position_size, notional_cap_enabled. They look active but have no reader. | **P1** | [A] census + #234 | **FIXING** — DS-105 `fix/knobs-truth`, wire or remove |
| 3 | The database lock wait (`busy_timeout`) reaches only 1 of the 4 DB connections, so the other 3 fail writes instantly under contention. This is the same family as the row-618 lost close. | P1 | [A] #230 + #233 | **FIXING** — DS-107 `fix/sys-robustness`, done and in final tests |
| 4 | A failed save of a decision record is logged at INFO and ignored at 12 call sites. | P1 | [A] | **FIXING** — DS-107 |
| 5 | Armed-order state writes are discarded silently (`_ =`) at about 20 sites. | P1 | [A] | **FIXING** — DS-107 (now logged, counted, and the slot fails closed) |
| 6 | A panic in one trader kills the **whole process**, because there is no recovery. | P1 | [A] | **FIXING** — DS-107 (recovered, and that trader goes safe) |
| 7 | The NT8 AddOn "right build" proof survives a reconnect, so a downgraded AddOn keeps the old proof. | P1 | [A] | **FIXING** — DS-103 `fix/farside-proof-reconnect` |
| 8 | The update-activation tool (v7) cannot install our flat folder layout. | P1 (tooling) | [A] | **FIXING** — DS-105 |
| 9 | Plan deaths: dead at publish (3), broken plan JSON (3), obstacle-list omission (2), identity/tape/confirm-shape (1 each). | P1 (your main problem) | [A] #238 + #242 | **FIXING** — DS-101 `fix/planner-killers` |
| 10 | Any logged-in user can make the bot write crypto-wallet keys into `.env` (onboarding route). | P2 security | [A] CTO read `api/server.go:196` | **FIXING** — DS-106 `fix/sec-hardening` |
| 11 | Logout does not survive a restart; `/login` has no rate limit; Telegram first `/start` binds any chat; the JWT boot line prints "configured" even when it is the default. | P2 security | [A] | **FIXING** — DS-106 |
| 12 | 3 setting conflicts: `sessions_enabled` list vs per-session enable; `FAST_MARKET_REASONING` (4 duplicate lines in .env); the stop-entry comment vs its value. | P2 | [A] | **FIXING** — DS-105, one source each |
| 13 | 13 overridden settings (day_plan wake/scenario/realign/structure…). | P2 | [A] | **FIXING** — DS-105: wire it, or take it off the editable surface |
| 14 | The saved prompt tells the AI that breakeven/trailing are SUSPENDED; they have been active since 09-15. | P2 | [A] | **FIXED on branch** — DS-105, prompt now reads the live posture |
| 15 | Order queue has no cap; the Picture consumer never stops; timer and goroutine leaks. | P2 | [A] | **FIXING** — DS-108 `fix/leaks-lifecycle` (all 5 done, final tests) |
| 16 | An unresolved trade's P&L shows blank instead of "unresolved". | P2 | [A] | **FIXING** — DS-108 |
| 17 | An unset `NT_TRANSPORT` silently falls back to the old CSV path. Live has it set. | P2 | [A] | **FIXING** — DS-108 |
| 18 | `/api/health` says OK even with a dead DB or dead NT8; `research.db` is not backed up; guardrail holds are not counted; several counters are write-only; the scheduler misses its minute. | P2 | [A] | **FIXING** — DS-104 `fix/ops-observability` |
| 19 | No data or log retention (logs 39 GB, about 1 GB/day). | P2 | [A] | **Your ruling: keep everything.** Retention is wired but OFF (0 = forever); trades, fills and plans can never be pruned |
| 20 | 261 dead functions (zero references). | P2 | [A] #241 | **LISTED** — removed in a later wave, after these fixes merge (to avoid conflicts) |
| 21 | 110 labels or guide lines that disagree with the code. | P2 | [A] census | After the dead/conflict work (DS-105) |
| 22 | A strategy row with an **empty id `""`** and `is_default=1`; two "MNQ SIM Default" rows both marked default. Your live strategy ("MNQ", `a5b7662e`) is not affected. | NOTE (data) | [A] CTO read | **Your decision** (§7) |

---

## 4. Claims we checked and threw out (do not act on these)

| Claim | Truth |
|---|---|
| "Stop-entry orders are ON against your 09-05 OFF ruling." | **Wrong.** You ruled them back ON on **2026-09-18 at 07:52 CT**, and it is recorded in `guards.ts`. All three 09-05 hazards are closed, and 22,252 order snapshots show zero stacking (#237). Only the `.env` comment is stale; replacement text is in §7. |
| "393 plans killed by the invalidation-wording rule." | **Wrong.** Those were warnings on plans that were **accepted**, from before the 09-23 fix, and each was logged 4 times (once per scenario). Since the fix, the live counter reads **0** grammar refusals. |
| "Trading pipeline: zero issues" (the first audit, 22 minutes). | **Incomplete.** The cross-check found the duplicate-entry path, which was then proven (#236). |
| "`NOFX_TIMEZONE` / `CLAW402_DEFAULT_MODEL` do something." | They are read by nothing. They are harmless and are being removed (DS-105). |

---

## 5. Fix waves — status at 01:00 CT Saturday

Every fix must meet all of the following before the CTO merges it:
- a test that FAILS without the fix;
- a mutation proof (break the fix on purpose and the test catches it);
- the full race test, tsc, vitest and CI;
- the new PR sections: **Rules compliance**, **Error traceability** and **Dead/conflict ledger**.

A fix that only changes a label or a comment, or only adds a display, does **not** count.

| Lane | Branch | What | Status |
|---|---|---|---|
| DS-102 | `fix/double-entry-replay` | Duplicate-entry guard (Go, always on) + AddOn duplicate check | Final race run |
| DS-107 | `fix/sys-robustness` | DB lock wait on all connections, loud writes, panic recovery | All 6 done; final race run |
| DS-106 | `fix/sec-hardening` | onboarding .env, logout, login limit, Telegram bind code, JWT line | All 5 done; final race run |
| DS-108 | `fix/leaks-lifecycle` | queue cap, stop paths, leaks, transport warning, "unresolved" P&L | All 5 done; final race run |
| DS-101 | `fix/planner-killers` | born-dead salvage, JSON repair hint, obstacle-list hint, confirm-shape hint | Done; race continuing. Zero-call re-judge: **3 of 14 dead reads now publish**, and 6 more get a targeted repair |
| DS-105 | `fix/knobs-truth` | 16 dead + 3 conflicts + 13 overridden; activation layout | In progress (risk controls first) |
| DS-104 | `fix/ops-observability` | health, backups, counters, scheduler, retention (off) | Started 00:50 |
| DS-103 | `fix/farside-proof-reconnect` | AddOn proof cleared on reconnect | Started 00:47; rebases after DS-102 |

---

## 6. Research results

**Flash vs Pro (#231, 14 real planner calls, $1.29)**
- Flash at max was **not faster**: a median of about 288 s against Pro's 292 s. Flash writes about twice as fast, but it writes twice as much.
- Flash hit the 65k output cap and produced **no plan** on 3 of 5 prompts.
- With a 131k cap, both Flash plans were **dead at publish**.
- **Decision: keep Pro, and keep max.**

**Why plans die (#238 + #242, since the 09-23 fix)**
- 80 rejected attempts across 36 reads, of which **14 reads died**. Row ids are in the reports.
- Top killers:

  | Killer | Kills |
  |---|---|
  | Dead at publish (staleness) | 3 |
  | Broken plan JSON | 3 (ids 357-359, 391-393, 394-396) |
  | Obstacle list | 2 (ids 349-351, 385-387) |
  | Identity unresolved | 1 (ids 352-354) |
  | Tape verify | 1 (ids 360-362) |
  | Confirm shape | 1 (ids 412-414) |
  | Other (see #238/#242) | 3 |

- A killed read costs 7.5 min on average (17.3 max), followed by a 30-minute hold.
- The fix (DS-101) salvages a plan when one setup is dead, gives the AI a precise repair hint for each failure type, and keeps reasoning at max.

**Partner repo (#239)**
- 171 of 173 files are byte-identical and 2 are excluded on purpose.
- The secret scan shows nothing new, and the clean build is green.
- **Build at commit `66a0111c`, not the branch tip `3ab8652`.** The tip carries the RELEASE stamp, so a build there fails the startup check and refuses to trade. The runbook is fixed in #240.

---

## 7. Your decisions

**Done:**
- **Row 618** (yes): fixed.
- **Sessions** (all ON): you set them.
- **Retention** (keep everything): OFF.

**Still open (none of them block the boot):**
1. **The empty-id strategy row** (`id=""`, "MNQ SIM Default", `is_default=1`, `is_active=1`) and the second "MNQ SIM Default" (`578ac8f6`, `is_default=1`). My recommendation: delete the empty-id row, which is broken data, and keep `578ac8f6` as the only default. Your live "MNQ" strategy is untouched. This is a guarded DB write and needs your "yes".
2. **The stop-entry comment in `.env`** (lines 40–42). You edit `.env`. Replace the stale comment with:
   `# STOP_ENTRY_SEAM: 2026-09-05 OFF (unconfirmed cancels) → 2026-09-18 07:52 CT ON by owner ruling (cancel-confirm wave + stop-price frames proven). See web/src/guide/content/guards.ts.`
3. **The duplicate `FAST_MARKET_REASONING=max` lines** in `.env` (4 of them). They are harmless, but keep one.

---

## 8. Boot plan (this weekend, before Sunday 17:00 CT)

1. The CTO merges each fix PR in order after its gate: duplicate entry → system → security → leaks → planner → settings → operations → AddOn proof. After each merge, the full suite is re-run on the merged code.
2. Clean-clone release build, dry run, then the **attended** cutover. You run the command; the CTO verifies the boot and pushes the marker; you release the lock.
3. **You recompile the NT8 AddOn (F5).** The duplicate-entry fix changes `VLTraderTCPClient.cs`. The Go-side guard protects you even before F5.
4. Then the partner repo is re-synced and **the partner machines are updated at commit-of-build, not the tip**.

**Watch at Sunday's open (17:00 CT):**
- the first planner read shows max reasoning;
- the new boot lines are present: retention off, attempted-entry guard on, transport, JWT;
- no ERROR lines.

---

## 9. Index

| PR | What |
|---|---|
| #228 / #229 / #230 | audits: settings / trading / system |
| #232 / #235 / #233 / #234 / #239 | cross-checks: trading exec / trading plan / system / settings / partner |
| #236 / #237 | duplicate entry (proven) / stop-entry (safe) |
| #231 / #238 / #242 | Flash vs Pro / plan-death count / its cross-check |
| #240 / #241 | partner runbook fix / dead-code list |
| Fix branches | see §5 |

All audit and verification PRs are **drafts, not for merge**; they are the record. Fix PRs are opened against `dev` and merged only by the CTO after the gate.
