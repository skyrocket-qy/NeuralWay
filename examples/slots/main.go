package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

//go:embed assets/*.png
var assetFS embed.FS

const (
	screenWidth  = 900
	screenHeight = 650
	reelCount    = 5
	rowCount     = 4
	symbolSize   = 90
	reelWidth    = 100
	reelSpacing  = 8
)

// Symbol represents a slot symbol.
type Symbol int

const (
	SymbolDragon Symbol = iota
	SymbolPhoenix
	SymbolTreasure
	SymbolWild
	SymbolCardK
	SymbolCardA
	SymbolCardQ
	SymbolCount
)

// SymbolInfo contains symbol display data.
type SymbolInfo struct {
	Name      string
	NameCN    string
	ImageFile string
	Payout    map[int]int // count -> payout multiplier
}

var SymbolData = map[Symbol]SymbolInfo{
	SymbolDragon:   {Name: "Dragon", NameCN: "龙", ImageFile: "slot_dragon_1766041758627.png", Payout: map[int]int{3: 50, 4: 150, 5: 500}},
	SymbolPhoenix:  {Name: "Phoenix", NameCN: "凤", ImageFile: "slot_phoenix_1766041775332.png", Payout: map[int]int{3: 40, 4: 120, 5: 400}},
	SymbolTreasure: {Name: "Treasure", NameCN: "聚宝盆", ImageFile: "slot_treasure_pot_1766041794217.png", Payout: map[int]int{3: 30, 4: 100, 5: 300}},
	SymbolWild:     {Name: "Wild", NameCN: "百搭", ImageFile: "slot_wild_1766041810081.png", Payout: map[int]int{3: 100, 4: 500, 5: 2000}},
	SymbolCardK:    {Name: "K", NameCN: "K", ImageFile: "slot_card_k_1766041844584.png", Payout: map[int]int{3: 10, 4: 30, 5: 100}},
	SymbolCardA:    {Name: "A", NameCN: "A", ImageFile: "slot_card_a_1766041858660.png", Payout: map[int]int{3: 10, 4: 30, 5: 100}},
	SymbolCardQ:    {Name: "Q", NameCN: "Q", ImageFile: "slot_card_q_1766041875271.png", Payout: map[int]int{3: 5, 4: 20, 5: 75}},
}

// Reel represents a single slot reel.
type Reel struct {
	Symbols      []Symbol
	Position     float64
	TargetPos    int
	Spinning     bool
	SpinSpeed    float64
	StopDelay    float64
	StopTimer    float64
	BounceFactor float64
	BouncePhase  float64
}

// Win represents a winning combination.
type Win struct {
	Symbol Symbol
	Count  int
	Payout int
	Rows   []int
}

// SlotMachine represents the slot game.
type SlotMachine struct {
	Reels        [reelCount]*Reel
	Credits      int64
	Bet          int
	TotalWin     int64
	Wins         []Win
	Spinning     bool
	AutoSpin     bool
	DisplayGrid  [reelCount][rowCount]Symbol
	SymbolImages map[Symbol]*ebiten.Image

	// Animation
	WinFlashTimer float64
	SpinTimer     float64

	// UI positions
	reelStartX float64
	reelStartY float64
}

// NewSlotMachine creates a new slot machine.
func NewSlotMachine() *SlotMachine {
	rand.Seed(time.Now().UnixNano())

	sm := &SlotMachine{
		Credits:      9999999999,
		Bet:          20000,
		reelStartX:   float64(screenWidth-reelCount*reelWidth-(reelCount-1)*reelSpacing) / 2,
		reelStartY:   100,
		SymbolImages: make(map[Symbol]*ebiten.Image),
	}

	// Load symbol images
	sm.loadAssets()

	// Initialize reels
	for i := 0; i < reelCount; i++ {
		sm.Reels[i] = &Reel{
			Symbols:   generateReelStrip(),
			SpinSpeed: 25 + rand.Float64()*10,
			StopDelay: float64(i) * 0.25,
		}
	}

	sm.randomizeDisplay()
	return sm
}

