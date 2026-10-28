package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/types"
)

func newTestRootModel() RootModel {
	mockClient := api.NewMockClient()
	repo := types.RepoContext{Owner: "octocat", Name: "hello-world"}
	viewer := types.Viewer{Login: "octocat", ID: "MDQ6VXNlcjE="}
	m := NewRootModel(mockClient, repo, viewer)

	// Populate initial data for PRs
	m.PRsView, _ = m.PRsView.Update(types.PRsLoadedMsg{
		PRs: []types.PullRequest{
			{
				ID:             "PR_1",
				Number:         101,
				Title:          "Support triage keys",
				Author:         "octocat",
				HeadRefOid:     "abc1234",
				ReviewDecision: "REVIEW_REQUIRED",
			},
		},
	})

	// Populate initial data for Issues
	m.IssuesView, _ = m.IssuesView.Update(types.IssuesLoadedMsg{
		Issues: []types.Issue{
			{
				ID:        "ISSUE_1",
				Number:    202,
				Title:     "Add vertical borders",
				Author:    "octocat",
				Labels:    []string{"ui"},
				Assignees: []string{},
			},
		},
	})

	// Populate initial data for Workflows
	m.WorkflowsView, _ = m.WorkflowsView.Update(types.WorkflowsLoadedMsg{
		Runs: []types.WorkflowRun{
			{
				ID:         303,
				Name:       "Continuous Integration",
				Status:     "completed",
				Conclusion: "failure",
				HeadSha:    "fedcba9",
				HeadBranch: "main",
				Event:      "push",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
	})

	return m
}

func TestRootModel_HandleMutationResult(t *testing.T) {
	tests := []struct {
		name         string
		activeView   types.ViewState
		resultMsg    types.MutationResultMsg
		wantErrMsg   string
		expectReload bool
	}{
		{
			name:       "error in mutation sets ErrMsg and halts refresh",
			activeView: types.ViewPRs,
			resultMsg: types.MutationResultMsg{
				Operation: "approve",
				Success:   false,
				Err:       errors.New("graphql permission denied"),
			},
			wantErrMsg:   "graphql permission denied",
			expectReload: false,
		},
		{
			name:       "successful mutation in PRs triggers reload and clears error",
			activeView: types.ViewPRs,
			resultMsg: types.MutationResultMsg{
				Operation: "assign",
				Success:   true,
				Message:   "Assigned @octocat to PR #101",
			},
			wantErrMsg:   "",
			expectReload: true,
		},
		{
			name:       "successful mutation in Issues triggers reload",
			activeView: types.ViewIssues,
			resultMsg: types.MutationResultMsg{
				Operation: "close",
				Success:   true,
				Message:   "Closed Issue #202",
			},
			wantErrMsg:   "",
			expectReload: true,
		},
		{
			name:       "successful mutation in Workflows triggers reload",
			activeView: types.ViewWorkflows,
			resultMsg: types.MutationResultMsg{
				Operation: "rerun",
				Success:   true,
				Message:   "Triggered rerun for workflow run #303",
			},
			wantErrMsg:   "",
			expectReload: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestRootModel()
			m.ActiveView = tt.activeView
			m.ErrMsg = "existing error"

			updatedModel, cmd := m.Update(tt.resultMsg)
			root := updatedModel.(RootModel)

			if root.ErrMsg != tt.wantErrMsg {
				t.Errorf("ErrMsg = %q, want %q", root.ErrMsg, tt.wantErrMsg)
			}

			if tt.expectReload && cmd == nil {
				t.Errorf("expected refresh command on success, got nil")
			} else if !tt.expectReload && cmd != nil {
				t.Errorf("expected no refresh command on error, got %v", cmd)
			}
		})
	}
}

func TestRootModel_PRsActions(t *testing.T) {
	tests := []struct {
		name          string
		action        string
		arg           string
		viewer        types.Viewer
		wantOp        string
		wantSuccess   bool
		wantErrSubstr string
		wantModal     bool
	}{
		{
			name:        "assign action assigns viewer to PR",
			action:      "assign",
			viewer:      types.Viewer{Login: "octocat", ID: "MDQ6VXNlcjE="},
			wantOp:      "assign",
			wantSuccess: true,
		},
		{
			name:          "assign action fails when viewer ID is not hydrated",
			action:        "assign",
			viewer:        types.Viewer{Login: "octocat", ID: ""},
			wantOp:        "assign",
			wantSuccess:   false,
			wantErrSubstr: "viewer identity not hydrated",
		},
		{
			name:        "approve action approves PR",
			action:      "approve",
			viewer:      types.Viewer{Login: "octocat", ID: "MDQ6VXNlcjE="},
			wantOp:      "approve",
			wantSuccess: true,
		},
		{
			name:        "close action closes PR",
			action:      "close",
			viewer:      types.Viewer{Login: "octocat", ID: "MDQ6VXNlcjE="},
			wantOp:      "close",
			wantSuccess: true,
		},
		{
			name:      "labels action opens label modal for PR",
			action:    "labels",
			viewer:    types.Viewer{Login: "octocat", ID: "MDQ6VXNlcjE="},
			wantModal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestRootModel()
			m.ActiveView = types.ViewPRs
			m.Viewer = tt.viewer

			updatedModel, cmd := m.Update(types.ActionMsg{Action: tt.action, Arg: tt.arg})
			root := updatedModel.(RootModel)

			if tt.wantModal {
				if !root.LabelModal.Active() {
					t.Fatalf("expected LabelModal to be active after action %q", tt.action)
				}
				if root.LabelModal.ItemNumber() != 101 {
					t.Errorf("LabelModal.ItemNumber = %d, want 101", root.LabelModal.ItemNumber())
				}
				if !root.LabelModal.IsPR() {
					t.Errorf("LabelModal.IsPR = false, want true")
				}
				return
			}

			if cmd == nil {
				t.Fatalf("expected tea.Cmd from action %q, got nil", tt.action)
			}

			msg := cmd()
			resultMsg, ok := msg.(types.MutationResultMsg)
			if !ok {
				t.Fatalf("expected types.MutationResultMsg, got %T", msg)
			}

			if resultMsg.Operation != tt.wantOp {
				t.Errorf("Operation = %q, want %q", resultMsg.Operation, tt.wantOp)
			}

			if resultMsg.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v (Err: %v)", resultMsg.Success, tt.wantSuccess, resultMsg.Err)
			}

			if tt.wantErrSubstr != "" {
				if resultMsg.Err == nil || !strings.Contains(resultMsg.Err.Error(), tt.wantErrSubstr) {
					t.Errorf("Err = %v, want substring %q", resultMsg.Err, tt.wantErrSubstr)
				}
			}
		})
	}
}

