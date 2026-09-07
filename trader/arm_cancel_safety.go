package trader

import (
	"strings"
	"time"

	nt "nofx/provider/ninjatrader"
	"nofx/store"
)

// ── NEVER CANCEL A FILLED ARM (2026-09-07) ───────────────────────────────────
//
// WHAT HAPPENED. 2026-09-06 23:35:03 an S1 long limit was placed at 29576 and
// filled at ~23:35:22. The reconciler saw the untracked position and
// materialized it at 23:36:42. One cycle later, at 23:37:02, the
// one-open-position guard observed "a position is open" and cancelled the arm
// that had CREATED it — because the LEDGER still read `working`: the fill was
// drained at the END of the cycle, after every guard had already decided.
//
// NT8 cancels by signal id, so that cancel took the bracket with it. Snapshot
// 8208, seven seconds before the guard fired, is the whole story:
//
//     aa07e583-…-sl   Accepted   stop 29554
//     aa07e583-…-tp   Working    limit 29623
//
// TWO CHILDREN AND NO ENTRY. The entry order was already gone — filled — and
// the only things left to cancel were the protections. Position 592 then ran
// naked for 8h18m, through the Monday open.
//
// THE RULE. A cancel targets the ENTRY. If the entry is no longer resting at
// the broker, there is nothing to cancel and the children are not ours to
// touch. The ledger's word is not evidence — it is a memory, and at 23:37:02 it
// was a minute out of date.

// armCancelVerdict is the whole adjudication as one value, so a caller cannot
// act on half of it.
type armCancelVerdict struct {
	Allow bool
	Why   string
}

// entryIsResting reports whether the book holds a WORKING order that IS this
// signal's entry — a bare name, not a bracket child.
//
// NT8 names the entry after the signal and its children "<signal>-sl" /
// "<signal>-tp", so the suffix is what separates "the order we placed" from
// "the protections it grew".
func entryIsResting(book []nt.NT8Order, signalID string) (found bool, childrenSeen bool) {
	sig := strings.ToLower(strings.TrimSpace(signalID))
	if sig == "" {
		return false, false
	}
	for i := range book {
		o := book[i]
		if !o.IsWorking() {
			continue
		}
		n := strings.ToLower(strings.TrimSpace(o.Name))
		switch {
		case n == sig:
			found = true
		case strings.HasPrefix(n, sig+"-"):
			childrenSeen = true
		}
	}
	return found, childrenSeen
}

// adjudicateArmCancel decides whether this arm's signal may be cancelled AT THE
// BROKER. PURE: the caller supplies the ledger state and the book.
//
// It refuses on ignorance. Cancelling is the destructive branch here — it can
// remove a live protective stop — so an unreadable book is a refusal, not a
// permission (A24: UNKNOWN never takes the destructive branch).
func adjudicateArmCancel(ledgerState, signalID string, book []nt.NT8Order, haveBook bool) armCancelVerdict {
	if strings.TrimSpace(signalID) == "" {
		return armCancelVerdict{false, "the arm has no signal id — nothing was ever placed under it"}
	}
	// THE LEDGER'S OWN WORD, first and cheapest. A filled arm is never cancelled.
	switch strings.ToLower(strings.TrimSpace(ledgerState)) {
	case store.StateFilled:
		return armCancelVerdict{false, "the arm is FILLED — its entry is a position, and the only orders left under this signal are its protections"}
	case store.StateArmed, store.StateWorking:
		// keep going: the ledger thinks it is live, and the book decides.
	default:
		return armCancelVerdict{false, "the arm is " + ledgerState + " — not a live order"}
	}
	// THE BOOK DECIDES. The ledger is a memory; at 23:37:02 it was a minute out
	// of date and that minute cost a stop.
	if !haveBook {
		return armCancelVerdict{false, "no broker book — the entry cannot be shown to be resting, and cancelling blind can take a live protective stop with it"}
	}
	resting, children := entryIsResting(book, signalID)
	if resting {
		return armCancelVerdict{true, "the entry is still resting at the broker"}
	}
	if children {
		return armCancelVerdict{false, "the entry has FILLED — only its OCO children remain, and a cancel would take the protection with it (2026-09-06 23:37:02)"}
	}
	// NOTHING AT ALL under this signal. This is NOT the dangerous case and must
	// not be refused: there is no protection to lose, and refusing would strand
	// the ledger row `working` forever with no way to retire it. The guard
	// exists to protect a FILLED arm's children — not to block every cancel
	// whose target has already gone. (Caught by
	// TestShadowedRestingOrderCancelledAtBoot, which retires exactly such a row.)
	return armCancelVerdict{true, "nothing under this signal is working at the broker — the cancel is a harmless no-op and lets the ledger row retire"}
}

// cancelSignalIfSafe is the SIGNAL-level gate, for the paths that hold only a
// signal id (the stale reaper, the re-request pass). The ledger state is
// unknown there, so the BOOK alone decides — and the same rule holds: if the
// entry is not resting, the only things under this signal are protections.
//
// Returns true when the cancel was actually sent.
func (at *AutoTrader) cancelSignalIfSafe(send func(string) error, signalID, who string, now time.Time) bool {
	book, have, _ := at.liveBook(now)
	v := adjudicateArmCancel(store.StateWorking, signalID, book, have)
	if !v.Allow {
		at.logWarnf("🛟 cancel REFUSED (%s) signal=%s — %s", who, shortID(signalID), v.Why)
		return false
	}
	if err := send(signalID); err != nil {
		at.logWarnf("✕ cancel SEND failed (%s) signal=%s: %v", who, shortID(signalID), err)
	}
	return true
}

// cancelSafetyFor is the production entry point: it reads THIS trader's live
// book and adjudicates one row. now is the caller's clock (A28).
func (at *AutoTrader) cancelSafetyFor(r store.ArmedOrderDB, now time.Time) armCancelVerdict {
	book, have, _ := at.liveBook(now)
	return adjudicateArmCancel(r.State, r.SignalID, book, have)
}
