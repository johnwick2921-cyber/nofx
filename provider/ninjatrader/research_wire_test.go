package ninjatrader

import (
	"context"
	"encoding/json"
	"math"
	"net"
	"nofx/researchsnapshot"
	"path/filepath"
	"testing"
	"time"
)

func TestStageAWireProductionReadLoop(t *testing.T) {
	a, err := researchsnapshot.Open(filepath.Join(t.TempDir(), "research.db"), "wire-fixture")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	r := researchsnapshot.NewRecorder(a, 32, func(string) {})
	researchsnapshot.Install(r)
	defer func() { researchsnapshot.Install(nil); r.Close() }()
	server := NewTCPServer(nil)
	left, right := net.Pipe()
	server.conn = left
	server.wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go server.readLoop(ctx, left)
	frames := []struct {
		kind    FrameType
		payload any
	}{
		{FrameBarUpdate, map[string]any{"symbol": "MNQ", "timeframe": "5m", "bars": []any{map[string]any{"t": 1788881400000, "o": 100, "h": 101, "l": 99, "v": 0}}}},
		{FrameOrderSnapshot, map[string]any{"build_id": "2026-09-07-h1", "emitted_at_ms": 1788881400123, "orders": []any{}}},
	}
	for _, f := range frames {
		if err := WriteFrame(right, f.kind, f.payload); err != nil {
			t.Fatal(err)
		}
	}
	right.Close()
	server.wg.Wait()
	wait, c := context.WithTimeout(context.Background(), time.Second)
	defer c()
	if err = r.Flush(wait); err != nil {
		t.Fatal(err)
	}
	data, err := a.Export(wait, 0, math.MaxInt64)
	if err != nil {
		t.Fatal(err)
	}
	var bundle researchsnapshot.Bundle
	_ = json.Unmarshal(data, &bundle)
	if len(bundle.Objects["market"]) != 1 || len(bundle.Objects["exec"]) != 1 {
		t.Fatalf("received frames never reached archive: counts=%v dropped=%d", bundle.Manifest.Counts, r.Dropped())
	}
	var fields map[string]any
	_ = json.Unmarshal(bundle.Objects["market"][0].Fields, &fields)
	if fields["close"] != nil || fields["volume"] != float64(0) {
		t.Fatalf("wire absence became fabricated value: close=%v volume=%v", fields["close"], fields["volume"])
	}
	if bundle.Objects["market"][0].Clocks.ObservationMS == nil || *bundle.Objects["market"][0].Clocks.ObservationMS != 1788881400000 {
		t.Fatal("source observation stamp discarded")
	}
}

func TestStageABackfillAvailableOnlyAtReceipt(t *testing.T) {
	w := newResearchWire()
	now := time.Date(2026, 9, 8, 16, 0, 0, 0, time.UTC)
	facts := w.facts(FrameBarsHistorical, json.RawMessage(`{"symbol":"MNQ","timeframe":"5m","bars":[{"t":1788881400000,"o":100,"h":101,"l":99,"c":100,"v":0}]}`), now)
	if len(facts) != 1 || *facts[0].Clocks.ReceiptMS != now.UnixMilli() || *facts[0].Clocks.ObservationMS != 1788881400000 {
		t.Fatalf("four-clock capture wrong: %+v", facts)
	}
	if facts[0].Fields["price_scale"] != nil {
		t.Fatal("historical adjustment policy invented")
	}
}
