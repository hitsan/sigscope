package render

// Unicode characters for waveform rendering

const (
	// Single-bit waveform characters (__/‾‾\__ style)
	CharHigh        = "‾"  // High level
	CharLow         = "_"  // Low level
	CharRisingEdge  = "/"  // Rising edge (0→1)
	CharFallingEdge = "\\" // Falling edge (1→0)
	CharUnknown     = "?"  // Unknown value
	CharHighZ       = "~"  // High impedance

	// Bus signal characters
	CharBusRise   = "X"  // Bus transition marker (single cell)
	CharBusFall   = "X"  // Kept for compatibility; not used when rendering single-cell markers
	CharBusHigh   = "▔"  // Bus top line
	CharBusLow    = "▁"  // Bus bottom line
	CharBusMiddle = " "  // Bus middle (for value display)

	// Cursor
	CharCursor = "│"

	// Box drawing
	CharVertical    = "│"
	CharHorizontal  = "─"
	CharTopLeft     = "┌"
	CharTopRight    = "┐"
	CharBottomLeft  = "└"
	CharBottomRight = "┘"
	CharTeeLeft     = "├"
	CharTeeRight    = "┤"
	CharTeeTop      = "┬"
	CharTeeBottom   = "┴"
	CharCross       = "┼"
)
