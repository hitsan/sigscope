package render

import (
	"fmt"
	"strconv"
	"strings"

	"sigscope/internal/vcd"
)

// RenderWaveformSingleLine renders a signal's waveform in single-line mode
func RenderWaveformSingleLine(sig *vcd.SignalData, startTime, endTime uint64, width int, classicStyle bool) string {
	if width <= 0 || endTime <= startTime {
		return ""
	}

	timePerChar := float64(endTime-startTime) / float64(width)
	result := make([]string, width)

	if sig.Signal.Width == 1 {
		renderSingleBitOneLine(sig, startTime, timePerChar, result, classicStyle)
	} else {
		renderBusOneLine(sig, startTime, timePerChar, result, width)
	}

	return strings.Join(result, "")
}

// renderSingleBitOneLine renders a single-bit signal in single-line mode
func renderSingleBitOneLine(sig *vcd.SignalData, startTime uint64, timePerChar float64, result []string, classicStyle bool) {
	for i := range result {
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

		if classicStyle {
			// Classic style: ▔▁│
			if hasTransition {
				result[i] = CharEdgeClassic // │
			} else {
				switch startValue {
				case "1":
					result[i] = CharHighClassic // ▔
				case "0":
					result[i] = CharLowClassic // ▁
				case "x", "X":
					result[i] = CharUnknown // ?
				case "z", "Z":
					result[i] = CharHighZ // ~
				default:
					result[i] = CharUnknown
				}
			}
		} else {
			// Modern style: __/‾‾\__
			if hasTransition {
				// Determine transition direction
				if startValue == "0" && transitionTo == "1" {
					result[i] = CharRisingEdge // /
				} else if startValue == "1" && transitionTo == "0" {
					result[i] = CharFallingEdge // \
				} else {
					result[i] = CharUnknown // ?
				}
			} else {
				// Stable state
				switch startValue {
				case "1":
					result[i] = CharHigh // ‾
				case "0":
					result[i] = CharLow // _
				case "x", "X":
					result[i] = CharUnknown // ?
				case "z", "Z":
					result[i] = CharHighZ // ~
				default:
					result[i] = CharUnknown
				}
			}
		}
	}
}

// renderBusOneLine renders a multi-bit bus signal in single-line mode
func renderBusOneLine(sig *vcd.SignalData, startTime uint64, timePerChar float64, result []string, width int) {
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
				result[i] = "="
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
			valueEnd := seg.endIdx

			// Center the hex value
			availableWidth := valueEnd - valueStart
			if availableWidth > 0 {
				displayValue := hexValue
				if len(displayValue) > availableWidth {
					displayValue = displayValue[:availableWidth]
				}

				// Fill with '=' first
				padding := (availableWidth - len(displayValue)) / 2
				for i := valueStart; i < valueEnd; i++ {
					result[i] = "-"
				}
				// Overwrite with hex value
				for idx, ch := range displayValue {
					pos := valueStart + padding + idx
					if pos < valueEnd {
						result[pos] = string(ch)
					}
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
