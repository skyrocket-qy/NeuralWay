package main

import (
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 640
	screenHeight = 480
	tileSize     = 32
	mapWidth     = 20
	mapHeight    = 15
)

// TileType represents map tiles.
type TileType int

const (
	TileFloor TileType = iota
	TileWall
	TileStairs
)

// ItemType represents item types.
type ItemType int

const (
	ItemPotion ItemType = iota
	ItemWeapon
	ItemArmor
	ItemGold
)

// Item represents a pickup.
type Item struct {
	X, Y  int
	Type  ItemType
	Value int
}

// Enemy represents a monster.
type Enemy struct {
	X, Y   int
	HP     int
	MaxHP  int
	Attack int
	Name   string
	Dead   bool
}

// Player represents the player.
type Player struct {
	X, Y    int
	HP      int
	MaxHP   int
	Attack  int
	Defense int
	Gold    int
	Level   int
	XP      int
}

// Game represents the roguelike.
type Game struct {
	tiles    [mapHeight][mapWidth]TileType
	player   *Player
	enemies  []*Enemy
	items    []*Item
	floor    int
	message  string
	messages []string
	gameOver bool
}

// NewGame creates a new game.
func NewGame() *Game {
	g := &Game{
		player: &Player{
			HP: 100, MaxHP: 100,
			Attack: 10, Defense: 5,
			Level: 1,
		},
		floor:    1,
		messages: make([]string, 0),
	}
	g.generateLevel()

	return g
}

func (g *Game) generateLevel() {
	// Fill with walls
	for y := range mapHeight {
		for x := range mapWidth {
			g.tiles[y][x] = TileWall
		}
	}

	// Generate rooms using simple BSP-like approach
	rooms := make([][4]int, 0) // x, y, w, h

	// Create random rooms
	for i := 0; i < 6+rand.Intn(4); i++ {
		w := 4 + rand.Intn(5)
		h := 3 + rand.Intn(4)
		x := 1 + rand.Intn(mapWidth-w-2)
		y := 1 + rand.Intn(mapHeight-h-2)

		// Carve room
		for ry := y; ry < y+h && ry < mapHeight-1; ry++ {
			for rx := x; rx < x+w && rx < mapWidth-1; rx++ {
				g.tiles[ry][rx] = TileFloor
			}
		}

		rooms = append(rooms, [4]int{x, y, w, h})
	}

	// Connect rooms with corridors
	for i := 1; i < len(rooms); i++ {
		x1 := rooms[i-1][0] + rooms[i-1][2]/2
		y1 := rooms[i-1][1] + rooms[i-1][3]/2
		x2 := rooms[i][0] + rooms[i][2]/2
		y2 := rooms[i][1] + rooms[i][3]/2

		// Horizontal corridor
		for x := min(x1, x2); x <= max(x1, x2); x++ {
			if y1 >= 0 && y1 < mapHeight && x >= 0 && x < mapWidth {
				g.tiles[y1][x] = TileFloor
			}
		}
		// Vertical corridor
		for y := min(y1, y2); y <= max(y1, y2); y++ {
			if y >= 0 && y < mapHeight && x2 >= 0 && x2 < mapWidth {
				g.tiles[y][x2] = TileFloor
			}
		}
	}

	// Place player in first room
	if len(rooms) > 0 {
		g.player.X = rooms[0][0] + 1
		g.player.Y = rooms[0][1] + 1
	}

	// Place stairs in last room
	if len(rooms) > 1 {
		stairRoom := rooms[len(rooms)-1]
		g.tiles[stairRoom[1]+1][stairRoom[0]+stairRoom[2]-2] = TileStairs
	}

	// Spawn enemies
	g.enemies = make([]*Enemy, 0)
	enemyNames := []string{"Goblin", "Rat", "Skeleton", "Orc", "Spider"}

	for i := 0; i < 5+g.floor*2; i++ {
		for range 50 {
			x := rand.Intn(mapWidth)

			y := rand.Intn(mapHeight)
			if g.tiles[y][x] == TileFloor && (x != g.player.X || y != g.player.Y) {
				hp := 20 + g.floor*10 + rand.Intn(20)
				g.enemies = append(g.enemies, &Enemy{
					X: x, Y: y,
					HP: hp, MaxHP: hp,
					Attack: 5 + g.floor*2,
					Name:   enemyNames[rand.Intn(len(enemyNames))],
				})

				break
			}
		}
	}

	// Spawn items
	g.items = make([]*Item, 0)

	for i := 0; i < 3+rand.Intn(3); i++ {
		for range 50 {
			x := rand.Intn(mapWidth)

			y := rand.Intn(mapHeight)
			if g.tiles[y][x] == TileFloor {
				itemType := ItemType(rand.Intn(4))
				value := 10 + rand.Intn(20)
				g.items = append(g.items, &Item{X: x, Y: y, Type: itemType, Value: value})

				break
			}
		}
	}

	g.addMessage("Entered floor " + formatInt(g.floor))
}

