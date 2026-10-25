package types

import (
	"errors"
	"testing"
)

func TestViewState_String(t *testing.T) {
	tests := []struct {
		name     string
		state    ViewState
		expected string
	}{
		{
			name:     "PRs view",
			state:    ViewPRs,
			expected: "Pull Requests",
		},
		{
			name:     "Issues view",
			state:    ViewIssues,
			expected: "Issues",
		},
		{
			name:     "Workflows view",
			state:    ViewWorkflows,
			expected: "Workflows",
		},
		{
			name:     "Job Logs view",
			state:    ViewJobLogs,
			expected: "Job Logs",
		},
		{
			name:     "Unknown view",
			state:    ViewState(999),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.expected {
				t.Errorf("ViewState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRepoContext_String(t *testing.T) {
	tests := []struct {
		name     string
		repo     RepoContext
		expected string
	}{
		{
			name:     "empty repo",
			repo:     RepoContext{},
			expected: "",
		},
		{
			name:     "valid repo",
			repo:     RepoContext{Owner: "charmbracelet", Name: "bubbletea"},
			expected: "charmbracelet/bubbletea",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.repo.String()
			if got != tt.expected {
				t.Errorf("RepoContext.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrMsg_Error(t *testing.T) {
	tests := []struct {
		name     string
		msg      ErrMsg
		expected string
	}{
		{
			name:     "nil error",
			msg:      ErrMsg{Err: nil},
			expected: "",
		},
		{
			name:     "with error",
			msg:      ErrMsg{Err: errors.New("failed network call")},
			expected: "failed network call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.Error()
			if got != tt.expected {
				t.Errorf("ErrMsg.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}
