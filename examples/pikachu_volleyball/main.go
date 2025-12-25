package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth    = 800
	screenHeight   = 600
	playerSize     = 50
	ballRadius     = 20
	netWidth       = 8
	netHeight      = 120
	groundY        = float64(screenHeight - 40)
	gravity        = 0.8
	jumpPower      = -18
	moveSpeed      = 5.0
	ballBounceDamp = 0.85
	maxJumps       = 2
	winScore       = 15
)

// Player represents Pikachu character.
type Player struct {
	X, Y         float64
	VX, VY       float64
	Size         float64
	Score        int
	JumpsLeft    int
	OnGround     bool
	Color        color.RGBA
	ControlLeft  ebiten.Key
	ControlRight ebiten.Key
	ControlJump  ebiten.Key
}

// Ball represents the volleyball.
type Ball struct {
	X, Y        float64
	VX, VY      float64
	Radius      float64
	LastTouched int // 1 for player1, 2 for player2
}

// VolleyballGame represents the game state.
type VolleyballGame struct {
	player1  *Player
	player2  *Player
	ball     *Ball
	paused   bool
	gameOver bool
	winScore int
	groundY  float64
}

// NewVolleyballGame creates a new volleyball game.
func NewVolleyballGame() *VolleyballGame {
	g := &VolleyballGame{
		player1: &Player{
			X:            150,
			Y:            groundY - playerSize,
			Size:         playerSize,
			Color:        color.RGBA{R: 255, G: 220, B: 0, A: 255}, // Yellow for Pikachu
			ControlLeft:  ebiten.KeyA,
			ControlRight: ebiten.KeyD,
			ControlJump:  ebiten.KeyW,
			OnGround:     true,
			JumpsLeft:    maxJumps,
		},
		player2: &Player{
			X:            screenWidth - 150 - playerSize,
			Y:            groundY - playerSize,
			Size:         playerSize,
			Color:        color.RGBA{R: 255, G: 100, B: 100, A: 255}, // Red opponent
			ControlLeft:  ebiten.KeyLeft,
			ControlRight: ebiten.KeyRight,
			ControlJump:  ebiten.KeyUp,
			OnGround:     true,
			JumpsLeft:    maxJumps,
		},
		ball: &Ball{
			Radius: ballRadius,
		},
		winScore: winScore,
		groundY:  groundY,
	}
	g.resetBall()

	return g
}

func (g *VolleyballGame) resetBall() {
	g.ball.X = screenWidth / 2
	g.ball.Y = 100
	g.ball.VX = 0
	g.ball.VY = 0
	g.ball.LastTouched = 0
}

func (g *VolleyballGame) Update() error {
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			*g = *NewVolleyballGame()
		}

		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.paused = !g.paused
	}

	if g.paused {
		return nil
	}

	// Update players
	g.updatePlayer(g.player1)
	g.updatePlayer(g.player2)

	// Update ball physics
	g.ball.VY += gravity
	g.ball.X += g.ball.VX
	g.ball.Y += g.ball.VY

	// Ball collision with ground
	if g.ball.Y+g.ball.Radius >= g.groundY {
		g.ball.Y = g.groundY - g.ball.Radius
		g.ball.VY = -g.ball.VY * ballBounceDamp
		g.ball.VX *= 0.95

		// Check scoring
		if g.ball.X < screenWidth/2 {
			g.player2.Score++
			g.resetBall()
		} else if g.ball.X > screenWidth/2 {
			g.player1.Score++
			g.resetBall()
		}

		// Check win condition
		if g.player1.Score >= g.winScore || g.player2.Score >= g.winScore {
			g.gameOver = true
		}
	}

	// Ball collision with net
	netX := float64(screenWidth/2 - netWidth/2)
	if g.ball.X+g.ball.Radius > netX &&
		g.ball.X-g.ball.Radius < netX+netWidth &&
		g.ball.Y+g.ball.Radius > g.groundY-netHeight {
		// Bounce off net
		if g.ball.X < screenWidth/2 {
			g.ball.X = netX - g.ball.Radius
		} else {
			g.ball.X = netX + netWidth + g.ball.Radius
		}

		g.ball.VX = -g.ball.VX * 0.8
	}

	// Ball collision with players
	g.checkBallPlayerCollision(g.player1, 1)
	g.checkBallPlayerCollision(g.player2, 2)

	// Ball bounds
	if g.ball.X < g.ball.Radius {
		g.ball.X = g.ball.Radius
		g.ball.VX = -g.ball.VX * 0.8
	}

	if g.ball.X > screenWidth-g.ball.Radius {
		g.ball.X = screenWidth - g.ball.Radius
		g.ball.VX = -g.ball.VX * 0.8
	}

	if g.ball.Y < g.ball.Radius {
		g.ball.Y = g.ball.Radius
		g.ball.VY = -g.ball.VY * 0.8
	}

	return nil
}

