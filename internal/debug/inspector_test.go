package debug

import (
	"testing"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

func TestInspectorToggle(t *testing.T) {
	world := ecs.NewWorld()
	inspector := NewInspector(&world)

	if inspector.Enabled() {
		t.Error("Inspector should be disabled by default")
	}

	inspector.Toggle()

	if !inspector.Enabled() {
		t.Error("Inspector should be enabled after toggle")
	}

	inspector.Toggle()

	if inspector.Enabled() {
		t.Error("Inspector should be disabled after second toggle")
	}
}

func TestInspectorSetCamera(t *testing.T) {
	world := ecs.NewWorld()
	inspector := NewInspector(&world)

	inspector.SetCamera(100, 200)

	// Camera values are private, but we can verify no panic
	// In a real test, we'd expose stats or use a getter
}

func TestInspectorStatsWithEntities(t *testing.T) {
	world := ecs.NewWorld()
	inspector := NewInspector(&world)

	// Create some test entities
	posMapper := ecs.NewMap1[components.Position](&world)
	spriteMapper := ecs.NewMap2[components.Position, components.Sprite](&world)

	// 3 position-only entities
	for i := range 3 {
		posMapper.NewEntity(&components.Position{X: float64(i * 10), Y: 0})
	}

	// 2 sprite entities (also have position)
	for i := range 2 {
		spriteMapper.NewEntity(
			&components.Position{X: float64(i * 10), Y: 100},
			&components.Sprite{Visible: true},
		)
	}

	// Enable and update to refresh stats
	inspector.Toggle()
	inspector.Update()

	// Stats should be updated (we can't access private fields, but no panic = success)
}

func TestInspectorUpdateWhenDisabled(t *testing.T) {
	world := ecs.NewWorld()
	inspector := NewInspector(&world)

	// Should not panic when updating while disabled
	inspector.Update()
}
