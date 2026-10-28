package views

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/victoriacheng15/octodeck/internal/api"
	"github.com/victoriacheng15/octodeck/internal/types"
)

// WorkflowLevel tracks whether the workflows view is displaying runs or jobs drill-down.
type WorkflowLevel int

const (
	LevelRuns WorkflowLevel = iota
	LevelJobs
)

// WorkflowsModel manages Actions workflow runs and jobs drill-down.
type WorkflowsModel struct {
	client      api.Client
	runsTable   table.Model
	jobsTable   table.Model
	level       WorkflowLevel
	runs        []types.WorkflowRun
	jobs        []types.Job
	selectedRun types.WorkflowRun
	commitSHA   string
	prNumber    int
	loading     bool
	err         error
	width       int
	height      int
}

// NewWorkflowsModel constructs a new WorkflowsModel initialized with Catppuccin tables.
func NewWorkflowsModel(client api.Client) WorkflowsModel {
	runCols := []table.Column{
		{Title: "#", Width: 10},
		{Title: "Workflow", Width: 26},
		{Title: "Event", Width: 12},
		{Title: "Status", Width: 16},
		{Title: "Branch", Width: 14},
		{Title: "Commit", Width: 9},
	}
	rt := table.New(
		table.WithColumns(runCols),
		table.WithFocused(true),
	)
	rt.SetStyles(DefaultTableStyles())

	jobCols := []table.Column{
		{Title: "#", Width: 10},
		{Title: "Job", Width: 32},
		{Title: "Status", Width: 18},
		{Title: "Steps", Width: 12},
	}
	jt := table.New(
		table.WithColumns(jobCols),
		table.WithFocused(true),
	)
	jt.SetStyles(DefaultTableStyles())

	return WorkflowsModel{
		client:    client,
		runsTable: rt,
		jobsTable: jt,
		level:     LevelRuns,
		loading:   true,
	}
}

// Init triggers initial fetch of workflow runs.
func (m WorkflowsModel) Init() tea.Cmd {
	return m.FetchRuns("")
}

// FetchRuns queries workflow runs, optionally filtered by commit SHA.
func (m WorkflowsModel) FetchRuns(commitSHA string) tea.Cmd {
	return func() tea.Msg {
		runs, err := m.client.ListWorkflowRuns(context.Background(), commitSHA, 30)
		return types.WorkflowsLoadedMsg{
			Runs:      runs,
			CommitSHA: commitSHA,
			Err:       err,
		}
	}
}

// FetchJobs queries jobs for a specific workflow run.
func (m WorkflowsModel) FetchJobs(runID int64) tea.Cmd {
	return func() tea.Msg {
		jobs, err := m.client.ListWorkflowJobs(context.Background(), runID)
		return types.JobsLoadedMsg{
			RunID: runID,
			Jobs:  jobs,
			Err:   err,
		}
	}
}

// SetFilter applies a commit SHA and optional PR number filter.
func (m *WorkflowsModel) SetFilter(commitSHA string, prNumber int) {
	m.commitSHA = commitSHA
	m.prNumber = prNumber
	m.level = LevelRuns
	m.loading = true
}

// ClearFilter resets any active commit SHA filter.
func (m *WorkflowsModel) ClearFilter() {
	m.commitSHA = ""
	m.prNumber = 0
}

// Level returns the current drill-down level (runs or jobs).
func (m WorkflowsModel) Level() WorkflowLevel {
	return m.level
}

