package query

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"sigscope/internal/vcd"
)

// RunList executes the list command
func RunList(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sigscope list <vcd-file>")
	}

	filename := args[0]

	// Parse VCD file
	vcdFile, err := vcd.Parse(filename)
	if err != nil {
		return fmt.Errorf("failed to parse VCD file: %w", err)
	}

	// Build signal list
	signalList := vcdFile.GetSignalList()
	signals := make([]SignalInfo, 0, len(signalList))

	for _, sig := range signalList {
		signals = append(signals, SignalInfo{
			Name:  sig.Signal.FullName,
			Width: sig.Signal.Width,
		})
	}

	// Sort by name for consistent output
	sort.Slice(signals, func(i, j int) bool {
		return signals[i].Name < signals[j].Name
	})

	// Build output
	output := ListOutput{
		Signals:   signals,
		Timescale: vcdFile.Timescale,
		TimeRange: [2]uint64{0, vcdFile.EndTime},
	}

	// Output JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
