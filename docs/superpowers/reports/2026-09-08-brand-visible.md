# Dispatch 102 — visible brand preflight (not shipped)

Branch `fix/brand-visible`; session `brand-visible-0b955fbc/root[unlisted]`; isolated worktree `/tmp/nofx-brand-visible`; accepted from dev `954f11b15f2e7615678f7d2b708c47895faebf1e`.

**STOP before production edits.** D2 requires the boot banner to read the same product-name source as the UI. Section B forbids changing “the Go module path or any import.” A shared Go/UI source needs import glue. The owner has been asked whether adding only display-name imports is allowed while every existing module/import path stays unchanged. No answer has been received. This is a dispatch conflict, not a skill or automatic-review rejection.

This report records preflight evidence only. No production strings, identifiers, runtime files, dist, RELEASE, Guide stamp, DB, or held rebrand branch have been changed. No swap or kill was attempted. This report is initially on the named branch, **not yet on dev**; it is not a completed implementation or cutover report.

## Running source and ownership

[A] `/api/health` returned `revision=6f677b55daa1`, `status=ok`. Independently, `go version -m /proc/3726840/exe` read `vcs.revision=6f677b55daa1c7da33b8c35f8bcc67883f36b470`, `vcs.modified=false`.

[A] At acceptance the main tree was clean on dev; Stage A held a fresh deploy lock for its archive-path repair. That lane owns `954f11b1`; this lane did not author its repair, take its lock, or change its worktree. The earlier archive request is being handled there. The held `fix/rebrand-phase-1-2` branch was not modified, merged, rebased, or mined for implementation.

## Census corrections and bounded scope

The basis is `docs/superpowers/reports/2026-09-03-rebrand-census.md` at `878f9e7f`. Its Group 1 table has **11 numbered rows**, although its summary says **17 locations / approximately 35 direct string hits**. Those are not a comparable denominator for all raw grep matches. The full per-file visible census is still in progress; no invented final count is claimed.

[A] Confirmed at the measured running source:

| Source | Current visible text | Intended canonical form |
| --- | --- | --- |
| `main.go:45` | `🚀 NOFX - AI-Powered Trading System` | `🚀 VL Intelligent - AI-Powered Trading System` |
| `agent/i18n.go:26` | `📊 *NOFXi Status*` | `📊 *VL Status*` |
| `agent/i18n.go:25` | `📊 *NOFXi 状态*` | `📊 *VL 状态*` |
| `web/src/components/agent/ChatMessages.tsx:133` | `NOFXi · <time>` | `VL · <time>` |
| `web/src/components/agent/ChatInput.tsx:124` | `Ask NOFXi anything...  ⌘K` | `Ask VL anything...  ⌘K` |
| `web/src/components/agent/ChatInput.tsx:123` | `跟 NOFXi 聊点什么...  ⌘K` | `跟 VL 聊点什么...  ⌘K` |
| `web/src/components/agent/ChatInput.tsx:187` | `NOFXi may make mistakes. Always verify trading decisions.` | Short form `VL`; remaining text unchanged |
| `web/src/components/agent/WelcomeScreen.tsx:118` | `跟 NOFXi 聊点什么` | `跟 VL 聊点什么` |
| `web/index.html:8` | `VL Trader - AI Trading System` | `VL Intelligent - AI Trading System` |
| `web/src/pages/TraderDashboardPage.tsx:461` | `VL Trader · …` | Full product form `VL Intelligent` |
| `web/src/guide/GuidePage.tsx:487` | `NOFX System Guide` | `VL Intelligent System Guide` |
| `web/src/guide/content/welcome.ts:14` | `NOFX / VL` | `VL Intelligent` |
| `README.md:1` | `NOFX` heading | `VL Intelligent` |

The status card is server-authored text from `Agent.handleStatus` (`agent/agent.go:878`), formatted from `agent/i18n.go`, then rendered by `ChatMessages` / `MessageRenderer`; it is not a separate frontend status-title literal. The test exercises those production call sites.

[A] Raw running-source checks reproduce `docs/i18n/` **379** case-insensitive matches in 18 files. `.github/SECURITY.md` has **13** raw matches. `web/src/index.css` has **26** raw matches, compared with the census's 25 CSS matches. These include paths, URLs, handles and invisible identifiers; they are not counts of authorized replacements. Section B controls over the older census: CSS identifiers, security-policy prose outside Section B, external handles/URLs and the real `nofx-bin` example are not a blanket rename scope.

[A] Additional current sources absent from the small census include `agent/i18n.go`, daily-report text in `agent/scheduler.go:75`, onboarding greeting text in `agent/onboard.go:531,535`, the Guide heading and dashboard label above. `agent/onboard.go:412` writes a trader name with a `NOFXi-` prefix; that stored-name producer is excluded by Section B, even though greeting text in the same file is in scope.

