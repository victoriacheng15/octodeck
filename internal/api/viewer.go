package api

import (
	"context"
	"fmt"

	"github.com/victoriacheng15/octodeck/internal/types"
)

type viewerResponse struct {
	Viewer struct {
		Login string `json:"login"`
		ID    string `json:"id"`
	} `json:"viewer"`
}

// ResolveViewer hydrates the current authenticated viewer's identity and Node ID via GraphQL.
func (c *GitHubClient) ResolveViewer(ctx context.Context) (types.Viewer, error) {
	if cached, ok := c.cache.Get("viewer"); ok {
		if v, ok := cached.(types.Viewer); ok {
			return v, nil
		}
	}

	var resp viewerResponse
	query := `query ViewerQuery { viewer { login id } }`

	if err := c.gql.DoWithContext(ctx, query, nil, &resp); err != nil {
		return types.Viewer{}, fmt.Errorf("failed to resolve viewer: %w", err)
	}

	viewer := types.Viewer{
		Login: resp.Viewer.Login,
		ID:    resp.Viewer.ID,
	}

	c.cache.Set("viewer", viewer, 0)
	return viewer, nil
}

// AssignToSelf adds the authenticated viewer as an assignee to an Issue or PR.
func (c *GitHubClient) AssignToSelf(ctx context.Context, assignableID string, viewerID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	mutation := `mutation AddAssignee($input: AddAssigneesToAssignableInput!) {
		addAssigneesToAssignable(input: $input) {
			clientMutationId
		}
	}`
	vars := map[string]interface{}{
		"input": map[string]interface{}{
			"assignableId": assignableID,
			"assigneeIds":  []string{viewerID},
		},
	}

	var resp struct{}
	if err := c.gql.DoWithContext(ctx, mutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to assign self: %w", err)
	}
	return nil
}

// ApprovePullRequest approves a pull request via GraphQL mutation.
func (c *GitHubClient) ApprovePullRequest(ctx context.Context, prID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	mutation := `mutation ApprovePR($input: AddPullRequestReviewInput!) {
		addPullRequestReview(input: $input) {
			clientMutationId
		}
	}`
	vars := map[string]interface{}{
		"input": map[string]interface{}{
			"pullRequestId": prID,
			"event":         "APPROVE",
		},
	}

	var resp struct{}
	if err := c.gql.DoWithContext(ctx, mutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to approve PR: %w", err)
	}
	return nil
}

// CloseItem closes an Issue or Pull Request via GraphQL mutation.
func (c *GitHubClient) CloseItem(ctx context.Context, id string, isPR bool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var mutation string
	var vars map[string]interface{}

	if isPR {
		mutation = `mutation ClosePR($input: ClosePullRequestInput!) {
			closePullRequest(input: $input) {
				clientMutationId
			}
		}`
		vars = map[string]interface{}{
			"input": map[string]interface{}{
				"pullRequestId": id,
			},
		}
	} else {
		mutation = `mutation CloseIssue($input: CloseIssueInput!) {
			closeIssue(input: $input) {
				clientMutationId
			}
		}`
		vars = map[string]interface{}{
			"input": map[string]interface{}{
				"issueId": id,
			},
		}
	}

	var resp struct{}
	if err := c.gql.DoWithContext(ctx, mutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to close item: %w", err)
	}
	return nil
}
