package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 450
	screenHeight = 550
	gridCols     = 8
	gridRows     = 8
	cellSize     = 50
	gridOffsetX  = 25
	gridOffsetY  = 80
)

type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
)

type GemType int

const (
	GemRed GemType = iota
	GemGreen
	GemBlue
	GemYellow
	GemPurple
	GemCount
)

var GemColors = []color.RGBA{
	{R: 255, G: 60, B: 60, A: 255},
	{R: 60, G: 200, B: 60, A: 255},
	{R: 60, G: 100, B: 255, A: 255},
	{R: 255, G: 220, B: 60, A: 255},
	{R: 180, G: 60, B: 200, A: 255},
}

type Gem struct {
	Type    GemType
	X, Y    float64
	TargetY float64
	Falling bool
	Matched bool
	Scale   float64 // For pop animation
}

type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64
	Color  color.RGBA
	Size   float64
}

type ScorePopup struct {
	X, Y  float64
	Value int
	Timer float64
	Combo int
}

type Game struct {
	grid           [gridRows][gridCols]*Gem
	selectedX      int
	selectedY      int
	selected       bool
	score          int
	highscore      int
	combo          int
	maxCombo       int
	moves          int
	animating      bool
	swapping       bool
	swapX1, swapY1 int
	swapX2, swapY2 int
	swapProgress   float64
	state          GameState
	titlePulse     float64
	particles      []Particle
	popups         []ScorePopup
}

func NewGame() *Game {
	return &Game{
		selectedX: -1,
		selectedY: -1,
		state:     StateTitle,
	}
}

func (g *Game) startGame() {
	g.score = 0
	g.combo = 0
	g.maxCombo = 0
	g.moves = 0
	g.particles = nil
	g.popups = nil
	g.initGrid()
	g.state = StatePlaying
}

func (g *Game) initGrid() {
	for y := range gridRows {
		for x := range gridCols {
			g.grid[y][x] = &Gem{
				Type:    GemType(rand.Intn(int(GemCount))),
				X:       float64(x),
				Y:       float64(y),
				TargetY: float64(y),
				Scale:   1.0,
			}
		}
	}

	for g.checkAndMarkMatches() {
		g.removeMatches()
		g.fillGrid()
	}
}

func (g *Game) spawnMatchParticles(x, y int, gemType GemType) {
	px := float64(gridOffsetX + x*cellSize + cellSize/2)
	py := float64(gridOffsetY + y*cellSize + cellSize/2)
	clr := GemColors[gemType]

	for i := range 8 {
		angle := float64(i) * math.Pi * 2 / 8
		speed := 2 + rand.Float64()*3
		g.particles = append(g.particles, Particle{
			X: px, Y: py,
			VX: math.Cos(angle) * speed, VY: math.Sin(angle) * speed,
			Life: 0.8, Color: clr, Size: 4 + rand.Float64()*3,
		})
	}
}

func (g *Game) addPopup(x, y, value, combo int) {
	px := float64(gridOffsetX + x*cellSize + cellSize/2)
	py := float64(gridOffsetY + y*cellSize)
	g.popups = append(g.popups, ScorePopup{X: px, Y: py, Value: value, Timer: 1.0, Combo: combo})
}

func (g *Game) Update() error {
	dt := 1.0 / 60.0
	g.titlePulse += dt * 2

	// Update particles
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.15

		p.Life -= dt * 2.5
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Update popups
	for i := len(g.popups) - 1; i >= 0; i-- {
		g.popups[i].Timer -= dt

		g.popups[i].Y -= 35 * dt
		if g.popups[i].Timer <= 0 {
			g.popups = append(g.popups[:i], g.popups[i+1:]...)
		}
	}

	switch g.state {
	case StateTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.startGame()
		}

	case StatePlaying:
		g.updateGameplay(dt)
	}

	return nil
}

