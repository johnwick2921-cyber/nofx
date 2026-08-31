package kernel

import (
	"nofx/logger"
)

// LogVolumeWaveBoot (Pack B, owner override 2026-08-26) — the boot-line ledger
// for the volume wave: one line per shipped knob so the boot block
// self-documents the cutover (dispatch: boot adds these). S-wave (2026-08-26):
// proximity is per-trader config now (owner retune 0.3 → ±~105pt), so the boot
// line states the resolver instead of a constant.
func LogVolumeWaveBoot() {
	logger.Infof("🎛 volume wave: detectors=on · seats=%d · proximity=cfg(resolved per-trader; retuned 0.3) · family-confluence(cap=%d) · zone-ladder=1.0/0.6/0.3/0.15 · roles=on(overrides=%v) · bias_ctx=on · tier1+=VAH/VAL/SETT/nPOC (R-A13)",
		DefaultMaxLevels, ConfluenceCap(), IsRoleOverridden())
	logger.Infof("🎯 touch telemetry: band=%dt(%.1fpt) max_bars=%d vol_lookback=%d approach=%d — advisory, zero gates",
		TouchBandTicks(), TouchBandPoints(), TouchEpisodeMaxBars(), TouchVolLookback(), TouchApproachBars())
	logger.Infof("📐 fvg_entry: on min_disp=%.1f×ATR ce_width=%.0fpt lookback=%d bars — advisory, zero gates",
		FvgEntryMinDispATR(), FvgEntryCEWidthPts(), FvgEntryLookbackBars())
	logger.Infof("🔧 S-wave (2026-08-26): stale_confirm=%.1f×ATR5m · eod_flat=session-end (NY 14:45 CT, R-A15)",
		StaleConfirmATR())
	logger.Infof("📜 planner playbook (2026-08-26): playbook=v2 bias_tree=on chain_after=on no_trade_gates=on killzone_weights=on stop_doing=on — ALL ADVISORY, zero new gates")
	logger.Infof("🛡 plan facts guards: 0-levels-on-a-side + empty machine map = fail-closed (2026-08-18 pathology guards) — per-side counts REMOVED entirely (owner ruling 2026-08-31; no quota, no WARN, no thin_side note)")
	logger.Infof("🚀 planner speed wave (2026-08-31): retry=%s stream=on stream_idle=%ds ttfb=on — reasoning/mode/cap unchanged until owner ruling",
		ResolvePlannerRetryMode(), PlannerStreamIdleSeconds())
}
