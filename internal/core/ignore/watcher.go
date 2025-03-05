package ignore

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// Watcher watches for changes to the .golineignore file and reloads the ignore controller
type Watcher struct {
	controller     *Controller
	ignoreFilePath string
	stopChan       chan struct{}
	interval       time.Duration
	lastModTime    time.Time
}

// NewWatcher creates a new watcher for the given controller
func NewWatcher(controller *Controller, cwd string) *Watcher {
	return &Watcher{
		controller:     controller,
		ignoreFilePath: filepath.Join(cwd, ".golineignore"),
		stopChan:       make(chan struct{}),
		interval:       2 * time.Second, // Check every 2 seconds
	}
}

// Start starts the watcher
func (w *Watcher) Start() {
	// Get initial modification time
	w.updateLastModTime()

	// Start watching for changes
	go w.watch()
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	close(w.stopChan)
}

// watch periodically checks for changes to the .golineignore file
func (w *Watcher) watch() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkForChanges()
		case <-w.stopChan:
			return
		}
	}
}

// checkForChanges checks if the .golineignore file has changed
func (w *Watcher) checkForChanges() {
	// Get current modification time
	fileInfo, err := os.Stat(w.ignoreFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File was deleted, reload controller
			if !w.lastModTime.IsZero() {
				w.lastModTime = time.Time{}
				err := w.controller.Reload()
				if err != nil {
					log.Printf("Error reloading ignore controller: %v", err)
				}
			}
		}
		return
	}

	// Check if file was modified
	modTime := fileInfo.ModTime()
	if modTime != w.lastModTime {
		w.lastModTime = modTime
		err := w.controller.Reload()
		if err != nil {
			log.Printf("Error reloading ignore controller: %v", err)
		}
	}
}

// updateLastModTime updates the last modification time of the .golineignore file
func (w *Watcher) updateLastModTime() {
	fileInfo, err := os.Stat(w.ignoreFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.lastModTime = time.Time{}
		}
		return
	}
	w.lastModTime = fileInfo.ModTime()
}
