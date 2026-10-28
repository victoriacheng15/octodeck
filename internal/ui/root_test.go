package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/types"
)

func TestRootModel_KeyNavigation(t *testing.T) {
	tests := []struct {
		name         string
		initialState types.ViewState
		key          string
		expectedView types.ViewState
		expectQuit   bool
	}{
		{
			name:         "pressing 1 directly does not change view",
			initialState: types.ViewIssues,
			key:          "1",
			expectedView: types.ViewIssues,
			expectQuit:   false,
		},
		{
			name:         "pressing 2 directly does not change view",
			initialState: types.ViewPRs,
			key:          "2",
			expectedView: types.ViewPRs,
			expectQuit:   false,
		},
		{
			name:         "pressing 3 directly does not change view",
			initialState: types.ViewPRs,
			key:          "3",
			expectedView: types.ViewPRs,
			expectQuit:   false,
		},
		{
			name:         "press esc returns to previous view",
			initialState: types.ViewWorkflows,
			key:          "esc",
			expectedView: types.ViewPRs,
			expectQuit:   false,
		},
		{
			name:         "press q quits",
			initialState: types.ViewPRs,
			key:          "q",
			expectedView: types.ViewPRs,
			expectQuit:   true,
		},
		{
			name:         "press ctrl+c quits",
			initialState: types.ViewPRs,
			key:          "ctrl+c",
			expectedView: types.ViewPRs,
			expectQuit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRootModel(api.NewMockClient(), types.RepoContext{Owner: "cli", Name: "go-gh"}, types.Viewer{Login: "octocat"})
			m.PrevView = types.ViewPRs
			m.ActiveView = tt.initialState

			updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if tt.key == "esc" {
				updatedModel, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
			} else if tt.key == "ctrl+c" {
				updatedModel, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
			}

			root := updatedModel.(RootModel)
			if root.ActiveView != tt.expectedView {
				t.Errorf("ActiveView = %v, want %v", root.ActiveView, tt.expectedView)
			}

			if tt.expectQuit && cmd == nil {
				t.Errorf("expected tea.Quit command, got nil")
			}
		})
	}
}

func TestRootModel_ContentHeight(t *testing.T) {
	tests := []struct {
		name           string
		terminalHeight int
		expectedHeight int
	}{
		{
			name:           "standard 24-line terminal",
			terminalHeight: 24,
			expectedHeight: 20, // 24 - 2 (header) - 2 (footer)
		},
		{
			name:           "tall 50-line terminal",
			terminalHeight: 50,
			expectedHeight: 46,
		},
		{
			name:           "too small terminal guards against negative height",
			terminalHeight: 3,
			expectedHeight: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRootModel(api.NewMockClient(), types.RepoContext{}, types.Viewer{})
			m.Height = tt.terminalHeight

			got := m.ContentHeight()
			if got != tt.expectedHeight {
				t.Errorf("ContentHeight() = %d, want %d", got, tt.expectedHeight)
			}
		})
	}
}

