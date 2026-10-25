package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/types"
	"github.com/victoriacheng15/octodeck/internal/ui/views"
)

// RootModel is the top-level application model managing the active view, layout dimensions, and global keybindings.
type RootModel struct {
	Client     api.Client
	Repo       types.RepoContext
	Viewer     types.Viewer
	ActiveView types.ViewState
	PrevView   types.ViewState

	PRsView    views.PRsModel
	IssuesView views.IssuesModel

	Width  int
	Height int
	Theme  Theme
	ErrMsg string
}

// NewRootModel constructs the top-level Bubble Tea model.
func NewRootModel(client api.Client, repo types.RepoContext, initialViewer types.Viewer) RootModel {
	if client == nil {
		client = api.NewMockClient()
	}

	prsModel := views.NewPRsModel(client)
	issuesModel := views.NewIssuesModel(client)
	prsModel.SetDimensions(80, 20)
	issuesModel.SetDimensions(80, 20)

	return RootModel{
		Client:     client,
		Repo:       repo,
		Viewer:     initialViewer,
		ActiveView: types.ViewPRs,
		PrevView:   types.ViewPRs,
		PRsView:    prsModel,
		IssuesView: issuesModel,
		Width:      80,
		Height:     24,
		Theme:      DefaultTheme(),
	}
}

// Init triggers initial asynchronous setup such as viewer hydration and loading views.
func (m RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.InitViewer(),
		m.PRsView.Init(),
		m.IssuesView.Init(),
	)
}

// InitViewer hydrates authenticated viewer identity.
func (m RootModel) InitViewer() tea.Cmd {
	return func() tea.Msg {
		if m.Client == nil {
			return nil
		}
		viewer, err := m.Client.ResolveViewer(context.Background())
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.ViewerResolvedMsg{Viewer: viewer}
	}
}

// Update processes incoming Bubble Tea messages and state transitions.
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		contentHeight := m.ContentHeight()
		m.PRsView.SetDimensions(m.Width, contentHeight)
		m.IssuesView.SetDimensions(m.Width, contentHeight)
		return m, nil

	case types.NavigateMsg:
		m.PrevView = m.ActiveView
		m.ActiveView = msg.Target
		return m, nil

	case types.ViewerResolvedMsg:
		if msg.Err != nil {
			m.ErrMsg = msg.Err.Error()
		} else {
			m.Viewer = msg.Viewer
		}
		return m, nil

	case types.PRsLoadedMsg:
		var cmd tea.Cmd
		m.PRsView, cmd = m.PRsView.Update(msg)
		return m, cmd

	case types.IssuesLoadedMsg:
		var cmd tea.Cmd
		m.IssuesView, cmd = m.IssuesView.Update(msg)
		return m, cmd

	case types.ErrMsg:
		if msg.Err != nil {
			m.ErrMsg = msg.Err.Error()
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.PrevView = m.ActiveView
			m.ActiveView = types.ViewPRs
			return m, nil
		case "2":
			m.PrevView = m.ActiveView
			m.ActiveView = types.ViewIssues
			return m, nil
		case "3":
			m.PrevView = m.ActiveView
			m.ActiveView = types.ViewWorkflows
			return m, nil
		case "esc":
			if m.ActiveView != m.PrevView {
				m.ActiveView = m.PrevView
			}
			return m, nil
		}
	}

	// Route unhandled keys to the active sub-model
	var subCmd tea.Cmd
	switch m.ActiveView {
	case types.ViewPRs:
		m.PRsView, subCmd = m.PRsView.Update(msg)
	case types.ViewIssues:
		m.IssuesView, subCmd = m.IssuesView.Update(msg)
	}
	if subCmd != nil {
		cmds = append(cmds, subCmd)
	}

	return m, tea.Batch(cmds...)
}

// ContentHeight calculates the vertical line budget available for child views.
func (m RootModel) ContentHeight() int {
	available := m.Height - HeaderHeight() - FooterHeight()
	if available < 1 {
		return 1
	}
	return available
}

// View renders the complete composite screen: header, inner content, and footer.
func (m RootModel) View() string {
	header := RenderHeader(m.Repo, m.Viewer, m.ActiveView, m.Width, m.Theme)
	footer := RenderFooter(m.Client.GetRateLimit(), m.ErrMsg, m.Width, m.Theme)

	contentHeight := m.ContentHeight()

	var innerContent string
	switch m.ActiveView {
	case types.ViewPRs:
		innerContent = m.PRsView.View()
	case types.ViewIssues:
		innerContent = m.IssuesView.View()
	case types.ViewWorkflows:
		innerContent = fmt.Sprintf("Workflows View\nPress 1 for PRs, 2 for Issues.\nActive Repo: %s", m.Repo.String())
	case types.ViewJobLogs:
		innerContent = "Job Logs View\nPress Esc to return."
	}

	framedContent := lipgloss.Place(
		m.Width,
		contentHeight,
		lipgloss.Left,
		lipgloss.Top,
		m.Theme.FrameBox.Render(innerContent),
	)

	return lipgloss.JoinVertical(lipgloss.Top, header, framedContent, footer)
}
