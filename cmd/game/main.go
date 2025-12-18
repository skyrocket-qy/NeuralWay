package main

import (
	"log"

	"td/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 800
	screenHeight = 480
)

// GameWrapper wraps the TDGame to implement ebiten.Game interface.
type GameWrapper struct {
	tdGame *game.TDGame
}

func (g *GameWrapper) Update() error {
	return g.tdGame.Update()
}

func (g *GameWrapper) Draw(screen *ebiten.Image) {
	g.tdGame.Draw(screen)
}

func (g *GameWrapper) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// Create TD game
	tdGame := game.NewTDGame(screenWidth, screenHeight)

	wrapper := &GameWrapper{tdGame: tdGame}

	// Configure window
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Tower Defense - AI ECS Framework Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run the game
	if err := ebiten.RunGame(wrapper); err != nil {
		log.Fatal(err)
	}
}
