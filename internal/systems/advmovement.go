package systems

import (
	"math"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// AdvancedMovementSystem handles walk, run, jump, dash, fly mechanics.
type AdvancedMovementSystem struct {
	// Filters for different entity configurations
	basicFilter    *ecs.Filter2[components.Position, components.Velocity]
	movementFilter *ecs.Filter3[components.Position, components.Velocity, components.Movement]
	jumpFilter     *ecs.Filter3[components.Position, components.Velocity, components.Jump]
	dashFilter     *ecs.Filter3[components.Position, components.Velocity, components.Dash]
	flightFilter   *ecs.Filter3[components.Position, components.Velocity, components.Flight]

	// Global settings
	Gravity         float64
	DefaultFriction float64
}

// NewAdvancedMovementSystem creates an advanced movement system.
func NewAdvancedMovementSystem(world *ecs.World) *AdvancedMovementSystem {
	return &AdvancedMovementSystem{
		basicFilter:     ecs.NewFilter2[components.Position, components.Velocity](world),
		movementFilter:  ecs.NewFilter3[components.Position, components.Velocity, components.Movement](world),
		jumpFilter:      ecs.NewFilter3[components.Position, components.Velocity, components.Jump](world),
		dashFilter:      ecs.NewFilter3[components.Position, components.Velocity, components.Dash](world),
		flightFilter:    ecs.NewFilter3[components.Position, components.Velocity, components.Flight](world),
		Gravity:         980,
		DefaultFriction: 0.9,
	}
}

// Update processes all movement for the frame.
func (s *AdvancedMovementSystem) Update(world *ecs.World, dt float64) {
	s.updateJump(dt)
	s.updateDash(dt)
	s.updateFlight(dt)
	s.updateMovement(dt)
	s.applyVelocity(dt)
}

// updateMovement handles basic movement acceleration/deceleration.
func (s *AdvancedMovementSystem) updateMovement(dt float64) {
	query := s.movementFilter.Query()
	for query.Next() {
		_, vel, mov := query.Get()

		if mov.Frozen || !mov.CanMove {
			continue
		}

		maxSpeed := mov.GetMaxSpeed()

		// Apply acceleration toward current speed
		if mov.IsMoving {
			speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
			if speed < maxSpeed {
				// Accelerate
				accelFactor := mov.Acceleration * dt
				vel.X += math.Cos(mov.Direction) * accelFactor
				vel.Y += math.Sin(mov.Direction) * accelFactor
			}
		} else {
			// Decelerate
			speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
			if speed > 0 {
				decel := mov.Deceleration * dt
				if decel >= speed {
					vel.X = 0
					vel.Y = 0
				} else {
					ratio := (speed - decel) / speed
					vel.X *= ratio
					vel.Y *= ratio
				}
			}
		}

		// Clamp to max speed
		speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
		if speed > maxSpeed {
			ratio := maxSpeed / speed
			vel.X *= ratio
			vel.Y *= ratio
		}

		mov.CurrentSpeed = math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
	}
}

// updateJump handles jumping and gravity.
func (s *AdvancedMovementSystem) updateJump(dt float64) {
	query := s.jumpFilter.Query()
	for query.Next() {
		_, vel, jump := query.Get()

		// Update timers
		if jump.CoyoteTimer > 0 {
			jump.CoyoteTimer -= dt
		}

		if jump.JumpBufferTimer > 0 {
			jump.JumpBufferTimer -= dt
		}

		// Apply gravity if not grounded
		if !jump.IsGrounded {
			gravity := s.Gravity * jump.GravityScale
			jump.VerticalVelocity += gravity * dt

			// Clamp to terminal velocity
			if jump.VerticalVelocity > jump.TerminalVelocity {
				jump.VerticalVelocity = jump.TerminalVelocity
			}

			// Detect falling
			if jump.VerticalVelocity > 0 {
				jump.IsFalling = true
				jump.IsJumping = false
			}
		}

		// Apply vertical velocity
		vel.Y = jump.VerticalVelocity
	}
}

// updateDash handles dash state and movement.
func (s *AdvancedMovementSystem) updateDash(dt float64) {
	query := s.dashFilter.Query()
	for query.Next() {
		_, vel, dash := query.Get()

		// Update cooldown timer
		if dash.CooldownTimer > 0 {
			dash.CooldownTimer -= dt
		}

		// Update dash
		if dash.IsDashing {
			dash.DashTimer -= dt

			if dash.DashTimer <= 0 {
				dash.EndDash()
			} else {
				// Apply dash velocity
				vel.X = math.Cos(dash.DashDirection) * dash.DashSpeed
				vel.Y = math.Sin(dash.DashDirection) * dash.DashSpeed
			}
		}
	}
}

// updateFlight handles flying state and fuel.
func (s *AdvancedMovementSystem) updateFlight(dt float64) {
	query := s.flightFilter.Query()
	for query.Next() {
		_, _, flight := query.Get()

		if flight.IsFlying && flight.FuelBased {
			// Drain fuel
			flight.Fuel -= flight.FuelDrain * dt
			if flight.Fuel <= 0 {
				flight.Fuel = 0
				flight.StopFlight()
			}
		} else if !flight.IsFlying && flight.FuelBased {
			// Regen fuel when not flying
			flight.Fuel += flight.FuelRegen * dt
			if flight.Fuel > flight.MaxFuel {
				flight.Fuel = flight.MaxFuel
			}
		}

		// Track flight time
		if flight.IsFlying && flight.FlyDuration > 0 {
			flight.FlightTimer += dt
			if flight.FlightTimer >= flight.FlyDuration {
				flight.StopFlight()
			}
		}
	}
}

// applyVelocity applies velocity to position for all entities.
func (s *AdvancedMovementSystem) applyVelocity(dt float64) {
	query := s.basicFilter.Query()
	for query.Next() {
		pos, vel := query.Get()
		pos.X += vel.X * dt
		pos.Y += vel.Y * dt
	}
}

// MoveDirection sets movement direction and marks as moving.
func MoveDirection(mov *components.Movement, dirX, dirY float64) {
	if dirX == 0 && dirY == 0 {
		mov.IsMoving = false

		return
	}

	mov.IsMoving = true
	mov.Direction = math.Atan2(dirY, dirX)
	mov.FacingDirection = mov.Direction
}

// MoveToward sets velocity to move toward a target position.
func MoveToward(pos *components.Position, vel *components.Velocity, targetX, targetY, speed float64) {
	dx := targetX - pos.X
	dy := targetY - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 1 {
		vel.X = 0
		vel.Y = 0

		return
	}

	vel.X = (dx / dist) * speed
	vel.Y = (dy / dist) * speed
}

// DistanceTo returns the distance between position and a target.
func DistanceTo(pos *components.Position, targetX, targetY float64) float64 {
	dx := targetX - pos.X
	dy := targetY - pos.Y

	return math.Sqrt(dx*dx + dy*dy)
}

// ApplyKnockback applies a knockback force to velocity.
func ApplyKnockback(vel *components.Velocity, fromX, fromY, toX, toY, force float64) {
	dx := toX - fromX
	dy := toY - fromY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist == 0 {
		return
	}

	vel.X += (dx / dist) * force
	vel.Y += (dy / dist) * force
}

// NormalizeVelocity limits velocity to max speed.
func NormalizeVelocity(vel *components.Velocity, maxSpeed float64) {
	speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
	if speed > maxSpeed {
		ratio := maxSpeed / speed
		vel.X *= ratio
		vel.Y *= ratio
	}
}

// ApplyFriction reduces velocity by a friction factor.
func ApplyFriction(vel *components.Velocity, friction float64) {
	vel.X *= friction
	vel.Y *= friction
}

// TopDownMovementSystem provides simplified 8-direction movement.
type TopDownMovementSystem struct {
	filter *ecs.Filter3[components.Position, components.Velocity, components.Movement]
}

// NewTopDownMovementSystem creates a top-down movement system.
func NewTopDownMovementSystem(world *ecs.World) *TopDownMovementSystem {
	return &TopDownMovementSystem{
		filter: ecs.NewFilter3[components.Position, components.Velocity, components.Movement](world),
	}
}

// Update processes top-down movement with input vector.
func (s *TopDownMovementSystem) Update(inputX, inputY, dt float64) {
	query := s.filter.Query()
	for query.Next() {
		pos, vel, mov := query.Get()

		if mov.Frozen || !mov.CanMove {
			vel.X = 0
			vel.Y = 0

			continue
		}

		// Normalize diagonal movement
		length := math.Sqrt(inputX*inputX + inputY*inputY)
		if length > 1 {
			inputX /= length
			inputY /= length
		}

		maxSpeed := mov.GetMaxSpeed()

		if length > 0 {
			mov.IsMoving = true
			mov.Direction = math.Atan2(inputY, inputX)
			mov.FacingDirection = mov.Direction

			// Set velocity
			vel.X = inputX * maxSpeed
			vel.Y = inputY * maxSpeed
		} else {
			mov.IsMoving = false

			// Apply friction/deceleration
			vel.X *= 0.8
			vel.Y *= 0.8

			// Stop if very slow
			if math.Abs(vel.X) < 1 {
				vel.X = 0
			}

			if math.Abs(vel.Y) < 1 {
				vel.Y = 0
			}
		}

		// Apply velocity
		pos.X += vel.X * dt
		pos.Y += vel.Y * dt

		mov.CurrentSpeed = math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
	}
}
