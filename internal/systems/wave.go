package systems

import (
	"github.com/mlange-42/ark/ecs"
)

// WaveConfig defines a wave of enemies.
type WaveConfig struct {
	WaveNumber     int
	Enemies        map[string]int // SpawnType -> count
	SpawnDelay     float64        // Delay between spawns
	WaveDelay      float64        // Delay before wave starts
	BossWave       bool
	DifficultyMult float64 // Difficulty multiplier for this wave
}

// NewWaveConfig creates a wave configuration.
func NewWaveConfig(waveNumber int) WaveConfig {
	return WaveConfig{
		WaveNumber:     waveNumber,
		Enemies:        make(map[string]int),
		SpawnDelay:     0.5,
		WaveDelay:      3.0,
		DifficultyMult: 1.0,
	}
}

// WaveState represents the current wave state.
type WaveState int

const (
	WaveIdle WaveState = iota
	WaveStarting
	WaveActive
	WaveClearing
	WaveComplete
	WavesFinal
)

// WaveSystem manages wave-based gameplay.
type WaveSystem struct {
	waves          []WaveConfig
	currentWave    int
	state          WaveState
	waveTimer      float64
	spawnTimer     float64
	spawnQueue     []string // Remaining spawns for current wave
	aliveCount     int      // Entities alive from current wave
	totalKills     int
	spawnerSystem  *SpawnerSystem
	onWaveStart    func(int)
	onWaveEnd      func(int)
	onAllWavesDone func()
}

// NewWaveSystem creates a wave system.
func NewWaveSystem(spawnerSystem *SpawnerSystem) *WaveSystem {
	return &WaveSystem{
		waves:         make([]WaveConfig, 0),
		spawnerSystem: spawnerSystem,
		state:         WaveIdle,
	}
}

// SetOnWaveStart sets the wave start callback.
func (s *WaveSystem) SetOnWaveStart(fn func(int)) {
	s.onWaveStart = fn
}

// SetOnWaveEnd sets the wave end callback.
func (s *WaveSystem) SetOnWaveEnd(fn func(int)) {
	s.onWaveEnd = fn
}

// SetOnAllWavesDone sets the callback for completing all waves.
func (s *WaveSystem) SetOnAllWavesDone(fn func()) {
	s.onAllWavesDone = fn
}

// AddWave adds a wave configuration.
func (s *WaveSystem) AddWave(wave WaveConfig) {
	s.waves = append(s.waves, wave)
}

// GenerateWaves generates waves with scaling difficulty.
func (s *WaveSystem) GenerateWaves(count int, baseEnemies map[string]int, scaling float64) {
	for i := 1; i <= count; i++ {
		wave := NewWaveConfig(i)
		wave.DifficultyMult = 1.0 + (float64(i-1) * scaling)

		for enemyType, baseCount := range baseEnemies {
			wave.Enemies[enemyType] = int(float64(baseCount) * wave.DifficultyMult)
		}

		// Every 5th wave is a boss wave
		if i%5 == 0 {
			wave.BossWave = true
			wave.WaveDelay = 5.0
		}

		s.waves = append(s.waves, wave)
	}
}

// StartWaves begins wave spawning.
func (s *WaveSystem) StartWaves() {
	s.currentWave = 0
	s.state = WaveStarting
	s.startCurrentWave()
}

// startCurrentWave prepares the current wave.
func (s *WaveSystem) startCurrentWave() {
	if s.currentWave >= len(s.waves) {
		s.state = WavesFinal
		if s.onAllWavesDone != nil {
			s.onAllWavesDone()
		}

		return
	}

	wave := &s.waves[s.currentWave]
	s.state = WaveStarting
	s.waveTimer = wave.WaveDelay

	// Build spawn queue
	s.spawnQueue = s.spawnQueue[:0]

	for enemyType, count := range wave.Enemies {
		for range count {
			s.spawnQueue = append(s.spawnQueue, enemyType)
		}
	}

	// Shuffle spawn queue
	for i := len(s.spawnQueue) - 1; i > 0; i-- {
		j := int(float64(i+1) * 0.5) // Simple shuffle
		s.spawnQueue[i], s.spawnQueue[j] = s.spawnQueue[j], s.spawnQueue[i]
	}

	s.aliveCount = 0

	if s.onWaveStart != nil {
		s.onWaveStart(wave.WaveNumber)
	}
}

