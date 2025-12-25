package systems

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/assets"
	"td/internal/components"
)

func TestTilemapRenderSystemSetCamera(t *testing.T) {
	world := ecs.NewWorld()
	system := NewTilemapRenderSystem(&world)

	system.SetCamera(100, 200, 640, 480)

	cameraX, cameraY, viewW, viewH := system.Stats()
	if cameraX != 100 || cameraY != 200 {
		t.Errorf("Camera position mismatch: got (%.1f, %.1f), want (100, 200)", cameraX, cameraY)
	}

	if viewW != 640 || viewH != 480 {
		t.Errorf("Viewport size mismatch: got (%d, %d), want (640, 480)", viewW, viewH)
	}
}

func TestTilemapComponent(t *testing.T) {
	// Create a mock TiledMap
	mockMap := &assets.TiledMap{
		Width:      10,
		Height:     10,
		TileWidth:  32,
		TileHeight: 32,
		Layers: []assets.TiledLayer{
			{Name: "ground", Type: "tilelayer", Width: 10, Height: 10, Visible: true},
			{Name: "objects", Type: "tilelayer", Width: 10, Height: 10, Visible: true},
		},
	}

	tilemap := components.NewTilemap(mockMap)

	if tilemap.Map != mockMap {
		t.Error("Map reference mismatch")
	}

	if tilemap.OffsetX != 0 || tilemap.OffsetY != 0 {
		t.Errorf("Default offset should be (0, 0), got (%.1f, %.1f)", tilemap.OffsetX, tilemap.OffsetY)
	}

	if tilemap.LayerMask != components.AllLayers {
		t.Errorf(
			"Default layer mask should be AllLayers (0x%x), got 0x%x",
			components.AllLayers,
			tilemap.LayerMask,
		)
	}
}

func TestTilemapComponentLayerMask(t *testing.T) {
	mockMap := &assets.TiledMap{
		Width:      5,
		Height:     5,
		TileWidth:  16,
		TileHeight: 16,
	}

	tilemap := components.Tilemap{
		Map:       mockMap,
		LayerMask: 0b0101, // Only layers 0 and 2
	}

	// Check layer 0 is enabled
	if tilemap.LayerMask&(1<<0) == 0 {
		t.Error("Layer 0 should be enabled")
	}
	// Check layer 1 is disabled
	if tilemap.LayerMask&(1<<1) != 0 {
		t.Error("Layer 1 should be disabled")
	}
	// Check layer 2 is enabled
	if tilemap.LayerMask&(1<<2) == 0 {
		t.Error("Layer 2 should be enabled")
	}
}

func TestTilemapRenderSystemWithEntity(t *testing.T) {
	world := ecs.NewWorld()
	system := NewTilemapRenderSystem(&world)

	// Create a minimal tilemap entity
	mockMap := &assets.TiledMap{
		Width:      20,
		Height:     15,
		TileWidth:  32,
		TileHeight: 32,
		Layers: []assets.TiledLayer{
			{
				Name:    "ground",
				Type:    "tilelayer",
				Width:   20,
				Height:  15,
				Visible: true,
				Data:    make([]int, 20*15),
			},
		},
		TileImages: make(map[int]*ebiten.Image),
	}

	mapper := ecs.NewMap1[components.Tilemap](&world)
	entity := mapper.NewEntity(&components.Tilemap{
		Map:       mockMap,
		OffsetX:   0,
		OffsetY:   0,
		LayerMask: components.AllLayers,
	})

	if !world.Alive(entity) {
		t.Error("Tilemap entity should be alive")
	}

	// Verify system can query the entity
	query := system.filter.Query()

	count := 0
	for query.Next() {
		count++
	}

	if count != 1 {
		t.Errorf("Expected 1 tilemap entity, got %d", count)
	}
}
