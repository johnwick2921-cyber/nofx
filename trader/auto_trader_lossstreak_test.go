package trader

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/store"
)

func g6Harness(t *testing.T, n int) (*AutoTrader, *store.Store) {
	t.Helper()
	st, err := store.New(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	at := &AutoTrader{
		id:       "t1",
		exchange: "ninjatrader",
		store:    st,
		config:   AutoTraderConfig{StrategyConfig: &store.StrategyConfig{Regime: &store.RegimeConfig{LossStreakN: &n}}},
	}
	return at, st
}

func g6Close(t *testing.T, st *store.Store, pnl float64, at int64) {
	t.Helper()
	pos := &store.TraderPosition{
		TraderID:   "t1",
		Symbol:     "MNQ",
		Side:       "SHORT",
		EntryPrice: 29400,
		Quantity:   1,
		EntryTime:  at - 60000,
		Status:     "OPEN",
	}
	if err := st.Position().Create(pos); err != nil {
		t.Fatalf("create pos: %v", err)
	}
	if _, err := st.Position().ClosePosition(pos.ID, 29410, "", pnl, 0, "test"); err != nil {
		t.Fatalf("close pos: %v", err)
	}
	_ = st.Position().ClosePosition // idempotency guard not needed
}

func TestG6LossStreakPause(t *testing.T) {
	at, st := g6Harness(t, 4)
	base := time.Now().Add(-6 * time.Minute).UnixMilli()
	for i := 0; i < 4; i++ {
		g6Close(t, st, -50, base+int64(i)*60_000)
	}
	ctx := &kernel.Context{}
	at.observeLossStreak(ctx)
	if !ctx.LossStreakPaused {
		t.Fatalf("4 consecutive losers must pause entries")
	}
	if !strings.Contains(ctx.LossStreakMsg, "loss_streak: 4 consecutive losers") || !strings.Contains(ctx.LossStreakMsg, "paused until") {
		t.Fatalf("bad refusal: %q", ctx.LossStreakMsg)
	}
}

func TestG6WinnerResetsStreak(t *testing.T) {
	at, st := g6Harness(t, 4)
	base := time.Now().Add(-8 * time.Minute).UnixMilli()
	// 4 losers, then a NEWER winner → streak reset.
	for i := 0; i < 4; i++ {
		g6Close(t, st, -50, base+int64(i)*60_000)
	}
	g6Close(t, st, 120, base+5*60_000)
	ctx := &kernel.Context{}
	at.observeLossStreak(ctx)
	if ctx.LossStreakPaused {
		t.Fatalf("a winner must reset the streak")
	}
}

func TestG6OffSwitch(t *testing.T) {
	at, st := g6Harness(t, 0) // 0 = off
	base := time.Now().Add(-6 * time.Minute).UnixMilli()
	for i := 0; i < 4; i++ {
		g6Close(t, st, -50, base+int64(i)*60_000)
	}
	ctx := &kernel.Context{}
	at.observeLossStreak(ctx)
	if ctx.LossStreakPaused {
		t.Fatalf("loss_streak_n=0 must disable the gate")
	}
}

func TestG6TimerExpiryResumes(t *testing.T) {
	t.Setenv("LOSS_STREAK_PAUSE_MIN", "0")
	at, st := g6Harness(t, 4)
	base := time.Now().Add(-6 * time.Minute).UnixMilli()
	for i := 0; i < 4; i++ {
		g6Close(t, st, -50, base+int64(i)*60_000)
	}
	ctx := &kernel.Context{}
	at.observeLossStreak(ctx)
	if ctx.LossStreakPaused {
		t.Fatalf("expired pause window must resume entries")
	}
}
