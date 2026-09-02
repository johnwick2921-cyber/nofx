# CLASS 38 — prompt/validator contract mismatch: the prompt offers what the validator refuses

Date: 2026-09-01 · Owner: hoang · Worktree `../nofx-class38` (branch `fix/class38-contract-mismatch`)
Evidence tiers: **[A]** directly verified · **[B]** inferred from strong evidence · **[C]** speculation.
All times CT (R8). Live rev at dispatch: `e42a0b43` (class 37, booted 21:19:49 CT).

## STATUS

| Item | State |
|---|---|
| Code | merged to dev — see the closeout for the sha |
| Build | clean clone `--no-local`, `vcs.modified=false` — quoted in §7 |
| Cutover | **NOT DONE — PARKED pending the owner's explicit GO (A3)** |
| Proof (A20) | **NOT YET OCCURRED.** The proving event is the next live read authored against the NEW prompt. Nothing to quote; see §8. |
| Lock (A2) | `~/nofx-main.lock` acquired 21:27 CT (no prior holder), released at closeout |
| Stop-lines | held: no validator logic, no enum, no retry-semantic, no normalization (F7 is class 39) |

---

## 1. The bug [A]

One ASIA read, three attempts, three DIFFERENT rejects — each one a place where the
prompt (or a hint it quotes) offers something the validator refuses.
Source: `planner_rejected_prompts` rows 78, 79, 80.

| Row | CT | Attempt | Reject (verbatim, truncated) |
|---|---|---|---|
| 78 | 17:47:32 | 1/3 full | `scenario[0].confirm2.rule "1m_mss" not allowed for breakdown_continue — entry law: 1 confirming close + displacement ≥ BD_MIN_DISP_ATR×ATR5m OR stop-entry (E7); 2x5m legal ONLY here` |
| 79 | 17:53:33 | 2/3 repair | `scenario[0].confirm2.rule "2x5m" invalid (touch\|1x5m_close\|2x5m_close\|1m_mss\|time_hold)` |
| 80 | 18:00:41 | 3/3 re-author | `arm legs on breakdown_continue — arm_legs_sweep_reclaim_only (the split entry is the sweep_reclaim contract; other conditions arm single)` |

**(a) → (b) is a closed loop.** The rejection the model reads quotes the entry-law
`Style` string verbatim (`entry_law.go:153`: `"… not allowed for %s — entry law: %s"`).
That Style said **"2x5m legal ONLY here"**. `confirm.rule`'s enum is
`touch|1x5m_close|2x5m_close|1m_mss|time_hold` — **"2x5m" is not in it**. It belongs
to the SEPARATE death/flip enum (`plan_doc.go:285` `conditionRules = {2x5m, 5m_close}`).
The model did exactly what the instruction said and was rejected for it. Its repair then
came back unparseable, and the journal recorded one bare sentence about that.

**(c) is an unstated rule.** The scenario schema line renders `"legs":[…]` on EVERY
scenario with no condition qualifier, while its siblings on the SAME line do carry one
(`fvg{} REQUIRED iff condition=="fvg_entry"`, `breakdown{} REQUIRED iff waterfall-class`).
The sweep_reclaim-only rule existed ONLY in `plan_doc.go:167`. Meanwhile the ARMED ORDERS
prose pushes breakdown_continue toward `arm{}`. Counts in the rendered prompt before this
wave [A]: `arm_legs` 0 · `split entry` 0 · `arm single` 0 · `EXACTLY 2 legs` 0 ·
`split contract` 0 · `ARM SPLIT` 0.

**Why the class-34 guard stayed green:** it checks CONDITION tokens only. "2x5m" is a
rule token, so the guard never looked at it.

**Scale (72h to 2026-09-01, 121 validator rejects) [A]:** legs on a non-sweep condition
= **35 (28.9%)** — breakdown_continue 24, reject 11; 7 landed on attempt 3/3; two
sessions fail-closed directly on it (2026-09-01 ASIA, 2026-08-31 NY).
Still live at 21:44:29 CT on the running binary: `📐 planner attempt 1/3 parse/schema
rejected: arm on S1 needs EXACTLY 2 legs (split contract), got 1` — 487.9 s of
max-reasoning burned on a contract the prompt never stated.

---

## 2. C5 — every condition-keyed validator restriction, before and after [A]

The wave's deliverable. "Prompt states it?" = the sentence exists in the RENDERED prompt
(`plannerOutputContract`), schema qualifier or prose. All 17 are now machine-checked by
`ValidatePromptContracts`; the BEFORE column is why the check was needed.

