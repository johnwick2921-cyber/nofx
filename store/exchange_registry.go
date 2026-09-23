package store

import (
	"errors"
	"fmt"
	"strings"
)

// supportedExchangeTypes is the ONE registry of broker types this build can
// construct (trader.NewAutoTrader's switch — a parity test pins the two
// together). Every path that meets a stored exchange_type — the trader load
// paths, the account-state probe, update-exchange, create-exchange and the
// boot cleanup below — consults it FIRST, so a type removed from the build is
// refused by name everywhere, never coerced and never rewritten.
//
// Matching is EXACT (no case-folding): NewAutoTrader's switch is exact, so a
// row the broker factory cannot construct is not "supported" here either.
var supportedExchangeTypes = []string{
	"bybit",
	"okx",
	"bitget",
	"gate",
	"kucoin",
	"hyperliquid",
	"aster",
	"lighter",
	"indodax",
	"ninjatrader",
}

// ErrUnsupportedExchangeType is wrapped by every refusal CheckSupportedExchangeType
// returns, so callers classify with errors.Is instead of matching text.
var ErrUnsupportedExchangeType = errors.New("unsupported exchange_type")

// SupportedExchangeTypes returns a copy of the registry, in registry order.
func SupportedExchangeTypes() []string {
	out := make([]string, len(supportedExchangeTypes))
	copy(out, supportedExchangeTypes)
	return out
}

// IsSupportedExchangeType reports whether exchangeType is a broker this build
// constructs. Exact match; "" is never supported.
func IsSupportedExchangeType(exchangeType string) bool {
	for _, t := range supportedExchangeTypes {
		if t == exchangeType {
			return true
		}
	}
	return false
}

// CheckSupportedExchangeType returns nil for a supported type and otherwise the
// named refusal — the stored value interpolated with %q, so the removed venue
// is named at runtime from the row itself, never from a literal in source.
func CheckSupportedExchangeType(exchangeType string) error {
	if IsSupportedExchangeType(exchangeType) {
		return nil
	}
	return fmt.Errorf("%w %q — not a broker this build constructs (supported: %s)",
		ErrUnsupportedExchangeType, exchangeType, strings.Join(supportedExchangeTypes, ", "))
}
