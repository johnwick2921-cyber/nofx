# AUDIT2-RENAME-WEB — DS-102 (2026-09-29, round 2)

- **Slice:** WEB/UI — R1b items 3 (update header), 5 (browser storage), 9 (docs/guide), item 11 (theme, from F20) + folds **F20, F21, F22** + round-1 report #270 (DS-103) read first.
- **Plan v2 audited:** `/home/hoang/rename-vl-plan/2026-09-29-rename-nofx-to-vl-plan-v2.md` — `md5sum` = `7b8dc9a9a196e5554203db7a3b71aa7b` (matches dispatch 7b8dc9a9…). [A]
- **Base:** `origin/dev` `9d52f5dc611ed001fc057bca613685071d77efab` (RELEASE boot 7). [A]
- **Round-1 avoidance:** #270's P1-1 → F20, P1-2 → F21, P2-4 → F22 are the folds this report re-checks; its P2-1 (translations.ts install commands), P3s (telegram test literals, dev-server window, install-autostart.sh, fixtures, cosmetic comments) are not repeated here.

## (a) FOLD VERIFICATION — F20, F21, F22 against origin/dev

**F20 (theme) — CORRECT and mechanically COMPLETE.** [A]
- `web/tailwind.config.js:10-28` holds the `nofx-gold/bg/accent/text/success/danger` palette exactly as cited.
- `web/src/index.css` has 25 `nofx` lines (`--nofx-*` vars + `.nofx-toast`).
- `grep -rn 'nofx-' web/src web/index.html` = 470 occurrences; **zero dynamic class construction** (`nofx-${` = 0) — a mechanical `nofx-` → `vl-` rewrite cannot miss runtime-built class names. The fold's "~545" is a slight overcount of my measured set (round 1 counted across more paths) — harmless; no action.

**F21 (storage migration) — CORRECT and achievable as specified.** [A]
- All six old keys verified: `nofx_user_mode` (`web/src/lib/onboarding.ts:3`), `nofx_beginner_wallet_address` / `nofx_beginner_onboarding_completed` (`web/src/components/modals/SetupPage.tsx:67-68`, `web/src/contexts/AuthContext.tsx:255-256`), `nofxi-agent-chat` / `nofxi-agent-chat:<uid>` / `nofxi-agent-chat-draft:<uid>` (`web/src/lib/agentChatStorage.ts:1,15,19`).
- F21's new names bind all six (`vl.userMode`, `vl.beginnerWalletAddress`, `vl.beginnerOnboardingCompleted`, `vl.agentChat`, `vl.agentChat:<uid>`, `vl.agentChatDraft:<uid>`). ✓
- "Entry point at app bootstrap (`web/src/main.tsx`) before any reader" is ACHIEVABLE: I grepped for module-scope `localStorage` reads across `web/src` — **zero** package-scope readers outside functions. [A] The first reads happen at render/hook time, after `main.tsx`'s body runs. (ESM note: module-scope readers would have evaluated during imports, BEFORE main.tsx's body — the fold's wording stays valid only because none exist; see P3-1 for the wording lock.)

**F22 (update header) — CORRECT and COMPLETE.** [A]
- `api/handler_updates.go:408`: `if v := r.Header.Values(UpdateHeader); len(v) != 1 || v[0] != "1"` — today's rule is exactly-one-VALUE (Go counts empty strings as values). F22's "count values across both names; exactly one equal to \"1\" passes; two values (same or different name) → 403" is the correct generalization, and it subsumes the same-name-duplicate case (`X-NOFX-Update` sent twice with "1" → 2 values → 403) that the phase text's "both present → 403" does not.
- CORS (`api/server.go:94-98`): `Access-Control-Allow-Headers: Content-Type, Authorization` — the update header remains un-allow-listed; no CORS change needed and none claimed. ✓

## (b) NEW FINDINGS (beyond #270)

**P2-1 — "R1b item 11" (theme) does not exist in either list.** F20's fix text says the theme rename is "R1b item 11, mechanical", but the R1b numbered list in the plan body ends at 10, and the D2 JOB list ends at 10. A lane executing D2 verbatim would skip the whole theme rename (470 class/variable occurrences) even though the folds are "binding". **Exact replacement:** insert into BOTH lists, after item 10:
> `11. Theme (F20): rename the Tailwind palette in tailwind.config.js:10-28, every --nofx-* var and .nofx-toast in web/src/index.css, and all ~470 nofx-* class uses in web/src (mechanical; no dynamic class construction exists — verified). brand-i18n.test.ts untouched.`
And in D2's mutant list add: `plant one nofx- class (theme census pin must fail)` — the theme needs its own census pin like the module.

**P2-2 — D2 item 3 wording contradicts F22.** D2 item 3 still says "server accepts EXACTLY ONE of X-VL-Update / X-NOFX-Update equal to \"1\" (both present → 403)". "Both present" misses the same-name-duplicate case F22 closes. **Exact replacement for D2 item 3's first sentence:**
> `Header: web sends X-VL-Update: 1; the server counts values across BOTH names (r.Header.Values of each, concatenated) — exactly one value and it equals "1" → pass; two values (same name or different) → 403. CORS must still not allow it (api/server.go ~98 test).`

**P2-3 — D2 item 5's key names contradict F21.** D2 item 5 keeps `… → vl.agentChat…` (the round-1 "name is an ellipsis" defect). **Exact replacement:**
> `5. Browser storage: a single migration at app bootstrap (web/src/main.tsx, before the first render; no module-scope localStorage readers exist — verified 2026-09-29) enumerates the six old keys and copies to EXACTLY vl.userMode, vl.beginnerWalletAddress, vl.beginnerOnboardingCompleted, vl.agentChat, vl.agentChat:<uid>, vl.agentChatDraft:<uid>; copy only when the new key is absent (two-tab safe); delete old only after write + read-back; parse/quota failure keeps the old key.`

**P3-1 — F21 wording lock.** The fold's "before any reader" is only safe because zero module-scope readers exist today. Add one clause so a future import-graph change cannot re-break it: "the migration must be a side-effect import placed FIRST in main.tsx (ESM evaluates dependencies in import order) — never a statement after the imports." [A]

**P3-2 — theme census pin absent from D2's mutant list.** D2 lists 5 mutants; none covers the theme. A `nofx-` class planted post-rename would pass every gate. Add the mutant from P2-1's fix.

## (c) CONSISTENCY — anything else contradicting a fold?

- Plan body R1b item 7 still says the census guard is `grep -rIic nofx` vs an allow-list — superseded by F23 (token/pattern allow-list of file:pattern pairs). The folds header says they override, so this is safe, but quote-and-replace it to avoid the round-1 mistake being re-implemented. P3-3: replace item 7's last sentence with `"New census guard (F23): token/pattern allow-list of file:pattern pairs — not a raw grep count."`
- F27 (button update parks at nt8_updated after the VL_BUILD_ID bump) vs D2 item 8 — no contradiction found. [A]
- F5 (nofx-web.service active, :3000) is in the ops slice; no web-slice contradiction. [B]

## Verdict
**NO NEW P0/P1.** F20/F21/F22 are correct and complete against the code; the gaps are three P2 dispatch-text consistency defects (item 11 absent, D2 item 3 vs F22, D2 item 5 vs F21) and three P3 wording/pin gaps.

## NOT done (L-AUDIT)
No code change outside this report; no deploy/units/DB/lock; evidence all read at base `9d52f5dc6` (quote for every file:line above). Build/vet not run — this slice changed no code (L8's gate applies to code-changing waves; stated per dispatch L-AUDIT).