| # | Field / rule | Legal for | Validator site | Prompt BEFORE | Prompt AFTER |
|---|---|---|---|---|---|
| 1 | `arm{}` enabled | fvg_entry, reject, breakdown_continue, breakup_continue (sweep_reclaim via wait_confirm) | `plan_doc.go:159` | PARTIAL — "every fvg_entry / reject … SHOULD carry arm{}", breakout_retest excluded; reclaim/hold/acceptance unstated | `legal ONLY on fvg_entry\|reject\|breakdown_continue\|breakup_continue` |
| 2 | `arm.legs[]` present | sweep_reclaim ONLY | `plan_doc.go:167` | **NONE** | `(ONLY if condition is sweep_reclaim …)` + prose |
| 3 | `legs` count | EXACTLY 2 | `plan_doc.go:170` | **NONE** | `EXACTLY 2 legs` |
| 4 | non-sweep arms | SINGLE, no legs | `plan_doc.go:167` | **NONE** | `must arm SINGLE`, `no legs` |
| 5 | split `confirm.rule` | touch at the sweep ref | `plan_doc.go:173` | PARTIAL (entry law) | `confirm=touch at the sweep ref` |
| 6 | split leg wait_confirm | leg 1 false, leg 2 true | `plan_doc.go:180,183` | **NONE** | stated in the ARM SPLIT prose |
| 7 | split leg 2 rule | 1m_mss \| 1x5m_close, = confirm2.rule | `plan_doc.go:185,188` | PARTIAL / **NONE** for the equality | `confirm2 = 1m_mss or 1x5m_close`, `EQUAL to confirm2.rule` |
| 8 | split top-level | mirrors leg 1 | `plan_doc.go:197` | **NONE** | `top-level entry/stop/target mirror leg 1` |
| 9 | breakdown arm | breakdown{} + entry_mode=pullback | `plan_doc.go:146,150` | YES | unchanged (+ restated in ARM SPLIT) |
| 10 | sweep single arm | wait_confirm:true | `plan_doc.go:137` | YES | unchanged |
| 11 | `fvg{}` | iff fvg_entry | `plan_doc.go` schema | YES | unchanged |
| 12 | `breakdown{}` | iff waterfall-class | `plan_doc.go` schema | YES | unchanged |
| 13 | fades touch-only | reject, fvg_entry | `entry_law.go:145` | YES | unchanged |
| 14 | `2x5m_close` | breakdown/breakup ONLY | `entry_law.go:149` | YES but **misspelled as "2x5m"** in the Style quoted into rejections | `2x5m_close` everywhere |
| 15 | armed fade stop | ≥2 ticks beyond level | `entry_law.go:~176` | YES | unchanged |
| 16 | breakout_retest | never arms (GAR-F4) | `armed.go:19` | YES | unchanged |
| 17 | death/flip.rule enum | 2x5m \| 5m_close, SEPARATE from confirm | `plan_doc.go:285` vs `1105` | **NONE — two vocabularies, undeclared** | `death/flip rules use their OWN vocabulary` |

Rows 2, 3, 4, 6, 7(equality), 8 and 17 were enforced and unstated. Row 14 was stated in
the wrong spelling — the single most expensive character in the contract.

---

## 3. The fix — file:line [A]

| File | Lines | Change |
|---|---|---|
| `kernel/entry_law.go` | 24-31 (comment), 54, 58, 62, 66, 70, 74 | F1 — every `Style` spells the confirm enum form (`1x5m_close`, `2x5m_close`); a comment records that Style is quoted verbatim into rejections |
| `kernel/validator_hints.go` | `RepairEntryConfirmLaw` | F1 — `2x5m` → `2x5m_close` |
| `kernel/planner_prompt.go` | 578 (`legs`, `arm{}` qualifiers) | F2 — `(ONLY if condition is sweep_reclaim — the split contract, EXACTLY 2 legs; EVERY other condition arms SINGLE: omit legs)`; `arm{}` names its armable set |
| `kernel/planner_prompt.go` | 582 (flip line) | F1 — `// NOTE: death/flip rules use their OWN vocabulary (2x5m \| 5m_close) — NEVER the confirm enum` |
| `kernel/planner_prompt.go` | 596 (ENTRY LAW prose) | F1 — `never 2x5m_close`, `1x5m_close legal for the break leg`, `1x5m_close as fallback` |
| `kernel/planner_prompt.go` | 603 (new paragraph) | F3 — `ARM SPLIT vs ARM SINGLE`: the whole split contract, and that breakdown_continue / breakup_continue / reject / fvg_entry arm SINGLE with no legs |
| `kernel/validator_hints.go` | `HintRuleField`, `ruleTokenScan`, `validateHintTokens`, registry | F4 — every rule token checked against the enum of ITS OWN field; the 9 entry-law `Style` strings registered as hints |
| `kernel/prompt_contract.go` | new file, 17 rows | F5 — `PromptContracts()` + `ValidatePromptContracts()` + `PromptContractBootLine()` |
| `kernel/levels_volume_boot.go` | class-34 line + new line | F8 — `📜 prompt/validator contract: N restrictions, all stated in prompt (class 38 guard)` |
| `trader/auto_trader_planner.go` | `repairUnparseableLine`, `clampLine`, call site | F6 — parse error + the defect being repaired + a bounded head of the raw response |
| `web/src/guide/content/plays.ts`, `status.ts` | ENTRY LAW table, new ARM SPLIT block, boot ledger | A12 |
| `docs/superpowers/AUDIT-CHECKLIST.md` | class 38 after 37 | A16 |

