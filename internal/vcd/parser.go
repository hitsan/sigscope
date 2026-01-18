package vcd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Parse reads and parses a VCD file
func Parse(filename string) (*VCDFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	vcd := NewVCDFile()
	scanner := bufio.NewScanner(file)

	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var currentScope []string
	inHeader := true
	var currentTime uint64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if inHeader {
			// Parse header section
			if strings.HasPrefix(line, "$version") {
				vcd.Version = parseHeaderValue(line, scanner, "$end")
			} else if strings.HasPrefix(line, "$date") {
				vcd.Date = parseHeaderValue(line, scanner, "$end")
			} else if strings.HasPrefix(line, "$timescale") {
				vcd.Timescale = parseHeaderValue(line, scanner, "$end")
			} else if strings.HasPrefix(line, "$scope") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					currentScope = append(currentScope, parts[2])
				}
			} else if strings.HasPrefix(line, "$upscope") {
				if len(currentScope) > 0 {
					currentScope = currentScope[:len(currentScope)-1]
				}
			} else if strings.HasPrefix(line, "$var") {
				sig := parseVar(line, currentScope)
				if sig != nil {
					vcd.Signals[sig.ID] = &SignalData{
						Signal:  *sig,
						Changes: make([]ValueChange, 0),
					}
				}
			} else if strings.HasPrefix(line, "$enddefinitions") {
				inHeader = false
			}
		} else {
			// Parse value changes
			if strings.HasPrefix(line, "#") {
				// Time stamp
				timeStr := strings.TrimPrefix(line, "#")
				t, err := strconv.ParseUint(timeStr, 10, 64)
				if err == nil {
					currentTime = t
					if t > vcd.EndTime {
						vcd.EndTime = t
					}
				}
			} else if strings.HasPrefix(line, "b") || strings.HasPrefix(line, "B") {
				// Binary value for bus signal
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					value := strings.TrimPrefix(strings.TrimPrefix(parts[0], "b"), "B")
					id := parts[1]
					if sig, ok := vcd.Signals[id]; ok {
						sig.Changes = append(sig.Changes, ValueChange{
							Time:  currentTime,
							Value: value,
						})
					}
				}
			} else if len(line) >= 2 {
				// Single-bit value change (e.g., "0!", "1#", "x$")
				value := string(line[0])
				id := line[1:]
				if value == "0" || value == "1" || value == "x" || value == "X" || value == "z" || value == "Z" {
					if sig, ok := vcd.Signals[id]; ok {
						sig.Changes = append(sig.Changes, ValueChange{
							Time:  currentTime,
							Value: strings.ToLower(value),
						})
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return vcd, nil
}

// parseHeaderValue extracts value from header sections that may span multiple lines
func parseHeaderValue(line string, scanner *bufio.Scanner, endMarker string) string {
	// Check if $end is on the same line
	if strings.Contains(line, endMarker) {
		// Extract value between keyword and $end
		parts := strings.SplitN(line, " ", 2)
		if len(parts) > 1 {
			value := strings.TrimSuffix(parts[1], endMarker)
			return strings.TrimSpace(value)
		}
		return ""
	}

	// Multi-line value
	var values []string
	parts := strings.SplitN(line, " ", 2)
	if len(parts) > 1 {
		values = append(values, strings.TrimSpace(parts[1]))
	}

	for scanner.Scan() {
		nextLine := strings.TrimSpace(scanner.Text())
		if strings.Contains(nextLine, endMarker) {
			nextLine = strings.TrimSuffix(nextLine, endMarker)
			if trimmed := strings.TrimSpace(nextLine); trimmed != "" {
				values = append(values, trimmed)
			}
			break
		}
		values = append(values, nextLine)
	}

	return strings.Join(values, " ")
}

// parseVar parses a $var line and returns a Signal
func parseVar(line string, currentScope []string) *Signal {
	// $var wire 1 ! clk $end
	// $var wire 8 " data [7:0] $end
	parts := strings.Fields(line)
	if len(parts) < 5 {
		return nil
	}

	// parts[0] = "$var"
	// parts[1] = type (wire, reg, etc.)
	// parts[2] = width
	// parts[3] = id
	// parts[4:] = name (may include [7:0]) and $end

	width, err := strconv.Atoi(parts[2])
	if err != nil {
		width = 1
	}

	id := parts[3]

	// Find name - everything after id until $end
	nameEndIdx := len(parts)
	for i := 4; i < len(parts); i++ {
		if parts[i] == "$end" {
			nameEndIdx = i
			break
		}
	}

	name := strings.Join(parts[4:nameEndIdx], " ")
	// Remove bit range from name if present (e.g., "[7:0]")
	if idx := strings.Index(name, " ["); idx != -1 {
		name = name[:idx]
	}

	scope := strings.Join(currentScope, ".")
	fullName := name
	if scope != "" {
		fullName = scope + "." + name
	}

	return &Signal{
		ID:       id,
		Name:     name,
		Width:    width,
		Scope:    scope,
		FullName: fullName,
	}
}
