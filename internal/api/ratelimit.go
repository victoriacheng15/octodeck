package api

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/victoriacheng15/octodeck/internal/types"
)

// RateLimitTracker maintains thread-safe state for the active GitHub API quota.
type RateLimitTracker struct {
	mu    sync.RWMutex
	quota types.RateLimit
}

// NewRateLimitTracker constructs a tracker initialized to zero quota.
func NewRateLimitTracker() *RateLimitTracker {
	return &RateLimitTracker{}
}

// Get returns a copy of the current rate limit quota.
func (t *RateLimitTracker) Get() types.RateLimit {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.quota
}

// Update explicitly sets the rate limit quota.
func (t *RateLimitTracker) Update(limit, remaining, used int, reset time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.quota = types.RateLimit{
		Limit:     limit,
		Remaining: remaining,
		Used:      used,
		Reset:     reset,
	}
}

// ParseHeaders extracts X-RateLimit-* headers from HTTP response headers.
func (t *RateLimitTracker) ParseHeaders(header http.Header) {
	limitStr := header.Get("X-Ratelimit-Limit")
	remStr := header.Get("X-Ratelimit-Remaining")
	usedStr := header.Get("X-Ratelimit-Used")
	resetStr := header.Get("X-Ratelimit-Reset")

	if limitStr == "" && remStr == "" {
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return
	}

	rem, err := strconv.Atoi(remStr)
	if err != nil {
		return
	}

	used, _ := strconv.Atoi(usedStr)

	var resetAt time.Time
	if unixSec, err := strconv.ParseInt(resetStr, 10, 64); err == nil && unixSec > 0 {
		resetAt = time.Unix(unixSec, 0)
	}

	t.Update(limit, rem, used, resetAt)
}

// RateLimitInterceptor is an http.RoundTripper that intercepts HTTP responses
// to capture rate limit quota metrics transparently.
type RateLimitInterceptor struct {
	Base    http.RoundTripper
	Tracker *RateLimitTracker
}

// RoundTrip executes the HTTP request and inspects response headers for quota updates.
func (i *RateLimitInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	base := i.Base
	if base == nil {
		base = http.DefaultTransport
	}

	resp, err := base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if i.Tracker != nil && resp != nil {
		i.Tracker.ParseHeaders(resp.Header)
	}

	return resp, nil
}
