package systems

import (
	"testing"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

func TestBatchRenderSystemStats(t *testing.T) {
	world := ecs.NewWorld()

	batchSystem := NewBatchRenderSystem(&world)

	// Create sprites with the same "image" (nil for testing)
	mapper := ecs.NewMap2[components.Position, components.Sprite](&world)

	// Create 5 entities without images (should be skipped)
	for i := range 5 {
		mapper.NewEntity(
			&components.Position{X: float64(i * 10), Y: float64(i * 10)},
			&components.Sprite{Image: nil, Visible: true, ScaleX: 1, ScaleY: 1},
		)
	}

	// Stats should be 0 since nil images are skipped
	textures, sprites := batchSystem.Stats()
	if textures != 0 || sprites != 0 {
		t.Errorf("Expected (0, 0) for nil images, got (%d, %d)", textures, sprites)
	}
}

func TestBatchRenderSystemAddToBatch(t *testing.T) {
	world := ecs.NewWorld()

	batchSystem := NewBatchRenderSystem(&world)

	// Test addToBatch directly
	pos := &components.Position{X: 100, Y: 200}
	sprite := &components.Sprite{
		Image:   nil, // nil image should be skipped
		Visible: true,
		ScaleX:  2,
		ScaleY:  2,
	}

	// This should not panic and should skip nil images
	batchSystem.addToBatch(pos, sprite, 0)

	// Test invisible sprite
	sprite.Visible = false
	batchSystem.addToBatch(pos, sprite, 0)

	// Batches should be empty
	textures, sprites := batchSystem.Stats()
	if textures != 0 || sprites != 0 {
		t.Errorf("Expected empty batches for nil/invisible sprites, got (%d, %d)", textures, sprites)
	}
}

func TestSortLayerComponent(t *testing.T) {
	world := ecs.NewWorld()

	mapper := ecs.NewMap3[components.Position, components.Sprite, components.SortLayer](&world)

	entity := mapper.NewEntity(
		&components.Position{X: 0, Y: 0},
		&components.Sprite{Visible: true, ScaleX: 1, ScaleY: 1},
		&components.SortLayer{Layer: 5},
	)

	if !world.Alive(entity) {
		t.Error("Entity should be alive")
	}

	layerMap := ecs.NewMap[components.SortLayer](&world)

	layer := layerMap.Get(entity)
	if layer.Layer != 5 {
		t.Errorf("Expected layer 5, got %d", layer.Layer)
	}
}

// BenchmarkBatchRenderSystemSetup benchmarks the batch grouping overhead without actual rendering.
func BenchmarkBatchRenderSystemSetup(b *testing.B) {
	world := ecs.NewWorld()

	batchSystem := NewBatchRenderSystem(&world)

	// Create 1000 sprites (no actual images for benchmark)
	mapper := ecs.NewMap2[components.Position, components.Sprite](&world)
	for i := range 1000 {
		mapper.NewEntity(
			&components.Position{X: float64(i % 100), Y: float64(i / 100)},
			&components.Sprite{
				Image:   nil, // Cannot create real images in headless test
				Visible: true,
				ScaleX:  1,
				ScaleY:  1,
			},
		)
	}

	for b.Loop() {
		// Simulate the query loop without Draw
		query := batchSystem.filter.Query()
		count := 0

		for query.Next() {
			query.Get()

			count++
		}
	}
}

func BenchmarkRenderSystemQuery(b *testing.B) {
	world := ecs.NewWorld()

	renderSystem := NewRenderSystem(&world)

	// Create 1000 sprites
	mapper := ecs.NewMap2[components.Position, components.Sprite](&world)
	for i := range 1000 {
		mapper.NewEntity(
			&components.Position{X: float64(i % 100), Y: float64(i / 100)},
			&components.Sprite{
				Image:   nil,
				Visible: true,
				ScaleX:  1,
				ScaleY:  1,
			},
		)
	}

	for b.Loop() {
		query := renderSystem.filter.Query()
		count := 0

		for query.Next() {
			query.Get()

			count++
		}
	}
}
