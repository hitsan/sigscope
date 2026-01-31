package query

// QueryOutput represents the JSON output for query command
type QueryOutput struct {
	Timescale string            `json:"timescale"`
	Clock     *ClockInfo        `json:"clock,omitempty"`
	Init      map[string]any    `json:"init"`
	Events    []Event           `json:"events"`
}

// ClockInfo contains detected clock information
type ClockInfo struct {
	Name   string `json:"name"`
	Period uint64 `json:"period"`
	Edge   string `json:"edge"` // "posedge" or "negedge"
}

// Event represents a timestamped set of signal changes
type Event struct {
	Time uint64         `json:"t"`
	Set  map[string]any `json:"set"`
}

// ValueWithMeta represents a multi-bit value with metadata
type ValueWithMeta struct {
	Value string `json:"v"`
	Width int    `json:"w"`
	Radix string `json:"radix"` // "hex" or "bin"
}

// ListOutput represents the JSON output for list command
type ListOutput struct {
	Signals   []SignalInfo `json:"signals"`
	Timescale string       `json:"timescale"`
	TimeRange [2]uint64    `json:"time_range"`
}

// SignalInfo contains signal metadata
type SignalInfo struct {
	Name  string `json:"name"`
	Width int    `json:"width"`
}
