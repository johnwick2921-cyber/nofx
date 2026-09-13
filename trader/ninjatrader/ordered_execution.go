package ninjatrader

import (
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
	"strings"
)

// InstallOrderedExecutions installs durable execution observation at successful
// trader construction, before entries are possible. Observers survive Stop and
// are replaced by the next successful owner, independently of its trading Run.
func (t *TCPTrader) InstallOrderedExecutions(traderID, exchangeID, exchangeType string, st *store.Store, order func(ntwire.OrderUpdatePayload)) {
	if st == nil || t.server == nil || t.boundAccount == "" {
		return
	}
	t.StartCloseSync(traderID, exchangeID, exchangeType, st)
	t.StartPositionReconcile(traderID, exchangeID, exchangeType, st)
	pb := store.NewPositionBuilder(st.Position())
	t.server.RegisterOrderedExecutionsFor(t.symbol, t.boundAccount, ntwire.OrderedExecutionHandlers{
		Order: func(p ntwire.OrderUpdatePayload) {
			// Positive cumulative entry evidence invalidates older flat snapshots
			// even when price is missing, or Cancelled has no companion Fill.
			if p.Quantity > 0 && p.SignalID != "" && (p.OrderName == "" || p.OrderName == p.SignalID) && (strings.EqualFold(p.State, "partfilled") || ntwire.ClassifyOrderState(p.State) == ntwire.LivenessTerminal) {
				t.noteEntrySnapshotFence(p.SignalID, p.Quantity)
			}
			order(p)
		}, Fill: t.handleFill,
		Close: func(p ntwire.PositionClosePayload) { t.recordClose(traderID, exchangeID, exchangeType, st, pb, p) },
	})
}

// Only growth of a cumulative entry observation advances the fence; repeated
// old frames cannot invalidate a newer broker snapshot forever.
func (t *TCPTrader) noteEntrySnapshotFence(signal string, quantity int) {
	if t.server != nil {
		t.server.NoteEntryExecution(t.symbol, t.boundAccount, signal, quantity)
	}
}
