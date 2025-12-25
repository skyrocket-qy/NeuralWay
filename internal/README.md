# Framework Components

This directory contains the modular game framework. Each package follows the **"Use-if-Needed"** philosophy - import only what your game requires.

## Package Overview

| Package | Purpose | Dependencies |
|---------|---------|--------------|
| `pool` | Generic object pooling | None |
| `engine` | ECS game loop integration | ark, ebiten |
| `components` | Common ECS component types | ebiten |
| `systems` | Pre-built ECS systems | components |
| `archetypes` | Entity creation helpers | components, systems |
| `assets` | Asset loading (images, audio, tilemaps) | ebiten |
| `game` | Tower defense example code | All above |

## Usage

Import only the packages you need:

```go
import (
    "td/internal/engine"
    "td/internal/components"
    "td/internal/systems"
)

func main() {
    game := engine.NewGame(800, 600, "My Game")
    game.AddDrawSystem(systems.NewRenderSystem(&game.World))
    game.Run()
}
```

## Package Details

### `pool` - Object Pooling
Standalone generic pool for reducing allocations.

### `engine` - Game Loop
Wraps Ebitengine + Ark ECS into a simple `Game` struct with `System` and `DrawSystem` interfaces.

### `components` - ECS Components
Core components: `Position`, `Velocity`, `Sprite`, `Collider`, `Health`, `Tag`, `SortLayer`, `Tilemap`.

### `systems` - ECS Systems
Pre-built systems:
- `RenderSystem` - Basic sprite rendering
- `BatchRenderSystem` - Batched sprite rendering with culling
- `TilemapRenderSystem` - Tilemap rendering with viewport culling
- `MovementSystem` - Position += Velocity
- `CollisionSystem` - AABB collision detection
- `AnimationSystem` - Sprite animation
- `InputSystem` - Keyboard/mouse input helpers

### `archetypes` - Entity Templates
- **Generic**: `Archetype2`, `Archetype3`, `Archetype4` - build custom archetypes
- **Game-specific**: `SpriteArchetype`, `MovableArchetype`, `CollidableArchetype`, etc.

### `assets` - Asset Loading
- `Loader` - Image loading with caching
- `TiledMap` - Tiled JSON/TMX map loading
- `SpriteSheet` - Sprite sheet parsing
- `AudioManager` - Sound loading and playback

### `game` - Example Code
Tower defense specific code (not framework). Use as reference.
