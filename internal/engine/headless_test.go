package engine

import (
	"testing"

	"github.com/mlange-42/ark/ecs"
)

// testSystem is a simple system that increments a counter.
type testSystem struct {
	counter *int
}

func (s *testSystem) Update(world *ecs.World) {
	*s.counter++
}

func TestHeadlessGameStep(t *testing.T) {
	game := NewHeadlessGame()

	counter := 0
	game.AddSystem(&testSystem{counter: &counter})

	game.Step()

	if counter != 1 {
		t.Errorf("Counter should be 1 after one step, got %d", counter)
	}

	game.Step()

	if counter != 2 {
		t.Errorf("Counter should be 2 after two steps, got %d", counter)
	}
}

func TestHeadlessGameStepN(t *testing.T) {
	game := NewHeadlessGame()

	counter := 0
	game.AddSystem(&testSystem{counter: &counter})

	game.StepN(100)

	if counter != 100 {
		t.Errorf("Counter should be 100 after StepN(100), got %d", counter)
	}
}

func TestHeadlessGameCurrentTick(t *testing.T) {
	game := NewHeadlessGame()

	if game.CurrentTick() != 0 {
		t.Error("Tick should start at 0")
	}

	game.StepN(50)

	if game.CurrentTick() != 50 {
		t.Errorf("Tick should be 50, got %d", game.CurrentTick())
	}
}

func TestHeadlessGameRunUntil(t *testing.T) {
	game := NewHeadlessGame()

	counter := 0
	game.AddSystem(&testSystem{counter: &counter})

	ticks := game.RunUntil(func(w *ecs.World) bool {
		return counter >= 25
	}, 1000)

	if counter != 25 {
		t.Errorf("Counter should be 25, got %d", counter)
	}

	if ticks != 25 {
		t.Errorf("Should have run 25 ticks, got %d", ticks)
	}
}

func TestHeadlessGameRunUntilMaxTicks(t *testing.T) {
	game := NewHeadlessGame()

	// Condition never becomes true
	ticks := game.RunUntil(func(w *ecs.World) bool {
		return false
	}, 50)

	if ticks != 50 {
		t.Errorf("Should stop at max ticks (50), got %d", ticks)
	}
}

func TestHeadlessGameRunWithCallback(t *testing.T) {
	game := NewHeadlessGame()

	ticks := make([]int64, 0)

	game.RunWithCallback(5, func(w *ecs.World, tick int64) {
		ticks = append(ticks, tick)
	})

	if len(ticks) != 5 {
		t.Errorf("Callback should be called 5 times, got %d", len(ticks))
	}

	if ticks[4] != 5 {
		t.Errorf("Last tick should be 5, got %d", ticks[4])
	}
}

func TestHeadlessGameReset(t *testing.T) {
	game := NewHeadlessGame()

	game.StepN(100)
	game.Reset()

	if game.CurrentTick() != 0 {
		t.Errorf("Tick should be 0 after reset, got %d", game.CurrentTick())
	}
}
