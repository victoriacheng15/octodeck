package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// CommandRunner executes an external command with context and returns combined output.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

var defaultRunner CommandRunner = func(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// runner is the active command runner, swappable in unit tests.
var runner = defaultRunner

// ErrDirtyWorkingTree indicates git status --porcelain detected uncommitted modifications.
var ErrDirtyWorkingTree = errors.New("working tree has uncommitted changes; commit or stash before checkout")

// CheckTreeCleanliness verifies that the git working tree has no uncommitted changes.
func CheckTreeCleanliness(ctx context.Context) error {
	out, err := runner(ctx, "git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("git status failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}

	trimmed := bytes.TrimSpace(out)
	if len(trimmed) > 0 {
		return ErrDirtyWorkingTree
	}
	return nil
}

// VerifyRepoMatch checks that the local repository origin remote matches targetOwner/targetName.
func VerifyRepoMatch(ctx context.Context, targetOwner, targetName string) error {
	if targetOwner == "" || targetName == "" {
		return nil
	}

	out, err := runner(ctx, "git", "remote", "get-url", "origin")
	if err != nil {
		return fmt.Errorf("cannot checkout: current directory is not a git repository with an origin remote (%w)", err)
	}

	remoteURL := strings.TrimSpace(string(out))
	repo, err := repository.Parse(remoteURL)
	if err != nil {
		return fmt.Errorf("failed to parse git remote URL %q: %w", remoteURL, err)
	}

	if !strings.EqualFold(repo.Owner, targetOwner) || !strings.EqualFold(repo.Name, targetName) {
		return fmt.Errorf("cannot checkout: local repo is %s/%s, not %s/%s", repo.Owner, repo.Name, targetOwner, targetName)
	}
	return nil
}

// CheckoutPR executes gh pr checkout for the target PR number if the working tree is clean and matches target repo.
func CheckoutPR(ctx context.Context, targetOwner, targetName string, prNumber int) error {
	if prNumber <= 0 {
		return fmt.Errorf("invalid pull request number: %d", prNumber)
	}

	if err := VerifyRepoMatch(ctx, targetOwner, targetName); err != nil {
		return err
	}

	if err := CheckTreeCleanliness(ctx); err != nil {
		return err
	}

	out, err := runner(ctx, "gh", "pr", "checkout", strconv.Itoa(prNumber))
	if err != nil {
		return fmt.Errorf("gh pr checkout failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}
