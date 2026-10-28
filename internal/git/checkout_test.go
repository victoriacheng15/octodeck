package git

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCheckTreeCleanliness(t *testing.T) {
	origRunner := runner
	defer func() { runner = origRunner }()

	tests := []struct {
		name        string
		output      []byte
		cmdErr      error
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "clean working tree returns nil",
			output:      []byte(""),
			cmdErr:      nil,
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "clean tree with whitespace returns nil",
			output:      []byte("   \n\t  \n"),
			cmdErr:      nil,
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "dirty tree with modified file returns ErrDirtyWorkingTree",
			output:      []byte(" M internal/git/checkout.go\n"),
			cmdErr:      nil,
			wantErr:     true,
			expectedErr: "working tree has uncommitted changes",
		},
		{
			name:        "dirty tree with untracked file returns ErrDirtyWorkingTree",
			output:      []byte("?? untracked.go\n"),
			cmdErr:      nil,
			wantErr:     true,
			expectedErr: "working tree has uncommitted changes",
		},
		{
			name:        "git status execution error returns formatted error",
			output:      []byte("fatal: not a git repository"),
			cmdErr:      errors.New("exit status 128"),
			wantErr:     true,
			expectedErr: "git status failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner = func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name != "git" {
					t.Fatalf("expected command git, got %s", name)
				}
				return tt.output, tt.cmdErr
			}

			err := CheckTreeCleanliness(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.expectedErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyRepoMatch(t *testing.T) {
	origRunner := runner
	defer func() { runner = origRunner }()

	tests := []struct {
		name        string
		targetOwner string
		targetName  string
		remoteURL   string
		cmdErr      error
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "matching HTTPS remote returns nil",
			targetOwner: "octocat",
			targetName:  "hello-world",
			remoteURL:   "https://github.com/octocat/hello-world.git\n",
			wantErr:     false,
		},
		{
			name:        "matching SSH remote case insensitive returns nil",
			targetOwner: "Octocat",
			targetName:  "Hello-World",
			remoteURL:   "git@github.com:octocat/hello-world.git\n",
			wantErr:     false,
		},
		{
			name:        "mismatched remote returns error",
			targetOwner: "chaos-mesh",
			targetName:  "chaos-mesh",
			remoteURL:   "https://github.com/octocat/hello-world.git\n",
			wantErr:     true,
			expectedErr: "cannot checkout: local repo is octocat/hello-world, not chaos-mesh/chaos-mesh",
		},
		{
			name:        "git remote failure returns error",
			targetOwner: "octocat",
			targetName:  "hello-world",
			remoteURL:   "",
			cmdErr:      errors.New("fatal: not a git repository"),
			wantErr:     true,
			expectedErr: "not a git repository with an origin remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner = func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" && len(args) == 3 && args[0] == "remote" && args[1] == "get-url" {
					return []byte(tt.remoteURL), tt.cmdErr
				}
				t.Fatalf("unexpected command: %s %v", name, args)
				return nil, nil
			}

			err := VerifyRepoMatch(context.Background(), tt.targetOwner, tt.targetName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.expectedErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCheckoutPR(t *testing.T) {
	origRunner := runner
	defer func() { runner = origRunner }()

	tests := []struct {
		name        string
		targetOwner string
		targetName  string
		prNumber    int
		runnerFn    CommandRunner
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "invalid negative PR number returns error",
			targetOwner: "octocat",
			targetName:  "hello-world",
			prNumber:    -1,
			runnerFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				t.Fatalf("runner should not be called for invalid prNumber")
				return nil, nil
			},
			wantErr:     true,
			expectedErr: "invalid pull request number",
		},
		{
			name:        "invalid zero PR number returns error",
			targetOwner: "octocat",
			targetName:  "hello-world",
			prNumber:    0,
			runnerFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				t.Fatalf("runner should not be called for zero prNumber")
				return nil, nil
			},
			wantErr:     true,
			expectedErr: "invalid pull request number",
		},
		{
			name:        "mismatched repository rejects checkout before checking cleanliness",
			targetOwner: "chaos-mesh",
			targetName:  "chaos-mesh",
			prNumber:    42,
			runnerFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" && args[0] == "remote" {
					return []byte("git@github.com:octocat/hello-world.git\n"), nil
				}
				t.Fatalf("status or gh should not be called when repository mismatches")
				return nil, nil
			},
			wantErr:     true,
			expectedErr: "cannot checkout: local repo is octocat/hello-world, not chaos-mesh/chaos-mesh",
		},
		{
			name:        "dirty working tree rejects checkout",
			targetOwner: "octocat",
			targetName:  "hello-world",
			prNumber:    42,
			runnerFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" && args[0] == "remote" {
					return []byte("git@github.com:octocat/hello-world.git\n"), nil
				}
				if name == "git" && args[0] == "status" {
					return []byte(" M file.go\n"), nil
				}
				t.Fatalf("gh should not be called when working tree is dirty")
				return nil, nil
			},
			wantErr:     true,
			expectedErr: "working tree has uncommitted changes",
		},
		{
			name:        "clean tree executes gh pr checkout successfully",
			targetOwner: "octocat",
			targetName:  "hello-world",
			prNumber:    42,
			runnerFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" && args[0] == "remote" {
					return []byte("git@github.com:octocat/hello-world.git\n"), nil
				}
				if name == "git" && args[0] == "status" {
					return []byte(""), nil
				}
				if name == "gh" {
					if len(args) != 3 || args[0] != "pr" || args[1] != "checkout" || args[2] != "42" {
						t.Fatalf("unexpected gh args: %v", args)
					}
					return []byte("Switched to branch 'pr-42'"), nil
				}
				t.Fatalf("unexpected command: %s", name)
				return nil, nil
			},
			wantErr: false,
		},
		{
			name:        "gh pr checkout command failure returns formatted error",
			targetOwner: "octocat",
			targetName:  "hello-world",
			prNumber:    99,
			runnerFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name == "git" && args[0] == "remote" {
					return []byte("git@github.com:octocat/hello-world.git\n"), nil
				}
				if name == "git" && args[0] == "status" {
					return []byte(""), nil
				}
				if name == "gh" {
					return []byte("pull request #99 not found"), errors.New("exit status 1")
				}
				return nil, nil
			},
			wantErr:     true,
			expectedErr: "gh pr checkout failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner = tt.runnerFn
			err := CheckoutPR(context.Background(), tt.targetOwner, tt.targetName, tt.prNumber)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.expectedErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
