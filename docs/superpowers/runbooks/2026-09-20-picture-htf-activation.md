# Two-picture (W-PICTURE-HTF) — owner-attended activation runbook

Wave: branch `fix/picture-htf` (pushed). Mode: deterministic 4H-pivot →
H1-close-break → 5m-swing entry with protective bracket, SIM-only. The AI is
commentary only. Nothing here runs unattended — every step below except the
NT8 compile is the bot operator's (owner's) action or explicitly acked.

## Step 0 — preconditions (verify BEFORE anything)

- `git log -1 origin/dev` = the dev tip this wave was tested against (report
  quotes it). If dev moved, re-merge/re-test at the merged head first.
- Bot flat gate: no open positions in the store, no working arms, NT8
  snapshots flat on both accounts.
- Backup the DB: the systemd timer does it daily; take one explicitly:
  `python3 -c "import sqlite3; s=sqlite3.connect('file:/home/hoang/nofx/data/data.db?mode=ro',uri=True); d=sqlite3.connect('/home/hoang/nofx-backups/manual/pre-picture.db'); s.backup(d)"`
- Strategy config quoted before touching anything: trader → strategy binding,
  `risk_control.min_risk_reward_ratio` (live: 2.0), `day_plan.picture_htf`
  (live: absent = resolved defaults once enabled).

## Step 1 — AddOn backup + copy (the C# half)

The AddOn compiles ONLY inside NT8 on Windows. Before copying, back up the
currently-running DLL source:

```bash
# 1. backup what NT8 is running today (WSL view of the NT8 AddOns folder)
mkdir -p ~/nofx-backups/addon/$(date +%Y%m%d-%H%M%S)
cp "/mnt/c/Users/hoang/Documents/NinjaTrader 8/bin/Custom/AddOns/"*.cs \
   ~/nofx-backups/addon/$(date +%Y%m%d-%H%M%S)/

# 2. copy the new sources (from the DEPLOYED tree, not the worktree)
cp /home/hoang/nofx/ninjascript/*.cs \
   "/mnt/c/Users/hoang/Documents/NinjaTrader 8/bin/Custom/AddOns/"

# 3. verify the copy
md5sum /home/hoang/nofx/ninjascript/*.cs \
       "/mnt/c/Users/hoang/Documents/NinjaTrader 8/bin/Custom/AddOns/"*.cs
```

Then IN NT8: **F5 compile**, then a **full NT8 restart** (AddOns do NOT
hot-reload). Until the heartbeat reports build `2026-09-20-p1`, the Go side
prints `addon=not proven` and the mode refuses every evaluation — by design.

## Step 2 — merge + merged-head suite

1. Merge `fix/picture-htf` into `dev` (owner/CTO-gated PR).
2. At the MERGED head: `go test ./... -count=1` and `cd web && npm run
   build` + `npx vitest run` — a branch green alone is not green merged.

## Step 3 — the cutover (owner-attended, no timers)

1. Acquire the main-tree lock: `deploy/nofx-lock.sh acquire <session> "<task>"`.
2. Five-leg gate, all on FRESH broker evidence: flat store · flat NT8
   snapshots · no working arms · API positions [] · owner present and acking.
3. Build the binary from a clean clone at the merged head (vcs.revision
   stamped, not `<no-vcs>`): the release chicken-and-egg = commit code S →
   stash `deploy/RELEASE` + `web/src/guide/types.ts` → build binary → pop →
   `sed` GUIDE_BUILT_REV + deploy/RELEASE = S → `npm run build` in web/ →
   commit artifacts.
4. Swap (mv old → `nofx-bin.old.<prev>`) then `kill -9 <pid>` (SIGTERM exits
   0 and does NOT relaunch; systemd `Restart=on-failure` boots the new one).
5. Read the boot lines from the journal, ALL of:
   - `🔐 BOOT INTEGRITY OK — rev <merged sha> · goldens PASS`
   - `📷 picture-htf: mode=… rule=v1 SIM-only data=native NT8 bars
     (final+emitted_at) addon=proven|not proven (build=…, need ≥ 2026-09-20-p1)`
   - UI bundle matches binary (no drift banner).
6. Flat-gate re-check post-boot. Push the post-boot marker BEFORE releasing
   the lock. Rollback = restore `nofx-bin.old.<prev>` + release + restart.

## Step 4 — enable the mode

Strategy Studio → Day Plan → Picture HTF → toggle ON (knobs blank = resolved
defaults; min R:R blank = inherit the risk-control floor, live 2.0). Quote
the saved config again after saving.

## Step 4.5 — native 4H data readiness (subscription config is NOT proof)

`defaultAutoBarsTimeframes` including `"4h"` proves only that the
SUBSCRIPTION was requested. Before relying on the mode, verify RECEIVED
data:

1. **Received native 4H bars** — the bars store must hold native `4h` rows
   for the live contract (query by symbol+tf). The replay only ever had a
   disclosed ETH-grid proxy because no native 4h rows existed in the store.
2. **Sufficient completed history** — at least the pivot window (default
   120) of COMPLETED 4h bars ending at the latest completed bar, so the
   level picture is real rather than bootstrapped from a few frames.
3. Quote the counts in the post-activation evidence (Step 5): newest 4h open
   time, completed-4h count in the window, freshest 4h bar age.

Until 1–2 hold, treat the mode as data-unready regardless of what the boot
line prints.

## Step 5 — evidence to collect after activation (report these back)

1. The received AddOn build id on the heartbeat (`build=2026-09-20-p1` in the
   boot line / `AddonBuildLine`).
2. The actual Go boot output lines (integrity, picture boot line, UI match).
3. The selected strategy mode + saved knob values.
4. Controlled SIM entry + protection receipts: from a natural qualifying
   setup (or a labeled NT8 replay through the LIVE subscription — which the
   sink treats as live? NO: an NT8 replay is historical and never fans out;
   a controlled entry therefore requires a real live setup or the owner's
   explicit test-seam command). Quote the opportunity row (intended
   geometry), the signal frame, and the received order_update/fill frames
   showing the bracket legs.
5. Native 4H readiness counts from Step 4.5: newest 4h open time, completed
   4h count in the pivot window, freshest 4h bar age.
6. Natural-market setup evidence is reported SEPARATELY. If none occurs,
   state that it remains pending.
