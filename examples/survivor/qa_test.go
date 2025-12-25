package main

import (
	"fmt"
	"testing"
	"time"

	"td/internal/ai"
)

// TestQASession runs automated QA testing on the Survivor game.
func TestQASession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping QA session in short mode")
	}

	// Create game and adapter
	game := NewGame()
	adapter := NewSurvivorAdapter(game)

	// Create QA session
	session := ai.NewQASession(adapter)
	session.SetPlayer(ai.NewRandomPlayer(time.Now().UnixNano()))
	session.SetConfig(ai.SessionConfig{
		Runs:        3,
		MaxTicks:    1800, // 30 seconds per run
		RecordEvery: 10,   // Record every 10 ticks
	})

	// Configure detector
	detector := ai.NewAnomalyDetector()
	detector.EntityLeakThreshold = 300 // More lenient for survivor
	detector.StuckThreshold = 300      // 5 seconds stuck
	detector.BoundsWidth = 10000       // Large world
	detector.BoundsHeight = 10000
	session.SetDetector(detector)

	// Run QA session
	report := session.Run()

	// Output report
	t.Log("\n" + report.GenerateMarkdown())

	// Check results
	if report.TotalAnomalies > 5 {
		t.Errorf("Too many anomalies detected: %d", report.TotalAnomalies)
	}
}

// TestSurvivorAdapter tests the adapter directly.
func TestSurvivorAdapter(t *testing.T) {
	t.Run("Adapter implements interface", func(t *testing.T) {
		game := NewGame()
		adapter := NewSurvivorAdapter(game)

		// Start game
		err := adapter.Reset()
		if err != nil {
			t.Fatalf("Reset failed: %v", err)
		}

		if adapter.Name() != "Dev Survivor" {
			t.Errorf("Name = %s, want Dev Survivor", adapter.Name())
		}
	})

	t.Run("GetState returns valid state", func(t *testing.T) {
		game := NewGame()
		adapter := NewSurvivorAdapter(game)
		adapter.Reset()

		state := adapter.GetState()
		if state.PlayerHealth[1] == 0 {
			t.Error("MaxHP should not be 0 after reset")
		}
	})

	t.Run("PerformAction updates movement", func(t *testing.T) {
		game := NewGame()
		adapter := NewSurvivorAdapter(game)
		adapter.Reset()

		initialState := adapter.GetState()
		initialX := initialState.PlayerPos[0]

		// Move right for several ticks
		for range 10 {
			adapter.PerformAction(ai.ActionMoveRight)
			adapter.Step()
		}

		newState := adapter.GetState()
		if newState.PlayerPos[0] <= initialX {
			t.Error("Player should have moved right")
		}
	})

	t.Run("IsGameOver detects death", func(t *testing.T) {
		game := NewGame()
		adapter := NewSurvivorAdapter(game)
		adapter.Reset()

		// Initially not game over
		if adapter.IsGameOver() {
			t.Error("Game should not be over immediately after reset")
		}
	})
}

// TestRandomPlaythrough runs a random playthrough for basic coverage.
func TestRandomPlaythrough(t *testing.T) {
	game := NewGame()
	adapter := NewSurvivorAdapter(game)
	adapter.Reset()

	player := ai.NewRandomPlayer(42)
	available := adapter.AvailableActions()

	// Run for 600 ticks (10 seconds)
	for i := range 600 {
		state := adapter.GetState()
		action := player.DecideAction(state, available)
		adapter.PerformAction(action)
		adapter.Step()

		if adapter.IsGameOver() {
			t.Logf("Game over at tick %d, score: %d", i, adapter.GetScore())

			return
		}
	}

	finalState := adapter.GetState()
	t.Logf("Survived 600 ticks, HP: %d/%d, Kills: %d, Enemies: %d",
		finalState.PlayerHealth[0], finalState.PlayerHealth[1],
		adapter.GetScore(), adapter.GetEnemyCount())
}

// BenchmarkGameStep benchmarks a single game step.
func BenchmarkGameStep(b *testing.B) {
	game := NewGame()
	adapter := NewSurvivorAdapter(game)
	adapter.Reset()

	for b.Loop() {
		adapter.PerformAction(ai.ActionMoveRight)
		adapter.Step()
	}
}

// ExampleQAReport shows how to generate a QA report.
func ExampleQAReport() {
	game := NewGame()
	adapter := NewSurvivorAdapter(game)

	session := ai.NewQASession(adapter)
	session.SetPlayer(ai.NewRandomPlayer(12345))
	session.SetConfig(ai.SessionConfig{
		Runs:     1,
		MaxTicks: 100,
	})

	report := session.Run()
	fmt.Println(report.Conclusion)
	// Output: PASS - No anomalies detected
}
