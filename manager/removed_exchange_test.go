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

// D1 + D3 — the legacy rows survive BOOT untouched, and the BOOT loader records
// the refusal. The store is closed and reopened (New → initTables: the old-
// schema migration, the account-name backfill and the incomplete-config
// cleanup all run), then LoadTradersFromStore — main.go's boot path — loads.
// Rows: a DISABLED exchange row naming the removed broker (registry-first: it
// must be refused by name, not skipped as "not enabled"), and an OLD-SCHEMA row
// whose id is the removed broker's name with no type (the migration must not
// rewrite it into a typed row). No row is deleted or modified at boot.
func TestLegacyRemovedExchangeRowsSurviveBootAndTheBootLoaderRecordsTheRefusal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "boot-removed.db")
	st, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	const user = "u-boot-removed"
	if err := st.User().Create(&store.User{ID: user, Email: "boot-removed@test", PasswordHash: "x"}); err != nil {
		t.Fatal(err)
	}
	if err := st.AIModel().Create(user, "m-br", "m", "deepseek", true, "sk-test-not-a-real-key", ""); err != nil {
		t.Fatal(err)
	}
	if err := st.Strategy().Create(&store.Strategy{ID: "s-br", UserID: user, Name: "s", Config: "{}"}); err != nil {
		t.Fatal(err)
	}
	seed := []struct{ traderID, exID, exType, want string }{
		{"t-typed", "ex-typed-removed", "binance", `unsupported trading platform "binance"`},
		{"t-oldschema", "binance", "", "no exchange configured"},
	}
	for _, s := range seed {
		ex := &store.Exchange{ID: s.exID, ExchangeType: s.exType, UserID: user, Name: "legacy", Type: "cex", Enabled: false, APIKey: "k", SecretKey: "s"}
		if err := st.GormDB().Create(ex).Error; err != nil {
			t.Fatal(err)
		}
		if err := st.Trader().Create(&store.Trader{ID: s.traderID, UserID: user, Name: s.traderID, AIModelID: "m-br", ExchangeID: s.exID, StrategyID: "s-br", InitialBalance: 1000}); err != nil {
			t.Fatal(err)
		}
	}
	type snap struct {
		typ, account string
		enabled      bool
	}
	read := func(s *store.Store, id string) (snap, bool) {
		var ex store.Exchange
		if err := s.GormDB().Where("id = ? AND user_id = ?", id, user).Limit(1).Find(&ex).Error; err != nil {
			t.Fatal(err)
		}
		return snap{ex.ExchangeType, ex.AccountName, ex.Enabled}, ex.ID != ""
	}
	before := map[string]snap{}
	for _, s := range seed {
		b, ok := read(st, s.exID)
		if !ok {
			t.Fatalf("%s: seed row missing", s.exID)
		}
		before[s.exID] = b
	}
	st.Plan().Close()
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	st2, err := store.New(path) // BOOT
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st2.Plan().Close(); _ = st2.Close() })
	for _, s := range seed {
		a, ok := read(st2, s.exID)
		if !ok {
			t.Fatalf("%s: the legacy exchange row was DELETED at boot — a data.db write for a type the build no longer knows", s.exID)
		}
		if a != before[s.exID] {
			t.Errorf("%s: the legacy exchange row was MODIFIED at boot: %+v → %+v", s.exID, before[s.exID], a)
		}
		tr, err := st2.Trader().GetByID(s.traderID)
		if err != nil || tr.ExchangeID != s.exID {
			t.Errorf("%s: the trader's exchange binding was rewritten at boot (want %q): %+v %v", s.traderID, s.exID, tr, err)
		}
	}

	tm := NewTraderManager()
	if err := tm.LoadTradersFromStore(st2); err != nil { // main.go's boot loader
		t.Fatal(err)
	}
	for _, s := range seed {
		if _, err := tm.GetTrader(s.traderID); err == nil {
			t.Errorf("%s: must NOT load", s.traderID)
		}
		lerr := tm.GetLoadError(s.traderID)
		if lerr == nil || !strings.Contains(lerr.Error(), "refused") || !strings.Contains(lerr.Error(), s.want) {
			t.Errorf("%s: the BOOT loader must record the named refusal (%q) — GetLoadError=%v", s.traderID, s.want, lerr)
		}
	}
}
