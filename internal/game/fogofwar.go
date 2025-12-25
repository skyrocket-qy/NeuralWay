package game

import (
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// FogState represents visibility state of a cell.
type FogState int

const (
	FogUnexplored FogState = iota // Never seen
	FogExplored                   // Seen before, currently hidden
	FogVisible                    // Currently visible
)

// FogOfWar manages visibility on a tile-based map.
type FogOfWar struct {
	Width, Height int
	CellSize      float64
	Grid          [][]FogState

	// Vision sources
	visionSources []VisionSource

	// Settings
	RememberExplored bool // If true, explored areas stay revealed
	DefaultState     FogState
}

// VisionSource represents something that reveals fog.
type VisionSource struct {
	X, Y   float64
	Range  float64 // Vision radius
	Entity ecs.Entity
	Active bool
}

// NewFogOfWar creates a fog of war system.
func NewFogOfWar(width, height int, cellSize float64) *FogOfWar {
	grid := make([][]FogState, height)
	for y := range grid {
		grid[y] = make([]FogState, width)
	}

	return &FogOfWar{
		Width:            width,
		Height:           height,
		CellSize:         cellSize,
		Grid:             grid,
		RememberExplored: true,
		DefaultState:     FogUnexplored,
		visionSources:    make([]VisionSource, 0),
	}
}

// AddVisionSource adds a vision source.
func (f *FogOfWar) AddVisionSource(x, y, visionRange float64, entity ecs.Entity) int {
	source := VisionSource{
		X:      x,
		Y:      y,
		Range:  visionRange,
		Entity: entity,
		Active: true,
	}
	f.visionSources = append(f.visionSources, source)

	return len(f.visionSources) - 1
}

// UpdateVisionSource updates a vision source position.
func (f *FogOfWar) UpdateVisionSource(index int, x, y float64) {
	if index >= 0 && index < len(f.visionSources) {
		f.visionSources[index].X = x
		f.visionSources[index].Y = y
	}
}

// RemoveVisionSource removes a vision source.
func (f *FogOfWar) RemoveVisionSource(index int) {
	if index >= 0 && index < len(f.visionSources) {
		f.visionSources[index].Active = false
	}
}

// Update recalculates fog based on vision sources.
func (f *FogOfWar) Update() {
	// First, hide all currently visible cells
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			if f.Grid[y][x] == FogVisible {
				if f.RememberExplored {
					f.Grid[y][x] = FogExplored
				} else {
					f.Grid[y][x] = FogUnexplored
				}
			}
		}
	}

	// Reveal cells around each vision source
	for _, source := range f.visionSources {
		if !source.Active {
			continue
		}

		f.revealAround(source.X, source.Y, source.Range)
	}
}

// revealAround reveals cells within range of a point.
func (f *FogOfWar) revealAround(wx, wy, radius float64) {
	// Convert world coords to grid coords
	centerX := int(wx / f.CellSize)
	centerY := int(wy / f.CellSize)
	cellRadius := int(radius/f.CellSize) + 1

	for dy := -cellRadius; dy <= cellRadius; dy++ {
		for dx := -cellRadius; dx <= cellRadius; dx++ {
			gx := centerX + dx
			gy := centerY + dy

			if gx < 0 || gx >= f.Width || gy < 0 || gy >= f.Height {
				continue
			}

			// Check if within circular range
			cellCenterX := (float64(gx) + 0.5) * f.CellSize
			cellCenterY := (float64(gy) + 0.5) * f.CellSize
			distSq := (cellCenterX-wx)*(cellCenterX-wx) + (cellCenterY-wy)*(cellCenterY-wy)

			if distSq <= radius*radius {
				f.Grid[gy][gx] = FogVisible
			}
		}
	}
}

// IsVisible returns true if a world position is visible.
func (f *FogOfWar) IsVisible(wx, wy float64) bool {
	gx := int(wx / f.CellSize)
	gy := int(wy / f.CellSize)

	if gx < 0 || gx >= f.Width || gy < 0 || gy >= f.Height {
		return false
	}

	return f.Grid[gy][gx] == FogVisible
}

// IsExplored returns true if a world position has been explored.
func (f *FogOfWar) IsExplored(wx, wy float64) bool {
	gx := int(wx / f.CellSize)
	gy := int(wy / f.CellSize)

	if gx < 0 || gx >= f.Width || gy < 0 || gy >= f.Height {
		return false
	}

	state := f.Grid[gy][gx]

	return state == FogVisible || state == FogExplored
}

// GetState returns the fog state at a grid position.
func (f *FogOfWar) GetState(gx, gy int) FogState {
	if gx < 0 || gx >= f.Width || gy < 0 || gy >= f.Height {
		return FogUnexplored
	}

	return f.Grid[gy][gx]
}

// GetAlpha returns the fog alpha (0=visible, 0.5=explored, 1=hidden).
func (f *FogOfWar) GetAlpha(gx, gy int) float64 {
	state := f.GetState(gx, gy)
	switch state {
	case FogVisible:
		return 0.0
	case FogExplored:
		return 0.5
	default:
		return 1.0
	}
}

// RevealAll reveals the entire map.
func (f *FogOfWar) RevealAll() {
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			f.Grid[y][x] = FogVisible
		}
	}
}

// Reset resets all fog to unexplored.
func (f *FogOfWar) Reset() {
	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			f.Grid[y][x] = f.DefaultState
		}
	}
}

// UpdateFromEntities updates vision sources from entity positions.
func (f *FogOfWar) UpdateFromEntities(world *ecs.World, posFilter *ecs.Filter1[components.Position]) {
	// This would typically be called per-frame to update vision source positions
	// based on entity movement
	for i := range f.visionSources {
		if !f.visionSources[i].Active {
			continue
		}

		entity := f.visionSources[i].Entity

		query := posFilter.Query()
		for query.Next() {
			if query.Entity() == entity {
				pos := query.Get()
				f.visionSources[i].X = pos.X
				f.visionSources[i].Y = pos.Y

				break
			}
		}
	}

	f.Update()
}

// GetExploredPercent returns the percentage of map explored.
func (f *FogOfWar) GetExploredPercent() float64 {
	explored := 0
	total := f.Width * f.Height

	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			if f.Grid[y][x] != FogUnexplored {
				explored++
			}
		}
	}

	return float64(explored) / float64(total) * 100.0
}
