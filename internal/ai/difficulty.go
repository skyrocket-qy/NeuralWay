package ai

import "math"

// DifficultyBalancer tracks player performance and adjusts game parameters.
type DifficultyBalancer struct {
	// Tracked metrics (e.g., "win_rate", "deaths", "completion_time")
	metrics map[string]*MetricTracker

	// Adjustable parameters (e.g., "enemy_speed", "spawn_rate", "damage_mult")
	parameters map[string]*Parameter

	// Global difficulty multiplier (1.0 = normal)
	difficultyLevel float64

	// Smoothing factor for adjustments (0-1, lower = smoother)
	adjustmentRate float64
}

// MetricTracker tracks a rolling average of a metric.
type MetricTracker struct {
	values     []float64
	maxSamples int
	target     float64 // Target value for this metric
	tolerance  float64 // Acceptable deviation from target
}

// Parameter represents an adjustable game parameter.
type Parameter struct {
	Value    float64
	Min      float64
	Max      float64
	Default  float64
	Inverted bool // If true, lower is harder
}

// NewDifficultyBalancer creates a new difficulty balancer.
func NewDifficultyBalancer() *DifficultyBalancer {
	return &DifficultyBalancer{
		metrics:         make(map[string]*MetricTracker),
		parameters:      make(map[string]*Parameter),
		difficultyLevel: 1.0,
		adjustmentRate:  0.1,
	}
}

// RegisterMetric registers a metric to track.
func (d *DifficultyBalancer) RegisterMetric(name string, target, tolerance float64, maxSamples int) {
	d.metrics[name] = &MetricTracker{
		values:     make([]float64, 0, maxSamples),
		maxSamples: maxSamples,
		target:     target,
		tolerance:  tolerance,
	}
}

// RegisterParameter registers an adjustable parameter.
func (d *DifficultyBalancer) RegisterParameter(
	name string,
	defaultVal, minVal, maxVal float64,
	inverted bool,
) {
	d.parameters[name] = &Parameter{
		Value:    defaultVal,
		Min:      minVal,
		Max:      maxVal,
		Default:  defaultVal,
		Inverted: inverted,
	}
}

// RecordMetric records a metric value.
func (d *DifficultyBalancer) RecordMetric(name string, value float64) {
	tracker, ok := d.metrics[name]
	if !ok {
		return
	}

	tracker.values = append(tracker.values, value)
	if len(tracker.values) > tracker.maxSamples {
		tracker.values = tracker.values[1:]
	}
}

// GetMetricAverage returns the rolling average of a metric.
func (d *DifficultyBalancer) GetMetricAverage(name string) float64 {
	tracker, ok := d.metrics[name]
	if !ok || len(tracker.values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range tracker.values {
		sum += v
	}

	return sum / float64(len(tracker.values))
}

// Update adjusts difficulty based on tracked metrics.
func (d *DifficultyBalancer) Update() {
	// Calculate overall performance score
	performanceScore := 0.0
	metricCount := 0

	for _, tracker := range d.metrics {
		if len(tracker.values) < 3 {
			continue // Need enough samples
		}

		avg := 0.0
		for _, v := range tracker.values {
			avg += v
		}

		avg /= float64(len(tracker.values))

		// How far from target? Positive = doing too well
		deviation := (avg - tracker.target) / tracker.tolerance
		performanceScore += deviation
		metricCount++
	}

	if metricCount == 0 {
		return
	}

	performanceScore /= float64(metricCount)

	// Adjust difficulty level
	// If performance is positive (doing too well), increase difficulty
	adjustment := performanceScore * d.adjustmentRate
	d.difficultyLevel += adjustment

	// Clamp difficulty level
	d.difficultyLevel = clamp(d.difficultyLevel, 0.5, 2.0)

	// Update all parameters based on difficulty level
	for _, param := range d.parameters {
		range_ := param.Max - param.Min
		normalized := (d.difficultyLevel - 0.5) / 1.5 // 0 to 1

		if param.Inverted {
			normalized = 1 - normalized
		}

		param.Value = param.Min + (range_ * normalized)
	}
}

// GetParameter returns the current value of a parameter.
func (d *DifficultyBalancer) GetParameter(name string) float64 {
	param, ok := d.parameters[name]
	if !ok {
		return 0
	}

	return param.Value
}

// GetDifficultyLevel returns the current difficulty multiplier.
func (d *DifficultyBalancer) GetDifficultyLevel() float64 {
	return d.difficultyLevel
}

// SetDifficultyLevel manually sets the difficulty level.
func (d *DifficultyBalancer) SetDifficultyLevel(level float64) {
	d.difficultyLevel = clamp(level, 0.5, 2.0)
}

// SetAdjustmentRate sets how quickly difficulty adjusts.
func (d *DifficultyBalancer) SetAdjustmentRate(rate float64) {
	d.adjustmentRate = clamp(rate, 0.01, 0.5)
}

// Reset resets all metrics and parameters to defaults.
func (d *DifficultyBalancer) Reset() {
	d.difficultyLevel = 1.0

	for _, tracker := range d.metrics {
		tracker.values = tracker.values[:0]
	}

	for _, param := range d.parameters {
		param.Value = param.Default
	}
}

// clamp restricts a value to a range.
func clamp(value, minVal, maxVal float64) float64 {
	return math.Max(minVal, math.Min(maxVal, value))
}
