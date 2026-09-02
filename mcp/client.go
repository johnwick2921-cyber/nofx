package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ProviderCustom = "custom"

	MCPClientTemperature = 0.5
)

var (
	// Overall request ceiling. Generous because reasoning models stream slowly
	// (hidden reasoning + long completions); the idle-timeout watchdog in the
	// streaming path catches genuinely hung connections sooner.
	DefaultTimeout = 300 * time.Second

	MaxRetryTimes = 3

	retryableErrors = []string{
		"EOF",
		"timeout",
		"connection reset",
		"connection refused",
		"temporary failure",
		"no such host",
		"stream error",   // HTTP/2 stream error
		"INTERNAL_ERROR", // Server internal error
		"status 429",     // Rate limit / upstream gateway throttling
		"rate_limit_error",
		"upstream_empty_output",
		"status 502", // Bad Gateway
		"status 503", // Service Unavailable
		"status 520", // Cloudflare origin error
		"status 524", // Cloudflare timeout
	}

	// TokenUsageCallback is called after each AI request with token usage info
	TokenUsageCallback func(usage TokenUsage)

	// TruncatedResponses counts finish_reason=length responses — the P0
	// 2026-08-19 disease where the whole output budget was spent on reasoning
	// and the decision block never got emitted. Surfaced in the startup AI-params
	// line and bumped with a WARN log on every occurrence.
	TruncatedResponses atomic.Int64
)

