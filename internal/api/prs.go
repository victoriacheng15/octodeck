package api

import (
	"context"
	"fmt"

	"github.com/victoriacheng15/octodeck/internal/types"
)

type prsGraphQLResponse struct {
	Repository struct {
		PullRequests struct {
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
				IsDraft   bool   `json:"isDraft"`
				CreatedAt string `json:"createdAt"`
				Author    struct {
					Login string `json:"login"`
				} `json:"author"`
				HeadRefName string `json:"headRefName"`
				BaseRefName string `json:"baseRefName"`
				Comments    struct {
					TotalCount int `json:"totalCount"`
				} `json:"comments"`
				Labels struct {
					Nodes []struct {
						Name  string `json:"name"`
						Color string `json:"color"`
					} `json:"nodes"`
				} `json:"labels"`
			} `json:"nodes"`
		} `json:"pullRequests"`
	} `json:"repository"`
}

// ListPullRequests queries open pull requests with cursor pagination.
func (c *GitHubClient) ListPullRequests(ctx context.Context, cursor string, limit int) ([]types.PullRequest, types.PageInfo, error) {
	select {
	case <-ctx.Done():
		return nil, types.PageInfo{}, ctx.Err()
	default:
	}

	if limit <= 0 {
		limit = 25
	}

	cacheKey := fmt.Sprintf("prs:%s:%s:%s:%d", c.repo.Owner, c.repo.Name, cursor, limit)
	if cached, ok := c.cache.Get(cacheKey); ok {
		if res, ok := cached.(struct {
			PRs      []types.PullRequest
			PageInfo types.PageInfo
		}); ok {
			return res.PRs, res.PageInfo, nil
		}
	}

	query := `query ListPRs($owner: String!, $name: String!, $first: Int!, $after: String) {
		repository(owner: $owner, name: $name) {
			pullRequests(first: $first, after: $after, states: [OPEN], orderBy: {field: CREATED_AT, direction: DESC}) {
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
					isDraft
					createdAt
					headRefName
					baseRefName
					author { login }
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

	var resp prsGraphQLResponse
	if err := c.gql.DoWithContext(ctx, query, vars, &resp); err != nil {
		return nil, types.PageInfo{}, fmt.Errorf("failed to list pull requests: %w", err)
	}

	var prs []types.PullRequest
	for _, n := range resp.Repository.PullRequests.Nodes {
		var labels []string
		for _, l := range n.Labels.Nodes {
			labels = append(labels, l.Name)
		}
		prs = append(prs, types.PullRequest{
			ID:          n.ID,
			Number:      n.Number,
			Title:       n.Title,
			Author:      n.Author.Login,
			HeadRefName: n.HeadRefName,
			Labels:      labels,
		})
	}

	pageInfo := types.PageInfo{
		HasNextPage:     resp.Repository.PullRequests.PageInfo.HasNextPage,
		HasPreviousPage: resp.Repository.PullRequests.PageInfo.HasPreviousPage,
		StartCursor:     resp.Repository.PullRequests.PageInfo.StartCursor,
		EndCursor:       resp.Repository.PullRequests.PageInfo.EndCursor,
	}

	c.cache.Set(cacheKey, struct {
		PRs      []types.PullRequest
		PageInfo types.PageInfo
	}{PRs: prs, PageInfo: pageInfo}, 0)

	return prs, pageInfo, nil
}
