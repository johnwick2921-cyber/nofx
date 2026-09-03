# NO-TRADE BAND (checklist class 51) — session-scoped, config-driven, judged at read time

**Wave:** no-trade-band · worktree `~/nofx-band` off `origin/dev` 0bba743b
**Commits:** 52903708 · c972075d · 091f83ee · d0b6e594 · fa15f2d9 (+ this report)
**Boot line:** `🗓 no-trade band: …` (rendered below, every field resolved)
**Executor diff:** none. No gate, arm, size or order path changed behaviour.

---

## 1. What was wrong

**The pin.** The ASIA card at 23:00 CT rendered three no-trade rules. None of
them could refuse an entry.

| shown as a live rule | what it actually was |
|---|---|
| first 5m after the open | closed at 17:05 — six hours earlier |
| lunch 12:00–13:30 CT | an NY window; ASIA runs 17:00–02:00 and never touches it |
| 🔴 ISM PMI 09:00 CT ±15m | fired fourteen hours before the card was opened |

The card was rendering `doc.no_trade`: prose the model wrote when the plan was
authored, stored, and displayed verbatim for the rest of the session. Nothing
in the path asked what time it was.

**The same defect's second face.** Each fixed window was defined more than once
and the copies could not disagree loudly:

| copy | where | consumed by |
|---|---|---|
| `cur < start+5` | `trader/auto_trader_session.go` `inSessionFirst5m` | the entry gate |
| `InBlackoutWindow(t, "12:00", "13:30")` | same file | the entry gate |
| its own first-5m test + `"12:00"`/`"13:30"` | `kernel/adherence.go:121` | the adherence grader |
| `"first 5m (CT)"`, `"12:00-13:30 CT lunch"` | `kernel/planner_prompt.go` schema example + rule sentence | the model |
| whatever the model wrote | the plan doc | the card |

Five statements of two windows. The gate refused one, the grader scored
another, the prompt taught a third and the card claimed a fourth. No test
could fail if any pair drifted, because nothing compared them.

**And a claim that was never measured.** `WidenCTWindows` appended
`+%dm (clock drift)` whenever `driftWidenMinutes` returned nonzero, and that
function rounds ANY nonzero offset up to at least one minute. This machine's
NTP offset is 108 ms — healthy. Every T1 line in every plan therefore carried
"+1m (clock drift)", baked into the stored text at authoring time, telling the
reader for the rest of the day that the machine's clock was drifting.

---

## 2. What changed

### 2.1 One definition each — gate, grader and card read the same window

`kernel/no_trade_band.go`:

```go
func FirstNoTradeMinutes() int                        { return 5 }
func LunchWindowCT() (startCT, endCT string)          { return "12:00", "13:30" }
func InFirstNoTradeMinutes(startCT string, t time.Time) bool
func InLunchNoTrade(t time.Time) bool
```

The gate (`sessionGateDecision`) and the grader (`adherence.go:121`) now call
these. Every payload built from them is stamped `source: "code-constant"`, so
the surface never implies a knob that does not exist. **These are not
owner-configurable and the card says so.**

A parity test asserts the gate, the grader and the card return the same answer
for the same clock across seven boundary minutes.

### 2.2 The machine writes the windows; the model does not

The plan doc gains an additive `no_trade_windows` field. It is written at plan
time from enforcing sources only:

| kind | source | comes from |
|---|---|---|
| `first_n` | `session-registry` | `BuildMachineNoTradeWindows(sess)` — the gate's own definition, keyed on the session's real open |
| `lunch` | `code-constant` | same |
| `t1` | `calendar` | `at.t1WindowsFor(tradeDate, sess)` |

`t1WindowsFor` was extracted from `currentT1Windows` for one reason: the old
function keyed on *the session active at `now`*, and a 16:55 read authors ASIA
while NY is still live. The plan card now carries **literally the windows the
entry gate will refuse inside** — clock widening and the P0.6 fail-closed
static fallback included.

`T1NoTradeWindowsFromCT` takes already-resolved `CTWindow`s rather than raw
calendar events on purpose. The widening decision is made once, by the
enforcer. A second computation in the kernel could disagree with the gate, and
a card that disagrees with the gate is the defect this wave exists to fix.

