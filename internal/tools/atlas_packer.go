package tools

import (
	"encoding/json"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AtlasEntry represents a single sprite in the atlas.
type AtlasEntry struct {
	Name   string `json:"name"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// AtlasMetadata represents the JSON metadata for the atlas.
type AtlasMetadata struct {
	Image   string       `json:"image"`
	Entries []AtlasEntry `json:"entries"`
}

// AtlasPacker packs multiple images into a single sprite sheet.
type AtlasPacker struct {
	MaxWidth  int
	MaxHeight int
	Padding   int
}

// NewAtlasPacker creates a new atlas packer.
func NewAtlasPacker(maxWidth, maxHeight, padding int) *AtlasPacker {
	return &AtlasPacker{
		MaxWidth:  maxWidth,
		MaxHeight: maxHeight,
		Padding:   padding,
	}
}

// PackDirectory packs all PNG images in a directory into an atlas.
func (p *AtlasPacker) PackDirectory(inputDir, outputName string) error {
	// 1. Collect images
	images := make(map[string]image.Image)

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".png" {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			img, _, err := image.Decode(file)
			if err != nil {
				return err
			}

			relPath, _ := filepath.Rel(inputDir, path)
			images[relPath] = img
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(images) == 0 {
		return nil
	}

	// 2. Sort images by height (simple packing heuristic)
	type sortedImage struct {
		Name string
		Img  image.Image
	}

	sorted := make([]sortedImage, 0, len(images))
	for name, img := range images {
		sorted = append(sorted, sortedImage{Name: name, Img: img})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Img.Bounds().Dy() > sorted[j].Img.Bounds().Dy()
	})

	// 3. Pack images
	entries := make([]AtlasEntry, 0, len(sorted))
	canvas := image.NewRGBA(image.Rect(0, 0, p.MaxWidth, p.MaxHeight))

	currentX, currentY := 0, 0
	rowHeight := 0

	for _, item := range sorted {
		bounds := item.Img.Bounds()
		w, h := bounds.Dx(), bounds.Dy()

		// New row if doesn't fit horizontally
		if currentX+w > p.MaxWidth {
			currentX = 0
			currentY += rowHeight + p.Padding
			rowHeight = 0
		}

		// Check if fits vertically
		if currentY+h > p.MaxHeight {
			// Skip images that don't fit vertically
			continue
		}

		// Draw image
		draw.Draw(
			canvas,
			image.Rect(currentX, currentY, currentX+w, currentY+h),
			item.Img,
			image.Point{},
			draw.Src,
		)

		// Record entry
		entries = append(entries, AtlasEntry{
			Name:   item.Name,
			X:      currentX,
			Y:      currentY,
			Width:  w,
			Height: h,
		})

		// Advance position
		currentX += w + p.Padding

		if h > rowHeight {
			rowHeight = h
		}
	}

	// 4. Save atlas image
	outFile, err := os.Create(outputName + ".png")
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := png.Encode(outFile, canvas); err != nil {
		return err
	}

	// 5. Save metadata
	meta := AtlasMetadata{
		Image:   filepath.Base(outputName + ".png"),
		Entries: entries,
	}

	jsonFile, err := os.Create(outputName + ".json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")

	return encoder.Encode(meta)
}
