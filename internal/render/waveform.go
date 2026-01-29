package render

import (
	"fmt"
	"strconv"
	"strings"

	"sigscope/internal/vcd"
)

// WaveformLines represents a 2-line waveform display
type WaveformLines struct {
	Upper string // Upper line
	Lower string // Lower line
}

// RenderWaveform renders a signal's waveform for the given time window
// Returns two lines: upper and lower for 2-line display
func RenderWaveform(sig *vcd.SignalData, startTime, endTime uint64, width int) WaveformLines {
	if width <= 0 || endTime <= startTime {
		return WaveformLines{"", ""}
	}

	timePerChar := float64(endTime-startTime) / float64(width)
	upperResult := make([]string, width)
	lowerResult := make([]string, width)

	if sig.Signal.Width == 1 {
		renderSingleBit(sig, startTime, timePerChar, upperResult, lowerResult)
	} else {
		renderBus(sig, startTime, timePerChar, upperResult, lowerResult, width)
	}

	return WaveformLines{
		Upper: strings.Join(upperResult, ""),
		Lower: strings.Join(lowerResult, ""),
	}
}

// renderSingleBit renders a single-bit signal in 2-line format
func renderSingleBit(sig *vcd.SignalData, startTime uint64, timePerChar float64, upper []string, lower []string) {
	for i := range upper {
		charStartTime := startTime + uint64(float64(i)*timePerChar)
		charEndTime := startTime + uint64(float64(i+1)*timePerChar)

		// Use the value at the start of this cell
		startValue := sig.GetValueAt(charStartTime)

		// Check for transitions within this character
		var transitionTo string
		hasTransition := false
		for _, change := range sig.Changes {
			if change.Time >= charStartTime && change.Time < charEndTime {
				hasTransition = true
				transitionTo = change.Value
				break
			}
		}

		if hasTransition {
			// Determine transition direction
			switch {
			case startValue == "1" && transitionTo == "0":
				// High to Low
				upper[i] = "┐"
				lower[i] = "└"
			case startValue == "0" && transitionTo == "1":
				// Low to High
				upper[i] = "┌"
				lower[i] = "┘"
			default:
				// Unknown transition
				upper[i] = "│"
				lower[i] = "│"
			}
		} else {
			// Stable state
			switch startValue {
			case "1":
				upper[i] = "─"
				lower[i] = " "
			case "0":
				upper[i] = " "
				lower[i] = "─"
			case "x", "X":
				upper[i] = "?"
				lower[i] = "?"
			case "z", "Z":
				upper[i] = "~"
				lower[i] = "~"
			default:
				upper[i] = "?"
				lower[i] = "?"
			}
		}
	}
}

// renderBus renders a multi-bit bus signal (GTKWave style) in 2-line format
func renderBus(sig *vcd.SignalData, startTime uint64, timePerChar float64, upper []string, lower []string, width int) {
	// Find all value changes in the visible window
	type segment struct {
		startIdx int
		endIdx   int
		value    string
	}

	segments := make([]segment, 0)
	currentValue := sig.GetValueAt(startTime)
	currentStartIdx := 0

	for i := 0; i < width; i++ {
		charTime := startTime + uint64(float64(i)*timePerChar)
		charEndTime := startTime + uint64(float64(i+1)*timePerChar)

		// Check for value change in this character
		for _, change := range sig.Changes {
			// Treat changes at the cell start as belonging to this cell,
			// and exclude the end to avoid right-shifted edges.
			if change.Time >= charTime && change.Time < charEndTime {
				// End current segment
				if i > currentStartIdx {
					segments = append(segments, segment{
						startIdx: currentStartIdx,
						endIdx:   i,
						value:    currentValue,
					})
				}
				currentValue = change.Value
				currentStartIdx = i
				break
			}
		}
	}

	// Add final segment
	if currentStartIdx < width {
		segments = append(segments, segment{
			startIdx: currentStartIdx,
			endIdx:   width,
			value:    currentValue,
		})
	}

	// Initialize with spaces
	for i := range upper {
		upper[i] = " "
		lower[i] = " "
	}

	// Render each segment with upper/lower half blocks
	for _, seg := range segments {
		segWidth := seg.endIdx - seg.startIdx

		// Convert binary value to hex
		hexValue := binaryToHex(seg.value, sig.Signal.Width)

		// Fill with lower half block (▄) in upper line and upper half block (▀) in lower line
		// This creates a filled band in the middle
		for i := seg.startIdx; i < seg.endIdx; i++ {
			upper[i] = "▄"
			lower[i] = "▀"
		}

		// Show transition as empty space at segment boundaries
		if seg.startIdx > 0 {
			upper[seg.startIdx] = " "
			lower[seg.startIdx] = " "
		}

		// Display value in the center if there's enough space
		if segWidth > len(hexValue)+2 {
			valueStart := seg.startIdx
			if seg.startIdx > 0 {
				valueStart = seg.startIdx + 1
			}
			valueEnd := seg.endIdx

			availableWidth := valueEnd - valueStart
			displayValue := hexValue
			if len(displayValue) > availableWidth {
				displayValue = displayValue[:availableWidth]
			}

			// Center the value across both lines for better visibility
			padding := (availableWidth - len(displayValue)) / 2
			for idx, ch := range displayValue {
				pos := valueStart + padding + idx
				if pos < seg.endIdx {
					// Place value characters on upper line
					upper[pos] = string(ch)
				}
			}
		}
	}
}

// binaryToHex converts a binary string to hexadecimal
func binaryToHex(binary string, width int) string {
	// Handle x and z values
	if strings.Contains(binary, "x") || strings.Contains(binary, "X") {
		return "XX"
	}
	if strings.Contains(binary, "z") || strings.Contains(binary, "Z") {
		return "ZZ"
	}

	// Pad binary to multiple of 4
	padded := binary
	remainder := len(binary) % 4
	if remainder != 0 {
		padded = strings.Repeat("0", 4-remainder) + binary
	}

	// Convert to hex
	val, err := strconv.ParseUint(padded, 2, 64)
	if err != nil {
		return "??"
	}

	// Format with appropriate width
	hexWidth := (width + 3) / 4
	format := fmt.Sprintf("%%0%dX", hexWidth)
	return fmt.Sprintf(format, val)
}

// RenderCursor returns a cursor marker at the given position
func RenderCursor(cursorTime, startTime, endTime uint64, width int) (int, bool) {
	if cursorTime < startTime || cursorTime > endTime {
		return -1, false
	}

	timeRange := float64(endTime - startTime)
	if timeRange == 0 {
		return 0, true
	}

	pos := int(float64(cursorTime-startTime) / timeRange * float64(width))
	if pos >= width {
		pos = width - 1
	}
	return pos, true
}
