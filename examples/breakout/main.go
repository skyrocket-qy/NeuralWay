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
	paddleWidth  = 100
	paddleHeight = 15
	ballSize     = 10
	brickRows    = 5
	brickCols    = 10
	brickWidth   = 56
	brickHeight  = 20
	brickPadding = 4
	brickOffsetX = 20
	brickOffsetY = 60
)

type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StateGameOver
	StateVictory
)

type Brick struct {
	X, Y   float64
	Width  float64
	Height float64
	Color  color.RGBA
	Alive  bool
	Points int
}

type Ball struct {
	X, Y   float64
	VX, VY float64
	Size   float64
}

type Paddle struct {
	X, Y   float64
	Width  float64
	Height float64
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
	Text  string
	Timer float64
	Color color.RGBA
}

type Breakout struct {
	paddle     *Paddle
	ball       *Ball
	bricks     []*Brick
	particles  []Particle
	popups     []ScorePopup
	score      int
	highscore  int
	lives      int
	state      GameState
	launched   bool
	combo      int
	comboTimer float64
	level      int
	titlePulse float64
	hitFlash   float64
	trails     []struct{ X, Y, A float64 }
}

func NewBreakout() *Breakout {
	b := &Breakout{
		paddle: &Paddle{
			X:      float64(screenWidth-paddleWidth) / 2,
			Y:      float64(screenHeight) - 40,
			Width:  paddleWidth,
			Height: paddleHeight,
		},
		ball:  &Ball{Size: ballSize},
		lives: 3,
		level: 1,
		state: StateTitle,
	}

	return b
}

func (b *Breakout) startGame() {
	b.score = 0
	b.lives = 3
	b.level = 1
	b.combo = 0
	b.createBricks()
	b.resetBall()
	b.state = StatePlaying
}

func (b *Breakout) resetBall() {
	b.ball.X = float64(screenWidth)/2 - ballSize/2
	b.ball.Y = b.paddle.Y - ballSize - 5
	b.ball.VX = 0
	b.ball.VY = 0
	b.launched = false
	b.trails = nil
}

func (b *Breakout) launchBall() {
	if !b.launched {
		angle := (rand.Float64()*60 - 30) * math.Pi / 180
		speed := 5.0 + float64(b.level)*0.3
		b.ball.VX = math.Sin(angle) * speed
		b.ball.VY = -math.Cos(angle) * speed
		b.launched = true
	}
}

func (b *Breakout) createBricks() {
	colors := []color.RGBA{
		{R: 255, G: 50, B: 50, A: 255},
		{R: 255, G: 150, B: 50, A: 255},
		{R: 255, G: 255, B: 50, A: 255},
		{R: 50, G: 255, B: 50, A: 255},
		{R: 50, G: 150, B: 255, A: 255},
	}
	points := []int{50, 40, 30, 20, 10}

	b.bricks = make([]*Brick, 0, brickRows*brickCols)
	for row := range brickRows {
		for col := range brickCols {
			b.bricks = append(b.bricks, &Brick{
				X:      float64(brickOffsetX + col*(brickWidth+brickPadding)),
				Y:      float64(brickOffsetY + row*(brickHeight+brickPadding)),
				Width:  brickWidth,
				Height: brickHeight,
				Color:  colors[row],
				Points: points[row] * b.level,
				Alive:  true,
			})
		}
	}
}

func (b *Breakout) spawnBrickParticles(brick *Brick) {
	for range 15 {
		angle := rand.Float64() * math.Pi * 2
		speed := 2 + rand.Float64()*3
		b.particles = append(b.particles, Particle{
			X: brick.X + brick.Width/2, Y: brick.Y + brick.Height/2,
			VX: math.Cos(angle) * speed, VY: math.Sin(angle) * speed,
			Life: 1.0, Color: brick.Color, Size: 3 + rand.Float64()*3,
		})
	}
}

func (b *Breakout) addPopup(x, y float64, text string, clr color.RGBA) {
	b.popups = append(b.popups, ScorePopup{X: x, Y: y, Text: text, Timer: 1.0, Color: clr})
}

