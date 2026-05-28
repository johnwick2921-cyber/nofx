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

	// Plan 4.4 Stage 2 — bar ingest. Bar frames from the C# AddOn
	// (bars_historical, bar_update) decode in the read loop and post to
	// barIngestCh via a NON-BLOCKING drop-oldest send (for bar_update) or
	// a bounded BLOCKING send (for bars_historical). The drainBarIngest
	// goroutine consumes from this channel and writes into barCache.
	// Backpressure invariant: a slow cache writer must NOT stall the
	// socket read (which would block heartbeats → spurious reconnect).
	barCache      *BarCache
	barIngestCh   chan barIngestMsg
	barsSubscribe BarsSubscribePayload // sent on each (re)connect after flushPending
	barsSubMu     sync.RWMutex         // protects barsSubscribe

	// Connection state — single concurrent client (spec L4359).
	connMu        sync.Mutex
	conn          net.Conn
	lastAckTime   time.Time
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

// barIngestMsg is the internal envelope passed from the socket read loop
// to the drain goroutine that writes into barCache. The historical flag
// distinguishes the one-shot seed batch from streaming updates so the
// drainer can apply the correct cache method.
type barIngestMsg struct {
	historical bool
	symbol     string
	timeframe  string
	bars       []Bar
}

// barIngestChannelBuffer caps the bar ingest channel. Sized so a brief
// scheduler hiccup in the drain goroutine doesn't drop bar_updates on the
// floor under normal load; a sustained backlog under pathological load
// will drop oldest-first to keep the socket read responsive.
const barIngestChannelBuffer = 256

// Plan 4.4 Stage 2 — default auto-subscribe parameters. These match the
// Balanced Strategy's SelectedTimeframes (store/strategy.go) for the
// active futures trader and prove the end-to-end bar pipe. Stage 3 will
// replace these with a per-trader strategy lookup; for now the constants
// validate the framing + cache + handler chain.
var (
	defaultAutoBarsSymbol     = "MNQ"
	defaultAutoBarsTimeframes = []string{"5m", "15m", "1h"}
	defaultAutoBarsBack       = 500
)

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
		addr:        TCPListenAddr,
		fillCh:      make(chan FillPayload, fillChannelBuffer),
		barCache:    NewBarCache(0),
		barIngestCh: make(chan barIngestMsg, barIngestChannelBuffer),
		barsSubscribe: BarsSubscribePayload{
			Symbol:     defaultAutoBarsSymbol,
			Timeframes: append([]string(nil), defaultAutoBarsTimeframes...),
			BarsBack:   defaultAutoBarsBack,
		},
		logger: logger,
	}
}

// BarCache exposes the bar cache for the kernel (Stage 3) and chart
// relay (Stage 4). The cache is goroutine-safe; readers receive snapshot
// copies via Get and never block writers for long.
func (s *TCPServer) BarCache() *BarCache { return s.barCache }

// SetBarsSubscribe overrides the auto-subscribe parameters sent on each
// (re)connect. Stage 3 will call this with the active strategy's
// SelectedTimeframes; until then the server uses Balanced Strategy
// defaults (MNQ at 5m/15m/1h, 500 bars back).
func (s *TCPServer) SetBarsSubscribe(p BarsSubscribePayload) {
	s.barsSubMu.Lock()
	defer s.barsSubMu.Unlock()
	// Defensive copy of the timeframes slice to detach from caller storage.
	p.Timeframes = append([]string(nil), p.Timeframes...)
	s.barsSubscribe = p
}

func (s *TCPServer) currentBarsSubscribe() BarsSubscribePayload {
	s.barsSubMu.RLock()
	defer s.barsSubMu.RUnlock()
	return BarsSubscribePayload{
		Symbol:     s.barsSubscribe.Symbol,
		Timeframes: append([]string(nil), s.barsSubscribe.Timeframes...),
		BarsBack:   s.barsSubscribe.BarsBack,
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
	s.wg.Add(2)
	go s.acceptLoop(cctx)
	// Plan 4.4 Stage 2 — bar ingest drain goroutine, decouples cache
	// writes from the socket read loop.
	go s.drainBarIngest(cctx)
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
		// Plan 4.4 Stage 2 — auto-subscribe to bars on every accept so
		// the C# AddOn starts emitting bars_historical + bar_update
		// without operator action. Idempotent on the C# side per the
		// protocol contract (re-subscribing returns immediately).
		s.sendAutoBarsSubscribe(c)
	}
}

// sendAutoBarsSubscribe emits the configured bars_subscribe frame on a
// freshly-accepted connection. Failure here is non-fatal — the C# side
// will simply not start streaming bars; the next reconnect retries. We
// log a warn so the operator sees the failure.
func (s *TCPServer) sendAutoBarsSubscribe(c net.Conn) {
	payload := s.currentBarsSubscribe()
	if payload.Symbol == "" || len(payload.Timeframes) == 0 {
		// No subscription configured — Stage 2 default always populates
		// this, but defend against an explicit empty override.
		return
	}
	s.writeMu.Lock()
	_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := WriteFrame(c, FrameBarsSubscribe, payload)
	s.writeMu.Unlock()
	if err != nil {
		s.logger.Warn("tcp_server: send bars_subscribe", "err", err,
			"symbol", payload.Symbol, "timeframes", payload.Timeframes)
		return
	}
	s.logger.Info("tcp_server: sent bars_subscribe",
		"symbol", payload.Symbol,
		"timeframes", payload.Timeframes,
		"bars_back", payload.BarsBack)
}