func (g *VolleyballGame) updatePlayer(p *Player) {
	// Horizontal movement
	if ebiten.IsKeyPressed(p.ControlLeft) {
		p.VX = -moveSpeed
	} else if ebiten.IsKeyPressed(p.ControlRight) {
		p.VX = moveSpeed
	} else {
		p.VX *= 0.8 // Friction
	}

	// Jump
	if inpututil.IsKeyJustPressed(p.ControlJump) && p.JumpsLeft > 0 {
		p.VY = jumpPower
		p.JumpsLeft--
		p.OnGround = false
	}

	// Apply gravity
	if !p.OnGround {
		p.VY += gravity
	}

	// Update position
	p.X += p.VX
	p.Y += p.VY

	// Ground collision
	if p.Y+p.Size >= g.groundY {
		p.Y = g.groundY - p.Size
		p.VY = 0
		p.OnGround = true
		p.JumpsLeft = maxJumps
	} else {
		p.OnGround = false
	}

	// Net collision
	netX := float64(screenWidth/2 - netWidth/2)
	netRight := netX + netWidth

	// Player 1 (left side) can't cross net
	if p == g.player1 {
		if p.X+p.Size > netX {
			p.X = netX - p.Size
			p.VX = 0
		}
	}

	// Player 2 (right side) can't cross net
	if p == g.player2 {
		if p.X < netRight {
			p.X = netRight
			p.VX = 0
		}
	}

	// Screen bounds
	p.X = clamp(p.X, 0, screenWidth-p.Size)
	p.Y = clamp(p.Y, 0, g.groundY-p.Size)
}

func (g *VolleyballGame) checkBallPlayerCollision(p *Player, playerNum int) {
	// Circle-rectangle collision
	closestX := clamp(g.ball.X, p.X, p.X+p.Size)
	closestY := clamp(g.ball.Y, p.Y, p.Y+p.Size)

	distX := g.ball.X - closestX
	distY := g.ball.Y - closestY
	distSq := distX*distX + distY*distY

	if distSq < g.ball.Radius*g.ball.Radius {
		// Collision detected
		dist := math.Sqrt(distSq)
		if dist == 0 {
			dist = 1
		}

		// Normalize and push ball out
		nx := distX / dist
		ny := distY / dist

		overlap := g.ball.Radius - dist
		g.ball.X += nx * overlap
		g.ball.Y += ny * overlap

		// Calculate new velocity based on hit position
		centerX := p.X + p.Size/2
		centerY := p.Y + p.Size/2
		hitAngle := math.Atan2(g.ball.Y-centerY, g.ball.X-centerX)

		speed := math.Sqrt(g.ball.VX*g.ball.VX + g.ball.VY*g.ball.VY)
		speed = math.Max(speed, 8) // Minimum hit speed

		// Add player velocity to ball
		speed += math.Abs(p.VX) * 0.3

		g.ball.VX = math.Cos(hitAngle) * speed
		g.ball.VY = math.Sin(hitAngle) * speed

		// Bias upward
		if g.ball.VY > 0 {
			g.ball.VY *= 0.5
		}

		g.ball.LastTouched = playerNum
	}
}

