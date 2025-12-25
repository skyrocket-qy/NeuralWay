package main

import (
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 500
	screenHeight = 580
	gridSize     = 16
	cellSize     = 28
	mineCount    = 40
	gridOffsetX  = 26
	gridOffsetY  = 80
)

// CellState represents the state of a cell.
type CellState int

const (
	StateHidden CellState = iota
	StateRevealed
	StateFlagged
)

// Cell represents a single cell.
type Cell struct {
	IsMine   bool
	State    CellState
	Adjacent int
}

// Game represents the minesweeper game.
type Game struct {
	grid       [gridSize][gridSize]*Cell
	gameOver   bool
	won        bool
	firstClick bool
	flagCount  int
	startTime  time.Time
	elapsed    int
}

// NewGame creates a new game.
func NewGame() *Game {
	g := &Game{
		firstClick: true,
	}
	g.initGrid()

	return g
}

func (g *Game) initGrid() {
	for i := range gridSize {
		for j := range gridSize {
			g.grid[i][j] = &Cell{}
		}
	}
}

func (g *Game) placeMines(excludeRow, excludeCol int) {
	placed := 0
	for placed < mineCount {
		r := rand.Intn(gridSize)
		c := rand.Intn(gridSize)

		// Don't place on first click or already mine
		if (r == excludeRow && c == excludeCol) || g.grid[r][c].IsMine {
			continue
		}

		g.grid[r][c].IsMine = true
		placed++
	}

	// Calculate adjacent counts
	for i := range gridSize {
		for j := range gridSize {
			if g.grid[i][j].IsMine {
				continue
			}

			count := 0

			for di := -1; di <= 1; di++ {
				for dj := -1; dj <= 1; dj++ {
					ni, nj := i+di, j+dj
					if ni >= 0 && ni < gridSize && nj >= 0 && nj < gridSize {
						if g.grid[ni][nj].IsMine {
							count++
						}
					}
				}
			}

			g.grid[i][j].Adjacent = count
		}
	}
}

func (g *Game) reset() {
	g.initGrid()
	g.gameOver = false
	g.won = false
	g.firstClick = true
	g.flagCount = 0
	g.elapsed = 0
}

func (g *Game) Update() error {
	if !g.gameOver && !g.won && !g.firstClick {
		g.elapsed = int(time.Since(g.startTime).Seconds())
	}

	if g.gameOver || g.won {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) ||
			inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.reset()
		}

		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()

		row, col := g.screenToGrid(mx, my)
		if row >= 0 && row < gridSize && col >= 0 && col < gridSize {
			g.reveal(row, col)
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()

		row, col := g.screenToGrid(mx, my)
		if row >= 0 && row < gridSize && col >= 0 && col < gridSize {
			g.toggleFlag(row, col)
		}
	}

	return nil
}

func (g *Game) screenToGrid(mx, my int) (int, int) {
	col := (mx - gridOffsetX) / cellSize
	row := (my - gridOffsetY) / cellSize

	return row, col
}

func (g *Game) reveal(row, col int) {
	cell := g.grid[row][col]

	if cell.State == StateFlagged || cell.State == StateRevealed {
		return
	}

	if g.firstClick {
		g.firstClick = false
		g.placeMines(row, col)
		g.startTime = time.Now()
	}

	cell.State = StateRevealed

	if cell.IsMine {
		g.gameOver = true
		g.revealAllMines()

		return
	}

	// Flood fill for empty cells
	if cell.Adjacent == 0 {
		for di := -1; di <= 1; di++ {
			for dj := -1; dj <= 1; dj++ {
				ni, nj := row+di, col+dj
				if ni >= 0 && ni < gridSize && nj >= 0 && nj < gridSize {
					if g.grid[ni][nj].State == StateHidden {
						g.reveal(ni, nj)
					}
				}
			}
		}
	}

	g.checkWin()
}

func (g *Game) toggleFlag(row, col int) {
	cell := g.grid[row][col]
	if cell.State == StateRevealed {
		return
	}

	if cell.State == StateFlagged {
		cell.State = StateHidden
		g.flagCount--
	} else {
		cell.State = StateFlagged
		g.flagCount++
	}
}

func (g *Game) revealAllMines() {
	for i := range gridSize {
		for j := range gridSize {
			if g.grid[i][j].IsMine {
				g.grid[i][j].State = StateRevealed
			}
		}
	}
}

func (g *Game) checkWin() {
	for i := range gridSize {
		for j := range gridSize {
			cell := g.grid[i][j]
			if !cell.IsMine && cell.State != StateRevealed {
				return
			}
		}
	}

	g.won = true
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 192, G: 192, B: 192, A: 255})

	// Header
	vector.FillRect(screen, 0, 0, screenWidth, 70, color.RGBA{R: 150, G: 150, B: 150, A: 255}, false)

	// Mine counter
	g.drawCounter(screen, 20, 15, mineCount-g.flagCount)

	// Timer
	g.drawCounter(screen, screenWidth-100, 15, g.elapsed)

	// Reset button
	g.drawResetButton(screen)

	// Grid border
	vector.FillRect(screen, float32(gridOffsetX-3), float32(gridOffsetY-3),
		float32(gridSize*cellSize+6), float32(gridSize*cellSize+6),
		color.RGBA{R: 128, G: 128, B: 128, A: 255}, false)

	// Draw cells
	for i := range gridSize {
		for j := range gridSize {
			g.drawCell(screen, i, j)
		}
	}

	// Game over / win message
	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "GAME OVER - Press R to restart", 120, screenHeight-25)
	} else if g.won {
		ebitenutil.DebugPrintAt(screen, "YOU WIN! - Press R to restart", 130, screenHeight-25)
	} else {
		ebitenutil.DebugPrintAt(screen, "Left click = Reveal | Right click = Flag", 100, screenHeight-25)
	}
}

