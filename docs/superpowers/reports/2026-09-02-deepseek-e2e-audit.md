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

## 5. THE TRANSPORT FAILURES THAT ARE REAL — WHAT THEY COST, AND WHAT THEY DO NOT

Transport failures are only 9.8% of failed attempts and have caused **zero** fail-closes, but they are not free. Three distinct mechanisms, all [RUNTIME] over the 7-day window:

**5.1 — The 600-second ceiling, hit 16 times.** Every occurrence:
`08-26 20:40:02 · 08-27 22:19:20 · 08-28 12:41:05 · 08-28 12:51:06 · 08-30 21:20:01 · 08-30 23:30:11 · 08-31 00:58:42 · 08-31 09:22:11 · 08-31 10:35:38 · 08-31 10:51:29 · 08-31 11:53:29 · 08-31 14:27:28 · 09-01 06:00:01 · 09-01 12:33:07 · 09-01 12:43:07 · 09-02 00:59:47` (CT).
These split into **two completely different failures wearing the same label**:

- **Streamed-and-starved (n=7 with counters):** e.g. `08-31 10:51:29 … duration_ms=600001 … ttfb_ms=562 reasoning_chars=140177 timeout_source=client deadline_s=600`. First byte in 562 ms, then **140,177 characters of reasoning received**, and the call was killed at the ceiling with nothing usable. Also 126,768 (10:35:38) · 134,792 (11:53:29) · 133,667 (14:27:28) · 134,322 (09-01 06:00:01) · 73,196 (09-01 12:33:07) · 71,414 (09-01 12:43:07). This is the class-37 disease and is what the planner's 1200 s split was built for — the boot line now reads `stream_total=1200s (class 37: planner ceiling split from the HTTP ceiling)`.
- **Never-started (the newest, 09-02 00:59:47):** `duration_ms=600000 … retries=1 ttfb_ms=0 reasoning_chars=0 timeout_source=client deadline_s=600 class=client_timeout http_status=200 err="failed to read response: context deadline exceeded"`. **Zero bytes in ten minutes.** DeepSeek documents that it closes a connection if inference has not started within ~10 minutes; our non-stream HTTP ceiling is 600 s = the same ten minutes. We do not distinguish "queued behind the provider's backlog and never started" from "the network died" — both land as `timeout_source=client`, and both cost the caller a full ten minutes.

**5.2 — The executor loses its cadence, 96 times.** `⏱ cycle overran the scan interval (…) — next tick delayed, in-flight work never cancelled; intervening ticks skipped` fired **96 times** in the window; the worst was `09-02 00:59:47 (10m0.047s > 2m0s)` — a single hung call cost **ten minutes of executor blindness** on a 2-minute cadence, with the in-flight work explicitly never cancelled. Others in the last two days: 2m41s · 4m54s · 2m19s · 3m0s · 2m44s · 3m30s · 3m9s. The executor is not being killed by the provider; it is being *held* by it.

**5.3 — The 503 burst is one hour of the whole week.** All **35** call-level 503s in seven days fall in a single hour, `09-02 01:xx CT` (= 15:xx Beijing) [RUNTIME]. There is no chronic 503 condition, and **0 fail-closes cite a 503**. But the burst exposed a real hazard: **3 planner attempts × 3 calls = 9 provider calls in 7 seconds** (§2), which on a *scheduled* read would have destroyed the session in the time it takes to read this sentence. That it landed on a wake read was luck, not design.

**The mislabelling that hides all of this** [RUNTIME]: `timeout_source=transport` appears on `09-02 01:03:52 … class=other http_status=200 err="fail to parse AI server response: API returned empty response"` — an HTTP 200 with an empty body after 244 s, which is not a transport event at all. §E4 enumerates every site with this defect; the census in §1 was built by reading durations and error text rather than trusting the label.

---

## 6. SECTION C — REQUEST CONSTRUCTION

### C1 Prompt assembly

Five calling paths build prompts [CODE]: planner **full-author** (`kernel.BuildPlannerPrompt`, `kernel/planner_prompt.go:319`, invoked `trader/auto_trader_planner.go:902`) · planner **repair** (`kernel/planner_repair.go:12`, invoked `:1436`) · planner **re-author+block** (`userPrompt = prompt + rejectBlock`, `:1438`, block from `plannerRejectBlock` `:1242`) · planner **resend-identical** (class-41 M0, `:1425-1430`) · **executor** (`kernel/engine_prompt_futures.go:24` + `kernel/engine_analysis.go:526`, called at `:676` via `CallWithMessages`, **non-stream**) · **weekly** (`kernel/weekly_prompt.go:276`, non-stream) · **Ask-Planner** (`api/handler_plan.go:1398/:2150`, non-stream). All planner variants share one 173-char system prompt (`trader/auto_trader_planner.go:24`).

