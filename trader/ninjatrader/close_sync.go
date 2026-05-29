package ninjatrader

import (
	"strings"
	"time"

	"nofx/logger"
	"nofx/market"
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// StartCloseSync consumes position_close frames from the TCP bridge and records
// each close into trader_positions (real exit price + futures realized PnL),
// then clears the in-memory fill so GetPositions reports flat.
//
// NT closes positions broker-side via the OCO bracket (SL/TP); the bot never
// issues a close_* order and — unlike every crypto broker — NT has no
// order-sync. So this is the ONLY path that transitions an open NT position to
// CLOSED, which is what populates the dashboard's position history.
func (t *TCPTrader) StartCloseSync(traderID, exchangeID, exchangeType string, st *store.Store) {
	if st == nil {
		return
	}
	pb := store.NewPositionBuilder(st.Position())
	t.closeSyncOnce.Do(func() {
		go func() {
			for p := range t.server.ClosedPositions() {
				t.recordClose(traderID, exchangeID, exchangeType, st, pb, p)
			}
		}()
		logger.Infof("🔄 NinjaTrader close-sync started (records SL/TP exits to position history)")
	})
}

func (t *TCPTrader) recordClose(
	traderID, exchangeID, exchangeType string,
	st *store.Store,
	pb *store.PositionBuilder,
	p ntwire.PositionClosePayload,
) {
	side, action := "LONG", "close_long"
	if strings.EqualFold(p.PositionSide, "short") {
		side, action = "SHORT", "close_short"
	}

	// The open record is keyed by the canonical bot symbol (root, e.g. "MNQ"),
	// not the resolved front-month contract — prefer t.symbol.
	symbol := t.symbol
	if symbol == "" {
		symbol = p.Symbol
	}
	qty := float64(p.Quantity)
	if qty <= 0 {
		qty = 1
	}

	// Realized PnL with the futures point value (PositionBuilder's fallback
	// formula omits it). Entry from the open record; fall back to last fill.
	entry := 0.0
	if open, err := st.Position().GetOpenPositionBySymbol(traderID, symbol, side); err == nil && open != nil {
		entry = open.EntryPrice
	}
	if entry == 0 {
		t.mu.Lock()
		entry = t.lastFill.FillPrice
		t.mu.Unlock()
	}
	pv := market.FuturesPointValue(symbol)
	if pv <= 0 {
		pv = 1
	}
	realizedPnL := 0.0
	if entry > 0 {
		if side == "LONG" {
			realizedPnL = (p.ExitPrice - entry) * qty * pv
		} else {
			realizedPnL = (entry - p.ExitPrice) * qty * pv
		}
	}

	exitMs := time.Now().UTC().UnixMilli()
	if p.ExitTime != "" {
		if ts, err := time.Parse(time.RFC3339, p.ExitTime); err == nil {
			exitMs = ts.UTC().UnixMilli()
		}
	}

	if err := pb.ProcessTrade(traderID, exchangeID, exchangeType, symbol, side, action,
		qty, p.ExitPrice, 0, realizedPnL, exitMs, p.SignalID); err != nil {
		logger.Warnf("ninjatrader/tcp: record close failed (%s %s): %v", symbol, side, err)
	} else {
		logger.Infof("📕 NT position closed: %s %s qty=%.0f exit=%.2f reason=%s pnl=%.2f",
			symbol, side, qty, p.ExitPrice, p.ExitReason, realizedPnL)
	}

	// Mark flat so GetPositions stops reporting the now-closed position.
	t.mu.Lock()
	t.hasFill = false
	t.mu.Unlock()
}
