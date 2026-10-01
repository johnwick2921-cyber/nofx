package store

import (
	"encoding/json"
	"strings"
	"testing"
)

// Item A (crypto removal): the hyper_* coin-source schema fields are gone, but
// stored rows carrying the legacy keys must still load — unknown JSON fields are
// ignored, never a load failure — and the kept fields survive intact. The codec
// must not re-emit the removed keys.
func TestLegacyStrategyRowWithHyperCoinSourceKeysStillLoads(t *testing.T) {
	blob := []byte(`{"strategy_type":"ai_trading","language":"en","ai_config":{"coin_source":{
		"source_type": "static",
		"static_coins": ["MNQ"],
		"use_hyper_all": true,
		"use_hyper_main": false,
		"hyper_main_limit": 20
	}}}`)
	var cfg StrategyConfig
	if err := json.Unmarshal(blob, &cfg); err != nil {
		t.Fatalf("legacy stored row must load (unknown hyper_* keys ignored), got: %v", err)
	}
	if cfg.CoinSource.SourceType != "static" || len(cfg.CoinSource.StaticCoins) != 1 || cfg.CoinSource.StaticCoins[0] != "MNQ" {
		t.Fatalf("kept coin-source fields must survive the legacy row: %+v", cfg.CoinSource)
	}

	out, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("marshal after legacy load: %v", err)
	}
	for _, key := range []string{"use_hyper_all", "use_hyper_main", "hyper_main_limit"} {
		if strings.Contains(string(out), key) {
			t.Fatalf("the removed hyper_* key must not be re-emitted (%s): %s", key, out)
		}
	}
}
