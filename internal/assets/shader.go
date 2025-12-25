package assets

import (
	"fmt"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
)

// Shader wraps an Ebitengine Kage shader.
type Shader struct {
	*ebiten.Shader

	name string
}

// ShaderManager loads and caches Kage shaders.
type ShaderManager struct {
	shaders map[string]*Shader
	fs      fs.FS
}

// NewShaderManager creates a shader manager.
func NewShaderManager(filesystem fs.FS) *ShaderManager {
	return &ShaderManager{
		shaders: make(map[string]*Shader),
		fs:      filesystem,
	}
}

// Load loads a Kage shader from a .kage file.
func (m *ShaderManager) Load(name, path string) (*Shader, error) {
	if s, ok := m.shaders[name]; ok {
		return s, nil
	}

	data, err := fs.ReadFile(m.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read shader %s: %w", path, err)
	}

	shader, err := ebiten.NewShader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to compile shader %s: %w", name, err)
	}

	s := &Shader{Shader: shader, name: name}
	m.shaders[name] = s

	return s, nil
}

// LoadFromString compiles a shader from source code.
func (m *ShaderManager) LoadFromString(name, source string) (*Shader, error) {
	if s, ok := m.shaders[name]; ok {
		return s, nil
	}

	shader, err := ebiten.NewShader([]byte(source))
	if err != nil {
		return nil, fmt.Errorf("failed to compile shader %s: %w", name, err)
	}

	s := &Shader{Shader: shader, name: name}
	m.shaders[name] = s

	return s, nil
}

// Get returns a cached shader by name.
func (m *ShaderManager) Get(name string) *Shader {
	return m.shaders[name]
}

// Common built-in shaders

// GrayscaleShader returns a shader that converts to grayscale.
const GrayscaleShader = "\n//kage:unit pixels\npackage main\n\nfunc Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {\n\tc := imageSrc0At(srcPos)\n\tgray := dot(c.rgb, vec3(0.299, 0.587, 0.114))\n\treturn vec4(gray, gray, c.a) * color"

// FlashShader returns a shader for damage flash effect.
const FlashShader = `
//kage:unit pixels
package main

var FlashColor vec4
var FlashAmount float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	c := imageSrc0At(srcPos)
	return mix(c, FlashColor, FlashAmount) * color
}
`

// WobbleShader returns a shader for wobble/wave effect.
const WobbleShader = `
//kage:unit pixels
package main

var Time float
var Amplitude float
var Frequency float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	offset := sin(srcPos.y * Frequency + Time) * Amplitude
	return imageSrc0At(vec2(srcPos.x + offset, srcPos.y)) * color
}
`

// OutlineShader returns a shader for adding outlines.
const OutlineShader = `
//kage:unit pixels
package main

var OutlineColor vec4
var OutlineWidth float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	c := imageSrc0At(srcPos)
	if c.a > 0.5 {
		return c * color
	}
	
	// Check neighbors
	for dx := -OutlineWidth; dx <= OutlineWidth; dx++ {
		for dy := -OutlineWidth; dy <= OutlineWidth; dy++ {
			neighbor := imageSrc0At(srcPos + vec2(dx, dy))
			if neighbor.a > 0.5 {
				return OutlineColor * color
			}
		}
	}
	
	return vec4(0)
}
`

// LoadBuiltinShaders loads all built-in shaders into the manager.
func (m *ShaderManager) LoadBuiltinShaders() error {
	builtins := map[string]string{
		"grayscale": GrayscaleShader,
		"flash":     FlashShader,
		"wobble":    WobbleShader,
		"outline":   OutlineShader,
	}

	for name, source := range builtins {
		if _, err := m.LoadFromString(name, source); err != nil {
			return fmt.Errorf("failed to load builtin shader %s: %w", name, err)
		}
	}

	return nil
}
