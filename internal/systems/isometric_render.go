package systems

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// IsometricRenderSystem renders entities in isometric space.
type IsometricRenderSystem struct {
	filter     *ecs.Filter2[components.IsometricPosition, components.Sprite]
	sortBuffer []isoSprite
	screen     *ebiten.Image
	TileWidth  float64
	TileHeight float64
	CameraX    float64
	CameraY    float64
	OffsetX    float64 // Screen offset for centering
	OffsetY    float64
}

type isoSprite struct {
	entity  ecs.Entity
	isoPos  *components.IsometricPosition
	sprite  *components.Sprite
	screenX float64
	screenY float64
	sortKey float64
}

// NewIsometricRenderSystem creates a new isometric render system.
func NewIsometricRenderSystem(world *ecs.World, tileWidth, tileHeight float64) *IsometricRenderSystem {
	return &IsometricRenderSystem{
		filter:     ecs.NewFilter2[components.IsometricPosition, components.Sprite](world),
		sortBuffer: make([]isoSprite, 0, 256),
		TileWidth:  tileWidth,
		TileHeight: tileHeight,
	}
}

// SetScreen sets the target screen.
func (s *IsometricRenderSystem) SetScreen(screen *ebiten.Image) {
	s.screen = screen
	if screen != nil {
		w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
		s.OffsetX = float64(w) / 2
		s.OffsetY = float64(h) / 4
	}
}

// Update collects and sorts entities.
func (s *IsometricRenderSystem) Update(world *ecs.World) {
	s.sortBuffer = s.sortBuffer[:0]

	isoPosMap := ecs.NewMap1[components.IsometricPosition](world)
	spriteMap := ecs.NewMap1[components.Sprite](world)

	query := s.filter.Query()
	for query.Next() {
		entity := query.Entity()
		isoPos := isoPosMap.Get(entity)
		sprite := spriteMap.Get(entity)

		screenX, screenY := isoPos.ToScreen(s.TileWidth, s.TileHeight)

		// Sort by isoX + isoY (depth) then by height
		sortKey := isoPos.IsoX + isoPos.IsoY - isoPos.IsoZ*0.001

		s.sortBuffer = append(s.sortBuffer, isoSprite{
			entity:  entity,
			isoPos:  isoPos,
			sprite:  sprite,
			screenX: screenX,
			screenY: screenY,
			sortKey: sortKey,
		})
	}

	// Sort by depth (back to front)
	sort.Slice(s.sortBuffer, func(i, j int) bool {
		return s.sortBuffer[i].sortKey < s.sortBuffer[j].sortKey
	})
}

// Draw renders all isometric sprites.
func (s *IsometricRenderSystem) Draw(world *ecs.World) {
	if s.screen == nil {
		return
	}

	for _, item := range s.sortBuffer {
		sprite := item.sprite
		if sprite.Image == nil {
			continue
		}

		opts := &ebiten.DrawImageOptions{}

		// Apply screen position with camera and offset
		drawX := item.screenX - s.CameraX + s.OffsetX
		drawY := item.screenY - s.CameraY + s.OffsetY

		opts.GeoM.Translate(drawX, drawY)
		s.screen.DrawImage(sprite.Image, opts)
	}
}

// WorldToScreen converts world isometric position to screen position.
func (s *IsometricRenderSystem) WorldToScreen(isoX, isoY, isoZ float64) (screenX, screenY float64) {
	pos := components.IsometricPosition{IsoX: isoX, IsoY: isoY, IsoZ: isoZ}
	sx, sy := pos.ToScreen(s.TileWidth, s.TileHeight)

	return sx - s.CameraX + s.OffsetX, sy - s.CameraY + s.OffsetY
}

// ScreenToWorld converts screen position to isometric coordinates.
func (s *IsometricRenderSystem) ScreenToWorld(screenX, screenY float64) (isoX, isoY float64) {
	// Adjust for camera and offset
	adjX := screenX + s.CameraX - s.OffsetX
	adjY := screenY + s.CameraY - s.OffsetY

	pos := components.FromScreen(adjX, adjY, s.TileWidth, s.TileHeight)

	return pos.IsoX, pos.IsoY
}
