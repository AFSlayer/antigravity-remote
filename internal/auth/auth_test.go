package auth

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testAuth(t *testing.T, password string) *Authenticator {
	t.Helper()

	dir := t.TempDir()
	creds, _, err := LoadOrCreateCredentials(filepath.Join(dir, "credentials.json"), password)
	if err != nil {
		t.Fatalf("credentials: %v", err)
	}
	sessions, err := NewSessionStore(filepath.Join(dir, "sessions.json"), time.Hour)
	if err != nil {
		t.Fatalf("sessions: %v", err)
	}

	return New(Options{
		Credentials: creds,
		Sessions:    sessions,
		LoginPage: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("login page"))
		},
		IsPublic: func(path string) bool { return path == "/favicon.ico" },
	})
}

func protected() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("secret"))
	})
}

func TestPasswordVerify(t *testing.T) {
	dir := t.TempDir()
	creds, generated, err := LoadOrCreateCredentials(filepath.Join(dir, "c.json"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(generated) < 12 {
		t.Fatalf("generated password too short: %q", generated)
	}
	if !creds.Verify(generated) {
		t.Error("generated password should verify")
	}
	if creds.Verify(generated + "x") {
		t.Error("wrong password must not verify")
	}
	if strings.Contains(creds.Hash, generated) {
		t.Error("stored hash must not contain the plaintext")
	}
}

func TestCredentialsPersistAndRotate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.json")

	creds, _, err := LoadOrCreateCredentials(path, "first-password")
	if err != nil {
		t.Fatal(err)
	}
	if err := creds.Set("second-password"); err != nil {
		t.Fatal(err)
	}

	reloaded, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reloaded.Verify("second-password") {
		t.Error("rotated password should verify after reload")
	}
	if reloaded.Verify("first-password") {
		t.Error("old password must stop working")
	}
}

func TestShortPasswordRejected(t *testing.T) {
	if _, _, err := LoadOrCreateCredentials(filepath.Join(t.TempDir(), "c.json"), "short"); err == nil {
		t.Error("expected short password to be rejected")
	}
}

func TestSessionLifecycle(t *testing.T) {
	store, err := NewSessionStore(filepath.Join(t.TempDir(), "s.json"), time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := store.Create("test-agent", "1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if !store.Validate(raw) {
		t.Fatal("fresh session should validate")
	}
	if store.Validate(raw + "x") {
		t.Error("tampered token must not validate")
	}

	for _, s := range store.List() {
		if s.Key == raw {
			t.Error("raw token must not be stored on disk")
		}
	}

	if err := store.Revoke(raw); err != nil {
		t.Fatal(err)
	}
	if store.Validate(raw) {
		t.Error("revoked session must not validate")
	}
}

func TestSessionExpiry(t *testing.T) {
	store, err := NewSessionStore(filepath.Join(t.TempDir(), "s.json"), -time.Second)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := store.Create("", "")
	if err != nil {
		t.Fatal(err)
	}
	if store.Validate(raw) {
		t.Error("expired session must not validate")
	}
}

func TestRevokeAllInvalidatesEveryDevice(t *testing.T) {
	store, _ := NewSessionStore(filepath.Join(t.TempDir(), "s.json"), time.Hour)

	a, _ := store.Create("phone", "1.1.1.1")
	b, _ := store.Create("laptop", "2.2.2.2")

	n, err := store.RevokeAll()
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("want 2 revoked, got %d", n)
	}
	if store.Validate(a) || store.Validate(b) {
		t.Error("all sessions must be invalid after RevokeAll")
	}
}

func TestSessionsSurviveRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.json")

	store, _ := NewSessionStore(path, time.Hour)
	raw, _ := store.Create("phone", "1.1.1.1")

	reopened, err := NewSessionStore(path, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !reopened.Validate(raw) {
		t.Error("sessions should survive a restart")
	}
}

func TestLimiterLocksOutAfterMaxFailures(t *testing.T) {
	l := NewLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if ok, _ := l.Allow("ip"); !ok {
			t.Fatalf("attempt %d should be allowed", i)
		}
		l.Fail("ip")
	}

	ok, retry := l.Allow("ip")
	if ok {
		t.Fatal("expected lockout after reaching max failures")
	}
	if retry <= 0 {
		t.Error("lockout should report a retry-after duration")
	}
}

func TestLimiterBackoffGrows(t *testing.T) {
	l := NewLimiter(1, time.Minute)

	l.Fail("ip")
	_, first := l.Allow("ip")

	l.mu.Lock()
	l.entries["ip"].lockedUntil = time.Now().Add(-time.Second)
	l.mu.Unlock()

	l.Fail("ip")
	_, second := l.Allow("ip")

	if second <= first {
		t.Errorf("backoff should grow: first=%s second=%s", first, second)
	}
}

func TestLimiterResetClearsLockout(t *testing.T) {
	l := NewLimiter(1, time.Minute)
	l.Fail("ip")
	l.Reset("ip")

	if ok, _ := l.Allow("ip"); !ok {
		t.Error("reset should clear the lockout")
	}
}

