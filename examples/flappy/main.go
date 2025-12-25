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
	screenWidth  = 400
	screenHeight = 600
	gravity      = 0.5
	jumpForce    = -8
	pipeWidth    = 60
	pipeGap      = 150
	pipeSpeed    = 3
	birdSize     = 30
)

type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StateGameOver
)

type Bird struct {
	X, Y      float64
	VelocityY float64
	Rotation  float64
}

type Pipe struct {
	X      float64
	GapY   float64
	Passed bool
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
	Timer float64
}

type Game struct {
	bird       *Bird
	pipes      []*Pipe
	particles  []Particle
	popups     []ScorePopup
	score      int
	highscore  int
	state      GameState
	pipeTimer  float64
	titlePulse float64
	deathTimer float64
	groundX    float64 // For scrolling ground
}

func NewGame() *Game {
	return &Game{
		bird:  &Bird{X: 100, Y: float64(screenHeight) / 2},
		pipes: make([]*Pipe, 0),
		state: StateTitle,
	}
}

func (g *Game) startGame() {
	g.bird = &Bird{X: 100, Y: float64(screenHeight) / 2}
	g.pipes = make([]*Pipe, 0)
	g.particles = nil
	g.popups = nil
	g.score = 0
	g.pipeTimer = 0
	g.deathTimer = 0
	g.bird.VelocityY = jumpForce
	g.state = StatePlaying
}

func (g *Game) spawnPipe() {
	minGap := float64(pipeGap/2 + 50)
	maxGap := float64(screenHeight - pipeGap/2 - 100)
	gapY := minGap + rand.Float64()*(maxGap-minGap)
	g.pipes = append(g.pipes, &Pipe{X: float64(screenWidth), GapY: gapY})
}

func (g *Game) spawnDeathParticles() {
	for range 20 {
		angle := rand.Float64() * math.Pi * 2
		speed := 2 + rand.Float64()*4
		g.particles = append(g.particles, Particle{
			X: g.bird.X + birdSize/2, Y: g.bird.Y + birdSize/2,
			VX: math.Cos(angle) * speed, VY: math.Sin(angle) * speed,
			Life: 1.0, Color: color.RGBA{R: 255, G: uint8(180 + rand.Intn(75)), B: 50, A: 255},
			Size: 4 + rand.Float64()*4,
		})
	}
}

func (g *Game) addScorePopup() {
	g.popups = append(g.popups, ScorePopup{X: g.bird.X + 40, Y: g.bird.Y - 20, Timer: 1.0})
}

func (g *Game) Update() error {
	dt := 1.0 / 60.0
	g.titlePulse += dt * 2

	g.groundX -= pipeSpeed
	if g.groundX <= -40 {
		g.groundX = 0
	}

	// Update particles
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.15

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
		// Hover animation
		g.bird.Y = float64(screenHeight)/2 + math.Sin(g.titlePulse*3)*15
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.startGame()
		}

	case StatePlaying:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.bird.VelocityY = jumpForce
		}

		g.bird.VelocityY += gravity
		g.bird.Y += g.bird.VelocityY

		g.bird.Rotation = g.bird.VelocityY * 3
		if g.bird.Rotation > 90 {
			g.bird.Rotation = 90
		}

		if g.bird.Rotation < -30 {
			g.bird.Rotation = -30
		}

		// Ground/ceiling
		if g.bird.Y < 0 || g.bird.Y > float64(screenHeight-50-birdSize) {
			g.spawnDeathParticles()

			if g.score > g.highscore {
				g.highscore = g.score
			}

			g.state = StateGameOver

			return nil
		}

		// Pipes
		g.pipeTimer += dt
		if g.pipeTimer >= 1.5 {
			g.spawnPipe()
			g.pipeTimer = 0
		}

		for i := len(g.pipes) - 1; i >= 0; i-- {
			pipe := g.pipes[i]
			pipe.X -= pipeSpeed

			if pipe.X < -pipeWidth {
				g.pipes = append(g.pipes[:i], g.pipes[i+1:]...)

				continue
			}

			if !pipe.Passed && pipe.X+pipeWidth < g.bird.X {
				pipe.Passed = true
				g.score++
				g.addScorePopup()
			}

			if g.checkCollision(pipe) {
				g.spawnDeathParticles()

				if g.score > g.highscore {
					g.highscore = g.score
				}

				g.state = StateGameOver
			}
		}

	case StateGameOver:
		g.deathTimer += dt
		// Bird falls
		g.bird.VelocityY += gravity

		g.bird.Y += g.bird.VelocityY
		if g.bird.Y > float64(screenHeight-50-birdSize) {
			g.bird.Y = float64(screenHeight - 50 - birdSize)
			g.bird.VelocityY = 0
		}

		if g.deathTimer > 0.5 &&
			(inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)) {
			g.startGame()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StateTitle
			g.bird = &Bird{X: 100, Y: float64(screenHeight) / 2}
			g.pipes = nil
		}
	}

	return nil
}

