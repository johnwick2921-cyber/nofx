package ninjatrader

import (
	"nofx/store"
	"testing"
)

func TestShortPartialCloseUsesActualQuantityAndFuturesPointValue(t *testing.T) {
	tr, st, row := partialCloseFixture(t)
	if err := st.GormDB().Model(&store.TraderPosition{}).Where("id = ?", row.ID).Update("side", "SHORT").Error; err != nil {
		t.Fatal(err)
	}
	p := closeFixtureFrame()
	p.PositionSide = "short"
	p.ExitPrice = 90
	tr.recordClose("owned", "ex", "ninjatrader", st, store.NewPositionBuilder(st.Position()), p)
	var got store.TraderPosition
	st.GormDB().First(&got, row.ID)
	if got.Status != "OPEN" || got.Quantity != 2 || got.RealizedPnL != 20 {
		t.Fatalf("short partial wrong: %+v", got)
	}
	var fill store.TraderFill
	st.GormDB().First(&fill)
	if fill.Side != "BUY" || fill.Quantity != 1 || fill.RealizedPnL != 20 {
		t.Fatalf("short fill wrong: %+v", fill)
	}
}