**The model's prose is untouched.** It is stored exactly as authored and now
renders under "Model notes".

### 2.3 Read-time evaluation

`EvaluateNoTradeWindows(wins, nowMin, sessionStartMin, eodMin)` stamps each
window `live` / `elapsed` / `other_session`, asking two independent questions:

1. **Does the window touch this session's tradeable span at all?** Geometry,
   measured from the session start so a session that wraps midnight stays
   monotonic. NY lunch on an ASIA card fails here, whatever the clock says.
2. **If it does, has it finished?** Measured from now on a signed ±12h axis —
   offsets-from-session-start alone cannot tell an 08:00 pre-open read from a
   16:00 post-close one, and the first draft of this function got exactly that
   wrong until the NY pre-open pin caught it.

`GET /plan/today` gains `no_trade_band`: every window with its status and its
CT bounds resolved server-side, so the card does no clock arithmetic. A doc
written before this wave has no machine windows, returns nil, and the card
keeps rendering prose as rules rather than claiming a plan with no limits.

### 2.4 The drift claim, fixed at the measurement

The widening is **byte-identical for every input**. Even a 108 ms offset can
carry an event across a minute boundary, so the one-minute guard is real
protection and the enforced blackout at `auto_trader_calendar.go:185` is
unchanged. Only the claim moved: `driftIsSkew` gates the label at 60 s, the
point where the offset alone can shift an event by a whole minute. Below that
the extra minute is silent boundary rounding.

| offset | widening | label |
|---|---|---|
| 108 ms (this machine) | ±1 min | none |
| 90 s | ±2 min | `+2m (clock drift)` |

### 2.5 The prompt

The `no_trade` schema example and its rule sentence are generated from
`FirstNoTradeMinutes()` and `LunchWindowCT()`. The sentence now also tells the
author the windows are enforced whether or not it lists them, so what it writes
is read as notes.

---

## 3. The boot line

```
🗓 no-trade band: first_n=5m lunch=12:00–13:30 (source=code-constant, shared by gate+grader+card) · T1 taken from the enforcing gate at plan time (widening and fail-closed fallback included) · every window judged against the reader's clock, not the write clock · model prose renders as notes
```

Every field resolved: `first_n` from `FirstNoTradeMinutes()`, the bounds from
`LunchWindowCT()`, the source from the constant that stamps the payload.

---

## 4. Tests

| id | what it pins | RED before |
|---|---|---|
| parity | gate, grader and card agree across 7 boundary minutes | — (shipped 52903708) |
| F1 | ASIA card at 23:00 → 0 live; same doc at 17:02 → first-5m live; pre-band doc → nil | the card had no band to evaluate |
| F2 | NY pre-open read at 08:00 → first-5m + lunch + the 10:00 T1 live; the 07:30 T1 not live | evaluator returned 0 live |
| F3 | a window straddling now reads live | evaluator returned elapsed |
| F4 | literal scan: exactly one copy of each bound at the definition site, none in the prompt or the grader | grader held its own `"12:00"`/`"13:30"` |
| F5 | 108 ms → no drift claim, widening kept; 90 s → claim + 2 min | label read `+1m (clock drift)` |
| F6 | T1 band entries map the enforcer's windows, bounds untouched | — |
| class-38 row | the rendered contract states the resolved windows and their independence | — |
| write-site | the machine stamps the doc; the model's prose is unmodified | **quoted RED**: "the machine wrote NO no_trade_windows" |
| FE ×5 | none-live at 23:00, count of 3, expansion with reasons, live rule at 17:02, pre-band fallback | the block joined prose unconditionally |

Suites: `kernel` `trader` `api` `market` green; vitest 39 files / 303 tests green.

---

## 5. E6 — lunch vs ny_pm overlap (REPORT ONLY, no change)

Measured from the live registry:

```
NY     window 08:30–14:45
        killzone ny_am        08:30–11:00
        killzone ny_pm        13:00–14:45   ⚠ OVERLAPS LUNCH
        band     first_n      08:30–08:35 (code-constant)
        band     lunch        12:00–13:30 (code-constant)
```

