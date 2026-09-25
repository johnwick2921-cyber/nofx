package trader

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
	ntTrader "nofx/trader/ninjatrader"
)

// W117 slice A — trader-side requirement pins (R2 + R4), at the production call
// sites: the REAL onArmedOrderUpdate, the REAL store methods the armed pass
// calls, and a REAL TCPServer fed by a fake client with the AddOn's REAL frame
// order (order_update, THEN order_snapshot).

// R2 — the full RED the CTO named: a fill lands, then the armed pass's
// RequestCancel must leave the row FILLED and the position must exist.
func TestLateFillSurvivesThePassCancelRequest(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}})
	ledger := st.ArmedOrders()
	if err := ledger.UpsertArm(&store.ArmedOrderDB{TraderID: at.id, PlanID: "2026-09-25:R2", Version: 1, Session: "NY", Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 95, TargetPx: 110, State: "working", SignalID: "sig-r2"}); err != nil {
		t.Fatal(err)
	}
	at.onArmedOrderUpdate(ntwire.OrderUpdatePayload{SignalID: "sig-r2", State: "filled", FillPrice: 29347.25, Symbol: "MNQ", Account: "Sim101", Quantity: 1}, ledger)

	row, err := ledger.FindBySignal(at.id, "sig-r2")
	if err != nil || row == nil {
		t.Fatalf("armed row: %v", err)
	}
	if row.State != "filled" {
		t.Fatalf("fill must land first: %+v", row)
	}

	// The armed pass races in with its own cancel call (the SAME store method
	// maybeManageArmedOrdersAt uses).
	if err := ledger.RequestCancel(row.ID, "pass invalidation", time.Now().UnixMilli()); err != nil {
		t.Fatalf("request cancel: %v", err)
	}
	row2, _ := ledger.FindBySignal(at.id, "sig-r2")
	if row2.State != "filled" {
		t.Fatalf("the pass's RequestCancel moved a FILLED row to %q — a late fill was unwound", row2.State)
	}
	// And the position the fill materialized must still exist.
	pos, err := st.Position().GetOpenPositionBySymbol(at.id, "MNQ", "LONG")
	if err != nil || pos == nil {
		t.Fatalf("the filled entry's position must exist after the pass's cancel: %v", err)
	}
	if pos.Status != "OPEN" {
		t.Fatalf("the position must stay OPEN, got %q", pos.Status)
	}
}

