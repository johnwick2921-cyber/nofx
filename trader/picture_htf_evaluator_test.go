package trader

import (
	"strings"
	"testing"
	"time"

	"nofx/market"
	"nofx/store"
)

// pictureHtfTestEnv drives the evaluator through the PRODUCTION call site
// (Evaluate/OnBars) with a seeded FuturesBarsProvider — the same ladder the
// live path reads. The submit seam records calls instead of sending wire.
type pictureHtfTestEnv struct {
	t       *testing.T
	at      *AutoTrader
	st      *store.Store
	eval    *PictureHtfEvaluator
	submits []string
	now     time.Time
}

func newPictureHtfEnv(t *testing.T, cfg store.PictureHtfConfig) *pictureHtfTestEnv {
	t.Helper()
	sc := store.StrategyConfig{DayPlan: &store.DayPlanConfig{PictureHtf: &cfg}}
	sc.RiskControl.MinRiskRewardRatio = 2.5
	at, st := resetTrader(t, sc)
	eval := NewPictureHtfEvaluator(at, store.PictureHtfResolved(&cfg))
	env := &pictureHtfTestEnv{t: t, at: at, st: st, eval: eval}
	orig := pictureHtfSubmitSeam
	pictureHtfSubmitSeam = func(e *PictureHtfEvaluator, row *store.PictureHtfOpportunityDB, stopPx, targetPx, qty float64) error {
		env.submits = append(env.submits, row.OppKey)
		return nil
	}
	origCap := pictureHtfCapabilityProven
	pictureHtfCapabilityProven = func(*AutoTrader) bool { return true }
	t.Cleanup(func() {
		pictureHtfSubmitSeam = orig
		pictureHtfCapabilityProven = origCap
		market.FuturesBarsProvider = nil
	})
	return env
}

// t4h0 is the 4H ladder base (2026-09-13 01:00Z). The chronology:
//
//	4H ladder i=0..8 → pivot 101 knowable at 09-13 17:00Z, pivot 110 knowable
//	at 09-14 09:00Z.
//	H1 ladder i=0..41 → last pair closes 17:00Z (prev 101.0) / 18:00Z
//	(cur 101.5) — after BOTH pivots are knowable.
//	5m ladder ends 18:59:59.999Z → the next 5m interval opens 19:00Z;
//	now = 19:00:00.5Z (14:00:00.5 CT) — inside the 10s entry window.
var t4h0 = time.Date(2026, 9, 13, 1, 0, 0, 0, time.UTC).UnixMilli()

func mkBar(open, span int64, o, h, l, c float64) market.Kline {
	return market.Kline{OpenTime: open, CloseTime: open + span - 1, Open: o, High: h, Low: l, Close: c, Final: true}
}

func tailOf(bars []market.Kline, n int) []market.Kline {
	if n <= 0 || len(bars) <= n {
		return bars
	}
	return bars[len(bars)-n:]
}

// pictureBars4H builds the full ladder: an OLD 110 resistance (i=1), a
// pullback, and the RECENT 101 resistance (i=7) that the H1 pair later
// breaks. No later 4H close exceeds 101, so the 101 level is still ACTIVE
// (not retired) when the H1 crosses it — and the 110 zone above is the
// opposing target. This mirrors reality: old high → consolidation →
// breakout → target the old high.
func pictureBars4H() []market.Kline {
	vals := [][4]float64{
		{108, 110, 106, 109}, {109, 111, 107, 110}, {107, 108, 105, 106}, {105, 106, 103, 104},
		{103, 104, 101, 102}, {100, 101, 99, 100.5}, {98.5, 100, 97.5, 99}, {99, 105, 96, 101},
		{96, 97, 94, 95}, {94, 95, 92, 93}, {93, 94, 91, 92},
	}
	out := make([]market.Kline, 0, len(vals))
	for i, v := range vals {
		out = append(out, mkBar(t4h0+int64(i)*4*3600*1000, 4*3600*1000, v[0], v[1], v[2], v[3]))
	}
	return out
}

// pictureBars4HNoTarget: the same recent 101 resistance but NO old high — the
// opposing-zone refusal fixture.
func pictureBars4HNoTarget() []market.Kline {
	vals := [][4]float64{
		{98, 99, 97, 98}, {99, 100, 98, 99.5}, {99, 105, 96, 101},
		{96, 97, 94, 95}, {94, 95, 92, 93}, {93, 94, 91, 92},
	}
	out := make([]market.Kline, 0, len(vals))
	for i, v := range vals {
		out = append(out, mkBar(t4h0+int64(i)*4*3600*1000, 4*3600*1000, v[0], v[1], v[2], v[3]))
	}
	return out
}

func pictureBarsH1() []market.Kline {
	out := make([]market.Kline, 0, 42)
	for i := 0; i < 42; i++ {
		c := 99.0 + float64(i)*0.0625
		out = append(out, mkBar(t4h0+int64(i)*3600*1000, 3600*1000, c-0.2, c+0.4, c-0.4, c))
	}
	out[40] = mkBar(t4h0+40*3600*1000, 3600*1000, 100.8, 101.3, 100.5, 101.0) // prev
	out[41] = mkBar(t4h0+41*3600*1000, 3600*1000, 101.0, 102.0, 100.6, 101.5) // cur
	return out
}

