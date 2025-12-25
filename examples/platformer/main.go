package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 640
	screenHeight = 480
	gravity      = 0.5
	jumpForce    = -12
	moveSpeed    = 4
	tileSize     = 32
)

// Player represents the player character.
type Player struct {
	X, Y       float64
	VX, VY     float64
	OnGround   bool
	FacingLeft bool
	JumpCount  int
}

// Coin represents a collectible.
type Coin struct {
	X, Y      float64
	Collected bool
}

// Game represents the platformer game.
type Game struct {
	player     *Player
	coins      []*Coin
	level      [][]int
	cameraX    float64
	score      int
	levelWidth int
	won        bool
}

// Level tiles: 0=empty, 1=ground, 2=platform, 3=coin, 4=goal.
var levelData = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 2, 2, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 0, 0, 0, 0, 0, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 2, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0},
	{0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
}

// NewGame creates a new game.
func NewGame() *Game {
	g := &Game{
		player: &Player{
			X: 50,
			Y: float64(len(levelData)-2)*tileSize - 32,
		},
		level:      levelData,
		levelWidth: len(levelData[0]) * tileSize,
		coins:      make([]*Coin, 0),
	}

	// Find coins in level
	for y, row := range levelData {
		for x, tile := range row {
			if tile == 3 {
				g.coins = append(g.coins, &Coin{
					X: float64(x*tileSize) + tileSize/2,
					Y: float64(y*tileSize) + tileSize/2,
				})
			}
		}
	}

	return g
}

func (g *Game) reset() {
	g.player.X = 50
	g.player.Y = float64(len(levelData)-2)*tileSize - 32
	g.player.VX = 0
	g.player.VY = 0
	g.player.OnGround = false
	g.player.JumpCount = 0
	g.score = 0
	g.won = false

	g.cameraX = 0
	for _, c := range g.coins {
		c.Collected = false
	}
}

func (g *Game) Update() error {
	if g.won {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.reset()
		}

		return nil
	}

	// Horizontal movement
	g.player.VX = 0
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.VX = -moveSpeed
		g.player.FacingLeft = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.VX = moveSpeed
		g.player.FacingLeft = false
	}

	// Jump (double jump allowed)
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if g.player.JumpCount < 2 {
			g.player.VY = jumpForce
			g.player.JumpCount++
			g.player.OnGround = false
		}
	}

	// Apply gravity
	g.player.VY += gravity
	if g.player.VY > 15 {
		g.player.VY = 15
	}

	// Move X
	g.player.X += g.player.VX
	g.resolveCollisionX()

	// Move Y
	g.player.Y += g.player.VY
	g.resolveCollisionY()

	// Keep in bounds
	if g.player.X < 16 {
		g.player.X = 16
	}

	if g.player.X > float64(g.levelWidth)-16 {
		g.player.X = float64(g.levelWidth) - 16
	}

	// Fall death
	if g.player.Y > float64(len(g.level)*tileSize) {
		g.reset()
	}

	// Collect coins
	for _, c := range g.coins {
		if !c.Collected {
			dx := g.player.X - c.X

			dy := g.player.Y - c.Y
			if math.Sqrt(dx*dx+dy*dy) < 24 {
				c.Collected = true
				g.score += 100
			}
		}
	}

	// Check goal
	for y, row := range g.level {
		for x, tile := range row {
			if tile == 4 {
				goalX := float64(x*tileSize) + tileSize/2
				goalY := float64(y*tileSize) + tileSize/2
				dx := g.player.X - goalX

				dy := g.player.Y - goalY
				if math.Sqrt(dx*dx+dy*dy) < 30 {
					g.won = true
				}
			}
		}
	}

	// Camera follow
	targetCam := g.player.X - float64(screenWidth)/2

	g.cameraX += (targetCam - g.cameraX) * 0.1
	if g.cameraX < 0 {
		g.cameraX = 0
	}

	if g.cameraX > float64(g.levelWidth-screenWidth) {
		g.cameraX = float64(g.levelWidth - screenWidth)
	}

	return nil
}

func (g *Game) resolveCollisionX() {
	px := int(g.player.X) / tileSize
	py := int(g.player.Y) / tileSize

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			tx, ty := px+dx, py+dy
			if ty >= 0 && ty < len(g.level) && tx >= 0 && tx < len(g.level[ty]) {
				tile := g.level[ty][tx]
				if tile == 1 || tile == 2 {
					if g.checkTileCollision(tx, ty) {
						if g.player.VX > 0 {
							g.player.X = float64(tx*tileSize) - 16
						} else if g.player.VX < 0 {
							g.player.X = float64((tx+1)*tileSize) + 16
						}
					}
				}
			}
		}
	}
}

