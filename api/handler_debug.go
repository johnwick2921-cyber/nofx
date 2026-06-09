package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nofx/config"
	ntTrader "nofx/trader/ninjatrader"
)

// handleNTBarsUnsubscribe drops an EXTRA bar subscription at runtime (P5.1
// dispose-one proof / P5.3 building block). Futures-mode only; the primary
// (trading) symbol is refused at the TCP-server layer, so this can never tear
// down the live trader's feed. Bars-only — no order/position path involved.
func (s *Server) handleNTBarsUnsubscribe(c *gin.Context) {
	if cfg := config.Get(); cfg == nil || cfg.TradingMode != "futures" {
		SafeBadRequest(c, "bars unsubscribe is futures-mode only")
		return
	}
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		SafeBadRequest(c, "Invalid trader ID")
		return
	}
	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		SafeNotFound(c, "Trader")
		return
	}
	nt, ok := trader.GetUnderlyingTrader().(*ntTrader.TCPTrader)
	if !ok {
		SafeBadRequest(c, "trader is not an NT8 TCP trader (NT_TRANSPORT=tcp required)")
		return
	}
	symbol := c.Query("symbol")
	if symbol == "" {
		SafeBadRequest(c, "symbol is required")
		return
	}
	if err := nt.DebugUnsubscribeBars(symbol); err != nil {
		SafeInternalError(c, "NT bars unsubscribe", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "symbol": symbol})
}

// handleNTTestTrade places a deterministic 1-contract SIM bracket order on the
// NT trader for end-to-end dashboard proof (Plan 4.11: signal → NT8 SIM fill →
// position). Futures/SIM only; bypasses the AI + risk gate. NOT a trading path
// — a debug harness so the operator can see a real position + balance move on
// the dashboard without depending on an AI decision.
func (s *Server) handleNTTestTrade(c *gin.Context) {
	if cfg := config.Get(); cfg == nil || cfg.TradingMode != "futures" {
		SafeBadRequest(c, "test trade is futures-mode only")
		return
	}
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		SafeBadRequest(c, "Invalid trader ID")
		return
	}
	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		SafeNotFound(c, "Trader")
		return
	}
	nt, ok := trader.GetUnderlyingTrader().(*ntTrader.TCPTrader)
	if !ok {
		SafeBadRequest(c, "trader is not an NT8 TCP trader (NT_TRANSPORT=tcp required)")
		return
	}

	side := c.DefaultQuery("side", "long")
	if side != "long" && side != "short" {
		SafeBadRequest(c, "side must be long or short")
		return
	}

	result, err := nt.DebugPlaceTestTrade(side)
	if err != nil {
		SafeInternalError(c, "NT test trade", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "side": side, "result": result})
}