// Update updates the wave system.
func (s *WaveSystem) Update(world *ecs.World, dt float64) {
	switch s.state {
	case WaveIdle:
		// Waiting to start
		return

	case WaveStarting:
		// Countdown before wave
		s.waveTimer -= dt
		if s.waveTimer <= 0 {
			s.state = WaveActive
			s.spawnTimer = 0
		}

	case WaveActive:
		// Spawning enemies
		if len(s.spawnQueue) > 0 {
			wave := &s.waves[s.currentWave]

			s.spawnTimer -= dt
			if s.spawnTimer <= 0 {
				// Spawn next enemy
				enemyType := s.spawnQueue[0]
				s.spawnQueue = s.spawnQueue[1:]

				// Use spawner system if available
				if s.spawnerSystem != nil {
					// Find a spawn point for this type or use default
					for id, sp := range s.spawnerSystem.spawnPoints {
						if sp.SpawnType == enemyType {
							s.spawnerSystem.SpawnNow(id)

							break
						}
					}
				}

				s.aliveCount++
				s.spawnTimer = wave.SpawnDelay
			}
		}

		// Check if wave is done spawning
		if len(s.spawnQueue) == 0 {
			s.state = WaveClearing
		}

	case WaveClearing:
		// Waiting for all enemies to die
		if s.aliveCount <= 0 {
			s.state = WaveComplete
			if s.onWaveEnd != nil {
				s.onWaveEnd(s.waves[s.currentWave].WaveNumber)
			}

			// Start next wave
			s.currentWave++
			s.startCurrentWave()
		}

	case WaveComplete, WavesFinal:
		// Done
		return
	}
}

// EnemyKilled notifies the system that an enemy died.
func (s *WaveSystem) EnemyKilled() {
	s.aliveCount--

	s.totalKills++
	if s.aliveCount < 0 {
		s.aliveCount = 0
	}
}

// GetCurrentWave returns the current wave number (1-indexed).
func (s *WaveSystem) GetCurrentWave() int {
	if s.currentWave < len(s.waves) {
		return s.waves[s.currentWave].WaveNumber
	}

	return s.currentWave + 1
}

// GetTotalWaves returns total number of waves.
func (s *WaveSystem) GetTotalWaves() int {
	return len(s.waves)
}

// GetWaveProgress returns progress through current wave (0.0-1.0).
func (s *WaveSystem) GetWaveProgress() float64 {
	if s.currentWave >= len(s.waves) {
		return 1.0
	}

	wave := &s.waves[s.currentWave]

	totalEnemies := 0
	for _, count := range wave.Enemies {
		totalEnemies += count
	}

	if totalEnemies == 0 {
		return 1.0
	}

	killed := totalEnemies - len(s.spawnQueue) - s.aliveCount

	return float64(killed) / float64(totalEnemies)
}

// GetRemainingEnemies returns enemies still to spawn + alive.
func (s *WaveSystem) GetRemainingEnemies() int {
	return len(s.spawnQueue) + s.aliveCount
}

// GetState returns the current wave state.
func (s *WaveSystem) GetState() WaveState {
	return s.state
}

// GetTotalKills returns total enemies killed.
func (s *WaveSystem) GetTotalKills() int {
	return s.totalKills
}

// IsBossWave returns true if current wave is a boss wave.
func (s *WaveSystem) IsBossWave() bool {
	if s.currentWave < len(s.waves) {
		return s.waves[s.currentWave].BossWave
	}

	return false
}

// SkipWave skips to the next wave.
func (s *WaveSystem) SkipWave() {
	s.aliveCount = 0
	s.spawnQueue = s.spawnQueue[:0]
	s.currentWave++
	s.startCurrentWave()
}

// Reset restarts from wave 1.
func (s *WaveSystem) Reset() {
	s.currentWave = 0
	s.state = WaveIdle
	s.totalKills = 0
	s.aliveCount = 0
	s.spawnQueue = s.spawnQueue[:0]
}
