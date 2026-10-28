package ui

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/victoriacheng15/octodeck/internal/git"
	"github.com/victoriacheng15/octodeck/internal/types"
	"github.com/victoriacheng15/octodeck/internal/ui/views"
)

// handleMutationResult processes MutationResultMsg and reloads active view data on success.
func (m RootModel) handleMutationResult(msg types.MutationResultMsg) (RootModel, tea.Cmd) {
	if msg.Err != nil {
		m.ErrMsg = msg.Err.Error()
		return m, nil
	}

	m.ErrMsg = ""

	switch m.ActiveView {
	case types.ViewPRs:
		return m, m.PRsView.FetchPRs("")
	case types.ViewIssues:
		return m, m.IssuesView.FetchIssues("")
	case types.ViewWorkflows:
		return m, m.WorkflowsView.FetchRuns(m.WorkflowsView.CommitSHA())
	default:
		return m, nil
	}
}

// actionMap routes triage action verbs to handler methods on RootModel.
var actionMap = map[string]func(m *RootModel, arg string) tea.Cmd{
	"checkout": (*RootModel).actionCheckout,
	"approve":  (*RootModel).actionApprove,
	"close":    (*RootModel).actionClose,
	"assign":   (*RootModel).actionAssign,
	"labels":   (*RootModel).actionLabels,
	"rerun":    (*RootModel).actionRerun,
	"cancel":   (*RootModel).actionCancel,
	"help":     (*RootModel).actionHelp,
}

// handleAction dispatches triage mutations originated from the command bar via actionMap lookup.
func (m *RootModel) handleAction(action string, arg string) tea.Cmd {
	if handler, ok := actionMap[action]; ok {
		return handler(m, arg)
	}
	return func() tea.Msg {
		return types.ErrMsg{Err: fmt.Errorf("unknown action: %s", action)}
	}
}

func (m *RootModel) actionCheckout(arg string) tea.Cmd {
	if m.ActiveView != types.ViewPRs {
		return func() tea.Msg {
			return types.ErrMsg{Err: errors.New(":checkout is only available in Pull Requests view")}
		}
	}
	return m.executeCheckoutPR(arg)
}

func (m *RootModel) actionApprove(_ string) tea.Cmd {
	if m.ActiveView != types.ViewPRs {
		return func() tea.Msg {
			return types.ErrMsg{Err: errors.New(":approve is only available in Pull Requests view")}
		}
	}
	return m.executeApprovePR()
}

func (m *RootModel) actionClose(_ string) tea.Cmd {
	switch m.ActiveView {
	case types.ViewPRs:
		return m.executeClosePR()
	case types.ViewIssues:
		return m.executeCloseIssue()
	default:
		return func() tea.Msg {
			return types.ErrMsg{Err: errors.New(":close is only available in Pull Requests and Issues views")}
		}
	}
}

func (m *RootModel) actionAssign(_ string) tea.Cmd {
	switch m.ActiveView {
	case types.ViewPRs:
		return m.executeAssignSelfPR()
	case types.ViewIssues:
		return m.executeAssignSelfIssue()
	default:
		return func() tea.Msg {
			return types.ErrMsg{Err: errors.New(":assign is only available in Pull Requests and Issues views")}
		}
	}
}

func (m *RootModel) actionLabels(_ string) tea.Cmd {
	switch m.ActiveView {
	case types.ViewPRs:
		m.openLabelModalPR()
		return nil
	case types.ViewIssues:
		m.openLabelModalIssue()
		return nil
	default:
		return func() tea.Msg {
			return types.ErrMsg{Err: errors.New(":labels is only available in Pull Requests and Issues views")}
		}
	}
}

func (m *RootModel) actionRerun(_ string) tea.Cmd {
	if m.ActiveView != types.ViewWorkflows {
		return func() tea.Msg {
			return types.ErrMsg{Err: errors.New(":rerun is only available in Workflows view")}
		}
	}
	if m.WorkflowsView.Level() != views.LevelRuns {
		return func() tea.Msg {
			return types.ErrMsg{Err: errors.New(":rerun must be executed from the runs table")}
		}
	}
	return m.executeRerunJobs()
}

func (m *RootModel) actionCancel(_ string) tea.Cmd {
	if m.ActiveView != types.ViewWorkflows {
		return func() tea.Msg {
			return types.ErrMsg{Err: errors.New(":cancel is only available in Workflows view")}
		}
	}
	if m.WorkflowsView.Level() != views.LevelRuns {
		return func() tea.Msg {
			return types.ErrMsg{Err: errors.New(":cancel must be executed from the runs table")}
		}
	}
	return m.executeCancelWorkflowRun()
}

func (m *RootModel) actionHelp(_ string) tea.Cmd {
	m.HelpModal.Open()
	return nil
}

func (m *RootModel) executeAssignSelfPR() tea.Cmd {
	pr, ok := m.PRsView.SelectedPR()
	if !ok {
		return nil
	}
	if m.Viewer.ID == "" {
		return func() tea.Msg {
			return types.MutationResultMsg{
				Operation: "assign",
				Success:   false,
				Err:       errors.New("viewer identity not hydrated; cannot assign self"),
			}
		}
	}

	client := m.Client
	viewerID := m.Viewer.ID
	viewerLogin := m.Viewer.Login
	return func() tea.Msg {
		err := client.AssignToSelf(context.Background(), pr.ID, viewerID)
		return types.MutationResultMsg{
			Operation: "assign",
			Success:   err == nil,
			Message:   fmt.Sprintf("Assigned @%s to PR #%d", viewerLogin, pr.Number),
			Err:       err,
		}
	}
}

