package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
	"td/internal/engine"
	"td/internal/systems"
)

// GameState represents the current game state.
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateCardSelect
	StatePaused
	StateGameOver
	StateVictory
)

// TDGame is the main tower defense game.
type TDGame struct {
	engine.BaseScene

	// ECS
	World *ecs.World

	// Game objects
	TDMap        *TDMap
	WaveManager  *WaveManager
	CardSelector *CardSelector
	Hero         *Hero
	HeroEntity   ecs.Entity

	// Monsters tracking
	MonsterMoveSystem *MonsterMovementSystem
	ActiveMonsters    map[ecs.Entity]*Monster

	// Systems
	Input *systems.InputManager

	// Game state
	State       GameState
	Lives       int
	Gold        int
	Score       int
	CurrentWave int
	DeltaTime   *engine.DeltaTime

	// Screen dimensions
	Width  int
	Height int
}

// NewTDGame creates a new tower defense game.
func NewTDGame(width, height int) *TDGame {
	world := ecs.NewWorld()

	game := &TDGame{
		World:          &world,
		TDMap:          CreateDefaultMap(),
		WaveManager:    NewWaveManager(),
		CardSelector:   NewCardSelector(),
		ActiveMonsters: make(map[ecs.Entity]*Monster),
		Input:          systems.NewInputManager(),
		State:          StatePlaying,
		Lives:          20,
		Gold:           100,
		DeltaTime:      engine.NewDeltaTime(60),
		Width:          width,
		Height:         height,
	}

	// Create monster movement system
	game.MonsterMoveSystem = NewMonsterMovementSystem(game.TDMap)

	// Create hero
	spawnX, spawnY := game.TDMap.TileToWorld(12, 7)
	game.Hero = NewHero("Guardian")
	game.HeroEntity = CreateHeroEntity(&world, game.Hero, spawnX, spawnY)

	// Setup input bindings
	game.Input.BindAction("pause", ebiten.KeyEscape)
	game.Input.BindAction("select", ebiten.Key1, ebiten.Key2, ebiten.Key3)

	return game
}

// Load implements Scene.
func (g *TDGame) Load() error {
	return nil
}

// Unload implements Scene.
func (g *TDGame) Unload() {
	// Cleanup
}

// Update implements Scene.
func (g *TDGame) Update() error {
	dt := g.DeltaTime.Update()
	g.Input.Update()

	switch g.State {
	case StatePlaying:
		g.updatePlaying(dt)
	case StateCardSelect:
		g.updateCardSelect()
	case StatePaused:
		if g.Input.IsActionJustPressed("pause") {
			g.State = StatePlaying
		}
	}

	return nil
}

func (g *TDGame) updatePlaying(dt float64) {
	// Handle pause
	if g.Input.IsActionJustPressed("pause") {
		g.State = StatePaused

		return
	}

	// Update wave spawning
	monsterType := g.WaveManager.Update(dt)
	if monsterType != "" {
		g.spawnMonster(monsterType)
	}

	// Update monsters
	g.updateMonsters(dt)

	// Update hero attacks
	g.updateHeroAttacks(dt)

	// Check for wave completion -> card selection
	if len(g.ActiveMonsters) == 0 && !g.WaveManager.WaveActive &&
		g.WaveManager.CurrentWave > g.CurrentWave {
		g.CurrentWave = g.WaveManager.CurrentWave
		if !g.WaveManager.AllComplete {
			g.CardSelector.GenerateChoices(g.CurrentWave)
			g.State = StateCardSelect
		}
	}

	// Check win/lose conditions
	if g.Lives <= 0 {
		g.State = StateGameOver
	}

	if g.WaveManager.AllComplete && len(g.ActiveMonsters) == 0 {
		g.State = StateVictory
	}
}

func (g *TDGame) updateCardSelect() {
	mx, my := g.Input.MousePosition()
	clicked := g.Input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	card := g.CardSelector.HandleInput(mx, my, clicked, g.Width, g.Height)
	if card != nil {
		g.CardSelector.ApplyCard(card, g.Hero)
		g.State = StatePlaying
	}
}

func (g *TDGame) spawnMonster(monsterType string) {
	spawnX, spawnY := g.TDMap.TileToWorld(g.TDMap.SpawnPoint.X, g.TDMap.SpawnPoint.Y)
	entity, monster := CreateMonsterEntity(g.World, monsterType, spawnX, spawnY)
	g.ActiveMonsters[entity] = monster
	g.MonsterMoveSystem.AddMonster(entity, monster)
}

func (g *TDGame) updateMonsters(dt float64) {
	posMapper := ecs.NewMap1[components.Position](g.World)

	for entity := range g.ActiveMonsters {
		pos := posMapper.Get(entity)
		if pos == nil {
			continue
		}

		reachedEnd := g.MonsterMoveSystem.UpdateMonster(entity, pos, dt)
		if reachedEnd {
			g.Lives--
			g.removeMonster(entity)
		}
	}
}

