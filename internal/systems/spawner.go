package systems

import (
	"math/rand"

	"github.com/mlange-42/ark/ecs"
)

// SpawnPoint defines where and what to spawn.
type SpawnPoint struct {
	ID        string
	X, Y      float64
	SpawnType string  // Type of entity to spawn
	Interval  float64 // Time between spawns (0 = one-shot)
	Timer     float64 // Current timer
	MaxSpawns int     // Maximum total spawns (-1 = infinite)
	Spawned   int     // Total spawned
	Active    bool
	Radius    float64        // Random spawn radius
	Data      map[string]any // Extra spawn data
}

// NewSpawnPoint creates a spawn point.
func NewSpawnPoint(id, spawnType string, x, y float64) SpawnPoint {
	return SpawnPoint{
		ID:        id,
		X:         x,
		Y:         y,
		SpawnType: spawnType,
		MaxSpawns: -1,
		Active:    true,
		Data:      make(map[string]any),
	}
}

// SpawnEvent represents an entity spawn.
type SpawnEvent struct {
	SpawnPoint string
	SpawnType  string
	X, Y       float64
	Entity     ecs.Entity // Set after entity is created
	Data       map[string]any
}

// EntityFactory is a function that creates an entity.
type EntityFactory func(world *ecs.World, event SpawnEvent) ecs.Entity

// SpawnerSystem manages entity spawning.
type SpawnerSystem struct {
	spawnPoints map[string]*SpawnPoint
	spawnQueue  []SpawnEvent
	factories   map[string]EntityFactory
	onSpawn     func(SpawnEvent)
}

// NewSpawnerSystem creates a spawner system.
func NewSpawnerSystem() *SpawnerSystem {
	return &SpawnerSystem{
		spawnPoints: make(map[string]*SpawnPoint),
		spawnQueue:  make([]SpawnEvent, 0),
		factories:   make(map[string]EntityFactory),
	}
}

// SetOnSpawn sets the spawn callback.
func (s *SpawnerSystem) SetOnSpawn(fn func(SpawnEvent)) {
	s.onSpawn = fn
}

// RegisterFactory registers an entity factory for a spawn type.
func (s *SpawnerSystem) RegisterFactory(spawnType string, factory EntityFactory) {
	s.factories[spawnType] = factory
}

// AddSpawnPoint adds a spawn point.
func (s *SpawnerSystem) AddSpawnPoint(sp SpawnPoint) {
	s.spawnPoints[sp.ID] = &sp
}

// RemoveSpawnPoint removes a spawn point.
func (s *SpawnerSystem) RemoveSpawnPoint(id string) {
	delete(s.spawnPoints, id)
}

// SetActive enables or disables a spawn point.
func (s *SpawnerSystem) SetActive(id string, active bool) {
	if sp, ok := s.spawnPoints[id]; ok {
		sp.Active = active
	}
}

// SpawnNow immediately spawns from a spawn point.
func (s *SpawnerSystem) SpawnNow(id string) *SpawnEvent {
	sp, ok := s.spawnPoints[id]
	if !ok || !sp.Active {
		return nil
	}

	if sp.MaxSpawns >= 0 && sp.Spawned >= sp.MaxSpawns {
		return nil
	}

	// Calculate spawn position with radius
	x, y := sp.X, sp.Y
	if sp.Radius > 0 {
		angle := rand.Float64() * 2 * 3.14159
		dist := rand.Float64() * sp.Radius
		x += dist * cosApprox(angle)
		y += dist * sinApprox(angle)
	}

	event := SpawnEvent{
		SpawnPoint: sp.ID,
		SpawnType:  sp.SpawnType,
		X:          x,
		Y:          y,
		Data:       sp.Data,
	}

	s.spawnQueue = append(s.spawnQueue, event)
	sp.Spawned++
	sp.Timer = sp.Interval

	return &event
}

// Update updates timers and processes spawns.
func (s *SpawnerSystem) Update(world *ecs.World, dt float64) {
	// Update spawn point timers
	for _, sp := range s.spawnPoints {
		if !sp.Active {
			continue
		}

		if sp.MaxSpawns >= 0 && sp.Spawned >= sp.MaxSpawns {
			continue
		}

		if sp.Interval <= 0 {
			continue // One-shot, needs manual trigger
		}

		sp.Timer -= dt
		if sp.Timer <= 0 {
			s.SpawnNow(sp.ID)
		}
	}

	// Process spawn queue
	for i := range s.spawnQueue {
		event := &s.spawnQueue[i]
		if factory, ok := s.factories[event.SpawnType]; ok {
			event.Entity = factory(world, *event)
		}

		if s.onSpawn != nil {
			s.onSpawn(*event)
		}
	}

	s.spawnQueue = s.spawnQueue[:0]
}

// GetSpawnQueue returns pending spawns (for custom handling).
func (s *SpawnerSystem) GetSpawnQueue() []SpawnEvent {
	return s.spawnQueue
}

// ClearSpawnQueue clears pending spawns.
func (s *SpawnerSystem) ClearSpawnQueue() {
	s.spawnQueue = s.spawnQueue[:0]
}

// ResetSpawnPoint resets a spawn point's counter.
func (s *SpawnerSystem) ResetSpawnPoint(id string) {
	if sp, ok := s.spawnPoints[id]; ok {
		sp.Spawned = 0
		sp.Timer = 0
	}
}

// GetSpawnCount returns how many entities a spawn point has created.
func (s *SpawnerSystem) GetSpawnCount(id string) int {
	if sp, ok := s.spawnPoints[id]; ok {
		return sp.Spawned
	}

	return 0
}

// Simple cos approximation to avoid importing math.
func cosApprox(x float64) float64 {
	// Taylor series approximation
	x2 := x * x

	return 1 - x2/2 + x2*x2/24
}

// Simple sin approximation.
func sinApprox(x float64) float64 {
	x2 := x * x

	return x - x*x2/6 + x*x2*x2/120
}

// SpawnCircle spawns entities in a circle pattern.
func (s *SpawnerSystem) SpawnCircle(
	world *ecs.World,
	spawnType string,
	cx, cy, radius float64,
	count int,
) []SpawnEvent {
	events := make([]SpawnEvent, 0, count)
	for i := range count {
		angle := float64(i) * (2 * 3.14159 / float64(count))
		x := cx + radius*cosApprox(angle)
		y := cy + radius*sinApprox(angle)

		event := SpawnEvent{
			SpawnType: spawnType,
			X:         x,
			Y:         y,
			Data:      make(map[string]any),
		}

		if factory, ok := s.factories[spawnType]; ok {
			event.Entity = factory(world, event)
		}

		events = append(events, event)
		if s.onSpawn != nil {
			s.onSpawn(event)
		}
	}

	return events
}

// SpawnLine spawns entities in a line.
func (s *SpawnerSystem) SpawnLine(
	world *ecs.World,
	spawnType string,
	x1, y1, x2, y2 float64,
	count int,
) []SpawnEvent {
	events := make([]SpawnEvent, 0, count)
	for i := range count {
		t := float64(i) / float64(count-1)
		if count == 1 {
			t = 0.5
		}

		x := x1 + (x2-x1)*t
		y := y1 + (y2-y1)*t

		event := SpawnEvent{
			SpawnType: spawnType,
			X:         x,
			Y:         y,
			Data:      make(map[string]any),
		}

		if factory, ok := s.factories[spawnType]; ok {
			event.Entity = factory(world, event)
		}

		events = append(events, event)
		if s.onSpawn != nil {
			s.onSpawn(event)
		}
	}

	return events
}
