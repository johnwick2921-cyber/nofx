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
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"

	"nofx/logger"
	"nofx/market"
	"nofx/provider/databento"
	ntwire "nofx/provider/ninjatrader"
	"nofx/trader/types"
) // reflect used in GetBalance to notify parent AutoTrader

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

	// closeSyncOnce guards StartCloseSync so a re-entrant AutoTrader.Run never
	// spawns a second consumer racing on the single ClosedPositions() channel.
	closeSyncOnce sync.Once

	// Plan 4 Stage 4 — reference to the parent AutoTrader (optional).
	// Used to notify the AutoTrader when the first account_balance frame arrives.
	// Set by transport.go after creating the trader.
	parentAutoTrader interface{} // *AutoTrader (avoid circular import)
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
	return t.sendClose("long", quantity)
}

func (t *TCPTrader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	return t.sendClose("short", quantity)
}

// sendClose asks the AddOn to flatten the symbol's position (close at market +
// cancel the protective bracket). The resulting market-exit fill returns as a
// position_close frame, which the close-sync records to history.
func (t *TCPTrader) sendClose(side string, quantity float64) (map[string]interface{}, error) {
	payload := ntwire.ClosePositionPayload{
		Symbol:   t.symbol,
		Side:     side,
		Quantity: int(quantity),
		SignalID: uuid.NewString(),
	}
	if err := t.server.SendClosePosition(payload); err != nil {
		return nil, fmt.Errorf("ninjatrader/tcp: send close: %w", err)
	}
	return map[string]interface{}{
		"status": "close_submitted",
		"symbol": t.symbol,
		"side":   side,
	}, nil
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
	// Plan 4.11 — serve the REAL NT SIM account from the latest account_balance
	// frame (C# AddOn → tcp_server). No more $50k mock. Until the first frame
	// arrives (AddOn connecting), report zeros so the dashboard shows "no data"
	// rather than a fabricated balance; the C# emits on connect + periodically.
	if acct, ok := t.server.AccountState(); ok {
		// Plan 4 Stage 4 — notify parent AutoTrader that balance has arrived
		// (used by defer-until-balance guard in runCycle).
		// Use reflection to avoid circular import (trader/ninjatrader → trader).
		if t.parentAutoTrader != nil {
			if method := reflect.ValueOf(t.parentAutoTrader).MethodByName("SetHasReceivedBalance"); method.IsValid() {
				method.Call([]reflect.Value{reflect.ValueOf(true)})
			}
		}

		equity := acct.NetLiquidation
		if equity == 0 {
			equity = acct.CashValue + acct.UnrealizedPnL
		}
		avail := acct.BuyingPower
		if avail == 0 {
			avail = acct.CashValue
		}
		return map[string]interface{}{
			"totalEquity":           equity,
			"availableBalance":      avail,
			"totalWalletBalance":    acct.CashValue,
			"totalUnrealizedProfit": acct.UnrealizedPnL,
		}, nil
	}
	return map[string]interface{}{
		"totalEquity":      0.0,
		"availableBalance": 0.0,
	}, nil
}

func (t *TCPTrader) GetPositions() ([]map[string]interface{}, error) {
	t.mu.Lock()
	if !t.hasFill {
		t.mu.Unlock()
		return []map[string]interface{}{}, nil
	}
	fill := t.lastFill
	t.mu.Unlock()

	side := upperSideStr(fill.Side)
	qty := float64(fill.Quantity)
	entry := fill.FillPrice

	// Mark price from the live BarCache (latest 5m bar close); fall back to the
	// entry fill if no bars are cached yet. This makes uPnL a real, moving
	// number instead of the {qty:1.0} stub with no mark.
	mark := entry
	if bars := t.server.BarCache().Get(t.symbol, "5m"); len(bars) > 0 {
		mark = bars[len(bars)-1].C
	}

	// Futures: contract point value (MNQ=$2/pt). uPnL = (mark-entry) × qty ×
	// pointValue × direction. positionAmt is signed (short < 0) per the
	// Binance convention GetAccountInfo expects.
	pv := market.FuturesPointValue(t.symbol)
	if pv <= 0 {
		pv = 1
	}
	dir := 1.0
	signedQty := qty
	if side == "SHORT" {
		dir = -1.0
		signedQty = -qty
	}
	uPnL := (mark - entry) * dir * qty * pv
	uPnLPct := 0.0
	if entry > 0 {
		uPnLPct = (mark - entry) / entry * 100 * dir
	}

	return []map[string]interface{}{{
		// snake_case — read by the UI Position type (web/src/types/trading.ts)
		"symbol":             t.symbol,
		"side":               side,
		"entry_price":        entry,
		"mark_price":         mark,
		"quantity":           qty,
		"leverage":           1.0, // futures: margin is per-contract, not crypto leverage
		"unrealized_pnl":     uPnL,
		"unrealized_pnl_pct": uPnLPct,
		"liquidation_price":  0.0,
		"margin_used":        0.0,
		// camelCase — read by AutoTrader.GetAccountInfo (Binance-style margin calc)
		"positionAmt":      signedQty,
		"entryPrice":       entry,
		"markPrice":        mark,
		"unRealizedProfit": uPnL,
	}}, nil
}

// DebugPlaceTestTrade places a deterministic 1-contract bracket order on the
// resolved SIM account for end-to-end proof (Plan 4.11 dashboard: signal →
// NT8 SIM fill → position). It BYPASSES the AI + risk gate — a debug harness,
// NOT a trading path — and is reachable only via the SIM-gated
// /api/debug/nt-test-trade endpoint. SL/TP are priced off the latest cached
// MNQ bar (20pt stop / 40pt target → ~2:1). side = "short" else long.
func (t *TCPTrader) DebugPlaceTestTrade(side string) (map[string]interface{}, error) {
	bars := t.server.BarCache().Get(t.symbol, "5m")
	if len(bars) == 0 {
		return nil, fmt.Errorf("ninjatrader/tcp: no %s bars cached; cannot price a test trade (recompile/connect NT8 first)", t.symbol)
	}
	price := bars[len(bars)-1].C
	tick := InstrumentTickSize(t.symbol)
	if side == "short" {
		_ = t.SetStopLoss(t.symbol, "short", 1, RoundToTick(price+20, tick))
		_ = t.SetTakeProfit(t.symbol, "short", 1, RoundToTick(price-40, tick))
		return t.OpenShort(t.symbol, 1, 1)
	}
	_ = t.SetStopLoss(t.symbol, "long", 1, RoundToTick(price-20, tick))
	_ = t.SetTakeProfit(t.symbol, "long", 1, RoundToTick(price+40, tick))
	return t.OpenLong(t.symbol, 1, 1)
}

// BarCache exposes the live NT8 bar cache for the Stage 4 SSE chart relay.
// Read-only access; the relay polls Get(symbol, timeframe) for snapshots +
// incremental updates.
func (t *TCPTrader) BarCache() *ntwire.BarCache { return t.server.BarCache() }

// GetServer exposes the underlying TCP server for API handlers (e.g., account selection).
// Used by the /api/accounts and /api/account/select handlers to interact with the NT AddOn.
func (t *TCPTrader) GetServer() *ntwire.TCPServer { return t.server }

// SetParentAutoTrader sets the parent AutoTrader reference (Plan 4 Stage 4).
// Called by transport.go after creating the TCPTrader. Used to notify when
// the first account_balance frame arrives (defer-until-balance guard).
func (t *TCPTrader) SetParentAutoTrader(parent interface{}) {
	t.parentAutoTrader = parent
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
