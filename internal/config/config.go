package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/auth"
	ghconfig "github.com/cli/go-gh/v2/pkg/config"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/victoriacheng15/octodeck/internal/types"
)

// Config encapsulates runtime configuration resolved from CLI flags, environment, and git.
type Config struct {
	Repo   types.RepoContext
	Viewer types.Viewer
	Debug  bool
}

// ResolveHost returns the provided host if non-empty, otherwise gh default host, falling back to "github.com".
func ResolveHost(host string) string {
	if host != "" {
		return host
	}
	if defaultHost, _ := auth.DefaultHost(); defaultHost != "" {
		return defaultHost
	}
	return "github.com"
}

// CheckAuth verifies that a valid token exists for the target host.
func CheckAuth(host string) (string, error) {
	targetHost := ResolveHost(host)
	token, _ := auth.TokenForHost(targetHost)
	if token != "" {
		return token, nil
	}
	return "", fmt.Errorf("no GitHub authentication found for host %q.\n\nTo authenticate, run:\n  gh auth login\n\nAlternatively, export a personal access token:\n  export GH_TOKEN=\"<your_token>\"", targetHost)
}

// ResolveOptimisticUser inspects hosts.yml synchronously to display @username on frame 0.
func ResolveOptimisticUser(host string) string {
	targetHost := ResolveHost(host)
	cfg, err := ghconfig.Read(nil)
	if err != nil {
		return ""
	}
	user, _ := cfg.Get([]string{"hosts", targetHost, "user"})
	return user
}

// ResolveRepo determines the target repository context from flag or local git remotes.
func ResolveRepo(repoFlag string) (types.RepoContext, error) {
	if repoFlag != "" {
		r, err := repository.Parse(repoFlag)
		if err != nil {
			return types.RepoContext{}, fmt.Errorf("invalid repository %q: %w", repoFlag, err)
		}
		return types.RepoContext{
			Owner: r.Owner,
			Name:  r.Name,
			Host:  r.Host,
		}, nil
	}

	r, err := repository.Current()
	if err != nil {
		return types.RepoContext{
			Host: ResolveHost(""),
		}, nil
	}

	return types.RepoContext{
		Owner: r.Owner,
		Name:  r.Name,
		Host:  r.Host,
	}, nil
}

// Load parses CLI flags, validates authentication, and discovers repository context.
func Load() (Config, error) {
	return LoadWithArgs(os.Args[1:])
}

// LoadWithArgs allows deterministic testing of flag parsing and configuration loading.
func LoadWithArgs(args []string) (Config, error) {
	fs := flag.NewFlagSet("octodeck", flag.ContinueOnError)
	var repoFlag string
	fs.StringVar(&repoFlag, "R", "", "Target repository in [HOST/]OWNER/REPO format")
	fs.StringVar(&repoFlag, "repo", "", "Target repository in [HOST/]OWNER/REPO format")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	repoCtx, err := ResolveRepo(repoFlag)
	if err != nil {
		return Config{}, err
	}

	targetHost := ResolveHost(repoCtx.Host)
	if _, err := CheckAuth(targetHost); err != nil {
		return Config{}, err
	}

	viewer := types.Viewer{
		Login: ResolveOptimisticUser(targetHost),
	}

	return Config{
		Repo:   repoCtx,
		Viewer: viewer,
		Debug:  os.Getenv("DEBUG") == "1",
	}, nil
}
