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
		printUsage()
		os.Exit(1)
	}

	// Check for help flag
	if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		printUsage()
		os.Exit(0)
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

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: sigscope <command> [options] [arguments]

Commands:
  list <vcd-file>              List all signals in VCD file
  query [OPTIONS] <vcd-file>   Query waveform data in differential event format
  <vcd-file>                   Launch TUI viewer (default)

Query Options:
  -s, --signals <pattern>      Signal name pattern (can be repeated)
  -t, --time-start <time>      Start time (default: 0)
  -e, --time-end <time>        End time (default: VCD end time)

Examples:
  sigscope waveform.vcd                           # Launch TUI
  sigscope list waveform.vcd                      # List all signals
  sigscope query waveform.vcd                     # Query all signals
  sigscope query -s clk -s data waveform.vcd      # Query specific signals
  sigscope query -t 1000 -e 5000 waveform.vcd     # Query time range

Use "sigscope <command> --help" for more information about a command.`)
}
