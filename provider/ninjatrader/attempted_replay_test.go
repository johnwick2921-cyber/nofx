package ninjatrader

import (
	"net"
	"strings"
	"testing"
	"time"
)

// FIX-DOUBLE-ENTRY — RED-first tests at the production flush path.
//
// The harness drives the REAL SendSignal → flushPendingReportFor path with a
// write hook that fails mid-flush (creating an `attempted` frame, exactly the
// production shape at tcp_server.go:2591) and then exercises the guard the
// flush must consult before re-sending it.

func newAttemptedHarness(t *testing.T) (*TCPServer, net.Conn, func(sig SignalPayload), func() net.Conn) {
	t.Helper()
	s := NewTCPServer(nil)
	s.attemptedVerifyWaitOverride = 200 * time.Millisecond
	srv, cli := net.Pipe()
	t.Cleanup(func() {
		_ = srv.Close()
		_ = cli.Close()
	})
	s.conn = srv
	// A failed flush calls closeConn (s.conn -> nil AND the socket is dead);
	// rearm builds a FRESH pipe pair (a reconnect) and returns the new peer.
	rearm := func() net.Conn {
		ns, nc := net.Pipe()
		t.Cleanup(func() {
			_ = ns.Close()
			_ = nc.Close()
		})
		s.conn = ns
		return nc
	}
	return s, cli, func(sig SignalPayload) { _ = s.SendSignal(sig) }, rearm
}

func failOnceHook(t *testing.T, real func(c net.Conn, sig SignalPayload) error) func(c net.Conn, sig SignalPayload) error {
	var once bool
	return func(c net.Conn, sig SignalPayload) error {
		if !once {
			once = true
			return &net.OpError{Op: "write", Err: errDeadPipe}
		}
		return real(c, sig)
	}
}

var errDeadPipe = &testErr{"dead pipe"}

type testErr struct{ s string }

func (e *testErr) Error() string { return e.s }

func sigNow(id string) string { return time.Now().UTC().Format(time.RFC3339Nano) }

func TestAttemptedEntryFoundAtBrokerIsNotResent(t *testing.T) {
	s, _, send, rearm := newAttemptedHarness(t)
	s.SetWriteFrameHookForTest(func(c net.Conn, sig SignalPayload) error { return errDeadPipe })
	send(SignalPayload{SignalID: "e1", Timestamp: sigNow("e1")}) // write fails → attempted, requeued
	rearm()
	if s.PendingSignalCount() != 1 || s.PendingAttemptedCount() != 1 {
		t.Fatalf("fixture: attempted frame expected in pending: %+v", s.pending)
	}
	rec := time.Now()
	s.SetReconnectAtForTest(rec)
	// A FRESH post-reconnect snapshot carrying the order named signal_id.
	s.orderSnaps.PutAt(OrderSnapshotPayload{Account: "Sim101", Orders: []NT8Order{{OrderID: "o1", Name: "e1", State: "working"}}}, rec.Add(10*time.Millisecond))
	wrote := false
	s.SetWriteFrameHookForTest(func(c net.Conn, sig SignalPayload) error { wrote = true; return nil })
	if err := s.flushPending(); err != nil {
		t.Fatal(err)
	}
	if wrote {
		t.Fatal("attempted entry found at the broker must NOT be resent")
	}
	if s.PendingSignalCount() != 0 {
		t.Fatalf("settled frame must leave the queue: %+v", s.pending)
	}
}

