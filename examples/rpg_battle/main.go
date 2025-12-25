package main

import (
	"image/color"
	"log"
	"math/rand"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 700
	screenHeight = 500
)

// Character represents a combatant.
type Character struct {
	Name    string
	HP      int
	MaxHP   int
	MP      int
	MaxMP   int
	Attack  int
	Defense int
	Speed   int
	IsEnemy bool
	X, Y    float64
	Skills  []Skill
}

// Skill represents an ability.
type Skill struct {
	Name   string
	Cost   int
	Damage int
	IsHeal bool
}

// BattleState represents the game state.
type BattleState int

const (
	StateSelectAction BattleState = iota
	StateSelectSkill
	StateSelectTarget
	StateEnemyTurn
	StateAnimation
	StateVictory
	StateDefeat
)

// Game represents the RPG battle.
type Game struct {
	party          []*Character
	enemies        []*Character
	state          BattleState
	turnOrder      []*Character
	currentIdx     int
	selectedAction int
	selectedSkill  int
	selectedTarget int
	message        string
	animTimer      float64
	battleCount    int
}

// NewGame creates a new game.
func NewGame() *Game {
	g := &Game{}
	g.initBattle()

	return g
}

func (g *Game) initBattle() {
	// Create party
	g.party = []*Character{
		{
			Name: "Warrior", HP: 150, MaxHP: 150, MP: 30, MaxMP: 30,
			Attack: 20, Defense: 10, Speed: 8, X: 100, Y: 200,
			Skills: []Skill{
				{Name: "Slash", Cost: 0, Damage: 25},
				{Name: "Power Strike", Cost: 10, Damage: 45},
			},
		},
		{
			Name: "Mage", HP: 80, MaxHP: 80, MP: 100, MaxMP: 100,
			Attack: 8, Defense: 5, Speed: 12, X: 100, Y: 280,
			Skills: []Skill{
				{Name: "Fireball", Cost: 15, Damage: 40},
				{Name: "Ice Storm", Cost: 25, Damage: 60},
				{Name: "Heal", Cost: 20, Damage: 50, IsHeal: true},
			},
		},
		{
			Name: "Rogue", HP: 100, MaxHP: 100, MP: 50, MaxMP: 50,
			Attack: 18, Defense: 6, Speed: 15, X: 100, Y: 360,
			Skills: []Skill{
				{Name: "Backstab", Cost: 0, Damage: 30},
				{Name: "Poison", Cost: 15, Damage: 35},
			},
		},
	}

	// Create enemies
	g.enemies = []*Character{
		{Name: "Goblin", HP: 60, MaxHP: 60, Attack: 12, Defense: 3, Speed: 10, IsEnemy: true, X: 550, Y: 200},
		{Name: "Orc", HP: 100, MaxHP: 100, Attack: 18, Defense: 8, Speed: 6, IsEnemy: true, X: 550, Y: 300},
	}
	if g.battleCount > 0 {
		// Stronger enemies in later battles
		for _, e := range g.enemies {
			e.HP += g.battleCount * 20
			e.MaxHP = e.HP
			e.Attack += g.battleCount * 3
		}
	}

	// Calculate turn order
	g.calculateTurnOrder()
	g.state = StateSelectAction
	g.message = g.turnOrder[0].Name + "'s turn"
}

func (g *Game) calculateTurnOrder() {
	g.turnOrder = make([]*Character, 0)
	for _, c := range g.party {
		if c.HP > 0 {
			g.turnOrder = append(g.turnOrder, c)
		}
	}

	for _, c := range g.enemies {
		if c.HP > 0 {
			g.turnOrder = append(g.turnOrder, c)
		}
	}
	// Sort by speed (simple bubble sort)
	for i := 0; i < len(g.turnOrder); i++ {
		for j := i + 1; j < len(g.turnOrder); j++ {
			if g.turnOrder[j].Speed > g.turnOrder[i].Speed {
				g.turnOrder[i], g.turnOrder[j] = g.turnOrder[j], g.turnOrder[i]
			}
		}
	}

	g.currentIdx = 0
}

func (g *Game) currentChar() *Character {
	if g.currentIdx < len(g.turnOrder) {
		return g.turnOrder[g.currentIdx]
	}

	return nil
}

