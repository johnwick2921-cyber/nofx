// PLACEMENT CONFIRMATION — class 81, one line lower (owner ruling 2026-09-07).
//
// THE INCIDENT. At 22:47:24 arm 117 was sent to NT8. PlaceLimitEntry returned
// nil, so the ledger wrote `working` — a word that means RESTING AT THE BROKER.
// In the same second NT8 answered:
//
//	VLTraderTCPClient: stale signal 9ba63cb5-… (age 1824.5s) — rejecting
//
// Nothing carried that refusal back to the ledger. For the next 33 minutes four
// surfaces disagreed: the ledger said `working`, the broker's fresh book was
// empty, the in-memory pending map had correctly dropped it, and the chart drew
// a line at the stop price labelled "Limit". The loudest of the four — a red
// [ERRO] — reached no surface the owner reads.
//
// THE RULE, the same one the cancel path already obeys: a distributed claim is
// settled by a RECEIVED far-side frame. A send is an intention. A placement
// lands in place_pending and only a frame that NAMES the signal promotes it,
// recording which frame did so; a refusal moves it terminal in the broker's own
// words; and a placement no frame ever settles retires as `unconfirmed`, never
// as the flattering guess in either direction.
package trader

import (
	"strings"
	"time"

	nt "nofx/provider/ninjatrader"
	"nofx/store"
	"nofx/telemetry"
)

// placeConfirmMaxWait is how long a sent placement may go unconfirmed before it
// is retired. RESOLVED from the snapshot cadence rather than picked: two
// snapshot intervals is the same bound the slot guard calls a stale book, so a
// placement cannot outlive the evidence that would have settled it.
func placeConfirmMaxWait() time.Duration { return snapshotMaxAge() }

// confirmPendingPlacements is the per-cycle settlement pass for placements. It
// mirrors confirmPendingCancels exactly, and like it, it is telemetry-shaped: a
// failed read WARNs and returns, never stops the loop, and never promotes a row
// on ignorance (A10 / class 23).
//
// Promotion evidence, in order of strength:
//  1. the broker's periodic BOOK naming the signal — the same frame leg 4 trusts
//  2. an order_update the wire already recorded against the signal
//
// Nothing else promotes. In particular the absence of a rejection is not
// evidence of acceptance.
func (at *AutoTrader) confirmPendingPlacements(ledger *store.ArmedOrderStore, now time.Time) (confirmed, stillPending, expired int) {
	if at == nil || ledger == nil {
		return 0, 0, 0
	}
	rows, err := ledger.ListPlacePending(at.id)
	if err != nil {
		at.logWarnf("📤 place-confirm: ledger read failed (%v) — nothing promoted this cycle", err)
		return 0, 0, 0
	}
	if len(rows) == 0 {
		return 0, 0, 0
	}

	book, haveBook, bookAge := at.liveBook(now)
	_, _, _, snapID := at.persistedBook(now)

	for _, r := range rows {
		sig := strings.TrimSpace(r.SignalID)
		if sig == "" {
			// Sent with no signal id is not representable — but if it ever
			// happens, it can never be confirmed, so retire it rather than let
			// it hold a slot forever.
			_ = ledger.ExpirePlacement(r.ID, now.Sub(r.CreatedAt))
			at.logWarnf("📤 place-confirm: %s leg %d had NO signal id — retired unconfirmed (nothing could ever name it)", r.Scenario, r.LegIndex+1)
			expired++
			continue
		}

		// 1. THE BOOK. A fresh book that names the signal is the strongest
		//    evidence there is, and it is the same evidence the flat gate's leg
		//    4 answers from.
		if haveBook && bookAge <= placeConfirmMaxWait() {
			if o, ok := bookOrderForSignal(book, sig); ok {
				frame := "order_snapshot " + snapshotLabel(snapID) + " (" + o.State + ", " + o.Type + ")"
				if err := ledger.ConfirmPlacement(r.ID, frame); err != nil {
					at.logWarnf("📤 place-confirm: promote failed for %s: %v", shortID(sig), err)
					continue
				}
				at.logInfof("📥 armed %s leg %d → WORKING — confirmed by %s (signal %s)", r.Scenario, r.LegIndex+1, frame, shortID(sig))
				telemetry.IncGateBlock(at.id, "place_confirmed")
				confirmed++
				continue
			}
		}

		// 2. NOT YET. A placement is only retired once the evidence that would
		//    have settled it has had its full bound to arrive.
		waited := now.Sub(r.CreatedAt)
		if waited <= placeConfirmMaxWait() {
			stillPending++
			continue
		}

		// 3. THE BOUND IS SPENT. Terminal, and NAMED — never silently working
		//    (which claims a broker state) and never silently gone (which hides
		//    that we sent something).
		if err := ledger.ExpirePlacement(r.ID, waited); err != nil {
			at.logWarnf("📤 place-confirm: expire failed for %s: %v", shortID(sig), err)
			continue
		}
		bookNote := "no book to check"
		if haveBook {
			bookNote = "book age " + bookAge.Round(time.Second).String() + ", signal absent from it"
		}
		at.logWarnf("📤 armed %s leg %d → UNCONFIRMED after %s — no order_update and no book frame ever named signal %s (%s). SENT, never confirmed, never refused.",
			r.Scenario, r.LegIndex+1, waited.Round(time.Second), shortID(sig), bookNote)
		telemetry.IncGateBlock(at.id, "place_unconfirmed_no_frame")
		expired++
	}
	return confirmed, stillPending, expired
}

// bookOrderForSignal finds a WORKING order in the broker's book carrying this
// signal's name. It reuses orderBelongsToSlot's join so the confirmation and the
// slot guard can never disagree about what "this signal's order" means.
func bookOrderForSignal(book []nt.NT8Order, signalID string) (nt.NT8Order, bool) {
	for _, o := range book {
		if !o.IsWorking() {
			continue
		}
		if _, ok := orderBelongsToSlot(o.Name, []string{signalID}); ok {
			return o, true
		}
	}
	return nt.NT8Order{}, false
}

// snapshotLabel renders a snapshot id, or says it has none rather than printing
// a zero that reads like snapshot 0 (A24).
func snapshotLabel(id int64) string {
	if id <= 0 {
		return "(live cache, unpersisted)"
	}
	return "#" + itoa64(id)
}

func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

// PlaceConfirmBootLine states the placement-confirmation contract once at boot,
// every field READ from the code that enforces it (A11).
//
// It carries one thing that is NOT yet true: the AddOn half of the broker-reason
// field. FillPayload.Reason and the C# SendFillFrame(reason:) shipped together,
// but NinjaScript only takes effect after the copy → F5 → full NT8 restart, so
// until that happens a rejection records the honest fallback rather than NT8's
// sentence. A boot line that implied otherwise would be the exact defect this
// wave exists to fix, one level up.
func PlaceConfirmBootLine(addonReasonLive bool) string {
	reason := "NOT LIVE until the next AddOn copy/F5/full NT8 restart — rejections record \"no reason text on the fill frame\" until then"
	if addonReasonLive {
		reason = "live — rejections carry NT8's own sentence"
	}
	return "place-confirm: a send writes place_pending, never working · promoted ONLY by a received frame naming the signal (recorded) · " +
		"reject → terminal in the broker's words · unconfirmed after " + placeConfirmMaxWait().String() + " → unconfirmed:no_frame · " +
		"broker-reason wire field: " + reason
}
