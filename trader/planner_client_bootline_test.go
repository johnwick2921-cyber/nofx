package trader

import (
	"strings"
	"testing"
)

// Class 37 (C7) — the per-trader boot line carries the RESOLVED planner client
// config: provider row, both stream deadlines, the HTTP ceiling that still
// governs non-stream paths, retries/backoff and the planner cap.
func TestClass37PlannerClientBootLineWording(t *testing.T) {
	line := plannerClientBootLine("8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek", 30, 1200, 600, 2, 2, 65536)
	for _, want := range []string{
		"🛰 planner client:",
		"provider_row=8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek",
		"stream_idle=30s",
		"stream_total=1200s (AI_PLAN_TOTAL_DEADLINE_SECS)",
		"http_ceiling=600s (non-stream paths only)",
		"retries=2 backoff=2s cap=65536",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("boot line missing %q: %s", want, line)
		}
	}
	if strings.Contains(strings.ToLower(line), "sk-") {
		t.Fatalf("boot line must never carry a key fragment: %s", line)
	}
}

func TestClass37PlannerStreamTotalResolvedInTrader(t *testing.T) {
	t.Setenv("AI_PLAN_TOTAL_DEADLINE_SECS", "")
	t.Setenv("AI_PLAN_STREAM_IDLE_SECS", "")
	if got := plannerStreamTotal().Seconds(); got != 1200 {
		t.Fatalf("plannerStreamTotal = %.0fs, want 1200", got)
	}
	if plannerStreamTotal() <= plannerStreamIdle() {
		t.Fatalf("total must exceed idle")
	}
}

// CLASS 46 — the class-41 stream-policy boot line and its fixture are GONE.
// Both asserted the same string literals (watchdog_log=on, keepalive=30s,
// serialize_executor=off, resend_identical=on), so the pair could only ever
// agree with each other, never with reality. The replacement lives in
// mcp/planner_policy_test.go, where every field is compared against the
// function that enforces it.
