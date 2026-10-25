package ui

import (
	"errors"
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
			name:         "press 1 switches to PRs",
			initialState: types.ViewIssues,
			key:          "1",
			expectedView: types.ViewPRs,
			expectQuit:   false,
		},
		{
			name:         "press 2 switches to Issues",
			initialState: types.ViewPRs,
			key:          "2",
			expectedView: types.ViewIssues,
			expectQuit:   false,
		},
		{
			name:         "press 3 switches to Workflows",
			initialState: types.ViewPRs,
			key:          "3",
			expectedView: types.ViewWorkflows,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRootModel(api.NewMockClient(), types.RepoContext{}, types.Viewer{})
			updatedModel, _ := m.Update(tt.msg)
			tt.assert(t, updatedModel.(RootModel))
		})
	}
}