func (g *Game) Update() error {
	dt := 1.0 / 60.0

	// Animation timer
	if g.state == StateAnimation {
		g.animTimer -= dt
		if g.animTimer <= 0 {
			g.nextTurn()
		}

		return nil
	}

	// Victory/Defeat
	if g.state == StateVictory || g.state == StateDefeat {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			if g.state == StateVictory {
				g.battleCount++
				g.initBattle()
			} else {
				g.battleCount = 0
				g.initBattle()
			}
		}

		return nil
	}

	current := g.currentChar()
	if current == nil {
		g.calculateTurnOrder()

		return nil
	}

	// Enemy turn
	if current.IsEnemy {
		g.state = StateEnemyTurn
		// Simple AI: attack random party member
		alive := make([]*Character, 0)

		for _, c := range g.party {
			if c.HP > 0 {
				alive = append(alive, c)
			}
		}

		if len(alive) > 0 {
			target := alive[rand.Intn(len(alive))]

			damage := max(current.Attack-target.Defense/2, 5)

			target.HP -= damage
			if target.HP < 0 {
				target.HP = 0
			}

			g.message = current.Name + " attacks " + target.Name + " for " + formatInt(damage) + " damage!"
		}

		g.state = StateAnimation
		g.animTimer = 1.0

		return nil
	}

	// Player turn
	switch g.state {
	case StateSelectAction:
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			g.selectedAction = 0 // Attack
			g.state = StateSelectTarget
		}

		if inpututil.IsKeyJustPressed(ebiten.Key2) {
			g.selectedAction = 1 // Skills
			g.state = StateSelectSkill
			g.selectedSkill = 0
		}

		if inpututil.IsKeyJustPressed(ebiten.Key3) {
			g.selectedAction = 2 // Defend
			current.Defense += 5
			g.message = current.Name + " defends!"
			g.state = StateAnimation
			g.animTimer = 0.8
		}

	case StateSelectSkill:
		for i := 0; i < len(current.Skills); i++ {
			if inpututil.IsKeyJustPressed(ebiten.Key(int(ebiten.Key1) + i)) {
				if current.MP >= current.Skills[i].Cost {
					g.selectedSkill = i
					if current.Skills[i].IsHeal {
						// Target party member
						g.selectedTarget = -1
						for j, c := range g.party {
							if c.HP > 0 {
								g.selectedTarget = j

								break
							}
						}
					}

					g.state = StateSelectTarget
				} else {
					g.message = "Not enough MP!"
				}
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StateSelectAction
		}

	case StateSelectTarget:
		// Select enemy or party member
		if g.selectedAction == 1 && current.Skills[g.selectedSkill].IsHeal {
			// Heal party
			for i := 0; i < len(g.party); i++ {
				if inpututil.IsKeyJustPressed(ebiten.Key(int(ebiten.Key1)+i)) && g.party[i].HP > 0 {
					g.executeSkill(current, g.party[i], current.Skills[g.selectedSkill])
					g.state = StateAnimation
					g.animTimer = 1.0

					break
				}
			}
		} else {
			// Attack enemy
			for i := 0; i < len(g.enemies); i++ {
				if inpututil.IsKeyJustPressed(ebiten.Key(int(ebiten.Key1)+i)) && g.enemies[i].HP > 0 {
					if g.selectedAction == 0 {
						// Basic attack
						damage := max(current.Attack-g.enemies[i].Defense/2, 5)

						g.enemies[i].HP -= damage
						g.message = current.Name + " attacks for " + formatInt(damage) + "!"
					} else {
						g.executeSkill(current, g.enemies[i], current.Skills[g.selectedSkill])
					}

					g.state = StateAnimation
					g.animTimer = 1.0

					break
				}
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StateSelectAction
		}
	}

	return nil
}

func (g *Game) executeSkill(user, target *Character, skill Skill) {
	user.MP -= skill.Cost
	if skill.IsHeal {
		target.HP += skill.Damage
		if target.HP > target.MaxHP {
			target.HP = target.MaxHP
		}

		g.message = user.Name + " heals " + target.Name + " for " + formatInt(skill.Damage) + "!"
	} else {
		target.HP -= skill.Damage
		if target.HP < 0 {
			target.HP = 0
		}

		g.message = user.Name + " uses " + skill.Name + " for " + formatInt(skill.Damage) + "!"
	}
}

