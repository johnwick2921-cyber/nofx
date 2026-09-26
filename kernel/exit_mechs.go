package kernel

import (
	"os"
	"strings"
)

// ExitMechsSuspended is the ONE canonical reader for the EXIT_MECHS_SUSPENDED
// posture (0B, 2026-09-02): when the env is set to 0/false/off/no the breakeven
// and ATR-trailing mechanisms are LIVE at the wire; otherwise (unset or any
// other value) they are SUSPENDED — triggers still evaluate but no move_stop
// frame is sent. Moved here from trader/exit_mechs_suspend.go (FIX-KNOBS P2-1,
// 2026-09-26) because the RENDERED system prompt (kernel/engine_prompt_futures.go)
// must READ the same value the mechanics gate on — one canonicalizer, called
// where the value enters; the trader-side call sites delegate to this.
func ExitMechsSuspended() bool {
	if v := strings.TrimSpace(os.Getenv("EXIT_MECHS_SUSPENDED")); v != "" {
		switch strings.ToLower(v) {
		case "0", "false", "off", "no":
			return false
		}
	}
	return true
}
