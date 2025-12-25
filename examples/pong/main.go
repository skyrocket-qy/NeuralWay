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
	screenWidth  = 640
	screenHeight = 480
	paddleWidth  = 15
	paddleHeight = 80
	ballSize     = 12
	paddleSpeed  = 6.0
	ballSpeed    = 5.0
)

// GameState for Pong.
type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
)

// Paddle represents a player paddle.
type Paddle struct {
	X, Y   float64
	Width  float64
	Height float64
	Score  int
}

// Ball represents the game ball.
type Ball struct {
	X, Y   float64
	VX, VY float64
	Size   float64
}

// Trail for ball effect.
type Trail struct {
	X, Y  float64
	Alpha float64
}

// ScorePopup for score animation.
type ScorePopup struct {
	X, Y  float64
	Text  string
	Timer float64
	Color color.RGBA
}

// Pong represents the pong game.
type Pong struct {
	player1    *Paddle
	player2    *Paddle
	ball       *Ball
	state      GameState
	winScore   int
	trails     []Trail
	popups     []ScorePopup
	serveTimer float64
	serving    int // 1 or 2 for which player serves
	titlePulse float64
	hitFlash   float64 // For paddle hit flash
	// AI Mode
	aiMode       bool    // true = vs AI, false = vs Player
	aiDifficulty int     // 0=Easy, 1=Medium, 2=Hard
	aiReactDelay float64 // Delay before AI reacts to ball
	aiTargetY    float64 // Where AI wants the paddle to be
}

// NewPong creates a new pong game.
func NewPong() *Pong {
	return &Pong{
		player1: &Paddle{
			X:      30,
			Y:      float64(screenHeight-paddleHeight) / 2,
			Width:  paddleWidth,
			Height: paddleHeight,
		},
		player2: &Paddle{
			X:      float64(screenWidth) - 30 - paddleWidth,
			Y:      float64(screenHeight-paddleHeight) / 2,
			Width:  paddleWidth,
			Height: paddleHeight,
		},
		ball: &Ball{
			Size: ballSize,
		},
		state:    StateTitle,
		winScore: 5,
		serving:  1,
	}
}

func (p *Pong) startGame() {
	p.player1.Score = 0
	p.player2.Score = 0
	p.serving = 1
	p.serveTimer = 1.0
	p.resetBall(0) // Ball stationary for serve
	p.state = StatePlaying
}

func (p *Pong) resetBall(direction float64) {
	p.ball.X = float64(screenWidth) / 2
	p.ball.Y = float64(screenHeight) / 2
	p.ball.VX = ballSpeed * direction
	p.ball.VY = ballSpeed * 0.5 * direction
	p.trails = nil
}

func (p *Pong) addPopup(x, y float64, text string, clr color.RGBA) {
	p.popups = append(p.popups, ScorePopup{
		X: x, Y: y, Text: text, Timer: 1.0, Color: clr,
	})
}

// updateAI controls the AI paddle with difficulty-based behavior.
func (p *Pong) updateAI(dt float64) {
	// AI only reacts when ball is moving towards it
	if p.ball.VX > 0 {
		// Predict where ball will be when it reaches AI paddle
		timeToReach := (p.player2.X - p.ball.X) / p.ball.VX
		predictedY := p.ball.Y + p.ball.VY*timeToReach

		// Add imperfection based on difficulty
		var (
			errorRange    float64
			reactionSpeed float64
		)

		switch p.aiDifficulty {
		case 0: // Easy - slow, imprecise
			errorRange = 60
			reactionSpeed = paddleSpeed * 0.5
		case 1: // Medium - moderate
			errorRange = 30
			reactionSpeed = paddleSpeed * 0.75
		case 2: // Hard - fast, precise
			errorRange = 10
			reactionSpeed = paddleSpeed * 0.95
		}

		// Add some randomness to target (AI isn't perfect)
		if p.aiReactDelay <= 0 {
			p.aiTargetY = predictedY + (rand.Float64()-0.5)*errorRange
			p.aiReactDelay = 0.1 + rand.Float64()*0.1 // Delay before next update
		}

		p.aiReactDelay -= dt

		// Move paddle towards target
		paddleCenter := p.player2.Y + p.player2.Height/2
		diff := p.aiTargetY - paddleCenter

		if math.Abs(diff) > 5 {
			if diff > 0 {
				p.player2.Y += reactionSpeed
			} else {
				p.player2.Y -= reactionSpeed
			}
		}
	} else {
		// Ball moving away - return to center slowly
		centerY := float64(screenHeight)/2 - p.player2.Height/2
		if p.player2.Y < centerY-10 {
			p.player2.Y += paddleSpeed * 0.3
		} else if p.player2.Y > centerY+10 {
			p.player2.Y -= paddleSpeed * 0.3
		}
	}
}

