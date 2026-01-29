package update

import (
	"time"

	"sigscope/internal/model"
	"sigscope/internal/vcd"
	"sigscope/internal/watcher"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all key events and returns updated model
func Update(m model.Model, msg tea.Msg) (model.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return handleKey(m, msg)
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case watcher.FileChangedMsg:
		return handleFileChanged(m, msg)
	case watcher.FileWatchErrorMsg:
		return handleWatchError(m, msg)
	}
	return m, nil
}

func handleKey(m model.Model, msg tea.KeyMsg) (model.Model, tea.Cmd) {
	// Handle search mode separately
	if m.Mode == model.ModeSearch {
		return handleSearchKey(m, msg)
	}

	switch msg.String() {
	// Quit
	case "q", "ctrl+c":
		return m, tea.Quit

	// Signal navigation (up/down)
	case "j", "down":
		m.MoveSignalDown()
	case "k", "up":
		m.MoveSignalUp()

	// Time navigation (left/right)
	case "h", "left":
		scrollAmount := (m.TimeEnd - m.TimeStart) / 10
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		m.ScrollTimeLeft(scrollAmount)
	case "l", "right":
		scrollAmount := (m.TimeEnd - m.TimeStart) / 10
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		m.ScrollTimeRight(scrollAmount)

	// Page navigation
	case "H", "shift+left":
		scrollAmount := (m.TimeEnd - m.TimeStart) / 2
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		m.ScrollTimeLeft(scrollAmount)
	case "L", "shift+right":
		scrollAmount := (m.TimeEnd - m.TimeStart) / 2
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		m.ScrollTimeRight(scrollAmount)

	// Zoom
	case "+", "=":
		m.ZoomIn()
	case "-", "_":
		m.ZoomOut()
	case "0":
		m.ResetZoom()

	// Jump to start/end
	case "g":
		m.GoToStart()
	case "G":
		m.GoToEnd()

	// Cursor movement mode (toggle)
	case "c":
		m.CursorVisible = !m.CursorVisible

	// Jump to prev/next value change
	case "[":
		m.PrevChange()
	case "]":
		m.NextChange()

	// Search mode
	case "/":
		m.Mode = model.ModeSearch
		m.SearchQuery = ""

	// Signal selection mode
	case "s":
		m.ToggleSelectMode()

	// Toggle 1-line/2-line display mode
	case "t":
		m.ToggleTwoLineMode()

	// Toggle signal visibility (select mode only)
	case " ":
		if m.SelectMode {
			m.ToggleSignalVisibility()
		}

	// Show all signals (select mode only)
	case "a":
		if m.SelectMode {
			m.SetAllSignalsVisible(true)
		}

	// Hide all signals (select mode only)
	case "A":
		if m.SelectMode {
			m.SetAllSignalsVisible(false)
		}
	}

	return m, nil
}

func handleSearchKey(m model.Model, msg tea.KeyMsg) (model.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.Search(m.SearchQuery)
		m.Mode = model.ModeNormal
	case "esc":
		m.Mode = model.ModeNormal
		m.SearchQuery = ""
	case "backspace":
		if len(m.SearchQuery) > 0 {
			m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
		}
	default:
		// Add character to search query
		if len(msg.String()) == 1 {
			m.SearchQuery += msg.String()
		}
	}
	return m, nil
}

func handleFileChanged(m model.Model, msg watcher.FileChangedMsg) (model.Model, tea.Cmd) {
	if msg.Error != nil {
		m.WatchError = msg.Error.Error()
		return m, watcher.WatchFile(m.Filename)
	}

	// VCDファイルを再パース
	vcdFile, err := vcd.Parse(m.Filename)
	if err != nil {
		m.ReloadError = err.Error()
		return m, watcher.WatchFile(m.Filename)
	}

	// 現在の状態を保存
	savedState := model.ViewState{
		CursorTime:         m.CursorTime,
		SelectedSignal:     m.SelectedSignal,
		TimeStart:          m.TimeStart,
		TimeEnd:            m.TimeEnd,
		Zoom:               m.Zoom,
		SignalScrollOffset: m.SignalScrollOffset,
		SelectMode:         m.SelectMode,
		TwoLineMode:        m.TwoLineMode,
		SignalVisible:      append([]bool{}, m.SignalVisible...),
		SignalNames:        m.ExtractSignalNames(),
	}

	// 新しいモデルを構築
	newModel := model.NewModel(vcdFile, m.Filename)

	// 状態を復元
	newModel.RestoreViewState(savedState)

	// 端末サイズを復元
	newModel.Width = m.Width
	newModel.Height = m.Height

	// 再読み込み成功を記録
	newModel.LastReloadTime = time.Now()
	newModel.ReloadError = ""
	newModel.WatchError = ""

	return newModel, watcher.WatchFile(m.Filename)
}

func handleWatchError(m model.Model, msg watcher.FileWatchErrorMsg) (model.Model, tea.Cmd) {
	m.WatchError = msg.Error.Error()
	return m, watcher.WatchFile(m.Filename)
}
