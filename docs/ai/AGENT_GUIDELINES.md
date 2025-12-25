# AI Agent Guidelines

This document provides rules and patterns for AI agents working with this codebase.

---

## Core Principles

1. **Use ECS Architecture** - All game entities must use the Entity Component System pattern
2. **Prefer Composition** - Add components to entities instead of inheritance
3. **Keep Components Small** - Each component should represent ONE data concept
4. **Systems Process Components** - Logic lives in systems, not components

---

## Code Conventions

### File Naming
```
internal/
├── components/     # ECS components (data only)
│   └── *.go        # One file per component group
├── systems/        # ECS systems (logic)
│   └── *.go        # One file per system
├── ai/             # AI infrastructure
│   └── *.go        # Adapters, observers, detectors
└── engine/         # Game loop infrastructure
```

### Component Pattern
```go
// CORRECT: Simple data struct
type Position struct {
    X, Y float64
}

// CORRECT: Constructor for complex initialization
func NewSprite(img *ebiten.Image) *Sprite {
    return &Sprite{Image: img, ScaleX: 1, ScaleY: 1, Visible: true}
}

// WRONG: Don't add methods that modify state
// Components are DATA, not behavior
```

### System Pattern
```go
// System interface
type System interface {
    Update(world *ecs.World)
}

// Implementation
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

## DO's and DON'Ts

### ✅ DO
- Use `ecs.NewMap[T]` for component access
- Use `ecs.NewFilter` for querying entities
- Add `AIMetadata` to entities that AI should observe
- Use `HeadlessGame` for testing/AI play
- Keep game logic deterministic for reproducibility

### ❌ DON'T
- Don't store entity references in components (use entity IDs)
- Don't call rendering functions in headless mode
- Don't use global state for game logic
- Don't modify components during iteration without proper handling
- Don't create circular dependencies between packages

---

## Adding New Features

### New Component
1. Create in `internal/components/`
2. Add tests in `*_test.go`
3. Document in `docs/ai/components.md`

### New System
1. Create in `internal/systems/`
2. Implement `System` or `DrawSystem` interface
3. Add to game via `game.AddSystem()`

### New Game Example
1. Create folder in `examples/`
2. Implement `GameAdapter` for QA testing
3. Add to Makefile for build targets

---

## ECS Query Patterns

### Filter by Component Type
```go
// All entities with Position
filter := ecs.NewFilter1[Position](world)
query := filter.Query()
for query.Next() {
    pos := query.Get()
    // ...
}
```

### Filter by Multiple Components
```go
// Entities with BOTH Position AND Velocity
filter := ecs.NewFilter2[Position, Velocity](world)
```

### Random Access
```go
// Get specific component from entity
posMap := ecs.NewMap[Position](world)
if posMap.Has(entity) {
    pos := posMap.Get(entity)
}
```

---

## AI Integration Points

| Interface | Purpose | Usage |
|-----------|---------|-------|
| `GameAdapter` | Universal game testing | Implement for each game |
| `Player` | AI decision making | Random, Strategy, LLM |
| `Observer` | State recording | Anomaly detection |
| `AnomalyDetector` | Bug finding | Stuck, leaks, etc. |

---

## Testing Guidelines

### Unit Tests
- Test components in `internal/components/*_test.go`
- Test systems with mock worlds in `internal/systems/*_test.go`

### Integration Tests
- Use `engine.HeadlessGame` for game loop tests
- Use `ai.QASession` for automated playtesting

### Run Tests
```bash
go test ./...                      # All tests
go test ./internal/...             # Framework tests
go test ./examples/survivor/...    # Game-specific tests
```
