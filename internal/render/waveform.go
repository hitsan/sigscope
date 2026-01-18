package render

import (
	"fmt"
	"strconv"
	"strings"

	"wave/internal/vcd"
)

// RenderWaveform renders a signal's waveform for the given time window
func RenderWaveform(sig *vcd.SignalData, startTime, endTime uint64, width int) string {
	if width <= 0 || endTime <= startTime {
		return ""
	}

	timePerChar := float64(endTime-startTime) / float64(width)
	result := make([]string, width)

	if sig.Signal.Width == 1 {
		renderSingleBit(sig, startTime, timePerChar, result)
	} else {
		renderBus(sig, startTime, timePerChar, result, width)
	}

	return strings.Join(result, "")
}

// renderSingleBit renders a single-bit signal
func renderSingleBit(sig *vcd.SignalData, startTime uint64, timePerChar float64, result []string) {
	for i := range result {
		charStartTime := startTime + uint64(float64(i)*timePerChar)
		charEndTime := startTime + uint64(float64(i+1)*timePerChar)

		// Get value at end of this character position
		endValue := sig.GetValueAt(charEndTime)

		// Check for transitions within this character
		hasTransition := false
		transitionUp := false
		for _, change := range sig.Changes {
			if change.Time > charStartTime && change.Time <= charEndTime {
				hasTransition = true
				transitionUp = change.Value == "1"
				break
			}
		}

		if hasTransition {
			if transitionUp {
				result[i] = CharRiseEdge
			} else {
				result[i] = CharFallEdge
			}
		} else {
			switch endValue {
			case "1":
				result[i] = CharHigh
			case "0":
				result[i] = CharLow
			case "x", "X":
				result[i] = CharUnknown
			case "z", "Z":
				result[i] = CharHighZ
			default:
				result[i] = CharUnknown
			}
		}
	}
}

// renderBus renders a multi-bit bus signal (GTKWave style)
func renderBus(sig *vcd.SignalData, startTime uint64, timePerChar float64, result []string, width int) {
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
			if change.Time > charTime && change.Time <= charEndTime {
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
	for i := range result {
		result[i] = " "
	}

	// Render each segment
	for _, seg := range segments {
		segWidth := seg.endIdx - seg.startIdx

		// Convert binary value to hex
		hexValue := binaryToHex(seg.value, sig.Signal.Width)

		if segWidth <= 2 {
			// Too narrow for value, just show transitions
			if seg.startIdx > 0 {
				result[seg.startIdx] = CharBusRise
			}
			for i := seg.startIdx + 1; i < seg.endIdx; i++ {
				result[i] = CharHigh
			}
		} else {
			// Show transition at start
			if seg.startIdx > 0 {
				result[seg.startIdx] = CharBusRise
			}

			// Calculate space for value
			valueStart := seg.startIdx
			if seg.startIdx > 0 {
				valueStart = seg.startIdx + 1
			}
			valueEnd := seg.endIdx - 1

			// Center the hex value
			availableWidth := valueEnd - valueStart
			if availableWidth > 0 {
				displayValue := hexValue
				if len(displayValue) > availableWidth {
					displayValue = displayValue[:availableWidth]
				}

				padding := (availableWidth - len(displayValue)) / 2
				for i := valueStart; i < valueEnd; i++ {
					relPos := i - valueStart
					if relPos >= padding && relPos < padding+len(displayValue) {
						result[i] = string(displayValue[relPos-padding])
					} else {
						result[i] = " "
					}
				}
			}

			// Show transition at end (if not last segment)
			if seg.endIdx < width {
				result[seg.endIdx-1] = CharBusFall
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