func (g *Game) updateGameplay(dt float64) {
	g.animating = false

	for y := range gridRows {
		for x := range gridCols {
			gem := g.grid[y][x]
			if gem == nil {
				continue
			}

			if gem.Falling {
				g.animating = true

				gem.Y += 10 * dt
				if gem.Y >= gem.TargetY {
					gem.Y = gem.TargetY
					gem.Falling = false
				}
			}
			// Scale animation
			if gem.Scale < 1.0 {
				gem.Scale += dt * 5
				if gem.Scale > 1.0 {
					gem.Scale = 1.0
				}
			}
		}
	}

	if g.swapping {
		g.animating = true

		g.swapProgress += dt * 5
		if g.swapProgress >= 1.0 {
			g.swapping = false

			g.swapProgress = 0
			if !g.checkAndMarkMatches() {
				g.grid[g.swapY1][g.swapX1], g.grid[g.swapY2][g.swapX2] = g.grid[g.swapY2][g.swapX2], g.grid[g.swapY1][g.swapX1]
			} else {
				g.combo = 1
				g.moves++
				g.processMatches()
			}
		}
	}

	if g.animating {
		return
	}

	// Check for cascades
	if g.checkAndMarkMatches() {
		g.combo++
		if g.combo > g.maxCombo {
			g.maxCombo = g.combo
		}

		g.processMatches()

		return
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		gridX := (mx - gridOffsetX) / cellSize
		gridY := (my - gridOffsetY) / cellSize

		if gridX >= 0 && gridX < gridCols && gridY >= 0 && gridY < gridRows {
			if !g.selected {
				g.selected = true
				g.selectedX = gridX
				g.selectedY = gridY
			} else {
				dx := abs(gridX - g.selectedX)

				dy := abs(gridY - g.selectedY)
				if (dx == 1 && dy == 0) || (dx == 0 && dy == 1) {
					g.startSwap(g.selectedX, g.selectedY, gridX, gridY)
				}

				g.selected = false
			}
		} else {
			g.selected = false
		}
	}

	// ESC to return to title
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.score > g.highscore {
			g.highscore = g.score
		}

		g.state = StateTitle
	}
}

func (g *Game) startSwap(x1, y1, x2, y2 int) {
	g.swapping = true
	g.swapX1, g.swapY1 = x1, y1
	g.swapX2, g.swapY2 = x2, y2
	g.swapProgress = 0
	g.grid[y1][x1], g.grid[y2][x2] = g.grid[y2][x2], g.grid[y1][x1]
}

func (g *Game) checkAndMarkMatches() bool {
	hasMatch := false

	for y := range gridRows {
		for x := range gridCols - 2 {
			if g.grid[y][x] == nil {
				continue
			}

			t := g.grid[y][x].Type

			count := 1
			for nx := x + 1; nx < gridCols && g.grid[y][nx] != nil && g.grid[y][nx].Type == t; nx++ {
				count++
			}

			if count >= 3 {
				hasMatch = true

				for i := 0; i < count; i++ {
					g.grid[y][x+i].Matched = true
				}
			}
		}
	}

	for x := range gridCols {
		for y := range gridRows - 2 {
			if g.grid[y][x] == nil {
				continue
			}

			t := g.grid[y][x].Type

			count := 1
			for ny := y + 1; ny < gridRows && g.grid[ny][x] != nil && g.grid[ny][x].Type == t; ny++ {
				count++
			}

			if count >= 3 {
				hasMatch = true

				for i := 0; i < count; i++ {
					g.grid[y+i][x].Matched = true
				}
			}
		}
	}

	return hasMatch
}

func (g *Game) processMatches() {
	g.removeMatches()
	g.dropGems()
	g.fillGrid()
}

func (g *Game) removeMatches() {
	for y := range gridRows {
		for x := range gridCols {
			if g.grid[y][x] != nil && g.grid[y][x].Matched {
				points := 10 * g.combo
				g.score += points
				g.spawnMatchParticles(x, y, g.grid[y][x].Type)

				if g.combo > 1 {
					g.addPopup(x, y, points, g.combo)
				}

				g.grid[y][x] = nil
			}
		}
	}
}

func (g *Game) dropGems() {
	for x := range gridCols {
		writePos := gridRows - 1
		for y := gridRows - 1; y >= 0; y-- {
			if g.grid[y][x] != nil {
				if writePos != y {
					g.grid[writePos][x] = g.grid[y][x]
					g.grid[writePos][x].TargetY = float64(writePos)
					g.grid[writePos][x].Falling = true
					g.grid[y][x] = nil
				}

				writePos--
			}
		}
	}
}