func (g *Game) drawCounter(screen *ebiten.Image, x, y, value int) {
	vector.FillRect(screen, float32(x), float32(y), 80, 40, color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)

	text := "000"
	if value >= 0 && value < 1000 {
		text = ""
		if value < 100 {
			text += "0"
		}

		if value < 10 {
			text += "0"
		}

		for v := value; v > 0; v /= 10 {
			text = string(rune('0'+v%10)) + text[len(text)-(len(text)-1):]
		}

		if value == 0 {
			text = "000"
		} else if value < 10 {
			text = "00" + string(rune('0'+value))
		} else if value < 100 {
			text = "0" + string(rune('0'+value/10)) + string(rune('0'+value%10))
		} else {
			text = string(rune('0'+value/100)) + string(rune('0'+(value/10)%10)) + string(rune('0'+value%10))
		}
	}

	ebitenutil.DebugPrintAt(screen, text, x+25, y+12)
}

func (g *Game) drawResetButton(screen *ebiten.Image) {
	btnX := float32(screenWidth/2 - 20)
	btnY := float32(15)
	vector.FillRect(screen, btnX, btnY, 40, 40, color.RGBA{R: 200, G: 200, B: 200, A: 255}, false)
	vector.StrokeRect(screen, btnX, btnY, 40, 40, 2, color.RGBA{R: 100, G: 100, B: 100, A: 255}, false)

	// Face
	face := ":)"
	if g.gameOver {
		face = "X("
	} else if g.won {
		face = ":D"
	}

	ebitenutil.DebugPrintAt(screen, face, int(btnX)+12, int(btnY)+12)
}

func (g *Game) drawCell(screen *ebiten.Image, row, col int) {
	cell := g.grid[row][col]
	x := float32(gridOffsetX + col*cellSize)
	y := float32(gridOffsetY + row*cellSize)

	switch cell.State {
	case StateHidden:
		// Raised button look
		vector.FillRect(
			screen,
			x,
			y,
			cellSize-1,
			cellSize-1,
			color.RGBA{R: 192, G: 192, B: 192, A: 255},
			false,
		)
		vector.FillRect(screen, x, y, cellSize-1, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
		vector.FillRect(screen, x, y, 2, cellSize-1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
		vector.FillRect(
			screen,
			x+float32(cellSize-2),
			y,
			2,
			cellSize-1,
			color.RGBA{R: 128, G: 128, B: 128, A: 255},
			false,
		)
		vector.FillRect(
			screen,
			x,
			y+float32(cellSize-2),
			cellSize-1,
			2,
			color.RGBA{R: 128, G: 128, B: 128, A: 255},
			false,
		)
	case StateFlagged:
		// Same as hidden but with flag
		vector.FillRect(
			screen,
			x,
			y,
			cellSize-1,
			cellSize-1,
			color.RGBA{R: 192, G: 192, B: 192, A: 255},
			false,
		)
		ebitenutil.DebugPrintAt(screen, "F", int(x)+10, int(y)+6)
	default:
		// Revealed
		vector.FillRect(
			screen,
			x,
			y,
			cellSize-1,
			cellSize-1,
			color.RGBA{R: 180, G: 180, B: 180, A: 255},
			false,
		)
		vector.StrokeRect(
			screen,
			x,
			y,
			cellSize-1,
			cellSize-1,
			1,
			color.RGBA{R: 128, G: 128, B: 128, A: 255},
			false,
		)

		if cell.IsMine {
			vector.FillCircle(
				screen,
				x+cellSize/2,
				y+cellSize/2,
				8,
				color.RGBA{R: 0, G: 0, B: 0, A: 255},
				false,
			)
		} else if cell.Adjacent > 0 {
			colors := []color.RGBA{
				{R: 0, G: 0, B: 255, A: 255},     // 1 - Blue
				{R: 0, G: 128, B: 0, A: 255},     // 2 - Green
				{R: 255, G: 0, B: 0, A: 255},     // 3 - Red
				{R: 0, G: 0, B: 128, A: 255},     // 4 - Dark Blue
				{R: 128, G: 0, B: 0, A: 255},     // 5 - Maroon
				{R: 0, G: 128, B: 128, A: 255},   // 6 - Teal
				{R: 0, G: 0, B: 0, A: 255},       // 7 - Black
				{R: 128, G: 128, B: 128, A: 255}, // 8 - Gray
			}
			_ = colors // TODO: use for colored numbers
			text := string(rune('0' + cell.Adjacent))
			ebitenutil.DebugPrintAt(screen, text, int(x)+10, int(y)+6)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Minesweeper")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