func mustCIDR(t *testing.T, s string) *net.IPNet {
	t.Helper()
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestClientIPIgnoresForwardedHeaderFromUntrustedPeer(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.9:1234"
	r.Header.Set("X-Forwarded-For", "1.1.1.1")

	if got := ClientIP(r, nil); got != "203.0.113.9" {
		t.Errorf("want direct peer when no proxies are trusted, got %s", got)
	}
}

func TestClientIPUsesForwardedHeaderFromTrustedProxy(t *testing.T) {
	trusted := []*net.IPNet{mustCIDR(t, "127.0.0.1/32")}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:5555"
	r.Header.Set("X-Forwarded-For", "203.0.113.7, 127.0.0.1")

	if got := ClientIP(r, trusted); got != "203.0.113.7" {
		t.Errorf("want the client behind the proxy, got %s", got)
	}
}

func TestClientIPRejectsSpoofedChain(t *testing.T) {
	trusted := []*net.IPNet{mustCIDR(t, "127.0.0.1/32")}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:5555"
	r.Header.Set("X-Forwarded-For", "10.0.0.1, 203.0.113.7")

	if got := ClientIP(r, trusted); got != "203.0.113.7" {
		t.Errorf("want the rightmost untrusted hop, got %s", got)
	}
}

func TestRequestIsSecureOnlyTrustsProxyHeader(t *testing.T) {
	trusted := []*net.IPNet{mustCIDR(t, "127.0.0.1/32")}

	spoofed := httptest.NewRequest(http.MethodGet, "/", nil)
	spoofed.RemoteAddr = "203.0.113.9:1"
	spoofed.Header.Set("X-Forwarded-Proto", "https")
	if RequestIsSecure(spoofed, trusted) {
		t.Error("must not trust X-Forwarded-Proto from an untrusted peer")
	}

	viaProxy := httptest.NewRequest(http.MethodGet, "/", nil)
	viaProxy.RemoteAddr = "127.0.0.1:1"
	viaProxy.Header.Set("X-Forwarded-Proto", "https")
	if !RequestIsSecure(viaProxy, trusted) {
		t.Error("should trust X-Forwarded-Proto from a trusted proxy")
	}
}

func TestMiddlewareBlocksUnauthenticated(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("want redirect for html navigation, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != LoginPath {
		t.Errorf("want redirect to %s, got %s", LoginPath, loc)
	}
}

func TestMiddlewareReturnsJSONForAPICalls(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	r := httptest.NewRequest(http.MethodPost, "/exa.language_server_pb.Service/Something", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401 for API call, got %d", w.Code)
	}
}

func TestMiddlewareBlocksUnauthenticatedWebsocket(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	r := httptest.NewRequest(http.MethodGet, "/connect-websocket", nil)
	r.Header.Set("Upgrade", "websocket")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401 for websocket upgrade, got %d", w.Code)
	}
}

func TestPublicPathsBypassAuth(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/favicon.ico", nil))

	if w.Code != http.StatusOK {
		t.Errorf("public assets should bypass auth, got %d", w.Code)
	}
}

func TestLoginFlowSetsHardenedCookie(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, LoginAPIPath, strings.NewReader(`{"password":"correct-horse"}`))
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("login should succeed, got %d: %s", w.Code, w.Body.String())
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("want one cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != CookieName || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie not hardened: %+v", c)
	}

	authed := httptest.NewRequest(http.MethodGet, "/", nil)
	authed.AddCookie(c)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, authed)

	if w2.Body.String() != "secret" {
		t.Errorf("authenticated request should reach the app, got %q", w2.Body.String())
	}
}

func TestLoginRateLimited(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	var last int
	for i := 0; i < loginMaxFailures+1; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, LoginAPIPath, strings.NewReader(`{"password":"nope"}`))
		r.RemoteAddr = "198.51.100.5:1000"
		h.ServeHTTP(w, r)
		last = w.Code
	}

	if last != http.StatusTooManyRequests {
		t.Errorf("want 429 after %d failures, got %d", loginMaxFailures, last)
	}
}

func TestCorrectPasswordRejectedWhileLockedOut(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	for i := 0; i < loginMaxFailures; i++ {
		r := httptest.NewRequest(http.MethodPost, LoginAPIPath, strings.NewReader(`{"password":"nope"}`))
		r.RemoteAddr = "198.51.100.6:1000"
		h.ServeHTTP(httptest.NewRecorder(), r)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, LoginAPIPath, strings.NewReader(`{"password":"correct-horse"}`))
	r.RemoteAddr = "198.51.100.6:1000"
	h.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("lockout must apply even to the correct password, got %d", w.Code)
	}
}

func TestEnrollmentTokenIsSingleUse(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	tok, err := a.NewEnrollmentToken()
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, LoginPath+"?t="+tok, nil))

	if w.Code != http.StatusFound || w.Header().Get("Location") != "/" {
		t.Fatalf("enrollment should log in and redirect home, got %d %s", w.Code, w.Header().Get("Location"))
	}
	if len(w.Result().Cookies()) != 1 {
		t.Fatal("enrollment should set a session cookie")
	}

	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, LoginPath+"?t="+tok, nil))

	if w2.Header().Get("Location") != LoginPath {
		t.Error("reused enrollment token must not grant a session")
	}
	if len(w2.Result().Cookies()) != 0 {
		t.Error("reused enrollment token must not set a cookie")
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, LoginAPIPath, strings.NewReader(`{"password":"correct-horse"}`)))
	cookie := w.Result().Cookies()[0]

	out := httptest.NewRecorder()
	logout := httptest.NewRequest(http.MethodPost, LogoutAPIPath, nil)
	logout.AddCookie(cookie)
	h.ServeHTTP(out, logout)

	if out.Code != http.StatusOK {
		t.Fatalf("logout should succeed, got %d", out.Code)
	}

	after := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(cookie)
	r.Header.Set("Accept", "text/html")
	h.ServeHTTP(after, r)

	if after.Code != http.StatusFound {
		t.Error("session should be invalid after logout")
	}
}

func TestLoginErrorDoesNotLeakWhetherPasswordExists(t *testing.T) {
	a := testAuth(t, "correct-horse")
	h := a.Middleware(protected())

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, LoginAPIPath, strings.NewReader(`{"password":""}`)))

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "Incorrect password." {
		t.Errorf("unexpected error message: %v", body["error"])
	}
}
