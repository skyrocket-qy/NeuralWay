package systems

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// ShaderRenderSystem renders sprites with shader effects.
// Entities must have Position, Sprite, and ShaderEffect components.
type ShaderRenderSystem struct {
	filter *ecs.Filter3[components.Position, components.Sprite, components.ShaderEffect]
	opts   *ebiten.DrawRectShaderOptions
}

// NewShaderRenderSystem creates a new shader render system.
func NewShaderRenderSystem(world *ecs.World) *ShaderRenderSystem {
	return &ShaderRenderSystem{
		filter: ecs.NewFilter3[components.Position, components.Sprite, components.ShaderEffect](world),
		opts:   &ebiten.DrawRectShaderOptions{},
	}
}

// Draw renders all entities with shader effects.
func (s *ShaderRenderSystem) Draw(world *ecs.World, screen *ebiten.Image) {
	query := s.filter.Query()
	for query.Next() {
		pos, sprite, effect := query.Get()

		if !sprite.Visible || sprite.Image == nil {
			continue
		}

		if !effect.Enabled || effect.Shader == nil {
			// Fall back to normal rendering if shader disabled
			s.drawNormal(screen, pos, sprite)

			continue
		}

		s.drawWithShader(screen, pos, sprite, effect)
	}
}

// drawNormal renders without shader effect.
func (s *ShaderRenderSystem) drawNormal(
	screen *ebiten.Image,
	pos *components.Position,
	sprite *components.Sprite,
) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sprite.ScaleX, sprite.ScaleY)
	op.GeoM.Translate(pos.X+sprite.OffsetX, pos.Y+sprite.OffsetY)
	screen.DrawImage(sprite.Image, op)
}

// drawWithShader renders the sprite with the shader effect.
func (s *ShaderRenderSystem) drawWithShader(
	screen *ebiten.Image,
	pos *components.Position,
	sprite *components.Sprite,
	effect *components.ShaderEffect,
) {
	bounds := sprite.Image.Bounds()
	w := float64(bounds.Dx()) * sprite.ScaleX
	h := float64(bounds.Dy()) * sprite.ScaleY

	// Reset options
	s.opts.GeoM.Reset()
	s.opts.GeoM.Scale(sprite.ScaleX, sprite.ScaleY)
	s.opts.GeoM.Translate(pos.X+sprite.OffsetX, pos.Y+sprite.OffsetY)

	// Set source image
	s.opts.Images[0] = sprite.Image

	// Copy uniforms
	s.opts.Uniforms = effect.Uniforms

	// Draw with shader
	screen.DrawRectShader(int(w), int(h), effect.Shader, s.opts)
}