**F4 is field-scoped, not a blanket ban** — `2x5m` is LEGAL in a death/flip hint and
ILLEGAL in a confirm-field hint; the guard encodes exactly that. `\b…\b` boundaries mean
`2x5m_close` never matches `\b2x5m\b` (the `_` is a word character), so only genuinely
bare spellings trip it.

---

## 4. Tests — E1 quoted failing, then passing (A8/E1-E6) [A]

**E1 RED, on the pre-fix tree** (`go test ./kernel -run TestClass38`), commit `d530260d`:
```
--- FAIL: TestClass38PinRows78to80
  entry law Style for "reclaim" names "1x5m" — not a confirm.rule enum member …
  entry law Style for "hold" names "1x5m" …
  entry law Style for "acceptance" names "1x5m" …
  entry law Style for "breakout_retest" names "1x5m" …
  entry law Style for "breakdown_continue" names "2x5m" …      ← row 78
  entry law Style for "breakup_continue" names "2x5m" …
  RepairEntryConfirmLaw names bare token "2x5m" …              ← row 79
  the "legs" schema field carries NO condition qualifier …     ← row 80
  the rendered prompt never says "arm SINGLE" / "no legs" / "EXACTLY 2 legs"
--- FAIL: TestClass38DeathFlipVocabularyIsDeclared
  … no statement that they are DIFFERENT enums
```
**GREEN after the fix** (`797a7d78`): all six class-38 tests pass.

| Test | Pins |
|---|---|
| `TestClass38PinRows78to80` | rows 78/79/80 against the live entry-law table and the live rendered prompt |
| `TestClass38DeathFlipVocabularyIsDeclared` | the two-enum declaration |
| `TestClass38HintGuardCatchesOutOfEnumToken` (E2) | the guard rejects the EXACT pre-fix text, accepts its fixed form, allows `2x5m` in a death/flip hint, rejects `1x5m_close` there |
| `TestClass38LiveHintRegistryIsClean` | the shipped registry passes AND the 9 entry-law Styles are actually registered |
| `TestClass38PromptContractsAllStated` (E3) | all 17 restrictions stated; every row carries a site and a non-empty fragment list |
| `TestClass38ContractTestFailsWhenPromptDropsARule` (E3 teeth) | removing any restriction's sentence FAILS the guard — no vacuous pass |
| `trader.TestClass38RepairUnparseableLineCarriesEvidence` (E4) | the F6 line carries parse error, defect and bounded raw head; a 30k response cannot flood the journal |

**E5 prompt delta:** output contract **10,219 → 11,284 chars (+1,065, ≈ +267 tokens)** —
about +4% of a ~6,400-token full-author prompt.

**E6:** `go test ./... -count=1` → **27 packages ok, 0 FAIL**. Goldens PASS. `tsc --noEmit`
clean. `npm run build` ✓ (4.28 s). `npx vitest run` → **295 passed, 0 failed**.

---

## 5. SURPRISE (A23) — a class-37 miss found by class 38 [A]

`web/src/guide/GuidePage.test.tsx` asserted **42** knob cards. Class 37 (`75130d59`)
added the "Planner stream total deadline" card and never bumped the census, so the web
suite has been failing on dev since that commit — **including the binary now live**
(class 37 ran `npm run build`, which passes, but not `vitest`). Verified by stashing only
the class-38 guide edits and re-running: the failure persists. Census bumped to **43** in
`18ef0a6e`, with the reason recorded in the test name.

**Standing lesson:** `npm run build` is not the web suite. A guide-content change needs
`npx vitest run` in the same wave.

---

## 6. F7 — REPORT ONLY (class 39), and the ruling needs a wording change [A]

The owner's ruling: *multiple legs on a non-sweep condition normalize to a single arm
with a WARN, keeping the leg whose confirm matches the single-arm contract, REJECT if
neither does; never normalize the reverse.*

