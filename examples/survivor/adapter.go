package main

import (
	"td/internal/ai"
)

// SurvivorAdapter implements ai.GameAdapter for automated QA testing.
type SurvivorAdapter struct {
	game   *Game
	tick   int64
	inputs InputState
}

// InputState holds simulated input for the game.
type InputState struct {
	Up, Down, Left, Right bool
}

// NewSurvivorAdapter creates an adapter for the given game.
func NewSurvivorAdapter(game *Game) *SurvivorAdapter {
	return &SurvivorAdapter{
		game: game,
	}
}

// Name returns the game's identifier.
func (a *SurvivorAdapter) Name() string {
	return "Dev Survivor"
}

// GetState returns the current game state.
func (a *SurvivorAdapter) GetState() ai.GameState {
	state := ai.GameState{
		Tick:        a.tick,
		Score:       a.game.killCount,
		EntityCount: len(a.game.enemies) + len(a.game.projectiles) + len(a.game.xpGems),
		CustomData:  make(map[string]any),
	}

	if a.game.player != nil {
		state.PlayerPos = [2]float64{a.game.player.X, a.game.player.Y}
		state.PlayerHealth = [2]int{a.game.player.HP, a.game.player.MaxHP}
		state.Score = a.game.killCount

		// Add custom data
		state.CustomData["level"] = a.game.player.Level
		state.CustomData["xp"] = a.game.player.XP
		state.CustomData["game_time"] = a.game.gameTime
		state.CustomData["weapons"] = len(a.game.player.Weapons)
		state.CustomData["enemies"] = len(a.game.enemies)
	}

	return state
}

// IsGameOver returns true if the game has ended.
func (a *SurvivorAdapter) IsGameOver() bool {
	return a.game.state == StateGameOver
}

// GetScore returns the current kill count.
func (a *SurvivorAdapter) GetScore() int {
	return a.game.killCount
}

// AvailableActions returns movement directions.
func (a *SurvivorAdapter) AvailableActions() []ai.ActionType {
	if a.game.state != StatePlaying {
		return []ai.ActionType{ai.ActionNone}
	}

	return []ai.ActionType{
		ai.ActionNone,
		ai.ActionMoveUp,
		ai.ActionMoveDown,
		ai.ActionMoveLeft,
		ai.ActionMoveRight,
	}
}

// PerformAction applies the given action as input.
func (a *SurvivorAdapter) PerformAction(action ai.ActionType) error {
	// Reset inputs
	a.inputs = InputState{}

	// Set direction
	switch action {
	case ai.ActionMoveUp:
		a.inputs.Up = true
	case ai.ActionMoveDown:
		a.inputs.Down = true
	case ai.ActionMoveLeft:
		a.inputs.Left = true
	case ai.ActionMoveRight:
		a.inputs.Right = true
	}

	return nil
}

// Step advances the game by one tick using the current inputs.
func (a *SurvivorAdapter) Step() error {
	a.tick++

	// Only step if we're in playing state
	if a.game.state != StatePlaying {
		return nil
	}

	dt := 1.0 / 60.0
	a.game.gameTime += dt
	a.game.player.HitTimer -= dt

	// Apply recovery
	if a.game.player.Recovery > 0 {
		a.game.player.HP += int(a.game.player.Recovery * dt)
		if a.game.player.HP > a.game.player.MaxHP {
			a.game.player.HP = a.game.player.MaxHP
		}
	}

	// Apply movement from inputs
	dx, dy := 0.0, 0.0
	if a.inputs.Up {
		dy = -1
	}

	if a.inputs.Down {
		dy = 1
	}

	if a.inputs.Left {
		dx = -1
	}

	if a.inputs.Right {
		dx = 1
	}

	if dx != 0 && dy != 0 {
		dx *= 0.707
		dy *= 0.707
	}

	a.game.player.X += dx * a.game.player.Speed
	a.game.player.Y += dy * a.game.player.Speed

	// Update camera
	a.game.cameraX = a.game.player.X - float64(screenWidth)/2
	a.game.cameraY = a.game.player.Y - float64(screenHeight)/2

	// Spawn enemies
	a.game.spawnTimer += dt

	spawnRate := 1.0 - a.game.gameTime*0.01
	if spawnRate < 0.05 {
		spawnRate = 0.05
	}

	for a.game.spawnTimer >= spawnRate {
		a.game.spawnEnemy()
		a.game.spawnTimer -= spawnRate
	}

	// Boss timer
	a.game.bossTimer += dt
	if a.game.bossTimer >= 180 {
		a.game.spawnBoss()
		a.game.bossTimer = 0
	}

	// Update weapons
	for _, w := range a.game.player.Weapons {
		def := WeaponDefs[w.Type]
		cooldown := def.Cooldown * a.game.player.CooldownMult

		w.Timer += dt
		if w.Timer >= cooldown {
			a.game.fireWeapon(w)
			w.Timer = 0
		}
	}

	// Update game components
	a.game.updateProjectiles(dt)
	a.game.updateEnemies(dt)
	a.game.collectXP(dt)
	a.game.updateParticles(dt)

	// Update damage numbers
	for i := len(a.game.damageNumbers) - 1; i >= 0; i-- {
		d := a.game.damageNumbers[i]
		d.X += d.VX * dt
		d.Y += d.VY * dt
		d.VY += 200 * dt

		d.Timer -= dt
		if d.Timer <= 0 {
			a.game.freeDamageNumber(d)
			a.game.damageNumbers = append(a.game.damageNumbers[:i], a.game.damageNumbers[i+1:]...)
		}
	}

	return nil
}

// Reset restarts the game.
func (a *SurvivorAdapter) Reset() error {
	a.tick = 0
	a.inputs = InputState{}
	// Start game with default character
	a.game.startGame(CharJunior)

	return nil
}

// GetGameTime returns the current in-game time.
func (a *SurvivorAdapter) GetGameTime() float64 {
	return a.game.gameTime
}

// GetPlayerLevel returns the player's current level.
func (a *SurvivorAdapter) GetPlayerLevel() int {
	if a.game.player == nil {
		return 0
	}

	return a.game.player.Level
}

// GetWeaponCount returns how many weapons the player has.
func (a *SurvivorAdapter) GetWeaponCount() int {
	if a.game.player == nil {
		return 0
	}

	return len(a.game.player.Weapons)
}

// GetEnemyCount returns the current enemy count.
func (a *SurvivorAdapter) GetEnemyCount() int {
	return len(a.game.enemies)
}
