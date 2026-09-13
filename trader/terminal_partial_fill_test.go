package trader

import (
	"gorm.io/gorm"
	"math"
	nt "nofx/provider/ninjatrader"
	"nofx/store"
	"sync"
	"testing"
	"time"
)

func TestIndependentTerminalCancelPositiveQuantity(t *testing.T) {
	at := class33Trader(t)
	class33Seed(t, at, "terminal-positive", "entry-terminal", store.ProcessBootID(), store.StateWorking)
	at.onArmedOrderUpdate(nt.OrderUpdatePayload{SignalID: "entry-terminal", OrderName: "entry-terminal", State: "cancelled", FillPrice: 29044, Quantity: 1, Symbol: "MNQ", Account: "Sim101"}, at.store.ArmedOrders())
	var arm store.ArmedOrderDB
	at.store.GormDB().Where("signal_id = ?", "entry-terminal").First(&arm)
	positions, _ := at.store.Position().GetOpenPositions(at.id)
	t.Logf("terminal cumulativeqty1 => armstate=%s fillprice=%g openrows=%d", arm.State, arm.FillPrice, len(positions))
	if len(positions) != 1 {
		t.Errorf("positive cumulative fill should materialize position before terminal entry retirement")
	}
}

func TestTerminalCumulativeFillLifecycle(t *testing.T) {
	for _, terminal := range []string{"cancelled", "rejected"} {
		if nt.ClassifyOrderState(terminal) != nt.LivenessTerminal {
			t.Fatal("fixture must be a terminal broker receipt")
		}
		t.Run(terminal, func(t *testing.T) {
			at := class33Trader(t)
			class33Seed(t, at, "terminal", "entry-terminal", store.ProcessBootID(), store.StateCancelPending)
			u := nt.OrderUpdatePayload{SignalID: "entry-terminal", OrderName: "entry-terminal", State: terminal, FillPrice: 29044, Quantity: 2, Symbol: "MNQ", Account: "Sim101"}
			ledger := at.store.ArmedOrders()
			at.onArmedOrderUpdate(u, ledger)
			at.onArmedOrderUpdate(u, ledger)
			rows, err := at.store.Position().GetOpenPositions(at.id)
			if err != nil || len(rows) != 1 {
				t.Fatalf("rows=%v err=%v", rows, err)
			}
			pos := rows[0]
			if pos.Quantity != 2 || pos.EntryQuantity != 2 || pos.EntryOrderID != u.SignalID || pos.Account != u.Account || pos.PlanID != "p1" || pos.CitedScenarioID != "terminal" {
				t.Fatalf("lost exact fill lineage: %+v", pos)
			}
			// An exit reduces held quantity, but lifetime entry remains two contracts.
			if err := at.store.GormDB().Model(&store.TraderPosition{}).Where("id = ?", pos.ID).Update("quantity", 1).Error; err != nil {
				t.Fatal(err)
			}
			at.onArmedOrderUpdate(u, ledger) // duplicate must not restore the exited contract
			u.Quantity = 3
			at.onArmedOrderUpdate(u, ledger) // one additional actual fill, add delta one
			u.Quantity = 1
			at.onArmedOrderUpdate(u, ledger) // reordered older cumulative frame
			if err := at.store.GormDB().First(pos, pos.ID).Error; err != nil {
				t.Fatal(err)
			}
			if pos.Quantity != 2 || pos.EntryQuantity != 3 {
				t.Fatalf("cumulative replay corrupted residual: %+v", pos)
			}
			var arm store.ArmedOrderDB
			if err := ledger.DB().Where("signal_id = ?", u.SignalID).First(&arm).Error; err != nil {
				t.Fatal(err)
			}
			if arm.State != "filled" || arm.FillQuantity != 3 || arm.FillPrice != 29044 {
				t.Fatalf("fill truth lost: %+v", arm)
			}
			if err := at.store.GormDB().Model(pos).Update("status", "CLOSED").Error; err != nil {
				t.Fatal(err)
			}
			u.Quantity = 3
			at.onArmedOrderUpdate(u, ledger)
			u.Quantity = 4 // newly reported evidence after close is recorded, not fabricated into a new hold
			at.onArmedOrderUpdate(u, ledger)
			open, _ := at.store.Position().GetOpenPositions(at.id)
			if len(open) != 0 {
				t.Fatalf("closed entry resurrected: %+v", open)
			}
			var count int64
			at.store.GormDB().Model(&store.TraderPosition{}).Where("trader_id = ?", at.id).Count(&count)
			if count != 1 {
				t.Fatalf("duplicate entry rows=%d", count)
			}
		})
	}
}

