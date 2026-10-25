package views

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/types"
)

// IssuesModel manages the state and table rendering for GitHub issues.
type IssuesModel struct {
	client   api.Client
	table    table.Model
	issues   []types.Issue
	pageInfo types.PageInfo
	loading  bool
	err      error
	width    int
	height   int
}

// NewIssuesModel constructs an IssuesModel initialized with column headers.
func NewIssuesModel(client api.Client) IssuesModel {
	columns := []table.Column{
		{Title: "#", Width: 6},
		{Title: "Title", Width: 40},
		{Title: "Author", Width: 14},
		{Title: "Labels", Width: 20},
		{Title: "Assignees", Width: 16},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)
	t.SetStyles(DefaultTableStyles())

	return IssuesModel{
		client:  client,
		table:   t,
		loading: true,
	}
}

// Init triggers initial fetch of issues.
func (m IssuesModel) Init() tea.Cmd {
	return m.FetchIssues("")
}

// FetchIssues dispatches an asynchronous command querying open issues.
func (m IssuesModel) FetchIssues(cursor string) tea.Cmd {
	return func() tea.Msg {
		issues, pageInfo, err := m.client.ListIssues(context.Background(), cursor, 30)
		return types.IssuesLoadedMsg{
			Issues:   issues,
			PageInfo: pageInfo,
			Err:      err,
		}
	}
}

// SetDimensions updates the table viewport width and height responsively.
func (m *IssuesModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height

	usableWidth := width - 4
	if usableWidth < 60 {
		usableWidth = 60
	}

	numWidth := 6
	remaining := usableWidth - numWidth
	if remaining < 30 {
		remaining = 30
	}

	authorWidth := remaining * 18 / 100
	if authorWidth < 12 {
		authorWidth = 12
	} else if authorWidth > 20 {
		authorWidth = 20
	}

	assigneesWidth := remaining * 18 / 100
	if assigneesWidth < 12 {
		assigneesWidth = 12
	} else if assigneesWidth > 22 {
		assigneesWidth = 22
	}

	labelsWidth := remaining * 24 / 100
	if labelsWidth < 16 {
		labelsWidth = 16
	} else if labelsWidth > 32 {
		labelsWidth = 32
	}

	titleWidth := remaining - authorWidth - assigneesWidth - labelsWidth
	if titleWidth < 15 {
		titleWidth = 15
	}

	m.table.SetColumns([]table.Column{
		{Title: "#", Width: numWidth},
		{Title: "Title", Width: titleWidth},
		{Title: "Author", Width: authorWidth},
		{Title: "Labels", Width: labelsWidth},
		{Title: "Assignees", Width: assigneesWidth},
	})

	m.table.SetWidth(width)
	m.table.SetHeight(height - 1)
}

// Update handles table navigation, pagination keystrokes, and loaded events.
func (m IssuesModel) Update(msg tea.Msg) (IssuesModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case types.IssuesLoadedMsg:
		m.loading = false
		m.err = msg.Err
		if msg.Err != nil {
			return m, nil
		}

		m.issues = msg.Issues
		m.pageInfo = msg.PageInfo

		var rows []table.Row
		for _, issue := range m.issues {
			rows = append(rows, table.Row{
				"#" + strconv.Itoa(issue.Number),
				issue.Title,
				issue.Author,
				FormatLabels(issue.Labels),
				FormatAssignees(issue.Assignees),
			})
		}
		m.table.SetRows(rows)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "]":
			if m.pageInfo.HasNextPage {
				m.loading = true
				return m, m.FetchIssues(m.pageInfo.EndCursor)
			}
		case "[":
			if m.pageInfo.HasPreviousPage {
				m.loading = true
				return m, m.FetchIssues(m.pageInfo.StartCursor)
			}
		case "r":
			m.loading = true
			return m, m.FetchIssues("")
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the table or state message.
func (m IssuesModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  Error loading issues: %v\n  Press 'r' to retry.", m.err)
	}

	if m.loading && len(m.issues) == 0 {
		return "\n  Loading issues..."
	}

	if len(m.issues) == 0 {
		return "\n  No open issues found."
	}

	return m.table.View()
}
