package systems

import (
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// MovementSystem updates entity positions based on velocity.
type MovementSystem struct {
	filter *ecs.Filter2[components.Position, components.Velocity]
}

// NewMovementSystem creates a new movement system.
func NewMovementSystem(world *ecs.World) *MovementSystem {
	return &MovementSystem{
		filter: ecs.NewFilter2[components.Position, components.Velocity](world),
	}
}

// Update applies velocity to position for all entities.
func (s *MovementSystem) Update(world *ecs.World) {
	query := s.filter.Query()
	for query.Next() {
		pos, vel := query.Get()
		pos.X += vel.X
		pos.Y += vel.Y
	}
}