// R4 — recordAcceptedRisk must not read the PRE-change broker book. The AddOn
// sends order_update THEN order_snapshot (VLTraderTCPClient.cs ~1948 then
// ~1955): the ordered worker waits for the post-update snapshot watermark, so
// the accepted-risk row carries the NEW book's prices, never the old ones.
func TestAcceptedRiskUsesThePostChangeBook(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "r4.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	srv := ntwire.NewTCPServer(nil)
	srv.SetAddrForTest("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := srv.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Stop() })

	var c net.Conn
	for i := 0; i < 100 && c == nil; i++ {
		cc, err := net.Dial("tcp", srv.ListenAddrForTest().String())
		if err == nil {
			deadline := time.Now().Add(20 * time.Millisecond)
			for time.Now().Before(deadline) && !srv.IsConnected() {
				time.Sleep(2 * time.Millisecond)
			}
			if srv.IsConnected() {
				c = cc
			} else {
				_ = cc.Close()
			}
		} else if cc != nil {
			_ = cc.Close()
		}
		time.Sleep(5 * time.Millisecond)
	}
	if c == nil {
		t.Fatal("fake client never connected")
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := ntwire.WriteFrame(c, ntwire.FrameHello, ntwire.HelloPayload{ProtocolVersion: ntwire.ProtocolVersion, Source: "vltrader-addon"}); err != nil {
		t.Fatal(err)
	}

	nt := ntTrader.NewTCPTrader(srv, "MNQ", "Sim101")
	at := &AutoTrader{id: "trader-1", store: st, exchange: "ninjatrader", trader: nt}
	ledger := st.ArmedOrders()
	if err := ledger.UpsertArm(&store.ArmedOrderDB{TraderID: at.id, PlanID: "2026-09-25:R4", Version: 1, Session: "NY", Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 95, TargetPx: 110, State: "working", SignalID: "sig-r4"}); err != nil {
		t.Fatal(err)
	}

	unreg, err := srv.RegisterOrderedExecutionsFor("MNQ", "Sim101", ntwire.OrderedExecutionHandlers{
		Order: func(u ntwire.OrderUpdatePayload) { at.onArmedOrderUpdate(u, ledger) },
	})
	if err != nil {
		t.Fatalf("register ordered owner: %v", err)
	}
	defer unreg()

	// The PRE-change book (an earlier snapshot already established it).
	if err := ntwire.WriteFrame(c, ntwire.FrameOrderSnapshot, ntwire.OrderSnapshotPayload{Account: "Sim101", BuildID: "test", Orders: []ntwire.NT8Order{
		{Name: "sig-r4-sl", Type: "stop", StopPrice: 29500, State: "Working"},
		{Name: "sig-r4-tp", Type: "limit", LimitPrice: 29600, State: "Working"},
	}}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(150 * time.Millisecond) // let the read loop file the old book

	// The AddOn's REAL frame order: order_update, THEN order_snapshot.
	if err := ntwire.WriteFrame(c, ntwire.FrameOrderUpdate, ntwire.OrderUpdatePayload{SignalID: "sig-r4", State: "accepted", Symbol: "MNQ", Account: "Sim101", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	if err := ntwire.WriteFrame(c, ntwire.FrameOrderSnapshot, ntwire.OrderSnapshotPayload{Account: "Sim101", BuildID: "test", Orders: []ntwire.NT8Order{
		{Name: "sig-r4-sl", Type: "stop", StopPrice: 29700, State: "Working"},
		{Name: "sig-r4-tp", Type: "limit", LimitPrice: 29800, State: "Working"},
	}}); err != nil {
		t.Fatal(err)
	}

	var rows []store.AcceptedRisk
	deadline := time.Now().Add(5 * time.Second)
	for {
		rows, _ = st.AcceptedRisk().ForSignal("sig-r4")
		if len(rows) > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(rows) == 0 {
		t.Fatal("no accepted-risk row recorded")
	}
	r := rows[0]
	if r.BookSource != "f12" {
		t.Fatalf("the post-change book must be read (BookSource=%q): the worker applied the update before the snapshot landed", r.BookSource)
	}
	if r.AcceptedStopPx == nil || *r.AcceptedStopPx != 29700 {
		t.Fatalf("accepted stop must come from the POST-change snapshot (29700), got %v — the PRE-change book leaked in", r.AcceptedStopPx)
	}
}

// R6 — an exit that arrives BEFORE its cumulative entry update is RETAINED
// (pending receipt) and applied after the entry lands. Production call sites:
// the worker's close handler (recordCloseOrdered → ApplyNT8Exit) and the
// worker's order handler with the post-entry retry (installNTOrderedExecutions).
func TestExitBeforeCumulativeEntryIsRetainedThenApplied(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "r6.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	srv := ntwire.NewTCPServer(nil)
	srv.SetAddrForTest("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := srv.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Stop() })

	var c net.Conn
	for i := 0; i < 100 && c == nil; i++ {
		cc, err := net.Dial("tcp", srv.ListenAddrForTest().String())
		if err == nil {
			deadline := time.Now().Add(20 * time.Millisecond)
			for time.Now().Before(deadline) && !srv.IsConnected() {
				time.Sleep(2 * time.Millisecond)
			}
			if srv.IsConnected() {
				c = cc
			} else {
				_ = cc.Close()
			}
		} else if cc != nil {
			_ = cc.Close()
		}
		time.Sleep(5 * time.Millisecond)
	}
	if c == nil {
		t.Fatal("fake client never connected")
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := ntwire.WriteFrame(c, ntwire.FrameHello, ntwire.HelloPayload{ProtocolVersion: ntwire.ProtocolVersion, Source: "vltrader-addon"}); err != nil {
		t.Fatal(err)
	}

	nt := ntTrader.NewTCPTrader(srv, "MNQ", "Sim101")
	at := &AutoTrader{id: "trader-1", store: st, exchange: "ninjatrader", trader: nt, exchangeID: "ex-1"}
	at.installNTOrderedExecutions(nt)
	t.Cleanup(func() {
		at.orderedExecMu.Lock()
		if at.orderedExecUnreg != nil {
			at.orderedExecUnreg()
		}
		at.orderedExecMu.Unlock()
	})

	// 1. The exit arrives FIRST — no open row exists anywhere.
	if err := ntwire.WriteFrame(c, ntwire.FramePositionClose, ntwire.PositionClosePayload{SignalID: "sig-r6", Symbol: "MNQ", PositionSide: "long", Account: "Sim101", Quantity: 1, ExitPrice: 29360, ExitReason: "sl", ExitTime: time.Now().UTC().Add(-time.Second).Format(time.RFC3339)}); err != nil {
		t.Fatal(err)
	}
	var pending []store.NT8ExitReceipt
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if pending, err = st.Position().PendingNT8Exits("Sim101"); err == nil && len(pending) == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(pending) != 1 {
		t.Fatalf("the early exit must be RETAINED as a pending receipt, got %d (err=%v)", len(pending), err)
	}
	if pending[0].Applied {
		t.Fatal("the pending receipt must not be applied while its entry is absent")
	}

	// 2. The cumulative entry update lands (with its snapshot), then the
	// worker's order handler retries the parked exit.
	ledger := st.ArmedOrders()
	if err := ledger.UpsertArm(&store.ArmedOrderDB{TraderID: at.id, PlanID: "2026-09-25:R6", Version: 1, Session: "NY", Scenario: "S1", Side: "long", EntryPx: 29350, StopPx: 29345, TargetPx: 29380, State: "working", SignalID: "sig-r6"}); err != nil {
		t.Fatal(err)
	}
	if err := ntwire.WriteFrame(c, ntwire.FrameOrderUpdate, ntwire.OrderUpdatePayload{SignalID: "sig-r6", State: "filled", FillPrice: 29350, Symbol: "MNQ", Account: "Sim101", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	if err := ntwire.WriteFrame(c, ntwire.FrameOrderSnapshot, ntwire.OrderSnapshotPayload{Account: "Sim101", BuildID: "test", Orders: []ntwire.NT8Order{}}); err != nil {
		t.Fatal(err)
	}

	// 3. The parked exit applies after the entry: the receipt flips, the
	// position exists and is CLOSED with the receipt's pnl.
	deadline = time.Now().Add(4 * time.Second)
	var applied []store.NT8ExitReceipt
	for time.Now().Before(deadline) {
		if pending, err = st.Position().PendingNT8Exits("Sim101"); err == nil && len(pending) == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(pending) != 0 {
		t.Fatalf("the parked exit must apply once the entry lands; still pending: %d", len(pending))
	}
	_ = applied
	closedRows, cerr := st.Position().GetUngradedClosedPositions(at.id, 0, 10)
	if cerr != nil || len(closedRows) == 0 {
		t.Fatalf("after the entry lands the exit must close the row: %v", cerr)
	}
	if closedRows[0].Status != "CLOSED" || closedRows[0].RealizedPnL == 0 {
		t.Fatalf("the applied exit must close the row with realized pnl: %+v", closedRows[0])
	}
}
