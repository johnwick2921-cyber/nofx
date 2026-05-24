// Package ninjatrader implements the trader.Trader interface by writing
// trade signals to a CSV file that NinjaTrader's claudetrader.cs strategy
// consumes. Reads fills back via a CSV tailer.
//
// Limitations (intentional for Plan 1 SIM paper trading; lifted by Plan 1.5):
//   - CloseLong/CloseShort return error — position closes via SL/TP only
//   - CancelAllOrders/Cancel*Orders return error — not supported by CSV bridge
//   - GetBalance returns hardcoded $50k SIM101 mock
//   - GetClosedPnL returns nil
//   - GetOpenOrders returns empty slice
package ninjatrader

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"nofx/provider/ninjatrader"
	"nofx/trader/types"
)

type Config struct {
	DataDir string // /mnt/c/Users/<u>/NofxTrader/data
	Symbol  string // e.g. "MNQ" (informational only; NT uses chart's instrument)
}

// Trader satisfies trader/types.Trader using the CSV bridge.
type Trader struct {
	cfg    Config
	writer *ninjatrader.CSVWriter
	tailer *ninjatrader.CSVTailer

	mu           sync.Mutex
	stopLoss     map[string]float64 // key: "<symbol>:<side>"
	takePrft     map[string]float64
	entryRefs    map[string]float64
	lastFill     ninjatrader.FillRow
	lastQuantity float64
	hasFill      bool
}

func New(cfg Config) *Trader {
	t := &Trader{
		cfg:      cfg,
		writer:   ninjatrader.NewCSVWriter(cfg.DataDir),
		tailer:   ninjatrader.NewCSVTailer(cfg.DataDir, time.Second),
		stopLoss: map[string]float64{},
		takePrft: map[string]float64{},
		entryRefs: map[string]float64{},
	}
	go func() {
		_ = t.tailer.TailFills(context.Background(), func(f ninjatrader.FillRow) {
			t.mu.Lock()
			defer t.mu.Unlock()
			t.lastFill = f
			t.hasFill = true
		})
	}()
	return t
}

// Compile-time check that we implement the interface. If signatures drift,
// the build fails here — not silently at runtime.
var _ types.Trader = (*Trader)(nil)

// --- Trader interface methods ---

func (t *Trader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	return t.placeEntry(symbol, "LONG", quantity)
}

func (t *Trader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	return t.placeEntry(symbol, "SHORT", quantity)
}

func (t *Trader) placeEntry(symbol, side string, quantity float64) (map[string]interface{}, error) {
	t.mu.Lock()
	key := keyFor(symbol, side)
	sl := t.stopLoss[key]
	tp := t.takePrft[key]
	entryRef := t.entryRefs[key]
	t.mu.Unlock()

	if sl == 0 || tp == 0 {
		return nil, fmt.Errorf("ninjatrader: SetStopLoss and SetTakeProfit must be called before %s", side)
	}
	if entryRef == 0 {
		return nil, fmt.Errorf("ninjatrader: SetEntryReference must be called before %s", side)
	}

	orderID := fmt.Sprintf("nt-%d", time.Now().UnixNano())
	sig := ninjatrader.SignalRow{
		DateTime:   time.Now().Format("01/02/2006 15:04:05"),
		Direction:  side,
		EntryPrice: entryRef,
		StopLoss:   sl,
		TakeProfit: tp,
	}
	if err := t.writer.WriteSignal(sig); err != nil {
		return nil, err
	}

	t.mu.Lock()
	t.lastQuantity = quantity
	delete(t.stopLoss, key)
	delete(t.takePrft, key)
	delete(t.entryRefs, key)
	t.mu.Unlock()

	return map[string]interface{}{
		"status":   "submitted",
		"orderId":  orderID,
		"symbol":   symbol,
		"side":     side,
		"quantity": quantity,
	}, nil
}

func (t *Trader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	return nil, fmt.Errorf("ninjatrader: manual CloseLong not supported via CSV bridge — position closes via SL/TP set at entry")
}

func (t *Trader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	return nil, fmt.Errorf("ninjatrader: manual CloseShort not supported via CSV bridge — position closes via SL/TP set at entry")
}

func (t *Trader) SetLeverage(symbol string, leverage int) error {
	return nil // futures leverage is set at the broker, not per-order
}

func (t *Trader) SetMarginMode(symbol string, isCrossMargin bool) error {
	return nil // n/a for futures
}

