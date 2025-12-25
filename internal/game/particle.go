package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// Particle represents a single particle.
type Particle struct {
	X, Y       float64
	VX, VY     float64
	Life       float64 // Remaining lifetime
	MaxLife    float64
	Size       float64
	StartSize  float64
	EndSize    float64
	Color      color.RGBA
	StartColor color.RGBA
	EndColor   color.RGBA
	Rotation   float64
	RotSpeed   float64
	GravityX   float64
	GravityY   float64
	Drag       float64 // Velocity damping (0-1)
	Active     bool
}

// ParticleConfig defines how particles are created.
type ParticleConfig struct {
	// Lifetime
	LifeMin, LifeMax float64

	// Size
	SizeStart    float64
	SizeEnd      float64
	SizeVariance float64

	// Velocity
	SpeedMin, SpeedMax float64
	AngleMin, AngleMax float64 // In radians
	SpreadAngle        float64 // Spread from direction

	// Color
	StartColor color.RGBA
	EndColor   color.RGBA

	// Physics
	GravityX, GravityY float64
	Drag               float64

	// Rotation
	RotationMin, RotationMax float64
	RotSpeedMin, RotSpeedMax float64

	// Emission
	EmissionRate float64 // Particles per second
	BurstCount   int     // Particles per burst

	// Blend mode (for rendering)
	Additive bool
}

// DefaultParticleConfig returns a default particle configuration.
func DefaultParticleConfig() ParticleConfig {
	return ParticleConfig{
		LifeMin:      0.5,
		LifeMax:      1.5,
		SizeStart:    8,
		SizeEnd:      2,
		SpeedMin:     50,
		SpeedMax:     150,
		AngleMin:     0,
		AngleMax:     2 * math.Pi,
		StartColor:   color.RGBA{255, 255, 255, 255},
		EndColor:     color.RGBA{255, 255, 255, 0},
		Drag:         0.98,
		EmissionRate: 50,
		BurstCount:   20,
	}
}

// ParticleEmitter emits and manages particles.
type ParticleEmitter struct {
	X, Y         float64
	Config       ParticleConfig
	Particles    []Particle
	EmitTimer    float64
	Active       bool
	Continuous   bool          // Keep emitting
	FollowTarget bool          // Particles follow emitter
	Image        *ebiten.Image // Optional particle image
}

// NewParticleEmitter creates an emitter with pooled particles.
func NewParticleEmitter(config ParticleConfig, poolSize int) *ParticleEmitter {
	return &ParticleEmitter{
		Config:    config,
		Particles: make([]Particle, poolSize),
		Active:    true,
	}
}

// SetPosition sets the emitter position.
func (e *ParticleEmitter) SetPosition(x, y float64) {
	e.X = x
	e.Y = y
}

// Emit spawns a single particle.
func (e *ParticleEmitter) Emit() {
	// Find inactive particle
	for i := range e.Particles {
		if !e.Particles[i].Active {
			e.initParticle(&e.Particles[i])

			return
		}
	}
}

// Burst emits multiple particles at once.
func (e *ParticleEmitter) Burst(count int) {
	for range count {
		e.Emit()
	}
}

// BurstAt emits particles at a specific position.
func (e *ParticleEmitter) BurstAt(x, y float64, count int) {
	oldX, oldY := e.X, e.Y
	e.X, e.Y = x, y
	e.Burst(count)
	e.X, e.Y = oldX, oldY
}

// initParticle initializes a particle from config.
func (e *ParticleEmitter) initParticle(p *Particle) {
	cfg := &e.Config

	// Position
	p.X = e.X
	p.Y = e.Y

	// Lifetime
	p.MaxLife = cfg.LifeMin + rand.Float64()*(cfg.LifeMax-cfg.LifeMin)
	p.Life = p.MaxLife

	// Size
	variance := cfg.SizeVariance * (rand.Float64()*2 - 1)
	p.StartSize = cfg.SizeStart + variance
	p.EndSize = cfg.SizeEnd + variance*0.5
	p.Size = p.StartSize

	// Velocity
	speed := cfg.SpeedMin + rand.Float64()*(cfg.SpeedMax-cfg.SpeedMin)
	angle := cfg.AngleMin + rand.Float64()*(cfg.AngleMax-cfg.AngleMin)
	p.VX = speed * math.Cos(angle)
	p.VY = speed * math.Sin(angle)

	// Color
	p.StartColor = cfg.StartColor
	p.EndColor = cfg.EndColor
	p.Color = p.StartColor

	// Physics
	p.GravityX = cfg.GravityX
	p.GravityY = cfg.GravityY
	p.Drag = cfg.Drag

	// Rotation
	p.Rotation = cfg.RotationMin + rand.Float64()*(cfg.RotationMax-cfg.RotationMin)
	p.RotSpeed = cfg.RotSpeedMin + rand.Float64()*(cfg.RotSpeedMax-cfg.RotSpeedMin)

	p.Active = true
}