func (g *Game) addMessage(msg string) {
	g.messages = append(g.messages, msg)
	if len(g.messages) > 5 {
		g.messages = g.messages[1:]
	}
}

func (g *Game) Update() error {
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			// Restart
			g.player = &Player{HP: 100, MaxHP: 100, Attack: 10, Defense: 5, Level: 1}
			g.floor = 1
			g.gameOver = false
			g.messages = make([]string, 0)
			g.generateLevel()
		}

		return nil
	}

	dx, dy := 0, 0
	moved := false

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		dy = -1
		moved = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		dy = 1
		moved = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		dx = -1
		moved = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		dx = 1
		moved = true
	}

	if moved {
		newX := g.player.X + dx
		newY := g.player.Y + dy

		// Check bounds and walls
		if newX >= 0 && newX < mapWidth && newY >= 0 && newY < mapHeight {
			tile := g.tiles[newY][newX]

			// Check for enemy
			enemy := g.getEnemyAt(newX, newY)
			if enemy != nil {
				// Attack
				damage := max(g.player.Attack-rand.Intn(5), 1)

				enemy.HP -= damage
				g.addMessage("Hit " + enemy.Name + " for " + formatInt(damage))

				if enemy.HP <= 0 {
					enemy.Dead = true
					xp := 10 + g.floor*5
					g.player.XP += xp
					g.addMessage(enemy.Name + " defeated! +" + formatInt(xp) + " XP")
					g.checkLevelUp()
				}
			} else if tile != TileWall {
				g.player.X = newX
				g.player.Y = newY

				// Check stairs
				if tile == TileStairs {
					g.floor++
					g.generateLevel()
				}

				// Check items
				for i := len(g.items) - 1; i >= 0; i-- {
					item := g.items[i]
					if item.X == g.player.X && item.Y == g.player.Y {
						g.pickupItem(item)
						g.items = append(g.items[:i], g.items[i+1:]...)
					}
				}
			}
		}

		// Enemy turns
		for _, e := range g.enemies {
			if e.Dead {
				continue
			}
			// Simple AI: move toward player if close
			edx := 0
			edy := 0

			if abs(e.X-g.player.X)+abs(e.Y-g.player.Y) <= 5 {
				if e.X < g.player.X {
					edx = 1
				} else if e.X > g.player.X {
					edx = -1
				}

				if e.Y < g.player.Y {
					edy = 1
				} else if e.Y > g.player.Y {
					edy = -1
				}
			}

			// Attack if adjacent
			if abs(e.X-g.player.X) <= 1 && abs(e.Y-g.player.Y) <= 1 {
				damage := max(e.Attack-g.player.Defense/2, 1)

				g.player.HP -= damage
				g.addMessage(e.Name + " hits you for " + formatInt(damage))

				if g.player.HP <= 0 {
					g.gameOver = true
					g.addMessage("You died!")
				}
			} else if edx != 0 || edy != 0 {
				// Move
				newX := e.X + edx

				newY := e.Y + edy
				if newX >= 0 && newX < mapWidth && newY >= 0 && newY < mapHeight &&
					g.tiles[newY][newX] == TileFloor && g.getEnemyAt(newX, newY) == nil {
					e.X = newX
					e.Y = newY
				}
			}
		}
	}

	return nil
}

func (g *Game) getEnemyAt(x, y int) *Enemy {
	for _, e := range g.enemies {
		if !e.Dead && e.X == x && e.Y == y {
			return e
		}
	}

	return nil
}

func (g *Game) pickupItem(item *Item) {
	switch item.Type {
	case ItemPotion:
		heal := item.Value

		g.player.HP += heal
		if g.player.HP > g.player.MaxHP {
			g.player.HP = g.player.MaxHP
		}

		g.addMessage("Found potion! +" + formatInt(heal) + " HP")
	case ItemWeapon:
		g.player.Attack += item.Value / 5
		g.addMessage("Found weapon! Attack +" + formatInt(item.Value/5))
	case ItemArmor:
		g.player.Defense += item.Value / 5
		g.addMessage("Found armor! Defense +" + formatInt(item.Value/5))
	case ItemGold:
		g.player.Gold += item.Value
		g.addMessage("Found " + formatInt(item.Value) + " gold!")
	}
}

