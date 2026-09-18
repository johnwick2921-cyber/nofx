package trader

import (
	"nofx/store"
)

// DeathRereadBootLine (W-DEATH-REREAD, 2026-09-18) renders the resolved knob at
// trader load — the moment the strategy config exists. READ from the resolver,
// never from a file default: nil = ON (the owner's 12:3x CT "fix all" default),
// a saved true = ON(saved), an explicit false = OFF (today's behaviour).
func DeathRereadBootLine(dp *store.DayPlanConfig) string {
	if dp == nil || dp.DeathReread == nil {
		return "🧬 death→reread=on(default) (W-DEATH-REREAD)"
	}
	if *dp.DeathReread {
		return "🧬 death→reread=on(saved) (W-DEATH-REREAD)"
	}
	return "🧬 death→reread=off (W-DEATH-REREAD)"
}
