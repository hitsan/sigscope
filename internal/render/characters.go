package render

// Unicode characters for waveform rendering

const (
	// Modern style (__/‾‾\__) - Default
	CharHigh        = "‾"  // High level
	CharLow         = "_"  // Low level
	CharRisingEdge  = "/"  // Rising edge (0→1)
	CharFallingEdge = "\\" // Falling edge (1→0)

	// Classic style (▔▁│)
	CharHighClassic   = "▔" // High level (classic)
	CharLowClassic    = "▁" // Low level (classic)
	CharEdgeClassic   = "│" // Edge (classic)

	// Common characters
	CharUnknown = "?" // Unknown value
	CharHighZ   = "~" // High impedance

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
