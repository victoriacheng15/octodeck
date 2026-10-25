package types

// NavigateMsg requests the root router to transition to another view.
type NavigateMsg struct {
	Target ViewState
}

// PRsLoadedMsg carries fetched pull requests and pagination state.
type PRsLoadedMsg struct {
	PRs      []PullRequest
	PageInfo PageInfo
	Err      error
}

// IssuesLoadedMsg carries fetched issues and pagination state.
type IssuesLoadedMsg struct {
	Issues   []Issue
	PageInfo PageInfo
	Err      error
}

// WorkflowsLoadedMsg carries workflow runs, optionally filtered by commit SHA.
type WorkflowsLoadedMsg struct {
	Runs      []WorkflowRun
	CommitSHA string
	Err       error
}

// JobsLoadedMsg carries the jobs belonging to a workflow run.
type JobsLoadedMsg struct {
	RunID int64
	Jobs  []Job
	Err   error
}

// LogsLoadedMsg delivers plain-text job logs for display in the viewport.
type LogsLoadedMsg struct {
	JobID   int64
	Content string
	Err     error
}

// RateLimitUpdatedMsg delivers updated GitHub API quota metrics to the UI.
type RateLimitUpdatedMsg struct {
	RateLimit RateLimit
}

// ViewerResolvedMsg delivers the authenticated viewer context after hydration.
type ViewerResolvedMsg struct {
	Viewer Viewer
	Err    error
}

// RepoSwitchedMsg delivers a newly targeted repository context after in-TUI switching.
type RepoSwitchedMsg struct {
	Repo RepoContext
}

// MutationResultMsg communicates the outcome of a triage mutation (assign, close, approve, rerun).
type MutationResultMsg struct {
	Operation string
	Success   bool
	Message   string
	Err       error
}

// ErrMsg represents an application-wide error message.
type ErrMsg struct {
	Err error
}

func (e ErrMsg) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}
