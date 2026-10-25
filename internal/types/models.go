package types

import "time"

// ViewState represents the currently active view in the TUI router.
type ViewState int

const (
	ViewPRs ViewState = iota
	ViewIssues
	ViewWorkflows
	ViewJobLogs
)

// String returns a human-readable name for the view state.
func (v ViewState) String() string {
	switch v {
	case ViewPRs:
		return "Pull Requests"
	case ViewIssues:
		return "Issues"
	case ViewWorkflows:
		return "Workflows"
	case ViewJobLogs:
		return "Job Logs"
	default:
		return "Unknown"
	}
}

// RepoContext encapsulates the target repository identity.
type RepoContext struct {
	Owner string
	Name  string
	Host  string
}

// String returns owner/name representation.
func (r RepoContext) String() string {
	if r.Owner == "" && r.Name == "" {
		return ""
	}
	return r.Owner + "/" + r.Name
}

// Viewer represents the authenticated GitHub user identity.
type Viewer struct {
	Login string
	ID    string // GraphQL global node ID (required for mutations)
	Name  string
}

// PageInfo holds cursor pagination metadata from GitHub GraphQL v4.
type PageInfo struct {
	HasNextPage     bool
	EndCursor       string
	HasPreviousPage bool
	StartCursor     string
}

// PullRequest represents a normalized GitHub pull request domain model.
type PullRequest struct {
	ID                string
	Number            int
	Title             string
	Author            string
	HeadRefName       string
	HeadRefOid        string // Commit SHA used to correlate with Actions runs
	ReviewDecision    string // APPROVED, CHANGES_REQUESTED, REVIEW_REQUIRED
	StatusCheckRollup string // SUCCESS, FAILURE, PENDING
	Labels            []string
	Assignees         []string
	UpdatedAt         time.Time
}

// Issue represents a normalized GitHub issue domain model.
type Issue struct {
	ID        string
	Number    int
	Title     string
	Author    string
	Labels    []string
	Assignees []string
	UpdatedAt time.Time
}

// WorkflowRun represents an Actions workflow run from REST v3.
type WorkflowRun struct {
	ID         int64
	Name       string
	Status     string // queued, in_progress, completed
	Conclusion string // success, failure, cancelled, timed_out
	HeadSha    string
	HeadBranch string
	Event      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// JobStep represents a single step within a workflow job.
type JobStep struct {
	Number     int
	Name       string
	Status     string
	Conclusion string
}

// Job represents an Actions job within a workflow run.
type Job struct {
	ID          int64
	RunID       int64
	Name        string
	Status      string
	Conclusion  string
	StartedAt   time.Time
	CompletedAt time.Time
	Steps       []JobStep
}

// RateLimit captures quota metrics from response headers.
type RateLimit struct {
	Remaining int
	Limit     int
	Reset     time.Time
	Used      int
}
