package kernel

import (
	"os"
	"strconv"
	"strings"
)

// ResolvePlannerRetryMode (planner-speed wave 3.5, 2026-08-31) reads RETRY_MODE
// (repair|reauthor, default repair) — the one-click revert knob for the repair
// retry without a deploy.
func ResolvePlannerRetryMode() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("RETRY_MODE"))) {
	case "reauthor":
		return "reauthor"
	default:
		return "repair"
	}
}

// PlannerStreamIdleSeconds (planner-speed wave 4.2, 2026-08-31) reads
// AI_PLAN_STREAM_IDLE_SECS (default 30) — the idle-chunk deadline for the
// planner's SSE stream. The whole-request ceiling stays http.Client.Timeout
// (600s): a live-but-slow stream is never killed, a stalled one dies in 30s.
func PlannerStreamIdleSeconds() int {
	d := 30
	if v := strings.TrimSpace(os.Getenv("AI_PLAN_STREAM_IDLE_SECS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			d = n
		}
	}
	return d
}