func TestTerminalFillDoesNotInventOrAdoptPosition(t *testing.T) {
	for _, mode := range []string{"zero quantity", "missing price", "missing account", "protective leg", "different symbol", "unrelated entry", "same signal wrong account", "prior cancellation"} {
		t.Run(mode, func(t *testing.T) {
			at := class33Trader(t)
			class33Seed(t, at, "terminal", "entry-terminal", store.ProcessBootID(), store.StateWorking)
			u := nt.OrderUpdatePayload{SignalID: "entry-terminal", OrderName: "entry-terminal", State: "cancelled", FillPrice: 29044, Quantity: 2, Symbol: "MNQ", Account: "Sim101"}
			switch mode {
			case "zero quantity":
				u.Quantity = 0
			case "missing price":
				u.FillPrice = 0
			case "missing account":
				u.Account = ""
			case "protective leg":
				u.OrderName += "-sl"
			case "different symbol":
				u.Symbol = "ES"
			case "prior cancellation":
				first := u
				first.Quantity = 0
				first.FillPrice = 0
				at.onArmedOrderUpdate(first, at.store.ArmedOrders())
			case "unrelated entry", "same signal wrong account":
				entry, account := "other-entry", "Sim101"
				if mode == "same signal wrong account" {
					entry = u.SignalID
					account = "SimOther"
				}
				if err := at.store.Position().Create(&store.TraderPosition{TraderID: at.id, Symbol: "MNQ", Side: "LONG", Account: account, EntryOrderID: entry, Quantity: 1, EntryQuantity: 1, EntryPrice: 29000, EntryTime: 1, Status: "OPEN", PlanID: "untouched"}); err != nil {
					t.Fatal(err)
				}
			}
			at.onArmedOrderUpdate(u, at.store.ArmedOrders())
			rows, _ := at.store.Position().GetOpenPositions(at.id)
			want := 0
			if mode == "unrelated entry" || mode == "same signal wrong account" || mode == "prior cancellation" {
				want = 1
			}
			if len(rows) != want {
				t.Fatalf("rows=%d want=%d", len(rows), want)
			}
			if mode == "unrelated entry" || mode == "same signal wrong account" {
				if rows[0].PlanID != "untouched" || rows[0].Quantity != 1 {
					t.Fatalf("adopted unrelated position: %+v", rows[0])
				}
			}
			if mode == "prior cancellation" && rows[0].Quantity != 2 {
				t.Fatalf("late terminal fill lost: %+v", rows[0])
			}
		})
	}
}

func TestCumulativeFillPartFilledThenTerminalConcurrentReplay(t *testing.T) {
	at := class33Trader(t)
	class33Seed(t, at, "terminal", "entry-terminal", store.ProcessBootID(), store.StateWorking)
	u := nt.OrderUpdatePayload{SignalID: "entry-terminal", OrderName: "entry-terminal", State: "partfilled", FillPrice: 29044, Quantity: 1, Symbol: "MNQ", Account: "Sim101"}
	at.onArmedOrderUpdate(u, at.store.ArmedOrders())
	u.State = "cancelled"
	u.Quantity = 2
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); at.onArmedOrderUpdate(u, at.store.ArmedOrders()) }()
	}
	wg.Wait()
	rows, _ := at.store.Position().GetOpenPositions(at.id)
	if len(rows) != 1 || rows[0].Quantity != 2 || rows[0].EntryQuantity != 2 {
		t.Fatalf("cumulative replay multiplied fill: %+v", rows)
	}
}

