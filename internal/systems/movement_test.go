package systems

import (
	"testing"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// TestMovementSystem tests the movement system.
func TestMovementSystem(t *testing.T) {
	t.Run("NewMovementSystem creates system", func(t *testing.T) {
		world := ecs.NewWorld()
		ms := NewMovementSystem(&world)

		if ms == nil {
			t.Fatal("NewMovementSystem returned nil")
		}

		if ms.filter == nil {
			t.Error("MovementSystem.filter should not be nil")
		}
	})

	t.Run("Update applies velocity to position", func(t *testing.T) {
		world := ecs.NewWorld()
		ms := NewMovementSystem(&world)

		// Create an entity with position and velocity
		mapper := ecs.NewMap2[components.Position, components.Velocity](&world)
		entity := mapper.NewEntity(&components.Position{X: 100, Y: 100}, &components.Velocity{X: 5, Y: -3})

		// Update the system
		ms.Update(&world)

		// Verify position changed
		posMapper := ecs.NewMap1[components.Position](&world)
		pos := posMapper.Get(entity)

		if pos.X != 105 {
			t.Errorf("Position.X = %v, want 105", pos.X)
		}

		if pos.Y != 97 {
			t.Errorf("Position.Y = %v, want 97", pos.Y)
		}
	})

	t.Run("Update handles zero velocity", func(t *testing.T) {
		world := ecs.NewWorld()
		ms := NewMovementSystem(&world)

		// Create an entity with zero velocity
		mapper := ecs.NewMap2[components.Position, components.Velocity](&world)
		entity := mapper.NewEntity(&components.Position{X: 50, Y: 50}, &components.Velocity{X: 0, Y: 0})

		ms.Update(&world)

		posMapper := ecs.NewMap1[components.Position](&world)
		pos := posMapper.Get(entity)

		if pos.X != 50 || pos.Y != 50 {
			t.Errorf("Position should not change with zero velocity, got (%v, %v)", pos.X, pos.Y)
		}
	})

	t.Run("Update handles negative velocity", func(t *testing.T) {
		world := ecs.NewWorld()
		ms := NewMovementSystem(&world)

		mapper := ecs.NewMap2[components.Position, components.Velocity](&world)
		entity := mapper.NewEntity(&components.Position{X: 100, Y: 100}, &components.Velocity{X: -10, Y: -20})

		ms.Update(&world)

		posMapper := ecs.NewMap1[components.Position](&world)
		pos := posMapper.Get(entity)

		if pos.X != 90 {
			t.Errorf("Position.X = %v, want 90", pos.X)
		}

		if pos.Y != 80 {
			t.Errorf("Position.Y = %v, want 80", pos.Y)
		}
	})

	t.Run("Update processes multiple entities", func(t *testing.T) {
		world := ecs.NewWorld()
		ms := NewMovementSystem(&world)

		mapper := ecs.NewMap2[components.Position, components.Velocity](&world)
		e1 := mapper.NewEntity(&components.Position{X: 0, Y: 0}, &components.Velocity{X: 1, Y: 1})
		e2 := mapper.NewEntity(&components.Position{X: 100, Y: 100}, &components.Velocity{X: 2, Y: 2})

		ms.Update(&world)

		posMapper := ecs.NewMap1[components.Position](&world)
		pos1 := posMapper.Get(e1)
		pos2 := posMapper.Get(e2)

		if pos1.X != 1 || pos1.Y != 1 {
			t.Errorf("Entity 1 position wrong: got (%v, %v), want (1, 1)", pos1.X, pos1.Y)
		}

		if pos2.X != 102 || pos2.Y != 102 {
			t.Errorf("Entity 2 position wrong: got (%v, %v), want (102, 102)", pos2.X, pos2.Y)
		}
	})
}
