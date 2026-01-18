package update

import (
	"wave/internal/model"

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