func pictureBars5M(withSwing bool) []market.Kline {
	// 28 bars, last close 18:59:59.999Z, closes rise to ~101.5.
	base := t4h0 + 39*3600*1000 + 40*60*1000 // 16:40Z
	out := make([]market.Kline, 0, 28)
	for i := 0; i < 28; i++ {
		c := 99.3 + float64(i)*0.08
		lo := c - 0.6
		if withSwing && i == 22 {
			lo = 98.5 // the strict swing low (neighbors higher)
		}
		if i >= 23 {
			lo = c - 0.4
		}
		out = append(out, mkBar(base+int64(i)*5*60*1000, 5*60*1000, c, c+0.6, lo, c+0.03))
	}
	return out
}

// seedPictureTape installs the full ladder (4H with both pivots, H1 pair, 5m
// with the swing) and sets now = 0.5s into the next 5m interval.
func (env *pictureHtfTestEnv) seedPictureTape() {
	env.t.Helper()
	env.seed(pictureBars4H(), pictureBarsH1(), pictureBars5M(true))
	env.now = time.UnixMilli(t4h0 + 42*3600*1000 + 500) // 19:00:00.5Z
}

func (env *pictureHtfTestEnv) seed(bars4h, barsH1, bars5m []market.Kline) {
	env.t.Helper()
	market.FuturesBarsProvider = func(symbol, tf string, count int) []market.Kline {
		switch tf {
		case "4h":
			return tailOf(bars4h, count)
		case "1h":
			return tailOf(barsH1, count)
		case "5m":
			return tailOf(bars5m, count)
		}
		return nil
	}
}

func (env *pictureHtfTestEnv) fresh5m() {
	env.t.Helper()
	env.eval.OnBars("MNQ", "5m", tailOf(market.FuturesBarsProvider("MNQ", "5m", 28), 1), env.now.Add(-500*time.Millisecond))
}

func TestPictureHtfEvaluatorSubmitsOnceAndAdmits(t *testing.T) {
	env := newPictureHtfEnv(t, store.PictureHtfConfig{Enabled: true, MinRR: 2.5})
	env.seedPictureTape()
	// The production call site is the 5m bar EVENT: OnBars → evaluate → submit.
	env.eval.OnBars("MNQ", "5m", tailOf(market.FuturesBarsProvider("MNQ", "5m", 28), 1), env.now)
	if len(env.submits) != 1 {
		t.Fatalf("exactly one submission from the 5m event, got %d", len(env.submits))
	}
	// A duplicate evaluation (tick fallback, restart): the durable claim
	// blocks a second submission and reports it.
	res2 := env.eval.Evaluate("MNQ", env.now)
	if res2.Stage != "watching" || !strings.Contains(res2.Reason, "already claimed") {
		t.Fatalf("duplicate evaluation must report the existing claim, got %+v", res2)
	}
	if len(env.submits) != 1 {
		t.Fatalf("duplicate evaluation must not re-submit")
	}
	row, ok, _ := env.st.PictureHtfGet(env.submits[0])
	if !ok || row.Stage != "place_pending" {
		t.Fatalf("the opportunity must sit place_pending awaiting broker evidence: %+v", row)
	}
	if row.StopPx <= 0 || row.TargetPx <= 0 || row.RREstimate < row.RRConfigured {
		t.Fatalf("geometry must be persisted: %+v", row)
	}
	if row.Direction != "long" || row.LevelRole != "resistance" || row.LevelBodyTop != 101 {
		t.Fatalf("the broken level must be the 101 resistance, got %+v", row)
	}
	if row.TargetPx != 110 {
		t.Fatalf("the opposing target must be the old 110 high, got %+v", row)
	}
}

func TestPictureHtfEvaluatorLateFrameExpires(t *testing.T) {
	env := newPictureHtfEnv(t, store.PictureHtfConfig{Enabled: true, MinRR: 2.5})
	env.seedPictureTape()
	// The freshest 5m frame is 5s old — outside the 2s freshness limit.
	env.eval.OnBars("MNQ", "5m", tailOf(market.FuturesBarsProvider("MNQ", "5m", 28), 1), env.now.Add(-5*time.Second))
	res := env.eval.Evaluate("MNQ", env.now)
	if res.Stage != "expired" {
		t.Fatalf("a late frame must expire, got %+v", res)
	}
	if len(env.submits) != 0 {
		t.Fatalf("a late frame must never submit")
	}
}

func TestPictureHtfEvaluatorPastWindowExpires(t *testing.T) {
	env := newPictureHtfEnv(t, store.PictureHtfConfig{Enabled: true, MinRR: 2.5})
	env.seedPictureTape()
	// 30s into the interval — outside the 10s entry window.
	env.eval.OnBars("MNQ", "5m", tailOf(market.FuturesBarsProvider("MNQ", "5m", 28), 1), env.now)
	late := env.now.Add(30 * time.Second)
	res := env.eval.Evaluate("MNQ", late)
	if res.Stage != "expired" {
		t.Fatalf("past the entry window must expire, got %+v", res)
	}
}

