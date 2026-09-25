// W117 F2 trader half (9b379c8c, re-derived on today's code) — the owning
// durable consumers run on the TCP read goroutine, in receive order, BEFORE
// advisory channel fanout. This file installs the registration; the three
// channel consumers skip the already-handled copy via the OrderedHandled
// flags.

package ninjatrader

import (
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// InstallOrderedExecutions installs durable execution observation at successful
// trader construction, before entries are possible. Observers survive Stop and
// are replaced by the next successful owner, independently of its trading Run.
// The registered callbacks MUST NOT wait for broker replies — they run on the
// TCP read goroutine.
func (t *TCPTrader) InstallOrderedExecutions(traderID, exchangeID, exchangeType string, st *store.Store, order func(ntwire.OrderUpdatePayload)) {
	if st == nil || t.server == nil || t.boundAccount == "" {
		return
	}
	t.StartCloseSync(traderID, exchangeID, exchangeType, st)
	t.StartPositionReconcile(traderID, exchangeID, exchangeType, st)
	pb := store.NewPositionBuilder(st.Position())
	t.server.RegisterOrderedExecutionsFor(t.symbol, t.boundAccount, ntwire.OrderedExecutionHandlers{
		Order: order,
		Fill:  t.handleFill,
		Close: func(p ntwire.PositionClosePayload) { t.recordClose(traderID, exchangeID, exchangeType, st, pb, p) },
	})
}
