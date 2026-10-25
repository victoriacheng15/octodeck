package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/victoriacheng15/octodeck/internal/types"
)

func TestGitHubClient_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel context

	repo := types.RepoContext{Owner: "cli", Name: "go-gh"}
	client := &GitHubClient{
		repo:        repo,
		cache:       NewCache(1 * time.Minute),
		rateTracker: NewRateLimitTracker(),
	}

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "ListPullRequests respects cancellation",
			call: func() error {
				_, _, err := client.ListPullRequests(ctx, "", 10)
				return err
			},
		},
		{
			name: "ListIssues respects cancellation",
			call: func() error {
				_, _, err := client.ListIssues(ctx, "", 10)
				return err
			},
		},
		{
			name: "ListWorkflowRuns respects cancellation",
			call: func() error {
				_, err := client.ListWorkflowRuns(ctx, "", 10)
				return err
			},
		},
		{
			name: "ListWorkflowJobs respects cancellation",
			call: func() error {
				_, err := client.ListWorkflowJobs(ctx, 123)
				return err
			},
		},
		{
			name: "GetJobLogs respects cancellation",
			call: func() error {
				_, err := client.GetJobLogs(ctx, 123, 1024)
				return err
			},
		},
		{
			name: "AssignToSelf respects cancellation",
			call: func() error {
				return client.AssignToSelf(ctx, "I_1", "U_1")
			},
		},
		{
			name: "ApprovePullRequest respects cancellation",
			call: func() error {
				return client.ApprovePullRequest(ctx, "PR_1")
			},
		},
		{
			name: "CloseItem respects cancellation",
			call: func() error {
				return client.CloseItem(ctx, "PR_1", true)
			},
		},
		{
			name: "RerunFailedJobs respects cancellation",
			call: func() error {
				return client.RerunFailedJobs(ctx, 123)
			},
		},
		{
			name: "CancelWorkflowRun respects cancellation",
			call: func() error {
				return client.CancelWorkflowRun(ctx, 123)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err != context.Canceled {
				t.Errorf("got %v, want context.Canceled", err)
			}
		})
	}
}

func TestGitHubClient_RateLimitTracking(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		respBody   string
		wantViewer types.Viewer
		wantLimit  int
		wantRem    int
	}{
		{
			name: "resolve viewer and capture rate limit headers",
			headers: map[string]string{
				"X-Ratelimit-Limit":     "5000",
				"X-Ratelimit-Remaining": "4800",
				"X-Ratelimit-Used":      "200",
				"X-Ratelimit-Reset":     "1700000000",
			},
			respBody: `{"data":{"viewer":{"login":"testuser","id":"MDQ6VXNlcjE="}}}`,
			wantViewer: types.Viewer{
				Login: "testuser",
				ID:    "MDQ6VXNlcjE=",
			},
			wantLimit: 5000,
			wantRem:   4800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for k, v := range tt.headers {
					w.Header().Set(k, v)
				}
				if strings.Contains(r.URL.Path, "graphql") {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.respBody))
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()

			repo := types.RepoContext{Owner: "testowner", Name: "testrepo"}
			client, err := NewGitHubClient(repo, GitHubClientOptions{
				Host:      server.Listener.Addr().String(),
				AuthToken: "mock-token",
				Transport: server.Client().Transport,
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			viewer, err := client.ResolveViewer(context.Background())
			if err != nil {
				t.Fatalf("ResolveViewer failed: %v", err)
			}

			if viewer.Login != tt.wantViewer.Login || viewer.ID != tt.wantViewer.ID {
				t.Errorf("got viewer %+v, want %+v", viewer, tt.wantViewer)
			}

			quota := client.GetRateLimit()
			if quota.Limit != tt.wantLimit || quota.Remaining != tt.wantRem {
				t.Errorf("quota = %+v, want Limit=%d, Remaining=%d", quota, tt.wantLimit, tt.wantRem)
			}

			cachedViewer, err := client.ResolveViewer(context.Background())
			if err != nil {
				t.Fatalf("cached ResolveViewer failed: %v", err)
			}
			if cachedViewer.Login != viewer.Login {
				t.Errorf("cached login = %s, want %s", cachedViewer.Login, viewer.Login)
			}
		})
	}
}
