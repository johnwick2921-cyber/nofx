// Package ninjatrader — TCP listener for Plan 1.5 pre-trigger build.
//
// Listens on 127.0.0.1:36974 (NOT NT's ATI port 36973). Accepts a single
// concurrent NT AddOn client; rejects (closes) any second simultaneous dial.
// Routes signal frames out, fill/heartbeat/ack frames in. Spec ref:
// docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md L4359 + L4374-4416.
//
// CSV bridge (Plan 1) remains the SIM-validated production path. This file
// is purely additive — Plan 1 critical files (csv_writer.go, csv_tailer.go,
// types.go, mock_nt.go) stay byte-identical per ADR-007.
package ninjatrader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"
)

// Wire-protocol constants per spec L4359 + L4408 + L4415 + L4414 + L4376.
const (
	// TCPListenAddr — loopback only; NT runs on the same Windows host (or
	// reaches WSL2 via mirrored networking). NOT 36973 (NT's ATI port).
	TCPListenAddr = "127.0.0.1:36974"

	// TCPHeartbeatInterval — server pings every 30s per spec L4408.
	TCPHeartbeatInterval = 30 * time.Second

	// TCPHeartbeatAckTimeout — server closes the connection after this long
	// without a heartbeat ack per spec L4415.
	TCPHeartbeatAckTimeout = 60 * time.Second

	// TCPStaleSignalAge — signals older than this are dropped silently from
	// the pending queue on reconnect per spec L4414 (C# AddOn may reject
	// stale; we proactively drop on the Go side to avoid wasted bandwidth).
	TCPStaleSignalAge = 60 * time.Second

	// TCPMaxFrameBytes — oversized frames are an error per spec L4376.
	TCPMaxFrameBytes = 1 << 20 // 1 MB

	// fillChannelBuffer — bounded so a stuck consumer can't OOM the server.
	fillChannelBuffer = 32

	// acceptLoopPollInterval — listener deadline so the accept loop can
	// observe ctx cancellation without blocking indefinitely on Accept().
	acceptLoopPollInterval = 250 * time.Millisecond
)

// TCPServer is the Go-side endpoint of the Plan 1.5 TCP bridge. It owns the
// listener, exactly one connected client at a time, the pending-signal queue
// flushed on (re)connect, and the inbound fill channel that TCPTrader reads.
type TCPServer struct {
	addr     string
	listener net.Listener

	// Pending signals to flush on (re)connect (spec L4414).
	pendingMu sync.Mutex
	pending   []timedSignal

	// Inbound fills — TCPTrader subscribes via Fills().
	fillCh chan FillPayload

	// Connection state — single concurrent client (spec L4359).
	connMu       sync.Mutex
	conn         net.Conn
	lastAckTime  time.Time
	staleOverride time.Duration // test hook: 0 means use TCPStaleSignalAge

	// writeMu serializes all writes to a connected client's net.Conn.
	// Both heartbeatLoop and readLoop (handling client heartbeats) write
	// to the same conn; without serialization, concurrent writes could
	// interleave frame bytes and corrupt the wire protocol. The C# side
	// already uses lock(writeLock) for the same reason. Added in
	// Plan 1.5.6 alongside the FrameHeartbeat ack-write deadline fix.
	writeMu sync.Mutex

	// Lifecycle.
	cancel context.CancelFunc
	wg     sync.WaitGroup
	logger *slog.Logger
}

// timedSignal pairs a signal payload with the wall-clock time SendSignal was
// called, so the flush path can drop stale-on-reconnect entries.
type timedSignal struct {
	payload   SignalPayload
	timestamp time.Time
}

// NewTCPServer constructs a server bound to TCPListenAddr. The listener is
// not opened until Start. Pass nil to use the default slog logger.
func NewTCPServer(logger *slog.Logger) *TCPServer {
	if logger == nil {
		logger = slog.Default()
	}
	return &TCPServer{
		addr:   TCPListenAddr,
		fillCh: make(chan FillPayload, fillChannelBuffer),
		logger: logger,
	}
}

