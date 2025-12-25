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
	screenWidth  = 800
	screenHeight = 600
)

// UnitType represents unit types.
type UnitType int

const (
	UnitSoldier UnitType = iota
	UnitArcher
	UnitTank
)

// Unit represents a game unit.
type Unit struct {
	X, Y      float64
	TargetX   float64
	TargetY   float64
	Health    int
	MaxHealth int
	Attack    int
	Range     float64
	Speed     float64
	Type      UnitType
	Team      int // 0=player, 1=enemy
	Selected  bool
	AttackCD  float64
	Moving    bool
}

// Game represents the mini RTS.
type Game struct {
	units         []*Unit
	selectedUnits []*Unit
	selectStartX  int
	selectStartY  int
	selecting     bool
	resources     int
	enemySpawnCD  float64
	wave          int
	message       string
	messageTimer  float64
}

// NewGame creates a new game.
func NewGame() *Game {
	g := &Game{
		units:     make([]*Unit, 0),
		resources: 500,
		wave:      1,
	}

	// Spawn starting units
	for i := range 5 {
		g.units = append(g.units, g.createUnit(100+float64(i*30), 300, 0, UnitSoldier))
	}

	return g
}

func (g *Game) createUnit(x, y float64, team int, uType UnitType) *Unit {
	u := &Unit{
		X:       x,
		Y:       y,
		TargetX: x,
		TargetY: y,
		Team:    team,
		Type:    uType,
	}

	switch uType {
	case UnitSoldier:
		u.Health, u.MaxHealth = 100, 100
		u.Attack = 15
		u.Range = 25
		u.Speed = 2
	case UnitArcher:
		u.Health, u.MaxHealth = 60, 60
		u.Attack = 20
		u.Range = 120
		u.Speed = 1.5
	case UnitTank:
		u.Health, u.MaxHealth = 200, 200
		u.Attack = 25
		u.Range = 20
		u.Speed = 1
	}

	return u
}

func (g *Game) Update() error {
	dt := 1.0 / 60.0

	// Message timer
	if g.messageTimer > 0 {
		g.messageTimer -= dt
	}

	// Enemy spawn
	g.enemySpawnCD -= dt
	if g.enemySpawnCD <= 0 {
		g.spawnEnemyWave()
		g.enemySpawnCD = 15.0
		g.wave++
	}

	// Unit buying
	if inpututil.IsKeyJustPressed(ebiten.Key1) && g.resources >= 50 {
		g.resources -= 50
		g.units = append(g.units, g.createUnit(50+rand.Float64()*80, 250+rand.Float64()*100, 0, UnitSoldier))
		g.showMessage("Soldier purchased!")
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) && g.resources >= 80 {
		g.resources -= 80
		g.units = append(g.units, g.createUnit(50+rand.Float64()*80, 250+rand.Float64()*100, 0, UnitArcher))
		g.showMessage("Archer purchased!")
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) && g.resources >= 150 {
		g.resources -= 150
		g.units = append(g.units, g.createUnit(50+rand.Float64()*80, 250+rand.Float64()*100, 0, UnitTank))
		g.showMessage("Tank purchased!")
	}

	// Selection box
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.selectStartX, g.selectStartY = ebiten.CursorPosition()
		g.selecting = true
	}

	// Selection box is drawn while selecting in Draw()

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if g.selecting {
			// Select units in box
			x1, y1 := min(g.selectStartX, mx), min(g.selectStartY, my)
			x2, y2 := max(g.selectStartX, mx), max(g.selectStartY, my)

			// Clear previous selection
			for _, u := range g.units {
				u.Selected = false
			}

			// Select new units
			g.selectedUnits = nil
			for _, u := range g.units {
				if u.Team == 0 && int(u.X) >= x1 && int(u.X) <= x2 && int(u.Y) >= y1 && int(u.Y) <= y2 {
					u.Selected = true
					g.selectedUnits = append(g.selectedUnits, u)
				}
			}
		}

		g.selecting = false
	}

	// Move command (right click)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()
		for _, u := range g.selectedUnits {
			u.TargetX = float64(mx)
			u.TargetY = float64(my)
			u.Moving = true
		}
	}

	// Update units
	for _, u := range g.units {
		// Movement
		if u.Moving {
			dx := u.TargetX - u.X
			dy := u.TargetY - u.Y

			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 5 {
				u.X += (dx / dist) * u.Speed
				u.Y += (dy / dist) * u.Speed
			} else {
				u.Moving = false
			}
		}

		// Attack cooldown
		u.AttackCD -= dt

		// Find enemy to attack
		var target *Unit

		minDist := u.Range
		for _, other := range g.units {
			if other.Team != u.Team && other.Health > 0 {
				dist := math.Sqrt((u.X-other.X)*(u.X-other.X) + (u.Y-other.Y)*(u.Y-other.Y))
				if dist < minDist {
					minDist = dist
					target = other
				}
			}
		}

		// Attack
		if target != nil && u.AttackCD <= 0 {
			target.Health -= u.Attack
			u.AttackCD = 1.0

			// Enemy AI: move toward player units
			if u.Team == 1 && !u.Moving {
				u.TargetX = target.X
				u.TargetY = target.Y
				u.Moving = true
			}
		}

		// AI for enemies without targets
		if u.Team == 1 && target == nil {
			// Move toward left side
			if !u.Moving {
				u.TargetX = 100
				u.TargetY = u.Y + rand.Float64()*50 - 25
				u.Moving = true
			}
		}
	}

	// Remove dead units
	for i := len(g.units) - 1; i >= 0; i-- {
		if g.units[i].Health <= 0 {
			if g.units[i].Team == 1 {
				g.resources += 20
			}

			g.units = append(g.units[:i], g.units[i+1:]...)
		}
	}

	return nil
}