func (b *Breakout) Update() error {
	dt := 1.0 / 60.0
	b.titlePulse += dt * 2

	b.comboTimer -= dt
	if b.comboTimer < 0 {
		b.combo = 0
	}

	if b.hitFlash > 0 {
		b.hitFlash -= dt * 5
	}

	// Update particles
	for i := len(b.particles) - 1; i >= 0; i-- {
		p := &b.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.15

		p.Life -= dt * 2
		if p.Life <= 0 {
			b.particles = append(b.particles[:i], b.particles[i+1:]...)
		}
	}

	// Update popups
	for i := len(b.popups) - 1; i >= 0; i-- {
		b.popups[i].Timer -= dt

		b.popups[i].Y -= 30 * dt
		if b.popups[i].Timer <= 0 {
			b.popups = append(b.popups[:i], b.popups[i+1:]...)
		}
	}

	// Update trails
	for i := len(b.trails) - 1; i >= 0; i-- {
		b.trails[i].A -= dt * 4
		if b.trails[i].A <= 0 {
			b.trails = append(b.trails[:i], b.trails[i+1:]...)
		}
	}

	switch b.state {
	case StateTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			b.startGame()
		}

	case StatePlaying:
		mx, _ := ebiten.CursorPosition()
		b.paddle.X = clamp(float64(mx)-b.paddle.Width/2, 0, float64(screenWidth)-b.paddle.Width)

		if !b.launched {
			b.ball.X = b.paddle.X + b.paddle.Width/2 - b.ball.Size/2

			b.ball.Y = b.paddle.Y - b.ball.Size - 2
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
				inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				b.launchBall()
			}

			return nil
		}

		// Ball trail
		b.trails = append(
			b.trails,
			struct{ X, Y, A float64 }{b.ball.X + b.ball.Size/2, b.ball.Y + b.ball.Size/2, 0.7},
		)
		if len(b.trails) > 15 {
			b.trails = b.trails[1:]
		}

		b.ball.X += b.ball.VX
		b.ball.Y += b.ball.VY

		// Walls
		if b.ball.X <= 0 || b.ball.X+b.ball.Size >= float64(screenWidth) {
			b.ball.VX = -b.ball.VX
			b.ball.X = clamp(b.ball.X, 0, float64(screenWidth)-b.ball.Size)
		}

		if b.ball.Y <= 0 {
			b.ball.VY = -b.ball.VY
			b.ball.Y = 0
		}

		// Fall
		if b.ball.Y > float64(screenHeight) {
			b.lives--

			b.combo = 0
			if b.lives <= 0 {
				if b.score > b.highscore {
					b.highscore = b.score
				}

				b.state = StateGameOver
			} else {
				b.resetBall()
			}

			return nil
		}

		// Paddle
		if b.ball.Y+b.ball.Size >= b.paddle.Y && b.ball.Y <= b.paddle.Y+b.paddle.Height &&
			b.ball.X+b.ball.Size >= b.paddle.X && b.ball.X <= b.paddle.X+b.paddle.Width && b.ball.VY > 0 {
			hitPos := (b.ball.X + b.ball.Size/2 - b.paddle.X) / b.paddle.Width
			angle := (hitPos - 0.5) * math.Pi * 0.6
			speed := math.Sqrt(b.ball.VX*b.ball.VX + b.ball.VY*b.ball.VY)
			b.ball.VX = math.Sin(angle) * speed
			b.ball.VY = -math.Abs(math.Cos(angle) * speed)
			b.ball.Y = b.paddle.Y - b.ball.Size
			b.hitFlash = 1.0
		}

		// Bricks
		for _, brick := range b.bricks {
			if !brick.Alive {
				continue
			}

			if b.ball.X+b.ball.Size >= brick.X && b.ball.X <= brick.X+brick.Width &&
				b.ball.Y+b.ball.Size >= brick.Y && b.ball.Y <= brick.Y+brick.Height {
				brick.Alive = false
				b.combo++
				b.comboTimer = 2.0
				points := brick.Points * b.combo
				b.score += points
				b.spawnBrickParticles(brick)

				popText := fmt.Sprintf("+%d", points)
				if b.combo > 1 {
					popText = fmt.Sprintf("+%d x%d", points, b.combo)
				}

				b.addPopup(brick.X+brick.Width/2, brick.Y, popText, brick.Color)

				overlapL := b.ball.X + b.ball.Size - brick.X
				overlapR := brick.X + brick.Width - b.ball.X
				overlapT := b.ball.Y + b.ball.Size - brick.Y

				overlapB := brick.Y + brick.Height - b.ball.Y
				if math.Min(overlapL, overlapR) < math.Min(overlapT, overlapB) {
					b.ball.VX = -b.ball.VX
				} else {
					b.ball.VY = -b.ball.VY
				}

				break
			}
		}

		// Victory check
		allDead := true

		for _, brick := range b.bricks {
			if brick.Alive {
				allDead = false

				break
			}
		}

		if allDead {
			if b.score > b.highscore {
				b.highscore = b.score
			}

			b.state = StateVictory
		}

	case StateGameOver, StateVictory:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			b.startGame()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			b.state = StateTitle
		}
	}

	return nil
}

