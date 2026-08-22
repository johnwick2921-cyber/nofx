package store

import (
	"strings"
	"testing"
)

// TestRegimeSurvivesCreateAndEditPaths is the G1 Studio-persistence proof (the
// cadence-drop lesson): the regime block must survive BOTH config paths.
//   - CREATE: handleCreateStrategy binds the body into StrategyConfig
//     (UnmarshalJSON) and persists it via json.Marshal (MarshalJSON) — the
//     round trip below is exactly that path.
//   - EDIT: handleUpdateStrategy runs every edit through MergeStrategyConfig —
//     an unrelated edit must NOT wipe regime, and a partial regime patch must
//     deep-merge.
//   - DEFAULT: nil regime / nil htf_veto resolves ON (dispatch 1.3).
func TestRegimeSurvivesCreateAndEditPaths(t *testing.T) {
	// Default semantics: absent regime → veto ON.
	def := GetDefaultStrategyConfig("en")
	if !def.HTFVetoEnabled() {
		t.Fatalf("shipped default must be ON (nil regime)")
	}
	if strings.Contains(string(mustMarshal(t, def)), "regime") {
		t.Fatalf("nil regime must emit no regime key (byte-identical legacy configs)")
	}

	// CREATE path: htf_veto=false round-trips through marshal/unmarshal.
	off := false
	cfg := def
	cfg.Regime = &RegimeConfig{HTFVeto: &off}
	blob := mustMarshal(t, cfg)
	if !strings.Contains(string(blob), `"regime":{"htf_veto":false}`) {
		t.Fatalf("marshal must carry regime.htf_veto=false, got %s", blob)
	}
	var back StrategyConfig
	if err := back.UnmarshalJSON(blob); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.HTFVetoEnabled() {
		t.Fatalf("htf_veto=false must survive the create round trip as OFF")
	}

	// EDIT path (1): an unrelated edit must NOT wipe regime.
	on := true
	base := def
	base.Regime = &RegimeConfig{HTFVeto: &on}
	merged, err := MergeStrategyConfig(base, map[string]any{"prompt_variant": "aggressive"})
	if err != nil {
		t.Fatalf("merge unrelated: %v", err)
	}
	if merged.Regime == nil || merged.Regime.HTFVeto == nil || !*merged.Regime.HTFVeto {
		t.Fatalf("unrelated edit wiped regime: %+v", merged.Regime)
	}

	// EDIT path (2): a partial regime patch toggles htf_veto off.
	merged2, err := MergeStrategyConfig(base, map[string]any{"regime": map[string]any{"htf_veto": false}})
	if err != nil {
		t.Fatalf("merge regime patch: %v", err)
	}
	if merged2.HTFVetoEnabled() {
		t.Fatalf("regime.htf_veto=false patch must land as OFF, got %+v", merged2.Regime)
	}

	// G6 field rides the same seam: loss_streak_n survives create + edit.
	n := 0
	cfg.Regime = &RegimeConfig{LossStreakN: &n}
	blob2 := mustMarshal(t, cfg)
	if !strings.Contains(string(blob2), `"loss_streak_n":0`) {
		t.Fatalf("marshal must carry regime.loss_streak_n, got %s", blob2)
	}
	var back2 StrategyConfig
	if err := back2.UnmarshalJSON(blob2); err != nil {
		t.Fatalf("unmarshal loss_streak_n: %v", err)
	}
	if back2.LossStreakNValue() != 0 {
		t.Fatalf("loss_streak_n=0 must survive the round trip as OFF")
	}
	merged3, err := MergeStrategyConfig(base, map[string]any{"regime": map[string]any{"loss_streak_n": 0}})
	if err != nil {
		t.Fatalf("merge loss_streak_n patch: %v", err)
	}
	if merged3.LossStreakNValue() != 0 {
		t.Fatalf("loss_streak_n=0 patch must land as OFF, got %+v", merged3.Regime)
	}
}

func mustMarshal(t *testing.T, c StrategyConfig) []byte {
	t.Helper()
	blob, err := c.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return blob
}
