package api

import (
	"context"
	"time"

	"github.com/victoriacheng15/octodeck/internal/types"
)

// MockClient is an in-memory implementation of Client for unit testing.
type MockClient struct {
	ViewerFn           func(ctx context.Context) (types.Viewer, error)
	ListPRsFn          func(ctx context.Context, cursor string, limit int) ([]types.PullRequest, types.PageInfo, error)
	ListIssuesFn       func(ctx context.Context, cursor string, limit int) ([]types.Issue, types.PageInfo, error)
	ListWorkflowRunsFn func(ctx context.Context, commitSHA string, limit int) ([]types.WorkflowRun, error)
	ListWorkflowJobsFn func(ctx context.Context, runID int64) ([]types.Job, error)
	GetJobLogsFn       func(ctx context.Context, jobID int64, maxBytes int64) (string, error)
	AssignToSelfFn     func(ctx context.Context, assignableID string, viewerID string) error
	ApprovePRFn        func(ctx context.Context, prID string) error
	CloseItemFn        func(ctx context.Context, id string, isPR bool) error
	RerunFailedJobsFn  func(ctx context.Context, runID int64) error
	CancelWorkflowFn   func(ctx context.Context, runID int64) error
	RateLimit          types.RateLimit
	Repo               types.RepoContext
}

var _ Client = (*MockClient)(nil)

// NewMockClient initializes a MockClient with realistic default fixture responses.
func NewMockClient() *MockClient {
	now := time.Now()
	return &MockClient{
		ViewerFn: func(ctx context.Context) (types.Viewer, error) {
			return types.Viewer{
				Login: "octocat",
				ID:    "U_kgDOABC123",
				Name:  "Monalisa Octocat",
			}, nil
		},
		ListPRsFn: func(ctx context.Context, cursor string, limit int) ([]types.PullRequest, types.PageInfo, error) {
			return []types.PullRequest{
				{
					ID:                "PR_kwDO123",
					Number:            101,
					Title:             "feat: add contextual workflows drilldown",
					Author:            "octocat",
					HeadRefName:       "feat/drilldown",
					HeadRefOid:        "a1b2c3d4e5f67890",
					ReviewDecision:    "APPROVED",
					StatusCheckRollup: "SUCCESS",
					Labels:            []string{"enhancement", "ui"},
					Assignees:         []string{"octocat"},
					UpdatedAt:         now.Add(-10 * time.Minute),
				},
				{
					ID:                "PR_kwDO124",
					Number:            102,
					Title:             "fix: handle nil rate limit headers",
					Author:            "contributor",
					HeadRefName:       "fix/ratelimit",
					HeadRefOid:        "f6e5d4c3b2a10987",
					ReviewDecision:    "CHANGES_REQUESTED",
					StatusCheckRollup: "FAILURE",
					Labels:            []string{"bug"},
					Assignees:         nil,
					UpdatedAt:         now.Add(-2 * time.Hour),
				},
			}, types.PageInfo{HasNextPage: false, EndCursor: ""}, nil
		},
		ListIssuesFn: func(ctx context.Context, cursor string, limit int) ([]types.Issue, types.PageInfo, error) {
			return []types.Issue{
				{
					ID:        "ISSUE_kwDO201",
					Number:    50,
					Title:     "Add keyboard navigation shortcuts",
					Author:    "octocat",
					Labels:    []string{"good first issue"},
					Assignees: []string{"octocat"},
					UpdatedAt: now.Add(-1 * time.Hour),
				},
			}, types.PageInfo{HasNextPage: false, EndCursor: ""}, nil
		},
		ListWorkflowRunsFn: func(ctx context.Context, commitSHA string, limit int) ([]types.WorkflowRun, error) {
			return []types.WorkflowRun{
				{
					ID:         987654,
					Name:       "Continuous Integration",
					Status:     "completed",
					Conclusion: "failure",
					HeadSha:    "f6e5d4c3b2a10987",
					HeadBranch: "fix/ratelimit",
					Event:      "pull_request",
					CreatedAt:  now.Add(-20 * time.Minute),
					UpdatedAt:  now.Add(-15 * time.Minute),
				},
			}, nil
		},
		ListWorkflowJobsFn: func(ctx context.Context, runID int64) ([]types.Job, error) {
			return []types.Job{
				{
					ID:          555123,
					RunID:       runID,
					Name:        "Lint Go",
					Status:      "completed",
					Conclusion:  "failure",
					StartedAt:   now.Add(-18 * time.Minute),
					CompletedAt: now.Add(-17 * time.Minute),
					Steps: []types.JobStep{
						{Number: 1, Name: "Set up Go", Status: "completed", Conclusion: "success"},
						{Number: 2, Name: "Run Go Lint", Status: "completed", Conclusion: "failure"},
					},
				},
			}, nil
		},
		GetJobLogsFn: func(ctx context.Context, jobID int64, maxBytes int64) (string, error) {
			return "2026-10-24T17:00:00Z Step 2: Run Go Lint\nError: unformatted file found: main.go\nexit status 1", nil
		},
		AssignToSelfFn: func(ctx context.Context, assignableID string, viewerID string) error {
			return nil
		},
		ApprovePRFn: func(ctx context.Context, prID string) error {
			return nil
		},
		CloseItemFn: func(ctx context.Context, id string, isPR bool) error {
			return nil
		},
		RerunFailedJobsFn: func(ctx context.Context, runID int64) error {
			return nil
		},
		CancelWorkflowFn: func(ctx context.Context, runID int64) error {
			return nil
		},
		RateLimit: types.RateLimit{
			Remaining: 4950,
			Limit:     5000,
			Reset:     now.Add(45 * time.Minute),
			Used:      50,
		},
	}
}

