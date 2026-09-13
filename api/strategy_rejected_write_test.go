package api

import (
	"encoding/json"
	"net/http"
	"nofx/store"
	"strings"
	"testing"
)

func TestRejectedStrategyUpdateDoesNotPersist(t *testing.T) {
	s, r := planModeStrategyServer(t)
	initial := &store.Strategy{ID: "atomic", UserID: "u1", Name: "Before", Config: `{"strategy_type":"ai_trading"}`}
	if err := s.store.Strategy().Create(initial); err != nil {
		t.Fatal(err)
	}
	maxContext := 0
	for _, n := range store.ModelContextLimits {
		if n > maxContext {
			maxContext = n
		}
	}
	body, err := json.Marshal(map[string]any{"name": "Rejected", "config": map[string]any{"custom_prompt": strings.Repeat("x", maxContext*5)}})
	if err != nil {
		t.Fatal(err)
	}
	rec := doStrategyRequest(r, http.MethodPut, "/api/strategies/atomic", string(body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected token rejection, got %d", rec.Code)
	}
	got, err := s.store.Strategy().Get("u1", "atomic")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != initial.Name || got.Config != initial.Config {
		t.Fatal("rejected request changed persisted strategy")
	}
}
