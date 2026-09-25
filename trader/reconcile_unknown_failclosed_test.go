// CTO addendum 2 (2026-09-25, CI job Backend Tests at 83f9fe1b): the reconcile
// consumer logged 'positions read failed — reconcile treats as flat' — the
// exact unknown-as-flat reading F4 exists to kill. ntHeldPosition returned ""
// on a read error and reconcileBeforeOpenNTReport read "" as "NT8 flat →
// proceed": the open went ahead on an UNREADABLE book. Fail-closed contract:
// unknown refuses the open (and never submits a flatten); only a POSITIVE
// empty book is flat.
package trader

import (
	"errors"
	"strings"
	"testing"
)

func TestReconcileBeforeOpenNTUnknownPositionsRefuses(t *testing.T) {
	at := &AutoTrader{
		id:       "recon-unknown",
		exchange: "ninjatrader",
		trader:   &stubTrader{err: errors.New("NT8 account positions unknown: no account snapshot or confirmed entry fill")},
	}
	flattenSent, err := at.reconcileBeforeOpenNTReport("MNQ", "long")
	if err == nil {
		t.Fatal("an UNKNOWN positions read must refuse the open — reconcile treated it as flat and proceeded")
	}
	if flattenSent {
		t.Fatal("unknown must refuse BEFORE any flatten is submitted (do nothing destructive on an unreadable book)")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("the refusal must say unknown, got: %v", err)
	}
}

func TestReconcileBeforeOpenNTKnownFlatProceeds(t *testing.T) {
	at := &AutoTrader{
		id:       "recon-flat",
		exchange: "ninjatrader",
		trader:   &stubTrader{positions: []map[string]interface{}{}}, // known-flat
	}
	flattenSent, err := at.reconcileBeforeOpenNTReport("MNQ", "long")
	if err != nil {
		t.Fatalf("a POSITIVE empty book is flat and must proceed, got err=%v", err)
	}
	if flattenSent {
		t.Fatal("a flat book must not submit a flatten")
	}
}
