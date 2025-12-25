package game

import (
	"math"
)

// TimeOfDay represents the current time phase.
type TimeOfDay int

const (
	TimeNight TimeOfDay = iota
	TimeDawn
	TimeDay
	TimeDusk
)

// String returns the time of day name.
func (t TimeOfDay) String() string {
	names := []string{"Night", "Dawn", "Day", "Dusk"}
	if int(t) < len(names) {
		return names[t]
	}

	return "Unknown"
}

// DayNightCycle manages time progression with day/night transitions.
type DayNightCycle struct {
	// Time settings
	CurrentTime float64 // Current time in hours (0-24)
	DayLength   float64 // Real seconds for a full day cycle
	TimeScale   float64 // Speed multiplier (1.0 = normal)
	Paused      bool

	// Transition times (in hours, 24h format)
	DawnStart  float64 // When dawn begins
	DayStart   float64 // When day begins
	DuskStart  float64 // When dusk begins
	NightStart float64 // When night begins

	// Visual parameters
	AmbientLight float64 // Current ambient light level (0-1)
	SunAngle     float64 // Sun position in degrees

	// Color temperatures (RGB multipliers)
	DayColor      [3]float64
	DawnDuskColor [3]float64
	NightColor    [3]float64
	CurrentColor  [3]float64

	// Callbacks
	onPhaseChange func(old, newPhase TimeOfDay)
	currentPhase  TimeOfDay
}

// NewDayNightCycle creates a day/night cycle with defaults.
func NewDayNightCycle(dayLengthSeconds float64) *DayNightCycle {
	return &DayNightCycle{
		CurrentTime:   12.0, // Start at noon
		DayLength:     dayLengthSeconds,
		TimeScale:     1.0,
		DawnStart:     5.0,
		DayStart:      7.0,
		DuskStart:     18.0,
		NightStart:    20.0,
		DayColor:      [3]float64{1.0, 1.0, 1.0},
		DawnDuskColor: [3]float64{1.0, 0.8, 0.6},
		NightColor:    [3]float64{0.3, 0.3, 0.5},
		currentPhase:  TimeDay,
	}
}

// SetOnPhaseChange sets the phase change callback.
func (d *DayNightCycle) SetOnPhaseChange(fn func(old, newPhase TimeOfDay)) {
	d.onPhaseChange = fn
}

// Update advances time.
func (d *DayNightCycle) Update(dt float64) {
	if d.Paused {
		return
	}

	// Convert real seconds to game hours
	hoursPerSecond := 24.0 / d.DayLength
	d.CurrentTime += dt * hoursPerSecond * d.TimeScale

	// Wrap around at midnight
	for d.CurrentTime >= 24.0 {
		d.CurrentTime -= 24.0
	}

	for d.CurrentTime < 0 {
		d.CurrentTime += 24.0
	}

	// Update phase
	newPhase := d.calculatePhase()
	if newPhase != d.currentPhase {
		oldPhase := d.currentPhase

		d.currentPhase = newPhase
		if d.onPhaseChange != nil {
			d.onPhaseChange(oldPhase, newPhase)
		}
	}

	// Update ambient light and colors
	d.updateLighting()
}

// calculatePhase determines current time of day.
func (d *DayNightCycle) calculatePhase() TimeOfDay {
	t := d.CurrentTime
	if t >= d.NightStart || t < d.DawnStart {
		return TimeNight
	}

	if t >= d.DawnStart && t < d.DayStart {
		return TimeDawn
	}

	if t >= d.DayStart && t < d.DuskStart {
		return TimeDay
	}

	return TimeDusk
}

// updateLighting calculates ambient light based on time.
func (d *DayNightCycle) updateLighting() {
	t := d.CurrentTime

	var (
		light float64
		color [3]float64
	)

	switch d.currentPhase {
	case TimeNight:
		light = 0.2
		color = d.NightColor

	case TimeDawn:
		// Transition from night to day
		progress := (t - d.DawnStart) / (d.DayStart - d.DawnStart)
		light = lerp(0.2, 1.0, progress)

		color = lerpColors(d.NightColor, d.DawnDuskColor, progress)
		if progress > 0.5 {
			color = lerpColors(d.DawnDuskColor, d.DayColor, (progress-0.5)*2)
		}

	case TimeDay:
		light = 1.0
		color = d.DayColor

	case TimeDusk:
		// Transition from day to night
		progress := (t - d.DuskStart) / (d.NightStart - d.DuskStart)

		light = lerp(1.0, 0.2, progress)
		if progress < 0.5 {
			color = lerpColors(d.DayColor, d.DawnDuskColor, progress*2)
		} else {
			color = lerpColors(d.DawnDuskColor, d.NightColor, (progress-0.5)*2)
		}
	}

	d.AmbientLight = light
	d.CurrentColor = color

	// Calculate sun angle (0 at midnight, 180 at noon)
	d.SunAngle = (d.CurrentTime / 24.0) * 360.0
}

