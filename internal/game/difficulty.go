package game

// DifficultyLevel represents a preset difficulty.
type DifficultyLevel int

const (
	DifficultyEasy DifficultyLevel = iota
	DifficultyNormal
	DifficultyHard
	DifficultyNightmare
	DifficultyCustom
)

// String returns the difficulty name.
func (d DifficultyLevel) String() string {
	names := []string{"Easy", "Normal", "Hard", "Nightmare", "Custom"}
	if int(d) < len(names) {
		return names[d]
	}

	return "Unknown"
}

// DifficultySettings contains all difficulty-related multipliers.
type DifficultySettings struct {
	Level DifficultyLevel

	// Enemy modifiers
	EnemyHealth float64 // Multiplier for enemy HP
	EnemyDamage float64 // Multiplier for enemy damage
	EnemySpeed  float64 // Multiplier for enemy movement speed
	EnemyCount  float64 // Multiplier for enemy spawn count

	// Player modifiers
	PlayerHealth float64 // Multiplier for player HP
	PlayerDamage float64 // Multiplier for player damage
	PlayerSpeed  float64 // Multiplier for player speed

	// Economy modifiers
	XPGain   float64 // Multiplier for XP gained
	GoldGain float64 // Multiplier for gold/currency gained
	DropRate float64 // Multiplier for item drop rate

	// Gameplay modifiers
	WaveInterval float64 // Multiplier for time between waves
	RespawnTime  float64 // Multiplier for respawn cooldown
	CooldownRate float64 // Multiplier for ability cooldowns

	// Dynamic difficulty
	DynamicEnabled bool
	TargetWinRate  float64 // Target win rate for dynamic (0.5 = 50%)
}

// NewDifficultySettings creates default (Normal) settings.
func NewDifficultySettings() *DifficultySettings {
	return &DifficultySettings{
		Level:         DifficultyNormal,
		EnemyHealth:   1.0,
		EnemyDamage:   1.0,
		EnemySpeed:    1.0,
		EnemyCount:    1.0,
		PlayerHealth:  1.0,
		PlayerDamage:  1.0,
		PlayerSpeed:   1.0,
		XPGain:        1.0,
		GoldGain:      1.0,
		DropRate:      1.0,
		WaveInterval:  1.0,
		RespawnTime:   1.0,
		CooldownRate:  1.0,
		TargetWinRate: 0.5,
	}
}

// SetLevel applies a preset difficulty level.
func (d *DifficultySettings) SetLevel(level DifficultyLevel) {
	d.Level = level

	switch level {
	case DifficultyEasy:
		d.EnemyHealth = 0.75
		d.EnemyDamage = 0.75
		d.EnemySpeed = 0.9
		d.EnemyCount = 0.8
		d.PlayerHealth = 1.25
		d.PlayerDamage = 1.1
		d.XPGain = 1.25
		d.GoldGain = 1.25
		d.DropRate = 1.5

	case DifficultyNormal:
		d.EnemyHealth = 1.0
		d.EnemyDamage = 1.0
		d.EnemySpeed = 1.0
		d.EnemyCount = 1.0
		d.PlayerHealth = 1.0
		d.PlayerDamage = 1.0
		d.XPGain = 1.0
		d.GoldGain = 1.0
		d.DropRate = 1.0

	case DifficultyHard:
		d.EnemyHealth = 1.5
		d.EnemyDamage = 1.35
		d.EnemySpeed = 1.15
		d.EnemyCount = 1.25
		d.PlayerHealth = 0.9
		d.PlayerDamage = 0.95
		d.XPGain = 1.5
		d.GoldGain = 1.5
		d.DropRate = 1.25

	case DifficultyNightmare:
		d.EnemyHealth = 2.5
		d.EnemyDamage = 2.0
		d.EnemySpeed = 1.3
		d.EnemyCount = 1.75
		d.PlayerHealth = 0.75
		d.PlayerDamage = 0.85
		d.XPGain = 2.5
		d.GoldGain = 2.5
		d.DropRate = 2.0

	case DifficultyCustom:
		// Keep current settings
	}
}

// DifficultyManager manages difficulty and dynamic adjustment.
type DifficultyManager struct {
	Settings *DifficultySettings

	// Dynamic difficulty tracking
	WinsRecent    int     // Recent wins
	LossesRecent  int     // Recent losses
	HistorySize   int     // How many results to track
	AdjustRate    float64 // How much to adjust per evaluation
	MinMultiplier float64 // Minimum overall difficulty
	MaxMultiplier float64 // Maximum overall difficulty
	CurrentScale  float64 // Current dynamic scale
}

