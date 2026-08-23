package kernel

import (
	"fmt"
	"os"
	"strconv"
)

// G4 (regime wave, 2026-08-21) — TRANSITION STAND-DOWN: when G2 emits a
// counter-trend CHoCH/MSS on the plan's bias TF (15m) that the plan's flip
// has NOT yet confirmed, the plan enters TRANSITION and NEW entries in the
// plan's own direction are paused until (a) the flip/death confirms (the
// planner re-plans — the normal path), (b) structure re-confirms the original
// trend (BOS resumption), or (c) TRANSITION_MAX_MIN expires. Entries in the
// COUNTER direction are NOT enabled by this — the flip does that job.
// FAIL-OPEN: no plan / no structure snapshot → the state simply never opens.
//
// OFF = today's pre-wave behavior (Studio toggle transition_standdown,
// default ON).

const DefaultTransitionMaxMin = 45

// TransitionMaxMin resolves the stand-down cap (env TRANSITION_MAX_MIN,
// default 45 minutes).
func TransitionMaxMin() int {
	if v := os.Getenv("TRANSITION_MAX_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return DefaultTransitionMaxMin
}

// TransitionState is the trader-maintained stand-down state. Runtime-only
// (mirrored to system_config for the plan card's chip).
type TransitionState struct {
	Active      bool   `json:"active"`
	Dir         string `json:"dir"`    // plan bias: "long" | "short" — entries in THIS direction pause
	Detail      string `json:"detail"` // "CHoCH-up 15m @29470.25 10:45"
	SinceMs     int64  `json:"since_ms"`
	PlanID      string `json:"plan_id,omitempty"`
	PlanVersion int    `json:"plan_version,omitempty"`
}

// TransitionStanddownVerdict is the pure gate: (blocked, refusal message).
// action is open_long/open_short; dir is the plan's bias direction; active +
// detail come from the trader's TransitionState. Empty refusal = pass.
func TransitionStanddownVerdict(action string, active bool, dir, detail string) (bool, string) {
	if !active {
		return false, ""
	}
	d := ""
	switch action {
	case "open_long":
		d = "long"
	case "open_short":
		d = "short"
	default:
		return false, ""
	}
	if d != dir {
		return false, "" // counter-direction entries are never paused by this
	}
	return true, fmt.Sprintf("transition_standdown: %s paused — unconfirmed %s, awaiting flip or resumption", d, detail)
}
