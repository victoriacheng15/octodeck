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

// PRsModel manages the state and table rendering for GitHub pull requests.
type PRsModel struct {
	client   api.Client
	table    table.Model
	prs      []types.PullRequest
	pageInfo types.PageInfo
	loading  bool
	err      error
	width    int
	height   int
}

// NewPRsModel constructs a PRsModel initialized with column headers.
func NewPRsModel(client api.Client) PRsModel {
	columns := []table.Column{
		{Title: "#", Width: 6},
		{Title: "Title", Width: 38},
		{Title: "Author", Width: 14},
		{Title: "Review", Width: 18},
		{Title: "CI", Width: 16},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)
	t.SetStyles(DefaultTableStyles())

	return PRsModel{
		client:  client,
		table:   t,
		loading: true,
	}
}

// Init triggers initial fetch of pull requests.
func (m PRsModel) Init() tea.Cmd {
	return m.FetchPRs("")
}

// FetchPRs dispatches an asynchronous command querying open pull requests.
func (m PRsModel) FetchPRs(cursor string) tea.Cmd {
	return func() tea.Msg {
		prs, pageInfo, err := m.client.ListPullRequests(context.Background(), cursor, 30)
		return types.PRsLoadedMsg{
			PRs:      prs,
			PageInfo: pageInfo,
			Err:      err,
		}
	}
}

// SetDimensions updates the table viewport width and height responsively.
func (m *PRsModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height

	// Available width inside FrameBox (padding: 1 left, 1 right => 2 chars)
	availableWidth := width - 2
	if availableWidth < 60 {
		availableWidth = 60
	}

	// 5 columns with cell padding (0, 1) and right border => 3 chars overhead per column
	numCols := 5
	colOverhead := numCols * 3
	usableWidth := availableWidth - colOverhead
	if usableWidth < 45 {
		usableWidth = 45
	}

	numWidth := 6
	reviewWidth := 16
	ciWidth := 16

	remaining := usableWidth - numWidth - reviewWidth - ciWidth
	if remaining < 20 {
		remaining = 20
	}

	authorWidth := remaining * 25 / 100
	if authorWidth < 12 {
		authorWidth = 12
	} else if authorWidth > 20 {
		authorWidth = 20
	}

	titleWidth := remaining - authorWidth
	if titleWidth < 15 {
		titleWidth = 15
	}

	m.table.SetColumns([]table.Column{
		{Title: "#", Width: numWidth},
		{Title: "Title", Width: titleWidth},
		{Title: "Author", Width: authorWidth},
		{Title: "Review", Width: reviewWidth},
		{Title: "CI", Width: ciWidth},
	})

	m.table.SetWidth(availableWidth)
	m.table.SetHeight(height - 1)
}

// Update handles table navigation, pagination keystrokes, and loaded events.
func (m PRsModel) Update(msg tea.Msg) (PRsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case types.PRsLoadedMsg:
		m.loading = false
		m.err = msg.Err
		if msg.Err != nil {
			return m, nil
		}

		m.prs = msg.PRs
		m.pageInfo = msg.PageInfo

		var rows []table.Row
		for _, pr := range m.prs {
			rows = append(rows, table.Row{
				"#" + strconv.Itoa(pr.Number),
				pr.Title,
				pr.Author,
				RenderReviewDecision(pr.ReviewDecision),
				RenderCIStatus(pr.StatusCheckRollup),
			})
		}
		m.table.SetRows(rows)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "]":
			if m.pageInfo.HasNextPage {
				m.loading = true
				return m, m.FetchPRs(m.pageInfo.EndCursor)
			}
		case "[":
			if m.pageInfo.HasPreviousPage {
				m.loading = true
				return m, m.FetchPRs(m.pageInfo.StartCursor)
			}
		case "r":
			m.loading = true
			return m, m.FetchPRs("")
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the table or state message.
func (m PRsModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  Error loading pull requests: %v\n  Press 'r' to retry.", m.err)
	}

	if m.loading && len(m.prs) == 0 {
		return "\n  Loading pull requests..."
	}

	if len(m.prs) == 0 {
		return "\n  No open pull requests found."
	}

	return m.table.View()
}
