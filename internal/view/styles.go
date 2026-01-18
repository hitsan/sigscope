package view

import "github.com/charmbracelet/lipgloss"

var (
	// Title bar style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	// Signal name styles
	SignalNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	SelectedSignalStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("46"))

	// Marker for selected signal
	SelectedMarker = "▶"
	NormalMarker   = " "

	// Waveform styles
	WaveformStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40"))

	BusValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	// Cursor style
	CursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	// Timeline style
	TimelineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	// Status bar style
	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	// Search input style
	SearchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Background(lipgloss.Color("236"))

	// Separator style
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// Border color
	BorderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
