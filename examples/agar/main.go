package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	worldSize    = 2000
)

// Cell represents a player or AI cell.
type Cell struct {
	X, Y   float64
	Radius float64
	Color  color.RGBA
	VX, VY float64
	IsAI   bool
	Name   string
}

// Food represents food pellets.
type Food struct {
	X, Y  float64
	Color color.RGBA
}

// Game represents the agar.io clone.
type Game struct {
	player    *Cell
	aiCells   []*Cell
	foods     []*Food
	cameraX   float64
	cameraY   float64
	score     int
	highscore int
	gameOver  bool
}

// NewGame creates a new game.
func NewGame() *Game {
	g := &Game{
		foods:   make([]*Food, 0),
		aiCells: make([]*Cell, 0),
	}
	g.reset()

	return g
}

func (g *Game) reset() {
	g.player = &Cell{
		X:      float64(worldSize) / 2,
		Y:      float64(worldSize) / 2,
		Radius: 20,
		Color:  color.RGBA{R: 50, G: 150, B: 255, A: 255},
		Name:   "Player",
	}

	g.foods = make([]*Food, 0)
	g.aiCells = make([]*Cell, 0)
	g.score = 0
	g.gameOver = false

	// Spawn initial food
	for range 200 {
		g.spawnFood()
	}

	// Spawn AI cells
	for range 10 {
		g.spawnAI()
	}
}

func (g *Game) spawnFood() {
	colors := []color.RGBA{
		{R: 255, G: 100, B: 100, A: 255},
		{R: 100, G: 255, B: 100, A: 255},
		{R: 100, G: 100, B: 255, A: 255},
		{R: 255, G: 255, B: 100, A: 255},
		{R: 255, G: 100, B: 255, A: 255},
		{R: 100, G: 255, B: 255, A: 255},
	}

	g.foods = append(g.foods, &Food{
		X:     rand.Float64() * worldSize,
		Y:     rand.Float64() * worldSize,
		Color: colors[rand.Intn(len(colors))],
	})
}

func (g *Game) spawnAI() {
	colors := []color.RGBA{
		{R: 200, G: 50, B: 50, A: 255},
		{R: 50, G: 200, B: 50, A: 255},
		{R: 200, G: 200, B: 50, A: 255},
		{R: 200, G: 50, B: 200, A: 255},
		{R: 50, G: 200, B: 200, A: 255},
	}
	names := []string{"Bot1", "Bot2", "Bot3", "Bot4", "Bot5", "Bot6", "Bot7", "Bot8"}

	g.aiCells = append(g.aiCells, &Cell{
		X:      rand.Float64() * worldSize,
		Y:      rand.Float64() * worldSize,
		Radius: 15 + rand.Float64()*30,
		Color:  colors[rand.Intn(len(colors))],
		IsAI:   true,
		Name:   names[rand.Intn(len(names))],
	})
}

