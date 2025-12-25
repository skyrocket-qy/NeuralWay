# System Contract Protocol

Define strict Input/Output for ECS systems to enable parallel AI development.

---

## Contract Definition

Each system declares:
- **Reads**: Components read (dependencies)
- **Writes**: Components modified
- **Queries**: Entity filters used

---

## Example: MovementSystem

```go
// Contract:
// Reads:  Velocity
// Writes: Position
// Query:  Filter2[Position, Velocity]

type MovementSystem struct {
    filter *ecs.Filter2[Position, Velocity]
}

func (s *MovementSystem) Update(world *ecs.World) {
    query := s.filter.Query()
    for query.Next() {
        pos, vel := query.Get()
        pos.X += vel.X
        pos.Y += vel.Y
    }
}
```

---

## Contract Table

| System | Reads | Writes |
|--------|-------|--------|
| MovementSystem | Velocity | Position |
| RenderSystem | Position, Sprite | - |
| CollisionSystem | Position, Collider | - |
| HealthSystem | Health | Health |
| AnimationSystem | Animation | Sprite |

---

## Benefits

1. **Parallel Development**: AI can generate systems without conflicts
2. **Clear Dependencies**: Know what components are needed
3. **Testing**: Mock only required components
