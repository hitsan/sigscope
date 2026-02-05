package view

import (
	"fmt"
	"strings"

	"sigscope/internal/model"
)

// RenderSignalList renders the signal name list (left pane)
func RenderSignalList(m model.Model) string {
	if m.SelectMode {
		return renderSelectModeListSingleLine(m)
	}
	return renderNormalModeListSingleLine(m)
}

// renderNormalModeListSingleLine renders signal list in normal mode (1-line per signal)
func renderNormalModeListSingleLine(m model.Model) string {
	var lines []string
	visibleCount := m.VisibleSignalCount()
	indices := m.VisibleSignalIndices()

	// Determine which signals to show
	startIdx := m.SignalScrollOffset
	endIdx := startIdx + visibleCount
	if endIdx > len(indices) {
		endIdx = len(indices)
	}

	for vi := startIdx; vi < endIdx; vi++ {
		globalIdx := indices[vi]
		sig := m.Signals[globalIdx]
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
		if globalIdx == m.SelectedSignal {
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

// renderSelectModeListSingleLine renders signal list in select mode (1-line per signal)
func renderSelectModeListSingleLine(m model.Model) string {
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

		// Checkbox marker
		checkbox := UncheckedMarker
		if m.SignalVisible[i] {
			checkbox = CheckedMarker
		}

		// Truncate or pad name to fit (reserve space for marker + checkbox + space)
		nameWidth := m.SignalPaneWidth - 4
		if len(name) > nameWidth {
			name = name[:nameWidth-1] + "…"
		} else {
			name = name + strings.Repeat(" ", nameWidth-len(name))
		}

		// Apply style based on selection
		var line string
		if i == m.SelectedSignal {
			line = SelectedSignalStyle.Render(SelectedMarker + checkbox + " " + name)
		} else {
			line = SignalNameStyle.Render(NormalMarker + checkbox + " " + name)
		}

		lines = append(lines, line)
	}

	// Pad with empty lines if needed
	for len(lines) < visibleCount {
		lines = append(lines, strings.Repeat(" ", m.SignalPaneWidth))
	}

	return strings.Join(lines, "\n")
}