`ny_pm` opens at 13:00; lunch refuses entries until 13:30. Thirty minutes of
every NY afternoon are advertised as a prime hunting window and cannot trade.

**The gate is safe.** In `sessionGateDecision` the lunch refusal sits above any
killzone consideration, so no entry fires before 13:30 regardless. The defect
is presentational: the registry and the guide promise a window that is a third
dead. **[A]** — read from the running registry and the gate's own ordering.

Two honest resolutions, owner's call: move `ny_pm` to 13:30, or state the
overlap on the surface. Neither is in this wave.

---

## 6. Known and untouched

- **`kernel/plan_render.go:169` renders `doc.NoTrade` into the plan text.**
  Whatever reads that rendering sees the model's prose, including windows that
  have elapsed — the same lie this wave fixed on the card, on a different
  surface. Out of this wave's footprint (zero executor diff); worth its own
  look. **[A]**
- **Lunch is built for ASIA and LONDON too.** It never intersects their
  windows, so the card collapses it as `other_session`. This mirrors the gate
  exactly: `InLunchNoTrade` is session-independent, so lunch is only ever
  effective during NY. **[A]**
- **`GUIDE_BUILT_REV`** is bumped in the deploy marker commit, per the
  established pattern (the Go binary is built at the content rev; the web
  bundle is built after the marker sets the rev).

---

## 7. Cutover

Not deployed. Preconditions, in order: the wave ahead of this one boots by
name, then `GET /api/cutover-gate` reads `ready: true` on all five legs with no
arm resting, then the owner's explicit GO. A19 all three halves: `deploy/RELEASE`
written in `~/nofx` before the kill, the marker committed from that same tree
after the boot line is observed, and the running binary preserved as
`nofx-bin.old.<its own rev>` first.

---

## 8. BLOCKED — the main tree is mid-cutover under a dead lock (A23: measured, not touched)

Measured at 22:12 CT, read-only. I did not touch `~/nofx`.

| fact | value |
|---|---|
| `~/nofx-main.lock` | `owner=weekly-refs-deploy pid=976198 expiry=1788410353` |
| PID 976198 | **DEAD** (`kill -0` fails); the lock does not expire for ~87 more minutes |
| main tree HEAD | `1cee77a8` (class-50b), 7 commits ahead of `origin/dev`, **unpushed** |
| `deploy/RELEASE` | `1cee77a8` — **uncommitted** (` M deploy/RELEASE`) |
| running process | PID 1093919, `vcs.revision=56904ec1`, up 37 min |
| `nofx-bin` | 56904ec1 (unchanged) |
| `nofx-bin.next` | clean build of `1cee77a8`, `vcs.modified=false`, built 21:39 |
| `nofx-bin.old.56904ec1…` | preserved 21:31 |

**The hazard.** `RELEASE` claims `1cee77a8` while the process runs `56904ec1`.
`kernel.AssertBootIntegrity` reads `deploy/RELEASE` relative to the unit's
`WorkingDirectory=/home/hoang/nofx`, prefix-matches it against the embedded
revision, and on mismatch latches `tradingRefused`. With `Restart=on-failure`,
**any crash or host restart right now boots the bot into TradingRefused.** This
is the exact failure mode recorded on 2026-09-02 07:32. **[A]** — read from the
lock file, `ps`, `go version -m`, the unit and `boot_integrity.go`.

The dead dispatch got as far as: build `.next`, preserve the old binary, write
`RELEASE` and `GUIDE_BUILT_REV`. It never swapped and never killed.

**Two safe resolutions, both the owner's call.** Finish that cutover (`mv` the
staged `.next` in, kill, observe its boot line, commit `RELEASE` from the main
tree), or revert `deploy/RELEASE` and `web/src/guide/types.ts` to `56904ec1` so
the file matches the binary that is actually running. I have done neither.

**Consequences for this wave.** My branch is off `origin/dev` (`0bba743b`) and
does not contain the seven class-50 commits. `git merge-tree` against
`1cee77a8` reports **zero conflicts**, so the rebase is clean whenever their
wave boots. The checklist collision is already resolved: they took 51, this
wave is 52.
