// Package ninjatrader — env-var router for the Plan 1/Plan 1.5 transport split.
//
// NT_TRANSPORT=csv  (default — Plan 1, SIM-validated)
// NT_TRANSPORT=tcp  (Plan 1.5 opt-in)
//
// Zero-downtime flip-back: change the env var, restart the bot, no code
// rebuild required. The TCP path starts a listener on 127.0.0.1:36974; the
// C# AddOn dials in on NT startup (see ninjascript/VLTraderTCPClient.cs
// scaffolded under Plan 1.5 spec L4370).
package ninjatrader

import (
	"context"
	"fmt"
	"os"
	"strings"

	ntwire "nofx/provider/ninjatrader"
	"nofx/trader/types"
)

// TransportEnvVar is the env-var name read by NewTraderFromEnv. Exported so
// tests / docs can reference the canonical key without string duplication.
const TransportEnvVar = "NT_TRANSPORT"

// NewTraderFromEnv returns either the CSV Trader (default — Plan 1 SIM-validated)
// or the TCPTrader (Plan 1.5 opt-in), based on the NT_TRANSPORT env var.
//
// Unknown values are an error — fail-fast instead of silently defaulting,
// because a typo'd env var on a live-money bot is a footgun.
func NewTraderFromEnv(cfg Config) (types.Trader, error) {
	transport := strings.ToLower(strings.TrimSpace(os.Getenv(TransportEnvVar)))
	switch transport {
	case "", "csv":
		return New(cfg), nil
	case "tcp":
		server := ntwire.NewTCPServer(nil)
		if err := server.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("transport: start tcp server: %w", err)
		}
		return NewTCPTrader(server, cfg.Symbol), nil
	default:
		return nil, fmt.Errorf("transport: unknown %s=%q (want csv or tcp)", TransportEnvVar, transport)
	}
}
