package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/victoriacheng15/octodeck/internal/types"
)

// LogsModel provides a scrollable, line-numbered viewport for inspecting Actions job logs.
type LogsModel struct {
	viewport   viewport.Model
	jobID      int64
	jobName    string
	rawContent string
	lines      []string
	loading    bool
	err        error
	width      int
	height     int
}

// NewLogsModel constructs an empty LogsModel initialized with a 80x20 viewport.
func NewLogsModel() LogsModel {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)

	return LogsModel{
		viewport: vp,
		loading:  false,
		width:    80,
		height:   20,
	}
}

// SetDimensions sets viewport width and height responsively.
func (m *LogsModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height

	contentHeight := height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	m.viewport.Width = width - 2
	m.viewport.Height = contentHeight
}

// SetLoading flags the viewer as fetching logs for a specific job.
func (m *LogsModel) SetLoading(jobID int64, jobName string) {
	m.jobID = jobID
	m.jobName = jobName
	m.loading = true
	m.err = nil
	m.rawContent = ""
	m.lines = nil
}

// JobName returns the active job name.
func (m LogsModel) JobName() string {
	return m.jobName
}

// SetLogs parses, sanitizes, numbers, and sets plain-text log content in the viewport.
func (m *LogsModel) SetLogs(jobID int64, jobName, content string) {
	m.jobID = jobID
	m.jobName = jobName
	m.rawContent = content
	m.loading = false
	m.err = nil

	sanitizedLines := SanitizeLogs(content)
	m.lines = sanitizedLines

	totalLines := len(sanitizedLines)
	if totalLines == 0 {
		m.viewport.SetContent("  (No log output available for this job)")
		return
	}

	gutterWidth := len(strconv.Itoa(totalLines))
	if gutterWidth < 2 {
		gutterWidth = 2
	}

	numStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))
	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#313244"))

	var formatted []string
	formatStr := fmt.Sprintf("%%%dd ", gutterWidth)
	bar := barStyle.Render("│ ")

	for i, line := range sanitizedLines {
		lineNum := numStyle.Render(fmt.Sprintf(formatStr, i+1))
		formatted = append(formatted, lineNum+bar+line)
	}

	m.viewport.SetContent(strings.Join(formatted, "\n"))
	m.viewport.GotoTop()
}

// SetError updates the viewer error banner.
func (m *LogsModel) SetError(err error) {
	m.err = err
	m.loading = false
}

// Update processes viewport scrolling, navigation, and exit keys.
func (m LogsModel) Update(msg tea.Msg) (LogsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg {
				return types.NavigateMsg{Target: types.ViewWorkflows}
			}
		case "g":
			m.viewport.GotoTop()
			return m, nil
		case "G":
			m.viewport.GotoBottom()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the log header and scrollable viewport.
func (m LogsModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  Error loading job logs: %v\n  Press Esc to return.", m.err)
	}

	if m.loading {
		return fmt.Sprintf("\n  Fetching logs for job %s (#%d)...", m.jobName, m.jobID)
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cba6f7")).
		Bold(true).
		Padding(0, 1)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6c7086"))

	pct := int(m.viewport.ScrollPercent() * 100)
	title := fmt.Sprintf("Logs > %s (#%d) [%d lines]", m.jobName, m.jobID, len(m.lines))
	hints := fmt.Sprintf(" — Esc: return, j/k: scroll, g/G: top/bottom (%d%%)", pct)

	header := headerStyle.Render(title) + dimStyle.Render(hints)
	return header + "\n" + m.viewport.View()
}
