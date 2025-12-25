package ai

import (
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// Action represents a game action that can be executed by an AI agent.
type Action interface {
	// Execute performs the action on the world.
	Execute(world *ecs.World) error
	// Name returns a human-readable action name.
	Name() string
}

// ActionExecutor handles execution of AI actions.
type ActionExecutor struct {
	world  *ecs.World
	posMap *ecs.Map[components.Position]
	velMap *ecs.Map[components.Velocity]
}

// NewActionExecutor creates an action executor for the given world.
func NewActionExecutor(world *ecs.World) *ActionExecutor {
	return &ActionExecutor{
		world:  world,
		posMap: ecs.NewMap[components.Position](world),
		velMap: ecs.NewMap[components.Velocity](world),
	}
}

// Execute runs a single action.
func (e *ActionExecutor) Execute(action Action) error {
	return action.Execute(e.world)
}

// ExecuteBatch runs multiple actions, returning any errors.
func (e *ActionExecutor) ExecuteBatch(actions []Action) []error {
	errors := make([]error, 0)

	for _, action := range actions {
		if err := action.Execute(e.world); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// ============================================================================
// Built-in Actions
// ============================================================================

// MoveAction moves an entity by a delta.
type MoveAction struct {
	Entity ecs.Entity
	DX, DY float64
}

func (a MoveAction) Name() string { return "move" }

func (a MoveAction) Execute(world *ecs.World) error {
	posMap := ecs.NewMap[components.Position](world)
	if !posMap.Has(a.Entity) {
		return nil // Entity doesn't have position, skip
	}

	pos := posMap.Get(a.Entity)
	pos.X += a.DX
	pos.Y += a.DY

	return nil
}

// SetPositionAction sets an entity's absolute position.
type SetPositionAction struct {
	Entity ecs.Entity
	X, Y   float64
}

func (a SetPositionAction) Name() string { return "set_position" }

func (a SetPositionAction) Execute(world *ecs.World) error {
	posMap := ecs.NewMap[components.Position](world)
	if !posMap.Has(a.Entity) {
		return nil
	}

	pos := posMap.Get(a.Entity)
	pos.X = a.X
	pos.Y = a.Y

	return nil
}

// SetVelocityAction sets an entity's velocity.
type SetVelocityAction struct {
	Entity ecs.Entity
	VX, VY float64
}

func (a SetVelocityAction) Name() string { return "set_velocity" }

func (a SetVelocityAction) Execute(world *ecs.World) error {
	velMap := ecs.NewMap[components.Velocity](world)
	if !velMap.Has(a.Entity) {
		return nil
	}

	vel := velMap.Get(a.Entity)
	vel.X = a.VX
	vel.Y = a.VY

	return nil
}

// RemoveEntityAction removes an entity from the world.
type RemoveEntityAction struct {
	Entity ecs.Entity
}

func (a RemoveEntityAction) Name() string { return "remove" }

func (a RemoveEntityAction) Execute(world *ecs.World) error {
	if world.Alive(a.Entity) {
		world.RemoveEntity(a.Entity)
	}

	return nil
}

// SetStateAction updates an entity's AIMetadata visual state.
type SetStateAction struct {
	Entity ecs.Entity
	State  string
}

func (a SetStateAction) Name() string { return "set_state" }

func (a SetStateAction) Execute(world *ecs.World) error {
	metaMap := ecs.NewMap[components.AIMetadata](world)
	if !metaMap.Has(a.Entity) {
		return nil
	}

	meta := metaMap.Get(a.Entity)
	meta.SetState(a.State)

	return nil
}
