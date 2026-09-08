package ninjatrader

import (
	"encoding/json"
	"fmt"
	"nofx/researchsnapshot"
	"time"
)

// One state per connection, accessed only by the research worker. Metadata
// belongs to the frames preceding the observation, never a later live cache.
type researchWire struct {
	contract, build string
	previous        map[string]Bar
}

func newResearchWire() *researchWire { return &researchWire{previous: map[string]Bar{}} }
func (w *researchWire) observe(kind FrameType, payload json.RawMessage) {
	w.observeAt(kind, payload, time.Now())
}
func (w *researchWire) observeAt(kind FrameType, payload json.RawMessage, received time.Time) {
	switch kind {
	case FrameHello, FrameSubscribed, FrameBarsHistorical, FrameBarUpdate, FrameFill, FrameOrderUpdate, FrameOrderSnapshot, FramePositionClose:
	default:
		return
	}
	researchsnapshot.Record("wire:"+string(kind), func() []researchsnapshot.Fact { return w.facts(kind, payload, received) })
}
func (w *researchWire) facts(kind FrameType, payload json.RawMessage, received time.Time) []researchsnapshot.Fact {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(payload, &m); err != nil {
		panic("research wire JSON decode")
	}
	text := func(key string) string { var v string; _ = json.Unmarshal(m[key], &v); return v }
	if kind == FrameHello {
		w.build = text("build_id")
		return nil
	}
	if kind == FrameSubscribed {
		if text("symbol") == "MNQ" {
			w.contract = text("resolved_contract")
		}
		return nil
	}
	if kind == FrameBarsHistorical || kind == FrameBarUpdate {
		if text("symbol") != "MNQ" {
			return nil
		}
		var bars []Bar
		if err := json.Unmarshal(m["bars"], &bars); err != nil {
			panic("research bars decode")
		}
		tf := text("timeframe")
		opened := OpenStampBars(bars, tf)
		out := make([]researchsnapshot.Fact, 0, len(bars))
		for i, b := range bars {
			f := researchsnapshot.NewFact("market", string(kind), nil, researchsnapshot.Clocks{ObservationMS: researchsnapshot.Value(b.T), ReceiptMS: researchsnapshot.Value(received.UnixMilli())})
			f.Set("root_symbol", "MNQ")
			f.Set("feed", "NT8 TCP")
			f.Set("source_timezone", "UTC epoch milliseconds")
			f.Set("timeframe", tf)
			f.Set("source_stamp_ms", b.T)
			f.Set("bar_open_ms", opened[i].T)
			f.Set("bar_close_ms", b.T)
			f.Set("open", b.O)
			f.Set("high", b.H)
			f.Set("low", b.L)
			f.Set("close", b.C)
			f.Set("volume", b.V)
			f.Set("finalized", b.T <= received.UnixMilli())
			f.Set("forming", b.T > received.UnixMilli())
			if w.contract != "" {
				f.Set("contract", w.contract)
				f.Set("contract_basis", "preceding received subscription acknowledgement; historical per-bar contract UNKNOWN")
			}
			if w.build != "" {
				f.Set("source_build_id", w.build)
			}
			f.Unknown("price_scale", "wire has no merge/back-adjust policy; cannot certify historical contract prices")
			key := fmt.Sprintf("%s:%d", tf, b.T)
			if prev, ok := w.previous[key]; ok {
				f.Set("previous_observation", prev)
				f.Set("correction", prev != b)
			} else {
				f.Unknown("correction", "no prior captured version in this connection's bounded comparison window")
			}
			if len(w.previous) >= 12000 {
				w.previous = map[string]Bar{}
			}
			w.previous[key] = b
			out = append(out, f)
		}
		return out
	}
	// Select only non-account fields. Never archive credentials or account names.
	f := researchsnapshot.NewFact("exec", string(kind), nil, researchsnapshot.Clocks{ReceiptMS: researchsnapshot.Value(received.UnixMilli())})
	var emitted int64
	_ = json.Unmarshal(m["emitted_at_ms"], &emitted)
	if emitted > 0 {
		f.Clocks.ObservationMS = researchsnapshot.Value(emitted)
		delete(f.Missing, "observation_ms")
	}
	if stamp := text("fill_time"); stamp != "" {
		if t, e := time.Parse(time.RFC3339Nano, stamp); e == nil {
			f.Clocks.ObservationMS = researchsnapshot.Value(t.UnixMilli())
			delete(f.Missing, "observation_ms")
		}
	}
	selected := map[string]json.RawMessage{}
	for _, key := range []string{"signal_id", "symbol", "order_name", "state", "fill_price", "fill_time", "side", "quantity", "slippage_ticks", "status", "seq", "orders", "build_id", "emitted_at_ms", "reason", "exit_price", "exit_time", "position_id"} {
		if v, ok := m[key]; ok {
			selected[key] = v
		}
	}
	f.Set("broker_frame", selected)
	for dest, src := range map[string]string{"root_symbol": "symbol", "signal_id": "signal_id", "size": "quantity", "reason": "reason", "source_build_id": "build_id", "position_id": "position_id", "exit_price": "exit_price"} {
		if v, ok := m[src]; ok {
			f.Set(dest, v)
		}
	}
	if kind == FrameFill {
		f.Set("fills", selected)
		if text("status") == "filled" || text("status") == "partial" {
			if v, ok := m["fill_price"]; ok {
				f.Set("attainable_entry", v)
				f.Set("entry_basis", "received fill; entry/exit role requires signal linkage")
			}
		}
	}
	if text("reason") == "" {
		f.Unknown("reason", "received frame omits reason; h1 does not transmit rejection check")
	}
	f.Unknown("costs", "no broker commission or cost field in this received frame")
	return []researchsnapshot.Fact{f}
}