func (g *VolleyballGame) Draw(screen *ebiten.Image) {
	// Sky background
	screen.Fill(color.RGBA{R: 135, G: 206, B: 235, A: 255})

	// Ground
	ground := ebiten.NewImage(screenWidth, int(screenHeight-g.groundY))
	ground.Fill(color.RGBA{R: 139, G: 69, B: 19, A: 255})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, g.groundY)
	screen.DrawImage(ground, op)

	// Net
	netX := screenWidth/2 - netWidth/2
	net := ebiten.NewImage(netWidth, netHeight)
	net.Fill(color.RGBA{R: 255, G: 255, B: 255, A: 255})

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(netX), g.groundY-netHeight)
	screen.DrawImage(net, op)

	// Net top
	netTop := ebiten.NewImage(netWidth*2, 10)
	netTop.Fill(color.RGBA{R: 200, G: 0, B: 0, A: 255})

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(netX-netWidth/2), g.groundY-netHeight-5)
	screen.DrawImage(netTop, op)

	// Players (draw as Pikachu-like characters)
	g.drawPlayer(screen, g.player1)
	g.drawPlayer(screen, g.player2)

	// Ball
	g.drawBall(screen)

	// Scores
	g.drawScore(screen, g.player1.Score, 100, 40, g.player1.Color)
	g.drawScore(screen, g.player2.Score, screenWidth-100, 40, g.player2.Color)

	// Controls hint
	ebitenutil.DebugPrintAt(screen, "P1: A/D/W  P2: Arrows  ESC: Pause", screenWidth/2-120, screenHeight-25)

	// Pause overlay
	if g.paused {
		vector.FillRect(
			screen,
			0,
			0,
			screenWidth,
			screenHeight,
			color.RGBA{R: 0, G: 0, B: 0, A: 150},
			false,
		)

		boxW, boxH := float32(250), float32(80)
		boxX, boxY := float32(screenWidth-250)/2, float32(screenHeight-80)/2
		vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 50, G: 50, B: 70, A: 255}, false)
		vector.StrokeRect(
			screen,
			boxX,
			boxY,
			boxW,
			boxH,
			2,
			color.RGBA{R: 255, G: 255, B: 100, A: 255},
			false,
		)

		ebitenutil.DebugPrintAt(screen, "PAUSED", int(boxX)+95, int(boxY)+25)
		ebitenutil.DebugPrintAt(screen, "Press ESC to resume", int(boxX)+60, int(boxY)+50)
	}

	// Game over overlay
	if g.gameOver {
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

		if g.player1.Score >= g.winScore {
			winner = "PLAYER 1 WINS!"
			winColor = g.player1.Color
		} else {
			winner = "PLAYER 2 WINS!"
			winColor = g.player2.Color
		}

		boxW, boxH := float32(280), float32(120)
		boxX, boxY := float32(screenWidth-280)/2, float32(screenHeight-120)/2
		vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 30, G: 30, B: 50, A: 255}, false)
		vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, winColor, false)

		ebitenutil.DebugPrintAt(screen, winner, int(boxX)+80, int(boxY)+25)
		ebitenutil.DebugPrintAt(
			screen,
			fmt.Sprintf("Score: %d - %d", g.player1.Score, g.player2.Score),
			int(boxX)+95,
			int(boxY)+55,
		)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to restart", int(boxX)+60, int(boxY)+90)
	}
}

