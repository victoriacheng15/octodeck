package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/victoriacheng15/octodeck/internal/types"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedMsg tea.Msg
		isQuit      bool
		checkErr    func(err error) bool
	}{
		{
			name:        "empty command produces nil",
			input:       "",
			expectedMsg: nil,
		},
		{
			name:        "only colon produces nil",
			input:       ":",
			expectedMsg: nil,
		},
		{
			name:        "whitespace produces nil",
			input:       "   ",
			expectedMsg: nil,
		},
		{
			name:        "switch view to PRs via :pr",
			input:       ":pr",
			expectedMsg: types.NavigateMsg{Target: types.ViewPRs},
		},
		{
			name:        "switch view to PRs via prs",
			input:       "prs",
			expectedMsg: types.NavigateMsg{Target: types.ViewPRs},
		},
		{
			name:        "switch view to Issues via :is",
			input:       ":is",
			expectedMsg: types.NavigateMsg{Target: types.ViewIssues},
		},
		{
			name:        "switch view to Issues via :issues",
			input:       ":issues",
			expectedMsg: types.NavigateMsg{Target: types.ViewIssues},
		},
		{
			name:        "switch view to Workflows via :wf",
			input:       ":wf",
			expectedMsg: types.NavigateMsg{Target: types.ViewWorkflows},
		},
		{
			name:        "switch view to Workflows via :workflows",
			input:       ":workflows",
			expectedMsg: types.NavigateMsg{Target: types.ViewWorkflows},
		},
		{
			name:   "quit via :q",
			input:  ":q",
			isQuit: true,
		},
		{
			name:   "quit via :quit",
			input:  ":quit",
			isQuit: true,
		},
		{
			name:  "switch repo happy path",
			input: ":repo charmbracelet/bubbletea",
			expectedMsg: types.RepoSwitchedMsg{
				Repo: types.RepoContext{
					Owner: "charmbracelet",
					Name:  "bubbletea",
					Host:  "github.com",
				},
			},
		},
		{
			name:  "switch repo alias :r",
			input: ":r cli/go-gh",
			expectedMsg: types.RepoSwitchedMsg{
				Repo: types.RepoContext{
					Owner: "cli",
					Name:  "go-gh",
					Host:  "github.com",
				},
			},
		},
		{
			name:  "switch repo missing argument",
			input: ":repo",
			checkErr: func(err error) bool {
				return err != nil
			},
		},
		{
			name:  "switch repo invalid format",
			input: ":repo a/b/c/d",
			checkErr: func(err error) bool {
				return err != nil
			},
		},
		{
			name:  "unknown command",
			input: ":unknown",
			checkErr: func(err error) bool {
				return err != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ParseCommand(tt.input)
			if tt.input == "" || tt.input == ":" || tt.input == "   " {
				if cmd != nil {
					t.Errorf("ParseCommand(%q) expected nil, got %v", tt.input, cmd)
				}
				return
			}

			if tt.isQuit {
				if cmd == nil {
					t.Fatalf("expected tea.Quit command, got nil")
				}
				return
			}

			if cmd == nil {
				t.Fatalf("ParseCommand(%q) returned nil cmd", tt.input)
			}

			msg := cmd()
			if tt.checkErr != nil {
				errMsg, ok := msg.(types.ErrMsg)
				if !ok || !tt.checkErr(errMsg.Err) {
					t.Errorf("ParseCommand(%q) expected error message, got %v", tt.input, msg)
				}
				return
			}

			if msg != tt.expectedMsg {
				t.Errorf("ParseCommand(%q) = %+v, want %+v", tt.input, msg, tt.expectedMsg)
			}
		})
	}
}

func TestCommandBarModel_StateTransitions(t *testing.T) {
	tests := []struct {
		name       string
		action     func(m *CommandBarModel) tea.Cmd
		assertPost func(t *testing.T, m CommandBarModel, cmd tea.Cmd)
	}{
		{
			name: "initial state is inactive",
			action: func(m *CommandBarModel) tea.Cmd {
				return nil
			},
			assertPost: func(t *testing.T, m CommandBarModel, cmd tea.Cmd) {
				if m.Active() {
					t.Errorf("expected inactive command bar")
				}
				if m.View() != "" {
					t.Errorf("expected empty view when inactive")
				}
			},
		},
		{
			name: "activate focuses input and sets active",
			action: func(m *CommandBarModel) tea.Cmd {
				return m.Activate()
			},
			assertPost: func(t *testing.T, m CommandBarModel, cmd tea.Cmd) {
				if !m.Active() {
					t.Errorf("expected active command bar")
				}
				if m.View() == "" {
					t.Errorf("expected non-empty view when active")
				}
			},
		},
		{
			name: "deactivate clears and blurs input",
			action: func(m *CommandBarModel) tea.Cmd {
				m.Activate()
				m.Deactivate()
				return nil
			},
			assertPost: func(t *testing.T, m CommandBarModel, cmd tea.Cmd) {
				if m.Active() {
					t.Errorf("expected inactive command bar after deactivate")
				}
			},
		},
		{
			name: "escape key deactivates command bar",
			action: func(m *CommandBarModel) tea.Cmd {
				m.Activate()
				updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
				*m = updated
				return cmd
			},
			assertPost: func(t *testing.T, m CommandBarModel, cmd tea.Cmd) {
				if m.Active() {
					t.Errorf("expected inactive command bar after escape")
				}
			},
		},
		{
			name: "enter key executes command and deactivates",
			action: func(m *CommandBarModel) tea.Cmd {
				m.Activate()
				m.input.SetValue(":pr")
				updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
				*m = updated
				return cmd
			},
			assertPost: func(t *testing.T, m CommandBarModel, cmd tea.Cmd) {
				if m.Active() {
					t.Errorf("expected inactive command bar after enter")
				}
				if cmd == nil {
					t.Fatalf("expected non-nil command after enter")
				}
				msg := cmd()
				nav, ok := msg.(types.NavigateMsg)
				if !ok || nav.Target != types.ViewPRs {
					t.Errorf("expected NavigateMsg(PRs), got %v", msg)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCommandBarModel()
			m.SetWidth(80)
			cmd := tt.action(&m)
			tt.assertPost(t, m, cmd)
		})
	}
}
