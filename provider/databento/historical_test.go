package databento

import (
	"testing"
	"time"
)

func TestParseOHLCVResponse_TwoBars(t *testing.T) {
	// Sample of what Databento returns for schema=ohlcv-1m, encoding=json.
	// Each line is one bar. Numeric fields are integer fixed-point (1e9 divisor).
	body := []byte(`{"ts_event":"1746360000000000000","open":"21500250000000","high":"21515750000000","low":"21498000000000","close":"21510000000000","volume":"4321"}
{"ts_event":"1746360060000000000","open":"21510000000000","high":"21525500000000","low":"21505000000000","close":"21522750000000","volume":"5102"}
`)

	bars, err := parseOHLCVResponse(body)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("want 2 bars, got %d", len(bars))
	}

	want0 := Bar{
		Timestamp: time.Unix(0, 1746360000000000000).UTC(),
		Open:      21500.25,
		High:      21515.75,
		Low:       21498.00,
		Close:     21510.00,
		Volume:    4321,
	}
	if bars[0] != want0 {
		t.Errorf("bar[0] = %+v, want %+v", bars[0], want0)
	}

	if bars[1].Close != 21522.75 {
		t.Errorf("bar[1].Close = %v, want 21522.75", bars[1].Close)
	}
}

func TestParseOHLCVResponse_Empty(t *testing.T) {
	bars, err := parseOHLCVResponse([]byte(""))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(bars) != 0 {
		t.Errorf("want 0 bars, got %d", len(bars))
	}
}

func TestParseOHLCVResponse_MalformedLine(t *testing.T) {
	body := []byte(`{"ts_event":"abc","open":"1","high":"1","low":"1","close":"1","volume":"1"}` + "\n")
	_, err := parseOHLCVResponse(body)
	if err == nil {
		t.Fatal("want error on malformed ts_event, got nil")
	}
}