func TestPictureHtfEvaluatorBeforeBoundaryWatches(t *testing.T) {
	env := newPictureHtfEnv(t, store.PictureHtfConfig{Enabled: true, MinRR: 2.5})
	env.seedPictureTape()
	early := env.now.Add(-1 * time.Second)
	env.eval.OnBars("MNQ", "5m", tailOf(market.FuturesBarsProvider("MNQ", "5m", 28), 1), early)
	if res := env.eval.Evaluate("MNQ", early); res.Stage != "watching" {
		t.Fatalf("before the boundary the mode watches, got %+v", res)
	}
	if len(env.submits) != 0 {
		t.Fatalf("no submission before the window")
	}
}

func TestPictureHtfEvaluatorNoSwingRefuses(t *testing.T) {
	env := newPictureHtfEnv(t, store.PictureHtfConfig{Enabled: true, MinRR: 2.5})
	env.seed(pictureBars4H(), pictureBarsH1(), pictureBars5M(false))
	env.now = time.UnixMilli(t4h0 + 42*3600*1000 + 500)
	env.fresh5m()
	res := env.eval.Evaluate("MNQ", env.now)
	if res.Stage != "refused" || res.Reason == "" {
		t.Fatalf("no swing must refuse with a reason, got %+v", res)
	}
	if len(env.submits) != 0 {
		t.Fatalf("a refused opportunity must never submit")
	}
}

func TestPictureHtfEvaluatorNoOpposingZoneRefuses(t *testing.T) {
	env := newPictureHtfEnv(t, store.PictureHtfConfig{Enabled: true, MinRR: 2.5})
	// Only the 101 pivot — no old high above the entry.
	env.seed(pictureBars4HNoTarget(), pictureBarsH1(), pictureBars5M(true))
	env.now = time.UnixMilli(t4h0 + 42*3600*1000 + 500)
	env.fresh5m()
	res := env.eval.Evaluate("MNQ", env.now)
	if res.Stage != "refused" {
		t.Fatalf("no opposing zone must refuse, got %+v", res)
	}
}

func TestPictureHtfEvaluatorLowRRRefuses(t *testing.T) {
	// The qualified tape offers ~2.6R; demanding 5R must refuse with the named
	// reason and NEVER skip the nearer zone to fix the number.
	env := newPictureHtfEnv(t, store.PictureHtfConfig{Enabled: true, MinRR: 5.0})
	env.seedPictureTape()
	env.fresh5m()
	res := env.eval.Evaluate("MNQ", env.now)
	if res.Stage != "refused" {
		t.Fatalf("R:R below the configured minimum must refuse, got %+v", res)
	}
	if len(env.submits) != 0 {
		t.Fatalf("a refused opportunity must never submit")
	}
}

func TestPictureHtfEvaluatorDisabledDoesNothing(t *testing.T) {
	env := newPictureHtfEnv(t, store.PictureHtfConfig{Enabled: false})
	env.seedPictureTape()
	if res := env.eval.Evaluate("MNQ", env.now); res.Stage != "watching" {
		t.Fatalf("disabled mode must watch, got %+v", res)
	}
	if len(env.submits) != 0 {
		t.Fatalf("disabled mode must never submit")
	}
}

func TestPictureHtfEvaluatorCapabilityGateBlocks(t *testing.T) {
	// The DEFAULT capability seam reads the concrete trader — resetTrader has
	// none, so capability is NOT proven and the mode must stay unavailable.
	at, _ := resetTrader(t, store.StrategyConfig{DayPlan: &store.DayPlanConfig{PictureHtf: &store.PictureHtfConfig{Enabled: true}}})
	ev := NewPictureHtfEvaluator(at, store.PictureHtfResolved(&store.PictureHtfConfig{Enabled: true}))
	res := ev.Evaluate("MNQ", time.Now())
	if res.Stage != "watching" || !strings.Contains(res.Reason, "mode unavailable") {
		t.Fatalf("an unproven AddOn must gate the mode off, got %+v", res)
	}
}

func TestPictureHtfEvaluatorIgnoresUnfinalizedBars(t *testing.T) {
	env := newPictureHtfEnv(t, store.PictureHtfConfig{Enabled: true, MinRR: 2.5})
	env.seedPictureTape()
	// Override the 5m ladder: the NEWEST bar is time-complete (CloseTime < now)
	// but the AddOn never finalized it — it must not count as a completed bar,
	// so the previous interval's window is past and the evaluation expires.
	bars5m := pictureBars5M(true)
	bars5m[len(bars5m)-1].Final = false
	env.seed(pictureBars4H(), pictureBarsH1(), bars5m)
	env.eval.OnBars("MNQ", "5m", tailOf(bars5m, 1), env.now)
	res := env.eval.Evaluate("MNQ", env.now)
	if res.Stage != "expired" {
		t.Fatalf("an unfinalized bar must not establish the current interval, got %+v", res)
	}
	if len(env.submits) != 0 {
		t.Fatalf("an unfinalized newest bar must never submit")
	}
}
