# 1E — MONTE CARLO DRAWDOWN RIG (offline, read-only, pre-registered)

**Status at this commit: PRE-REGISTRATION ONLY. No analysis has been run.**
Committed 2026-09-03 before any computation, per the dispatch. Results are
appended in a later commit; this header is not edited afterwards.

Scripts: `~/nofx-analysis/mc-drawdown/` (stdlib + numpy, seeded, re-runnable).
Engine: untouched. DB opened `mode=ro`. No knob writes, no lock, no live calls.

---

## PREMISES — VERIFIED BEFORE PRE-REGISTERING

| # | claim | measured | verdict |
|---|---|---|---|
| P1 | ~60–70 usable closed trades since 2026-08-15 | **n = 64** | **holds** |
| P2 | max_daily_trades=3, daily_loss_limit=$450, profit target $900, master OFF | all present; **and every guardrail is individually disabled too** | **holds, with an addition** |
| P3 | size clamped to 1 contract, $2/pt MNQ | confirmed exactly on 4 rows | **holds** |
| P4 | stops "30–40 pre-0B, ~20–80 post" | n=34 arms: median **26.75**, p25 21.96, p75 40.00, **max 150.00**, mean 37.75 | **broadly holds, fatter right tail** |

**P1 sample construction** (`trader_positions`, `mode=ro`):
`status != 'OPEN'` AND `created_at >= 2026-08-15` AND `pnl_corrected IS NOT NULL`
AND `source IN ('system','armed_entry','reconcile')` → **64 rows**.
70 closed in the window; **3 UNRESOLVED** (`pnl_corrected IS NULL`) excluded and
counted; **3 test-seam** (`source='e7_farside_test'`) excluded. Raw
`realized_pnl` is never read (A22).

**P2 resolved from the BOUND strategy** (`traders.strategy_id` → `a5b7662e…`,
name "MNQ"), never `LIMIT 1` — checklist class 9:

```
guardrails_enabled        = False      daily_loss_limit_usd      = 450
daily_loss_enabled        = False      daily_profit_target_usd   = 900
daily_profit_enabled      = False      max_daily_trades          = 3
max_daily_trades_enabled  = False      max_contracts_per_order   = 2
max_contracts_enabled     = False
day_plan.sessions max_trades = 10 / 7 / 10
```

> **Surprise, recorded before analysis:** the guardrails are not merely
> master-off — **each one is individually disabled as well**. Q3 is therefore
> entirely counterfactual: it asks what these limits WOULD have done, not what
> they did. Nothing in this report describes a control that is currently
> protecting the account. Also note `max_contracts_per_order = 2` in config
> while 0B's `ClampStageAContracts` caps at 1 — the clamp governs, so size is 1.

**P3** verified on ids 587–590: e.g. **590** entry 29193.25 → exit 29143.75 LONG
= −49.50 pts × $2 = **−$99.00** at qty 1. Point value $2/pt confirmed.

---

## QUESTIONS (pre-registered verbatim)

**Q1** Given the realized per-trade P&L distribution (`pnl_corrected`, n=64),
what is the distribution of MAX DRAWDOWN over the next 20 / 50 / 100 trades?
(p50, p90, p95, p99, worst)

**Q2** What is the probability of a losing streak of k = 4, 6, 8, 10 consecutive
trades within 50 trades, given the realized win rate?

**Q3** With the daily loss limit at $450 and 1 contract: on what fraction of
simulated days does the limit trip, and how much of the realized expectancy does
it forfeit (trades lost after a trip that would have been winners)? Same for
`max_daily_trades = 3`.

**Q4** Expectancy per trade with a 95% CI, and the n required to distinguish it
from zero at power 0.8 (formula stated).

**Q5** How do Q1–Q3 change if the per-trade distribution is drawn from the
pre-0B era vs post-0B (stops widened, BE/trail off)? Descriptive — post-0B n is
tiny; say so.

## METHOD (pre-registered)

- **M1** Trade-sequence bootstrap: IID resample of realized `pnl_corrected`,
  **B = 10,000** paths per horizon, AND a stationary block bootstrap
  (**mean block 5**) to preserve streakiness. Both reported; if they disagree
  materially, say which and why.
- **M2** Day simulation for Q3: group realized trades by session-day, resample
  **days** (not trades) to preserve within-day clustering, apply the daily limit
  / trade cap in sequence order, count trips, and compute forfeited P&L as the
  sum of post-trip trades' realized `pnl_corrected` on that day.
- **M3** Q2 streaks: exact via the realized p(win) and n, **plus** the bootstrap
  empirical count; both reported.
- **M4** Q4: mean, sd, t; `n_required = ((z_α + z_β)·sd/μ)²`, α = 0.05
  two-sided, power 0.8. If μ ≈ 0, report "n required → ∞ at the current
  expectancy" honestly.
- **M5** Multiplicity: none applied — every figure here is descriptive, and this
  is stated rather than silently assumed.
- **M6** Sensitivity: re-run Q1 with the single largest loss and the single
  largest win removed. **If the picture flips, the sample is too thin to carry a
  verdict — say so.**

## PASS RULES (pre-registered)

- A figure ships only with its **n**, its interval, and the **row ids** behind
  it (A21). No rate without n (A24).
- Ambiguity and exclusions are **counted and shown**, never dropped.
- **M6 flip ⇒ the report's headline becomes "sample too thin", regardless of how
  clean the central estimates look.**
- This wave issues **no verdict** on size, limits or exits. It is the INPUT to
  those rulings (dispatch stop-line).

---

*Results appended below in a later commit. Nothing above is edited after this
commit — the git history is the pre-registration's proof.*
