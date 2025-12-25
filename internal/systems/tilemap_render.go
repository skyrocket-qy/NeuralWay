package systems

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// TilemapRenderSystem renders Tilemap components with viewport culling.
// Only tiles visible within the camera viewport are drawn for performance.
type TilemapRenderSystem struct {
	filter  *ecs.Filter1[components.Tilemap]
	cameraX float64
	cameraY float64
	viewW   int
	viewH   int
	opts    *ebiten.DrawImageOptions
}

// NewTilemapRenderSystem creates a new tilemap render system.
func NewTilemapRenderSystem(world *ecs.World) *TilemapRenderSystem {
	return &TilemapRenderSystem{
		filter: ecs.NewFilter1[components.Tilemap](world),
		viewW:  800,
		viewH:  600,
		opts:   &ebiten.DrawImageOptions{},
	}
}

// SetCamera sets the camera position and viewport dimensions.
// Call this before Draw() to update the visible area.
func (s *TilemapRenderSystem) SetCamera(x, y float64, width, height int) {
	s.cameraX = x
	s.cameraY = y
	s.viewW = width
	s.viewH = height
}

// Draw renders all visible tilemaps with viewport culling.
func (s *TilemapRenderSystem) Draw(world *ecs.World, screen *ebiten.Image) {
	query := s.filter.Query()
	for query.Next() {
		tilemap := query.Get()
		s.drawTilemap(screen, tilemap)
	}
}

// drawTilemap renders a single tilemap with culling.
func (s *TilemapRenderSystem) drawTilemap(screen *ebiten.Image, tilemap *components.Tilemap) {
	tm := tilemap.Map
	if tm == nil {
		return
	}

	tileW := tm.TileWidth
	tileH := tm.TileHeight

	// Calculate visible tile range based on camera
	worldLeft := s.cameraX
	worldTop := s.cameraY
	worldRight := s.cameraX + float64(s.viewW)
	worldBottom := s.cameraY + float64(s.viewH)

	// Adjust for tilemap offset
	localLeft := worldLeft - tilemap.OffsetX
	localTop := worldTop - tilemap.OffsetY
	localRight := worldRight - tilemap.OffsetX
	localBottom := worldBottom - tilemap.OffsetY

	// Convert to tile coordinates with 1-tile padding for partial visibility
	minTileX := int(localLeft/float64(tileW)) - 1
	minTileY := int(localTop/float64(tileH)) - 1
	maxTileX := int(localRight/float64(tileW)) + 1
	maxTileY := int(localBottom/float64(tileH)) + 1

	// Clamp to map bounds
	if minTileX < 0 {
		minTileX = 0
	}

	if minTileY < 0 {
		minTileY = 0
	}

	if maxTileX > tm.Width {
		maxTileX = tm.Width
	}

	if maxTileY > tm.Height {
		maxTileY = tm.Height
	}

	// Draw visible layers
	for layerIdx, layer := range tm.Layers {
		// Check layer mask
		if tilemap.LayerMask&(1<<uint(layerIdx)) == 0 {
			continue
		}

		if layer.Type != "tilelayer" || !layer.Visible {
			continue
		}

		// Draw only visible tiles
		for y := minTileY; y < maxTileY; y++ {
			for x := minTileX; x < maxTileX; x++ {
				gid := tm.GetTileAt(&layer, x, y)
				if gid == 0 {
					continue
				}

				tileImg := tm.TileImages[gid]
				if tileImg == nil {
					continue
				}

				// Calculate screen position
				screenX := float64(x*tileW) + tilemap.OffsetX - s.cameraX
				screenY := float64(y*tileH) + tilemap.OffsetY - s.cameraY

				s.opts.GeoM.Reset()
				s.opts.GeoM.Translate(screenX, screenY)
				screen.DrawImage(tileImg, s.opts)
			}
		}
	}
}

// Stats returns the current camera position and visible area.
func (s *TilemapRenderSystem) Stats() (cameraX, cameraY float64, viewW, viewH int) {
	return s.cameraX, s.cameraY, s.viewW, s.viewH
}
