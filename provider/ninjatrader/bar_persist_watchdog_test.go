package ninjatrader

import (
	"os"
	"testing"
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