// GetPhase returns the current time of day phase.
func (d *DayNightCycle) GetPhase() TimeOfDay {
	return d.currentPhase
}

// GetTimeString returns current time as "HH:MM".
func (d *DayNightCycle) GetTimeString() string {
	hours := int(d.CurrentTime)
	minutes := int((d.CurrentTime - float64(hours)) * 60)

	return formatTime(hours, minutes)
}

func formatTime(h, m int) string {
	return string(rune('0'+h/10)) + string(rune('0'+h%10)) + ":" +
		string(rune('0'+m/10)) + string(rune('0'+m%10))
}

// SetTime sets the current time directly.
func (d *DayNightCycle) SetTime(hours float64) {
	d.CurrentTime = hours
	for d.CurrentTime >= 24.0 {
		d.CurrentTime -= 24.0
	}

	for d.CurrentTime < 0 {
		d.CurrentTime += 24.0
	}

	d.updateLighting()
}

// IsNight returns true if it's currently night.
func (d *DayNightCycle) IsNight() bool {
	return d.currentPhase == TimeNight
}

// IsDay returns true if it's currently day.
func (d *DayNightCycle) IsDay() bool {
	return d.currentPhase == TimeDay
}

// GetAmbientColor returns RGBA values for ambient lighting.
func (d *DayNightCycle) GetAmbientColor() (r, g, b, a float64) {
	return d.CurrentColor[0], d.CurrentColor[1], d.CurrentColor[2], d.AmbientLight
}

func lerpColors(a, b [3]float64, t float64) [3]float64 {
	return [3]float64{
		lerp(a[0], b[0], t),
		lerp(a[1], b[1], t),
		lerp(a[2], b[2], t),
	}
}

// WeatherType represents different weather conditions.
type WeatherType int

const (
	WeatherClear WeatherType = iota
	WeatherCloudy
	WeatherRain
	WeatherStorm
	WeatherSnow
	WeatherFog
)

// String returns the weather name.
func (w WeatherType) String() string {
	names := []string{"Clear", "Cloudy", "Rain", "Storm", "Snow", "Fog"}
	if int(w) < len(names) {
		return names[w]
	}

	return "Unknown"
}

// WeatherEffects defines gameplay effects of weather.
type WeatherEffects struct {
	VisibilityMult float64 // Visibility multiplier
	MovementMult   float64 // Movement speed multiplier
	AccuracyMult   float64 // Projectile accuracy multiplier
	SoundRangeMult float64 // Sound detection range multiplier
	FireDamageMult float64 // Fire damage multiplier
	ColdDamageMult float64 // Cold damage multiplier
}

// WeatherSystem manages weather conditions.
type WeatherSystem struct {
	Current         WeatherType
	Intensity       float64 // 0.0 - 1.0
	TransitionTime  float64 // Time to transition between weather
	TransitionTimer float64
	targetWeather   WeatherType
	targetIntensity float64

	// Effects
	Effects WeatherEffects

	// Particle hints (for rendering)
	ParticleCount int
	WindX, WindY  float64

	// Automation
	AutoChange    bool
	MinDuration   float64
	MaxDuration   float64
	durationTimer float64

	onWeatherChange func(old, newWeather WeatherType)
}

// NewWeatherSystem creates a weather system.
func NewWeatherSystem() *WeatherSystem {
	return &WeatherSystem{
		Current:        WeatherClear,
		Intensity:      0.0,
		TransitionTime: 5.0,
		MinDuration:    30.0,
		MaxDuration:    120.0,
		Effects: WeatherEffects{
			VisibilityMult: 1.0,
			MovementMult:   1.0,
			AccuracyMult:   1.0,
			SoundRangeMult: 1.0,
			FireDamageMult: 1.0,
			ColdDamageMult: 1.0,
		},
	}
}

// SetOnWeatherChange sets the weather change callback.
func (w *WeatherSystem) SetOnWeatherChange(fn func(old, newWeather WeatherType)) {
	w.onWeatherChange = fn
}

