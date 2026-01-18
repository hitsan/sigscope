package view

import (
	"fmt"
	"strings"

	"wave/internal/model"

	"github.com/charmbracelet/lipgloss"
)

// Render creates the complete view
func Render(m model.Model) string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	var sections []string

	// Title bar
	title := renderTitle(m)
	sections = append(sections, title)

	// Main content (signals + waveforms)
	content := renderMainContent(m)
	sections = append(sections, content)

	// Status bar
	status := renderStatusBar(m)
	sections = append(sections, status)

	return strings.Join(sections, "\n")
}

// renderTitle renders the title bar
func renderTitle(m model.Model) string {
	title := fmt.Sprintf(" wave - %s ", m.Filename)
	width := m.Width
	if len(title) < width {
		title = title + strings.Repeat(" ", width-len(title))
	}
	return TitleStyle.Render(title[:min(len(title), width)])
}

// renderMainContent renders the signal list and waveform area side by side
func renderMainContent(m model.Model) string {
	// Render timeline header
	timelineLabel := strings.Repeat(" ", m.SignalPaneWidth) + "│"
	timeline := RenderTimeline(m)
	timelineRow := timelineLabel + timeline

	// Render separator line
	separator := strings.Repeat("─", m.SignalPaneWidth) + "┼" + strings.Repeat("─", m.WaveformWidth())

	// Render signal names and waveforms
	signalList := RenderSignalList(m)
	waveforms := RenderWaveforms(m)

	// Combine signal names and waveforms line by line
	signalLines := strings.Split(signalList, "\n")
	waveformLines := strings.Split(waveforms, "\n")

	var contentLines []string
	maxLines := max(len(signalLines), len(waveformLines))

	for i := 0; i < maxLines; i++ {
		sigLine := ""
		waveLine := ""

		if i < len(signalLines) {
			sigLine = signalLines[i]
		}
		if i < len(waveformLines) {
			waveLine = waveformLines[i]
		}

		// Pad signal line to fixed width
		sigLine = padRight(sigLine, m.SignalPaneWidth)

		contentLines = append(contentLines, sigLine+SeparatorStyle.Render("│")+waveLine)
	}

	// Combine all parts
	var result []string
	result = append(result, timelineRow)
	result = append(result, SeparatorStyle.Render(separator))
	result = append(result, contentLines...)

	return strings.Join(result, "\n")
}

// renderStatusBar renders the status bar at the bottom
func renderStatusBar(m model.Model) string {
	var status string

	if m.Mode == model.ModeSearch {
		// Search mode
		status = fmt.Sprintf(" Search: %s█", m.SearchQuery)
	} else {
		// Normal mode - show cursor time and value
		timeStr := formatTimeStatus(m.CursorTime)

		// Get current value of selected signal
		valueStr := ""
		if sig := m.SelectedSignalData(); sig != nil {
			value := sig.GetValueAt(m.CursorTime)
			if sig.Signal.Width == 1 {
				valueStr = fmt.Sprintf("Value: %s", value)
			} else {
				valueStr = fmt.Sprintf("Value: 0x%s", binaryToHexStatus(value, sig.Signal.Width))
			}
		}

		zoomStr := fmt.Sprintf("Zoom: %.1fx", m.Zoom)
		helpStr := "j/k:↑↓ h/l:←→ +/-:zoom /:search q:quit"

		status = fmt.Sprintf(" Time: %s | %s | %s | %s", timeStr, valueStr, zoomStr, helpStr)
	}

	// Pad to full width
	width := m.Width
	if len(status) < width {
		status = status + strings.Repeat(" ", width-len(status))
	} else if len(status) > width {
		status = status[:width]
	}

	return StatusStyle.Render(status)
}

// formatTimeStatus formats time for status bar display
func formatTimeStatus(t uint64) string {
	if t >= 1000000000 {
		return fmt.Sprintf("%.2fms", float64(t)/1000000000)
	}
	if t >= 1000000 {
		return fmt.Sprintf("%.2fus", float64(t)/1000000)
	}
	if t >= 1000 {
		return fmt.Sprintf("%.2fns", float64(t)/1000)
	}
	return fmt.Sprintf("%dps", t)
}

// binaryToHexStatus converts binary string to hex for status display
func binaryToHexStatus(binary string, width int) string {
	if strings.Contains(binary, "x") || strings.Contains(binary, "X") {
		return "XX"
	}
	if strings.Contains(binary, "z") || strings.Contains(binary, "Z") {
		return "ZZ"
	}

	// Pad binary to proper width
	for len(binary) < width {
		binary = "0" + binary
	}

	// Pad to multiple of 4
	remainder := len(binary) % 4
	if remainder != 0 {
		binary = strings.Repeat("0", 4-remainder) + binary
	}

	// Convert each nibble
	var hex strings.Builder
	for i := 0; i < len(binary); i += 4 {
		nibble := binary[i : i+4]
		val := 0
		for _, c := range nibble {
			val = val*2
			if c == '1' {
				val++
			}
		}
		hex.WriteString(fmt.Sprintf("%X", val))
	}

	return hex.String()
}

// padRight pads a string to the specified width
func padRight(s string, width int) string {
	// Count actual display width (accounting for ANSI codes)
	displayWidth := lipgloss.Width(s)
	if displayWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-displayWidth)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
