package view

import (
	"strings"

	"sigscope/internal/model"
	"sigscope/internal/render"
)

// RenderWaveforms renders all visible signal waveforms (right pane)
func RenderWaveforms(m model.Model) string {
	if m.SelectMode {
		return renderSelectModeWaveforms(m)
	}
	return renderNormalModeWaveforms(m)
}

// renderNormalModeWaveforms renders waveforms in normal mode (visible signals only)
func renderNormalModeWaveforms(m model.Model) string {
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

		// Render waveform for this signal (2 lines)
		waveform := render.RenderWaveform(sig, m.TimeStart, m.TimeEnd, width)
		upperRunes := []rune(waveform.Upper)
		lowerRunes := []rune(waveform.Lower)

		// Apply grid lines to both lines
		for _, pos := range gridPositions {
			if pos < len(upperRunes) && upperRunes[pos] == ' ' {
				upperRunes[pos] = '┊'
			}
			if pos < len(lowerRunes) && lowerRunes[pos] == ' ' {
				lowerRunes[pos] = '┊'
			}
		}

		// Apply cursor overlay if visible to both lines
		if m.CursorVisible && cursorVisible && cursorPos >= 0 && cursorPos < len(upperRunes) {
			upperRunes[cursorPos] = '│'
			if cursorPos < len(lowerRunes) {
				lowerRunes[cursorPos] = '│'
			}
		}

		upperLine := string(upperRunes)
		lowerLine := string(lowerRunes)

		// Apply different style for selected signal
		if globalIdx == m.SelectedSignal {
			lines = append(lines, SelectedSignalStyle.Render(upperLine))
			lines = append(lines, SelectedSignalStyle.Render(lowerLine))
		} else {
			lines = append(lines, WaveformStyle.Render(upperLine))
			lines = append(lines, WaveformStyle.Render(lowerLine))
		}
	}

	// Pad with empty lines if needed
	for len(lines) < visibleCount {
		lines = append(lines, strings.Repeat(" ", width))
	}

	return strings.Join(lines, "\n")
}

// renderSelectModeWaveforms renders waveforms in select mode (all signals, hidden ones as empty)
func renderSelectModeWaveforms(m model.Model) string {
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
		var upperRunes, lowerRunes []rune

		if m.SignalVisible[i] {
			sig := m.Signals[i]
			waveform := render.RenderWaveform(sig, m.TimeStart, m.TimeEnd, width)
			upperRunes = []rune(waveform.Upper)
			lowerRunes = []rune(waveform.Lower)
		} else {
			upperRunes = []rune(strings.Repeat(" ", width))
			lowerRunes = []rune(strings.Repeat(" ", width))
		}

		// Apply grid lines to both lines
		for _, pos := range gridPositions {
			if pos < len(upperRunes) && upperRunes[pos] == ' ' {
				upperRunes[pos] = '┊'
			}
			if pos < len(lowerRunes) && lowerRunes[pos] == ' ' {
				lowerRunes[pos] = '┊'
			}
		}

		// Apply cursor overlay if visible to both lines
		if m.CursorVisible && cursorVisible && cursorPos >= 0 && cursorPos < len(upperRunes) {
			upperRunes[cursorPos] = '│'
			if cursorPos < len(lowerRunes) {
				lowerRunes[cursorPos] = '│'
			}
		}

		upperLine := string(upperRunes)
		lowerLine := string(lowerRunes)

		// Apply different style for selected signal
		if i == m.SelectedSignal {
			lines = append(lines, SelectedSignalStyle.Render(upperLine))
			lines = append(lines, SelectedSignalStyle.Render(lowerLine))
		} else {
			lines = append(lines, WaveformStyle.Render(upperLine))
			lines = append(lines, WaveformStyle.Render(lowerLine))
		}
	}

	// Pad with empty lines if needed
	for len(lines) < visibleCount {
		lines = append(lines, strings.Repeat(" ", width))
	}

	return strings.Join(lines, "\n")
}
