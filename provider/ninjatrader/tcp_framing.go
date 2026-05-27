// Package ninjatrader — TCP framing codec for Plan 1.5.
//
// Wire format: 4-byte big-endian length prefix followed by UTF-8 JSON payload.
// Max frame size: 1 MB. Oversized frames are an error (server closes the
// connection per spec L4376, L4416).
//
// Four message types per spec L4378-4410: signal | fill | heartbeat | ack.
// SEPARATE FILE from tcp_server.go per spec file-manifest L4360.
package ninjatrader

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// FrameType enumerates the 4 wire envelope types per spec L4382.
type FrameType string

const (
	FrameSignal    FrameType = "signal"
	FrameFill      FrameType = "fill"
	FrameHeartbeat FrameType = "heartbeat"
	FrameAck       FrameType = "ack"
)

// Envelope wraps every frame on the TCP stream per spec L4381-4385.
type Envelope struct {
	Type    FrameType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SignalPayload is the Go-server → C#-AddOn signal frame per spec L4387-4396.
type SignalPayload struct {
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`        // "long" | "short"
	Quantity   int     `json:"quantity"`
	Entry      float64 `json:"entry"`       // tick-rounded
	StopLoss   float64 `json:"stop_loss"`
	TakeProfit float64 `json:"take_profit"`
	SignalID   string  `json:"signal_id"`   // UUID
	Timestamp  string  `json:"timestamp"`   // RFC3339
}

// FillPayload is the C#-AddOn → Go-server fill frame per spec L4398-4406.
type FillPayload struct {
	SignalID      string  `json:"signal_id"`
	FillPrice     float64 `json:"fill_price"`
	FillTime      string  `json:"fill_time"` // RFC3339
	Side          string  `json:"side"`
	Quantity      int     `json:"quantity"`
	SlippageTicks float64 `json:"slippage_ticks"`
	Status        string  `json:"status"` // "filled" | "rejected" | "partial"
}

// AckPayload acknowledges a heartbeat or a specific signal_id per spec L4410.
type AckPayload struct {
	Acks string `json:"acks"` // "heartbeat" or "<signal_id>"
}

// HeartbeatPayload is an empty struct — spec L4408 says empty payload.
type HeartbeatPayload struct{}

// ErrFrameTooLarge signals the peer sent a length header > TCPMaxFrameBytes.
// Per spec L4416, the server logs and closes the connection on this error.
var ErrFrameTooLarge = errors.New("tcp_framing: frame exceeds max size")

// WriteFrame writes a single length-prefixed JSON frame to w.
// Returns an error if the marshalled body exceeds TCPMaxFrameBytes.
func WriteFrame(w io.Writer, frameType FrameType, payload any) error {
	body, err := marshalEnvelope(frameType, payload)
	if err != nil {
		return err
	}
	if len(body) > TCPMaxFrameBytes {
		return fmt.Errorf("tcp_framing: frame too large (%d > %d)", len(body), TCPMaxFrameBytes)
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(body)))
	if _, err := w.Write(hdr[:]); err != nil {
		return fmt.Errorf("tcp_framing: write header: %w", err)
	}
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("tcp_framing: write body: %w", err)
	}
	return nil
}

// ReadFrame reads a single length-prefixed JSON frame from r.
// Returns ErrFrameTooLarge if the length header exceeds TCPMaxFrameBytes
// (server closes the connection per spec L4416).
func ReadFrame(r io.Reader) (Envelope, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Envelope{}, err
	}
	length := binary.BigEndian.Uint32(hdr[:])
	if length > TCPMaxFrameBytes {
		return Envelope{}, ErrFrameTooLarge
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return Envelope{}, fmt.Errorf("tcp_framing: read body: %w", err)
	}
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return Envelope{}, fmt.Errorf("tcp_framing: bad JSON: %w", err)
	}
	return env, nil
}

func marshalEnvelope(frameType FrameType, payload any) ([]byte, error) {
	var raw json.RawMessage
	if payload == nil {
		raw = json.RawMessage("{}")
	} else {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("tcp_framing: marshal payload: %w", err)
		}
		raw = b
	}
	body, err := json.Marshal(Envelope{Type: frameType, Payload: raw})
	if err != nil {
		return nil, fmt.Errorf("tcp_framing: marshal envelope: %w", err)
	}
	return body, nil
}
