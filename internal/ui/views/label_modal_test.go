package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLabelModal_OpenAndClose(t *testing.T) {
	tests := []struct {
		name          string
		number        int
		title         string
		id            string
		isPR          bool
		currentLabels []string
		wantSelected  []string
	}{
		{
			name:          "open issue with existing labels",
			number:        10,
			title:         "Fix login crash",
			id:            "I_10",
			isPR:          false,
			currentLabels: []string{"bug", "ui"},
			wantSelected:  []string{"bug", "ui"},
		},
		{
			name:          "open PR with no initial labels",
			number:        25,
			title:         "Implement mutations",
			id:            "PR_25",
			isPR:          true,
			currentLabels: []string{},
			wantSelected:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modal := NewLabelModal()
			if modal.Active() {
				t.Fatalf("expected new modal to be inactive")
			}

			modal.Open(tt.number, tt.title, tt.id, tt.isPR, tt.currentLabels)
			if !modal.Active() {
				t.Fatalf("expected modal to be active after Open")
			}
			if modal.ItemNumber() != tt.number {
				t.Errorf("ItemNumber = %d, want %d", modal.ItemNumber(), tt.number)
			}
			if modal.ItemID() != tt.id {
				t.Errorf("ItemID = %s, want %s", modal.ItemID(), tt.id)
			}
			if modal.IsPR() != tt.isPR {
				t.Errorf("IsPR = %v, want %v", modal.IsPR(), tt.isPR)
			}

			selected := modal.SelectedLabels()
			if len(selected) != len(tt.wantSelected) {
				t.Fatalf("SelectedLabels len = %d, want %d (%v vs %v)", len(selected), len(tt.wantSelected), selected, tt.wantSelected)
			}

			modal.Close()
			if modal.Active() {
				t.Fatalf("expected modal to be inactive after Close")
			}
		})
	}
}

func TestLabelModal_KeyHandling(t *testing.T) {
	tests := []struct {
		name         string
		keys         []tea.KeyMsg
		expectActive bool
		toggleCount  int
	}{
		{
			name: "esc key dismisses modal",
			keys: []tea.KeyMsg{
				{Type: tea.KeyEscape},
			},
			expectActive: false,
		},
		{
			name: "down arrow moves cursor and space toggles label",
			keys: []tea.KeyMsg{
				{Type: tea.KeyDown},
				{Type: tea.KeySpace},
			},
			expectActive: true,
			toggleCount:  1,
		},
		{
			name: "j navigates down, k navigates up, enter toggles",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("j")},
				{Type: tea.KeyRunes, Runes: []rune("k")},
				{Type: tea.KeyEnter},
			},
			expectActive: true,
			toggleCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modal := NewLabelModal()
			modal.Open(1, "Test Item", "ID_1", false, []string{})

			initialCount := len(modal.SelectedLabels())

			for _, key := range tt.keys {
				var cmd tea.Cmd
				modal, cmd = modal.Update(key)
				if cmd != nil {
					t.Errorf("unexpected cmd from modal.Update")
				}
			}

			if modal.Active() != tt.expectActive {
				t.Errorf("modal.Active() = %v, want %v", modal.Active(), tt.expectActive)
			}

			if tt.expectActive && tt.toggleCount > 0 {
				newCount := len(modal.SelectedLabels())
				if newCount == initialCount {
					t.Errorf("expected selected count to change, initial %d, got %d", initialCount, newCount)
				}
			}
		})
	}
}

func TestLabelModal_RenderOverlay(t *testing.T) {
	tests := []struct {
		name         string
		active       bool
		number       int
		title        string
		isPR         bool
		expectSubstr []string
	}{
		{
			name:         "inactive modal returns empty string",
			active:       false,
			expectSubstr: nil,
		},
		{
			name:   "active modal renders header and labels",
			active: true,
			number: 42,
			title:  "Implement triage shortcuts",
			isPR:   true,
			expectSubstr: []string{
				"Manage Labels: PR #42",
				"enhancement",
				"Space: toggle",
				"Esc: close",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modal := NewLabelModal()
			modal.SetDimensions(80, 24)
			if tt.active {
				modal.Open(tt.number, tt.title, "ID_42", tt.isPR, []string{"enhancement"})
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
