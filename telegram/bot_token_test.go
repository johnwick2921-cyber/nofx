package telegram

// M3 red-team H2 (CTO ruling 1790231205208): authMiddleware refuses, on every
// protected route, a token issued at or before the account's last credential
// change — the Telegram bot's own token included. The bot must therefore
// re-mint when its token would be refused, or the owner's password change
// silently kills the bot until a restart. Proven against the PRODUCTION
// server (api.NewServer + Server.Start on a loopback socket) with the token
// the production bot mints (agent.GenerateBotToken) and the predicate
// resolveBotUser calls (botTokenStale).

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"nofx/api"
	"nofx/auth"
	"nofx/manager"
	"nofx/store"
	"nofx/telegram/agent"
	"nofx/trader"

	"github.com/golang-jwt/jwt/v5"
)

const (
	btOwnerID    = "eeeeeeee-1111-2222-3333-444444444444"
	btOwnerEmail = "owner@example.test"
	btOwnerPass  = "owner-original-pass"
	btSecret     = "bot-token-test-secret-not-the-default-0123"
)

func btBoot(t *testing.T) (string, *store.Store) {
	t.Helper()
	dataDir := filepath.Join(t.TempDir(), "data")
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	st, err := store.New(filepath.Join(dataDir, "data.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Plan().Close(); _ = st.Close() })
	prevDir := trader.MaintenanceDataDir()
	trader.SetMaintenanceDataDir(dataDir)
	t.Cleanup(func() { trader.SetMaintenanceDataDir(prevDir) })
	prev := auth.JWTSecret
	auth.SetJWTSecret(btSecret)
	t.Cleanup(func() { auth.JWTSecret = prev })
	hash, err := auth.HashPassword(btOwnerPass)
	if err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-time.Hour).UTC()
	if err := st.User().Create(&store.User{ID: btOwnerID, Email: btOwnerEmail, PasswordHash: hash, CreatedAt: past, UpdatedAt: past}); err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	s := api.NewServer(manager.NewTraderManager(), st, nil, "127.0.0.1", port)
	go func() { _ = s.Start() }()
	t.Cleanup(func() { _ = s.Shutdown() })
	base := "http://127.0.0.1:" + strconv.Itoa(port)
	deadline := time.Now().Add(10 * time.Second)
	for {
		resp, err := http.Get(base + "/api/health")
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("production server never came up: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	return base, st
}

func btCall(t *testing.T, base, method, path, tok, body string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, base+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func btRow(t *testing.T, st *store.Store) store.User {
	t.Helper()
	u, err := st.User().GetByID(btOwnerID)
	if err != nil {
		t.Fatal(err)
	}
	return *u
}

func TestBotReMintsItsTokenWhenAPasswordChangeRetiresIt(t *testing.T) {
	base, st := btBoot(t)
	bot, err := agent.GenerateBotToken(btOwnerID)
	if err != nil {
		t.Fatal(err)
	}
	if botTokenStale(bot, btRow(t, st)) {
		t.Fatal("a freshly minted bot token reads as stale")
	}
	if code, _ := btCall(t, base, "GET", "/api/my-traders", bot, ""); code != http.StatusOK {
		t.Fatalf("control: the bot token on GET /api/my-traders = %d", code)
	}

	// The owner changes the password through the web flow.
	code, body := btCall(t, base, "POST", "/api/login", "", `{"email":"`+btOwnerEmail+`","password":"`+btOwnerPass+`"}`)
	if code != http.StatusOK {
		t.Fatalf("owner login = %d", code)
	}
	var login struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal([]byte(body), &login)
	if code, body := btCall(t, base, "PUT", "/api/user/password", login.Token, `{"current_password":"`+btOwnerPass+`","new_password":"owner-rotated-pass-1"}`); code != http.StatusOK {
		t.Fatalf("the owner's password change = %d %s", code, body)
	}

	// The API now refuses the bot's old token, and the bot's predicate says so.
	if code, _ := btCall(t, base, "GET", "/api/my-traders", bot, ""); code != http.StatusUnauthorized {
		t.Fatalf("the pre-change bot token on GET /api/my-traders = %d — want 401 (H2)", code)
	}
	if !botTokenStale(bot, btRow(t, st)) {
		t.Fatal("botTokenStale = false for a token the API refuses — the bot would go dark until a restart")
	}

	// Re-minted after the change's second: usable again, and not stale.
	now := time.Now()
	time.Sleep(now.Truncate(time.Second).Add(time.Second + 20*time.Millisecond).Sub(now))
	again, err := agent.GenerateBotToken(btOwnerID)
	if err != nil {
		t.Fatal(err)
	}
	if botTokenStale(again, btRow(t, st)) {
		t.Fatal("the re-minted bot token reads as stale")
	}
	if code, _ := btCall(t, base, "GET", "/api/my-traders", again, ""); code != http.StatusOK {
		t.Fatalf("the re-minted bot token on GET /api/my-traders = %d — want 200", code)
	}
}

func TestBotTokenStaleOnAnythingTheAPIWouldRefuse(t *testing.T) {
	_, st := btBoot(t)
	u := btRow(t, st)
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{UserID: btOwnerID, Email: auth.BotInternalEmail, Scope: auth.ScopeTelegram,
		RegisteredClaims: jwt.RegisteredClaims{IssuedAt: jwt.NewNumericDate(time.Now().Add(-25 * time.Hour)), ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))}}).SignedString([]byte(btSecret))
	if err != nil {
		t.Fatal(err)
	}
	other, err := agent.GenerateBotToken("ffffffff-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatal(err)
	}
	for name, tok := range map[string]string{"empty": "", "expired": expired, "another user's": other, "garbage": "not.a.jwt"} {
		if !botTokenStale(tok, u) {
			t.Errorf("%s token: botTokenStale = false, want true", name)
		}
	}
}
