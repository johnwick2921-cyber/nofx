package ninjatrader

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"

	"nofx/discipline"
	"nofx/logger"
	"nofx/market"
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// StartCloseSync consumes position_close frames from the TCP bridge and records
// each close into trader_positions (real exit price + futures realized PnL),
// preserves residual rows and requires fresh account snapshots after executions.
//
// An exit order's completed quantity can be smaller than the held position.
// Record actual executions atomically; whole-position hooks run only when the
// owned row reaches zero. Account-flat proof belongs to the broker snapshot.
func (t *TCPTrader) StartCloseSync(traderID, exchangeID, exchangeType string, st *store.Store) {
	if st == nil {
		return
	}
	// A2 (G1) — record the owning trader id so outbound order frames can stamp it.
	t.mu.Lock()
	t.traderID = traderID
	t.st = st // Entry rejection receipts need the ledger before any placement.
	t.mu.Unlock()
	t.loadExitSnapshotFence(st)
	pb := store.NewPositionBuilder(st.Position())
	t.closeSyncOnce.Do(func() {
		go func() {
			for p := range t.server.SubscribeClosesFor(t.symbol, t.boundAccount) { // P5.4 router-fed (per-symbol)
				t.recordClose(traderID, exchangeID, exchangeType, st, pb, p)
			}
		}()
		// Rejected exit/flatten watcher. The SIM/broker refused a close (e.g. "no
		// market data" with the feed down). The position is STILL OPEN in NT8 — do
		// NOT record a close; raise a loud alarm. Because decision-driven closes no
		// longer mark the DB CLOSED off the mark (see auto_trader_decision.go), the
		// position simply stays open: the next decision cycle re-issues the close (a
		// natural bounded retry) and the periodic reconcile keeps the DB anchored to
		// NT8 truth, so the orphan can't be netted onto by the next entry.
		go func() {
			for r := range t.server.SubscribeRejectsFor(t.symbol, t.boundAccount) { // P5.4 router-fed (per-symbol)
				logger.Warnf("🚨 NT close REJECTED: %s %s — STILL OPEN in NT8, NOT recording closed (reason: %q, account: %s). Will retry on next decision cycle / reconnect.",
					r.Symbol, r.PositionSide, r.Reason, r.Account)
			}
		}()
		// Instrument-info watcher (Phase 4): NT8 reports the RESOLVED instrument's
		// real specs. Cross-check the hardcoded tables — they are CME-correct, so a
		// divergence is a drift signal worth surfacing. For a parked/unknown symbol
		// (no table entry) NT8 is the only source. The tables stay authoritative for
		// the math; this is defense-in-depth + drift detection.
		go func() {
			for in := range t.server.SubscribeInstrumentInfoFor(t.symbol, t.boundAccount) { // P5.4 router-fed (per-symbol)
				tablePV := market.FuturesPointValue(in.Symbol)
				tableTick := market.FuturesTickSize(in.Symbol)
				switch {
				case tablePV <= 0:
					logger.Infof("📐 NT8 instrument_info %s (%s): point_value=%.4g tick=%.6g — no table entry (parked/unknown); NT8 is the source.",
						in.Symbol, in.Contract, in.PointValue, in.TickSize)
				case in.PointValue != tablePV || (tableTick > 0 && in.TickSize != tableTick):
					logger.Warnf("🚨 instrument spec DRIFT %s (%s): NT8 point_value=%.4g tick=%.6g vs table point_value=%.4g tick=%.6g — verify the table.",
						in.Symbol, in.Contract, in.PointValue, in.TickSize, tablePV, tableTick)
				default:
					logger.Infof("📐 NT8 instrument_info %s (%s): point_value=%.4g tick=%.6g — matches table ✓",
						in.Symbol, in.Contract, in.PointValue, in.TickSize)
				}
			}
		}()
		logger.Infof("🔄 NinjaTrader close-sync started (records SL/TP + manual exits; alarms on rejected flattens; cross-checks NT8 instrument specs)")
	})
}

