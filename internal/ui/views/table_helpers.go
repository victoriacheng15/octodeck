package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha Color Palette constants
var (
	colorMauve   = lipgloss.Color("#cba6f7")
	colorGreen   = lipgloss.Color("#a6e3a1")
	colorRed     = lipgloss.Color("#f38ba8")
	colorYellow  = lipgloss.Color("#f9e2af")
	colorOverlay = lipgloss.Color("#6c7086")
	colorSurface = lipgloss.Color("#313244")
	colorBase    = lipgloss.Color("#1e1e2e")
	colorText    = lipgloss.Color("#cdd6f4")
)

// DefaultTableStyles constructs unified styling for all tables.
func DefaultTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorSurface).
		BorderBottom(true).
		Bold(true).
		Foreground(colorMauve)

	s.Cell = s.Cell.
		Foreground(colorText)

	s.Selected = s.Selected.
		Foreground(colorMauve).
		Background(colorSurface).
		Bold(true)

	return s
}

// RenderCIStatus returns an ANSI-colored indicator for pull request CI status.
func RenderCIStatus(status string) string {
	switch strings.ToUpper(status) {
	case "SUCCESS":
		return "\x1b[32m✓ pass\x1b[0m"
	case "FAILURE":
		return "\x1b[31m✗ fail\x1b[0m"
	case "PENDING":
		return "\x1b[33m● pending\x1b[0m"
	default:
		return "\x1b[90m-\x1b[0m"
	}
}

// RenderReviewDecision returns an ANSI-colored label for pull request review status.
func RenderReviewDecision(decision string) string {
	switch strings.ToUpper(decision) {
	case "APPROVED":
		return "\x1b[32mApproved\x1b[0m"
	case "CHANGES_REQUESTED":
		return "\x1b[31mChanges\x1b[0m"
	case "REVIEW_REQUIRED":
		return "\x1b[33mRequired\x1b[0m"
	default:
		return "\x1b[90m-\x1b[0m"
	}
}

// FormatLabels formats a slice of label names into a comma-separated string.
func FormatLabels(labels []string) string {
	if len(labels) == 0 {
		return "-"
	}
	return strings.Join(labels, ", ")
}

// FormatAssignees formats a slice of logins into a comma-separated @login string.
func FormatAssignees(assignees []string) string {
	if len(assignees) == 0 {
		return "-"
	}
	var formatted []string
	for _, a := range assignees {
		formatted = append(formatted, "@"+a)
	}
	return strings.Join(formatted, ", ")
}
