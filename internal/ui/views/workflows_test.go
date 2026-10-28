package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/types"
)

func TestWorkflowsModel_StateTransitions(t *testing.T) {
	mockClient := api.NewMockClient()

	tests := []struct {
		name       string
		setup      func(m *WorkflowsModel) tea.Cmd
		inputMsg   tea.Msg
		assertPost func(t *testing.T, m WorkflowsModel, cmd tea.Cmd)
	}{
		{
			name: "initial state has loading flag and level runs",
			setup: func(m *WorkflowsModel) tea.Cmd {
				return nil
			},
			inputMsg: nil,
			assertPost: func(t *testing.T, m WorkflowsModel, cmd tea.Cmd) {
				if m.Level() != LevelRuns {
					t.Errorf("Level() = %v, want LevelRuns", m.Level())
				}
				if !m.loading {
					t.Errorf("expected loading = true initially")
				}
			},
		},
		{
			name: "WorkflowsLoadedMsg populates runs table and resets loading",
			setup: func(m *WorkflowsModel) tea.Cmd {
				return nil
			},
			inputMsg: types.WorkflowsLoadedMsg{
				Runs: []types.WorkflowRun{
					{
						ID:         101,
						Name:       "CI / Build & Test",
						Event:      "pull_request",
						Status:     "completed",
						Conclusion: "success",
						HeadSha:    "a1b2c3d4e5",
						HeadBranch: "feat/logs",
					},
				},
				CommitSHA: "",
			},
			assertPost: func(t *testing.T, m WorkflowsModel, cmd tea.Cmd) {
				if m.loading {
					t.Errorf("expected loading = false after load")
				}
				if len(m.runs) != 1 {
					t.Fatalf("len(m.runs) = %d, want 1", len(m.runs))
				}
				if m.runs[0].ID != 101 {
					t.Errorf("runs[0].ID = %d, want 101", m.runs[0].ID)
				}
				rendered := m.View()
				if !strings.Contains(rendered, "CI / Build & Test") {
					t.Errorf("View() missing workflow name: %s", rendered)
				}
			},
		},
		{
			name: "enter key on run drills down and requests jobs",
			setup: func(m *WorkflowsModel) tea.Cmd {
				m.runs = []types.WorkflowRun{
					{
						ID:         202,
						Name:       "Deploy Preview",
						Event:      "push",
						Status:     "completed",
						Conclusion: "success",
					},
				}
				m.runsTable.SetRows([]table.Row{
					{"#202", "Deploy Preview", "push", "✓ success", "main", "a1b2c3d"},
				})
				m.loading = false
				return nil
			},
			inputMsg: tea.KeyMsg{Type: tea.KeyEnter},
			assertPost: func(t *testing.T, m WorkflowsModel, cmd tea.Cmd) {
				if cmd == nil {
					t.Fatalf("expected command from enter key, got nil")
				}
				if m.selectedRun.ID != 202 {
					t.Errorf("selectedRun.ID = %d, want 202", m.selectedRun.ID)
				}
				// Execute command returned by enter key
				msg := cmd()
				jobsMsg, ok := msg.(types.JobsLoadedMsg)
				if !ok {
					t.Fatalf("expected JobsLoadedMsg, got %T", msg)
				}
				if jobsMsg.RunID != 202 {
					t.Errorf("jobsMsg.RunID = %d, want 202", jobsMsg.RunID)
				}
			},
		},
		{
			name: "JobsLoadedMsg sets LevelJobs and renders jobs table",
			setup: func(m *WorkflowsModel) tea.Cmd {
				m.selectedRun = types.WorkflowRun{ID: 202, Name: "Deploy Preview"}
				return nil
			},
			inputMsg: types.JobsLoadedMsg{
				RunID: 202,
				Jobs: []types.Job{
					{
						ID:         303,
						RunID:      202,
						Name:       "build-job",
						Status:     "completed",
						Conclusion: "success",
						Steps: []types.JobStep{
							{Number: 1, Name: "Checkout"},
							{Number: 2, Name: "Build"},
						},
					},
				},
			},
			assertPost: func(t *testing.T, m WorkflowsModel, cmd tea.Cmd) {
				if m.Level() != LevelJobs {
					t.Fatalf("Level() = %v, want LevelJobs", m.Level())
				}
				if len(m.jobs) != 1 {
					t.Fatalf("len(m.jobs) = %d, want 1", len(m.jobs))
				}
				rendered := m.View()
				if !strings.Contains(rendered, "build-job") {
					t.Errorf("View() missing job name: %s", rendered)
				}
				if !strings.Contains(rendered, "2 steps") {
					t.Errorf("View() missing step count: %s", rendered)
				}
			},
		},
		{
			name: "esc in LevelJobs returns to LevelRuns",
			setup: func(m *WorkflowsModel) tea.Cmd {
				m.level = LevelJobs
				m.jobs = []types.Job{{ID: 303, Name: "build-job"}}
				return nil
			},
			inputMsg: tea.KeyMsg{Type: tea.KeyEscape},
			assertPost: func(t *testing.T, m WorkflowsModel, cmd tea.Cmd) {
				if m.Level() != LevelRuns {
					t.Errorf("Level() = %v, want LevelRuns after Esc", m.Level())
				}
			},
		},
		{
			name: "enter in LevelJobs emits OpenJobLogsMsg",
			setup: func(m *WorkflowsModel) tea.Cmd {
				m.level = LevelJobs
				m.jobs = []types.Job{{ID: 303, Name: "build-job"}}
				m.jobsTable.SetRows([]table.Row{
					{"#303", "build-job", "✓ success", "2 steps"},
				})
				return nil
			},
			inputMsg: tea.KeyMsg{Type: tea.KeyEnter},
			assertPost: func(t *testing.T, m WorkflowsModel, cmd tea.Cmd) {
				if cmd == nil {
					t.Fatalf("expected non-nil command on enter in LevelJobs")
				}
				msg := cmd()
				logsMsg, ok := msg.(types.OpenJobLogsMsg)
				if !ok {
					t.Fatalf("expected OpenJobLogsMsg, got %T", msg)
				}
				if logsMsg.JobID != 303 || logsMsg.JobName != "build-job" {
					t.Errorf("unexpected logsMsg: %+v", logsMsg)
				}
			},
		},
		{
			name: "c key in LevelRuns clears commit filter and re-fetches",
			setup: func(m *WorkflowsModel) tea.Cmd {
				m.level = LevelRuns
				m.SetFilter("a1b2c3d4e5", 42)
				return nil
			},
			inputMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")},
			assertPost: func(t *testing.T, m WorkflowsModel, cmd tea.Cmd) {
				if m.commitSHA != "" || m.prNumber != 0 {
					t.Errorf("expected filter cleared, got commitSHA=%q prNumber=%d", m.commitSHA, m.prNumber)
				}
				if cmd == nil {
					t.Fatalf("expected fetch command after clearing filter")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewWorkflowsModel(mockClient)
			m.SetDimensions(100, 24)
			if tt.setup != nil {
				tt.setup(&m)
			}

			var cmd tea.Cmd
			if tt.inputMsg != nil {
				m, cmd = m.Update(tt.inputMsg)
			}
			tt.assertPost(t, m, cmd)
		})
	}
}
