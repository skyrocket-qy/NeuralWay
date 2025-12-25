package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 450
	screenHeight = 550
	gridSize     = 4
	tileSize     = 90
	tilePadding  = 10
	gridOffsetX  = 25
	gridOffsetY  = 120
)

type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StateGameOver
	StateWin
)

var TileColors = map[int]color.RGBA{
	0:    {R: 205, G: 193, B: 180, A: 255},
	2:    {R: 238, G: 228, B: 218, A: 255},
	4:    {R: 237, G: 224, B: 200, A: 255},
	8:    {R: 242, G: 177, B: 121, A: 255},
	16:   {R: 245, G: 149, B: 99, A: 255},
	32:   {R: 246, G: 124, B: 95, A: 255},
	64:   {R: 246, G: 94, B: 59, A: 255},
	128:  {R: 237, G: 207, B: 114, A: 255},
	256:  {R: 237, G: 204, B: 97, A: 255},
	512:  {R: 237, G: 200, B: 80, A: 255},
	1024: {R: 237, G: 197, B: 63, A: 255},
	2048: {R: 237, G: 194, B: 46, A: 255},
}

type TileAnim struct {
	Row, Col int
	Scale    float64
	Pop      bool // true = spawn pop, false = merge
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
}

type Game struct {
	grid         [gridSize][gridSize]int
	score        int
	highscore    int
	state        GameState
	moved        bool
	animations   []TileAnim
	particles    []Particle
	popups       []ScorePopup
	titlePulse   float64
	moveCount    int
	bestTile     int
	continuePlay bool // Continue after winning
}

func NewGame() *Game {
	return &Game{state: StateTitle}
}

func (g *Game) startGame() {
	for i := range gridSize {
		for j := range gridSize {
			g.grid[i][j] = 0
		}
	}

	g.score = 0
	g.moveCount = 0
	g.bestTile = 0
	g.state = StatePlaying
	g.continuePlay = false
	g.spawnTile()
	g.spawnTile()
}

func (g *Game) spawnTile() {
	empty := make([][2]int, 0)

	for i := range gridSize {
		for j := range gridSize {
			if g.grid[i][j] == 0 {
				empty = append(empty, [2]int{i, j})
			}
		}
	}

	if len(empty) == 0 {
		return
	}

	pos := empty[rand.Intn(len(empty))]

	value := 2
	if rand.Float64() < 0.1 {
		value = 4
	}

	g.grid[pos[0]][pos[1]] = value
	g.animations = append(g.animations, TileAnim{Row: pos[0], Col: pos[1], Scale: 0, Pop: true})
}

func (g *Game) spawnMergeParticles(row, col, value int) {
	x := float64(gridOffsetX + tilePadding + col*(tileSize+tilePadding) + tileSize/2)
	y := float64(gridOffsetY + tilePadding + row*(tileSize+tilePadding) + tileSize/2)

	clr := TileColors[value]
	if _, ok := TileColors[value]; !ok {
		clr = color.RGBA{R: 60, G: 58, B: 50, A: 255}
	}

	for i := range 8 {
		angle := float64(i) * math.Pi * 2 / 8
		speed := 2 + rand.Float64()*2
		g.particles = append(g.particles, Particle{
			X: x, Y: y,
			VX: math.Cos(angle) * speed, VY: math.Sin(angle) * speed,
			Life: 0.8, Color: clr, Size: 4 + rand.Float64()*3,
		})
	}
}

func (g *Game) addPopup(row, col, value int) {
	x := float64(gridOffsetX + tilePadding + col*(tileSize+tilePadding) + tileSize/2)
	y := float64(gridOffsetY + tilePadding + row*(tileSize+tilePadding))
	g.popups = append(g.popups, ScorePopup{X: x, Y: y, Value: value, Timer: 1.0})
}

