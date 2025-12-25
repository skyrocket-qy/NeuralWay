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
	screenWidth  = 600
	screenHeight = 500
)

// Card represents a playing card.
type Card struct {
	Suit  int // 0=Hearts, 1=Diamonds, 2=Clubs, 3=Spades
	Value int // 1-13 (Ace=1, J=11, Q=12, K=13)
}

func (c Card) Name() string {
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	suits := []string{"♥", "♦", "♣", "♠"}

	return values[c.Value-1] + suits[c.Suit]
}

func (c Card) BJValue() int {
	if c.Value >= 10 {
		return 10
	}

	return c.Value
}

// Hand represents a hand of cards.
type Hand struct {
	Cards []Card
}

func (h *Hand) Add(c Card) {
	h.Cards = append(h.Cards, c)
}

func (h *Hand) Value() int {
	total := 0
	aces := 0

	for _, c := range h.Cards {
		if c.Value == 1 {
			aces++
			total += 11
		} else {
			total += c.BJValue()
		}
	}
	// Convert aces from 11 to 1 if busting
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}

	return total
}

func (h *Hand) IsBusted() bool {
	return h.Value() > 21
}

func (h *Hand) IsBlackjack() bool {
	return len(h.Cards) == 2 && h.Value() == 21
}

// Game represents the blackjack game.
type Game struct {
	playerHand *Hand
	dealerHand *Hand
	deck       []Card
	chips      int
	bet        int
	gameState  int // 0=betting, 1=playing, 2=dealer turn, 3=result
	message    string
	wins       int
	losses     int
}

// NewGame creates a new game.
func NewGame() *Game {
	g := &Game{
		chips: 1000,
		bet:   100,
	}

	return g
}

func (g *Game) shuffleDeck() {
	g.deck = make([]Card, 0, 52)

	for suit := range 4 {
		for value := 1; value <= 13; value++ {
			g.deck = append(g.deck, Card{Suit: suit, Value: value})
		}
	}

	rand.Shuffle(len(g.deck), func(i, j int) {
		g.deck[i], g.deck[j] = g.deck[j], g.deck[i]
	})
}

func (g *Game) drawCard() Card {
	if len(g.deck) == 0 {
		g.shuffleDeck()
	}

	card := g.deck[len(g.deck)-1]
	g.deck = g.deck[:len(g.deck)-1]

	return card
}

func (g *Game) startRound() {
	if g.chips < g.bet {
		g.message = "Not enough chips!"

		return
	}

	g.chips -= g.bet
	g.shuffleDeck()
	g.playerHand = &Hand{}
	g.dealerHand = &Hand{}

	g.playerHand.Add(g.drawCard())
	g.dealerHand.Add(g.drawCard())
	g.playerHand.Add(g.drawCard())
	g.dealerHand.Add(g.drawCard())

	g.gameState = 1
	g.message = ""

	// Check for blackjack
	if g.playerHand.IsBlackjack() {
		g.dealerTurn()
	}
}

func (g *Game) hit() {
	g.playerHand.Add(g.drawCard())

	if g.playerHand.IsBusted() {
		g.endRound("Bust! You lose.")
		g.losses++
	}
}

func (g *Game) stand() {
	g.dealerTurn()
}

func (g *Game) dealerTurn() {
	g.gameState = 2
	// Dealer draws until 17
	for g.dealerHand.Value() < 17 {
		g.dealerHand.Add(g.drawCard())
	}

	g.determineWinner()
}

func (g *Game) determineWinner() {
	playerVal := g.playerHand.Value()
	dealerVal := g.dealerHand.Value()

	if g.playerHand.IsBlackjack() && !g.dealerHand.IsBlackjack() {
		g.endRound("Blackjack! You win 3:2!")
		g.chips += int(float64(g.bet) * 2.5)
		g.wins++
	} else if g.dealerHand.IsBusted() {
		g.endRound("Dealer busts! You win!")
		g.chips += g.bet * 2
		g.wins++
	} else if playerVal > dealerVal {
		g.endRound("You win!")
		g.chips += g.bet * 2
		g.wins++
	} else if playerVal < dealerVal {
		g.endRound("Dealer wins!")
		g.losses++
	} else {
		g.endRound("Push - tie!")
		g.chips += g.bet
	}
}

func (g *Game) endRound(msg string) {
	g.message = msg
	g.gameState = 3
}