func (sm *SlotMachine) loadAssets() {
	for sym, info := range SymbolData {
		path := "assets/" + info.ImageFile
		data, err := assetFS.ReadFile(path)
		if err != nil {
			log.Printf("Warning: Could not load %s: %v", path, err)
			// Create fallback colored rectangle
			img := ebiten.NewImage(symbolSize, symbolSize)
			colors := []color.RGBA{
				{R: 255, G: 215, B: 0, A: 255},  // Dragon - Gold
				{R: 255, G: 100, B: 0, A: 255},  // Phoenix - Orange
				{R: 218, G: 165, B: 32, A: 255}, // Treasure - Goldenrod
				{R: 255, G: 20, B: 147, A: 255}, // Wild - Pink
				{R: 128, G: 0, B: 128, A: 255},  // K - Purple
				{R: 200, G: 0, B: 0, A: 255},    // A - Red
				{R: 0, G: 100, B: 200, A: 255},  // Q - Blue
			}
			if int(sym) < len(colors) {
				img.Fill(colors[sym])
			}
			sm.SymbolImages[sym] = img
			continue
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			log.Printf("Warning: Could not decode %s: %v", path, err)
			continue
		}
		sm.SymbolImages[sym] = ebiten.NewImageFromImage(img)
	}
}

func generateReelStrip() []Symbol {
	weights := map[Symbol]int{
		SymbolDragon:   3,
		SymbolPhoenix:  4,
		SymbolTreasure: 5,
		SymbolWild:     2,
		SymbolCardK:    8,
		SymbolCardA:    8,
		SymbolCardQ:    10,
	}

	var strip []Symbol
	for sym, weight := range weights {
		for j := 0; j < weight; j++ {
			strip = append(strip, sym)
		}
	}

	rand.Shuffle(len(strip), func(i, j int) {
		strip[i], strip[j] = strip[j], strip[i]
	})

	return strip
}

func (sm *SlotMachine) randomizeDisplay() {
	for i := 0; i < reelCount; i++ {
		for j := 0; j < rowCount; j++ {
			sm.DisplayGrid[i][j] = Symbol(rand.Intn(int(SymbolCount)))
		}
	}
}

func (sm *SlotMachine) spin() {
	if sm.Spinning || sm.Credits < int64(sm.Bet) {
		return
	}

	sm.Credits -= int64(sm.Bet)
	sm.Spinning = true
	sm.TotalWin = 0
	sm.Wins = nil
	sm.SpinTimer = 0

	for i, reel := range sm.Reels {
		reel.Spinning = true
		reel.Position = 0
		reel.TargetPos = rand.Intn(len(reel.Symbols))
		reel.SpinSpeed = 25 + rand.Float64()*10
		reel.StopTimer = 0
		reel.StopDelay = float64(i) * 0.3
		reel.BounceFactor = 0
		reel.BouncePhase = 0
	}
}

func (sm *SlotMachine) Update() error {
	// Handle keyboard input
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && !sm.Spinning {
		sm.spin()
	}

	// Handle mouse click on spin button
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !sm.Spinning {
		mx, my := ebiten.CursorPosition()
		btnX := screenWidth - 200
		btnY := screenHeight - 130 + 20
		btnW := 160
		btnH := 80
		if mx >= btnX && mx <= btnX+btnW && my >= btnY && my <= btnY+btnH {
			sm.spin()
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		sm.AutoSpin = !sm.AutoSpin
	}

	// Bet adjustment
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) && !sm.Spinning {
		sm.Bet = min(sm.Bet+10000, 100000)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) && !sm.Spinning {
		sm.Bet = max(sm.Bet-10000, 10000)
	}

	dt := 1.0 / 60.0
	sm.SpinTimer += dt

	allStopped := true
	for i, reel := range sm.Reels {
		if reel.Spinning {
			allStopped = false

			reel.StopTimer += dt
			if reel.StopTimer >= reel.StopDelay {
				reel.SpinSpeed *= 0.94

				if reel.SpinSpeed < 0.5 {
					reel.Spinning = false
					reel.BounceFactor = 8

					for j := 0; j < rowCount; j++ {
						idx := (reel.TargetPos + j) % len(reel.Symbols)
						sm.DisplayGrid[i][j] = reel.Symbols[idx]
					}
				}
			}

			reel.Position += reel.SpinSpeed * dt
		} else if reel.BounceFactor > 0 {
			reel.BouncePhase += dt * 15
			reel.BounceFactor *= 0.92
			if reel.BounceFactor < 0.1 {
				reel.BounceFactor = 0
			}
		}
	}

	if sm.Spinning && allStopped {
		sm.Spinning = false
		sm.checkWins()

		if sm.AutoSpin && sm.Credits >= int64(sm.Bet) && sm.TotalWin == 0 {
			go func() {
				time.Sleep(800 * time.Millisecond)
			}()
		}
	}

	if len(sm.Wins) > 0 {
		sm.WinFlashTimer += dt
	}

	return nil
}

