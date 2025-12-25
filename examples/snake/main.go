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
	screenWidth  = 640
	screenHeight = 480
	gridSize     = 20
	gridWidth    = screenWidth / gridSize
	gridHeight   = screenHeight / gridSize
)

// GameState represents the current game state.
type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StateGameOver
)

// Direction represents movement direction.
type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

// Point represents a grid position.
type Point struct {
	X, Y int
}

// Particle represents a visual effect particle.
type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64
	Color  color.RGBA
	Size   float64
}

// Snake represents the snake game.
type Snake struct {
	body       []Point
	direction  Direction
	nextDir    Direction
	food       Point
	score      int
	highscore  int
	state      GameState
	moveTimer  float64
	moveDelay  float64
	particles  []Particle
	foodPulse  float64 // For food animation
	titlePulse float64 // For title animation
	deathTimer float64 // For death animation
}

// NewSnake creates a new snake game.
func NewSnake() *Snake {
	return &Snake{
		state:     StateTitle,
		moveDelay: 0.1,
	}
}

func (s *Snake) startGame() {
	s.body = []Point{
		{X: gridWidth / 2, Y: gridHeight / 2},
		{X: gridWidth/2 - 1, Y: gridHeight / 2},
		{X: gridWidth/2 - 2, Y: gridHeight / 2},
	}
	s.direction = DirRight
	s.nextDir = DirRight
	s.score = 0
	s.moveDelay = 0.1
	s.moveTimer = 0
	s.particles = nil
	s.deathTimer = 0
	s.state = StatePlaying
	s.spawnFood()
}

func (s *Snake) spawnFood() {
	for {
		s.food = Point{
			X: rand.Intn(gridWidth),
			Y: rand.Intn(gridHeight),
		}
		onSnake := false

		for _, p := range s.body {
			if p.X == s.food.X && p.Y == s.food.Y {
				onSnake = true

				break
			}
		}

		if !onSnake {
			break
		}
	}
}

func (s *Snake) spawnFoodParticles() {
	fx := float64(s.food.X*gridSize + gridSize/2)
	fy := float64(s.food.Y*gridSize + gridSize/2)

	for i := range 12 {
		angle := float64(i) * math.Pi * 2 / 12
		speed := 2.0 + rand.Float64()*2
		s.particles = append(s.particles, Particle{
			X: fx, Y: fy,
			VX:    math.Cos(angle) * speed,
			VY:    math.Sin(angle) * speed,
			Life:  1.0,
			Color: color.RGBA{R: 255, G: uint8(100 + rand.Intn(100)), B: 50, A: 255},
			Size:  4 + rand.Float64()*3,
		})
	}
}

func (s *Snake) spawnDeathParticles() {
	for _, p := range s.body {
		px := float64(p.X*gridSize + gridSize/2)
		py := float64(p.Y*gridSize + gridSize/2)

		for range 4 {
			angle := rand.Float64() * math.Pi * 2
			speed := 1.0 + rand.Float64()*3
			s.particles = append(s.particles, Particle{
				X: px, Y: py,
				VX:   math.Cos(angle) * speed,
				VY:   math.Sin(angle) * speed,
				Life: 1.0,
				Color: color.RGBA{
					R: uint8(50 + rand.Intn(50)),
					G: uint8(150 + rand.Intn(100)),
					B: uint8(50 + rand.Intn(50)),
					A: 255,
				},
				Size: 3 + rand.Float64()*4,
			})
		}
	}
}

func (s *Snake) Update() error {
	dt := 1.0 / 60.0
	s.foodPulse += dt * 3
	s.titlePulse += dt * 2

	// Update particles
	for i := len(s.particles) - 1; i >= 0; i-- {
		p := &s.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.1 // gravity

		p.Life -= dt * 2
		if p.Life <= 0 {
			s.particles = append(s.particles[:i], s.particles[i+1:]...)
		}
	}

	switch s.state {
	case StateTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			s.startGame()
		}

	case StatePlaying:
		// Handle input
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			if s.direction != DirDown {
				s.nextDir = DirUp
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			if s.direction != DirUp {
				s.nextDir = DirDown
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			if s.direction != DirRight {
				s.nextDir = DirLeft
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			if s.direction != DirLeft {
				s.nextDir = DirRight
			}
		}

		// Update movement timer
		s.moveTimer += dt
		if s.moveTimer >= s.moveDelay {
			s.moveTimer = 0
			s.direction = s.nextDir
			s.move()
		}

	case StateGameOver:
		s.deathTimer += dt
		if s.deathTimer > 0.5 &&
			(inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter)) {
			if s.score > s.highscore {
				s.highscore = s.score
			}

			s.startGame()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			if s.score > s.highscore {
				s.highscore = s.score
			}

			s.state = StateTitle
		}
	}

	return nil
}

