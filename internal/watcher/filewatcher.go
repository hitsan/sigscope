package watcher

import (
	"time"

	"github.com/fsnotify/fsnotify"
	tea "github.com/charmbracelet/bubbletea"
)

// FileChangedMsg is sent when the watched file changes
type FileChangedMsg struct {
	Filename string
	Error    error
}

// FileWatchErrorMsg is sent when the file watcher encounters an error
type FileWatchErrorMsg struct {
	Error error
}

// WatchFile creates a command that watches a file for changes
func WatchFile(filename string) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return FileWatchErrorMsg{Error: err}
		}
		defer watcher.Close()

		err = watcher.Add(filename)
		if err != nil {
			return FileWatchErrorMsg{Error: err}
		}

		// Debounce timer
		const debounceDelay = 200 * time.Millisecond
		var debounceTimer <-chan time.Time

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return FileWatchErrorMsg{Error: err}
				}

				// Filter for relevant events: WRITE, CREATE, RENAME
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Rename == fsnotify.Rename {

					// Reset debounce timer
					debounceTimer = time.After(debounceDelay)
				}

			case <-debounceTimer:
				// Debounce period elapsed, send file changed message
				return FileChangedMsg{
					Filename: filename,
					Error:    nil,
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return FileWatchErrorMsg{Error: err}
				}
				return FileWatchErrorMsg{Error: err}
			}
		}
	}
}
