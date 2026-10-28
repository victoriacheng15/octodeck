package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/victoriacheng15/octodeck/internal/types"
)

// RenderHeader renders the persistent top header showing target repo, viewer, and active tab.
func RenderHeader(repo types.RepoContext, viewer types.Viewer, activeView types.ViewState, width int, theme Theme) string {
	repoStr := repo.String()
	if repoStr == "" {
		repoStr = "octodeck"
	}
	repoPart := theme.RepoName.Render(repoStr)

	userPart := ""
	if viewer.Login != "" {
		userPart = " " + theme.ViewerBadge.Render("@"+viewer.Login)
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center, repoPart, userPart)

	tabs := []struct {
		state types.ViewState
		label string
	}{
		{types.ViewPRs, "PRs"},
		{types.ViewIssues, "Issues"},
		{types.ViewWorkflows, "Workflows"},
	}

	var renderedTabs []string
	for _, t := range tabs {
		if t.state == activeView {
			renderedTabs = append(renderedTabs, theme.ActiveTab.Render(t.label))
		} else {
			renderedTabs = append(renderedTabs, theme.InactiveTab.Render(t.label))
		}
	}
	right := strings.Join(renderedTabs, " ")

	space := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if space < 1 {
		space = 1
	}

	content := lipgloss.JoinHorizontal(lipgloss.Center, left, strings.Repeat(" ", space), right)
	return theme.BaseHeader.Width(width).Render(content)
}

// HeaderHeight returns the fixed line height consumed by the header.
func HeaderHeight() int {
	return 2
}