func (s *Snake) move() {
	head := s.body[0]

	var newHead Point

	switch s.direction {
	case DirUp:
		newHead = Point{X: head.X, Y: head.Y - 1}
	case DirDown:
		newHead = Point{X: head.X, Y: head.Y + 1}
	case DirLeft:
		newHead = Point{X: head.X - 1, Y: head.Y}
	case DirRight:
		newHead = Point{X: head.X + 1, Y: head.Y}
	}

	// Wrap around screen edges
	if newHead.X < 0 {
		newHead.X = gridWidth - 1
	}

	if newHead.X >= gridWidth {
		newHead.X = 0
	}

	if newHead.Y < 0 {
		newHead.Y = gridHeight - 1
	}

	if newHead.Y >= gridHeight {
		newHead.Y = 0
	}

	// Check self collision
	for _, p := range s.body {
		if p.X == newHead.X && p.Y == newHead.Y {
			s.spawnDeathParticles()
			s.state = StateGameOver

			return
		}
	}

	// Move snake
	s.body = append([]Point{newHead}, s.body...)

	// Check food collision
	if newHead.X == s.food.X && newHead.Y == s.food.Y {
		s.score += 10
		s.spawnFoodParticles()
		s.spawnFood()

		if s.moveDelay > 0.05 {
			s.moveDelay -= 0.005
		}
	} else {
		s.body = s.body[:len(s.body)-1]
	}
}

func (s *Snake) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 12, G: 18, B: 30, A: 255})

	// Draw grid lines (subtle)
	gridColor := color.RGBA{R: 22, G: 28, B: 42, A: 255}
	for x := 0; x <= gridWidth; x++ {
		vector.StrokeLine(
			screen,
			float32(x*gridSize),
			0,
			float32(x*gridSize),
			screenHeight,
			1,
			gridColor,
			false,
		)
	}

	for y := 0; y <= gridHeight; y++ {
		vector.StrokeLine(
			screen,
			0,
			float32(y*gridSize),
			screenWidth,
			float32(y*gridSize),
			1,
			gridColor,
			false,
		)
	}

	// Draw particles
	for _, p := range s.particles {
		alpha := uint8(p.Life * 255)
		c := color.RGBA{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
		vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(p.Size*p.Life), c, false)
	}

	switch s.state {
	case StateTitle:
		s.drawTitle(screen)
	case StatePlaying:
		s.drawGame(screen)
	case StateGameOver:
		s.drawGame(screen)
		s.drawGameOver(screen)
	}
}

func (s *Snake) drawTitle(screen *ebiten.Image) {
	// Animated background snake
	for i := range 20 {
		x := float32(50 + i*30)
		y := float32(200 + math.Sin(s.titlePulse+float64(i)*0.3)*30)
		radius := float32(12 - i/3)
		alpha := uint8(200 - i*8)
		vector.FillCircle(
			screen,
			x,
			y,
			radius,
			color.RGBA{R: 50, G: uint8(180 - i*5), B: 50, A: alpha},
			false,
		)
	}

	// Title box
	boxW, boxH := float32(350), float32(200)
	boxX, boxY := float32(screenWidth-350)/2, float32(screenHeight-200)/2-30

	// Glow effect
	glowPulse := float32(0.7 + 0.3*math.Sin(s.titlePulse*2))
	vector.FillRect(
		screen,
		boxX-5,
		boxY-5,
		boxW+10,
		boxH+10,
		color.RGBA{R: 50, G: 200, B: 50, A: uint8(40 * glowPulse)},
		false,
	)
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 20, G: 30, B: 45, A: 240}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 80, G: 220, B: 80, A: 255}, false)

	// Title
	ebitenutil.DebugPrintAt(screen, "S N A K E", int(boxX)+130, int(boxY)+30)

	// High score
	if s.highscore > 0 {
		ebitenutil.DebugPrintAt(
			screen,
			fmt.Sprintf("High Score: %d", s.highscore),
			int(boxX)+115,
			int(boxY)+70,
		)
	}

	// Pulsing start prompt
	if int(s.titlePulse*2)%2 == 0 {
		ebitenutil.DebugPrintAt(screen, "Press SPACE to Start", int(boxX)+95, int(boxY)+110)
	}

	// Controls
	ebitenutil.DebugPrintAt(screen, "Controls: WASD or Arrow Keys", int(boxX)+70, int(boxY)+150)
	ebitenutil.DebugPrintAt(screen, "Eat food, grow longer, don't crash!", int(boxX)+50, int(boxY)+170)
}

