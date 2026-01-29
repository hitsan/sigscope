package view

import (
	"fmt"
	"strings"

	"sigscope/internal/model"
)

// RenderTimeline renders the time axis header
func RenderTimeline(m model.Model) string {
	width := m.WaveformWidth()
	if width <= 0 {
		return ""
	}

	// Calculate tick interval (aim for roughly 5-10 ticks)
	timeRange := m.TimeEnd - m.TimeStart
	if timeRange == 0 {
		return strings.Repeat(" ", width)
	}

	// Find a nice tick interval
	tickInterval := findNiceInterval(timeRange, width/15)
	if tickInterval == 0 {
		tickInterval = 1
	}

	// Build timeline string
	result := make([]byte, width)
	for i := range result {
		result[i] = ' '
	}

	// Calculate first tick position
	firstTick := ((m.TimeStart / tickInterval) + 1) * tickInterval
	if m.TimeStart == 0 {
		firstTick = 0
	}

	for t := firstTick; t <= m.TimeEnd; t += tickInterval {
		// Calculate position
		pos := int(float64(t-m.TimeStart) / float64(timeRange) * float64(width))
		if pos >= width {
			break
		}

		// Format time label
		label := formatTime(t)

		// Place label centered on tick position
		labelStart := pos - len(label)/2
		if labelStart < 0 {
			labelStart = 0
		}
		if labelStart+len(label) > width {
			labelStart = width - len(label)
		}

		// Write label if it fits
		if labelStart >= 0 {
			for i, c := range label {
				if labelStart+i < width {
					result[labelStart+i] = byte(c)
				}
			}
		}
	}

	return TimelineStyle.Render(string(result))
}

// findNiceInterval finds a nice tick interval
func findNiceInterval(timeRange uint64, targetTicks int) uint64 {
	if targetTicks <= 0 {
		targetTicks = 5
	}

	rough := timeRange / uint64(targetTicks)
	if rough == 0 {
		return 1
	}

	// Round to nice numbers (1, 2, 5, 10, 20, 50, 100, ...)
	niceIntervals := []uint64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000}

	// Find multiplier
	multiplier := uint64(1)
	for rough > niceIntervals[len(niceIntervals)-1] {
		rough /= 10
		multiplier *= 10
	}

	// Find closest nice interval
	for _, nice := range niceIntervals {
		if nice >= rough {
			return nice * multiplier
		}
	}

	return rough * multiplier
}

// GetGridPositions returns positions for vertical grid lines
func GetGridPositions(m model.Model) []int {
	width := m.WaveformWidth()
	timeRange := m.TimeEnd - m.TimeStart
	if timeRange == 0 || width <= 0 {
		return nil
	}

	tickInterval := findNiceInterval(timeRange, width/15)
	if tickInterval == 0 {
		return nil
	}

	var positions []int
	firstTick := ((m.TimeStart / tickInterval) + 1) * tickInterval
	if m.TimeStart == 0 {
		firstTick = tickInterval
	}

	for t := firstTick; t <= m.TimeEnd; t += tickInterval {
		pos := int(float64(t-m.TimeStart) / float64(timeRange) * float64(width))
		if pos > 0 && pos < width {
			positions = append(positions, pos)
		}
	}
	return positions
}

// formatTime formats a time value with appropriate unit
func formatTime(t uint64) string {
	if t == 0 {
		return "0"
	}
	if t >= 1000000000 {
		return fmt.Sprintf("%dms", t/1000000000)
	}
	if t >= 1000000 {
		return fmt.Sprintf("%dus", t/1000000)
	}
	if t >= 1000 {
		return fmt.Sprintf("%dns", t/1000)
	}
	return fmt.Sprintf("%dps", t)
}
