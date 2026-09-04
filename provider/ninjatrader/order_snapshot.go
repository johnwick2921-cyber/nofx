// F12 — the order_snapshot frame: the BROKER's working-order book.
//
// WHY THIS EXISTS. The AddOn emitted position, fill, order_update, bar and
// account frames — everything except the one thing a cutover needs to know:
// what orders are resting at the broker RIGHT NOW. Class 33 made cutover leg 4
// read our own armed_orders ledger, and the gate's own note admitted it: "the
// armed_orders ledger (no NT8 order frame — F12 open)". A gate leg answered by
// our own bookkeeping cannot catch the case where the ledger and the broker
// disagree, which is exactly the case a flat gate exists to catch.
//
// order_update frames are per-EVENT: a Go restart loses the picture until the
// next event happens, which may be never on a quiet book. A periodic snapshot
// makes the broker's book re-derivable at any moment, and an EMPTY book arrives
// as `orders: []` — an explicit empty, never an absent frame. That distinction
// is the whole reason leg 4 can be trusted: "no orders" and "no answer" are
// different claims and must never render as the same one.
package ninjatrader

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// NT8Order is one order in the broker's book, as the AddOn sees it.
type NT8Order struct {
	OrderID    string  `json:"order_id"`
	Name       string  `json:"name,omitempty"`
	Action     string  `json:"action,omitempty"` // buy | sell
	Type       string  `json:"type,omitempty"`   // limit | stop | stop_limit | market
	LimitPrice float64 `json:"limit_price,omitempty"`
	StopPrice  float64 `json:"stop_price,omitempty"`
	Quantity   int     `json:"quantity,omitempty"`
	Filled     int     `json:"filled,omitempty"`
	State      string  `json:"state,omitempty"`
	OCO        string  `json:"oco,omitempty"`
	TimeMs     int64   `json:"time_ms,omitempty"`
}

// terminalOrderStates are the states that mean "this order is history". Held as
// a set rather than inline in a condition so leg 4 and the override guard
// cannot drift apart on what "working" means.
var terminalOrderStates = map[string]bool{
	"filled":          true,
	"cancelled":       true,
	"canceled":        true,
	"rejected":        true,
	"expired":         true,
	"unknown":         true,
	"partfilled_done": true,
}

// IsWorking reports whether the order still stands at the broker.
func (o NT8Order) IsWorking() bool {
	return !terminalOrderStates[strings.ToLower(strings.TrimSpace(o.State))]
}

// OrderSnapshotPayload is the frame body.
type OrderSnapshotPayload struct {
	Account string `json:"account"`
	Symbol  string `json:"symbol"`
	// BuildID is the AddOn's VL_BUILD_ID. It rides every snapshot so the running
	// DLL can be identified from a RECEIVED frame — the only proof that a
	// distributed change actually landed (class 6). A build id read from our own
	// source proves nothing about what NT8 loaded.
	BuildID   string     `json:"build_id"`
	EmittedMs int64      `json:"emitted_at_ms"`
	Reason    string     `json:"reason,omitempty"` // periodic | state_change
	Orders    []NT8Order `json:"orders"`
}

// WorkingOrders is the non-terminal subset — what leg 4 counts.
func (p OrderSnapshotPayload) WorkingOrders() []NT8Order {
	out := make([]NT8Order, 0, len(p.Orders))
	for _, o := range p.Orders {
		if o.IsWorking() {
			out = append(out, o)
		}
	}
	return out
}

// ParseOrderSnapshot decodes a frame body. A malformed frame is an ERROR the
// caller logs and drops (A10) — never a panic, and never a half-parsed snapshot
// that would poison the cache with a book that was never received.
func ParseOrderSnapshot(b []byte) (OrderSnapshotPayload, error) {
	var p OrderSnapshotPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return OrderSnapshotPayload{}, fmt.Errorf("order_snapshot: bad payload: %w", err)
	}
	// A frame with no symbol cannot be filed against an instrument. Filing it
	// under "" would let it answer for every symbol at once.
	if strings.TrimSpace(p.Symbol) == "" {
		return OrderSnapshotPayload{}, fmt.Errorf("order_snapshot: no symbol — unaddressable frame")
	}
	if p.Orders == nil {
		// An absent list and an empty list must not collapse into each other:
		// the AddOn sends [] for an empty book, and a nil here would later read
		// as "we never got a book".
		p.Orders = []NT8Order{}
	}
	return p, nil
}

// cachedSnapshot is a received frame plus the instant WE received it. Age is
// measured against receipt, not against the AddOn's emitted_at: the two clocks
// are different machines, and a Windows clock skew must not make a stale book
// look fresh.
type cachedSnapshot struct {
	Payload    OrderSnapshotPayload
	ReceivedAt time.Time
}

// OrderSnapshotCache holds the latest book per account+symbol.
type OrderSnapshotCache struct {
	mu    sync.RWMutex
	byKey map[string]cachedSnapshot
}

// NOTE ON build_id: this cache deliberately does NOT keep one. The server has
// owned the far-side build id since the E7 handshake (TCPServer.farSideBuild,
// exposed as FarSideBuildID) and the snapshot frame feeds THAT field rather
// than a second copy here. Two stores of the same received value are free to
// disagree, and a boot line would then have to pick one — which is how a
// "received" value quietly becomes a guess.

func NewOrderSnapshotCache() *OrderSnapshotCache {
	return &OrderSnapshotCache{byKey: map[string]cachedSnapshot{}}
}

func snapKey(account, symbol string) string {
	return strings.ToUpper(strings.TrimSpace(account)) + "|" + strings.ToUpper(strings.TrimSpace(symbol))
}

// PutAt stores a snapshot. The receipt instant is passed IN (class 60 / A28):
// nothing under here reads a clock, so a test states its own and the boot line,
// the gate and the guard all agree on one instant.
func (c *OrderSnapshotCache) PutAt(p OrderSnapshotPayload, receivedAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byKey[snapKey(p.Account, p.Symbol)] = cachedSnapshot{Payload: p, ReceivedAt: receivedAt}
}

// Latest returns the newest snapshot for the key.
func (c *OrderSnapshotCache) Latest(account, symbol string) (OrderSnapshotPayload, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byKey[snapKey(account, symbol)]
	return s.Payload, ok
}

// AgeAt is how long ago the latest snapshot for the key was RECEIVED. The
// second return is false when there is no snapshot at all — the caller must
// distinguish "no book" from "a book of age 0", which is precisely the
// plausible-zero the gate rules forbid (A24).
func (c *OrderSnapshotCache) AgeAt(account, symbol string, now time.Time) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.byKey[snapKey(account, symbol)]
	if !ok {
		return 0, false
	}
	return now.Sub(s.ReceivedAt), true
}

// OrderSnapshots exposes the cache so the trader layer can read the broker's
// book for cutover leg 4 and the override guard.
func (s *TCPServer) OrderSnapshots() *OrderSnapshotCache { return s.orderSnaps }

// SetOrderSnapshotSink installs the persistence hook. Called once at wiring
// time; nil disables persistence without disabling the cache, so a store
// failure can never cost the gate its broker read.
func (s *TCPServer) SetOrderSnapshotSink(fn func(OrderSnapshotPayload)) { s.orderSnapCB = fn }
