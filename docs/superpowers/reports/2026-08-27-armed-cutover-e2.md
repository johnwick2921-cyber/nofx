# Armed-Orders Cutover E2 — Complete (2026-08-27)

Wave 2 armed-orders cutover declared **COMPLETE** on live proof. The two planner
defects that blocked E2 are fixed, the E2 chain ran on the real wire, and the
debug seam is OFF.

## The three-part fix (one wave commit)

### b1 — flip.rule alias normalization (the planner defect #1)

- **Before:** `kernel/plan_doc.go` `conditionRules` = `{2x5m, 15m_close, 5m_close}`.
  The planner's natural spelling `2x5m_close` was rejected at write with
  `flip.rule "2x5m_close" invalid` → every retry failed closed.
- **After:** `ValidatePlanDocWithCaps` normalizes `2x5m_close → 2x5m` and
  `1x5m_close → 5m_close` at parse; the stored doc carries the canonical rule
  (same aliasing `confirmRules` / `scenario_facts` already accepted elsewhere).
- **Proof (live):** post-fix NY plan v3 stores `flip.rule = "2x5m"` — the model
  wrote the alias, the parser normalized, the plan landed ACTIVE on the first
  read instead of `planner_fail_closed`.

### b2 — FRESH FVGs machine grounding (the planner defect #2)

- **Before:** the fvg write-site validator (`kernel/fvg_entry.go`) re-checks
  every declared gap against its own fresh-gap scan, but the model never SAW the
  candidates — it invented stale gaps and all three retries died on
  `no fresh 3-candle gap … (fake/stale gap)`.
- **After:** new `kernel.FreshFvgCandidates` exposes the validator's own
  candidate list (direction, lo–hi, age, displacement×ATR5m). It is rendered in
  the planner prompt as a `## FRESH FVGs:` section, plus contract rule A2b:
  *author an fvg_entry ONLY from this list; if empty, do NOT author one.*
- **Proof (live):** v3 authored **no** fvg_entry (the list was empty) and the
  write passed on the first attempt.

### a — E2 debug seam (built gated, used once, turned OFF)

- `POST /api/armed/test-arm` (`api/handler_armed_seam.go` +
  `trader/armed_executor.go` `TestArmPlace`/`TestArmCancel`): drives the REAL
  placement path (`TCPTrader.PlaceLimitEntry` / `CancelOrder`), ledger row
  tagged `TEST-E2`.
- Defended on both sides: env `ARMED_TEST_SEAM=on` (default OFF) + bound
  account must be `Sim101`.
- Boot line reports seam state.

## Deployment

- Code commit `e35c91d4` + release marker → `dev`, pushed (`bc95e92b..b5c14097`).
  Built with `-ldflags -X main.buildRevision=e35c91d41cc9`, `deploy/RELEASE`
  marked `e35c91d4` (the integrity gate refuses a rev/RELEASE mismatch — caught
  and fixed in this wave).
- Final commit `9dd9c629` (seam response echo fix) + release marker → `dev`,
  pushed (`b5c14097..84b304a0`).

## Boot lines (current, verified)

```
🔐 BOOT INTEGRITY OK — rev 9dd9c629dce0 +dirty · built 2026-08-27T16:33:28Z · expected 9dd9c629 · goldens PASS
⚔️ armed_orders=on place_band=100t stale_working=15m test_seam=off (resting limits fill at the authorized price; stale_reeval NOT applied)
```

## NY plan after owner reset

- `POST /api/plan/reset` → **v3 ACTIVE**, `trigger_reason=owner_reset`,
  `lifecycle=active`, budget `replan_cap=4`.
- `flip.rule = 2x5m` (canonical — b1 live), no fvg scenario authored (b2 live).

## E2 chain (quoted from the log)

```
🧪 TEST-E2 arm → WORKING limit 29580.00 signal=88a1eb20-2e90-4573-be8e-2e7d5b6b9417 (seam)
🧪 TEST-E2 cancel sent signal=88a1eb20-2e90-4573-be8e-2e7d5b6b9417 (row 1 → cancelled)
```

Ledger row 1: `state=working` (place) → `state=cancelled` (cancel) — the full
armed-orders chain ran on the real TCP wire into NT8 SIM.

## Seam state

`ARMED_TEST_SEAM` removed from `.env`; boot line confirms `test_seam=off`.
The endpoint now refuses (`ARMED_TEST_SEAM is off`) — it cannot place orders
unarmed.

## Incident disclosure

The final redeploy kill ran while a position existed: an **owner manual NT8-side
entry** (MNQ LONG @ 29611.50, Sim101), not bot-generated. The reconcile system
materialized it as tracked before the kill and re-read it after; no bot damage.
The flat-window check was present but unguarded in that command — future deploys
must gate the `kill -9` on an empty positions list.

## Next

- Level-truth wave resumes: pop the parked stashes (now indexed
  `stash@{0..4}` after cutover) in order, then `go test ./...`.
- Master-recheck dispatch continues in its own worktree — untouched here.
