package api

import (
	"net/http"
	"time"

	"nofx/kernel"
	"nofx/market"

	"github.com/gin-gonic/gin"
)

// handleKlinesSVP serves the server-computed Session Volume Profile (SVP) for a
// CME futures symbol: POC / VAH / VAL and the histogram bins for the developing
// and prior RTH sessions.
//
// ONE source of truth: it reads the SAME live NT8 bars the AI kernel uses
// (market.FuturesBarsProvider at the shared kernel.SVPBarInterval / SVPBarCount)
// and runs the SAME kernel.BuildSVPProfile engine — the chart never recomputes
// the profile itself. The SVP is always built from the fixed SVPBarInterval (1m,
// the interval the AI uses) regardless of the chart's display interval, so the
// ?interval query param is intentionally ignored here; this is what guarantees
// the POC/VAH/VAL the trader sees equals the POC/VAH/VAL the AI acts on.
//
// A cold/empty cache, an unbound provider, or a non-futures symbol returns a
// well-formed zero-value 200 (empty bins) — NEVER a 500 and never a crypto
// fallthrough — mirroring getKlinesFromNinjaTrader so the chart degrades to
// "no profile" gracefully.
func (s *Server) handleKlinesSVP(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		symbol = "MNQ"
	}

	empty := kernel.SVPProfile{RowHeight: kernel.SVPRowHeight, Sessions: []kernel.SVPSession{}}

	if !market.IsCMEFuturesSymbol(symbol) {
		c.JSON(http.StatusOK, empty)
		return
	}
	provider := market.FuturesBarsProvider
	if provider == nil {
		c.JSON(http.StatusOK, empty)
		return
	}
	// ONE source of truth: build the profile from the fixed SVPBarInterval /
	// SVPBarCount the AI prompt uses (kernel/engine_analysis.go), NOT the chart's
	// display timeframe. The ?interval query param is intentionally IGNORED — this
	// is what guarantees the POC/VAH/VAL the trader sees equals the values the AI
	// acts on, regardless of which candle timeframe the chart is set to.
	// 1m/2000 matches the AI's exact input (~2 session profiles / ~1.4 days).
	bars := provider(symbol, kernel.SVPBarInterval, kernel.SVPBarCount)
	if len(bars) == 0 {
		c.JSON(http.StatusOK, empty)
		return
	}

	c.JSON(http.StatusOK, kernel.BuildSVPProfile(bars, time.Now()))
}