func (g *TDGame) updateHeroAttacks(dt float64) {
	if !g.Hero.CanAttack(dt) {
		return
	}

	// Find closest monster in range
	posMapper := ecs.NewMap1[components.Position](g.World)

	heroPos := posMapper.Get(g.HeroEntity)
	if heroPos == nil {
		return
	}

	var (
		closestEntity  ecs.Entity
		closestDist    = math.MaxFloat64
		closestMonster *Monster
	)

	for entity, monster := range g.ActiveMonsters {
		pos := posMapper.Get(entity)
		if pos == nil {
			continue
		}

		dx := pos.X - heroPos.X
		dy := pos.Y - heroPos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist <= g.Hero.AttackRange && dist < closestDist {
			closestDist = dist
			closestEntity = entity
			closestMonster = monster
		}
	}

	// Attack the closest monster
	if closestMonster != nil {
		healthMapper := ecs.NewMap1[components.Health](g.World)
		health := healthMapper.Get(closestEntity)

		if health != nil {
			health.Current -= g.Hero.AttackDamage
			if health.Current <= 0 {
				g.Gold += closestMonster.Experience / 2

				g.Score += closestMonster.Experience
				g.Hero.GainExp(closestMonster.Experience) // Level ups handled automatically

				g.removeMonster(closestEntity)
			}
		}
	}
}

func (g *TDGame) removeMonster(entity ecs.Entity) {
	delete(g.ActiveMonsters, entity)
	g.MonsterMoveSystem.RemoveMonster(entity)
	g.World.RemoveEntity(entity)
}

// Draw implements Scene.
func (g *TDGame) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{R: 30, G: 30, B: 40, A: 255})

	// Draw map
	g.TDMap.Draw(screen)

	// Draw entities (monsters, hero)
	g.drawEntities(screen)

	// Draw UI
	g.drawUI(screen)

	// Draw card selector if active
	g.CardSelector.Draw(screen, g.Width, g.Height)
}

func (g *TDGame) drawEntities(screen *ebiten.Image) {
	spriteFilter := ecs.NewFilter2[components.Position, components.Sprite](g.World)
	query := spriteFilter.Query()

	for query.Next() {
		pos, sprite := query.Get()
		if !sprite.Visible || sprite.Image == nil {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(sprite.ScaleX, sprite.ScaleY)
		op.GeoM.Translate(pos.X+sprite.OffsetX, pos.Y+sprite.OffsetY)
		screen.DrawImage(sprite.Image, op)
	}

	// Draw health bars for monsters
	healthFilter := ecs.NewFilter2[components.Position, components.Health](g.World)
	hQuery := healthFilter.Query()

	for hQuery.Next() {
		pos, health := hQuery.Get()
		if health.Current >= health.Max {
			continue
		}

		// Health bar background
		bgWidth := 20
		bg := ebiten.NewImage(bgWidth, 4)
		bg.Fill(color.RGBA{R: 100, G: 0, B: 0, A: 255})

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pos.X-float64(bgWidth)/2, pos.Y-15)
		screen.DrawImage(bg, op)

		// Health bar foreground
		fgWidth := int(float64(bgWidth) * float64(health.Current) / float64(health.Max))
		if fgWidth > 0 {
			fg := ebiten.NewImage(fgWidth, 4)
			fg.Fill(color.RGBA{R: 0, G: 200, B: 0, A: 255})
			screen.DrawImage(fg, op)
		}
	}
}

func (g *TDGame) drawUI(screen *ebiten.Image) {
	// UI panel at top
	panel := ebiten.NewImage(g.Width, 40)
	panel.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 220})
	screen.DrawImage(panel, nil)

	// Draw state-specific overlays
	switch g.State {
	case StatePaused:
		g.drawCenteredOverlay(
			screen,
			"PAUSED - Press ESC to resume",
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
		)
	case StateGameOver:
		g.drawCenteredOverlay(screen, "GAME OVER", color.RGBA{R: 255, G: 50, B: 50, A: 255})
	case StateVictory:
		g.drawCenteredOverlay(screen, "VICTORY!", color.RGBA{R: 50, G: 255, B: 50, A: 255})
	}
}

func (g *TDGame) drawCenteredOverlay(screen *ebiten.Image, text string, clr color.RGBA) {
	overlay := ebiten.NewImage(g.Width, g.Height)
	overlay.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 180})
	screen.DrawImage(overlay, nil)

	// Draw a simple colored rectangle as placeholder for text
	textBg := ebiten.NewImage(300, 50)
	textBg.Fill(clr)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.Width-300)/2, float64(g.Height-50)/2)
	screen.DrawImage(textBg, op)
}
