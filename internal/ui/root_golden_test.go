package ui

import (
	"strings"
	"testing"

	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/types"
)

func TestRootModel_GoldenLayout(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		activeView  types.ViewState
		mustContain []string
	}{
		{
			name:       "standard 80x24 layout on PRs",
			width:      80,
			height:     24,
			activeView: types.ViewPRs,
			mustContain: []string{
				"octocat",
				"1:PRs",
				"2:Issues",
				"3:Workflows",
				"Pull Requests View",
				"1-3:tabs",
				"API: 4950/5000",
			},
		},
		{
			name:       "wide 120x40 layout on Issues",
			width:      120,
			height:     40,
			activeView: types.ViewIssues,
			mustContain: []string{
				"charmbracelet/bubbletea",
				"1:PRs",
				"2:Issues",
				"3:Workflows",
				"Issues View",
				"1-3:tabs",
			},
		},
		{
			name:       "layout on Workflows",
			width:      100,
			height:     30,
			activeView: types.ViewWorkflows,
			mustContain: []string{
				"Workflows View",
				"1:PRs",
				"2:Issues",
				"3:Workflows",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := types.RepoContext{Owner: "charmbracelet", Name: "bubbletea"}
			viewer := types.Viewer{Login: "octocat"}
			m := NewRootModel(api.NewMockClient(), repo, viewer)
			m.Width = tt.width
			m.Height = tt.height
			m.ActiveView = tt.activeView

			rendered := m.View()

			for _, expected := range tt.mustContain {
				if !strings.Contains(rendered, expected) {
					t.Errorf("rendered output missing %q", expected)
				}
			}

			lines := strings.Split(rendered, "\n")
			if len(lines) != tt.height {
				t.Errorf("rendered lines = %d, want exact terminal height %d", len(lines), tt.height)
			}
		})
	}
}
