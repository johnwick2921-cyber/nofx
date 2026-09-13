package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nofx/auth"
	"nofx/manager"
	"nofx/store"
)

// Exercise the actual setupRoutes/auth/middleware/handler chain, not a copy of it.
func TestRepositoryOwnershipProductionRoutes(t *testing.T) {
	s := newSecurityTestServer(t)
	for _, row := range []store.Trader{{ID: "own", UserID: "reader", Name: "Own", AIModelID: "m", ExchangeID: "e"}, {ID: "foreign", UserID: "owner", Name: "Foreign", AIModelID: "m", ExchangeID: "e"}} {
		if err := s.store.Trader().Create(&row); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.store.Equity().Save(&store.EquitySnapshot{TraderID: "foreign", TotalEquity: 123}); err != nil {
		t.Fatal(err)
	}
	previousSecret := append([]byte(nil), auth.JWTSecret...)
	auth.SetJWTSecret("offline-repository-ownership-test")
	t.Cleanup(func() { auth.JWTSecret = previousSecret })
	token, err := auth.GenerateJWT("reader", "reader@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	previousRoutes := routeRegistry
	t.Cleanup(func() { routeRegistry = previousRoutes })
	live := NewServer(manager.NewTraderManager(), s.store, nil, "", 0)
	for _, tc := range []struct{ method, url, body string }{
		{"GET", "/api/config/resolved?trader_id=foreign", ""},
		{"GET", "/api/desk?trader_id=foreign", ""},
		{"GET", "/api/accounts?trader_id=foreign", ""},
		{"GET", "/api/ai-costs?trader_id=foreign", ""},
		{"GET", "/api/audit/decisions?trader_id=foreign", ""},
		{"GET", "/api/traders/foreign/grid-risk", ""},
		{"POST", "/api/account/select?trader_id=foreign", `{"account":"Sim101"}`},
		{"POST", "/api/plan/overlay?trader_id=own", `{"trader_id":"foreign"}`},
		{"DELETE", "/api/traders/foreign", ""},
	} {
		t.Run(tc.method+tc.url, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			live.router.ServeHTTP(rec, req)
			if rec.Code != http.StatusNotFound {
				t.Errorf("expected ownership 404, got %d: %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"error":"Trader not found"`) {
				t.Errorf("expected ownership refusal, not a missing runtime or plan: %s", rec.Body.String())
			}
		})
	}
	request := httptest.NewRequest("GET", "/api/config/resolved?trader_id=own", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	live.router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("own trader read must remain available: %d", response.Code)
	}
	var count int64
	if err := s.store.GormDB().Model(&store.EquitySnapshot{}).Where("trader_id = ?", "foreign").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("foreign equity modified: count=%d", count)
	}
	if row, err := s.store.Trader().GetByID("foreign"); err != nil || row == nil {
		t.Fatalf("foreign trader modified: %v", err)
	}
}