func (b *Breakout) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 10, G: 8, B: 22, A: 255})

	// Particles
	for _, p := range b.particles {
		alpha := uint8(p.Life * 255)
		c := color.RGBA{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
		vector.FillCircle(screen, float32(p.X), float32(p.Y), float32(p.Size*p.Life), c, false)
	}

	switch b.state {
	case StateTitle:
		b.drawTitle(screen)
	case StatePlaying:
		b.drawGame(screen)
	case StateGameOver:
		b.drawGame(screen)
		b.drawOverlay(screen, "GAME OVER", color.RGBA{R: 255, G: 80, B: 80, A: 255})
	case StateVictory:
		b.drawGame(screen)
		b.drawOverlay(screen, "VICTORY!", color.RGBA{R: 80, G: 255, B: 100, A: 255})
	}
}

func (b *Breakout) drawTitle(screen *ebiten.Image) {
	// Demo bricks
	for col := range 8 {
		x := float32(160 + col*40)
		y := float32(120 + math.Sin(b.titlePulse+float64(col)*0.4)*15)
		colors := []color.RGBA{
			{255, 50, 50, 255},
			{255, 150, 50, 255},
			{255, 255, 50, 255},
			{50, 255, 50, 255},
			{50, 150, 255, 255},
			{255, 50, 255, 255},
			{50, 255, 255, 255},
			{255, 255, 255, 255},
		}
		vector.FillRect(screen, x, y, 35, 18, colors[col], false)
	}

	boxW, boxH := float32(380), float32(220)
	boxX, boxY := float32(screenWidth-380)/2, float32(screenHeight-220)/2

	pulse := float32(0.7 + 0.3*math.Sin(b.titlePulse*2))
	vector.FillRect(
		screen,
		boxX-4,
		boxY-4,
		boxW+8,
		boxH+8,
		color.RGBA{R: 100, G: 200, B: 255, A: uint8(40 * pulse)},
		false,
	)
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 15, G: 20, B: 35, A: 240}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 100, G: 200, B: 255, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "B R E A K O U T", int(boxX)+110, int(boxY)+30)

	if b.highscore > 0 {
		ebitenutil.DebugPrintAt(
			screen,
			fmt.Sprintf("High Score: %d", b.highscore),
			int(boxX)+120,
			int(boxY)+65,
		)
	}

	if int(b.titlePulse*2)%2 == 0 {
		ebitenutil.DebugPrintAt(screen, "Click or SPACE to Start", int(boxX)+95, int(boxY)+105)
	}

	ebitenutil.DebugPrintAt(screen, "Move: Mouse | Launch: Click/Space", int(boxX)+60, int(boxY)+150)
	ebitenutil.DebugPrintAt(screen, "Break all bricks! Chain hits for combos!", int(boxX)+40, int(boxY)+180)
}

