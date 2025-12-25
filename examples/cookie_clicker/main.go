package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 600
	screenHeight = 500
)

// Upgrade represents a purchasable upgrade.
type Upgrade struct {
	Name     string
	BaseCost float64
	CPS      float64 // Cookies per second
	Owned    int
}

// Game represents the cookie clicker game.
type Game struct {
	cookies      float64
	cps          float64 // Cookies per second
	totalCookies float64
	clicks       int
	upgrades     []*Upgrade
	clickPower   float64

	// Animation
	cookieScale  float64
	clickEffects []ClickEffect
}

// ClickEffect represents a floating +1 effect.
type ClickEffect struct {
	X, Y  float64
	Value float64
	Alpha float64
	Timer float64
}

// NewGame creates a new game.
func NewGame() *Game {
	return &Game{
		cookies:     0,
		clickPower:  1,
		cookieScale: 1.0,
		upgrades: []*Upgrade{
			{Name: "Cursor", BaseCost: 15, CPS: 0.1},
			{Name: "Grandma", BaseCost: 100, CPS: 1},
			{Name: "Farm", BaseCost: 1100, CPS: 8},
			{Name: "Mine", BaseCost: 12000, CPS: 47},
			{Name: "Factory", BaseCost: 130000, CPS: 260},
			{Name: "Bank", BaseCost: 1400000, CPS: 1400},
		},
	}
}

func (u *Upgrade) Cost() float64 {
	return math.Floor(u.BaseCost * math.Pow(1.15, float64(u.Owned)))
}

func (g *Game) Update() error {
	dt := 1.0 / 60.0

	// Passive cookie generation
	g.cookies += g.cps * dt
	g.totalCookies += g.cps * dt

	// Cookie click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()

		// Check cookie click (center area)
		cookieX, cookieY := 150, 250
		dx := float64(mx - cookieX)
		dy := float64(my - cookieY)
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 80 {
			g.cookies += g.clickPower
			g.totalCookies += g.clickPower
			g.clicks++
			g.cookieScale = 0.9 // Shrink on click

			// Add click effect
			g.clickEffects = append(g.clickEffects, ClickEffect{
				X:     float64(mx),
				Y:     float64(my),
				Value: g.clickPower,
				Alpha: 1.0,
			})
		}

		// Check upgrade clicks
		for i, upgrade := range g.upgrades {
			uy := 80 + i*60
			if mx >= 320 && mx <= 580 && my >= uy && my <= uy+50 {
				cost := upgrade.Cost()
				if g.cookies >= cost {
					g.cookies -= cost
					upgrade.Owned++
					g.cps += upgrade.CPS
				}
			}
		}
	}

	// Cookie bounce back
	if g.cookieScale < 1.0 {
		g.cookieScale += dt * 2
		if g.cookieScale > 1.0 {
			g.cookieScale = 1.0
		}
	}

	// Update click effects
	for i := len(g.clickEffects) - 1; i >= 0; i-- {
		effect := &g.clickEffects[i]
		effect.Timer += dt
		effect.Y -= 30 * dt

		effect.Alpha -= dt * 1.5
		if effect.Alpha <= 0 {
			g.clickEffects = append(g.clickEffects[:i], g.clickEffects[i+1:]...)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 40, G: 30, B: 50, A: 255})

	// Left panel - Cookie area
	vector.FillRect(screen, 0, 0, 300, screenHeight, color.RGBA{R: 50, G: 40, B: 60, A: 255}, false)

	// Cookie count
	ebitenutil.DebugPrintAt(screen, formatBigNumber(g.cookies)+" cookies", 60, 30)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("per second: %.1f", g.cps), 80, 50)

	// Big cookie
	g.drawCookie(screen, 150, 250, g.cookieScale)

	// Click effects
	for _, effect := range g.clickEffects {
		if effect.Alpha > 0 {
			alpha := uint8(effect.Alpha * 255)
			vector.FillCircle(
				screen,
				float32(effect.X),
				float32(effect.Y),
				15,
				color.RGBA{R: 255, G: 200, B: 50, A: alpha},
				false,
			)
		}
	}

	// Right panel - Upgrades
	vector.FillRect(screen, 300, 0, 300, screenHeight, color.RGBA{R: 60, G: 50, B: 70, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "UPGRADES", 400, 30)

	for i, upgrade := range g.upgrades {
		g.drawUpgrade(screen, upgrade, 320, 80+i*60)
	}

	// Stats at bottom
	vector.FillRect(
		screen,
		0,
		screenHeight-40,
		screenWidth,
		40,
		color.RGBA{R: 30, G: 25, B: 40, A: 255},
		false,
	)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total: %s | Clicks: %d | Click Power: %.0f",
		formatBigNumber(g.totalCookies), g.clicks, g.clickPower), 100, screenHeight-25)
}

func (g *Game) drawCookie(screen *ebiten.Image, x, y int, scale float64) {
	radius := float32(70 * scale)

	// Cookie base
	vector.FillCircle(
		screen,
		float32(x),
		float32(y),
		radius,
		color.RGBA{R: 200, G: 150, B: 80, A: 255},
		false,
	)

	// Cookie edge
	vector.StrokeCircle(
		screen,
		float32(x),
		float32(y),
		radius,
		4,
		color.RGBA{R: 150, G: 100, B: 50, A: 255},
		false,
	)

	// Chocolate chips
	chips := [][2]float32{
		{-25, -20}, {20, -30}, {-10, 10}, {25, 15}, {-30, 25}, {5, -5}, {30, -10},
	}
	for _, chip := range chips {
		cx := float32(x) + chip[0]*float32(scale)
		cy := float32(y) + chip[1]*float32(scale)
		vector.FillCircle(
			screen,
			cx,
			cy,
			8*float32(scale),
			color.RGBA{R: 80, G: 50, B: 30, A: 255},
			false,
		)
	}
}

func (g *Game) drawUpgrade(screen *ebiten.Image, upgrade *Upgrade, x, y int) {
	cost := upgrade.Cost()
	canAfford := g.cookies >= cost

	// Background
	bgColor := color.RGBA{R: 80, G: 70, B: 90, A: 255}
	if canAfford {
		bgColor = color.RGBA{R: 60, G: 120, B: 60, A: 255}
	}

	vector.FillRect(screen, float32(x), float32(y), 260, 50, bgColor, false)
	vector.StrokeRect(
		screen,
		float32(x),
		float32(y),
		260,
		50,
		2,
		color.RGBA{R: 150, G: 140, B: 160, A: 255},
		false,
	)

	// Name and count
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s (%d)", upgrade.Name, upgrade.Owned), x+10, y+8)

	// Cost
	ebitenutil.DebugPrintAt(screen, "Cost: "+formatBigNumber(cost), x+10, y+28)

	// CPS
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("+%.1f/s", upgrade.CPS), x+180, y+28)
}

func formatBigNumber(n float64) string {
	if n >= 1e15 {
		return fmt.Sprintf("%.2fQa", n/1e15)
	}

	if n >= 1e12 {
		return fmt.Sprintf("%.2fT", n/1e12)
	}

	if n >= 1e9 {
		return fmt.Sprintf("%.2fB", n/1e9)
	}

	if n >= 1e6 {
		return fmt.Sprintf("%.2fM", n/1e6)
	}

	if n >= 1e3 {
		return fmt.Sprintf("%.2fK", n/1e3)
	}

	return fmt.Sprintf("%.0f", n)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Cookie Clicker")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
