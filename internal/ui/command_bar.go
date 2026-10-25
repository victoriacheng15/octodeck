package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/victoriacheng15/octodeck/internal/config"
	"github.com/victoriacheng15/octodeck/internal/types"
)

// CommandBarModel provides a k9s-style ":" command bar prompt overlay.
type CommandBarModel struct {
	input  textinput.Model
	active bool
	width  int
}

// NewCommandBarModel constructs a new command bar model initialized with Catppuccin styling.
func NewCommandBarModel() CommandBarModel {
	ti := textinput.New()
	ti.Prompt = ": "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
	ti.Placeholder = "repo <owner/repo>, pr, is, wf, q"
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))
	ti.CharLimit = 120

	return CommandBarModel{
		input:  ti,
		active: false,
		width:  80,
	}
}

// Active reports whether the command bar is currently focused and receiving input.
func (m CommandBarModel) Active() bool {
	return m.active
}

// Activate opens and focuses the command bar.
func (m *CommandBarModel) Activate() tea.Cmd {
	m.active = true
	m.input.SetValue("")
	return m.input.Focus()
}

// Deactivate clears and blurs the command bar.
func (m *CommandBarModel) Deactivate() {
	m.active = false
	m.input.Blur()
	m.input.SetValue("")
}

// SetWidth sets the available display width for the command bar.
func (m *CommandBarModel) SetWidth(width int) {
	m.width = width
	m.input.Width = width - 4
}

// Update handles keystrokes when the command bar is active.
func (m CommandBarModel) Update(msg tea.Msg) (CommandBarModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEscape:
			m.Deactivate()
			return m, nil

		case tea.KeyEnter:
			cmdString := strings.TrimSpace(m.input.Value())
			m.Deactivate()
			return m, ParseCommand(cmdString)
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the command bar input row.
func (m CommandBarModel) View() string {
	if !m.active {
		return ""
	}

	style := lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#313244")).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("#45475a")).
		Padding(0, 1)

	return style.Render(m.input.View())
}

// ParseCommand parses a typed command and returns the corresponding tea.Cmd.
func ParseCommand(input string) tea.Cmd {
	clean := strings.TrimSpace(input)
	clean = strings.TrimPrefix(clean, ":")
	if clean == "" {
		return nil
	}

	parts := strings.Fields(clean)
	verb := strings.ToLower(parts[0])

	switch verb {
	case "repo", "r":
		if len(parts) < 2 {
			return func() tea.Msg {
				return types.ErrMsg{Err: fmt.Errorf("usage: :repo <owner/repo>")}
			}
		}
		repoTarget := parts[1]
		repoCtx, err := config.ResolveRepo(repoTarget)
		if err != nil {
			return func() tea.Msg {
				return types.ErrMsg{Err: fmt.Errorf("invalid repository %q: %w", repoTarget, err)}
			}
		}
		return func() tea.Msg {
			return types.RepoSwitchedMsg{Repo: repoCtx}
		}

	case "pr", "prs", "1":
		return func() tea.Msg {
			return types.NavigateMsg{Target: types.ViewPRs}
		}

	case "is", "issues", "2":
		return func() tea.Msg {
			return types.NavigateMsg{Target: types.ViewIssues}
		}

	case "wf", "workflows", "3":
		return func() tea.Msg {
			return types.NavigateMsg{Target: types.ViewWorkflows}
		}

	case "q", "quit":
		return tea.Quit

	default:
		return func() tea.Msg {
			return types.ErrMsg{Err: fmt.Errorf("unknown command :%s (try :repo, :pr, :is, :wf, :q)", verb)}
		}
	}
}
