package views

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DefaultRepoLabels provides standard GitHub labels when repo labels are not separately hydrated.
var DefaultRepoLabels = []string{
	"bug",
	"documentation",
	"duplicate",
	"enhancement",
	"good first issue",
	"help wanted",
	"invalid",
	"question",
	"ui",
	"wontfix",
}

// LabelModal provides an interactive coordinate-aware overlay for viewing and toggling item labels.
type LabelModal struct {
	active     bool
	itemNumber int
	itemTitle  string
	itemID     string
	isPR       bool
	labels     []string
	selected   map[string]bool
	cursor     int
	width      int
	height     int
}

// NewLabelModal constructs an inactive label modal.
func NewLabelModal() LabelModal {
	return LabelModal{
		selected: make(map[string]bool),
	}
}

// Active reports whether the modal overlay is currently open.
func (m LabelModal) Active() bool {
	return m.active
}

// ItemNumber returns the number of the item currently being edited.
func (m LabelModal) ItemNumber() int {
	return m.itemNumber
}

// ItemID returns the GraphQL node ID of the item currently being edited.
func (m LabelModal) ItemID() string {
	return m.itemID
}

// IsPR reports whether the target item is a pull request.
func (m LabelModal) IsPR() bool {
	return m.isPR
}

// Open activates the modal for a specific issue or pull request.
func (m *LabelModal) Open(number int, title string, itemID string, isPR bool, currentLabels []string) {
	m.active = true
	m.itemNumber = number
	m.itemTitle = title
	m.itemID = itemID
	m.isPR = isPR
	m.cursor = 0
	m.selected = make(map[string]bool)

	// Build unique sorted list of labels from defaults and current item labels
	labelSet := make(map[string]struct{})
	for _, l := range DefaultRepoLabels {
		labelSet[l] = struct{}{}
	}
	for _, l := range currentLabels {
		labelSet[l] = struct{}{}
		m.selected[l] = true
	}

	m.labels = make([]string, 0, len(labelSet))
	for l := range labelSet {
		m.labels = append(m.labels, l)
	}
	sort.Strings(m.labels)
}

// Close dismisses the modal overlay without applying uncommitted changes.
func (m *LabelModal) Close() {
	m.active = false
}

// SelectedLabels returns the slice of currently checked label names.
func (m LabelModal) SelectedLabels() []string {
	var checked []string
	for _, l := range m.labels {
		if m.selected[l] {
			checked = append(checked, l)
		}
	}
	return checked
}

// SetDimensions updates the viewport bounding box available for modal placement.
func (m *LabelModal) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

// Update handles navigation and toggling keystrokes within the modal.
func (m LabelModal) Update(msg tea.Msg) (LabelModal, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.Close()
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else if len(m.labels) > 0 {
				m.cursor = len(m.labels) - 1
			}
		case "down", "j":
			if m.cursor < len(m.labels)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case " ", "enter":
			if len(m.labels) > 0 && m.cursor >= 0 && m.cursor < len(m.labels) {
				lbl := m.labels[m.cursor]
				m.selected[lbl] = !m.selected[lbl]
			}
		}
	}

	return m, nil
}

// RenderOverlay renders the centered modal box using lipgloss.Place.
func (m LabelModal) RenderOverlay(width, height int) string {
	if !m.active {
		return ""
	}

	boxBorder := lipgloss.RoundedBorder()
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#cba6f7")).
		MarginBottom(1)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6c7086")).
		MarginBottom(1)

	cursorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#89b4fa"))

	checkedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#a6e3a1"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cdd6f4"))

	itemType := "Issue"
	if m.isPR {
		itemType = "PR"
	}

	titleText := fmt.Sprintf("Manage Labels: %s #%d", itemType, m.itemNumber)
	if m.itemTitle != "" {
		displayTitle := m.itemTitle
		if len(displayTitle) > 30 {
			displayTitle = displayTitle[:27] + "..."
		}
		titleText += fmt.Sprintf(" (%s)", displayTitle)
	}

	var lines []string
	lines = append(lines, titleStyle.Render(titleText))
	lines = append(lines, hintStyle.Render("↑/↓/j/k: navigate • Space: toggle • Esc: close"))

	maxVisible := 8
	startIdx := 0
	if m.cursor >= maxVisible {
		startIdx = m.cursor - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(m.labels) {
		endIdx = len(m.labels)
	}

	for i := startIdx; i < endIdx; i++ {
		l := m.labels[i]
		box := "[ ]"
		if m.selected[l] {
			box = checkedStyle.Render("[x]")
		}

		line := fmt.Sprintf("%s %s", box, l)
		if i == m.cursor {
			line = cursorStyle.Render("> " + line)
		} else {
			line = normalStyle.Render("  " + line)
		}
		lines = append(lines, line)
	}

	dialogWidth := 50
	if width < 54 {
		dialogWidth = width - 4
		if dialogWidth < 20 {
			dialogWidth = 20
		}
	}

	modalBox := lipgloss.NewStyle().
		Border(boxBorder).
		BorderForeground(colorMauve).
		Padding(1, 2).
		Width(dialogWidth).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
	)
}
