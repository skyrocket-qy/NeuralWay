package systems

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// YSortRenderSystem renders sprites sorted by Y position for 2.5D depth effect.
type YSortRenderSystem struct {
	filter      *ecs.Filter2[components.Position, components.Sprite]
	sortBuffer  []sortableSprite
	screen      *ebiten.Image
	CameraX     float64
	CameraY     float64
	YSortOffset float64 // Offset added to Y for sorting (e.g., sprite height)
}

type sortableSprite struct {
	entity   ecs.Entity
	position *components.Position
	sprite   *components.Sprite
	sortY    float64
}

// NewYSortRenderSystem creates a new Y-sorted render system.
func NewYSortRenderSystem(world *ecs.World) *YSortRenderSystem {
	return &YSortRenderSystem{
		filter:      ecs.NewFilter2[components.Position, components.Sprite](world),
		sortBuffer:  make([]sortableSprite, 0, 256),
		YSortOffset: 0,
	}
}

// SetScreen sets the target screen for rendering.
func (s *YSortRenderSystem) SetScreen(screen *ebiten.Image) {
	s.screen = screen
}

// Update collects and sorts entities.
func (s *YSortRenderSystem) Update(world *ecs.World) {
	// Collect all renderable entities
	s.sortBuffer = s.sortBuffer[:0]

	posMap := ecs.NewMap1[components.Position](world)
	spriteMap := ecs.NewMap1[components.Sprite](world)

	query := s.filter.Query()
	for query.Next() {
		entity := query.Entity()
		pos := posMap.Get(entity)
		sprite := spriteMap.Get(entity)

		s.sortBuffer = append(s.sortBuffer, sortableSprite{
			entity:   entity,
			position: pos,
			sprite:   sprite,
			sortY:    pos.Y + s.YSortOffset,
		})
	}

	// Sort by Y position (lower Y = further back = drawn first)
	sort.Slice(s.sortBuffer, func(i, j int) bool {
		return s.sortBuffer[i].sortY < s.sortBuffer[j].sortY
	})
}

// Draw renders all sprites in Y-sorted order.
func (s *YSortRenderSystem) Draw(world *ecs.World) {
	if s.screen == nil {
		return
	}

	for _, item := range s.sortBuffer {
		pos := item.position
		sprite := item.sprite

		if sprite.Image == nil {
			continue
		}

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(pos.X-s.CameraX, pos.Y-s.CameraY)
		s.screen.DrawImage(sprite.Image, opts)
	}
}
