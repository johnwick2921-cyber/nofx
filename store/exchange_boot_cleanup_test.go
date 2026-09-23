package store

import (
	"errors"
	"path/filepath"
	"testing"
)

// A boot cleanup must never delete or modify an exchange row because its type
// is one this build does not construct (a broker removed from the build, or no
// type at all). The rows are written straight to the table — the way a legacy
// database holds them — then the store is CLOSED and REOPENED: New → initTables
// is the production boot path that runs migrateToMultiAccount, the account-name
// backfill and cleanupIncompleteExchangeConfigs. Supported-type rows in the same
// table prove the cleanup really ran (an incomplete one is still deleted, a
// complete disabled one is still enabled), so survival is not vacuous.
func TestBootCleanupLeavesAnUnsupportedExchangeRowUntouched(t *testing.T) {
	path := filepath.Join(t.TempDir(), "boot.db")
	st, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	const user = "u-boot-cleanup"
	rows := []*Exchange{
		// Unsupported: a type the registry does not list (the name is arbitrary
		// on purpose — the rule is the registry, not a list of removed names).
		{ID: "ex-retired", ExchangeType: "retired-venue", UserID: user, Name: "legacy", Type: "cex", Enabled: false, APIKey: "k", SecretKey: "s"},
		// Unsupported: no type at all (used to be DELETED by the default branch).
		{ID: "ex-untyped", ExchangeType: "", UserID: user, Name: "legacy", Type: "cex", Enabled: false},
		// Negative controls, supported type: prove the cleanup ran.
		{ID: "ex-bybit-incomplete", ExchangeType: "bybit", UserID: user, Name: "Bybit Futures", Type: "cex", Enabled: true, AccountName: "A"},
		{ID: "ex-bybit-complete", ExchangeType: "bybit", UserID: user, Name: "Bybit Futures", Type: "cex", Enabled: false, AccountName: "B", APIKey: "k", SecretKey: "s"},
	}
	for _, r := range rows {
		if err := st.GormDB().Create(r).Error; err != nil {
			t.Fatal(err)
		}
	}
	read := func(s *Store, id string) (*Exchange, bool) {
		var ex Exchange
		if err := s.GormDB().Where("id = ?", id).Limit(1).Find(&ex).Error; err != nil {
			t.Fatal(err)
		}
		return &ex, ex.ID != ""
	}
	before := map[string]Exchange{}
	for _, id := range []string{"ex-retired", "ex-untyped"} {
		ex, ok := read(st, id)
		if !ok {
			t.Fatalf("%s: seed row missing", id)
		}
		before[id] = *ex
	}
	st.Plan().Close()
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	// Boot.
	st2, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st2.Plan().Close(); _ = st2.Close() })

	for id, b := range before {
		a, ok := read(st2, id)
		if !ok {
			t.Fatalf("%s: an exchange row of an unsupported type was DELETED at boot — a boot cleanup must never write a row because its type is unknown", id)
		}
		if a.ExchangeType != b.ExchangeType || a.Enabled != b.Enabled || a.AccountName != b.AccountName || !a.UpdatedAt.Equal(b.UpdatedAt) || string(a.APIKey) != string(b.APIKey) {
			t.Errorf("%s: row MODIFIED at boot: before type=%q enabled=%v account=%q updated=%v, after type=%q enabled=%v account=%q updated=%v",
				id, b.ExchangeType, b.Enabled, b.AccountName, b.UpdatedAt, a.ExchangeType, a.Enabled, a.AccountName, a.UpdatedAt)
		}
	}
	if _, ok := read(st2, "ex-bybit-incomplete"); ok {
		t.Error("control: an incomplete SUPPORTED row must still be cleaned up at boot (the cleanup did not run?)")
	}
	if c, ok := read(st2, "ex-bybit-complete"); !ok || !c.Enabled {
		t.Error("control: a complete SUPPORTED row must still be enabled at boot (the cleanup did not run?)")
	}
}

// The registry and the credential switch move together: every supported type
// has its own credential case (a type that fell to the default would be
// reported as "missing exchange_type" — the shape that used to delete rows),
// the old-schema migration only rewrites supported types, and the refusal is
// named with the stored value.
func TestExchangeRegistryParity(t *testing.T) {
	for _, typ := range SupportedExchangeTypes() {
		m := MissingRequiredExchangeCredentialFields(typ, "", "", "", "", "", "", "", "", "", "")
		if len(m) == 1 && m[0] == "exchange_type" {
			t.Errorf("%q is in the registry but has no credential case", typ)
		}
		if err := CheckSupportedExchangeType(typ); err != nil {
			t.Errorf("%q: %v", typ, err)
		}
	}
	for _, id := range legacyExchangeTypeIDs {
		if !IsSupportedExchangeType(id) {
			t.Errorf("old-schema migration would rewrite %q, which this build does not construct", id)
		}
	}
	for _, bad := range []string{"", "Bybit", " bybit", "retired-venue"} {
		err := CheckSupportedExchangeType(bad)
		if err == nil || !errors.Is(err, ErrUnsupportedExchangeType) {
			t.Fatalf("%q must be refused with ErrUnsupportedExchangeType, got %v", bad, err)
		}
		if want := `unsupported exchange_type "` + bad + `"`; len(err.Error()) < len(want) || err.Error()[:len(want)] != want {
			t.Errorf("refusal must name the stored value: want prefix %q, got %q", want, err.Error())
		}
	}
}
