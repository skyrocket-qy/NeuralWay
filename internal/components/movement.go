package components

// MovementMode represents the current movement state.
type MovementMode int

const (
	MovementWalk MovementMode = iota
	MovementRun
	MovementCrouch
	MovementSwim
	MovementFly
	MovementClimb
)

// Movement represents advanced movement capabilities.
type Movement struct {
	// Speed settings
	WalkSpeed   float64
	RunSpeed    float64
	CrouchSpeed float64
	SwimSpeed   float64
	FlySpeed    float64
	ClimbSpeed  float64

	// Current state
	CurrentMode     MovementMode
	CurrentSpeed    float64
	IsMoving        bool
	IsSprinting     bool
	Direction       float64 // Radians, 0 = right
	FacingDirection float64 // Which way entity faces

	// Acceleration
	Acceleration float64 // How fast to reach max speed
	Deceleration float64 // How fast to stop
	AirControl   float64 // Control factor while airborne (0-1)

	// Movement modifiers
	SpeedMultiplier float64
	Frozen          bool
	CanMove         bool
}

// NewMovement creates a movement component with defaults.
func NewMovement(walkSpeed float64) Movement {
	return Movement{
		WalkSpeed:       walkSpeed,
		RunSpeed:        walkSpeed * 1.5,
		CrouchSpeed:     walkSpeed * 0.5,
		SwimSpeed:       walkSpeed * 0.7,
		FlySpeed:        walkSpeed * 1.2,
		ClimbSpeed:      walkSpeed * 0.4,
		CurrentMode:     MovementWalk,
		CurrentSpeed:    0,
		Acceleration:    1000, // Pixels per second squared
		Deceleration:    800,
		AirControl:      0.3,
		SpeedMultiplier: 1.0,
		CanMove:         true,
	}
}

// GetMaxSpeed returns the max speed for current mode.
func (m *Movement) GetMaxSpeed() float64 {
	var speed float64

	switch m.CurrentMode {
	case MovementWalk:
		if m.IsSprinting {
			speed = m.RunSpeed
		} else {
			speed = m.WalkSpeed
		}
	case MovementRun:
		speed = m.RunSpeed
	case MovementCrouch:
		speed = m.CrouchSpeed
	case MovementSwim:
		speed = m.SwimSpeed
	case MovementFly:
		speed = m.FlySpeed
	case MovementClimb:
		speed = m.ClimbSpeed
	default:
		speed = m.WalkSpeed
	}

	return speed * m.SpeedMultiplier
}

// SetMode changes movement mode.
func (m *Movement) SetMode(mode MovementMode) {
	m.CurrentMode = mode
}

// StartSprint enables sprinting.
func (m *Movement) StartSprint() {
	m.IsSprinting = true
}

// StopSprint disables sprinting.
func (m *Movement) StopSprint() {
	m.IsSprinting = false
}

// Jump represents jumping/falling state.
type Jump struct {
	JumpForce        float64 // Initial jump velocity
	MaxJumps         int     // Total jumps allowed (1 = single, 2 = double jump)
	JumpsRemaining   int     // Jumps left before landing
	IsJumping        bool    // Currently in jump motion
	IsFalling        bool    // Falling (not jumping)
	IsGrounded       bool    // On ground
	CoyoteTime       float64 // Grace period after leaving edge
	CoyoteTimer      float64
	JumpBufferTime   float64 // Buffer for jump input
	JumpBufferTimer  float64
	GravityScale     float64 // Multiplier for gravity (1.0 = normal)
	TerminalVelocity float64 // Max fall speed
	VerticalVelocity float64 // Current Y velocity
	CanJump          bool
}

// NewJump creates a jumping component with defaults.
func NewJump(jumpForce float64) Jump {
	return Jump{
		JumpForce:        jumpForce,
		MaxJumps:         1,
		JumpsRemaining:   1,
		CoyoteTime:       0.1,
		JumpBufferTime:   0.1,
		GravityScale:     1.0,
		TerminalVelocity: 800,
		CanJump:          true,
	}
}

// TryJump attempts to perform a jump.
func (j *Jump) TryJump() bool {
	if !j.CanJump {
		return false
	}

	// Check coyote time (recently grounded)
	canCoyote := j.CoyoteTimer > 0 && !j.IsJumping

	if j.IsGrounded || canCoyote || j.JumpsRemaining > 0 {
		j.VerticalVelocity = -j.JumpForce
		j.IsJumping = true
		j.IsGrounded = false
		j.IsFalling = false
		j.CoyoteTimer = 0

		if j.JumpsRemaining > 0 {
			j.JumpsRemaining--
		}

		return true
	}

	// Buffer the jump for later
	j.JumpBufferTimer = j.JumpBufferTime

	return false
}

// Land resets jump state when landing.
func (j *Jump) Land() {
	j.IsGrounded = true
	j.IsJumping = false
	j.IsFalling = false
	j.VerticalVelocity = 0
	j.JumpsRemaining = j.MaxJumps

	// Check jump buffer
	if j.JumpBufferTimer > 0 {
		j.TryJump()
	}
}

// LeaveGround called when walking off an edge.
func (j *Jump) LeaveGround() {
	if !j.IsJumping {
		j.CoyoteTimer = j.CoyoteTime
	}

	j.IsGrounded = false
}

