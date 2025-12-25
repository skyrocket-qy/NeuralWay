package systems

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// parallaxEntry holds data for sorting layers.
type parallaxEntry struct {
	pos   *components.Position
	layer *components.ParallaxLayer
}

// ParallaxRenderSystem renders parallax background layers with depth-based scrolling.
// Layers are sorted by their Layer value (lower = drawn first/furthest back).
type ParallaxRenderSystem struct {
	filter  *ecs.Filter2[components.Position, components.ParallaxLayer]
	cameraX float64
	cameraY float64
	viewW   int
	viewH   int
	layers  []parallaxEntry
	opts    *ebiten.DrawImageOptions
}

// NewParallaxRenderSystem creates a new parallax render system.
func NewParallaxRenderSystem(world *ecs.World) *ParallaxRenderSystem {
	return &ParallaxRenderSystem{
		filter: ecs.NewFilter2[components.Position, components.ParallaxLayer](world),
		viewW:  800,
		viewH:  600,
		layers: make([]parallaxEntry, 0, 16),
		opts:   &ebiten.DrawImageOptions{},
	}
}

// SetCamera sets the camera position and viewport size.
func (s *ParallaxRenderSystem) SetCamera(x, y float64, width, height int) {
	s.cameraX = x
	s.cameraY = y
	s.viewW = width
	s.viewH = height
}

// Draw renders all parallax layers sorted by depth.
func (s *ParallaxRenderSystem) Draw(world *ecs.World, screen *ebiten.Image) {
	// Collect all parallax layers
	s.layers = s.layers[:0]

	query := s.filter.Query()
	for query.Next() {
		pos, layer := query.Get()
		if layer.Image == nil {
			continue
		}

		s.layers = append(s.layers, parallaxEntry{pos: pos, layer: layer})
	}

	// Sort by layer (ascending = background first)
	sort.Slice(s.layers, func(i, j int) bool {
		return s.layers[i].layer.Layer < s.layers[j].layer.Layer
	})

	// Draw each layer
	for _, entry := range s.layers {
		s.drawLayer(screen, entry.pos, entry.layer)
	}
}

// drawLayer renders a single parallax layer with optional tiling.
func (s *ParallaxRenderSystem) drawLayer(
	screen *ebiten.Image,
	pos *components.Position,
	layer *components.ParallaxLayer,
) {
	img := layer.Image
	imgW := float64(img.Bounds().Dx())
	imgH := float64(img.Bounds().Dy())

	// Calculate parallax offset
	offsetX := pos.X - s.cameraX*layer.SpeedFactor
	offsetY := pos.Y - s.cameraY*layer.SpeedFactor

	if layer.RepeatX || layer.RepeatY {
		s.drawTiledLayer(screen, img, imgW, imgH, offsetX, offsetY, layer.RepeatX, layer.RepeatY)
	} else {
		// Single image draw
		s.opts.GeoM.Reset()
		s.opts.GeoM.Translate(offsetX, offsetY)
		screen.DrawImage(img, s.opts)
	}
}

// drawTiledLayer tiles an image across the viewport.
func (s *ParallaxRenderSystem) drawTiledLayer(
	screen, img *ebiten.Image,
	imgW, imgH, offsetX, offsetY float64,
	repeatX, repeatY bool,
) {
	// Normalize offset to prevent large coordinate values
	if repeatX && imgW > 0 {
		for offsetX > 0 {
			offsetX -= imgW
		}

		for offsetX < -imgW {
			offsetX += imgW
		}
	}

	if repeatY && imgH > 0 {
		for offsetY > 0 {
			offsetY -= imgH
		}

		for offsetY < -imgH {
			offsetY += imgH
		}
	}

	// Calculate tile range
	startX := offsetX
	endX := float64(s.viewW)
	startY := offsetY
	endY := float64(s.viewH)

	if !repeatX {
		endX = startX + imgW
	}

	if !repeatY {
		endY = startY + imgH
	}

	// Draw tiles
	for y := startY; y < endY; y += imgH {
		for x := startX; x < endX; x += imgW {
			s.opts.GeoM.Reset()
			s.opts.GeoM.Translate(x, y)
			screen.DrawImage(img, s.opts)

			if !repeatX {
				break
			}
		}

		if !repeatY {
			break
		}
	}
}

// Stats returns current camera position.
func (s *ParallaxRenderSystem) Stats() (cameraX, cameraY float64, viewW, viewH int) {
	return s.cameraX, s.cameraY, s.viewW, s.viewH
}