**The class-38 contract text** [DOC-internal `docs/superpowers/reports/2026-09-01-class38-contract-mismatch.md`, merged `c0580011`, cut over 22:22:58 CT 09-01] found **17 condition-keyed validator restrictions**, 7 of them enforced-but-unstated and one stated in the wrong spelling (`2x5m` where the enum is `2x5m_close`). It is live [RUNTIME 07:32:15]: `📜 prompt/validator contract: 17 restrictions, all stated in prompt (class 38 guard)`. Its cost: the output contract grew 10,219 → 11,284 chars (+267 tokens, ≈+4%).

**C1.5 — the token estimator is wrong by a third, always in the same direction.** [CODE] `trader/auto_trader_planner.go:1373-1379` `estimatePromptTokens = (len(s)+3)/4` — a flat 4 chars/token. Measured against the provider's own `prompt=` count on 10 paired calls [RUNTIME]:

| path | n | min err | median err | max err | real chars/token |
|---|---|---|---|---|---|
| planner (T2 estimate vs `prompt=`) | 10 | −25.10% | **−32.02%** | −32.76% | 2.69–3.00 |
| executor (exact DB char counts vs `prompt=`) | 10 | −41.66% | **−41.70%** | −42.16% | 2.31–2.33 |

Zero over-estimates in 20 pairs. The executor prompt is denser (numeric OHLCV tables) so a single divisor cannot serve both. **Consequence, stated honestly:** no cap, budget or gate reads this number today [CODE, exhaustive grep — its three call sites are all log statements], so the error costs nothing now; it only makes three log lines wrong. The moment anyone sizes a context budget or a cost estimate on it they will be low by a third to two-fifths.

### C2 Request body — what we send, verbatim

Planner stream [CODE] `mcp/client.go:976-1085` + `applyDeepSeekThinkingDefaults` `:1087-1099`:
```json
{"model":"deepseek-v4-pro",
 "messages":[{"role":"system","content":"…173 chars…"},{"role":"user","content":"…~26.7 KB…"}],
 "temperature":0.5, "max_tokens":65536, "stream":true,
 "thinking":{"type":"enabled"}, "reasoning_effort":"max"}
```
Executor non-stream [CODE] `mcp/client.go:482-530` — **a second, independent builder**:
```json
{"model":"deepseek-v4-pro",
 "messages":[{"role":"system","content":"…~11.0 KB…"},{"role":"user","content":"…~16.2 KB…"}],
 "temperature":0.5, "max_tokens":32768,
 "thinking":{"type":"enabled"}, "reasoning_effort":"low"}
```
(`stream` is absent entirely on this path; its `messages` are `[]map[string]string` and therefore **structurally cannot** carry `reasoning_content` or `tool_calls`.)

Omitted by zero-value/`omitempty`: `top_p` (boot says `top_p=omitted`), penalties, `stop`, `tools`, `response_format`, `logprobs`, `n`, `seed`, and **`stream_options` — the key exists nowhere in the codebase** [CODE, exhaustive grep for `stream_options|include_usage` → zero hits].

### C3 Headers — exactly two, and they are the documented two

`Content-Type: application/json` [CODE] `mcp/client.go:642` and `Authorization: Bearer <redacted>` [CODE] `:478-480` → `mcp/provider/deepseek.go:66-68`. Both paths share one builder; **there is no separate header path for SSE**. Not set: `Accept` (so **no `Accept: text/event-stream`**), `Connection`, `User-Agent` (Go emits `Go-http-client/1.1`), `Accept-Encoding` (Go auto-adds gzip). **Headers DeepSeek documents that we omit: none** — [DOC] their documented request is exactly those two headers, and streaming is selected by the body's `"stream": true`, not by `Accept`.

### C4 Provider rows [DB `ai_models`, 11 rows — key material never read or printed]

Two DeepSeek rows, **both enabled**, both with a key present, both with empty `custom_api_url`/`custom_model_name` (so both resolve to the package defaults `https://api.deepseek.com` + `deepseek-v4-pro` [CODE] `mcp/providers.go:19-20`):

| row | name | enabled | key | bound to |
|---|---|---|---|---|
| `8ef641a7-…_deepseek` | DeepSeek AI | 1 | present | **the only trader** (`hoang`), via `ai_model_id` |
| `396db319-…_deepseek_0751c0b6` | **DeepSeek 2** | **1** | present | **nothing** — no trader, and `planner_model` is empty on all 9 strategies [DB] |