func (g *Game) checkCollision(pipe *Pipe) bool {
	birdLeft, birdRight := g.bird.X, g.bird.X+birdSize
	birdTop, birdBottom := g.bird.Y, g.bird.Y+birdSize
	pipeLeft, pipeRight := pipe.X, pipe.X+pipeWidth
	gapTop, gapBottom := pipe.GapY-pipeGap/2, pipe.GapY+pipeGap/2

	if birdRight > pipeLeft && birdLeft < pipeRight {
		if birdTop < gapTop || birdBottom > gapBottom {
			return true
		}
	}

	return false
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Sky gradient
	for y := range screenHeight - 50 {
		t := float64(y) / float64(screenHeight-50)
		r := uint8(100 - t*50)
		gr := uint8(180 - t*80)
		b := uint8(255 - t*55)
		vector.FillRect(
			screen,
			0,
			float32(y),
			float32(screenWidth),
			1,
			color.RGBA{R: r, G: gr, B: b, A: 255},
			false,
		)
	}

	// Pipes
	for _, pipe := range g.pipes {
		g.drawPipe(screen, pipe)
	}

	// Particles
	for _, p := range g.particles {
		alpha := uint8(p.Life * 255)
		c := color.RGBA{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
		vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(p.Size*p.Life), c, false)
	}

	// Ground (scrolling)
	vector.FillRect(
		screen,
		0,
		float32(screenHeight-50),
		float32(screenWidth),
		50,
		color.RGBA{R: 139, G: 119, B: 101, A: 255},
		false,
	)
	vector.FillRect(
		screen,
		0,
		float32(screenHeight-50),
		float32(screenWidth),
		5,
		color.RGBA{R: 34, G: 139, B: 34, A: 255},
		false,
	)
	// Grass pattern
	for x := int(g.groundX); x < screenWidth+40; x += 20 {
		vector.FillRect(
			screen,
			float32(x),
			float32(screenHeight-48),
			2,
			8,
			color.RGBA{R: 50, G: 160, B: 50, A: 255},
			false,
		)
	}

	// Bird
	g.drawBird(screen)

	// Popups
	for _, pop := range g.popups {
		ebitenutil.DebugPrintAt(screen, "+1", int(pop.X), int(pop.Y))
	}

	switch g.state {
	case StateTitle:
		g.drawTitle(screen)
	case StatePlaying:
		g.drawHUD(screen)
	case StateGameOver:
		g.drawHUD(screen)
		g.drawGameOver(screen)
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	boxW, boxH := float32(300), float32(220)
	boxX, boxY := float32(screenWidth-300)/2, float32(screenHeight-220)/2

	pulse := float32(0.7 + 0.3*math.Sin(g.titlePulse*2))
	vector.FillRect(
		screen,
		boxX-4,
		boxY-4,
		boxW+8,
		boxH+8,
		color.RGBA{R: 255, G: 220, B: 50, A: uint8(50 * pulse)},
		false,
	)
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 50, G: 50, B: 70, A: 240}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 255, G: 220, B: 50, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "FLAPPY BIRD", int(boxX)+95, int(boxY)+25)

	if g.highscore > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best: %d", g.highscore), int(boxX)+120, int(boxY)+60)
	}

	if int(g.titlePulse*2)%2 == 0 {
		ebitenutil.DebugPrintAt(screen, "Tap or SPACE to Start", int(boxX)+70, int(boxY)+100)
	}

	ebitenutil.DebugPrintAt(screen, "Tap/Click or SPACE to flap", int(boxX)+50, int(boxY)+150)
	ebitenutil.DebugPrintAt(screen, "Avoid the pipes!", int(boxX)+90, int(boxY)+180)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Score with shadow
	scoreText := strconv.Itoa(g.score)
	ebitenutil.DebugPrintAt(screen, scoreText, screenWidth/2-len(scoreText)*3+1, 31)
	ebitenutil.DebugPrintAt(screen, scoreText, screenWidth/2-len(scoreText)*3, 30)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	alpha := uint8(min(int(g.deathTimer*300), 180))
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: alpha},
		false,
	)

	if g.deathTimer > 0.3 {
		boxW, boxH := float32(280), float32(160)
		boxX, boxY := float32(screenWidth-280)/2, float32(screenHeight-160)/2
		vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 40, G: 40, B: 60, A: 250}, false)
		vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 255, G: 80, B: 80, A: 255}, false)

		ebitenutil.DebugPrintAt(screen, "GAME OVER", int(boxX)+95, int(boxY)+25)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.score), int(boxX)+105, int(boxY)+55)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best: %d", g.highscore), int(boxX)+110, int(boxY)+80)

		if g.score == g.highscore && g.score > 0 {
			ebitenutil.DebugPrintAt(screen, "NEW BEST!", int(boxX)+100, int(boxY)+105)
		}

		if g.deathTimer > 0.5 {
			ebitenutil.DebugPrintAt(screen, "Tap: Retry  ESC: Menu", int(boxX)+60, int(boxY)+135)
		}
	}
}