func TestRootModel_IssuesActions(t *testing.T) {
	tests := []struct {
		name          string
		action        string
		viewer        types.Viewer
		wantOp        string
		wantSuccess   bool
		wantErrSubstr string
		wantModal     bool
	}{
		{
			name:        "assign action assigns viewer to Issue",
			action:      "assign",
			viewer:      types.Viewer{Login: "octocat", ID: "MDQ6VXNlcjE="},
			wantOp:      "assign",
			wantSuccess: true,
		},
		{
			name:          "assign action fails when viewer ID is not hydrated",
			action:        "assign",
			viewer:        types.Viewer{Login: "octocat", ID: ""},
			wantOp:        "assign",
			wantSuccess:   false,
			wantErrSubstr: "viewer identity not hydrated",
		},
		{
			name:        "close action closes Issue",
			action:      "close",
			viewer:      types.Viewer{Login: "octocat", ID: "MDQ6VXNlcjE="},
			wantOp:      "close",
			wantSuccess: true,
		},
		{
			name:      "labels action opens label modal for Issue",
			action:    "labels",
			viewer:    types.Viewer{Login: "octocat", ID: "MDQ6VXNlcjE="},
			wantModal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestRootModel()
			m.ActiveView = types.ViewIssues
			m.Viewer = tt.viewer

			updatedModel, cmd := m.Update(types.ActionMsg{Action: tt.action})
			root := updatedModel.(RootModel)

			if tt.wantModal {
				if !root.LabelModal.Active() {
					t.Fatalf("expected LabelModal to be active after action %q", tt.action)
				}
				if root.LabelModal.ItemNumber() != 202 {
					t.Errorf("LabelModal.ItemNumber = %d, want 202", root.LabelModal.ItemNumber())
				}
				if root.LabelModal.IsPR() {
					t.Errorf("LabelModal.IsPR = true, want false")
				}
				return
			}

			if cmd == nil {
				t.Fatalf("expected tea.Cmd from action %q, got nil", tt.action)
			}

			msg := cmd()
			resultMsg, ok := msg.(types.MutationResultMsg)
			if !ok {
				t.Fatalf("expected types.MutationResultMsg, got %T", msg)
			}

			if resultMsg.Operation != tt.wantOp {
				t.Errorf("Operation = %q, want %q", resultMsg.Operation, tt.wantOp)
			}

			if resultMsg.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v (Err: %v)", resultMsg.Success, tt.wantSuccess, resultMsg.Err)
			}

			if tt.wantErrSubstr != "" {
				if resultMsg.Err == nil || !strings.Contains(resultMsg.Err.Error(), tt.wantErrSubstr) {
					t.Errorf("Err = %v, want substring %q", resultMsg.Err, tt.wantErrSubstr)
				}
			}
		})
	}
}

