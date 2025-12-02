package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type SpinnerTickMsg time.Time

type SearchProgress struct {
	spinner   spinner.Model
	message   string
	startTime time.Time
	elapsed   time.Duration
}

func NewSearchProgress(message string) SearchProgress {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return SearchProgress{
		spinner:   s,
		message:   message,
		startTime: time.Now(),
		elapsed:   0,
	}
}

func (sp SearchProgress) Init() tea.Cmd {
	return tea.Batch(
		sp.spinner.Tick,
		sp.tickCmd(),
	)
}

func (sp *SearchProgress) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case SpinnerTickMsg:
		sp.elapsed = time.Since(sp.startTime)
		return sp.tickCmd()

	default:
		var cmd tea.Cmd
		sp.spinner, cmd = sp.spinner.Update(msg)
		return cmd
	}
}

func (sp SearchProgress) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return SpinnerTickMsg(t)
	})
}

func (sp SearchProgress) View() string {
	return fmt.Sprintf("%s %s (%.1fs)",
		sp.spinner.View(),
		sp.message,
		sp.elapsed.Seconds(),
	)
}
