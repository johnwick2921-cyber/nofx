// Package ninjatrader — alternative Trader implementation that emits signals
// through the Plan 1.5 TCP bridge instead of the Plan 1 CSV bridge.
//
// SEPARATE TYPE from the existing CSV `Trader` (per ADR-007: additive, not a
// modification). Implements the same 19-method `trader/types.Trader`
// interface. Selection is via NT_TRANSPORT env var (transport.go).
//
// Wire-protocol contract: provider/ninjatrader/tcp_framing.go + tcp_server.go.
// Tick rounding: shared with the CSV trader via in-package call to
// RoundToTick / InstrumentTickSize (tick_rounding.go).
package ninjatrader

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"nofx/logger"
	"nofx/provider/databento"
	ntwire "nofx/provider/ninjatrader"
	"nofx/trader/types"
)

// TCPTrader satisfies trader/types.Trader using the TCP bridge.
type TCPTrader struct {
	server *ntwire.TCPServer
	symbol string // for tick rounding selection

	mu       sync.Mutex
	stopLoss map[string]float64 // key: "<symbol>:<side>"
	takePrft map[string]float64
	lastFill ntwire.FillPayload
	hasFill  bool

	// pending tracks signal_id → side so we can correlate fills back to
	// position state. Optional: not strictly needed for the 19-method
	// interface, but cheap insurance.
	pendingMu sync.Mutex
	pending   map[string]string // signal_id -> side
}

// Compile-time interface guard (ADR-007 pattern — mirrors CSV Trader).
var _ types.Trader = (*TCPTrader)(nil)

// NewTCPTrader wraps an already-Started TCPServer with the Trader interface.
// The server's lifecycle is owned by the caller (transport.go).
func NewTCPTrader(server *ntwire.TCPServer, symbol string) *TCPTrader {
	t := &TCPTrader{
		server:   server,
		symbol:   symbol,
		stopLoss: map[string]float64{},
		takePrft: map[string]float64{},
		pending:  map[string]string{},
	}
	// Subscribe to inbound fills — update lastFill cache (mirrors CSV Trader).
	go func() {
		for fill := range server.Fills() {
			t.mu.Lock()
			t.lastFill = fill
			t.hasFill = true
			t.mu.Unlock()
		}
	}()
	return t
}

// --- Trader interface methods (19 total) ---

func (t *TCPTrader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	return t.placeEntry(symbol, "long", quantity)
}

func (t *TCPTrader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	return t.placeEntry(symbol, "short", quantity)
}