func (g *Game) Update() error {
	dt := 1.0 / 60.0
	g.titlePulse += dt * 2

	// Update animations
	for i := len(g.animations) - 1; i >= 0; i-- {
		a := &g.animations[i]

		a.Scale += dt * 8
		if a.Scale >= 1.0 {
			g.animations = append(g.animations[:i], g.animations[i+1:]...)
		}
	}

	// Update particles
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.1

		p.Life -= dt * 2
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Update popups
	for i := len(g.popups) - 1; i >= 0; i-- {
		g.popups[i].Timer -= dt

		g.popups[i].Y -= 40 * dt
		if g.popups[i].Timer <= 0 {
			g.popups = append(g.popups[:i], g.popups[i+1:]...)
		}
	}

	switch g.state {
	case StateTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.startGame()
		}

	case StatePlaying:
		g.moved = false
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			g.moveLeft()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			g.moveRight()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			g.moveUp()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.moveDown()
		}

		if g.moved {
			g.moveCount++
			g.spawnTile()
			g.checkGameOver()
			g.updateBestTile()
		}

	case StateGameOver, StateWin:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if g.score > g.highscore {
				g.highscore = g.score
			}

			g.startGame()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			if g.score > g.highscore {
				g.highscore = g.score
			}

			g.state = StateTitle
		}

		if g.state == StateWin && inpututil.IsKeyJustPressed(ebiten.KeyC) {
			g.continuePlay = true
			g.state = StatePlaying
		}
	}

	return nil
}

func (g *Game) updateBestTile() {
	for i := range gridSize {
		for j := range gridSize {
			if g.grid[i][j] > g.bestTile {
				g.bestTile = g.grid[i][j]
			}
		}
	}
}

func (g *Game) moveLeft() {
	for i := range gridSize {
		newRow := g.slideAndMerge(g.grid[i][:], i, true)
		for j := range gridSize {
			if g.grid[i][j] != newRow[j] {
				g.moved = true
			}

			g.grid[i][j] = newRow[j]
		}
	}
}

func (g *Game) moveRight() {
	for i := range gridSize {
		row := make([]int, gridSize)
		for j := range gridSize {
			row[j] = g.grid[i][gridSize-1-j]
		}

		newRow := g.slideAndMerge(row, i, false)
		for j := range gridSize {
			if g.grid[i][gridSize-1-j] != newRow[j] {
				g.moved = true
			}

			g.grid[i][gridSize-1-j] = newRow[j]
		}
	}
}

func (g *Game) moveUp() {
	for j := range gridSize {
		col := make([]int, gridSize)
		for i := range gridSize {
			col[i] = g.grid[i][j]
		}

		newCol := g.slideAndMergeCol(col, j, true)
		for i := range gridSize {
			if g.grid[i][j] != newCol[i] {
				g.moved = true
			}

			g.grid[i][j] = newCol[i]
		}
	}
}

func (g *Game) moveDown() {
	for j := range gridSize {
		col := make([]int, gridSize)
		for i := range gridSize {
			col[i] = g.grid[gridSize-1-i][j]
		}

		newCol := g.slideAndMergeCol(col, j, false)
		for i := range gridSize {
			if g.grid[gridSize-1-i][j] != newCol[i] {
				g.moved = true
			}

			g.grid[gridSize-1-i][j] = newCol[i]
		}
	}
}

func (g *Game) slideAndMerge(line []int, row int, leftward bool) []int {
	nonZero := make([]int, 0)

	for _, v := range line {
		if v != 0 {
			nonZero = append(nonZero, v)
		}
	}

	merged := make([]int, 0)

	skip := false
	for i := 0; i < len(nonZero); i++ {
		if skip {
			skip = false

			continue
		}

		if i+1 < len(nonZero) && nonZero[i] == nonZero[i+1] {
			newVal := nonZero[i] * 2
			merged = append(merged, newVal)
			g.score += newVal

			col := len(merged) - 1
			if !leftward {
				col = gridSize - 1 - col
			}

			g.spawnMergeParticles(row, col, newVal)
			g.addPopup(row, col, newVal)

			if newVal == 2048 && !g.continuePlay {
				g.state = StateWin
			}

			skip = true
		} else {
			merged = append(merged, nonZero[i])
		}
	}

	result := make([]int, gridSize)
	copy(result, merged)

	return result
}

