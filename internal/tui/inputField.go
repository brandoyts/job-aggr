package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputField struct {
	model       textinput.Model
	label       string
	placeholder string
	submitted   bool
	value       string
}

func NewInputField(label, placeholder string) InputField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()

	return InputField{
		model:       ti,
		label:       label,
		placeholder: placeholder,
		submitted:   false,
		value:       "",
	}
}

func (f *InputField) Focus() tea.Cmd {
	f.model.Focus()
	return textinput.Blink
}

func (f *InputField) Blur() {
	f.model.Blur()
}

func (f *InputField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.model, cmd = f.model.Update(msg)
	return cmd
}

func (f *InputField) Submit() {
	f.value = f.model.Value()
	f.submitted = true
	f.model.Blur()
}

func (f InputField) View() string {
	if f.submitted {
		return f.label + " " + f.value
	}
	return f.label + "\n" + f.model.View()
}

func (f InputField) Value() string {
	if f.submitted {
		return f.value
	}
	return f.model.Value()
}

func (f InputField) IsValid() bool {
	return len(f.model.Value()) > 0
}

func (f InputField) IsSubmitted() bool {
	return f.submitted
}