## Live surface evidence and limits

[A] A real browser navigation to `http://localhost:8080` reached `/login`; its title was `VL Trader - AI Trading System`. A full authenticated read-only surface walk is **NOT YET PROVEN**. The standalone browser could not start because its local Chromium could not load `libnspr4.so`. No fixture output is presented as a live authenticated UI observation.

[A] The eight rendered pins below cover the status card, sender label, input and disclaimer in EN/ZH/ID, the Vite-transformed page title and the real Guide component. Indonesian status currently uses the existing English fallback; the test does not add or invent an Indonesian translation.

A15 after-boot inventory is pending because there has been no boot. Known retained technical names include `nofx-bin` in the Guide's architecture example, nofx paths/commands and GitHub URLs in technical help/README, and external NofxOS provider labels. Historical user messages and stored trader names are not rewritten. Held identifier work must be named from its existing branch provenance, not attributed to this lane.

## RED evidence, then scope baseline

[A] `npm --prefix web test -- src/brand-visible.test.tsx`: **8 failed / 8** on unmodified production source. The initial title harness failed on a Node/jsdom environment mismatch; that result was discarded. After moving Vite's transformation into its normal Node environment, all eight failures concern actual old names:

- EN/ID status: `expected '⚡📊 NOFXi Status…' to contain 'VL Status'`.
- ZH status: `expected '⚡📊 NOFXi 状态…' to contain 'VL 状态'`.
- EN/ID placeholder: `expected 'Ask NOFXi anything...  ⌘K' to contain 'Ask VL anything'`.
- ZH placeholder: `expected '跟 NOFXi 聊点什么...  ⌘K' to contain '跟 VL 聊点什么'`.
- Title: `expected 'VL Trader - AI Trading System' to be 'VL Intelligent - AI Trading System'`.
- Guide: expected the `VL Intelligent System Guide` heading query not to be null.

The browser test obtains status text by executing `TestBrandStatusRenderFixture`, which calls the real `Agent.handleStatus` for each language. It then renders `ChatMessages` with that result. It does not hand-build the status message. No live chat request is sent.

[A] `npm --prefix web test -- src/brand-scope.test.ts`: **17 passed / 17**. Sixteen protected source files are byte-pinned, including units, binary ExecStart, JWT issuer/validator, lock/claim/backup scripts, module path, log naming, TCP strings, NT8 and browser chat-storage keys. The negative pin removes the exact text `&& token.Valid` in memory and asserts rejection by the scope checker. No protected production file was edited for the mutation. This is an E2 guard-removal check, **not** the unperformed E5 shared-constant mutation.

No GREEN rename result is claimed. E3 final language counts, E4 new-constant call-site grep, E5 constant mutation, full merged suite, clean-clone build, binary-derived Guide stamp, fresh five-leg gate, window/in-flight check and cutover proof all remain outstanding. The pending production source has not been built.

## C3 initial protected-identifier evidence

- `deploy/nofx.service:41` uses `ExecStart=__NOFX_DIR__/nofx-bin`; user units reference the current repo path in `deploy/systemd-user/nofx-backup.service:7`.
- `deploy/nofx-lock.sh:43` chooses the `nofx-main.lock.d` lock path.
- `auth/auth.go:93` mints `Issuer: "nofxAI"`. Correction: `ValidateJWT` at line 102 does not itself require a particular issuer; it checks signature/method and token validity. Thus an issuer-specific reader requirement is **not established from this function**, and no issuer/auth edit is made.
- `go.mod:1` declares `module nofx`, used by the existing Go imports.
- `logger/logger.go:90` constructs `nofx_%s.log`; a specific current reader glob has not yet been established.
- `provider/ninjatrader/tcp_server.go:1749` emits `Source: "nofx-go"`; `tcp_framing.go:114` documents the pair with `vltrader-addon`. No wire field or AddOn changes.
- `web/src/lib/agentChatStorage.ts` retains persisted `nofxi-agent-chat` and draft keys.

## Numbering and rollback

[A] Both checklist numbering formats were extracted, then `sort -n | uniq -c` was run. Highest occupied: **93**, not dispatch's 91. Duplicates: **75, 76, 77**, each count 2. Numbers 92 and 93 each count 1. No number was assigned; repeat at merge, never renumber another lane.

Rollback is not applicable yet: no runtime artifact was changed. Any eventual cutover must back up the binary actually running at that time, verify its embedded revision, preserve that revision in its rollback filename, then follow RELEASE → mv → VERIFY → owner kill. Prior cutover artifacts are not an authorization or fresh gate for this wave.

## Spec and source freshness at acceptance

The following are exact `git log -1 --format='%h %aI %s' -- <file>` results. Base is dev `954f11b1`. Re-read changed specifications after the next dev refresh before implementation.