func (t *TCPTrader) placeEntry(symbol, side string, quantity float64) (map[string]interface{}, error) {
	// Expiry warning — defense in depth (mirrors CSV Trader).
	if days := databento.DaysUntilExpiry(symbol, time.Now()); days >= 0 && days <= 5 {
		logger.Warnf("ninjatrader/tcp: placing %s entry on %s within %d days of expiry — verify contract roll", side, symbol, days)
	}

	// CSV Trader keys SL/TP by "LONG"/"SHORT" uppercase; mirror that here.
	upperSide := upperSideStr(side)
	t.mu.Lock()
	sl := t.stopLoss[keyFor(symbol, upperSide)]
	tp := t.takePrft[keyFor(symbol, upperSide)]
	t.mu.Unlock()
	if sl == 0 || tp == 0 {
		return nil, fmt.Errorf("ninjatrader/tcp: SetStopLoss and SetTakeProfit must be called before %s", side)
	}

	// Mirror CSV Trader: entry price for the wire is a reference value
	// computed as SL/TP midpoint. NT places market orders; entry on the
	// wire is for the AddOn's logging + protective bracket reference.
	entryRef := (sl + tp) / 2.0

	tick := InstrumentTickSize(t.symbol)
	entry := RoundToTick(entryRef, tick)
	sl = RoundToTick(sl, tick)
	tp = RoundToTick(tp, tick)

	signalID := uuid.NewString()
	payload := ntwire.SignalPayload{
		Symbol:     t.symbol,
		Side:       side, // lowercase per spec L4390
		Quantity:   int(quantity),
		Entry:      entry,
		StopLoss:   sl,
		TakeProfit: tp,
		SignalID:   signalID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	t.pendingMu.Lock()
	t.pending[signalID] = upperSide
	t.pendingMu.Unlock()

	if err := t.server.SendSignal(payload); err != nil {
		return nil, fmt.Errorf("ninjatrader/tcp: send signal: %w", err)
	}
	return map[string]interface{}{
		"status":    "submitted",
		"symbol":    symbol,
		"side":      side,
		"quantity":  quantity,
		"signal_id": signalID,
	}, nil
}

func (t *TCPTrader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	// Plan 1.5 wire protocol does not (yet) define a "close" frame; per spec
	// L4398-4406 only signal/fill/heartbeat/ack are wire types. Match CSV
	// Trader semantics: positions close via SL/TP only. A future spec rev
	// can add a close-side signal frame; until then this returns the same
	// error the CSV Trader returns to keep behaviour consistent.
	return nil, fmt.Errorf("ninjatrader/tcp: manual CloseLong not supported — position closes via SL/TP set at entry")
}

func (t *TCPTrader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	return nil, fmt.Errorf("ninjatrader/tcp: manual CloseShort not supported — position closes via SL/TP set at entry")
}

func (t *TCPTrader) SetLeverage(symbol string, leverage int) error {
	return nil // futures leverage is set at the broker, not per-order
}

func (t *TCPTrader) SetMarginMode(symbol string, isCrossMargin bool) error {
	return nil // n/a for futures
}

func (t *TCPTrader) SetStopLoss(symbol, positionSide string, quantity, stopPrice float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopLoss[keyFor(symbol, upperSideStr(positionSide))] = stopPrice
	return nil
}

func (t *TCPTrader) SetTakeProfit(symbol, positionSide string, quantity, takeProfitPrice float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.takePrft[keyFor(symbol, upperSideStr(positionSide))] = takeProfitPrice
	return nil
}

func (t *TCPTrader) CancelAllOrders(symbol string) error {
	return fmt.Errorf("ninjatrader/tcp: CancelAllOrders not supported")
}

func (t *TCPTrader) GetBalance() (map[string]interface{}, error) {
	// Mirror CSV Trader — TCP protocol doesn't expose balance either.
	return map[string]interface{}{
		"totalEquity":      50000.0,
		"availableBalance": 50000.0,
	}, nil
}

func (t *TCPTrader) GetPositions() ([]map[string]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.hasFill {
		return []map[string]interface{}{}, nil
	}
	return []map[string]interface{}{{
		"symbol":     t.symbol,
		"side":       upperSideStr(t.lastFill.Side),
		"entryPrice": t.lastFill.FillPrice,
		"quantity":   float64(t.lastFill.Quantity),
	}}, nil
}

func (t *TCPTrader) FormatQuantity(symbol string, quantity float64) (string, error) {
	return fmt.Sprintf("%.0f", quantity), nil
}

func (t *TCPTrader) GetOrderStatus(symbol, orderID string) (map[string]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.hasFill {
		return map[string]interface{}{"status": "pending"}, nil
	}
	return map[string]interface{}{
		"status": t.lastFill.Status,
		"price":  t.lastFill.FillPrice,
		"side":   upperSideStr(t.lastFill.Side),
	}, nil
}

func (t *TCPTrader) GetClosedPnL(start time.Time, limit int) ([]types.ClosedPnLRecord, error) {
	return nil, nil
}

func (t *TCPTrader) GetMarketPrice(symbol string) (float64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.hasFill {
		return 0, fmt.Errorf("ninjatrader/tcp: no fill yet, market price unavailable; use Databento client directly")
	}
	return t.lastFill.FillPrice, nil
}

func (t *TCPTrader) CancelStopLossOrders(symbol string) error {
	return fmt.Errorf("ninjatrader/tcp: CancelStopLossOrders not supported — SL is set at entry")
}

func (t *TCPTrader) CancelTakeProfitOrders(symbol string) error {
	return fmt.Errorf("ninjatrader/tcp: CancelTakeProfitOrders not supported — TP is set at entry")
}

func (t *TCPTrader) CancelStopOrders(symbol string) error {
	return fmt.Errorf("ninjatrader/tcp: CancelStopOrders not supported")
}

func (t *TCPTrader) GetOpenOrders(symbol string) ([]types.OpenOrder, error) {
	return []types.OpenOrder{}, nil
}

// upperSideStr normalises "long"/"LONG"/"Long" → "LONG" for SL/TP map keys.
// CSV Trader keys are uppercase, so we mirror that to keep behaviour parallel.
func upperSideStr(side string) string {
	switch side {
	case "long", "LONG", "Long":
		return "LONG"
	case "short", "SHORT", "Short":
		return "SHORT"
	default:
		return side
	}
}