// TokenUsage represents token usage from AI API response
type TokenUsage struct {
	Provider         string // payment channel: "claw402" or native provider name
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Channel returns the payment channel category for telemetry.
// Returns "claw402" or "native" based on the provider.
func (u TokenUsage) Channel() string {
	switch u.Provider {
	case ProviderClaw402:
		return "claw402"
	default:
		return "native"
	}
}

// Client AI API configuration
type Client struct {
	Provider   string
	APIKey     string
	BaseURL    string
	Model      string
	UseFullURL bool // Whether to use full URL (without appending /chat/completions)
	MaxTokens  int  // Maximum tokens for AI response

	HTTPClient *http.Client // Exported for sub-packages
	Log        Logger       // Exported for sub-packages
	Cfg        *Config      // Exported for sub-packages

	// Hooks are used to implement dynamic dispatch (polymorphism)
	// When provider.DeepSeekClient embeds Client, Hooks point to DeepSeekClient
	// This way methods called in Call() are automatically dispatched to the overridden version
	Hooks ClientHooks

	// lastFinishReason holds the most recent response's finish_reason for the
	// structured ai_call log (atomic: the planner client is shared between the
	// executor loop and the Ask-Planner API). Claude overrides ParseMCPResponse
	// and does not set it — those calls log finish_reason=unknown.
	lastFinishReason atomic.Value

	// Per-call latency telemetry (atomics — the shared client serves the
	// executor loop, the planner goroutine, and API ask-planner concurrently;
	// these are read immediately after each single call returns).
	lastTTFBMs         atomic.Int64 // time-to-first-byte of the last single call (0 = never measured)
	lastReasoningChars atomic.Int64 // reasoning_content chars of the last single call

	// Class 37 (2026-09-01) — failure-class telemetry for the ai_call line: the
	// HTTP status and provider request id of the last response (0/"" when no
	// response arrived) and the classified error of the last FAILED call.
	lastHTTPStatus atomic.Int64
	lastRequestID  atomic.Value // string
	lastErrClass   atomic.Value // string

	// reasoningTokensAbsentLogged ensures the explicit one-time "the provider
	// usage carries no reasoning_tokens field" note is printed once per client.
	reasoningTokensAbsentLogged sync.Once
}

// New creates default client (backward compatible)
//
// Deprecated: Recommend using NewClient(...opts) for better flexibility
func New() AIClient {
	return NewClient()
}

// NewClient creates client (supports options pattern)
//
// Usage examples:
//
//	// Basic usage (backward compatible)
//	client := mcp.NewClient()
//
//	// Custom logger
//	client := mcp.NewClient(mcp.WithLogger(customLogger))
//
//	// Custom timeout
//	client := mcp.NewClient(mcp.WithTimeout(60*time.Second))
//
//	// Combine multiple options
//	client := mcp.NewClient(
//	    mcp.WithDeepSeekConfig("sk-xxx"),
//	    mcp.WithLogger(customLogger),
//	    mcp.WithTimeout(60*time.Second),
//	)
func NewClient(opts ...ClientOption) AIClient {
	// 1. Create default config
	cfg := DefaultConfig()

	// 2. Apply user options
	for _, opt := range opts {
		opt(cfg)
	}

	// 3. Create client instance
	client := &Client{
		Provider:   cfg.Provider,
		APIKey:     cfg.APIKey,
		BaseURL:    cfg.BaseURL,
		Model:      cfg.Model,
		MaxTokens:  cfg.MaxTokens,
		UseFullURL: cfg.UseFullURL,
		HTTPClient: cfg.HTTPClient,
		Log:        cfg.Logger,
		Cfg:        cfg,
	}

	// 4. Set default Provider (if not set)
	if client.Provider == "" {
		client.Provider = ProviderDeepSeek
		client.BaseURL = DefaultDeepSeekBaseURL
		client.Model = DefaultDeepSeekModel
	}

	// 5. Set hooks to point to self
	client.Hooks = client

	return client
}

// SetCustomAPI sets custom OpenAI-compatible API
func (client *Client) SetAPIKey(apiKey, apiURL, customModel string) {
	client.Provider = ProviderCustom
	client.APIKey = apiKey

	// Check if URL ends with #, if so use full URL (without appending /chat/completions)
	if strings.HasSuffix(apiURL, "#") {
		client.BaseURL = strings.TrimSuffix(apiURL, "#")
		client.UseFullURL = true
	} else {
		client.BaseURL = apiURL
		client.UseFullURL = false
	}

	client.Model = customModel
}

// ResolvedModel returns the EXACT model string this client calls (client.Model,
// set to the provider default or a custom override) — never a provider alias.
func (client *Client) ResolvedModel() string { return client.Model }

func (client *Client) SetTimeout(timeout time.Duration) {
	client.HTTPClient.Timeout = timeout
}

// Class 37 (2026-09-01) — the two split deadlines on the planner stream carry
// sentinels so the ai_call line can name the killer instead of the generic
// net/http "context deadline exceeded" text.
var (
	// ErrStreamIdleDeadline — the idle watchdog cancelled a SILENT stream.
	ErrStreamIdleDeadline = errors.New("stream idle deadline exceeded")
	// ErrStreamTotalDeadline — the planner's whole-call ceiling fired on a
	// LIVE stream (distinct from http.Client.Timeout, the executor's ceiling).
	ErrStreamTotalDeadline = errors.New("stream total deadline exceeded")
)

// classifyAIError maps a failed call's error to ONE class token for the
// ai_call line: total_deadline · idle_deadline · client_timeout · http_status ·
// auth_config · transport · context · other. Most-specific first.
func classifyAIError(err error) string {
	if err == nil {
		return "ok"
	}
	msg := err.Error()
	switch {
	case errors.Is(err, ErrStreamTotalDeadline):
		return "total_deadline"
	case errors.Is(err, ErrStreamIdleDeadline):
		return "idle_deadline"
	case strings.Contains(msg, "Client.Timeout"):
		return "client_timeout"
	case strings.Contains(msg, "(status "):
		return "http_status"
	case strings.Contains(msg, "API key not set"):
		return "auth_config"
	case strings.Contains(msg, "connection reset"), strings.Contains(msg, "broken pipe"),
		strings.Contains(msg, "connection refused"), strings.Contains(msg, "no such host"),
		strings.Contains(msg, "i/o timeout"), strings.Contains(msg, "TLS handshake"),
		strings.Contains(msg, "EOF"), strings.Contains(msg, "stream error"):
		return "transport"
	case strings.Contains(msg, "context deadline exceeded"), strings.Contains(msg, "context canceled"):
		return "context"
	}
	return "other"
}

// ClassifyAIError (class 41) exports the ai_call class token for callers
// outside mcp (the planner attempt loop decides resend-vs-repair on it).
func ClassifyAIError(err error) string { return classifyAIError(err) }

// IsProviderFailure (class 41, owner ruling class 37 M4) — true when a failed
// call never produced a model answer to repair: transport cut, idle/total
// deadline, http.Client.Timeout, context. The planner re-sends the IDENTICAL
// prompt on these; validator/parse rejects (class=other, http_status…) keep
// the reject/repair flow.
func IsProviderFailure(err error) bool {
	if err == nil {
		return false
	}
	switch classifyAIError(err) {
	case "transport", "idle_deadline", "total_deadline", "client_timeout", "context":
		return true
	case "http_status":
		// 2026-09-02 01:15 CT: "Server Overloaded" 503 ×3 burned all three
		// planner attempts in 9 s, each re-authored with the 503 text as its
		// "validator reason". A 5xx or 429 is the provider, not the prompt.
		st := httpStatusFrom(err.Error())
		return st >= 500 || st == 429
	}
	return false
}

// httpStatusFrom extracts NNN from "(status NNN)" in an API error message; 0
// when absent.
func httpStatusFrom(msg string) int {
	i := strings.Index(msg, "(status ")
	if i < 0 {
		return 0
	}
	n := 0
	for _, r := range msg[i+len("(status "):] {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
		if n > 999 {
			return 0
		}
	}
	return n
}

// requestIDFrom returns the provider's request id header when present (the
// attributable handle for a provider-side ticket); "" otherwise.
func requestIDFrom(h http.Header) string {
	for _, k := range []string{"X-Request-Id", "Request-Id", "X-Amzn-Requestid", "Cf-Ray"} {
		if v := strings.TrimSpace(h.Get(k)); v != "" {
			return v
		}
	}
	return ""
}

func (client *Client) lastRequestIDString() string {
	if v := client.lastRequestID.Load(); v != nil {
		s, _ := v.(string)
		return s
	}
	return ""
}

// resetCallTelemetry clears the per-call atomics so a FAILED call never
// reports the previous call's ttfb/reasoning/status/request-id (observed
// 2026-09-01 06:10:36 CT: an executor transport reset logged
// reasoning_chars=7092 inherited from the planner stream that ended 06:09:33).
func (client *Client) resetCallTelemetry() {
	client.lastTTFBMs.Store(0)
	client.lastReasoningChars.Store(0)
	client.lastHTTPStatus.Store(0)
	client.lastRequestID.Store("")
}

// LastErrClass / LastHTTPStatus / LastRequestID — read-only helpers for the
// planner's failure line (class 37): the class of the last FAILED call ("" after
// a success), and the status / provider request id of the last response.
func LastErrClass(c AIClient) string {
	bc, ok := c.(interface{ BaseClient() *Client })
	if !ok {
		return ""
	}
	if v := bc.BaseClient().lastErrClass.Load(); v != nil {
		s, _ := v.(string)
		return s
	}
	return ""
}

func LastHTTPStatus(c AIClient) int {
	bc, ok := c.(interface{ BaseClient() *Client })
	if !ok {
		return 0
	}
	return int(bc.BaseClient().lastHTTPStatus.Load())
}

func LastRequestID(c AIClient) string {
	bc, ok := c.(interface{ BaseClient() *Client })
	if !ok {
		return ""
	}
	return bc.BaseClient().lastRequestIDString()
}

// logAICall emits ONE structured line per AI call so the next timeout is
// self-diagnosing instead of a forensic hunt (incident 2026-08-18: the only
// evidence was a bare duration and a generic net/http error string).
//
//	ai_call model=<m> duration_ms=<d> finish_reason=<r> ok=<bool>
//	  retries=<n> ttfb_ms=<t> reasoning_chars=<c>
//	  + on failure: timeout_source=planner_total|stream_idle|client|context|transport
//	    deadline_s=<n> class=<token> http_status=<n> request_id=<id> (class 37)
func (client *Client) logAICall(start time.Time, callErr error, retries int) {
	if client.Log == nil {
		return
	}
	durMs := time.Since(start).Milliseconds()
	finish := "unknown"
	if v, ok := client.lastFinishReason.Load().(string); ok && callErr == nil {
		finish = v
	}
	if callErr == nil {
		client.lastErrClass.Store("")
		client.Log.Infof("ai_call model=%s duration_ms=%d finish_reason=%s ok=true retries=%d ttfb_ms=%d reasoning_chars=%d http_status=%d request_id=%q",
			client.Model, durMs, finish, retries, client.lastTTFBMs.Load(), client.lastReasoningChars.Load(), client.lastHTTPStatus.Load(), client.lastRequestIDString())
		return
	}
	// Which deadline actually fired. net/http wraps them all in the same
	// "context deadline exceeded" text, hence the incident's ambiguity.
	msg := callErr.Error()
	source := "transport"
	switch {
	case errors.Is(callErr, ErrStreamTotalDeadline):
		source = "planner_total" // class 37 — the planner's own whole-call ceiling
	case errors.Is(callErr, ErrStreamIdleDeadline):
		source = "stream_idle" // the idle watchdog on a silent stream
	case strings.Contains(msg, "Client.Timeout"):
		source = "client" // http.Client.Timeout — the whole-request ceiling
	case strings.Contains(msg, "context deadline exceeded"), strings.Contains(msg, "context canceled"):
		source = "context" // a caller-supplied ctx (agent paths)
	}
	deadline := int64(0)
	if client.HTTPClient != nil {
		deadline = int64(client.HTTPClient.Timeout / time.Second)
	}
	// Class 37 — ONE class token per failure so "the API keeps failing" can
	// never again be the whole diagnosis; status + provider request id ride
	// along when a response arrived at all.
	class := classifyAIError(callErr)
	client.lastErrClass.Store(class)
	client.Log.Warnf("ai_call model=%s duration_ms=%d finish_reason=n/a ok=false retries=%d ttfb_ms=%d reasoning_chars=%d timeout_source=%s deadline_s=%d class=%s http_status=%d request_id=%q err=%q",
		client.Model, durMs, retries, client.lastTTFBMs.Load(), client.lastReasoningChars.Load(), source, deadline, class, client.lastHTTPStatus.Load(), client.lastRequestIDString(), msg)
}

// CallWithMessages template method - fixed retry flow (cannot be overridden)
func (client *Client) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	if client.APIKey == "" {
		return "", fmt.Errorf("AI API key not set, please call SetAPIKey first")
	}

	// Fixed retry flow
	var lastErr error
	maxRetries := client.Cfg.MaxRetries

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			client.Log.Warnf("⚠️  AI API call failed, retrying (%d/%d)...", attempt, maxRetries)
		}

		// Call the fixed single-call flow
		callStart := time.Now()
		result, err := client.Hooks.Call(systemPrompt, userPrompt)
		client.logAICall(callStart, err, attempt)
		if err == nil {
			if attempt > 1 {
				client.Log.Infof("✓ AI API retry succeeded")
			}
			return result, nil
		}

		lastErr = err
		// Check if error is retryable via hooks (supports custom retry strategy)
		if !client.Hooks.IsRetryableError(err) {
			return "", err
		}

		// Wait before retry
		if attempt < maxRetries {
			waitTime := client.Cfg.RetryWaitBase * time.Duration(attempt)
			client.Log.Infof("⏳ Waiting %v before retry...", waitTime)
			if err := sleepWithContext(context.Background(), waitTime); err != nil {
				return "", err
			}
		}
	}

	return "", fmt.Errorf("still failed after %d retries: %w", maxRetries, lastErr)
}