func (g *Game) resolveCollisionY() {
	px := int(g.player.X) / tileSize
	py := int(g.player.Y) / tileSize

	g.player.OnGround = false

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			tx, ty := px+dx, py+dy
			if ty >= 0 && ty < len(g.level) && tx >= 0 && tx < len(g.level[ty]) {
				tile := g.level[ty][tx]
				if tile == 1 || tile == 2 {
					if g.checkTileCollision(tx, ty) {
						if g.player.VY > 0 {
							g.player.Y = float64(ty*tileSize) - 16
							g.player.VY = 0
							g.player.OnGround = true
							g.player.JumpCount = 0
						} else if g.player.VY < 0 {
							g.player.Y = float64((ty+1)*tileSize) + 16
							g.player.VY = 0
						}
					}
				}
			}
		}
	}
}

func (g *Game) checkTileCollision(tx, ty int) bool {
	tileLeft := float64(tx * tileSize)
	tileRight := tileLeft + tileSize
	tileTop := float64(ty * tileSize)
	tileBottom := tileTop + tileSize

	playerLeft := g.player.X - 14
	playerRight := g.player.X + 14
	playerTop := g.player.Y - 14
	playerBottom := g.player.Y + 14

	return playerRight > tileLeft && playerLeft < tileRight &&
		playerBottom > tileTop && playerTop < tileBottom
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Sky gradient
	for y := range screenHeight {
		t := float64(y) / float64(screenHeight)
		r := uint8(100 + t*50)
		gCol := uint8(150 + t*50)
		b := uint8(255 - t*55)
		vector.FillRect(
			screen,
			0,
			float32(y),
			float32(screenWidth),
			1,
			color.RGBA{R: r, G: gCol, B: b, A: 255},
			false,
		)
	}

	// Draw level
	for y, row := range g.level {
		for x, tile := range row {
			screenX := float32(float64(x*tileSize) - g.cameraX)
			screenY := float32(y * tileSize)

			if screenX < -tileSize || screenX > screenWidth {
				continue
			}

			switch tile {
			case 1: // Ground
				vector.FillRect(
					screen,
					screenX,
					screenY,
					tileSize,
					tileSize,
					color.RGBA{R: 100, G: 80, B: 60, A: 255},
					false,
				)
				vector.FillRect(
					screen,
					screenX,
					screenY,
					tileSize,
					6,
					color.RGBA{R: 80, G: 180, B: 80, A: 255},
					false,
				)
			case 2: // Platform
				vector.FillRect(
					screen,
					screenX,
					screenY,
					tileSize,
					tileSize/2,
					color.RGBA{R: 139, G: 90, B: 43, A: 255},
					false,
				)
			case 4: // Goal
				vector.FillRect(
					screen,
					screenX,
					screenY,
					tileSize,
					tileSize,
					color.RGBA{R: 255, G: 215, B: 0, A: 255},
					false,
				)
				ebitenutil.DebugPrintAt(screen, "GOAL", int(screenX)+2, int(screenY)+10)
			}
		}
	}

	// Draw coins
	for _, c := range g.coins {
		if !c.Collected {
			screenX := float32(c.X - g.cameraX)
			screenY := float32(c.Y)
			vector.FillCircle(
				screen,
				screenX,
				screenY,
				10,
				color.RGBA{R: 255, G: 215, B: 0, A: 255},
				false,
			)
			vector.StrokeCircle(
				screen,
				screenX,
				screenY,
				10,
				2,
				color.RGBA{R: 200, G: 160, B: 0, A: 255},
				false,
			)
		}
	}

	// Draw player
	g.drawPlayer(screen)

	// UI
	vector.FillRect(screen, 0, 0, screenWidth, 35, color.RGBA{R: 0, G: 0, B: 0, A: 150}, false)
	ebitenutil.DebugPrintAt(screen, "Score: "+formatInt(g.score), 10, 10)
	ebitenutil.DebugPrintAt(screen, "WASD/Arrows = Move | Space = Jump (x2)", 200, 10)

	if g.won {
		vector.FillRect(
			screen,
			0,
			0,
			screenWidth,
			screenHeight,
			color.RGBA{R: 0, G: 0, B: 0, A: 150},
			false,
		)
		ebitenutil.DebugPrintAt(
			screen,
			"YOU WIN! Score: "+formatInt(g.score),
			screenWidth/2-60,
			screenHeight/2-10,
		)
		ebitenutil.DebugPrintAt(screen, "Press R to restart", screenWidth/2-60, screenHeight/2+20)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	screenX := float32(g.player.X - g.cameraX)
	screenY := float32(g.player.Y)

	// Body
	vector.FillRect(
		screen,
		screenX-12,
		screenY-14,
		24,
		28,
		color.RGBA{R: 50, G: 150, B: 255, A: 255},
		false,
	)

	// Head
	vector.FillCircle(
		screen,
		screenX,
		screenY-20,
		10,
		color.RGBA{R: 255, G: 220, B: 180, A: 255},
		false,
	)

	// Eye
	eyeOffset := float32(3)
	if g.player.FacingLeft {
		eyeOffset = -3
	}

	vector.FillCircle(
		screen,
		screenX+eyeOffset,
		screenY-22,
		3,
		color.RGBA{R: 0, G: 0, B: 0, A: 255},
		false,
	)
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}

	return result
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Platformer")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
