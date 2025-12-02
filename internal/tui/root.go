package tui

import (
	"context"
	"strings"

	"github.com/brandoyts/job-aggr/internal/service/aggregator"
	"github.com/brandoyts/job-aggr/internal/service/scraper/indeed"
	"github.com/brandoyts/job-aggr/internal/service/scraper/linkedin"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Root struct {
	currentStep Step
	title       InputField
	location    InputField
	jobs        JobsList
	progress    SearchProgress
	err         error
}

func NewRoot() *Root {
	title := NewInputField("Job Title:", "e.g. Software Engineer")
	location := NewInputField("Location:", "e.g. San Francisco, CA")

	return &Root{
		currentStep: StepTitle,
		title:       title,
		location:    location,
		jobs:        NewJobsList(),
		progress:    NewSearchProgress("🔎 Searching for jobs..."),
	}
}

func (m *Root) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msgTyped := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msgTyped)

	case JobsMsg:
		m.jobs.SetItems(msgTyped)
		m.currentStep = StepJobs
		return m, nil

	case ErrMsg:
		m.err = msgTyped
		return m, nil

	}

	return m.updateCurrentField(msg)
}

func (m *Root) handleKeyPress(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit

	case tea.KeyEnter:
		return m.handleEnter()
	}

	return m.updateCurrentField(key)
}

func (m *Root) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentStep {
	case StepTitle:
		if m.title.IsValid() {
			m.title.Submit()
			m.currentStep = StepLocation
			return m, m.location.Focus()
		}

	case StepLocation:
		m.location.Submit()
		m.currentStep = StepSearching
		m.progress = NewSearchProgress("🔎 Searching for jobs...")
		return m, tea.Batch(m.progress.Init(), m.performSearch())

	case StepJobs:
		// Do nothing on Enter when viewing jobs
		return m, nil
	}

	return m, nil
}

func (m *Root) updateCurrentField(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentStep {
	case StepTitle:
		cmd = m.title.Update(msg)
	case StepLocation:
		cmd = m.location.Update(msg)
	case StepSearching:
		cmd = m.progress.Update(msg)
	case StepJobs:
		cmd = m.jobs.Update(msg)
	}

	return m, cmd
}

func (m Root) View() string {
	var b strings.Builder

	// Always show all fields (submitted ones show as completed)
	b.WriteString(m.title.View())
	b.WriteString("\n\n")

	if m.currentStep >= StepLocation {
		b.WriteString(m.location.View())
		b.WriteString("\n\n")
	}

	// Show progress during search
	if m.currentStep == StepSearching {
		b.WriteString(m.progress.View())
		b.WriteString("\n\n")
	}

	if m.currentStep >= StepJobs {
		b.WriteString(m.jobs.View())
	}

	// Show instructions based on current step
	if m.currentStep != StepSearching {
		b.WriteString(m.getInstructions())
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString("Error: " + m.err.Error())
	}

	return b.String()
}

func (m Root) getInstructions() string {
	switch m.currentStep {
	case StepTitle:
		return "(enter to continue, esc to quit)"
	case StepLocation:
		return "(enter to search, esc to quit)"
	case StepSearching:
		return ""
	case StepJobs:
		return "(↑/↓ to navigate, esc to quit)"
	}
	return ""
}

func (m Root) performSearch() tea.Cmd {
	return func() tea.Msg {
		in := indeed.NewScraper()
		li := linkedin.NewScraper()

		aggr := aggregator.NewAggregatorService(in, li)

		result, err := aggr.FetchJobs(context.Background(), m.title.Value(), m.location.Value())
		if err != nil {
			return ErrMsg(err)
		}

		var jobs []Job
		for _, job := range result {
			jobs = append(jobs, Job{
				Title:    job.Title,
				Company:  job.Company,
				Location: job.Location,
				Source:   job.Source,
				Link:     job.Url,
			})
		}
		return JobsMsg(jobs)
	}
}

// Getters for accessing state
func (m Root) GetTitle() string {
	return m.title.Value()
}

func (m Root) GetLocation() string {
	return m.location.Value()
}

func (m Root) GetJobs() []Job {
	return m.jobs.GetJobs()
}

func (m Root) GetCurrentStep() Step {
	return m.currentStep
}
