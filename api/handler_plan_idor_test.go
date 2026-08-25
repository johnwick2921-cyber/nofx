package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"nofx/store"

	"github.com/gin-gonic/gin"
)

// C1 (2026-08-25) — the /api/plan/* ownership middleware must 404 (never 403,
// never 200) when the JWT user does not own the queried trader.
func TestPlanTraderOwnershipBlocksCrossUser(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "idor.db"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	if err := st.Trader().Create(&store.Trader{ID: "a_trader", UserID: "user-a", Name: "A", AIModelID: "m", ExchangeID: "e"}); err != nil {
		t.Fatalf("create trader A: %v", err)
	}
	if err := st.Trader().Create(&store.Trader{ID: "b_trader", UserID: "user-b", Name: "B", AIModelID: "m", ExchangeID: "e"}); err != nil {
		t.Fatalf("create trader B: %v", err)
	}

	s := &Server{store: st}
	router := gin.New()
	router.GET("/api/plan/today",
		func(c *gin.Context) { c.Set("user_id", c.Query("as_user")); c.Next() },
		s.planTraderOwnership(),
		func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	cases := []struct {
		name     string
		asUser   string
		traderID string
		want     int
	}{
		{"cross-user read blocked", "user-b", "a_trader", http.StatusNotFound},
		{"owner read allowed", "user-a", "a_trader", http.StatusOK},
		{"empty trader_id passes to handler", "user-a", "", http.StatusOK},
	}
	for _, tc := range cases {
		url := "/api/plan/today?as_user=" + tc.asUser
		if tc.traderID != "" {
			url += "&trader_id=" + tc.traderID
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
		if rec.Code != tc.want {
			t.Fatalf("%s: expected %d, got %d body=%s", tc.name, tc.want, rec.Code, rec.Body.String())
		}
	}
}

// C3 (2026-08-25) — getTraderFromQuery: no global-first-trader fallback and no
// cross-user trader resolution. A user with no traders gets an error; an
// explicit trader of ANOTHER user gets an error.
func TestGetTraderFromQueryNoGlobalFallback(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "gtr.db"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	if err := st.Trader().Create(&store.Trader{ID: "a_trader", UserID: "user-a", Name: "A", AIModelID: "m", ExchangeID: "e"}); err != nil {
		t.Fatalf("create trader A: %v", err)
	}

	s := &Server{store: st}
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.Set("user_id", c.Query("as_user"))
		_, traderID, err := s.getTraderFromQuery(c)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no trader"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"trader_id": traderID})
	})

	cases := []struct {
		name   string
		asUser string
		query  string
		want   int
		wantID string
	}{
		{"defaults to own first trader", "user-a", "", http.StatusOK, "a_trader"},
		{"explicit own trader ok", "user-a", "&trader_id=a_trader", http.StatusOK, "a_trader"},
		{"cross-user trader refused", "user-a", "&trader_id=b_trader", http.StatusNotFound, ""},
		{"user with no traders refused", "user-x", "", http.StatusNotFound, ""},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test?as_user="+tc.asUser+tc.query, nil))		if rec.Code != tc.want {
			t.Fatalf("%s: expected %d, got %d body=%s", tc.name, tc.want, rec.Code, rec.Body.String())
		}
		if tc.wantID != "" && !strings.Contains(rec.Body.String(), tc.wantID) {
			t.Fatalf("%s: response %q missing trader id %q", tc.name, rec.Body.String(), tc.wantID)
		}
	}
}