func (g *Game) slideAndMergeCol(line []int, col int, upward bool) []int {
	nonZero := make([]int, 0)

	for _, v := range line {
		if v != 0 {
			nonZero = append(nonZero, v)
		}
	}

	merged := make([]int, 0)

	skip := false
	for i := 0; i < len(nonZero); i++ {
		if skip {
			skip = false

			continue
		}

		if i+1 < len(nonZero) && nonZero[i] == nonZero[i+1] {
			newVal := nonZero[i] * 2
			merged = append(merged, newVal)
			g.score += newVal

			row := len(merged) - 1
			if !upward {
				row = gridSize - 1 - row
			}

			g.spawnMergeParticles(row, col, newVal)
			g.addPopup(row, col, newVal)

			if newVal == 2048 && !g.continuePlay {
				g.state = StateWin
			}

			skip = true
		} else {
			merged = append(merged, nonZero[i])
		}
	}

	result := make([]int, gridSize)
	copy(result, merged)

	return result
}

func (g *Game) checkGameOver() {
	for i := range gridSize {
		for j := range gridSize {
			if g.grid[i][j] == 0 {
				return
			}
		}
	}

	for i := range gridSize {
		for j := range gridSize {
			if j+1 < gridSize && g.grid[i][j] == g.grid[i][j+1] {
				return
			}

			if i+1 < gridSize && g.grid[i][j] == g.grid[i+1][j] {
				return
			}
		}
	}

	if g.score > g.highscore {
		g.highscore = g.score
	}

	g.state = StateGameOver
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 250, G: 248, B: 239, A: 255})

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
	case StateGameOver:
		g.drawGame(screen)
		g.drawOverlay(screen, "Game Over!", color.RGBA{R: 119, G: 110, B: 101, A: 220}, false)
	case StateWin:
		g.drawGame(screen)
		g.drawOverlay(screen, "You Win!", color.RGBA{R: 237, G: 194, B: 46, A: 220}, true)
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	// Animated demo tiles
	for i := range 4 {
		x := float32(90 + i*90)
		y := float32(80 + math.Sin(g.titlePulse+float64(i)*0.5)*10)
		vals := []int{2, 4, 8, 16}
		vector.FillRect(screen, x, y, 60, 60, TileColors[vals[i]], false)
		ebitenutil.DebugPrintAt(screen, strconv.Itoa(vals[i]), int(x)+22, int(y)+22)
	}

	boxW, boxH := float32(350), float32(250)
	boxX, boxY := float32(screenWidth-350)/2, float32(screenHeight-250)/2+20

	pulse := float32(0.7 + 0.3*math.Sin(g.titlePulse*2))
	vector.FillRect(
		screen,
		boxX-4,
		boxY-4,
		boxW+8,
		boxH+8,
		color.RGBA{R: 237, G: 194, B: 46, A: uint8(50 * pulse)},
		false,
	)
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 187, G: 173, B: 160, A: 250}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 119, G: 110, B: 101, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "2 0 4 8", int(boxX)+135, int(boxY)+30)

	if g.highscore > 0 {
		ebitenutil.DebugPrintAt(
			screen,
			fmt.Sprintf("Best Score: %d", g.highscore),
			int(boxX)+115,
			int(boxY)+70,
		)
	}

	if int(g.titlePulse*2)%2 == 0 {
		ebitenutil.DebugPrintAt(screen, "Press SPACE to Start", int(boxX)+90, int(boxY)+115)
	}

	ebitenutil.DebugPrintAt(screen, "Controls: Arrow Keys / WASD", int(boxX)+70, int(boxY)+165)
	ebitenutil.DebugPrintAt(screen, "Combine tiles to reach 2048!", int(boxX)+60, int(boxY)+195)
	ebitenutil.DebugPrintAt(screen, "Merge matching numbers!", int(boxX)+85, int(boxY)+220)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Header
	vector.FillRect(screen, 0, 0, screenWidth, 100, color.RGBA{R: 187, G: 173, B: 160, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "2048", 20, 15)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Moves: %d", g.moveCount), 20, 40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best Tile: %d", g.bestTile), 20, 60)

	g.drawScoreBox(screen, 200, 15, "SCORE", g.score)
	g.drawScoreBox(screen, 320, 15, "BEST", g.highscore)

	// Grid
	gridW := float32(gridSize*tileSize + (gridSize+1)*tilePadding)
	vector.FillRect(
		screen,
		float32(gridOffsetX),
		float32(gridOffsetY),
		gridW,
		gridW,
		color.RGBA{R: 187, G: 173, B: 160, A: 255},
		false,
	)

	for i := range gridSize {
		for j := range gridSize {
			g.drawTile(screen, i, j)
		}
	}

	// Popups
	for _, pop := range g.popups {
		alpha := uint8(pop.Timer * 255)
		_ = alpha

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("+%d", pop.Value), int(pop.X)-15, int(pop.Y))
	}

	ebitenutil.DebugPrintAt(screen, "Arrow Keys / WASD | R to restart", 80, screenHeight-25)
}

