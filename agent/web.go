package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type storeUserIDContextKey struct{}
type sessionPolicyContextKey struct{}

type SessionPolicy struct {
	Authenticated           bool
	IsAdmin                 bool
	CanExecuteTrade         bool
	CanViewSensitiveSecrets bool
}

// WithStoreUserID annotates an HTTP request context with the authenticated store user ID.
func WithStoreUserID(ctx context.Context, storeUserID string) context.Context {
	return context.WithValue(ctx, storeUserIDContextKey{}, storeUserID)
}

func storeUserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(storeUserIDContextKey{}).(string); ok && v != "" {
		return v
	}
	return "default"
}

func WithSessionPolicy(ctx context.Context, policy SessionPolicy) context.Context {
	return context.WithValue(ctx, sessionPolicyContextKey{}, policy)
}

func sessionPolicyFromContext(ctx context.Context) SessionPolicy {
	if v, ok := ctx.Value(sessionPolicyContextKey{}).(SessionPolicy); ok {
		return v
	}
	return SessionPolicy{}
}

// validSymbolRe matches a well-formed ticker symbol (e.g. MNQ, NQM6, NQ.c.0).
var validSymbolRe = regexp.MustCompile(`^[A-Za-z0-9.\-_]{1,20}$`)

// WebHandler provides HTTP endpoints for the NOFXi agent.
type WebHandler struct {
	agent  *Agent
	logger *slog.Logger
}

func NewWebHandler(agent *Agent, logger *slog.Logger) *WebHandler {
	return &WebHandler{agent: agent, logger: logger}
}

// HandleHealth handles GET /api/agent/health.
func (w *WebHandler) HandleHealth(rw http.ResponseWriter, r *http.Request) {
	writeJSON(rw, 200, map[string]string{"status": "ok", "agent": "NOFXi", "time": time.Now().Format(time.RFC3339)})
}

// HandleChat handles POST /api/agent/chat.
func (w *WebHandler) HandleChat(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "method not allowed", 405)
		return
	}
	var req struct {
		Message string `json:"message"`
		UserID  int64  `json:"user_id"`
		UserKey string `json:"user_key"`
		Lang    string `json:"lang"`
	}
	// Limit request body to 64KB to prevent abuse
	if err := json.NewDecoder(io.LimitReader(r.Body, 64*1024)).Decode(&req); err != nil {
		writeJSON(rw, 400, map[string]string{"error": "invalid request"})
		return
	}
	if req.Message == "" {
		writeJSON(rw, 400, map[string]string{"error": "message required"})
		return
	}
	if req.UserID == 0 {
		req.UserID = SessionUserIDFromKey(storeUserIDFromContext(r.Context()))
	}
	msg := req.Message
	if req.Lang != "" {
		msg = "[lang:" + req.Lang + "] " + msg
	}

	ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
	defer cancel()

	resp, err := w.agent.HandleMessageForStoreUser(ctx, storeUserIDFromContext(r.Context()), req.UserID, msg)
	if err != nil {
		w.logger.Error("agent HandleMessage failed", "error", err, "user_id", req.UserID)
		writeJSON(rw, 500, map[string]string{"error": "I ran into a problem while handling that message. Please try again."})
		return
	}
	writeJSON(rw, 200, map[string]string{"response": resp})
}

