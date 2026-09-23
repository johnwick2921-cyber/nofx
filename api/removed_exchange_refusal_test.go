package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"nofx/auth"
	"nofx/manager"
	"nofx/store"
)

// ── A stored exchange row of a type this build cannot construct ─────────────
//
// A broker removed from the build can still be named by a legacy row in
// data.db. Every API surface that meets that row must REFUSE it by name — the
// stored type interpolated — never report "disabled" or "missing exchange_type"
// and never re-enable it:
//   - POST /api/traders/:id/start → trader.start.load_failed, reason_key
//     trader.reason.exchange_unsupported, and the reason NAMES the type (the
//     classifier used to replace it with a fixed sentence);
//   - PUT /api/exchanges on the row → 400 UNSUPPORTED_EXCHANGE naming it, row
//     unchanged;
//   - GET /api/exchanges/account-state → UNSUPPORTED_EXCHANGE naming it, even
//     when the row is disabled;
//   - POST /api/exchanges creating that type → 400.
// The type name is arbitrary on purpose ("retired-venue"): the rule is the
// registry (store.IsSupportedExchangeType), not a list of removed names. The
// rows are written straight to the table — the way a legacy database holds
// them — and loaded through the production router, JWT and TraderManager.

const (
	rxUser    = "u-retired-exchange"
	rxType    = "retired-venue"
	rxTrader  = "t-retired"
	rxEnabled = "ex-retired-enabled"
	rxOff     = "ex-retired-disabled"
)

func newRetiredExchangeServer(t *testing.T) (*Server, *store.Store, string) {
	t.Helper()
	st, err := store.New(filepath.Join(t.TempDir(), "rx.db"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { st.Plan().Close(); _ = st.Close() })
	if err := st.AIModel().Create(rxUser, "m-rx", "m", "deepseek", true, "sk-test-not-a-real-key", ""); err != nil {
		t.Fatalf("ai model: %v", err)
	}
	if err := st.Strategy().Create(&store.Strategy{ID: "s-rx", UserID: rxUser, Name: "s", Config: "{}"}); err != nil {
		t.Fatalf("strategy: %v", err)
	}
	for _, ex := range []*store.Exchange{
		{ID: rxEnabled, ExchangeType: rxType, UserID: rxUser, AccountName: "A", Name: "legacy", Type: "cex", Enabled: true, APIKey: "k", SecretKey: "s"},
		{ID: rxOff, ExchangeType: rxType, UserID: rxUser, AccountName: "B", Name: "legacy", Type: "cex", Enabled: false, APIKey: "k", SecretKey: "s"},
	} {
		if err := st.GormDB().Create(ex).Error; err != nil {
			t.Fatalf("seed exchange: %v", err)
		}
	}
	for id, exID := range map[string]string{rxTrader: rxEnabled, rxTrader + "-off": rxOff} {
		if err := st.Trader().Create(&store.Trader{ID: id, UserID: rxUser, Name: id, AIModelID: "m-rx", ExchangeID: exID, StrategyID: "s-rx", InitialBalance: 1000}); err != nil {
			t.Fatalf("trader: %v", err)
		}
	}
	tm := manager.NewTraderManager()
	auth.SetJWTSecret("retired-exchange-test-secret")
	tok, err := auth.GenerateJWT(rxUser, "rx@test")
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	return NewServer(tm, st, nil, "127.0.0.1", 0), st, tok
}

func rxDo(t *testing.T, s *Server, tok, method, url, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec, out
}

func TestStartOnARetiredExchangeTypeIsANamedLoadRefusal(t *testing.T) {
	s, _, tok := newRetiredExchangeServer(t)
	for _, id := range []string{rxTrader, rxTrader + "-off"} {
		rec, out := rxDo(t, s, tok, http.MethodPost, "/api/traders/"+id+"/start", "")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: start must be refused with 400, got %d %s", id, rec.Code, rec.Body.String())
		}
		if out["error_key"] != "trader.start.load_failed" {
			t.Fatalf("%s: want error_key trader.start.load_failed (not disabled / setup_invalid), got %v — %s", id, out["error_key"], rec.Body.String())
		}
		params, _ := out["error_params"].(map[string]any)
		if params["reason_key"] != "trader.reason.exchange_unsupported" {
			t.Fatalf("%s: want reason_key trader.reason.exchange_unsupported, got %v", id, params["reason_key"])
		}
		if reason, _ := params["reason"].(string); !strings.Contains(reason, `"`+rxType+`"`) {
			t.Fatalf("%s: the reason must NAME the stored exchange type %q (the manager's refusal reaches the response), got %q", id, rxType, reason)
		}
		if msg, _ := out["error"].(string); !strings.Contains(msg, rxType) {
			t.Fatalf("%s: the public message must name the stored type too, got %q", id, msg)
		}
	}
	if _, err := s.traderManager.GetTrader(rxTrader); err == nil {
		t.Fatal("a trader on a retired exchange type must never be in memory")
	}
}

