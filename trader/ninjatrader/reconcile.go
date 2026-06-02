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
	for _, p := range snap {
		if p.Quantity == 0 {
			continue
		}
		side := "LONG"
		if strings.EqualFold(p.Side, "short") {
			side = "SHORT"
		}
		held[market.Normalize(p.Symbol)+"|"+side] = p.AvgPrice
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
		} else {
			// Orphan — NT8 reports flat for this (symbol,side) but the row is OPEN.
			// Skip very fresh rows (positions frame may still be in flight).
			if nowMs-row.EntryTime < orphanGraceMs {
				continue
			}
			if err := st.Position().ClosePosition(row.ID, row.EntryPrice, "reconcile", 0, 0, "reconcile_flat"); err != nil {
				logger.Warnf("ninjatrader/tcp: reconcile orphan-close failed (row %d): %v", row.ID, err)
			} else {
				logger.Infof("🔧 reconcile: closed orphan %s %s (NT8 flat) row=%d", row.Symbol, row.Side, row.ID)
			}
		}
	}
}
