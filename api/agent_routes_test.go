package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"nofx/agent"

	"github.com/gin-gonic/gin"
)

// agentRoutesServer registers the agent routes exactly as main.go does
// (Server.RegisterAgentHandler) on a bare router.
func agentRoutesServer() *Server {
	gin.SetMode(gin.TestMode)
	s := &Server{router: gin.New()}
	s.RegisterAgentHandler(agent.NewWebHandler(nil, slog.Default()))
	return s
}

// The public external-exchange proxies are gone: nothing is registered at
// /api/agent/klines or /api/agent/ticker.
func TestAgentRoutes_ExternalExchangeProxiesAreNotRegistered(t *testing.T) {
	s := agentRoutesServer()
	for _, path := range []string{"/api/agent/klines?symbol=MNQ", "/api/agent/ticker?symbol=MNQ"} {
		rec := httptest.NewRecorder()
		s.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET %s = %d, want 404 (route removed) — %s", path, rec.Code, rec.Body.String())
		}
	}
}

// /api/agent/tickers stays public and answers a non-CME symbol with the
// contract row: price null + the named NinjaTrader-only reason.
func TestAgentRoutes_TickersNonCMEIsNamedUnavailable(t *testing.T) {
	s := agentRoutesServer()
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/agent/tickers?symbols=BTCUSDT", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — %s", rec.Code, rec.Body.String())
	}
	var rows []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil {
		t.Fatalf("body is not a JSON array: %v — %s", err, rec.Body.String())
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %s", rec.Body.String())
	}
	row := rows[0]
	if row["symbol"] != "BTCUSDT" || row["source"] != "nt8" {
		t.Fatalf("row = %v", row)
	}
	if v, present := row["price"]; !present || v != nil {
		t.Fatalf("price = %v (present=%v), want a present null", v, present)
	}
	if row["unavailable"] != "only CME futures symbols are available (NinjaTrader market data)" {
		t.Fatalf("unavailable = %v", row["unavailable"])
	}
}
