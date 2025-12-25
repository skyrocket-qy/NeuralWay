package components

// Combat represents combat stats for an entity.
type Combat struct {
	AttackPower  float64
	Defense      float64
	AttackSpeed  float64 // Attacks per second
	Range        float64 // Attack range
	DamageType   DamageType
	CanAttack    bool
	LastAttackAt float64 // Time of last attack
}

// DamageType represents the type of damage dealt.
type DamageType string

const (
	DamagePhysical  DamageType = "physical"
	DamageFire      DamageType = "fire"
	DamageCold      DamageType = "cold"
	DamageLightning DamageType = "lightning"
	DamageChaos     DamageType = "chaos"
)

// NewCombat creates combat stats with default values.
func NewCombat(attackPower, defense float64) Combat {
	return Combat{
		AttackPower:  attackPower,
		Defense:      defense,
		AttackSpeed:  1.0,
		Range:        50,
		DamageType:   DamagePhysical,
		CanAttack:    true,
		LastAttackAt: 0,
	}
}

// Cooldown represents an ability with a cooldown timer.
type Cooldown struct {
	Duration    float64 // Total cooldown duration in seconds
	Remaining   float64 // Time remaining until ready
	Charges     int     // Current charges (for multi-charge abilities)
	MaxCharges  int     // Maximum charges
	ChargeTime  float64 // Time to regain one charge
	ChargeTimer float64 // Current charge regeneration timer
}

// NewCooldown creates a cooldown with the given duration.
func NewCooldown(duration float64) Cooldown {
	return Cooldown{
		Duration:   duration,
		Remaining:  0,
		Charges:    1,
		MaxCharges: 1,
	}
}

// NewCooldownWithCharges creates a multi-charge cooldown.
func NewCooldownWithCharges(chargeTime float64, maxCharges int) Cooldown {
	return Cooldown{
		Duration:   chargeTime,
		ChargeTime: chargeTime,
		Charges:    maxCharges,
		MaxCharges: maxCharges,
	}
}

// IsReady returns true if the ability can be used.
func (c *Cooldown) IsReady() bool {
	if c.MaxCharges > 1 {
		return c.Charges > 0
	}

	return c.Remaining <= 0
}

// Use triggers the cooldown if ready, returns true if successful.
func (c *Cooldown) Use() bool {
	if !c.IsReady() {
		return false
	}

	if c.MaxCharges > 1 {
		c.Charges--
	} else {
		c.Remaining = c.Duration
	}

	return true
}

// Combo tracks combo attacks for an entity.
type Combo struct {
	Count       int     // Current combo count
	MaxCount    int     // Maximum combo before reset
	Timer       float64 // Time remaining before combo resets
	Window      float64 // Time window to continue combo
	Multiplier  float64 // Damage multiplier per combo
	BonusPerHit float64 // Additional multiplier per hit
}

// NewCombo creates a combo tracker.
func NewCombo(window float64) Combo {
	return Combo{
		Window:      window,
		Multiplier:  1.0,
		BonusPerHit: 0.1, // 10% per hit
	}
}

// AddHit registers a combo hit and returns the new multiplier.
func (c *Combo) AddHit() float64 {
	c.Count++

	c.Timer = c.Window
	if c.MaxCount > 0 && c.Count > c.MaxCount {
		c.Count = c.MaxCount
	}

	return c.Multiplier + (c.BonusPerHit * float64(c.Count-1))
}

// Reset clears the combo.
func (c *Combo) Reset() {
	c.Count = 0
	c.Timer = 0
}

// CriticalHit represents critical strike chance and damage.
type CriticalHit struct {
	Chance     float64 // 0.0 to 1.0 chance to crit
	Multiplier float64 // Damage multiplier on crit (e.g., 2.0 = 200%)
	Guaranteed bool    // Next hit is guaranteed crit
}

// NewCriticalHit creates a crit component with default values.
func NewCriticalHit(chance, multiplier float64) CriticalHit {
	return CriticalHit{
		Chance:     chance,
		Multiplier: multiplier,
	}
}

// Mana represents a resource pool for abilities.
type Mana struct {
	Current    float64
	Max        float64
	Regen      float64 // Mana regenerated per second
	RegenDelay float64 // Delay before regen starts after use
	DelayTimer float64 // Current delay timer
}

// NewMana creates a mana pool.
func NewMana(maxVal float64) Mana {
	return Mana{
		Current: maxVal,
		Max:     maxVal,
		Regen:   1.0,
	}
}

// Use consumes mana if available, returns true if successful.
func (m *Mana) Use(cost float64) bool {
	if m.Current < cost {
		return false
	}

	m.Current -= cost
	m.DelayTimer = m.RegenDelay

	return true
}

// Buff represents a temporary stat modification.
type Buff struct {
	Name       string
	StatType   string  // Which stat is modified
	Modifier   float64 // Additive modifier
	Multiplier float64 // Multiplicative modifier (1.0 = no change)
	Duration   float64 // Time remaining
	Stacks     int     // Current stacks
	MaxStacks  int     // Maximum stacks
}

// NewBuff creates a new buff.
func NewBuff(name, statType string, modifier, duration float64) Buff {
	return Buff{
		Name:       name,
		StatType:   statType,
		Modifier:   modifier,
		Multiplier: 1.0,
		Duration:   duration,
		Stacks:     1,
		MaxStacks:  1,
	}
}

// BuffContainer holds active buffs/debuffs on an entity.
type BuffContainer struct {
	Buffs []Buff
}

// NewBuffContainer creates an empty buff container.
func NewBuffContainer() BuffContainer {
	return BuffContainer{
		Buffs: make([]Buff, 0),
	}
}

// AddBuff adds or stacks a buff.
func (bc *BuffContainer) AddBuff(buff Buff) {
	// Check for existing buff with same name
	for i := range bc.Buffs {
		if bc.Buffs[i].Name == buff.Name {
			// Stack or refresh
			if bc.Buffs[i].Stacks < bc.Buffs[i].MaxStacks {
				bc.Buffs[i].Stacks++
			}

			bc.Buffs[i].Duration = buff.Duration

			return
		}
	}

	bc.Buffs = append(bc.Buffs, buff)
}

// GetModifier returns the total modifier for a stat type.
func (bc *BuffContainer) GetModifier(statType string) (additive, multiplicative float64) {
	multiplicative = 1.0

	for _, b := range bc.Buffs {
		if b.StatType == statType {
			additive += b.Modifier * float64(b.Stacks)
			multiplicative *= b.Multiplier
		}
	}

	return additive, multiplicative
}

// RemoveExpired removes all expired buffs.
func (bc *BuffContainer) RemoveExpired() {
	active := bc.Buffs[:0]
	for _, b := range bc.Buffs {
		if b.Duration > 0 {
			active = append(active, b)
		}
	}

	bc.Buffs = active
}
