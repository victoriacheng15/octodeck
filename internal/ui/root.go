package ui

import (
	"context"

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

	PRsView       views.PRsModel
	IssuesView    views.IssuesModel
	WorkflowsView views.WorkflowsModel
	LogsView      views.LogsModel
	CommandBar    CommandBarModel
	LabelModal    views.LabelModal
	HelpModal     views.HelpModal

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
	workflowsModel := views.NewWorkflowsModel(client)
	logsModel := views.NewLogsModel()

	prsModel.SetDimensions(80, 20)
	issuesModel.SetDimensions(80, 20)
	workflowsModel.SetDimensions(80, 20)
	logsModel.SetDimensions(80, 20)

	cmdBar := NewCommandBarModel()
	cmdBar.SetWidth(80)

	labelModal := views.NewLabelModal()
	labelModal.SetDimensions(80, 20)

	helpModal := views.NewHelpModal()
	helpModal.SetDimensions(80, 20)

	return RootModel{
		Client:        client,
		Repo:          repo,
		Viewer:        initialViewer,
		ActiveView:    types.ViewPRs,
		PrevView:      types.ViewPRs,
		PRsView:       prsModel,
		IssuesView:    issuesModel,
		WorkflowsView: workflowsModel,
		LogsView:      logsModel,
		CommandBar:    cmdBar,
		LabelModal:    labelModal,
		HelpModal:     helpModal,
		Width:         80,
		Height:        24,
		Theme:         DefaultTheme(),
	}
}

// Init triggers initial asynchronous setup such as viewer hydration and loading views.
func (m RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.InitViewer(),
		m.PRsView.Init(),
		m.IssuesView.Init(),
		m.WorkflowsView.Init(),
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
	if m.HelpModal.Active() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			m.HelpModal, cmd = m.HelpModal.Update(keyMsg)
			return m, cmd
		}
	}

	if m.LabelModal.Active() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			m.LabelModal, cmd = m.LabelModal.Update(keyMsg)
			return m, cmd
		}
	}

	if m.CommandBar.Active() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			m.CommandBar, cmd = m.CommandBar.Update(keyMsg)
			return m, cmd
		}
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.CommandBar.SetWidth(m.Width)
		contentHeight := m.ContentHeight()
		m.PRsView.SetDimensions(m.Width, contentHeight)
		m.IssuesView.SetDimensions(m.Width, contentHeight)
		m.WorkflowsView.SetDimensions(m.Width, contentHeight)
		m.LogsView.SetDimensions(m.Width, contentHeight)
		m.LabelModal.SetDimensions(m.Width, contentHeight)
		m.HelpModal.SetDimensions(m.Width, contentHeight)
		return m, nil

	case types.RepoSwitchedMsg:
		m.Repo = msg.Repo
		m.ErrMsg = ""
		m.Client.SetRepo(msg.Repo)
		return m, tea.Batch(
			m.PRsView.FetchPRs(""),
			m.IssuesView.FetchIssues(""),
			m.WorkflowsView.FetchRuns(""),
		)

	case types.NavigateMsg:
		m.PrevView = m.ActiveView
		m.ActiveView = msg.Target
		return m, nil

	case types.FilterWorkflowsMsg:
		m.PrevView = m.ActiveView
		m.ActiveView = types.ViewWorkflows
		m.WorkflowsView.SetFilter(msg.CommitSHA, msg.PRNumber)
		return m, m.WorkflowsView.FetchRuns(msg.CommitSHA)

	case types.WorkflowsLoadedMsg:
		var cmd tea.Cmd
		m.WorkflowsView, cmd = m.WorkflowsView.Update(msg)
		return m, cmd

	case types.JobsLoadedMsg:
		var cmd tea.Cmd
		m.WorkflowsView, cmd = m.WorkflowsView.Update(msg)
		return m, cmd

	case types.OpenJobLogsMsg:
		m.PrevView = m.ActiveView
		m.ActiveView = types.ViewJobLogs
		m.LogsView.SetLoading(msg.JobID, msg.JobName)
		return m, func() tea.Msg {
			content, err := m.Client.GetJobLogs(context.Background(), msg.JobID, 2*1024*1024)
			return types.LogsLoadedMsg{
				JobID:   msg.JobID,
				Content: content,
				Err:     err,
			}
		}

	case types.LogsLoadedMsg:
		if msg.Err != nil {
			m.LogsView.SetError(msg.Err)
		} else {
			m.LogsView.SetLogs(msg.JobID, m.LogsView.JobName(), msg.Content)
		}
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

	case types.ActionMsg:
		m.ErrMsg = ""
		cmd := m.handleAction(msg.Action, msg.Arg)
		return m, cmd

	case types.MutationResultMsg:
		return m.handleMutationResult(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case ":":
			m.ErrMsg = ""
			cmd := m.CommandBar.Activate()
			return m, cmd
		case "?":
			m.HelpModal.Open()
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
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
	case types.ViewWorkflows:
		m.WorkflowsView, subCmd = m.WorkflowsView.Update(msg)
	case types.ViewJobLogs:
		m.LogsView, subCmd = m.LogsView.Update(msg)
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
	var footer string
	if m.CommandBar.Active() {
		footer = m.CommandBar.View()
	} else {
		footer = RenderFooter(m.Client.GetRateLimit(), m.ErrMsg, m.Width, m.Theme)
	}

	contentHeight := m.ContentHeight()

	var innerContent string
	switch m.ActiveView {
	case types.ViewPRs:
		innerContent = m.PRsView.View()
	case types.ViewIssues:
		innerContent = m.IssuesView.View()
	case types.ViewWorkflows:
		innerContent = m.WorkflowsView.View()
	case types.ViewJobLogs:
		innerContent = m.LogsView.View()
	}

	framedContent := lipgloss.Place(
		m.Width,
		contentHeight,
		lipgloss.Left,
		lipgloss.Top,
		m.Theme.FrameBox.Render(innerContent),
	)

	if m.LabelModal.Active() {
		framedContent = m.LabelModal.RenderOverlay(m.Width, contentHeight)
	} else if m.HelpModal.Active() {
		framedContent = m.HelpModal.RenderOverlay(m.Width, contentHeight)
	}

	return lipgloss.JoinVertical(lipgloss.Top, header, framedContent, footer)
}
