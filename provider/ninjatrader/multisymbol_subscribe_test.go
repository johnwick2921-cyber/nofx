package ninjatrader

import (
	"reflect"
	"testing"
)

// TestBarsSubscribePayloads_PrimaryOnly_ByteIdentical is the P5.1 golden: with
// NO extra symbols registered, the auto-subscribe payload list is exactly ONE
// frame whose content equals the pre-P5 single payload (currentBarsSubscribe).
// The single-symbol [MNQ] deployment therefore behaves byte-identically.
func TestBarsSubscribePayloads_PrimaryOnly_ByteIdentical(t *testing.T) {
	s := NewTCPServer(nil)
	s.SetBarsSubscribeSymbol("MNQ")

	got := s.barsSubscribePayloads()
	if len(got) != 1 {
		t.Fatalf("no extras must yield exactly 1 subscribe frame; got %d: %+v", len(got), got)
	}
	want := s.currentBarsSubscribe() // the pre-P5 payload shape
	if !reflect.DeepEqual(got[0], want) {
		t.Fatalf("primary frame must equal the legacy payload\n got:  %+v\n want: %+v", got[0], want)
	}
	if got[0].Symbol != "MNQ" || len(got[0].Timeframes) == 0 || got[0].BarsBack <= 0 {
		t.Fatalf("primary frame malformed: %+v", got[0])
	}
}

// TestAddBarsSubscribeSymbols locks the extras semantics: blanks, duplicates
// (case-insensitive), and the primary itself are skipped; each extra clones the
// primary's timeframes + bars-back with only the symbol swapped; the primary is
// always FIRST (the trading feed subscribes before any extra).
func TestAddBarsSubscribeSymbols(t *testing.T) {
	s := NewTCPServer(nil)
	s.SetBarsSubscribeSymbol("MNQ")
	s.AddBarsSubscribeSymbols("ES", " es ", "", "mnq", "NQ", "ES")

	got := s.barsSubscribePayloads()
	if len(got) != 3 {
		t.Fatalf("want primary+ES+NQ (3 frames); got %d: %+v", len(got), got)
	}
	if got[0].Symbol != "MNQ" || got[1].Symbol != "ES" || got[2].Symbol != "NQ" {
		t.Fatalf("order must be primary first then extras in add order; got %s,%s,%s",
			got[0].Symbol, got[1].Symbol, got[2].Symbol)
	}
	for i := 1; i < len(got); i++ {
		if !reflect.DeepEqual(got[i].Timeframes, got[0].Timeframes) || got[i].BarsBack != got[0].BarsBack {
			t.Fatalf("extra %s must clone the primary's timeframes+bars_back; got %+v vs primary %+v",
				got[i].Symbol, got[i], got[0])
		}
	}
}

// TestUnsubscribeBarsSymbol locks the teardown semantics: the PRIMARY symbol is
// refused (never tear down the live trading feed); an extra is removed from the
// auto-subscribe list even when disconnected (so a reconnect won't re-subscribe
// it), with the send error reported.
func TestUnsubscribeBarsSymbol(t *testing.T) {
	s := NewTCPServer(nil)
	s.SetBarsSubscribeSymbol("MNQ")
	s.AddBarsSubscribeSymbols("ES")

	// Primary refused.
	if err := s.UnsubscribeBarsSymbol("mnq"); err == nil {
		t.Fatal("unsubscribing the PRIMARY symbol must be refused")
	}
	if got := s.barsSubscribePayloads(); len(got) != 2 {
		t.Fatalf("refused unsubscribe must not mutate state; got %d frames", len(got))
	}

	// Extra removed from the list even though not connected (send fails).
	if err := s.UnsubscribeBarsSymbol("ES"); err == nil {
		t.Fatal("expected a not-connected send error (state still updated)")
	}
	got := s.barsSubscribePayloads()
	if len(got) != 1 || got[0].Symbol != "MNQ" {
		t.Fatalf("ES must be removed from the auto-subscribe list; got %+v", got)
	}
}
