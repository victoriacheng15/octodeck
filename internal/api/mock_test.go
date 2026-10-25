package api

import (
	"context"
	"errors"
	"testing"

	"github.com/victoriacheng15/octodeck/internal/types"
)

func TestMockClient_InterfaceCompliance(t *testing.T) {
	var _ Client = (*MockClient)(nil)
}

func TestMockClient_ResolveViewer(t *testing.T) {
	tests := []struct {
		name     string
		override func(ctx context.Context) (types.Viewer, error)
		expected types.Viewer
		wantErr  bool
	}{
		{
			name:     "default fixture",
			override: nil,
			expected: types.Viewer{
				Login: "octocat",
				ID:    "U_kgDOABC123",
				Name:  "Monalisa Octocat",
			},
			wantErr: false,
		},
		{
			name: "custom error",
			override: func(ctx context.Context) (types.Viewer, error) {
				return types.Viewer{}, errors.New("auth failed")
			},
			expected: types.Viewer{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockClient()
			if tt.override != nil {
				client.ViewerFn = tt.override
			}

			got, err := client.ResolveViewer(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveViewer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ResolveViewer() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMockClient_ListPullRequests(t *testing.T) {
	tests := []struct {
		name     string
		override func(ctx context.Context, cursor string, limit int) ([]types.PullRequest, types.PageInfo, error)
		ctx      context.Context
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "default fixtures",
			override: nil,
			ctx:      context.Background(),
			wantLen:  2,
			wantErr:  false,
		},
		{
			name: "context cancelled error",
			override: func(ctx context.Context, cursor string, limit int) ([]types.PullRequest, types.PageInfo, error) {
				return nil, types.PageInfo{}, ctx.Err()
			},
			ctx: func() context.Context {
				c, cancel := context.WithCancel(context.Background())
				cancel()
				return c
			}(),
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockClient()
			if tt.override != nil {
				client.ListPRsFn = tt.override
			}

			got, _, err := client.ListPullRequests(tt.ctx, "", 10)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListPullRequests() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != tt.wantLen {
				t.Errorf("ListPullRequests() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestMockClient_ListIssues(t *testing.T) {
	tests := []struct {
		name     string
		override func(ctx context.Context, cursor string, limit int) ([]types.Issue, types.PageInfo, error)
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "default fixtures",
			override: nil,
			wantLen:  1,
			wantErr:  false,
		},
		{
			name: "custom error",
			override: func(ctx context.Context, cursor string, limit int) ([]types.Issue, types.PageInfo, error) {
				return nil, types.PageInfo{}, errors.New("rate limited")
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockClient()
			if tt.override != nil {
				client.ListIssuesFn = tt.override
			}

			got, _, err := client.ListIssues(context.Background(), "", 10)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListIssues() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != tt.wantLen {
				t.Errorf("ListIssues() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestMockClient_ListWorkflowRuns(t *testing.T) {
	tests := []struct {
		name      string
		commitSHA string
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "filter by commit sha",
			commitSHA: "f6e5d4c3b2a10987",
			wantLen:   1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockClient()
			got, err := client.ListWorkflowRuns(context.Background(), tt.commitSHA, 10)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListWorkflowRuns() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != tt.wantLen {
				t.Errorf("ListWorkflowRuns() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestMockClient_ListWorkflowJobs(t *testing.T) {
	tests := []struct {
		name    string
		runID   int64
		wantLen int
		wantErr bool
	}{
		{
			name:    "jobs for run",
			runID:   987654,
			wantLen: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockClient()
			got, err := client.ListWorkflowJobs(context.Background(), tt.runID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListWorkflowJobs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != tt.wantLen {
				t.Errorf("ListWorkflowJobs() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestMockClient_GetJobLogs(t *testing.T) {
	tests := []struct {
		name     string
		jobID    int64
		maxBytes int64
		wantErr  bool
	}{
		{
			name:     "fetch logs",
			jobID:    555123,
			maxBytes: 2 * 1024 * 1024,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockClient()
			got, err := client.GetJobLogs(context.Background(), tt.jobID, tt.maxBytes)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetJobLogs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got == "" {
				t.Error("GetJobLogs() returned empty string")
			}
		})
	}
}

func TestMockClient_Mutations(t *testing.T) {
	tests := []struct {
		name    string
		execute func(client *MockClient) error
		wantErr bool
	}{
		{
			name: "AssignToSelf happy path",
			execute: func(client *MockClient) error {
				return client.AssignToSelf(context.Background(), "ISSUE_1", "U_1")
			},
			wantErr: false,
		},
		{
			name: "ApprovePullRequest happy path",
			execute: func(client *MockClient) error {
				return client.ApprovePullRequest(context.Background(), "PR_1")
			},
			wantErr: false,
		},
		{
			name: "CloseItem happy path",
			execute: func(client *MockClient) error {
				return client.CloseItem(context.Background(), "PR_1", true)
			},
			wantErr: false,
		},
		{
			name: "RerunFailedJobs happy path",
			execute: func(client *MockClient) error {
				return client.RerunFailedJobs(context.Background(), 123)
			},
			wantErr: false,
		},
		{
			name: "CancelWorkflowRun happy path",
			execute: func(client *MockClient) error {
				return client.CancelWorkflowRun(context.Background(), 123)
			},
			wantErr: false,
		},
		{
			name: "CloseItem error path",
			execute: func(client *MockClient) error {
				client.CloseItemFn = func(ctx context.Context, id string, isPR bool) error {
					return errors.New("mutation error")
				}
				return client.CloseItem(context.Background(), "PR_1", true)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockClient()
			err := tt.execute(client)
			if (err != nil) != tt.wantErr {
				t.Fatalf("mutation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockClient_GetRateLimit(t *testing.T) {
	tests := []struct {
		name          string
		expectedLimit int
	}{
		{
			name:          "default quota",
			expectedLimit: 5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockClient()
			got := client.GetRateLimit()
			if got.Limit != tt.expectedLimit {
				t.Errorf("GetRateLimit().Limit = %v, want %v", got.Limit, tt.expectedLimit)
			}
		})
	}
}