func TestAttemptedEntryAbsentFromFreshSnapshotResendsOnce(t *testing.T) {
	s, _, send, rearm := newAttemptedHarness(t)
	s.SetWriteFrameHookForTest(func(c net.Conn, sig SignalPayload) error { return errDeadPipe })
	send(SignalPayload{SignalID: "e2", Timestamp: sigNow("e2")})
	cli := rearm()
	if s.PendingSignalCount() != 1 || s.PendingAttemptedCount() != 1 {
		t.Fatalf("fixture: attempted frame expected in pending: %+v", s.pending)
	}
	rec := time.Now()
	s.SetReconnectAtForTest(rec)
	s.orderSnaps.PutAt(OrderSnapshotPayload{Account: "Sim101", Orders: []NT8Order{}}, rec.Add(10*time.Millisecond))
	s.SetWriteFrameHookForTest(nil)
	frames := make(chan Envelope, 1)
	go func() {
		env, err := ReadFrame(cli)
		if err == nil {
			frames <- env
		}
		close(frames)
	}()
	if err := s.flushPending(); err != nil {
		t.Fatal(err)
	}
	if s.PendingSignalCount() != 0 {
		t.Fatalf("resent frame must leave the queue: %+v", s.pending)
	}
	select {
	case env := <-frames:
		if env.Type != FrameSignal {
			t.Fatalf("frame type %v, want signal", env.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the absent-from-book attempted entry must be resent exactly once")
	}
}

func TestAttemptedEntryNoFreshSnapshotDropsAfterWait(t *testing.T) {
	s, _, send, rearm := newAttemptedHarness(t)
	s.SetWriteFrameHookForTest(func(c net.Conn, sig SignalPayload) error { return errDeadPipe })
	send(SignalPayload{SignalID: "e3", TraderID: "t3", Timestamp: sigNow("e3")})
	rearm()
	s.SetReconnectAtForTest(time.Now())
	s.SetWriteFrameHookForTest(nil)
	if err := s.flushPending(); err != nil {
		t.Fatal(err)
	}
	if s.PendingSignalCount() != 1 {
		t.Fatalf("within the wait the frame must be HELD, got %d", s.PendingSignalCount())
	}
	time.Sleep(400 * time.Millisecond) // past the 200ms wait → fail closed
	if err := s.flushPending(); err != nil {
		t.Fatal(err)
	}
	if s.PendingSignalCount() != 0 {
		t.Fatalf("after the wait the unverified frame must be DROPPED, got %d", s.PendingSignalCount())
	}
	if got := GateBlockCountForTest("t3", attemptedGateName); got != 1 {
		t.Fatalf("refusal must be counted (attempted_entry_unverified), got %d", got)
	}
}

func TestAttemptedEntryStaleSnapshotTreatedAsNoSnapshot(t *testing.T) {
	s, _, send, rearm := newAttemptedHarness(t)
	s.SetWriteFrameHookForTest(func(c net.Conn, sig SignalPayload) error { return errDeadPipe })
	send(SignalPayload{SignalID: "e4", Timestamp: sigNow("e4")})
	rearm()
	rec := time.Now()
	// A snapshot received BEFORE the reconnect proves nothing about it.
	s.orderSnaps.PutAt(OrderSnapshotPayload{Account: "Sim101", Orders: []NT8Order{{OrderID: "o4", Name: "e4", State: "working"}}}, rec.Add(-time.Minute))
	s.SetReconnectAtForTest(rec)
	s.SetWriteFrameHookForTest(nil)
	if err := s.flushPending(); err != nil {
		t.Fatal(err)
	}
	if s.PendingSignalCount() != 1 {
		t.Fatalf("a pre-reconnect snapshot must be treated as NO snapshot (hold), got %d", s.PendingSignalCount())
	}
}

func TestAttemptedEntryVariantsLimitAndStopEntry(t *testing.T) {
	for _, kind := range []string{"limit", "stop_entry"} {
		t.Run(kind, func(t *testing.T) {
			s, _, send, rearm := newAttemptedHarness(t)
			s.SetWriteFrameHookForTest(func(c net.Conn, sig SignalPayload) error { return errDeadPipe })
			send(SignalPayload{SignalID: "e5-" + kind, OrderType: kind, LimitPrice: 100, StopPrice: 101, Timestamp: sigNow("e5")})
			rearm()
			rec := time.Now()
			s.SetReconnectAtForTest(rec)
			// found at broker → not resent
			s.orderSnaps.PutAt(OrderSnapshotPayload{Account: "Sim101", Orders: []NT8Order{{OrderID: "o5", Name: "e5-" + kind, State: "working"}}}, rec.Add(10*time.Millisecond))
			wrote := false
			s.SetWriteFrameHookForTest(func(c net.Conn, sig SignalPayload) error { wrote = true; return nil })
			if err := s.flushPending(); err != nil {
				t.Fatal(err)
			}
			if wrote {
				t.Fatalf("%s attempted entry found at broker must NOT be resent", kind)
			}
			if s.PendingSignalCount() != 0 {
				t.Fatalf("%s settled frame must leave the queue", kind)
			}
		})
	}
}

func TestFreshEntryFlushesWithoutGuard(t *testing.T) {
	// A never-attempted entry keeps today's behaviour: no snapshot, no wait.
	s, cli, send, _ := newAttemptedHarness(t)
	s.SetReconnectAtForTest(time.Now())
	frames := make(chan Envelope, 1)
	go func() {
		env, err := ReadFrame(cli)
		if err == nil {
			frames <- env
		}
		close(frames)
	}()
	send(SignalPayload{SignalID: "e6", Timestamp: sigNow("e6")})
	if s.PendingSignalCount() != 0 {
		t.Fatalf("fresh frame should flush immediately, got %d queued", s.PendingSignalCount())
	}
	select {
	case env := <-frames:
		if env.Type != FrameSignal {
			t.Fatalf("fresh entry must flush unguarded: %v", env.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("fresh entry must flush unguarded (no frame)")
	}
}

func TestAttemptedGuardBootLineValues(t *testing.T) {
	// The boot line reads its values; the wait knob is parsed from env.
	t.Setenv(attemptedEntryVerifyWaitEnv, "7")
	if attemptedVerifyWait() != 7*time.Second {
		t.Fatalf("env knob must tune the wait: %v", attemptedVerifyWait())
	}
	t.Setenv(attemptedEntryVerifyWaitEnv, "bogus")
	if attemptedVerifyWait() != attemptedVerifyWaitDefault {
		t.Fatalf("bogus env must fall back to the default: %v", attemptedVerifyWait())
	}
	if !strings.Contains(attemptedGateName, "attempted_entry") {
		t.Fatal("gate name drift")
	}
}