func (client *Client) SetAuthHeader(reqHeader http.Header) {
	reqHeader.Set("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))
}

func (client *Client) BuildMCPRequestBody(systemPrompt, userPrompt string) map[string]any {
	// Build messages array
	messages := []map[string]string{}

	// If system prompt exists, add system message
	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}
	// Add user message
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": userPrompt,
	})

	// Guard: truncate messages if they would exceed the model's context window
	if client.Cfg.MaxContext > 0 {
		truncated, removed := truncateMessages(messages, client.Cfg.MaxContext, client.MaxTokens)
		if removed > 0 {
			client.Log.Warnf("⚠️  [%s] Context guard: truncated %d oldest messages to fit within %d token limit",
				client.String(), removed, client.Cfg.MaxContext)
			messages = truncated
		}
	}

	// Build request body
	requestBody := map[string]interface{}{
		"model":       client.Model,
		"messages":    messages,
		"temperature": client.Cfg.Temperature, // Use configured temperature
	}
	// P0 2026-08-19 — top_p only when the operator set it (0 = omit, provider default).
	if client.Cfg.TopP > 0 {
		requestBody["top_p"] = client.Cfg.TopP
	}
	// OpenAI newer models use max_completion_tokens instead of max_tokens
	if client.Provider == ProviderOpenAI {
		requestBody["max_completion_tokens"] = client.MaxTokens
	} else {
		requestBody["max_tokens"] = client.MaxTokens
	}
	if client.Provider == ProviderDeepSeek {
		applyDeepSeekThinkingDefaults(requestBody, client.Cfg)
	}
	return requestBody
}

// MarshalRequestBody can be used to marshal the request body and can be overridden
func (client *Client) MarshalRequestBody(requestBody map[string]any) ([]byte, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}
	return jsonData, nil
}

func (client *Client) ParseMCPResponse(body []byte) (string, error) {
	r, err := client.ParseMCPResponseFull(body)
	if err != nil {
		return "", err
	}
	return r.Content, nil
}

