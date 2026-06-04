package ninjatrader

import (
	"strings"
	"time"

	"nofx/logger"
	"nofx/market"
	"nofx/store"
)

// Position reconcile — the durable single-source-of-truth anchor for NT8.
//
// The AI-decision write records entry_price = marketData.CurrentPrice (the last
// 5m BarCache close), a decision-time *reference* that goes stale (and freezes
// entirely when NT8 stops streaming bars). This loop periodically compares the
// open trader_positions rows against the NT8 positions snapshot (the truth) and:
//
//   (b) ENTRY TRUTH  — for a row NT8 still holds, replaces a stale entry_price
//       with the NT8 position average (Position.AveragePrice). close-sync's
//       realized-PnL formula is unchanged; it simply consumes a correct entry.
//   ORPHAN CLEAR     — for an OPEN row NT8 reports FLAT, marks it closed so the
//       phantom clears and GetPositions / the risk gate agree with NT8.
//
// SAFETY: acts ONLY on a positive snapshot (PositionsFor ok == true). When NT8
// has not reported positions (disconnected / not-yet-deployed), it does NOTHING
// — never closing a real position on missing data. A grace window also exempts
// freshly-opened rows whose positions frame may not have arrived yet, and rows
// are only judged against the account NT8 is currently reporting.
const (
	reconcileInterval = 20 * time.Second
	// orphanGraceMs: a row younger than this is not orphan-closed — its positions
	// frame (PositionUpdate) may simply not have arrived yet after the open.
	orphanGraceMs = 120_000
)

// StartPositionReconcile launches the periodic reconcile goroutine (idempotent
// via reconcileOnce). Wired next to StartCloseSync for the ninjatrader exchange.
func (t *TCPTrader) StartPositionReconcile(traderID, exchangeID, exchangeType string, st *store.Store) {
	if st == nil {
		return
	}
	t.reconcileOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(reconcileInterval)
			defer ticker.Stop()
			for range ticker.C {
				t.reconcilePositions(traderID, st)
			}
		}()
		logger.Infof("🔧 NinjaTrader position-reconcile started (anchors entry_price to NT8 avg + clears orphan rows)")
	})
}

// reconcilePositions runs one reconcile pass. See file header for semantics.
func (t *TCPTrader) reconcilePositions(traderID string, st *store.Store) {
	acct := ""
	if a := t.server.CurrentAccount(); a != nil {
		acct = *a
	}
	snap, ok := t.server.PositionsFor(acct)
	if !ok {
		// NT8 has not reported positions for this account — do NOT touch the DB.
		return
	}

	// NT8-held positions keyed "SYMBOL|SIDE" → average price (canonical form).
	held := make(map[string]float64, len(snap))
	heldQty := make(map[string]float64, len(snap))
	for _, p := range snap {
		if p.Quantity == 0 {
			continue
		}
		side := "LONG"
		if strings.EqualFold(p.Side, "short") {
			side = "SHORT"
		}
		key := market.Normalize(p.Symbol) + "|" + side
		held[key] = p.AvgPrice
		heldQty[key] = float64(p.Quantity)
	}

	rows, err := st.Position().GetOpenPositions(traderID)
	if err != nil {
		logger.Warnf("ninjatrader/tcp: reconcile list open positions failed: %v", err)
		return
	}
	nowMs := time.Now().UTC().UnixMilli()
	for _, row := range rows {
		// Only judge rows for the account NT8 is currently reporting (avoids
		// closing another account's positions during a transient account switch).
		if row.Account != "" && acct != "" && row.Account != acct {
			continue
		}
		key := row.Symbol + "|" + row.Side
		if avg, isHeld := held[key]; isHeld {
			// (b) Entry truth — anchor a stale entry to the NT8 average.
			if avg > 0 && row.EntryPrice != avg {
				if err := st.Position().UpdateEntryPrice(row.ID, avg); err != nil {
					logger.Warnf("ninjatrader/tcp: reconcile entry update failed (row %d): %v", row.ID, err)
				} else {
					logger.Infof("🔧 reconcile: %s %s entry %.2f → NT8 avg %.2f (row %d)",
						row.Symbol, row.Side, row.EntryPrice, avg, row.ID)
				}
			}
			// (c) QTY TRUTH — alarm when NT8's held contract count diverges from the
			// DB row. This is the id=45→id=46 net-2 tell: an unclosed prior contract
			// (rejected flatten) the next entry netted onto, so NT8 holds 2 while the
			// row says 1. Log-only (no auto-flatten): loudly flagging the divergence
			// is the safe SIM action; the close routing now prevents the phantom that
			// causes it, and the entry-anchor above keeps the price honest.
			if hq, ok := heldQty[key]; ok && row.Quantity > 0 && hq != row.Quantity {
				logger.Warnf("🚨 reconcile QTY DIVERGENCE: %s %s — NT8 holds %.0f but DB row %d has %.0f. Likely an unclosed prior contract netted onto by a later entry (id=45→46 class). DB qty NOT auto-changed — investigate.",
					row.Symbol, row.Side, hq, row.ID, row.Quantity)
			}
		} else {
			// Orphan — NT8 reports flat for this (symbol,side) but the row is OPEN.
			// Skip very fresh rows (positions frame may still be in flight).
			if nowMs-row.EntryTime < orphanGraceMs {
				continue
			}
			// NT8 is FLAT but the exit fill was never captured (close-sync missed the
			// position_close frame). We must clear the phantom, but the real exit price
			// and realized P&L are UNKNOWN — record the marker (entry-as-exit / 0 are
			// placeholders only). Every P&L stat/UI treats this as "unknown", NOT a
			// false $0 breakeven. Never fabricate an exit NT8 didn't give us.
			if err := st.Position().ClosePosition(row.ID, row.EntryPrice, "reconcile", 0, 0, store.CloseReasonReconcileFlat); err != nil {
				logger.Warnf("ninjatrader/tcp: reconcile orphan-close failed (row %d): %v", row.ID, err)
			} else {
				logger.Infof("🔧 reconcile: closed orphan %s %s (NT8 flat) row=%d", row.Symbol, row.Side, row.ID)
			}
		}
	}
}
