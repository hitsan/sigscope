package view

import (
	"fmt"
	"strings"

	"wave/internal/model"
)

// RenderSignalList renders the signal name list (left pane)
func RenderSignalList(m model.Model) string {
	var lines []string
	visibleCount := m.VisibleSignalCount()

	// Determine which signals to show
	startIdx := m.SignalScrollOffset
	endIdx := startIdx + visibleCount
	if endIdx > len(m.Signals) {
		endIdx = len(m.Signals)
	}

	for i := startIdx; i < endIdx; i++ {
		sig := m.Signals[i]
		name := sig.Signal.Name

		// Add bit width indicator for buses
		if sig.Signal.Width > 1 {
			name = fmt.Sprintf("%s[%d:0]", name, sig.Signal.Width-1)
		}

		// Truncate or pad name to fit
		nameWidth := m.SignalPaneWidth - 2 // Reserve space for marker
		if len(name) > nameWidth {
			name = name[:nameWidth-1] + "…"
		} else {
			name = name + strings.Repeat(" ", nameWidth-len(name))
		}

		// Apply style based on selection
		var line string
		if i == m.SelectedSignal {
			line = SelectedSignalStyle.Render(SelectedMarker + name)
		} else {
			line = SignalNameStyle.Render(NormalMarker + name)
		}

		lines = append(lines, line)
	}

	// Pad with empty lines if needed
	for len(lines) < visibleCount {
		lines = append(lines, strings.Repeat(" ", m.SignalPaneWidth))
	}

	return strings.Join(lines, "\n")
}