func (g *Game) spawnEnemyWave() {
	count := 3 + g.wave
	for range count {
		uType := UnitSoldier
		if rand.Float64() < 0.3 {
			uType = UnitArcher
		}

		if g.wave > 3 && rand.Float64() < 0.2 {
			uType = UnitTank
		}

		g.units = append(g.units, g.createUnit(
			float64(screenWidth)-50,
			150+rand.Float64()*300,
			1, uType))
	}

	g.showMessage("Wave " + formatInt(g.wave) + " incoming!")
}

func (g *Game) showMessage(msg string) {
	g.message = msg
	g.messageTimer = 2.0
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 60, G: 80, B: 60, A: 255})

	// Draw units
	for _, u := range g.units {
		g.drawUnit(screen, u)
	}

	// Selection box
	if g.selecting {
		mx, my := ebiten.CursorPosition()
		x1, y1 := float32(min(g.selectStartX, mx)), float32(min(g.selectStartY, my))
		x2, y2 := float32(max(g.selectStartX, mx)), float32(max(g.selectStartY, my))
		vector.StrokeRect(screen, x1, y1, x2-x1, y2-y1, 2, color.RGBA{R: 0, G: 255, B: 0, A: 200}, false)
	}

	// UI Panel
	vector.FillRect(screen, 0, 0, screenWidth, 50, color.RGBA{R: 40, G: 40, B: 40, A: 220}, false)

	ebitenutil.DebugPrintAt(screen, "Resources: "+formatInt(g.resources), 10, 10)
	ebitenutil.DebugPrintAt(screen, "Wave: "+formatInt(g.wave), 150, 10)
	ebitenutil.DebugPrintAt(screen, "Selected: "+formatInt(len(g.selectedUnits)), 250, 10)

	ebitenutil.DebugPrintAt(screen, "[1] Soldier $50", 400, 10)
	ebitenutil.DebugPrintAt(screen, "[2] Archer $80", 530, 10)
	ebitenutil.DebugPrintAt(screen, "[3] Tank $150", 660, 10)

	ebitenutil.DebugPrintAt(screen, "Left drag = Select | Right click = Move", 200, 30)

	// Message
	if g.messageTimer > 0 {
		vector.FillRect(
			screen,
			float32(screenWidth/2-100),
			float32(screenHeight/2-15),
			200,
			30,
			color.RGBA{R: 0, G: 0, B: 0, A: 180},
			false,
		)
		ebitenutil.DebugPrintAt(screen, g.message, screenWidth/2-len(g.message)*3, screenHeight/2-7)
	}
}

func (g *Game) drawUnit(screen *ebiten.Image, u *Unit) {
	var (
		c    color.RGBA
		size float32 = 12
	)

	switch u.Type {
	case UnitSoldier:
		if u.Team == 0 {
			c = color.RGBA{R: 50, G: 150, B: 255, A: 255}
		} else {
			c = color.RGBA{R: 255, G: 100, B: 100, A: 255}
		}
	case UnitArcher:
		size = 10

		if u.Team == 0 {
			c = color.RGBA{R: 100, G: 200, B: 100, A: 255}
		} else {
			c = color.RGBA{R: 255, G: 150, B: 100, A: 255}
		}
	case UnitTank:
		size = 18

		if u.Team == 0 {
			c = color.RGBA{R: 100, G: 100, B: 200, A: 255}
		} else {
			c = color.RGBA{R: 200, G: 50, B: 50, A: 255}
		}
	}

	// Body
	vector.FillCircle(screen, float32(u.X), float32(u.Y), size, c, false)

	// Selection ring
	if u.Selected {
		vector.StrokeCircle(
			screen,
			float32(u.X),
			float32(u.Y),
			size+3,
			2,
			color.RGBA{R: 0, G: 255, B: 0, A: 255},
			false,
		)
	}

	// Health bar
	barW := float32(size * 2)
	barH := float32(4)
	barX := float32(u.X) - barW/2
	barY := float32(u.Y) - size - 8
	healthRatio := float32(u.Health) / float32(u.MaxHealth)

	vector.FillRect(screen, barX, barY, barW, barH, color.RGBA{R: 60, G: 60, B: 60, A: 255}, false)
	vector.FillRect(
		screen,
		barX,
		barY,
		barW*healthRatio,
		barH,
		color.RGBA{R: 50, G: 200, B: 50, A: 255},
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
	ebiten.SetWindowTitle("Mini RTS")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
