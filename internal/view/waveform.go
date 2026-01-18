package view

import (
	"strings"

	"wave/internal/model"
	"wave/internal/render"
)

// RenderWaveforms renders all visible signal waveforms (right pane)
func RenderWaveforms(m model.Model) string {
	var lines []string
	visibleCount := m.VisibleSignalCount()
	width := m.WaveformWidth()

	// Determine which signals to show
	startIdx := m.SignalScrollOffset
	endIdx := startIdx + visibleCount
	if endIdx > len(m.Signals) {
		endIdx = len(m.Signals)
	}

	// Get cursor position
	cursorPos, cursorVisible := render.RenderCursor(m.CursorTime, m.TimeStart, m.TimeEnd, width)

	for i := startIdx; i < endIdx; i++ {
		sig := m.Signals[i]

		// Render waveform for this signal
		waveform := render.RenderWaveform(sig, m.TimeStart, m.TimeEnd, width)

		// Apply cursor overlay if visible
		if m.CursorVisible && cursorVisible && cursorPos >= 0 && cursorPos < len(waveform) {
			// Convert to runes for proper handling
			runes := []rune(waveform)
			if cursorPos < len(runes) {
				// Replace character at cursor position with cursor marker
				cursorLine := string(runes[:cursorPos]) + CursorStyle.Render(render.CharCursor) + string(runes[cursorPos+1:])
				waveform = cursorLine
			}
		}

		// Apply different style for selected signal
		if i == m.SelectedSignal {
			lines = append(lines, SelectedSignalStyle.Render(waveform))
		} else {
			lines = append(lines, WaveformStyle.Render(waveform))
		}
	}

	// Pad with empty lines if needed
	for len(lines) < visibleCount {
		lines = append(lines, strings.Repeat(" ", width))
	}

	return strings.Join(lines, "\n")
}
