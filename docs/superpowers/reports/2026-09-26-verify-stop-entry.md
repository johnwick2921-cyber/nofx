# VERIFY-0926 — STOP-ENTRY SEAM (DS-106, second pair of eyes)

Follow-up to the knobs cross-check (`2026-09-26-verify-knobs.md`, which CONFIRMED P1-1 as a contradiction of the 2026-09-05 ruling). CTO's question: `.env:40-43` says the seam "Stays off until a cancel-confirmation wave lands", then sets `STOP_ENTRY_SEAM=on`; boot-3's cancel path reads `confirm=broker-snapshot` — so has the precondition been met, and is the seam ON by ruling or by drift?

READ-ONLY (L17). Base `04ae1c2f`, worktree `/home/hoang/nofx-ds-106-verify0926` @ branch `verify/0926-stop-entry`, claim `6526bc44`. Logs `/home/hoang/nofx/data/nofx_2026-09-{18..25}.log` read-only. No code changes, no tests.

## 1. Answer in one paragraph

The seam is ON by a **second owner ruling, 2026-09-18 07:52 CT** — not by drift. The ruling is recorded in the stop-entry guide clause (`web/src/guide/content/guards.ts`, merged in `99349544`, 2026-09-18 13:23:23 CT): *"Two owner rulings set it: 2026-09-05 OFF — no stop entry is placed at all until a wave lands that confirms a broker cancel instead of assuming one — and 2026-09-18 07:52 CT ON, once the far side proved its stop-price frames."* [A]. At the flip-moment boot both 09-05 preconditions held on the wire [A]: the cancel-confirmation wave was live (`cancels: confirm=broker-snapshot · slot-guard=on(refuse-on-live|stale)`) and the far side was proven (`addon build_id=2026-09-07-h1 expected=2026-09-07-h1 match=yes`). The residual problem is **stale prose, not a live-config contradiction**: the `.env` comment predates the second ruling, and my knobs addendum's "no later re-enable ruling exists in docs/" was wrong in scope — the ruling lives in the guide, not in `docs/` reports.

## 2. The two rulings

| date | ruling | recorded where |
|---|---|---|
| 2026-09-05 | OFF — "Until a cancel-confirmation wave lands, the AddOn build floor is not a brake the owner will rely on." Reason: `nt.CancelOrder` reports success on a SEND, not on a confirmation; broker once held **nine** working stops for one slot (snapshot id 1664) on a 2-contract cap. | `docs/superpowers/reports/2026-09-05-wave-b-stop-entry.md:5,247-249,541`; `.env:40-42` comment |
| 2026-09-18 07:52 CT | ON — "once the far side proved its stop-price frames" | `web/src/guide/content/guards.ts` (merged `99349544`) [A] |

## 3. Precondition ledger at the flip moment (09-18 07:52:20 boot, `nofx` rev 249ff3a5eea6)

- **Cancel-confirmation wave: LANDED.** Merged to dev 2026-09-06 00:24–06:48 CT (`871083cc` adjudicator → `22af989e` settlement/reconciliation/boot line → `f61b23a3` per-slot invariant → `ea3f33fe`), first-parent of `04ae1c2f` [A]. Rule A20: *"a cancel is proven by the ORDER'S ABSENCE FROM A FRESH BROKER SNAPSHOT"* (`trader/cancel_confirm.go:21-24`). Live at the flip boot: `🧾 cancels: confirm=broker-snapshot · pending=0 · unconfirmed=0 · slot-guard=on(refuse-on-live|stale) · timeout=1m30s · stale-bound=1m0s · rerequest-cap=5` (log line 23499) [A]. The slot-guard (`refuse-on-live|stale`) is the exact broker-side-stacking brake the 09-05 ruling named.
- **Far-side proof: HELD.** `🎯 stop-entry: seam=on · slots=stop_price · guard=stop-side · unknown=no-op · addon build_id=2026-09-07-h1 expected=2026-09-07-h1 match=yes` (log line 23563) [A]. `PlaceStopEntry` refuses any build below `MinAddonBuildStopSlot` — the build floor the ruling required before relying on the seam.
- Same-boot entry-law line: `stop_entry_seam=ON` (23506); arms line `stop-entry=on(reclaim)` (23503) [A].

## 4. Seam-state timeline (log-verified)

| boot (CT) | seam | build / match | confirm line |
|---|---|---|---|
| 09-18 00:09:58 | OFF (09-05 literal) | 2026-09-07-h1 **match=yes** | broker-snapshot ✓ |
| 09-18 01:32:39 | OFF | match=yes | — |
| 09-18 02:25:22 | OFF | match=yes | — |
| **09-18 07:52:20** | **ON** ← second ruling | match=yes | broker-snapshot ✓ |
| 09-18 13:30:41 | ON | match=yes | ✓ |
| 09-19 01:55:04 | ON | match=yes | ✓ |
| 09-22 00:48:11 | ON | 2026-09-20-p1 match=yes | ✓ |
| 09-23 00:46:33 | ON | build=2026-09-20-p1 expected=2026-09-22-m2 **match=NO** | ✓ |
| 09-24 10:19:48 | ON | 2026-09-23-m21 match=yes | ✓ |
| 09-25 01:40:13 | ON | build_id=none → `slots=unproven(addon build)` | ✓ |
| 09-25 08:21:42 | ON | 2026-09-23-m21 match=yes | ✓ |
| 09-25 20:14:26 (boot 3) | ON | — (`stop-entry=on(reclaim)`, entry law ON) | ✓ |

Notable: on 09-18 00:09–02:25 both preconditions **already held** (match=yes, broker-snapshot live) yet the seam stayed OFF until the 07:52 ruling — the flip was a decision, not an accident. On 09-23 the seam was ON with `match=NO` (expected constant bumped ahead of the AddOn): placement still refused below the build floor (`MinAddonBuildStopSlot`), and the next day's build closed the gap — a transient, floor-protected, not a violation.

## 5. Verdicts

- **P1-1 as CONFIRMED in the knobs addendum: REFUTED here (corrected).** The seam ON is not a contradiction of the 09-05 ruling; the 09-05 ruling was superseded by the 2026-09-18 07:52 CT owner ruling, which the guide clause records. My knobs addendum statement "No later re-enable ruling exists anywhere in docs/" was over-narrow: I grepped `docs/` reports only; the ruling lives in `web/src/guide/content/guards.ts`.
- **NEW P2 (docs staleness):** `.env:40-42` still reads "…Stays off until a cancel-confirmation wave lands…" directly above `STOP_ENTRY_SEAM=on` (`.env:43`). After the second ruling the comment is wrong and actively misleads — it should cite both rulings (09-05 OFF → 09-18 07:52 ON). The 09-05 report remains the only report-level ruling text; recommend a one-line ruling note in a report or the comment fix. Not a runtime defect — the binary reads the env, not the comment.
- **What the CTO observed, verified:** boot-3 cancel path `confirm=broker-snapshot` — true since 09-18 00:09:58 at least [A]; the wave is live and the seam's remaining precondition has held since the flip.

## 6. What I did NOT verify

The second ruling's out-of-band provenance (the owner's 07:52 instruction channel — bridge mail older than 09-25 is not retained; the guide clause is the in-repo record). Whether the 09-23 `match=NO` day placed any stop entry (the floor would have refused; I did not replay that day's placements). No code, tests, or checklist changes (L17).