**Sample honesty (A21/A24):** of the 35 journal instances, exactly **n=1** still has the
model's output retained — `planner_rejected_prompts` is capped at 20 rows
(`store/planner_rejected.go:43`) and the rest were trimmed. That one is row **69**
(2026-09-01 11:59:57 CT, the repair that embeds attempt 1's verbatim output). Everything
below is n=1 and must not be read as a rate.

Row 69, scenario S1, run against the real validator:
```
leg count as authored: 1                      ← the ruling says "multiple legs"
leg[0].rule = "touch" · single-arm allowed set for breakdown_continue = [1x5m_close 2x5m_close]
AS AUTHORED    → arm legs on breakdown_continue — arm_legs_sweep_reclaim_only
RULE-AS-WORDED → no leg matches the contract → would STILL REJECT
LEGS DROPPED   → <nil>   (valid)
top-level mirrors leg 0: entry 29130.00==29130.00 stop 29168.00==29168.00 target 29040.00==29040.00
```

Three findings for class 39:

1. **The validator branch is count-agnostic.** `plan_doc.go:167` fires on `len(legs) > 0`
   for any non-sweep condition. A rule scoped to "multiple legs" would not cover the one
   instance we can actually see, which had ONE leg.
2. **"Keep the matching leg" is the wrong test for the 1-leg case.** A leg's `rule` is the
   leg-2 CHAIN rule of a split; a single arm has no leg rule at all. Testing it against the
   condition's confirm set rejects a scenario that is otherwise valid.
3. **The right normalization is simpler: drop `legs[]` and re-validate the single arm.**
   In row 69 the arm already carried `wait_confirm:true`, a legal `confirm{}` of
   `1x5m_close`, `breakdown{}` with `entry_mode=pullback`, and top-level prices that
   already mirrored the leg. Removing `legs` alone makes it pass.

**Recommended class-39 wording:** *when a non-sweep condition carries `legs[]`, drop the
array and re-run `ArmSpecValid` on the remaining single arm; accept with a WARN naming
what was dropped, and REJECT unchanged if the arm without legs is still invalid. Never
synthesize a leg (one → two is authoring).* The 2-leg case is then covered too: the
top-level already mirrors leg 1 by contract, so dropping the array keeps the leg-1 entry.
Not shipped here — F7 is class 39, with its own gate-level fixture, guide text and
checklist entry.

---

## 7. Build, stage, rollback (A4/A13)

See the closeout section for the merged sha, the `go version -m` stamps and the staged
binary's sha256.

**Rollback (exact):**
```
cd ~/nofx && mv nofx-bin nofx-bin.bad.class38 && cp nofx-bin.prev.boot nofx-bin \
  && printf 'e42a0b43b4bead2c5d2207958d8a0bde2d65be11' > deploy/RELEASE \
  && kill -9 $(systemctl show -p MainPID --value nofx) \
  && git checkout -- deploy/RELEASE web/src/guide/types.ts
```
`nofx-bin.prev.boot` is written at cutover from the running class-37 binary (`e42a0b43`).
There is no env knob to soft-revert this wave: it is prompt text plus two guards.

---

## 8. Proof (A20) — NOT YET OCCURRED

The proving event is a live read authored against the NEW prompt landing on attempt 1 or
2 with no leg-contract or confirm-token reject. **No cutover has happened, so there is
nothing to quote.** The last read on the running binary (21:44:29 CT, class 37 rev
`e42a0b43`) burned attempt 1/3 in 487.9 s on `arm on S1 needs EXACTLY 2 legs (split
contract), got 1` — the defect class, still live, on the old prompt.

Separately owed: **class 37's** own behavioural proof is still outstanding — no read has
crossed 600 s and no `class=total_deadline` line has appeared since its 21:19:49 boot
(longest since: 487.9 s).

---

## 9. What the owner will STILL see wrong (A15)

- **Until cutover:** the running binary carries the OLD prompt. Every leg-contract and
  confirm-token reject continues, roughly 29% of validator rejects, at ~450 s of
  max-reasoning per burned attempt.
- **The guide is ahead of the binary.** `GUIDE_BUILT_REV` is bumped in the marker commit,
  so the drift banner shows until the cutover — expected, and the vite dev server serves
  the new guide text from source immediately.
- **F7 is not fixed.** Legs on a non-sweep condition still REJECT rather than normalize.
  The prompt now states the rule, which should reduce the shape; it does not eliminate it.
- **The 2-leg case is unverified.** n=1 retained sample, and it was a 1-leg arm. The class-39
  fixture must cover both.
- **The rejected-prompt store trims at 20 rows**, so this forensics window closes fast. If
  the owner wants a wider F7 sample, raise `plannerRejectedCap` BEFORE the next session.
- Row 71/72 (class 37's timeout text stored as a "validator reason") remains: a
  transport/deadline failure is still fed to attempt 2 as if it were a validator defect.
  Unchanged here — prompt paths were the scope, retry semantics were not.