func TestRootModel_Messages(t *testing.T) {
	tests := []struct {
		name   string
		msg    tea.Msg
		assert func(t *testing.T, m RootModel)
	}{
		{
			name: "WindowSizeMsg updates dimensions",
			msg:  tea.WindowSizeMsg{Width: 100, Height: 40},
			assert: func(t *testing.T, m RootModel) {
				if m.Width != 100 || m.Height != 40 {
					t.Errorf("got %dx%d, want 100x40", m.Width, m.Height)
				}
			},
		},
		{
			name: "NavigateMsg updates active and previous view",
			msg:  types.NavigateMsg{Target: types.ViewIssues},
			assert: func(t *testing.T, m RootModel) {
				if m.ActiveView != types.ViewIssues || m.PrevView != types.ViewPRs {
					t.Errorf("unexpected navigation state: active=%v prev=%v", m.ActiveView, m.PrevView)
				}
			},
		},
		{
			name: "ViewerResolvedMsg updates viewer identity",
			msg:  types.ViewerResolvedMsg{Viewer: types.Viewer{Login: "monalisa", ID: "U_1"}},
			assert: func(t *testing.T, m RootModel) {
				if m.Viewer.Login != "monalisa" {
					t.Errorf("got login %q, want %q", m.Viewer.Login, "monalisa")
				}
			},
		},
		{
			name: "ErrMsg updates error banner",
			msg:  types.ErrMsg{Err: errors.New("connection reset")},
			assert: func(t *testing.T, m RootModel) {
				if m.ErrMsg != "connection reset" {
					t.Errorf("got error %q, want %q", m.ErrMsg, "connection reset")
				}
			},
		},
		{
			name: "RepoSwitchedMsg updates active repo context",
			msg: types.RepoSwitchedMsg{
				Repo: types.RepoContext{Owner: "golang", Name: "go", Host: "github.com"},
			},
			assert: func(t *testing.T, m RootModel) {
				if m.Repo.String() != "golang/go" {
					t.Errorf("got repo %q, want %q", m.Repo.String(), "golang/go")
				}
				if m.ErrMsg != "" {
					t.Errorf("expected ErrMsg to be empty, got %q", m.ErrMsg)
				}
			},
		},
		{
			name: "FilterWorkflowsMsg routes to Workflows view with filter",
			msg: types.FilterWorkflowsMsg{
				CommitSHA: "abcdef123456",
				PRNumber:  12,
			},
			assert: func(t *testing.T, m RootModel) {
				if m.ActiveView != types.ViewWorkflows {
					t.Errorf("ActiveView = %v, want ViewWorkflows", m.ActiveView)
				}
				if m.PrevView != types.ViewPRs {
					t.Errorf("PrevView = %v, want ViewPRs", m.PrevView)
				}
			},
		},
		{
			name: "OpenJobLogsMsg routes to JobLogs view",
			msg: types.OpenJobLogsMsg{
				JobID:   999,
				JobName: "unit-tests",
			},
			assert: func(t *testing.T, m RootModel) {
				if m.ActiveView != types.ViewJobLogs {
					t.Errorf("ActiveView = %v, want ViewJobLogs", m.ActiveView)
				}
				if m.LogsView.JobName() != "unit-tests" {
					t.Errorf("LogsView.JobName() = %q, want unit-tests", m.LogsView.JobName())
				}
			},
		},
		{
			name: "LogsLoadedMsg populates logs content in LogsView",
			msg: types.LogsLoadedMsg{
				JobID:   999,
				Content: "Build finished successfully\n",
			},
			assert: func(t *testing.T, m RootModel) {
				rendered := m.LogsView.View()
				if !strings.Contains(rendered, "Build finished successfully") {
					t.Errorf("LogsView missing rendered content: %s", rendered)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRootModel(api.NewMockClient(), types.RepoContext{}, types.Viewer{})
			updatedModel, _ := m.Update(tt.msg)
			tt.assert(t, updatedModel.(RootModel))
		})
	}
}

func TestRootModel_CommandBarInteraction(t *testing.T) {
	tests := []struct {
		name          string
		initialActive bool
		initialValue  string
		inputKey      tea.KeyMsg
		expectActive  bool
		expectQuit    bool
		expectView    types.ViewState
		checkRepo     string
	}{
		{
			name:          "colon activates command bar",
			initialActive: false,
			inputKey:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")},
			expectActive:  true,
			expectQuit:    false,
			expectView:    types.ViewPRs,
		},
		{
			name:          "typing 2 while command bar active does not switch tab",
			initialActive: true,
			inputKey:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")},
			expectActive:  true,
			expectQuit:    false,
			expectView:    types.ViewPRs,
		},
		{
			name:          "typing q while command bar active does not quit",
			initialActive: true,
			inputKey:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
			expectActive:  true,
			expectQuit:    false,
			expectView:    types.ViewPRs,
		},
		{
			name:          "escape deactivates command bar",
			initialActive: true,
			inputKey:      tea.KeyMsg{Type: tea.KeyEscape},
			expectActive:  false,
			expectQuit:    false,
			expectView:    types.ViewPRs,
		},
		{
			name:          "ctrl+c while command bar active quits application",
			initialActive: true,
			inputKey:      tea.KeyMsg{Type: tea.KeyCtrlC},
			expectActive:  true,
			expectQuit:    true,
			expectView:    types.ViewPRs,
		},
		{
			name:          "enter with :repo submits RepoSwitchedMsg and deactivates",
			initialActive: true,
			initialValue:  ":repo charmbracelet/lipgloss",
			inputKey:      tea.KeyMsg{Type: tea.KeyEnter},
			expectActive:  false,
			expectQuit:    false,
			expectView:    types.ViewPRs,
			checkRepo:     "charmbracelet/lipgloss",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRootModel(api.NewMockClient(), types.RepoContext{Owner: "cli", Name: "go-gh"}, types.Viewer{})
			m.ActiveView = types.ViewPRs
			if tt.initialActive {
				m.CommandBar.Activate()
				if tt.initialValue != "" {
					m.CommandBar.input.SetValue(tt.initialValue)
				}
			}

			updated, cmd := m.Update(tt.inputKey)
			root := updated.(RootModel)

			if root.CommandBar.Active() != tt.expectActive {
				t.Errorf("CommandBar.Active() = %v, want %v", root.CommandBar.Active(), tt.expectActive)
			}

			if tt.expectQuit && cmd == nil {
				t.Errorf("expected tea.Quit command, got nil")
			}

			if root.ActiveView != tt.expectView {
				t.Errorf("ActiveView = %v, want %v", root.ActiveView, tt.expectView)
			}

			if tt.checkRepo != "" {
				if cmd == nil {
					t.Fatalf("expected command from enter key, got nil")
				}
				msg := cmd()
				repoMsg, ok := msg.(types.RepoSwitchedMsg)
				if !ok {
					t.Fatalf("expected RepoSwitchedMsg, got %T (%v)", msg, msg)
				}
				// Feed RepoSwitchedMsg back into root model
				updatedAfterSwitch, _ := root.Update(repoMsg)
				root = updatedAfterSwitch.(RootModel)
				if root.Repo.String() != tt.checkRepo {
					t.Errorf("Repo = %s, want %s", root.Repo.String(), tt.checkRepo)
				}
			}
		})
	}
}

