package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderForeground(lipgloss.Color("240"))

type JobsList struct {
	table   table.Model
	jobs    []Job
	visible bool
}

type Job struct {
	Title    string
	Company  string
	Location string
	Link     string
	Source   string
}

type OpenLinkMsg struct {
	URL string
}

func NewJobsList() JobsList {
	columns := []table.Column{
		{Title: "Title", Width: 50},
		{Title: "Company", Width: 25},
		{Title: "Location", Width: 50},
		{Title: "Source", Width: 20},
		{Title: "Link", Width: 50},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)

	return JobsList{
		table:   t,
		jobs:    []Job{},
		visible: false,
	}
}

func (j *JobsList) SetItems(jobs []Job) {
	j.jobs = jobs
	j.visible = true

	rows := make([]table.Row, len(jobs))
	for i, job := range jobs {
		rows[i] = table.Row{job.Title, job.Company, job.Location, job.Source, job.Link}
	}

	j.table.SetRows(rows)
}

func (j *JobsList) Update(msg tea.Msg) tea.Cmd {
	if !j.visible {
		return nil
	}

	var cmd tea.Cmd
	j.table, cmd = j.table.Update(msg)
	return cmd
}

func (j JobsList) View() string {
	if !j.visible || len(j.jobs) == 0 {
		return "\n📄 Job Results:\nNo jobs found.\n"
	}

	header := fmt.Sprintf("\n📄 Job Results: Found %d job(s)\n\n", len(j.jobs))
	return header + baseStyle.Render(j.table.View()) + "\n"
}

func (j JobsList) GetJobs() []Job {
	return j.jobs
}

func (j JobsList) GetSelected() Job {
	selectedIdx := j.table.Cursor()
	if selectedIdx < len(j.jobs) {
		return j.jobs[selectedIdx]
	}
	return Job{}
}

func (j JobsList) IsVisible() bool {
	return j.visible
}
