package ninjatrader

import (
	"errors"

	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// W117 slice A (F2) — the owning trader's durable execution consumers ride the
// server's per-(symbol,account) ordered worker, in TCP receive order. The
// worker calls exactly these three closures; the legacy advisory consumers
// skip frames the worker owns (OrderedOwned), so nothing is double-applied.
//
// Lifecycle (R5): InstallOrderedExecutions is called from AutoTrader.Run —
// never from NewTCPTrader — and the returned closure MUST be called on Stop.
// A second LIVE registration for the same (symbol, account) is refused loudly
// by the server (two traders on one account would double-apply the same
// fills); the refusal is an error here, not a silent eviction.

// InstallOrderedExecutions wires this trader's durable consumers into the
// server's ordered-execution worker for its (symbol, account) owner. The
// returned func unregisters them (idempotent; call on Stop).
func (t *TCPTrader) InstallOrderedExecutions(traderID, exchangeID, exchangeType string, st *store.Store, orderFn func(ntwire.OrderUpdatePayload)) (func(), error) {
	if t == nil || t.server == nil {
		return nil, errors.New("install ordered executions: nil trader or server")
	}
	if st == nil {
		return nil, errors.New("install ordered executions: nil store")
	}
	// The durable close consumer is the SAME recordClose the legacy close-sync
	// path calls — one funnel, so the worker's receive-order close applies the
	// same ownership routing, pnl attribution and (R6) ApplyNT8Exit parking.
	pb := store.NewPositionBuilder(st.Position())
	unreg, err := t.server.RegisterOrderedExecutionsFor(t.symbol, t.boundAccount, ntwire.OrderedExecutionHandlers{
		Order: orderFn,
		Fill:  t.handleFillInbound,
		Close: func(p ntwire.PositionClosePayload) {
			t.recordClose(traderID, exchangeID, exchangeType, st, pb, p)
		},
	})
	if err != nil {
		return nil, err
	}
	t.mu.Lock()
	t.traderID = traderID
	t.st = st
	t.mu.Unlock()
	return unreg, nil
}