func (b *Breakout) drawGame(screen *ebiten.Image) {
	// Bricks
	for _, brick := range b.bricks {
		if !brick.Alive {
			continue
		}

		glow := color.RGBA{R: brick.Color.R, G: brick.Color.G, B: brick.Color.B, A: 40}
		vector.FillRect(
			screen,
			float32(brick.X)-1,
			float32(brick.Y)-1,
			float32(brick.Width)+2,
			float32(brick.Height)+2,
			glow,
			false,
		)
		vector.FillRect(
			screen,
			float32(brick.X),
			float32(brick.Y),
			float32(brick.Width),
			float32(brick.Height),
			brick.Color,
			false,
		)
	}

	// Ball trail
	for _, t := range b.trails {
		alpha := uint8(t.A * 100)
		vector.FillCircle(
			screen,
			float32(t.X),
			float32(t.Y),
			float32(ballSize/2),
			color.RGBA{R: 255, G: 255, B: 255, A: alpha},
			false,
		)
	}

	// Paddle
	paddleGlow := uint8(60)
	if b.hitFlash > 0 {
		paddleGlow = uint8(60 + b.hitFlash*100)
	}

	vector.FillRect(
		screen,
		float32(b.paddle.X)-2,
		float32(b.paddle.Y)-2,
		float32(b.paddle.Width)+4,
		float32(b.paddle.Height)+4,
		color.RGBA{R: 100, G: 200, B: 255, A: paddleGlow},
		false,
	)
	vector.FillRect(
		screen,
		float32(b.paddle.X),
		float32(b.paddle.Y),
		float32(b.paddle.Width),
		float32(b.paddle.Height),
		color.RGBA{R: 100, G: 200, B: 255, A: 255},
		false,
	)

	// Ball
	ballX, ballY := float32(b.ball.X+b.ball.Size/2), float32(b.ball.Y+b.ball.Size/2)
	vector.FillCircle(
		screen,
		ballX,
		ballY,
		float32(b.ball.Size/2)+3,
		color.RGBA{R: 255, G: 255, B: 255, A: 60},
		false,
	)
	vector.FillCircle(
		screen,
		ballX,
		ballY,
		float32(b.ball.Size/2),
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
		false,
	)

	// UI
	vector.FillRect(screen, 0, 0, screenWidth, 40, color.RGBA{R: 0, G: 0, B: 0, A: 180}, false)
	ebitenutil.DebugPrintAt(screen, "LIVES:", 10, 12)

	for i := 0; i < b.lives; i++ {
		vector.FillCircle(
			screen,
			float32(70+i*20),
			20,
			7,
			color.RGBA{R: 255, G: 60, B: 100, A: 255},
			false,
		)
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SCORE: %d", b.score), 200, 12)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HIGH: %d", b.highscore), 350, 12)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("LVL %d", b.level), screenWidth-60, 12)

	// Combo
	if b.combo > 1 && b.comboTimer > 0 {
		comboAlpha := uint8(min(int(b.comboTimer*127+128), 255))
		vector.FillRect(
			screen,
			screenWidth/2-50,
			45,
			100,
			25,
			color.RGBA{R: 255, G: 200, B: 50, A: comboAlpha / 2},
			false,
		)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("COMBO x%d!", b.combo), screenWidth/2-35, 50)
	}

	// Popups
	for _, pop := range b.popups {
		ebitenutil.DebugPrintAt(screen, pop.Text, int(pop.X)-15, int(pop.Y))
	}

	if !b.launched {
		ebitenutil.DebugPrintAt(screen, "Click or SPACE to launch", screenWidth/2-80, screenHeight/2+60)
	}
}

func (b *Breakout) drawOverlay(screen *ebiten.Image, title string, titleColor color.RGBA) {
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: 180},
		false,
	)

	boxW, boxH := float32(280), float32(150)
	boxX, boxY := float32(screenWidth-280)/2, float32(screenHeight-150)/2
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 20, G: 25, B: 40, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, titleColor, false)

	ebitenutil.DebugPrintAt(screen, title, int(boxX)+95, int(boxY)+25)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Final Score: %d", b.score), int(boxX)+85, int(boxY)+55)

	if b.score >= b.highscore && b.score > 0 {
		ebitenutil.DebugPrintAt(screen, "NEW HIGH SCORE!", int(boxX)+85, int(boxY)+80)
	}

	ebitenutil.DebugPrintAt(screen, "SPACE: Retry  ESC: Menu", int(boxX)+55, int(boxY)+115)
}

func (b *Breakout) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func clamp(v, minVal, maxVal float64) float64 {
	return math.Max(minVal, math.Min(maxVal, v))
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Breakout")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	if err := ebiten.RunGame(NewBreakout()); err != nil {
		log.Fatal(err)
	}
}
