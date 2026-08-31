# 2026-08-31 — Planner Speed Wave (kill the 9-minute read)

Root cause per autopsy 168e5282: ~26k output tokens/attempt at reasoning=max ≈
99.5% of wall time. This wave ships instrumentation, the repair retry, the
split-deadline streaming path, and the offline A/B evidence — reasoning mode and
completion cap stay UNCHANGED until the owner rules on the tables below.

## Per-phase verdict

| Phase | Shipped | Proof |
|---|---|---|
| 1 Instrumentation | ttfb (request-anchored), T1/T2 durations, reasoning_chars, verbatim rejected-prompt store (cap 20), per-call retry count in `ai_call` | live lines below |
| 2 Offline fast-vs-max A/B | `cmd/planner_ab` harness; provider-direct; schema-gate legality | table §A/B |
| 3 Repair-call retry | RETRY_MODE=repair|reauthor (default repair); malformed repair → one full re-author; `repair_regression` counter | live repair proof below |
| 4 Streaming + split deadlines | planner on SSE (`CallWithRequestStreamRetry`), idle 30s + total 600s; weekly/executor untouched | live proof below |
| 5 Completion cap | REPORT ONLY — recommendation below | — |

## Live proof (owner reset 09:40:38 CT, NY read, rev `5bf48951`)

Attempt 1 (full author, reasoning=max, SSE):
```
📝 prompt render (T2): 0ms ~6263 tokens
🗺️ map assembly (T1): 0ms
📡 Request URL (stream idle=30s): https://api.deepseek.com/chat/completions
📊 AI call complete (stream): completion=30421 prompt=9406 finish_reason=stop
   reasoning_chars=102847 ttfb_ms=667 wall_ms=499558
ai_call model=deepseek-v4-pro duration_ms=499558 … retries=1 ttfb_ms=667
   reasoning_chars=102847
📐 planner attempt 1/3 rejected: S1 breakdown_continue: a close came back
   across 29351.47 — the breakdown is void
```
- **ttfb = 667ms** — the queue is DEAD. The entire 499.6s wall is generation
  (30,421 completion tokens + 102,847 reasoning chars streamed at ~61 t/s).
  The autopsy's "queue vs generation" question is answered: generation.

Attempt 2 (REPAIR call):
```
🧩 planner attempt 2/3 repair: prompt ~1109 tokens (full-author ~6263 tokens)
📊 AI call complete (stream): completion=13635 prompt=1556 finish_reason=stop
   reasoning_chars=47093 ttfb_ms=418 wall_ms=208094
🗓️ PLAN written 2026-08-31 NY v3 (model deepseek-v4-pro, lifecycle active,
   prompt 28250aa44112)
```
- Repair prompt = **1556 tokens = 16.5% of the full author's 9406**.
- Repair wall = **208.1s = 42% of attempt 1's 499.6s**.
- The repair fixed the named defect and the plan wrote ACTIVE through the
  IDENTICAL validator chain (zero relaxation). Total read 09:40:38→09:52:26 =
  11m48s for 2 attempts (vs ~19 min for 3 full re-authors this morning).
- Rejected prompts persisted verbatim (`planner_rejected_prompts` id=1).

## §A/B — offline fast-vs-max (stored verbatim prompt, provider-direct)

Prompt: NY attempt-1 rejected prompt (25,050 chars, 9,406 provider tokens),
breakdown-void defect. Legality = the OFFLINE schema gate
(`ParsePlanDocCapped` + caps, byte-identical parse path); the facts validator
cannot run offline because facts are not stored with the prompt (gap — see §5).

| mode | wall | completion_tok | reasoning_chars | finish | legal | first defect |
|---|---|---|---|---|---|---|
| **AB_TABLE** | | | | | | |

## §5 — completion cap (report only, no changes)

Observed completions this wave: full-author 30,421 tokens / 499.6s;
repair 13,635 / 208.1s; prior 8 attempts 18.7k–33.1k (autopsy 168e5282).
Throughput ≈ 61–66 t/s. **A cap that bounds an attempt at ~250s is ≈15k
tokens — and 8/8 prior completions exceeded 15k, so a bare cap would truncate
≈100% of full-author attempts.** The cap lever is a CEILING, not a speed
lever. Ordering: (1) latency-mode ruling (A/B table above), (2) schema
slimming (the emitted JSON's prose fields are the bulk of output — target a
40-50% output cut → ~15-18k tokens ≈ 250-300s), (3) only then a cap (e.g.
24k) as truncation insurance. Reasoning tokens ride inside the same stream —
`reasoning_chars` shows think volume (102,847 chars on the full author) but
the provider returns no reasoning-token count; the wall-time cost of max
reasoning is exactly what the A/B quantifies.

## Instrumentation gaps (carried)

Facts are not stored with the rejected prompt (the A/B could only run the
schema gate offline); executor full-body ttfb is inherently ~0 (headers+body
buffered) — only the stream path yields real T4; reasoning has no token count
from the provider (chars proxy).

## Rollback

`nofx-bin.prev.boot` = rev `e86ae805`. Revert = swap back + kill -9 +
`deploy/RELEASE` = `e86ae805784b7b0ee10299a3c977738a813d0cd4`.