func (g *Game) fillGrid() {
	for x := range gridCols {
		emptyCount := 0

		for y := range gridRows {
			if g.grid[y][x] == nil {
				emptyCount++
			}
		}

		fillY := 0

		for y := range gridRows {
			if g.grid[y][x] == nil {
				g.grid[y][x] = &Gem{
					Type:    GemType(rand.Intn(int(GemCount))),
					X:       float64(x),
					Y:       float64(-emptyCount + fillY),
					TargetY: float64(y),
					Falling: true,
					Scale:   0.0,
				}
				fillY++
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 35, G: 25, B: 45, A: 255})

	// Particles
	for _, p := range g.particles {
		alpha := uint8(p.Life * 255)
		c := color.RGBA{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
		vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(p.Size*p.Life), c, false)
	}

	switch g.state {
	case StateTitle:
		g.drawTitle(screen)
	case StatePlaying:
		g.drawGame(screen)
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	// Animated gems
	for i := range 5 {
		x := float32(100 + i*60)
		y := float32(60 + math.Sin(g.titlePulse+float64(i)*0.5)*15)
		vector.FillCircle(screen, x, y, 18, GemColors[i], false)
		vector.FillCircle(screen, x-4, y-4, 5, color.RGBA{R: 255, G: 255, B: 255, A: 80}, false)
	}

	boxW, boxH := float32(350), float32(250)
	boxX, boxY := float32(screenWidth-350)/2, float32(screenHeight-250)/2

	pulse := float32(0.7 + 0.3*math.Sin(g.titlePulse*2))
	vector.FillRect(
		screen,
		boxX-4,
		boxY-4,
		boxW+8,
		boxH+8,
		color.RGBA{R: 180, G: 60, B: 200, A: uint8(50 * pulse)},
		false,
	)
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 50, G: 40, B: 60, A: 245}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 180, G: 60, B: 200, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "M A T C H  3", int(boxX)+110, int(boxY)+30)

	if g.highscore > 0 {
		ebitenutil.DebugPrintAt(
			screen,
			fmt.Sprintf("Best Score: %d", g.highscore),
			int(boxX)+115,
			int(boxY)+70,
		)
	}

	if int(g.titlePulse*2)%2 == 0 {
		ebitenutil.DebugPrintAt(screen, "Click or SPACE to Start", int(boxX)+80, int(boxY)+110)
	}

	ebitenutil.DebugPrintAt(screen, "Click two adjacent gems to swap", int(boxX)+55, int(boxY)+160)
	ebitenutil.DebugPrintAt(screen, "Match 3+ of the same color!", int(boxX)+65, int(boxY)+190)
	ebitenutil.DebugPrintAt(screen, "Chain matches for combo bonus!", int(boxX)+55, int(boxY)+220)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Header
	vector.FillRect(screen, 0, 0, screenWidth, 70, color.RGBA{R: 55, G: 45, B: 65, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "Match 3", 15, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.score), 15, 30)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Moves: %d", g.moves), 15, 50)

	if g.combo > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("COMBO x%d!", g.combo), 150, 30)
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best Combo: %d", g.maxCombo), 280, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("High: %d", g.highscore), 280, 30)

	// Grid background
	vector.FillRect(screen, float32(gridOffsetX), float32(gridOffsetY),
		float32(gridCols*cellSize), float32(gridRows*cellSize),
		color.RGBA{R: 25, G: 20, B: 35, A: 255}, false)

	// Gems
	for y := range gridRows {
		for x := range gridCols {
			gem := g.grid[y][x]
			if gem == nil {
				continue
			}

			drawX := float32(gridOffsetX + x*cellSize + cellSize/2)
			drawY := float32(gridOffsetY + int(gem.Y*float64(cellSize)) + cellSize/2)

			if g.selected && x == g.selectedX && y == g.selectedY {
				vector.FillRect(
					screen,
					drawX-float32(cellSize/2)+2,
					drawY-float32(cellSize/2)+2,
					float32(
						cellSize-4,
					),
					float32(cellSize-4),
					color.RGBA{R: 255, G: 255, B: 255, A: 100},
					false,
				)
			}

			radius := float32(cellSize/2-4) * float32(gem.Scale)
			gemColor := GemColors[gem.Type]
			vector.FillCircle(screen, drawX, drawY, radius, gemColor, false)

			if gem.Scale >= 0.8 {
				vector.FillCircle(
					screen,
					drawX-5,
					drawY-5,
					5,
					color.RGBA{R: 255, G: 255, B: 255, A: 80},
					false,
				)
			}
		}
	}

	// Grid lines
	for i := 0; i <= gridCols; i++ {
		x := float32(gridOffsetX + i*cellSize)
		vector.FillRect(
			screen,
			x,
			float32(gridOffsetY),
			1,
			float32(gridRows*cellSize),
			color.RGBA{R: 55, G: 45, B: 65, A: 255},
			false,
		)
	}

	for i := 0; i <= gridRows; i++ {
		y := float32(gridOffsetY + i*cellSize)
		vector.FillRect(
			screen,
			float32(gridOffsetX),
			y,
			float32(gridCols*cellSize),
			1,
			color.RGBA{R: 55, G: 45, B: 65, A: 255},
			false,
		)
	}

	// Popups
	for _, pop := range g.popups {
		text := fmt.Sprintf("+%d", pop.Value)
		if pop.Combo > 1 {
			text = fmt.Sprintf("+%d x%d", pop.Value, pop.Combo)
		}

		ebitenutil.DebugPrintAt(screen, text, int(pop.X)-15, int(pop.Y))
	}

	ebitenutil.DebugPrintAt(screen, "Click gems to swap | ESC: Menu", 85, screenHeight-25)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Match 3")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
