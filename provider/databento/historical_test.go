package databento

// CRITICAL: Fixtures in this file MUST match the actual Databento
// API response shape, NOT a fabricated approximation. A prior
// fabricated fixture allowed a struct-shape bug to ship to Plan 1
// SHIPPED status (cf. smoke-test 2026-05-25). Verify any fixture
// updates against a live curl response before committing.
//
// Reference shape (ohlcv-1m, GLBX.MDP3, NQ.c.0):
//   - hd.ts_event: nanoseconds since epoch as string
//   - hd.rtype: 33 for ohlcv-1m
//   - open/high/low/close: price × 1e9 as string
//   - volume: integer as string

import (
	"testing"
	"time"
)

// realDatabentoResponseLine is a real OHLCV-1m record captured from a live
// GLBX.MDP3 NQ.c.0 curl on 2026-05-25 (smoke-test diagnosis window).
const realDatabentoResponseLine = `{"hd":{"ts_event":"1779387780000000000","rtype":33,"publisher_id":1,"instrument_id":42004058},"open":"29456250000000","high":"29458750000000","low":"29445500000000","close":"29454750000000","volume":"581"}`

func TestParseOHLCVResponse_TwoBars(t *testing.T) {
	// Two real-shape OHLCV-1m records (second is the first with a
	// 60-second-later ts_event + slightly different prices).
	body := []byte(realDatabentoResponseLine + "\n" +
		`{"hd":{"ts_event":"1779387840000000000","rtype":33,"publisher_id":1,"instrument_id":42004058},"open":"29454250000000","high":"29473250000000","low":"29453750000000","close":"29473000000000","volume":"713"}` + "\n")

	bars, err := parseOHLCVResponse(body)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("want 2 bars, got %d", len(bars))
	}

	want0 := Bar{
		Timestamp: time.Unix(0, 1779387780000000000).UTC(),
		Open:      29456.25,
		High:      29458.75,
		Low:       29445.50,
		Close:     29454.75,
		Volume:    581,
	}
	if bars[0] != want0 {
		t.Errorf("bar[0] = %+v, want %+v", bars[0], want0)
	}

	if bars[1].Close != 29473.00 {
		t.Errorf("bar[1].Close = %v, want 29473.00", bars[1].Close)
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
	// Real-shape record but with a non-numeric ts_event — should fail parsing.
	body := []byte(`{"hd":{"ts_event":"abc","rtype":33,"publisher_id":1,"instrument_id":42004058},"open":"1","high":"1","low":"1","close":"1","volume":"1"}` + "\n")
	_, err := parseOHLCVResponse(body)
	if err == nil {
		t.Fatal("want error on malformed ts_event, got nil")
	}
}

func TestParseOHLCVResponse_UnexpectedRtype(t *testing.T) {
	// Real-shape record but with the wrong rtype — the sanity check in
	// toBar() should reject it so we never silently parse a non-OHLCV-1m
	// record as a bar.
	body := []byte(`{"hd":{"ts_event":"1779387780000000000","rtype":1,"publisher_id":1,"instrument_id":42004058},"open":"29456250000000","high":"29458750000000","low":"29445500000000","close":"29454750000000","volume":"581"}` + "\n")
	_, err := parseOHLCVResponse(body)
	if err == nil {
		t.Fatal("want error on unexpected rtype, got nil")
	}
}
