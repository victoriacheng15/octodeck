package views

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/victoriacheng15/octodeck/internal/api"
)

type workflowsTestWrapper struct {
	WorkflowsModel
}

func (w workflowsTestWrapper) Init() tea.Cmd {
	return w.WorkflowsModel.Init()
}

func (w workflowsTestWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := w.WorkflowsModel.Update(msg)
	w.WorkflowsModel = updated
	return w, cmd
}

func (w workflowsTestWrapper) View() string {
	return w.WorkflowsModel.View()
}

func TestWorkflowsModel_TeatestDrillDown(t *testing.T) {
	mockClient := api.NewMockClient()
	m := NewWorkflowsModel(mockClient)
	m.SetDimensions(80, 24)

	tm := teatest.NewTestModel(
		t,
		workflowsTestWrapper{WorkflowsModel: m},
		teatest.WithInitialTermSize(80, 24),
	)

	// 1. Wait for initial workflow runs to load and render from Init()
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("Continuous"))
	}, teatest.WithDuration(2*time.Second))

	// 2. Drill down into Jobs with Enter key (executes FetchJobs via MockClient)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for the jobs list to render in drill-down
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("Lint Go"))
	}, teatest.WithDuration(2*time.Second))

	// 3. Return to Runs with Escape
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})

	// Wait for Runs level to re-render
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("Continuous"))
	}, teatest.WithDuration(2*time.Second))

	_ = tm.Quit()
}