func (t *TCPTrader) recordClose(
	traderID, exchangeID, exchangeType string,
	st *store.Store,
	pb *store.PositionBuilder,
	p ntwire.PositionClosePayload,
) {
	side := strings.ToUpper(strings.TrimSpace(p.PositionSide))
	symbol := t.symbol
	if symbol == "" {
		symbol = p.Symbol
	}
	if p.Symbol != "" && !equalSymbol(p.Symbol, symbol) {
		logger.Warnf("NT8 exit refused: symbol mismatch")
		return
	}
	exitMs := time.Now().UTC().UnixMilli()
	if p.ExitTime != "" {
		ts, err := time.Parse(time.RFC3339Nano, p.ExitTime)
		if err != nil {
			logger.Warnf("NT8 exit refused: invalid exit time: %v", err)
			return
		}
		exitMs = ts.UnixMilli()
	}
	// Completed order receipts use the broker order ID, independent of echo seq.
	// Legacy frames are accepted only with a known bot entry/operation identity;
	// generic manual names are not unique across separate orders.
	reason := strings.ToLower(strings.TrimSpace(p.ExitReason))
	leg := "exit"
	if reason == "sl" || reason == "tp" {
		leg = reason
	}
	identityParts := []any{p.Account, symbol, side, p.ExitOrderID}
	if strings.TrimSpace(p.ExitOrderID) == "" {
		if !t.knownLegacyExitIdentity(st, p, symbol, side, reason) {
			logger.Warnf("NT8 exit refused: missing unique exit order identity account=%s signal=%s", p.Account, p.SignalID)
			return
		}
		identityParts = []any{p.Account, symbol, side, p.SignalID, leg, p.Seq}
	}
	identity, _ := json.Marshal(identityParts)
	key := fmt.Sprintf("nt8-exit-v2-%x", sha256.Sum256(identity))
	receipt := store.NT8ExitReceipt{ID: key, Account: p.Account, Symbol: symbol, Side: side, SignalID: p.SignalID, TraderID: p.TraderID,
		ExchangeID: exchangeID, ExchangeType: exchangeType, Reason: reason, Quantity: float64(p.Quantity), Price: p.ExitPrice,
		PointValue: market.FuturesPointValue(symbol), ExitMs: exitMs, ReceivedMs: time.Now().UnixMilli()}
	// No commission is present on this wire frame. Preserve existing row fees;
	// do not invent or charge the entry commission again for each partial exit.
	result, err := st.Position().ApplyNT8Exit(receipt)
	if err != nil {
		logger.Warnf("NT8 exit receipt refused/uncommitted account=%s signal=%s qty=%d: %v", p.Account, p.SignalID, p.Quantity, err)
		return
	}
	if result.Pending || result.Applied {
		t.noteExitSnapshotFence(receipt)
	}
	if result.Pending {
		logger.Warnf("NT8 exit receipt pending owned entry evidence account=%s signal=%s qty=%d (not flat)", p.Account, p.SignalID, p.Quantity)
		return
	}
	t.finishNT8Exit(receipt, result)
}

func (t *TCPTrader) noteExitSnapshotFence(receipt store.NT8ExitReceipt) {
	if !strings.EqualFold(t.boundAccount, receipt.Account) {
		return
	}
	t.mu.Lock()
	if receipt.ReceivedMs > t.positionsAfterMs {
		t.positionsAfterMs = receipt.ReceivedMs
	}
	t.mu.Unlock()
}

func (t *TCPTrader) finishNT8Exit(receipt store.NT8ExitReceipt, result store.NT8ExitResult) {
	if !result.Applied {
		return
	} // durable duplicate; no hooks or cache mutation
	owner := result.Position
	if strings.EqualFold(t.boundAccount, receipt.Account) {
		t.mu.Lock()
		// Only reduce the cache belonging to this entry. Snapshot-backed positions
		// remain authoritative; an unrelated or unidentified cached fill is not ours.
		if t.hasFill && owner.EntryOrderID != "" && t.lastFill.SignalID == owner.EntryOrderID {
			if result.Closed {
				t.hasFill = false
			} else if int(owner.Quantity) < t.lastFill.Quantity {
				t.lastFill.Quantity = int(owner.Quantity)
			}
		}
		t.mu.Unlock()
	}
	if !result.Closed {
		logger.Warnf("NT8 partial exit recorded: row=%d actual_qty=%.0f residual=%.0f price=%.2f pnl=%.2f (still OPEN)", owner.ID, result.Quantity, owner.Quantity, receipt.Price, result.RealizedPnL)
		return
	}
	if OnPositionClosed != nil {
		OnPositionClosed(owner.TraderID, owner.ID)
	}
	if strings.EqualFold(receipt.Reason, "sl") {
		discipline.NoteStopLossExit(owner.TraderID, receipt.Symbol, receipt.Side, receipt.Price, receipt.ExitMs)
	}
	logger.Warnf("📕 NT position closed: %s %s qty=%.2f exit=%.2f reason=%s pnl=%.2f (owner=%s row=%d final_execution_qty=%.0f)", receipt.Symbol, receipt.Side, owner.Quantity, owner.ExitPrice, receipt.Reason, owner.RealizedPnL, owner.TraderID, owner.ID, result.Quantity)
}

func (t *TCPTrader) retryPendingNT8Exits(st *store.Store) {
	receipts, err := st.Position().PendingNT8Exits(t.boundAccount)
	if err != nil {
		logger.Warnf("NT8 pending exit read failed: %v", err)
		return
	}
	for _, receipt := range receipts {
		result, err := st.Position().ApplyNT8Exit(receipt)
		if err != nil {
			logger.Warnf("NT8 pending exit unresolved signal=%s: %v", receipt.SignalID, err)
			continue
		}
		if result.Applied {
			t.noteExitSnapshotFence(receipt)
		}
		t.finishNT8Exit(receipt, result)
	}
}

// OnPositionClosed (Phase 4, final-bundle 2026-08-19) is the package-level
// close-event hook: the trader layer registers a dispatcher so a confirmed
// position close (close_sync priced close, reconcile priced/flat close)
// triggers exactly one post-exit rescan for the OWNING trader. Nil = no-op.
var OnPositionClosed func(traderID string, positionID int64)

// knownLegacyExitIdentity limits pre-exit_order_id compatibility to bot-owned
// bracket lineage or UUID operation names. Generic manual order names refuse.
func (t *TCPTrader) knownLegacyExitIdentity(st *store.Store, p ntwire.PositionClosePayload, symbol, side, reason string) bool {
	if _, err := uuid.Parse(p.SignalID); err == nil {
		return true
	}
	if reason != "sl" && reason != "tp" {
		return false
	}
	var count int64
	err := st.GormDB().Model(&store.TraderPosition{}).Where("account = ? AND symbol = ? AND UPPER(side) = ? AND entry_order_id = ?", p.Account, symbol, side, p.SignalID).Count(&count).Error
	return err == nil && count > 0 && p.SignalID != ""
}