func (g *Game) Update() error {
	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			if g.score > g.highscore {
				g.highscore = g.score
			}

			g.reset()
		}

		return nil
	}

	// Get mouse position relative to center
	mx, my := ebiten.CursorPosition()
	dx := float64(mx) - float64(screenWidth)/2
	dy := float64(my) - float64(screenHeight)/2

	// Normalize and apply speed (smaller = faster)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist > 0 {
		speed := 5.0 / (1 + g.player.Radius/50)
		g.player.VX = (dx / dist) * speed
		g.player.VY = (dy / dist) * speed
	}

	// Move player
	g.player.X += g.player.VX
	g.player.Y += g.player.VY

	// Keep in world bounds
	g.player.X = clamp(g.player.X, g.player.Radius, worldSize-g.player.Radius)
	g.player.Y = clamp(g.player.Y, g.player.Radius, worldSize-g.player.Radius)

	// Update camera
	g.cameraX = g.player.X - float64(screenWidth)/2
	g.cameraY = g.player.Y - float64(screenHeight)/2

	// Eat food
	for i := len(g.foods) - 1; i >= 0; i-- {
		food := g.foods[i]

		dist := math.Sqrt((g.player.X-food.X)*(g.player.X-food.X) + (g.player.Y-food.Y)*(g.player.Y-food.Y))
		if dist < g.player.Radius {
			g.player.Radius += 0.5
			g.score += 10
			g.foods = append(g.foods[:i], g.foods[i+1:]...)
			g.spawnFood()
		}
	}

	// Update AI
	for _, ai := range g.aiCells {
		// Find nearest food or smaller cell
		var targetX, targetY float64

		minDist := math.MaxFloat64

		for _, food := range g.foods {
			dist := math.Sqrt((ai.X-food.X)*(ai.X-food.X) + (ai.Y-food.Y)*(ai.Y-food.Y))
			if dist < minDist {
				minDist = dist
				targetX, targetY = food.X, food.Y
			}
		}

		// Move toward target
		if minDist < math.MaxFloat64 {
			dx := targetX - ai.X
			dy := targetY - ai.Y

			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 0 {
				speed := 3.0 / (1 + ai.Radius/50)
				ai.X += (dx / dist) * speed
				ai.Y += (dy / dist) * speed
			}
		}

		// Keep in bounds
		ai.X = clamp(ai.X, ai.Radius, worldSize-ai.Radius)
		ai.Y = clamp(ai.Y, ai.Radius, worldSize-ai.Radius)

		// AI eats food
		for i := len(g.foods) - 1; i >= 0; i-- {
			food := g.foods[i]

			dist := math.Sqrt((ai.X-food.X)*(ai.X-food.X) + (ai.Y-food.Y)*(ai.Y-food.Y))
			if dist < ai.Radius {
				ai.Radius += 0.3

				g.foods = append(g.foods[:i], g.foods[i+1:]...)
				g.spawnFood()
			}
		}
	}

	// Player eats AI or gets eaten
	for i := len(g.aiCells) - 1; i >= 0; i-- {
		ai := g.aiCells[i]
		dist := math.Sqrt((g.player.X-ai.X)*(g.player.X-ai.X) + (g.player.Y-ai.Y)*(g.player.Y-ai.Y))

		if g.player.Radius > ai.Radius*1.1 && dist < g.player.Radius {
			// Player eats AI
			g.player.Radius += ai.Radius * 0.3
			g.score += int(ai.Radius * 10)
			g.aiCells = append(g.aiCells[:i], g.aiCells[i+1:]...)
			g.spawnAI()
		} else if ai.Radius > g.player.Radius*1.1 && dist < ai.Radius {
			// AI eats player
			g.gameOver = true
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 240, G: 240, B: 245, A: 255})

	// Grid lines
	gridSpacing := 50.0
	for x := 0.0; x < worldSize; x += gridSpacing {
		screenX := x - g.cameraX
		if screenX >= -10 && screenX <= screenWidth+10 {
			vector.FillRect(
				screen,
				float32(screenX),
				0,
				1,
				screenHeight,
				color.RGBA{R: 220, G: 220, B: 225, A: 255},
				false,
			)
		}
	}

	for y := 0.0; y < worldSize; y += gridSpacing {
		screenY := y - g.cameraY
		if screenY >= -10 && screenY <= screenHeight+10 {
			vector.FillRect(
				screen,
				0,
				float32(screenY),
				screenWidth,
				1,
				color.RGBA{R: 220, G: 220, B: 225, A: 255},
				false,
			)
		}
	}

	// Draw food
	for _, food := range g.foods {
		screenX := food.X - g.cameraX

		screenY := food.Y - g.cameraY
		if screenX >= -10 && screenX <= screenWidth+10 && screenY >= -10 && screenY <= screenHeight+10 {
			vector.FillCircle(screen, float32(screenX), float32(screenY), 5, food.Color, false)
		}
	}

	// Draw AI cells
	for _, ai := range g.aiCells {
		screenX := ai.X - g.cameraX

		screenY := ai.Y - g.cameraY
		if screenX >= -float64(ai.Radius) && screenX <= float64(screenWidth)+ai.Radius &&
			screenY >= -float64(ai.Radius) && screenY <= float64(screenHeight)+ai.Radius {
			vector.FillCircle(
				screen,
				float32(screenX),
				float32(screenY),
				float32(ai.Radius),
				ai.Color,
				false,
			)
			vector.StrokeCircle(
				screen,
				float32(screenX),
				float32(screenY),
				float32(ai.Radius),
				2,
				color.RGBA{R: 0, G: 0, B: 0, A: 50},
				false,
			)
			ebitenutil.DebugPrintAt(screen, ai.Name, int(screenX)-15, int(screenY)-5)
		}
	}

	// Draw player
	if !g.gameOver {
		screenX := g.player.X - g.cameraX
		screenY := g.player.Y - g.cameraY
		vector.FillCircle(
			screen,
			float32(screenX),
			float32(screenY),
			float32(g.player.Radius),
			g.player.Color,
			false,
		)
		vector.StrokeCircle(
			screen,
			float32(screenX),
			float32(screenY),
			float32(g.player.Radius),
			3,
			color.RGBA{R: 255, G: 255, B: 255, A: 150},
			false,
		)
		ebitenutil.DebugPrintAt(screen, g.player.Name, int(screenX)-20, int(screenY)-5)
	}

	// UI
	ebitenutil.DebugPrintAt(screen, "Score: "+formatInt(g.score), 10, 10)
	ebitenutil.DebugPrintAt(screen, "Mass: "+formatInt(int(g.player.Radius)), 10, 30)
	ebitenutil.DebugPrintAt(screen, "Best: "+formatInt(g.highscore), 10, 50)

	if g.gameOver {
		vector.FillRect(
			screen,
			0,
			0,
			screenWidth,
			screenHeight,
			color.RGBA{R: 0, G: 0, B: 0, A: 150},
			false,
		)
		ebitenutil.DebugPrintAt(screen, "GAME OVER - You were eaten!", screenWidth/2-100, screenHeight/2-20)
		ebitenutil.DebugPrintAt(screen, "Score: "+formatInt(g.score), screenWidth/2-40, screenHeight/2)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to restart", screenWidth/2-80, screenHeight/2+30)
	}
}

func clamp(v, minVal, maxVal float64) float64 {
	if v < minVal {
		return minVal
	}

	if v > maxVal {
		return maxVal
	}

	return v
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
	ebiten.SetWindowTitle("Agar.io Clone")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
