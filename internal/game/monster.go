package game

import (
	"image/color"
	"math"
	"td/internal/components"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
)

// Monster represents an enemy that follows the path.
type Monster struct {
	Name       string
	Health     int
	MaxHealth  int
	Speed      float64
	PathIndex  int
	Experience int // Exp granted when killed
	ReachedEnd bool
}

// NewMonster creates a new monster.
func NewMonster(name string, health int, speed float64, exp int) *Monster {
	return &Monster{
		Name:       name,
		Health:     health,
		MaxHealth:  health,
		Speed:      speed,
		PathIndex:  0,
		Experience: exp,
	}
}

// TakeDamage applies damage and returns true if monster is dead.
func (m *Monster) TakeDamage(damage int) bool {
	m.Health -= damage
	return m.Health <= 0
}

// MonsterType defines different monster variants.
type MonsterType struct {
	Name   string
	Color  color.RGBA
	Health int
	Speed  float64
	Exp    int
	Size   int
}

// Predefined monster types.
var MonsterTypes = map[string]MonsterType{
	"goblin": {
		Name:   "Goblin",
		Color:  color.RGBA{R: 0, G: 200, B: 0, A: 255},
		Health: 30,
		Speed:  50,
		Exp:    10,
		Size:   16,
	},
	"orc": {
		Name:   "Orc",
		Color:  color.RGBA{R: 100, G: 150, B: 0, A: 255},
		Health: 80,
		Speed:  30,
		Exp:    25,
		Size:   20,
	},
	"troll": {
		Name:   "Troll",
		Color:  color.RGBA{R: 150, G: 100, B: 50, A: 255},
		Health: 200,
		Speed:  20,
		Exp:    50,
		Size:   28,
	},
	"boss": {
		Name:   "Boss",
		Color:  color.RGBA{R: 200, G: 0, B: 50, A: 255},
		Health: 500,
		Speed:  15,
		Exp:    200,
		Size:   32,
	},
}

// CreateMonsterEntity creates an ECS entity for a monster.
func CreateMonsterEntity(world *ecs.World, monsterType string, x, y float64) (ecs.Entity, *Monster) {
	mt := MonsterTypes[monsterType]
	if mt.Name == "" {
		mt = MonsterTypes["goblin"] // Default
	}

	monster := NewMonster(mt.Name, mt.Health, mt.Speed, mt.Exp)

	img := ebiten.NewImage(mt.Size, mt.Size)
	img.Fill(mt.Color)

	mapper := ecs.NewMap4[components.Position, components.Velocity, components.Sprite, components.Health](world)
	entity := mapper.NewEntity(
		&components.Position{X: x, Y: y},
		&components.Velocity{X: 0, Y: 0},
		&components.Sprite{
			Image:   img,
			OffsetX: -float64(mt.Size) / 2,
			OffsetY: -float64(mt.Size) / 2,
			ScaleX:  1,
			ScaleY:  1,
			Visible: true,
		},
		&components.Health{
			Current: mt.Health,
			Max:     mt.Health,
		},
	)

	return entity, monster
}

// Wave defines a spawning wave of monsters.
type Wave struct {
	Monsters   []WaveMonster
	SpawnDelay float64 // Delay between monster spawns
	WaveDelay  float64 // Delay before wave starts
}

// WaveMonster defines a monster spawn in a wave.
type WaveMonster struct {
	Type  string
	Count int
}

// WaveManager manages monster wave spawning.
type WaveManager struct {
	Waves        []Wave
	CurrentWave  int
	SpawnTimer   float64
	WaveTimer    float64
	SpawnIndex   int
	MonsterIndex int
	WaveActive   bool
	AllComplete  bool
}

