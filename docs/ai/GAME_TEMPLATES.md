# Game Example Templates

Reference patterns for creating new game examples.

---

## Minimal Game Structure

```
examples/mygame/
├── main.go         # Game implementation
├── adapter.go      # GameAdapter for QA (optional)
├── main_test.go    # Unit tests
├── qa_test.go      # QA session tests (optional)
└── assets/         # Images, sounds
    └── *.png
```

---

## Main Game Template

```go
package main

import (
    "log"
    "github.com/hajimehoshi/ebiten/v2"
)

const (
    screenWidth  = 640
    screenHeight = 480
)

type GameState int

const (
    StateTitle GameState = iota
    StatePlaying
    StateGameOver
)

type Game struct {
    state GameState
    score int
    // Add game-specific fields
}

func NewGame() *Game {
    return &Game{
        state: StateTitle,
    }
}

func (g *Game) Update() error {
    switch g.state {
    case StateTitle:
        // Handle title screen input
    case StatePlaying:
        // Game logic
    case StateGameOver:
        // Handle restart
    }
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    switch g.state {
    case StateTitle:
        // Draw title
    case StatePlaying:
        // Draw game
    case StateGameOver:
        // Draw game over
    }
}

func (g *Game) Layout(w, h int) (int, int) {
    return screenWidth, screenHeight
}

func main() {
    ebiten.SetWindowSize(screenWidth, screenHeight)
    ebiten.SetWindowTitle("My Game")
    if err := ebiten.RunGame(NewGame()); err != nil {
        log.Fatal(err)
    }
}
```

---

## GameAdapter Template

```go
package main

import "td/internal/ai"

type MyGameAdapter struct {
    game *Game
    tick int64
}

func NewMyGameAdapter(game *Game) *MyGameAdapter {
    return &MyGameAdapter{game: game}
}

func (a *MyGameAdapter) Name() string {
    return "My Game"
}

func (a *MyGameAdapter) GetState() ai.GameState {
    return ai.GameState{
        Tick:        a.tick,
        Score:       a.game.score,
        PlayerPos:   [2]float64{a.game.playerX, a.game.playerY},
        PlayerHealth: [2]int{a.game.hp, a.game.maxHp},
        EntityCount: len(a.game.entities),
    }
}

func (a *MyGameAdapter) IsGameOver() bool {
    return a.game.state == StateGameOver
}

func (a *MyGameAdapter) GetScore() int {
    return a.game.score
}

func (a *MyGameAdapter) AvailableActions() []ai.ActionType {
    if a.game.state != StatePlaying {
        return []ai.ActionType{ai.ActionNone}
    }
    return []ai.ActionType{
        ai.ActionMoveUp,
        ai.ActionMoveDown,
        ai.ActionMoveLeft,
        ai.ActionMoveRight,
    }
}

func (a *MyGameAdapter) PerformAction(action ai.ActionType) error {
    // Map action to game input
    return nil
}

func (a *MyGameAdapter) Step() error {
    a.tick++
    // Step game logic (without rendering)
    return nil
}

func (a *MyGameAdapter) Reset() error {
    a.tick = 0
    // Reset game state
    return nil
}
```

---

## QA Test Template

```go
package main

import (
    "testing"
    "time"
    "td/internal/ai"
)

func TestQASession(t *testing.T) {
    game := NewGame()
    adapter := NewMyGameAdapter(game)

    session := ai.NewQASession(adapter)
    session.SetPlayer(ai.NewRandomPlayer(time.Now().UnixNano()))
    session.SetConfig(ai.SessionConfig{
        Runs:     5,
        MaxTicks: 1800,
    })

    report := session.Run()
    t.Log("\n" + report.GenerateMarkdown())

    if report.TotalAnomalies > 3 {
        t.Errorf("Too many anomalies: %d", report.TotalAnomalies)
    }
}
```

---

## Visual Polish Checklist

Every game should include:

- [ ] Title screen with game name
- [ ] Instructions/controls display
- [ ] Score display during gameplay
- [ ] Game over screen with final score
- [ ] Particle effects for feedback
- [ ] Smooth animations
- [ ] Audio feedback (optional)

---

## Makefile Entry

Add to project Makefile:
```makefile
run-mygame:
	go run ./examples/mygame

build-mygame-wasm:
	GOOS=js GOARCH=wasm go build -o examples/mygame/web/mygame.wasm ./examples/mygame
```
