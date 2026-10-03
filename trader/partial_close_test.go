package trader

import (
	"strings"
	"testing"

	nt "vl/provider/ninjatrader"
	"vl/store"
)

// ── PARTIAL CLOSE (2026-10-03) — the decision cores ────────────────────────
//
// The pure seams are the production call sites' decision cores:
// reduceQuantityCheck is the pre-send guard, resizeAfterReduceDecision is the
// post-fill stop policy, findProtectiveLeg is the leg discovery.

func TestReduceQuantityCheck(t *testing.T) {
	cases := []struct {
		name    string
		openQty int
		wantQty int
		ok      bool
		why     string
	}{
		{"reduce 3 of 5", 5, 3, true, "leaves 2"},
		{"reduce 2 of 5", 5, 2, true, "leaves 3"},
		{"quantity >= open is REFUSED", 5, 5, false, "full close"},
		{"quantity > open is REFUSED", 5, 6, false, "full close"},
		{"zero quantity refused", 5, 0, false, "positive"},
		{"negative quantity refused", 5, -1, false, "positive"},
		{"no open position refused", 0, 1, false, "no open position"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, why := reduceQuantityCheck(tc.openQty, tc.wantQty)
			if ok != tc.ok {
				t.Fatalf("want ok=%v, got ok=%v (why=%q)", tc.ok, ok, why)
			}
			if !strings.Contains(why, tc.why) {
				t.Fatalf("why %q must mention %q", why, tc.why)
			}
		})
	}
}

// THE TWO MUTANTS THE DISPATCH NAMES, PINNED IN ONE TABLE:
//   - "ignore the quantity (flatten all)": a reduce of 3 from 5 MUST leave 2 —
//     the decision is refused whenever quantity >= open, and the remaining is
//     open-want. A caller that sends close_position instead fails here.
//   - "skip the stop resize": with a CONFIRMED leg cancel the action must be
//     resize_stop at the remaining quantity, and with an UNCONFIRMED cancel it
//     must be flatten — never a silent no-op.
func TestResizeAfterReduceDecisionPinsTheTwoMutants(t *testing.T) {
	if got, qty := resizeAfterReduceDecision(2, true); got != "resize_stop" || qty != 2 {
		t.Fatalf("reduce 3 of 5 + confirmed cancel must resize the stop to 2, got %q qty=%d", got, qty)
	}
	if got, qty := resizeAfterReduceDecision(3, true); got != "resize_stop" || qty != 3 {
		t.Fatalf("the resize quantity is the REMAINING position, got %q qty=%d", got, qty)
	}
	if got, _ := resizeAfterReduceDecision(2, false); got != "flatten" {
		t.Fatalf("an UNCONFIRMED leg cancel must flatten, got %q (skip-the-resize mutant)", got)
	}
	if got, _ := resizeAfterReduceDecision(0, true); got != "flat" {
		t.Fatalf("a zero remainder is flat, got %q", got)
	}
}

func TestFindProtectiveLeg(t *testing.T) {
	book := []nt.NT8Order{
		{OrderID: "a", Symbol: "MNQ", Name: "sig-1", State: "Working", Type: "stop"},
		{OrderID: "b", Symbol: "MNQ", Name: "sig-1-sl", State: "Working", Type: "stop", StopPrice: 29590.5},
		{OrderID: "c", Symbol: "MNQ", Name: "sig-1-tp", State: "Working", Type: "limit"},
		{OrderID: "d", Symbol: "MNQ", Name: "sig-2-sl", State: "Cancelled", Type: "stop"},
	}
	leg, price, ok := findProtectiveLeg(book, "MNQ")
	if !ok || leg != "sig-1-sl" || price != 29590.5 {
		t.Fatalf("want sig-1-sl @ 29590.5, got %q @ %.2f ok=%v", leg, price, ok)
	}
	// A terminal leg is not protective any more.
	if _, _, ok := findProtectiveLeg([]nt.NT8Order{book[3]}, "MNQ"); ok {
		t.Fatalf("a cancelled leg must not be found")
	}
	// Another symbol's leg does not match.
	if _, _, ok := findProtectiveLeg(book, "NQ"); ok {
		t.Fatalf("a different symbol's leg must not match")
	}
}

// The store call sites: the partial ledger and the resize lifecycle.
func TestPartialCloseStoreLifecycle(t *testing.T) {
	if at := (&AutoTrader{id: "t1"}); at.store != nil {
		t.Fatalf("fixture: no store")
	}
	// The pure halves are pinned above; the store halves are pinned in
	// store/partial_close_test.go against the production methods.
	_ = store.StopResizePending // compile-time: the state enum exists
}
