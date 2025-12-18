package game

import (
	"image/color"
	"td/internal/components"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
)

// TDMap represents the tower defense map.
type TDMap struct {
	Width      int
	Height     int
	TileSize   int
	PathGrid   *PathGrid
	SpawnPoint Point
	EndPoint   Point
	Tiles      [][]TileType
	Path       []Point // Cached path from spawn to end
}

// TileType defines the type of map tile.
type TileType int

const (
	TileGround TileType = iota
	TilePath
	TileWater
	TileWall
	TileSpawn
	TileEnd
)

// TileColors for rendering different tile types.
var TileColors = map[TileType]color.RGBA{
	TileGround: {R: 34, G: 139, B: 34, A: 255},   // Forest green
	TilePath:   {R: 194, G: 178, B: 128, A: 255}, // Sand
	TileWater:  {R: 65, G: 105, B: 225, A: 255},  // Royal blue
	TileWall:   {R: 105, G: 105, B: 105, A: 255}, // Dim gray
	TileSpawn:  {R: 255, G: 69, B: 0, A: 255},    // Orange red
	TileEnd:    {R: 50, G: 205, B: 50, A: 255},   // Lime green
}

// NewTDMap creates a new tower defense map.
func NewTDMap(width, height, tileSize int) *TDMap {
	m := &TDMap{
		Width:    width,
		Height:   height,
		TileSize: tileSize,
		PathGrid: NewPathGrid(width, height),
		Tiles:    make([][]TileType, height),
	}

	for y := 0; y < height; y++ {
		m.Tiles[y] = make([]TileType, width)
		for x := 0; x < width; x++ {
			m.Tiles[y][x] = TileGround
		}
	}

	return m
}

// SetTile sets a tile type and updates walkability.
func (m *TDMap) SetTile(x, y int, tileType TileType) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}

	m.Tiles[y][x] = tileType

	// Update walkability
	switch tileType {
	case TileGround, TilePath, TileSpawn, TileEnd:
		m.PathGrid.SetWalkable(x, y, true)
	case TileWater, TileWall:
		m.PathGrid.SetWalkable(x, y, false)
	}

	// Update special points
	if tileType == TileSpawn {
		m.SpawnPoint = Point{X: x, Y: y}
	}
	if tileType == TileEnd {
		m.EndPoint = Point{X: x, Y: y}
	}
}

// CreatePath creates a walkable path from spawn to end.
func (m *TDMap) CreatePath(points []Point) {
	for _, p := range points {
		m.SetTile(p.X, p.Y, TilePath)
	}
}

// CalculatePath calculates and caches the path from spawn to end.
func (m *TDMap) CalculatePath() {
	m.Path = m.PathGrid.FindPath(
		m.SpawnPoint.X, m.SpawnPoint.Y,
		m.EndPoint.X, m.EndPoint.Y,
	)
}

// WorldToTile converts world coordinates to tile coordinates.
func (m *TDMap) WorldToTile(wx, wy float64) (int, int) {
	return int(wx) / m.TileSize, int(wy) / m.TileSize
}

// TileToWorld converts tile coordinates to world coordinates (center of tile).
func (m *TDMap) TileToWorld(tx, ty int) (float64, float64) {
	return float64(tx*m.TileSize) + float64(m.TileSize)/2,
		float64(ty*m.TileSize) + float64(m.TileSize)/2
}

// Draw renders the map to the screen.
func (m *TDMap) Draw(screen *ebiten.Image) {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			tileType := m.Tiles[y][x]
			col := TileColors[tileType]

			tile := ebiten.NewImage(m.TileSize, m.TileSize)
			tile.Fill(col)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*m.TileSize), float64(y*m.TileSize))
			screen.DrawImage(tile, op)
		}
	}
}

// CreateDefaultMap creates a simple test map.
func CreateDefaultMap() *TDMap {
	m := NewTDMap(25, 15, 32)

	// Set spawn and end
	m.SetTile(0, 7, TileSpawn)
	m.SetTile(24, 7, TileEnd)

	// Create a winding path
	path := []Point{
		{1, 7}, {2, 7}, {3, 7}, {4, 7}, {5, 7},
		{5, 6}, {5, 5}, {5, 4}, {5, 3},
		{6, 3}, {7, 3}, {8, 3}, {9, 3}, {10, 3},
		{10, 4}, {10, 5}, {10, 6}, {10, 7}, {10, 8}, {10, 9}, {10, 10}, {10, 11},
		{11, 11}, {12, 11}, {13, 11}, {14, 11}, {15, 11},
		{15, 10}, {15, 9}, {15, 8}, {15, 7},
		{16, 7}, {17, 7}, {18, 7}, {19, 7},
		{19, 6}, {19, 5}, {19, 4}, {19, 3},
		{20, 3}, {21, 3}, {22, 3}, {23, 3}, {23, 4}, {23, 5}, {23, 6}, {23, 7},
	}
	m.CreatePath(path)

	// Add some obstacles
	for x := 7; x <= 9; x++ {
		for y := 5; y <= 9; y++ {
			m.SetTile(x, y, TileWall)
		}
	}

	for x := 12; x <= 14; x++ {
		for y := 5; y <= 9; y++ {
			m.SetTile(x, y, TileWater)
		}
	}

	m.CalculatePath()
	return m
}

// Hero represents a hero unit that can attack monsters.
type Hero struct {
	Name         string
	AttackDamage int
	AttackRange  float64
	AttackSpeed  float64 // Attacks per second
	AttackTimer  float64
	Level        int
	Experience   int
	ExpToLevel   int
}

// NewHero creates a new hero with default stats.
func NewHero(name string) *Hero {
	return &Hero{
		Name:         name,
		AttackDamage: 10,
		AttackRange:  100,
		AttackSpeed:  1.0,
		AttackTimer:  0,
		Level:        1,
		Experience:   0,
		ExpToLevel:   100,
	}
}

// GainExp adds experience and handles leveling.
func (h *Hero) GainExp(amount int) bool {
	h.Experience += amount
	if h.Experience >= h.ExpToLevel {
		h.Experience -= h.ExpToLevel
		h.Level++
		h.ExpToLevel = h.ExpToLevel * 3 / 2 // 50% more exp needed each level
		h.AttackDamage += 5
		h.AttackRange += 10
		return true // Leveled up
	}
	return false
}

// CanAttack returns true if the hero can attack this frame.
func (h *Hero) CanAttack(dt float64) bool {
	h.AttackTimer += dt
	interval := 1.0 / h.AttackSpeed
	if h.AttackTimer >= interval {
		h.AttackTimer -= interval
		return true
	}
	return false
}

// CreateHeroEntity creates an ECS entity for a hero.
func CreateHeroEntity(world *ecs.World, hero *Hero, x, y float64) ecs.Entity {
	img := ebiten.NewImage(24, 24)
	img.Fill(color.RGBA{R: 0, G: 100, B: 255, A: 255}) // Blue hero

	mapper := ecs.NewMap3[components.Position, components.Sprite, components.Collider](world)
	return mapper.NewEntity(
		&components.Position{X: x, Y: y},
		&components.Sprite{
			Image:   img,
			ScaleX:  1,
			ScaleY:  1,
			Visible: true,
		},
		&components.Collider{
			Width:  24,
			Height: 24,
			Layer:  1, // Hero layer
			Mask:   2, // Collides with monsters
		},
	)
}