func (sm *SlotMachine) checkWins() {
	sm.Wins = nil
	sm.TotalWin = 0

	// Check each row for winning combinations (left to right)
	for row := 0; row < rowCount; row++ {
		firstSym := sm.DisplayGrid[0][row]
		if firstSym == SymbolWild {
			for col := 1; col < reelCount; col++ {
				if sm.DisplayGrid[col][row] != SymbolWild {
					firstSym = sm.DisplayGrid[col][row]
					break
				}
			}
		}

		count := 0
		for col := 0; col < reelCount; col++ {
			sym := sm.DisplayGrid[col][row]
			if sym == firstSym || sym == SymbolWild {
				count++
			} else {
				break
			}
		}

		if count >= 3 {
			info := SymbolData[firstSym]
			payout := int64(info.Payout[count]) * int64(sm.Bet) / 100
			if payout > 0 {
				sm.Wins = append(sm.Wins, Win{
					Symbol: firstSym,
					Count:  count,
					Payout: int(payout),
					Rows:   []int{row},
				})
				sm.TotalWin += payout
			}
		}
	}

	sm.Credits += sm.TotalWin
}

func (sm *SlotMachine) Draw(screen *ebiten.Image) {
	// Dark blue gradient background
	for y := 0; y < screenHeight; y++ {
		t := float64(y) / float64(screenHeight)
		r := uint8(10 + t*20)
		g := uint8(15 + t*25)
		b := uint8(40 + t*50)
		vector.DrawFilledRect(screen, 0, float32(y), float32(screenWidth), 1, color.RGBA{R: r, G: g, B: b, A: 255}, false)
	}

	// Draw ornate frame
	sm.drawFrame(screen)

	// Draw reels
	sm.drawReels(screen)

	// Draw UI panel
	sm.drawUI(screen)
}

