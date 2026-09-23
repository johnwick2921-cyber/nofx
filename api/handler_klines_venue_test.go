package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// ── /api/klines: the venue is REQUIRED; no default, no cross-venue read ────
//
// Through the production route (router → JWT → handleKlines): a request with
// no exchange is a named 400 (there is no default venue); an exchange with no
// CoinAnk book of its own is a named 400 with ZERO outbound requests; a
// CoinAnk failure on the right venue is ONE request, never a second read of
// another venue's book. http.DefaultTransport (what the CoinAnk client uses)
// is replaced so nothing leaves the process.

type klinesVenueTrap struct {
	mu     sync.Mutex
	venues []string
}

func (k *klinesVenueTrap) RoundTrip(r *http.Request) (*http.Response, error) {
	k.mu.Lock()
	k.venues = append(k.venues, r.URL.Query().Get("exchange"))
	k.mu.Unlock()
	return nil, errors.New("offline: no network in this test")
}

func klinesGet(t *testing.T, s *Server, tok, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/klines?"+query, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

func TestKlinesRouteRefusesAMissingOrBooklessVenue(t *testing.T) {
	s, tok := newRollHoleServer(t)
	trap := &klinesVenueTrap{}
	prev := http.DefaultTransport
	http.DefaultTransport = trap
	t.Cleanup(func() { http.DefaultTransport = prev })

	for _, q := range []string{"symbol=BTCUSDT&interval=5m&limit=10", "symbol=BTCUSDT&interval=5m&limit=10&exchange=%20"} {
		rec := klinesGet(t, s, tok, q)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "exchange parameter is required") {
			t.Fatalf("%s: a missing exchange must be a named 400 (no default venue), got %d %s", q, rec.Code, rec.Body.String())
		}
	}
	for _, ex := range []string{"kucoin", "lighter", "indodax"} {
		rec := klinesGet(t, s, tok, "symbol=BTCUSDT&interval=5m&limit=10&exchange="+ex)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `\"`+ex+`\"`) || !strings.Contains(rec.Body.String(), "NO_KLINE_VENUE") {
			t.Fatalf("exchange %s has no CoinAnk book: want a 400 naming it, got %d %s", ex, rec.Code, rec.Body.String())
		}
	}
	trap.mu.Lock()
	n := len(trap.venues)
	trap.mu.Unlock()
	if n != 0 {
		t.Fatalf("refused venues must make ZERO outbound requests, got %v", trap.venues)
	}

	// A venue WITH a book that fails: one request on that venue, then an
	// error — never a second request to another venue.
	rec := klinesGet(t, s, tok, "symbol=BTCUSDT&interval=5m&limit=10&exchange=bybit")
	if rec.Code == http.StatusOK {
		t.Fatalf("an offline CoinAnk must not answer 200, got %s", rec.Body.String())
	}
	trap.mu.Lock()
	defer trap.mu.Unlock()
	if len(trap.venues) != 1 || trap.venues[0] != "Bybit" {
		t.Fatalf("exactly ONE request on Bybit — no cross-venue fallback; got %v", trap.venues)
	}

	// Control: the futures venue still serves the NT8 ring by name.
	if rec := klinesGet(t, s, tok, "symbol=MNQ&interval=5m&limit=10&exchange=ninjatrader"); rec.Code != http.StatusOK {
		t.Fatalf("control: ninjatrader klines must still answer 200, got %d %s", rec.Code, rec.Body.String())
	}
}
