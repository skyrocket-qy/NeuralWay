# Components Reference

ECS components available in this framework.

---

## Core Components

### Position
```go
type Position struct { X, Y float64 }
```
2D world position.

### Velocity
```go
type Velocity struct { X, Y float64 }
```
Movement speed per tick.

### Sprite
```go
type Sprite struct {
    Image   *ebiten.Image
    OffsetX, OffsetY float64
    ScaleX, ScaleY float64
    Visible bool
}
```

### Health
```go
type Health struct { Current, Max int }
```

### Collider
```go
type Collider struct {
    Width, Height float64
    Layer, Mask uint32
}
```

---

## AI Components

### AIMetadata
```go
type AIMetadata struct {
    EntityType  string   // player, enemy, item, projectile, ui, effect
    Description string   // Human-readable for LLM
    VisualState string   // idle, attacking, damaged, dead
    Tags        []string // hostile, animated, collidable
    ActiveEffects []string
}
```

**Helpers**:
- `NewPlayerMetadata(desc)` → controllable, collidable
- `NewEnemyMetadata(desc)` → hostile, collidable, animated
- `NewItemMetadata(desc)` → collectible, collidable
- `NewProjectileMetadata(desc)` → collidable, temporary

---

## Rendering Components

### ParallaxLayer
```go
type ParallaxLayer struct {
    Image       *ebiten.Image
    SpeedFactor float64 // 0=static, 1=full scroll
    RepeatX, RepeatY bool
    Layer       int     // Draw order
}
```

### ShaderEffect
```go
type ShaderEffect struct {
    Shader   *ebiten.Shader
    Uniforms map[string]interface{}
    Enabled  bool
}
```
