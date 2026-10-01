package store

import (
	"path/filepath"
	"testing"

	"vl/crypto"
)

// TestLegacyCryptoRowsLoadWithFuturesSettingsIntact is the C1 owner-data load
// proof (test side): a stored strategy / exchange / ai_model row carrying EVERY
// legacy crypto field must LOAD through its production call site with the
// futures-effective settings unchanged — the legacy fields are read-and-ignored,
// never a load failure and never a corruption of the futures configuration.
// Synthetic fixture only (never data.db or a copy of it; the real-copy check is
// the CTO's).
//
// Boot-cleanup survival is NOT asserted for the exchange row here: the C1 P0
// fix (cleanupIncompleteExchangeConfigs skips unsupported types, PR #188 D1–D3)
// is not yet ported at this head — do not pin the pre-fix deletion.
func TestLegacyCryptoRowsLoadWithFuturesSettingsIntact(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "c1-legacy.db")
	st, err := New(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	const userID = "u1"

	// 1) STRATEGY — every legacy crypto field, plus explicit futures settings.
	legacyStrategyJSON := `{
		"coin_source": {
			"source_type": "ai500",
			"static_coins": [],
			"use_ai500": true,
			"ai500_limit": 50,
			"oi_top_limit": 20,
			"use_oi_top": true,
			"use_hyper_all": true,
			"use_hyper_main": true,
			"netflow_source": true
		},
		"enable_sentinel": true,
		"watch_symbols": ["BTCUSDT", "ETHUSDT"],
		"indicators": {
			"klines": {
				"primary_timeframe": "5m",
				"selected_timeframes": ["5m", "15m", "1h"],
				"primary_count": 200
			}
		},
		"risk_control": {
			"max_margin_usage": 0.8,
			"min_position_size": 1,
			"btc_eth_max_leverage": 3,
			"altcoin_max_leverage": 5
		}
	}`
	if err := st.Strategy().Create(&Strategy{
		ID:       "s-legacy",
		UserID:   userID,
		Name:     "legacy crypto strategy",
		Config:   legacyStrategyJSON,
		IsActive: true,
	}); err != nil {
		t.Fatalf("plant legacy strategy: %v", err)
	}
	loaded, err := st.Strategy().Get(userID, "s-legacy")
	if err != nil {
		t.Fatalf("load legacy strategy: %v", err)
	}
	cfg, err := loaded.ParseConfig()
	if err != nil {
		t.Fatalf("parse legacy strategy config: %v", err)
	}
	// Futures-effective settings unchanged.
	if cfg.Indicators.Klines.PrimaryTimeframe != "5m" ||
		len(cfg.Indicators.Klines.SelectedTimeframes) != 3 ||
		cfg.Indicators.Klines.SelectedTimeframes[0] != "5m" ||
		cfg.Indicators.Klines.PrimaryCount != 200 {
		t.Fatalf("futures klines settings changed by legacy fields: %+v", cfg.Indicators.Klines)
	}
	if cfg.RiskControl.MaxMarginUsage != 0.8 || cfg.RiskControl.MinPositionSize != 1 {
		t.Fatalf("futures risk settings changed by legacy fields: %+v", cfg.RiskControl)
	}

	// 2) EXCHANGE — a pre-existing crypto row with every legacy credential
	// column set, planted at the DB level exactly as a migrated row would be.
	if err := st.gdb.Create(&Exchange{
		ID:                      "e-legacy",
		UserID:                  userID,
		Name:                    "legacy binance",
		Type:                    "cex",
		ExchangeType:            "binance",
		AccountName:             "Default",
		Enabled:                 true,
		APIKey:                  crypto.EncryptedString("legacy-api-key"),
		SecretKey:               crypto.EncryptedString("legacy-secret"),
		Passphrase:              crypto.EncryptedString(""),
		HyperliquidWalletAddr:   "0xhyper",
		HyperliquidUnifiedAcct:  true,
		AsterUser:               "aster-user",
		AsterSigner:             "aster-signer",
		AsterPrivateKey:         crypto.EncryptedString("aster-pk"),
		LighterWalletAddr:       "0xlighter",
		LighterPrivateKey:       crypto.EncryptedString("lighter-pk"),
		LighterAPIKeyPrivateKey: crypto.EncryptedString("lighter-apk"),
		LighterAPIKeyIndex:      7,
	}).Error; err != nil {
		t.Fatalf("plant legacy exchange row: %v", err)
	}
	ex, err := st.Exchange().GetByID(userID, "e-legacy")
	if err != nil {
		t.Fatalf("load legacy exchange row: %v", err)
	}
	if ex.ExchangeType != "binance" || !ex.Enabled {
		t.Fatalf("legacy exchange identity changed: type=%q enabled=%v", ex.ExchangeType, ex.Enabled)
	}
	if ex.HyperliquidWalletAddr != "0xhyper" || ex.AsterUser != "aster-user" ||
		ex.AsterSigner != "aster-signer" || ex.LighterWalletAddr != "0xlighter" || ex.LighterAPIKeyIndex != 7 {
		t.Fatalf("legacy credential columns changed on load: %+v", ex)
	}
	if len(string(ex.APIKey)) == 0 || len(string(ex.SecretKey)) == 0 {
		t.Fatalf("legacy credential columns did not round-trip (empty after decrypt)")
	}

	// 3) AI MODEL — the disabled claw402 row (C1: it stays, it never crashes).
	if err := st.AIModel().Create(userID, "m-legacy", "Claw402 legacy", "claw402", false, "0xlegacy", "https://claw402.ai"); err != nil {
		t.Fatalf("plant legacy ai_model row: %v", err)
	}
	model, err := st.AIModel().Get(userID, "m-legacy")
	if err != nil {
		t.Fatalf("load legacy ai_model row: %v", err)
	}
	if model.Provider != "claw402" || model.Enabled {
		t.Fatalf("legacy ai_model changed on load: provider=%q enabled=%v", model.Provider, model.Enabled)
	}

	// 4) REOPEN — initTables + cleanup run again; the strategy and the ai_model
	// must survive byte-for-byte (the exchange-row survival assertion lands with
	// the C1 P0 cleanup-skip port — see the header comment).
	if err := st.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	st2, err := New(dbPath)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer st2.Close()
	loaded2, err := st2.Strategy().Get(userID, "s-legacy")
	if err != nil {
		t.Fatalf("strategy row did not survive the reopen: %v", err)
	}
	if loaded2.Config != legacyStrategyJSON {
		t.Fatalf("strategy config changed across the reopen:\n%q\n%q", loaded2.Config, legacyStrategyJSON)
	}
	if _, err := st2.AIModel().Get(userID, "m-legacy"); err != nil {
		t.Fatalf("ai_model row did not survive the reopen: %v", err)
	}
}