func (g *Game) Update() error {
	switch g.gameState {
	case 0: // Betting
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.bet += 50
			if g.bet > g.chips {
				g.bet = g.chips
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.bet -= 50
			if g.bet < 50 {
				g.bet = 50
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.startRound()
		}
	case 1: // Playing
		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			g.hit()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.stand()
		}
	case 3: // Result
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if g.chips > 0 {
				g.gameState = 0
			} else {
				g.chips = 1000
				g.wins = 0
				g.losses = 0
				g.gameState = 0
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 0, G: 100, B: 50, A: 255})

	// Table felt pattern
	vector.FillRect(
		screen,
		50,
		50,
		screenWidth-100,
		screenHeight-100,
		color.RGBA{R: 0, G: 120, B: 60, A: 255},
		false,
	)
	vector.StrokeRect(
		screen,
		50,
		50,
		screenWidth-100,
		screenHeight-100,
		3,
		color.RGBA{R: 139, G: 90, B: 43, A: 255},
		false,
	)

	// Dealer area
	ebitenutil.DebugPrintAt(screen, "DEALER", 270, 70)

	if g.dealerHand != nil {
		g.drawHand(screen, g.dealerHand, 150, 100, g.gameState < 2)

		if g.gameState >= 2 {
			ebitenutil.DebugPrintAt(screen, "Value: "+formatInt(g.dealerHand.Value()), 420, 130)
		}
	}

	// Player area
	ebitenutil.DebugPrintAt(screen, "PLAYER", 270, 250)

	if g.playerHand != nil {
		g.drawHand(screen, g.playerHand, 150, 280, false)
		ebitenutil.DebugPrintAt(screen, "Value: "+formatInt(g.playerHand.Value()), 420, 310)
	}

	// UI panel
	vector.FillRect(
		screen,
		0,
		screenHeight-80,
		screenWidth,
		80,
		color.RGBA{R: 40, G: 30, B: 20, A: 255},
		false,
	)

	// Chips and bet
	ebitenutil.DebugPrintAt(screen, "Chips: $"+formatInt(g.chips), 20, screenHeight-65)
	ebitenutil.DebugPrintAt(screen, "Bet: $"+formatInt(g.bet), 20, screenHeight-45)
	ebitenutil.DebugPrintAt(screen, "W: "+formatInt(g.wins)+" L: "+formatInt(g.losses), 20, screenHeight-25)

	// Controls
	switch g.gameState {
	case 0:
		ebitenutil.DebugPrintAt(screen, "UP/DOWN = Adjust Bet | SPACE = Deal", 200, screenHeight-45)
	case 1:
		ebitenutil.DebugPrintAt(screen, "H = Hit | S = Stand", 250, screenHeight-45)
	case 3:
		ebitenutil.DebugPrintAt(screen, g.message+" | SPACE = Continue", 180, screenHeight-45)
	}
}

func (g *Game) drawHand(screen *ebiten.Image, hand *Hand, startX, y int, hideSecond bool) {
	for i, card := range hand.Cards {
		x := startX + i*60
		g.renderCard(screen, card, x, y, hideSecond && i == 1)
	}
}

func (g *Game) renderCard(screen *ebiten.Image, card Card, x, y int, faceDown bool) {
	cardW := float32(50)
	cardH := float32(70)

	// Card background
	if faceDown {
		vector.FillRect(
			screen,
			float32(x),
			float32(y),
			cardW,
			cardH,
			color.RGBA{R: 50, G: 50, B: 200, A: 255},
			false,
		)
		vector.StrokeRect(
			screen,
			float32(x),
			float32(y),
			cardW,
			cardH,
			2,
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
			false,
		)
		ebitenutil.DebugPrintAt(screen, "?", x+20, y+28)
	} else {
		vector.FillRect(screen, float32(x), float32(y), cardW, cardH, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
		vector.StrokeRect(screen, float32(x), float32(y), cardW, cardH, 2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)

		// Card text
		cardText := card.Name()

		textColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
		if card.Suit == 0 || card.Suit == 1 {
			textColor = color.RGBA{R: 200, G: 0, B: 0, A: 255}
		}

		_ = textColor // Note: ebitenutil uses fixed color

		ebitenutil.DebugPrintAt(screen, cardText, x+10, y+28)
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
	ebiten.SetWindowTitle("Blackjack")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
