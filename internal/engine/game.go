package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
)

// Game wraps Ebitengine and Ark ECS world for the game loop.
type Game struct {
	World  ecs.World
	width  int
	height int
	title  string

	// Systems to run each frame
	updateSystems []System
	drawSystems   []DrawSystem
}

// System is an interface for ECS systems that run during Update.
type System interface {
	Update(world *ecs.World)
}

// DrawSystem is an interface for ECS systems that run during Draw.
type DrawSystem interface {
	Draw(world *ecs.World, screen *ebiten.Image)
}

// NewGame creates a new game instance with given dimensions.
func NewGame(width, height int, title string) *Game {
	return &Game{
		World:         ecs.NewWorld(),
		width:         width,
		height:        height,
		title:         title,
		updateSystems: make([]System, 0),
		drawSystems:   make([]DrawSystem, 0),
	}
}

// AddSystem adds an update system to the game loop.
func (g *Game) AddSystem(s System) {
	g.updateSystems = append(g.updateSystems, s)
}

// AddDrawSystem adds a draw system to the game loop.
func (g *Game) AddDrawSystem(s DrawSystem) {
	g.drawSystems = append(g.drawSystems, s)
}

// Update implements ebiten.Game interface.
func (g *Game) Update() error {
	for _, s := range g.updateSystems {
		s.Update(&g.World)
	}

	return nil
}

// Draw implements ebiten.Game interface.
func (g *Game) Draw(screen *ebiten.Image) {
	for _, s := range g.drawSystems {
		s.Draw(&g.World, screen)
	}
}

// Layout implements ebiten.Game interface.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.width, g.height
}

// Run starts the game loop.
func (g *Game) Run() error {
	ebiten.SetWindowSize(g.width, g.height)
	ebiten.SetWindowTitle(g.title)

	return ebiten.RunGame(g)
}
