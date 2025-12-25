package ai

import (
	"fmt"
	"strings"
)

// AssetStyle defines the visual style for generated assets.
type AssetStyle string

const (
	StylePixelArt  AssetStyle = "pixel-art"
	StyleCartoon   AssetStyle = "cartoon"
	StyleRealistic AssetStyle = "realistic"
	StyleAnime     AssetStyle = "anime"
	StyleLowPoly   AssetStyle = "low-poly"
)

// SpritePrompt describes a sprite to generate.
type SpritePrompt struct {
	Subject     string     // What to draw (e.g., "warrior", "tree")
	Style       AssetStyle // Visual style
	Animation   string     // Animation state (idle, walk, attack)
	Direction   string     // Facing direction (front, side, back)
	Size        string     // Sprite size (16x16, 32x32, 64x64)
	Variations  int        // Number of variations to generate
	Transparent bool       // Use transparent background
}

// SFXPrompt describes a sound effect to generate.
type SFXPrompt struct {
	Description string  // What the sound is (e.g., "coin pickup")
	Duration    float64 // Length in seconds
	Intensity   string  // quiet, normal, loud
	Category    string  // ui, ambient, action, music
}

// AssetGenerator generates prompts for AI image/audio generation.
type AssetGenerator struct {
	defaultStyle AssetStyle
}

// NewAssetGenerator creates a new asset generator.
func NewAssetGenerator(style AssetStyle) *AssetGenerator {
	return &AssetGenerator{
		defaultStyle: style,
	}
}

// GenerateSpritePrompt creates a prompt for DALL-E/Stable Diffusion.
func (g *AssetGenerator) GenerateSpritePrompt(spec SpritePrompt) string {
	style := spec.Style
	if style == "" {
		style = g.defaultStyle
	}

	var parts []string

	// Core subject
	parts = append(parts, spec.Subject)

	// Animation/pose
	if spec.Animation != "" {
		parts = append(parts, spec.Animation+" pose")
	}

	// Direction
	if spec.Direction != "" {
		parts = append(parts, spec.Direction+" view")
	}

	// Style
	switch style {
	case StylePixelArt:
		parts = append(parts, "pixel art style", "16-bit", "retro game sprite")
	case StyleCartoon:
		parts = append(parts, "cartoon style", "bold outlines", "vibrant colors")
	case StyleRealistic:
		parts = append(parts, "realistic style", "detailed", "high quality")
	case StyleAnime:
		parts = append(parts, "anime style", "Japanese animation", "cel shaded")
	case StyleLowPoly:
		parts = append(parts, "low poly style", "3D rendered", "minimal polygons")
	}

	// Size hint
	if spec.Size != "" {
		parts = append(parts, spec.Size+" sprite")
	}

	// Background
	if spec.Transparent {
		parts = append(parts, "transparent background", "PNG format")
	} else {
		parts = append(parts, "solid color background")
	}

	// Game art keywords
	parts = append(parts, "game asset", "sprite sheet ready")

	return strings.Join(parts, ", ")
}

// GenerateSFXPrompt creates a prompt for ElevenLabs/audio AI.
func (g *AssetGenerator) GenerateSFXPrompt(spec SFXPrompt) string {
	var parts []string

	// Core description
	parts = append(parts, spec.Description)

	// Duration
	if spec.Duration > 0 {
		parts = append(parts, fmt.Sprintf("%.1f seconds", spec.Duration))
	}

	// Intensity
	switch spec.Intensity {
	case "quiet":
		parts = append(parts, "soft", "subtle")
	case "loud":
		parts = append(parts, "loud", "impactful")
	default:
		parts = append(parts, "medium volume")
	}

	// Category keywords
	switch spec.Category {
	case "ui":
		parts = append(parts, "UI sound effect", "clean", "digital")
	case "ambient":
		parts = append(parts, "ambient sound", "background", "atmospheric")
	case "action":
		parts = append(parts, "action sound", "dynamic", "energetic")
	case "music":
		parts = append(parts, "musical", "melodic")
	}

	// Audio format
	parts = append(parts, "game audio", "WAV format")

	return strings.Join(parts, ", ")
}

// GenerateSpriteSheet generates prompts for a full animation sprite sheet.
func (g *AssetGenerator) GenerateSpriteSheet(subject string, style AssetStyle) []string {
	animations := []string{"idle", "walk", "run", "jump", "attack", "hurt", "death"}
	directions := []string{"front", "side"}

	prompts := make([]string, 0)

	for _, anim := range animations {
		for _, dir := range directions {
			prompt := g.GenerateSpritePrompt(SpritePrompt{
				Subject:     subject,
				Style:       style,
				Animation:   anim,
				Direction:   dir,
				Size:        "32x32",
				Transparent: true,
			})
			prompts = append(prompts, prompt)
		}
	}

	return prompts
}

// GenerateGameAssets generates prompts for common game assets.
func (g *AssetGenerator) GenerateGameAssets(theme string) map[string]string {
	assets := map[string]string{
		"player": g.GenerateSpritePrompt(
			SpritePrompt{
				Subject:     theme + " hero character",
				Style:       g.defaultStyle,
				Animation:   "idle",
				Transparent: true,
			},
		),
		"enemy": g.GenerateSpritePrompt(
			SpritePrompt{
				Subject:     theme + " enemy creature",
				Style:       g.defaultStyle,
				Animation:   "idle",
				Transparent: true,
			},
		),
		"coin": g.GenerateSpritePrompt(
			SpritePrompt{Subject: "gold coin collectible", Style: g.defaultStyle, Transparent: true},
		),
		"background": g.GenerateSpritePrompt(
			SpritePrompt{Subject: theme + " game background", Style: g.defaultStyle, Transparent: false},
		),
		"tile": g.GenerateSpritePrompt(
			SpritePrompt{
				Subject:     theme + " platform tile",
				Style:       g.defaultStyle,
				Size:        "32x32",
				Transparent: true,
			},
		),
	}

	return assets
}

// ExportAssetList generates a markdown list of asset prompts.
func (g *AssetGenerator) ExportAssetList(assets map[string]string) string {
	var sb strings.Builder

	sb.WriteString("# AI Asset Prompts\n\n")

	for name, prompt := range assets {
		sb.WriteString(fmt.Sprintf("## %s\n\n", name))
		sb.WriteString("```\n" + prompt + "\n```\n\n")
	}

	return sb.String()
}
