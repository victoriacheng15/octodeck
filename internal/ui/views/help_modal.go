package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpEntry describes a single command or shortcut in the help menu.
type HelpEntry struct {
	Command string
	Desc    string
}

// HelpSection groups related commands together.
type HelpSection struct {
	Title   string
	Entries []HelpEntry
}

// DefaultHelpSections returns the full organized command registry for octodeck.
func DefaultHelpSections() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation (:)",
			Entries: []HelpEntry{
				{Command: ":pr, :prs", Desc: "Switch to Pull Requests view"},
				{Command: ":is, :issues", Desc: "Switch to Issues view"},
				{Command: ":wf, :workflows", Desc: "Switch to Workflows view"},
				{Command: ":repo <owner/repo>", Desc: "Switch target repository"},
			},
		},
		{
			Title: "Triage Actions (:)",
			Entries: []HelpEntry{
				{Command: ":co, :checkout [num]", Desc: "Safely check out PR locally"},
				{Command: ":approve", Desc: "Approve focused pull request"},
				{Command: ":close", Desc: "Close focused PR or Issue"},
				{Command: ":assign", Desc: "Assign authenticated viewer"},
				{Command: ":labels, :l", Desc: "Manage item labels dialog"},
			},
		},
		{
			Title: "Workflow Actions (:)",
			Entries: []HelpEntry{
				{Command: ":rerun", Desc: "Rerun failed workflow jobs"},
				{Command: ":cancel", Desc: "Cancel active workflow run"},
			},
		},
		{
			Title: "Table Navigation & General",
			Entries: []HelpEntry{
				{Command: "↑/↓, j/k", Desc: "Navigate table rows"},
				{Command: "PgUp/PgDn, b/f", Desc: "Page up / down table"},
				{Command: "r", Desc: "Refresh active view"},
				{Command: ":", Desc: "Open command bar"},
				{Command: "?, :help, :h", Desc: "Open this help dialog"},
				{Command: "q, :q, :quit", Desc: "Exit octodeck"},
				{Command: "Esc", Desc: "Close dialog / back"},
			},
		},
	}
}

// HelpModal displays an interactive coordinate-aware overlay listing all commands.
type HelpModal struct {
	active bool
	width  int
	height int
}

// NewHelpModal constructs an inactive help modal.
func NewHelpModal() HelpModal {
	return HelpModal{}
}

// Active reports whether the help modal is open.
func (m HelpModal) Active() bool {
	return m.active
}

// Open activates the help modal overlay.
func (m *HelpModal) Open() {
	m.active = true
}

// Close dismisses the help modal overlay.
func (m *HelpModal) Close() {
	m.active = false
}

// SetDimensions sets available layout bounds.
func (m *HelpModal) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

// Update handles dismissal keystrokes (Esc, q, ?).
func (m HelpModal) Update(msg tea.Msg) (HelpModal, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "q", "?":
			m.Close()
			return m, nil
		}
	}
	return m, nil
}

// RenderOverlay renders the help dialog centered using lipgloss.Place.
func (m HelpModal) RenderOverlay(width, height int) string {
	if !m.active {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorMauve).
		MarginBottom(1)

	sectionTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorBlue).
		MarginTop(1)

	cmdStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorGreen)

	descStyle := lipgloss.NewStyle().
		Foreground(colorText)

	hintStyle := lipgloss.NewStyle().
		Foreground(colorOverlay).
		MarginTop(1)

	var lines []string
	lines = append(lines, titleStyle.Render("Octodeck Command Reference"))

	sections := DefaultHelpSections()
	for _, sec := range sections {
		lines = append(lines, sectionTitleStyle.Render(sec.Title))
		for _, e := range sec.Entries {
			cmdCol := cmdStyle.Width(24).Render(e.Command)
			descCol := descStyle.Render(e.Desc)
			lines = append(lines, "  "+cmdCol+" "+descCol)
		}
	}

	lines = append(lines, hintStyle.Render("Press Esc or q to close"))

	dialogWidth := 66
	if width < 70 {
		dialogWidth = width - 4
		if dialogWidth < 20 {
			dialogWidth = 20
		}
	}

	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMauve).
		Padding(1, 2).
		Width(dialogWidth).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
	)
}