- `docs/superpowers/reports/2026-09-03-rebrand-census.md` — `878f9e7f 2026-09-03T21:53:31-05:00 docs: rebrand census 2026-09-03 — every nofx identity mapped (read-only)`
- `docs/superpowers/SYSTEM-MAP.md` — `954f11b1 2026-09-08T18:38:16-05:00 fix(research): resolve archive paths before SQLite URI construction`
- `docs/superpowers/AUDIT-CHECKLIST.md` — `954f11b1 2026-09-08T18:38:16-05:00 fix(research): resolve archive paths before SQLite URI construction`
- `main.go` — `896aeea5 2026-09-08T15:55:50-05:00 feat(research): capture scorer evidence and stage append-only snapshot writers`
- `agent/i18n.go` — `c7cd5aae 2026-04-25T16:18:45+08:00 change v1`
- `agent/agent.go` — `6677b5b0 2026-08-22T18:24:37-05:00 feat(wave4): API auto max — per-model DeepSeek thinking knobs (4.5)`
- `agent/scheduler.go` — `c7cd5aae 2026-04-25T16:18:45+08:00 change v1`
- `agent/onboard.go` — `9f82938e 2026-06-11T11:32:24-05:00 feat(ai): deepseek-v4-pro is the system-wide DeepSeek default`
- `web/src/components/agent/ChatMessages.tsx` — `6288d061 2026-05-11T23:51:27+08:00 fix(web): fix UI bugs and unify design tokens`
- `web/src/components/agent/MessageRenderer.tsx` — `a7c31f5e 2026-04-21T23:47:55+08:00 feat: port NOFXi agent module onto latest dev base (#1485)`
- `web/src/components/agent/ChatInput.tsx` — `6288d061 2026-05-11T23:51:27+08:00 fix(web): fix UI bugs and unify design tokens`
- `web/src/components/agent/WelcomeScreen.tsx` — `258d3863 2026-05-27T17:12:54-05:00 fix(web): plan 4.7 — AgentChat tickers BTC/ETH/SOL -> MNQ`
- `web/index.html` — `68393cf0 2026-08-22T17:19:41-05:00 feat(wave2): clarity quick-wins — rev exposure, no-data honesty, venue badge, grid honesty, duplicate hint, approve button, favicon`
- `web/src/pages/TraderDashboardPage.tsx` — `83892c74 2026-09-06T16:23:53-05:00 fix(dashboard): show market and equity together without remounting Planner`
- `web/src/guide/GuidePage.tsx` — `09e110b5 2026-09-03T22:39:03-05:00 fix(guide): the drift banner could never have been right, plus the label rules`
- `web/src/guide/content/welcome.ts` — `c3405721 2026-09-03T23:01:08-05:00 docs(guide): a canon section — the files that ARE the answer, incl. clock-seams.list`
- `README.md` — `d453bc2e 2026-08-26T19:50:46-05:00 PLANNER CONTRACT WAVE: S/D+FVG playbook (A1-A5) + MPM look-ahead rule (B1) + README §9 UI fixes (C1-C9) (#81)`
- `web/src/index.css` — `1d015b89 2026-08-15T08:32:04-05:00 feat(dayplan): P4 FE foundation — tokens, plan API client, i18n, hooks`
- `.github/SECURITY.md` — `85794a72 2026-03-15T11:50:08+08:00 feat: add X-Client-ID header for claw402 monitoring`
- `deploy/nofx.service` — `b03debf4 2026-06-10T21:40:33-05:00 fix(deploy): autostart units go JOURNAL-ONLY — kills the 209/STDOUT loop for good`
- `deploy/systemd-user/nofx-backup.service` — `1b29263c 2026-08-13T17:56:29-05:00 feat(ops): C1 — twice-daily SQLite auto-backup user timer + RESTORE.md`
- `deploy/nofx-lock.sh` — `bd20be31 2026-09-03T22:04:13-05:00 feat(lock): reclaim on the record, and push-empty-at-accept (class 70)`
- `auth/auth.go` — `85794a72 2026-03-15T11:50:08+08:00 feat: add X-Client-ID header for claw402 monitoring`
- `go.mod` — `294d7a13 2026-08-29T22:47:49-05:00 security(f1): dependency vuln scan + safe bumps + CI automation + class 22`
- `logger/logger.go` — `85794a72 2026-03-15T11:50:08+08:00 feat: add X-Client-ID header for claw402 monitoring`
- `provider/ninjatrader/tcp_server.go` — `0babd090 2026-09-08T16:22:32-05:00 feat(research): link attempts, permissions, broker receipts and corrected outcomes`
- `provider/ninjatrader/tcp_framing.go` — `6262bf42 2026-09-07T23:34:58-05:00 fix: stamp entry creation time and await received placement truth`
- `web/src/lib/agentChatStorage.ts` — `73b9b1ad 2026-05-02T22:55:10+08:00 Improve NOFXi agent product handling`
