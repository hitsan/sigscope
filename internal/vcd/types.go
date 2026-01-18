package vcd

// Signal represents a VCD signal definition
type Signal struct {
	ID       string // VCD identifier (e.g., "!", "#", etc.)
	Name     string // Signal name
	Width    int    // Bit width (1 for single bit, >1 for bus)
	Scope    string // Hierarchical scope (e.g., "top.module")
	FullName string // Scope + Name
}

// ValueChange represents a value change event
type ValueChange struct {
	Time  uint64 // Time in timescale units
	Value string // Value (binary string for buses, "0"/"1"/"x"/"z" for single bit)
}

// SignalData contains a signal definition and its value changes
type SignalData struct {
	Signal  Signal
	Changes []ValueChange
}

// VCDFile represents a parsed VCD file
type VCDFile struct {
	Version   string
	Date      string
	Timescale string
	Signals   map[string]*SignalData // Key: signal ID
	EndTime   uint64                 // Maximum time in the file
}

// NewVCDFile creates a new VCDFile instance
func NewVCDFile() *VCDFile {
	return &VCDFile{
		Signals: make(map[string]*SignalData),
	}
}

// GetSignalList returns all signals as a slice, sorted by full name
func (v *VCDFile) GetSignalList() []*SignalData {
	result := make([]*SignalData, 0, len(v.Signals))
	for _, sig := range v.Signals {
		result = append(result, sig)
	}
	return result
}

// GetValueAt returns the value of a signal at a given time
func (sd *SignalData) GetValueAt(time uint64) string {
	if len(sd.Changes) == 0 {
		return "x"
	}

	// Find the last change at or before the given time
	var lastValue string = "x"
	for _, change := range sd.Changes {
		if change.Time > time {
			break
		}
		lastValue = change.Value
	}
	return lastValue
}
