// FIX-KNOBS A-priority (DS-105, 2026-09-26; census 883fe0fc) — the four DEAD
// risk knobs must be GONE, not displayed:
//
//	max_contracts_enabled / notional_cap_enabled — parse-only toggles, no gate
//	  ever read them (the D3 futures clamps are always-on venue safety and must
//	  never be toggle-able);
//	max_margin_usage — prompt text only, never a gate;
//	min_position_size — a configurable floor BEHIND the live hardcoded 12/60
//	  (kernel/engine_position.go), so editing it mostly does nothing.
//
// A risk control that looks active and does nothing is the worst kind, so they
// are removed from the struct, the registry, the schema surface and the UI —
// while old stored rows must still load (unknown keys are ignored, pinned here).
package store

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestDeadRiskKnobsAbsentFromStructAndRegistry(t *testing.T) {
	for _, name := range []string{
		"MaxMarginUsage", "MinPositionSize", "MaxContractsEnabled", "NotionalCapEnabled",
	} {
		if _, ok := reflect.TypeOf(RiskControlConfig{}).FieldByName(name); ok {
			t.Errorf("dead knob field %s still present on RiskControlConfig — a knob that looks active and does nothing is a defect", name)
		}
	}
	for _, path := range []string{
		"max_margin_usage", "min_position_size", "max_contracts_enabled", "notional_cap_enabled",
	} {
		if e, ok := knobRegistry[path]; ok {
			t.Errorf("dead knob %q still in the knob registry (%+v)", path, e)
		}
	}
}

func TestDeadRiskKnobsOldRowsStillLoad(t *testing.T) {
	blob := []byte(`{"strategy_type":"ai_trading","ai_config":{"risk_control":{"max_margin_usage":0.9,"min_position_size":12,"max_contracts_enabled":true,"notional_cap_enabled":false,"min_risk_reward_ratio":3,"min_confidence":65}}}`)
	var c StrategyConfig
	if err := json.Unmarshal(blob, &c); err != nil {
		t.Fatalf("a stored row carrying the removed knobs must still load: %v", err)
	}
	if c.RiskControl.MinRiskRewardRatio != 3 || c.RiskControl.MinConfidence != 65 {
		t.Errorf("the surviving fields must parse, got %+v", c.RiskControl)
	}
	// And the marshal half must never re-emit the dead keys.
	out, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"max_margin_usage", "min_position_size", "max_contracts_enabled", "notional_cap_enabled"} {
		if containsKey(out, key) {
			t.Errorf("marshal re-emitted the dead knob %q: %s", key, out)
		}
	}
}

func containsKey(blob []byte, key string) bool {
	var probe map[string]any
	if err := json.Unmarshal(blob, &probe); err != nil {
		return false
	}
	ai, _ := probe["ai_config"].(map[string]any)
	if ai == nil {
		return false
	}
	rc, _ := ai["risk_control"].(map[string]any)
	if rc == nil {
		return false
	}
	_, ok := rc[key]
	return ok
}

// FIX-KNOBS A (2026-09-26): external_data_sources.* — FetchExternalData has NO
// production caller, so the whole knob family is removed from the surface.
func TestExternalDataSourcesAbsentFromStruct(t *testing.T) {
	if _, ok := reflect.TypeOf(IndicatorConfig{}).FieldByName("ExternalDataSources"); ok {
		t.Error("dead knob field ExternalDataSources still present on IndicatorConfig")
	}
}

func TestExternalDataSourcesOldRowsStillLoad(t *testing.T) {
	blob := []byte(`{"strategy_type":"ai_trading","ai_config":{"indicators":{"external_data_sources":[{"name":"x","type":"api","url":"https://example.invalid","method":"GET","headers":{"Authorization":"Bearer secret"},"data_path":"a.b","refresh_secs":60}],"enable_ema":true}}}`)
	var c StrategyConfig
	if err := json.Unmarshal(blob, &c); err != nil {
		t.Fatalf("a stored row carrying external_data_sources must still load: %v", err)
	}
	if !c.Indicators.EnableEMA {
		t.Errorf("surviving sibling field must parse, got %+v", c.Indicators)
	}
	out, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, []byte("external_data_sources")) {
		t.Errorf("marshal re-emitted the removed external_data_sources: %s", out)
	}
}
