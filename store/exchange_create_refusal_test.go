package store

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// ExchangeStore.Create consults the supported-exchange registry FIRST: a type
// this build cannot construct is refused by NAME (the value interpolated, and
// errors.Is ErrUnsupportedExchangeType) — never the misleading "missing
// required exchange fields: exchange_type" the credential check's default
// branch used to answer — and nothing is written. Supported types keep their
// existing casing tolerance.
func TestCreateRefusesAnUnsupportedExchangeTypeByName(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "create.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Plan().Close(); _ = st.Close() })
	const user = "u-create-refusal"

	for _, typ := range []string{"retired-venue", ""} {
		id, err := st.Exchange().Create(user, typ, "A", true, "k", "s", "", false, "", false, "", "", "", "", "", "", 0, "", "", 0)
		if err == nil {
			t.Fatalf("type %q must be refused, got id %q", typ, id)
		}
		if !errors.Is(err, ErrUnsupportedExchangeType) || !strings.Contains(err.Error(), `"`+typ+`"`) || strings.Contains(err.Error(), "missing required") {
			t.Fatalf("type %q: want the named registry refusal, got %v", typ, err)
		}
	}
	list, err := st.Exchange().List(user)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("a refused create must write nothing, found %d rows", len(list))
	}

	// Controls: a supported type still creates, in any casing it did before.
	for _, typ := range []string{"bybit", "Bybit"} {
		if _, err := st.Exchange().Create(user, typ, "acct-"+typ, true, "k", "s", "", false, "", false, "", "", "", "", "", "", 0, "", "", 0); err != nil {
			t.Fatalf("supported type %q must still create: %v", typ, err)
		}
	}
}
