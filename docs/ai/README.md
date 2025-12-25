# AI-Native Game Engine

RAG-ready documentation for LLM agents integrating with this game framework.

---

## 📚 Documentation Index

### For AI Coding Agents
| Document | Purpose |
|----------|---------|
| [AGENT_GUIDELINES.md](AGENT_GUIDELINES.md) | **START HERE** - Code conventions, patterns, rules |
| [components.md](components.md) | ECS component reference |
| [actions.md](actions.md) | AI action reference |

### For AI QA Agents
| Document | Purpose |
|----------|---------|
| [QA_GUIDE.md](QA_GUIDE.md) | Automated QA testing guide |
| [GAME_TEMPLATES.md](GAME_TEMPLATES.md) | Templates for new games |

### Standards & Protocols
| Document | Purpose |
|----------|---------|
| [../protocols/state.md](../protocols/state.md) | State export format |
| [../protocols/actions.md](../protocols/actions.md) | Action protocol |
| [../standards/naming.md](../standards/naming.md) | Asset naming conventions |

---

## 🚀 Quick Start

```go
import (
    "td/internal/ai"
    "td/internal/engine"
    "td/internal/components"
)

// 1. Create headless game (no GPU)
game := engine.NewHeadlessGame()

// 2. Add AI metadata to entities
mapper := ecs.NewMap2[components.Position, components.AIMetadata](&game.World)
mapper.NewEntity(
    &components.Position{X: 100, Y: 200},
    components.NewPlayerMetadata("Hero character"),
)

// 3. Export state for LLM
exporter := ai.NewStateExporter(&game.World)
json := exporter.ExportJSON(&game.World, game.CurrentTick())

// 4. Execute AI actions
executor := ai.NewActionExecutor(&game.World)
executor.Execute(ai.MoveAction{Entity: entity, DX: 10, DY: 0})
```

---

## 📦 Core Packages

| Package | Purpose |
|---------|---------|
| `internal/ai` | GameAdapter, Observer, AnomalyDetector, Players, QASession |
| `internal/engine` | HeadlessGame, Game, SceneManager |
| `internal/components` | ECS components including AIMetadata |
| `internal/systems` | Render, physics, animation systems |

---

## 🎮 Example Games with QA Adapters

| Game | Adapter | QA Tests |
|------|---------|----------|
| Survivor | `examples/survivor/adapter.go` | `qa_test.go` |

---

## 🔧 Key AI Interfaces

### GameAdapter
```go
type GameAdapter interface {
    Name() string
    GetState() GameState
    IsGameOver() bool
    GetScore() int
    AvailableActions() []ActionType
    PerformAction(ActionType) error
    Step() error
    Reset() error
}
```

### Player
```go
type Player interface {
    DecideAction(state GameState, available []ActionType) ActionType
}
```

---

## 🧪 Running QA Tests

```bash
# Run all tests
go test ./...

# Run QA session on Survivor
go test ./examples/survivor/... -run TestQASession -v

# Quick adapter tests
go test ./examples/survivor/... -run TestSurvivorAdapter
```
