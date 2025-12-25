package archetypes

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// ============================================================================
// Game-Specific Archetypes
// These are convenience archetypes using framework components.
// Use them as-is or as templates for your own archetypes.
// ============================================================================

// SpriteArchetype creates entities with Position and Sprite components.
type SpriteArchetype struct {
	arch *Archetype2[components.Position, components.Sprite]
}

// NewSpriteArchetype creates a sprite archetype.
func NewSpriteArchetype(world *ecs.World) *SpriteArchetype {
	return &SpriteArchetype{
		arch: NewArchetype2[components.Position, components.Sprite](world),
	}
}

// New creates a static sprite entity.
func (a *SpriteArchetype) New(x, y float64, img *ebiten.Image) ecs.Entity {
	return a.arch.New(
		&components.Position{X: x, Y: y},
		&components.Sprite{
			Image:   img,
			ScaleX:  1,
			ScaleY:  1,
			Visible: true,
		},
	)
}

// MovableArchetype creates entities with Position, Velocity, and Sprite components.
type MovableArchetype struct {
	arch *Archetype3[components.Position, components.Velocity, components.Sprite]
}

// NewMovableArchetype creates a movable sprite archetype.
func NewMovableArchetype(world *ecs.World) *MovableArchetype {
	return &MovableArchetype{
		arch: NewArchetype3[components.Position, components.Velocity, components.Sprite](world),
	}
}

// New creates a movable sprite entity.
func (a *MovableArchetype) New(x, y, vx, vy float64, img *ebiten.Image) ecs.Entity {
	return a.arch.New(
		&components.Position{X: x, Y: y},
		&components.Velocity{X: vx, Y: vy},
		&components.Sprite{
			Image:   img,
			ScaleX:  1,
			ScaleY:  1,
			Visible: true,
		},
	)
}

// CollidableArchetype creates entities with Position, Velocity, Sprite, and Collider.
type CollidableArchetype struct {
	arch *Archetype4[components.Position, components.Velocity, components.Sprite, components.Collider]
}

// NewCollidableArchetype creates a collidable sprite archetype.
func NewCollidableArchetype(world *ecs.World) *CollidableArchetype {
	return &CollidableArchetype{
		arch: NewArchetype4[components.Position, components.Velocity, components.Sprite, components.Collider](
			world,
		),
	}
}

// New creates a collidable sprite entity.
func (a *CollidableArchetype) New(
	x, y, vx, vy float64,
	img *ebiten.Image,
	width, height float64,
	layer, mask uint32,
) ecs.Entity {
	return a.arch.New(
		&components.Position{X: x, Y: y},
		&components.Velocity{X: vx, Y: vy},
		&components.Sprite{
			Image:   img,
			ScaleX:  1,
			ScaleY:  1,
			Visible: true,
		},
		&components.Collider{
			Width:  width,
			Height: height,
			Layer:  layer,
			Mask:   mask,
		},
	)
}

// ProjectileArchetype creates projectile entities.
type ProjectileArchetype struct {
	arch *Archetype3[components.Position, components.Velocity, components.Sprite]
}

// NewProjectileArchetype creates a projectile archetype.
func NewProjectileArchetype(world *ecs.World) *ProjectileArchetype {
	return &ProjectileArchetype{
		arch: NewArchetype3[components.Position, components.Velocity, components.Sprite](world),
	}
}

// New creates a projectile entity moving from origin toward target.
func (a *ProjectileArchetype) New(x, y, vx, vy float64, img *ebiten.Image) ecs.Entity {
	return a.arch.New(
		&components.Position{X: x, Y: y},
		&components.Velocity{X: vx, Y: vy},
		&components.Sprite{
			Image:   img,
			ScaleX:  1,
			ScaleY:  1,
			Visible: true,
		},
	)
}

// PickupArchetype creates pickup/collectible entities.
type PickupArchetype struct {
	arch *Archetype3[components.Position, components.Sprite, components.Tag]
}

// NewPickupArchetype creates a pickup archetype.
func NewPickupArchetype(world *ecs.World) *PickupArchetype {
	return &PickupArchetype{
		arch: NewArchetype3[components.Position, components.Sprite, components.Tag](world),
	}
}

// New creates a pickup entity.
func (a *PickupArchetype) New(x, y float64, img *ebiten.Image, tag string) ecs.Entity {
	return a.arch.New(
		&components.Position{X: x, Y: y},
		&components.Sprite{
			Image:   img,
			ScaleX:  1,
			ScaleY:  1,
			Visible: true,
		},
		&components.Tag{Name: tag},
	)
}