// SetDoubleJump enables double jump.
func (j *Jump) SetDoubleJump(enabled bool) {
	if enabled {
		j.MaxJumps = 2
	} else {
		j.MaxJumps = 1
	}

	j.JumpsRemaining = j.MaxJumps
}

// Dash represents dashing abilities.
type Dash struct {
	DashSpeed       float64 // Speed during dash
	DashDuration    float64 // How long dash lasts
	DashCooldown    float64 // Time between dashes
	MaxDashes       int     // Max dashes before recharge
	DashesRemaining int

	// State
	IsDashing     bool
	DashTimer     float64 // Current dash time remaining
	CooldownTimer float64 // Time until next dash available
	DashDirection float64 // Direction of dash (radians)

	// Options
	InvincibleDash   bool // Invincible during dash
	CanDashInAir     bool // Can dash while airborne
	PreserveVelocity bool // Keep velocity after dash ends
}

// NewDash creates a dash component with defaults.
func NewDash(dashSpeed, duration, cooldown float64) Dash {
	return Dash{
		DashSpeed:       dashSpeed,
		DashDuration:    duration,
		DashCooldown:    cooldown,
		MaxDashes:       1,
		DashesRemaining: 1,
		CanDashInAir:    true,
	}
}

// TryDash attempts to perform a dash.
func (d *Dash) TryDash(direction float64) bool {
	if d.IsDashing || d.DashesRemaining <= 0 || d.CooldownTimer > 0 {
		return false
	}

	d.IsDashing = true
	d.DashTimer = d.DashDuration
	d.DashDirection = direction
	d.DashesRemaining--

	return true
}

// EndDash ends the current dash.
func (d *Dash) EndDash() {
	d.IsDashing = false
	d.CooldownTimer = d.DashCooldown
}

// ResetDashes restores all dashes (e.g., on landing).
func (d *Dash) ResetDashes() {
	d.DashesRemaining = d.MaxDashes
}

// IsOnCooldown returns true if dash is on cooldown.
func (d *Dash) IsOnCooldown() bool {
	return d.CooldownTimer > 0
}

// Flight represents flying abilities.
type Flight struct {
	CanFly       bool
	IsFlying     bool
	FlySpeed     float64
	Altitude     float64 // Current height
	MaxAltitude  float64 // Max flying height (0 = unlimited)
	FlyDuration  float64 // Max fly time (0 = unlimited)
	FlightTimer  float64 // Current flight time used
	FuelBased    bool    // If true, uses fuel instead of timer
	Fuel         float64 // Current fuel
	MaxFuel      float64
	FuelDrain    float64 // Fuel per second while flying
	FuelRegen    float64 // Fuel regen per second while grounded
	HoverEnabled bool    // Can hover in place
	IsHovering   bool
}

// NewFlight creates a flight component with unlimited flight.
func NewFlight(flySpeed float64) Flight {
	return Flight{
		CanFly:   true,
		FlySpeed: flySpeed,
	}
}

// NewFlightWithFuel creates a fuel-based flight component.
func NewFlightWithFuel(flySpeed, maxFuel, drain, regen float64) Flight {
	return Flight{
		CanFly:    true,
		FlySpeed:  flySpeed,
		FuelBased: true,
		Fuel:      maxFuel,
		MaxFuel:   maxFuel,
		FuelDrain: drain,
		FuelRegen: regen,
	}
}

// StartFlight begins flying.
func (f *Flight) StartFlight() bool {
	if !f.CanFly {
		return false
	}

	if f.FuelBased && f.Fuel <= 0 {
		return false
	}

	f.IsFlying = true
	f.IsHovering = false

	return true
}

// StopFlight ends flying.
func (f *Flight) StopFlight() {
	f.IsFlying = false
	f.IsHovering = false
}

// Hover toggles hover mode.
func (f *Flight) Hover(enabled bool) {
	if f.HoverEnabled && f.IsFlying {
		f.IsHovering = enabled
	}
}

// GetFuelPercent returns remaining fuel as 0.0-1.0.
func (f *Flight) GetFuelPercent() float64 {
	if !f.FuelBased || f.MaxFuel == 0 {
		return 1.0
	}

	return f.Fuel / f.MaxFuel
}

// PlatformerPhysics contains constants for platformer-style movement.
type PlatformerPhysics struct {
	Gravity        float64 // Downward acceleration
	FrictionGround float64 // Ground friction
	FrictionAir    float64 // Air friction
	MaxFallSpeed   float64 // Terminal velocity

	// Wall interactions
	WallSlideSpeed float64 // Speed limit when sliding on walls
	WallJumpForceX float64 // Horizontal force from wall jump
	WallJumpForceY float64 // Vertical force from wall jump
	CanWallSlide   bool
	CanWallJump    bool

	// State
	IsTouchingWall bool
	WallDirection  int // -1 = left wall, 1 = right wall
	IsWallSliding  bool
}

// NewPlatformerPhysics creates platformer physics with common defaults.
func NewPlatformerPhysics() PlatformerPhysics {
	return PlatformerPhysics{
		Gravity:        980, // Approximate Earth gravity in pixels/s²
		FrictionGround: 0.85,
		FrictionAir:    0.95,
		MaxFallSpeed:   600,
		WallSlideSpeed: 100,
		WallJumpForceX: 300,
		WallJumpForceY: 400,
		CanWallSlide:   false,
		CanWallJump:    false,
	}
}