func TestRootModel_WorkflowsActions(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		wantOp      string
		wantSuccess bool
	}{
		{
			name:        "rerun action triggers rerun of failed jobs",
			action:      "rerun",
			wantOp:      "rerun",
			wantSuccess: true,
		},
		{
			name:        "cancel action cancels active workflow run",
			action:      "cancel",
			wantOp:      "cancel",
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestRootModel()
			m.ActiveView = types.ViewWorkflows

			updatedModel, cmd := m.Update(types.ActionMsg{Action: tt.action})
			_ = updatedModel.(RootModel)

			if cmd == nil {
				t.Fatalf("expected tea.Cmd from action %q, got nil", tt.action)
			}

			msg := cmd()
			resultMsg, ok := msg.(types.MutationResultMsg)
			if !ok {
				t.Fatalf("expected types.MutationResultMsg, got %T", msg)
			}

			if resultMsg.Operation != tt.wantOp {
				t.Errorf("Operation = %q, want %q", resultMsg.Operation, tt.wantOp)
			}

			if resultMsg.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", resultMsg.Success, tt.wantSuccess)
			}
		})
	}
}

func TestRootModel_PRCheckout(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		wantOp      string
		wantSuccess bool
	}{
		{
			name:        "checkout action with selected PR",
			arg:         "",
			wantOp:      "checkout",
			wantSuccess: true,
		},
		{
			name:        "checkout action with explicit number",
			arg:         "42",
			wantOp:      "checkout",
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestRootModel()
			m.ActiveView = types.ViewPRs

			updatedModel, cmd := m.Update(types.ActionMsg{Action: "checkout", Arg: tt.arg})
			_ = updatedModel.(RootModel)

			if cmd == nil {
				t.Fatalf("expected tea.Cmd from checkout action, got nil")
			}

			msg := cmd()
			resultMsg, ok := msg.(types.MutationResultMsg)
			if !ok {
				t.Fatalf("expected MutationResultMsg, got %T", msg)
			}

			if resultMsg.Operation != tt.wantOp {
				t.Errorf("Operation = %q, want %q", resultMsg.Operation, tt.wantOp)
			}
		})
	}
}

func TestRootModel_DirectKeysDoNotTriggerMutations(t *testing.T) {
	keys := []string{"a", "v", "x", "c", "l", "R", "X"}

	for _, k := range keys {
		t.Run("key "+k+" is ignored as direct mutation", func(t *testing.T) {
			m := newTestRootModel()
			m.ActiveView = types.ViewPRs

			updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
			root := updatedModel.(RootModel)

			// Modal should not open and no error or mutation should occur
			if root.LabelModal.Active() {
				t.Errorf("key %q opened LabelModal directly; mutations should only run via command bar", k)
			}
		})
	}
}