func (p *Pong) Update() error {
	dt := 1.0 / 60.0
	p.titlePulse += dt * 2

	// Update trails
	for i := len(p.trails) - 1; i >= 0; i-- {
		p.trails[i].Alpha -= dt * 3
		if p.trails[i].Alpha <= 0 {
			p.trails = append(p.trails[:i], p.trails[i+1:]...)
		}
	}

	// Update popups
	for i := len(p.popups) - 1; i >= 0; i-- {
		p.popups[i].Timer -= dt

		p.popups[i].Y -= dt * 30
		if p.popups[i].Timer <= 0 {
			p.popups = append(p.popups[:i], p.popups[i+1:]...)
		}
	}

	// Hit flash decay
	if p.hitFlash > 0 {
		p.hitFlash -= dt * 5
	}

	switch p.state {
	case StateTitle:
		// Mode selection: 1 = vs AI, 2 = 2 Players
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			p.aiMode = true
			p.aiDifficulty = 1 // Medium by default
			p.startGame()
		}

		if inpututil.IsKeyJustPressed(ebiten.Key2) {
			p.aiMode = false
			p.startGame()
		}
		// Difficulty selection for AI mode
		if inpututil.IsKeyJustPressed(ebiten.KeyE) {
			p.aiMode = true
			p.aiDifficulty = 0 // Easy
			p.startGame()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			p.aiMode = true
			p.aiDifficulty = 2 // Hard
			p.startGame()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			p.aiMode = true
			p.aiDifficulty = 1
			p.startGame()
		}

	case StatePlaying:
		// Serve countdown
		if p.serveTimer > 0 {
			p.serveTimer -= dt
			if p.serveTimer <= 0 {
				dir := 1.0
				if p.serving == 2 {
					dir = -1.0
				}

				p.ball.VX = ballSpeed * dir
				p.ball.VY = (rand.Float64() - 0.5) * ballSpeed
			}

			return nil
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			p.state = StatePaused
		}

		// Player 1 controls (always human)
		if ebiten.IsKeyPressed(ebiten.KeyW) {
			p.player1.Y -= paddleSpeed
		}

		if ebiten.IsKeyPressed(ebiten.KeyS) {
			p.player1.Y += paddleSpeed
		}

		// Player 2 controls (human or AI)
		if p.aiMode {
			// AI opponent logic
			p.updateAI(dt)
		} else {
			// Human player 2
			if ebiten.IsKeyPressed(ebiten.KeyUp) {
				p.player2.Y -= paddleSpeed
			}

			if ebiten.IsKeyPressed(ebiten.KeyDown) {
				p.player2.Y += paddleSpeed
			}
		}

		// Clamp paddles
		p.player1.Y = clamp(p.player1.Y, 0, float64(screenHeight)-p.player1.Height)
		p.player2.Y = clamp(p.player2.Y, 0, float64(screenHeight)-p.player2.Height)

		// Add ball trail
		p.trails = append(
			p.trails,
			Trail{X: p.ball.X + p.ball.Size/2, Y: p.ball.Y + p.ball.Size/2, Alpha: 0.8},
		)
		if len(p.trails) > 20 {
			p.trails = p.trails[1:]
		}

		// Update ball
		p.ball.X += p.ball.VX
		p.ball.Y += p.ball.VY

		// Wall collision
		if p.ball.Y <= 0 || p.ball.Y+p.ball.Size >= float64(screenHeight) {
			p.ball.VY = -p.ball.VY
			p.ball.Y = clamp(p.ball.Y, 0, float64(screenHeight)-p.ball.Size)
		}

		// Paddle collision
		if p.ball.X <= p.player1.X+p.player1.Width &&
			p.ball.Y+p.ball.Size >= p.player1.Y &&
			p.ball.Y <= p.player1.Y+p.player1.Height &&
			p.ball.VX < 0 {
			p.ball.VX = -p.ball.VX * 1.05
			relativeY := (p.ball.Y + p.ball.Size/2 - p.player1.Y) / p.player1.Height
			p.ball.VY = (relativeY - 0.5) * ballSpeed * 2
			p.hitFlash = 1.0
		}

		if p.ball.X+p.ball.Size >= p.player2.X &&
			p.ball.Y+p.ball.Size >= p.player2.Y &&
			p.ball.Y <= p.player2.Y+p.player2.Height &&
			p.ball.VX > 0 {
			p.ball.VX = -p.ball.VX * 1.05
			relativeY := (p.ball.Y + p.ball.Size/2 - p.player2.Y) / p.player2.Height
			p.ball.VY = (relativeY - 0.5) * ballSpeed * 2
			p.hitFlash = 1.0
		}

		// Clamp speed
		maxSpeed := 12.0
		p.ball.VX = clamp(p.ball.VX, -maxSpeed, maxSpeed)
		p.ball.VY = clamp(p.ball.VY, -maxSpeed, maxSpeed)

		// Scoring
		if p.ball.X < 0 {
			p.player2.Score++
			p.addPopup(float64(screenWidth)*3/4, 100, "+1", color.RGBA{R: 255, G: 100, B: 100, A: 255})
			p.serving = 1
			p.serveTimer = 1.0
			p.resetBall(0)

			if p.player2.Score >= p.winScore {
				p.state = StateGameOver
			}
		}

		if p.ball.X > float64(screenWidth) {
			p.player1.Score++
			p.addPopup(float64(screenWidth)/4, 100, "+1", color.RGBA{R: 100, G: 150, B: 255, A: 255})
			p.serving = 2
			p.serveTimer = 1.0
			p.resetBall(0)

			if p.player1.Score >= p.winScore {
				p.state = StateGameOver
			}
		}

	case StatePaused:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			p.state = StatePlaying
		}

	case StateGameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			p.startGame()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			p.state = StateTitle
		}
	}

	return nil
}

