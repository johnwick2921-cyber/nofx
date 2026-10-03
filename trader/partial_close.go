package trader

import (
	"fmt"
	"strings"
	"time"

	nt "vl/provider/ninjatrader"
	"vl/store"
	ntTrader "vl/trader/ninjatrader"
)

// ── PARTIAL CLOSE (2026-10-03, mentor mode) ────────────────────────────────
//
// The owner's mentor sizing scales OUT HALF at 1:1, then break-even, then
// trail (owner order 10-02 22:2x CT). Today that is impossible on the wire:
// close_position flattens the whole position. This file is the EXECUTION half:
// the exact-quantity market exit (reduce_position), the per-partial ledger,
// and the protective-stop resize that follows a CONFIRMED reduce fill through
// the #309 report regime. EVERYTHING sits behind CANCEL_CONFIRM_REQUIRE_REPORT
// (default OFF) — nothing changes for the AI mode, which keeps its
// 1-contract rule.

// partialCloseEnabled: the partial-close feature rides the SAME knob as the
// report regime (dispatch: "everything stays behind the #309 knob").
func partialCloseEnabled() bool { return cancelConfirmRequireReport() }

// reduceQuantityCheck is the pre-send guard, PURE. A reduce is refused when it
// would close the whole position — a full close still goes through
// close_position — or is not a positive partial.
func reduceQuantityCheck(openQty, want int) (ok bool, reason string) {
	switch {
	case want <= 0:
		return false, "reduce quantity must be positive"
	case openQty <= 0:
		return false, "no open position to reduce"
	case want >= openQty:
		return false, fmt.Sprintf("reduce quantity %d >= open %d — a full close goes through close_position", want, openQty)
	}
	return true, fmt.Sprintf("reduce %d of %d leaves %d", want, openQty, openQty-want)
}

// resizeAfterReduceDecision is the post-fill decision, PURE. The protective
// stop is resized ONLY when the leg cancel is CONFIRMED through the report
// regime; a resize that cannot be confirmed FAILS CLOSED and the remainder is
// flattened — never a naked remainder, never a blind re-place (dispatch item
// 2).
func resizeAfterReduceDecision(remaining int, stopResizeConfirmed bool) (action string, qty int) {
	if !stopResizeConfirmed {
		return "flatten", 0
	}
	if remaining <= 0 {
		return "flat", 0
	}
	return "resize_stop", remaining
}

// findProtectiveLeg is PURE: the book entry that is the protective stop for
// the symbol — the order named "<signal>-sl". Returns the leg's full order
// name and its stop price (captured BEFORE the cancel, so the replacement can
// be placed at the same price once the report regime confirms the cancel).
func findProtectiveLeg(book []nt.NT8Order, symbol string) (legName string, stopPrice float64, ok bool) {
	for i := range book {
		o := book[i]
		if !o.IsWorking() {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(o.Symbol), strings.TrimSpace(symbol)) {
			continue
		}
		if strings.HasSuffix(strings.ToLower(strings.TrimSpace(o.Name)), "-sl") {
			return o.Name, o.StopPrice, true
		}
	}
	return "", 0, false
}

// ReducePosition runs the guarded partial close: knob, the quantity guard,
// the typed send (which itself refuses on an AddOn that never advertised
// reduce_position), and the request-time ledger row.
func (at *AutoTrader) ReducePosition(side string, qty, openQty int, who string) error {
	if !partialCloseEnabled() {
		return fmt.Errorf("partial close is disabled (CANCEL_CONFIRM_REQUIRE_REPORT off)")
	}
	nt := at.armedTrader()
	if nt == nil {
		return fmt.Errorf("partial close: no bound NT8 trader")
	}
	if ok, why := reduceQuantityCheck(openQty, qty); !ok {
		return fmt.Errorf("partial close REFUSED: %s", why)
	}
	side = strings.ToLower(strings.TrimSpace(side))
	if side != "long" && side != "short" {
		return fmt.Errorf("partial close: side must be long or short, got %q", side)
	}
	clientID := fmt.Sprintf("rx-%d-%s", time.Now().UnixMilli(), strings.ToLower(who))
	if err := nt.ReducePosition(side, qty, clientID); err != nil {
		return fmt.Errorf("partial close send refused: %w", err)
	}
	if at.store != nil {
		if err := at.store.PartialClose().RecordReduce(&store.PositionReduction{
			TraderID: at.id, Symbol: at.futuresSymbol(), Side: side,
			ClientID: clientID, Quantity: qty, Remaining: -1, Who: who,
		}); err != nil {
			at.logWarnf("🧩 partial close: ledger write failed for %s: %v", clientID, err)
		}
	}
	at.logInfof("🧩 partial close REQUESTED %s %s qty=%d of %d (client=%s)", at.futuresSymbol(), side, qty, openQty, clientID)
	return nil
}

