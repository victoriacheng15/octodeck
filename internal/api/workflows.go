package api

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/victoriacheng15/octodeck/internal/types"
)

type workflowRunsRESTResponse struct {
	WorkflowRuns []struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		HeadBranch string `json:"head_branch"`
		HeadSHA    string `json:"head_sha"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		Event      string `json:"event"`
		CreatedAt  string `json:"created_at"`
		Actor      struct {
			Login string `json:"login"`
		} `json:"actor"`
	} `json:"workflow_runs"`
}

type workflowJobsRESTResponse struct {
	Jobs []struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		StartedAt  string `json:"started_at"`
		Steps      []struct {
			Name       string `json:"name"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
			Number     int    `json:"number"`
		} `json:"steps"`
	} `json:"jobs"`
}

// ListWorkflowRuns queries workflow runs, optionally filtered by commit SHA.
func (c *GitHubClient) ListWorkflowRuns(ctx context.Context, commitSHA string, limit int) ([]types.WorkflowRun, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if limit <= 0 {
		limit = 25
	}

	cacheKey := fmt.Sprintf("runs:%s:%s:%s:%d", c.repo.Owner, c.repo.Name, commitSHA, limit)
	if cached, ok := c.cache.Get(cacheKey); ok {
		if runs, ok := cached.([]types.WorkflowRun); ok {
			return runs, nil
		}
	}

	path := fmt.Sprintf("repos/%s/%s/actions/runs?per_page=%d", c.repo.Owner, c.repo.Name, limit)
	if commitSHA != "" {
		path += fmt.Sprintf("&head_sha=%s", commitSHA)
	}

	var resp workflowRunsRESTResponse
	if err := c.rest.Get(path, &resp); err != nil {
		return nil, fmt.Errorf("failed to list workflow runs: %w", err)
	}

	var runs []types.WorkflowRun
	for _, r := range resp.WorkflowRuns {
		runs = append(runs, types.WorkflowRun{
			ID:         r.ID,
			Name:       r.Name,
			HeadBranch: r.HeadBranch,
			HeadSha:    r.HeadSHA,
			Status:     r.Status,
			Conclusion: r.Conclusion,
			Event:      r.Event,
		})
	}

	c.cache.Set(cacheKey, runs, 0)
	return runs, nil
}

// ListWorkflowJobs queries individual jobs for a workflow run.
func (c *GitHubClient) ListWorkflowJobs(ctx context.Context, runID int64) ([]types.Job, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path := fmt.Sprintf("repos/%s/%s/actions/runs/%d/jobs", c.repo.Owner, c.repo.Name, runID)
	var resp workflowJobsRESTResponse
	if err := c.rest.Get(path, &resp); err != nil {
		return nil, fmt.Errorf("failed to list workflow jobs: %w", err)
	}

	var jobs []types.Job
	for _, j := range resp.Jobs {
		var steps []types.JobStep
		for _, s := range j.Steps {
			steps = append(steps, types.JobStep{
				Number:     s.Number,
				Name:       s.Name,
				Status:     s.Status,
				Conclusion: s.Conclusion,
			})
		}
		jobs = append(jobs, types.Job{
			ID:         j.ID,
			Name:       j.Name,
			Status:     j.Status,
			Conclusion: j.Conclusion,
			Steps:      steps,
		})
	}

	return jobs, nil
}

// GetJobLogs fetches plain-text logs for a job, capped at maxBytes.
func (c *GitHubClient) GetJobLogs(ctx context.Context, jobID int64, maxBytes int64) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if maxBytes <= 0 {
		maxBytes = 2 * 1024 * 1024 // 2MB default cap
	}

	path := fmt.Sprintf("repos/%s/%s/actions/jobs/%d/logs", c.repo.Owner, c.repo.Name, jobID)
	httpResp, err := c.rest.RequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to request job logs: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code fetching logs: %d", httpResp.StatusCode)
	}

	limitReader := io.LimitReader(httpResp.Body, maxBytes)
	content, err := io.ReadAll(limitReader)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return string(content), nil
}

// RerunFailedJobs triggers a re-run of failed jobs in a workflow run via REST.
func (c *GitHubClient) RerunFailedJobs(ctx context.Context, runID int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	path := fmt.Sprintf("repos/%s/%s/actions/runs/%d/rerun-failed-jobs", c.repo.Owner, c.repo.Name, runID)
	if err := c.rest.Post(path, nil, nil); err != nil {
		return fmt.Errorf("failed to rerun failed jobs: %w", err)
	}
	return nil
}

// CancelWorkflowRun cancels an active workflow run via REST.
func (c *GitHubClient) CancelWorkflowRun(ctx context.Context, runID int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	path := fmt.Sprintf("repos/%s/%s/actions/runs/%d/cancel", c.repo.Owner, c.repo.Name, runID)
	if err := c.rest.Post(path, nil, nil); err != nil {
		return fmt.Errorf("failed to cancel workflow run: %w", err)
	}
	return nil
}