func (p *Pong) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 8, G: 12, B: 25, A: 255})

	switch p.state {
	case StateTitle:
		p.drawTitle(screen)
	case StatePlaying:
		p.drawGame(screen)
	case StatePaused:
		p.drawGame(screen)
		p.drawPaused(screen)
	case StateGameOver:
		p.drawGame(screen)
		p.drawGameOver(screen)
	}
}

func (p *Pong) drawTitle(screen *ebiten.Image) {
	// Animated ball bouncing
	ballY := float32(screenHeight/2 + math.Sin(p.titlePulse*3)*50)
	vector.FillCircle(
		screen,
		screenWidth/2,
		ballY,
		15,
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
		false,
	)

	// Title box
	boxW, boxH := float32(380), float32(220)
	boxX, boxY := float32(screenWidth-380)/2, float32(screenHeight-220)/2-20

	pulse := float32(0.7 + 0.3*math.Sin(p.titlePulse*2))
	vector.FillRect(
		screen,
		boxX-4,
		boxY-4,
		boxW+8,
		boxH+8,
		color.RGBA{R: 100, G: 100, B: 255, A: uint8(40 * pulse)},
		false,
	)
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 20, G: 25, B: 40, A: 240}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 150, G: 150, B: 255, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "P O N G", int(boxX)+150, int(boxY)+30)

	// Game mode selection
	if int(p.titlePulse*2)%2 == 0 {
		ebitenutil.DebugPrintAt(screen, "Select Mode:", int(boxX)+135, int(boxY)+65)
	}

	ebitenutil.DebugPrintAt(screen, "[1] vs AI   [2] 2 Players", int(boxX)+85, int(boxY)+90)
	ebitenutil.DebugPrintAt(screen, "[E]asy  [SPACE]Medium  [H]ard", int(boxX)+60, int(boxY)+115)

	ebitenutil.DebugPrintAt(screen, "P1: W/S    P2: Up/Down", int(boxX)+90, int(boxY)+155)
	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf("First to %d wins!", p.winScore),
		int(boxX)+120,
		int(boxY)+180,
	)
}