// Start binds the listener and spins up the accept loop. The returned error
// is non-nil only if the bind fails. Stop releases the port.
func (s *TCPServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("tcp_server: listen %s: %w", s.addr, err)
	}
	s.listener = ln
	cctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.wg.Add(1)
	go s.acceptLoop(cctx)
	s.logger.Info("tcp_server: listening", "addr", s.addr)
	return nil
}

// Stop cancels the accept loop, closes the listener, drops any active
// connection, and waits for all goroutines to exit. Safe to call once.
func (s *TCPServer) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.closeConn()
	s.wg.Wait()
	return nil
}

// SendSignal queues a signal for delivery to the connected client. If no
// client is currently connected, the signal is buffered and flushed on the
// next successful accept (subject to TCPStaleSignalAge).
func (s *TCPServer) SendSignal(payload SignalPayload) error {
	s.pendingMu.Lock()
	s.pending = append(s.pending, timedSignal{payload: payload, timestamp: time.Now()})
	s.pendingMu.Unlock()
	return s.flushPending()
}

// Fills returns the inbound fill channel. TCPTrader subscribes here.
func (s *TCPServer) Fills() <-chan FillPayload { return s.fillCh }

// IsConnected reports whether a client is currently connected.
func (s *TCPServer) IsConnected() bool {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	return s.conn != nil
}

// Addr returns the actual listening address (useful for tests using :0).
func (s *TCPServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

// SetStaleSignalAgeForTest overrides the stale-signal cutoff for the
// TestTCPServer_StaleSignalDropped test. Production code never calls this.
func (s *TCPServer) SetStaleSignalAgeForTest(d time.Duration) {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	s.staleOverride = d
}

// SetAddrForTest overrides the bind address before Start (used by tests that
// need an ephemeral port via "127.0.0.1:0"). Production callers leave the
// constructor default in place.
func (s *TCPServer) SetAddrForTest(addr string) { s.addr = addr }

func (s *TCPServer) staleAge() time.Duration {
	s.connMu.Lock()
	override := s.staleOverride
	s.connMu.Unlock()
	if override > 0 {
		return override
	}
	return TCPStaleSignalAge
}

func (s *TCPServer) acceptLoop(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if t, ok := s.listener.(*net.TCPListener); ok {
			_ = t.SetDeadline(time.Now().Add(acceptLoopPollInterval))
		}
		c, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				continue
			}
			s.logger.Warn("tcp_server: accept error", "err", err)
			continue
		}

		// Single concurrent client (spec L4359). Any second simultaneous
		// dial is closed immediately.
		s.connMu.Lock()
		if s.conn != nil {
			s.logger.Warn("tcp_server: rejecting concurrent client", "addr", c.RemoteAddr())
			_ = c.Close()
			s.connMu.Unlock()
			continue
		}
		s.conn = c
		s.lastAckTime = time.Now()
		s.connMu.Unlock()

		s.logger.Info("tcp_server: client connected", "addr", c.RemoteAddr())
		s.wg.Add(2)
		go s.readLoop(ctx, c)
		go s.heartbeatLoop(ctx, c)
		// Flush any signals queued during the disconnect window.
		_ = s.flushPending()
	}
}

