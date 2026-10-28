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
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("#45475a")).
		Padding(0, 1)

	return style.Render(m.input.View())
}

// commandMap routes command names and aliases to their execution handlers.
var commandMap = map[string]func(parts []string) tea.Cmd{
	"repo": handleRepoCommand,
	"r":    handleRepoCommand,

	"pr":  navCommand(types.ViewPRs),
	"prs": navCommand(types.ViewPRs),
	"1":   navCommand(types.ViewPRs),

	"is":     navCommand(types.ViewIssues),
	"issues": navCommand(types.ViewIssues),
	"2":      navCommand(types.ViewIssues),

	"wf":        navCommand(types.ViewWorkflows),
	"workflows": navCommand(types.ViewWorkflows),
	"3":         navCommand(types.ViewWorkflows),

	"q":    func([]string) tea.Cmd { return tea.Quit },
	"quit": func([]string) tea.Cmd { return tea.Quit },

	"checkout": actionCommand("checkout"),
	"co":       actionCommand("checkout"),
	"approve":  actionCommand("approve"),
	"close":    actionCommand("close"),
	"assign":   actionCommand("assign"),
	"labels":   actionCommand("labels"),
	"label":    actionCommand("labels"),
	"l":        actionCommand("labels"),
	"rerun":    actionCommand("rerun"),
	"cancel":   actionCommand("cancel"),

	"help": actionCommand("help"),
	"h":    actionCommand("help"),
	"?":    actionCommand("help"),
}

func navCommand(target types.ViewState) func([]string) tea.Cmd {
	return func([]string) tea.Cmd {
		return func() tea.Msg {
			return types.NavigateMsg{Target: target}
		}
	}
}

func actionCommand(action string) func([]string) tea.Cmd {
	return func(parts []string) tea.Cmd {
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}
		return func() tea.Msg {
			return types.ActionMsg{Action: action, Arg: arg}
		}
	}
}

func handleRepoCommand(parts []string) tea.Cmd {
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

	if handler, ok := commandMap[verb]; ok {
		return handler(parts)
	}

	return func() tea.Msg {
		return types.ErrMsg{Err: fmt.Errorf("unknown command :%s (try :repo, :pr, :is, :wf, :co, :close, :approve, :assign, :labels, :q)", verb)}
	}
}