func (s *Snake) drawGame(screen *ebiten.Image) {
	// Draw food with pulsing glow
	foodX := float32(s.food.X*gridSize + gridSize/2)
	foodY := float32(s.food.Y*gridSize + gridSize/2)
	pulseSize := float32(2 + 2*math.Sin(s.foodPulse))
	vector.FillCircle(
		screen,
		foodX,
		foodY,
		float32(gridSize/2)+pulseSize,
		color.RGBA{R: 255, G: 100, B: 100, A: 50},
		false,
	)
	vector.FillCircle(
		screen,
		foodX,
		foodY,
		float32(gridSize/2-2),
		color.RGBA{R: 255, G: 60, B: 60, A: 255},
		false,
	)

	// Draw snake
	for i, p := range s.body {
		x := float32(p.X*gridSize + gridSize/2)
		y := float32(p.Y*gridSize + gridSize/2)
		radius := float32(gridSize/2 - 1)

		if i == 0 {
			// Head with glow
			vector.FillCircle(screen, x, y, radius+3, color.RGBA{R: 100, G: 255, B: 100, A: 60}, false)
			vector.FillCircle(screen, x, y, radius, color.RGBA{R: 80, G: 220, B: 80, A: 255}, false)
			// Eyes
			ex1, ey1 := x-4, y-2
			ex2, ey2 := x+4, y-2

			vector.FillCircle(screen, ex1, ey1, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
			vector.FillCircle(screen, ex2, ey2, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
		} else {
			// Body gradient
			t := float64(i) / float64(len(s.body))
			g := uint8(200 - t*80)
			vector.FillCircle(screen, x, y, radius, color.RGBA{R: 40, G: g, B: 40, A: 255}, false)
		}
	}

	// Score panel
	vector.FillRect(screen, 5, 5, 140, 35, color.RGBA{R: 0, G: 0, B: 0, A: 180}, false)
	vector.StrokeRect(screen, 5, 5, 140, 35, 2, color.RGBA{R: 80, G: 220, B: 80, A: 200}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SCORE: %d", s.score), 15, 8)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HIGH:  %d", s.highscore), 15, 22)

	// Length indicator
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Length: %d", len(s.body)), screenWidth-90, 10)
}

func (s *Snake) drawGameOver(screen *ebiten.Image) {
	// Fade in effect
	alpha := uint8(min(int(s.deathTimer*400), 180))
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: alpha},
		false,
	)

	if s.deathTimer > 0.2 {
		boxW, boxH := float32(280), float32(150)
		boxX, boxY := float32(screenWidth-280)/2, float32(screenHeight-150)/2

		// Box with red glow
		vector.FillRect(
			screen,
			boxX-3,
			boxY-3,
			boxW+6,
			boxH+6,
			color.RGBA{R: 255, G: 50, B: 50, A: 80},
			false,
		)
		vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 25, G: 25, B: 40, A: 250}, false)
		vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2, color.RGBA{R: 255, G: 80, B: 80, A: 255}, false)

		ebitenutil.DebugPrintAt(screen, "GAME OVER", int(boxX)+95, int(boxY)+20)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Final Score: %d", s.score), int(boxX)+85, int(boxY)+50)
		ebitenutil.DebugPrintAt(
			screen,
			fmt.Sprintf("Snake Length: %d", len(s.body)),
			int(boxX)+80,
			int(boxY)+70,
		)

		if s.score > s.highscore {
			ebitenutil.DebugPrintAt(screen, "NEW HIGH SCORE!", int(boxX)+85, int(boxY)+95)
		}

		if s.deathTimer > 0.5 {
			ebitenutil.DebugPrintAt(screen, "SPACE: Retry  ESC: Menu", int(boxX)+55, int(boxY)+120)
		}
	}
}

func (s *Snake) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewSnake()); err != nil {
		log.Fatal(err)
	}
}
