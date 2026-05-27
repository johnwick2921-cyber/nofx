package ninjatrader

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRoundTrip_Signal(t *testing.T) {
	want := SignalPayload{
		Symbol:     "MNQ",
		Side:       "long",
		Quantity:   1,
		Entry:      21500.25,
		StopLoss:   21485.00,
		TakeProfit: 21525.00,
		SignalID:   "test-uuid-signal",
		Timestamp:  "2026-05-26T12:00:00Z",
	}
	var buf bytes.Buffer
	if err := WriteFrame(&buf, FrameSignal, want); err != nil {
		t.Fatalf("WriteFrame: %v", err)
	}
	env, err := ReadFrame(&buf)
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	if env.Type != FrameSignal {
		t.Errorf("type: want %q got %q", FrameSignal, env.Type)
	}
	var got SignalPayload
	if err := jsonUnmarshalForTest(env.Payload, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if got != want {
		t.Errorf("signal mismatch:\n got=%+v\nwant=%+v", got, want)
	}
}

func TestRoundTrip_Fill(t *testing.T) {
	want := FillPayload{
		SignalID:      "test-uuid-fill",
		FillPrice:     21500.50,
		FillTime:      "2026-05-26T12:00:01Z",
		Side:          "long",
		Quantity:     1,
		SlippageTicks: 1.0,
		Status:        "filled",
	}
	var buf bytes.Buffer
	if err := WriteFrame(&buf, FrameFill, want); err != nil {
		t.Fatalf("WriteFrame: %v", err)
	}
	env, err := ReadFrame(&buf)
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	if env.Type != FrameFill {
		t.Errorf("type: want %q got %q", FrameFill, env.Type)
	}
	var got FillPayload
	if err := jsonUnmarshalForTest(env.Payload, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if got != want {
		t.Errorf("fill mismatch:\n got=%+v\nwant=%+v", got, want)
	}
}

func TestRoundTrip_Heartbeat(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteFrame(&buf, FrameHeartbeat, HeartbeatPayload{}); err != nil {
		t.Fatalf("WriteFrame: %v", err)
	}
	env, err := ReadFrame(&buf)
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	if env.Type != FrameHeartbeat {
		t.Errorf("type: want %q got %q", FrameHeartbeat, env.Type)
	}
	// Empty payload is an empty JSON object per the encoder.
	if string(env.Payload) != "{}" {
		t.Errorf("payload: want \"{}\" got %q", string(env.Payload))
	}
}

func TestRoundTrip_Ack(t *testing.T) {
	cases := []AckPayload{
		{Acks: "heartbeat"},
		{Acks: "abc-123-uuid"},
	}
	for _, want := range cases {
		var buf bytes.Buffer
		if err := WriteFrame(&buf, FrameAck, want); err != nil {
			t.Fatalf("WriteFrame %q: %v", want.Acks, err)
		}
		env, err := ReadFrame(&buf)
		if err != nil {
			t.Fatalf("ReadFrame %q: %v", want.Acks, err)
		}
		if env.Type != FrameAck {
			t.Errorf("type: want %q got %q", FrameAck, env.Type)
		}
		var got AckPayload
		if err := jsonUnmarshalForTest(env.Payload, &got); err != nil {
			t.Fatalf("unmarshal %q: %v", want.Acks, err)
		}
		if got != want {
			t.Errorf("ack mismatch: got=%+v want=%+v", got, want)
		}
	}
}

func TestReadFrame_OversizedRejected(t *testing.T) {
	var buf bytes.Buffer
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], TCPMaxFrameBytes+1)
	buf.Write(hdr[:])
	// No body needed — length-prefix check fires first.
	_, err := ReadFrame(&buf)
	if !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("want ErrFrameTooLarge, got %v", err)
	}
}

func TestWriteFrame_OversizedRejected(t *testing.T) {
	// Payload string > 1 MB after marshalling.
	big := strings.Repeat("x", TCPMaxFrameBytes+10)
	var buf bytes.Buffer
	err := WriteFrame(&buf, FrameSignal, map[string]string{"big": big})
	if err == nil {
		t.Fatal("want error for oversized frame, got nil")
	}
	if !strings.Contains(err.Error(), "frame too large") {
		t.Errorf("want 'frame too large' error, got %v", err)
	}
}

func TestReadFrame_MalformedJSON(t *testing.T) {
	var buf bytes.Buffer
	body := []byte("not json at all {{{")
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(body)))
	buf.Write(hdr[:])
	buf.Write(body)
	_, err := ReadFrame(&buf)
	if err == nil {
		t.Fatal("want JSON parse error, got nil")
	}
	if !strings.Contains(err.Error(), "bad JSON") {
		t.Errorf("want 'bad JSON' error, got %v", err)
	}
}

func TestReadFrame_TruncatedHeader(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0x00, 0x00}) // only 2 of 4 header bytes
	_, err := ReadFrame(buf)
	if err == nil {
		t.Fatal("want EOF/unexpected-EOF, got nil")
	}
	if !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		t.Errorf("want io.EOF or io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestReadFrame_TruncatedBody(t *testing.T) {
	var buf bytes.Buffer
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], 100) // says 100 bytes follow
	buf.Write(hdr[:])
	buf.Write([]byte("short")) // only 5 bytes follow
	_, err := ReadFrame(&buf)
	if err == nil {
		t.Fatal("want truncated-body error, got nil")
	}
}

// jsonUnmarshalForTest wraps encoding/json.Unmarshal for assertion clarity.
func jsonUnmarshalForTest(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
