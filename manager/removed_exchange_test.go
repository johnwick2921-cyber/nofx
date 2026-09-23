package manager

import (
	"path/filepath"
	"strings"
	"testing"

	"nofx/store"
)

// ── W-NO-BINANCE B — migration truth ────────────────────────────────────────
//
// The Binance broker is deleted. A trader row that still names it (a legacy
// row in data.db) must load as a REFUSED trader with a named reason — never a
// crash, never silently coerced onto another broker. A trader row with NO
// exchange type used to be HANDED the Binance broker by default; it is now
// refused too (CTO ruling). The rows are written straight to the table, the
// way a legacy database holds them (the store's Create validation would refuse
// them today). This file is on the zero-grep guard's allowlist: it is the one
// place that must spell the removed exchange's name to prove the refusal.
func TestLegacyTraderRowsForARemovedOrEmptyExchangeLoadAsRefused(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "removed.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Plan().Close(); _ = st.Close() })
	const user = "u-removed-exchange"
	if err := st.User().Create(&store.User{ID: user, Email: "removed@test", PasswordHash: "x"}); err != nil {
		t.Fatal(err)
	}
	if err := st.AIModel().Create(user, "m-rm", "m", "deepseek", true, "sk-test-not-a-real-key", ""); err != nil {
		t.Fatal(err)
	}
	if err := st.Strategy().Create(&store.Strategy{ID: "s-rm", UserID: user, Name: "s", Config: "{}"}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ id, exchangeType, want string }{
		{"t-removed", "binance", `unsupported trading platform "binance"`},
		{"t-empty", "", "no exchange configured"},
	} {
		ex := &store.Exchange{ID: "ex-" + tc.id, ExchangeType: tc.exchangeType, UserID: user, Name: "legacy", Type: "cex", Enabled: true}
		if err := st.GormDB().Create(ex).Error; err != nil {
			t.Fatal(err)
		}
		if err := st.Trader().Create(&store.Trader{ID: tc.id, UserID: user, Name: tc.id, AIModelID: "m-rm", ExchangeID: ex.ID, StrategyID: "s-rm", InitialBalance: 1000}); err != nil {
			t.Fatal(err)
		}
	}
	tm := NewTraderManager()
	_ = tm.LoadUserTradersFromStore(st, user) // must not panic
	for _, tc := range []struct{ id, want string }{
		{"t-removed", `unsupported trading platform "binance"`},
		{"t-empty", "no exchange configured"},
	} {
		if _, err := tm.GetTrader(tc.id); err == nil {
			t.Errorf("%s: a trader on a removed or empty exchange must NOT load", tc.id)
		}
		lerr := tm.GetLoadError(tc.id)
		if lerr == nil || !strings.Contains(lerr.Error(), "refused") || !strings.Contains(lerr.Error(), tc.want) {
			t.Errorf("%s: the load error must name the refusal (%q), got %v", tc.id, tc.want, lerr)
		}
	}
}
