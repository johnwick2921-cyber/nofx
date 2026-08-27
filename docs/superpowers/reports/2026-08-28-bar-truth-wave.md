# BAR-TRUTH WAVE (S-1 + A-1 + A-2) — fix built, cutover pending owner ack

Branch `fix/bar-truth` off dev. Commit `dd3da1c9f2577e461e5ceac665fe137b4e50f1a9`.
Read-only arbiter evidence below; the full three-way diff runs AT CUTOVER via
the new owner endpoint.

## 1. ARBITER — pre-fix evidence (two-way) + the three-way mechanism

Pre-fix (persisted vs live engine):
- Engine at 11:56:15 CT: `stale_reeval_refused: drift_too_big (|5.25| >= 4.96 =
  0.25 x ATR 19.84)` — live ATR14(5m) = **19.84**.
- Independent recompute from the bars table at the same cut: **29.78**
  (stable across last-2000/3000/all windows; formula byte-verified against
  `market/data_indicators.go:86-116`).
- Conclusion: the persisted series is NOT the series the engine traded on —
  the per-frame persistence path (goroutine-per-frame + single-connection
  GORM + ingest-channel drops) punctured it.

Three-way mechanism (this wave): `POST /api/nt/bar-arbiter` (owner-only).
- `action=backfill` → one-shot deep `bars_subscribe` (MNQ 1m, bars_back 8640)
  on the LIVE connection; the AddOn replays deep history; a capture window
  records count + FNV-1a hash per (symbol,tf) = the **NT8-truth** series.
- `action=diff` → returns `{replay, cache{count,hash,atr5m}, db{count,hash,
  atr5m}, deltas[≤5 cache-vs-DB bars], drops}`. The ATR14(5m) is reimplemented
  independently in the handler (R2 — never calls the engine).

## 2. INGEST FIX (A-1)

- `fanOutBarPersist` no longer spawns a goroutine per frame. ONE worker drains
  a bounded 1024 queue, batches (256 rows or 300ms), and writes CLOSED bars
  only (empty batches are no-ops — intra-bar updates stay in-memory BY
  CONSTRUCTION).
- Queue-full drops + ingest-channel drops are counted (atomics) and logged
  **1-line/min** (`bars: persist queue summary: dropped=N flushed=M` /
  `bars: ingest drop summary: …`) — the per-drop WARN flood is gone.
- A dropped close self-heals: the next live frame re-derives the closed cache
  tail (`ClosedCacheTail`, window 8), and the deep backfill covers anything
  older.
- Tests: worker drains + survives a panicking persister; queue-full drops
  counted; no-op without a persister.

## 3. BACKFILL (S-1 repair)

The deep `bars_subscribe` replay (8640 × 1m ≈ 6 days) reseeds the kernel cache
AND persists through the historical path — repairing the punctured 08-24→now
window without a restart. The ATR probe re-runs on `action=diff`:
**stored vs live must match to 2dp** (E-proof slot).

## 4. RETENTION

Post-fix journald projection is measured at cutover: line-rate over one hour
× 2G cap. Target ≥ 7 days. (Pre-fix rate ~19k lines/5min ≈ 0.4 days — A-1.)

## 5. A-2 CLOSURE — the 354 NULL rows

Class split (fresh query): **317 `sync` + 37 `reconcile_flat`** — none have
verifiable stored prices (trader_fills carries no position_id; the boot
backfill already stamped the 171 rows that DID have stored prices). Per class:
**excluded** (not backfillable this wave).
Enforcement: `GetPositionStats`, `GetSessionDayActivity`,
`CountConsecutiveLossesSince` now carry `WHERE … pnl_corrected IS NOT NULL`
with A-2 comments; the excluded count surfaces as `excluded_null_pnl` — no
silent NULLs in any table we rule from. Fixture tests updated to seed verified
rows.

## 6. HYGIENE

- Removed the stale `/tmp/nofx-dev-check` worktree (detached a52de628).
- PR triage: closed 9 heads fully contained in dev — #45, #47, #48, #49, #50,
  #55, #57, #58, #59.
- Survivors (11): #64 regime-wave · #63 research-import · #62 brand-census ·
  #61 partner-sync · #60 pnl-record-integrity · #56 decision-anatomy ·
  #54 controls-runtime-verify · #53 strategy-controls-census · #52
  breakeven-audit · #51 ledger-close-sep · #46 forensics-zerotrade.

## 7. E-proof slots (filled AT CUTOVER — owner ack required)

1. Three-way diff table (`action=diff` JSON — replay/cache/db counts+hashes).
2. ATR probe: db.atr5m vs cache.atr5m to 2dp.
3. Drop counters: `ingest_*` + `persist_queue` = 0 across one full session.
4. Retention number: ≥7-day projection from the post-fix line rate.

## Cutover ask

Flat window + owner ack → build at this commit, `deploy/RELEASE` bump, kill -9
per canon (all-origin flat gate). On boot: run `action=backfill`, wait ~30s,
`action=diff`; quote all four E-proofs into this report.
