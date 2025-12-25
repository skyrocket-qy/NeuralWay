package main

import (
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
	screenWidth  = 480
	screenHeight = 640
	playerSpeed  = 5
	bulletSpeed  = 8
	enemySpeed   = 2
)

// Entity represents a game entity.
type Entity struct {
	X, Y   float64
	W, H   float64
	VX, VY float64
	Health int
	Active bool
}

// Game represents the space shooter game.
type Game struct {
	player        *Entity
	bullets       []*Entity
	enemies       []*Entity
	particles     []*Particle
	score         int
	highscore     int
	lives         int
	gameOver      bool
	spawnTimer    float64
	shootCooldown float64
	level         int
}

// Particle for explosions.
type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64
	Color  color.RGBA
}

// NewGame creates a new game.
func NewGame() *Game {
	return &Game{
		player: &Entity{
			X:      float64(screenWidth) / 2,
			Y:      float64(screenHeight) - 80,
			W:      40,
			H:      40,
			Health: 1,
			Active: true,
		},
		bullets:   make([]*Entity, 0),
		enemies:   make([]*Entity, 0),
		particles: make([]*Particle, 0),
		lives:     3,
		level:     1,
	}
}

func (g *Game) reset() {
	g.player.X = float64(screenWidth) / 2
	g.player.Y = float64(screenHeight) - 80
	g.player.Active = true
	g.bullets = make([]*Entity, 0)
	g.enemies = make([]*Entity, 0)
	g.particles = make([]*Particle, 0)
	g.score = 0
	g.lives = 3
	g.gameOver = false
	g.level = 1
}

func (g *Game) Update() error {
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			if g.score > g.highscore {
				g.highscore = g.score
			}

			g.reset()
		}

		return nil
	}

	dt := 1.0 / 60.0

	// Player movement
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.X -= playerSpeed
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.X += playerSpeed
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.Y -= playerSpeed
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.player.Y += playerSpeed
	}

	// Clamp player position
	g.player.X = clamp(g.player.X, g.player.W/2, float64(screenWidth)-g.player.W/2)
	g.player.Y = clamp(g.player.Y, g.player.H/2, float64(screenHeight)-g.player.H/2)

	// Shooting
	g.shootCooldown -= dt
	if ebiten.IsKeyPressed(ebiten.KeySpace) && g.shootCooldown <= 0 {
		g.shoot()
		g.shootCooldown = 0.15
	}

	// Update bullets
	for i := len(g.bullets) - 1; i >= 0; i-- {
		b := g.bullets[i]

		b.Y += b.VY
		if b.Y < -10 {
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
		}
	}

	// Spawn enemies
	g.spawnTimer += dt

	spawnRate := 1.5 - float64(g.level)*0.1
	if spawnRate < 0.3 {
		spawnRate = 0.3
	}

	if g.spawnTimer >= spawnRate {
		g.spawnEnemy()
		g.spawnTimer = 0
	}

	// Update enemies
	for i := len(g.enemies) - 1; i >= 0; i-- {
		e := g.enemies[i]
		e.Y += e.VY
		e.X += math.Sin(e.Y*0.02) * 2 // Wavy motion

		if e.Y > screenHeight+50 {
			g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)

			continue
		}

		// Collision with player
		if g.checkCollision(g.player, e) {
			g.lives--
			g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
			g.spawnExplosion(g.player.X, g.player.Y)

			if g.lives <= 0 {
				g.gameOver = true
			}

			continue
		}

		// Collision with bullets
		for j := len(g.bullets) - 1; j >= 0; j-- {
			b := g.bullets[j]
			if g.checkCollision(b, e) {
				g.score += 100
				g.bullets = append(g.bullets[:j], g.bullets[j+1:]...)
				g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
				g.spawnExplosion(e.X, e.Y)

				// Level up every 500 points
				if g.score > 0 && g.score%500 == 0 {
					g.level++
				}

				break
			}
		}
	}

	// Update particles
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.X += p.VX
		p.Y += p.VY

		p.Life -= dt * 2
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	return nil
}

func (g *Game) shoot() {
	g.bullets = append(g.bullets, &Entity{
		X:      g.player.X,
		Y:      g.player.Y - g.player.H/2,
		W:      6,
		H:      15,
		VY:     -bulletSpeed,
		Active: true,
	})
}

func (g *Game) spawnEnemy() {
	x := rand.Float64()*(screenWidth-60) + 30
	speed := enemySpeed + rand.Float64()*float64(g.level)*0.5

	g.enemies = append(g.enemies, &Entity{
		X:      x,
		Y:      -30,
		W:      35,
		H:      35,
		VY:     speed,
		Active: true,
	})
}

