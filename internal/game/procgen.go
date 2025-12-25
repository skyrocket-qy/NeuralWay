package game

import (
	"math/rand"
)

// ProceduralConfig defines parameters for procedural generation.
type ProceduralConfig struct {
	Width, Height int
	Seed          int64

	// Noise settings
	NoiseScale  float64
	Octaves     int
	Persistence float64
	Lacunarity  float64
}

// DefaultProceduralConfig returns default settings.
func DefaultProceduralConfig(width, height int) ProceduralConfig {
	return ProceduralConfig{
		Width:       width,
		Height:      height,
		NoiseScale:  0.1,
		Octaves:     4,
		Persistence: 0.5,
		Lacunarity:  2.0,
	}
}

// ProceduralMap uses TileType from tdmap.go:
// TileGround, TilePath, TileWater, TileWall, TileSpawn, TileEnd

// ProceduralMap represents a generated map.
type ProceduralMap struct {
	Width, Height int
	Tiles         [][]TileType
	SpawnPoints   [][2]int // [x, y] spawn locations
	ExitPoints    [][2]int // [x, y] exit locations
}

// NewProceduralMap creates an empty map.
func NewProceduralMap(width, height int) *ProceduralMap {
	tiles := make([][]TileType, height)
	for y := range tiles {
		tiles[y] = make([]TileType, width)
	}

	return &ProceduralMap{
		Width:       width,
		Height:      height,
		Tiles:       tiles,
		SpawnPoints: make([][2]int, 0),
		ExitPoints:  make([][2]int, 0),
	}
}

// GetTile returns the tile at a position.
func (m *ProceduralMap) GetTile(x, y int) TileType {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return TileGround
	}

	return m.Tiles[y][x]
}

// SetTile sets the tile at a position.
func (m *ProceduralMap) SetTile(x, y int, tile TileType) {
	if x >= 0 && x < m.Width && y >= 0 && y < m.Height {
		m.Tiles[y][x] = tile
	}
}

// IsWalkable returns true if the tile can be walked on.
func (m *ProceduralMap) IsWalkable(x, y int) bool {
	tile := m.GetTile(x, y)

	return tile != TileWall && tile != TileWater
}

// DungeonGenerator generates dungeon-style maps.
type DungeonGenerator struct {
	rng *rand.Rand
}