// Force the real exit transaction between the entry callback's position read
// and its UPDATE. The Go race detector alone cannot detect SQL lost updates.
func TestCumulativeEntryDeltaComposesWithConcurrentExitTransaction(t *testing.T) {
	at := class33Trader(t)
	class33Seed(t, at, "terminal", "entry-terminal", store.ProcessBootID(), store.StateWorking)
	u := nt.OrderUpdatePayload{SignalID: "entry-terminal", OrderName: "entry-terminal", State: "partfilled", FillPrice: 29044, Quantity: 2, Symbol: "MNQ", Account: "Sim101"}
	at.onArmedOrderUpdate(u, at.store.ArmedOrders())
	db := at.store.GormDB()
	if err := db.AutoMigrate(&store.NT8ExitReceipt{}); err != nil {
		t.Fatal(err)
	}
	fired := false
	const callback = "test:exit-between-entry-read-update"
	if err := db.Callback().Update().Before("gorm:begin_transaction").Register(callback, func(tx *gorm.DB) {
		fields, ok := tx.Statement.Dest.(map[string]any)
		if fired || !ok || tx.Statement.Table != "trader_positions" || fields["entry_quantity"] != 3 {
			return
		}
		fired = true
		result, err := at.store.Position().ApplyNT8Exit(store.NT8ExitReceipt{ID: "interleaved-exit", Account: "Sim101", Symbol: "MNQ", Side: "LONG", SignalID: u.SignalID, TraderID: at.id, Reason: "tp", Quantity: 1, Price: 29050, PointValue: 2, ExitMs: time.Now().Add(time.Second).UnixMilli()})
		if err != nil || !result.Applied || result.Closed {
			t.Errorf("exit interleave failed: %+v %v", result, err)
			tx.AddError(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Update().Remove(callback) })
	u.Quantity = 3
	u.State = "cancelled"
	at.onArmedOrderUpdate(u, at.store.ArmedOrders())
	rows, _ := at.store.Position().GetOpenPositions(at.id)
	if !fired || len(rows) != 1 || rows[0].Quantity != 2 || rows[0].EntryQuantity != 3 || rows[0].RealizedPnL != 12 {
		t.Fatalf("entry update lost exit: fired=%v rows=%+v", fired, rows)
	}
}

func TestCumulativeEntryRefetchesAfterCompareFailure(t *testing.T) {
	at := class33Trader(t)
	class33Seed(t, at, "terminal", "entry-terminal", store.ProcessBootID(), store.StateWorking)
	u := nt.OrderUpdatePayload{SignalID: "entry-terminal", OrderName: "entry-terminal", State: "partfilled", FillPrice: 29044, Quantity: 1, Symbol: "MNQ", Account: "Sim101"}
	at.onArmedOrderUpdate(u, at.store.ArmedOrders())
	rows, _ := at.store.Position().GetOpenPositions(at.id)
	if len(rows) != 1 {
		t.Fatal("entry missing")
	}
	db := at.store.GormDB()
	fired := false
	const callback = "test:entry-compare-conflict"
	if err := db.Callback().Update().Before("gorm:begin_transaction").Register(callback, func(tx *gorm.DB) {
		fields, ok := tx.Statement.Dest.(map[string]any)
		if fired || !ok || tx.Statement.Table != "trader_positions" || fields["entry_quantity"] != 3 {
			return
		}
		fired = true
		if err := at.store.Position().UpdatePositionQuantityAndPrice(rows[0].ID, 1, 29044, 0); err != nil {
			t.Error(err)
			tx.AddError(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Update().Remove(callback) })
	u.Quantity = 3
	u.State = "cancelled"
	at.onArmedOrderUpdate(u, at.store.ArmedOrders())
	rows, _ = at.store.Position().GetOpenPositions(at.id)
	if !fired || len(rows) != 1 || rows[0].Quantity != 3 || rows[0].EntryQuantity != 3 {
		t.Fatalf("stale delta was reused or dropped: fired=%v rows=%+v", fired, rows)
	}
}

func TestCumulativeEntryNotionalPreservesResidualBasisAcrossExits(t *testing.T) {
	at := class33Trader(t)
	class33Seed(t, at, "terminal", "entry-terminal", store.ProcessBootID(), store.StateWorking)
	u := nt.OrderUpdatePayload{SignalID: "entry-terminal", OrderName: "entry-terminal", State: "partfilled", FillPrice: 100, Quantity: 2, Symbol: "MNQ", Account: "Sim101"}
	at.onArmedOrderUpdate(u, at.store.ArmedOrders())
	exit := func(id string, price, qty float64) {
		t.Helper()
		res, err := at.store.Position().ApplyNT8Exit(store.NT8ExitReceipt{ID: id, Account: "Sim101", Symbol: "MNQ", Side: "LONG", SignalID: u.SignalID, TraderID: at.id, Reason: "tp", Quantity: qty, Price: price, PointValue: 2, ExitMs: time.Now().Add(time.Second).UnixMilli()})
		if err != nil || !res.Applied {
			t.Fatalf("exit: %+v %v", res, err)
		}
	}
	check := func(qty, entryQty, basis, notional float64) {
		t.Helper()
		rows, _ := at.store.Position().GetOpenPositions(at.id)
		if len(rows) != 1 {
			t.Fatalf("rows=%v", rows)
		}
		p := rows[0]
		if p.Quantity != qty || p.EntryQuantity != entryQty || math.Abs(p.EntryPrice-basis) > 1e-9 || p.EntryNotional == nil || math.Abs(*p.EntryNotional-notional) > 1e-9 {
			t.Fatalf("wrong residual basis: %+v", p)
		}
	}
	exit("basis-exit1", 120, 1)
	u.Quantity = 3
	u.FillPrice = 310.0 / 3
	at.onArmedOrderUpdate(u, at.store.ArmedOrders())
	check(2, 3, 105, 310)
	exit("basis-exit2", 115, 1)
	u.Quantity = 4
	u.FillPrice = 430.0 / 4
	u.State = "cancelled"
	at.onArmedOrderUpdate(u, at.store.ArmedOrders())
	check(2, 4, 112.5, 430)
	exit("basis-exit3", 130, 2)
	var closed store.TraderPosition
	if err := at.store.GormDB().Where("trader_id = ? AND entry_order_id = ?", at.id, u.SignalID).First(&closed).Error; err != nil {
		t.Fatal(err)
	}
	if closed.Status != "CLOSED" || closed.Quantity != 4 || closed.EntryQuantity != 4 || math.Abs(closed.EntryPrice-107.5) > 1e-9 || math.Abs(closed.RealizedPnL-130) > 1e-9 {
		t.Fatalf("closed history lost cumulative notional/PnL: %+v", closed)
	}
}

func TestCumulativeEntryLegacyUnknownNotional(t *testing.T) {
	for _, exited := range []bool{false, true} {
		t.Run(map[bool]string{false: "no exits derive", true: "after exit refuse"}[exited], func(t *testing.T) {
			at := class33Trader(t)
			class33Seed(t, at, "terminal", "entry-terminal", store.ProcessBootID(), store.StateWorking)
			u := nt.OrderUpdatePayload{SignalID: "entry-terminal", OrderName: "entry-terminal", State: "partfilled", FillPrice: 100, Quantity: 2, Symbol: "MNQ", Account: "Sim101"}
			at.onArmedOrderUpdate(u, at.store.ArmedOrders())
			rows, _ := at.store.Position().GetOpenPositions(at.id)
			updates := map[string]any{"entry_notional": nil}
			if exited {
				updates["quantity"] = 1
			}
			if err := at.store.GormDB().Model(&store.TraderPosition{}).Where("id = ?", rows[0].ID).Updates(updates).Error; err != nil {
				t.Fatal(err)
			}
			u.Quantity = 3
			u.FillPrice = 310.0 / 3
			u.State = "cancelled"
			at.onArmedOrderUpdate(u, at.store.ArmedOrders())
			rows, _ = at.store.Position().GetOpenPositions(at.id)
			p := rows[0]
			if exited {
				if p.Quantity != 1 || p.EntryQuantity != 2 || p.EntryPrice != 100 || p.EntryNotional != nil {
					t.Fatalf("invented legacy residual basis: %+v", p)
				}
			} else {
				if p.Quantity != 3 || p.EntryQuantity != 3 || p.EntryNotional == nil || *p.EntryNotional != 310 {
					t.Fatalf("no-exit legacy basis not derived: %+v", p)
				}
			}
		})
	}
}
