// Package archetypes provides predefined entity templates for common game object patterns.
// Archetypes wrap ark ECS's component mappers to provide type-safe, convenient entity creation.
//
// This package provides:
//   - Generic builders (Archetype2, Archetype3, Archetype4) for custom archetypes
//   - Game-specific archetypes (SpriteArchetype, MovableArchetype, etc.) in game_archetypes.go
package archetypes

import (
	"github.com/mlange-42/ark/ecs"
)

// Archetype2 creates entities with 2 components.
type Archetype2[A, B any] struct {
	mapper *ecs.Map2[A, B]
}

// NewArchetype2 creates a new 2-component archetype.
func NewArchetype2[A, B any](world *ecs.World) *Archetype2[A, B] {
	return &Archetype2[A, B]{
		mapper: ecs.NewMap2[A, B](world),
	}
}

// New creates an entity with the given component values.
func (a *Archetype2[A, B]) New(c1 *A, c2 *B) ecs.Entity {
	return a.mapper.NewEntity(c1, c2)
}

// Archetype3 creates entities with 3 components.
type Archetype3[A, B, C any] struct {
	mapper *ecs.Map3[A, B, C]
}

// NewArchetype3 creates a new 3-component archetype.
func NewArchetype3[A, B, C any](world *ecs.World) *Archetype3[A, B, C] {
	return &Archetype3[A, B, C]{
		mapper: ecs.NewMap3[A, B, C](world),
	}
}

// New creates an entity with the given component values.
func (a *Archetype3[A, B, C]) New(c1 *A, c2 *B, c3 *C) ecs.Entity {
	return a.mapper.NewEntity(c1, c2, c3)
}

// Archetype4 creates entities with 4 components.
type Archetype4[A, B, C, D any] struct {
	mapper *ecs.Map4[A, B, C, D]
}

// NewArchetype4 creates a new 4-component archetype.
func NewArchetype4[A, B, C, D any](world *ecs.World) *Archetype4[A, B, C, D] {
	return &Archetype4[A, B, C, D]{
		mapper: ecs.NewMap4[A, B, C, D](world),
	}
}

// New creates an entity with the given component values.
func (a *Archetype4[A, B, C, D]) New(c1 *A, c2 *B, c3 *C, c4 *D) ecs.Entity {
	return a.mapper.NewEntity(c1, c2, c3, c4)
}
