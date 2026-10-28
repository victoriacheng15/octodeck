package views

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/victoriacheng15/octodeck/internal/types"
)

func TestLogsModel_StateTransitions(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(m *LogsModel)
		inputMsg   tea.Msg
		assertPost func(t *testing.T, m LogsModel, cmd tea.Cmd)
	}{
		{
			name: "initial state has empty content and loading false",
			setup: func(m *LogsModel) {
			},
			inputMsg: nil,
			assertPost: func(t *testing.T, m LogsModel, cmd tea.Cmd) {
				if m.loading {
					t.Errorf("expected loading = false initially")
				}
				if m.JobName() != "" {
					t.Errorf("expected empty jobName initially")
				}
			},
		},
		{
			name: "SetLoading updates status and job metadata",
			setup: func(m *LogsModel) {
				m.SetLoading(404, "integration-tests")
			},
			inputMsg: nil,
			assertPost: func(t *testing.T, m LogsModel, cmd tea.Cmd) {
				if !m.loading {
					t.Errorf("expected loading = true")
				}
				if m.JobName() != "integration-tests" {
					t.Errorf("JobName() = %q, want integration-tests", m.JobName())
				}
				rendered := m.View()
				if !strings.Contains(rendered, "integration-tests") {
					t.Errorf("View() missing job name in loading state: %s", rendered)
				}
			},
		},
		{
			name: "SetLogs sanitizes, formats line numbers, and populates viewport",
			setup: func(m *LogsModel) {
				content := "2026-10-27T12:00:00Z Step 1: Initialize\n2026-10-27T12:00:01Z Step 2: \x1b[32mSuccess\x1b[0m\n"
				m.SetLogs(404, "integration-tests", content)
			},
			inputMsg: nil,
			assertPost: func(t *testing.T, m LogsModel, cmd tea.Cmd) {
				if m.loading {
					t.Errorf("expected loading = false after SetLogs")
				}
				if len(m.lines) != 2 {
					t.Fatalf("len(m.lines) = %d, want 2", len(m.lines))
				}
				rendered := m.View()
				if !strings.Contains(rendered, "integration-tests") {
					t.Errorf("View() missing job name: %s", rendered)
				}
				if !strings.Contains(rendered, "1 │ ") {
					t.Errorf("View() missing line 1 gutter: %s", rendered)
				}
				if !strings.Contains(rendered, "Step 2:") {
					t.Errorf("View() missing log line content: %s", rendered)
				}
			},
		},
		{
			name: "esc key navigates back to ViewWorkflows",
			setup: func(m *LogsModel) {
				m.SetLogs(404, "integration-tests", "line 1\nline 2")
			},
			inputMsg: tea.KeyMsg{Type: tea.KeyEscape},
			assertPost: func(t *testing.T, m LogsModel, cmd tea.Cmd) {
				if cmd == nil {
					t.Fatalf("expected non-nil command on esc key")
				}
				msg := cmd()
				navMsg, ok := msg.(types.NavigateMsg)
				if !ok {
					t.Fatalf("expected NavigateMsg, got %T", msg)
				}
				if navMsg.Target != types.ViewWorkflows {
					t.Errorf("navMsg.Target = %v, want ViewWorkflows", navMsg.Target)
				}
			},
		},
		{
			name: "SetError renders error message",
			setup: func(m *LogsModel) {
				m.SetError(errors.New("log stream expired"))
			},
			inputMsg: nil,
			assertPost: func(t *testing.T, m LogsModel, cmd tea.Cmd) {
				rendered := m.View()
				if !strings.Contains(rendered, "log stream expired") {
					t.Errorf("View() missing error message: %s", rendered)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewLogsModel()
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