// consumeReduceFills drains the trader's reduce_fill stream: every fill is
// applied to the ledger (latest-wins per client_id), and a fill with a
// positive remaining quantity opens the protective-stop resize through the
// #309 confirm path.
func (at *AutoTrader) consumeReduceFills(nt *ntTrader.TCPTrader) {
	ch := nt.ReduceFills()
	if ch == nil {
		return
	}
	for {
		select {
		case f, open := <-ch:
			if !open {
				return
			}
			at.onReduceFill(nt, f)
		default:
			return
		}
	}
}

func (at *AutoTrader) onReduceFill(nt *ntTrader.TCPTrader, f nt.ReduceFillPayload) {
	if at.store != nil {
		if err := at.store.PartialClose().ApplyReduceFill(f.ClientID, f.FillPrice, f.Remaining); err != nil {
			at.logWarnf("🧩 partial close: fill ledger write failed for %s: %v", f.ClientID, err)
		}
	}
	at.logInfof("🧩 partial close FILLED %s %s client=%s qty=%d @ %.2f remaining=%d", f.Symbol, f.Side, f.ClientID, f.Quantity, f.FillPrice, f.Remaining)
	if f.Remaining <= 0 {
		return // the remainder is flat: nothing to resize
	}
	at.requestStopResize(nt, f.Remaining, f.Side)
}

// requestStopResize cancels the protective leg for the reduced position and
// opens a pending stop_resizes row. The replacement stop is placed ONLY after
// confirmStopResizes sees the AddOn's terminal report for the leg's order id —
// and a resize that cannot be confirmed fails closed (the remainder is
// flattened), never a blind re-place.
func (at *AutoTrader) requestStopResize(nt *ntTrader.TCPTrader, remaining int, side string) {
	book, have, age := at.liveBook(time.Now())
	if !have || age > snapshotMaxAge() {
		at.logWarnf("🧩 stop resize: cannot read the broker book (age=%s) — failing closed", age.Round(time.Second))
		at.flattenRemainderAfterResizeFailure("the broker book is unreadable")
		return
	}
	legName, stopPrice, ok := findProtectiveLeg(book, at.futuresSymbol())
	if !ok {
		// No protective leg in the book: the position is already naked. The
		// only honest action is the flatten — a stop we cannot see cannot be
		// resized.
		at.logWarnf("🧩 stop resize: no protective leg found for %s — failing closed", at.futuresSymbol())
		at.flattenRemainderAfterResizeFailure("no protective leg in the broker book")
		return
	}
	entrySignal := strings.TrimSuffix(legName, "-sl")
	now := time.Now()
	if at.store != nil {
		if err := at.store.PartialClose().RecordStopResizeRequest(&store.StopResize{
			TraderID: at.id, Symbol: at.futuresSymbol(), LegSignalID: legName,
			Side: side, Quantity: remaining, StopPrice: stopPrice,
			RequestMs: now.UnixMilli(),
		}); err != nil {
			at.logWarnf("🧩 stop resize: request ledger write failed: %v", err)
		}
	}
	if err := nt.CancelBracketLeg(entrySignal, "sl"); err != nil {
		at.logWarnf("🧩 stop resize: leg cancel SEND failed (%v) — the resize stays pending on the report regime", err)
	}
	at.logInfof("🧩 stop resize REQUESTED leg=%s -> qty=%d (cancel sent; the replacement waits for the AddOn's terminal report)", legName, remaining)
}

