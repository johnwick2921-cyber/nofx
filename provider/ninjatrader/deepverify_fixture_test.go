// deep-verify-22 provider fixture — WriteFrame wire bytes for modify_bracket.
package ninjatrader

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

// G2 2.4 — SendModifyBracket's frame is written by WriteFrame (pure): 4-byte
// big-endian length + JSON envelope with type modify_bracket and the payload.
func TestDeepG2ModifyBracketFrameBytes(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteFrame(&buf, FrameModifyBracket, ModifyBracketPayload{
		TraderID: "tr1", Account: "Sim101", SignalID: "sig-9", NewStopLoss: 95.5, NewTakeProfit: 110.25,
	}); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	if len(raw) < 5 {
		t.Fatalf("frame too short: %d", len(raw))
	}
	n := binary.BigEndian.Uint32(raw[:4])
	if int(n) != len(raw)-4 {
		t.Fatalf("length prefix %d != body %d", n, len(raw)-4)
	}
	var env map[string]json.RawMessage
	if err := json.Unmarshal(raw[4:], &env); err != nil {
		t.Fatal(err)
	}
	var typ string
	_ = json.Unmarshal(env["type"], &typ)
	if typ != string(FrameModifyBracket) {
		t.Fatalf("frame type = %q", typ)
	}
	var p ModifyBracketPayload
	_ = json.Unmarshal(env["payload"], &p)
	if p.SignalID != "sig-9" || p.NewStopLoss != 95.5 || p.NewTakeProfit != 110.25 {
		t.Fatalf("payload mismatch: %+v", p)
	}
}
