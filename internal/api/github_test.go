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

	t.Run("ListPullRequests respects cancellation", func(t *testing.T) {
		_, _, err := client.ListPullRequests(ctx, "", 10)
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})

	t.Run("ListIssues respects cancellation", func(t *testing.T) {
		_, _, err := client.ListIssues(ctx, "", 10)
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})

	t.Run("ListWorkflowRuns respects cancellation", func(t *testing.T) {
		_, err := client.ListWorkflowRuns(ctx, "", 10)
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})

	t.Run("ListWorkflowJobs respects cancellation", func(t *testing.T) {
		_, err := client.ListWorkflowJobs(ctx, 123)
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})

	t.Run("GetJobLogs respects cancellation", func(t *testing.T) {
		_, err := client.GetJobLogs(ctx, 123, 1024)
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})

	t.Run("AssignToSelf respects cancellation", func(t *testing.T) {
		err := client.AssignToSelf(ctx, "I_1", "U_1")
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})

	t.Run("ApprovePullRequest respects cancellation", func(t *testing.T) {
		err := client.ApprovePullRequest(ctx, "PR_1")
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})

	t.Run("CloseItem respects cancellation", func(t *testing.T) {
		err := client.CloseItem(ctx, "PR_1", true)
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})

	t.Run("RerunFailedJobs respects cancellation", func(t *testing.T) {
		err := client.RerunFailedJobs(ctx, 123)
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})

	t.Run("CancelWorkflowRun respects cancellation", func(t *testing.T) {
		err := client.CancelWorkflowRun(ctx, 123)
		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})
}

func TestGitHubClient_RateLimitTracking(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ratelimit-Limit", "5000")
		w.Header().Set("X-Ratelimit-Remaining", "4800")
		w.Header().Set("X-Ratelimit-Used", "200")
		w.Header().Set("X-Ratelimit-Reset", "1700000000")

		if strings.Contains(r.URL.Path, "graphql") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"viewer":{"login":"testuser","id":"MDQ6VXNlcjE="}}}`))
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

	if viewer.Login != "testuser" || viewer.ID != "MDQ6VXNlcjE=" {
		t.Errorf("unexpected viewer: %+v", viewer)
	}

	// Verify rate limit tracker was updated from response headers
	quota := client.GetRateLimit()
	if quota.Limit != 5000 || quota.Remaining != 4800 {
		t.Errorf("quota = %+v, want Limit=5000, Remaining=4800", quota)
	}

	// Verify caching returns same viewer without another network call
	cachedViewer, err := client.ResolveViewer(context.Background())
	if err != nil {
		t.Fatalf("cached ResolveViewer failed: %v", err)
	}
	if cachedViewer.Login != viewer.Login {
		t.Errorf("cached login = %s, want %s", cachedViewer.Login, viewer.Login)
	}
}
