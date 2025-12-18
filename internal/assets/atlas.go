package assets

import (
	"fmt"
	"image"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

// AtlasRegion represents a rectangular region in a texture atlas.
type AtlasRegion struct {
	X, Y          int
	Width, Height int
	Atlas         *ebiten.Image
}

// SubImage returns an Ebitengine sub-image for this region.
func (r *AtlasRegion) SubImage() *ebiten.Image {
	return r.Atlas.SubImage(image.Rect(r.X, r.Y, r.X+r.Width, r.Y+r.Height)).(*ebiten.Image)
}

// TextureAtlas packs multiple images into a single texture for efficient rendering.
type TextureAtlas struct {
	Image   *ebiten.Image
	Regions map[string]*AtlasRegion
	Width   int
	Height  int
}

// AtlasBuilder helps construct texture atlases from multiple images.
type AtlasBuilder struct {
	images map[string]*ebiten.Image
}

// NewAtlasBuilder creates a new atlas builder.
func NewAtlasBuilder() *AtlasBuilder {
	return &AtlasBuilder{
		images: make(map[string]*ebiten.Image),
	}
}

// Add adds an image to be packed into the atlas.
func (b *AtlasBuilder) Add(name string, img *ebiten.Image) {
	b.images[name] = img
}

// Build creates the texture atlas using a simple row-based packing algorithm.
// maxWidth specifies the maximum width of the atlas.
func (b *AtlasBuilder) Build(maxWidth int) (*TextureAtlas, error) {
	if len(b.images) == 0 {
		return nil, fmt.Errorf("no images to pack")
	}

	// Sort images by height (tallest first) for better packing
	type imgEntry struct {
		name string
		img  *ebiten.Image
	}
	entries := make([]imgEntry, 0, len(b.images))
	for name, img := range b.images {
		entries = append(entries, imgEntry{name, img})
	}
	sort.Slice(entries, func(i, j int) bool {
		_, hi := entries[i].img.Bounds().Dx(), entries[i].img.Bounds().Dy()
		_, hj := entries[j].img.Bounds().Dx(), entries[j].img.Bounds().Dy()
		return hi > hj
	})

	// Simple row-based packing
	regions := make(map[string]*AtlasRegion)
	x, y := 0, 0
	rowHeight := 0
	atlasHeight := 0

	for _, entry := range entries {
		w := entry.img.Bounds().Dx()
		h := entry.img.Bounds().Dy()

		// Move to next row if needed
		if x+w > maxWidth {
			x = 0
			y += rowHeight
			rowHeight = 0
		}

		regions[entry.name] = &AtlasRegion{
			X:      x,
			Y:      y,
			Width:  w,
			Height: h,
		}

		x += w
		if h > rowHeight {
			rowHeight = h
		}
		if y+rowHeight > atlasHeight {
			atlasHeight = y + rowHeight
		}
	}

	// Create the atlas image
	atlas := ebiten.NewImage(maxWidth, atlasHeight)

	// Draw all images into the atlas
	for name, region := range regions {
		img := b.images[name]
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(region.X), float64(region.Y))
		atlas.DrawImage(img, op)
		region.Atlas = atlas
	}

	return &TextureAtlas{
		Image:   atlas,
		Regions: regions,
		Width:   maxWidth,
		Height:  atlasHeight,
	}, nil
}

// Get returns a region by name.
func (a *TextureAtlas) Get(name string) (*AtlasRegion, bool) {
	r, ok := a.Regions[name]
	return r, ok
}

// GetSubImage returns a sub-image for a named region.
func (a *TextureAtlas) GetSubImage(name string) *ebiten.Image {
	if r, ok := a.Regions[name]; ok {
		return r.SubImage()
	}
	return nil
}
