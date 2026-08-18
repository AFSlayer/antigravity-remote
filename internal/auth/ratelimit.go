package auth

import (
	"sync"
	"time"
)

const (
	maxLimiterEntries = 4096
	maxBackoff        = 30 * time.Minute
)

type limiterEntry struct {
	failures    int
	windowStart time.Time
	lockedUntil time.Time
	strikes     int
	seen        time.Time
}

// Limiter counts failures per key and locks the key out with exponential backoff
// once they exceed max within window.
type Limiter struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
	max     int
	window  time.Duration
}

// NewLimiter returns a Limiter allowing max failures per window.
func NewLimiter(max int, window time.Duration) *Limiter {
	return &Limiter{
		entries: map[string]*limiterEntry{},
		max:     max,
		window:  window,
	}
}

// Allow reports whether key may attempt again, and if not, how long to wait.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[key]
	if !ok {
		return true, 0
	}
	e.seen = now

	if now.Before(e.lockedUntil) {
		return false, e.lockedUntil.Sub(now).Round(time.Second)
	}
	if now.Sub(e.windowStart) > l.window {
		e.failures = 0
		e.windowStart = now
	}
	return true, 0
}

// Fail records a failed attempt, extending the lockout if the limit is reached.
func (l *Limiter) Fail(key string) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.pruneLocked(now)

	e, ok := l.entries[key]
	if !ok {
		e = &limiterEntry{windowStart: now}
		l.entries[key] = e
	}
	e.seen = now

	if now.Sub(e.windowStart) > l.window {
		e.failures = 0
		e.windowStart = now
	}
	e.failures++

	if e.failures >= l.max {
		e.strikes++
		backoff := time.Duration(1<<uint(min(e.strikes-1, 5))) * time.Minute
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		e.lockedUntil = now.Add(backoff)
		e.failures = 0
		e.windowStart = now
	}
}

// Reset clears a key's history after a successful attempt.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	delete(l.entries, key)
	l.mu.Unlock()
}

func (l *Limiter) pruneLocked(now time.Time) {
	if len(l.entries) < maxLimiterEntries {
		return
	}
	for key, e := range l.entries {
		if now.Sub(e.seen) > l.window && now.After(e.lockedUntil) {
			delete(l.entries, key)
		}
	}
}
