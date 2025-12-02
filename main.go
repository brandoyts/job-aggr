package main

import (
	"fmt"
	"os"

	"github.com/brandoyts/job-aggr/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(tui.NewRoot())
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
