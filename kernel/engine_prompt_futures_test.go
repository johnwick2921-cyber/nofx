package kernel

import (
	"strings"
	"testing"
)

func TestBuildFuturesSystemPrompt_NoCryptoVocab(t *testing.T) {
	p := BuildFuturesSystemPrompt(FuturesPromptConfig{
		Symbol:             "MNQ",
		ContractMultiplier: 2.0, // MNQ = $2/point
		TickSize:           0.25,
		MinStopPoints:      15,
		MaxStopPoints:      50,
		MinRiskReward:      1.5,
	})

	// Must NOT contain crypto vocabulary
	forbidden := []string{
		"cryptocurrency", "altcoin", "BTC", "ETH", "USDT", "perpetual",
		"funding rate", "coins simultaneously",
	}
	for _, f := range forbidden {
		if strings.Contains(p, f) {
			t.Errorf("futures prompt contains forbidden crypto term %q", f)
		}
	}

	// Must contain futures-specific framing
	required := []string{
		"NQ", "tick", "contract", "stop loss", "take profit", "MNQ",
	}
	for _, r := range required {
		if !strings.Contains(p, r) {
			t.Errorf("futures prompt missing required term %q", r)
		}
	}
}

func TestBuildFuturesUserPrompt_IncludesIndicators(t *testing.T) {
	p := BuildFuturesUserPrompt(FuturesContext{
		Symbol:       "MNQ",
		CurrentPrice: 21500.00,
		EMA20:        21495.00,
		EMA50:        21480.00,
		RSI14:        58.3,
		MACD:         3.21,
		ATR14:        12.5,
		BollUpper:    21540.00,
		BollLower:    21460.00,
	})
	for _, s := range []string{"21500.00", "EMA20", "RSI14", "ATR14", "Bollinger"} {
		if !strings.Contains(p, s) {
			t.Errorf("user prompt missing %q", s)
		}
	}
}