func (g *Game) spawnExplosion(x, y float64) {
	for range 15 {
		angle := rand.Float64() * math.Pi * 2
		speed := rand.Float64()*3 + 1
		g.particles = append(g.particles, &Particle{
			X:     x,
			Y:     y,
			VX:    math.Cos(angle) * speed,
			VY:    math.Sin(angle) * speed,
			Life:  1.0,
			Color: color.RGBA{R: 255, G: uint8(rand.Intn(200)), B: 0, A: 255},
		})
	}
}

func (g *Game) checkCollision(a, b *Entity) bool {
	return a.X-a.W/2 < b.X+b.W/2 &&
		a.X+a.W/2 > b.X-b.W/2 &&
		a.Y-a.H/2 < b.Y+b.H/2 &&
		a.Y+a.H/2 > b.Y-b.H/2
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Starfield background
	screen.Fill(color.RGBA{R: 5, G: 5, B: 20, A: 255})
	g.drawStars(screen)

	// Draw particles
	for _, p := range g.particles {
		alpha := uint8(p.Life * 255)
		c := color.RGBA{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
		vector.FillCircle(screen, float32(p.X), float32(p.Y), 3, c, false)
	}

	// Draw bullets
	for _, b := range g.bullets {
		vector.FillRect(
			screen,
			float32(b.X-b.W/2),
			float32(b.Y-b.H/2),
			float32(b.W),
			float32(b.H),
			color.RGBA{R: 255, G: 255, B: 100, A: 255},
			false,
		)
	}

	// Draw enemies
	for _, e := range g.enemies {
		g.drawEnemy(screen, e)
	}

	// Draw player
	if g.player.Active && !g.gameOver {
		g.drawPlayer(screen)
	}

	// UI
	g.drawUI(screen)

	if g.gameOver {
		g.drawGameOver(screen)
	}
}

func (g *Game) drawStars(screen *ebiten.Image) {
	// Simple pseudo-random stars
	for i := range 50 {
		x := (i*47 + 31) % screenWidth
		y := (i*89 + 17) % screenHeight
		brightness := uint8(100 + (i*7)%155)
		vector.FillCircle(
			screen,
			float32(x),
			float32(y),
			1,
			color.RGBA{R: brightness, G: brightness, B: brightness, A: 255},
			false,
		)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	x := float32(g.player.X)
	y := float32(g.player.Y)

	// Ship body
	vector.FillRect(screen, x-15, y-10, 30, 25, color.RGBA{R: 100, G: 150, B: 255, A: 255}, false)
	// Cockpit
	vector.FillCircle(screen, x, y-5, 8, color.RGBA{R: 50, G: 200, B: 255, A: 255}, false)
	// Wings
	vector.FillRect(screen, x-25, y+5, 50, 8, color.RGBA{R: 80, G: 120, B: 200, A: 255}, false)
	// Engine glow
	vector.FillRect(screen, x-8, y+15, 16, 8, color.RGBA{R: 255, G: 150, B: 50, A: 255}, false)
}

func (g *Game) drawEnemy(screen *ebiten.Image, e *Entity) {
	x := float32(e.X)
	y := float32(e.Y)

	// Enemy body
	vector.FillCircle(screen, x, y, float32(e.W/2), color.RGBA{R: 200, G: 50, B: 50, A: 255}, false)
	// Eyes
	vector.FillCircle(screen, x-8, y-5, 5, color.RGBA{R: 255, G: 255, B: 0, A: 255}, false)
	vector.FillCircle(screen, x+8, y-5, 5, color.RGBA{R: 255, G: 255, B: 0, A: 255}, false)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Top bar
	vector.FillRect(screen, 0, 0, screenWidth, 40, color.RGBA{R: 20, G: 20, B: 40, A: 200}, false)

	ebitenutil.DebugPrintAt(screen, "Score: "+formatInt(g.score), 10, 12)
	ebitenutil.DebugPrintAt(screen, "Level: "+formatInt(g.level), screenWidth/2-30, 12)
	ebitenutil.DebugPrintAt(screen, "Lives: "+formatInt(g.lives), screenWidth-80, 12)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: 180},
		false,
	)

	boxW := float32(250)
	boxH := float32(120)
	boxX := float32(screenWidth-250) / 2
	boxY := float32(screenHeight-120) / 2

	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 40, G: 40, B: 60, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2, color.RGBA{R: 255, G: 100, B: 100, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "GAME OVER", int(boxX)+80, int(boxY)+25)
	ebitenutil.DebugPrintAt(screen, "Score: "+formatInt(g.score), int(boxX)+80, int(boxY)+50)
	ebitenutil.DebugPrintAt(screen, "Press SPACE to restart", int(boxX)+35, int(boxY)+80)
}

func formatInt(n int) string {
	result := ""

	if n == 0 {
		return "0"
	}

	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}

	return result
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

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Space Shooter")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
