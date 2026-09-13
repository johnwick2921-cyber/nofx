package ninjatrader

import "strings"

// OrderedExecutionHandlers runs the owning durable consumers on the TCP read
// goroutine, before advisory fanout. Callbacks must not wait for broker replies.
// Database backpressure deliberately delays feed reception rather than losing
// or reordering execution evidence. No router/cache/connection lock is held.
type OrderedExecutionHandlers struct {
	Order func(OrderUpdatePayload)
	Fill  func(FillPayload)
	Close func(PositionClosePayload)
}

// RegisterOrderedExecutionsFor atomically replaces one exact account/symbol
// owner. Ordinary trading pause must retain this observation registration.
func (s *TCPServer) RegisterOrderedExecutionsFor(symbol, account string, h OrderedExecutionHandlers) func() {
	if strings.TrimSpace(symbol) == "" || strings.TrimSpace(account) == "" {
		return func() {}
	}
	key := subKey(symbol, account)
	owner := &h
	s.executionMu.Lock()
	if s.executionOwners == nil {
		s.executionOwners = make(map[string]*OrderedExecutionHandlers)
	}
	s.executionOwners[key] = owner
	s.executionMu.Unlock()
	return func() {
		s.executionMu.Lock()
		defer s.executionMu.Unlock()
		if s.executionOwners[key] == owner {
			delete(s.executionOwners, key)
		}
	}
}
func (s *TCPServer) dispatchOrderedOrder(p *OrderUpdatePayload) {
	s.executionMu.Lock()
	defer s.executionMu.Unlock()
	if h := s.executionOwners[subKey(p.Symbol, p.Account)]; h != nil && h.Order != nil {
		p.OrderedArrival = true
		h.Order(*p)
		p.OrderedHandled = true
	}
}
func (s *TCPServer) dispatchOrderedFill(p *FillPayload) {
	s.executionMu.Lock()
	defer s.executionMu.Unlock()
	if h := s.executionOwners[subKey(p.Symbol, p.Account)]; h != nil && h.Fill != nil {
		h.Fill(*p)
		p.OrderedHandled = true
	}
}
func (s *TCPServer) dispatchOrderedClose(p *PositionClosePayload) {
	s.executionMu.Lock()
	defer s.executionMu.Unlock()
	if h := s.executionOwners[subKey(p.Symbol, p.Account)]; h != nil && h.Close != nil {
		h.Close(*p)
		p.OrderedHandled = true
	}
}