func (sm *SlotMachine) drawFrame(screen *ebiten.Image) {
	frameX := sm.reelStartX - 25
	frameY := sm.reelStartY - 15
	frameW := float32(reelCount*reelWidth + (reelCount-1)*reelSpacing + 50)
	frameH := float32(rowCount*symbolSize + 30)

	// Outer gold border
	vector.DrawFilledRect(screen, float32(frameX)-8, float32(frameY)-8, frameW+16, frameH+16, color.RGBA{R: 139, G: 69, B: 19, A: 255}, false)
	vector.DrawFilledRect(screen, float32(frameX)-4, float32(frameY)-4, frameW+8, frameH+8, color.RGBA{R: 218, G: 165, B: 32, A: 255}, false)
	// Inner dark
	vector.DrawFilledRect(screen, float32(frameX), float32(frameY), frameW, frameH, color.RGBA{R: 15, G: 20, B: 45, A: 255}, false)

	// Title area - "4096 WAYS" style
	titleW := float32(100)
	titleH := float32(80)
	// Left side
	vector.DrawFilledRect(screen, float32(frameX)-titleW-5, float32(frameY)+frameH/2-titleH/2, titleW, titleH, color.RGBA{R: 139, G: 69, B: 19, A: 255}, false)
	vector.DrawFilledRect(screen, float32(frameX)-titleW, float32(frameY)+frameH/2-titleH/2+5, titleW-10, titleH-10, color.RGBA{R: 218, G: 165, B: 32, A: 255}, false)
	// Right side
	vector.DrawFilledRect(screen, float32(frameX)+frameW+5, float32(frameY)+frameH/2-titleH/2, titleW, titleH, color.RGBA{R: 139, G: 69, B: 19, A: 255}, false)
	vector.DrawFilledRect(screen, float32(frameX)+frameW+10, float32(frameY)+frameH/2-titleH/2+5, titleW-10, titleH-10, color.RGBA{R: 218, G: 165, B: 32, A: 255}, false)
}
func (sm *SlotMachine) drawReels(screen *ebiten.Image) {
	for i, reel := range sm.Reels {
		x := sm.reelStartX + float64(i)*(reelWidth+reelSpacing)

		// Reel background
		vector.DrawFilledRect(screen, float32(x), float32(sm.reelStartY), float32(reelWidth), float32(rowCount*symbolSize), color.RGBA{R: 25, G: 30, B: 55, A: 255}, false)

		symX := x + (reelWidth-symbolSize)/2

		if reel.Spinning {
			// Smooth scrolling during spin
			// Position increases continuously, we use modulo to wrap
			baseOffset := math.Mod(reel.Position*100, float64(symbolSize))

			// Draw extra symbols for seamless scrolling
			for j := -1; j <= rowCount+1; j++ {
				// Calculate which symbol to show based on position
				symIdx := int(reel.Position*2) + j
				symIdx = ((symIdx % len(reel.Symbols)) + len(reel.Symbols)) % len(reel.Symbols)
				sym := reel.Symbols[symIdx]
				symImg := sm.SymbolImages[sym]

				symY := sm.reelStartY + float64(j)*symbolSize + baseOffset

				// Only draw if visible
				if symY >= sm.reelStartY-symbolSize && symY <= sm.reelStartY+float64(rowCount)*symbolSize+symbolSize {
					if symImg != nil {
						op := &ebiten.DrawImageOptions{}
						imgW := float64(symImg.Bounds().Dx())
						imgH := float64(symImg.Bounds().Dy())
						scaleX := float64(symbolSize-4) / imgW
						scaleY := float64(symbolSize-6) / imgH
						op.GeoM.Scale(scaleX, scaleY)
						op.GeoM.Translate(symX+2, symY+2)

						// Motion blur when spinning fast
						if reel.SpinSpeed > 15 {
							op.ColorScale.ScaleAlpha(0.6)
						} else if reel.SpinSpeed > 5 {
							op.ColorScale.ScaleAlpha(0.8)
						}

						screen.DrawImage(symImg, op)
					}
				}
			}
		} else {
			// Stopped - show final symbols with bounce
			for j := 0; j < rowCount; j++ {
				sym := sm.DisplayGrid[i][j]
				symImg := sm.SymbolImages[sym]
				symY := sm.reelStartY + float64(j)*symbolSize

				// Bounce animation after stop
				bounceOffset := 0.0
				if reel.BounceFactor > 0 {
					bounceOffset = math.Sin(reel.BouncePhase) * reel.BounceFactor
				}

				if symImg != nil {
					op := &ebiten.DrawImageOptions{}
					imgW := float64(symImg.Bounds().Dx())
					imgH := float64(symImg.Bounds().Dy())
					scaleX := float64(symbolSize-4) / imgW
					scaleY := float64(symbolSize-6) / imgH
					op.GeoM.Scale(scaleX, scaleY)
					op.GeoM.Translate(symX+2, symY+bounceOffset+2)
					screen.DrawImage(symImg, op)
				}

				// Symbol border
				vector.StrokeRect(screen, float32(symX), float32(symY+bounceOffset), float32(symbolSize), float32(symbolSize-4), 1, color.RGBA{R: 100, G: 100, B: 120, A: 100}, false)
			}
		}

		// Reel mask - cover overflow at top and bottom
		vector.DrawFilledRect(screen, float32(x)-2, float32(sm.reelStartY)-symbolSize-10, float32(reelWidth)+4, float32(symbolSize)+12, color.RGBA{R: 15, G: 20, B: 45, A: 255}, false)
		vector.DrawFilledRect(screen, float32(x)-2, float32(sm.reelStartY)+float32(rowCount*symbolSize)-2, float32(reelWidth)+4, float32(symbolSize)+12, color.RGBA{R: 15, G: 20, B: 45, A: 255}, false)

		// Reel divider line
		if i < reelCount-1 {
			divX := x + reelWidth + reelSpacing/2
			vector.DrawFilledRect(screen, float32(divX), float32(sm.reelStartY), 2, float32(rowCount*symbolSize), color.RGBA{R: 218, G: 165, B: 32, A: 150}, false)
		}
	}

	// Win highlighting
	if len(sm.Wins) > 0 {
		alpha := uint8(100 + 100*math.Sin(sm.WinFlashTimer*8))
		for _, win := range sm.Wins {
			for _, row := range win.Rows {
				for col := 0; col < win.Count; col++ {
					x := sm.reelStartX + float64(col)*(reelWidth+reelSpacing)
					y := sm.reelStartY + float64(row)*symbolSize
					vector.StrokeRect(screen, float32(x), float32(y), float32(reelWidth), float32(symbolSize), 3, color.RGBA{R: 255, G: 215, B: 0, A: alpha}, false)
				}
			}
		}
	}
}

