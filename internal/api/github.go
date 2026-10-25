package api

import (
	"context"
	"net/http"
	"time"

	ghapi "github.com/cli/go-gh/v2/pkg/api"
	"github.com/victoriacheng15/octodeck/internal/types"
)

// GitHubClient implements the Client interface communicating with GitHub v4 and v3 APIs.
type GitHubClient struct {
	repo        types.RepoContext
	gql         *ghapi.GraphQLClient
	rest        *ghapi.RESTClient
	cache       *Cache
	rateTracker *RateLimitTracker
}

var _ Client = (*GitHubClient)(nil)

// GitHubClientOptions customizes the GitHub client construction.
type GitHubClientOptions struct {
	Host       string
	AuthToken  string
	Transport  http.RoundTripper
	DefaultTTL time.Duration
}

// NewGitHubClient constructs a new live GitHub client connected to GitHub APIs.
func NewGitHubClient(repo types.RepoContext, opts GitHubClientOptions) (*GitHubClient, error) {
	tracker := NewRateLimitTracker()

	baseTransport := opts.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	interceptor := &RateLimitInterceptor{
		Base:    baseTransport,
		Tracker: tracker,
	}

	host := opts.Host
	if host == "" {
		host = repo.Host
	}
	if host == "" {
		host = "github.com"
	}

	clientOpts := ghapi.ClientOptions{
		Host:      host,
		AuthToken: opts.AuthToken,
		Transport: interceptor,
	}

	gqlClient, err := ghapi.NewGraphQLClient(clientOpts)
	if err != nil {
		return nil, err
	}

	restClient, err := ghapi.NewRESTClient(clientOpts)
	if err != nil {
		return nil, err
	}

	ttl := opts.DefaultTTL
	if ttl <= 0 {
		ttl = 30 * time.Second
	}

	return &GitHubClient{
		repo:        repo,
		gql:         gqlClient,
		rest:        restClient,
		cache:       NewCache(ttl),
		rateTracker: tracker,
	}, nil
}

// GetRateLimit returns the current rate limit quota recorded by the tracker.
func (c *GitHubClient) GetRateLimit() types.RateLimit {
	if c.rateTracker == nil {
		return types.RateLimit{}
	}
	return c.rateTracker.Get()
}

// ListPullRequests is stubbed until the PRs feature PR.
func (c *GitHubClient) ListPullRequests(ctx context.Context, cursor string, limit int) ([]types.PullRequest, types.PageInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, types.PageInfo{}, err
	}
	return nil, types.PageInfo{}, nil
}

// ListWorkflowRuns is stubbed until the Workflows feature PR.
func (c *GitHubClient) ListWorkflowRuns(ctx context.Context, commitSHA string, limit int) ([]types.WorkflowRun, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

// ListWorkflowJobs is stubbed until the Workflows feature PR.
func (c *GitHubClient) ListWorkflowJobs(ctx context.Context, runID int64) ([]types.Job, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

// GetJobLogs is stubbed until the Workflows feature PR.
func (c *GitHubClient) GetJobLogs(ctx context.Context, jobID int64, maxBytes int64) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "", nil
}

// RerunFailedJobs is stubbed until the Workflows feature PR.
func (c *GitHubClient) RerunFailedJobs(ctx context.Context, runID int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

// CancelWorkflowRun is stubbed until the Workflows feature PR.
func (c *GitHubClient) CancelWorkflowRun(ctx context.Context, runID int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
