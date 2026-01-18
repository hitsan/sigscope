package main

import (
	"fmt"
	"os"

	"wave/internal/model"
	"wave/internal/update"
	"wave/internal/vcd"
	"wave/internal/view"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: wave <vcd-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Parse VCD file
	vcdFile, err := vcd.Parse(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing VCD file: %v\n", err)
		os.Exit(1)
	}

	// Create model
	m := model.NewModel(vcdFile, filename)

	// Create and run Bubble Tea program
	p := tea.NewProgram(appModel{m}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// appModel wraps model.Model to implement tea.Model interface
type appModel struct {
	model.Model
}

func (a appModel) Init() tea.Cmd {
	return nil
}

func (a appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := update.Update(a.Model, msg)
	return appModel{m}, cmd
}

func (a appModel) View() string {
	return view.Render(a.Model)
}
