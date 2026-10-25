package ui

import "github.com/charmbracelet/lipgloss"

// Theme holds centralized styles for the TUI layout and widgets.
type Theme struct {
	BaseHeader  lipgloss.Style
	RepoName    lipgloss.Style
	ViewerBadge lipgloss.Style

	ActiveTab   lipgloss.Style
	InactiveTab lipgloss.Style

	FooterBar lipgloss.Style
	KeyHint   lipgloss.Style
	KeyDesc   lipgloss.Style
	QuotaText lipgloss.Style

	FrameBox lipgloss.Style
	ErrorBox lipgloss.Style
}

// DefaultTheme returns the Catppuccin Mocha styling palette.
func DefaultTheme() Theme {
	// Catppuccin Mocha Palette
	mauve := lipgloss.Color("#cba6f7")
	blue := lipgloss.Color("#89b4fa")
	green := lipgloss.Color("#a6e3a1")
	red := lipgloss.Color("#f38ba8")
	text := lipgloss.Color("#cdd6f4")
	subtext := lipgloss.Color("#a6adc8")
	overlay := lipgloss.Color("#6c7086")
	surface0 := lipgloss.Color("#313244")
	base := lipgloss.Color("#1e1e2e")

	return Theme{
		BaseHeader: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(surface0),

		RepoName: lipgloss.NewStyle().
			Bold(true).
			Foreground(blue),

		ViewerBadge: lipgloss.NewStyle().
			Foreground(green),

		ActiveTab: lipgloss.NewStyle().
			Bold(true).
			Foreground(base).
			Background(mauve).
			Padding(0, 1),

		InactiveTab: lipgloss.NewStyle().
			Foreground(overlay).
			Padding(0, 1),

		FooterBar: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(surface0),

		KeyHint: lipgloss.NewStyle().
			Bold(true).
			Foreground(text),

		KeyDesc: lipgloss.NewStyle().
			Foreground(overlay),

		QuotaText: lipgloss.NewStyle().
			Foreground(subtext),

		FrameBox: lipgloss.NewStyle().
			Padding(0, 1),

		ErrorBox: lipgloss.NewStyle().
			Foreground(red).
			Bold(true).
			Padding(0, 1),
	}
}
