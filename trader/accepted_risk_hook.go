package trader

import (
	"strconv"
	"strings"
	"time"

	nt "nofx/provider/ninjatrader"
	"nofx/store"
)

// ── WAVE A / D4 — RECORDING WHAT THE BROKER ACCEPTED ─────────────────────────
//
// NT8 sends an order_update the moment it accepts an order, and the AddOn fires
// an order_snapshot on the SAME state change (VLTraderTCPClient.cs:1600), so the
// broker's own prices are in hand at exactly the instant we need them. Until
// now the acceptance event was received and dropped: onArmedOrderUpdate handled
// filled / partfilled / rejected / cancelled and had NO case for "accepted".
//
// CLASS 23 / A10 GOVERNS THIS FILE. It is telemetry. It writes one row and
// decides nothing — a panic, a stale book or a locked database WARNs and the
// trading loop continues.

// recordAcceptedRisk appends the immutable record of an accepted order. It is
// called once per acceptance; a re-authorization appends ANOTHER row rather
// than editing this one, so ledger drift becomes a query.
func (at *AutoTrader) recordAcceptedRisk(r store.ArmedOrderDB, u nt.OrderUpdatePayload) {
	if at == nil || at.store == nil {
		return
	}
	defer func() {
		if rec := recover(); rec != nil {
			at.logWarnf("🧾 accepted-risk recording panicked and was contained: %v", rec)
		}
	}()

	now := time.Now().UTC()
	row := &store.AcceptedRisk{
		TraderID: at.id, SignalID: u.SignalID, OrderName: u.OrderName,
		Symbol: u.Symbol, Account: u.Account, Side: r.Side,
		Quantity: u.Quantity,
		// The ledger's terms AT THIS INSTANT — recorded so the divergence from
		// the broker is measurable instead of reconstructed from a log file.
		LedgerEntryPx: r.EntryPx, LedgerStopPx: r.StopPx, LedgerTargetPx: r.TargetPx,
		AcceptedAtMs: now.UnixMilli(),
		BookSource:   "none",
	}

	// THE BROKER'S OWN PRICES, from the same F12 book cutover leg 4 answers
	// from. Absent or stale leaves them NULL: an unknown accepted price is
	// never 0, and never the ledger's number wearing the broker's name.
	if cache, account, _ := at.brokerBook(); cache != nil {
		if age, ok := cache.AgeAt(account, now); ok {
			row.BookAgeMs = age.Milliseconds()
		}
		if snap, ok := cache.Latest(account); ok {
			row.BookSource = "f12"
			applyBrokerTerms(row, snap.WorkingOrders(), u.SignalID)
		}
	}

	if err := at.store.AcceptedRisk().Append(row); err != nil {
		at.logWarnf("🧾 accepted-risk write failed (signal %s): %v", u.SignalID, err)
		return
	}
	at.logInfof("🧾 accepted risk recorded: signal=%s order=%s entry=%s stop=%s target=%s (ledger stop %.4f) · book=%s age=%dms",
		u.SignalID, u.OrderName, fmtPx(row.AcceptedEntryPx), fmtPx(row.AcceptedStopPx),
		fmtPx(row.AcceptedTargetPx), row.LedgerStopPx, row.BookSource, row.BookAgeMs)
}

// applyBrokerTerms fills the accepted prices from the broker's book. NT8 names
// the bracket children "<signal>-sl" and "<signal>-tp", so the leg each order
// IS comes from its own name rather than from a guess about its price.
func applyBrokerTerms(row *store.AcceptedRisk, orders []nt.NT8Order, signalID string) {
	if signalID == "" {
		return
	}
	for i := range orders {
		o := orders[i]
		if !strings.HasPrefix(o.Name, signalID) {
			continue
		}
		switch {
		case strings.HasSuffix(strings.ToLower(o.Name), "-sl"):
			if o.StopPrice > 0 {
				v := o.StopPrice
				row.AcceptedStopPx = &v
			}
		case strings.HasSuffix(strings.ToLower(o.Name), "-tp"):
			if o.LimitPrice > 0 {
				v := o.LimitPrice
				row.AcceptedTargetPx = &v
			}
		default:
			// the entry itself
			if row.OrderType == "" {
				row.OrderType = o.Type
			}
			if p := entryPriceOf(o); p > 0 {
				v := p
				row.AcceptedEntryPx = &v
			}
		}
	}
}

// entryPriceOf is the price the entry order rests at: a stop's trigger,
// otherwise its limit. Same rule orderPrice uses for leg 4 — one definition.
func entryPriceOf(o nt.NT8Order) float64 {
	if o.StopPrice > 0 {
		return o.StopPrice
	}
	return o.LimitPrice
}

// fmtPx renders a nullable price so a log line cannot print 0.00 for "unknown".
func fmtPx(p *float64) string {
	if p == nil {
		return "n/a"
	}
	return strconv.FormatFloat(*p, 'f', -1, 64)
}
