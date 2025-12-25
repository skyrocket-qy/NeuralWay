package components

import "td/internal/assets"

// Tilemap wraps a TiledMap asset for ECS entities.
// Allows tilemap-based backgrounds and levels to be rendered as entities.
type Tilemap struct {
	// Map is the loaded Tiled map data.
	Map *assets.TiledMap

	// OffsetX and OffsetY specify the world position of the tilemap's top-left corner.
	OffsetX float64
	OffsetY float64

	// LayerMask is a bitmask specifying which layers to render.
	// Bit 0 = layer 0, Bit 1 = layer 1, etc.
	// Use 0xFFFFFFFF (AllLayers) to render all layers.
	LayerMask uint32
}

// AllLayers is a LayerMask that enables all layers.
const AllLayers uint32 = 0xFFFFFFFF

// NewTilemap creates a Tilemap component with default settings.
func NewTilemap(m *assets.TiledMap) Tilemap {
	return Tilemap{
		Map:       m,
		OffsetX:   0,
		OffsetY:   0,
		LayerMask: AllLayers,
	}
}
