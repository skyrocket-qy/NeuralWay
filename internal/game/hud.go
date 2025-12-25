package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// HUDElement represents a UI element type.
type HUDElement int

const (
	HUDHealthBar HUDElement = iota
	HUDManaBar
	HUDXPBar
	HUDMinimapElem
	HUDTextElem
	HUDIconElem
)

// HUDAlign represents alignment for UI elements.
type HUDAlign int

const (
	AlignLeft HUDAlign = iota
	AlignCenter
	AlignRight
)

// HUDBar represents a progress bar.
type HUDBar struct {
	X, Y          float64
	Width, Height float64
	Current       float64 // 0.0 - 1.0
	FillColor     color.RGBA
	BackColor     color.RGBA
	BorderColor   color.RGBA
	ShowText      bool
	TextFormat    string // e.g., "%d/%d" or "%.0f%%"
	Visible       bool
}

// NewHUDBar creates a new bar.
func NewHUDBar(x, y, width, height float64) *HUDBar {
	return &HUDBar{
		X:           x,
		Y:           y,
		Width:       width,
		Height:      height,
		FillColor:   color.RGBA{0, 200, 0, 255},
		BackColor:   color.RGBA{50, 50, 50, 200},
		BorderColor: color.RGBA{255, 255, 255, 255},
		Visible:     true,
	}
}

// SetValue sets the bar value (clamped 0-1).
func (b *HUDBar) SetValue(current, maxVal float64) {
	if maxVal <= 0 {
		b.Current = 0

		return
	}

	b.Current = current / maxVal
	if b.Current < 0 {
		b.Current = 0
	}

	if b.Current > 1 {
		b.Current = 1
	}
}

// Draw renders the bar.
func (b *HUDBar) Draw(screen *ebiten.Image) {
	if !b.Visible {
		return
	}

	// Background
	drawRect(screen, b.X, b.Y, b.Width, b.Height, b.BackColor)

	// Fill
	fillWidth := b.Width * b.Current
	if fillWidth > 0 {
		drawRect(screen, b.X, b.Y, fillWidth, b.Height, b.FillColor)
	}

	// Border
	drawRectOutline(screen, b.X, b.Y, b.Width, b.Height, b.BorderColor)
}

// HUDText represents a text label.
type HUDText struct {
	X, Y    float64
	Text    string
	Color   color.RGBA
	Align   HUDAlign
	Scale   float64
	Visible bool
	Font    *text.GoTextFace
	Shadow  bool
}

// NewHUDText creates a text element.
func NewHUDText(x, y float64, txt string) *HUDText {
	return &HUDText{
		X:       x,
		Y:       y,
		Text:    txt,
		Color:   color.RGBA{255, 255, 255, 255},
		Scale:   1.0,
		Visible: true,
	}
}

// Draw renders the text.
func (t *HUDText) Draw(screen *ebiten.Image) {
	if !t.Visible || t.Text == "" {
		return
	}

	// Simple text rendering without fonts (placeholder)
	// In production, use text/v2 with proper fonts
}

// HUDIcon represents an icon with optional cooldown overlay.
type HUDIcon struct {
	X, Y          float64
	Size          float64
	Image         *ebiten.Image
	Cooldown      float64 // 0.0 - 1.0 (0 = ready)
	Charges       int
	MaxCharges    int
	Visible       bool
	BorderColor   color.RGBA
	CooldownColor color.RGBA
}

// NewHUDIcon creates an icon.
func NewHUDIcon(x, y, size float64, img *ebiten.Image) *HUDIcon {
	return &HUDIcon{
		X:             x,
		Y:             y,
		Size:          size,
		Image:         img,
		Visible:       true,
		BorderColor:   color.RGBA{200, 200, 200, 255},
		CooldownColor: color.RGBA{0, 0, 0, 180},
	}
}

// Draw renders the icon.
func (i *HUDIcon) Draw(screen *ebiten.Image) {
	if !i.Visible {
		return
	}

	if i.Image != nil {
		opts := &ebiten.DrawImageOptions{}
		w, h := i.Image.Bounds().Dx(), i.Image.Bounds().Dy()
		scaleX := i.Size / float64(w)
		scaleY := i.Size / float64(h)
		opts.GeoM.Scale(scaleX, scaleY)
		opts.GeoM.Translate(i.X, i.Y)
		screen.DrawImage(i.Image, opts)
	}

	// Cooldown overlay
	if i.Cooldown > 0 {
		// Draw pie-chart style cooldown
		drawCooldownOverlay(screen, i.X, i.Y, i.Size, i.Cooldown, i.CooldownColor)
	}

	// Border
	drawRectOutline(screen, i.X, i.Y, i.Size, i.Size, i.BorderColor)
}