func (g *Game) drawBird(screen *ebiten.Image) {
	// Body
	vector.FillCircle(
		screen,
		float32(g.bird.X+birdSize/2),
		float32(g.bird.Y+birdSize/2),
		birdSize/2,
		color.RGBA{R: 255, G: 220, B: 50, A: 255},
		false,
	)
	// Eye
	vector.FillCircle(
		screen,
		float32(g.bird.X+birdSize*0.7),
		float32(g.bird.Y+birdSize*0.3),
		5,
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
		false,
	)
	vector.FillCircle(
		screen,
		float32(g.bird.X+birdSize*0.75),
		float32(g.bird.Y+birdSize*0.35),
		2,
		color.RGBA{R: 0, G: 0, B: 0, A: 255},
		false,
	)
	// Beak
	vector.FillRect(
		screen,
		float32(g.bird.X+birdSize*0.8),
		float32(g.bird.Y+birdSize*0.45),
		10,
		6,
		color.RGBA{R: 255, G: 150, B: 0, A: 255},
		false,
	)
	// Wing
	wingY := g.bird.Y + birdSize*0.5
	if g.bird.VelocityY < 0 {
		wingY -= 5
	}

	vector.FillCircle(
		screen,
		float32(g.bird.X+birdSize*0.3),
		float32(wingY),
		8,
		color.RGBA{R: 255, G: 180, B: 50, A: 255},
		false,
	)
}

func (g *Game) drawPipe(screen *ebiten.Image, pipe *Pipe) {
	pipeColor := color.RGBA{R: 50, G: 180, B: 50, A: 255}
	pipeEdge := color.RGBA{R: 30, G: 140, B: 30, A: 255}
	pipeHighlight := color.RGBA{R: 80, G: 220, B: 80, A: 255}

	gapTop := pipe.GapY - pipeGap/2
	gapBottom := pipe.GapY + pipeGap/2

	// Top pipe
	vector.FillRect(screen, float32(pipe.X), 0, float32(pipeWidth), float32(gapTop), pipeColor, false)
	vector.FillRect(screen, float32(pipe.X), 0, 5, float32(gapTop), pipeHighlight, false)
	vector.FillRect(
		screen,
		float32(pipe.X-5),
		float32(gapTop-30),
		float32(pipeWidth+10),
		30,
		pipeEdge,
		false,
	)

	// Bottom pipe
	vector.FillRect(
		screen,
		float32(pipe.X),
		float32(gapBottom),
		float32(pipeWidth),
		float32(screenHeight-50)-float32(gapBottom),
		pipeColor,
		false,
	)
	vector.FillRect(
		screen,
		float32(pipe.X),
		float32(gapBottom),
		5,
		float32(screenHeight-50)-float32(gapBottom),
		pipeHighlight,
		false,
	)
	vector.FillRect(
		screen,
		float32(pipe.X-5),
		float32(gapBottom),
		float32(pipeWidth+10),
		30,
		pipeEdge,
		false,
	)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Flappy Bird")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
