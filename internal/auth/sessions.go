package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

const maxSessions = 200

// Session is one signed-in device. Key is the SHA-256 of the cookie value, never
// the value itself.
type Session struct {
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	LastSeen  time.Time `json:"last_seen"`
	UserAgent string    `json:"user_agent,omitempty"`
	IP        string    `json:"ip,omitempty"`
}

func (s Session) Expired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

// SessionStore holds active sessions and persists them across restarts.
type SessionStore struct {
	mu       sync.Mutex
	path     string
	ttl      time.Duration
	sessions map[string]*Session
	dirty    bool
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// NewSessionStore opens the store at path, dropping anything already expired.
func NewSessionStore(path string, ttl time.Duration) (*SessionStore, error) {
	s := &SessionStore{
		path:     path,
		ttl:      ttl,
		sessions: map[string]*Session{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}

	var stored []*Session
	if err := json.Unmarshal(data, &stored); err != nil {
		return s, nil
	}

	now := time.Now()
	for _, sess := range stored {
		if sess.Key == "" || sess.Expired(now) {
			continue
		}
		s.sessions[sess.Key] = sess
	}
	return s, nil
}

// Create issues a new session and returns the raw token for the cookie. Only its
// hash is retained.
func (s *SessionStore) Create(userAgent, ip string) (string, error) {
	raw, err := randomToken()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	sess := &Session{
		Key:       hashToken(raw),
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
		LastSeen:  now,
		UserAgent: truncate(userAgent, 200),
		IP:        ip,
	}

	s.mu.Lock()
	s.pruneLocked(now)
	s.sessions[sess.Key] = sess
	if len(s.sessions) > maxSessions {
		s.evictOldestLocked(len(s.sessions) - maxSessions)
	}
	s.mu.Unlock()

	return raw, s.persist()
}

// Validate reports whether raw names a live session, refreshing its last-seen
// time at most once a minute to avoid writing on every request.
func (s *SessionStore) Validate(raw string) bool {
	if raw == "" {
		return false
	}
	key := hashToken(raw)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[key]
	if !ok {
		return false
	}
	if sess.Expired(now) {
		delete(s.sessions, key)
		s.dirty = true
		return false
	}

	if now.Sub(sess.LastSeen) > time.Minute {
		sess.LastSeen = now
		s.dirty = true
	}
	return true
}

// Revoke signs out one device.
func (s *SessionStore) Revoke(raw string) error {
	if raw == "" {
		return nil
	}
	s.mu.Lock()
	delete(s.sessions, hashToken(raw))
	s.mu.Unlock()
	return s.persist()
}

// RevokeAll signs out every device and reports how many were affected.
func (s *SessionStore) RevokeAll() (int, error) {
	s.mu.Lock()
	n := len(s.sessions)
	s.sessions = map[string]*Session{}
	s.mu.Unlock()
	return n, s.persist()
}

// List returns active sessions, newest first.
func (s *SessionStore) List() []Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		out = append(out, *sess)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

// Count returns the number of active sessions.
func (s *SessionStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions)
}

func (s *SessionStore) pruneLocked(now time.Time) {
	for key, sess := range s.sessions {
		if sess.Expired(now) {
			delete(s.sessions, key)
			s.dirty = true
		}
	}
}

func (s *SessionStore) evictOldestLocked(n int) {
	type kv struct {
		key string
		at  time.Time
	}
	all := make([]kv, 0, len(s.sessions))
	for key, sess := range s.sessions {
		all = append(all, kv{key, sess.LastSeen})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].at.Before(all[j].at) })

	for i := 0; i < n && i < len(all); i++ {
		delete(s.sessions, all[i].key)
	}
	s.dirty = true
}

func (s *SessionStore) persist() error {
	s.mu.Lock()
	list := make([]*Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		list = append(list, sess)
	}
	s.dirty = false
	s.mu.Unlock()

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(data, '\n'), 0o600)
}

// Flush writes pending last-seen updates. Call it periodically and on shutdown.
func (s *SessionStore) Flush() error {
	s.mu.Lock()
	dirty := s.dirty
	s.mu.Unlock()

	if !dirty {
		return nil
	}
	return s.persist()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
