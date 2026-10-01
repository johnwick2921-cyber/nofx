package store

import (
	"encoding/json"
	"strings"
	"testing"
)

// C1 (crypto removal): legacy stored strategy rows carry btc_eth_max_leverage
// and altcoin_max_leverage under ai_config.risk_control. Those knobs are gone
// from RiskControlConfig, but stored rows MUST still load — unknown JSON fields
// are ignored, never a load failure, and the kept fields survive intact. The
// codec also must not re-emit the removed keys.
func TestLegacyStrategyRowWithCryptoLeverageKnobsStillLoads(t *testing.T) {
	blob := []byte(`{"strategy_type":"ai_trading","language":"en","ai_config":{"risk_control":{
		"btc_eth_max_leverage": 10,
		"altcoin_max_leverage": 5,
		"min_confidence": 65,
		"min_risk_reward_ratio": 3.0,
		"max_positions": 3
	}}}`)
	var cfg StrategyConfig
	if err := json.Unmarshal(blob, &cfg); err != nil {
		t.Fatalf("legacy stored row must load (unknown crypto knobs ignored), got: %v", err)
	}
	if cfg.RiskControl.MinConfidence != 65 || cfg.RiskControl.MaxPositions != 3 {
		t.Fatalf("kept fields must survive the legacy row: %+v", cfg.RiskControl)
	}

	out, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("marshal after legacy load: %v", err)
	}
	if strings.Contains(string(out), "btc_eth_max_leverage") || strings.Contains(string(out), "altcoin_max_leverage") {
		t.Fatalf("the removed crypto knobs must not be re-emitted: %s", out)
	}
}