func (m *RootModel) executeApprovePR() tea.Cmd {
	pr, ok := m.PRsView.SelectedPR()
	if !ok {
		return nil
	}
	client := m.Client
	return func() tea.Msg {
		err := client.ApprovePullRequest(context.Background(), pr.ID)
		return types.MutationResultMsg{
			Operation: "approve",
			Success:   err == nil,
			Message:   fmt.Sprintf("Approved PR #%d", pr.Number),
			Err:       err,
		}
	}
}

func (m *RootModel) executeClosePR() tea.Cmd {
	pr, ok := m.PRsView.SelectedPR()
	if !ok {
		return nil
	}
	client := m.Client
	return func() tea.Msg {
		err := client.CloseItem(context.Background(), pr.ID, true)
		return types.MutationResultMsg{
			Operation: "close",
			Success:   err == nil,
			Message:   fmt.Sprintf("Closed PR #%d", pr.Number),
			Err:       err,
		}
	}
}

func (m *RootModel) executeCheckoutPR(arg string) tea.Cmd {
	prNumber := 0
	if arg != "" {
		n, err := strconv.Atoi(arg)
		if err != nil || n <= 0 {
			return func() tea.Msg {
				return types.ErrMsg{Err: fmt.Errorf("invalid pull request number: %s", arg)}
			}
		}
		prNumber = n
	} else {
		pr, ok := m.PRsView.SelectedPR()
		if !ok {
			return func() tea.Msg {
				return types.ErrMsg{Err: errors.New("no pull request selected")}
			}
		}
		prNumber = pr.Number
	}

	owner := m.Repo.Owner
	name := m.Repo.Name
	return func() tea.Msg {
		err := git.CheckoutPR(context.Background(), owner, name, prNumber)
		return types.MutationResultMsg{
			Operation: "checkout",
			Success:   err == nil,
			Message:   fmt.Sprintf("Checked out PR #%d", prNumber),
			Err:       err,
		}
	}
}

func (m *RootModel) openLabelModalPR() {
	pr, ok := m.PRsView.SelectedPR()
	if !ok {
		return
	}
	m.LabelModal.Open(pr.Number, pr.Title, pr.ID, true, pr.Labels)
}

func (m *RootModel) executeAssignSelfIssue() tea.Cmd {
	issue, ok := m.IssuesView.SelectedIssue()
	if !ok {
		return nil
	}
	if m.Viewer.ID == "" {
		return func() tea.Msg {
			return types.MutationResultMsg{
				Operation: "assign",
				Success:   false,
				Err:       errors.New("viewer identity not hydrated; cannot assign self"),
			}
		}
	}

	client := m.Client
	viewerID := m.Viewer.ID
	viewerLogin := m.Viewer.Login
	return func() tea.Msg {
		err := client.AssignToSelf(context.Background(), issue.ID, viewerID)
		return types.MutationResultMsg{
			Operation: "assign",
			Success:   err == nil,
			Message:   fmt.Sprintf("Assigned @%s to Issue #%d", viewerLogin, issue.Number),
			Err:       err,
		}
	}
}

func (m *RootModel) executeCloseIssue() tea.Cmd {
	issue, ok := m.IssuesView.SelectedIssue()
	if !ok {
		return nil
	}
	client := m.Client
	return func() tea.Msg {
		err := client.CloseItem(context.Background(), issue.ID, false)
		return types.MutationResultMsg{
			Operation: "close",
			Success:   err == nil,
			Message:   fmt.Sprintf("Closed Issue #%d", issue.Number),
			Err:       err,
		}
	}
}

func (m *RootModel) openLabelModalIssue() {
	issue, ok := m.IssuesView.SelectedIssue()
	if !ok {
		return
	}
	m.LabelModal.Open(issue.Number, issue.Title, issue.ID, false, issue.Labels)
}

func (m *RootModel) executeRerunJobs() tea.Cmd {
	run, ok := m.WorkflowsView.SelectedRun()
	if !ok {
		return nil
	}
	client := m.Client
	runID := run.ID
	return func() tea.Msg {
		err := client.RerunFailedJobs(context.Background(), runID)
		return types.MutationResultMsg{
			Operation: "rerun",
			Success:   err == nil,
			Message:   fmt.Sprintf("Triggered rerun for workflow run #%d", runID),
			Err:       err,
		}
	}
}

func (m *RootModel) executeCancelWorkflowRun() tea.Cmd {
	run, ok := m.WorkflowsView.SelectedRun()
	if !ok {
		return nil
	}
	client := m.Client
	runID := run.ID
	return func() tea.Msg {
		err := client.CancelWorkflowRun(context.Background(), runID)
		return types.MutationResultMsg{
			Operation: "cancel",
			Success:   err == nil,
			Message:   fmt.Sprintf("Cancelled workflow run #%d", runID),
			Err:       err,
		}
	}
}
