package mcp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Planner-speed wave phase 4 (2026-08-31) — the split-deadline contract on the
// SSE path: a live-but-slow stream survives the idle watchdog; a silent stream
// dies at the idle deadline (not 600s); the transport-reset class retries via
// the fixed retry flow; the ai_call line carries ttfb/reasoning/retries.

func streamClient(t *testing.T, url string, retries int) *Client {
	t.Helper()
	c := NewClient().(*Client)
	c.APIKey = "test-key"
	c.BaseURL = url
	c.UseFullURL = true
	c.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	c.Cfg.MaxRetries = retries
	c.Log = NewMockLogger()
	return c
}

func writeSSE(w http.ResponseWriter, data string) {
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// TestStreamSlowButAliveSurvivesIdle — chunks keep arriving (gap < idle) so a
// stream longer than the idle window completes; ttfb/reasoning/finish captured.
func TestStreamSlowButAliveSurvivesIdle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		time.Sleep(30 * time.Millisecond) // nonzero queue+think so ttfb is measurable
		writeSSE(w, `{"choices":[{"delta":{"reasoning_content":"thinking hard"},"finish_reason":null}]}`)
		for i := 0; i < 8; i++ { // 8 chunks × 60ms = 480ms > 150ms idle
			time.Sleep(60 * time.Millisecond)
			writeSSE(w, fmt.Sprintf(`{"choices":[{"delta":{"content":"chunk%d"},"finish_reason":null}]}`, i))
		}
		writeSSE(w, `{"choices":[{"delta":{"content":""},"finish_reason":"stop"}],"usage":{"prompt_tokens":100,"completion_tokens":200,"total_tokens":300}}`)
		writeSSE(w, `[DONE]`)
	}))
	defer srv.Close()

	c := streamClient(t, srv.URL, 1)
	req := &Request{Messages: []Message{NewUserMessage("hi")}}
	out, err := c.CallWithRequestStreamRetry(req, nil, 150*time.Millisecond)
	if err != nil {
		t.Fatalf("slow-but-alive stream must survive the idle watchdog, got %v", err)
	}
	if !strings.Contains(out, "chunk7") {
		t.Fatalf("stream text incomplete: %q", out)
	}
	if c.lastTTFBMs.Load() <= 0 {
		t.Fatalf("ttfb not stamped, got %d", c.lastTTFBMs.Load())
	}
	if c.lastReasoningChars.Load() <= 0 {
		t.Fatalf("reasoning chars not captured, got %d", c.lastReasoningChars.Load())
	}
	if f, _ := c.lastFinishReason.Load().(string); f != "stop" {
		t.Fatalf("finish_reason = %q, want stop", f)
	}
	// The retry wrapper emits the enriched ai_call line.
	got := c.Log.(*MockLogger)
	if !logContains(got, "ai_call model=") || !logContains(got, "ttfb_ms=") || !logContains(got, "reasoning_chars=") {
		t.Fatalf("ai_call line missing new fields: %+v", got.Logs)
	}
}

// TestStreamIdleSilenceAborts — a silent stream dies at the idle deadline with
// a context cancel (the retry fires immediately instead of burning 600s).
func TestStreamIdleSilenceAborts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		writeSSE(w, `{"choices":[{"delta":{"content":"first"},"finish_reason":null}]}`)
		time.Sleep(2 * time.Second) // silence ≫ idle
	}))
	defer srv.Close()

	c := streamClient(t, srv.URL, 1)
	req := &Request{Messages: []Message{NewUserMessage("hi")}}
	_, err := c.CallWithRequestStreamIdle(req, nil, 200*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("silent stream must die at the idle deadline (context canceled), got %v", err)
	}
}

// TestStreamTransportResetRetryEngages — the first connection is reset
// mid-stream; the fixed retry flow re-sends and wins (phase 4.4).
func TestStreamTransportResetRetryEngages(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		if n == 1 {
			writeSSE(w, `{"choices":[{"delta":{"content":"partial"},"finish_reason":null}]}`)
			if hj, ok := w.(http.Hijacker); ok {
				conn, _, _ := hj.Hijack()
				_ = conn.Close() // transport reset, mid-stream
				return
			}
			return
		}
		writeSSE(w, `{"choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":2,"total_tokens":12}}`)
		writeSSE(w, `[DONE]`)
	}))
	defer srv.Close()

	c := streamClient(t, srv.URL, 2)
	req := &Request{Messages: []Message{NewUserMessage("hi")}}
	out, err := c.CallWithRequestStreamRetry(req, nil, 5*time.Second)
	if err != nil {
		t.Fatalf("transport reset must retry and win, got %v", err)
	}
	if !strings.Contains(out, "ok") {
		t.Fatalf("retried stream text = %q", out)
	}
	if calls.Load() < 2 {
		t.Fatalf("expected ≥2 server hits, got %d", calls.Load())
	}
	got := c.Log.(*MockLogger)
	if !logContains(got, "AI API stream failed, retrying (2/2)") {
		t.Fatalf("retry warn missing: %+v", got.Logs)
	}
}

func logContains(m *MockLogger, sub string) bool {
	for _, e := range m.Logs {
		if strings.Contains(e.Message, sub) {
			return true
		}
	}
	return false
}