func (m *MockClient) ResolveViewer(ctx context.Context) (types.Viewer, error) {
	if m.ViewerFn != nil {
		return m.ViewerFn(ctx)
	}
	return types.Viewer{}, nil
}

func (m *MockClient) ListPullRequests(ctx context.Context, cursor string, limit int) ([]types.PullRequest, types.PageInfo, error) {
	if m.ListPRsFn != nil {
		return m.ListPRsFn(ctx, cursor, limit)
	}
	return nil, types.PageInfo{}, nil
}

func (m *MockClient) ListIssues(ctx context.Context, cursor string, limit int) ([]types.Issue, types.PageInfo, error) {
	if m.ListIssuesFn != nil {
		return m.ListIssuesFn(ctx, cursor, limit)
	}
	return nil, types.PageInfo{}, nil
}

func (m *MockClient) ListWorkflowRuns(ctx context.Context, commitSHA string, limit int) ([]types.WorkflowRun, error) {
	if m.ListWorkflowRunsFn != nil {
		return m.ListWorkflowRunsFn(ctx, commitSHA, limit)
	}
	return nil, nil
}

func (m *MockClient) ListWorkflowJobs(ctx context.Context, runID int64) ([]types.Job, error) {
	if m.ListWorkflowJobsFn != nil {
		return m.ListWorkflowJobsFn(ctx, runID)
	}
	return nil, nil
}

func (m *MockClient) GetJobLogs(ctx context.Context, jobID int64, maxBytes int64) (string, error) {
	if m.GetJobLogsFn != nil {
		return m.GetJobLogsFn(ctx, jobID, maxBytes)
	}
	return "", nil
}

func (m *MockClient) AssignToSelf(ctx context.Context, assignableID string, viewerID string) error {
	if m.AssignToSelfFn != nil {
		return m.AssignToSelfFn(ctx, assignableID, viewerID)
	}
	return nil
}

func (m *MockClient) ApprovePullRequest(ctx context.Context, prID string) error {
	if m.ApprovePRFn != nil {
		return m.ApprovePRFn(ctx, prID)
	}
	return nil
}

func (m *MockClient) CloseItem(ctx context.Context, id string, isPR bool) error {
	if m.CloseItemFn != nil {
		return m.CloseItemFn(ctx, id, isPR)
	}
	return nil
}

func (m *MockClient) RerunFailedJobs(ctx context.Context, runID int64) error {
	if m.RerunFailedJobsFn != nil {
		return m.RerunFailedJobsFn(ctx, runID)
	}
	return nil
}

func (m *MockClient) CancelWorkflowRun(ctx context.Context, runID int64) error {
	if m.CancelWorkflowFn != nil {
		return m.CancelWorkflowFn(ctx, runID)
	}
	return nil
}

func (m *MockClient) SetRepo(repo types.RepoContext) {
	m.Repo = repo
}

func (m *MockClient) GetRateLimit() types.RateLimit {
	return m.RateLimit
}