// NewDifficultyManager creates a difficulty manager.
func NewDifficultyManager() *DifficultyManager {
	return &DifficultyManager{
		Settings:      NewDifficultySettings(),
		HistorySize:   10,
		AdjustRate:    0.05,
		MinMultiplier: 0.5,
		MaxMultiplier: 2.0,
		CurrentScale:  1.0,
	}
}

// SetDifficulty sets the difficulty level.
func (dm *DifficultyManager) SetDifficulty(level DifficultyLevel) {
	dm.Settings.SetLevel(level)
}

// RecordWin records a player win for dynamic adjustment.
func (dm *DifficultyManager) RecordWin() {
	dm.WinsRecent++
	dm.evaluate()
}

// RecordLoss records a player loss for dynamic adjustment.
func (dm *DifficultyManager) RecordLoss() {
	dm.LossesRecent++
	dm.evaluate()
}

// evaluate adjusts difficulty based on win rate.
func (dm *DifficultyManager) evaluate() {
	if !dm.Settings.DynamicEnabled {
		return
	}

	total := dm.WinsRecent + dm.LossesRecent
	if total < dm.HistorySize/2 {
		return // Not enough data
	}

	winRate := float64(dm.WinsRecent) / float64(total)
	target := dm.Settings.TargetWinRate

	if winRate > target+0.1 {
		// Player winning too much, increase difficulty
		dm.CurrentScale += dm.AdjustRate
	} else if winRate < target-0.1 {
		// Player losing too much, decrease difficulty
		dm.CurrentScale -= dm.AdjustRate
	}

	// Clamp
	if dm.CurrentScale < dm.MinMultiplier {
		dm.CurrentScale = dm.MinMultiplier
	}

	if dm.CurrentScale > dm.MaxMultiplier {
		dm.CurrentScale = dm.MaxMultiplier
	}

	// Reset history periodically
	if total >= dm.HistorySize {
		dm.WinsRecent = dm.WinsRecent / 2
		dm.LossesRecent = dm.LossesRecent / 2
	}
}

// GetEnemyHealth returns adjusted enemy health.
func (dm *DifficultyManager) GetEnemyHealth(base float64) float64 {
	return base * dm.Settings.EnemyHealth * dm.CurrentScale
}

// GetEnemyDamage returns adjusted enemy damage.
func (dm *DifficultyManager) GetEnemyDamage(base float64) float64 {
	return base * dm.Settings.EnemyDamage * dm.CurrentScale
}

// GetEnemySpeed returns adjusted enemy speed.
func (dm *DifficultyManager) GetEnemySpeed(base float64) float64 {
	return base * dm.Settings.EnemySpeed * dm.CurrentScale
}

// GetEnemyCount returns adjusted enemy count.
func (dm *DifficultyManager) GetEnemyCount(base int) int {
	return int(float64(base) * dm.Settings.EnemyCount * dm.CurrentScale)
}

// GetPlayerHealth returns adjusted player health.
func (dm *DifficultyManager) GetPlayerHealth(base float64) float64 {
	return base * dm.Settings.PlayerHealth
}

// GetPlayerDamage returns adjusted player damage.
func (dm *DifficultyManager) GetPlayerDamage(base float64) float64 {
	return base * dm.Settings.PlayerDamage
}

// GetXPGain returns adjusted XP gain.
func (dm *DifficultyManager) GetXPGain(base int64) int64 {
	return int64(float64(base) * dm.Settings.XPGain)
}

// GetGoldGain returns adjusted gold gain.
func (dm *DifficultyManager) GetGoldGain(base int64) int64 {
	return int64(float64(base) * dm.Settings.GoldGain)
}

// GetDropRate returns adjusted drop rate.
func (dm *DifficultyManager) GetDropRate(base float64) float64 {
	return base * dm.Settings.DropRate
}

// EnableDynamic enables dynamic difficulty adjustment.
func (dm *DifficultyManager) EnableDynamic(enable bool) {
	dm.Settings.DynamicEnabled = enable
}

// SetTargetWinRate sets the target win rate for dynamic difficulty.
func (dm *DifficultyManager) SetTargetWinRate(rate float64) {
	dm.Settings.TargetWinRate = rate
}

// GetCurrentScale returns the current dynamic scale.
func (dm *DifficultyManager) GetCurrentScale() float64 {
	return dm.CurrentScale
}

// Reset resets the difficulty manager.
func (dm *DifficultyManager) Reset() {
	dm.WinsRecent = 0
	dm.LossesRecent = 0
	dm.CurrentScale = 1.0
}

// GetDescription returns a human-readable difficulty description.
func (dm *DifficultyManager) GetDescription() string {
	level := dm.Settings.Level.String()
	if dm.Settings.DynamicEnabled {
		return level + " (Dynamic)"
	}

	return level
}
