package mcp

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// E1 — THE HONESTY PIN. Every field on the boot line is compared against the
// function that ENFORCES it. No literal appears in this test except the field
// NAMES. On the class-41 line this fails: watchdog_log, keepalive,
// serialize_executor and resend_identical were string literals, and its
// fixture asserted those same literals, so the line could drift from reality
// undetectably.
func TestClass46BootLineEveryFieldReadFromEnforcer(t *testing.T) {
	line := PlannerClientPolicyLine()

	want := map[string]string{
		"tries":            fmt.Sprintf("%d", StreamRetryTries()),
		"keepalive_set":    DialerKeepAliveSet().String(),
		"resend_identical": fmt.Sprintf("%t", ResendIdenticalOnProviderFailure()),
		"serialize":        fmt.Sprintf("%t", SerializeExecutorDuringPlannerStream()),
		"storm_cap":        fmt.Sprintf("%d", StormCapPerRead()),
		"trace":            fmt.Sprintf("%t", TransportTraceEnabled()),
	}
	for field, v := range want {
		token := field + "=" + v
		if !strings.Contains(line, token) {
			t.Errorf("field %q must be READ from its enforcer: want %q in\n%s", field, token, line)
		}
	}
	// backoff is the schedule, joined from the resolver.
	sched := StreamRetryBackoffSchedule()
	parts := make([]string, 0, len(sched))
	for _, d := range sched {
		parts = append(parts, d.String())
	}
	if !strings.Contains(line, "backoff="+strings.Join(parts, "→")) {
		t.Errorf("backoff not read from StreamRetryBackoffSchedule():\n%s", line)
	}
	// watchdog carries BOTH resolved timers and says it measures data.
	if !strings.Contains(line, fmt.Sprintf("watchdog=pre%ds/post%ds(data)", WatchdogPreTokenSeconds(), WatchdogPostTokenSeconds())) {
		t.Errorf("watchdog fields not read from their resolvers:\n%s", line)
	}
	// The observed keepalive is honest about being unobserved.
	if _, ok := ObservedKeepAlive(); !ok && !strings.Contains(line, "observed=n/a") {
		t.Errorf("an unobserved keepalive must print n/a, never the set value:\n%s", line)
	}
}

// The line must MOVE when the enforcers move. A literal cannot do that.
func TestClass46BootLineTracksTheResolvers(t *testing.T) {
	base := PlannerClientPolicyLine()
	t.Setenv("AI_PLAN_STORM_CAP", "9")
	t.Setenv("AI_PLAN_WATCHDOG_POST_SECS", "45")
	t.Setenv("AI_PLAN_STREAM_TRIES", "4")
	t.Setenv("AI_PLAN_TRACE", "off")
	moved := PlannerClientPolicyLine()
	if moved == base {
		t.Fatal("the boot line did not change when its resolvers did — a field is a literal")
	}
	for _, want := range []string{"storm_cap=9", "post45s(data)", "tries=4", "trace=false"} {
		if !strings.Contains(moved, want) {
			t.Errorf("want %q in\n%s", want, moved)
		}
	}
}

// A24 — no bare "on"/"off" words on the line: those were the class-41 tell.
func TestClass46BootLineHasNoLiteralOnOff(t *testing.T) {
	line := PlannerClientPolicyLine()
	for _, banned := range []string{"watchdog_log=on", "serialize_executor=off", "resend_identical=on", "keepalive=30s"} {
		if strings.Contains(line, banned) {
			t.Fatalf("class-41 literal survived: %q in\n%s", banned, line)
		}
	}
	// Every "field=value" pair must have a non-empty value.
	for _, m := range regexp.MustCompile(`(\w+)=(\S*)`).FindAllStringSubmatch(line, -1) {
		if m[2] == "" {
			t.Fatalf("field %q printed an empty value:\n%s", m[1], line)
		}
	}
}

// Resolvers clamp instead of trusting junk (A11).
func TestClass46PolicyResolversClamp(t *testing.T) {
	t.Setenv("AI_PLAN_STORM_CAP", "0")
	if StormCapPerRead() != 5 {
		t.Fatalf("storm cap 0 must fall back to the default, got %d", StormCapPerRead())
	}
	t.Setenv("AI_PLAN_STORM_CAP", "999")
	if StormCapPerRead() != 5 {
		t.Fatalf("storm cap 999 is out of range and must fall back, got %d", StormCapPerRead())
	}
	t.Setenv("AI_PLAN_WATCHDOG_PRE_SECS", "junk")
	if WatchdogPreTokenSeconds() != 600 {
		t.Fatalf("junk must fall back, got %d", WatchdogPreTokenSeconds())
	}
	// The pre-token limit must not exceed DeepSeek's ~10 min queue close.
	if WatchdogPreTokenSeconds() > 1200 {
		t.Fatal("pre-token limit exceeds the resolver's own ceiling")
	}
}
