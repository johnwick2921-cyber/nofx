package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"nofx/auth"
	"nofx/manager"
	"nofx/store"
	"strings"
	"testing"
)

func TestOrderFillsProductionRouteScopesOrderAndFills(t *testing.T) {
	s := newSecurityTestServer(t)
	for _, row := range []store.Trader{{ID: "own", UserID: "reader", Name: "Own", AIModelID: "m", ExchangeID: "e"}, {ID: "foreign", UserID: "owner", Name: "Foreign", AIModelID: "m", ExchangeID: "e"}} {
		if err := s.store.Trader().Create(&row); err != nil {
			t.Fatal(err)
		}
	}
	own := store.TraderOrder{TraderID: "own", ExchangeOrderID: "own-order", Symbol: "MNQ", Side: "BUY", Type: "LIMIT", Quantity: 1}
	foreign := store.TraderOrder{TraderID: "foreign", ExchangeOrderID: "foreign-order", Symbol: "MNQ", Side: "BUY", Type: "LIMIT", Quantity: 1}
	for _, row := range []*store.TraderOrder{&own, &foreign} {
		if err := s.store.Order().CreateOrder(row); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []store.TraderFill{{TraderID: "own", OrderID: own.ID, ExchangeTradeID: "own-fill", ExchangeOrderID: "own-order", Symbol: "MNQ", Side: "BUY", Price: 100, Quantity: 1}, {TraderID: "foreign", OrderID: foreign.ID, ExchangeTradeID: "foreign-fill", ExchangeOrderID: "foreign-order", Symbol: "MNQ", Side: "BUY", Price: 900, Quantity: 1}, {TraderID: "foreign", OrderID: own.ID, ExchangeTradeID: "malformed-foreign-fill", ExchangeOrderID: "own-order", Symbol: "MNQ", Side: "BUY", Price: 999, Quantity: 1}} {
		if err := s.store.Order().CreateFill(&f); err != nil {
			t.Fatal(err)
		}
	}
	prevSecret := append([]byte(nil), auth.JWTSecret...)
	auth.SetJWTSecret("offline-order-fill-scope")
	t.Cleanup(func() { auth.JWTSecret = prevSecret })
	token, err := auth.GenerateJWT("reader", "reader@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	previousRoutes := routeRegistry
	t.Cleanup(func() { routeRegistry = previousRoutes })
	live := NewServer(manager.NewTraderManager(), s.store, nil, "", 0)
	for _, tc := range []struct {
		id   int64
		code int
	}{{own.ID, http.StatusOK}, {foreign.ID, http.StatusNotFound}} {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/orders/%d/fills?trader_id=own", tc.id), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		live.router.ServeHTTP(rec, req)
		if rec.Code != tc.code {
			t.Fatalf("order%d: code=%d want=%d body=%s", tc.id, rec.Code, tc.code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "foreign-fill") {
			t.Fatalf("foreign fill exposed: %s", rec.Body.String())
		}
		if tc.code == http.StatusOK && !strings.Contains(rec.Body.String(), "own-fill") {
			t.Fatalf("owned fills missing: %s", rec.Body.String())
		}
	}
}
