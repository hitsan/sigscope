package view

import (
	"strings"

	"sigscope/internal/model"
	"sigscope/internal/render"
)

// RenderWaveforms renders all visible signal waveforms (right pane)
func RenderWaveforms(m model.Model) string {
	if m.SelectMode {
		return renderSelectModeWaveformsSingleLine(m)
	}
	return renderNormalModeWaveformsSingleLine(m)
}

// renderNormalModeWaveformsSingleLine renders waveforms in normal mode (1-line per signal)
func renderNormalModeWaveformsSingleLine(m model.Model) string {
	var lines []string
	visibleCount := m.VisibleSignalCount()
	width := m.WaveformWidth()
	indices := m.VisibleSignalIndices()

	// Determine which signals to show
	startIdx := m.SignalScrollOffset
	endIdx := startIdx + visibleCount
	if endIdx > len(indices) {
		endIdx = len(indices)
	}

	// Get cursor and grid positions
	cursorPos, cursorVisible := render.RenderCursor(m.CursorTime, m.TimeStart, m.TimeEnd, width)
	gridPositions := GetGridPositions(m)

	for vi := startIdx; vi < endIdx; vi++ {
		globalIdx := indices[vi]
		sig := m.Signals[globalIdx]

		// Render waveform for this signal (single line)
		waveform := render.RenderWaveformSingleLine(sig, m.TimeStart, m.TimeEnd, width, m.ClassicStyle)
		runes := []rune(waveform)

		// Apply grid lines
		for _, pos := range gridPositions {
			if pos < len(runes) && runes[pos] == ' ' {
				runes[pos] = '┊'
			}
		}

		// Apply cursor overlay if visible
		if m.CursorVisible && cursorVisible && cursorPos >= 0 && cursorPos < len(runes) {
			runes[cursorPos] = '│'
		}

		waveform = string(runes)

		// Apply different style for selected signal
		if globalIdx == m.SelectedSignal {
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


// renderSelectModeWaveformsSingleLine renders waveforms in select mode (1-line per signal)
func renderSelectModeWaveformsSingleLine(m model.Model) string {
	var lines []string
	visibleCount := m.VisibleSignalCount()
	width := m.WaveformWidth()

	// Determine which signals to show
	startIdx := m.SignalScrollOffset
	endIdx := startIdx + visibleCount
	if endIdx > len(m.Signals) {
		endIdx = len(m.Signals)
	}

	// Get cursor and grid positions
	cursorPos, cursorVisible := render.RenderCursor(m.CursorTime, m.TimeStart, m.TimeEnd, width)
	gridPositions := GetGridPositions(m)

	for i := startIdx; i < endIdx; i++ {
		var runes []rune

		if m.SignalVisible[i] {
			sig := m.Signals[i]
			waveform := render.RenderWaveformSingleLine(sig, m.TimeStart, m.TimeEnd, width, m.ClassicStyle)
			runes = []rune(waveform)
		} else {
			runes = []rune(strings.Repeat(" ", width))
		}

		// Apply grid lines
		for _, pos := range gridPositions {
			if pos < len(runes) && runes[pos] == ' ' {
				runes[pos] = '┊'
			}
		}

		// Apply cursor overlay if visible
		if m.CursorVisible && cursorVisible && cursorPos >= 0 && cursorPos < len(runes) {
			runes[cursorPos] = '│'
		}

		waveform := string(runes)

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