func TestUpdateOfARetiredExchangeRowIsANamedRefusalAndWritesNothing(t *testing.T) {
	s, st, tok := newRetiredExchangeServer(t)
	body := `{"exchanges":{"` + rxOff + `":{"enabled":true,"api_key":"k2","secret_key":"s2"}}}`
	rec, out := rxDo(t, s, tok, http.MethodPut, "/api/exchanges", body)
	if rec.Code != http.StatusBadRequest || out["code"] != "UNSUPPORTED_EXCHANGE" {
		t.Fatalf("update of a retired-type row must be a 400 UNSUPPORTED_EXCHANGE, got %d %s", rec.Code, rec.Body.String())
	}
	if msg, _ := out["error"].(string); !strings.Contains(msg, `"`+rxType+`"`) || strings.Contains(msg, "missing") {
		t.Fatalf("the refusal must name the stored type and not claim missing fields, got %q", msg)
	}
	ex, err := st.Exchange().GetByID(rxUser, rxOff)
	if err != nil {
		t.Fatalf("the row must still exist: %v", err)
	}
	if ex.Enabled || string(ex.APIKey) != "k" {
		t.Fatalf("a refused update must write nothing: enabled=%v", ex.Enabled)
	}
}

func TestAccountStateNamesARetiredExchangeType(t *testing.T) {
	s, _, tok := newRetiredExchangeServer(t)
	rec, out := rxDo(t, s, tok, http.MethodGet, "/api/exchanges/account-state", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("account-state: %d %s", rec.Code, rec.Body.String())
	}
	states, _ := out["states"].(map[string]any)
	for _, id := range []string{rxEnabled, rxOff} {
		st, _ := states[id].(map[string]any)
		if st == nil {
			t.Fatalf("%s: no state reported", id)
		}
		if st["error_code"] != "UNSUPPORTED_EXCHANGE" || st["status"] != exchangeAccountStatusUnavailable {
			t.Fatalf("%s: want unavailable/UNSUPPORTED_EXCHANGE (the registry is consulted before the enabled check), got %v", id, st)
		}
		if msg, _ := st["error_message"].(string); !strings.Contains(msg, `"`+rxType+`"`) {
			t.Fatalf("%s: the state must name the stored type, got %q", id, msg)
		}
	}
}

func TestCreatingAnExchangeOfAnUnsupportedTypeIsRefused(t *testing.T) {
	s, _, tok := newRetiredExchangeServer(t)
	rec, _ := rxDo(t, s, tok, http.MethodPost, "/api/exchanges", `{"exchange_type":"`+rxType+`","account_name":"C","enabled":true,"api_key":"k","secret_key":"s"}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), rxType) {
		t.Fatalf("create of an unsupported type must be a named 400, got %d %s", rec.Code, rec.Body.String())
	}
	// Control: every registry type passes the same gate.
	for _, typ := range store.SupportedExchangeTypes() {
		if !store.IsSupportedExchangeType(typ) {
			t.Fatalf("registry type %q must be supported", typ)
		}
	}
}

// The classifier keeps the refusal's own text: the reason names what the
// manager named. A fixed sentence alone would lose the stored type.
func TestClassifyTraderSetupReasonKeepsTheNamedRefusal(t *testing.T) {
	raw := `refused: unsupported trading platform "` + rxType + `" — this trader does not load`
	key, msg := classifyTraderSetupReason(raw)
	if key != "trader.reason.exchange_unsupported" || !strings.Contains(msg, `"`+rxType+`"`) {
		t.Fatalf("got key=%q msg=%q", key, msg)
	}
	key, msg = classifyTraderSetupReason(`refused: trader "x" has no exchange configured — a trader never gets a broker by default`)
	if key != "trader.reason.exchange_missing" || !strings.Contains(msg, "no exchange configured") {
		t.Fatalf("empty exchange: got key=%q msg=%q", key, msg)
	}
}
