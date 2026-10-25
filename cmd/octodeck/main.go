package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/config"
	"github.com/victoriacheng15/octodeck/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if cfg.Debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err == nil {
			defer f.Close()
		}
	}

	token, err := config.CheckAuth(cfg.Repo.Host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client, err := api.NewGitHubClient(cfg.Repo, api.GitHubClientOptions{
		Host:      cfg.Repo.Host,
		AuthToken: token,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
		os.Exit(1)
	}
	model := ui.NewRootModel(client, cfg.Repo, cfg.Viewer)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
