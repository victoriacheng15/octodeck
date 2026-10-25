package api

import (
	"context"

	"github.com/victoriacheng15/octodeck/internal/types"
)

// Client defines the contract for communicating with GitHub APIs (GraphQL v4 and REST v3).
type Client interface {
	// ResolveViewer hydrates the current authenticated viewer's identity and Node ID.
	ResolveViewer(ctx context.Context) (types.Viewer, error)

	// ListPullRequests queries open pull requests with cursor pagination.
	ListPullRequests(ctx context.Context, cursor string, limit int) ([]types.PullRequest, types.PageInfo, error)

	// ListIssues queries open issues with cursor pagination.
	ListIssues(ctx context.Context, cursor string, limit int) ([]types.Issue, types.PageInfo, error)

	// ListWorkflowRuns queries workflow runs, optionally filtered by commit SHA.
	ListWorkflowRuns(ctx context.Context, commitSHA string, limit int) ([]types.WorkflowRun, error)

	// ListWorkflowJobs queries the individual jobs for a workflow run.
	ListWorkflowJobs(ctx context.Context, runID int64) ([]types.Job, error)

	// GetJobLogs fetches plain-text logs for a job, capped at maxBytes.
	GetJobLogs(ctx context.Context, jobID int64, maxBytes int64) (string, error)

	// AssignToSelf adds the authenticated viewer as an assignee to an Issue or PR.
	AssignToSelf(ctx context.Context, assignableID string, viewerID string) error

	// ApprovePullRequest approves a pull request via GraphQL mutation.
	ApprovePullRequest(ctx context.Context, prID string) error

	// CloseItem closes an Issue or Pull Request via GraphQL mutation.
	CloseItem(ctx context.Context, id string, isPR bool) error

	// RerunFailedJobs triggers a re-run of failed jobs in a workflow run via REST.
	RerunFailedJobs(ctx context.Context, runID int64) error

	// CancelWorkflowRun cancels an active workflow run via REST.
	CancelWorkflowRun(ctx context.Context, runID int64) error

	// GetRateLimit returns the most recently recorded rate-limit quota state.
	GetRateLimit() types.RateLimit
}
