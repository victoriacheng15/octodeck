package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/victoriacheng15/octodeck/internal/types"
)

func TestRateLimitTracker_ParseHeaders(t *testing.T) {
	tests := []struct {
		name        string
		initial     types.RateLimit
		headers     map[string]string
		wantLimit   int
		wantRem     int
		wantUsed    int
		wantResetAt time.Time
	}{
		{
			name: "parse standard rate limit headers",
			headers: map[string]string{
				"X-Ratelimit-Limit":     "5000",
				"X-Ratelimit-Remaining": "4950",
				"X-Ratelimit-Used":      "50",
				"X-Ratelimit-Reset":     "1700000000",
			},
			wantLimit:   5000,
			wantRem:     4950,
			wantUsed:    50,
			wantResetAt: time.Unix(1700000000, 0),
		},
		{
			name: "ignore empty headers and preserve existing quota",
			initial: types.RateLimit{
				Limit:     5000,
				Remaining: 5000,
			},
			headers:     map[string]string{},
			wantLimit:   5000,
			wantRem:     5000,
			wantUsed:    0,
			wantResetAt: time.Time{},
		},
		{
			name: "ignore non-integer headers",
			initial: types.RateLimit{
				Limit:     5000,
				Remaining: 5000,
			},
			headers: map[string]string{
				"X-Ratelimit-Limit":     "invalid",
				"X-Ratelimit-Remaining": "invalid",
			},
			wantLimit: 5000,
			wantRem:   5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewRateLimitTracker()
			if tt.initial.Limit > 0 {
				tracker.Update(tt.initial.Limit, tt.initial.Remaining, tt.initial.Used, tt.initial.Reset)
			}

			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}

			tracker.ParseHeaders(h)
			got := tracker.Get()

			if got.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", got.Limit, tt.wantLimit)
			}
			if got.Remaining != tt.wantRem {
				t.Errorf("Remaining = %d, want %d", got.Remaining, tt.wantRem)
			}
			if got.Used != tt.wantUsed {
				t.Errorf("Used = %d, want %d", got.Used, tt.wantUsed)
			}
			if !tt.wantResetAt.IsZero() && got.Reset != tt.wantResetAt {
				t.Errorf("Reset = %v, want %v", got.Reset, tt.wantResetAt)
			}
		})
	}
}

func TestRateLimitInterceptor(t *testing.T) {
	tests := []struct {
		name         string
		headerLimit  string
		headerRem    string
		wantRemQuota int
	}{
		{
			name:         "intercept standard response",
			headerLimit:  "5000",
			headerRem:    "4999",
			wantRemQuota: 4999,
		},
		{
			name:         "intercept rate-limited warning response",
			headerLimit:  "5000",
			headerRem:    "50",
			wantRemQuota: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewRateLimitTracker()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Ratelimit-Limit", tt.headerLimit)
				w.Header().Set("X-Ratelimit-Remaining", tt.headerRem)
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
			if quota.Remaining != tt.wantRemQuota {
				t.Errorf("quota.Remaining = %d, want %d", quota.Remaining, tt.wantRemQuota)
			}
		})
	}
}