func (t *Trader) SetStopLoss(symbol, positionSide string, quantity, stopPrice float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopLoss[keyFor(symbol, positionSide)] = stopPrice
	return nil
}

func (t *Trader) SetTakeProfit(symbol, positionSide string, quantity, takeProfitPrice float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.takePrft[keyFor(symbol, positionSide)] = takeProfitPrice
	return nil
}

func (t *Trader) SetEntryReference(symbol, positionSide string, entryPrice float64) error {
	if entryPrice <= 0 {
		return fmt.Errorf("ninjatrader: entry reference must be positive")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entryRefs[keyFor(symbol, positionSide)] = entryPrice
	return nil
}

func (t *Trader) CancelAllOrders(symbol string) error {
	return fmt.Errorf("ninjatrader: CancelAllOrders not supported via CSV bridge")
}

func (t *Trader) GetBalance() (map[string]interface{}, error) {
	// claudetrader.cs doesn't expose balance via CSV. For paper-mode v1, return
	// a fixed SIM101 balance ($50k default) so the trader loop doesn't fail
	// balance checks. Documented in plan; lifted by Plan 1.5.
	return map[string]interface{}{
		"total_equity":          50000.0,
		"available_balance":     50000.0,
		"wallet_balance":        50000.0,
		"total_pnl":             0.0,
		"totalEquity":           50000.0,
		"availableBalance":      50000.0,
		"totalWalletBalance":    50000.0,
		"totalUnrealizedProfit": 0.0,
	}, nil
}

func (t *Trader) GetPositions() ([]map[string]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.hasFill {
		return []map[string]interface{}{}, nil
	}
	side := strings.ToLower(t.lastFill.Direction)
	qty := t.lastQuantity
	if qty <= 0 {
		qty = 1.0
	}
	positionAmt := qty
	if side == "short" {
		positionAmt = -qty
	}
	return []map[string]interface{}{{
		"symbol":             t.cfg.Symbol,
		"side":               side,
		"entryPrice":         t.lastFill.EntryPrice,
		"markPrice":          t.lastFill.EntryPrice,
		"positionAmt":        positionAmt,
		"quantity":           qty,
		"unRealizedProfit":   0.0,
		"leverage":           1.0,
		"liquidationPrice":   0.0,
		"createdTime":        time.Now().UnixMilli(),
		"updatedTime":        time.Now().UnixMilli(),
	}}, nil
}

func (t *Trader) FormatQuantity(symbol string, quantity float64) (string, error) {
	return fmt.Sprintf("%.0f", quantity), nil
}

func (t *Trader) GetOrderStatus(symbol, orderID string) (map[string]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.hasFill {
		return map[string]interface{}{"status": "NEW"}, nil
	}
	qty := t.lastQuantity
	if qty <= 0 {
		qty = 1.0
	}
	return map[string]interface{}{
		"status":      "FILLED",
		"price":       t.lastFill.EntryPrice,
		"avgPrice":    t.lastFill.EntryPrice,
		"executedQty": qty,
		"commission":  0.0,
		"side":        t.lastFill.Direction,
	}, nil
}

func (t *Trader) GetClosedPnL(start time.Time, limit int) ([]types.ClosedPnLRecord, error) {
	return nil, nil
}

func (t *Trader) GetMarketPrice(symbol string) (float64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.hasFill {
		return 0, fmt.Errorf("ninjatrader: no fill yet, market price unavailable; use Databento client directly")
	}
	return t.lastFill.EntryPrice, nil
}

func (t *Trader) CancelStopLossOrders(symbol string) error {
	return fmt.Errorf("ninjatrader: CancelStopLossOrders not supported via CSV bridge — SL is set at entry")
}

func (t *Trader) CancelTakeProfitOrders(symbol string) error {
	return fmt.Errorf("ninjatrader: CancelTakeProfitOrders not supported via CSV bridge — TP is set at entry")
}

func (t *Trader) CancelStopOrders(symbol string) error {
	return fmt.Errorf("ninjatrader: CancelStopOrders not supported via CSV bridge")
}

func (t *Trader) GetOpenOrders(symbol string) ([]types.OpenOrder, error) {
	// CSV protocol doesn't expose pending orders. Signal file holds at most one
	// unconsumed signal; SL/TP live on NT's side after entry fills.
	return []types.OpenOrder{}, nil
}

func keyFor(symbol, side string) string {
	return symbol + ":" + side
}
