package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/png" // Register PNG decoder
	"io/fs"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// Loader handles loading and caching game assets.
type Loader struct {
	images map[string]*ebiten.Image
	fs     fs.FS
}

// NewLoader creates a new asset loader.
// Pass nil to load from disk, or an embed.FS to load from embedded files.
func NewLoader(fileSystem fs.FS) *Loader {
	return &Loader{
		images: make(map[string]*ebiten.Image),
		fs:     fileSystem,
	}
}

// NewEmbedLoader creates a loader from an embed.FS.
func NewEmbedLoader(efs embed.FS, root string) (*Loader, error) {
	subFS, err := fs.Sub(efs, root)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub filesystem: %w", err)
	}
	return NewLoader(subFS), nil
}

// LoadImage loads an image from the filesystem and caches it.
func (l *Loader) LoadImage(path string) (*ebiten.Image, error) {
	// Check cache first
	if img, ok := l.images[path]; ok {
		return img, nil
	}

	// Read file
	data, err := fs.ReadFile(l.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read image %s: %w", path, err)
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image %s: %w", path, err)
	}

	// Convert to Ebitengine image
	eImg := ebiten.NewImageFromImage(img)
	l.images[path] = eImg

	return eImg, nil
}

// LoadAllImages loads all PNG images from a directory.
func (l *Loader) LoadAllImages(dir string) (map[string]*ebiten.Image, error) {
	result := make(map[string]*ebiten.Image)

	err := fs.WalkDir(l.fs, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
			img, err := l.LoadImage(path)
			if err != nil {
				return err
			}
			result[path] = img
		}
		return nil
	})

	return result, err
}

// GetImage returns a cached image or nil if not loaded.
func (l *Loader) GetImage(path string) *ebiten.Image {
	return l.images[path]
}

// Clear removes all cached images.
func (l *Loader) Clear() {
	l.images = make(map[string]*ebiten.Image)
}

// ImageCount returns the number of cached images.
func (l *Loader) ImageCount() int {
	return len(l.images)
}