func (s *TCPServer) readLoop(ctx context.Context, c net.Conn) {
	defer s.wg.Done()
	defer s.closeConn()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// Modest read deadline so the loop notices ctx cancellation between frames.
		_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
		env, err := ReadFrame(c)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				continue
			}
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				s.logger.Info("tcp_server: client disconnected")
				return
			}
			if errors.Is(err, ErrFrameTooLarge) {
				s.logger.Warn("tcp_server: oversized frame, closing connection")
				return
			}
			s.logger.Warn("tcp_server: read frame error", "err", err)
			return
		}

		switch env.Type {
		case FrameFill:
			var fill FillPayload
			if err := json.Unmarshal(env.Payload, &fill); err != nil {
				s.logger.Warn("tcp_server: bad fill payload", "err", err)
				continue
			}
			select {
			case s.fillCh <- fill:
			default:
				s.logger.Warn("tcp_server: fill channel full, dropping", "signal_id", fill.SignalID)
			}

		case FrameAck:
			s.connMu.Lock()
			s.lastAckTime = time.Now()
			s.connMu.Unlock()

		case FrameHeartbeat:
			// Respond to peer heartbeat with an ack (spec L4410). Set our
			// own write deadline — Go's net.Conn deadlines are PERSISTENT
			// until reset, so without this the ack inherits whatever
			// deadline heartbeatLoop most recently set. Once that stale
			// deadline expires (5s after the last scheduled server
			// heartbeat), every subsequent write fails with i/o timeout
			// and the connection dies. (Plan 1.5.6, fixes NEW-1.)
			s.writeMu.Lock()
			_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			err := WriteFrame(c, FrameAck, AckPayload{Acks: "heartbeat"})
			s.writeMu.Unlock()
			if err != nil {
				s.logger.Warn("tcp_server: write heartbeat ack", "err", err)
				return
			}

		case FrameSignal:
			// Clients don't send signals; ignore but log.
			s.logger.Warn("tcp_server: unexpected signal frame from client")

		default:
			s.logger.Warn("tcp_server: unknown frame type", "type", env.Type)
		}
	}
}

func (s *TCPServer) heartbeatLoop(ctx context.Context, c net.Conn) {
	defer s.wg.Done()
	ticker := time.NewTicker(TCPHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.connMu.Lock()
			if s.conn != c {
				s.connMu.Unlock()
				return
			}
			since := time.Since(s.lastAckTime)
			s.connMu.Unlock()
			if since > TCPHeartbeatAckTimeout {
				s.logger.Warn("tcp_server: heartbeat ack timeout, closing", "since", since)
				s.closeConn()
				return
			}
			s.writeMu.Lock()
			_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			err := WriteFrame(c, FrameHeartbeat, HeartbeatPayload{})
			s.writeMu.Unlock()
			if err != nil {
				s.logger.Warn("tcp_server: write heartbeat", "err", err)
				s.closeConn()
				return
			}
		}
	}
}

// flushPending writes any non-stale queued signals to the connected client.
// No-op if disconnected. Stale entries (>TCPStaleSignalAge) are dropped.
func (s *TCPServer) flushPending() error {
	s.connMu.Lock()
	c := s.conn
	s.connMu.Unlock()
	if c == nil {
		return nil
	}

	cutoff := s.staleAge()
	now := time.Now()

	s.pendingMu.Lock()
	kept := s.pending[:0]
	toSend := make([]SignalPayload, 0, len(s.pending))
	for _, ts := range s.pending {
		if now.Sub(ts.timestamp) > cutoff {
			s.logger.Warn("tcp_server: dropping stale signal", "signal_id", ts.payload.SignalID, "age", now.Sub(ts.timestamp))
			continue
		}
		toSend = append(toSend, ts.payload)
		_ = kept
	}
	// Drain pending: anything not stale is being sent now; anything stale was logged + dropped.
	s.pending = s.pending[:0]
	s.pendingMu.Unlock()

	for _, sig := range toSend {
		s.writeMu.Lock()
		_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
		err := WriteFrame(c, FrameSignal, sig)
		s.writeMu.Unlock()
		if err != nil {
			// Re-queue the un-sent signals (including this one) for next flush.
			s.pendingMu.Lock()
			s.pending = append(s.pending, timedSignal{payload: sig, timestamp: now})
			s.pendingMu.Unlock()
			s.logger.Warn("tcp_server: flush signal failed", "err", err, "signal_id", sig.SignalID)
			s.closeConn()
			return err
		}
	}
	return nil
}

func (s *TCPServer) closeConn() {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	if s.conn != nil {
		_ = s.conn.Close()
		s.conn = nil
	}
}