// confirmStopResizes is the per-cycle settlement pass for pending resize rows
// — the #309 report regime applied to the leg: a row promotes ONLY on the
// AddOn's positive terminal report for the leg's order id; past the timeout it
// is re-requested up to the cap; at the cap it FAILS CLOSED and the remainder
// is flattened. It never places a stop beside an unconfirmed live one.
func (at *AutoTrader) confirmStopResizes(nt *ntTrader.TCPTrader, now time.Time) {
	if !partialCloseEnabled() || at.store == nil || nt == nil {
		return
	}
	rows, err := at.store.PartialClose().ListStopResizePending(at.id)
	if err != nil || len(rows) == 0 {
		return
	}
	timeout := cancelConfirmTimeout()
	cap := cancelReRequestMax()
	for i := range rows {
		r := rows[i]
		ok, why := cancelReportQualifies(store.ArmedOrderDB{
			CancelReportMs: r.ReportMs, CancelReportState: r.ReportState,
			CancelRequestedAtMs: r.RequestMs,
		})
		if ok {
			confirmed, err := at.store.PartialClose().ConfirmStopResizeByReport(r.ID)
			if err != nil {
				at.logWarnf("🧩 stop resize confirm failed for leg=%s: %v", r.LegSignalID, err)
				continue
			}
			if err := nt.PlaceProtectiveStop(at.futuresSymbol(), confirmed.Side, confirmed.Quantity, confirmed.StopPrice,
				strings.TrimSuffix(confirmed.LegSignalID, "-sl"), "resize after reduce"); err != nil {
				at.logWarnf("🧩 stop resize REPLACE failed for leg=%s (%v) — failing closed", r.LegSignalID, err)
				at.flattenRemainderAfterResizeFailure("replacement stop placement failed")
				_ = at.store.PartialClose().MarkStopResizeFailed(r.ID)
				continue
			}
			_ = at.store.PartialClose().MarkStopResizeDone(r.ID)
			at.logInfof("🧩 stop resize CONFIRMED+PLACED leg=%s -> qty=%d — %s", r.LegSignalID, confirmed.Quantity, why)
			continue
		}
		reqAge := time.Duration(0)
		if r.RequestMs > 0 {
			reqAge = time.Duration(now.UnixMilli()-r.RequestMs) * time.Millisecond
		}
		if reqAge < timeout {
			continue
		}
		if r.Attempts >= cap {
			at.logWarnf("🧩 stop resize UNCONFIRMED leg=%s after %s and %d attempt(s) — FAILING CLOSED: the remainder is flattened, never left naked (%s)",
				r.LegSignalID, reqAge.Round(time.Second), r.Attempts, why)
			at.flattenRemainderAfterResizeFailure(why)
			_ = at.store.PartialClose().MarkStopResizeFailed(r.ID)
			continue
		}
		entrySignal := strings.TrimSuffix(r.LegSignalID, "-sl")
		if cerr := nt.CancelBracketLeg(entrySignal, "sl"); cerr != nil {
			at.logWarnf("🧩 stop resize re-request SEND FAILED leg=%s: %v", r.LegSignalID, cerr)
		}
		at.logWarnf("🧩 stop resize UNCONFIRMED leg=%s after %s (%s) — re-requested, attempt %d of %d; the replacement waits",
			r.LegSignalID, reqAge.Round(time.Second), why, r.Attempts+1, cap)
	}
}

// flattenRemainderAfterResizeFailure is the fail-closed exit: close_position
// flattens the WHOLE remaining position at market. It runs only when a resize
// cannot be confirmed — the remainder is never left naked and the old stop is
// never blindly replaced.
func (at *AutoTrader) flattenRemainderAfterResizeFailure(why string) {
	nt := at.armedTrader()
	if nt == nil {
		at.logErrorf("🧩 stop resize FAIL-CLOSED but no bound trader to flatten: %s", why)
		return
	}
	at.logWarnf("🧩 stop resize FAIL-CLOSED — flattening the remainder: %s", why)
	if _, err := nt.CloseLong(at.futuresSymbol(), 0); err != nil {
		at.logErrorf("🧩 stop resize FAIL-CLOSED flatten SEND failed: %v", err)
	}
}

// recordStopResizeReport routes a terminal order_update for a protective leg
// ("<signal>-sl") into the pending resize's report columns — the SAME evidence
// the #309 settlement pass reads. Called from onArmedOrderUpdate for frames it
// would otherwise drop (leg frames do not match the entry's signal id).
func (at *AutoTrader) recordStopResizeReport(u nt.OrderUpdatePayload) {
	if !partialCloseEnabled() || at.store == nil || u.OrderName == "" {
		return
	}
	if !strings.EqualFold(strings.TrimSpace(u.State), "cancelled") && !strings.EqualFold(strings.TrimSpace(u.State), "filled") {
		return
	}
	rows, err := at.store.PartialClose().ListStopResizePending(at.id)
	if err != nil {
		return
	}
	for i := range rows {
		if strings.EqualFold(strings.TrimSpace(rows[i].LegSignalID), strings.TrimSpace(u.OrderName)) {
			if err := at.store.PartialClose().RecordStopResizeReport(rows[i].ID, armedReportNow(), strings.ToLower(strings.TrimSpace(u.State))); err != nil {
				at.logWarnf("🧩 stop resize report record failed for leg=%s: %v", u.OrderName, err)
			}
			return
		}
	}
}
