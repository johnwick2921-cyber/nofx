# VLTrader TCP Wire Protocol

This is the local C#-side copy of the wire protocol. The authoritative implementation is `provider/ninjatrader/tcp_server.go` in the Go repo; any drift between this doc and the Go impl is a bug — file an issue.

Spec source: `docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md` lines 4343-4447.

## Transport

- TCP loopback: `127.0.0.1:36974` (note: NOT NinjaTrader's ATI port 36973).
- Single concurrent client: NT8 AddOn → Go server. A second client is rejected.
- Connection is opened by the C# AddOn at NT8 startup; reconnects every 5 seconds on disconnect.

## Framing

Each frame on the TCP stream is:

1. 4-byte **big-endian** unsigned 32-bit length prefix.
2. UTF-8 JSON payload of exactly `length` bytes.

Maximum frame size: **1 MB (1,048,576 bytes)**. Oversized frames are a protocol error — the receiving side closes the connection.

## Envelope

Every JSON body has the same outer shape:

```json
{
  "type": "signal" | "fill" | "heartbeat" | "ack",
  "payload": { }
}
```

## Message types

### 1. `signal` (Go server → C# AddOn)

Outgoing trade signal. Each numeric field is tick-rounded by the Go side before emission (see `trader/ninjatrader/tick_rounding.go`).

```json
{
  "type": "signal",
  "payload": {
    "symbol": "MNQ",
    "side": "long",
    "quantity": 1,
    "entry": 21500.25,
    "stop_loss": 21450.0,
    "take_profit": 21550.0,
    "signal_id": "uuid-v4-string",
    "timestamp": "2026-05-26T18:00:00Z"
  }
}
```

**Field semantics:**

- `symbol`: NT8 instrument symbol (e.g. `MNQ`, `NQ`, `ES`, `MES`).
- `side`: `long` (buy) or `short` (sell short).
- `quantity`: number of contracts.
- `entry`: market entry reference (NT8 uses market orders; this is the AI's planned entry — used only for slippage attribution in the fill frame).
- `stop_loss`: stop price (tick-rounded).
- `take_profit`: limit price (tick-rounded).
- `signal_id`: UUID. Used as the OCO group ID and as the fill correlation key.
- `timestamp`: RFC3339 UTC. The AddOn rejects signals older than 60 seconds as stale.

### 2. `fill` (C# AddOn → Go server)

Outgoing fill notification.

```json
{
  "type": "fill",
  "payload": {
    "signal_id": "matching-uuid",
    "fill_price": 21500.5,
    "fill_time": "2026-05-26T18:00:01Z",
    "side": "long",
    "quantity": 1,
    "slippage_ticks": 1.0,
    "status": "filled"
  }
}
```

**Field semantics:**

- `signal_id`: matches the originating signal's `signal_id`.
- `fill_price`: actual average fill price.
- `fill_time`: RFC3339 UTC.
- `slippage_ticks`: `(fill_price - entry) / tick_size`, signed (positive = paid more than planned for a long, less than planned for a short).
- `status`: `filled`, `rejected`, or `partial`. Rejected fills indicate NT8 refused the order; the Go side does NOT retry — manual operator intervention is required.

### 3. `heartbeat` (bidirectional)

Empty payload. Sent every 30 seconds by both sides.

```json
{ "type": "heartbeat", "payload": {} }
```

### 4. `ack` (bidirectional)

```json
{ "type": "ack", "payload": { "acks": "heartbeat" } }
```

Or for a specific signal:

```json
{ "type": "ack", "payload": { "acks": "uuid-v4-string" } }
```

## Failure modes

- **TCP disconnect**: the server holds the signal queue; on reconnect, it sends pending signals with their original timestamps. The C# AddOn may reject signals older than 60s as stale (emits a `status=rejected` fill).
- **Heartbeat timeout**: the server closes the connection after 60s without an ack; the C# AddOn reconnects every 5s.
- **Invalid frame** (bad length, >1 MB, malformed JSON): the receiver logs a warning and closes the connection; the other side reconnects.
- **Order rejection by NT8**: the AddOn emits a fill frame with `status=rejected`; the Go side logs the rejection and does NOT retry.

## Cross-references

- Go-side authoritative implementation: `provider/ninjatrader/tcp_server.go`.
- Go-side framing codec: `provider/ninjatrader/tcp_framing.go` (length-prefix + JSON envelope shared with this AddOn).
- Architectural rationale: `docs/adr/ADR-001-csv-bridge-vs-tcp.md`.
- Plan 1.5 spec: `docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md` lines 4343-4447.
- Plan 1 critical-file integrity guard: `docs/adr/ADR-007-plan1-critical-file-integrity.md` (Plan 1.5 is purely additive — none of the CSV bridge files are modified).
