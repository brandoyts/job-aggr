package tui

type Step int

const (
	StepAPIKey Step = iota
	StepTitle
	StepLocation
	StepSearching
	StepJobs
)

type (
	JobsMsg []Job
	ErrMsg  error

	DoneMsg struct {
		APIKey   string
		Title    string
		Location string
		Jobs     []Job
	}
)
