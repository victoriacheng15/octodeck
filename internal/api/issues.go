package api

import (
	"context"
	"fmt"

	"github.com/victoriacheng15/octodeck/internal/types"
)

type issuesGraphQLResponse struct {
	Repository struct {
		Issues struct {
			PageInfo struct {
				HasNextPage     bool   `json:"hasNextPage"`
				HasPreviousPage bool   `json:"hasPreviousPage"`
				StartCursor     string `json:"startCursor"`
				EndCursor       string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []struct {
				ID        string `json:"id"`
				Number    int    `json:"number"`
				Title     string `json:"title"`
				State     string `json:"state"`
				CreatedAt string `json:"createdAt"`
				Author    struct {
					Login string `json:"login"`
				} `json:"author"`
				Assignees struct {
					Nodes []struct {
						Login string `json:"login"`
					} `json:"nodes"`
				} `json:"assignees"`
				Comments struct {
					TotalCount int `json:"totalCount"`
				} `json:"comments"`
				Labels struct {
					Nodes []struct {
						Name  string `json:"name"`
						Color string `json:"color"`
					} `json:"nodes"`
				} `json:"labels"`
			} `json:"nodes"`
		} `json:"issues"`
	} `json:"repository"`
}

// ListIssues queries open issues with cursor pagination.
func (c *GitHubClient) ListIssues(ctx context.Context, cursor string, limit int) ([]types.Issue, types.PageInfo, error) {
	select {
	case <-ctx.Done():
		return nil, types.PageInfo{}, ctx.Err()
	default:
	}

	if limit <= 0 {
		limit = 25
	}

	cacheKey := fmt.Sprintf("issues:%s:%s:%s:%d", c.repo.Owner, c.repo.Name, cursor, limit)
	if cached, ok := c.cache.Get(cacheKey); ok {
		if res, ok := cached.(struct {
			Issues   []types.Issue
			PageInfo types.PageInfo
		}); ok {
			return res.Issues, res.PageInfo, nil
		}
	}

	query := `query ListIssues($owner: String!, $name: String!, $first: Int!, $after: String) {
		repository(owner: $owner, name: $name) {
			issues(first: $first, after: $after, states: [OPEN], orderBy: {field: CREATED_AT, direction: DESC}) {
				pageInfo {
					hasNextPage
					hasPreviousPage
					startCursor
					endCursor
				}
				nodes {
					id
					number
					title
					state
					createdAt
					author { login }
					assignees(first: 5) {
						nodes { login }
					}
					comments { totalCount }
					labels(first: 10) {
						nodes { name color }
					}
				}
			}
		}
	}`

	vars := map[string]interface{}{
		"owner": c.repo.Owner,
		"name":  c.repo.Name,
		"first": limit,
	}
	if cursor != "" {
		vars["after"] = cursor
	}

	var resp issuesGraphQLResponse
	if err := c.gql.DoWithContext(ctx, query, vars, &resp); err != nil {
		return nil, types.PageInfo{}, fmt.Errorf("failed to list issues: %w", err)
	}

	var issues []types.Issue
	for _, n := range resp.Repository.Issues.Nodes {
		var labels []string
		for _, l := range n.Labels.Nodes {
			labels = append(labels, l.Name)
		}
		var assignees []string
		for _, a := range n.Assignees.Nodes {
			assignees = append(assignees, a.Login)
		}
		issues = append(issues, types.Issue{
			ID:        n.ID,
			Number:    n.Number,
			Title:     n.Title,
			Author:    n.Author.Login,
			Assignees: assignees,
			Labels:    labels,
		})
	}

	pageInfo := types.PageInfo{
		HasNextPage:     resp.Repository.Issues.PageInfo.HasNextPage,
		HasPreviousPage: resp.Repository.Issues.PageInfo.HasPreviousPage,
		StartCursor:     resp.Repository.Issues.PageInfo.StartCursor,
		EndCursor:       resp.Repository.Issues.PageInfo.EndCursor,
	}

	c.cache.Set(cacheKey, struct {
		Issues   []types.Issue
		PageInfo types.PageInfo
	}{Issues: issues, PageInfo: pageInfo}, 0)

	return issues, pageInfo, nil
}