// NewDungeonGenerator creates a dungeon generator.
func NewDungeonGenerator(seed int64) *DungeonGenerator {
	return &DungeonGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Room represents a dungeon room.
type Room struct {
	X, Y, Width, Height int
	Connected           bool
}

// Center returns the center of the room.
func (r *Room) Center() (int, int) {
	return r.X + r.Width/2, r.Y + r.Height/2
}

// Intersects checks if two rooms overlap.
func (r *Room) Intersects(other *Room, padding int) bool {
	return r.X-padding < other.X+other.Width &&
		r.X+r.Width+padding > other.X &&
		r.Y-padding < other.Y+other.Height &&
		r.Y+r.Height+padding > other.Y
}

// GenerateDungeon creates a dungeon with rooms and corridors.
func (g *DungeonGenerator) GenerateDungeon(
	width, height, roomCount, minRoomSize, maxRoomSize int,
) *ProceduralMap {
	m := NewProceduralMap(width, height)

	// Fill with walls
	for y := range height {
		for x := range width {
			m.Tiles[y][x] = TileWall
		}
	}

	rooms := make([]Room, 0, roomCount)

	// Generate rooms
	attempts := 0
	for len(rooms) < roomCount && attempts < roomCount*10 {
		attempts++

		w := minRoomSize + g.rng.Intn(maxRoomSize-minRoomSize+1)
		h := minRoomSize + g.rng.Intn(maxRoomSize-minRoomSize+1)
		x := 1 + g.rng.Intn(width-w-2)
		y := 1 + g.rng.Intn(height-h-2)

		room := Room{X: x, Y: y, Width: w, Height: h}

		// Check for overlap
		overlaps := false

		for _, other := range rooms {
			if room.Intersects(&other, 1) {
				overlaps = true

				break
			}
		}

		if !overlaps {
			rooms = append(rooms, room)
			// Carve room
			for ry := y; ry < y+h; ry++ {
				for rx := x; rx < x+w; rx++ {
					m.Tiles[ry][rx] = TilePath
				}
			}
		}
	}

	// Connect rooms with corridors
	for i := 1; i < len(rooms); i++ {
		g.createCorridor(m, &rooms[i-1], &rooms[i])
	}

	// Set spawn and exit
	if len(rooms) > 0 {
		sx, sy := rooms[0].Center()
		m.SpawnPoints = append(m.SpawnPoints, [2]int{sx, sy})

		ex, ey := rooms[len(rooms)-1].Center()
		m.ExitPoints = append(m.ExitPoints, [2]int{ex, ey})
	}

	return m
}

// createCorridor carves a corridor between two rooms.
func (g *DungeonGenerator) createCorridor(m *ProceduralMap, from, to *Room) {
	x1, y1 := from.Center()
	x2, y2 := to.Center()

	// Randomly decide whether to go horizontal or vertical first
	if g.rng.Float32() < 0.5 {
		g.carveHorizontal(m, x1, x2, y1)
		g.carveVertical(m, y1, y2, x2)
	} else {
		g.carveVertical(m, y1, y2, x1)
		g.carveHorizontal(m, x1, x2, y2)
	}
}

func (g *DungeonGenerator) carveHorizontal(m *ProceduralMap, x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	for x := x1; x <= x2; x++ {
		m.Tiles[y][x] = TilePath
	}
}

func (g *DungeonGenerator) carveVertical(m *ProceduralMap, y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	for y := y1; y <= y2; y++ {
		m.Tiles[y][x] = TilePath
	}
}

// CaveGenerator generates cave-style maps using cellular automata.
type CaveGenerator struct {
	rng *rand.Rand
}

// NewCaveGenerator creates a cave generator.
func NewCaveGenerator(seed int64) *CaveGenerator {
	return &CaveGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// GenerateCave creates a cave using cellular automata.
func (g *CaveGenerator) GenerateCave(
	width, height int,
	fillProbability float64,
	iterations int,
) *ProceduralMap {
	m := NewProceduralMap(width, height)

	// Random fill
	for y := range height {
		for x := range width {
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				m.Tiles[y][x] = TileWall
			} else if g.rng.Float64() < fillProbability {
				m.Tiles[y][x] = TileWall
			} else {
				m.Tiles[y][x] = TilePath
			}
		}
	}

	// Cellular automata iterations
	for range iterations {
		g.caStep(m)
	}

	// Find spawn point (first open floor)
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			if m.Tiles[y][x] == TilePath {
				m.SpawnPoints = append(m.SpawnPoints, [2]int{x, y})

				break
			}
		}

		if len(m.SpawnPoints) > 0 {
			break
		}
	}

	return m
}

// caStep performs one cellular automata step.
func (g *CaveGenerator) caStep(m *ProceduralMap) {
	newTiles := make([][]TileType, m.Height)
	for y := range newTiles {
		newTiles[y] = make([]TileType, m.Width)
		copy(newTiles[y], m.Tiles[y])
	}

	for y := 1; y < m.Height-1; y++ {
		for x := 1; x < m.Width-1; x++ {
			walls := g.countWallNeighbors(m, x, y)

			if walls > 4 {
				newTiles[y][x] = TileWall
			} else if walls < 4 {
				newTiles[y][x] = TilePath
			}
		}
	}

	m.Tiles = newTiles
}

// countWallNeighbors counts wall tiles in 3x3 area.
func (g *CaveGenerator) countWallNeighbors(m *ProceduralMap, x, y int) int {
	count := 0

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if m.Tiles[y+dy][x+dx] == TileWall {
				count++
			}
		}
	}

	return count
}

// ArenaGenerator generates arena-style maps.
type ArenaGenerator struct {
	rng *rand.Rand
}

// NewArenaGenerator creates an arena generator.
func NewArenaGenerator(seed int64) *ArenaGenerator {
	return &ArenaGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// GenerateArena creates an open arena with optional obstacles.
func (g *ArenaGenerator) GenerateArena(width, height, obstacleCount int) *ProceduralMap {
	m := NewProceduralMap(width, height)

	// Fill with floor
	for y := range height {
		for x := range width {
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				m.Tiles[y][x] = TileWall
			} else {
				m.Tiles[y][x] = TilePath
			}
		}
	}

	// Add random obstacles
	for range obstacleCount {
		ox := 2 + g.rng.Intn(width-4)
		oy := 2 + g.rng.Intn(height-4)
		ow := 1 + g.rng.Intn(3)
		oh := 1 + g.rng.Intn(3)

		for dy := range oh {
			for dx := range ow {
				if oy+dy < height-1 && ox+dx < width-1 {
					m.Tiles[oy+dy][ox+dx] = TileWall
				}
			}
		}
	}

	// Spawn in center
	m.SpawnPoints = append(m.SpawnPoints, [2]int{width / 2, height / 2})

	return m
}
