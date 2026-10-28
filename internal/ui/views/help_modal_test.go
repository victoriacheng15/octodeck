package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpModal_OpenClose(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		wantActive bool
	}{
		{
			name:       "new modal is inactive",
			action:     "init",
			wantActive: false,
		},
		{
			name:       "open modal activates",
			action:     "open",
			wantActive: true,
		},
		{
			name:       "close modal deactivates",
			action:     "close",
			wantActive: false,
		},
	}

	modal := NewHelpModal()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.action {
			case "open":
				modal.Open()
			case "close":
				modal.Close()
			}
			if modal.Active() != tt.wantActive {
				t.Errorf("modal.Active() = %v, want %v", modal.Active(), tt.wantActive)
			}
		})
	}
}

func TestHelpModal_KeyDismissal(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{
			name: "dismiss via esc",
			key:  tea.KeyMsg{Type: tea.KeyEscape},
		},
		{
			name: "dismiss via q",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
		},
		{
			name: "dismiss via ?",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modal := NewHelpModal()
			modal.Open()
			if !modal.Active() {
				t.Fatalf("expected modal to be active")
			}

			updated, _ := modal.Update(tt.key)
			if updated.Active() {
				t.Errorf("modal still active after pressing %v", tt.key)
			}
		})
	}
}

func TestHelpModal_RenderOverlay(t *testing.T) {
	tests := []struct {
		name         string
		active       bool
		expectSubstr []string
	}{
		{
			name:         "inactive modal renders empty string",
			active:       false,
			expectSubstr: nil,
		},
		{
			name:   "active modal renders header sections and command names",
			active: true,
			expectSubstr: []string{
				"Octodeck Command Reference",
				":pr, :prs",
				":co, :checkout",
				":approve",
				":close",
				":assign",
				":labels",
				"Press Esc or q to close",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modal := NewHelpModal()
			modal.SetDimensions(80, 24)
			if tt.active {
				modal.Open()
			}

			rendered := modal.RenderOverlay(80, 24)
			if !tt.active {
				if rendered != "" {
					t.Errorf("expected empty string for inactive modal, got %q", rendered)
				}
				return
			}

			for _, sub := range tt.expectSubstr {
				if !strings.Contains(rendered, sub) {
					t.Errorf("rendered modal overlay missing substring %q\nRendered:\n%s", sub, rendered)
				}
			}
		})
	}
}
