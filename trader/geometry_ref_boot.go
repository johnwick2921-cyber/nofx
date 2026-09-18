package trader

import (
	"fmt"

	"nofx/store"
)

// GeometryRefBootLine (W-GEOMETRY-REFUSAL, 2026-09-18) is the per-trader boot
// line for day_plan.geometry_reference_levels — READ from the resolved knob:
// the owner ruled the contract fix ON by default ("both fix now", 08:1x CT);
// explicit false = OFF (today's behaviour byte-identical). The process-level
// 🎛 entry law line prints n/a because the knob is per-strategy.
func GeometryRefBootLine(dp *store.DayPlanConfig) string {
	state := "on(default)"
	if !dp.GeometryRefIDsEnabled() {
		state = "off"
	}
	return fmt.Sprintf("🎛 geom_ref_ids=%s (day_plan.geometry_reference_levels; W-GEOMETRY-REFUSAL)", state)
}
