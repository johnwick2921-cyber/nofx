package trader

import (
	"fmt"
	"strings"
	"time"

	nt "nofx/provider/ninjatrader"
)

// ── CANCEL-CONFIRMATION (2026-09-06) ─────────────────────────────────────────
//
// THE DEFECT, IN ONE SENTENCE: a send is not a settlement.
//
// nt.CancelOrder is a one-line pass-through to SendCancelOrder
// (trader/ninjatrader/tcp_trader.go:558-562); its error is the result of PUTTING
// A FRAME ON A SOCKET. Five call sites in armed_executor.go treated `cerr == nil`
// as proof the order was gone and wrote the ledger `cancelled` on the strength of
// it (:342, :447, :490, :517, :837), and the sync helper wrote `cancelled` even
// on ACK TIMEOUT (:1933, "flatten proceeds"). So the ledger could say cancelled
// while the order still rested at the broker — which is exactly what
// nt8_order_snapshots id 1664 recorded: NINE working stop orders for ONE arm slot
// (2026-09-04:NY v3 S2 leg 0 SHORT), against nine ledger rows all reading
// 'cancelled'. They were harmless only because the order was malformed and inert.
//
// THE RULE (A20, class 6): a cancel is proven by the ORDER'S ABSENCE FROM A FRESH
// BROKER SNAPSHOT. Never by a return value, never by a log line, never by a
// ledger row. This file is the adjudication, kept PURE so a test can drive every
// branch with a real book; the clock is the caller's (A28).
//
// SIBLING TO class 79 (the reaper: silence is not death) and class 33 (the boot
// sweep). Their shared shape: an ABSENCE OF EVIDENCE is not evidence of absence.

// slotAction is what the adjudicator decided. It is deliberately not a bool:
// "allowed" and "refused because we cannot see" must not collapse together.
type slotAction string

const (
	// slotFree — a fresh book was read and it holds nothing for this slot.
	slotFree slotAction = "free"
	// slotLive — a fresh book still holds a non-terminal order for this slot.
	slotLive slotAction = "live"
	// slotUnverifiable — no book, or a book too old to believe. AN
	// UNVERIFIABLE SLOT IS NOT AN EMPTY SLOT (A24: UNKNOWN never takes the
	// destructive branch, and here PLACING is the destructive branch).
	slotUnverifiable slotAction = "unverifiable"
)

// slotVerdict is the whole adjudication as one value, so a caller cannot act on
// half of it and a test can assert all of it.
type slotVerdict struct {
	Action     slotAction
	Why        string
	SignalID   string // the live signal that blocked it, when Action == slotLive
	OrderID    string
	State      string // the broker's own word for that order
	SnapshotID int64  // which snapshot settled it; 0 when none did
	BookAge    time.Duration
}

// Allowed reports whether a placement may proceed. ONLY slotFree allows.
func (v slotVerdict) Allowed() bool { return v.Action == slotFree }

// Refusal renders the refusal exactly as D3 specifies, naming the signal, the
// order and the snapshot — the three ids a reader needs to check the claim.
func (v slotVerdict) Refusal() string {
	switch v.Action {
	case slotLive:
		return fmt.Sprintf("refused: slot already live at the broker (signal %s, order %s, state %s, snapshot %d, book age %s)",
			shortID(v.SignalID), shortID(v.OrderID), v.State, v.SnapshotID, v.BookAge.Round(time.Second))
	case slotUnverifiable:
		return fmt.Sprintf("refused: slot unverifiable — %s (an unverifiable slot is not an empty slot)", v.Why)
	}
	return ""
}

// orderBelongsToSlot matches a book order to one of a slot's signal ids. NT8
// names the entry after the signal and its bracket children "<signal>-sl" /
// "<signal>-tp", so the prefix is the join key — the same identity the exit-fill
// audit used to resolve 54 of 58 closes with zero ambiguity.
func orderBelongsToSlot(name string, slotSignalIDs []string) (string, bool) {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return "", false
	}
	for _, sid := range slotSignalIDs {
		s := strings.ToLower(strings.TrimSpace(sid))
		if s == "" {
			continue
		}
		if n == s || strings.HasPrefix(n, s+"-") {
			return sid, true
		}
	}
	return "", false
}

// adjudicateSlot decides whether an arm slot may be placed into.
//
// haveBook says a snapshot was actually READ (not that it was non-empty — an
// account with no orders legitimately emits an empty list, and that is a FREE
// slot, not an unverifiable one). maxAge is the resolved staleness bound.
//
// PURE: no clock, no store, no wire. The caller supplies the book, its age and
// the bound (A28).
func adjudicateSlot(
	book []nt.NT8Order, haveBook bool, bookAge, maxAge time.Duration,
	snapshotID int64, slotSignalIDs []string,
) slotVerdict {
	if !haveBook {
		return slotVerdict{Action: slotUnverifiable, Why: "no broker snapshot has been received", BookAge: bookAge}
	}
	if maxAge > 0 && bookAge > maxAge {
		return slotVerdict{
			Action: slotUnverifiable,
			Why: fmt.Sprintf("broker book is %s old, older than the %s bound",
				bookAge.Round(time.Second), maxAge.Round(time.Second)),
			BookAge: bookAge, SnapshotID: snapshotID,
		}
	}
	for i := range book {
		o := book[i]
		// IsWorking is the ONE definition of non-terminal (leg 4 and the override
		// guard read the same set) — reused, never re-implemented (A24).
		// CancelSubmitted and CancelPending are NOT terminal, which is why a
		// cancel still in flight correctly keeps the slot locked: measured 130
		// and 22 occurrences respectively across 360 live frames.
		if !o.IsWorking() {
			continue
		}
		if sid, ok := orderBelongsToSlot(o.Name, slotSignalIDs); ok {
			return slotVerdict{
				Action: slotLive, SignalID: sid, OrderID: o.OrderID, State: o.State,
				SnapshotID: snapshotID, BookAge: bookAge,
				Why: "a non-terminal order for this slot is still at the broker",
			}
		}
	}
	return slotVerdict{Action: slotFree, SnapshotID: snapshotID, BookAge: bookAge,
		Why: "no non-terminal order for this slot in a fresh book"}
}

// cancelSettled decides whether a requested cancel is CONFIRMED GONE.
//
// The same rule from the other side: the order is gone when a FRESH book no
// longer lists it as non-terminal. A stale or absent book settles NOTHING — it
// must never promote a row to cancelled, because 'cancelled' is what unlocks a
// replacement (D2).
func cancelSettled(
	book []nt.NT8Order, haveBook bool, bookAge, maxAge time.Duration, signalID string,
) (settled bool, why string) {
	if strings.TrimSpace(signalID) == "" {
		return false, "no signal id — nothing to look for"
	}
	if !haveBook {
		return false, "no broker snapshot has been received"
	}
	if maxAge > 0 && bookAge > maxAge {
		return false, fmt.Sprintf("book is %s old, older than the %s bound", bookAge.Round(time.Second), maxAge.Round(time.Second))
	}
	for i := range book {
		o := book[i]
		if !o.IsWorking() {
			continue
		}
		if _, ok := orderBelongsToSlot(o.Name, []string{signalID}); ok {
			return false, "still at the broker as " + o.State
		}
	}
	return true, "absent from a fresh book"
}

// shortID keeps a log line readable without inventing a value.
func shortID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "n/a"
	}
	if len(s) > 8 {
		return s[:8]
	}
	return s
}
