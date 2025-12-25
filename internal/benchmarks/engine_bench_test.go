package benchmarks_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
	"td/internal/systems"
)

// BenchmarkECSUpdate benchmarks a simple system update with many entities.
func BenchmarkECSUpdate(b *testing.B) {
	world := ecs.NewWorld()

	// Setup system (using MovementSystem as generic update example)
	sys := systems.NewMovementSystem(&world)
	// sys.SetWorldBounds(0, 0, 10000, 10000)

	// Create 10,000 entities
	mapper := ecs.NewMap2[components.Position, components.Velocity](&world)
	for i := range 10000 {
		mapper.NewEntity(
			&components.Position{X: float64(i), Y: float64(i)},
			&components.Velocity{X: 1.0, Y: 1.0},
		)
	}

	for b.Loop() {
		sys.Update(&world)
	}
}

// BenchmarkComponentCreation benchmarks entity creation cost.
func BenchmarkComponentCreation(b *testing.B) {
	for b.Loop() {
		world := ecs.NewWorld()
		mapper := ecs.NewMap2[components.Position, components.Sprite](&world)

		for range 1000 {
			mapper.NewEntity(
				&components.Position{},
				&components.Sprite{},
			)
		}
	}
}

// BenchmarkQueryIteration benchmarks pure query iteration speed.
func BenchmarkQueryIteration(b *testing.B) {
	world := ecs.NewWorld()
	mapper := ecs.NewMap2[components.Position, components.Velocity](&world)

	for i := range 10000 {
		mapper.NewEntity(
			&components.Position{X: float64(i), Y: float64(i)},
			&components.Velocity{X: 1.0, Y: 1.0},
		)
	}

	filter := ecs.NewFilter2[components.Position, components.Velocity](&world)

	for b.Loop() {
		query := filter.Query()
		for query.Next() {
			pos, vel := query.Get()
			pos.X += vel.X
		}
	}
}
