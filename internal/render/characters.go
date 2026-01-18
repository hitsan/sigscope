package render

// Unicode characters for waveform rendering

const (
	// Single-bit waveform characters
	CharHigh      = "─"  // High level
	CharLow       = "_"  // Low level (using underscore for visibility)
	CharRiseEdge  = "┌"  // Rising edge
	CharFallEdge  = "┐"  // Falling edge
	CharRiseLow   = "┘"  // Low after rise
	CharFallHigh  = "└"  // High after fall
	CharUnknown   = "?"  // Unknown value
	CharHighZ     = "~"  // High impedance

	// Bus signal characters (GTKWave style)
	CharBusRise   = "╱"  // Bus transition start
	CharBusFall   = "╲"  // Bus transition end
	CharBusHigh   = "▔"  // Bus top line
	CharBusLow    = "▁"  // Bus bottom line
	CharBusMiddle = " "  // Bus middle (for value display)

	// Cursor
	CharCursor = "│"

	// Box drawing
	CharVertical   = "│"
	CharHorizontal = "─"
	CharTopLeft    = "┌"
	CharTopRight   = "┐"
	CharBottomLeft = "└"
	CharBottomRight = "┘"
	CharTeeLeft    = "├"
	CharTeeRight   = "┤"
	CharTeeTop     = "┬"
	CharTeeBottom  = "┴"
	CharCross      = "┼"
)