// NewWaveManager creates a wave manager.
func NewWaveManager() *WaveManager {
	return &WaveManager{
		Waves: []Wave{
			// Wave 1: Basic goblins
			{
				Monsters:   []WaveMonster{{Type: "goblin", Count: 5}},
				SpawnDelay: 1.0,
				WaveDelay:  3.0,
			},
			// Wave 2: More goblins
			{
				Monsters:   []WaveMonster{{Type: "goblin", Count: 8}},
				SpawnDelay: 0.8,
				WaveDelay:  5.0,
			},
			// Wave 3: Mixed goblins and orcs
			{
				Monsters: []WaveMonster{
					{Type: "goblin", Count: 5},
					{Type: "orc", Count: 3},
				},
				SpawnDelay: 0.7,
				WaveDelay:  5.0,
			},
			// Wave 4: Orcs
			{
				Monsters:   []WaveMonster{{Type: "orc", Count: 6}},
				SpawnDelay: 0.6,
				WaveDelay:  5.0,
			},
			// Wave 5: Boss wave
			{
				Monsters: []WaveMonster{
					{Type: "troll", Count: 2},
					{Type: "boss", Count: 1},
				},
				SpawnDelay: 2.0,
				WaveDelay:  8.0,
			},
		},
		CurrentWave: 0,
		WaveActive:  false,
	}
}

// Update updates the wave manager and returns monster type to spawn (or empty).
func (w *WaveManager) Update(dt float64) string {
	if w.AllComplete {
		return ""
	}

	if !w.WaveActive {
		// Waiting for wave to start
		w.WaveTimer += dt
		if w.WaveTimer >= w.Waves[w.CurrentWave].WaveDelay {
			w.WaveActive = true
			w.WaveTimer = 0
			w.SpawnTimer = 0
			w.SpawnIndex = 0
			w.MonsterIndex = 0
		}
		return ""
	}

	// Wave is active, spawn monsters
	wave := w.Waves[w.CurrentWave]
	w.SpawnTimer += dt

	if w.SpawnTimer >= wave.SpawnDelay {
		w.SpawnTimer = 0

		// Get current monster type
		if w.SpawnIndex < len(wave.Monsters) {
			spawner := wave.Monsters[w.SpawnIndex]
			monsterType := spawner.Type

			w.MonsterIndex++
			if w.MonsterIndex >= spawner.Count {
				w.MonsterIndex = 0
				w.SpawnIndex++
			}

			// Check if wave is complete
			if w.SpawnIndex >= len(wave.Monsters) {
				w.WaveActive = false
				w.CurrentWave++
				if w.CurrentWave >= len(w.Waves) {
					w.AllComplete = true
				}
			}

			return monsterType
		}
	}

	return ""
}

// MonsterMovementSystem moves monsters along the path.
type MonsterMovementSystem struct {
	TDMap    *TDMap
	Monsters map[ecs.Entity]*Monster
}

// NewMonsterMovementSystem creates a monster movement system.
func NewMonsterMovementSystem(tdMap *TDMap) *MonsterMovementSystem {
	return &MonsterMovementSystem{
		TDMap:    tdMap,
		Monsters: make(map[ecs.Entity]*Monster),
	}
}

// AddMonster registers a monster entity.
func (s *MonsterMovementSystem) AddMonster(entity ecs.Entity, monster *Monster) {
	s.Monsters[entity] = monster
}

// RemoveMonster unregisters a monster entity.
func (s *MonsterMovementSystem) RemoveMonster(entity ecs.Entity) {
	delete(s.Monsters, entity)
}

// UpdateMonster updates a single monster's movement.
func (s *MonsterMovementSystem) UpdateMonster(entity ecs.Entity, pos *components.Position, dt float64) bool {
	monster := s.Monsters[entity]
	if monster == nil || monster.ReachedEnd {
		return false
	}

	path := s.TDMap.Path
	if monster.PathIndex >= len(path) {
		monster.ReachedEnd = true
		return true // Reached end
	}

	// Get target tile center
	target := path[monster.PathIndex]
	tx, ty := s.TDMap.TileToWorld(target.X, target.Y)

	// Calculate direction to target
	dx := tx - pos.X
	dy := ty - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 2.0 {
		// Reached waypoint, move to next
		monster.PathIndex++
		return false
	}

	// Move towards target
	speed := monster.Speed * dt
	pos.X += (dx / dist) * speed
	pos.Y += (dy / dist) * speed

	return false
}
