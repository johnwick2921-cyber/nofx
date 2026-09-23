package api

import (
	"nofx/agent"

	"github.com/gin-gonic/gin"
)

// RegisterAgentHandler registers NOFXi agent API routes on the main router.
// Chat endpoints require authentication; /api/agent/health and the NinjaTrader
// price ticker (/api/agent/tickers) are public and read-only. The former
// external-exchange proxies /api/agent/klines and /api/agent/ticker are gone
// — the ticker reads NT8 bars only and names why a symbol is
// unavailable.
func (s *Server) RegisterAgentHandler(h *agent.WebHandler) {
	// Chat requires auth — can trigger trades and access account data
	s.router.POST("/api/agent/chat", s.authMiddleware(), func(c *gin.Context) {
		isAdmin := c.GetString("user_id") == "admin"
		ctx := agent.WithStoreUserID(c.Request.Context(), c.GetString("user_id"))
		ctx = agent.WithSessionPolicy(ctx, agent.SessionPolicy{
			Authenticated:           true,
			IsAdmin:                 isAdmin,
			CanExecuteTrade:         true,
			CanViewSensitiveSecrets: false,
		})
		req := c.Request.WithContext(ctx)
		h.HandleChat(c.Writer, req)
	})
	s.router.POST("/api/agent/chat/stream", s.authMiddleware(), func(c *gin.Context) {
		isAdmin := c.GetString("user_id") == "admin"
		ctx := agent.WithStoreUserID(c.Request.Context(), c.GetString("user_id"))
		ctx = agent.WithSessionPolicy(ctx, agent.SessionPolicy{
			Authenticated:           true,
			IsAdmin:                 isAdmin,
			CanExecuteTrade:         true,
			CanViewSensitiveSecrets: false,
		})
		req := c.Request.WithContext(ctx)
		h.HandleChatStream(c.Writer, req)
	})
	// Public endpoints — read-only; the ticker reads NinjaTrader bars only.
	s.router.GET("/api/agent/health", gin.WrapF(h.HandleHealth))
	s.router.GET("/api/agent/tickers", gin.WrapF(h.HandleTickers))
}