// Minimap represents a minimap display.
type Minimap struct {
	X, Y          float64
	Width, Height float64
	MapWidth      float64 // World dimensions
	MapHeight     float64
	BorderColor   color.RGBA
	BackColor     color.RGBA
	PlayerColor   color.RGBA
	EnemyColor    color.RGBA
	ItemColor     color.RGBA
	Visible       bool

	// Markers
	Markers []MinimapMarker
}

// MinimapMarker represents a point on the minimap.
type MinimapMarker struct {
	WorldX, WorldY float64
	Color          color.RGBA
	Size           float64
}

// NewMinimap creates a minimap.
func NewMinimap(x, y, width, height, mapWidth, mapHeight float64) *Minimap {
	return &Minimap{
		X:           x,
		Y:           y,
		Width:       width,
		Height:      height,
		MapWidth:    mapWidth,
		MapHeight:   mapHeight,
		BorderColor: color.RGBA{255, 255, 255, 255},
		BackColor:   color.RGBA{0, 0, 0, 150},
		PlayerColor: color.RGBA{0, 255, 0, 255},
		EnemyColor:  color.RGBA{255, 0, 0, 255},
		ItemColor:   color.RGBA{255, 255, 0, 255},
		Visible:     true,
		Markers:     make([]MinimapMarker, 0),
	}
}

// AddMarker adds a marker to the minimap.
func (m *Minimap) AddMarker(worldX, worldY float64, c color.RGBA, size float64) {
	m.Markers = append(m.Markers, MinimapMarker{
		WorldX: worldX,
		WorldY: worldY,
		Color:  c,
		Size:   size,
	})
}

// ClearMarkers removes all markers.
func (m *Minimap) ClearMarkers() {
	m.Markers = m.Markers[:0]
}

// Draw renders the minimap.
func (m *Minimap) Draw(screen *ebiten.Image) {
	if !m.Visible {
		return
	}

	// Background
	drawRect(screen, m.X, m.Y, m.Width, m.Height, m.BackColor)

	// Markers
	for _, marker := range m.Markers {
		// Convert world to minimap coords
		mx := m.X + (marker.WorldX/m.MapWidth)*m.Width
		my := m.Y + (marker.WorldY/m.MapHeight)*m.Height
		drawRect(screen, mx-marker.Size/2, my-marker.Size/2, marker.Size, marker.Size, marker.Color)
	}

	// Border
	drawRectOutline(screen, m.X, m.Y, m.Width, m.Height, m.BorderColor)
}

// HUD manages all UI elements.
type HUD struct {
	Bars    map[string]*HUDBar
	Texts   map[string]*HUDText
	Icons   map[string]*HUDIcon
	Minimap *Minimap
	Visible bool
}

// NewHUD creates a new HUD.
func NewHUD() *HUD {
	return &HUD{
		Bars:    make(map[string]*HUDBar),
		Texts:   make(map[string]*HUDText),
		Icons:   make(map[string]*HUDIcon),
		Visible: true,
	}
}

// AddBar adds a bar to the HUD.
func (h *HUD) AddBar(name string, bar *HUDBar) {
	h.Bars[name] = bar
}

// AddText adds text to the HUD.
func (h *HUD) AddText(name string, txt *HUDText) {
	h.Texts[name] = txt
}

// AddIcon adds an icon to the HUD.
func (h *HUD) AddIcon(name string, icon *HUDIcon) {
	h.Icons[name] = icon
}

// SetMinimap sets the minimap.
func (h *HUD) SetMinimap(minimap *Minimap) {
	h.Minimap = minimap
}

// Update updates all HUD elements.
func (h *HUD) Update(dt float64) {
	// Animation updates could go here
}

// Draw renders the entire HUD.
func (h *HUD) Draw(screen *ebiten.Image) {
	if !h.Visible {
		return
	}

	for _, bar := range h.Bars {
		bar.Draw(screen)
	}

	for _, txt := range h.Texts {
		txt.Draw(screen)
	}

	for _, icon := range h.Icons {
		icon.Draw(screen)
	}

	if h.Minimap != nil {
		h.Minimap.Draw(screen)
	}
}

