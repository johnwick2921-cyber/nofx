package trader

import (
	"context"
	"io"
	"net"
	nt "nofx/provider/ninjatrader"
	"nofx/store"
	adapter "nofx/trader/ninjatrader"
	"testing"
	"time"
)

func TestOrderedTCPEntryExitChronology(t *testing.T) {
	for _, inverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "entry_then_exit", true: "exit_then_entry"}[inverse], func(t *testing.T) {
			at := class33Trader(t)
			class33Seed(t, at, "S1", "ordered-entry", store.ProcessBootID(), store.StateWorking)
			srv := nt.NewTCPServer(nil)
			srv.SetAddrForTest("127.0.0.1:0")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := srv.Start(ctx); err != nil {
				t.Fatal(err)
			}
			defer srv.Stop()
			tr := adapter.NewTCPTrader(srv, "MNQ", "Sim101")
			at.trader = tr
			at.installNTOrderedExecutions(tr)
			conn, err := net.Dial("tcp", srv.ListenAddrForTest().String())
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			go io.Copy(io.Discard, conn)
			first := nt.OrderUpdatePayload{SignalID: "ordered-entry", OrderName: "ordered-entry", State: "partfilled", Quantity: 1, FillPrice: 100, Symbol: "MNQ", Account: "Sim101"}
			second := first
			second.State = "filled"
			second.Quantity = 2
			second.FillPrice = 105
			close := nt.PositionClosePayload{SignalID: first.SignalID, ExitOrderID: "ordered-exit", Symbol: "MNQ", Account: "Sim101", PositionSide: "LONG", Quantity: 1, ExitPrice: 120, ExitReason: "tp"}
			write := func(kind nt.FrameType, p any) {
				t.Helper()
				if err := nt.WriteFrame(conn, kind, p); err != nil {
					t.Fatal(err)
				}
			}
			// A positions frame is a readLoop barrier, not a fabricated DB expectation.
			barrier := func(account string) {
				t.Helper()
				write(nt.FramePositions, map[string]any{"account": account, "positions": []nt.OpenPosition{}})
				until := time.Now().Add(3 * time.Second)
				for time.Now().Before(until) {
					if _, ok := srv.PositionsFor(account); ok {
						return
					}
					time.Sleep(time.Millisecond)
				}
				t.Fatal("readLoop barrier timed out")
			}
			write(nt.FrameOrderUpdate, first)
			write(nt.FrameFill, nt.FillPayload{SignalID: first.SignalID, Symbol: first.Symbol, Account: first.Account, Side: "long", Quantity: 1, FillPrice: 100, Status: "partial"})
			if inverse {
				write(nt.FramePositionClose, close)
				write(nt.FrameOrderUpdate, second)
			} else {
				write(nt.FrameOrderUpdate, second)
				write(nt.FramePositionClose, close)
			}
			barrier("barrier-one")
			check := func() {
				t.Helper()
				rows, err := at.store.Position().GetOpenPositions(at.id)
				if err != nil || len(rows) != 1 {
					t.Fatalf("open rows=%+v err=%v", rows, err)
				}
				p := rows[0]
				basis, pnl := 105.0, 30.0
				if inverse {
					basis, pnl = 110, 40
				}
				if p.PnlCorrected != nil || p.ExitTime != 0 && inverse {
					t.Fatalf("continuation retained final fields: %+v", p)
				}
				if p.Quantity != 1 || p.EntryQuantity != 2 || p.EntryPrice != basis || p.RealizedPnL != pnl || p.EntryNotional == nil || *p.EntryNotional != 210 {
					t.Fatalf("wrong causal position: %+v", p)
				}
			}
			check()
			// Replay the entire stream, including an older cumulative frame and receipt.
			write(nt.FrameOrderUpdate, first)
			write(nt.FrameOrderUpdate, second)
			write(nt.FramePositionClose, close)
			alien := close
			alien.ExitOrderID = "alien-exit"
			alien.SignalID = "unrelated-entry"
			write(nt.FramePositionClose, alien)
			barrier("barrier-two")
			check()
			manual := close
			manual.ExitOrderID = "manual-final"
			manual.SignalID = "Close"
			manual.ExitReason = "manual"
			manual.ExitPrice = 130
			write(nt.FramePositionClose, manual)
			barrier("barrier-three")
			var final store.TraderPosition
			if err := at.store.GormDB().Where("entry_order_id = ?", first.SignalID).First(&final).Error; err != nil {
				t.Fatal(err)
			}
			if final.Status != "CLOSED" || final.RealizedPnL != 80 || final.EntryPrice != 105 || final.ExitPrice != 125 || final.PnlCorrected == nil || *final.PnlCorrected != 80 {
				t.Fatalf("manual final accounting: %+v", final)
			}
			write(nt.FrameOrderUpdate, second)
			write(nt.FramePositionClose, manual)
			barrier("barrier-four")
			write(nt.FramePositions, map[string]any{"account": "Sim101", "positions": []nt.OpenPosition{}})
			barrier("barrier-flat")
			positions, err := tr.GetPositions()
			if err != nil || len(positions) != 0 {
				t.Fatalf("final broker flat snapshot: %+v %v", positions, err)
			}
			rows, _ := at.store.Position().GetOpenPositions(at.id)
			if len(rows) != 0 {
				t.Fatal("duplicate reopened terminal position")
			}

		})
	}
}
