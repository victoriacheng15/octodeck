package api

import (
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

// SetRepo updates the target repository context and flushes in-memory cache.
func (c *GitHubClient) SetRepo(repo types.RepoContext) {
	c.repo = repo
	if c.cache != nil {
		c.cache.Clear()
	}
}

// GetRateLimit returns the current rate limit quota recorded by the tracker.
func (c *GitHubClient) GetRateLimit() types.RateLimit {
	if c.rateTracker == nil {
		return types.RateLimit{}
	}
	return c.rateTracker.Get()
}
