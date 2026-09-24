package api

// PR #200 fold F6 (CTO 1790252194343, token-retire lens): a credential epoch
// in the FUTURE — the clock stepped back after a password change — refuses
// every NEW sign-in until the clock passes it: a token minted now is not
// strictly after the epoch (auth.RetiredBy). That stays fail-closed. The fix
// is observability: the refusal's log line says the epoch is N s ahead, that
// the clock stepped back, and that sign-in is refused until then — so the
// owner who cannot sign in is told why. A refusal by an epoch in the PAST
// keeps the plain category.
//
// authMiddleware (tokenRetirement) is pinned at the PRODUCTION router (canon
// 53). The credential guard (credentialActorRefusal) applies the same
// predicate but is unreachable for this case through the router —
// authMiddleware refuses first — so it is pinned by driving the production
// handler directly with the context authMiddleware would have set.

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"nofx/auth"

	"github.com/gin-gonic/gin"
)

var futureEpochLine = regexp.MustCompile(`token predates the account's last credential change — credential epoch is (\d+)s in the future — clock stepped back; sign-in refused until then`)

func setAdminUpdatedAt(t *testing.T, e *updEnv, at time.Time) {
	t.Helper()
	if err := e.st.GormDB().Exec(`UPDATE users SET updated_at = ? WHERE id = ?`, at.UTC(), updAdminID).Error; err != nil {
		t.Fatal(err)
	}
}

func assertAheadAbout(t *testing.T, where, logs string, want time.Duration) {
	t.Helper()
	m := futureEpochLine.FindStringSubmatch(logs)
	if m == nil {
		t.Fatalf("%s: no refusal line saying the credential epoch is in the future:\n%s", where, logs)
	}
	n, _ := strconv.Atoi(m[1])
	if w := int(want / time.Second); n > w || n < w-10 {
		t.Fatalf("%s: the line says the epoch is %ds ahead — want ~%ds", where, n, w)
	}
}

func TestFutureCredentialEpochRefusalSaysTheClockSteppedBack(t *testing.T) {
	logs := captureLogs(t)
	e := newUpdEnv(t)

	// A password change stamped an hour ahead of this clock.
	setAdminUpdatedAt(t, e, time.Now().Add(time.Hour))

	// A NEW sign-in: login still answers 200, its token is refused on use.
	tok, code := credLogin(t, e, updAdminEmail, updAdminPass)
	if code != http.StatusOK {
		t.Fatalf("login = %d", code)
	}
	if w := credCall(t, e, "GET", "/api/my-traders", tok, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("a token minted now, epoch an hour ahead: GET /api/my-traders = %d %s — want 401 (fail-closed, kept)", w.Code, w.Body.String())
	}
	all := logs()
	if !strings.Contains(all, "[auth] refused GET /api/my-traders from 127.0.0.1: token predates") {
		t.Fatalf("no authMiddleware refusal line:\n%s", all)
	}
	assertAheadAbout(t, "authMiddleware", all, time.Hour)

	// Control: an epoch in the PAST refuses an older token with the plain
	// category — no clock claim.
	setAdminUpdatedAt(t, e, time.Now().Add(-10*time.Minute))
	older := mintJWT(t, updAdminID, updAdminEmail, time.Now().Add(-20*time.Minute), time.Now().Add(time.Hour), updSecret)
	mark := len(logs())
	if w := credCall(t, e, "GET", "/api/config/resolved", older, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("control: a token older than a past epoch = %d, want 401", w.Code)
	}
	tail := logs()[mark:]
	if !strings.Contains(tail, "[auth] refused GET /api/config/resolved from 127.0.0.1: token predates the account's last credential change") {
		t.Fatalf("control: no plain refusal line:\n%s", tail)
	}
	if strings.Contains(tail, "in the future") || strings.Contains(tail, "clock stepped back") {
		t.Fatalf("control: an epoch in the PAST was reported as a stepped-back clock:\n%s", tail)
	}
}

func TestCredentialGuardFutureEpochRefusalSaysTheClockSteppedBack(t *testing.T) {
	logs := captureLogs(t)
	e := newUpdEnv(t)
	setAdminUpdatedAt(t, e, time.Now().Add(30*time.Minute))

	tok := mintJWT(t, updAdminID, updAdminEmail, time.Now().Add(-2*time.Second), time.Now().Add(time.Hour), updSecret)
	cl, err := auth.ValidateJWT(tok)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/user/password", strings.NewReader(`{"current_password":"`+updAdminPass+`","new_password":"guard-future-pass-1"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	setAuthContext(c, cl) // exactly what authMiddleware records
	e.s.handleChangePassword(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("credential guard, epoch 30 min ahead: handleChangePassword = %d %s — want 403", w.Code, w.Body.String())
	}
	all := logs()
	if !strings.Contains(all, "[credentials] refused") {
		t.Fatalf("no credential-guard refusal line:\n%s", all)
	}
	assertAheadAbout(t, "credential guard", all, 30*time.Minute)
}
