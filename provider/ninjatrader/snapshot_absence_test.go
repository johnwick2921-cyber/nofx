package ninjatrader

import "testing"

func TestSnapshotMissingOrdersIsNotEmptyBook(t *testing.T) {
	for _, payload := range []string{`{"account":"Sim101"}`, `{"account":"Sim101","orders":null}`} {
		if p, err := ParseOrderSnapshot([]byte(payload)); err == nil {
			t.Errorf("uncomputed book accepted as empty: payload=%s result=%+v", payload, p)
		}
	}
	if p, err := ParseOrderSnapshot([]byte(`{"account":"Sim101","orders":[]}`)); err != nil || p.Orders == nil {
		t.Fatalf("explicit empty book must remain valid: %+v %v", p, err)
	}
}