func (g *Game) nextTurn() {
	// Check victory/defeat
	allEnemiesDead := true

	for _, e := range g.enemies {
		if e.HP > 0 {
			allEnemiesDead = false

			break
		}
	}

	if allEnemiesDead {
		g.state = StateVictory
		g.message = "Victory! Press SPACE for next battle"

		return
	}

	allPartyDead := true

	for _, c := range g.party {
		if c.HP > 0 {
			allPartyDead = false

			break
		}
	}

	if allPartyDead {
		g.state = StateDefeat
		g.message = "Defeat! Press SPACE to restart"

		return
	}

	// Next character
	g.currentIdx++
	if g.currentIdx >= len(g.turnOrder) {
		g.calculateTurnOrder()
	}

	current := g.currentChar()
	if current != nil {
		g.message = current.Name + "'s turn"
		if !current.IsEnemy {
			g.state = StateSelectAction
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 40, G: 50, B: 60, A: 255})

	// Battle area
	vector.FillRect(screen, 50, 100, 600, 280, color.RGBA{R: 60, G: 70, B: 80, A: 255}, false)

	// Draw party
	for i, c := range g.party {
		g.drawCharacter(screen, c, i, false)
	}

	// Draw enemies
	for i, c := range g.enemies {
		g.drawCharacter(screen, c, i, true)
	}

	// UI Panel
	vector.FillRect(screen, 0, 400, screenWidth, 100, color.RGBA{R: 30, G: 35, B: 45, A: 255}, false)

	// Message
	ebitenutil.DebugPrintAt(screen, g.message, 20, 410)

	// Actions
	if g.state == StateSelectAction && g.currentChar() != nil && !g.currentChar().IsEnemy {
		ebitenutil.DebugPrintAt(screen, "[1] Attack  [2] Skills  [3] Defend", 20, 440)
	} else if g.state == StateSelectSkill {
		current := g.currentChar()

		text := ""

		var textSb393 strings.Builder
		for i, s := range current.Skills {
			textSb393.WriteString("[" + formatInt(i+1) + "] " + s.Name + " (" + formatInt(s.Cost) + "MP)  ")
		}

		text += textSb393.String()

		ebitenutil.DebugPrintAt(screen, text, 20, 440)
		ebitenutil.DebugPrintAt(screen, "[ESC] Back", 20, 460)
	} else if g.state == StateSelectTarget {
		if g.selectedAction == 1 && g.currentChar().Skills[g.selectedSkill].IsHeal {
			text := "Select ally: "

			var textSb401 strings.Builder

			for i, c := range g.party {
				if c.HP > 0 {
					textSb401.WriteString("[" + formatInt(i+1) + "] " + c.Name + "  ")
				}
			}

			text += textSb401.String()

			ebitenutil.DebugPrintAt(screen, text, 20, 440)
		} else {
			text := "Select enemy: "

			var textSb409 strings.Builder

			for i, e := range g.enemies {
				if e.HP > 0 {
					textSb409.WriteString("[" + formatInt(i+1) + "] " + e.Name + "  ")
				}
			}

			text += textSb409.String()

			ebitenutil.DebugPrintAt(screen, text, 20, 440)
		}

		ebitenutil.DebugPrintAt(screen, "[ESC] Back", 20, 460)
	}

	// Party stats
	for i, c := range g.party {
		x := 20 + i*220
		ebitenutil.DebugPrintAt(screen, c.Name, x, 480)
		ebitenutil.DebugPrintAt(
			screen,
			"HP:"+formatInt(c.HP)+"/"+formatInt(c.MaxHP)+" MP:"+formatInt(c.MP),
			x,
			495,
		)
	}
}

func (g *Game) drawCharacter(screen *ebiten.Image, c *Character, idx int, isEnemy bool) {
	if c.HP <= 0 {
		return
	}

	x := c.X
	y := c.Y

	// Body
	bodyColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	if isEnemy {
		bodyColor = color.RGBA{R: 200, G: 100, B: 100, A: 255}
	}

	vector.FillRect(screen, float32(x)-20, float32(y)-30, 40, 60, bodyColor, false)

	// Head
	vector.FillCircle(
		screen,
		float32(x),
		float32(y)-45,
		15,
		color.RGBA{R: 255, G: 220, B: 180, A: 255},
		false,
	)

	// Name
	ebitenutil.DebugPrintAt(screen, c.Name, int(x)-20, int(y)+35)

	// Health bar
	barW := float32(50)
	barH := float32(6)
	barX := float32(x) - barW/2
	barY := float32(y) - 70
	hpRatio := float32(c.HP) / float32(c.MaxHP)

	vector.FillRect(screen, barX, barY, barW, barH, color.RGBA{R: 60, G: 60, B: 60, A: 255}, false)

	barColor := color.RGBA{R: 50, G: 200, B: 50, A: 255}
	if hpRatio < 0.3 {
		barColor = color.RGBA{R: 200, G: 50, B: 50, A: 255}
	}

	vector.FillRect(screen, barX, barY, barW*hpRatio, barH, barColor, false)

	// Target indicator
	current := g.currentChar()
	if current != nil && current == c {
		vector.StrokeRect(
			screen,
			float32(x)-25,
			float32(y)-75,
			50,
			115,
			2,
			color.RGBA{R: 255, G: 255, B: 0, A: 255},
			false,
		)
	}
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	result := ""

	neg := n < 0
	if neg {
		n = -n
	}

	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}

	if neg {
		result = "-" + result
	}

	return result
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Turn-Based RPG Battle")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