func (g *VolleyballGame) drawPlayer(screen *ebiten.Image, p *Player) {
	// Body (rounded rectangle approximation)
	bodyImg := ebiten.NewImage(int(p.Size), int(p.Size))
	bodyImg.Fill(p.Color)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(bodyImg, op)

	// Ears (two triangular shapes approximated as rectangles)
	earSize := p.Size * 0.3
	leftEar := ebiten.NewImage(int(earSize), int(earSize*1.5+0.5))
	leftEar.Fill(p.Color)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X+p.Size*0.15, p.Y-earSize)
	screen.DrawImage(leftEar, op)

	rightEar := ebiten.NewImage(int(earSize), int(earSize*1.5+0.5))
	rightEar.Fill(p.Color)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X+p.Size*0.7, p.Y-earSize)
	screen.DrawImage(rightEar, op)

	// Ear tips (black)
	earTip := ebiten.NewImage(int(earSize*0.8+0.5), int(earSize*0.5+0.5))
	earTip.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 255})

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X+p.Size*0.17, p.Y-earSize)
	screen.DrawImage(earTip, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X+p.Size*0.72, p.Y-earSize)
	screen.DrawImage(earTip, op)

	// Eyes
	eyeSize := p.Size * 0.12
	leftEye := ebiten.NewImage(int(eyeSize+0.5), int(eyeSize+0.5))
	leftEye.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 255})

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X+p.Size*0.3, p.Y+p.Size*0.3)
	screen.DrawImage(leftEye, op)

	rightEye := ebiten.NewImage(int(eyeSize+0.5), int(eyeSize+0.5))
	rightEye.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 255})

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X+p.Size*0.6, p.Y+p.Size*0.3)
	screen.DrawImage(rightEye, op)

	// Cheeks (red circles)
	cheekSize := p.Size * 0.15
	leftCheek := ebiten.NewImage(int(cheekSize+0.5), int(cheekSize+0.5))
	leftCheek.Fill(color.RGBA{R: 255, G: 100, B: 100, A: 255})

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X+p.Size*0.1, p.Y+p.Size*0.45)
	screen.DrawImage(leftCheek, op)

	rightCheek := ebiten.NewImage(int(cheekSize+0.5), int(cheekSize+0.5))
	rightCheek.Fill(color.RGBA{R: 255, G: 100, B: 100, A: 255})

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X+p.Size*0.75, p.Y+p.Size*0.45)
	screen.DrawImage(rightCheek, op)
}

func (g *VolleyballGame) drawBall(screen *ebiten.Image) {
	// Draw circle using vector package
	ballImg := ebiten.NewImage(int(g.ball.Radius*2), int(g.ball.Radius*2))

	// White volleyball
	vector.FillCircle(
		ballImg,
		float32(g.ball.Radius),
		float32(g.ball.Radius),
		float32(g.ball.Radius),
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
		false,
	)

	// Volleyball pattern lines
	vector.StrokeLine(
		ballImg,
		float32(g.ball.Radius),
		0,
		float32(g.ball.Radius),
		float32(g.ball.Radius*2),
		2,
		color.RGBA{R: 200, G: 200, B: 200, A: 255},
		false,
	)
	vector.StrokeLine(
		ballImg,
		0,
		float32(g.ball.Radius),
		float32(g.ball.Radius*2),
		float32(g.ball.Radius),
		2,
		color.RGBA{R: 200, G: 200, B: 200, A: 255},
		false,
	)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.ball.X-g.ball.Radius, g.ball.Y-g.ball.Radius)
	screen.DrawImage(ballImg, op)
}

func (g *VolleyballGame) drawScore(screen *ebiten.Image, score, x, y int, clr color.RGBA) {
	// Score background box
	vector.FillRect(
		screen,
		float32(x-30),
		float32(y),
		60,
		45,
		color.RGBA{R: 0, G: 0, B: 0, A: 150},
		false,
	)
	vector.StrokeRect(screen, float32(x-30), float32(y), 60, 45, 2, clr, false)

	// Score number
	ebitenutil.DebugPrintAt(screen, strconv.Itoa(score), x-5, y+15)
}

func (g *VolleyballGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func clamp(v, minVal, maxVal float64) float64 {
	return math.Max(minVal, math.Min(maxVal, v))
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Pikachu Volleyball - Framework Example")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewVolleyballGame()); err != nil {
		log.Fatal(err)
	}
}
