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
	logger.Infof("🚀 planner speed wave (2026-08-31): retry=%s stream=on stream_idle=%ds stream_total=%ds (class 37: planner ceiling split from the HTTP ceiling) ttfb=on — reasoning/mode/cap unchanged until owner ruling",
		ResolvePlannerRetryMode(), PlannerStreamIdleSeconds(), PlannerStreamTotalSeconds())
	logger.Infof("🗓 session reads (owner ruling 2026-08-31, open−30): ASIA 16:30 · LONDON 01:30 · NY 08:00 CT — windows/flats unchanged; Sunday weekly 16:30 → ASIA follows")
	logger.Infof("🪢 netting-orphan wave (class 27, 2026-08-31): netting-flat cancels brackets (C# sweep + Go desync cancel_order) · exit reconstruction from the netting fill (never exit=entry; unknown → UNRESOLVED+alarm) · dedupe 577/578 class · one-live-arm guard (opposite-side refused while open; kind=exit escapes) · split legs > capacity rejected (capacity=1 unless max_contracts_per_order raises)")
	// 0C shadow demotion (owner ruling 2026-08-31): render the RESOLVED map —
	// process level sees defaults+env; the per-trader resolved map prints at the
	// trader's first arm cycle (class-8: resolved, never a literal).
	logger.Infof("%s (process-level: defaults+env; per-trader resolved map prints at first arm cycle)", ConditionStatusLedger(nil, nil, ShadowConditionsEnv()))
	// CLASS 34 (owner ruling 2026-08-31): every validator hint's condition
	// tokens must be legal AND live. The table test is the hard build gate;
	// this boot re-check makes a broken registry impossible to miss live.
	if err := ValidateValidatorHints(); err != nil {
		logger.Errorf("🧪 validator hints BROKEN: %v — a hint names an unknown/shadowed condition or a rule token illegal in its own field (class 34 + 38 guard)", err)
	} else {
		logger.Infof("🧪 validator hints: %d sites — every condition token legal + live, every rule token in its own field enum (class 34 + 38 guard)", len(ValidatorHints()))
	}
	// CLASS 38 (2026-09-01) — every validator restriction keyed by
	// condition/field must be STATED in the rendered prompt. The prompt offering
	// what the validator refuses is the whole defect class (rows 78/79/80).
	logger.Infof("%s", PromptContractBootLine())
	// CLASS 39 (owner ruling 2026-09-01) — normalize, don't reject: legs on a
	// non-sweep condition collapse to the single top-level arm with a WARN;
	// sweep_reclaim keeps its split contract untouched; never synthesize.
	logger.Infof("⚖ arm normalizer: legs on non-sweep → single arm + WARN (class 39); sweep_reclaim contract unchanged; counter arms_normalized_class39 recorded in system_config")
	logger.Infof("🛡 cutover safety (class 33): flat gate legs=5 (db_open_positions · api_positions · nt8_positions_snapshot · working_orders · planner_in_flight) via GET /api/cutover-gate; leg4 reads the armed_orders LEDGER (was a stub returning empty — passed vacuously at cutovers 35→41); boot sweep cancels pre-boot arms before ANY re-arm, counter arms_boot_swept_class33")
}
