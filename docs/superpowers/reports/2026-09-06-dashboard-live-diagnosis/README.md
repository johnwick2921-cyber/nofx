# Dashboard live diagnosis — report before fixes

**Status: diagnosis and proposed changes only. No fix, configuration change, order action, backfill, merge, or restart is included.** The owner requested a careful report with evidence published on GitHub before approving any changes.

## Provenance and evidence boundary

Running source: `a3471b3f4cd1fd07989b28cfd2fa7490ffed327e`. Documentation branch starts at current dev `eb2c65c882bc14c35f292322f0afc82691b185da`. Checklist freshness: `dd35112117d0b06ca90387236a1c636666fd4025 docs(checklist): class 83 and R10 — an acknowledgement mistaken for a settlement, outside the broker`. Applied AUDIT-CHECKLIST classes 28, 52, 68, 74 and R10 where relevant: canonical time, replay merge, real serving path, separate verified deployment, pinned publication.

Evidence tiers: **A** directly read or calculated; **B** inference; **C** untested hypothesis. Observations are timestamped snapshots, not a promise of current state. [Machine-readable evidence](evidence.json) contains scoped API facts, database rows, selected NinjaTrader and Go log excerpts, source revision, and arithmetic. [Verifier](verify.py) recomputes the candle gap and scenario arithmetic without network or production access. Account/user identifiers, tokens, credentials, full prompts and unrelated trades are excluded.

## 1. Internal arm presented as a resting broker order [A]

Row **106**, ASIA **S2**, `kind=stop_entry`, `state=armed`, `signal_id=""`, `placement_seq=0`. Created 17:01:34 CT. Broker snapshots report zero working orders while the internal ledger contains one arm. This is a prepared intent that has not been submitted; the mismatch alone does not establish a lost broker order.

**Direct cause:** deployed `STOP_ENTRY_SEAM=off`, confirmed by the single-key configuration read and the boot log. The placement branch explicitly continues without sending or changing the arm when this switch is off. Restarting NinjaTrader does not change that switch.

**Presentation defect:** `deskArms` counts all nonterminal rows and renders `%d resting`, including rows whose detail says `NOT placed at broker`. `deskBook` reuses the cutover gate; that gate compares broker orders with all nonterminal internal rows. It is conservative for deployment safety, but its generic MISMATCH wording does not explain a deliberately unsent arm.

**Proposal:** distinguish prepared, blocked (with exact reason), submitted, broker-acknowledged and terminal states. Show prepared count separately from broker working-order count. Preserve the conservative cutover gate; do not silently exclude an unresolved intent just to show green. A nonempty send identifier alone must not become proof of broker acknowledgement.

**Not proposed:** enabling stop-entry, deleting/cancelling row 106, changing thresholds or sending an order. Those are separate behavioral decisions requiring review.

## 2. Price-feed status mislabeled as transport link [A]; underlying feed cause unresolved

`deskFeed` and the MODE row print `FeedStatus()` as `link`. That function returns the last NT8 **price feed status**, not the TCP socket state. The system separately implements `IsConnected()` for TCP and `IsFeedConnected()` for the execution gate. The latter can accept a fresh live bar update despite an older disconnected status. Historical replay does not set that live-update freshness stamp.

Observed after the owner reset: historical frames and fresh broker snapshots arrived, while the last price status was Disconnected. This establishes transport traffic; it does **not** by itself establish a healthy live price feed. Bar timestamp freshness also cannot distinguish a recent historical download from continuing live delivery. Live bar updates are sampled at DEBUG, so their absence in INFO logs is not proof of absence.

NinjaTrader log excerpts show Connected at 17:06:55, Disconnected at 17:06:57, and subsequent one-minute historical replay emissions. The C# callback forwards connection events into one price-status field. A status from another connection is a **C-tier hypothesis**, not a proven cause in this incident; source-specific connection identity is missing from the supplied feed frame.

**Proposal:** render TCP transport, reported price-feed status, historical replay and latest received live-update time separately. Give discrepancies an explicit explanation; do not turn the display green merely because any historical frame arrived. Verify single-source and multiple-connection cases before changing status aggregation.

## 3. Missing 17:02, 17:03 and 17:04 one-minute candles [A]

The reported span is 17:01–17:05 CT on September 6. Both endpoints exist. The API and persisted MNQ one-minute bars omit **three interior timestamps**, not four interior candles:

| Open time CT | Chart API | Persisted bars |
|---|---|---|
| 17:01 | present | present |
| 17:02 | missing | missing |
| 17:03 | missing | missing |
| 17:04 | missing | missing |
| 17:05 | present | present |

The four-minute jump between endpoint timestamps contains three absent minute slots. The chart's `/api/klines` NT8 branch reads the same FuturesBarsProvider cache used by the kernel. Therefore this is **not solely a chart layout/rendering problem**. API read and database read were separate snapshots; comparison is restricted to the already closed 17:00–17:12 window.

NinjaTrader reports MNQ 1m historical emissions of 2000 bars at 17:06:54, 2 bars at 17:06:55, and zero at 17:07:36 and 17:07:53. Counts do not reveal the individual timestamps carried. Go ingest summaries report zero drops on the measured queue path. Historical seeding merges timestamps and preserves existing-only rows; empty reseeds preserve history. Live `Upsert` ignores older-than-tail timestamps. These facts narrow investigation but **do not prove where the three minutes were first lost**.

