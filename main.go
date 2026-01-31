package main

import (
	"fmt"
	"os"

	"sigscope/internal/cmd/query"
	"sigscope/internal/model"
	"sigscope/internal/update"
	"sigscope/internal/vcd"
	"sigscope/internal/view"
	"sigscope/internal/watcher"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: sigscope <vcd-file>")
		os.Exit(1)
	}

	// Check for subcommands
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "list":
			if err := query.RunList(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "query":
			if err := query.Run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Default: launch TUI
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
	return watcher.WatchFile(a.Model.Filename)
}

func (a appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := update.Update(a.Model, msg)
	return appModel{m}, cmd
}

func (a appModel) View() string {
	return view.Render(a.Model)
}