func (p *Pong) drawGame(screen *ebiten.Image) {
	// Center line
	for y := 0; y < screenHeight; y += 25 {
		vector.FillRect(
			screen,
			float32(screenWidth)/2-2,
			float32(y),
			4,
			15,
			color.RGBA{R: 40, G: 50, B: 70, A: 255},
			false,
		)
	}

	// Ball trail
	for _, t := range p.trails {
		alpha := uint8(t.Alpha * 100)
		vector.FillCircle(
			screen,
			float32(t.X),
			float32(t.Y),
			float32(ballSize/2),
			color.RGBA{R: 255, G: 255, B: 255, A: alpha},
			false,
		)
	}

	// Paddles with hit flash
	p1Glow := uint8(60)
	p2Glow := uint8(60)

	if p.hitFlash > 0 && p.ball.VX > 0 {
		p1Glow = uint8(60 + p.hitFlash*100)
	}

	if p.hitFlash > 0 && p.ball.VX < 0 {
		p2Glow = uint8(60 + p.hitFlash*100)
	}

	vector.FillRect(
		screen,
		float32(p.player1.X)-2,
		float32(p.player1.Y)-2,
		float32(p.player1.Width)+4,
		float32(p.player1.Height)+4,
		color.RGBA{R: 80, G: 130, B: 255, A: p1Glow},
		false,
	)
	vector.FillRect(
		screen,
		float32(p.player1.X),
		float32(p.player1.Y),
		float32(p.player1.Width),
		float32(p.player1.Height),
		color.RGBA{R: 100, G: 150, B: 255, A: 255},
		false,
	)

	vector.FillRect(
		screen,
		float32(p.player2.X)-2,
		float32(p.player2.Y)-2,
		float32(p.player2.Width)+4,
		float32(p.player2.Height)+4,
		color.RGBA{R: 255, G: 80, B: 80, A: p2Glow},
		false,
	)
	vector.FillRect(
		screen,
		float32(p.player2.X),
		float32(p.player2.Y),
		float32(p.player2.Width),
		float32(p.player2.Height),
		color.RGBA{R: 255, G: 100, B: 100, A: 255},
		false,
	)

	// Ball with glow
	ballX := float32(p.ball.X + p.ball.Size/2)
	ballY := float32(p.ball.Y + p.ball.Size/2)
	vector.FillCircle(
		screen,
		ballX,
		ballY,
		float32(p.ball.Size/2)+4,
		color.RGBA{R: 255, G: 255, B: 255, A: 60},
		false,
	)
	vector.FillCircle(
		screen,
		ballX,
		ballY,
		float32(p.ball.Size/2),
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
		false,
	)

	// Scores
	ebitenutil.DebugPrintAt(screen, strconv.Itoa(p.player1.Score), screenWidth/4-5, 30)
	ebitenutil.DebugPrintAt(screen, strconv.Itoa(p.player2.Score), screenWidth*3/4-5, 30)

	// Score dots
	for i := 0; i < p.player1.Score; i++ {
		vector.FillCircle(
			screen,
			float32(screenWidth/4-20+i*15),
			60,
			4,
			color.RGBA{R: 100, G: 150, B: 255, A: 255},
			false,
		)
	}

	for i := 0; i < p.player2.Score; i++ {
		vector.FillCircle(
			screen,
			float32(screenWidth*3/4-20+i*15),
			60,
			4,
			color.RGBA{R: 255, G: 100, B: 100, A: 255},
			false,
		)
	}

	// Score popups
	for _, pop := range p.popups {
		alpha := uint8(pop.Timer * 255)
		ebitenutil.DebugPrintAt(screen, pop.Text, int(pop.X), int(pop.Y))

		_ = alpha // Would use for proper alpha rendering
	}

	// Serve indicator
	if p.serveTimer > 0 {
		countdown := min(int(p.serveTimer)+1, 3)

		text := strconv.Itoa(countdown)
		ebitenutil.DebugPrintAt(screen, text, screenWidth/2-3, screenHeight/2-20)

		// Arrow showing serve direction
		arrowX := float32(screenWidth / 2)
		if p.serving == 1 {
			vector.FillRect(
				screen,
				arrowX-30,
				screenHeight/2,
				20,
				3,
				color.RGBA{R: 100, G: 150, B: 255, A: 200},
				false,
			)
		} else {
			vector.FillRect(screen, arrowX+10, screenHeight/2, 20, 3, color.RGBA{R: 255, G: 100, B: 100, A: 200}, false)
		}
	}

	ebitenutil.DebugPrintAt(screen, "ESC: Pause", screenWidth/2-35, screenHeight-20)
}

func (p *Pong) drawPaused(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: 150},
		false,
	)

	boxW, boxH := float32(220), float32(100)
	boxX, boxY := float32(screenWidth-220)/2, float32(screenHeight-100)/2
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 30, G: 30, B: 50, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2, color.RGBA{R: 255, G: 255, B: 100, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "PAUSED", int(boxX)+80, int(boxY)+30)
	ebitenutil.DebugPrintAt(screen, "Press ESC or SPACE", int(boxX)+45, int(boxY)+60)
}

func (p *Pong) drawGameOver(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: 180},
		false,
	)

	var (
		winner   string
		winColor color.RGBA
	)

	if p.player1.Score >= p.winScore {
		winner = "PLAYER 1 WINS!"
		winColor = color.RGBA{R: 100, G: 150, B: 255, A: 255}
	} else {
		winner = "PLAYER 2 WINS!"
		winColor = color.RGBA{R: 255, G: 100, B: 100, A: 255}
	}

	boxW, boxH := float32(280), float32(140)
	boxX, boxY := float32(screenWidth-280)/2, float32(screenHeight-140)/2
	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 20, G: 25, B: 40, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, winColor, false)

	ebitenutil.DebugPrintAt(screen, winner, int(boxX)+75, int(boxY)+25)
	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf("Final: %d - %d", p.player1.Score, p.player2.Score),
		int(boxX)+90,
		int(boxY)+55,
	)
	ebitenutil.DebugPrintAt(screen, "SPACE: Play Again", int(boxX)+70, int(boxY)+90)
	ebitenutil.DebugPrintAt(screen, "ESC: Menu", int(boxX)+95, int(boxY)+110)
}

func (p *Pong) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func clamp(v, minVal, maxVal float64) float64 {
	return math.Max(minVal, math.Min(maxVal, v))
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Pong")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewPong()); err != nil {
		log.Fatal(err)
	}
}