The other 9 rows are disabled and keyless. **Every path shares one client object**: `resolvePlannerClient` returns `at.mcpClient` verbatim when the planner binding is empty [CODE] `trader/auto_trader_planner.go:64-105`; [RUNTIME] `🧠 planner model: empty binding → using primary, pinned "deepseek-v4-pro"` and boot `🛰 planner client: provider_row=8ef641a7-…_deepseek`.

**DeepSeek 2 is enabled, keyed, and unreachable.** The only selector that could pick it (`store.PickProviderModel` → `betterProviderModel` `store/ai_model.go:227-252`) fires **only when no row's id matches the trader's `ai_model_id`** — which never happens today. **There is no runtime failover anywhere** [CODE, exhaustive]: nothing re-resolves the client after a failure; a 401/402/503 on the bound row never reaches for the second. The tiebreak, were it ever to run, separates the two rows by **120 microseconds** of `updated_at`.

## 7. SECTION G1 — DOCUMENTED PROVIDER BEHAVIOR vs OURS

| # | documented behavior | our handling | verdict |
|---|---|---|---|
| a | Under load the server holds the stream with `: keep-alive` **comment lines** | `onLine()` fires on **every** scanned line — blanks and `:` comments included — **before** the `data: ` filter, and pushes to `resetCh` [CODE] `mcp/client.go:1411-1424`, watchdog `:1207-1240`. **Keep-alives DO reset our idle watchdog.** | **COMPLY** |
| a3 | Rate limiting is by **concurrency** (cap 500 for this model), not RPM; overflow → 429 | We run ≤2 concurrent calls; 429 and all 5xx are treated as provider-side (`IsProviderFailure`, resend identical) [CODE] `:263-280` | COMPLY |
| **b** | **"If the request has not started inference after 10 minutes, the server will close the connection."** | Planner: idle 30 s + total **1200 s** → survives a queue (keep-alives reset the timer), dies at 1200 s or on the server's close → `class=transport` → class-41 resend-identical. **Executor: `http.Client.Timeout = 600 s` — exactly the documented 10 minutes.** | **COMPLY (planner) / GAP (executor)** — zero margin, and our timer and theirs fire at the same instant, so the two causes are indistinguishable in our logs. This is precisely the `09-02 00:59:47` zero-byte 600 s hang in §5.1. |
| c2 | `finish_reason: "length"` | Non-stream: counter + `🚨` WARN [CODE] `:591-597`. **Stream path stores the reason and does nothing else** [CODE] `:1263-1276` | **PARTIAL GAP** — `truncated-responses=N` on the boot line counts **executor truncations only**; a truncated *plan* would never appear in it (latent, n=0) |
| c3 | `finish_reason: "content_filter"` | **Not handled on either path**; partial content is parsed as if complete | **GAP** (latent, n=0 of 3,557) |
| c4 | `finish_reason: "insufficient_system_resource"` | **Not handled, and not in `retryableErrors`** — an interrupted response is fed to the plan parser as a normal answer, most likely failing schema validation and burning an attempt on a *provider outage* dressed as a validator reason | **GAP** (latent, n=0) — the same shape class-41 fixed for HTTP status |
| d | `reasoning_content` must **not** be echoed back without `tools` (ignored); **must** be echoed with `tools` (else HTTP 400) | Planner/repair never echo it (reasoning is accumulated separately, only `ReasoningChars` kept) [CODE] `:1404-1487`; executor structurally cannot; the agent path (the only one sending `tools`) **does** echo | **COMPLY on all four paths** — though two code comments assert the opposite of the docs (§Contradictions) |
| e | Max output for `deepseek-v4-pro` = **384K**; context 1M | 65,536 (planner) / 32,768 (executor) | **COMPLY** — 65536 is legal, 17% of the ceiling |
| f2 | Thinking mode **does not support** `temperature`/`top_p`/penalties | We send `temperature: 0.5` on **every** call | **GAP (benign on the wire, misleading in the ledger)** — the provider ignores it, but the boot line advertises `temperature=0.50` as an AI param "in force". It is not in force. |
| g1 | HTTP **500** = retry after a brief pause | **`500` is missing from `retryableErrors`** [CODE] `mcp/client.go:34-52` (429/502/503/520/524 are present). The planner survives via `IsProviderFailure` (`st >= 500`), but **the executor's `CallWithMessages` loop would give up on a 500** | **PARTIAL GAP** |
| g2 | 401/402/422 = do not retry | Not retried | COMPLY |

---
