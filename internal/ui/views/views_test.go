package views

import (
	"context"
	"strings"
	"testing"

	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/types"
)

func TestRenderCIStatus(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		mustContain string
	}{
		{name: "success status", status: "SUCCESS", mustContain: "pass"},
		{name: "failure status", status: "FAILURE", mustContain: "fail"},
		{name: "pending status", status: "PENDING", mustContain: "pending"},
		{name: "unknown status", status: "UNKNOWN", mustContain: "-"},
		{name: "empty status", status: "", mustContain: "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderCIStatus(tt.status)
			if !strings.Contains(got, tt.mustContain) {
				t.Errorf("RenderCIStatus(%q) = %q, want to contain %q", tt.status, got, tt.mustContain)
			}
		})
	}
}

func TestRenderReviewDecision(t *testing.T) {
	tests := []struct {
		name        string
		decision    string
		mustContain string
	}{
		{name: "approved status", decision: "APPROVED", mustContain: "Approved"},
		{name: "changes requested status", decision: "CHANGES_REQUESTED", mustContain: "Changes"},
		{name: "review required status", decision: "REVIEW_REQUIRED", mustContain: "Required"},
		{name: "empty decision", decision: "", mustContain: "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderReviewDecision(tt.decision)
			if !strings.Contains(got, tt.mustContain) {
				t.Errorf("RenderReviewDecision(%q) = %q, want to contain %q", tt.decision, got, tt.mustContain)
			}
		})
	}
}

func TestFormatLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{name: "nil labels", input: nil, expected: "-"},
		{name: "empty labels", input: []string{}, expected: "-"},
		{name: "single label", input: []string{"bug"}, expected: "bug"},
		{name: "multiple labels", input: []string{"bug", "p0", "cli"}, expected: "bug, p0, cli"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatLabels(tt.input)
			if got != tt.expected {
				t.Errorf("FormatLabels(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatAssignees(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{name: "nil assignees", input: nil, expected: "-"},
		{name: "empty assignees", input: []string{}, expected: "-"},
		{name: "single assignee", input: []string{"octocat"}, expected: "@octocat"},
		{name: "multiple assignees", input: []string{"octocat", "mona"}, expected: "@octocat, @mona"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAssignees(tt.input)
			if got != tt.expected {
				t.Errorf("FormatAssignees(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIssuesModel_TableRendering(t *testing.T) {
	client := api.NewMockClient()
	issues, pageInfo, err := client.ListIssues(context.Background(), "", 10)
	if err != nil {
		t.Fatalf("unexpected mock error: %v", err)
	}

	tests := []struct {
		name        string
		width       int
		height      int
		issues      []types.Issue
		mustContain []string
	}{
		{
			name:   "standard 80x20 table layout",
			width:  80,
			height: 20,
			issues: issues,
			mustContain: []string{
				"#",
				"Title",
				"Author",
				"Labels",
				"Assignees",
			},
		},
		{
			name:   "wide 120x30 table layout",
			width:  120,
			height: 30,
			issues: issues,
			mustContain: []string{
				"#",
				"Title",
				"Author",
				"Labels",
				"Assignees",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewIssuesModel(client)
			m.SetDimensions(tt.width, tt.height)

			updated, _ := m.Update(types.IssuesLoadedMsg{
				Issues:   tt.issues,
				PageInfo: pageInfo,
			})

			rendered := updated.View()
			for _, exp := range tt.mustContain {
				if !strings.Contains(rendered, exp) {
					t.Errorf("rendered view missing %q: %s", exp, rendered)
				}
			}

			if len(updated.table.Rows()) != len(tt.issues) {
				t.Errorf("rows = %d, want %d", len(updated.table.Rows()), len(tt.issues))
			}
		})
	}
}

func TestPRsModel_TableRendering(t *testing.T) {
	client := api.NewMockClient()
	prs, pageInfo, err := client.ListPullRequests(context.Background(), "", 10)
	if err != nil {
		t.Fatalf("unexpected mock error: %v", err)
	}

	tests := []struct {
		name        string
		width       int
		height      int
		prs         []types.PullRequest
		mustContain []string
	}{
		{
			name:   "standard 80x20 table layout",
			width:  80,
			height: 20,
			prs:    prs,
			mustContain: []string{
				"#",
				"Title",
				"Author",
				"Review",
				"CI",
			},
		},
		{
			name:   "wide 120x30 table layout",
			width:  120,
			height: 30,
			prs:    prs,
			mustContain: []string{
				"#",
				"Title",
				"Author",
				"Review",
				"CI",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPRsModel(client)
			m.SetDimensions(tt.width, tt.height)

			updated, _ := m.Update(types.PRsLoadedMsg{
				PRs:      tt.prs,
				PageInfo: pageInfo,
			})

			rendered := updated.View()
			for _, exp := range tt.mustContain {
				if !strings.Contains(rendered, exp) {
					t.Errorf("rendered view missing %q: %s", exp, rendered)
				}
			}

			if len(updated.table.Rows()) != len(tt.prs) {
				t.Errorf("rows = %d, want %d", len(updated.table.Rows()), len(tt.prs))
			}
		})
	}
}
