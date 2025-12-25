package assets

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// HotReloader watches files for changes and triggers reloads.
// Uses polling to avoid external dependencies and CGO.
type HotReloader struct {
	paths    []string
	modTimes map[string]time.Time
	interval time.Duration
	loader   *Loader
	onChange func(path string)
	stopChan chan struct{}
	running  bool
}

// NewHotReloader creates a new hot reloader.
func NewHotReloader(loader *Loader, interval time.Duration) *HotReloader {
	return &HotReloader{
		paths:    make([]string, 0),
		modTimes: make(map[string]time.Time),
		interval: interval,
		loader:   loader,
		stopChan: make(chan struct{}),
	}
}

// Watch adds a directory or file to watch.
func (h *HotReloader) Watch(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return filepath.Walk(path, func(p string, d os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				h.addFile(p, d.ModTime())
			}

			return nil
		})
	}

	h.addFile(path, info.ModTime())

	return nil
}

func (h *HotReloader) addFile(path string, modTime time.Time) {
	h.paths = append(h.paths, path)
	h.modTimes[path] = modTime
}

// Start begins the polling loop.
func (h *HotReloader) Start(onChange func(path string)) {
	h.onChange = onChange

	h.running = true
	go h.loop()
}

// Stop stops the polling loop.
func (h *HotReloader) Stop() {
	if h.running {
		close(h.stopChan)
		h.running = false
	}
}

func (h *HotReloader) loop() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopChan:
			return
		case <-ticker.C:
			h.checkChanges()
		}
	}
}

func (h *HotReloader) checkChanges() {
	for _, path := range h.paths {
		info, err := os.Stat(path)
		if err != nil {
			continue // File might have been deleted temporarily
		}

		if info.ModTime().After(h.modTimes[path]) {
			log.Printf("Asset changed: %s", path)
			h.modTimes[path] = info.ModTime()

			// Clear from loader cache logic would go here if Loader exposed it selectively
			// For now, we just notify
			if h.onChange != nil {
				h.onChange(path)
			}
		}
	}
}

// ReloadAsset forces reload of a specific asset in the loader.
// Note: This assumes Loader has a method to clear/reload specific paths.
// Since standard Loader only has Clear(), we might need to extend Loader or just Clear() all.
func (h *HotReloader) ReloadAsset(path string) {
	// In a real implementation: h.loader.Unload(path)
	// Then: h.loader.LoadImage(path)
}