// SetDimensions calculates responsive column widths for both runs and jobs tables.
func (m *WorkflowsModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height

	availableWidth := width - 2
	if availableWidth < 60 {
		availableWidth = 60
	}

	// 1. Runs Table: 6 columns => 18 char overhead
	runsCols := 6
	runsOverhead := runsCols * 3
	runsUsable := availableWidth - runsOverhead
	if runsUsable < 45 {
		runsUsable = 45
	}

	idWidth := 10
	eventWidth := 12
	statusWidth := 16
	branchWidth := 14
	commitWidth := 9
	fixedRuns := idWidth + eventWidth + statusWidth + branchWidth + commitWidth
	nameWidth := runsUsable - fixedRuns
	if nameWidth < 15 {
		nameWidth = 15
	}

	m.runsTable.SetColumns([]table.Column{
		{Title: "#", Width: idWidth},
		{Title: "Workflow", Width: nameWidth},
		{Title: "Event", Width: eventWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Branch", Width: branchWidth},
		{Title: "Commit", Width: commitWidth},
	})
	m.runsTable.SetWidth(availableWidth)
	m.runsTable.SetHeight(height - 2)

	// 2. Jobs Table: 4 columns => 12 char overhead
	jobsCols := 4
	jobsOverhead := jobsCols * 3
	jobsUsable := availableWidth - jobsOverhead
	if jobsUsable < 40 {
		jobsUsable = 40
	}

	jobIDWidth := 10
	jobStatusWidth := 18
	stepsWidth := 12
	fixedJobs := jobIDWidth + jobStatusWidth + stepsWidth
	jobNameWidth := jobsUsable - fixedJobs
	if jobNameWidth < 15 {
		jobNameWidth = 15
	}

	m.jobsTable.SetColumns([]table.Column{
		{Title: "#", Width: jobIDWidth},
		{Title: "Job", Width: jobNameWidth},
		{Title: "Status", Width: jobStatusWidth},
		{Title: "Steps", Width: stepsWidth},
	})
	m.jobsTable.SetWidth(availableWidth)
	m.jobsTable.SetHeight(height - 2)
}

// Update processes table keystrokes and loaded data events.
func (m WorkflowsModel) Update(msg tea.Msg) (WorkflowsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case types.WorkflowsLoadedMsg:
		m.loading = false
		m.err = msg.Err
		if msg.Err != nil {
			return m, nil
		}

		m.runs = msg.Runs
		m.commitSHA = msg.CommitSHA

		var rows []table.Row
		for _, r := range m.runs {
			shortCommit := r.HeadSha
			if len(shortCommit) > 7 {
				shortCommit = shortCommit[:7]
			}
			rows = append(rows, table.Row{
				"#" + strconv.FormatInt(r.ID, 10),
				r.Name,
				r.Event,
				RenderWorkflowStatus(r.Status, r.Conclusion),
				r.HeadBranch,
				shortCommit,
			})
		}
		m.runsTable.SetRows(rows)
		return m, nil

	case types.JobsLoadedMsg:
		m.loading = false
		m.err = msg.Err
		if msg.Err != nil {
			return m, nil
		}

		m.jobs = msg.Jobs
		m.level = LevelJobs

		var rows []table.Row
		for _, j := range m.jobs {
			stepText := fmt.Sprintf("%d steps", len(j.Steps))
			rows = append(rows, table.Row{
				"#" + strconv.FormatInt(j.ID, 10),
				j.Name,
				RenderWorkflowStatus(j.Status, j.Conclusion),
				stepText,
			})
		}
		m.jobsTable.SetRows(rows)
		return m, nil

	case tea.KeyMsg:
		switch m.level {
		case LevelRuns:
			switch msg.String() {
			case "enter":
				if len(m.runs) > 0 {
					cursor := m.runsTable.Cursor()
					if cursor >= 0 && cursor < len(m.runs) {
						m.selectedRun = m.runs[cursor]
						m.loading = true
						return m, m.FetchJobs(m.selectedRun.ID)
					}
				}
			case "r":
				m.loading = true
				return m, m.FetchRuns(m.commitSHA)
			case "c":
				if m.commitSHA != "" {
					m.ClearFilter()
					m.loading = true
					return m, m.FetchRuns("")
				}
			}

			var cmd tea.Cmd
			m.runsTable, cmd = m.runsTable.Update(msg)
			return m, cmd

		case LevelJobs:
			switch msg.String() {
			case "enter", "l", "L":
				if len(m.jobs) > 0 {
					cursor := m.jobsTable.Cursor()
					if cursor >= 0 && cursor < len(m.jobs) {
						selectedJob := m.jobs[cursor]
						return m, func() tea.Msg {
							return types.OpenJobLogsMsg{
								JobID:   selectedJob.ID,
								JobName: selectedJob.Name,
							}
						}
					}
				}
			case "esc", "backspace":
				m.level = LevelRuns
				return m, nil
			case "r":
				m.loading = true
				return m, m.FetchJobs(m.selectedRun.ID)
			}

			var cmd tea.Cmd
			m.jobsTable, cmd = m.jobsTable.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// View renders the workflow runs or jobs table, along with context banners.
func (m WorkflowsModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  Error loading workflows: %v\n  Press 'r' to retry.", m.err)
	}

	if m.loading && len(m.runs) == 0 && m.level == LevelRuns {
		return "\n  Loading workflow runs..."
	}

	if m.loading && len(m.jobs) == 0 && m.level == LevelJobs {
		return fmt.Sprintf("\n  Loading jobs for workflow #%d...", m.selectedRun.ID)
	}

	bannerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89b4fa")).
		Bold(true).
		Padding(0, 1)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6c7086"))

	if m.level == LevelJobs {
		header := bannerStyle.Render(fmt.Sprintf("Runs > %s (#%d)", m.selectedRun.Name, m.selectedRun.ID)) +
			dimStyle.Render(" — Enter/L: view logs, Esc: back to runs")
		return header + "\n" + m.jobsTable.View()
	}

	var header string
	if m.commitSHA != "" {
		shortSHA := m.commitSHA
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}
		if m.prNumber > 0 {
			header = bannerStyle.Render(fmt.Sprintf("Filtered by PR #%d (commit: %s)", m.prNumber, shortSHA)) +
				dimStyle.Render(" — 'c': clear filter, Enter: view jobs") + "\n"
		} else {
			header = bannerStyle.Render(fmt.Sprintf("Filtered by commit: %s", shortSHA)) +
				dimStyle.Render(" — 'c': clear filter, Enter: view jobs") + "\n"
		}
	}

	if len(m.runs) == 0 {
		if m.commitSHA != "" {
			return header + "\n  No workflow runs found for this commit."
		}
		return "\n  No workflow runs found."
	}

	if header != "" {
		return header + m.runsTable.View()
	}

	return m.runsTable.View()
}
