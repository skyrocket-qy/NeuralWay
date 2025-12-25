package systems

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// SpatialHash provides fast spatial queries using a grid-based hash.
type SpatialHash struct {
	cellSize int
	cells    map[int64][]ecs.Entity
}

// NewSpatialHash creates a spatial hash with the given cell size.
func NewSpatialHash(cellSize int) *SpatialHash {
	return &SpatialHash{
		cellSize: cellSize,
		cells:    make(map[int64][]ecs.Entity),
	}
}

// hash returns a unique key for a cell position.
func (s *SpatialHash) hash(x, y int) int64 {
	return int64(x)<<32 | int64(uint32(y))
}

// cellCoords returns the cell coordinates for a world position.
func (s *SpatialHash) cellCoords(x, y float64) (int, int) {
	return int(x) / s.cellSize, int(y) / s.cellSize
}

// Clear removes all entities from the spatial hash.
func (s *SpatialHash) Clear() {
	for k := range s.cells {
		delete(s.cells, k)
	}
}

// Insert adds an entity at the given position.
func (s *SpatialHash) Insert(entity ecs.Entity, x, y, width, height float64) {
	// Calculate cell range covered by the AABB
	minCX, minCY := s.cellCoords(x, y)
	maxCX, maxCY := s.cellCoords(x+width, y+height)

	for cx := minCX; cx <= maxCX; cx++ {
		for cy := minCY; cy <= maxCY; cy++ {
			key := s.hash(cx, cy)
			s.cells[key] = append(s.cells[key], entity)
		}
	}
}

// Query returns all entities that might intersect the given AABB.
func (s *SpatialHash) Query(x, y, width, height float64) []ecs.Entity {
	minCX, minCY := s.cellCoords(x, y)
	maxCX, maxCY := s.cellCoords(x+width, y+height)

	// Use a map to deduplicate entities spanning multiple cells
	seen := make(map[ecs.Entity]bool)

	var result []ecs.Entity

	for cx := minCX; cx <= maxCX; cx++ {
		for cy := minCY; cy <= maxCY; cy++ {
			key := s.hash(cx, cy)
			for _, e := range s.cells[key] {
				if !seen[e] {
					seen[e] = true
					result = append(result, e)
				}
			}
		}
	}

	return result
}

// CollisionSystem detects collisions between entities.
type CollisionSystem struct {
	posFilter      *ecs.Filter2[components.Position, components.Collider]
	spatialHash    *SpatialHash
	onCollision    func(a, b ecs.Entity)
	collisionPairs []CollisionPair
}

// CollisionPair represents two colliding entities.
type CollisionPair struct {
	A, B ecs.Entity
}

// NewCollisionSystem creates a collision detection system.
func NewCollisionSystem(world *ecs.World, cellSize int) *CollisionSystem {
	return &CollisionSystem{
		posFilter:   ecs.NewFilter2[components.Position, components.Collider](world),
		spatialHash: NewSpatialHash(cellSize),
	}
}

// SetCallback sets the collision callback function.
func (s *CollisionSystem) SetCallback(fn func(a, b ecs.Entity)) {
	s.onCollision = fn
}

// GetCollisions returns all collision pairs from the last update.
func (s *CollisionSystem) GetCollisions() []CollisionPair {
	return s.collisionPairs
}

// Update detects collisions and calls the callback for each pair.
func (s *CollisionSystem) Update(world *ecs.World) {
	// Clear spatial hash
	s.spatialHash.Clear()
	s.collisionPairs = s.collisionPairs[:0]

	// Insert all collidable entities
	query := s.posFilter.Query()

	type entityData struct {
		entity   ecs.Entity
		pos      *components.Position
		collider *components.Collider
	}

	var entities []entityData

	for query.Next() {
		pos, col := query.Get()
		entity := query.Entity()
		entities = append(entities, entityData{entity, pos, col})
		s.spatialHash.Insert(entity, pos.X, pos.Y, col.Width, col.Height)
	}

	// Check collisions
	checked := make(map[string]bool)

	for _, e := range entities {
		candidates := s.spatialHash.Query(e.pos.X, e.pos.Y, e.collider.Width, e.collider.Height)

		for _, other := range candidates {
			if e.entity == other {
				continue
			}

			// Create unique pair key using entity comparison
			// Use fmt.Sprintf for deduplication
			var pairKey string

			eID := fmt.Sprintf("%v", e.entity)

			oID := fmt.Sprintf("%v", other)
			if eID < oID {
				pairKey = eID + "-" + oID
			} else {
				pairKey = oID + "-" + eID
			}

			if checked[pairKey] {
				continue
			}

			checked[pairKey] = true

			// Find other entity's data
			for _, o := range entities {
				if o.entity == other {
					// Check layer masks
					if (e.collider.Layer&o.collider.Mask) == 0 && (o.collider.Layer&e.collider.Mask) == 0 {
						continue
					}

					// AABB collision test
					if aabbCollision(
						e.pos.X, e.pos.Y, e.collider.Width, e.collider.Height,
						o.pos.X, o.pos.Y, o.collider.Width, o.collider.Height,
					) {
						s.collisionPairs = append(s.collisionPairs, CollisionPair{e.entity, o.entity})
						if s.onCollision != nil {
							s.onCollision(e.entity, o.entity)
						}
					}

					break
				}
			}
		}
	}
}

// aabbCollision tests if two axis-aligned bounding boxes overlap.
func aabbCollision(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// PointInAABB tests if a point is inside an AABB.
func PointInAABB(px, py, x, y, w, h float64) bool {
	return px >= x && px <= x+w && py >= y && py <= y+h
}

// CircleCollision tests if two circles overlap.
func CircleCollision(x1, y1, r1, x2, y2, r2 float64) bool {
	dx := x2 - x1
	dy := y2 - y1
	dist := dx*dx + dy*dy
	radii := r1 + r2

	return dist < radii*radii
}
