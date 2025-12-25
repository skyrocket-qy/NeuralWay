package engine

import (
	"fmt"
	"reflect"

	"github.com/mlange-42/ark/ecs"
)

// ComponentRequirement defines a required component for a system.
type ComponentRequirement struct {
	Name     string
	Required bool
}

// ComponentValidator validates that entities have required components.
type ComponentValidator struct {
	rules   map[string][]reflect.Type
	world   *ecs.World
	hasFunc map[reflect.Type]func(ecs.Entity) bool
}

// NewComponentValidator creates a new component validator.
func NewComponentValidator(world *ecs.World) *ComponentValidator {
	return &ComponentValidator{
		rules:   make(map[string][]reflect.Type),
		world:   world,
		hasFunc: make(map[reflect.Type]func(ecs.Entity) bool),
	}
}

// RegisterHasFunc registers a function to check if an entity has a component.
// This is needed because Go generics can't be used dynamically.
func (v *ComponentValidator) RegisterHasFunc(compType reflect.Type, hasFunc func(ecs.Entity) bool) {
	v.hasFunc[compType] = hasFunc
}

// RequireComponents registers required components for a system.
func (v *ComponentValidator) RequireComponents(systemName string, compTypes ...reflect.Type) {
	v.rules[systemName] = append(v.rules[systemName], compTypes...)
}

// ValidateEntity checks if an entity has all required components for a system.
func (v *ComponentValidator) ValidateEntity(systemName string, entity ecs.Entity) []error {
	requirements, ok := v.rules[systemName]
	if !ok {
		return nil
	}

	var errors []error

	for _, compType := range requirements {
		hasFunc, ok := v.hasFunc[compType]
		if !ok {
			continue // No check function registered, skip
		}

		if !hasFunc(entity) {
			errors = append(errors, fmt.Errorf("entity %v missing required component %s for system %s",
				entity, compType.Name(), systemName))
		}
	}

	return errors
}

// ValidateAll checks all registered rules and returns any violations.
func (v *ComponentValidator) ValidateAll(entity ecs.Entity) map[string][]error {
	result := make(map[string][]error)

	for systemName := range v.rules {
		if errs := v.ValidateEntity(systemName, entity); len(errs) > 0 {
			result[systemName] = errs
		}
	}

	return result
}

// GetRules returns all registered rules.
func (v *ComponentValidator) GetRules() map[string][]reflect.Type {
	return v.rules
}

// ClearRules removes all rules.
func (v *ComponentValidator) ClearRules() {
	v.rules = make(map[string][]reflect.Type)
}

// SystemContract documents a system's component dependencies.
type SystemContract struct {
	Name   string
	Reads  []string // Component names the system reads
	Writes []string // Component names the system writes
}

// ContractRegistry tracks all system contracts.
type ContractRegistry struct {
	contracts map[string]SystemContract
}

// NewContractRegistry creates a new contract registry.
func NewContractRegistry() *ContractRegistry {
	return &ContractRegistry{
		contracts: make(map[string]SystemContract),
	}
}

// Register adds a system contract.
func (r *ContractRegistry) Register(contract SystemContract) {
	r.contracts[contract.Name] = contract
}

// Get returns a system contract.
func (r *ContractRegistry) Get(name string) (SystemContract, bool) {
	c, ok := r.contracts[name]

	return c, ok
}

// All returns all contracts.
func (r *ContractRegistry) All() map[string]SystemContract {
	return r.contracts
}

// FindConflicts detects systems that write to the same components.
func (r *ContractRegistry) FindConflicts() map[string][]string {
	// Map component -> systems that write to it
	writers := make(map[string][]string)

	for name, contract := range r.contracts {
		for _, comp := range contract.Writes {
			writers[comp] = append(writers[comp], name)
		}
	}

	// Find conflicts (multiple writers)
	conflicts := make(map[string][]string)

	for comp, systems := range writers {
		if len(systems) > 1 {
			conflicts[comp] = systems
		}
	}

	return conflicts
}
