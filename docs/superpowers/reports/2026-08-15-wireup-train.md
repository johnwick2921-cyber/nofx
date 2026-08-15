# VL Day-Plan — wire-up train (fixes the audit's 10 dead wires)

**LINE 1 — W1+W2 DONE + DEPLOY-READY (URGENT, pre-Sun-17:00); W3–W10 in progress.**
Read-only audit `7344820d` → this train wires the dead wires end-to-end. Additive,
SIM-only, guardrails untouched, own commit per item.

---

## ✅ W1 — Sunday-read guard + digest window (`57b518d8`) [A]
The read scheduler used `timeReachedCT` (fires for ANY time ≥ ReadCT), so at
**Sunday 17:00 CT** (CME reopens = Monday's trade-date) the NY 08:25 read fired
early, building Monday's plan from stale Sunday-evening data; the real 08:25 read
was then deduped away.
- `inSessionReadWindow(now, ReadCT, WindowEnd)` — fire ONLY inside the session's
  own window `[ReadCT, WindowEnd)`, wrap-aware (ASIA 16:55→02:00). NY `[08:25,15:00)`
  → Sunday 17:00 is past it → NY never fires there. + `IsCMEOpen(now)` (holiday/
  weekend-aware) gate. Monday's plan is now built fresh at 08:25.
- Daily digest moved from `>=16:00` (inside the CME break → fired 17:00+ with the
  NEW day's empty P&L window; Friday never fired) to `[15:00,16:00)` (RTH-close→
  break) where the trade-date + P&L window are still the CLOSING day's. Reachable
  Mon–Fri.
- **Proven:** `TestW1SundayReadGuard` + a full **Sat+Sun sweep** (NY read false at
  every weekend step) + daily-roll window + ASIA wrap. `timeReachedCT` kept for the
  last-entry/eod-flat gates (correct all-day-after use).

## ✅ W2 — pin the EXACT planner model + stats reset (`0c8bae59`) [A]
Plans stamped the provider alias `deepseek`, not an exact string (§125 violation).
- `mcp.AIClient.ResolvedModel()` exposes the exact model; `DefaultModelForAlias`/
  `IsProviderAlias` map/detect aliases (`deepseek`→`deepseek-v4-pro`).
  `resolvePlannerClient` pins the exact id on all 3 return paths (prefer the
  client's model, else map the alias, else warn) → every plan row carries an exact
  model string.
- §128 no cross-model pooling: `maybeResetStatsOnModelChange` records the pinned
  model in system_config; on a change → `MatchedRandomStore.ResetWindow` (clears
  verdicts + weekly), logged.
- **Proven:** alias→exact mapping/detection + live `ResolvedModel` + `ResetWindow`
  clears both tables + the resolver now asserts an exact id. 26 pkgs green.

---

## ⏫ DEPLOY NOW (owner, before Sunday 17:00 CT — Go touched, rebuild+restart):
```bash
cd /home/hoang/nofx && git pull && go build -o nofx-bin ./... && echo BUILD OK
kill -9 $(pgrep -f nofx-bin)   # systemd Restart=on-failure respawns the new binary
```
(`sudo systemctl restart nofx` is classifier-blocked; SIGKILL is the deploy per
CLAUDE.md. Do it in the flat/CME-closed weekend window. No AddOn F5 — no `.cs`
changed.) This makes Monday's 08:25 read fire correctly + exact model pinned.

---

## ⏳ Remaining (W3–W10) — second rebuild before Mon 08:00
W3 calendar producer + red-news · W4 overlays→executor · W5 learning-loop on real
exits · W6 alert Emit call-sites · W7 level-state writer · W8 registry loading ·
W9 config readers · W10 regime feeds. Prompt-content changes → goldens regenerated
deliberately, diffs listed. Filled in as each lands.
