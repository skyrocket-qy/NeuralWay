package assets

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteSheet represents an animated sprite sheet with multiple frames.
type SpriteSheet struct {
	Image       *ebiten.Image
	FrameWidth  int
	FrameHeight int
	Columns     int
	Rows        int
	Frames      []*ebiten.Image
}

// NewSpriteSheet creates a sprite sheet from an image with uniform frame sizes.
func NewSpriteSheet(img *ebiten.Image, frameWidth, frameHeight int) *SpriteSheet {
	bounds := img.Bounds()
	cols := bounds.Dx() / frameWidth
	rows := bounds.Dy() / frameHeight

	ss := &SpriteSheet{
		Image:       img,
		FrameWidth:  frameWidth,
		FrameHeight: frameHeight,
		Columns:     cols,
		Rows:        rows,
		Frames:      make([]*ebiten.Image, 0, cols*rows),
	}

	// Extract individual frames
	for row := range rows {
		for col := range cols {
			x := col * frameWidth
			y := row * frameHeight
			frame := img.SubImage(
				ebiten.NewImage(frameWidth, frameHeight).Bounds().Add(
					struct{ X, Y int }{x, y},
				),
			).(*ebiten.Image)
			ss.Frames = append(ss.Frames, frame)
		}
	}

	return ss
}

// Frame returns a specific frame by index.
func (ss *SpriteSheet) Frame(index int) *ebiten.Image {
	if index < 0 || index >= len(ss.Frames) {
		return nil
	}

	return ss.Frames[index]
}

// FrameAt returns a frame by column and row.
func (ss *SpriteSheet) FrameAt(col, row int) *ebiten.Image {
	return ss.Frame(row*ss.Columns + col)
}

// Animation defines a sequence of frames for animation playback.
type Animation struct {
	Name     string
	Frames   []int   // Frame indices
	Duration float64 // Duration per frame in seconds
	Loop     bool
}

// AnimationSet holds multiple named animations.
type AnimationSet struct {
	Animations map[string]*Animation
	Sheet      *SpriteSheet
}

// NewAnimationSet creates a new animation set.
func NewAnimationSet(sheet *SpriteSheet) *AnimationSet {
	return &AnimationSet{
		Animations: make(map[string]*Animation),
		Sheet:      sheet,
	}
}

// Add adds an animation to the set.
func (as *AnimationSet) Add(name string, frames []int, duration float64, loop bool) {
	as.Animations[name] = &Animation{
		Name:     name,
		Frames:   frames,
		Duration: duration,
		Loop:     loop,
	}
}

// Get returns an animation by name.
func (as *AnimationSet) Get(name string) *Animation {
	return as.Animations[name]
}

// AsepriteJSON represents Aseprite JSON export format.
type AsepriteJSON struct {
	Frames map[string]AsepriteFrame `json:"frames"`
	Meta   AsepriteMeta             `json:"meta"`
}

type AsepriteFrame struct {
	Frame struct {
		X int `json:"x"`
		Y int `json:"y"`
		W int `json:"w"`
		H int `json:"h"`
	} `json:"frame"`
	Duration int `json:"duration"`
}

type AsepriteMeta struct {
	FrameTags []AsepriteTag `json:"frameTags"`
}

type AsepriteTag struct {
	Name      string `json:"name"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Direction string `json:"direction"`
}

// LoadAseprite loads an Aseprite JSON sprite sheet.
func LoadAseprite(filesystem fs.FS, jsonPath, imagePath string) (*AnimationSet, error) {
	// Load JSON
	data, err := fs.ReadFile(filesystem, jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read aseprite JSON: %w", err)
	}

	var aseData AsepriteJSON
	if err := json.Unmarshal(data, &aseData); err != nil {
		return nil, fmt.Errorf("failed to parse aseprite JSON: %w", err)
	}

	// Load image
	loader := NewLoader(filesystem)

	img, err := loader.LoadImage(imagePath)
	if err != nil {
		return nil, err
	}

	// Create sprite sheet (we'll handle frames differently for Aseprite)
	// For simplicity, assume first frame dimensions apply to all
	var frameWidth, frameHeight int
	for _, f := range aseData.Frames {
		frameWidth = f.Frame.W
		frameHeight = f.Frame.H

		break
	}

	sheet := NewSpriteSheet(img, frameWidth, frameHeight)
	animSet := NewAnimationSet(sheet)

	// Create animations from frame tags
	for _, tag := range aseData.Meta.FrameTags {
		frames := make([]int, 0, tag.To-tag.From+1)
		for i := tag.From; i <= tag.To; i++ {
			frames = append(frames, i)
		}

		// Get duration from first frame of animation
		frameName := strconv.Itoa(tag.From)
		duration := 0.1 // Default

		for name, f := range aseData.Frames {
			if strings.Contains(name, frameName) {
				duration = float64(f.Duration) / 1000.0 // Convert ms to seconds

				break
			}
		}

		animSet.Add(tag.Name, frames, duration, true)
	}

	return animSet, nil
}

// ParseFrameRange parses a frame range string like "0-5" or "0,1,2,3".
func ParseFrameRange(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range: %s", s)
		}

		start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}

		end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}

		frames := make([]int, 0, end-start+1)
		for i := start; i <= end; i++ {
			frames = append(frames, i)
		}

		return frames, nil
	}

	// Comma separated
	parts := strings.Split(s, ",")

	frames := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, err
		}

		frames = append(frames, n)
	}

	return frames, nil
}