// ParseMCPResponseFull parses the OpenAI-format response body and returns both
// the text content and any tool calls.
func (client *Client) ParseMCPResponseFull(body []byte) (*LLMResponse, error) {
	var result struct {
		Choices []struct {
			Message struct {
				Content          string     `json:"content"`
				ReasoningContent string     `json:"reasoning_content"`
				ToolCalls        []ToolCall `json:"tool_calls"`
			} `json:"message"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("API returned empty response")
	}

	msg := result.Choices[0].Message

	// Report token usage if callback is set
	if TokenUsageCallback != nil && result.Usage.TotalTokens > 0 {
		TokenUsageCallback(TokenUsage{
			Provider:         client.Provider,
			Model:            client.Model,
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		})
	}

	// P0 2026-08-19 — finish_reason=length means the decision was TRUNCATED
	// (the max_tokens=2000 disease). Count it, shout about it, and surface the
	// effective parameters so the owner can see what is in force.
	if len(result.Choices) > 0 && result.Choices[0].FinishReason != nil {
		if *result.Choices[0].FinishReason == "length" {
			TruncatedResponses.Add(1)
			client.Log.Warnf("🚨 [%s] finish_reason=length — response TRUNCATED at %d completion tokens (prompt %d). The decision block may be missing.",
				client.String(), result.Usage.CompletionTokens, result.Usage.PromptTokens)
		}
	}

	// P0 2026-08-19 — local per-call evidence for the live-verification report:
	// completion tokens + finish_reason on every response, so median tokens and
	// stop/length counts are computable from journalctl (telemetry only ships
	// off-box, which left this invisible before).
	if client.Log != nil {
		finish := "unset"
		if len(result.Choices) > 0 && result.Choices[0].FinishReason != nil {
			finish = *result.Choices[0].FinishReason
		}
		rc := len(msg.ReasoningContent)
		client.lastReasoningChars.Store(int64(rc))
		// PLANNER SPEED WAVE 1.3 — the provider's usage block carries NO
		// reasoning-token count (deepseek returns reasoning_content text only);
		// reasoning_chars is the proxy. Say so once per client.
		client.reasoningTokensAbsentLogged.Do(func() {
			client.Log.Infof("📊 provider usage carries no reasoning_tokens field (deepseek) — reasoning_chars is the logged proxy")
		})
		client.Log.Infof("📊 AI call complete: completion=%d prompt=%d finish_reason=%s reasoning_chars=%d",
			result.Usage.CompletionTokens, result.Usage.PromptTokens, finish, rc)
		client.lastFinishReason.Store(finish)
	}

	return &LLMResponse{
		Content:          msg.Content,
		ReasoningContent: msg.ReasoningContent,
		ToolCalls:        msg.ToolCalls,
	}, nil
}

func (client *Client) BuildUrl() string {
	if client.UseFullURL {
		return client.BaseURL
	}
	return fmt.Sprintf("%s/chat/completions", client.BaseURL)
}

func (client *Client) BuildRequest(url string, jsonData []byte) (*http.Request, error) {
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("fail to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set auth header via hooks (supports overriding)
	client.Hooks.SetAuthHeader(req.Header)

	return req, nil
}

func contextFromRequest(req *Request) context.Context {
	if req != nil && req.Ctx != nil {
		return req.Ctx
	}
	return context.Background()
}

func (client *Client) buildHTTPRequestWithContext(ctx context.Context, url string, jsonData []byte) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	httpReq, err := client.Hooks.BuildRequest(url, jsonData)
	if err != nil {
		return nil, err
	}
	return httpReq.WithContext(ctx), nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Call single AI API call (fixed flow, cannot be overridden)
func (client *Client) Call(systemPrompt, userPrompt string) (string, error) {
	// Print current AI configuration
	client.Log.Infof("📡 [%s] Request AI Server: BaseURL: %s", client.String(), client.BaseURL)
	client.Log.Debugf("[%s] UseFullURL: %v", client.String(), client.UseFullURL)
	if len(client.APIKey) > 8 {
		client.Log.Debugf("[%s]   API Key: %s...%s", client.String(), client.APIKey[:4], client.APIKey[len(client.APIKey)-4:])
	}

	client.resetCallTelemetry() // class 37: a failed call must not inherit the previous call's numbers

	// Step 1: Build request body (via hooks for dynamic dispatch)
	requestBody := client.Hooks.BuildMCPRequestBody(systemPrompt, userPrompt)

	// Step 2: Serialize request body (via hooks for dynamic dispatch)
	jsonData, err := client.Hooks.MarshalRequestBody(requestBody)
	if err != nil {
		return "", err
	}

	// Step 3: Build URL (via hooks for dynamic dispatch)
	url := client.Hooks.BuildUrl()
	client.Log.Infof("📡 [MCP %s] Request URL: %s", client.String(), url)

	// Step 4: Create HTTP request (fixed logic)
	req, err := client.Hooks.BuildRequest(url, jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Step 5: Send HTTP request (fixed logic)
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	client.lastHTTPStatus.Store(int64(resp.StatusCode))
	client.lastRequestID.Store(requestIDFrom(resp.Header))

	// Step 6: Read response body (fixed logic) — stamp time-to-first-byte.
	body, err := readWithTTFB(resp.Body, &client.lastTTFBMs)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Step 7: Check HTTP status code (fixed logic)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned error (status %d): %s", resp.StatusCode, string(body))
	}

	// Step 8: Parse response (via hooks for dynamic dispatch)
	result, err := client.Hooks.ParseMCPResponse(body)
	if err != nil {
		return "", fmt.Errorf("fail to parse AI server response: %w", err)
	}

	return result, nil
}

// ttfbReader stamps elapsed-since-request time on the first Read — the
// time-to-first-byte (T4) evidence for the queue-vs-generation split.
type ttfbReader struct {
	r       io.Reader
	start   time.Time
	stamped *atomic.Int64
	once    bool
}

func (t *ttfbReader) Read(p []byte) (int, error) {
	if !t.once {
		t.once = true
		t.stamped.Store(time.Since(t.start).Milliseconds())
	}
	return t.r.Read(p)
}

func readWithTTFB(r io.Reader, stamped *atomic.Int64) ([]byte, error) {
	return io.ReadAll(&ttfbReader{r: r, start: time.Now(), stamped: stamped})
}

func (client *Client) String() string {
	return fmt.Sprintf("[Provider: %s, Model: %s]",
		client.Provider, client.Model)
}

// BaseClient returns the underlying *Client (satisfies ClientEmbedder interface).
func (c *Client) BaseClient() *Client { return c }

// ApplyMaxTokens overrides the completion cap on a concrete client for the
// duration of one call scope and returns a restore func (LONDON-FORENSICS F1a,
// 2026-08-28): planner reads get a bigger budget (AI_PLAN_MAX_TOKENS) without
// permanently changing the shared executor client's cap. Follows the existing
// ApplyThinking precedent (per-call mutation of the shared client).
func ApplyMaxTokens(c AIClient, tokens int) func() {
	bc, ok := c.(interface{ BaseClient() *Client })
	if !ok || tokens <= 0 {
		return func() {}
	}
	cl := bc.BaseClient()
	prev := cl.MaxTokens
	cl.MaxTokens = tokens
	return func() { cl.MaxTokens = prev }
}

// LastFinishReason returns the most recent response's finish_reason for a
// client ("" when the client never completed a call). Read-only helper for
// the planner's truncation-aware diagnostics.
func LastFinishReason(c AIClient) string {
	bc, ok := c.(interface{ BaseClient() *Client })
	if !ok {
		return ""
	}
	if v := bc.BaseClient().lastFinishReason.Load(); v != nil {
		s, _ := v.(string)
		return s
	}
	return ""
}

// IsRetryableError determines if error is retryable (network errors, timeouts, etc.)
func (client *Client) IsRetryableError(err error) bool {
	errStr := err.Error()
	// Network errors, timeouts, EOF, etc. can be retried
	for _, retryable := range client.Cfg.RetryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}
	return false
}

// ============================================================
// Builder Pattern API (Advanced Features)
// ============================================================

// CallWithRequest calls AI API using Request object (supports advanced features)
func (client *Client) CallWithRequest(req *Request) (string, error) {
	if client.APIKey == "" {
		return "", fmt.Errorf("AI API key not set, please call SetAPIKey first")
	}

	// If Model is not set in Request, use Client's Model
	if req.Model == "" {
		req.Model = client.Model
	}

	// Fixed retry flow
	var lastErr error
	maxRetries := client.Cfg.MaxRetries

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			client.Log.Warnf("⚠️  AI API call failed, retrying (%d/%d)...", attempt, maxRetries)
		}

		// Call single request
		result, err := client.callWithRequest(req)
		if err == nil {
			if attempt > 1 {
				client.Log.Infof("✓ AI API retry succeeded")
			}
			return result, nil
		}

		lastErr = err
		// Check if error is retryable
		if !client.Hooks.IsRetryableError(err) {
			return "", err
		}

		// Wait before retry
		if attempt < maxRetries {
			waitTime := client.Cfg.RetryWaitBase * time.Duration(attempt)
			client.Log.Infof("⏳ Waiting %v before retry...", waitTime)
			if err := sleepWithContext(contextFromRequest(req), waitTime); err != nil {
				return "", err
			}
		}
	}

	return "", fmt.Errorf("still failed after %d retries: %w", maxRetries, lastErr)
}

// CallWithRequestFull calls the AI API and returns both text content and tool calls.
func (client *Client) CallWithRequestFull(req *Request) (*LLMResponse, error) {
	if client.APIKey == "" {
		return nil, fmt.Errorf("AI API key not set, please call SetAPIKey first")
	}
	if req.Model == "" {
		req.Model = client.Model
	}

	var lastErr error
	maxRetries := client.Cfg.MaxRetries
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			client.Log.Warnf("⚠️  AI API call failed, retrying (%d/%d)...", attempt, maxRetries)
		}
		result, err := client.callWithRequestFull(req)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !client.Hooks.IsRetryableError(err) {
			return nil, err
		}
		if attempt < maxRetries {
			waitTime := client.Cfg.RetryWaitBase * time.Duration(attempt)
			if err := sleepWithContext(contextFromRequest(req), waitTime); err != nil {
				return nil, err
			}
		}
	}
	return nil, fmt.Errorf("still failed after %d retries: %w", maxRetries, lastErr)
}

// callWithRequestFull single call that returns LLMResponse (content + tool calls).
func (client *Client) callWithRequestFull(req *Request) (*LLMResponse, error) {
	client.Log.Infof("📡 [%s] Request AI Server (full): BaseURL: %s", client.String(), client.BaseURL)

	requestBody := client.Hooks.BuildRequestBodyFromRequest(req)
	jsonData, err := client.Hooks.MarshalRequestBody(requestBody)
	if err != nil {
		return nil, err
	}

	url := client.Hooks.BuildUrl()
	httpReq, err := client.buildHTTPRequestWithContext(contextFromRequest(req), url, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := readWithTTFB(resp.Body, &client.lastTTFBMs)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error (status %d): %s", resp.StatusCode, string(body))
	}

	return client.Hooks.ParseMCPResponseFull(body)
}

// callWithRequest single AI API call (using Request object)
func (client *Client) callWithRequest(req *Request) (string, error) {
	// Print current AI configuration
	client.Log.Infof("📡 [%s] Request AI Server with Builder: BaseURL: %s", client.String(), client.BaseURL)
	client.Log.Debugf("[%s] Messages count: %d", client.String(), len(req.Messages))

	requestBody := client.Hooks.BuildRequestBodyFromRequest(req)

	jsonData, err := client.Hooks.MarshalRequestBody(requestBody)
	if err != nil {
		return "", err
	}

	url := client.Hooks.BuildUrl()
	client.Log.Infof("📡 [MCP %s] Request URL: %s", client.String(), url)

	httpReq, err := client.buildHTTPRequestWithContext(contextFromRequest(req), url, jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := readWithTTFB(resp.Body, &client.lastTTFBMs)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned error (status %d): %s", resp.StatusCode, string(body))
	}

	result, err := client.Hooks.ParseMCPResponse(body)
	if err != nil {
		return "", fmt.Errorf("fail to parse AI server response: %w", err)
	}

	return result, nil
}

// BuildRequestBodyFromRequest builds request body from Request object
func (client *Client) BuildRequestBodyFromRequest(req *Request) map[string]any {
	// Convert Message to API format — must use map[string]any to support
	// tool-call messages (tool_calls, tool_call_id fields).
	messages := make([]map[string]any, 0, len(req.Messages))
	for _, msg := range req.Messages {
		m := map[string]any{"role": msg.Role}
		if len(msg.ToolCalls) > 0 {
			// Assistant message that contains tool invocations.
			// content must be null/omitted for OpenAI compatibility.
			m["tool_calls"] = msg.ToolCalls
		} else if msg.ToolCallID != "" {
			// Tool result message (role="tool").
			m["tool_call_id"] = msg.ToolCallID
			m["content"] = msg.Content
		} else {
			m["content"] = msg.Content
		}
		// DeepSeek thinking models require reasoning_content to be echoed back
		// in multi-turn conversations when present in assistant messages.
		if msg.ReasoningContent != "" {
			m["reasoning_content"] = msg.ReasoningContent
		}
		messages = append(messages, m)
	}

	// Guard: truncate messages if they would exceed the model's context window
	maxOut := client.MaxTokens
	if req.MaxTokens != nil {
		maxOut = *req.MaxTokens
	}
	if client.Cfg.MaxContext > 0 {
		truncated, removed := truncateMessagesAny(messages, client.Cfg.MaxContext, maxOut)
		if removed > 0 {
			client.Log.Warnf("⚠️  [%s] Context guard: truncated %d oldest messages to fit within %d token limit",
				client.String(), removed, client.Cfg.MaxContext)
			messages = truncated
		}
	}

	// Build basic request body
	requestBody := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
	}

	// Add optional parameters (only add non-nil parameters)
	if req.Temperature != nil {
		requestBody["temperature"] = *req.Temperature
	} else {
		// If not set in Request, use Client's configuration
		requestBody["temperature"] = client.Cfg.Temperature
	}

	// OpenAI newer models use max_completion_tokens instead of max_tokens
	tokenKey := "max_tokens"
	if client.Provider == ProviderOpenAI {
		tokenKey = "max_completion_tokens"
	}
	if req.MaxTokens != nil {
		requestBody[tokenKey] = *req.MaxTokens
	} else {
		// If not set in Request, use Client's MaxTokens
		requestBody[tokenKey] = client.MaxTokens
	}

	if req.TopP != nil {
		requestBody["top_p"] = *req.TopP
	} else if client.Cfg != nil && client.Cfg.TopP > 0 {
		// Same fallback temperature and max_tokens already have on this path:
		// the Request builder used to DROP the configured AI_TOP_P entirely
		// (class 7 — the Call path honored it, this one didn't).
		requestBody["top_p"] = client.Cfg.TopP
	}

	if req.FrequencyPenalty != nil {
		requestBody["frequency_penalty"] = *req.FrequencyPenalty
	}

	if req.PresencePenalty != nil {
		requestBody["presence_penalty"] = *req.PresencePenalty
	}

	if len(req.Stop) > 0 {
		requestBody["stop"] = req.Stop
	}

	if len(req.Tools) > 0 {
		requestBody["tools"] = req.Tools
	}

	if req.ToolChoice != "" {
		requestBody["tool_choice"] = req.ToolChoice
	}

	if req.Stream {
		requestBody["stream"] = true
	}

	if client.Provider == ProviderDeepSeek {
		applyDeepSeekThinkingDefaults(requestBody, client.Cfg)
	}

	return requestBody
}

// applyDeepSeekThinkingDefaults injects the DeepSeek thinking-mode parameters
// into the request body for deepseek providers. Docs:
// https://api-docs.deepseek.com/guides/thinking_mode — thinking {type: enabled/
// disabled} + reasoning_effort low|high|max (max is the true maximum; medium/
// xhigh map to high). Empty cfg values omit the key so a per-request override
// stays possible. These params only take effect for deepseek models.
func applyDeepSeekThinkingDefaults(body map[string]any, cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.ThinkingMode != "" {
		body["thinking"] = map[string]any{"type": cfg.ThinkingMode}
	}
	if cfg.ReasoningEffort != "" {
		body["reasoning_effort"] = cfg.ReasoningEffort
	}
}

// SetThinking overrides the env-default DeepSeek thinking knobs with per-model
// values (4.5 API auto max). Empty keeps the env-derived default.
func (c *Client) SetThinking(mode, effort string) {
	if c == nil {
		return
	}
	if mode != "" {
		c.Cfg.ThinkingMode = mode
	}
	if effort != "" {
		c.Cfg.ReasoningEffort = effort
	}
}

// ValidateThinkingKnobs whitelists the two DeepSeek fields; empty = inherit.
func ValidateThinkingKnobs(mode, effort string) error {
	switch mode {
	case "", "enabled", "disabled":
	default:
		return fmt.Errorf("thinking_mode must be enabled|disabled (got %q)", mode)
	}
	switch effort {
	case "", "low", "high", "max":
	default:
		return fmt.Errorf("reasoning_effort must be low|high|max (got %q)", effort)
	}
	return nil
}

// CallWithRequestStream streams the LLM response via SSE (Server-Sent Events).
// onChunk is called with the full accumulated text so far after each received chunk.
// Returns the complete final text when the stream ends.
//
// Idle timeout: if no chunk arrives for AI_STREAM_IDLE_TIMEOUT_SECS (default
// 180s — generous because reasoning models think before the first token) the
// stream is cancelled automatically. The planner path passes its own ~30s idle
// via CallWithRequestStreamIdle (split deadlines, planner-speed wave 4.2).
func (client *Client) CallWithRequestStream(req *Request, onChunk func(string)) (string, error) {
	idle := 180 * time.Second
	if v := os.Getenv("AI_STREAM_IDLE_TIMEOUT_SECS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			idle = time.Duration(n) * time.Second
		}
	}
	return client.CallWithRequestStreamIdle(req, onChunk, idle)
}

// CallWithRequestStreamIdle is CallWithRequestStream with an explicit
// idle-chunk deadline and NO planner total deadline — the whole-request ceiling
// stays http.Client.Timeout (legacy callers: agent paths, tests). The planner
// uses CallWithRequestStreamDeadlines (class 37).
func (client *Client) CallWithRequestStreamIdle(req *Request, onChunk func(string), idle time.Duration) (string, error) {
	return client.CallWithRequestStreamDeadlines(req, onChunk, idle, 0)
}

// CallWithRequestStreamDeadlines (class 37, 2026-09-01) is the split-deadline
// stream call with an EXPLICIT whole-call ceiling: `idle` kills a silent
// stream, `total` kills a live-but-endless one. When total > 0 the request runs
// on a shallow copy of the HTTP client with Timeout=0 (same Transport, same
// pool) so http.Client.Timeout — the executor's 600s ceiling — no longer bounds
// the body read: it killed 11 of 80 live max-reasoning planner streams at
// exactly 600.0s between 2026-08-30 and 2026-09-01 while the speed-wave
// comments claimed "a live-but-slow stream is never killed". total <= 0 keeps
// the legacy behaviour (http.Client.Timeout applies). Every failure is
// classified (idle_deadline / total_deadline / client_timeout / transport /
// http_status) on the ai_call line via context.Cause + classifyAIError.
func (client *Client) CallWithRequestStreamDeadlines(req *Request, onChunk func(string), idle, total time.Duration) (string, error) {
	if client.APIKey == "" {
		return "", fmt.Errorf("AI API key not set")
	}
	if req.Model == "" {
		req.Model = client.Model
	}
	req.Stream = true
	client.resetCallTelemetry()

	requestBody := client.Hooks.BuildRequestBodyFromRequest(req)
	jsonData, err := client.Hooks.MarshalRequestBody(requestBody)
	if err != nil {
		return "", err
	}

	url := client.Hooks.BuildUrl()
	if total > 0 {
		client.Log.Infof("📡 [MCP %s] Request URL (stream idle=%ds total=%ds): %s", client.String(), int(idle.Seconds()), int(total.Seconds()), url)
	} else {
		client.Log.Infof("📡 [MCP %s] Request URL (stream idle=%ds): %s", client.String(), int(idle.Seconds()), url)
	}

	if idle <= 0 {
		idle = 180 * time.Second
	}
	parent := contextFromRequest(req)
	hc := client.HTTPClient
	if total > 0 {
		var cancelTotal context.CancelFunc
		parent, cancelTotal = context.WithTimeoutCause(parent, total, ErrStreamTotalDeadline)
		defer cancelTotal()
		if hc != nil {
			// A shallow copy shares Transport/CheckRedirect/Jar; only the
			// whole-request Timeout is lifted for THIS call. The shared client
			// is never mutated — the executor keeps its ceiling.
			cp := *hc
			cp.Timeout = 0
			hc = &cp
		}
	}
	httpReq, err := client.buildHTTPRequestWithContext(parent, url, jsonData)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithCancelCause(parent)
	defer cancel(nil)
	resetCh := make(chan struct{}, 1)
	callStart := time.Now()
	go func() {
		t := time.NewTimer(idle)
		defer t.Stop()
		lastLine := callStart // per CALL: a fresh timer + fresh clock each call (never carried across retries)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				// CLASS 41 M2 (2026-09-02): the watchdog LOGS when it fires, with
				// the measured gap, so "0 idle kills" is evidence, not silence.
				// It measures time since the last SSE LINE scanned (bufio line,
				// including blank/keep-alive lines), not since the last byte.
				if client.Log != nil {
					client.Log.Warnf("⏱ stream idle watchdog FIRED: %.1fs since last SSE line (idle=%ds, call age %.1fs) — cancelling the stream (class=idle_deadline)",
						time.Since(lastLine).Seconds(), int(idle.Seconds()), time.Since(callStart).Seconds())
				}
				cancel(ErrStreamIdleDeadline) // idle timeout: kill the connection
				return
			case <-resetCh:
				// received a line — reset the idle timer
				lastLine = time.Now()
				if !t.Stop() {
					select {
					case <-t.C:
					default:
					}
				}
				t.Reset(idle)
			}
		}
	}()

	httpReq = httpReq.WithContext(ctx)
	reqStart := time.Now()
	resp, err := hc.Do(httpReq)
	if err != nil {
		return "", wrapStreamDeadlineErr(ctx, fmt.Errorf("streaming request failed: %w", err), idle, total)
	}
	defer resp.Body.Close()
	client.lastHTTPStatus.Store(int64(resp.StatusCode))
	client.lastRequestID.Store(requestIDFrom(resp.Header))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	sr, err := ParseSSEStreamFull(resp.Body, onChunk, reqStart, func() {
		select {
		case resetCh <- struct{}{}:
		default:
		}
	})
	if sr != nil {
		client.lastTTFBMs.Store(sr.TTFBMs)
		client.lastReasoningChars.Store(int64(sr.ReasoningChars))
		if sr.FinishReason != "" {
			client.lastFinishReason.Store(sr.FinishReason)
		}
		if client.Log != nil {
			pt, ct := 0, 0
			if sr.Usage != nil {
				pt, ct = sr.Usage.PromptTokens, sr.Usage.CompletionTokens
			}
			client.Log.Infof("📊 AI call complete (stream): completion=%d prompt=%d finish_reason=%s reasoning_chars=%d ttfb_ms=%d wall_ms=%d",
				ct, pt, sr.FinishReason, sr.ReasoningChars, sr.TTFBMs, time.Since(reqStart).Milliseconds())
		}
	}
	if sr != nil && sr.Usage != nil {
		ReportStreamUsage(sr.Usage, client.Provider, client.Model)
	}
	if err != nil {
		return "", wrapStreamDeadlineErr(ctx, err, idle, total)
	}
	if sr == nil {
		return "", fmt.Errorf("stream produced no result")
	}
	return sr.Text, nil
}

// wrapStreamDeadlineErr names WHICH of the two split deadlines cancelled the
// stream (context.Cause carries the sentinel; a parent total-deadline cause
// propagates to the idle child). Every other error passes through untouched so
// the transport/retry classification keeps working.
func wrapStreamDeadlineErr(ctx context.Context, err error, idle, total time.Duration) error {
	// Go ≥1.21 net/http surfaces context.Cause (our sentinel) as the read
	// error, so ctx.Err() is appended explicitly: the legacy
	// "context canceled" / "context deadline exceeded" greps keep working.
	cause := context.Cause(ctx)
	switch {
	case errors.Is(cause, ErrStreamTotalDeadline):
		return fmt.Errorf("%w (total %v, stream was live, %v): %v", ErrStreamTotalDeadline, total, ctx.Err(), err)
	case errors.Is(cause, ErrStreamIdleDeadline):
		return fmt.Errorf("%w (idle %v of silence, %v): %v", ErrStreamIdleDeadline, idle, ctx.Err(), err)
	}
	return err
}

// CallWithRequestStreamRetry wraps the streaming call in the same fixed retry
// flow CallWithMessages uses (MaxRetries + exponential backoff + retryable
// classification). Phase 4.4 — the transport-reset class must still retry.
func (client *Client) CallWithRequestStreamRetry(req *Request, onChunk func(string), idle time.Duration) (string, error) {
	return client.CallWithRequestStreamRetryDeadlines(req, onChunk, idle, 0)
}

// CallWithRequestStreamRetryDeadlines (class 37) is CallWithRequestStreamRetry
// with the planner's whole-call ceiling (see CallWithRequestStreamDeadlines).
// Deadline kills (idle / total / client_timeout) are NOT retried here — none of
// them match the retryable transport tokens — so the planner loop, not the
// client, owns that retry (3 attempts); transport resets still retry in place.
// Worst case per planner attempt therefore stays 1 + (MaxRetries−1) calls on
// resets only, never MaxRetries × total.
func (client *Client) CallWithRequestStreamRetryDeadlines(req *Request, onChunk func(string), idle, total time.Duration) (string, error) {
	if client.APIKey == "" {
		return "", fmt.Errorf("AI API key not set")
	}
	var lastErr error
	// CLASS 41 (2026-09-02): the stream path counts CALLS via StreamTries and
	// waits an EXPONENTIAL schedule (StreamBackoff, default 2s → 15s → 45s)
	// instead of RetryWaitBase×attempt. MaxRetries (AI_MAX_RETRIES) still
	// governs the non-stream paths only.
	maxRetries := client.Cfg.StreamTries
	if maxRetries < 1 {
		maxRetries = client.Cfg.MaxRetries
	}
	if maxRetries < 1 {
		maxRetries = 1
	}
	sched := client.Cfg.StreamBackoff
	if len(sched) == 0 {
		sched = StreamRetryBackoffSchedule()
	}
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			client.Log.Warnf("⚠️  AI API stream failed, retrying (%d/%d)...", attempt, maxRetries)
		}
		start := time.Now()
		result, err := client.CallWithRequestStreamDeadlines(req, onChunk, idle, total)
		client.logAICall(start, err, attempt)
		if err == nil {
			if attempt > 1 {
				client.Log.Infof("✓ AI API stream retry succeeded")
			}
			return result, nil
		}
		lastErr = err
		if !client.Hooks.IsRetryableError(err) {
			return "", err
		}
		if attempt < maxRetries {
			waitTime := streamBackoffFor(attempt, sched) // class 41: 2s → 15s → 45s
			client.Log.Infof("⏳ Waiting %v before retry...", waitTime)
			if err := sleepWithContext(contextFromRequest(req), waitTime); err != nil {
				return "", err
			}
		}
	}
	return "", fmt.Errorf("still failed after %d retries: %w", maxRetries, lastErr)
}

// SSEStreamResult is ParseSSEStreamFull's output: accumulated text, reasoning
// chars, finish_reason, usage, and time-to-first-byte.
type SSEStreamResult struct {
	Text           string
	ReasoningChars int
	FinishReason   string
	Usage          *TokenUsage
	TTFBMs         int64
}

// ParseSSEStream reads an SSE response body, accumulates text deltas,
// and calls onChunk with the full accumulated text after each chunk.
// If onLine is non-nil, it is called after each raw SSE line is scanned
// (useful for resetting idle-timeout watchdogs).
// Returns the complete accumulated text and any parsed token usage (nil if absent).
func ParseSSEStream(body io.Reader, onChunk func(string), onLine func()) (string, *TokenUsage, error) {
	sr, err := ParseSSEStreamFull(body, onChunk, time.Now(), onLine)
	if sr == nil {
		return "", nil, err
	}
	return sr.Text, sr.Usage, err
}

// ParseSSEStreamFull is the planner-speed-wave (2026-08-31) extension:
// additionally captures reasoning_content chars, finish_reason, and
// time-to-first-byte (the T4 evidence the latency autopsy was missing).
// `start` anchors the ttfb measurement — pass the request-sent time so the
// queue (Do → first chunk) is included, not just the body-read latency.
func ParseSSEStreamFull(body io.Reader, onChunk func(string), start time.Time, onLine func()) (*SSEStreamResult, error) {
	var accumulated strings.Builder
	var reasoning strings.Builder
	var usage *TokenUsage
	var finish string
	scanner := bufio.NewScanner(body)
	var ttfbMs int64
	haveFirst := false

	for scanner.Scan() {
		if onLine != nil {
			onLine()
		}
		if !haveFirst {
			haveFirst = true
			ttfbMs = time.Since(start).Milliseconds()
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage,omitempty"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // skip malformed chunks
		}

		if chunk.Usage != nil && chunk.Usage.TotalTokens > 0 {
			usage = &TokenUsage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		if chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != "" {
			finish = *chunk.Choices[0].FinishReason
		}
		delta := chunk.Choices[0].Delta
		if delta.ReasoningContent != "" {
			reasoning.WriteString(delta.ReasoningContent)
		}
		if delta.Content == "" {
			continue
		}

		accumulated.WriteString(delta.Content)
		if onChunk != nil {
			onChunk(accumulated.String())
		}
	}

	if err := scanner.Err(); err != nil {
		return &SSEStreamResult{
			Text: accumulated.String(), Usage: usage, FinishReason: finish,
			ReasoningChars: reasoning.Len(), TTFBMs: ttfbMs,
		}, fmt.Errorf("stream interrupted: %w", err)
	}

	return &SSEStreamResult{
		Text: accumulated.String(), Usage: usage, FinishReason: finish,
		ReasoningChars: reasoning.Len(), TTFBMs: ttfbMs,
	}, nil
}

// ReportStreamUsage fires TokenUsageCallback with the given usage, provider, and model.
// No-op if usage is nil or callback is unset.
func ReportStreamUsage(usage *TokenUsage, provider, model string) {
	if usage == nil || TokenUsageCallback == nil || usage.TotalTokens <= 0 {
		return
	}
	TokenUsageCallback(TokenUsage{
		Provider:         provider,
		Model:            model,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	})
}
