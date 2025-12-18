package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// CardEffect defines the type of effect a card has.
type CardEffect int

const (
	CardEffectDamage CardEffect = iota
	CardEffectRange
	CardEffectSpeed
	CardEffectHeal
	CardEffectGold
	CardEffectExp
)

// Card represents a rogue-like upgrade card.
type Card struct {
	Name        string
	Description string
	Effect      CardEffect
	Value       int
	Rarity      CardRarity
}

// CardRarity defines card rarity levels.
type CardRarity int

const (
	RarityCommon CardRarity = iota
	RarityUncommon
	RarityRare
	RarityLegendary
)

// RarityColors for card rendering.
var RarityColors = map[CardRarity]color.RGBA{
	RarityCommon:    {R: 180, G: 180, B: 180, A: 255}, // Gray
	RarityUncommon:  {R: 0, G: 200, B: 100, A: 255},   // Green
	RarityRare:      {R: 50, G: 150, B: 255, A: 255},  // Blue
	RarityLegendary: {R: 255, G: 180, B: 0, A: 255},   // Gold
}

// Predefined card pool.
var CardPool = []Card{
	// Common cards
	{Name: "Sharp Blade", Description: "+5 Attack Damage", Effect: CardEffectDamage, Value: 5, Rarity: RarityCommon},
	{Name: "Keen Eyes", Description: "+10 Attack Range", Effect: CardEffectRange, Value: 10, Rarity: RarityCommon},
	{Name: "Quick Hands", Description: "+10% Attack Speed", Effect: CardEffectSpeed, Value: 10, Rarity: RarityCommon},
	{Name: "Gold Pouch", Description: "+25 Gold", Effect: CardEffectGold, Value: 25, Rarity: RarityCommon},

	// Uncommon cards
	{Name: "Forged Steel", Description: "+10 Attack Damage", Effect: CardEffectDamage, Value: 10, Rarity: RarityUncommon},
	{Name: "Eagle Vision", Description: "+25 Attack Range", Effect: CardEffectRange, Value: 25, Rarity: RarityUncommon},
	{Name: "Battle Frenzy", Description: "+25% Attack Speed", Effect: CardEffectSpeed, Value: 25, Rarity: RarityUncommon},
	{Name: "Training Manual", Description: "+50 Experience", Effect: CardEffectExp, Value: 50, Rarity: RarityUncommon},

	// Rare cards
	{Name: "Legendary Sword", Description: "+20 Attack Damage", Effect: CardEffectDamage, Value: 20, Rarity: RarityRare},
	{Name: "Sniper Scope", Description: "+50 Attack Range", Effect: CardEffectRange, Value: 50, Rarity: RarityRare},
	{Name: "Berserker Rage", Description: "+50% Attack Speed", Effect: CardEffectSpeed, Value: 50, Rarity: RarityRare},

	// Legendary cards
	{Name: "Divine Blade", Description: "+40 Attack Damage", Effect: CardEffectDamage, Value: 40, Rarity: RarityLegendary},
	{Name: "All-Seeing Eye", Description: "+100 Attack Range", Effect: CardEffectRange, Value: 100, Rarity: RarityLegendary},
}

// CardSelector handles the card selection UI.
type CardSelector struct {
	Cards       [3]*Card
	Selected    int
	Active      bool
	CardWidth   int
	CardHeight  int
	CardSpacing int
}

// NewCardSelector creates a card selector.
func NewCardSelector() *CardSelector {
	return &CardSelector{
		Selected:    -1,
		Active:      false,
		CardWidth:   150,
		CardHeight:  200,
		CardSpacing: 20,
	}
}

// GenerateChoices generates 3 random cards with weighted rarity.
func (s *CardSelector) GenerateChoices(waveNumber int) {
	// Higher waves = better chances at rare cards
	for i := 0; i < 3; i++ {
		s.Cards[i] = s.pickRandomCard(waveNumber)
	}
	s.Selected = -1
	s.Active = true
}

// pickRandomCard picks a random card with rarity weighting.
func (s *CardSelector) pickRandomCard(waveNumber int) *Card {
	// Simple weighting based on wave number
	// In a real game, use proper random with weights
	maxRarity := RarityCommon
	if waveNumber >= 2 {
		maxRarity = RarityUncommon
	}
	if waveNumber >= 4 {
		maxRarity = RarityRare
	}
	if waveNumber >= 6 {
		maxRarity = RarityLegendary
	}

	// Find eligible cards
	var eligible []Card
	for _, card := range CardPool {
		if card.Rarity <= maxRarity {
			eligible = append(eligible, card)
		}
	}

	if len(eligible) == 0 {
		return &CardPool[0]
	}

	// Pick random (simple implementation - in real game use proper random)
	idx := waveNumber % len(eligible)
	return &eligible[idx]
}

// ApplyCard applies a card's effect to a hero.
func (s *CardSelector) ApplyCard(card *Card, hero *Hero) {
	switch card.Effect {
	case CardEffectDamage:
		hero.AttackDamage += card.Value
	case CardEffectRange:
		hero.AttackRange += float64(card.Value)
	case CardEffectSpeed:
		hero.AttackSpeed *= 1.0 + float64(card.Value)/100.0
	case CardEffectExp:
		hero.GainExp(card.Value)
	}
}

// Draw renders the card selection UI.
func (s *CardSelector) Draw(screen *ebiten.Image, screenWidth, screenHeight int) {
	if !s.Active {
		return
	}

	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 180})
	screen.DrawImage(overlay, nil)

	// Calculate card positions
	totalWidth := 3*s.CardWidth + 2*s.CardSpacing
	startX := (screenWidth - totalWidth) / 2
	startY := (screenHeight - s.CardHeight) / 2

	for i := 0; i < 3; i++ {
		card := s.Cards[i]
		if card == nil {
			continue
		}

		x := startX + i*(s.CardWidth+s.CardSpacing)
		y := startY

		// Card background
		cardImg := ebiten.NewImage(s.CardWidth, s.CardHeight)
		cardImg.Fill(RarityColors[card.Rarity])

		// Card border (selected)
		if i == s.Selected {
			border := ebiten.NewImage(s.CardWidth+4, s.CardHeight+4)
			border.Fill(color.RGBA{R: 255, G: 255, B: 255, A: 255})
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x-2), float64(y-2))
			screen.DrawImage(border, op)
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(cardImg, op)
	}
}

// HandleInput processes card selection input.
func (s *CardSelector) HandleInput(mouseX, mouseY int, clicked bool, screenWidth, screenHeight int) *Card {
	if !s.Active {
		return nil
	}

	totalWidth := 3*s.CardWidth + 2*s.CardSpacing
	startX := (screenWidth - totalWidth) / 2
	startY := (screenHeight - s.CardHeight) / 2

	s.Selected = -1

	for i := 0; i < 3; i++ {
		x := startX + i*(s.CardWidth+s.CardSpacing)
		y := startY

		if mouseX >= x && mouseX < x+s.CardWidth &&
			mouseY >= y && mouseY < y+s.CardHeight {
			s.Selected = i
			if clicked {
				s.Active = false
				return s.Cards[i]
			}
		}
	}

	return nil
}