func (g *Game) drawScoreBox(screen *ebiten.Image, x, y int, label string, value int) {
	vector.FillRect(
		screen,
		float32(x),
		float32(y),
		100,
		60,
		color.RGBA{R: 143, G: 122, B: 102, A: 255},
		false,
	)
	ebitenutil.DebugPrintAt(screen, label, x+32, y+8)
	ebitenutil.DebugPrintAt(screen, strconv.Itoa(value), x+30, y+30)
}

func (g *Game) drawTile(screen *ebiten.Image, row, col int) {
	value := g.grid[row][col]
	x := float32(gridOffsetX + tilePadding + col*(tileSize+tilePadding))
	y := float32(gridOffsetY + tilePadding + row*(tileSize+tilePadding))

	// Check for animation
	scale := float32(1.0)

	for _, a := range g.animations {
		if a.Row == row && a.Col == col {
			scale = float32(a.Scale)

			break
		}
	}

	tileColor := TileColors[value]
	if _, ok := TileColors[value]; !ok {
		tileColor = color.RGBA{R: 60, G: 58, B: 50, A: 255}
	}

	// Draw with scale
	ts := float32(tileSize) * scale
	ox := x + float32(tileSize-ts)/2
	oy := y + float32(tileSize-ts)/2
	vector.FillRect(screen, ox, oy, ts, ts, tileColor, false)

	if value > 0 && scale >= 0.5 {
		text := strconv.Itoa(value)
		textX := int(x) + (tileSize-len(text)*8)/2
		textY := int(y) + tileSize/2 - 6
		ebitenutil.DebugPrintAt(screen, text, textX, textY)
	}
}

func (g *Game) drawOverlay(screen *ebiten.Image, msg string, bgColor color.RGBA, showContinue bool) {
	vector.FillRect(screen, 0, 0, screenWidth, screenHeight, bgColor, false)

	boxW, boxH := float32(280), float32(150)
	boxX, boxY := float32(screenWidth-280)/2, float32(screenHeight-150)/2
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2, color.RGBA{R: 119, G: 110, B: 101, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, msg, int(boxX)+95, int(boxY)+25)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.score), int(boxX)+105, int(boxY)+55)

	if showContinue {
		ebitenutil.DebugPrintAt(screen, "C: Continue  SPACE: New", int(boxX)+50, int(boxY)+90)
		ebitenutil.DebugPrintAt(screen, "ESC: Menu", int(boxX)+100, int(boxY)+115)
	} else {
		ebitenutil.DebugPrintAt(screen, "SPACE: New Game", int(boxX)+80, int(boxY)+95)
		ebitenutil.DebugPrintAt(screen, "ESC: Menu", int(boxX)+100, int(boxY)+120)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("2048")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