// Helper functions

func drawRect(screen *ebiten.Image, x, y, w, h float64, c color.RGBA) {
	img := ebiten.NewImage(int(w), int(h))
	img.Fill(c)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x, y)
	screen.DrawImage(img, opts)
}

func drawRectOutline(screen *ebiten.Image, x, y, w, h float64, c color.RGBA) {
	// Top
	drawRect(screen, x, y, w, 1, c)
	// Bottom
	drawRect(screen, x, y+h-1, w, 1, c)
	// Left
	drawRect(screen, x, y, 1, h, c)
	// Right
	drawRect(screen, x+w-1, y, 1, h, c)
}

func drawCooldownOverlay(screen *ebiten.Image, x, y, size, cooldown float64, c color.RGBA) {
	// Simple overlay using rectangles
	// A proper implementation would use shaders or pre-rendered images
	overlayHeight := size * cooldown
	drawRect(screen, x, y, size, overlayHeight, c)
}

// Damage number for floating combat text.
type DamageNumber struct {
	X, Y      float64
	Text      string
	Color     color.RGBA
	Life      float64
	MaxLife   float64
	VelocityY float64
	Scale     float64
	Active    bool
}

// DamageNumberSystem manages floating damage numbers.
type DamageNumberSystem struct {
	Numbers  []DamageNumber
	PoolSize int
}

// NewDamageNumberSystem creates a damage number system.
func NewDamageNumberSystem(poolSize int) *DamageNumberSystem {
	return &DamageNumberSystem{
		Numbers:  make([]DamageNumber, poolSize),
		PoolSize: poolSize,
	}
}

// Spawn creates a floating damage number.
func (s *DamageNumberSystem) Spawn(x, y float64, txt string, c color.RGBA) {
	for i := range s.Numbers {
		if !s.Numbers[i].Active {
			s.Numbers[i] = DamageNumber{
				X:         x,
				Y:         y,
				Text:      txt,
				Color:     c,
				Life:      1.0,
				MaxLife:   1.0,
				VelocityY: -50,
				Scale:     1.0,
				Active:    true,
			}

			return
		}
	}
}

// Update updates all damage numbers.
func (s *DamageNumberSystem) Update(dt float64) {
	for i := range s.Numbers {
		n := &s.Numbers[i]
		if !n.Active {
			continue
		}

		n.Life -= dt
		if n.Life <= 0 {
			n.Active = false

			continue
		}

		n.Y += n.VelocityY * dt
		n.VelocityY *= 0.95 // Slow down

		// Fade out
		progress := 1.0 - (n.Life / n.MaxLife)
		n.Color.A = uint8(255 * (1.0 - progress))
		n.Scale = 1.0 + progress*0.5
	}
}

// Draw renders all damage numbers.
func (s *DamageNumberSystem) Draw(screen *ebiten.Image) {
	for i := range s.Numbers {
		n := &s.Numbers[i]
		if !n.Active {
			continue
		}

		// Draw text (placeholder - would use proper text rendering)
		size := 8.0 * n.Scale
		drawRect(screen, n.X-size/2, n.Y-size/2, size, size, n.Color)
	}
}

// WorldToScreen converts world position to screen position.
func WorldToScreen(worldX, worldY, cameraX, cameraY, zoom float64, screenW, screenH int) (float64, float64) {
	screenX := (worldX-cameraX)*zoom + float64(screenW)/2
	screenY := (worldY-cameraY)*zoom + float64(screenH)/2

	return screenX, screenY
}

// FormatNumber formats large numbers with suffixes.
func FormatNumber(n int64) string {
	if n < 1000 {
		return formatInt(n)
	}

	if n < 1000000 {
		return formatFloat(float64(n)/1000) + "K"
	}

	if n < 1000000000 {
		return formatFloat(float64(n)/1000000) + "M"
	}

	return formatFloat(float64(n)/1000000000) + "B"
}

func formatInt(n int64) string {
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

func formatFloat(f float64) string {
	whole := int64(f)

	frac := int64((f - float64(whole)) * 10)
	if frac == 0 {
		return formatInt(whole)
	}

	return formatInt(whole) + "." + string(rune('0'+frac))
}

// Clamp clamps a value between min and max.
func Clamp(value, minVal, maxVal float64) float64 {
	return math.Max(minVal, math.Min(maxVal, value))
}
