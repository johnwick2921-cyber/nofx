package kernel

import (
	"nofx/logger"
)

// LogVolumeWaveBoot (Pack B, owner override 2026-08-26) — the boot-line ledger
// for the volume wave: one line per shipped knob so the boot block
// self-documents the cutover (dispatch: boot adds these).
func LogVolumeWaveBoot() {
	logger.Infof("🎛 volume wave: detectors=on · seats=%d · proximity=%.1f · family-confluence(cap=%d) · zone-ladder=1.0/0.6/0.3/0.15 · roles=on(overrides=%v) · bias_ctx=on",
		DefaultMaxLevels, ActivationWindowK, ConfluenceCap(), IsRoleOverridden())
}
