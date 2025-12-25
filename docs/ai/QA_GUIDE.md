# QA Agent Guide

Instructions for AI agents performing automated QA testing on games.

---

## Quick Start

```go
// 1. Create game and adapter
game := NewGame()
adapter := NewSurvivorAdapter(game)

// 2. Create QA session
session := ai.NewQASession(adapter)
session.SetPlayer(ai.NewRandomPlayer(time.Now().UnixNano()))
session.SetConfig(ai.SessionConfig{
    Runs:     10,
    MaxTicks: 3600, // 1 minute
})

// 3. Run and analyze
report := session.Run()
fmt.Println(report.GenerateMarkdown())
```

---

## Game Adapter Requirements

When implementing `GameAdapter` for a new game:

### Required Methods

| Method | Returns | Purpose |
|--------|---------|---------|
| `Name()` | string | Game identifier |
| `GetState()` | GameState | Current observable state |
| `IsGameOver()` | bool | Has game ended? |
| `GetScore()` | int | Current score/kills |
| `AvailableActions()` | []ActionType | Valid actions now |
| `PerformAction(a)` | error | Apply action |
| `Step()` | error | Advance one tick |
| `Reset()` | error | Restart game |

### GameState Fields

```go
type GameState struct {
    Tick         int64          // Current tick
    Score        int            // Points/kills
    PlayerPos    [2]float64     // X, Y position
    PlayerHealth [2]int         // [current, max]
    EntityCount  int            // Total entities
    CustomData   map[string]any // Game-specific data
}
```

### CustomData Recommendations

Add game-specific metrics:
```go
state.CustomData["level"] = game.player.Level
state.CustomData["enemies"] = len(game.enemies)
state.CustomData["weapons"] = len(game.player.Weapons)
state.CustomData["game_time"] = game.gameTime
```

---

## Action Types

Standard actions an AI can perform:

| Action | Usage |
|--------|-------|
| `ActionNone` | Do nothing |
| `ActionMoveUp/Down/Left/Right` | Movement |
| `ActionJump` | Platformers |
| `ActionAttack` | Combat |
| `ActionUse` | Interact |

### Custom Actions

Games can define additional actions:
```go
const (
    ActionShoot ActionType = "shoot"
    ActionDodge ActionType = "dodge"
)
```

---

## Anomaly Detection

### Built-in Detectors

| Type | Configurable | Default |
|------|--------------|---------|
| Stuck | StuckThreshold | 120 ticks |
| Entity Leak | EntityLeakThreshold | 500 entities |
| Score Regression | - | Any decrease |
| Health Drain | HealthDrainRate | 0.5 HP/tick |
| Boundary Violation | BoundsWidth/Height | 1920×1080 |

### Tuning for Your Game

```go
detector := ai.NewAnomalyDetector()
detector.StuckThreshold = 300        // 5 seconds for slower games
detector.EntityLeakThreshold = 1000  // More entities expected
detector.BoundsWidth = 10000         // Larger world
detector.BoundsHeight = 10000
session.SetDetector(detector)
```

### Custom Detection Rules

```go
detector.AddRule(func(history []ai.Observation) []ai.Anomaly {
    // Check for specific game condition
    for _, obs := range history {
        if obs.State.CustomData["enemies"].(int) > 100 {
            return []ai.Anomaly{{
                Type: "too_many_enemies",
                Severity: ai.SeverityMedium,
                Tick: obs.Tick,
            }}
        }
    }
    return nil
})
```

---

## Player Strategies

### RandomPlayer
Best for: Fuzz testing, finding edge cases
```go
ai.NewRandomPlayer(seed)
```

### WeightedRandomPlayer
Best for: Biased exploration
```go
ai.NewWeightedRandomPlayer(seed, map[ai.ActionType]float64{
    ai.ActionMoveRight: 2.0, // Prefer moving right
})
```

### StrategyPlayer
Best for: Targeted testing scenarios
```go
ai.NewStrategyPlayer(behaviorTree)
```

### ReplayPlayer
Best for: Regression testing
```go
ai.NewReplayPlayer(recordedActions)
```

---

## Report Interpretation

### Conclusion Values

| Value | Meaning |
|-------|---------|
| `PASS` | No anomalies detected |
| `WARNING` | 1-3 anomalies, likely minor |
| `FAIL` | 4+ anomalies, needs investigation |

### Analyzing Anomalies

1. **Stuck** → Check collision boundaries
2. **Entity Leak** → Check entity cleanup on death
3. **Score Regression** → Intentional or bug?
4. **Health Drain** → Balance issue or bug?
5. **Boundary Violation** → Add bounds clamping