**Remaining evidence required:** original NT8/provider minute rows for the missing interval and the emitted timestamp sequence, compared with accepted cache/persisted rows. Investigate historical cursor filtering, provider data availability, timestamp conversion, placeholder filtering and late live updates one at a time. Do not label any of those the root cause yet. No forced backfill or capture-reset endpoint was called in this review.

**Trading implication [B]:** calculations or confirmations using an incomplete minute window may be unreliable. Their actual impact must be measured by comparing the same calculation with authoritative recovered bars. Do not invent flat candles or assume a missing minute had zero volume.

**Proposal:** make coverage gaps visible separately from last-bar freshness. Test late arrivals, empty reseeds, recovery merges, open/close timestamp conversion and incomplete confirmation windows. Recovery should merge authoritative bars without deleting unaffected history. Report repair success only after the exact missing timestamps and OHLCV agree with the source.

## 4. Planner risk/reward and execution wording observations

All arithmetic below is from ASIA plan version 1, generated 16:39:24 CT, before the reported market-open gap. It is **not** a recomputed plan or evidence of profitability.

| Scenario | Entry | Stop | Selected target | Risk points | Selected-target R | First-target R |
|---|---:|---:|---:|---:|---:|---:|
| S1 long | 29468.25 | 29440.00 | 29534.33 | 28.25 | 2.339 | 2.000 |
| S2 long | 29526.00 | 29497.00 | 29657.38 | 29.00 | 4.530 | 0.287 |
| S3 short | 29657.38 | 29687.00 | 29534.33 | 29.62 | 4.154 | 3.984 |

S2's first two target-chain levels are 29534.33 and 29539.38, only 8.33 and 13.38 points above planned entry. A distant target yielding 4.53R does not establish that nearer levels can be traversed. The source plan labels the nearer supply C while retaining machine grade A. Its significance needs review; this report does not assume the downgrade is wrong. All R values exclude fees/slippage and use authored prices, not broker-accepted tick-rounded prices or actual fills.

The current plan API reports strict mode, but saved executor decision **37770** includes preferred/off-plan-permitted wording. This is an observed wording discrepancy, not proof that execution bypasses strict enforcement. Trace configuration resolution and the enforcement call site before proposing a behavioral change.

The Desk DAY row explicitly reports guardrail master OFF and no enforced daily limit; the separate risk API carries an environment limit and a derived armed flag with caveats. Do not mistake those nominal values for an active daily stop. No risk switch was changed.

## Proposed acceptance checks before any fix approval

1. Prepared arm with seam off: no order sent; UI names disabled placement, no broker-acknowledged claim.
2. Submitted, rejected, acknowledged and filled orders remain distinguishable; broker mismatch never silently passes a cutover.
3. Fresh TCP with disconnected price source; historical-only download; true live updates; stale live updates; multiple connection events produce distinct truthful states.
4. Exact 17:02–17:04 gap is reproducible in fixture data, with endpoint candles preserved. Late arrival and replay merge recover only authoritative rows without duplicating or deleting unrelated rows.
5. An incomplete minute window cannot be presented as complete without explicit coverage evidence; any change to trading gates is separately approved and tested.
6. Planner/executor policy wording is checked against the effective configuration, with no assumption that a text inconsistency proves an execution defect.
7. Proposed implementation ships as a separate reviewed diff with evidence; no new deployment is authorized by this report publication.

## Source references (immutable running revision)

- [Stop-entry placement switch](https://github.com/johnwick2921-cyber/nofx/blob/a3471b3f4cd1fd07989b28cfd2fa7490ffed327e/trader/armed_executor.go#L949)
- [Arm display](https://github.com/johnwick2921-cyber/nofx/blob/a3471b3f4cd1fd07989b28cfd2fa7490ffed327e/trader/desk_facts.go#L524)
- [Broker/ledger count comparison](https://github.com/johnwick2921-cyber/nofx/blob/a3471b3f4cd1fd07989b28cfd2fa7490ffed327e/trader/f12_leg4.go#L79)
- [Feed status versus usable live feed](https://github.com/johnwick2921-cyber/nofx/blob/a3471b3f4cd1fd07989b28cfd2fa7490ffed327e/provider/ninjatrader/tcp_server.go#L1161)
- [Chart data route](https://github.com/johnwick2921-cyber/nofx/blob/a3471b3f4cd1fd07989b28cfd2fa7490ffed327e/api/handler_klines.go#L74)
- [Historical merge](https://github.com/johnwick2921-cyber/nofx/blob/a3471b3f4cd1fd07989b28cfd2fa7490ffed327e/provider/ninjatrader/bar_cache.go#L220)
- [Live upsert ordering](https://github.com/johnwick2921-cyber/nofx/blob/a3471b3f4cd1fd07989b28cfd2fa7490ffed327e/provider/ninjatrader/bar_cache.go#L299)
- [NT8 connection status producer](https://github.com/johnwick2921-cyber/nofx/blob/a3471b3f4cd1fd07989b28cfd2fa7490ffed327e/ninjascript/VLTraderTCPClient.cs#L315)
