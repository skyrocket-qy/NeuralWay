package engine

import (
	"time"

	"github.com/mlange-42/ark/ecs"
)

// HeadlessGame runs the game loop without GPU/rendering for fast AI training.
// It's typically 1000x+ faster than real-time simulation.
type HeadlessGame struct {
	World         ecs.World
	updateSystems []System
	tickRate      int   // Target ticks per second (0 = max speed)
	currentTick   int64 // Current simulation tick
}

// NewHeadlessGame creates a new headless game instance.
func NewHeadlessGame() *HeadlessGame {
	return &HeadlessGame{
		World:         ecs.NewWorld(),
		updateSystems: make([]System, 0),
		tickRate:      0, // Max speed by default
		currentTick:   0,
	}
}

// AddSystem adds an update system to the headless game loop.
func (g *HeadlessGame) AddSystem(s System) {
	g.updateSystems = append(g.updateSystems, s)
}

// SetTickRate sets the target ticks per second.
// Set to 0 for maximum speed (no throttling).
func (g *HeadlessGame) SetTickRate(tps int) {
	g.tickRate = tps
}

// CurrentTick returns the current simulation tick count.
func (g *HeadlessGame) CurrentTick() int64 {
	return g.currentTick
}

// Step runs one update tick.
func (g *HeadlessGame) Step() {
	for _, s := range g.updateSystems {
		s.Update(&g.World)
	}

	g.currentTick++
}

// StepN runs N update ticks.
func (g *HeadlessGame) StepN(n int) {
	for range n {
		g.Step()
	}
}

// RunFor runs the simulation for the specified duration at the current tick rate.
// If tickRate is 0, runs as fast as possible for the equivalent of that duration at 60 TPS.
func (g *HeadlessGame) RunFor(duration time.Duration) {
	tps := g.tickRate
	if tps == 0 {
		tps = 60 // Default to 60 TPS equivalent for duration calculation
	}

	ticks := int(duration.Seconds() * float64(tps))
	g.StepN(ticks)
}

// RunUntil runs the simulation until the condition returns true.
// Returns the number of ticks executed.
// maxTicks prevents infinite loops (0 = no limit).
func (g *HeadlessGame) RunUntil(cond func(*ecs.World) bool, maxTicks int) int {
	ticks := 0

	for {
		if cond(&g.World) {
			return ticks
		}

		if maxTicks > 0 && ticks >= maxTicks {
			return ticks
		}

		g.Step()

		ticks++
	}
}

// RunWithCallback runs ticks and calls callback after each step.
// Useful for recording state or checking conditions.
func (g *HeadlessGame) RunWithCallback(n int, callback func(*ecs.World, int64)) {
	for range n {
		g.Step()
		callback(&g.World, g.currentTick)
	}
}

// Reset clears the world and resets tick counter.
func (g *HeadlessGame) Reset() {
	g.World = ecs.NewWorld()
	g.currentTick = 0
}