func (sm *SlotMachine) drawUI(screen *ebiten.Image) {
	// Bottom panel
	panelY := float32(screenHeight - 130)
	vector.DrawFilledRect(screen, 0, panelY, float32(screenWidth), 130, color.RGBA{R: 30, G: 20, B: 50, A: 255}, false)
	vector.DrawFilledRect(screen, 0, panelY, float32(screenWidth), 4, color.RGBA{R: 218, G: 165, B: 32, A: 255}, false)

	// Credits display - 押分
	sm.drawInfoPanel(screen, 30, int(panelY)+15, "押分", fmt.Sprintf("%d", sm.Bet), color.RGBA{R: 255, G: 255, B: 0, A: 255})

	// Total credits - 总分
	sm.drawInfoPanel(screen, 200, int(panelY)+15, "总分", sm.formatNumber(sm.Credits), color.RGBA{R: 0, G: 255, B: 150, A: 255})

	// Win display - 赢分
	winColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if sm.TotalWin > 0 {
		flash := uint8(150 + 105*math.Sin(sm.WinFlashTimer*10))
		winColor = color.RGBA{R: 255, G: flash, B: 0, A: 255}
	}
	sm.drawInfoPanel(screen, 420, int(panelY)+15, "赢分", sm.formatNumber(sm.TotalWin), winColor)

	// Spin button - 转动
	btnX := float32(screenWidth - 200)
	btnY := panelY + 20
	btnW := float32(160)
	btnH := float32(80)

	btnColor := color.RGBA{R: 180, G: 50, B: 50, A: 255}
	btnInner := color.RGBA{R: 220, G: 80, B: 80, A: 255}
	if sm.Spinning {
		btnColor = color.RGBA{R: 80, G: 80, B: 80, A: 255}
		btnInner = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	}

	// Button with 3D effect
	vector.DrawFilledRect(screen, btnX, btnY+5, btnW, btnH, color.RGBA{R: 100, G: 30, B: 30, A: 255}, false)
	vector.DrawFilledRect(screen, btnX, btnY, btnW, btnH-5, btnColor, false)
	vector.DrawFilledRect(screen, btnX+5, btnY+5, btnW-10, btnH-15, btnInner, false)
	vector.StrokeRect(screen, btnX, btnY, btnW, btnH, 2, color.RGBA{R: 218, G: 165, B: 32, A: 255}, false)

	// Spin button text
	spinText := "转动 SPIN"
	if sm.Spinning {
		spinText = "..."
	}
	ebitenutil.DebugPrintAt(screen, spinText, int(btnX)+45, int(btnY)+30)

	// Auto-spin indicator
	if sm.AutoSpin {
		vector.DrawFilledCircle(screen, 620, panelY+55, 12, color.RGBA{R: 0, G: 255, B: 0, A: 255}, false)
		vector.StrokeCircle(screen, 620, panelY+55, 12, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
		ebitenutil.DebugPrintAt(screen, "AUTO", 640, int(panelY)+50)
	}

	// Controls hint at very bottom
	vector.DrawFilledRect(screen, 0, float32(screenHeight-25), float32(screenWidth), 25, color.RGBA{R: 0, G: 0, B: 0, A: 180}, false)
	ebitenutil.DebugPrintAt(screen, "SPACE=Spin | UP/DOWN=Bet | A=Auto", 300, screenHeight-20)
}

func (sm *SlotMachine) drawInfoPanel(screen *ebiten.Image, x, y int, label, value string, valueColor color.RGBA) {
	panelW := float32(160)
	panelH := float32(80)

	// Panel background
	vector.DrawFilledRect(screen, float32(x), float32(y), panelW, panelH, color.RGBA{R: 40, G: 30, B: 60, A: 255}, false)
	vector.StrokeRect(screen, float32(x), float32(y), panelW, panelH, 2, color.RGBA{R: 218, G: 165, B: 32, A: 255}, false)

	// Label area
	vector.DrawFilledRect(screen, float32(x)+5, float32(y)+5, panelW-10, 25, color.RGBA{R: 60, G: 50, B: 80, A: 255}, false)

	// Value area
	vector.DrawFilledRect(screen, float32(x)+10, float32(y)+35, panelW-20, 35, color.RGBA{R: 20, G: 15, B: 35, A: 255}, false)

	// Draw label text
	ebitenutil.DebugPrintAt(screen, label, x+60, y+9)

	// Draw value text (larger position)
	ebitenutil.DebugPrintAt(screen, value, x+30, y+45)
}

func (sm *SlotMachine) formatNumber(n int64) string {
	if n >= 1000000000 {
		return fmt.Sprintf("%.1fB", float64(n)/1000000000)
	}
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func (sm *SlotMachine) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("财神到 - Fortune Arrives")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewSlotMachine()); err != nil {
		log.Fatal(err)
	}
}
