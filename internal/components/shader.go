package components

import "github.com/hajimehoshi/ebiten/v2"

// ShaderEffect applies a shader effect to an entity's sprite rendering.
// Entities with both Sprite and ShaderEffect will be rendered using the shader.
type ShaderEffect struct {
	// Shader is the compiled Kage shader to apply.
	Shader *ebiten.Shader

	// Uniforms contains shader parameter values.
	// Keys must match uniform variable names in the shader.
	// Values can be float32, []float32, or [N]float32.
	Uniforms map[string]any

	// Enabled controls whether the shader effect is active.
	Enabled bool
}

// NewShaderEffect creates an enabled shader effect with the given shader.
func NewShaderEffect(shader *ebiten.Shader) ShaderEffect {
	return ShaderEffect{
		Shader:   shader,
		Uniforms: make(map[string]any),
		Enabled:  true,
	}
}

// SetUniform sets a shader uniform value.
func (s *ShaderEffect) SetUniform(name string, value any) {
	if s.Uniforms == nil {
		s.Uniforms = make(map[string]any)
	}

	s.Uniforms[name] = value
}

// SetFloat sets a float uniform.
func (s *ShaderEffect) SetFloat(name string, value float32) {
	s.SetUniform(name, value)
}

// SetVec4 sets a vec4 uniform (e.g., for colors).
func (s *ShaderEffect) SetVec4(name string, r, g, b, a float32) {
	s.SetUniform(name, []float32{r, g, b, a})
}
