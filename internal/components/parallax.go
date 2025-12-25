package components

import "github.com/hajimehoshi/ebiten/v2"

// ParallaxLayer defines a background layer that scrolls at a different speed than the camera.
// Lower Layer values are drawn first (furthest back), higher values are drawn on top.
type ParallaxLayer struct {
	// Image is the background image for this layer.
	Image *ebiten.Image

	// SpeedFactor controls scroll speed relative to camera.
	// 0.0 = static (doesn't move with camera)
	// 0.5 = moves at half camera speed (distant background)
	// 1.0 = moves at full camera speed (same as camera)
	SpeedFactor float64

	// RepeatX tiles the image horizontally across the viewport.
	RepeatX bool

	// RepeatY tiles the image vertically across the viewport.
	RepeatY bool

	// Layer controls draw order. Lower = drawn first (background).
	Layer int
}

// NewParallaxLayer creates a parallax layer with default settings.
func NewParallaxLayer(img *ebiten.Image, speedFactor float64, layer int) ParallaxLayer {
	return ParallaxLayer{
		Image:       img,
		SpeedFactor: speedFactor,
		RepeatX:     true,
		RepeatY:     false,
		Layer:       layer,
	}
}
