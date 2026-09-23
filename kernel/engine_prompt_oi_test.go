package kernel

import (
	"strings"
	"testing"

	"nofx/store"
)

// TestWriteAvailableIndicators_OIGatedByEnableOI locks the OI honesty fix: the
// "Open Interest (OI) data" availability line appears IFF EnableOI is true. A
// futures strategy (EnableOI=false via applyFuturesIndicatorDefaults) therefore
// omits it from the prompt. Crypto (EnableOI=true) lists it — and, since this
// build has no crypto open-interest source (the exchange-hosted fetcher was
// removed with its broker), the line says n/a instead of advertising a feed
// that is not there (a CORRECTION: the old "- Open Interest (OI) data" line
// promised a value every coin now reports absent). The user-prompt OI value
// line is gated by the same EnableOI flag (engine_prompt.go).
func TestWriteAvailableIndicators_OIGatedByEnableOI(t *testing.T) {
	const oiLine = "- Open Interest (OI) data"
	const cryptoOILine = "- Open Interest (OI) data: n/a (no open-interest source)\n"

	withOI := &store.StrategyConfig{Indicators: store.IndicatorConfig{
		Klines:   store.KlineConfig{SelectedTimeframes: []string{"5m"}},
		EnableOI: true,
	}}
	var on strings.Builder
	NewStrategyEngine(withOI).writeAvailableIndicators(&on)
	if !strings.Contains(on.String(), cryptoOILine) {
		t.Fatalf("EnableOI=true must list OI, as n/a (no crypto OI source); got:\n%s", on.String())
	}

	withoutOI := &store.StrategyConfig{Indicators: store.IndicatorConfig{
		Klines:   store.KlineConfig{SelectedTimeframes: []string{"5m"}},
		EnableOI: false,
	}}
	var off strings.Builder
	NewStrategyEngine(withoutOI).writeAvailableIndicators(&off)
	if strings.Contains(off.String(), oiLine) {
		t.Fatalf("EnableOI=false (futures default) must NOT list OI; got:\n%s", off.String())
	}
}
