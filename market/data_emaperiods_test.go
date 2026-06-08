package market

import (
	"reflect"
	"testing"
)

func synthKlines(n int) []Kline {
	ks := make([]Kline, n)
	for i := 0; i < n; i++ {
		// a gently varying price so EMAs differ across periods
		c := 100.0 + float64(i)*0.5 + float64(i%5)*0.3
		ks[i] = Kline{OpenTime: int64(i) * 60000, Open: c, High: c + 1, Low: c - 1, Close: c, Volume: 10}
	}
	return ks
}

// TestCalculateTimeframeSeries_EMAPeriods proves the configured-period path is a
// byte-identical superset of the legacy fixed path: with periods [20,50] the new
// EMAByPeriod entries equal the legacy EMA20Values/EMA50Values exactly (same calc,
// same guards), and a custom [9,21,200] populates those keys. nil → no map.
func TestCalculateTimeframeSeries_EMAPeriods(t *testing.T) {
	klines := synthKlines(60)

	// [20,50] → EMAByPeriod must equal the legacy fixed series, value-for-value.
	s := calculateTimeframeSeries(klines, "5m", 20, []int{20, 50})
	if !reflect.DeepEqual(s.EMAByPeriod[20], s.EMA20Values) {
		t.Fatalf("EMAByPeriod[20] must equal EMA20Values (identical calc)\n got: %v\nwant: %v", s.EMAByPeriod[20], s.EMA20Values)
	}
	if !reflect.DeepEqual(s.EMAByPeriod[50], s.EMA50Values) {
		t.Fatalf("EMAByPeriod[50] must equal EMA50Values\n got: %v\nwant: %v", s.EMAByPeriod[50], s.EMA50Values)
	}

	// [9,21,200] → all three keys present; shorter period has >= as many points.
	s3 := calculateTimeframeSeries(klines, "5m", 20, []int{9, 21, 200})
	for _, p := range []int{9, 21, 200} {
		if _, ok := s3.EMAByPeriod[p]; !ok {
			t.Fatalf("EMAByPeriod missing configured period %d", p)
		}
	}
	if len(s3.EMAByPeriod[9]) < len(s3.EMAByPeriod[200]) {
		t.Fatalf("EMA9 should have >= values than EMA200; got %d vs %d", len(s3.EMAByPeriod[9]), len(s3.EMAByPeriod[200]))
	}
	// The legacy fixed fields are STILL computed (back-compat readers depend on them).
	if len(s3.EMA20Values) == 0 {
		t.Fatalf("legacy EMA20Values must still be computed for back-compat readers")
	}

	// nil periods → no EMAByPeriod map (legacy output path, byte-identical).
	s0 := calculateTimeframeSeries(klines, "5m", 20, nil)
	if s0.EMAByPeriod != nil {
		t.Fatalf("nil periods must leave EMAByPeriod nil; got %v", s0.EMAByPeriod)
	}
}
