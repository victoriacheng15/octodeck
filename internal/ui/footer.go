package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/victoriacheng15/octodeck/internal/types"
)

// RenderFooter renders the bottom status bar with hotkeys, rate limit quota, and error notifications.
func RenderFooter(quota types.RateLimit, errMsg string, width int, theme Theme) string {
	if errMsg != "" {
		return theme.ErrorBox.Width(width).Render("Error: " + errMsg)
	}

	shortcuts := []struct {
		key  string
		desc string
	}{
		{":", "command"},
		{"/", "filter"},
		{"r", "refresh"},
		{"q", "quit"},
	}

	var renderedShortcuts []string
	for _, s := range shortcuts {
		renderedShortcuts = append(renderedShortcuts,
			theme.KeyHint.Render(s.key)+theme.KeyDesc.Render(":"+s.desc),
		)
	}
	left := strings.Join(renderedShortcuts, "  ")

	right := ""
	if quota.Limit > 0 {
		right = theme.QuotaText.Render(fmt.Sprintf("API: %d/%d", quota.Remaining, quota.Limit))
	}

	space := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if space < 1 {
		space = 1
	}

	content := lipgloss.JoinHorizontal(lipgloss.Center, left, strings.Repeat(" ", space), right)
	return theme.FooterBar.Width(width).Render(content)
}

// FooterHeight returns the line height consumed by the footer.
func FooterHeight() int {
	return 2
}
