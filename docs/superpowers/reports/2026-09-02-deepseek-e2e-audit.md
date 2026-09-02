# END-TO-END AUDIT OF THE DEEPSEEK CALL SYSTEM — ROOT CAUSE (2026-09-02)

**READ-ONLY.** No code, config, knob, key or DB write; no restart, cancel, reset or cutover; no lock taken; **no live provider call made** (A3 — the running system's own calls are the only evidence). Work on `docs/deepseek-e2e-audit-0902` in worktree `~/nofx-dsaudit`; `~/nofx` untouched.

Evidence class on every line: **[RUNTIME]** journal/log/live state · **[DB]** query + result · **[CODE]** file:line · **[CONFIG]** resolved value · **[NET]** socket observation · **[DOC]** provider documentation. A [CODE]-only claim about live behavior is UNVERIFIED until a [RUNTIME]/[NET] line confirms it.

---

## 0. PREMISE CORRECTION AND LIVE STATE

**The dispatch says "Live rev 8a756bba (class 33)". It is not.** [RUNTIME] at audit start 2026-09-02 07:35 CT:

| item | value |
|---|---|
| PID | 2461883, started `Wed Sep 2 07:32:15 2026` |
| running binary | `vcs.revision=0d093c3b3a11fb6ea6cb19454ffa59a9f7bd9f8b`, `vcs.time=2026-09-02T12:22:31Z`, `vcs.modified=false` |
| `deploy/RELEASE` | `0d093c3b…` (equal) |
| boot line | `🔐 BOOT INTEGRITY OK — rev 0d093c3b3a11 · built 2026-09-02T12:22:31Z · expected 0d093c3b3a11 · goldens PASS` |
| dev tip | `7e7556b9` |
| main tree | porcelain clean; **no `~/nofx-main.lock` present** (A2 satisfied without taking one) |

`8a756bba` was the **06:57:49** boot (class 33); it was superseded 35 minutes later by the root-fix cutover. **Four boots today** [RUNTIME]: 00:11:47 `23f56f49` (pnl-truth) · 06:27:45 `d5a6e138` (class-41 transport wave) · 06:57:49 `8a756bba` (class 33) · 07:32:15 `0d093c3b` (root-fix). Everything below distinguishes pre- and post-**06:27:45**, the class-41 policy boundary.

Resolved AI configuration in force [RUNTIME boot 07:32:15]:
```
🧠 AI params in force: model=deepseek-v4-pro client_max_tokens=32768 planner_max_tokens=65536
   temperature=0.50 top_p=omitted timeout=600s (HTTP ceiling; non-stream paths)
   planner_stream_idle=30s planner_stream_total=1200s retries=2 backoff=2s
🛰 planner client: provider_row=8ef641a7-…_deepseek stream_idle=30s stream_total=1200s
   (AI_PLAN_TOTAL_DEADLINE_SECS) http_ceiling=600s (non-stream paths only) retries=2 backoff=2s cap=65536
🔁 planner stream policy (class 41): stream_tries=3 (AI_PLAN_STREAM_TRIES, counts CALLS;
   AI_MAX_RETRIES=calls non-stream only) backoff=2s→15s→45s (AI_PLAN_STREAM_BACKOFF)
   watchdog_log=on (⏱ line on fire, per SSE line) keepalive=30s (dialer)
   serialize_executor=off resend_identical=on (transport/deadline → same prompt, no reject block)
```

---

## 1. THE HEADLINE — WHAT ACTUALLY FAILS

**The call system is not primarily failing at the transport. It is failing at the validator, and that is what costs sessions.**

Census of every planner attempt that did not yield a usable plan, 2026-08-26 → 2026-09-02 07:40 CT. Source: the `📐 planner attempt N/3 (rejected|parse/schema rejected|failed)` lines in `data/nofx_2026-08-2[6-9].log`, `nofx_2026-08-3*.log`, `nofx_2026-09-0*.log` (the file logs are authoritative — the journal only reaches back to 08-27 13:37). **n = 194 failed attempts** [RUNTIME].

| family | n | share |
|---|---|---|
| `breakdown_continue: a close came back across X — the breakdown is void` | 37 | 19.1% |
| `arm_legs_sweep_reclaim_only` (split legs on a non-`sweep_reclaim` condition) | 28 | 14.4% |
| `flip.rule "2x5m_close" invalid (2x5m\|15m_close\|5m_close)` | 18 | 9.3% |
| `confirm/confirm2.rule not allowed for <condition>` | 13 | 6.7% |
| `arm on S<n> needs EXACTLY 2 legs (split contract), got 1` | 11 | 5.7% |
| `no JSON object found in planner output` | 7 | 3.6% |
| `fade_requires_touch` | 6 | 3.1% |
| `leg 2 must chain (wait_confirm true)` | 4 | 2.1% |
| **the 51 remaining, individually classified** — `breakdown_continue` displacement below `BD_MIN_DISP_ATR` (8), `fvg: no fresh 3-candle gap` / stale gap (7), `only 3 levels on a side, ≥4 required` (10), gap-day trigger-side rules (4), `reject_retest` invalid (2), `death.rule invalid` (2), `arm enabled on non-armable condition` (2), `confirm2.rule "displacement"/"2x5m" invalid` (5), others (11) | 51 | 26.3% |
| **MODEL-OUTPUT SUBTOTAL** | **175** | **90.2%** |
| `503 Server Overloaded` | 3 | 1.5% |
| transport / deadline / EOF / reset / empty-response | 16 | 8.2% |
| **PROVIDER + TRANSPORT SUBTOTAL** | **19** | **9.8%** |

Every one of the 51 "remaining" was inspected individually and is **also** a model-output reject — there is no unclassified residue.

**And the outcome that matters:** 14 `🚨 PLANNER FAIL-CLOSED` events in the window. Grouping their quoted reasons — **14 caused by model output, 0 caused by transport, a 503, or a deadline** [RUNTIME]. Verbatim, every fail-close in the window:

| when (CT) | session | fail-close reason (verbatim, truncated) |
|---|---|---|
| 08-31 01:06:50 | 2026-08-30 ASIA | `only 3 levels below price 29430.25 but the machine table offered 35 — the plan must carry ≥4 on EACH side` |
| 08-31 06:50:50 | LONDON | `S3 breakdown_continue: a close came back across 29437.00 — the breakdown is void; author a reject/retest play instead` |
| 08-31 08:33:54 | LONDON | `arm on S2 split requires confirm=touch at the sweep ref (leg 1 rests AT the level)` |
| 08-31 08:47:19 | NY | `arm legs on reject — arm_legs_sweep_reclaim_only` |
| 08-31 09:33:18 | NY | `scenario[0].confirm2.rule "1m_mss" not allowed for breakdown_continue` |
| 08-31 17:18:32 | ASIA | `arm legs on reject — arm_legs_sweep_reclaim_only` |
| 09-01 01:01:09 | 2026-08-31 ASIA | `scenario[4].confirm2.rule "touch" not allowed for breakdown_continue` |
| 09-01 01:49:31 | LONDON | `S3 breakdown_continue: a close came back across 29502.25 — the breakdown is void` |
| 09-01 17:23:14 | ASIA | `S2 breakdown_continue: a close came back across 29130.50 — the breakdown is void` |
| 09-01 18:00:41 | ASIA | `arm legs on breakdown_continue — arm_legs_sweep_reclaim_only` |
| 09-01 21:14:52 | ASIA | `S4 breakdown_continue: a close came back across 29047.75 — the breakdown is void` |
| 09-01 21:53:46 | ASIA | `arm on S1 leg 2 must chain (wait_confirm true) on its confirm rule` |
| 09-02 01:37:44 | LONDON | `S2 breakdown_continue: a close came back across 29021.25 — the breakdown is void` |

(13 rows carry a distinct timestamp; the 14th match is the duplicate reason line emitted with the NO-TRADE write.)

**Session cost.** 2026-09-01 ASIA fail-closed **four times** (17:23, 18:00, 21:14, 21:53) before an owner reset at 23:12:44 rescued it; 2026-09-02 LONDON fail-closed at 01:37:44 and needed an owner reset at 07:15:26 [DB `plans`]. Both rescues were manual.

---

## 2. G2 — THE 09-02 01:0x 503 BURST, AND WHY IT DID **NOT** CAUSE LONDON'S FAIL-CLOSE

The dispatch asks for the burst "in full and its relation to LONDON's fail-close". They are **unrelated**, and the evidence is unambiguous.

**The burst** [RUNTIME, `data/nofx_2026-09-02.log`]. In the window 00:55–01:40 CT there were **22 `class=http_status http_status=503`** failures plus one `class=client_timeout` and one `class=other`. The 503s land on an **ASIA level-event wake read**, not on LONDON:

```
09-02 01:15:47 🗓️ level wake seated OB(bull)·1h invalidated: close 29049.00 below 29085.00 (noise 4.30)
               on ASIA 2026-09-01 — waking the planner (W6, 5th wake-up).
09-02 01:15:47 🧠 planner mode: fast-market (drift 50.2 pts = 2.8×ATR5m) — reasoning downgraded to fast→low (F3)
09-02 01:15:50 📐 planner attempt 1/3 failed: still failed after 2 retries: API error (status 503):
               {"error":{"message":"Server Overloaded","type":"service_unavailable_error",…}}
09-02 01:15:50 🧩 planner attempt 2/3 reauthor+block: prompt ~6864 tokens (full-author ~6744 tokens)
09-02 01:15:53 📐 planner attempt 2/3 failed: still failed after 2 retries: API error (status 503): …
09-02 01:15:53 🧩 planner attempt 3/3 reauthor+block: prompt ~6864 tokens (full-author ~6744 tokens)
09-02 01:15:57 📐 planner attempt 3/3 failed: still failed after 2 retries: API error (status 503): …
09-02 01:15:57 🗓️ wake re-read failed for 2026-09-01 ASIA (benign — active plan kept): … (status 503) …
```

**The whole three-attempt planner budget was consumed in 7 seconds** (01:15:50 → 01:15:57) — 3 attempts × 3 calls = **9 provider calls in 7 seconds**, every one a 503, with only the pre-class-41 2-second client backoff between them. Because this was a **wake** read (`failClosed=false`) the outcome was benign and the active plan was kept. **This is a near-miss, not a non-event: the identical burst on a scheduled read would have fail-closed the session in 7 seconds.** See §J1-2.

**LONDON's fail-close was a different read entirely**, 22 minutes later, and every one of its three attempts was a *validator* reject [RUNTIME]:

```
09-02 01:32:54 🧠 planner call (reasoning=fast→low …) completed in 173.2s
09-02 01:32:54 📐 planner attempt 1/3 rejected: S1 breakdown_continue: a close came back across 29021.25 — the breakdown is void …
09-02 01:32:54 🧩 planner attempt 2/3 repair: prompt ~922 tokens (full-author ~6555 tokens)
09-02 01:35:04 🧠 planner call … completed in 130.4s
09-02 01:35:04 📐 planner attempt 2/3 parse/schema rejected: scenario[0].confirm.rule "1x5m_close" — fade_requires_touch …
09-02 01:35:04 🧩 planner attempt 3/3 reauthor+block: prompt ~6682 tokens (full-author ~6555 tokens)
09-02 01:37:44 🧠 planner call … completed in 160.3s
09-02 01:37:44 📐 planner attempt 3/3 rejected: S2 breakdown_continue: a close came back across 29021.25 — the breakdown is void …
09-02 01:37:44 🚨 PLANNER FAIL-CLOSED 2026-09-02 LONDON: S2 breakdown_continue … — writing a NO-TRADE plan
09-02 01:37:44 🗓️ PLAN written 2026-09-02 LONDON v1 (model deepseek-v4-pro, lifecycle no_trade, prompt ae539d43b6ce…)
```

All three LONDON calls **succeeded at the transport** (173.2 s, 130.4 s, 160.3 s of streamed reasoning, no 503, no cut). **Verdict: the 503 burst cost nothing; the model's inability to satisfy the `breakdown_continue` rule cost the session.**

Two further live defects visible in the same window [RUNTIME]:
- `09-02 00:59:47 ai_call model=deepseek-v4-pro duration_ms=600000 finish_reason=n/a ok=false retries=1 ttfb_ms=0 reasoning_chars=0 timeout_source=client deadline_s=600 class=client_timeout http_status=200 err="failed to read response: context deadline exceeded …"` — a **600-second executor hang**, immediately followed by `⏱ cycle overran the scan interval (10m0.047s > 2m0s) — next tick delayed, in-flight work never cancelled; intervening ticks skipped`. Ten minutes of executor blindness from one call.
- `09-02 01:03:52 … duration_ms=244270 … timeout_source=transport … class=other http_status=200 err="fail to parse AI server response: API returned empty response"` — **HTTP 200 with an empty body after 244 seconds**, labelled `class=other` and `timeout_source=transport`, which is the mislabelling described in §E4.

---

## 3. G4 — PROVIDER STATUS FOR THE WINDOW

[DOC] `https://status.deepseek.com/` fetched during this audit: **no incidents, no degraded-performance events and no maintenance windows are reported between 2026-08-26 and 2026-09-02.** Current status "All systems are operating as expected"; API uptime for Jun–Sep 2026 shown as 99.81%–100%; no incident mentions the chat-completions API, streaming, or capacity. **So the 503 bursts we observed are below the provider's own incident threshold** — they are ordinary load-shedding, not an outage, and no fallback should be justified on the basis of a published incident.

---

## 4. THE DOMINANT ROOT CAUSE — A STANDING PROMPT INSTRUCTION THAT FIGHTS A TAPE-AWARE VALIDATOR

`breakdown_continue` / `breakup_continue` appear in **82 of the 194 failed attempts (42.3%)** and in **8 of the 13 fail-closes** [RUNTIME]. Of those 82, **38 are the same rule**: `a close came back across <level> — the breakdown is void`. This is not a transport problem and not a model-quality problem. It is a **contradiction between two parts of our own system**, and the mechanism is fully traceable:

**1. The prompt issues a standing, unconditional MUST** [CODE] `kernel/planner_prompt.go:589`:
> `"If price sits BELOW PDL you MUST write a continuation short; ABOVE PDH, a continuation long. "`

**2. "Continuation" is bound to the two waterfall conditions** [CODE] `kernel/planner_prompt.go:623`:
> `"WATERFALL PLAY (F1): author breakdown_continue|breakup_continue when the tape shows one-sided delivery, a >1.2×ATR gap-and-go, or a waterfall after a failed rally — the momentum-follow class."`

**3. The prompt NEVER tells the model that a given breakdown level is already void.** [CODE] A grep of `kernel/planner_prompt.go` for `void|reclaimed|came back across` returns only `:592`, an unrelated flip/death rule. The facts snapshot carries levels, ATR, regime — but not "the tape has already closed back across 29021.25, so a continuation there is illegal".

**4. The validator computes exactly that fact at WRITE time and rejects on it** [CODE] `kernel/breakdown_continue.go:254`:
> `return fmt.Errorf("%s %s: a close came back across %.2f — the breakdown is void; %s", s.ID, s.Condition, bd.Level, BreakdownReclaimedHint)`

**5. The correction is outgunned by the instruction.** A full re-author sends the **entire standing prompt plus a reject tail** [RUNTIME]: `🧩 planner attempt 3/3 reauthor+block: prompt ~6682 tokens (full-author ~6555 tokens)` — the reject block is **~92–127 tokens against a ~6,341–6,691-token prompt, i.e. ~1.5%**, and the standing MUST at `:589` is still inside it. The model is told "you MUST write a continuation short" in the body and "do not write the continuation you just wrote" in a footnote.

**6. The live regression, in one session** [RUNTIME] 2026-09-02 LONDON:
```
01:32:54  attempt 1/3 rejected: S1 breakdown_continue: a close came back across 29021.25 — the breakdown is void
01:32:54  attempt 2/3 repair: prompt ~922 tokens          → model switches to a reject fade
01:35:04  attempt 2/3 parse/schema rejected: scenario[0].confirm.rule "1x5m_close" — fade_requires_touch
01:35:04  attempt 3/3 reauthor+block: prompt ~6682 tokens (full-author ~6555)
01:37:44  attempt 3/3 rejected: S2 breakdown_continue: a close came back across 29021.25 — the breakdown is void
01:37:44  🚨 PLANNER FAIL-CLOSED … LONDON
```
Attempt 2 **obeyed** the hint and moved to a `reject` fade — and was then killed by a *different* rule (`fade_requires_touch`, the reject fade must confirm on touch, not on a close). Attempt 3, a full re-author, **regressed to `breakdown_continue` at the very same level 29021.25**. Three calls, 463.9 s of successful streaming, session lost.

**Is this a hard deadlock?** No — and the report will not overstate it (A12). The gap-day validator only requires *a short in any condition* [CODE] `kernel/plan_doc.go:841-842` calling `hasDirection(d.Scenarios, "short")` (`:877-884`), so a plain `reject` short satisfies it. **The deadlock is in the wording, not the code**: the error message says `the plan MUST include a continuation/breakdown short scenario` and the prompt says `you MUST write a continuation short`, both of which name the one condition family the tape may have already invalidated. This is the class-34 disease (a hint that names a target the validator then punishes), one layer up: **the instruction, not the hint, is now the thing that cannot be obeyed.**

**Owner ruling candidates (no change made — A1):**
- Make `:589` tape-aware, or soften it to "a short-direction scenario" and let the model choose the condition; and align the `plan_doc.go:842` message with what the code actually requires.
- Feed the validator's own knowledge forward: put "breakdown at X is VOID (close came back at HH:MM)" into the facts block so the model never authors it.
- Put the reject block at the **top** of the re-author prompt, or drop the conflicting standing line from re-author prompts specifically (lost-in-the-middle: a 1.5% tail after 6.5k tokens is the weakest position in the context).

---
