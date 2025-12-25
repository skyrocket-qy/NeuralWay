package archetypes

import (
	"testing"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

func TestArchetype2(t *testing.T) {
	world := ecs.NewWorld()

	arch := NewArchetype2[components.Position, components.Velocity](&world)
	entity := arch.New(
		&components.Position{X: 10, Y: 20},
		&components.Velocity{X: 1, Y: 2},
	)

	if !world.Alive(entity) {
		t.Error("Entity should be alive")
	}

	// Verify components via Map
	posMap := ecs.NewMap[components.Position](&world)
	velMap := ecs.NewMap[components.Velocity](&world)

	pos := posMap.Get(entity)
	if pos.X != 10 || pos.Y != 20 {
		t.Errorf("Position mismatch: got (%v, %v), want (10, 20)", pos.X, pos.Y)
	}

	vel := velMap.Get(entity)
	if vel.X != 1 || vel.Y != 2 {
		t.Errorf("Velocity mismatch: got (%v, %v), want (1, 2)", vel.X, vel.Y)
	}
}

func TestArchetype3(t *testing.T) {
	world := ecs.NewWorld()

	arch := NewArchetype3[components.Position, components.Velocity, components.Health](&world)
	entity := arch.New(
		&components.Position{X: 5, Y: 10},
		&components.Velocity{X: 0.5, Y: -0.5},
		&components.Health{Current: 80, Max: 100},
	)

	if !world.Alive(entity) {
		t.Error("Entity should be alive")
	}

	healthMap := ecs.NewMap[components.Health](&world)

	health := healthMap.Get(entity)
	if health.Current != 80 || health.Max != 100 {
		t.Errorf("Health mismatch: got (%v/%v), want (80/100)", health.Current, health.Max)
	}
}

func TestArchetype4(t *testing.T) {
	world := ecs.NewWorld()

	arch := NewArchetype4[components.Position, components.Velocity, components.Health, components.Tag](&world)
	entity := arch.New(
		&components.Position{X: 0, Y: 0},
		&components.Velocity{X: 3, Y: 4},
		&components.Health{Current: 50, Max: 50},
		&components.Tag{Name: "player"},
	)

	if !world.Alive(entity) {
		t.Error("Entity should be alive")
	}

	tagMap := ecs.NewMap[components.Tag](&world)

	tag := tagMap.Get(entity)
	if tag.Name != "player" {
		t.Errorf("Tag mismatch: got %q, want \"player\"", tag.Name)
	}
}

func TestSpriteArchetype(t *testing.T) {
	world := ecs.NewWorld()

	arch := NewSpriteArchetype(&world)
	entity := arch.New(100, 200, nil) // nil image for testing

	if !world.Alive(entity) {
		t.Error("Entity should be alive")
	}

	posMap := ecs.NewMap[components.Position](&world)

	pos := posMap.Get(entity)
	if pos.X != 100 || pos.Y != 200 {
		t.Errorf("Position mismatch: got (%v, %v), want (100, 200)", pos.X, pos.Y)
	}

	spriteMap := ecs.NewMap[components.Sprite](&world)

	sprite := spriteMap.Get(entity)
	if !sprite.Visible {
		t.Error("Sprite should be visible by default")
	}

	if sprite.ScaleX != 1 || sprite.ScaleY != 1 {
		t.Errorf("Scale mismatch: got (%v, %v), want (1, 1)", sprite.ScaleX, sprite.ScaleY)
	}
}

func TestMovableArchetype(t *testing.T) {
	world := ecs.NewWorld()

	arch := NewMovableArchetype(&world)
	entity := arch.New(50, 60, 1.5, -2.5, nil)

	velMap := ecs.NewMap[components.Velocity](&world)

	vel := velMap.Get(entity)
	if vel.X != 1.5 || vel.Y != -2.5 {
		t.Errorf("Velocity mismatch: got (%v, %v), want (1.5, -2.5)", vel.X, vel.Y)
	}
}

func TestCollidableArchetype(t *testing.T) {
	world := ecs.NewWorld()

	arch := NewCollidableArchetype(&world)
	entity := arch.New(0, 0, 1, 1, nil, 32, 32, 0x01, 0x02)

	colliderMap := ecs.NewMap[components.Collider](&world)

	collider := colliderMap.Get(entity)
	if collider.Width != 32 || collider.Height != 32 {
		t.Errorf("Collider size mismatch: got (%v, %v), want (32, 32)", collider.Width, collider.Height)
	}

	if collider.Layer != 0x01 || collider.Mask != 0x02 {
		t.Errorf(
			"Collider layer/mask mismatch: got (0x%x, 0x%x), want (0x01, 0x02)",
			collider.Layer,
			collider.Mask,
		)
	}
}

func TestPickupArchetype(t *testing.T) {
	world := ecs.NewWorld()

	arch := NewPickupArchetype(&world)
	entity := arch.New(10, 20, nil, "health_pack")

	tagMap := ecs.NewMap[components.Tag](&world)

	tag := tagMap.Get(entity)
	if tag.Name != "health_pack" {
		t.Errorf("Tag mismatch: got %q, want \"health_pack\"", tag.Name)
	}
}
