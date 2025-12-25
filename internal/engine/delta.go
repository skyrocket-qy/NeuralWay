package engine

import "time"

// DeltaTime manages frame-rate independent timing.
type DeltaTime struct {
	lastTime    time.Time
	delta       float64
	accumulator float64
	fixedStep   float64
}

// NewDeltaTime creates a delta time tracker with a fixed physics step.
func NewDeltaTime(fixedStepHz float64) *DeltaTime {
	return &DeltaTime{
		lastTime:  time.Now(),
		delta:     0,
		fixedStep: 1.0 / fixedStepHz,
	}
}

// Update calculates the delta time since last frame.
// Returns the delta in seconds.
func (dt *DeltaTime) Update() float64 {
	now := time.Now()
	dt.delta = now.Sub(dt.lastTime).Seconds()
	dt.lastTime = now

	return dt.delta
}

// Delta returns the current delta time in seconds.
func (dt *DeltaTime) Delta() float64 {
	return dt.delta
}

// Accumulate adds delta to the accumulator for fixed timestep physics.
// Returns true if a fixed step should be processed.
func (dt *DeltaTime) Accumulate() bool {
	dt.accumulator += dt.delta
	if dt.accumulator >= dt.fixedStep {
		dt.accumulator -= dt.fixedStep

		return true
	}

	return false
}

// FixedStep returns the fixed timestep duration.
func (dt *DeltaTime) FixedStep() float64 {
	return dt.fixedStep
}