func (g *Game) checkLevelUp() {
	xpNeeded := g.player.Level * 50
	if g.player.XP >= xpNeeded {
		g.player.Level++
		g.player.XP -= xpNeeded
		g.player.MaxHP += 20
		g.player.HP = g.player.MaxHP
		g.player.Attack += 3
		g.player.Defense += 2
		g.addMessage("Level up! You are now level " + formatInt(g.player.Level))
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Draw map
	for y := range mapHeight {
		for x := range mapWidth {
			screenX := float32(x * tileSize)
			screenY := float32(y * tileSize)

			switch g.tiles[y][x] {
			case TileFloor:
				vector.FillRect(
					screen,
					screenX,
					screenY,
					tileSize-1,
					tileSize-1,
					color.RGBA{R: 60, G: 60, B: 70, A: 255},
					false,
				)
			case TileWall:
				vector.FillRect(
					screen,
					screenX,
					screenY,
					tileSize-1,
					tileSize-1,
					color.RGBA{R: 40, G: 40, B: 50, A: 255},
					false,
				)
			case TileStairs:
				vector.FillRect(
					screen,
					screenX,
					screenY,
					tileSize-1,
					tileSize-1,
					color.RGBA{R: 100, G: 100, B: 50, A: 255},
					false,
				)
				ebitenutil.DebugPrintAt(screen, ">", int(screenX)+10, int(screenY)+8)
			}
		}
	}

	// Draw items
	for _, item := range g.items {
		screenX := float32(item.X*tileSize) + tileSize/2
		screenY := float32(item.Y*tileSize) + tileSize/2

		var c color.RGBA

		switch item.Type {
		case ItemPotion:
			c = color.RGBA{R: 255, G: 100, B: 100, A: 255}
		case ItemWeapon:
			c = color.RGBA{R: 200, G: 200, B: 200, A: 255}
		case ItemArmor:
			c = color.RGBA{R: 100, G: 150, B: 200, A: 255}
		case ItemGold:
			c = color.RGBA{R: 255, G: 215, B: 0, A: 255}
		}

		vector.FillCircle(screen, screenX, screenY, 8, c, false)
	}

	// Draw enemies
	for _, e := range g.enemies {
		if e.Dead {
			continue
		}

		screenX := float32(e.X*tileSize) + tileSize/2
		screenY := float32(e.Y*tileSize) + tileSize/2
		vector.FillCircle(screen, screenX, screenY, 12, color.RGBA{R: 200, G: 50, B: 50, A: 255}, false)
		// Health bar
		hpRatio := float32(e.HP) / float32(e.MaxHP)
		vector.FillRect(
			screen,
			float32(e.X*tileSize),
			float32(e.Y*tileSize)-4,
			tileSize*hpRatio,
			3,
			color.RGBA{R: 255, G: 50, B: 50, A: 255},
			false,
		)
	}

	// Draw player
	playerX := float32(g.player.X*tileSize) + tileSize/2
	playerY := float32(g.player.Y*tileSize) + tileSize/2
	vector.FillCircle(screen, playerX, playerY, 12, color.RGBA{R: 50, G: 150, B: 255, A: 255}, false)

	// UI - Stats
	vector.FillRect(
		screen,
		0,
		screenHeight-100,
		screenWidth,
		100,
		color.RGBA{R: 30, G: 30, B: 40, A: 255},
		false,
	)

	ebitenutil.DebugPrintAt(
		screen,
		"HP: "+formatInt(g.player.HP)+"/"+formatInt(g.player.MaxHP),
		10,
		screenHeight-95,
	)
	ebitenutil.DebugPrintAt(
		screen,
		"Atk: "+formatInt(g.player.Attack)+" Def: "+formatInt(g.player.Defense),
		10,
		screenHeight-75,
	)
	ebitenutil.DebugPrintAt(
		screen,
		"Lv: "+formatInt(g.player.Level)+" XP: "+formatInt(g.player.XP)+"/"+formatInt(g.player.Level*50),
		10,
		screenHeight-55,
	)
	ebitenutil.DebugPrintAt(
		screen,
		"Floor: "+formatInt(g.floor)+" Gold: "+formatInt(g.player.Gold),
		10,
		screenHeight-35,
	)

	// Messages
	for i, msg := range g.messages {
		ebitenutil.DebugPrintAt(screen, msg, 250, screenHeight-95+i*15)
	}

	// Game over
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
		ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-40, screenHeight/2-20)
		ebitenutil.DebugPrintAt(
			screen,
			"You reached floor "+formatInt(g.floor),
			screenWidth/2-60,
			screenHeight/2,
		)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to restart", screenWidth/2-80, screenHeight/2+30)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
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
	ebiten.SetWindowTitle("Roguelike Dungeon")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