// Update updates all particles.
func (e *ParticleEmitter) Update(dt float64) {
	// Continuous emission
	if e.Active && e.Continuous && e.Config.EmissionRate > 0 {
		e.EmitTimer += dt

		interval := 1.0 / e.Config.EmissionRate
		for e.EmitTimer >= interval {
			e.Emit()
			e.EmitTimer -= interval
		}
	}

	// Update particles
	for i := range e.Particles {
		p := &e.Particles[i]
		if !p.Active {
			continue
		}

		// Update life
		p.Life -= dt
		if p.Life <= 0 {
			p.Active = false

			continue
		}

		// Progress through lifetime (0 = start, 1 = end)
		t := 1.0 - (p.Life / p.MaxLife)

		// Update physics
		p.VX += p.GravityX * dt
		p.VY += p.GravityY * dt
		p.VX *= p.Drag
		p.VY *= p.Drag
		p.X += p.VX * dt
		p.Y += p.VY * dt

		// Update rotation
		p.Rotation += p.RotSpeed * dt

		// Interpolate size
		p.Size = lerp(p.StartSize, p.EndSize, t)

		// Interpolate color
		p.Color = lerpColor(p.StartColor, p.EndColor, t)
	}
}

// Draw renders all particles.
func (e *ParticleEmitter) Draw(screen *ebiten.Image) {
	for i := range e.Particles {
		p := &e.Particles[i]
		if !p.Active {
			continue
		}

		if e.Image != nil {
			// Draw image particle
			opts := &ebiten.DrawImageOptions{}

			// Center and scale
			w, h := e.Image.Bounds().Dx(), e.Image.Bounds().Dy()
			scale := p.Size / float64(max(w, h))

			opts.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			opts.GeoM.Rotate(p.Rotation)
			opts.GeoM.Scale(scale, scale)
			opts.GeoM.Translate(p.X, p.Y)

			// Apply color
			opts.ColorScale.Scale(
				float32(p.Color.R)/255,
				float32(p.Color.G)/255,
				float32(p.Color.B)/255,
				float32(p.Color.A)/255,
			)

			if e.Config.Additive {
				opts.Blend = ebiten.BlendLighter
			}

			screen.DrawImage(e.Image, opts)
		} else {
			// Draw simple circle
			drawCircle(screen, p.X, p.Y, p.Size/2, p.Color)
		}
	}
}

// GetActiveCount returns number of active particles.
func (e *ParticleEmitter) GetActiveCount() int {
	count := 0

	for i := range e.Particles {
		if e.Particles[i].Active {
			count++
		}
	}

	return count
}

// Clear deactivates all particles.
func (e *ParticleEmitter) Clear() {
	for i := range e.Particles {
		e.Particles[i].Active = false
	}
}

// ParticleSystem manages multiple emitters.
type ParticleSystem struct {
	Emitters []*ParticleEmitter
}

// NewParticleSystem creates a particle system.
func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		Emitters: make([]*ParticleEmitter, 0),
	}
}

// AddEmitter adds an emitter to the system.
func (ps *ParticleSystem) AddEmitter(emitter *ParticleEmitter) {
	ps.Emitters = append(ps.Emitters, emitter)
}

// Update updates all emitters.
func (ps *ParticleSystem) Update(dt float64) {
	for _, e := range ps.Emitters {
		e.Update(dt)
	}
}

// Draw renders all emitters.
func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	for _, e := range ps.Emitters {
		e.Draw(screen)
	}
}

// Helper functions

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(lerp(float64(a.R), float64(b.R), t)),
		G: uint8(lerp(float64(a.G), float64(b.G), t)),
		B: uint8(lerp(float64(a.B), float64(b.B), t)),
		A: uint8(lerp(float64(a.A), float64(b.A), t)),
	}
}

func drawCircle(screen *ebiten.Image, x, y, radius float64, c color.RGBA) {
	// Simple filled circle using ebitenutil would be better
	// This is a placeholder - in production use a circle image or shader
	size := max(int(radius*2), 1)

	img := ebiten.NewImage(size, size)
	img.Fill(c)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x-radius, y-radius)
	screen.DrawImage(img, opts)
}