// HandleChatStream handles POST /api/agent/chat/stream — SSE streaming chat.
// Sends server-sent events with types including planning, plan, step_start,
// step_complete, replan, tool, delta, done, error.
func (w *WebHandler) HandleChatStream(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "method not allowed", 405)
		return
	}
	var req struct {
		Message string `json:"message"`
		UserID  int64  `json:"user_id"`
		UserKey string `json:"user_key"`
		Lang    string `json:"lang"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 64*1024)).Decode(&req); err != nil {
		writeJSON(rw, 400, map[string]string{"error": "invalid request"})
		return
	}
	if req.Message == "" {
		writeJSON(rw, 400, map[string]string{"error": "message required"})
		return
	}
	if req.UserID == 0 {
		req.UserID = SessionUserIDFromKey(storeUserIDFromContext(r.Context()))
	}
	msg := req.Message
	if req.Lang != "" {
		msg = "[lang:" + req.Lang + "] " + msg
	}

	// Set SSE headers
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	rw.WriteHeader(200)

	flusher, ok := rw.(http.Flusher)
	if !ok {
		writeSSE(rw, nil, "error", "streaming not supported")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	resp, err := w.agent.HandleMessageStreamForStoreUser(ctx, storeUserIDFromContext(r.Context()), req.UserID, msg, func(event, data string) {
		if ctx.Err() != nil {
			return
		}
		writeSSE(rw, flusher, event, data)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			w.logger.Info("agent stream cancelled", "user_id", req.UserID, "error", err)
			return
		}
		w.logger.Error("agent HandleMessageStream failed", "error", err, "user_id", req.UserID)
		writeSSE(rw, flusher, "error", "I ran into a problem while handling that message. Please try again.")
		return
	}
	if ctx.Err() != nil {
		return
	}
	// Send final done event with complete response
	writeSSE(rw, flusher, "done", resp)
}

// writeSSE writes a single SSE event.
func writeSSE(w http.ResponseWriter, flusher http.Flusher, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, sseEscape(data))
	if flusher != nil {
		flusher.Flush()
	}
}

// sseEscape escapes newlines in SSE data (each line needs a "data: " prefix).
func sseEscape(s string) string {
	// SSE spec: multi-line data uses multiple "data:" lines
	// But we use JSON encoding to avoid this complexity
	b, _ := json.Marshal(s)
	return string(b)
}

// HandleTickers handles GET /api/agent/tickers?symbols=MNQ,ES — the chat
// page's price ticker. It returns a JSON array with ONE object per requested
// symbol, in request order (see nt8Quote for the shape). A CME futures symbol
// is read from NinjaTrader (market.GetWithExchange(sym, "ninjatrader")); any
// other symbol, or a malformed one, gets price null and a named "unavailable"
// reason. There is no external or crypto fallback and no fabricated 0: an
// unread value is null. A request with no symbols is refused (400) — there is
// no default list.
func (w *WebHandler) HandleTickers(rw http.ResponseWriter, r *http.Request) {
	symbols := splitComma(r.URL.Query().Get("symbols"))
	if len(symbols) == 0 {
		writeJSON(rw, 400, map[string]string{"error": "symbols required: comma-separated CME futures symbols, e.g. symbols=MNQ"})
		return
	}
	if len(symbols) > maxTickerSymbols {
		writeJSON(rw, 400, map[string]string{"error": fmt.Sprintf("max %d symbols", maxTickerSymbols)})
		return
	}

	out := make([]nt8Quote, 0, len(symbols))
	for _, sym := range symbols {
		if !validSymbolRe.MatchString(sym) {
			out = append(out, nt8Quote{Symbol: sym, Source: nt8QuoteSource, Unavailable: invalidTickerSymbolReason})
			continue
		}
		out = append(out, readNT8Quote(sym))
	}
	writeJSON(rw, 200, out)
}

// maxTickerSymbols caps one /api/agent/tickers request.
const maxTickerSymbols = 20

// invalidTickerSymbolReason is the named reason for a symbol that fails validSymbolRe.
const invalidTickerSymbolReason = "invalid symbol: 1-20 letters, digits, '.', '-' or '_'"

// commaRe is pre-compiled for splitComma — avoids recompiling on every call.
var commaRe = regexp.MustCompile(`\s*,\s*`)

// splitComma splits a comma-separated string, trims whitespace, skips empty.
func splitComma(s string) []string {
	var parts []string
	for _, p := range commaRe.Split(strings.TrimSpace(s), -1) {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	// CORS is handled by the gin middleware — no need to set it here
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
