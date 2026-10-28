package ui

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/types"
)

type rootTestWrapper struct {
	RootModel
}

func (w rootTestWrapper) Init() tea.Cmd {
	return w.RootModel.Init()
}

func (w rootTestWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := w.RootModel.Update(msg)
	w.RootModel = updated.(RootModel)
	return w, cmd
}

func (w rootTestWrapper) View() string {
	return w.RootModel.View()
}

func TestModal_TeatestActivationAndDismissal(t *testing.T) {
	mockClient := api.NewMockClient()
	repo := types.RepoContext{Owner: "cli", Name: "go-gh"}
	viewer := types.Viewer{Login: "octocat", ID: "MDQ6VXNlcjE="}
	m := NewRootModel(mockClient, repo, viewer)
	m.Width = 80
	m.Height = 24

	tm := teatest.NewTestModel(
		t,
		rootTestWrapper{RootModel: m},
		teatest.WithInitialTermSize(80, 24),
	)

	// 1. Wait for PRs to load and render from Init()
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("#101"))
	}, teatest.WithDuration(2*time.Second))

	// 2. Open command bar and execute :labels
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	for _, r := range "labels" {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// 3. Assert modal overlay renders
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("Manage Labels")) &&
			bytes.Contains(out, []byte("Space: toggle"))
	}, teatest.WithDuration(2*time.Second))

	// 4. Press 'Esc' to dismiss modal overlay
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})

	// 5. Assert modal overlay is dismissed and table view is restored
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return !bytes.Contains(out, []byte("Manage Labels")) &&
			bytes.Contains(out, []byte("#101"))
	}, teatest.WithDuration(2*time.Second))

	_ = tm.Quit()
}