// drainBarIngest consumes the bar ingest channel and applies updates to
// the cache. Runs as a background goroutine for the server's lifetime.
// Decoupling the cache writes from the socket read loop guarantees the
// read loop never stalls on cache lock contention — which would block
// heartbeat reads and trigger a spurious reconnect.
func (s *TCPServer) drainBarIngest(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.barIngestCh:
			if msg.historical {
				s.barCache.SeedHistorical(msg.symbol, msg.timeframe, msg.bars)
			} else {
				s.barCache.Upsert(msg.symbol, msg.timeframe, msg.bars)
			}
		}
	}
}

// enqueueBarUpdate posts a bar_update to the ingest channel using a
// drop-oldest, non-blocking send. If the channel is full (slow drain
// goroutine), we discard the OLDEST pending update to make room, log
// the drop once, and retry. If the channel is still full after the
// drop, we discard THIS update too — under sustained pressure the
// freshest in-flight update will still arrive on the next tick.
//
// This path MUST NOT block the socket read loop. Blocking the read
// loop stalls heartbeat receive → 60s ack timeout → server closes the
// conn → spurious reconnect cycle.
func (s *TCPServer) enqueueBarUpdate(symbol, timeframe string, bars []Bar) {
	msg := barIngestMsg{historical: false, symbol: symbol, timeframe: timeframe, bars: bars}
	select {
	case s.barIngestCh <- msg:
		return
	default:
	}
	// Channel full — drop oldest, log once, retry.
	select {
	case <-s.barIngestCh:
		s.logger.Warn("tcp_server: bar ingest backpressure — dropped oldest update",
			"symbol", symbol, "timeframe", timeframe)
	default:
	}
	select {
	case s.barIngestCh <- msg:
	default:
		s.logger.Warn("tcp_server: bar ingest backpressure — dropping current update",
			"symbol", symbol, "timeframe", timeframe)
	}
}

// enqueueBarHistorical posts a bars_historical message to the ingest
// channel. Historical batches are one-shot per (symbol, timeframe) and
// carry the full warmup window the kernel needs for indicator
// computation — they MUST NOT be dropped. We use a bounded blocking
// send guarded by a short timeout: if the drain goroutine is stuck
// longer than 2s, log + drop (better to lose this batch than deadlock
// the read loop forever).
func (s *TCPServer) enqueueBarHistorical(symbol, timeframe string, bars []Bar) {
	msg := barIngestMsg{historical: true, symbol: symbol, timeframe: timeframe, bars: bars}
	select {
	case s.barIngestCh <- msg:
		return
	case <-time.After(2 * time.Second):
		s.logger.Warn("tcp_server: bar ingest backpressure — dropped historical (drain stuck)",
			"symbol", symbol, "timeframe", timeframe, "bars", len(bars))
	}
}

func (s *TCPServer) readLoop(ctx context.Context, c net.Conn) {
	defer s.wg.Done()
	defer s.closeConn()

	// Plan 1.5.7 — per-connection ctx-cancel watcher unblocks the blocked
	// ReadFrame on server shutdown by closing this specific conn. Replaces
	// the previous SetReadDeadline(2s) + continue-on-timeout pattern that
	// caused frame desync: when a frame arrived near the 2s deadline,
	// io.ReadFull would consume partial bytes (the consumed bytes are gone
	// from the socket) then return a timeout; the `continue` then re-entered
	// ReadFrame which read 4 fresh bytes starting at the WRONG offset (mid-
	// header or mid-body), decoded as a garbage big-endian length usually
	// > 1 MB, and triggered ErrFrameTooLarge → spurious "oversized frame,
	// closing connection" warnings (~49 over 16h on quiet traffic).
	//
	// connCtx is derived from the server ctx so it cancels on shutdown.
	// defer connCancel() runs on normal disconnect, releasing the watcher
	// goroutine without a leak. c.Close() is idempotent.
	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel()
	go func() {
		<-connCtx.Done()
		_ = c.Close()
	}()

	for {
		// No SetReadDeadline — ReadFrame blocks until a frame arrives OR
		// the connection is closed (by the watcher above on shutdown, or
		// by the peer). The watcher provides the shutdown-responsiveness
		// the old 2s polling deadline used to.
		env, err := ReadFrame(c)
		if err != nil {
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

		case FrameBarsHistorical:
			// Plan 4.4 Stage 2 — one-shot per (symbol, timeframe) after
			// the initial BarsRequest load. Seeds the cache.
			var p BarsHistoricalPayload
			if err := json.Unmarshal(env.Payload, &p); err != nil {
				s.logger.Warn("tcp_server: bad bars_historical payload", "err", err)
				continue
			}
			s.enqueueBarHistorical(p.Symbol, p.Timeframe, p.Bars)

		case FrameBarUpdate:
			// Plan 4.4 Stage 2 — streaming updates. The bars array may
			// carry multiple bar indices per the NT8 multi-bar gotcha
			// (a single tick can update MinIndex..MaxIndex); the cache's
			// Upsert handles each in order via dedup-by-t.
			var p BarUpdatePayload
			if err := json.Unmarshal(env.Payload, &p); err != nil {
				s.logger.Warn("tcp_server: bad bar_update payload", "err", err)
				continue
			}
			s.enqueueBarUpdate(p.Symbol, p.Timeframe, p.Bars)

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
