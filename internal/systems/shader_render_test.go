package systems

import (
	"testing"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

func TestShaderRenderSystemCreation(t *testing.T) {
	world := ecs.NewWorld()
	system := NewShaderRenderSystem(&world)

	if system.filter == nil {
		t.Error("Filter should not be nil")
	}

	if system.opts == nil {
		t.Error("Options should not be nil")
	}
}

func TestShaderEffectComponent(t *testing.T) {
	effect := components.NewShaderEffect(nil)

	if !effect.Enabled {
		t.Error("Effect should be enabled by default")
	}

	if effect.Uniforms == nil {
		t.Error("Uniforms map should be initialized")
	}
}

func TestShaderEffectSetUniforms(t *testing.T) {
	effect := components.NewShaderEffect(nil)

	effect.SetFloat("Time", 1.5)

	if effect.Uniforms["Time"] != float32(1.5) {
		t.Error("Float uniform not set correctly")
	}

	effect.SetVec4("Color", 1.0, 0.5, 0.0, 1.0)

	vec4 := effect.Uniforms["Color"].([]float32)
	if vec4[0] != 1.0 || vec4[1] != 0.5 || vec4[2] != 0.0 || vec4[3] != 1.0 {
		t.Error("Vec4 uniform not set correctly")
	}
}

func TestShaderRenderSystemWithEntities(t *testing.T) {
	world := ecs.NewWorld()
	system := NewShaderRenderSystem(&world)

	mapper := ecs.NewMap3[components.Position, components.Sprite, components.ShaderEffect](&world)

	entity := mapper.NewEntity(
		&components.Position{X: 100, Y: 200},
		&components.Sprite{Visible: true, ScaleX: 1, ScaleY: 1},
		&components.ShaderEffect{Enabled: true, Uniforms: make(map[string]any)},
	)

	if !world.Alive(entity) {
		t.Error("Entity should be alive")
	}

	// Verify filter can find the entity
	query := system.filter.Query()

	count := 0
	for query.Next() {
		count++
	}

	if count != 1 {
		t.Errorf("Expected 1 shader entity, got %d", count)
	}
}
