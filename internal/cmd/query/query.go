package query

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"sigscope/internal/vcd"
)

// stringSlice is a custom flag type for repeated string flags
type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// Run executes the query command
func Run(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	var signals stringSlice
	fs.Var(&signals, "s", "Signal name pattern (can be repeated)")
	fs.Var(&signals, "signals", "Signal name pattern (can be repeated)")

	var timeStart uint64
	fs.Uint64Var(&timeStart, "t", 0, "Start time")
	fs.Uint64Var(&timeStart, "time-start", 0, "Start time")

	var timeEnd uint64
	fs.Uint64Var(&timeEnd, "e", 0, "End time (0 = use VCD end time)")
	fs.Uint64Var(&timeEnd, "time-end", 0, "End time (0 = use VCD end time)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: sigscope query [OPTIONS] <vcd-file>")
	}

	filename := fs.Arg(0)

	// Parse VCD file
	vcdFile, err := vcd.Parse(filename)
	if err != nil {
		return fmt.Errorf("failed to parse VCD file: %w", err)
	}

	// Set default end time
	if timeEnd == 0 {
		timeEnd = vcdFile.EndTime
	}

	// Validate time range
	if timeStart > timeEnd {
		return fmt.Errorf("invalid time range: start (%d) > end (%d)", timeStart, timeEnd)
	}

	// Match signals
	matchedSignals := matchSignals(vcdFile, signals)

	// Detect clock from all signals (not just matched ones)
	allSignals := vcdFile.GetSignalList()
	clock := detectClock(allSignals, timeStart, timeEnd)

	// Build initial values
	init := buildInit(matchedSignals, timeStart)

	// Build events
	events := buildEvents(matchedSignals, timeStart, timeEnd, clock)

	// Build output
	output := QueryOutput{
		Timescale: vcdFile.Timescale,
		Clock:     clock,
		Init:      init,
		Events:    events,
	}

	// Output JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// matchSignals filters signals based on patterns
func matchSignals(vcdFile *vcd.VCDFile, patterns []string) []*vcd.SignalData {
	allSignals := vcdFile.GetSignalList()

	// If no patterns, return all signals
	if len(patterns) == 0 {
		return allSignals
	}

	// Filter by patterns
	var matched []*vcd.SignalData
	for _, sig := range allSignals {
		for _, pattern := range patterns {
			if strings.Contains(sig.Signal.FullName, pattern) {
				matched = append(matched, sig)
				break
			}
		}
	}

	return matched
}

// detectClock attempts to detect a clock signal
func detectClock(signals []*vcd.SignalData, startTime, endTime uint64) *ClockInfo {
	for _, sig := range signals {
		// Only consider 1-bit signals
		if sig.Signal.Width != 1 {
			continue
		}

		// Collect transitions in the time range
		var transitions []uint64
		for _, ch := range sig.Changes {
			if ch.Time >= startTime && ch.Time <= endTime {
				transitions = append(transitions, ch.Time)
			}
		}

		// Need at least 3 transitions to detect a period
		if len(transitions) < 3 {
			continue
		}

		// Check if intervals are periodic
		intervals := make([]uint64, len(transitions)-1)
		for i := 1; i < len(transitions); i++ {
			intervals[i-1] = transitions[i] - transitions[i-1]
		}

		// Find the most common interval (period/2)
		intervalCounts := make(map[uint64]int)
		for _, interval := range intervals {
			intervalCounts[interval]++
		}

		var maxCount int
		var halfPeriod uint64
		for interval, count := range intervalCounts {
			if count > maxCount {
				maxCount = count
				halfPeriod = interval
			}
		}

		// If majority of intervals match, we found a clock
		if maxCount >= len(intervals)*2/3 {
			// Determine edge by checking first transition value
			edge := "posedge"
			if len(sig.Changes) > 0 {
				firstVal := sig.Changes[0].Value
				if firstVal == "1" {
					edge = "negedge"
				}
			}

			return &ClockInfo{
				Name:   shortName(sig.Signal.FullName),
				Period: halfPeriod * 2,
				Edge:   edge,
			}
		}
	}

	return nil
}

// buildInit constructs the initial value map
func buildInit(signals []*vcd.SignalData, startTime uint64) map[string]any {
	init := make(map[string]any)

	for _, sig := range signals {
		name := shortName(sig.Signal.FullName)
		value := sig.GetValueAt(startTime)
		init[name] = formatValue(value, sig.Signal.Width)
	}

	return init
}

// Change represents a signal value change
type Change struct {
	Time   uint64
	Signal string
	Value  any
}

// buildEvents constructs the event list
func buildEvents(signals []*vcd.SignalData, startTime, endTime uint64, clock *ClockInfo) []Event {
	var changes []Change

	for _, sig := range signals {
		name := shortName(sig.Signal.FullName)

		// Skip clock signal
		if clock != nil && name == clock.Name {
			continue
		}

		// Collect changes in time range
		for _, ch := range sig.Changes {
			if ch.Time >= startTime && ch.Time <= endTime {
				changes = append(changes, Change{
					Time:   ch.Time,
					Signal: name,
					Value:  formatValue(ch.Value, sig.Signal.Width),
				})
			}
		}
	}

	// Sort by time
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Time < changes[j].Time
	})

	// Group by time
	var events []Event
	if len(changes) == 0 {
		return events
	}

	currentTime := changes[0].Time
	currentSet := make(map[string]any)

	for _, ch := range changes {
		if ch.Time != currentTime {
			events = append(events, Event{
				Time: currentTime,
				Set:  currentSet,
			})
			currentTime = ch.Time
			currentSet = make(map[string]any)
		}
		currentSet[ch.Signal] = ch.Value
	}

	// Add last event
	if len(currentSet) > 0 {
		events = append(events, Event{
			Time: currentTime,
			Set:  currentSet,
		})
	}

	return events
}

// shortName extracts the last component of a hierarchical name
func shortName(fullName string) string {
	idx := strings.LastIndex(fullName, ".")
	if idx == -1 {
		return fullName
	}
	return fullName[idx+1:]
}

// formatValue formats a value based on width
func formatValue(value string, width int) any {
	if width == 1 {
		return value // "0", "1", "x", "z"
	}

	// Check for x/z
	if strings.ContainsAny(value, "xXzZ") {
		return ValueWithMeta{
			Value: value,
			Width: width,
			Radix: "bin",
		}
	}

	// Convert to hex
	n, err := strconv.ParseUint(value, 2, 64)
	if err != nil {
		// Fallback to binary
		return ValueWithMeta{
			Value: value,
			Width: width,
			Radix: "bin",
		}
	}

	hexStr := fmt.Sprintf("%X", n)
	return ValueWithMeta{
		Value: hexStr,
		Width: width,
		Radix: "hex",
	}
}