// SetWeather immediately changes weather.
func (w *WeatherSystem) SetWeather(weather WeatherType, intensity float64) {
	old := w.Current
	w.Current = weather
	w.Intensity = intensity
	w.updateEffects()

	if w.onWeatherChange != nil && old != weather {
		w.onWeatherChange(old, weather)
	}
}

// TransitionTo starts a gradual weather transition.
func (w *WeatherSystem) TransitionTo(weather WeatherType, intensity float64) {
	w.targetWeather = weather
	w.targetIntensity = intensity
	w.TransitionTimer = w.TransitionTime
}

// Update updates the weather system.
func (w *WeatherSystem) Update(dt float64) {
	// Handle transition
	if w.TransitionTimer > 0 {
		w.TransitionTimer -= dt
		progress := 1.0 - (w.TransitionTimer / w.TransitionTime)
		w.Intensity = lerp(w.Intensity, w.targetIntensity, progress)

		if w.TransitionTimer <= 0 {
			old := w.Current
			w.Current = w.targetWeather

			w.Intensity = w.targetIntensity
			if w.onWeatherChange != nil && old != w.Current {
				w.onWeatherChange(old, w.Current)
			}
		}
	}

	// Auto-change weather
	if w.AutoChange && w.TransitionTimer <= 0 {
		w.durationTimer -= dt
		if w.durationTimer <= 0 {
			w.randomizeWeather()
		}
	}

	w.updateEffects()
}

// updateEffects calculates gameplay effects based on weather.
func (w *WeatherSystem) updateEffects() {
	e := &w.Effects
	i := w.Intensity

	// Reset to defaults
	*e = WeatherEffects{
		VisibilityMult: 1.0,
		MovementMult:   1.0,
		AccuracyMult:   1.0,
		SoundRangeMult: 1.0,
		FireDamageMult: 1.0,
		ColdDamageMult: 1.0,
	}

	switch w.Current {
	case WeatherRain:
		e.VisibilityMult = 1.0 - (0.3 * i)
		e.AccuracyMult = 1.0 - (0.1 * i)
		e.FireDamageMult = 1.0 - (0.2 * i)
		e.SoundRangeMult = 1.0 - (0.2 * i)
		w.ParticleCount = int(500 * i)
		w.WindX = 20 * i

	case WeatherStorm:
		e.VisibilityMult = 1.0 - (0.5 * i)
		e.AccuracyMult = 1.0 - (0.3 * i)
		e.MovementMult = 1.0 - (0.1 * i)
		e.FireDamageMult = 1.0 - (0.4 * i)
		e.SoundRangeMult = 1.0 - (0.5 * i)
		w.ParticleCount = int(1000 * i)
		w.WindX = 50 * i

	case WeatherSnow:
		e.VisibilityMult = 1.0 - (0.2 * i)
		e.MovementMult = 1.0 - (0.15 * i)
		e.ColdDamageMult = 1.0 + (0.3 * i)
		e.FireDamageMult = 1.0 - (0.1 * i)
		w.ParticleCount = int(300 * i)

	case WeatherFog:
		e.VisibilityMult = 1.0 - (0.6 * i)
		e.SoundRangeMult = 1.0 - (0.3 * i)
		w.ParticleCount = 0

	case WeatherCloudy:
		e.VisibilityMult = 1.0 - (0.1 * i)
		w.ParticleCount = 0

	case WeatherClear:
		w.ParticleCount = 0
		w.WindX = 0
		w.WindY = 0
	}
}

// randomizeWeather picks a random weather for auto-change.
func (w *WeatherSystem) randomizeWeather() {
	// Simple weighted random
	weather := WeatherType(int(math.Mod(float64(int(w.Current)+1+int(w.Intensity*3)), 6)))

	intensity := 0.3 + (math.Mod(w.durationTimer*1000, 7) / 10.0)
	if intensity > 1.0 {
		intensity = 1.0
	}

	w.TransitionTo(weather, intensity)
	w.durationTimer = w.MinDuration + (w.MaxDuration-w.MinDuration)*(intensity)
}

// IsRaining returns true if raining or storming.
func (w *WeatherSystem) IsRaining() bool {
	return w.Current == WeatherRain || w.Current == WeatherStorm
}

// IsCold returns true if snowing.
func (w *WeatherSystem) IsCold() bool {
	return w.Current == WeatherSnow
}

// GetVisibility returns the visibility multiplier.
func (w *WeatherSystem) GetVisibility() float64 {
	return w.Effects.VisibilityMult
}
