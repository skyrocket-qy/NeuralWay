package systems

import (
	"td/internal/components"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
)

// RenderSystem draws all entities with Position and Sprite components.
type RenderSystem struct {
	filter *ecs.Filter2[components.Position, components.Sprite]
}

// NewRenderSystem creates a new render system.
func NewRenderSystem(world *ecs.World) *RenderSystem {
	return &RenderSystem{
		filter: ecs.NewFilter2[components.Position, components.Sprite](world),
	}
}

// Draw renders all visible sprites to the screen.
func (s *RenderSystem) Draw(world *ecs.World, screen *ebiten.Image) {
	query := s.filter.Query()
	for query.Next() {
		pos, sprite := query.Get()

		if !sprite.Visible || sprite.Image == nil {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(sprite.ScaleX, sprite.ScaleY)
		op.GeoM.Translate(pos.X+sprite.OffsetX, pos.Y+sprite.OffsetY)

		screen.DrawImage(sprite.Image, op)
	}
}
