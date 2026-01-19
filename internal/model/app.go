package model

import (
	"sort"
	"strings"

	"wave/internal/vcd"

	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the current application mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
)

// Model is the main application state
type Model struct {
	// VCD data
	VCD      *vcd.VCDFile
	Signals  []*vcd.SignalData // Sorted signal list
	Filename string

	// Viewport state
	TimeStart   uint64  // Start time of visible window
	TimeEnd     uint64  // End time of visible window
	Zoom        float64 // Zoom level (1.0 = default)
	TimePerChar uint64  // Time units per character

	// Cursor state
	CursorTime    uint64 // Current cursor position in time
	CursorVisible bool

	// Selection state
	SelectedSignal int // Index of selected signal

	// Signal visibility
	SignalVisible []bool // 各信号の表示/非表示（Signalsと同じ長さ）
	SelectMode    bool   // true: 全信号選択モード

	// Display state
	Width           int // Terminal width
	Height          int // Terminal height
	SignalPaneWidth int // Width of signal name pane

	// Mode
	Mode         Mode
	SearchQuery  string
	SearchResult []int // Indices of matching signals

	// Scroll state for signal list
	SignalScrollOffset int
}

// NewModel creates a new Model with VCD data
func NewModel(vcdFile *vcd.VCDFile, filename string) Model {
	signals := vcdFile.GetSignalList()

	// Sort signals by full name
	sort.Slice(signals, func(i, j int) bool {
		return signals[i].Signal.FullName < signals[j].Signal.FullName
	})

	// Initialize signal visibility (all visible by default)
	signalVisible := make([]bool, len(signals))
	for i := range signalVisible {
		signalVisible[i] = true
	}

	// Calculate initial time per char (show entire waveform by default)
	timePerChar := uint64(1)
	if vcdFile.EndTime > 0 {
		timePerChar = vcdFile.EndTime / 80 // Assume ~80 chars for waveform
		if timePerChar < 1 {
			timePerChar = 1
		}
	}

	return Model{
		VCD:             vcdFile,
		Signals:         signals,
		Filename:        filename,
		TimeStart:       0,
		TimeEnd:         vcdFile.EndTime,
		Zoom:            1.0,
		TimePerChar:     timePerChar,
		CursorTime:      0,
		CursorVisible:   true,
		SelectedSignal:  0,
		SignalVisible:   signalVisible,
		SelectMode:      false,
		Width:           80,
		Height:          24,
		SignalPaneWidth: 22,
		Mode:            ModeNormal,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// VisibleSignalCount returns the number of signals that can be displayed
func (m Model) VisibleSignalCount() int {
	// Reserve lines for: title, timeline, separator, status bar
	available := m.Height - 4
	if available < 1 {
		return 1
	}
	return available
}

// WaveformWidth returns the width available for waveform display
func (m Model) WaveformWidth() int {
	// Total width minus signal pane and separator
	w := m.Width - m.SignalPaneWidth - 3
	if w < 10 {
		return 10
	}
	return w
}

// ScrollTimeLeft scrolls the time window left
func (m *Model) ScrollTimeLeft(amount uint64) {
	if m.TimeStart >= amount {
		m.TimeStart -= amount
		m.TimeEnd -= amount
	} else {
		diff := m.TimeEnd - m.TimeStart
		m.TimeStart = 0
		m.TimeEnd = diff
	}
}

// ScrollTimeRight scrolls the time window right
func (m *Model) ScrollTimeRight(amount uint64) {
	if m.TimeEnd+amount <= m.VCD.EndTime {
		m.TimeStart += amount
		m.TimeEnd += amount
	} else {
		diff := m.TimeEnd - m.TimeStart
		m.TimeEnd = m.VCD.EndTime
		if m.VCD.EndTime > diff {
			m.TimeStart = m.VCD.EndTime - diff
		} else {
			m.TimeStart = 0
		}
	}
}

// ZoomIn increases zoom level
func (m *Model) ZoomIn() {
	m.Zoom *= 2
	m.recalculateTimeWindow()
}

// ZoomOut decreases zoom level
func (m *Model) ZoomOut() {
	m.Zoom /= 2
	if m.Zoom < 0.125 {
		m.Zoom = 0.125
	}
	m.recalculateTimeWindow()
}

// ResetZoom shows the entire waveform
func (m *Model) ResetZoom() {
	m.Zoom = 1.0
	m.TimeStart = 0
	m.TimeEnd = m.VCD.EndTime
	m.recalculateTimeWindow()
}

// recalculateTimeWindow recalculates time window based on zoom and cursor
func (m *Model) recalculateTimeWindow() {
	waveWidth := uint64(m.WaveformWidth())
	visibleDuration := waveWidth * m.TimePerChar / uint64(m.Zoom)

	// Center on cursor if visible, otherwise on current window center
	center := m.CursorTime
	if center < m.TimeStart || center > m.TimeEnd {
		center = (m.TimeStart + m.TimeEnd) / 2
	}

	halfDuration := visibleDuration / 2
	if center > halfDuration {
		m.TimeStart = center - halfDuration
	} else {
		m.TimeStart = 0
	}

	m.TimeEnd = m.TimeStart + visibleDuration
	if m.TimeEnd > m.VCD.EndTime {
		m.TimeEnd = m.VCD.EndTime
		if m.VCD.EndTime > visibleDuration {
			m.TimeStart = m.VCD.EndTime - visibleDuration
		} else {
			m.TimeStart = 0
		}
	}
}

// MoveSignalUp moves selection up
func (m *Model) MoveSignalUp() {
	if m.SelectMode {
		// 選択モード: 全信号内で移動
		if m.SelectedSignal > 0 {
			m.SelectedSignal--
			m.adjustSignalScroll()
		}
	} else {
		// 通常モード: 表示信号内で移動
		indices := m.VisibleSignalIndices()
		currentVisibleIdx := m.GlobalIndexToVisible(m.SelectedSignal)
		if currentVisibleIdx > 0 {
			m.SelectedSignal = indices[currentVisibleIdx-1]
			m.adjustSignalScroll()
		}
	}
}

// MoveSignalDown moves selection down
func (m *Model) MoveSignalDown() {
	if m.SelectMode {
		// 選択モード: 全信号内で移動
		if m.SelectedSignal < len(m.Signals)-1 {
			m.SelectedSignal++
			m.adjustSignalScroll()
		}
	} else {
		// 通常モード: 表示信号内で移動
		indices := m.VisibleSignalIndices()
		currentVisibleIdx := m.GlobalIndexToVisible(m.SelectedSignal)
		if currentVisibleIdx >= 0 && currentVisibleIdx < len(indices)-1 {
			m.SelectedSignal = indices[currentVisibleIdx+1]
			m.adjustSignalScroll()
		}
	}
}

// adjustSignalScroll adjusts scroll to keep selected signal visible
func (m *Model) adjustSignalScroll() {
	visibleCount := m.VisibleSignalCount()

	if m.SelectMode {
		// 選択モード: 全信号を対象にスクロール
		if m.SelectedSignal < m.SignalScrollOffset {
			m.SignalScrollOffset = m.SelectedSignal
		} else if m.SelectedSignal >= m.SignalScrollOffset+visibleCount {
			m.SignalScrollOffset = m.SelectedSignal - visibleCount + 1
		}
	} else {
		// 通常モード: 表示信号リスト内での位置を計算
		visibleIdx := m.GlobalIndexToVisible(m.SelectedSignal)
		if visibleIdx < 0 {
			return
		}
		if visibleIdx < m.SignalScrollOffset {
			m.SignalScrollOffset = visibleIdx
		} else if visibleIdx >= m.SignalScrollOffset+visibleCount {
			m.SignalScrollOffset = visibleIdx - visibleCount + 1
		}
	}
}

// GoToStart moves to time 0
func (m *Model) GoToStart() {
	m.CursorTime = 0
	duration := m.TimeEnd - m.TimeStart
	m.TimeStart = 0
	m.TimeEnd = duration
	if m.TimeEnd > m.VCD.EndTime {
		m.TimeEnd = m.VCD.EndTime
	}
}

// GoToEnd moves to end time
func (m *Model) GoToEnd() {
	m.CursorTime = m.VCD.EndTime
	duration := m.TimeEnd - m.TimeStart
	m.TimeEnd = m.VCD.EndTime
	if m.VCD.EndTime > duration {
		m.TimeStart = m.VCD.EndTime - duration
	} else {
		m.TimeStart = 0
	}
}

// NextChange moves cursor to next value change of selected signal
func (m *Model) NextChange() {
	if m.SelectedSignal >= len(m.Signals) {
		return
	}
	sig := m.Signals[m.SelectedSignal]
	for _, change := range sig.Changes {
		if change.Time > m.CursorTime {
			m.CursorTime = change.Time
			m.ensureCursorVisible()
			return
		}
	}
}

// PrevChange moves cursor to previous value change of selected signal
func (m *Model) PrevChange() {
	if m.SelectedSignal >= len(m.Signals) {
		return
	}
	sig := m.Signals[m.SelectedSignal]
	var prevTime uint64 = 0
	for _, change := range sig.Changes {
		if change.Time >= m.CursorTime {
			break
		}
		prevTime = change.Time
	}
	m.CursorTime = prevTime
	m.ensureCursorVisible()
}

// ensureCursorVisible scrolls time window to make cursor visible
func (m *Model) ensureCursorVisible() {
	if m.CursorTime < m.TimeStart {
		duration := m.TimeEnd - m.TimeStart
		m.TimeStart = m.CursorTime
		m.TimeEnd = m.TimeStart + duration
	} else if m.CursorTime > m.TimeEnd {
		duration := m.TimeEnd - m.TimeStart
		m.TimeEnd = m.CursorTime
		m.TimeStart = m.TimeEnd - duration
		if m.TimeStart > m.TimeEnd {
			m.TimeStart = 0
		}
	}
}

// SelectedSignalData returns the currently selected signal data
func (m Model) SelectedSignalData() *vcd.SignalData {
	if m.SelectedSignal >= 0 && m.SelectedSignal < len(m.Signals) {
		return m.Signals[m.SelectedSignal]
	}
	return nil
}

// Search performs a signal name search
func (m *Model) Search(query string) {
	m.SearchQuery = query
	m.SearchResult = nil

	if query == "" {
		return
	}

	query = strings.ToLower(query)
	for i, sig := range m.Signals {
		if strings.Contains(strings.ToLower(sig.Signal.FullName), query) {
			m.SearchResult = append(m.SearchResult, i)
		}
	}

	// Jump to first result
	if len(m.SearchResult) > 0 {
		m.SelectedSignal = m.SearchResult[0]
		m.adjustSignalScroll()
	}
}

// VisibleSignalIndices returns indices of visible signals
func (m *Model) VisibleSignalIndices() []int {
	var indices []int
	for i, visible := range m.SignalVisible {
		if visible {
			indices = append(indices, i)
		}
	}
	return indices
}

// ToggleSignalVisibility toggles visibility of the selected signal
func (m *Model) ToggleSignalVisibility() {
	if m.SelectedSignal >= 0 && m.SelectedSignal < len(m.SignalVisible) {
		m.SignalVisible[m.SelectedSignal] = !m.SignalVisible[m.SelectedSignal]
	}
}

// SetAllSignalsVisible sets visibility for all signals
func (m *Model) SetAllSignalsVisible(visible bool) {
	for i := range m.SignalVisible {
		m.SignalVisible[i] = visible
	}
}

// VisibleIndexToGlobal converts a visible signal index to global index
func (m *Model) VisibleIndexToGlobal(visibleIdx int) int {
	indices := m.VisibleSignalIndices()
	if visibleIdx >= 0 && visibleIdx < len(indices) {
		return indices[visibleIdx]
	}
	return 0
}

// GlobalIndexToVisible converts a global signal index to visible index (-1 if not visible)
func (m *Model) GlobalIndexToVisible(globalIdx int) int {
	indices := m.VisibleSignalIndices()
	for i, idx := range indices {
		if idx == globalIdx {
			return i
		}
	}
	return -1
}

// EnterSelectMode enters signal selection mode
func (m *Model) EnterSelectMode() {
	m.SelectMode = true
	m.SignalScrollOffset = 0
	m.adjustSignalScroll()
}

// ExitSelectMode exits signal selection mode
func (m *Model) ExitSelectMode() {
	m.SelectMode = false
	// 現在選択中の信号が非表示の場合、最初の表示信号を選択
	if !m.SignalVisible[m.SelectedSignal] {
		indices := m.VisibleSignalIndices()
		if len(indices) > 0 {
			m.SelectedSignal = indices[0]
		}
	}
	m.SignalScrollOffset = 0
	m.adjustSignalScroll()
}

// ToggleSelectMode toggles signal selection mode
func (m *Model) ToggleSelectMode() {
	if m.SelectMode {
		m.ExitSelectMode()
	} else {
		m.EnterSelectMode()
	}
}

// DisplaySignalCount returns the number of signals to display (depends on mode)
func (m *Model) DisplaySignalCount() int {
	if m.SelectMode {
		return len(m.Signals)
	}
	return len(m.VisibleSignalIndices())
}