func TestRootModel_PRToWorkflowsDrillDown(t *testing.T) {
	tests := []struct {
		name          string
		prs           []types.PullRequest
		pressKey      string
		expectNav     bool
		wantCommitSHA string
		wantPRNum     int
	}{
		{
			name: "press w on selected PR emits FilterWorkflowsMsg with PR commit SHA",
			prs: []types.PullRequest{
				{
					Number:     42,
					Title:      "Fix streaming logs",
					Author:     "octocat",
					HeadRefOid: "9f8e7d6c5b",
				},
			},
			pressKey:      "w",
			expectNav:     true,
			wantCommitSHA: "9f8e7d6c5b",
			wantPRNum:     42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRootModel(api.NewMockClient(), types.RepoContext{}, types.Viewer{})
			m.ActiveView = types.ViewPRs

			// Populate PRsView
			updated, _ := m.Update(types.PRsLoadedMsg{PRs: tt.prs})
			m = updated.(RootModel)

			// Press 'w'
			updatedAfterKey, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.pressKey)})
			m = updatedAfterKey.(RootModel)

			if tt.expectNav {
				if cmd == nil {
					t.Fatalf("expected non-nil command after pressing %q", tt.pressKey)
				}
				msg := cmd()
				filterMsg, ok := msg.(types.FilterWorkflowsMsg)
				if !ok {
					t.Fatalf("expected FilterWorkflowsMsg, got %T (%v)", msg, msg)
				}
				if filterMsg.CommitSHA != tt.wantCommitSHA {
					t.Errorf("CommitSHA = %q, want %q", filterMsg.CommitSHA, tt.wantCommitSHA)
				}
				if filterMsg.PRNumber != tt.wantPRNum {
					t.Errorf("PRNumber = %d, want %d", filterMsg.PRNumber, tt.wantPRNum)
				}

				// Deliver FilterWorkflowsMsg to RootModel
				updatedAfterFilter, _ := m.Update(filterMsg)
				root := updatedAfterFilter.(RootModel)
				if root.ActiveView != types.ViewWorkflows {
					t.Errorf("ActiveView = %v, want ViewWorkflows", root.ActiveView)
				}
			}
		})
	}
}
