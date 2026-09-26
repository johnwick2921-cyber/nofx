package kernel

import (
	"strings"
	"testing"

	"nofx/store"
)

// P2-1 (FIX-KNOBS, DS-105, audit 2026-09-26): the RENDERED prompt must state the
// LIVE exit-mech posture, read from the env the mechanics actually gate on —
// never let a stale saved prompt_section (which claims "SUSPENDED at the wire")
// be the only truth the model sees. The live box runs EXIT_MECHS_SUSPENDED=0
// (mechanics ACTIVE) while the saved trading_frequency text still says
// SUSPENDED.
func TestFuturesPromptStatesTheLiveExitMechPosture(t *testing.T) {
	stale := store.PromptSectionsConfig{
		TradingFrequency: "FREQ_BOX — breakeven and trailing are both SUSPENDED at the wire (stale saved text).",
	}

	t.Run("default env suspends — the prompt must say SUSPENDED", func(t *testing.T) {
		t.Setenv("EXIT_MECHS_SUSPENDED", "")
		e := futuresTestEngine()
		e.config.PromptSections = stale
		p := e.BuildFuturesDecisionSystemPrompt("MNQ", 50000)
		if !strings.Contains(p, "SUSPENDED at the wire") {
			t.Fatalf("the rendered prompt never states the live posture (suspended):\n%s", tailPrompt(p))
		}
	})

	t.Run("env unsuspends — the prompt must say ACTIVE", func(t *testing.T) {
		t.Setenv("EXIT_MECHS_SUSPENDED", "0")
		e := futuresTestEngine()
		e.config.PromptSections = stale
		p := e.BuildFuturesDecisionSystemPrompt("MNQ", 50000)
		if !strings.Contains(p, "ACTIVE at the wire") {
			t.Fatalf("the rendered prompt does not reflect the LIVE unsuspended posture:\n%s", tailPrompt(p))
		}
	})
}

func tailPrompt(p string) string {
	if len(p) > 600 {
		return p[len(p)-600:]
	}
	return p
}
