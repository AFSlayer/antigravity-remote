// Package auth guards the public listener with a password, server-side sessions
// and rate limiting.
//
// Reaching Antigravity means being able to read files and run commands on the
// host, so the defaults are deliberately strict: passwords are only ever stored
// hashed, session tokens are stored as hashes so a leaked file grants nothing,
// failed logins lock out with exponential backoff, and forwarded-header trust is
// opt-in. Administrative endpoints are not defined here at all — they live on a
// separate loopback-only listener so they cannot be reached from the network.
package auth

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Endpoints live under /__agy/ so they cannot collide with an Antigravity route
// now or after a future update.
const (
	CookieName    = "agy_session"
	LoginPath     = "/__agy/login"
	LoginAPIPath  = "/__agy/api/login"
	LogoutAPIPath = "/__agy/api/logout"

	loginWindow       = 5 * time.Minute
	loginMaxFailures  = 5
	globalMaxFailures = 60
	globalWindow      = time.Minute
	globalKey         = "*"

	// EnrollmentTTL bounds how long a QR code stays usable.
	EnrollmentTTL = 10 * time.Minute
)

// LoginPageFunc renders the password prompt.
type LoginPageFunc func(w http.ResponseWriter, r *http.Request)

// Options configures an Authenticator.
type Options struct {
	Credentials *Credentials
	Sessions    *SessionStore
	// Trusted lists proxy networks whose X-Forwarded-* headers may be believed.
	// Empty means trust nothing, which is the safe default.
	Trusted   []*net.IPNet
	LoginPage LoginPageFunc
	// IsPublic marks paths served without authentication, such as icons.
	IsPublic func(path string) bool
}

// Authenticator wraps a handler with password authentication.
type Authenticator struct {
	creds    *Credentials
	sessions *SessionStore
	trusted  []*net.IPNet
	perIP    *Limiter
	global   *Limiter
	login    LoginPageFunc
	isPublic func(string) bool

	mu     sync.Mutex
	enroll map[string]time.Time
}

// New builds an Authenticator from opts.
func New(opts Options) *Authenticator {
	return &Authenticator{
		creds:    opts.Credentials,
		sessions: opts.Sessions,
		trusted:  opts.Trusted,
		perIP:    NewLimiter(loginMaxFailures, loginWindow),
		global:   NewLimiter(globalMaxFailures, globalWindow),
		login:    opts.LoginPage,
		isPublic: opts.IsPublic,
		enroll:   map[string]time.Time{},
	}
}

// Sessions exposes the session store for the control panel and CLI.
func (a *Authenticator) Sessions() *SessionStore { return a.sessions }

// Credentials exposes the stored password hash for rotation.
func (a *Authenticator) Credentials() *Credentials { return a.creds }

// NewEnrollmentToken mints a single-use token that signs a device in without
// typing the password. It is what the QR code encodes, and it expires after
// EnrollmentTTL.
func (a *Authenticator) NewEnrollmentToken() (string, error) {
	tok, err := randomToken()
	if err != nil {
		return "", err
	}

	now := time.Now()
	a.mu.Lock()
	for t, exp := range a.enroll {
		if now.After(exp) {
			delete(a.enroll, t)
		}
	}
	a.enroll[hashToken(tok)] = now.Add(EnrollmentTTL)
	a.mu.Unlock()

	return tok, nil
}

func (a *Authenticator) consumeEnrollment(tok string) bool {
	if tok == "" {
		return false
	}
	key := hashToken(tok)

	a.mu.Lock()
	defer a.mu.Unlock()

	exp, ok := a.enroll[key]
	if !ok {
		return false
	}
	delete(a.enroll, key)
	return time.Now().Before(exp)
}

// Authenticated reports whether the request carries a valid session cookie.
func (a *Authenticator) Authenticated(r *http.Request) bool {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return false
	}
	return a.sessions.Validate(cookie.Value)
}

// Middleware serves the login endpoints and gates everything else behind a
// valid session.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if a.isPublic != nil && a.isPublic(path) {
			next.ServeHTTP(w, r)
			return
		}

		switch {
		case path == LoginPath:
			a.handleLoginPage(w, r)
			return
		case path == LoginAPIPath:
			a.handleLoginAPI(w, r)
			return
		case path == LogoutAPIPath:
			a.handleLogout(w, r)
			return
		}

		if a.Authenticated(r) {
			next.ServeHTTP(w, r)
			return
		}

		a.reject(w, r)
	})
}

func (a *Authenticator) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if tok := r.URL.Query().Get("t"); tok != "" {
		ip := ClientIP(r, a.trusted)

		if ok, retry := a.allow(ip); !ok {
			a.tooMany(w, retry)
			return
		}

		if a.consumeEnrollment(tok) {
			if err := a.startSession(w, r); err != nil {
				http.Error(w, "could not start session", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		a.perIP.Fail(ip)
		a.global.Fail(globalKey)
		http.Redirect(w, r, LoginPath, http.StatusFound)
		return
	}

	if a.Authenticated(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	a.login(w, r)
}

func (a *Authenticator) handleLoginAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "Method not allowed."})
		return
	}

	ip := ClientIP(r, a.trusted)
	if ok, retry := a.allow(ip); !ok {
		a.tooMany(w, retry)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		req.Password = r.FormValue("password")
	}

	if req.Password == "" || !a.creds.Verify(req.Password) {
		a.perIP.Fail(ip)
		a.global.Fail(globalKey)
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "Incorrect password."})
		return
	}

	a.perIP.Reset(ip)

	if err := a.startSession(w, r); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Could not start session."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (a *Authenticator) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false})
		return
	}

	if cookie, err := r.Cookie(CookieName); err == nil {
		_ = a.sessions.Revoke(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   RequestIsSecure(r, a.trusted),
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (a *Authenticator) startSession(w http.ResponseWriter, r *http.Request) error {
	raw, err := a.sessions.Create(r.UserAgent(), ClientIP(r, a.trusted))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    raw,
		Path:     "/",
		MaxAge:   int(a.sessions.ttl.Seconds()),
		HttpOnly: true,
		Secure:   RequestIsSecure(r, a.trusted),
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (a *Authenticator) allow(ip string) (bool, time.Duration) {
	if ok, retry := a.global.Allow(globalKey); !ok {
		return false, retry
	}
	return a.perIP.Allow(ip)
}

func (a *Authenticator) tooMany(w http.ResponseWriter, retry time.Duration) {
	w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retry.Seconds())))
	writeJSON(w, http.StatusTooManyRequests, map[string]any{
		"success": false,
		"error":   fmt.Sprintf("Too many attempts. Try again in %s.", retry),
	})
}

// reject answers an unauthenticated request in the shape its caller expects: a
// redirect for page loads, a bare status for API and websocket calls.
func (a *Authenticator) reject(w http.ResponseWriter, r *http.Request) {
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet || !strings.Contains(r.Header.Get("Accept"), "text/html") {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}

	http.Redirect(w, r, LoginPath, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
