package components

import "github.com/hajimehoshi/ebiten/v2"

// Position represents the 2D position of an entity.
type Position struct {
	X, Y float64
}

// Velocity represents the 2D velocity of an entity.
type Velocity struct {
	X, Y float64
}

// Sprite represents a renderable image for an entity.
type Sprite struct {
	Image   *ebiten.Image
	OffsetX float64
	OffsetY float64
	ScaleX  float64
	ScaleY  float64
	Visible bool
}

// NewSprite creates a sprite with default values.
func NewSprite(img *ebiten.Image) Sprite {
	return Sprite{
		Image:   img,
		OffsetX: 0,
		OffsetY: 0,
		ScaleX:  1,
		ScaleY:  1,
		Visible: true,
	}
}

// Collider represents a collision bounding box.
type Collider struct {
	Width  float64
	Height float64
	Layer  uint32 // Collision layer bitmask
	Mask   uint32 // Layers this collider interacts with
}

// Health represents entity health.
type Health struct {
	Current int
	Max     int
}

// NewHealth creates a health component with full health.
func NewHealth(max int) Health {
	return Health{Current: max, Max: max}
}

// Tag is a marker component for entity categorization.
type Tag struct {
	Name string
}
