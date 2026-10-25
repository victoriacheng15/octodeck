package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitTracker_ParseHeaders(t *testing.T) {
	tracker := NewRateLimitTracker()

	headers := http.Header{}
	headers.Set("X-Ratelimit-Limit", "5000")
	headers.Set("X-Ratelimit-Remaining", "4950")
	headers.Set("X-Ratelimit-Used", "50")
	headers.Set("X-Ratelimit-Reset", "1700000000")

	tracker.ParseHeaders(headers)
	quota := tracker.Get()

	if quota.Limit != 5000 {
		t.Errorf("Limit = %d, want 5000", quota.Limit)
	}
	if quota.Remaining != 4950 {
		t.Errorf("Remaining = %d, want 4950", quota.Remaining)
	}
	if quota.Used != 50 {
		t.Errorf("Used = %d, want 50", quota.Used)
	}
	if quota.Reset != time.Unix(1700000000, 0) {
		t.Errorf("Reset = %v, want %v", quota.Reset, time.Unix(1700000000, 0))
	}
}

func TestRateLimitTracker_IgnoreEmpty(t *testing.T) {
	tracker := NewRateLimitTracker()
	tracker.Update(5000, 5000, 0, time.Time{})

	tracker.ParseHeaders(http.Header{})
	quota := tracker.Get()

	if quota.Limit != 5000 || quota.Remaining != 5000 {
		t.Errorf("expected quota to remain unchanged when empty headers passed")
	}
}

func TestRateLimitInterceptor(t *testing.T) {
	tracker := NewRateLimitTracker()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ratelimit-Limit", "5000")
		w.Header().Set("X-Ratelimit-Remaining", "4999")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &RateLimitInterceptor{
			Base:    http.DefaultTransport,
			Tracker: tracker,
		},
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	defer resp.Body.Close()

	quota := tracker.Get()
	if quota.Remaining != 4999 {
		t.Errorf("quota.Remaining = %d, want 4999", quota.Remaining)
	}
}
