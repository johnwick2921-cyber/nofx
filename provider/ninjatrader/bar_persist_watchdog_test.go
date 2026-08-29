package ninjatrader

import (
	"os"
	"strings"
	"testing"
	"time"
)

// PRE-REOPEN F2 (2026-08-28) — the persist-stall watchdog env knob: default
// 60s, garbage → default, and a 10s floor so a mistyped value can't make the
// alarm scream every second.
func TestPersistWatchdogSeconds(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want int64
	}{
		{"unset defaults to 60", "", 60},
		{"explicit value", "45", 45},
		{"below floor falls back to default", "5", 60},
		{"negative clamps to default", "-20", 60},
		{"garbage falls back to default", "fast", 60},
		{"whitespace tolerated", " 90 ", 90},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.env == "" {
				_ = os.Unsetenv("PERSIST_STALL_WATCHDOG_S")
			} else {
				_ = os.Setenv("PERSIST_STALL_WATCHDOG_S", tc.env)
			}
			if got := persistWatchdogSeconds(); got != tc.want {
				t.Fatalf("persistWatchdogSeconds() = %d, want %d", got, tc.want)
			}
		})
	}
	_ = os.Unsetenv("PERSIST_STALL_WATCHDOG_S")
}

// S1 WIRE-UP (2026-08-29) — the BEHAVIOR fixture the half-ship escaped without:
// the alarm must FIRE (return the exact ERROR text) exactly once on a simulated
// 61s stall, stay silent in the same watchdog window (dedup via persistAlarmAt),
// and NOT re-alarm after the flush stamp advances. Existence tests can't catch
// a declared-but-unwired guard — this one proves firing.
func TestPersistWatchdogAlarmFiresOnceAndRecovers(t *testing.T) {
	_ = os.Setenv("PERSIST_STALL_WATCHDOG_S", "10")
	defer os.Unsetenv("PERSIST_STALL_WATCHDOG_S")

	now := time.Now().Unix()

	// Cold start (no flush ever) must never alarm.
	persistLastFlushAt.Store(0)
	persistAlarmAt.Store(0)
	if m := persistWatchdogAlarmAt(now); m != "" {
		t.Fatalf("cold start must not alarm, got %q", m)
	}

	// Simulated 61s flush stall → the alarm FIRES, with the loud ERROR text.
	persistLastFlushAt.Store(now - 61)
	a1 := persistWatchdogAlarmAt(now)
	if a1 == "" {
		t.Fatal("61s stall must fire the watchdog alarm")
	}
	if !strings.Contains(a1, "🔕 PERSIST WATCHDOG") || !strings.Contains(a1, "61s") {
		t.Fatalf("alarm text must carry the marker + duration, got %q", a1)
	}

	// Dedup: within the same watchdog window the alarm is silent (exactly once).
	if a2 := persistWatchdogAlarmAt(now + 5); a2 != "" {
		t.Fatalf("dedup must silence the same window, got %q", a2)
	}

	// Resumed flush → the stamp advances → the next window is clean.
	persistLastFlushAt.Store(now + 65)
	if a3 := persistWatchdogAlarmAt(now + 70); a3 != "" {
		t.Fatalf("resumed flush must not re-alarm, got %q", a3)
	}

	// A new stall AFTER recovery fires again (the guard is alive, not a one-shot).
	persistLastFlushAt.Store(now + 130)
	persistAlarmAt.Store(0)
	if a4 := persistWatchdogAlarmAt(now + 200); a4 == "" {
		t.Fatal("a fresh stall after recovery must re-fire the alarm")
	}
}
