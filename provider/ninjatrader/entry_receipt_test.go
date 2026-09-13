package ninjatrader

import (
	"sync"
	"testing"
)

func TestEntryReceiptScopeAndMonotoneQuantity(t *testing.T) {
	s := NewTCPServer(nil)
	for _, in := range []struct {
		symbol, account, signal string
		qty                     int
	}{{"", "Sim101", "entry", 1}, {"MNQ", "", "entry", 1}, {"MNQ", "Sim101", "", 1}, {"MNQ", "Sim101", "entry", 0}, {"MNQ", "Sim101", "entry", -1}} {
		s.NoteEntryExecution(in.symbol, in.account, in.signal, in.qty)
	}
	_, _, r, _ := s.PositionsForExecutionReceipt("Sim101", "MNQ")
	if !r.IsZero() {
		t.Fatal("invalid input created execution receipt")
	}
	s.NoteEntryExecution("MNQ", "Sim101", "entry", 2)
	_, _, original, _ := s.PositionsForExecutionReceipt("Sim101", "MNQ")
	if original.IsZero() {
		t.Fatal("entry receipt missing")
	}
	for _, scope := range [][2]string{{"SimOther", "MNQ"}, {"Sim101", "MES"}} {
		_, _, r, _ := s.PositionsForExecutionReceipt(scope[0], scope[1])
		if !r.IsZero() {
			t.Fatal("entry receipt crossed account/symbol boundary")
		}
	}
	s.SeedPositionsForTest("Sim101", []OpenPosition{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.NoteEntryExecution("mnq", "Sim101", "entry", 1)
			s.NoteEntryExecution("MNQ", "Sim101", "entry", 2)
			s.PositionsForExecutionReceipt("Sim101", "MNQ")
		}()
	}
	wg.Wait()
	ps, received, after, ok := s.PositionsForExecutionReceipt("Sim101", "MNQ")
	if !ok || ps == nil || len(ps) != 0 || !after.Equal(original) || !received.After(after) {
		t.Fatalf("replay invalidated newer known snapshot: %+v %v %v %v", ps, received, after, ok)
	}
	s.NoteEntryExecution("MNQ", "Sim101", "entry", 3)
	_, _, grown, _ := s.PositionsForExecutionReceipt("Sim101", "MNQ")
	if !grown.After(after) {
		t.Fatal("new cumulative execution did not advance receipt")
	}
	fresh := NewTCPServer(nil)
	if _, _, r, ok := fresh.PositionsForExecutionReceipt("Sim101", "MNQ"); ok || !r.IsZero() {
		t.Fatal("new server fabricated snapshot or execution receipt")
	}
}
