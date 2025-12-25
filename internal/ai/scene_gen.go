package ai

import (
	"fmt"
	"strings"
)

// EntitySpec describes an entity to generate.
type EntitySpec struct {
	Name       string
	Type       string // "sprite", "player", "enemy", "collectible"
	Position   [2]float64
	Sprite     string
	Properties map[string]string
}

// SceneSpec describes a scene to generate.
type SceneSpec struct {
	Name        string
	Width       int
	Height      int
	Background  string
	Entities    []EntitySpec
	Description string
}

// SceneGenerator generates game code from specifications.
type SceneGenerator struct {
	templates map[string]string
}

// NewSceneGenerator creates a new scene generator.
func NewSceneGenerator() *SceneGenerator {
	return &SceneGenerator{
		templates: defaultTemplates(),
	}
}

// defaultTemplates returns built-in code templates.
func defaultTemplates() map[string]string {
	return map[string]string{
		"main": `package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
)

const (
	screenWidth  = %d
	screenHeight = %d
)

type Game struct {
	world ecs.World
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw entities
}

func (g *Game) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}
	game.world = ecs.NewWorld()
	
	// Initialize entities
%s
	
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("%s")
	ebiten.RunGame(game)
}
`,
		"entity": `	// Create %s
	// Position: (%.0f, %.0f)
	// Type: %s
`,
	}
}

// GenerateCode generates Go code from a SceneSpec.
func (g *SceneGenerator) GenerateCode(spec SceneSpec) string {
	// Generate entity initialization code
	var entityCode strings.Builder
	for _, e := range spec.Entities {
		entityCode.WriteString(fmt.Sprintf(g.templates["entity"],
			e.Name, e.Position[0], e.Position[1], e.Type))
	}

	// Generate main code
	code := fmt.Sprintf(g.templates["main"],
		spec.Width, spec.Height,
		entityCode.String(),
		spec.Name)

	return code
}

// ParsePrompt extracts a SceneSpec from natural language.
// This is a simplified parser - in production, use an LLM.
func (g *SceneGenerator) ParsePrompt(prompt string) SceneSpec {
	spec := SceneSpec{
		Name:        "Generated Game",
		Width:       800,
		Height:      600,
		Entities:    make([]EntitySpec, 0),
		Description: prompt,
	}

	promptLower := strings.ToLower(prompt)

	// Extract entities from keywords
	if strings.Contains(promptLower, "player") {
		spec.Entities = append(spec.Entities, EntitySpec{
			Name:     "Player",
			Type:     "player",
			Position: [2]float64{400, 300},
		})
	}

	if strings.Contains(promptLower, "enemy") || strings.Contains(promptLower, "enemies") {
		spec.Entities = append(spec.Entities, EntitySpec{
			Name:     "Enemy",
			Type:     "enemy",
			Position: [2]float64{600, 300},
		})
	}

	if strings.Contains(promptLower, "coin") || strings.Contains(promptLower, "collectible") {
		spec.Entities = append(spec.Entities, EntitySpec{
			Name:     "Coin",
			Type:     "collectible",
			Position: [2]float64{200, 200},
		})
	}

	if strings.Contains(promptLower, "platform") {
		spec.Entities = append(spec.Entities, EntitySpec{
			Name:     "Platform",
			Type:     "sprite",
			Position: [2]float64{400, 500},
		})
	}

	// Extract name
	if strings.Contains(promptLower, "platformer") {
		spec.Name = "Platformer Game"
	} else if strings.Contains(promptLower, "shooter") {
		spec.Name = "Shooter Game"
	} else if strings.Contains(promptLower, "puzzle") {
		spec.Name = "Puzzle Game"
	}

	return spec
}

// GenerateFromPrompt generates code directly from a prompt.
func (g *SceneGenerator) GenerateFromPrompt(prompt string) string {
	spec := g.ParsePrompt(prompt)

	return g.GenerateCode(spec)
}

// ExportMarkdown generates a markdown description of a scene.
func (g *SceneGenerator) ExportMarkdown(spec SceneSpec) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", spec.Name))
	sb.WriteString(fmt.Sprintf("**Size**: %dx%d\n\n", spec.Width, spec.Height))

	if spec.Description != "" {
		sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", spec.Description))
	}

	if len(spec.Entities) > 0 {
		sb.WriteString("## Entities\n\n")
		sb.WriteString("| Name | Type | Position |\n")
		sb.WriteString("|------|------|----------|\n")

		for _, e := range spec.Entities {
			sb.WriteString(fmt.Sprintf("| %s | %s | (%.0f, %.0f) |\n",
				e.Name, e.Type, e.Position[0], e.Position[1]))
		}
	}

	return sb.String()
}
