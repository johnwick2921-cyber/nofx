package trader

import (
	"fmt"
	"strings"
	"sync"

	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
	ntTrader "nofx/trader/ninjatrader"
)

// PICTURE-HTF BROKER STATE (2026-09-20) — the missing halves the CTO named:
// the live consumer of RECEIVED broker events and the restart-safe
// reconciliation sweep. Both write only through PictureHtfMarkBrokerState —
// stages move FORWARD, fills are never downgraded, and a missing broker
// record leaves the row place_pending (blocked re-entry, never a blind
// resend).

// pictureEntryStage maps an order_update wire state onto the ledger stage
// vocabulary. accepted/partfilled carry the row forward like working;
// cancelled is a terminal refusal.
func pictureEntryStage(u string) string {
	switch u {
	case "accepted", "partfilled":
		return store.StateWorking
	case "cancelled":
		return store.StateRejected
	default:
		return strings.ToLower(u)
	}
}

// pictureStagePriority orders ledger stages so a late/duplicate event can
// never downgrade a row: terminal = 2, working = 1, everything else 0. The
// canonical predicate store.IsTerminalArmState is the ONLY terminal source
// (no re-typed state lists — arm_state_source_guard).
func pictureStagePriority(stage string) int {
	if store.IsTerminalArmState(stage) {
		return 2
	}
	if stage == store.StateWorking {
		return 1
	}
	return 0
}

// pictureLeg returns ("sl"|"tp", true) for a protective-order frame; the C#
// strips the suffix from signal_id but keeps it in order_name.
func pictureLeg(orderName, signalID string) (string, bool) {
	if strings.HasSuffix(orderName, "-sl") && strings.TrimSuffix(orderName, "-sl") == signalID {
		return "sl", true
	}
	if strings.HasSuffix(orderName, "-tp") && strings.TrimSuffix(orderName, "-tp") == signalID {
		return "tp", true
	}
	return "", false
}

// pictureHtfFillRR is the actual-fill R:R from the row's own geometry.
func pictureHtfFillRR(row store.PictureHtfOpportunityDB, fill float64) float64 {
	if fill <= 0 || row.StopPx <= 0 || row.TargetPx <= 0 {
		return 0
	}
	risk := fill - row.StopPx
	reward := row.TargetPx - fill
	if row.Direction == "short" {
		risk = row.StopPx - fill
		reward = fill - row.TargetPx
	}
	if risk <= 0 {
		return 0
	}
	return reward / risk
}

// pictureHtfConsumeOrderUpdate routes ONE received order_update frame into the
// opportunity it names. Entry-leg events move the row forward (working →
// partfilled → filled; rejected is terminal) and stamp fill price/qty and the
// actual-fill R:R. Protective-leg events (-sl/-tp) update the protection
// status WITHOUT touching the entry's stage or fill — a filled SL leg is the
// proof the protective order worked, recorded as protection_sl_filled.
func pictureHtfConsumeOrderUpdate(at *AutoTrader, u ntwire.OrderUpdatePayload) {
	if at == nil || at.store == nil || u.SignalID == "" {
		return
	}
	rows, err := at.store.PictureHtfBySignal(u.SignalID)
	if err != nil || len(rows) == 0 {
		return
	}
	for i := range rows {
		row := rows[i]
		if leg, isLeg := pictureLeg(u.OrderName, u.SignalID); isLeg {
			note := fmt.Sprintf("protection_%s_%s", leg, strings.ToLower(u.State))
			if u.State == "rejected" && u.Reason != "" {
				note += ": " + u.Reason
			}
			_ = at.store.PictureHtfMarkBrokerState(row.OppKey, row.Stage, "", note, u.Reason, 0, 0, 0)
			continue
		}
		// Entry leg: only move FORWARD, never downgrade.
		stage := pictureEntryStage(u.State)
		pri := pictureStagePriority(stage)
		if pri <= pictureStagePriority(row.Stage) && row.Stage != store.StatePlacePending {
			continue
		}
		_ = at.store.PictureHtfMarkBrokerState(row.OppKey, stage, "", u.State, u.Reason, u.FillPrice, float64(u.Quantity), pictureHtfFillRR(row, u.FillPrice))
	}
}

// pictureHtfBrokerLookup is the reconciliation's broker view: the latest
// RECEIVED order snapshot for the trader's bound account. Tests override it.
var pictureHtfBrokerLookup = func(at *AutoTrader, signalID string) (ntwire.NT8Order, bool) {
	tcp, ok := at.trader.(*ntTrader.TCPTrader)
	if !ok {
		return ntwire.NT8Order{}, false
	}
	return tcp.OrderSnapshotLookup(signalID)
}

// pictureHtfReconcilePending recovers place_pending rows across disconnects
// and restarts WITHOUT another entry: for each pending row the broker book is
// consulted and the row moves forward to what the BROKER says (filled →
// filled + fill fields; rejected/cancelled → rejected; working → working).
// A pending row whose order is ABSENT from the book stays place_pending —
// the ambiguous case is never blindly resent (the atomic claim already blocks
// re-entry; this sweep only ever observes).
func pictureHtfReconcilePending(at *AutoTrader) {
	if at == nil || at.store == nil {
		return
	}
	rows, err := at.store.PictureHtfPendingByTrader(at.id)
	if err != nil || len(rows) == 0 {
		return
	}
	for _, row := range rows {
		if row.SignalID == "" {
			continue
		}
		order, found := pictureHtfBrokerLookup(at, row.SignalID)
		if !found {
			continue // absent from the broker book — stay place_pending, never resend
		}
		state := strings.ToLower(order.State)
		if state == store.StateFilled {
			fill := float64(order.Filled)
			px := order.LimitPrice
			_ = at.store.PictureHtfMarkBrokerState(row.OppKey, store.StateFilled, order.OrderID, order.State, "", px, fill, pictureHtfFillRR(row, px))
		} else if store.IsTerminalArmState(state) {
			_ = at.store.PictureHtfMarkBrokerState(row.OppKey, store.StateRejected, order.OrderID, order.State, "reconciled from broker book", 0, 0, 0)
		} else {
			// Still working on the broker's book: record it, stay pending.
			_ = at.store.PictureHtfMarkBrokerState(row.OppKey, store.StateWorking, order.OrderID, order.State, "", 0, 0, 0)
		}
	}
}

// pictureHtfBrokerConsumers guards the per-trader consumer goroutine
// (keyed by trader id — a restarted trader replaces its entry).
var pictureHtfBrokerConsumers sync.Map

// ensurePictureHtfBrokerConsumer starts the live order_update consumer for the
// concrete NT8 trader (once per trader). Frames with no matching row are
// ignored; the consumer is the only writer of broker state besides the
// reconciliation sweep.
func (at *AutoTrader) ensurePictureHtfBrokerConsumer() {
	if at == nil || at.id == "" {
		return
	}
	if _, loaded := pictureHtfBrokerConsumers.LoadOrStore(at.id, struct{}{}); loaded {
		return
	}
	tcp, ok := at.trader.(*ntTrader.TCPTrader)
	if !ok || tcp == nil {
		return
	}
	go func() {
		for u := range tcp.OrderUpdates() {
			pictureHtfConsumeOrderUpdate(at, u)
		}
	}()
}
