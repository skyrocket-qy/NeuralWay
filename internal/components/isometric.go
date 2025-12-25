package components

// IsometricPosition represents a position in isometric space.
type IsometricPosition struct {
	IsoX float64 // Isometric X (horizontal)
	IsoY float64 // Isometric Y (depth into screen)
	IsoZ float64 // Height above ground
}

// ToScreen converts isometric coordinates to screen coordinates.
// Uses 2:1 isometric projection (30 degree angle).
func (p IsometricPosition) ToScreen(tileWidth, tileHeight float64) (screenX, screenY float64) {
	// Standard 2:1 isometric projection
	screenX = (p.IsoX - p.IsoY) * (tileWidth / 2)
	screenY = (p.IsoX+p.IsoY)*(tileHeight/2) - p.IsoZ

	return screenX, screenY
}

// FromScreen converts screen coordinates to isometric coordinates.
func FromScreen(screenX, screenY, tileWidth, tileHeight float64) IsometricPosition {
	isoX := (screenX/(tileWidth/2) + screenY/(tileHeight/2)) / 2
	isoY := (screenY/(tileHeight/2) - screenX/(tileWidth/2)) / 2

	return IsometricPosition{IsoX: isoX, IsoY: isoY, IsoZ: 0}
}

// IsometricTile represents a tile in an isometric tilemap.
type IsometricTile struct {
	TileID   int
	Height   int // Stack height (for multi-level)
	Walkable bool
}

// IsometricMap represents an isometric tilemap.
type IsometricMap struct {
	Width      int
	Height     int
	TileWidth  float64
	TileHeight float64
	Tiles      [][]IsometricTile
	Layout     IsoLayout
}

// IsoLayout defines the isometric map layout type.
type IsoLayout int

const (
	// LayoutDiamond is standard diamond isometric layout.
	LayoutDiamond IsoLayout = iota
	// LayoutStaggered is staggered isometric layout (offset rows).
	LayoutStaggered
)

// NewIsometricMap creates a new isometric map.
func NewIsometricMap(width, height int, tileWidth, tileHeight float64) *IsometricMap {
	tiles := make([][]IsometricTile, height)
	for y := range tiles {
		tiles[y] = make([]IsometricTile, width)
		for x := range tiles[y] {
			tiles[y][x] = IsometricTile{Walkable: true}
		}
	}

	return &IsometricMap{
		Width:      width,
		Height:     height,
		TileWidth:  tileWidth,
		TileHeight: tileHeight,
		Tiles:      tiles,
		Layout:     LayoutDiamond,
	}
}

// GetTile returns the tile at the given isometric coordinates.
func (m *IsometricMap) GetTile(x, y int) *IsometricTile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return nil
	}

	return &m.Tiles[y][x]
}

// SetTile sets a tile at the given isometric coordinates.
func (m *IsometricMap) SetTile(x, y int, tile IsometricTile) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}

	m.Tiles[y][x] = tile
}

// TileToScreen converts tile coordinates to screen position.
func (m *IsometricMap) TileToScreen(tileX, tileY int) (screenX, screenY float64) {
	pos := IsometricPosition{IsoX: float64(tileX), IsoY: float64(tileY)}

	return pos.ToScreen(m.TileWidth, m.TileHeight)
}

// ScreenToTile converts screen position to tile coordinates.
func (m *IsometricMap) ScreenToTile(screenX, screenY float64) (tileX, tileY int) {
	pos := FromScreen(screenX, screenY, m.TileWidth, m.TileHeight)

	return int(pos.IsoX + 0.5), int(pos.IsoY + 0.5)
}
